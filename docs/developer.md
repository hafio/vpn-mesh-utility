# Developer guide

## Contents

- [Toolchain](#toolchain)
- [Dev script tasks](#dev-script-tasks)
- [Goldens](#goldens)
- [Releases](#releases)
- [Extending the code](#extending-the-code)
  - [Add a validator](#add-a-validator)
  - [Add a topology rule](#add-a-topology-rule)
  - [Add an agent API route](#add-an-agent-api-route)
  - [Add an apply step](#add-an-apply-step)
  - [Add a command](#add-a-command)
  - [Add a web UI page](#add-a-web-ui-page)
  - [Add a database migration](#add-a-database-migration)

## Toolchain

Go is pinned exactly in `go.mod`: `go 1.27` is the language floor and `toolchain go1.27.3` is
what every build uses. Both dev scripts export `GOTOOLCHAIN` from that line when it is unset,
so a laptop and both CI runners compile with the same stdlib. The scripts contain no other
version logic. To bump Go: edit the `toolchain` line, the `golang:` base image in the
Dockerfile (it builds with `GOTOOLCHAIN=local` and fails loudly on a mismatch), and re-run the
gates.

Tools the tasks invoke are `tool` directives in `go.mod` and run as `go tool <name>`, never
`go run pkg@version` (which ignores go.mod and its toolchain pin):

```
go get -tool golang.org/x/vuln/cmd/govulncheck@v1.9.0
go get -tool honnef.co/go/tools/cmd/staticcheck@v0.8.1
go mod tidy
```

Library dependencies are pinned the same way, by `go.mod` and `go.sum`: `modernc.org/sqlite`
(server store), `golang.org/x/crypto` (argon2) and `github.com/pquerna/otp` (TOTP) for the
server's auth, `golang.org/x/term` (CLI password prompts), `golang.zx2c4.com/wireguard/wgctrl`,
`golang.zx2c4.com/wireguard` (wireguard-go, the userspace backend) and
`github.com/vishvananda/netlink` (agent, Linux), `go.yaml.in/yaml/v3` (export, and the
tests that read compose files and docs/containers.md's samples; the maintained continuation of
`gopkg.in/yaml.v3`). Code both sides need that is not part
of the wire contract (atomic file writes) lives in the stdlib-only `internal/fsutil`.

## Dev script tasks

`scripts/dev.sh` and `scripts/dev.ps1` are behaviourally identical and are the only place
that knows how to build, test or scan. Each task logs to `scripts/logs/<task>.log` and ends
with a `<timestamp> | <task> | <seconds>s | OK|FAILED (exit N)` footer. The log holds all the
task printed: every tool's stdout and stderr, and the scripts' own `==>`, `[ ok ]`, `[warn]` and
`[fail]` lines (dev.ps1 adds where a PowerShell error stopped the task). Data a tool writes on
stdout is kept beside it: `scan.json` (govulncheck), `vet.gofmt.txt` (`gofmt -l`) and
`cov.func.txt` (coverage per function). Several tasks may be given at once; the run stops at
the first failure.

`.gitattributes` keeps every text file LF on Windows too: goldens, embedded templates and
`gofmt -l` compare bytes, so a checkout with CRLF would fail `vet` and `test`. A Windows
checkout made before it existed keeps its CRLF files until they are checked out again: with no
local changes, `git rm -r --cached -q . && git reset --hard`.

| Task | What it does |
| --- | --- |
| `build` | `dist/meshctl-<os>-<arch>[.exe]` for `TARGET_OS`/`TARGET_ARCH` (unset: the host), `CGO_ENABLED=0`, `-trimpath`, version from `VERSION`, else `git describe`, else `dev` |
| `vet` | `go vet` for the host and the other shipped OS, `go tool staticcheck` for the host, `gofmt -l` |
| `test` | `go test -count=1 -timeout=3m ./...`, with `-race` by default in dev.sh (`MESHCTL_RACE=0` turns it off; dev.ps1 defaults it off, `MESHCTL_RACE=1` turns it on). The timeout is per package, so a deadlocked test fails in minutes |
| `cov` | coverage profile to `coverage/` (same 3 min per-package timeout), printed total, compared with the previous total in `scripts/logs/cov.log` (the floor) |
| `scan` | `go tool govulncheck -format json` judged by `internal/tools/vulnjudge` (fails on a fixable vulnerability the code calls, warns on one with no fix), then trivy over a freshly built image |
| `image` | `docker build` of the root Dockerfile as `vpn-mesh-utility:<version>` |
| `up` / `down` | the local dev stack in `deploy/compose.dev.yml` |
| `dist` | cross-compile linux/amd64, linux/arm64, windows/amd64 (local convenience) |
| `regen` | rewrite golden files with `-update` in planner, the phone profiles, the web UI pages, the firewall argv, the userspace backend's UAPI text, the local UI's dashboard, the mesh firewall's table, deploy (install.sh, install.ps1) and cmd/meshctl (docs/commands.md, and every doc's table of contents); not in any aggregate |
| `e2e` | builds `dist/e2e.test` and `dist/meshctl-e2e` as you, then runs the network-namespace tests against kernel WireGuard under `sudo -n` (Linux; run `sudo -v` first, then the task as yourself -- under `sudo` it fails, as root's PATH has no `go`; not in any aggregate) |
| `graphify` | refresh the knowledge graph (local only) |
| `all` | `build vet test` -- CI runs `all scan` |
| `full` | `all` + `cov image scan graphify` -- the pre-tag sweep |

The image half of `image` and `scan` is skipped with a warning when no Linux-container docker
daemon is reachable (the Windows CI runner, or a user without docker access). A skipped task
still ends `OK`, so the reason is written to its log too, as a `[warn]` line above the footer.

## Goldens

Golden files live in each package's `testdata/` (and `docs/commands.md`) and are reviewed by
hand: the planner goldens were derived from the topology rules in
`internal/server/planner/planner.go`, not captured from the code. `docs/commands.md` is the
exception that is purely generated: after changing a command or a flag, `regen` rewrites it.
Every doc (README.md and docs/*.md) opens with a `## Contents` section listing its `##` and
`###` headings; after adding, renaming or removing a heading, `regen` rewrites it. When a test reports a stale golden and the new output is what you intended, run
`./scripts/dev.sh regen` and review the diff before committing it. `regen` stays out of
`all`/`full` because a gate that rewrote its own expectations could never fail.

## Releases

The GitHub tag is the only version source. Pushing `v1.2.3` runs `.github/workflows/tag.yml`:
the gates on ubuntu and windows (`ci.yml`), a multi-arch image pushed to
`ghcr.io/hafio/vpn-mesh-utility:1.2.3`, the binaries for every `BUILD_TARGETS` entry, then a
GitHub release with `SHA256SUMS.txt`. Nothing runs on an ordinary push or pull request (except
pull requests that touch CI files), so run `./scripts/dev.sh full` on a clean checkout before
tagging.

`tag.yml` is the init-devops template with one approved divergence: the image build passes
`VERSION=<tag>` as a build argument, because the image cross-compiles the binaries it serves at
`/dl/*` and its build context has no git history.

Repository variable (Settings > Secrets and variables > Actions > Variables):

```
BUILD_TARGETS = [{"os":"linux","arch":"amd64"},{"os":"linux","arch":"arm64"},{"os":"windows","arch":"amd64"}]
```

## Extending the code

### Add a validator

Every value from outside the process goes through `internal/validate`. Add a function there
that accepts a positive grammar (a regexp or a parser), returns the normalised value, and fails
with `fail(field, value, reason)` -- masking the value with `mask()` if it is a secret. Add the
accept and reject cases to `validate_test.go`, including the shared `hostile` payloads.

### Add a topology rule

The planner (`internal/server/planner`) is pure: it takes `Input` and returns per-node
`proto.NodeConfig`. A rule change is three edits: the rule in `planner.go` (keep every emitted
slice sorted), any new cross-node constraint in `check.go` (error if it must block a change,
warning if the plan still applies), and the goldens. Re-derive the expected output by hand,
update `testdata/*.golden.json`, and make sure `TestEveryPairCanTalk` and
`TestComputedConfigsPassAgentValidation` still pass -- they are the properties every topology
must keep.

### Add an agent API route

Add the handler in `internal/server/api`, register it in `routes()` (or `adminRoutes()`) with its
auth kind, and list it in [agent-protocol.md](agent-protocol.md) as `` `METHOD path` `` --
`TestRoutesAreDocumented` fails until you do. Signed handlers start with `a.authenticate`, which
does every cheap check before the database; writes to desired state go through
`a.Reconciler.Apply`, never straight to the store. Payload types shared with the agent go in
`internal/proto` with a `Validate` method.

### Add an apply step

Steps live in `internal/agent/apply/apply.go` in their fixed order. A step compares desired with
observed state and returns how many changes it made, so `TestApplyIsIdempotent` keeps holding;
host access goes through an interface the engine's `Deps` carry, with a fake in
`internal/agent/agentfake`. Decide whether a failure is fatal (the config is not recorded as
applied) or degrades (add the step to `proto.DegradableSteps`). A step that must run a program
calls `execrun` and adds the program to its allowlist -- nothing else may import `os/exec`. A
step that remembers what it found on the host (the firewall's zone check) implements
`apply.Rechecker`, so the drift check's `Engine.Reassert` makes it look again.

### Add a command

Add an entry to `commands()` (or a group's list) in `cmd/meshctl/commands.go` and its `setup`
function: register flags, return the action. Take secrets from the environment, never as flag
defaults (`-h` prints defaults). Then `./scripts/dev.sh regen` to refresh docs/commands.md.

### Add a web UI page

Pages live in `internal/server/ui`: a handler in `pages.go` (or `login.go`), registered in
`New` (everything but the sign-in pages goes through `signedIn`, which enforces the session, the
same origin and the CSRF token on changes), and a template in `templates/<page>.html` that
defines `content`; shared blocks are `*.partial.html`. Compute display strings in Go, not in the
template. Write through `api.Nodes` or the reconciler, never the store, so the change is
versioned and audited. No inline script or style: the CSP blocks them and
`TestTemplatesHaveNoInlineScriptOrStyle` fails. Add the page to `TestEveryPageRenders`, and to
`TestPageGoldens` if its layout matters, then `./scripts/dev.sh regen` and review the golden.

The node's local UI is `internal/agent/localui` (amendment A7): a handler in `pages.go`
registered in `Handler` (`signedIn` for pages, `change` for every POST, which checks the
session, the Origin and the CSRF token before the handler runs), a template in
`templates/<page>.html` that uses the `head` and `foot` blocks of `layout.html`, and only
`static/app.css` beside it -- no script at all (`TestTemplatesHaveNoScript`). It reaches the
agent only through `localui.Funcs`; a change of the node's settings goes through the same
`prefs` relay as `meshctl exit`.

A Windows binary embeds `wintun.dll` (amendment A34). `build` and `dist` run
`go run ./internal/tools/fetchwintun` first, which downloads the pinned release, checks the zip
and the DLL against the pins in `internal/agent/wintun/wintun.go` and writes
`internal/agent/wintun/wintun_amd64.dll` (gitignored; a matching one is kept), then build the
Windows target with `-tags wintun`. `vet`, `test` and `cov` build without the tag and need no
network: such a binary refuses to start WireGuard with a message saying so. To update wintun:
change `Version`, `ZipURL`, `ZipSHA256` and the DLL pins together, from the hashes wintun.net
publishes, and replace `docs/third-party/wintun-LICENSE.txt`.

Every binary embeds Go's time zone database (`time/tzdata`, imported by `internal/validate`), so the
mesh firewall's schedules (amendment A44) name IANA zones on the distroless image and on Windows,
which carry none; a host's own database, when it has one, is read first.

To update htmx: download the release tarball from the npm registry, check it against the
registry's `dist.integrity`, copy `dist/htmx.min.js` to `internal/server/ui/static/`, update
`HTMXSHA256` in `render.go` and `docs/third-party/htmx-LICENSE.txt`, and re-run the gates.

### Add a database migration

Add `internal/server/store/migrations/NNNN_<name>.sql` with the next number. Never edit a
shipped migration. Each file runs in its own transaction with its `schema_version` row; a
binary refuses to open a database newer than its newest migration. If the change touches
desired state, update `store.Snapshot` and `Restore` and the round-trip test with it.
