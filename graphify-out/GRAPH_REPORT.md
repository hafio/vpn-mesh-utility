# Graph Report - vpn-mesh-utility  (2026-10-20)

## Corpus Check
- 245 files · ~247,193 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 2993 nodes · 12875 edges · 122 communities (102 shown, 20 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 733 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6d76df37`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- openTest
- entry
- htmx.min.js
- env
- meshctl — Design Spec
- Extending the code
- dev.sh
- dev.ps1
- archtest.go
- wg.go
- Tx
- github.com/hafio/vpn-mesh-utility
- newHarness
- testing.T
- newDir
- Input
- io.Writer
- Dir
- .install
- WGKey
- net/netip.Prefix
- meshctl/main.go
- Health
- LocalStatus
- lab
- Handler
- Session
- Command reference
- time.Time
- Host
- context.Context
- newRig
- .AddRule
- agent/agent_test.go
- Sealer
- Test catalogue
- runArgs
- Policy
- Tx
- Snapshot
- Agent protocol
- net/http.Request
- newRig
- planner_test.go
- dns_test.go
- ServerConfig
- CLAUDE.md
- Security
- otherNet
- Service
- Client
- Registration
- TestContainerNamesAreReported
- CookieValue
- time.Duration
- NodeChange
- DNS
- Service
- errorf
- containers.go
- .Apply
- wgdev
- UI
- fakeServer
- session
- loadMigrations
- PhoneConfig
- toAPIError
- otherPolicy
- validate.go
- StepError
- Writer
- mustAffect
- net/netip.AddrPort
- container.go
- net/http.HandlerFunc
- 7. Agent
- nullInt
- WriteEvent
- .show
- netlinkNet
- .renderDNS
- NewUserspace
- FromSnapshot
- fetchwintun/main_test.go
- Document
- 6. Control plane
- 8. Server
- .Seen
- Join
- .names
- resolved
- Tx
- Tx
- NodeConfig
- userspace
- File
- meshctl
- serveRuntime
- .renderFirewall
- Ring
- .Restore
- fakeService
- log/slog.Logger
- Diff
- Node
- NewKernel
- net/netip.Addr
- droppedLogs
- NewNet
- commands.go
- DefaultSocket
- Import
- .validate
- Hub
- Runtime
- agent/agent.go
- service_windows.go
- IdentityKey

## God Nodes (most connected - your core abstractions)
1. `ServerConfig()` - 60 edges
2. `Session` - 55 edges
3. `NodeConfig` - 53 edges
4. `newHarness()` - 51 edges
5. `env` - 44 edges
6. `runner` - 39 edges
7. `ui` - 39 edges
8. `runArgs()` - 31 edges
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

## Communities (122 total, 20 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (28): AdminError, Error, nasFile, nasHealth, nasService, DNSRecord, FirewallRule, Node (+20 more)

### Community 1 - "newRig"
Cohesion: 0.14
Nodes (13): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+5 more)

### Community 2 - "openTest"
Cohesion: 0.13
Nodes (29): TestExportOnceSkipsUnchangedAndKeepsFourteen(), Store, Open(), insertTopology(), node(), openTest(), TestAudit(), TestContainerNames() (+21 more)

### Community 3 - "entry"
Cohesion: 0.11
Nodes (5): entry, NewNet(), NewPolicy(), NewPolicy(), NewPolicy()

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

### Community 11 - "wg.go"
Cohesion: 0.08
Nodes (41): TestRunNeedsAnEnrollment(), NewHost(), peer(), pfx(), TestConfigureHasKernelSemantics(), TestLinksAddressesAndFailures(), TestNetHasKernelSemantics(), AddrPort() (+33 more)

### Community 14 - "newHarness"
Cohesion: 0.08
Nodes (52): brokenStream, countingDB, ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode() (+44 more)

### Community 15 - "testing.T"
Cohesion: 0.04
Nodes (77): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams, TestDNS() (+69 more)

### Community 16 - "newDir"
Cohesion: 0.15
Nodes (20): enroll(), enrollAs(), newDir(), newFakeServer(), shortTemp(), TestRunExplainsRefusals(), TestRunRefusesAnUnusableDataDir(), NewRoutes() (+12 more)

### Community 17 - "Input"
Cohesion: 0.16
Nodes (11): Change, PlanError, Result, Check(), checkContainers(), Problem, sortedByID(), Compute() (+3 more)

### Community 18 - "io.Writer"
Cohesion: 0.17
Nodes (23): TestBytesize(), dnsLine(), printExit(), adminFlags(), cleared(), joinNonEmpty(), nodeAction(), nodeName() (+15 more)

### Community 19 - "Dir"
Cohesion: 0.10
Nodes (15): TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestLocalSettings(), TestMeshCache() (+7 more)

### Community 21 - "WGKey"
Cohesion: 0.12
Nodes (30): IsStepError(), New(), meshDNS(), newRig(), TestApplyIsIdempotent(), TestApplyStopsAtTheFailingStep(), TestApplyValidatesBeforeTouchingTheHost(), TestContainersStep() (+22 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.06
Nodes (28): Net, Routes, NewTable(), compare(), Desired(), Table, Sync(), prefixes() (+20 more)

### Community 23 - "meshctl/main.go"
Cohesion: 0.22
Nodes (10): passwords(), runWith(), TestUserCommands(), find(), isHelp(), main(), printUsage(), run() (+2 more)

### Community 24 - "Health"
Cohesion: 0.15
Nodes (11): TestJudgeHubs(), Health(), phoneHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeForm (+3 more)

### Community 25 - "LocalStatus"
Cohesion: 0.10
Nodes (18): Client, Logs, logs(), refusal(), laptop(), PrefsRequest, PrefsResponse, LocalMesh (+10 more)

### Community 26 - "lab"
Cohesion: 0.13
Nodes (16): host, stream, lab, TestContainerNetworks(), deleteNamespaces(), eventually(), lab, inNS() (+8 more)

### Community 27 - "Handler"
Cohesion: 0.11
Nodes (29): TestStatusAndAgentHealthcheck(), TestLogsCommand(), TestUICommand(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe() (+21 more)

### Community 28 - "Session"
Cohesion: 0.12
Nodes (13): CheckServerDomains(), TestCheckServerDomains(), Session, atoi(), before(), changed(), ui, stamp() (+5 more)

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (30): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+22 more)

### Community 30 - "time.Time"
Cohesion: 0.07
Nodes (18): pendingApply, safeApply, syncBuffer, replayCache, failures, hitWindow, Lockout, LockoutError (+10 more)

### Community 31 - "Host"
Cohesion: 0.13
Nodes (5): link, exitHost, rig, Host, Sysctl

### Community 32 - "context.Context"
Cohesion: 0.09
Nodes (16): Deps, Deps, runner, kick(), oneLine(), Run(), TestOneLine(), DNS (+8 more)

### Community 33 - "newRig"
Cohesion: 0.19
Nodes (24): TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled(), TestChangesNeedOriginAndCSRF(), TestContainerModePages(), TestDashboardFragmentAndExport() (+16 more)

### Community 34 - ".AddRule"
Cohesion: 0.31
Nodes (5): MeshChange, RecordRequest, RuleRequest, Nodes, meshChange()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.20
Nodes (29): rig, eventually(), freePort(), joinedTo(), start(), startLogged(), startWith(), TestDNSFailureIsReported() (+21 more)

### Community 37 - "Test catalogue"
Cohesion: 0.10
Nodes (20): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+12 more)

### Community 38 - "runArgs"
Cohesion: 0.12
Nodes (30): TestAgentRunRuntimeSocket(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentRunNeedsAJoin(), TestExportCommand(), TestHealthcheckServer() (+22 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (22): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+14 more)

### Community 40 - "Tx"
Cohesion: 0.21
Nodes (3): boolInt(), Session, Tx

### Community 41 - "Snapshot"
Cohesion: 0.18
Nodes (15): SnapNode, Snapshot, Compare(), describeDNS(), describeRules(), fields(), fingerprint(), names() (+7 more)

### Community 42 - "Agent protocol"
Cohesion: 0.15
Nodes (13): Agent protocol, Apply (agent), Config and events, Container names, Errors, Join, Local API, Mesh view (+5 more)

### Community 43 - "net/http.Request"
Cohesion: 0.14
Nodes (12): decode(), reply(), cli(), api, api, decodeJSON(), readBody(), writeJSON() (+4 more)

### Community 44 - "newRig"
Cohesion: 0.17
Nodes (16): rig, newRig(), TestAccountChangesEndSessions(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures(), TestLoginRefusals(), TestLoginUpgradesAnOldHash() (+8 more)

### Community 45 - "planner_test.go"
Cohesion: 0.18
Nodes (28): deliver(), deliverAddr(), encode(), firstDiff(), hasProblem(), lookup(), mustCompute(), node() (+20 more)

### Community 46 - "dns_test.go"
Cohesion: 0.21
Nodes (22): desired(), same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects() (+14 more)

### Community 47 - "ServerConfig"
Cohesion: 0.22
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestApplyIsIdempotent() (+17 more)

### Community 49 - "CLAUDE.md"
Cohesion: 0.25
Nodes (7): Determinism and goldens, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates

### Community 50 - "Security"
Cohesion: 0.12
Nodes (16): Agent service (systemd), Changes a node asks for, Container-mode agent, DNS, Input and commands, Interfaces the agent did not create, Keys and tokens, NAS container (+8 more)

### Community 52 - "Service"
Cohesion: 0.09
Nodes (30): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), resetTOTP() (+22 more)

### Community 53 - "Client"
Cohesion: 0.18
Nodes (11): Options, apiError(), clip(), Client, Join(), New(), TestJoinIsUnsigned(), TestUnreachableAndMalformedAnswers() (+3 more)

### Community 54 - "Registration"
Cohesion: 0.17
Nodes (14): Domain, fakeRegistrar, nrpt, nrptRule, Registration, domainName(), fromRules(), nrptKey() (+6 more)

### Community 55 - "TestContainerNamesAreReported"
Cohesion: 0.17
Nodes (12): fakeRuntime, newFakeRuntime(), TestContainerNamesAreReported(), TestContainerNamesNeedAContainerNetwork(), webContainer(), Container, Of(), TestOf() (+4 more)

### Community 56 - "CookieValue"
Cohesion: 0.23
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 57 - "time.Duration"
Cohesion: 0.16
Nodes (8): Failover, hubMove, hubState, waitApplied(), Exec, Tx, HashToken(), NewToken()

### Community 58 - "NodeChange"
Cohesion: 0.13
Nodes (16): NetworkInfo, NodeList, PhoneProfile, statusDetail, heard(), Rollback, AdminClient, NewAdminClient() (+8 more)

### Community 59 - "DNS"
Cohesion: 0.19
Nodes (5): nameServer, DNS, Want, newNameTable(), MeshName

### Community 60 - "Service"
Cohesion: 0.17
Nodes (7): Registrar, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange()

### Community 61 - "errorf"
Cohesion: 0.09
Nodes (28): apiError, ComposeFile, ComposeOptions, NodeAddRequest, NodeSetRequest, PhoneRequest, checkName(), containerSubnet() (+20 more)

### Community 62 - "containers.go"
Cohesion: 0.13
Nodes (23): network, Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard() (+15 more)

### Community 63 - ".Apply"
Cohesion: 0.06
Nodes (30): CustomRule, ServerFirewall, hostService(), TestUnitPosture(), SystemdUnit(), applied, Firewall, policy (+22 more)

### Community 64 - "wgdev"
Cohesion: 0.08
Nodes (9): lookup(), TestUAPISetGolden(), UAPISet(), userspace, NewUserspace(), Link, kernel, links (+1 more)

### Community 65 - "UI"
Cohesion: 0.15
Nodes (10): UI, hostOf(), New(), oldest(), validOrNot(), loadPages(), UIToken, Config (+2 more)

### Community 66 - "fakeServer"
Cohesion: 0.22
Nodes (10): fakeServer, TestMeshNamesReachTheResolver(), TestMeshViewIsRefreshed(), StatusReport, TestStatusReportValidate(), validStatus(), MeshView, TestMeshViewValidate() (+2 more)

### Community 67 - "session"
Cohesion: 0.34
Nodes (5): bytesize(), UI, prefsFlash(), page, session

### Community 68 - "loadMigrations"
Cohesion: 0.27
Nodes (5): Store, Tx, loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 69 - "PhoneConfig"
Cohesion: 0.24
Nodes (11): Hub(), NewKey(), Render(), aliceKey(), golden(), TestHub(), TestNewKey(), TestProfiles() (+3 more)

### Community 70 - "toAPIError"
Cohesion: 0.22
Nodes (10): ErrorMessage(), TestAdminCompose(), toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion(), harness, meshHarness(), refused() (+2 more)

### Community 72 - "validate.go"
Cohesion: 0.07
Nodes (57): check(), Containers, DNS, Firewall, Interface, NodeConfig, validateMeshNames(), newNode() (+49 more)

### Community 74 - "Writer"
Cohesion: 0.40
Nodes (3): New(), New(), Writer

### Community 75 - "mustAffect"
Cohesion: 0.23
Nodes (4): Tx, mustAffect(), classify(), nullStr()

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.15
Nodes (12): resolver, bindBoth(), bindPair(), serve(), Listen(), Listen(), listenSocket(), Listen() (+4 more)

### Community 77 - "container.go"
Cohesion: 0.09
Nodes (43): composeNetwork, containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers() (+35 more)

### Community 78 - "net/http.HandlerFunc"
Cohesion: 0.26
Nodes (8): Route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented(), ui, sameOrigin(), handler

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "nullInt"
Cohesion: 0.16
Nodes (6): Tx, DNSRecord, Tx, nullInt(), AuditEntry, FirewallRule

### Community 81 - "WriteEvent"
Cohesion: 0.26
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 82 - ".show"
Cohesion: 0.27
Nodes (4): ui, writeHTML(), page, renderer

### Community 83 - "netlinkNet"
Cohesion: 0.11
Nodes (13): netlinkNet, blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule(), IPNet(), Prefix() (+5 more)

### Community 84 - ".renderDNS"
Cohesion: 0.36
Nodes (3): ui, dnsData, recordForm

### Community 86 - "FromSnapshot"
Cohesion: 0.36
Nodes (7): FromSnapshot(), snapshot(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes(), TestToSnapshotRefusesHostileDocuments(), TestToSnapshotRoundTrips(), counter

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.12
Nodes (19): NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall(), download() (+11 more)

### Community 88 - "Document"
Cohesion: 0.18
Nodes (5): watchdog, Document, Read(), Write(), ImportConfig

### Community 89 - "6. Control plane"
Cohesion: 0.40
Nodes (5): 6.1 Enrollment, 6.2 Request authentication, 6.3 Config sync, 6.4 Safe apply, 6. Control plane

### Community 90 - "8. Server"
Cohesion: 0.40
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 92 - "Join"
Cohesion: 0.27
Nodes (8): installAgent(), installAgent(), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), cleanHostname(), JoinOptions, Service, Join()

### Community 93 - ".names"
Cohesion: 0.14
Nodes (15): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), api, LoadNames(), PublishNames() (+7 more)

### Community 94 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 95 - "Tx"
Cohesion: 0.33
Nodes (3): Tx, PeerStat, StatusUpdate

### Community 97 - "NodeConfig"
Cohesion: 0.11
Nodes (10): NodeDetail, Rechecker, Engine, Result, JoinResponse, PeerStatus, NodeConfig, Role (+2 more)

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

### Community 106 - "log/slog.Logger"
Cohesion: 0.11
Nodes (24): Deps, NewNodes(), api, New(), NewAdmin(), newAPI(), Reconciler, NewReconciler() (+16 more)

### Community 107 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 108 - "Node"
Cohesion: 0.22
Nodes (9): Peer, Record, Sysctl, SysctlKeys(), ContainerNetwork(), Node, sortPeers(), sortPrefixes() (+1 more)

### Community 110 - "net/netip.Addr"
Cohesion: 0.15
Nodes (13): nameTable, zone, TestIsPublic(), parseServers(), serverList(), TestNRPTServers(), isPublic(), PublicIPv4() (+5 more)

### Community 113 - "commands.go"
Cohesion: 0.21
Nodes (15): agentCommands(), argSynopsis(), commands(), newFlags(), parseArgs(), parseFlags(), cell(), renderCommands() (+7 more)

### Community 118 - "Import"
Cohesion: 0.24
Nodes (12): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses(), Import(), document(), linuxOnly() (+4 more)

### Community 121 - "Hub"
Cohesion: 0.27
Nodes (3): Hub, subscription, ConfigEvent

### Community 122 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 126 - "agent/agent.go"
Cohesion: 0.13
Nodes (5): linkDNS, TestOpen(), Open(), NewKernel(), TestOpenRefusesUnknownBackends()

### Community 129 - "service_windows.go"
Cohesion: 0.09
Nodes (13): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), executes(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths() (+5 more)

### Community 134 - "IdentityKey"
Cohesion: 0.14
Nodes (20): newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestNamesRoundTrip(), TestPrefsRoundTrip(), TestStatusRoundTrip() (+12 more)

## Knowledge Gaps
- **165 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `listed`, `inspected` (+160 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 385 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NodeConfig` connect `NodeConfig` to `ui/ui_test.go`, `wg.go`, `newDir`, `Input`, `Dir`, `WGKey`, `net/netip.Prefix`, `time.Time`, `context.Context`, `agent/agent_test.go`, `planner_test.go`, `ServerConfig`, `Client`, `errorf`, `containers.go`, `fakeServer`, `PhoneConfig`, `toAPIError`, `Node`, `Hub`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Why does `resolver` connect `net/netip.AddrPort` to `Service`, `agent/agent.go`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **Why does `Host` connect `Host` to `wg.go`, `agent/agent_test.go`, `agent/agent.go`, `time.Time`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _165 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09311235596319624 - nodes in this community are weakly interconnected._
- **Should `newRig` be split into smaller, more focused modules?**
  _Cohesion score 0.14130434782608695 - nodes in this community are weakly interconnected._
- **Should `openTest` be split into smaller, more focused modules?**
  _Cohesion score 0.13109243697478992 - nodes in this community are weakly interconnected._