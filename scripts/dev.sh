#!/usr/bin/env bash
# scripts/dev.sh -- build/vet/test/cov/scan/image tasks for meshctl.
#
# Mirror of scripts/dev.ps1 (behaviourally identical). The USER runs this; CI
# (.github/workflows/ci.yml and tag.yml) calls task names only, so local == CI
# structurally. Works from any cwd. Accepts multiple tasks per invocation and
# stops at the first failure.
#
#   ./scripts/dev.sh build vet test
#   ./scripts/dev.sh all          # build vet test                      (CI runs: all scan)
#   ./scripts/dev.sh full         # all + cov image scan graphify       (pre-tag sweep)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
LOG_DIR="${SCRIPT_DIR}/logs"
DIST_DIR="${REPO_ROOT}/dist"
COV_DIR="${REPO_ROOT}/coverage"
# The binary is always meshctl, never the directory name: the repo is called
# vpn-mesh-utility, and the release merges every leg by file name.
BIN_NAME="meshctl"
IMAGE_REPO="${IMAGE_REPO:-vpn-mesh-utility}"

# Version stamp. The release workflow exports VERSION from the tag ref, so CI
# never depends on git describe seeing history through a shallow checkout;
# locally git describe is used, then "dev" when git or a repo is unavailable.
# The value reaches -ldflags and an image tag, so it is held to a strict
# charset here rather than trusted.
if [[ -z "${VERSION:-}" ]]; then
  VERSION="$(git -C "${REPO_ROOT}" describe --tags --always --dirty 2>/dev/null || true)"
  [[ -z "${VERSION}" ]] && VERSION="dev"
fi
if [[ ! "${VERSION}" =~ ^[A-Za-z0-9][A-Za-z0-9._+-]{0,63}$ ]]; then
  printf 'error: VERSION %q has characters outside [A-Za-z0-9._+-]; export a plain tag such as v1.2.3\n' "${VERSION}" >&2
  exit 1
fi
# Images drop the leading v (registry convention); '+' is not legal in a tag.
IMAGE_TAG="${VERSION#v}"
IMAGE_TAG="${IMAGE_TAG//+/-}"

# Local convenience cross-compile set (`dist`). CI does not use it: tag.yml
# matrixes over `build` with TARGET_OS/TARGET_ARCH from BUILD_TARGETS instead.
DIST_TARGETS=(linux/amd64 linux/arm64 windows/amd64)

# Packages whose tests own golden files (`regen` rewrites them with -update).
REGEN_PKGS=(./internal/server/planner ./internal/server/phone ./internal/server/ui ./internal/agent/firewall ./internal/agent/wg ./internal/agent/localui ./deploy ./cmd/meshctl)

# -race needs cgo + a C compiler; on by default here for `test`, disable with
# MESHCTL_RACE=0 (dev.ps1 defaults it off -- the windows runner's cgo is the
# slow, fragile leg). `cov` never races: `test` is the race run.
RACE_FLAG=()
if [[ "${MESHCTL_RACE:-1}" != "0" ]]; then RACE_FLAG=(-race); fi
# A race-built binary sleeps atexit_sleep_ms (default 1000) before every clean
# exit. Default it to 0 only when unset, so an exported value wins.
if [[ ${#RACE_FLAG[@]} -gt 0 && -z "${GORACE:-}" ]]; then export GORACE="atexit_sleep_ms=0"; fi

# Toolchain parity: go.mod's `toolchain` pin is what local and CI must agree
# on. Apply it only when GOTOOLCHAIN is unset, so an explicit value wins.
if [[ -z "${GOTOOLCHAIN:-}" && -f "${REPO_ROOT}/go.mod" ]]; then
  _tc="$(awk '$1=="toolchain"{print $2; exit}' "${REPO_ROOT}/go.mod")"
  [[ -n "${_tc}" ]] && export GOTOOLCHAIN="${_tc}"
  unset _tc
fi

# Keep captured logs clean: no color, no progress spinners from tools.
export NO_COLOR=1

# The previous coverage total is the floor (S5). Read it before any task runs,
# because `cov` truncates its own log when it starts.
PREV_COV="$(awk '/^total:/{v=$NF} END{print v}' "${LOG_DIR}/cov.log" 2>/dev/null || true)"

# --- pretty helpers (console only; logs are stripped of ANSI) -----------------
if [[ -t 1 ]]; then
  B=$'\033[1m'; GRN=$'\033[32m'; YLW=$'\033[33m'; RED=$'\033[31m'; CYN=$'\033[36m'; RST=$'\033[0m'
else
  B=""; GRN=""; YLW=""; RED=""; CYN=""; RST=""
fi
step() { printf '%s==>%s %s\n' "${CYN}${B}" "${RST}" "$*" >&2; }
ok()   { printf '%s[ ok ]%s %s\n' "${GRN}" "${RST}" "$*" >&2; }
warn() { printf '%s[warn]%s %s\n' "${YLW}" "${RST}" "$*" >&2; }
die()  { printf '%s[fail]%s %s\n' "${RED}${B}" "${RST}" "$*" >&2; exit 1; }

now() { date +%Y-%m-%dT%H:%M:%S%z; }

LOGFILE=""
log_init() {
  mkdir -p "${LOG_DIR}"
  LOGFILE="${LOG_DIR}/$1.log"
  # Replaced, not truncated: a log a root shell left behind is not ours to write.
  rm -f "${LOGFILE}"
  printf '=== %s | %s ===\n' "$(now)" "$1" > "${LOGFILE}"
}
# note <msg> -- a plain line on the console and in the task log.
note() { printf '%s\n' "$*" | tee -a "${LOGFILE}"; }
# skip <msg> -- a task that does not run says why in its log too, so an OK
# footer never stands for a gate that did not run without a reason beside it.
skip() { warn "$*"; printf '[warn] %s\n' "$*" >> "${LOGFILE}"; }
# cap runs a command, streaming combined stdout+stderr to the console and
# appending an ANSI-stripped copy to LOGFILE. Returns the command's exit code.
# One synchronous pipeline, so the footer can never overtake the task's output.
cap() {
  "$@" 2>&1 | sed -E $'s/\x1b\\[[0-9;?]*[a-zA-Z]//g' | tee -a "${LOGFILE}"
  return "${PIPESTATUS[0]}"
}
# finish <task> <exit-code> <elapsed-seconds> -- contract footer, log + console.
finish() {
  local status="OK"
  (( $2 != 0 )) && status="FAILED (exit $2)"
  printf '%s | %s | %ss | %s\n' "$(now)" "$1" "$3" "${status}" | tee -a "${LOG_DIR}/$1.log"
}

host_os()   { go env GOHOSTOS; }
host_arch() { go env GOHOSTARCH; }

# Docker can only build this project's Linux image when the daemon runs Linux
# containers. GitHub's windows runners are Windows-container only, so the image
# half of `image`/`scan` warn-skips there. Same rule in both scripts.
has_linux_docker() {
  command -v docker >/dev/null 2>&1 || { DOCKER_WHY="docker not on PATH"; return 1; }
  local t
  t="$(docker info --format '{{.OSType}}' 2>&1)"
  [[ "${t}" == "linux" ]] && return 0
  DOCKER_WHY="docker info: ${t}"
  return 1
}

build_image() {
  cap docker build --progress=plain --build-arg "VERSION=${VERSION}" \
    -t "${IMAGE_REPO}:${IMAGE_TAG}" "${REPO_ROOT}"
}

# trivy natively when it is on PATH, otherwise the container. --pull=always
# matters more than the tag: without it docker reuses whatever :latest resolved
# to months ago -- unpinned AND stale. The docker.sock mount lets the
# containerised trivy see the image just built locally.
trivy_image() {
  if command -v trivy >/dev/null 2>&1; then
    cap trivy image --quiet --exit-code 1 --severity HIGH,CRITICAL --ignore-unfixed "$1"
  else
    cap docker run --rm --pull=always -e NO_COLOR=1 \
      -v /var/run/docker.sock:/var/run/docker.sock \
      aquasec/trivy:latest image --quiet --exit-code 1 \
      --severity HIGH,CRITICAL --ignore-unfixed "$1"
  fi
}

# --- tasks --------------------------------------------------------------------
# Tasks return non-zero on failure; the dispatcher writes the footer and stops.

# build_one <os> <arch> -- compile meshctl for one target into dist/. The target
# is in the file name because the release job merges every leg with
# merge-multiple: identical names would silently overwrite.
build_one() {
  local os=$1 arch=$2 out="${DIST_DIR}/${BIN_NAME}-$1-$2"
  [[ "${os}" == "windows" ]] && out="${out}.exe"
  mkdir -p "${DIST_DIR}"
  step "  ${os}/${arch} -> ${out}"
  CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" \
    cap go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o "${out}" ./cmd/meshctl
}

# CI sets TARGET_OS/TARGET_ARCH (from the BUILD_TARGETS repo variable); unset
# means build for the host, so local and CI output are the same shape.
task_build() {
  build_one "${TARGET_OS:-$(host_os)}" "${TARGET_ARCH:-$(host_arch)}"
}

task_dist() {
  local t
  for t in "${DIST_TARGETS[@]}"; do
    build_one "${t%/*}" "${t#*/}" || return 1
  done
  ok "binaries in ${DIST_DIR}"
}

# vet: go vet for the host AND the other shipped OS (the platform code is split
# by build tags, so a host-only vet never compiles the other half), staticcheck
# for the host (each CI leg covers its own OS), then gofmt. Every check runs and
# reports; the task fails if any did.
task_vet() {
  local rc=0 other=windows
  [[ "$(host_os)" == "windows" ]] && other=linux
  cap go vet ./... || rc=1
  GOOS="${other}" cap go vet ./... || rc=1
  cap go tool staticcheck ./... || rc=1
  local unformatted
  unformatted="$(gofmt -l "${REPO_ROOT}/cmd" "${REPO_ROOT}/internal" "${REPO_ROOT}/deploy" "${REPO_ROOT}/ui" "${REPO_ROOT}/test" 2>/dev/null || true)"
  if [[ -n "${unformatted}" ]]; then
    note "gofmt: these files need formatting (run gofmt -w on them):"
    note "${unformatted}"
    rc=1
  fi
  return "${rc}"
}

# -timeout bounds each package's test binary: a deadlocked test fails in
# minutes instead of hanging for go test's default 10.
TEST_TIMEOUT="3m"
task_test() { cap go test "${RACE_FLAG[@]}" -count=1 -timeout="${TEST_TIMEOUT}" ./...; }

task_cov() {
  mkdir -p "${COV_DIR}"
  local prof="${COV_DIR}/coverage.out" html="${COV_DIR}/coverage.html"
  # -count=1 forces a real run so a cached result cannot report a stale total.
  cap go test -covermode=atomic -coverprofile="${prof}" -count=1 -timeout="${TEST_TIMEOUT}" ./... || return 1
  cap go tool cover -html="${prof}" -o "${html}" || return 1
  local total; total="$(go tool cover -func="${prof}" | tail -n1)"
  note "${total}"
  local now_pct="${total##*[[:space:]]}"
  if [[ -z "${PREV_COV}" ]]; then
    note "coverage floor: no previous total in logs/cov.log; this run is the baseline"
  elif awk -v a="${now_pct%\%}" -v b="${PREV_COV%\%}" 'BEGIN{exit !(a+0 < b+0)}'; then
    warn "coverage DROPPED: ${PREV_COV} -> ${now_pct}; add tests or state in the change why it moved (S5)"
    note "coverage floor: previous ${PREV_COV}, now ${now_pct} (DROPPED)"
  else
    note "coverage floor: previous ${PREV_COV}, now ${now_pct}"
  fi
  ok "coverage -> ${html}  (${now_pct})"
}

# scan: every applicable check, FATAL on a fixable CVE.
#  1. govulncheck through the go.mod tool directive (`go tool`, never
#     `go run pkg@version`, which ignores the toolchain pin) in JSON mode, judged
#     by internal/tools/vulnjudge: fail on a fixable called finding, warn on one
#     with no released fix.
#  2. trivy over a freshly built image (never a stale one).
task_scan() {
  local raw="${LOG_DIR}/scan.json" code=0
  # `-format json` exits 0 even with findings, so non-zero is a real tool
  # failure (bad flags, packages that will not load) and stays fatal.
  if ! go tool govulncheck -format json ./... >"${raw}"; then
    warn "govulncheck failed to run; see the output above"
    return 1
  fi
  cap go run ./internal/tools/vulnjudge "${raw}" || return $?
  if [[ ! -f "${REPO_ROOT}/Dockerfile" ]]; then
    note "no Dockerfile; skipping the image scan"
    return 0
  fi
  if ! has_linux_docker; then
    skip "no linux-container docker daemon (${DOCKER_WHY}); skipping the image scan"
    return 0
  fi
  build_image || return $?
  trivy_image "${IMAGE_REPO}:${IMAGE_TAG}" || code=$?
  return "${code}"
}

task_image() {
  [[ -f "${REPO_ROOT}/Dockerfile" ]] || { skip "no Dockerfile; skipping image"; return 0; }
  has_linux_docker || { skip "no linux-container docker daemon (${DOCKER_WHY}); skipping image"; return 0; }
  build_image
}

# up/down: the local single-node dev stack (server only, bridge network).
COMPOSE_DEV="${REPO_ROOT}/deploy/compose.dev.yml"
task_up() {
  [[ -f "${COMPOSE_DEV}" ]] || { skip "no ${COMPOSE_DEV}; nothing to bring up"; return 0; }
  IMAGE="${IMAGE_REPO}:${IMAGE_TAG}" cap docker compose -f "${COMPOSE_DEV}" up -d
}
task_down() {
  [[ -f "${COMPOSE_DEV}" ]] || { skip "no ${COMPOSE_DEV}; nothing to bring down"; return 0; }
  IMAGE="${IMAGE_REPO}:${IMAGE_TAG}" cap docker compose -f "${COMPOSE_DEV}" down
}

# regen rewrites the committed goldens from the code that generates them.
# Deliberately NOT in all/full: a gate that rewrote what it compares against
# could never fail. Review the diff before committing it.
task_regen() {
  local rc=0 pkg code
  for pkg in "${REGEN_PKGS[@]}"; do
    cap go test "${pkg}" -count=1 -update
    code=$?
    if (( code != 0 && rc == 0 )); then rc=${code}; fi
  done
  return "${rc}"
}

# e2e: network-namespace tests against real kernel WireGuard (test/e2e, build
# tag e2e). Needs root, so the test binary is compiled as you and only run under
# `sudo -n` -- run `sudo -v` first. Local only, outside every aggregate.
# Under sudo itself it refuses: root's secure_path has no go, and the build
# would leave root-owned files in dist/ and logs/.
task_e2e() {
  if (( EUID == 0 )); then
    note "e2e must not run as root (root's PATH has no go, and the build would leave root-owned files):"
    note "leave the root shell and run './scripts/dev.sh e2e' as your own user; it uses 'sudo -n' itself"
    return 1
  fi
  [[ "$(host_os)" == "linux" ]] || { skip "e2e runs on Linux only; skipping"; return 0; }
  [[ -d "${REPO_ROOT}/test/e2e" ]] || { skip "no test/e2e yet; skipping"; return 0; }
  if [[ ! -d /sys/module/wireguard ]] && ! modinfo wireguard >/dev/null 2>&1; then
    skip "no wireguard kernel module on this host; skipping"; return 0
  fi
  if ! sudo -n true 2>/dev/null; then
    note "e2e needs root: run 'sudo -v' first, then re-run this task"
    return 1
  fi
  mkdir -p "${DIST_DIR}"
  cap go test -c -tags e2e -o "${DIST_DIR}/e2e.test" ./test/e2e || return 1
  CGO_ENABLED=0 cap go build -trimpath -ldflags "-X main.version=${VERSION}" -o "${DIST_DIR}/meshctl-e2e" ./cmd/meshctl || return 1
  cap sudo -n "${DIST_DIR}/e2e.test" -test.v -test.count=1 -test.timeout=15m -meshctl="${DIST_DIR}/meshctl-e2e"
}

# Local only: the graph is a developer artifact, not a CI output.
task_graphify() {
  [[ -n "${CI:-}" ]] && { skip "graphify is local-only; skipping in CI"; return 0; }
  command -v graphify >/dev/null 2>&1 || { skip "graphify not on PATH; skipping"; return 0; }
  cap graphify update .
}

# --- dispatch -------------------------------------------------------------------
ALL="build vet test"
FULL="build vet test cov image scan graphify"

usage() {
  cat >&2 <<EOF
${B}dev.sh${RST} -- build/test/scan tooling for meshctl

Usage: ${0##*/} <task> [task...]

Tasks:
  build    compile -> dist/${BIN_NAME}-<os>-<arch>[.exe]; TARGET_OS/TARGET_ARCH pick
           the target, unset means host; stamps \`meshctl version\` from VERSION,
           else git describe, else "dev"
  vet      go vet (host + cross OS), staticcheck (host), gofmt -l
  test     go test ${RACE_FLAG[*]:-} -count=1 ./...
  cov      coverage profile -> coverage/coverage.html + printed total, compared
           with the previous total in logs/cov.log (the floor)
  scan     govulncheck (fatal on a fixable called vulnerability; no released
           fix warns) + trivy over a freshly built image
  image    docker build -> ${IMAGE_REPO}:<version>
  up/down  local dev stack (deploy/compose.dev.yml)
  dist     cross-compile ${DIST_TARGETS[*]}
  regen    rewrite goldens in ${REGEN_PKGS[*]} (not in all/full)
  e2e      root-only netns tests against kernel WireGuard (sudo -v first; not in all/full)
  graphify refresh graphify-out/ (local only; skipped when CI is set)
  all      ${ALL}   (what CI runs, as: all scan)
  full     ${FULL}   (pre-tag sweep)

Env: VERSION overrides the stamp; MESHCTL_RACE=0 disables -race on \`test\`;
     GORACE defaults to atexit_sleep_ms=0 on race runs; GOTOOLCHAIN defaults to
     go.mod's \`toolchain\` pin; IMAGE_REPO overrides the local image name.
Logs: ${LOG_DIR}/<task>.log (each run closes with a timestamped footer)
EOF
}

main() {
  cd "${REPO_ROOT}"
  [[ $# -eq 0 ]] && { usage; exit 0; }
  case "$1" in -h|--help|help) usage; exit 0;; esac

  local tasks="" a t start code
  for a in "$@"; do
    case "$a" in
      all)  tasks+=" ${ALL}" ;;
      full) tasks+=" ${FULL}" ;;
      *)    tasks+=" $a" ;;
    esac
  done

  for t in ${tasks}; do
    declare -F "task_${t}" >/dev/null || die "unknown task: ${t} (try: ${0##*/} help)"
    step "${t}"
    log_init "${t}"
    start=${SECONDS}
    code=0
    "task_${t}" || code=$?
    finish "${t}" "${code}" "$((SECONDS - start))"
    if (( code != 0 )); then
      warn "${t} failed; stopping"
      exit 1
    fi
    ok "${t}"
  done
  ok "done: $*"
}

main "$@"
