# Graph Report - vpn-mesh-utility  (2026-10-23)

## Corpus Check
- 286 files · ~326,045 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3490 nodes · 17051 edges · 133 communities (108 shown, 25 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2015 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6fc9763a`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- Node
- Role
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- net/netip.AddrPort
- meshHarness
- github.com/hafio/vpn-mesh-utility
- `internal/server/api`
- Run
- net/http.ResponseWriter
- .do
- PublishNames
- Dir
- ServerConfig
- Agent (`internal/agent/...`)
- net/netip.Prefix
- Net
- ui
- AdminError
- lab
- Snapshot
- .nodeView
- Command reference
- time.Time
- flag.FlagSet
- runner
- newRig
- Service
- agent/agent_test.go
- service_test.go
- `internal/server/auth`
- testing.T
- Policy
- .exec
- openTest
- .mesh
- netlinkNet
- wgdev
- planner_test.go
- Run
- CookieValue
- README.md
- Host
- Input
- otherNet
- .Tx
- context.Context
- Compose and stack files
- meshctl/main.go
- Nodes
- Client
- io.Writer
- net/netip.Addr
- meshDNS
- Tx
- Net
- service_test.go
- Container
- .query
- .Apply
- UI
- Registration
- Hub
- 7. Agent
- Segmentation
- Name
- StepError
- New
- .SetIPv6
- Name
- RenderContainerCompose
- newRig
- commands.go
- resolved
- errorf
- Join
- Agent protocol
- TestPageGoldens
- deploy_test.go
- nullInt
- fetchwintun/main_test.go
- Security
- walk
- Ruleset
- service_windows.go
- Diff
- Error
- Runtime
- Host
- Extending the code
- userspace
- TestRenderRefuses
- conn
- serveRuntime
- .add
- .queryRow
- Schedule
- Policy
- .Sync
- exprs
- PolicyRule
- noKernel
- nameTable
- NodeConfig
- invalid
- DNS
- `internal/validate`
- printStatus
- NewKernel
- droppedLogs
- ns
- loadMigrations
- PlanError
- `internal/server` and `internal/server/export`
- TestMain
- entry
- Error
- NewNet
- NewNet
- NewKernel
- NewKernel
- `internal/proto`
- NewKernel
- Refusal
- StatusError

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 239 edges
2. `ui` - 90 edges
3. ``internal/server/api`` - 78 edges
4. `Session` - 74 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 64 edges
7. `newHarness()` - 56 edges
8. `env` - 48 edges
9. `runner` - 46 edges
10. `api` - 42 edges

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

## Communities (133 total, 25 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (23): composeNetwork, containerFile, containerService, nasFile, nasHealth, nasService, linkDNS, DNSRecord (+15 more)

### Community 1 - "NodeConfig"
Cohesion: 0.16
Nodes (33): rig, NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS(), newRig() (+25 more)

### Community 2 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 3 - "Role"
Cohesion: 0.24
Nodes (13): Role, blockOf(), Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), Range, RangeFor() (+5 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (106): Status, TestCreateLinkNeedsPrivilege(), A(), Ae(), an(), at(), B(), Be() (+98 more)

### Community 5 - "env"
Cohesion: 0.20
Nodes (25): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+17 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.08
Nodes (25): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem, 2. Goals (+17 more)

### Community 7 - "AddrEntry"
Cohesion: 0.13
Nodes (17): Service, activeAt(), zoneOf(), checkSegmentation(), hosts(), mergePorts(), mergeServices(), newSegResolver() (+9 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.16
Nodes (15): Rule, Violation, Packages, and which way they may import, Input and commands, Coverage, `internal/archtest`, `internal/tools/vulnjudge`, Test catalogue (+7 more)

### Community 11 - "net/netip.AddrPort"
Cohesion: 0.13
Nodes (15): resolver, bindBoth(), bindPair(), serve(), Listen(), isPipe(), Listen(), listenSocket() (+7 more)

### Community 12 - "meshHarness"
Cohesion: 0.19
Nodes (19): cli(), TestIPv6DNSRecords(), TestReportedHostFamilies(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness (+11 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.07
Nodes (77): countingDB, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket() (+69 more)

### Community 15 - "Run"
Cohesion: 0.11
Nodes (24): Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New(), NewAdmin() (+16 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.15
Nodes (9): fakeServer, api, decode(), reply(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+1 more)

### Community 17 - ".do"
Cohesion: 0.23
Nodes (7): NodeSetRequest, SegmentView, ZoneView, AdminClient, nodePath(), Enrollment, NodeChange

### Community 18 - "PublishNames"
Cohesion: 0.18
Nodes (11): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), PublishNames(), DNSView (+3 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (25): `internal/fsutil`, TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious() (+17 more)

### Community 20 - "ServerConfig"
Cohesion: 0.24
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+17 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.05
Nodes (64): Agent (`internal/agent/...`), Exec, TestCleanHostname(), TestDefaultSocket(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine(), TestReportKeepsToTheServersLimits() (+56 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.11
Nodes (14): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+6 more)

### Community 24 - "ui"
Cohesion: 0.10
Nodes (13): Session, listOf(), meshChanged(), title(), atoi(), before(), changed(), versionNumber() (+5 more)

### Community 26 - "lab"
Cohesion: 0.17
Nodes (9): End to end -- `test/e2e`, host, eventually(), lab, stream, inNS(), TestServerMesh(), TestSpokesAndExit() (+1 more)

### Community 27 - "Snapshot"
Cohesion: 0.15
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - ".nodeView"
Cohesion: 0.16
Nodes (9): ago(), bytesize(), writeHTML(), TestHelpers(), dashboardData, nodeForm, nodeRow, nodesData (+1 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.05
Nodes (30): syncBuffer, Deps, Failover, replayCache, failures, hitWindow, Lockout, LockoutError (+22 more)

### Community 31 - "flag.FlagSet"
Cohesion: 0.28
Nodes (22): setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), adminFlags(), cleared(), joinNonEmpty(), nodeAction() (+14 more)

### Community 32 - "runner"
Cohesion: 0.12
Nodes (7): pendingApply, runner, safeApply, kick(), oneLine(), drain(), TestSafeApplyStateMachine()

### Community 33 - "newRig"
Cohesion: 0.13
Nodes (36): brokenStream, The web UI's pages (internal/server/ui), `internal/server/ui`, TestDashboardHealth(), TestFirewallPages(), TestNodeGroupsOnPages(), newQR(), TestPhonePage() (+28 more)

### Community 34 - "Service"
Cohesion: 0.16
Nodes (7): Registrar, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (56): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), newFakeRuntime() (+48 more)

### Community 36 - "service_test.go"
Cohesion: 0.19
Nodes (9): SystemdUnit(), Runner, exists(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere(), TestInstallWritesTheUnitAndStartsIt(), Installer (+1 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.10
Nodes (36): Config, Params, rig, 11. CLI reference, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions() (+28 more)

### Community 38 - "testing.T"
Cohesion: 0.11
Nodes (47): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+39 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.12
Nodes (4): mustAffect(), joinList(), boolInt(), classify()

### Community 41 - "openTest"
Cohesion: 0.10
Nodes (51): `internal/server/store`, TestExportOnceSkipsUnchangedAndKeepsFourteen(), DefaultPolicy(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules() (+43 more)

### Community 42 - ".mesh"
Cohesion: 0.15
Nodes (14): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+6 more)

### Community 43 - "netlinkNet"
Cohesion: 0.09
Nodes (14): netlinkNet, Owned(), TestGatewayRules(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule() (+6 more)

### Community 44 - "wgdev"
Cohesion: 0.24
Nodes (3): TestUAPISetGolden(), UAPISet(), wgdev

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (46): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), Problem, Compute(), deliver(), deliverAddr() (+38 more)

### Community 46 - "Run"
Cohesion: 0.10
Nodes (12): Deps, Deps, Run(), DNS, Firewall, MeshFirewall, Sysctls, Source (+4 more)

### Community 47 - "CookieValue"
Cohesion: 0.21
Nodes (5): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), group()

### Community 48 - "README.md"
Cohesion: 0.14
Nodes (9): graphify, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together, meshctl (+1 more)

### Community 49 - "Host"
Cohesion: 0.07
Nodes (33): link, exitHost, Agent service (systemd), Host, peer(), TestLinksAddressesAndFailures(), hexKey(), ParseUAPI() (+25 more)

### Community 50 - "Input"
Cohesion: 0.13
Nodes (13): Routes, Record, containers6(), Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable() (+5 more)

### Community 52 - ".Tx"
Cohesion: 0.10
Nodes (25): fakeService, DB, Pending, Sealer, addUser(), newPassword(), passwdUser(), resetTOTP() (+17 more)

### Community 53 - "context.Context"
Cohesion: 0.08
Nodes (25): Add a web UI page, Client, Logs, Handler(), logs(), refusal(), New(), laptop() (+17 more)

### Community 54 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 55 - "meshctl/main.go"
Cohesion: 0.17
Nodes (17): agentCommands(), commands(), runWith(), firewallCommands(), healthCommands(), find(), isHelp(), main() (+9 more)

### Community 56 - "Nodes"
Cohesion: 0.12
Nodes (19): NodeList, PhoneProfile, PhoneRequest, addresses(), heard(), checkName(), Nodes, nodeByName() (+11 more)

### Community 57 - "Client"
Cohesion: 0.16
Nodes (12): statusDetail, apiError(), clip(), Client, Rollback, StatusReport, StatusResponse, TestStatusReportValidate() (+4 more)

### Community 58 - "io.Writer"
Cohesion: 0.14
Nodes (13): agentLog(), printExit(), printChange(), printEnrollment(), ttyPassword(), OpenFile(), TestFileRotates(), Write() (+5 more)

### Community 59 - "net/netip.Addr"
Cohesion: 0.11
Nodes (27): TestIsPublic(), firstAddr(), isPublic(), isPublic6(), PublicIPv4(), PublicIPv6(), contains(), ofFamily() (+19 more)

### Community 60 - "meshDNS"
Cohesion: 0.15
Nodes (24): Apply (agent), desired(), same(), answerOf(), ask(), meshDNS(), newRegistrar(), query() (+16 more)

### Community 61 - "Tx"
Cohesion: 0.15
Nodes (23): Change, hubMove, hubState, NodeAddRequest, checkGroups(), containerSubnet(), containerSubnet6(), defaultMTU() (+15 more)

### Community 62 - "Net"
Cohesion: 0.12
Nodes (22): network, Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard() (+14 more)

### Community 63 - "service_test.go"
Cohesion: 0.15
Nodes (12): applied, Firewall, policy, Firewall, atLeast(), normalize(), richRule(), RuleArgv() (+4 more)

### Community 64 - "Container"
Cohesion: 0.31
Nodes (4): fakeRuntime, Container, ContainerName, NamesRequest

### Community 65 - ".query"
Cohesion: 0.14
Nodes (6): scanAll(), splitList(), AuditEntry, PeerStat, StatusUpdate, Version

### Community 66 - ".Apply"
Cohesion: 0.16
Nodes (13): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+5 more)

### Community 67 - "UI"
Cohesion: 0.15
Nodes (11): hostOf(), oldest(), validOrNot(), ago(), bytesize(), prefsFlash(), ULAPrefix(), handler (+3 more)

### Community 68 - "Registration"
Cohesion: 0.14
Nodes (17): Domain, fakeRegistrar, nrpt, nrptRule, Registration, domainName(), fromRules(), nrptKey() (+9 more)

### Community 69 - "Hub"
Cohesion: 0.25
Nodes (3): Hub, subscription, ConfigEvent

### Community 70 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 72 - "Name"
Cohesion: 0.15
Nodes (31): check(), newNode(), validateSet(), ToSnapshot(), checkDNS(), Endpoint6(), IP(), IPv6() (+23 more)

### Community 74 - "New"
Cohesion: 0.16
Nodes (10): loadPages(), newRenderer(), staticHandler(), headers(), New(), sameOrigin(), TestEveryPageRenders(), pages (+2 more)

### Community 75 - ".SetIPv6"
Cohesion: 0.23
Nodes (10): IPv6Request, IPv6View, addIPv6(), ipv6View(), Allocate6(), AllocateContainers6(), Block6(), CheckMeshCIDR6() (+2 more)

### Community 76 - "Name"
Cohesion: 0.17
Nodes (10): validateMeshNames(), Ref(), ServiceEntry(), TestRef(), TestServiceEntry(), Name(), TestName(), TestText() (+2 more)

### Community 77 - "RenderContainerCompose"
Cohesion: 0.11
Nodes (28): containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers(), serviceExample() (+20 more)

### Community 78 - "newRig"
Cohesion: 0.20
Nodes (12): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+4 more)

### Community 79 - "commands.go"
Cohesion: 0.22
Nodes (12): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), setupVersion(), shownDefault() (+4 more)

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.11
Nodes (22): apiError, RecordRequest, segLiteral, segDoc, Actor, segLiterals(), errorf(), PolicyZone() (+14 more)

### Community 82 - "Join"
Cohesion: 0.23
Nodes (10): hostService(), installAgent(), installAgent(), cleanHostname(), JoinOptions, Service, Join(), meshAddresses() (+2 more)

### Community 83 - "Agent protocol"
Cohesion: 0.22
Nodes (9): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Signed requests (+1 more)

### Community 84 - "TestPageGoldens"
Cohesion: 0.08
Nodes (26): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+18 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.13
Nodes (34): containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposePosture(), TestContainerComposeRefusesHostileValues(), TestContainerRendererHoldsItsStructure(), checkPosture(), keys() (+26 more)

### Community 86 - "nullInt"
Cohesion: 0.21
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.14
Nodes (20): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+12 more)

### Community 88 - "Security"
Cohesion: 0.15
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Changes a node asks for, DNS, Interfaces the agent did not create, Security, The host firewall (+6 more)

### Community 89 - "walk"
Cohesion: 0.32
Nodes (7): addrText(), Rule, merged(), portText(), services(), Range, service

### Community 90 - "Ruleset"
Cohesion: 0.17
Nodes (7): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, packet

### Community 92 - "service_windows.go"
Cohesion: 0.11
Nodes (10): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows() (+2 more)

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 95 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 96 - "Host"
Cohesion: 0.31
Nodes (4): Host, SyncAddrs(), SyncLink(), TestSyncLinkRefusesAForeignInterface()

### Community 97 - "Extending the code"
Cohesion: 0.22
Nodes (9): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+1 more)

### Community 98 - "userspace"
Cohesion: 0.11
Nodes (6): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace

### Community 99 - "TestRenderRefuses"
Cohesion: 0.28
Nodes (9): Keys and tokens, NewKey(), Render(), aliceKey(), golden(), TestNewKey(), TestProfiles(), TestRenderRefuses() (+1 more)

### Community 100 - "conn"
Cohesion: 0.43
Nodes (3): conn(), hasLegacy(), nftKernel

### Community 101 - "serveRuntime"
Cohesion: 0.23
Nodes (11): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+3 more)

### Community 103 - ".queryRow"
Cohesion: 0.13
Nodes (4): scanNode(), HashToken(), scanUser(), scanner

### Community 104 - "Schedule"
Cohesion: 0.19
Nodes (7): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), weekday()

### Community 105 - "Policy"
Cohesion: 0.13
Nodes (27): NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile(), Local, compile(), dualPolicy() (+19 more)

### Community 107 - "exprs"
Cohesion: 0.13
Nodes (11): build(), exprs(), ifname(), intervalAddrElems(), intervalElems(), TestBuildBatch(), TestExprs(), TestIntervalAddrElems() (+3 more)

### Community 108 - "PolicyRule"
Cohesion: 0.33
Nodes (5): Matrix, MatrixCell, PolicyRule, Schedule, SegPolicy

### Community 111 - "NodeConfig"
Cohesion: 0.06
Nodes (28): NodeDetail, dialing, IPv6Host, Rechecker, Determinism and goldens, Add an apply step, Shared fixtures -- `internal/testfix`, forHost() (+20 more)

### Community 112 - "invalid"
Cohesion: 0.16
Nodes (12): ComposeFile, ComposeOptions, NetworkInfo, ReachRequest, invalid(), plannerNode(), flowEnd(), flowEnds() (+4 more)

### Community 113 - "DNS"
Cohesion: 0.19
Nodes (5): nameServer, DNS, Want, newNameTable(), MeshName

### Community 114 - "`internal/validate`"
Cohesion: 0.09
Nodes (22): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), LocalToken(), mask() (+14 more)

### Community 115 - "printStatus"
Cohesion: 0.50
Nodes (4): dnsLine(), setupDNS(), bytesize(), printStatus()

### Community 118 - "ns"
Cohesion: 0.14
Nodes (12): lab, TestContainerNetworks(), TestDualStack(), deleteNamespaces(), newLab(), ns(), tail(), lab (+4 more)

### Community 119 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.09
Nodes (31): watchdog, `internal/server` and `internal/server/export`, NewAdminClient(), FromSnapshot(), Document, Read(), ReadSegmentation(), dualSnapshot() (+23 more)

### Community 134 - "`internal/proto`"
Cohesion: 0.06
Nodes (50): Options, `internal/proto`, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents() (+42 more)

## Knowledge Gaps
- **5 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `stream`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 333 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **25 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `NodeConfig`, `Node`, `htmx.min.js`, ``internal/proto``, `archtest.go`, `net/netip.AddrPort`, `Run`, `Dir`, `ServerConfig`, `time.Time`, `runner`, `agent/agent_test.go`, `service_test.go`, `testing.T`, `Policy`, `.mesh`, `netlinkNet`, `wgdev`, `Run`, `Host`, `context.Context`, `io.Writer`, `net/netip.Addr`, `meshDNS`, `Net`, `.Apply`, `Registration`, `newRig`, `resolved`, `fetchwintun/main_test.go`, `service_windows.go`, `Host`, `serveRuntime`, `Policy`, `exprs`, `NodeConfig`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 238 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 238 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _5 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07909564020587519 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect ``internal/server/api`` to `newRig`, `htmx.min.js`, `openTest`, `archtest.go`, `meshHarness`, `Run`, `.do`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.07389491242702252 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `.mesh`, `testing.T`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._