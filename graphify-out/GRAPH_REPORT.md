# Graph Report - vpn-mesh-utility  (2026-10-23)

## Corpus Check
- 285 files · ~324,535 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3485 nodes · 16984 edges · 117 communities (96 shown, 21 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2006 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `28eedecb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- IPNet
- Role
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
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
- nas_test.go
- runner
- newRig
- inNS
- agent/agent_test.go
- service_test.go
- testing.T
- testing.T
- Policy
- .exec
- openTest
- .mesh
- netlinkNet
- wgdev
- planner_test.go
- log/slog.Logger
- ui
- README.md
- Host
- container_test.go
- otherNet
- .Tx
- context.Context
- Compose and stack files
- Import
- NodeInfo
- Client
- .routes
- AddrRange
- Service
- Tx
- Net
- service_test.go
- Container
- .ExportOnce
- .renderVersion
- UI
- Registration
- Hub
- 7. Agent
- Segmentation
- Name
- StepError
- Table
- 8. The web UI
- Exit
- container.go
- newRig
- errorf
- Agent protocol
- TestPageGoldens
- deploy_test.go
- fetchwintun/main_test.go
- Security
- walk
- Ruleset
- service_windows.go
- Diff
- Error
- userspace
- serveRuntime
- .queryRow
- Schedule
- Policy
- exprs
- net/netip.Addr
- NodeConfig
- .Reach
- sync.Mutex
- `internal/validate`
- NewKernel
- ns
- loadMigrations
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
3. ``internal/server/api`` - 77 edges
4. `Session` - 74 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 64 edges
7. `newHarness()` - 55 edges
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

## Communities (117 total, 21 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, DNSRequest (+9 more)

### Community 1 - "NodeConfig"
Cohesion: 0.16
Nodes (32): rig, NewRoutes(), IsStepError(), New(), isStep(), meshDNS(), newRig(), TestApplyIsIdempotent() (+24 more)

### Community 2 - "IPNet"
Cohesion: 0.11
Nodes (10): Status, IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), TestCreateLinkNeedsPrivilege(), lookup(), Link (+2 more)

### Community 3 - "Role"
Cohesion: 0.18
Nodes (18): Role, blockOf(), Allocate(), Allocate6(), AllocateContainers(), AllocateContainers6(), Block(), Block6() (+10 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.05
Nodes (95): watchdog, agentCommands(), agentLog(), joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall() (+87 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (27): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments, 1. Problem (+19 more)

### Community 7 - "AddrEntry"
Cohesion: 0.18
Nodes (13): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), newSegResolver(), parseSpans(), ZoneAddresses() (+5 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.18
Nodes (14): Rule, Violation, Packages, and which way they may import, Input and commands, `internal/archtest`, `internal/tools/vulnjudge`, Test catalogue, Check() (+6 more)

### Community 11 - "net.Listener"
Cohesion: 0.18
Nodes (14): bindBoth(), bindPair(), dial(), Listen(), dial(), isPipe(), Listen(), listenPipe() (+6 more)

### Community 12 - "meshHarness"
Cohesion: 0.12
Nodes (21): apiError, ErrorMessage(), NewNodes(), toAPIError(), TestIPv6DNSRecords(), harness, meshHarness(), refused() (+13 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.08
Nodes (71): countingDB, `internal/server/api`, ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose() (+63 more)

### Community 15 - "Run"
Cohesion: 0.18
Nodes (15): Deps, New(), NewAdmin(), newAPI(), LoadOrCreateKey(), TestLoadOrCreateKey(), hostOf(), Routes() (+7 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.14
Nodes (11): api, decode(), reply(), ETag(), cli(), hostFamilies(), decodeJSON(), readBody() (+3 more)

### Community 17 - ".do"
Cohesion: 0.16
Nodes (11): IPv6Request, IPv6View, NetworkInfo, NodeAddRequest, NodeSetRequest, AdminClient, NewAdminClient(), nodePath() (+3 more)

### Community 18 - "PublishNames"
Cohesion: 0.15
Nodes (13): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), PublishNames(), DNSView (+5 more)

### Community 19 - "Dir"
Cohesion: 0.10
Nodes (21): `internal/fsutil`, Dir, Enrollment, mkdirs(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestKeysAreCreatedOnceAndKept(), TestLoadKeysRefusesLooseOrDamagedFiles() (+13 more)

### Community 20 - "ServerConfig"
Cohesion: 0.24
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+17 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.05
Nodes (54): Agent (`internal/agent/...`), capped, TestCleanHostname(), TestDefaultSocket(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine() (+46 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.13
Nodes (11): Routes, pfx(), TestNetHasKernelSemantics(), index(), route(), compare(), luidOf(), onLink() (+3 more)

### Community 24 - "ui"
Cohesion: 0.10
Nodes (16): Session, anyText(), listOf(), listText(), meshChanged(), objectRows(), title(), atoi() (+8 more)

### Community 26 - "lab"
Cohesion: 0.27
Nodes (6): Coverage, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - ".nodeView"
Cohesion: 0.21
Nodes (7): Health(), ago(), dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.09
Nodes (17): pendingApply, safeApply, Failover, hubMove, hubState, replayCache, failures, hitWindow (+9 more)

### Community 31 - "nas_test.go"
Cohesion: 0.27
Nodes (14): keys(), sameSet(), comment(), nasHeader(), renderNAS(), nasParams(), parseNAS(), TestNASComposeGolden() (+6 more)

### Community 32 - "runner"
Cohesion: 0.07
Nodes (16): Deps, fakeService, runner, Deps, resetTOTP(), kick(), oneLine(), Run() (+8 more)

### Community 33 - "newRig"
Cohesion: 0.11
Nodes (41): brokenStream, Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, TestTemplatesHaveNoScript(), TestFirewallPages(), TestNodeGroupsOnPages(), group() (+33 more)

### Community 34 - "inNS"
Cohesion: 0.19
Nodes (6): End to end -- `test/e2e`, stream, inNS(), lab, stream, TestMeshFirewall()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (59): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+51 more)

### Community 36 - "service_test.go"
Cohesion: 0.23
Nodes (10): TestUnitPosture(), SystemdUnit(), Agent service (systemd), exists(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere(), TestInstallWritesTheUnitAndStartsIt() (+2 more)

### Community 37 - "testing.T"
Cohesion: 0.05
Nodes (72): Params, rig, 11. CLI reference, `internal/server/auth`, TestDNS(), Serve(), serve(), shortDir() (+64 more)

### Community 38 - "testing.T"
Cohesion: 0.11
Nodes (43): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+35 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.09
Nodes (8): mustAffect(), DNSRecord, joinList(), boolInt(), classify(), nullInt(), nullStr(), StatusUpdate

### Community 41 - "openTest"
Cohesion: 0.12
Nodes (46): `internal/server/store`, DefaultPolicy(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation() (+38 more)

### Community 42 - ".mesh"
Cohesion: 0.14
Nodes (14): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+6 more)

### Community 43 - "netlinkNet"
Cohesion: 0.15
Nodes (8): netlinkNet, Owned(), TestGatewayRules(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule()

### Community 44 - "wgdev"
Cohesion: 0.24
Nodes (3): TestUAPISetGolden(), UAPISet(), wgdev

### Community 45 - "planner_test.go"
Cohesion: 0.05
Nodes (75): Matrix, MatrixCell, PlanError, Config and events, Add a topology rule, 4. Roles and topology, `internal/server/planner`, Peer (+67 more)

### Community 46 - "log/slog.Logger"
Cohesion: 0.18
Nodes (9): TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), Reconciler, NewReconciler(), Config, Options (+1 more)

### Community 47 - "ui"
Cohesion: 0.13
Nodes (8): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), writeHTML(), sameOrigin(), handler, page

### Community 48 - "README.md"
Cohesion: 0.15
Nodes (9): graphify, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together, meshctl (+1 more)

### Community 49 - "Host"
Cohesion: 0.06
Nodes (41): link, exitHost, TestRunNeedsAnEnrollment(), Host, NewHost(), peer(), TestConfigureHasKernelSemantics(), TestLinksAddressesAndFailures() (+33 more)

### Community 50 - "container_test.go"
Cohesion: 0.33
Nodes (11): containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposeHeader(), TestContainerComposePosture(), TestContainerComposeRefusesHostileValues(), TestContainerRendererHoldsItsStructure(), TestHostModeNetworkSetup() (+3 more)

### Community 52 - ".Tx"
Cohesion: 0.10
Nodes (22): Config, DB, Lockout, Pending, Sealer, addUser(), newPassword(), passwdUser() (+14 more)

### Community 53 - "context.Context"
Cohesion: 0.09
Nodes (23): statusDetail, Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest (+15 more)

### Community 54 - "Compose and stack files"
Cohesion: 0.15
Nodes (13): exampleContainer(), renderComposeDoc(), TestComposeDoc(), A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker (+5 more)

### Community 55 - "Import"
Cohesion: 0.24
Nodes (9): Document, Import(), document(), linuxOnly(), TestImportRefusesInvalidDocumentsUntouched(), TestImportRefusesWhileTheServerRuns(), TestImportRestoresAnExport(), TestLockIsExclusive() (+1 more)

### Community 56 - "NodeInfo"
Cohesion: 0.18
Nodes (10): NodeList, PhoneProfile, addresses(), heard(), withGroups(), NodeInfo, phoneHealth(), phoneState() (+2 more)

### Community 57 - "Client"
Cohesion: 0.14
Nodes (9): apiError(), clip(), Client, clip(), meshFilter(), StatusResponse, NamesResponse, Runtime (+1 more)

### Community 58 - ".routes"
Cohesion: 0.42
Nodes (7): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 59 - "AddrRange"
Cohesion: 0.13
Nodes (16): contains(), inRanges(), serves(), inRange(), AddrEntry(), AddrRange, Service, LastOf() (+8 more)

### Community 60 - "Service"
Cohesion: 0.08
Nodes (33): Registrar, resolver, Service, Apply (agent), desired(), New(), same(), answerOf() (+25 more)

### Community 61 - "Tx"
Cohesion: 0.11
Nodes (37): Change, ComposeFile, ComposeOptions, PhoneRequest, addIPv6(), checkGroups(), checkName(), containerSubnet() (+29 more)

### Community 62 - "Net"
Cohesion: 0.14
Nodes (20): Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard(), parse() (+12 more)

### Community 63 - "service_test.go"
Cohesion: 0.14
Nodes (12): applied, Firewall, policy, Firewall, Runner, normalize(), richRule(), RuleArgv() (+4 more)

### Community 64 - "Container"
Cohesion: 0.18
Nodes (10): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+2 more)

### Community 65 - ".ExportOnce"
Cohesion: 0.25
Nodes (6): DB, NextExport(), TestExporterRunStopsWithItsContext(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestNextExport(), Exporter

### Community 67 - "UI"
Cohesion: 0.11
Nodes (17): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), bytesize(), loadPages(), prefsFlash() (+9 more)

### Community 68 - "Registration"
Cohesion: 0.11
Nodes (19): Domain, fakeRegistrar, linkDomain, nrpt, nrptRule, Registration, resolved, domainName() (+11 more)

### Community 69 - "Hub"
Cohesion: 0.24
Nodes (3): Hub, subscription, ConfigEvent

### Community 70 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 72 - "Name"
Cohesion: 0.08
Nodes (52): RecordRequest, RenderContainerCompose(), RenderNASCompose(), check(), ParseAddr(), TestParseAddr(), validOrNot(), validateMeshNames() (+44 more)

### Community 74 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 75 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

### Community 77 - "container.go"
Cohesion: 0.23
Nodes (16): composeNetwork, containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), resolvers(), serviceExample() (+8 more)

### Community 78 - "newRig"
Cohesion: 0.24
Nodes (11): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+3 more)

### Community 81 - "errorf"
Cohesion: 0.12
Nodes (19): segLiteral, segDoc, Actor, segLiterals(), PolicyZone(), MeshChange, meshChange(), movePolicy() (+11 more)

### Community 83 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+2 more)

### Community 84 - "TestPageGoldens"
Cohesion: 0.08
Nodes (22): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), readZoneForm(), ruleRows(), scheduleText(), dnsState() (+14 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.24
Nodes (14): checkPosture(), mustRead(), service(), TestDevComposePosture(), TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues() (+6 more)

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.14
Nodes (20): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+12 more)

### Community 88 - "Security"
Cohesion: 0.20
Nodes (9): Changes a node asks for, DNS, Interfaces the agent did not create, Security, The host firewall, The local UI (internal/agent/localui), Web UI sign-in (internal/server/auth), What a node sees of the others (+1 more)

### Community 89 - "walk"
Cohesion: 0.23
Nodes (10): addrText(), Rule, merged(), portText(), services(), admit(), walk(), packet (+2 more)

### Community 90 - "Ruleset"
Cohesion: 0.11
Nodes (7): watchingFirewall, Kernel, Ruleset, Chain, Enforcer, Kernel, noKernel

### Community 92 - "service_windows.go"
Cohesion: 0.09
Nodes (13): hostService(), hostService(), uninstallAgent(), Service, Systemd(), commandLine(), RunWindowsService(), stop() (+5 more)

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 98 - "userspace"
Cohesion: 0.13
Nodes (5): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), userspace

### Community 101 - "serveRuntime"
Cohesion: 0.21
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 103 - ".queryRow"
Cohesion: 0.07
Nodes (10): scanNode(), scanAll(), splitList(), HashToken(), NewToken(), scanUser(), AuditEntry, PeerStat (+2 more)

### Community 104 - "Schedule"
Cohesion: 0.15
Nodes (10): The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds() (+2 more)

### Community 105 - "Policy"
Cohesion: 0.15
Nodes (25): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily() (+17 more)

### Community 107 - "exprs"
Cohesion: 0.07
Nodes (22): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+14 more)

### Community 110 - "net/netip.Addr"
Cohesion: 0.12
Nodes (18): network, nameTable, zone, TestIsPublic(), host(), parseServers(), serverList(), TestNRPTServers() (+10 more)

### Community 111 - "NodeConfig"
Cohesion: 0.06
Nodes (31): NodeDetail, dialing, IPv6Host, Rechecker, Determinism and goldens, Add an apply step, Keys and tokens, Shared fixtures -- `internal/testfix` (+23 more)

### Community 112 - ".Reach"
Cohesion: 0.18
Nodes (8): ReachRequest, SegmentView, ZoneView, ReachResult, zoneOf(), matrixCell, matrixRow, reachView

### Community 113 - "sync.Mutex"
Cohesion: 0.08
Nodes (11): nameServer, syncBuffer, DNS, Sysctl, Sysctls, Want, newNameTable(), clip() (+3 more)

### Community 114 - "`internal/validate`"
Cohesion: 0.10
Nodes (21): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), TestServiceEntry(), Arg(), PipePath() (+13 more)

### Community 118 - "ns"
Cohesion: 0.17
Nodes (9): host, lab, TestContainerNetworks(), deleteNamespaces(), newLab(), ns(), tail(), lab (+1 more)

### Community 119 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.20
Nodes (17): `internal/server` and `internal/server/export`, FromSnapshot(), dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes() (+9 more)

### Community 134 - "`internal/proto`"
Cohesion: 0.05
Nodes (60): Options, `internal/proto`, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents() (+52 more)

## Knowledge Gaps
- **5 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `stream`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 333 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `NodeConfig`, `IPNet`, `env`, ``internal/proto``, `archtest.go`, `net.Listener`, `net/http.ResponseWriter`, `Dir`, `ServerConfig`, `net/netip.Prefix`, `time.Time`, `newRig`, `agent/agent_test.go`, `service_test.go`, `testing.T`, `Policy`, `netlinkNet`, `wgdev`, `planner_test.go`, `log/slog.Logger`, `Host`, `.routes`, `AddrRange`, `Service`, `Net`, `Container`, `UI`, `Registration`, `Name`, `newRig`, `fetchwintun/main_test.go`, `service_windows.go`, `serveRuntime`, `Policy`, `exprs`, `net/netip.Addr`, `NodeConfig`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Are the 238 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 238 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _5 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08179699088789998 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `IPNet`, `.mesh`, `testing.T`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Should `IPNet` be split into smaller, more focused modules?**
  _Cohesion score 0.11067193675889328 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect ``internal/server/api`` to `agent/agent_test.go`, `htmx.min.js`, `serveRuntime`, `testing.T`, ``internal/proto``, `archtest.go`, `meshHarness`, `.do`, `.routes`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._