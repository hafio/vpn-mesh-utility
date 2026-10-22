# CLAUDE.md

Guidance for Claude Code working in this repository.

This file holds **cross-task conventions only** -- the rules that bind you before you touch
any part of the tree. Per-feature mechanics live in `docs/` and in the packages' own comments.
If a rule here and a doc disagree, fix the drift; do not pick one.

## What this repo is

`meshctl`: one Go binary (server + agent + CLI) that builds and operates a private WireGuard
mesh -- three VPS servers for Docker Swarm, a NAS behind CGNAT, laptops that switch their
internet exit, phones by QR. It never publishes, proxies or deploys applications.

- [docs/spec.md](docs/spec.md) is the design spec. Its **Amendments** table records every
  deliberate deviation from the original draft; read it before changing a design decision.
- [README.md](README.md) is the quick start, [docs/developer.md](docs/developer.md) the build
  and dev-script guide, [docs/test.md](docs/test.md) the catalogue of tests, fixtures and seams.

Delivery is milestone-gated (spec S15, M1..M5): a milestone starts only after the previous
one's gate is green on the owner's machine.

## You do not run the gates

Build, vet, test, cov, scan, image, e2e, `go mod tidy`, `go get` and every git or tag action are
the owner's. After a change, name the task (`./scripts/dev.sh all scan`, or the single task
that applies) and continue from the output reported back.

## Packages, and which way they may import

- **`internal/server/...` and `internal/agent/...` never import each other.** They talk only
  through `internal/proto` (the wire contract) and the stdlib-only `internal/validate`.
- **`internal/proto` imports only `internal/validate`; `internal/validate` imports only the
  stdlib.**
- **`internal/testfix` is for tests only** and imports only `internal/proto`, so both sides'
  tests can share it. `internal/agent/agentfake` (in-memory host fakes) is test-only too.
- `deploy` (unit, install.sh, compose files) imports only `internal/validate` and yaml.v3; both sides use it.
- `internal/fsutil` (atomic file writes) is stdlib-only; both sides use it.
- The server never imports data-plane libraries (wgctrl, wireguard-go, netlink, nftables,
  miekg/dns, godbus, x/sys/windows, go-winio); the agent never imports server ones (sqlite,
  argon2, otp, barcode, qrcode).
- `cmd/meshctl` wires everything; nothing imports it.

`internal/archtest` (`TestModuleBoundaries`) enforces all of this for linux and windows builds,
test files included. Change the rules there and here together.

## Security invariants

- **Every external value passes `internal/validate`** before it is stored, rendered into an
  artifact (compose file, systemd unit, install.sh, firewall rule) or handed to a command. Add a
  validator there rather than parsing ad hoc; validators accept a positive grammar.
- **Commands are never run through a shell.** The agent has exactly one `exec` call site,
  `internal/agent/execrun` (allowlist: `firewall-cmd`, `systemctl`), and it re-checks every argv
  element with `validate.Arg`. Nothing else under `internal/` or `cmd/` imports `os/exec`.
- **Secrets never reach logs, errors, audit details or exports**: join tokens and session
  cookies are stored as sha256 and masked in errors; passwords are argon2id hashes read from a
  terminal only; TOTP secrets are sealed with `secret.key`; private keys never leave the node
  (phones are the one documented exception, spec Amendments A6).
- **Container artifacts state their posture**: privileged false, `cap_drop: [ALL]`,
  `no-new-privileges=true`, explicit user, read-only root. Any widening token (the NAS agent's
  `network_mode: host`, `cap_add`, the tun device) is named in docs/security.md and pinned by a
  golden plus a test that forbids every other widening token.

## Determinism and goldens

The planner is a pure function and its output is hashed: every slice it emits is sorted, and
`proto.NodeConfig`'s field order is part of the hash (`TestHashIsPinned`). Golden files under
`testdata/` are hand-reviewed; `./scripts/dev.sh regen` rewrites them and is deliberately not
in `all`/`full` -- a gate must not rewrite what it compares against.

## Style

Plain ASCII everywhere. No global state: dependencies and the version arrive in Config/Deps
structs. Errors say what failed, why (the offending value, secrets masked) and what to do next,
and wrap the cause with `%w`. Logs use `log/slog`. Tests are stdlib-only and table-driven;
cover security and error boundaries by asserting the bad input is rejected **and** the
downstream call was not made.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- The graph is refreshed by the `graphify` task (part of `full`), which the owner runs like every
  other gate; when the graph is stale, propose `./scripts/dev.sh graphify` instead of running it.
