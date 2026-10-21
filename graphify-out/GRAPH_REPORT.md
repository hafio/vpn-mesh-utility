# Graph Report - vpn-mesh-utility  (2026-10-21)

## Corpus Check
- 247 files · ~250,937 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3011 nodes · 12949 edges · 127 communities (100 shown, 27 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 744 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `2bdc9226`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- Run
- entry
- htmx.min.js
- env
- meshctl — Design Spec
- Extending the code
- dev.sh
- dev.ps1
- archtest.go
- Key
- Tx
- github.com/hafio/vpn-mesh-utility
- newHarness
- testing.T
- net/http.Request
- NodeChange
- io.Writer
- Dir
- nas.go
- ServerConfig
- net/netip.Prefix
- meshctl/main.go
- .nodeView
- LocalStatus
- lab
- versions/versions.go
- Session
- Command reference
- time.Time
- Host
- context.Context
- newRig
- .Firewall
- agent/agent_test.go
- Sealer
- Test catalogue
- runArgs
- Policy
- Tx
- store_test.go
- Agent protocol
- netlinkNet
- newRig
- planner_test.go
- dns_test.go
- CookieValue
- CLAUDE.md
- Security
- otherNet
- Service
- Client
- nrpt.go
- TestContainerNamesAreReported
- net/http.ResponseWriter
- .Tx
- Runtime
- DNS
- Service
- errorf
- Sync
- firewall_test.go
- wgdev
- UI
- fakeServer
- session
- loadMigrations
- Hub
- NewNodes
- otherPolicy
- fail
- StepError
- Writer
- mustAffect
- net/netip.AddrPort
- container.go
- Ring
- 7. Agent
- nullInt
- DNSName
- .show
- IPNet
- changed
- deploy_test.go
- .BeginEnrollment
- fetchwintun/main_test.go
- io.Reader
- 6. Control plane
- 8. Server
- .Seen
- Join
- .names
- resolved
- Version
- Tx
- userspace
- File
- meshctl
- serveRuntime
- Allocate
- NodeConfig
- Deps
- Registration
- NewKernel
- net/netip.Addr
- DefaultSocket
- NewNet
- commands.go
- Arg
- Import
- log/slog.Logger
- TestMain
- AdminError
- NewKernel
- Error
- service_windows.go
- NewNet
- NewPolicy
- NewPolicy
- NewPolicy
- IdentityKey
- Refusal
- StatusError
- Error
- ExitRouter

## God Nodes (most connected - your core abstractions)
1. `ServerConfig()` - 60 edges
2. `Session` - 55 edges
3. `NodeConfig` - 53 edges
4. `newHarness()` - 52 edges
5. `env` - 44 edges
6. `ui` - 40 edges
7. `runner` - 39 edges
8. `runArgs()` - 32 edges
9. `joinedTo()` - 31 edges
10. `eventually()` - 31 edges

## Surprising Connections (you probably didn't know these)
- `Status` --references--> `on()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/static/htmx.min.js
- ``cmd/meshctl`` --references--> `on()`  [INFERRED]
  docs/test.md → internal/server/ui/static/htmx.min.js
- `1. Run the server (Dokploy)` --references--> `v()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js
- `6. DNS` --references--> `on()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js
- `main()` --calls--> `stop()`  [INFERRED]
  cmd/meshctl/main.go → internal/agent/service/service_windows.go

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
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/sign_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`

## Communities (127 total, 27 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, Exit, Interface, logsData, inspected (+9 more)

### Community 1 - "newRig"
Cohesion: 0.17
Nodes (12): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+4 more)

### Community 2 - "Run"
Cohesion: 0.11
Nodes (19): LoadOrCreateKey(), NewSealer(), TestLoadOrCreateKey(), TestSealerBindsTheSecretToItsUser(), hostOf(), ListenAdmin(), Routes(), Run() (+11 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (119): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+111 more)

### Community 5 - "env"
Cohesion: 0.23
Nodes (29): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+21 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.11
Nodes (18): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments (+10 more)

### Community 7 - "Extending the code"
Cohesion: 0.15
Nodes (13): Add a command, Add a database migration, Add a topology rule, Add a validator, Add a web UI page, Add an agent API route, Add an apply step, Dev script tasks (+5 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.27
Nodes (9): Rule, Violation, Check(), importsOf(), ModuleRoot(), TestCheckerCatchesViolations(), TestModuleBoundaries(), under() (+1 more)

### Community 11 - "Key"
Cohesion: 0.08
Nodes (36): peer(), pfx(), TestConfigureHasKernelSemantics(), TestNetHasKernelSemantics(), AddrPort(), TestAddrPort(), hexKey(), ParseUAPI() (+28 more)

### Community 14 - "newHarness"
Cohesion: 0.08
Nodes (54): brokenStream, countingDB, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes() (+46 more)

### Community 15 - "testing.T"
Cohesion: 0.04
Nodes (85): TestStatusAndAgentHealthcheck(), TestLogsCommand(), TestUICommand(), TestDNS(), TestPolicyHasKernelSemantics(), TestKernelNetWithoutPrivilege(), TestRegistrarNeedsTheInterface(), allowed() (+77 more)

### Community 16 - "net/http.Request"
Cohesion: 0.15
Nodes (12): Role, cli(), api, api, decodeJSON(), readBody(), writeJSON(), api (+4 more)

### Community 17 - "NodeChange"
Cohesion: 0.12
Nodes (18): NetworkInfo, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, statusDetail, heard() (+10 more)

### Community 18 - "io.Writer"
Cohesion: 0.14
Nodes (26): TestBytesize(), TestPrintEnrollmentInTheLocalZone(), TestPrintStatusOfALaptop(), dnsLine(), adminFlags(), cleared(), joinNonEmpty(), nodeAction() (+18 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (16): TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestLocalSettings(), TestMeshCache() (+8 more)

### Community 20 - "nas.go"
Cohesion: 0.18
Nodes (16): comment(), nasHeader(), renderNAS(), RenderNASCompose(), nasParams(), TestNASComposeGolden(), TestNASComposeRefusesHostileValues(), nasFile (+8 more)

### Community 21 - "ServerConfig"
Cohesion: 0.13
Nodes (31): rig, TestRunNeedsAnEnrollment(), NewHost(), NewRoutes(), TestRoutesSysctlAndFirewall(), IsStepError(), New(), meshDNS() (+23 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (16): Net, Routes, host(), NewTable(), compare(), Desired(), Table, Sync() (+8 more)

### Community 23 - "meshctl/main.go"
Cohesion: 0.23
Nodes (13): agentCommands(), commands(), healthCommands(), find(), isHelp(), main(), printUsage(), run() (+5 more)

### Community 24 - ".nodeView"
Cohesion: 0.17
Nodes (10): Health(), phoneHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeForm, nodeRow (+2 more)

### Community 25 - "LocalStatus"
Cohesion: 0.09
Nodes (22): printExit(), Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest (+14 more)

### Community 26 - "lab"
Cohesion: 0.13
Nodes (16): host, stream, lab, TestContainerNetworks(), deleteNamespaces(), eventually(), lab, inNS() (+8 more)

### Community 27 - "versions/versions.go"
Cohesion: 0.09
Nodes (29): FromSnapshot(), snapshot(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes(), TestToSnapshotRefusesHostileDocuments(), TestToSnapshotRoundTrips(), SnapNode, Snapshot (+21 more)

### Community 28 - "Session"
Cohesion: 0.14
Nodes (9): CheckServerDomains(), TestCheckServerDomains(), Session, atoi(), before(), ui, versionNumber(), message() (+1 more)

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (30): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+22 more)

### Community 30 - "time.Time"
Cohesion: 0.05
Nodes (25): pendingApply, safeApply, syncBuffer, Failover, hubMove, hubState, replayCache, failures (+17 more)

### Community 31 - "Host"
Cohesion: 0.15
Nodes (5): link, exitHost, Host, TestLinksAddressesAndFailures(), PeerStats

### Community 32 - "context.Context"
Cohesion: 0.06
Nodes (21): Deps, fakeService, Deps, Rechecker, runner, kick(), oneLine(), Run() (+13 more)

### Community 33 - "newRig"
Cohesion: 0.16
Nodes (25): Store, TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled(), TestChangesNeedOriginAndCSRF(), TestContainerModePages() (+17 more)

### Community 34 - ".Firewall"
Cohesion: 0.33
Nodes (5): CustomRule, ServerFirewall, FirewallView, firewallData, ruleForm

### Community 35 - "agent/agent_test.go"
Cohesion: 0.14
Nodes (45): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), shortTemp() (+37 more)

### Community 37 - "Test catalogue"
Cohesion: 0.10
Nodes (20): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+12 more)

### Community 38 - "runArgs"
Cohesion: 0.11
Nodes (31): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentRunNeedsAJoin() (+23 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (22): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+14 more)

### Community 40 - "Tx"
Cohesion: 0.18
Nodes (5): boolInt(), Session, Tx, User, scanUser()

### Community 41 - "store_test.go"
Cohesion: 0.15
Nodes (33): TestReconcilerVersionsOnlyRealChanges(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), DecodeSnapshot(), Open(), insertTopology(), node(), openTest(), TestAudit() (+25 more)

### Community 42 - "Agent protocol"
Cohesion: 0.15
Nodes (13): Agent protocol, Apply (agent), Config and events, Container names, Errors, Join, Local API, Mesh view (+5 more)

### Community 43 - "netlinkNet"
Cohesion: 0.18
Nodes (7): netlinkNet, Owned(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule()

### Community 44 - "newRig"
Cohesion: 0.19
Nodes (14): rig, newRig(), TestAccountChangesEndSessions(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures(), TestLoginRefusals(), TestLoginUpgradesAnOldHash() (+6 more)

### Community 45 - "planner_test.go"
Cohesion: 0.08
Nodes (45): PlanError, Peer, Record, Check(), checkContainers(), Problem, sortedByID(), Compute() (+37 more)

### Community 46 - "dns_test.go"
Cohesion: 0.25
Nodes (19): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceAnswersTheContainers() (+11 more)

### Community 47 - "CookieValue"
Cohesion: 0.21
Nodes (6): CookieValue(), ui, group(), ui, sameOrigin(), handler

### Community 49 - "CLAUDE.md"
Cohesion: 0.25
Nodes (7): Determinism and goldens, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates

### Community 50 - "Security"
Cohesion: 0.12
Nodes (16): Agent service (systemd), Changes a node asks for, Container-mode agent, DNS, Input and commands, Interfaces the agent did not create, Keys and tokens, NAS container (+8 more)

### Community 52 - "Service"
Cohesion: 0.13
Nodes (21): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), resetTOTP() (+13 more)

### Community 53 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 54 - "nrpt.go"
Cohesion: 0.18
Nodes (16): nrpt, nrptRule, domainName(), same(), TestSameRegistration(), fromRules(), nrptKey(), ownedKey() (+8 more)

### Community 55 - "TestContainerNamesAreReported"
Cohesion: 0.17
Nodes (12): fakeRuntime, newFakeRuntime(), TestContainerNamesAreReported(), TestContainerNamesNeedAContainerNetwork(), webContainer(), Container, Of(), TestOf() (+4 more)

### Community 56 - "net/http.ResponseWriter"
Cohesion: 0.16
Nodes (7): decode(), reply(), api, readSums(), TestCookieAndCSRF(), ClearCookie(), SetCookie()

### Community 57 - ".Tx"
Cohesion: 0.30
Nodes (6): fullExpiry(), load(), session(), hashCookie(), newSecret(), matchStep()

### Community 58 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 59 - "DNS"
Cohesion: 0.17
Nodes (6): nameServer, DNS, desired(), Want, newNameTable(), MeshName

### Community 60 - "Service"
Cohesion: 0.17
Nodes (7): Registrar, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange()

### Community 61 - "errorf"
Cohesion: 0.09
Nodes (31): ComposeFile, ComposeOptions, MeshChange, PhoneRequest, RecordRequest, checkName(), containerSubnet(), defaultMTU() (+23 more)

### Community 62 - "Sync"
Cohesion: 0.13
Nodes (21): network, Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard() (+13 more)

### Community 63 - "firewall_test.go"
Cohesion: 0.06
Nodes (47): hostService(), TestUnitPosture(), SystemdUnit(), applied, fakeRunner, Firewall, policy, Firewall (+39 more)

### Community 64 - "wgdev"
Cohesion: 0.17
Nodes (5): TestUAPISetGolden(), UAPISet(), userspace, NewUserspace(), wgdev

### Community 65 - "UI"
Cohesion: 0.20
Nodes (7): UI, hostOf(), oldest(), validOrNot(), LocalToken(), TestLocalTokenMasksTheValue(), handler

### Community 66 - "fakeServer"
Cohesion: 0.17
Nodes (13): fakeServer, newFakeServer(), TestMeshNamesReachTheResolver(), TestMeshViewIsCached(), TestMeshViewIsRefreshed(), ErrorBody, StatusReport, TestStatusReportValidate() (+5 more)

### Community 67 - "session"
Cohesion: 0.34
Nodes (5): bytesize(), UI, prefsFlash(), page, session

### Community 68 - "loadMigrations"
Cohesion: 0.31
Nodes (4): Store, Tx, loadMigrations(), migration

### Community 69 - "Hub"
Cohesion: 0.24
Nodes (3): Hub, subscription, ConfigEvent

### Community 70 - "NewNodes"
Cohesion: 0.16
Nodes (12): apiError, ErrorMessage(), NewNodes(), TestAdminCompose(), toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion(), harness (+4 more)

### Community 72 - "fail"
Cohesion: 0.10
Nodes (34): RuleRequest, check(), Containers, Exit, Firewall, Interface, NodeConfig, validateMeshNames() (+26 more)

### Community 74 - "Writer"
Cohesion: 0.29
Nodes (4): New(), New(), SysctlKeys(), Writer

### Community 75 - "mustAffect"
Cohesion: 0.12
Nodes (8): Tx, mustAffect(), scanNode(), Tx, classify(), PeerStat, scanner, StatusUpdate

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.15
Nodes (12): resolver, bindBoth(), bindPair(), serve(), Listen(), Listen(), listenSocket(), Listen() (+4 more)

### Community 77 - "container.go"
Cohesion: 0.17
Nodes (22): composeNetwork, containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers() (+14 more)

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "nullInt"
Cohesion: 0.12
Nodes (8): Tx, DNSRecord, Tx, Tx, nullInt(), nullStr(), AuditEntry, FirewallRule

### Community 81 - "DNSName"
Cohesion: 0.33
Nodes (6): DNS, checkDNS(), DNSName(), Domains(), TestDNSName(), TestDomains()

### Community 82 - ".show"
Cohesion: 0.19
Nodes (6): loadPages(), ui, writeHTML(), pages, page, renderer

### Community 83 - "IPNet"
Cohesion: 0.09
Nodes (11): IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), index(), route(), lookup(), Link (+3 more)

### Community 84 - "changed"
Cohesion: 0.15
Nodes (8): ui, ui, changed(), ui, newQR(), TestQRCode(), phoneData, qrCode

### Community 85 - "deploy_test.go"
Cohesion: 0.20
Nodes (20): parseContainer(), TestContainerComposePosture(), checkPosture(), keys(), mustRead(), sameSet(), service(), TestDevComposePosture() (+12 more)

### Community 86 - ".BeginEnrollment"
Cohesion: 0.53
Nodes (3): newTOTPKey(), TestMatchStepAllowsOneStepOfSkew(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.13
Nodes (19): NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall(), download() (+11 more)

### Community 89 - "6. Control plane"
Cohesion: 0.40
Nodes (5): 6.1 Enrollment, 6.2 Request authentication, 6.3 Config sync, 6.4 Safe apply, 6. Control plane

### Community 90 - "8. Server"
Cohesion: 0.40
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 92 - "Join"
Cohesion: 0.24
Nodes (9): installAgent(), hostService(), installAgent(), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), cleanHostname(), JoinOptions, Service (+1 more)

### Community 93 - ".names"
Cohesion: 0.11
Nodes (17): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), api, LoadNames(), PublishNames() (+9 more)

### Community 94 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 95 - "Version"
Cohesion: 0.67
Nodes (3): SemVer(), TestVersionAndSemVer(), Version()

### Community 98 - "userspace"
Cohesion: 0.19
Nodes (4): defaultInterface(), exitRoutes(), userspace, unspecified()

### Community 99 - "File"
Cohesion: 0.24
Nodes (6): agentLog(), TestAgentLog(), ttyPassword(), OpenFile(), TestFileRotates(), File

### Community 100 - "meshctl"
Cohesion: 0.40
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 101 - "serveRuntime"
Cohesion: 0.20
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 102 - "Allocate"
Cohesion: 0.27
Nodes (11): Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateContainers(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan() (+3 more)

### Community 103 - "NodeConfig"
Cohesion: 0.10
Nodes (18): NodeConfig, goldenConfig(), TestHashIgnoresVersionOnly(), TestHashIsPinned(), TestValidateAcceptsFixtures(), withDomains(), Hub(), NewKey() (+10 more)

### Community 106 - "Deps"
Cohesion: 0.18
Nodes (15): Deps, Route, AdminRoutes(), describe(), api, New(), NewAdmin(), newAPI() (+7 more)

### Community 107 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 110 - "net/netip.Addr"
Cohesion: 0.20
Nodes (8): nameTable, zone, TestIsPublic(), isPublic(), PublicIPv4(), hostAddr(), TestHostAddr(), Record

### Community 113 - "commands.go"
Cohesion: 0.20
Nodes (13): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), shownDefault(), cell() (+5 more)

### Community 114 - "Arg"
Cohesion: 0.40
Nodes (4): Arg(), FuzzArg(), FuzzName(), TestArg()

### Community 118 - "Import"
Cohesion: 0.24
Nodes (12): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses(), Import(), document(), linuxOnly() (+4 more)

### Community 121 - "log/slog.Logger"
Cohesion: 0.08
Nodes (20): Change, TestImportCommand(), TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), DB, Reconciler (+12 more)

### Community 129 - "service_windows.go"
Cohesion: 0.12
Nodes (9): uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows(), handler (+1 more)

### Community 134 - "IdentityKey"
Cohesion: 0.06
Nodes (44): Options, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents(), TestEventsRefusesAnythingButAStream() (+36 more)

## Knowledge Gaps
- **165 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `listed`, `inspected` (+160 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 388 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **27 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Dir` connect `Dir` to `context.Context`, `ui/ui_test.go`, `agent/agent_test.go`, `LocalStatus`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **Why does `Service` connect `Service` to `ui/ui_test.go`, `net/netip.AddrPort`, `net/netip.Addr`, `dns_test.go`, `log/slog.Logger`, `DNS`, `time.Time`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **Why does `fakeServer` connect `fakeServer` to `agent/agent_test.go`, `Hub`, `NodeConfig`, `testing.T`, `TestContainerNamesAreReported`, `net/http.ResponseWriter`, `LocalStatus`, `time.Time`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _165 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08006831424552943 - nodes in this community are weakly interconnected._
- **Should `Run` be split into smaller, more focused modules?**
  _Cohesion score 0.10952380952380952 - nodes in this community are weakly interconnected._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.06358543417366946 - nodes in this community are weakly interconnected._