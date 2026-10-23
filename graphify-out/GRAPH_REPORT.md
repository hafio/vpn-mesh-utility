# Graph Report - vpn-mesh-utility  (2026-10-23)

## Corpus Check
- 286 files · ~336,481 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3553 nodes · 17305 edges · 129 communities (105 shown, 24 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2051 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7477ac27`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- Node
- commands.go
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
- AddrRange
- github.com/hafio/vpn-mesh-utility
- `internal/server/api`
- Run
- net/http.ResponseWriter
- .do
- PublishNames
- Dir
- ServerConfig
- testing.T
- net/netip.Prefix
- Net
- Session
- AdminError
- lab
- Snapshot
- .nodeView
- Command reference
- time.Time
- Net
- runner
- newRig
- Service
- agent/agent_test.go
- Nodes
- `internal/server/auth`
- testing.T
- Policy
- .exec
- openTest
- io.Writer
- netlinkNet
- Allocate
- planner_test.go
- Operations
- Registration
- README.md
- Host
- Input
- time.Duration
- context.Context
- context.Context
- Compose and stack files
- Tx
- TestMigrationTurnsRulesIntoPolicies
- Client
- .routes
- net/netip.Addr
- meshDNS
- Run
- Sync
- .Apply
- NodeInfo
- .query
- Writer
- UI
- Registration
- Hub
- TestPageGoldens
- Container
- validate.go
- StepError
- CookieValue
- .renderVersion
- Exit
- container.go
- rig
- 7. Agent
- resolved
- errorf
- Join
- Agent protocol
- .renderPolicies
- deploy_test.go
- stop
- fetchwintun/main_test.go
- 8. The web UI
- Open
- Ruleset
- handler
- Diff
- Error
- Runtime
- IPv6View
- .Reach
- userspace
- Test catalogue
- wgdev
- serveRuntime
- loadMigrations
- .queryRow
- Schedule
- `internal/proto`
- github.com/miekg/dns.Msg
- exprs
- PolicyRule
- TestQRCode
- fakeService
- Engine
- .teardown
- DNS
- validate_test.go
- TestTemplatesHaveNoInlineScriptOrStyle
- NewKernel
- ns
- `internal/server` and `internal/server/export`
- TestMain
- entry
- Segmentation
- NewNet
- NewKernel
- NewKernel
- WriteEvent
- NewKernel
- Refusal
- stream

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 249 edges
2. `ui` - 90 edges
3. ``internal/server/api`` - 80 edges
4. `Session` - 74 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 67 edges
7. `newHarness()` - 58 edges
8. `env` - 48 edges
9. `runner` - 47 edges
10. `api` - 42 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Packages, and which way they may import` --references--> `TestModuleBoundaries()`  [INFERRED]
  CLAUDE.md → internal/archtest/archtest_test.go
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

## Communities (129 total, 24 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (20): Error, linkDNS, DNSRecord, FirewallRule, Node, NewNet(), logsData, inspected (+12 more)

### Community 1 - "NodeConfig"
Cohesion: 0.07
Nodes (57): Deps, IPv6Host, rig, Determinism and goldens, NewHost(), NewRoutes(), forHost(), MeshFirewall (+49 more)

### Community 2 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 3 - "commands.go"
Cohesion: 0.12
Nodes (26): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+18 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.15
Nodes (50): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+42 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (28): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments (+20 more)

### Community 7 - "AddrEntry"
Cohesion: 0.16
Nodes (14): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), Overlaps(), parseSpans(), Summarize() (+6 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.15
Nodes (15): Rule, Violation, Input and commands, Coverage, `internal/archtest`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue (+7 more)

### Community 11 - "net.Listener"
Cohesion: 0.33
Nodes (6): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), serve()

### Community 12 - "AddrRange"
Cohesion: 0.15
Nodes (17): addrs(), addrText(), contains(), Rule, merged(), portText(), services(), inRanges() (+9 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.05
Nodes (99): apiError, brokenStream, countingDB, `internal/server/api`, ErrorBody, ErrorMessage(), NewNodes(), adminCall() (+91 more)

### Community 15 - "Run"
Cohesion: 0.13
Nodes (19): Deps, NewUserspace(), DB, New(), NewAdmin(), newAPI(), Reconciler, NewReconciler() (+11 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.15
Nodes (9): api, decode(), reply(), cli(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+1 more)

### Community 17 - ".do"
Cohesion: 0.19
Nodes (9): NetworkInfo, NodeAddRequest, NodeSetRequest, SegmentView, ZoneView, AdminClient, nodePath(), Enrollment (+1 more)

### Community 18 - "PublishNames"
Cohesion: 0.14
Nodes (14): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+6 more)

### Community 19 - "Dir"
Cohesion: 0.08
Nodes (20): `internal/fsutil`, Enrollment, mkdirs(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestKeysAreCreatedOnceAndKept(), TestLoadKeysRefusesLooseOrDamagedFiles(), TestLocalSettings() (+12 more)

### Community 20 - "ServerConfig"
Cohesion: 0.15
Nodes (30): fakeRunner, allowed(), Check(), TestCheckAllowsFirewallArgv(), TestCheckRefuses(), appliedHost(), encodeCalls(), freshHost() (+22 more)

### Community 21 - "testing.T"
Cohesion: 0.04
Nodes (109): SystemdUnit(), Agent (`internal/agent/...`), DefaultSocket(), freePort(), TestDefaultSocket(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx() (+101 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.07
Nodes (15): otherNet, Routes, index(), NewTable(), route(), compare(), Table, Sync() (+7 more)

### Community 24 - "Session"
Cohesion: 0.13
Nodes (9): Session, listOf(), meshChanged(), atoi(), changed(), message(), actor(), settingsData (+1 more)

### Community 26 - "lab"
Cohesion: 0.22
Nodes (7): host, eventually(), lab, newLab(), TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - ".nodeView"
Cohesion: 0.16
Nodes (9): ago(), bytesize(), writeHTML(), TestHelpers(), dashboardData, nodeForm, nodeRow, nodesData (+1 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.09
Nodes (15): syncBuffer, replayCache, failures, hitWindow, Lockout, LockoutError, Sysctl, serves() (+7 more)

### Community 31 - "Net"
Cohesion: 0.16
Nodes (8): Sysctls, ensureAddr(), Net, Rule, Sync(), syncDropped(), syncProxies(), syncRules()

### Community 32 - "runner"
Cohesion: 0.10
Nodes (11): pendingApply, runner, safeApply, 12. Troubleshooting, kick(), oneLine(), TestOneLine(), TestReportKeepsToTheServersLimits() (+3 more)

### Community 33 - "newRig"
Cohesion: 0.22
Nodes (26): `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled() (+18 more)

### Community 34 - "Service"
Cohesion: 0.17
Nodes (8): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), serve()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.09
Nodes (64): fakeServer, rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime() (+56 more)

### Community 36 - "Nodes"
Cohesion: 0.18
Nodes (11): Change, PhoneRequest, checkName(), Nodes, nodeByName(), nodeInfo(), problemStrings(), withGroups() (+3 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (33): Config, Params, rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates() (+25 more)

### Community 38 - "testing.T"
Cohesion: 0.11
Nodes (41): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+33 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (6): mustAffect(), joinList(), boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.20
Nodes (32): `internal/server/store`, TestRestoreOfASnapshotWithFirewallRules(), TestSegScheduleString(), TestSetNodeGroups(), DecodeSnapshot(), insertTopology(), mustNode(), node() (+24 more)

### Community 42 - "io.Writer"
Cohesion: 0.10
Nodes (18): agentLog(), TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), printExit(), printChange(), printEnrollment(), bytesize() (+10 more)

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (13): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+5 more)

### Community 44 - "Allocate"
Cohesion: 0.21
Nodes (16): Allocate(), Allocate6(), AllocateContainers(), AllocateContainers6(), Block(), Block6(), CheckMeshCIDR(), CheckMeshCIDR6() (+8 more)

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (44): PlanError, Add a topology rule, `internal/server/planner`, Problem, Compute(), deliver(), deliverAddr(), encode() (+36 more)

### Community 46 - "Operations"
Cohesion: 0.21
Nodes (11): 10. Backups, 11. When the server is down, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 5. Day to day, 6. DNS (+3 more)

### Community 47 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 48 - "README.md"
Cohesion: 0.12
Nodes (11): graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates, Building, Documentation (+3 more)

### Community 49 - "Host"
Cohesion: 0.05
Nodes (38): link, exitHost, Host, peer(), TestLinksAddressesAndFailures(), AddrPort(), TestAddrPort(), hexKey() (+30 more)

### Community 50 - "Input"
Cohesion: 0.13
Nodes (13): Record, Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable(), sortedByID(), Input (+5 more)

### Community 51 - "time.Duration"
Cohesion: 0.23
Nodes (5): Failover, waitApplied(), Exec, NewRateLimiter(), TestRateLimiter()

### Community 52 - "context.Context"
Cohesion: 0.08
Nodes (29): DB, Pending, Sealer, addUser(), newPassword(), passwdUser(), resetTOTP(), dial() (+21 more)

### Community 53 - "context.Context"
Cohesion: 0.08
Nodes (24): statusDetail, Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest (+16 more)

### Community 54 - "Compose and stack files"
Cohesion: 0.18
Nodes (11): TestComposeDoc(), A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker (+3 more)

### Community 55 - "Tx"
Cohesion: 0.13
Nodes (34): ComposeFile, ComposeOptions, hubMove, hubState, Role, addIPv6(), blockOf(), checkGroups() (+26 more)

### Community 56 - "TestMigrationTurnsRulesIntoPolicies"
Cohesion: 0.31
Nodes (9): DefaultPolicy(), LegacySegmentation(), insertAt5(), TestMigrationTurnsRulesIntoPolicies(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSetSegmentationRefusesAStranger() (+1 more)

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - ".routes"
Cohesion: 0.42
Nodes (7): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 59 - "net/netip.Addr"
Cohesion: 0.10
Nodes (24): hostAddrs, nameTable, zone, Apply (agent), TestIsPublic(), gateway(), host(), records() (+16 more)

### Community 60 - "meshDNS"
Cohesion: 0.25
Nodes (22): desired(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters() (+14 more)

### Community 61 - "Run"
Cohesion: 0.13
Nodes (11): Deps, Options, Run(), DNS, Firewall, Sysctls, Join(), New() (+3 more)

### Community 62 - "Sync"
Cohesion: 0.17
Nodes (23): network, NewNet(), Gateway(), Guard(), Owned(), parse(), TestGatewayRules(), TestGatewayRulesWithIPv6() (+15 more)

### Community 63 - ".Apply"
Cohesion: 0.10
Nodes (17): hostService(), applied, Firewall, policy, Firewall, Runner, atLeast(), normalize() (+9 more)

### Community 64 - "NodeInfo"
Cohesion: 0.15
Nodes (12): NodeDetail, NodeList, PhoneProfile, addresses(), heard(), PeerStatus, NodeInfo, Health() (+4 more)

### Community 65 - ".query"
Cohesion: 0.12
Nodes (8): ContainerName, scanNode(), scanAll(), splitList(), AuditEntry, PeerStat, scanner, StatusUpdate

### Community 66 - "Writer"
Cohesion: 0.29
Nodes (3): New(), New(), Writer

### Community 67 - "UI"
Cohesion: 0.11
Nodes (17): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize() (+9 more)

### Community 68 - "Registration"
Cohesion: 0.17
Nodes (16): nrpt, nrptRule, domainName(), same(), TestSameRegistration(), fromRules(), nrptKey(), ownedKey() (+8 more)

### Community 69 - "Hub"
Cohesion: 0.24
Nodes (3): Hub, subscription, ConfigEvent

### Community 70 - "TestPageGoldens"
Cohesion: 0.40
Nodes (5): dnsState(), newRenderer(), golden(), TestEveryPageRenders(), TestPageGoldens()

### Community 71 - "Container"
Cohesion: 0.15
Nodes (6): fakeRuntime, clip(), Ring, Container, Source, ContainerName

### Community 72 - "validate.go"
Cohesion: 0.12
Nodes (41): check(), ParseAddr(), TestParseAddr(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), checkDNS() (+33 more)

### Community 74 - "CookieValue"
Cohesion: 0.11
Nodes (9): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), group(), sameOrigin(), handler, page (+1 more)

### Community 77 - "container.go"
Cohesion: 0.20
Nodes (20): composeNetwork, checkIPv6(), containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose() (+12 more)

### Community 78 - "rig"
Cohesion: 0.22
Nodes (4): capped, syncBuffer, faultyDB, rig

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.12
Nodes (20): RecordRequest, segLiteral, segDoc, Actor, segLiterals(), PolicyZone(), MeshChange, meshChange() (+12 more)

### Community 82 - "Join"
Cohesion: 0.21
Nodes (11): hostService(), installAgent(), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), cleanHostname(), JoinOptions, Service (+3 more)

### Community 83 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+2 more)

### Community 84 - ".renderPolicies"
Cohesion: 0.09
Nodes (24): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+16 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.11
Nodes (42): containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposeHeader(), TestContainerComposePosture(), TestContainerComposeRefusesHostileValues(), TestContainerRendererHoldsItsStructure(), TestHostModeNetworkSetup() (+34 more)

### Community 86 - "stop"
Cohesion: 0.24
Nodes (5): uninstallAgent(), stop(), TestAvailableNeedsElevation(), Windows(), SCM

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.14
Nodes (22): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+14 more)

### Community 88 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

### Community 89 - "Open"
Cohesion: 0.39
Nodes (6): openAt(), Store, Open(), TestOpenIsIdempotent(), TestOpenRefusesNewerSchema(), TestOpenRejectsDSNInjection()

### Community 90 - "Ruleset"
Cohesion: 0.10
Nodes (10): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel (+2 more)

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 95 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 96 - "IPv6View"
Cohesion: 0.38
Nodes (3): IPv6Request, IPv6View, ipv6View()

### Community 97 - ".Reach"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 98 - "userspace"
Cohesion: 0.07
Nodes (12): Status, 4. Spokes and exits, Container mode, lookup(), NewUserspace(), defaultInterface(), exitRoutes(), unspecified() (+4 more)

### Community 99 - "Test catalogue"
Cohesion: 0.21
Nodes (16): Keys and tokens, `internal/server/phone`, Hub(), joinSet(), NewKey(), Render(), aliceKey(), dualPhone() (+8 more)

### Community 101 - "serveRuntime"
Cohesion: 0.19
Nodes (13): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+5 more)

### Community 102 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 103 - ".queryRow"
Cohesion: 0.13
Nodes (3): HashToken(), NewToken(), Version

### Community 104 - "Schedule"
Cohesion: 0.11
Nodes (16): Changes a node asks for, DNS, Interfaces the agent did not create, Security, The host firewall, The local UI (internal/agent/localui), The mesh firewall, Web UI sign-in (internal/server/auth) (+8 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.06
Nodes (55): `internal/proto`, NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily() (+47 more)

### Community 107 - "exprs"
Cohesion: 0.08
Nodes (19): Dev script tasks, Developer guide, Goldens, Releases, Toolchain, build(), conn(), exprs() (+11 more)

### Community 108 - "PolicyRule"
Cohesion: 0.22
Nodes (8): Matrix, MatrixCell, Policy, PolicyRule, Schedule, activeAt(), zoneOf(), SegPolicy

### Community 109 - "TestQRCode"
Cohesion: 0.50
Nodes (4): newQR(), TestQRCode(), phoneData, qrCode

### Community 111 - "Engine"
Cohesion: 0.13
Nodes (10): dialing, Rechecker, Add a command, Add a database migration, Add a validator, Add a web UI page, Add an apply step, Extending the code (+2 more)

### Community 113 - "DNS"
Cohesion: 0.17
Nodes (5): nameServer, DNS, Want, DNS, MeshName

### Community 114 - "validate_test.go"
Cohesion: 0.10
Nodes (33): `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), AddrEntry(), TestAddrEntry(), TestAddrRangePrefix(), TestRef() (+25 more)

### Community 118 - "ns"
Cohesion: 0.14
Nodes (12): End to end -- `test/e2e`, lab, TestContainerNetworks(), lab, TestDualStack(), inNS(), ns(), lab (+4 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.08
Nodes (34): watchdog, `internal/server` and `internal/server/export`, FromSnapshot(), Document, Read(), ReadSegmentation(), dualSnapshot(), legacy() (+26 more)

### Community 134 - "WriteEvent"
Cohesion: 0.27
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

## Knowledge Gaps
- **6 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `lab` (+1 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 338 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does ``internal/server/api`` connect ``internal/server/api`` to `htmx.min.js`, `serveRuntime`, `archtest.go`, `.do`, `PublishNames`, `testing.T`, `.routes`, `time.Time`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Are the 248 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 248 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _6 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07674866310160428 - nodes in this community are weakly interconnected._
- **Why does `at()` connect `htmx.min.js` to ``internal/server/api``?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.06672519754170325 - nodes in this community are weakly interconnected._
- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `NodeConfig`, `Node`, `WriteEvent`, `archtest.go`, `Dir`, `ServerConfig`, `time.Time`, `runner`, `agent/agent_test.go`, `Policy`, `io.Writer`, `netlinkNet`, `Operations`, `Host`, `.routes`, `net/netip.Addr`, `meshDNS`, `Run`, `Sync`, `UI`, `Registration`, `validate.go`, `resolved`, `Join`, `stop`, `fetchwintun/main_test.go`, `serveRuntime`, ``internal/proto``, `exprs`, `Engine`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._