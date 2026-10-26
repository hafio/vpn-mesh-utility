#!/bin/sh
# meshctl installer, served by the meshctl control plane at /install.sh:
#
#   curl -fsSL https://<server>/install.sh | sudo sh -s -- --mesh <mesh> --token <token>
#
# It downloads meshctl for this machine from the same server, checks the
# download against the SHA-256 baked into this script when the server
# rendered it, installs it to /usr/local/bin, installs the agent's systemd
# unit meshctl-agent@.service, runs `meshctl agent join` and starts the
# mesh's instance, meshctl-agent@<mesh> (amendment A51). The token may also
# come from the MESHCTL_TOKEN environment variable, the mesh from
# MESHCTL_MESH.
set -eu

ORIGIN='https://mesh.example.com'
VERSION='v1.2.3'
SHA256_AMD64='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
SHA256_ARM64='bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb'

die() {
  echo "install.sh: $*" >&2
  exit 1
}

usage() {
  echo "usage: install.sh --token <token> [--mesh <mesh>] [--name <name>] [--endpoint <ipv4:port>] [--endpoint6 <[ipv6]:port>]"
}

# unit writes the agent's systemd unit, unchanged, to $1.
unit() {
  cat >"$1" <<'MESHCTL_UNIT'
# meshctl agent, one instance per mesh (amendment A51): meshctl-agent@<mesh>
# runs the agent of the mesh <mesh> on /var/lib/meshctl/<mesh>. install.sh
# installs this file unchanged to /etc/systemd/system/meshctl-agent@.service.
# The agent runs as root (it owns its interface, routes, sysctls and, on
# servers, firewalld), but only with the two capabilities that needs, on a
# read-only system. docs/security.md explains each setting; deploy_test.go
# pins them.
[Unit]
Description=meshctl mesh agent (%i)
Documentation=https://github.com/hafio/vpn-mesh-utility
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/meshctl agent run --mesh %i
Restart=always
RestartSec=5
# Exit 3: the control plane refused this node (amendment A50); it stays
# stopped until an operator acts.
RestartPreventExitStatus=3
StateDirectory=meshctl/%i
StateDirectoryMode=0700
RuntimeDirectory=meshctl wireguard
RuntimeDirectoryMode=0755
RuntimeDirectoryPreserve=yes
UMask=0077
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths=/var/lib/meshctl
ProtectHome=yes
PrivateTmp=yes
ProtectKernelModules=yes
ProtectKernelLogs=yes
ProtectControlGroups=yes
ProtectClock=yes
ProtectHostname=yes
RestrictNamespaces=yes
RestrictRealtime=yes
RestrictSUIDSGID=yes
LockPersonality=yes
MemoryDenyWriteExecute=yes
SystemCallArchitectures=native
DevicePolicy=closed
DeviceAllow=/dev/net/tun rw
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK

[Install]
WantedBy=multi-user.target
MESHCTL_UNIT
}

main() {
  token="${MESHCTL_TOKEN:-}"
  mesh="${MESHCTL_MESH:-default}"
  name=""
  endpoint=""
  endpoint6=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --token) [ $# -ge 2 ] || die "--token needs a value"; token="$2"; shift 2 ;;
      --mesh) [ $# -ge 2 ] || die "--mesh needs a value"; mesh="$2"; shift 2 ;;
      --name) [ $# -ge 2 ] || die "--name needs a value"; name="$2"; shift 2 ;;
      --endpoint) [ $# -ge 2 ] || die "--endpoint needs a value"; endpoint="$2"; shift 2 ;;
      --endpoint6) [ $# -ge 2 ] || die "--endpoint6 needs a value"; endpoint6="$2"; shift 2 ;;
      -h | --help) usage; exit 0 ;;
      *) usage >&2; die "unknown argument: $1" ;;
    esac
  done
  [ -n "$token" ] || die "no join token: pass --token <token> or set MESHCTL_TOKEN"
  # The mesh names the unit instance and the data dir: a mesh name only.
  case "$mesh" in
    '' | -* | *- | *[!a-z0-9-]*) die "--mesh must be a mesh name (a-z, 0-9 and '-'), not '$mesh'" ;;
  esac
  [ "${#mesh}" -le 32 ] || die "--mesh must be at most 32 characters"
  [ "$(id -u)" -eq 0 ] || die "must run as root: pipe into 'sudo sh -s -- --mesh <mesh> --token <token>'"
  [ "$(uname -s)" = Linux ] || die "this installer is for Linux; on Windows, use install.ps1"
  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64; sum="$SHA256_AMD64" ;;
    aarch64 | arm64) arch=arm64; sum="$SHA256_ARM64" ;;
    *) die "unsupported CPU $(uname -m); meshctl ships linux/amd64 and linux/arm64" ;;
  esac
  command -v curl >/dev/null 2>&1 || die "curl is required"
  command -v sha256sum >/dev/null 2>&1 || die "sha256sum is required (coreutils)"
  case "$ORIGIN" in
    https://*) protos='=https' ;;
    *) protos='=http,https' ;; # the server only allows plain http on loopback
  esac

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  curl -fsSL --proto "$protos" -o "$tmp/meshctl" "$ORIGIN/dl/meshctl-linux-$arch" ||
    die "downloading meshctl-linux-$arch from $ORIGIN failed"
  echo "$sum  $tmp/meshctl" | sha256sum -c - >/dev/null 2>&1 ||
    die "checksum mismatch for meshctl-linux-$arch: the download is corrupt or was altered; nothing was installed"
  install -m 0755 "$tmp/meshctl" /usr/local/bin/meshctl
  echo "installed meshctl $VERSION to /usr/local/bin/meshctl"

  service="meshctl-agent@$mesh.service"
  systemd=no
  running=no
  if [ -d /run/systemd/system ]; then
    systemd=yes
    unit "$tmp/unit"
    install -m 0644 "$tmp/unit" /etc/systemd/system/meshctl-agent@.service
    systemctl daemon-reload
    # The join needs the mesh's data dir to itself; a failed join starts
    # the running agent again below.
    if systemctl is-active --quiet "$service"; then
      running=yes
      systemctl stop "$service"
    fi
  fi

  set -- --server "$ORIGIN" --mesh "$mesh"
  [ -z "$name" ] || set -- "$@" --name "$name"
  [ -z "$endpoint" ] || set -- "$@" --endpoint "$endpoint"
  [ -z "$endpoint6" ] || set -- "$@" --endpoint6 "$endpoint6"
  if ! MESHCTL_TOKEN="$token" /usr/local/bin/meshctl agent join "$@"; then
    if [ "$running" = yes ]; then
      systemctl start "$service" && echo "started $service again on its earlier enrollment" >&2
    fi
    die "meshctl agent join failed (above)"
  fi

  if [ "$systemd" = yes ]; then
    systemctl enable "$service" >/dev/null
    systemctl restart "$service"
    echo "started $service"
  else
    echo "systemd is not running here: start the agent with: meshctl agent run --mesh $mesh"
  fi
}

main "$@"
