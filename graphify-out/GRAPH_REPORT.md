# Graph Report - vpn-mesh-utility  (2026-10-23)

## Corpus Check
- 286 files · ~328,037 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3503 nodes · 17104 edges · 119 communities (97 shown, 22 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2026 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1c2624f2`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- Node
- Host
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
- Agent (`internal/agent/...`)
- net/netip.Prefix
- Net
- ui
- AdminError
- lab
- Snapshot
- nodeRow
- Command reference
- time.Time
- ServerURL
- runner
- newRig
- Service
- agent/agent_test.go
- Installer
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
- Registration
- README.md
- Host
- Input
- otherNet
- .Tx
- context.Context
- Compose and stack files
- officeHours
- PeerStat
- Client
- Version
- net/netip.Addr
- meshDNS
- Tx
- Net
- service_test.go
- .query
- .Apply
- UI
- Registration
- log/slog.Logger
- Name
- StepError
- New
- Exit
- MeshNetworkSetup
- newRig
- meshctl/docs_test.go
- resolved
- errorf
- Agent protocol
- TestPageGoldens
- deploy_test.go
- nullInt
- fetchwintun/main_test.go
- Security
- .add
- Ruleset
- service_windows.go
- Diff
- Error
- Container
- Extending the code
- userspace
- Test catalogue
- conn
- serveRuntime
- .queryRow
- Schedule
- Policy
- exprs
- CompilePolicy
- noKernel
- .Apply
- DNS
- `internal/validate`
- NewKernel
- ns
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
1. `Agent (`internal/agent/...`)` - 241 edges
2. `ui` - 90 edges
3. ``internal/server/api`` - 78 edges
4. `Session` - 74 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 66 edges
7. `newHarness()` - 56 edges
8. `env` - 48 edges
9. `runner` - 46 edges
10. `api` - 42 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the server (Dokploy)` --references--> `v()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js
- `Add a validator` --references--> `mask()`  [INFERRED]
  docs/developer.md → internal/validate/validate.go

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

## Communities (119 total, 22 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (23): composeNetwork, containerFile, containerService, nasFile, nasHealth, nasService, linkDNS, DNSRecord (+15 more)

### Community 1 - "NodeConfig"
Cohesion: 0.14
Nodes (37): rig, NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS(), newRig() (+29 more)

### Community 2 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 3 - "Host"
Cohesion: 0.15
Nodes (4): link, exitHost, Host, TestLinksAddressesAndFailures()

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.05
Nodes (96): watchdog, agentCommands(), agentLog(), joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall() (+88 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "AddrEntry"
Cohesion: 0.20
Nodes (10): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), parseSpans(), portRange, segResolver (+2 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.21
Nodes (12): Rule, Violation, Packages, and which way they may import, Input and commands, `internal/archtest`, Check(), importsOf(), ModuleRoot() (+4 more)

### Community 11 - "net.Listener"
Cohesion: 0.48
Nodes (5): bindBoth(), bindPair(), Listen(), Listen(), listenSocket()

### Community 12 - "AddrRange"
Cohesion: 0.22
Nodes (12): addrs(), contains(), ofFamily(), inRanges(), inRange(), AddrEntry(), AddrRange, MergeRanges() (+4 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.05
Nodes (100): countingDB, `internal/server/api`, ErrorBody, cli(), ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput() (+92 more)

### Community 15 - "Run"
Cohesion: 0.11
Nodes (21): Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New(), NewAdmin() (+13 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.14
Nodes (6): api, hostFamilies(), decodeJSON(), readBody(), writeJSON(), readSums()

### Community 17 - ".do"
Cohesion: 0.07
Nodes (30): IPv6Request, IPv6View, NetworkInfo, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile (+22 more)

### Community 18 - "PublishNames"
Cohesion: 0.16
Nodes (12): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), PublishNames(), DNSView (+4 more)

### Community 19 - "Dir"
Cohesion: 0.12
Nodes (17): TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip() (+9 more)

### Community 20 - "ServerConfig"
Cohesion: 0.24
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+17 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.05
Nodes (64): Agent (`internal/agent/...`), TestCleanHostname(), TestDefaultSocket(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig() (+56 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (32): Routes, NewTable(), compare(), Desired(), Table, Sync(), prefixes(), TestDesiredSkipsTheDefaultRoutesAndDuplicates() (+24 more)

### Community 24 - "ui"
Cohesion: 0.09
Nodes (19): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), anyText() (+11 more)

### Community 26 - "lab"
Cohesion: 0.15
Nodes (11): Coverage, host, deleteNamespaces(), eventually(), lab, stream, newLab(), tail() (+3 more)

### Community 27 - "Snapshot"
Cohesion: 0.15
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "nodeRow"
Cohesion: 0.40
Nodes (5): dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.06
Nodes (24): pendingApply, safeApply, syncBuffer, Failover, hubMove, hubState, replayCache, failures (+16 more)

### Community 31 - "ServerURL"
Cohesion: 0.29
Nodes (10): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams, isLoopbackHost() (+2 more)

### Community 32 - "runner"
Cohesion: 0.09
Nodes (13): fakeService, runner, kick(), oneLine(), Result, dial(), dial(), isPipe() (+5 more)

### Community 33 - "newRig"
Cohesion: 0.11
Nodes (41): brokenStream, Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, TestTemplatesHaveNoScript(), TestFirewallPages(), TestNodeGroupsOnPages(), group() (+33 more)

### Community 34 - "Service"
Cohesion: 0.14
Nodes (9): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+1 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (59): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+51 more)

### Community 36 - "Installer"
Cohesion: 0.32
Nodes (4): hostService(), Runner, Systemd(), Installer

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (29): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures() (+21 more)

### Community 38 - "testing.T"
Cohesion: 0.11
Nodes (48): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+40 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.12
Nodes (4): mustAffect(), joinList(), boolInt(), classify()

### Community 41 - "openTest"
Cohesion: 0.12
Nodes (48): `internal/server/store`, DefaultPolicy(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation() (+40 more)

### Community 42 - ".mesh"
Cohesion: 0.09
Nodes (20): Status, 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh (+12 more)

### Community 43 - "netlinkNet"
Cohesion: 0.08
Nodes (16): netlinkNet, Owned(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule(), IPNet() (+8 more)

### Community 44 - "wgdev"
Cohesion: 0.24
Nodes (3): TestUAPISetGolden(), UAPISet(), wgdev

### Community 45 - "planner_test.go"
Cohesion: 0.12
Nodes (43): Add a topology rule, `internal/server/planner`, NodeConfig, TestCheckRequested(), TestRangeFor(), Compute(), deliver(), deliverAddr() (+35 more)

### Community 46 - "Run"
Cohesion: 0.10
Nodes (12): Deps, Deps, Run(), DNS, Firewall, MeshFirewall, Sysctls, Source (+4 more)

### Community 47 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 48 - "README.md"
Cohesion: 0.14
Nodes (9): graphify, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together, meshctl (+1 more)

### Community 49 - "Host"
Cohesion: 0.08
Nodes (37): dialing, peer(), ParseAddr(), TestParseAddr(), AddrPort(), TestAddrPort(), hexKey(), ParseUAPI() (+29 more)

### Community 50 - "Input"
Cohesion: 0.11
Nodes (15): PlanError, Routes, Record, containers6(), Check(), checkContainers(), checkContainers6(), checkNode6() (+7 more)

### Community 52 - ".Tx"
Cohesion: 0.09
Nodes (29): Config, DB, Lockout, Params, Pending, Sealer, addUser(), newPassword() (+21 more)

### Community 53 - "context.Context"
Cohesion: 0.08
Nodes (25): statusDetail, dnsLine(), printExit(), Client, Logs, Handler(), logs(), refusal() (+17 more)

### Community 54 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 55 - "officeHours"
Cohesion: 0.50
Nodes (4): officeHours(), sgt(), TestPolicyRuleString(), TestPolicySchedules()

### Community 57 - "Client"
Cohesion: 0.18
Nodes (11): Options, apiError(), clip(), Client, Join(), New(), TestJoinIsUnsigned(), TestUnreachableAndMalformedAnswers() (+3 more)

### Community 59 - "net/netip.Addr"
Cohesion: 0.11
Nodes (20): hostAddrs, nameTable, zone, TestIsPublic(), records(), firstAddr(), isPublic(), isPublic6() (+12 more)

### Community 60 - "meshDNS"
Cohesion: 0.25
Nodes (22): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+14 more)

### Community 61 - "Tx"
Cohesion: 0.09
Nodes (45): apiError, Change, ComposeFile, ComposeOptions, PhoneRequest, Role, addIPv6(), blockOf() (+37 more)

### Community 62 - "Net"
Cohesion: 0.12
Nodes (23): network, Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard() (+15 more)

### Community 63 - "service_test.go"
Cohesion: 0.15
Nodes (12): applied, Firewall, policy, Firewall, atLeast(), normalize(), richRule(), RuleArgv() (+4 more)

### Community 65 - ".query"
Cohesion: 0.22
Nodes (4): ContainerName, scanAll(), splitList(), AuditEntry

### Community 66 - ".Apply"
Cohesion: 0.16
Nodes (13): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+5 more)

### Community 67 - "UI"
Cohesion: 0.10
Nodes (19): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize() (+11 more)

### Community 68 - "Registration"
Cohesion: 0.19
Nodes (15): nrpt, nrptRule, desired(), domainName(), fromRules(), nrptKey(), ownedKey(), parseServers() (+7 more)

### Community 69 - "log/slog.Logger"
Cohesion: 0.09
Nodes (16): Deps, Hub, subscription, NewUserspace(), NewUserspace(), NewUserspace(), ConfigEvent, DB (+8 more)

### Community 72 - "Name"
Cohesion: 0.14
Nodes (33): check(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), Endpoint6(), IP(), IPv6() (+25 more)

### Community 74 - "New"
Cohesion: 0.19
Nodes (8): newRenderer(), staticHandler(), headers(), New(), sameOrigin(), handler, page, renderer

### Community 77 - "MeshNetworkSetup"
Cohesion: 0.36
Nodes (8): containerHeader(), MeshNetworkSetup(), networkCommands(), resolvers(), serviceExample(), TestHostModeNetworkSetup(), NetworkParams, NetworkSetup

### Community 78 - "newRig"
Cohesion: 0.18
Nodes (13): capped, golden(), laptop(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage() (+5 more)

### Community 79 - "meshctl/docs_test.go"
Cohesion: 0.36
Nodes (7): cell(), renderCommands(), sentence(), TestCommandsDoc(), admit(), walk(), packet

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.09
Nodes (26): RecordRequest, segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange (+18 more)

### Community 83 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+2 more)

### Community 84 - "TestPageGoldens"
Cohesion: 0.09
Nodes (21): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), readZoneForm(), ruleRows(), scheduleText(), title() (+13 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.10
Nodes (43): hostSocket(), renderContainer(), RenderContainerCompose(), containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposeHeader(), TestContainerComposePosture() (+35 more)

### Community 86 - "nullInt"
Cohesion: 0.21
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.10
Nodes (27): `internal/fsutil`, `internal/tools/fetchwintun`, Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+19 more)

### Community 88 - "Security"
Cohesion: 0.15
Nodes (15): TestUnitPosture(), SystemdUnit(), Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens, Security (+7 more)

### Community 89 - ".add"
Cohesion: 0.16
Nodes (11): clip(), Ring, addrText(), Rule, merged(), portText(), services(), ServiceEntry() (+3 more)

### Community 90 - "Ruleset"
Cohesion: 0.15
Nodes (6): watchingFirewall, Kernel, Ruleset, Chain, Enforcer, Kernel

### Community 92 - "service_windows.go"
Cohesion: 0.11
Nodes (10): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows() (+2 more)

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 95 - "Container"
Cohesion: 0.13
Nodes (10): fakeRuntime, Container, Of(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack(), clip(), meshFilter() (+2 more)

### Community 97 - "Extending the code"
Cohesion: 0.14
Nodes (12): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+4 more)

### Community 98 - "userspace"
Cohesion: 0.12
Nodes (5): defaultInterface(), exitRoutes(), unspecified(), Link, userspace

### Community 99 - "Test catalogue"
Cohesion: 0.17
Nodes (18): `internal/server/phone`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue, Hub(), joinSet(), NewKey(), Render() (+10 more)

### Community 100 - "conn"
Cohesion: 0.43
Nodes (3): conn(), hasLegacy(), nftKernel

### Community 101 - "serveRuntime"
Cohesion: 0.23
Nodes (11): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+3 more)

### Community 103 - ".queryRow"
Cohesion: 0.16
Nodes (4): scanNode(), User, scanUser(), scanner

### Community 104 - "Schedule"
Cohesion: 0.21
Nodes (7): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), weekday()

### Community 105 - "Policy"
Cohesion: 0.16
Nodes (23): NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile(), Local, compile(), dualPolicy() (+15 more)

### Community 107 - "exprs"
Cohesion: 0.17
Nodes (8): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), TestIntervalAddrElems(), TestIntervalElems(), nftOp

### Community 108 - "CompilePolicy"
Cohesion: 0.16
Nodes (15): Matrix, MatrixCell, PolicyRule, Schedule, activeAt(), zoneOf(), CompilePolicy(), newSegResolver() (+7 more)

### Community 111 - ".Apply"
Cohesion: 0.18
Nodes (5): IPv6Host, Rechecker, Add an apply step, forHost(), Engine

### Community 113 - "DNS"
Cohesion: 0.17
Nodes (6): nameServer, DNS, Want, newNameTable(), DNS, MeshName

### Community 114 - "`internal/validate`"
Cohesion: 0.09
Nodes (27): Security invariants, `internal/validate`, checkDNS(), TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), DNSName() (+19 more)

### Community 118 - "ns"
Cohesion: 0.14
Nodes (12): End to end -- `test/e2e`, lab, TestContainerNetworks(), lab, TestDualStack(), inNS(), ns(), lab (+4 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.11
Nodes (28): `internal/server` and `internal/server/export`, NewAdminClient(), FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport() (+20 more)

### Community 134 - "`internal/proto`"
Cohesion: 0.06
Nodes (52): Determinism and goldens, `internal/proto`, newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse() (+44 more)

## Knowledge Gaps
- **6 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `lab` (+1 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 334 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **22 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does ``internal/server/api`` connect ``internal/server/api`` to `Test catalogue`, `htmx.min.js`, ``internal/server/auth``, ``internal/proto``, `Run`, `.do`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Are the 240 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 240 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _6 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08164562710017255 - nodes in this community are weakly interconnected._
- **Why does `at()` connect `htmx.min.js` to ``internal/server/api``?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.13630229419703105 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `.mesh`, `testing.T`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._