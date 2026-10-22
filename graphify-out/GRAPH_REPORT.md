# Graph Report - vpn-mesh-utility  (2026-10-22)

## Corpus Check
- 272 files · ~286,801 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3301 nodes · 15705 edges · 120 communities (100 shown, 20 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 1820 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `2924301b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- storeNode
- Node
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- Ruleset
- meshHarness
- github.com/hafio/vpn-mesh-utility
- `internal/server/api`
- Run
- net/http.ResponseWriter
- NodeChange
- action
- Dir
- ServerConfig
- Agent (`internal/agent/...`)
- net/netip.Prefix
- meshctl/main.go
- Add a web UI page
- Role
- lab
- Snapshot
- Session
- Command reference
- time.Time
- Key
- runner
- newRig
- ns
- agent/agent_test.go
- Input
- `internal/server/auth`
- testing.T
- Policy
- .exec
- openTest
- .mesh
- netlinkNet
- wgdev
- planner_test.go
- meshDNS
- net/http.ResponseWriter
- README.md
- serve
- Security
- otherNet
- Service
- context.Context
- .Set
- Net
- Diff
- Client
- loadMigrations
- DNS
- Service
- Tx
- Sync
- service_test.go
- TestZoneAnswersExtraDomains
- Container
- Compose and stack files
- UI
- Registration
- log/slog.Logger
- Agent protocol
- otherPolicy
- Name
- StepError
- Host
- TestMain
- net/netip.AddrPort
- container.go
- ReachResult
- scanNode
- .query
- errorf
- resolved
- Error
- .fail
- deploy_test.go
- Run
- fetchwintun/main_test.go
- nft_linux.go
- AddrEntry
- NewNet
- Join
- route
- NewKernel
- Refusal
- .queryRow
- 7. Agent
- userspace
- File
- kernel
- serveRuntime
- Extending the code
- net/http.HandlerFunc
- Policy
- Ring
- NewHub
- conn
- TestFailoverMovesRelays
- JoinRequest
- net/netip.Addr
- Exit
- validate_test.go
- NewNet
- NewKernel
- `internal/server` and `internal/server/export`
- validate.go
- AdminError
- `internal/proto`
- StatusError

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 225 edges
2. `ui` - 87 edges
3. `Session` - 72 edges
4. ``internal/server/api`` - 68 edges
5. `ServerConfig()` - 62 edges
6. `NodeConfig` - 59 edges
7. `newHarness()` - 52 edges
8. `env` - 47 edges
9. `runner` - 44 edges
10. `api` - 40 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Dev script tasks` --references--> `build()`  [INFERRED]
  docs/developer.md → internal/agent/policy/nft_linux.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the server (Dokploy)` --references--> `v()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js

## Import Cycles
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/net_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/net_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/containers/net_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/containers/net_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/policy_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/sign_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`

## Communities (120 total, 20 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, SegPolicy (+10 more)

### Community 1 - "NodeConfig"
Cohesion: 0.08
Nodes (47): rig, Determinism and goldens, Keys and tokens, NewHost(), NewRoutes(), IsStepError(), New(), isStep() (+39 more)

### Community 2 - "storeNode"
Cohesion: 0.19
Nodes (21): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+13 more)

### Community 3 - "Node"
Cohesion: 0.27
Nodes (9): Config and events, 4. Roles and topology, Peer, Record, ContainerNetwork(), Node, sortPeers(), sortPrefixes() (+1 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.25
Nodes (27): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+19 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (28): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments (+20 more)

### Community 7 - "AddrEntry"
Cohesion: 0.17
Nodes (16): Service, checkSegmentation(), host(), mergeServices(), newSegResolver(), Overlaps(), parseSpans(), spanOf() (+8 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.13
Nodes (17): Rule, Violation, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates (+9 more)

### Community 11 - "Ruleset"
Cohesion: 0.09
Nodes (12): watchingFirewall, TestEveryCommandHasHelp(), Kernel, Ruleset, admit(), TestCompileAgreesWithDecide(), walk(), Chain (+4 more)

### Community 12 - "meshHarness"
Cohesion: 0.15
Nodes (17): apiError, ErrorMessage(), toAPIError(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness (+9 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.15
Nodes (31): `internal/server/api`, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose(), TestAdminContainerMode(), TestAdminErrorCodes() (+23 more)

### Community 15 - "Run"
Cohesion: 0.16
Nodes (15): Deps, NewNodes(), New(), NewAdmin(), newAPI(), LoadOrCreateKey(), TestLoadOrCreateKey(), hostOf() (+7 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.18
Nodes (7): api, cli(), decodeJSON(), readBody(), writeJSON(), readSums(), ClientIP()

### Community 17 - "NodeChange"
Cohesion: 0.10
Nodes (23): NetworkInfo, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, PhoneRequest, SegmentView (+15 more)

### Community 18 - "action"
Cohesion: 0.14
Nodes (28): dnsLine(), printExit(), setupFirewallApply(), setupFirewallShow(), adminFlags(), cleared(), heard(), joinNonEmpty() (+20 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (26): `internal/fsutil`, TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious() (+18 more)

### Community 20 - "ServerConfig"
Cohesion: 0.21
Nodes (24): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+16 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.04
Nodes (76): Agent (`internal/agent/...`), DefaultSocket(), TestCleanHostname(), TestDefaultSocket(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx(), TestConfigureHasKernelSemantics() (+68 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.13
Nodes (10): Routes, NewTable(), compare(), Table, Sync(), NewTable(), luidOf(), NewTable() (+2 more)

### Community 23 - "meshctl/main.go"
Cohesion: 0.12
Nodes (25): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+17 more)

### Community 24 - "Add a web UI page"
Cohesion: 0.08
Nodes (23): Add a web UI page, TestTemplatesHaveNoScript(), loadPages(), Health(), newPolicyForm(), ago(), bytesize(), dnsState() (+15 more)

### Community 25 - "Role"
Cohesion: 0.18
Nodes (15): JoinResponse, Role, Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateContainers() (+7 more)

### Community 26 - "lab"
Cohesion: 0.15
Nodes (12): Coverage, End to end -- `test/e2e`, host, eventually(), lab, stream, inNS(), meshNames() (+4 more)

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "Session"
Cohesion: 0.11
Nodes (11): Session, group(), atoi(), before(), changed(), versionNumber(), message(), writeHTML() (+3 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (33): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+25 more)

### Community 30 - "time.Time"
Cohesion: 0.09
Nodes (14): syncBuffer, Failover, hubMove, hubState, hitWindow, Shared fixtures -- `internal/testfix`, Exec, TestJudgeHubs() (+6 more)

### Community 31 - "Key"
Cohesion: 0.06
Nodes (36): link, exitHost, Host, peer(), TestLinksAddressesAndFailures(), AddrPort(), TestAddrPort(), hexKey() (+28 more)

### Community 32 - "runner"
Cohesion: 0.11
Nodes (10): pendingApply, runner, safeApply, kick(), oneLine(), TestOneLine(), TestReportKeepsToTheServersLimits(), Result (+2 more)

### Community 33 - "newRig"
Cohesion: 0.12
Nodes (35): The web UI's pages (internal/server/ui), `internal/server/ui`, capped, TestFirewallPages(), TestNodeGroupsOnPages(), newQR(), TestPhonePage(), TestQRCode() (+27 more)

### Community 34 - "ns"
Cohesion: 0.15
Nodes (10): lab, TestContainerNetworks(), deleteNamespaces(), ns(), tail(), lab, TestRelayFailover(), lab (+2 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.11
Nodes (54): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), newFakeRuntime() (+46 more)

### Community 36 - "Input"
Cohesion: 0.18
Nodes (12): PlanError, Check(), checkContainers(), checkDNS(), Problem, sortedByID(), Compute(), Input (+4 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (35): Config, Params, rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates() (+27 more)

### Community 38 - "testing.T"
Cohesion: 0.07
Nodes (72): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+64 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (22): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+14 more)

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (6): mustAffect(), joinList(), boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.14
Nodes (41): `internal/server/store`, DefaultPolicy(), LegacySegmentation(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation(), TestSegmentationRoundTrip() (+33 more)

### Community 42 - ".mesh"
Cohesion: 0.14
Nodes (15): Routes, 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh (+7 more)

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (13): netlinkNet, blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule(), IPNet(), Prefix() (+5 more)

### Community 45 - "planner_test.go"
Cohesion: 0.16
Nodes (35): Add a topology rule, `internal/server/planner`, TestCheckRequested(), deliver(), deliverAddr(), encode(), firstDiff(), hasProblem() (+27 more)

### Community 46 - "meshDNS"
Cohesion: 0.21
Nodes (13): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestServiceAnswersAndRegisters(), TestServiceAnswersTheContainers(), TestServiceBinding() (+5 more)

### Community 47 - "net/http.ResponseWriter"
Cohesion: 0.21
Nodes (7): fakeServer, decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie()

### Community 48 - "README.md"
Cohesion: 0.17
Nodes (7): `internal/tools/vulnjudge`, Test catalogue, Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "serve"
Cohesion: 0.22
Nodes (11): brokenStream, ErrorBody, serve(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums(), containerNAS() (+3 more)

### Community 50 - "Security"
Cohesion: 0.17
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Security (+6 more)

### Community 52 - "Service"
Cohesion: 0.08
Nodes (29): replayCache, DB, failures, Lockout, LockoutError, Pending, Sealer, addUser() (+21 more)

### Community 53 - "context.Context"
Cohesion: 0.05
Nodes (31): fakeService, waitApplied(), resetTOTP(), dial(), dial(), isPipe(), Client, Logs (+23 more)

### Community 54 - ".Set"
Cohesion: 0.39
Nodes (5): nrpt, ownedKey(), deleteRule(), flush(), ownedKeys()

### Community 56 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 59 - "DNS"
Cohesion: 0.13
Nodes (10): nameServer, DNS, Want, TestNewZoneRejects(), TestZoneAnswersContainerNames(), newNameTable(), newZone(), parent() (+2 more)

### Community 60 - "Service"
Cohesion: 0.18
Nodes (8): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), serve()

### Community 61 - "Tx"
Cohesion: 0.11
Nodes (30): Change, ComposeFile, ComposeOptions, checkGroups(), checkName(), containerSubnet(), defaultMTU(), Nodes (+22 more)

### Community 62 - "Sync"
Cohesion: 0.13
Nodes (21): Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard(), Owned() (+13 more)

### Community 63 - "service_test.go"
Cohesion: 0.09
Nodes (18): hostService(), SystemdUnit(), applied, Firewall, policy, Firewall, Runner, atLeast() (+10 more)

### Community 64 - "TestZoneAnswersExtraDomains"
Cohesion: 0.29
Nodes (11): nrptRule, desired(), domainName(), same(), TestSameRegistration(), TestZoneAnswersExtraDomains(), TestZoneFollowsTheConfig(), fromRules() (+3 more)

### Community 65 - "Container"
Cohesion: 0.14
Nodes (6): fakeRuntime, Container, clip(), Source, meshFilter(), Runtime

### Community 66 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 67 - "UI"
Cohesion: 0.16
Nodes (9): hostOf(), oldest(), validOrNot(), bytesize(), prefsFlash(), handler, page, session (+1 more)

### Community 68 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 69 - "log/slog.Logger"
Cohesion: 0.08
Nodes (15): Hub, subscription, NewUserspace(), NewUserspace(), NewUserspace(), ConfigEvent, DB, Reconciler (+7 more)

### Community 70 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+2 more)

### Community 72 - "Name"
Cohesion: 0.17
Nodes (19): RecordRequest, check(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), Ref(), ServiceEntry() (+11 more)

### Community 74 - "Host"
Cohesion: 0.36
Nodes (3): Host, SyncAddr(), SyncLink()

### Community 75 - "TestMain"
Cohesion: 0.14
Nodes (5): entry, NewKernel(), NewKernel(), runProxy(), TestMain()

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.26
Nodes (9): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), Listen(), ParseAddr(), TestParseAddr() (+1 more)

### Community 77 - "container.go"
Cohesion: 0.13
Nodes (26): composeNetwork, containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers() (+18 more)

### Community 78 - "ReachResult"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 79 - "scanNode"
Cohesion: 0.24
Nodes (3): scanNode(), scanUser(), scanner

### Community 80 - ".query"
Cohesion: 0.08
Nodes (18): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), PublishNames(), TestContainerAddr(), TestPublishNames() (+10 more)

### Community 81 - "errorf"
Cohesion: 0.11
Nodes (21): segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), MeshChange, meshChange(), movePolicy() (+13 more)

### Community 82 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 84 - ".fail"
Cohesion: 0.08
Nodes (21): anyText(), listOf(), listText(), meshChanged(), objectRows(), readPolicyForm(), readZoneForm(), ruleRows() (+13 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.13
Nodes (36): parseContainer(), TestContainerComposePosture(), TestContainerRendererHoldsItsStructure(), checkPosture(), keys(), mustRead(), sameSet(), service() (+28 more)

### Community 86 - "Run"
Cohesion: 0.10
Nodes (13): Deps, Deps, Rechecker, Add an apply step, Run(), Sysctl, DNS, Engine (+5 more)

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.20
Nodes (15): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), main(), run() (+7 more)

### Community 88 - "nft_linux.go"
Cohesion: 0.08
Nodes (25): NewPolicy(), NewPolicy(), NewPolicy(), build(), exprs(), ifname(), intervalElems(), TestBuildBatch() (+17 more)

### Community 89 - "AddrEntry"
Cohesion: 0.10
Nodes (24): addrs(), addrText(), contains(), Rule, merged(), portText(), prefixRange(), services() (+16 more)

### Community 92 - "Join"
Cohesion: 0.21
Nodes (10): installAgent(), hostService(), installAgent(), uninstallAgent(), TestJoinValidatesBeforeCallingTheServer(), JoinOptions, Service, Join() (+2 more)

### Community 93 - "route"
Cohesion: 0.42
Nodes (7): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 96 - ".queryRow"
Cohesion: 0.12
Nodes (5): TestDeleteNodeCascades(), TestTokenLifecycle(), HashToken(), NewToken(), Version

### Community 97 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 98 - "userspace"
Cohesion: 0.16
Nodes (4): defaultInterface(), exitRoutes(), unspecified(), userspace

### Community 99 - "File"
Cohesion: 0.27
Nodes (5): agentLog(), ttyPassword(), OpenFile(), TestFileRotates(), File

### Community 100 - "kernel"
Cohesion: 0.14
Nodes (5): Status, lookup(), Link, kernel, links

### Community 101 - "serveRuntime"
Cohesion: 0.21
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 102 - "Extending the code"
Cohesion: 0.22
Nodes (9): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+1 more)

### Community 104 - "Policy"
Cohesion: 0.33
Nodes (4): Matrix, MatrixCell, Policy, PolicyRule

### Community 106 - "NewHub"
Cohesion: 0.50
Nodes (5): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 109 - "JoinRequest"
Cohesion: 0.67
Nodes (3): JoinRequest, pickEndpoint(), TestPickEndpoint()

### Community 110 - "net/netip.Addr"
Cohesion: 0.16
Nodes (12): network, nameTable, zone, TestIsPublic(), host(), parseServers(), serverList(), TestNRPTServers() (+4 more)

### Community 114 - "validate_test.go"
Cohesion: 0.15
Nodes (24): `internal/validate`, FuzzArg(), FuzzName(), TestArg(), TestDNSName(), TestDomains(), TestEndpoint(), TestEnum() (+16 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.09
Nodes (29): watchdog, setupImport(), `internal/server` and `internal/server/export`, FromSnapshot(), Document, Read(), ReadSegmentation(), legacy() (+21 more)

### Community 123 - "validate.go"
Cohesion: 0.20
Nodes (16): Arg(), Domains(), fail(), Hostname(), isLoopbackHost(), Key32(), LocalToken(), mask() (+8 more)

### Community 134 - "`internal/proto`"
Cohesion: 0.05
Nodes (53): Options, `internal/proto`, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents() (+45 more)

## Knowledge Gaps
- **5 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `stream`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 330 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `on()` connect `htmx.min.js` to `.mesh`, `kernel`, `testing.T`?**
  _High betweenness centrality (0.076) - this node is a cross-community bridge._
- **Are the 224 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 224 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _5 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07609592132086336 - nodes in this community are weakly interconnected._
- **Why does ``cmd/meshctl`` connect `testing.T` to `htmx.min.js`, `Ruleset`, `README.md`, `meshctl/main.go`, `Service`?**
  _High betweenness centrality (0.058) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.07756813417190776 - nodes in this community are weakly interconnected._
- **Why does `Command reference` connect `Command reference` to `README.md`, `Security`, `meshctl/main.go`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._