#!/bin/sh
# meshctl uninstaller, served by the meshctl control plane at /uninstall.sh
# (amendment A55):
#
#   curl -fsSL https://<server>/uninstall.sh | sudo sh -s -- --mesh <mesh>
#
# It stops and disables meshctl-agent@<mesh>, has `meshctl agent reset` take
# the agent's tunnel, routes, rules, firewalld zone and DNS registration off
# the host, and deletes the mesh's data dir: /var/lib/meshctl/<mesh>, or for
# an install.sh --data install the directory its link names (--data <dir>
# names it when the link is gone), with the link and the unit's drop-in. It
# deletes only a directory that holds meshctl's data. With the last mesh gone
# it also removes /usr/local/bin/meshctl and the unit template. It asks first
# unless --yes is given. The node stays on the control plane: remove it there
# with `meshctl node rm <name>`.
set -eu
SCRIPT=uninstall.sh

# The part install.sh, reinstall.sh and uninstall.sh share (amendment A55):
# the control plane writes it into each of them when it serves them, so
# every script is one file a host pipes into sh.

ORIGIN='https://mesh.example.com'
VERSION='v1.2.3'
SHA256_AMD64='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
SHA256_ARM64='bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb'

# What a host install keeps where (amendment A51): one data dir (with
# --data, a link to one) and one unit instance per mesh, one binary and one
# unit template for all of them.
ROOT=/var/lib/meshctl
BIN=/usr/local/bin/meshctl
UNIT=/etc/systemd/system/meshctl-agent@.service

die() {
  echo "$SCRIPT: $*" >&2
  exit 1
}

# cleanup runs at exit: the download goes, and a token prompt cut short gives
# the terminal its echo back.
cleanup() {
  [ -z "${tmp:-}" ] || rm -rf "$tmp"
  [ "${hushed:-}" != yes ] || stty echo </dev/tty 2>/dev/null || true
}
trap cleanup EXIT
trap 'exit 130' INT TERM

# checkmesh holds --mesh to a mesh name: it names the unit instance and the
# data dir.
checkmesh() {
  case "$mesh" in
    '' | -* | *- | *[!a-z0-9-]*) die "--mesh must be a mesh name (a-z, 0-9 and '-'), not '$mesh'" ;;
  esac
  [ "${#mesh}" -le 32 ] || die "--mesh must be at most 32 characters"
}

# paths names what belongs to the mesh (checked by checkmesh).
paths() {
  link="$ROOT/$mesh"
  service="meshctl-agent@$mesh.service"
  dropins="/etc/systemd/system/$service.d"
}

# hidden fails when the agent's unit cannot keep data in $1: ProtectHome
# hides /home, /root and /run/user, PrivateTmp and boot empty /tmp,
# /var/tmp and /run, /proc, /sys and /dev are the kernel's, and
# /var/lib/meshctl holds the link to the data.
hidden() {
  case "$1/" in
    /home/* | /root/* | /run/* | /tmp/* | /var/tmp/* | /proc/* | /sys/* | /dev/* | /var/lib/meshctl/*) return 0 ;;
  esac
  return 1
}

# datadir checks a --data value before it reaches a command, a link or a unit
# file: an absolute path of letters, digits, '.', '_', '-' and '/', which a
# unit file takes without quoting, somewhere the agent's unit can write.
datadir() {
  case "$1" in
    /*) ;;
    *) die "--data must be an absolute path, not '$1'" ;;
  esac
  case "$1" in
    *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._/-]*)
      die "--data may hold only letters, digits, '.', '_', '-' and '/', not '$1'" ;;
    */ | *//* | */./* | */../* | */. | */..) die "--data must not hold '.' or '..' parts, '//' or a trailing '/': '$1'" ;;
  esac
  if hidden "$1"; then
    die "--data cannot be under /home, /root, /run, /tmp, /var/tmp, /proc, /sys, /dev or /var/lib/meshctl, which the agent's unit hides, empties or links from: '$1'"
  fi
}

# datacheck holds the --data dir to what a join can use: new, empty or an
# earlier meshctl data dir, and no other mesh's.
datacheck() {
  real="$(readlink -f "$data" 2>/dev/null || true)"
  if [ -n "$real" ] && hidden "$real"; then
    die "--data $data is $real, which the agent's unit hides, empties or links from"
  fi
  [ -e "$data" ] || return 0
  [ -d "$data" ] || die "--data $data is not a directory"
  if [ -n "$(ls -A "$data")" ] && [ ! -e "$data/agent.lock" ]; then
    die "--data $data holds files but no meshctl data: pick a new or empty directory"
  fi
  for other in "$ROOT"/*; do
    if [ -L "$other" ] && [ "$other" != "$link" ] && [ "$(readlink -f "$other")" = "$real" ]; then
      die "--data $data is the data dir of the mesh ${other##*/} already: each mesh's agent needs its own"
    fi
  done
}

# dropin prints the unit drop-in for the data dir $1 (checked by datadir):
# the instance waits for $1's file system, makes no state directory of its
# own (/var/lib/meshctl/<mesh> is the link), may write in $1 and runs the
# agent on $1, with the socket local commands find through the link.
dropin() {
  cat <<MESHCTL_DROPIN
# Written by install.sh or reinstall.sh with --data (amendments A54, A55):
# the agent of the mesh $mesh keeps its data in $1, which
# /var/lib/meshctl/$mesh links to. uninstall.sh removes this file.
[Unit]
RequiresMountsFor=$1

[Service]
StateDirectory=
ReadWritePaths=$1
ExecStart=
ExecStart=/usr/local/bin/meshctl agent run --data $1 --socket /run/meshctl/%i.sock
MESHCTL_DROPIN
}

# unit writes the agent's systemd unit, unchanged, to $1.
unit() {
  cat >"$1" <<'MESHCTL_UNIT'
# meshctl agent, one instance per mesh (amendment A51): meshctl-agent@<mesh>
# runs the agent of the mesh <mesh> on /var/lib/meshctl/<mesh>. install.sh
# installs this file unchanged to /etc/systemd/system/meshctl-agent@.service;
# with --data <dir> it links /var/lib/meshctl/<mesh> to <dir> and adds the
# drop-in meshctl-agent@<mesh>.service.d/data.conf, which runs the instance
# on <dir> and lets it write there (amendment A54). reinstall.sh and
# uninstall.sh keep the link and the drop-in in step (amendment A55).
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

# host stops unless this is Linux and the script runs as root ($1: the
# arguments to show), and finds out whether systemd runs here.
host() {
  [ "$(id -u)" -eq 0 ] || die "must run as root: pipe it into 'sudo sh -s -- $1'"
  [ "$(uname -s)" = Linux ] || die "this script is for Linux; on Windows, use install.ps1"
  systemd=no
  [ ! -d /run/systemd/system ] || systemd=yes
}

# tools finds the download for this CPU and the commands download needs.
tools() {
  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64; sum="$SHA256_AMD64" ;;
    aarch64 | arm64) arch=arm64; sum="$SHA256_ARM64" ;;
    *) die "unsupported CPU $(uname -m); meshctl ships linux/amd64 and linux/arm64" ;;
  esac
  command -v curl >/dev/null 2>&1 || die "curl is required"
  command -v sha256sum >/dev/null 2>&1 || die "sha256sum is required (coreutils)"
}

# download fetches meshctl for this machine from the control plane and
# checks it against the SHA-256 baked into this script, before anything on
# the host changes.
download() {
  case "$ORIGIN" in
    https://*) protos='=https' ;;
    *) protos='=http,https' ;; # the server only allows plain http on loopback
  esac
  tmp="$(mktemp -d)"
  curl -fsSL --proto "$protos" -o "$tmp/meshctl" "$ORIGIN/dl/meshctl-linux-$arch" ||
    die "downloading meshctl-linux-$arch from $ORIGIN failed; nothing was changed"
  echo "$sum  $tmp/meshctl" | sha256sum -c - >/dev/null 2>&1 ||
    die "checksum mismatch for meshctl-linux-$arch: the download is corrupt or was altered; nothing was changed"
}

# place installs the checked download to /usr/local/bin and, where systemd
# runs, the agent's unit template.
place() {
  install -m 0755 "$tmp/meshctl" "$BIN"
  echo "installed meshctl $VERSION to $BIN"
  if [ "$systemd" = yes ]; then
    unit "$tmp/unit"
    install -m 0644 "$tmp/unit" "$UNIT"
    systemctl daemon-reload
  fi
}

# joined is true when the mesh $1 has a joined node here: its data dir (or
# the directory its link names) holds an enrollment.
joined() { [ -e "$ROOT/$1/state.json" ]; }

# meshes lists the meshes with a joined node here.
meshes() {
  for d in "$ROOT"/*; do
    [ ! -e "$d/state.json" ] || printf '%s ' "${d##*/}"
  done
}

# field prints one plain value of the mesh's enrollment, state.json, which
# the agent writes a field to a line; nothing when it is not plain.
field() {
  sed -n "s|^  \"$1\": \"\([A-Za-z0-9.:/_-]*\)\",\{0,1\}\$|\1|p" "$link/state.json" 2>/dev/null | head -n 1
}

# describe says what of the mesh is on this host; dir is its data dir after.
describe() {
  dir="$link"
  [ ! -L "$link" ] || dir="$(readlink "$link")"
  node="$(field name)"
  ver="$("$BIN" version 2>/dev/null || true)"
  active="(no systemd)"
  [ "$systemd" = no ] || active="$(systemctl is-active "$service" 2>/dev/null || true)"
  echo "mesh $mesh is installed here: node ${node:-?} of $(field server), data in $dir, ${ver:-no $BIN}, $service ${active:-unknown}"
}

# interactive is true when there is a terminal to ask on (the script itself
# arrives on stdin).
interactive() { (exec </dev/tty) 2>/dev/null; }

# asktoken reads a join token at the terminal, not echoed, unless --token or
# MESHCTL_TOKEN gave one.
asktoken() {
  [ -z "$token" ] || return 0
  interactive || die "no join token: pass --token <token> or set MESHCTL_TOKEN"
  printf 'Join token for the node (meshctl node token <name> --mesh %s prints one): ' "$mesh" >/dev/tty
  hushed=yes
  stty -echo </dev/tty 2>/dev/null || true
  IFS= read -r token </dev/tty || token=""
  stty echo </dev/tty 2>/dev/null || true
  hushed=no
  echo >/dev/tty
  [ -n "$token" ] || die "no join token given; nothing was changed"
}

# start writes or removes the instance's drop-in to match the mesh's link,
# then enables and (re)starts the instance.
start() {
  if [ "$systemd" = no ]; then
    echo "systemd is not running here: start the agent with: meshctl agent run --mesh $mesh"
    return 0
  fi
  if [ -L "$link" ]; then
    target="$(readlink "$link")"
    datadir "$target"
    mkdir -p "$dropins"
    dropin "$target" >"$dropins/data.conf"
    echo "the agent of $mesh keeps its data in $target ($link links to it)"
  elif [ -e "$dropins/data.conf" ]; then
    rm -f "$dropins/data.conf"
    rmdir "$dropins" 2>/dev/null || true
  fi
  systemctl daemon-reload
  systemctl enable "$service" >/dev/null
  systemctl restart "$service"
  echo "started $service"
}

# join stops the mesh's agent when it runs, points the mesh's data dir at
# --data when given, joins with the token and starts the agent. A failed join
# puts the link back and starts again what it stopped.
join() {
  running=no
  if [ "$systemd" = yes ] && systemctl is-active --quiet "$service"; then
    running=yes
    systemctl stop "$service"
  fi
  previous=""
  if [ -n "$data" ]; then
    [ ! -L "$link" ] || previous="$(readlink "$link")"
    mkdir -p "$data" "$ROOT"
    chmod 0700 "$data"
    rm -f "$link"
    ln -s "$data" "$link"
  fi
  set -- --server "$ORIGIN" --mesh "$mesh"
  [ -z "$name" ] || set -- "$@" --name "$name"
  [ -z "$endpoint" ] || set -- "$@" --endpoint "$endpoint"
  [ -z "$endpoint6" ] || set -- "$@" --endpoint6 "$endpoint6"
  if ! MESHCTL_TOKEN="$token" "$BIN" agent join "$@"; then
    if [ -n "$data" ]; then
      rm -f "$link"
      [ -z "$previous" ] || ln -s "$previous" "$link"
    fi
    if [ "$running" = yes ]; then
      systemctl start "$service" && echo "started $service again on its earlier enrollment" >&2
    fi
    die "meshctl agent join failed (above)${gone:-}"
  fi
  start
}

# locate finds the mesh's data dir, target: where its link points, else
# /var/lib/meshctl/<mesh>, else $1 (uninstall.sh --data) when the link is
# gone; $1, when given, must be that dir. A directory that is not meshctl's
# stops the script before anything is deleted.
locate() {
  target=""
  if [ -L "$link" ]; then
    target="$(readlink "$link")"
    (datadir "$target") 2>/dev/null || die "$link links to '$target', which install.sh does not make: remove it by hand"
    if [ -n "$1" ] && [ "$(readlink -f "$1" 2>/dev/null || true)" != "$(readlink -f "$target" 2>/dev/null || true)" ]; then
      die "--data $1 is not where $link links to ($target)"
    fi
  elif [ -d "$link" ]; then
    [ -z "$1" ] || die "--data $1: the data of $mesh is in $link (it was installed without --data)"
    target="$link"
  elif [ -n "$1" ]; then
    target="$1"
  fi
  if [ -n "$target" ] && [ -e "$target" ]; then
    [ -d "$target" ] || die "$target is not a directory"
    if [ -n "$(ls -A "$target")" ] && [ ! -e "$target/state.json" ] && [ ! -e "$target/agent.lock" ]; then
      die "$target holds files but no meshctl data: nothing was removed"
    fi
  fi
}

# remove takes the mesh's agent off this host: it stops and disables the
# instance, has `meshctl agent reset` take off what the agent made, then
# deletes the drop-in, the link and the data dir (target, from locate).
# Nothing is deleted when the reset fails.
remove() {
  if [ "$systemd" = yes ]; then
    if systemctl is-active --quiet "$service"; then
      systemctl stop "$service"
    fi
    if systemctl is-enabled --quiet "$service" 2>/dev/null; then
      systemctl disable "$service" >/dev/null
    fi
  fi
  if [ -n "$target" ] && [ -e "$target/state.json" ]; then
    [ -x "$BIN" ] || die "$BIN is missing, so nothing can take the agent's tunnel, routes and rules off this host: run reinstall.sh --mesh $mesh --action reinstall first; nothing was deleted"
    "$BIN" agent reset --data "$target" ||
      die "meshctl agent reset could not take everything off this host (above): fix that and run this again; nothing was deleted"
  fi
  rm -f "$dropins/data.conf"
  rmdir "$dropins" 2>/dev/null || true
  [ ! -L "$link" ] || rm -f "$link"
  if [ -n "$target" ] && [ -d "$target" ]; then
    find "$target" -mindepth 1 -maxdepth 1 -exec rm -rf -- {} +
    rmdir "$target" 2>/dev/null || echo "kept the empty directory $target (a mount point?)"
  fi
  [ "$systemd" = no ] || systemctl daemon-reload
  echo "removed the agent of $mesh and its data${target:+ ($target)}"
}

usage() {
  echo "usage: uninstall.sh [--mesh <mesh>] [--data <dir>] [--yes]"
}

# confirm asks at the terminal before anything is deleted.
confirm() {
  interactive || die "no terminal to ask on: pass --yes to remove mesh $mesh from this host"
  printf 'Remove mesh %s from this host? Its keys and data are deleted; joining again needs a new token. [y/N]: ' "$mesh" >/dev/tty
  read -r answer </dev/tty || answer=""
  case "$answer" in
    y | Y | yes | YES) ;;
    *) echo "nothing changed"; exit 0 ;;
  esac
}

# last removes the binary and the unit template once no mesh is left here.
last() {
  if [ -n "$(ls -A "$ROOT" 2>/dev/null)" ]; then
    echo "$ROOT still holds $(ls -A "$ROOT" | tr '\n' ' ')so $BIN and $UNIT stay"
    return 0
  fi
  rm -f "$UNIT" "$BIN"
  rmdir "$ROOT" 2>/dev/null || true
  [ "$systemd" = no ] || systemctl daemon-reload
  echo "no mesh is left on this host: removed $BIN and $UNIT"
}

main() {
  mesh="${MESHCTL_MESH:-default}"
  data="${MESHCTL_DATA:-}"
  yes=no
  while [ $# -gt 0 ]; do
    case "$1" in
      --mesh) [ $# -ge 2 ] || die "--mesh needs a value"; mesh="$2"; shift 2 ;;
      --data) [ $# -ge 2 ] || die "--data needs a value"; data="$2"; shift 2 ;;
      --yes) yes=yes; shift ;;
      -h | --help) usage; exit 0 ;;
      *) usage >&2; die "unknown argument: $1" ;;
    esac
  done
  checkmesh
  [ -z "$data" ] || datadir "$data"
  host "--mesh <mesh>"
  paths
  locate "$data"

  enabled=no
  if [ "$systemd" = yes ] && systemctl is-enabled --quiet "$service" 2>/dev/null; then
    enabled=yes
  fi
  if [ -z "$target" ] && [ ! -e "$dropins" ] && [ "$enabled" = no ]; then
    installed="$(meshes)"
    echo "mesh $mesh is not installed here${installed:+ (installed: $installed)}; nothing to remove"
    exit 0
  fi
  node=""
  if joined "$mesh"; then
    describe
  else
    echo "mesh $mesh has no joined node here; removing what is left of it"
  fi
  [ "$yes" = yes ] || confirm
  remove
  last
  echo "On the control plane, remove the node too: meshctl node rm ${node:-<name>} --mesh $mesh (or on its page in the web UI)"
}

main "$@"
