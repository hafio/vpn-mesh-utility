# Graph Report - vpn-mesh-utility  (2026-10-21)

## Corpus Check
- 245 files · ~249,709 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3002 nodes · 12912 edges · 139 communities (108 shown, 31 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 741 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `0aa00e3e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- newHarness
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
- storeNode
- testing.T
- net/http.Request
- NodeChange
- adminFlags
- Dir
- .install
- WGKey
- net/netip.Prefix
- meshctl/main.go
- .nodeView
- LocalStatus
- lab
- Handler
- Session
- Command reference
- time.Time
- Host
- runner
- newRig
- errorf
- agent/agent_test.go
- Sealer
- Test catalogue
- runArgs
- Policy
- Tx
- store_test.go
- Agent protocol
- net/http.ResponseWriter
- newRig
- planner_test.go
- meshDNS
- firewall_test.go
- CLAUDE.md
- Security
- otherNet
- Service
- Client
- nrpt.go
- Container
- CookieValue
- Node
- context.Context
- DNS
- Service
- Tx
- Sync
- service_test.go
- wgdev
- UI
- netlinkNet
- session
- loadMigrations
- NodeConfig
- meshHarness
- otherPolicy
- fail
- StepError
- Writer
- mustAffect
- net/netip.AddrPort
- container.go
- net/http.HandlerFunc
- 7. Agent
- nullInt
- NewReader
- .show
- IPNet
- .renderDNS
- deploy_test.go
- Net
- fetchwintun/main_test.go
- io.Reader
- 6. Control plane
- 8. Server
- .Seen
- Join
- .names
- resolved
- Tx
- Tx
- Host
- userspace
- File
- meshctl
- serveRuntime
- Allocate
- PhoneConfig
- Exporter
- Tx
- Run
- Registration
- Table
- NewKernel
- net/netip.Addr
- Ring
- NewNet
- commands.go
- Arg
- Input
- meshctl/docs_test.go
- syncBuffer
- NewHub
- .validate
- serve
- log/slog.Logger
- printStatus
- changed
- TestMain
- AdminError
- NewKernel
- Error
- .exec
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

## Communities (139 total, 31 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, Exit, Interface, logsData, inspected (+9 more)

### Community 1 - "newRig"
Cohesion: 0.24
Nodes (13): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+5 more)

### Community 2 - "newHarness"
Cohesion: 0.10
Nodes (30): apiError, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose(), TestAdminContainerMode() (+22 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (119): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+111 more)

### Community 5 - "env"
Cohesion: 0.22
Nodes (30): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+22 more)

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
Cohesion: 0.10
Nodes (33): TestRunNeedsAnEnrollment(), NewHost(), peer(), pfx(), TestConfigureHasKernelSemantics(), TestLinksAddressesAndFailures(), TestNetHasKernelSemantics(), hexKey() (+25 more)

### Community 14 - "storeNode"
Cohesion: 0.18
Nodes (22): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+14 more)

### Community 15 - "testing.T"
Cohesion: 0.03
Nodes (76): TestCleanHostname(), TestDefaultSocket(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestOneLine(), TestOpenRuntime(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig() (+68 more)

### Community 16 - "net/http.Request"
Cohesion: 0.17
Nodes (8): ETag(), TestETagQuotesTheHash(), api, decodeJSON(), writeJSON(), ClientIP(), inAny(), TestClientIP()

### Community 17 - "NodeChange"
Cohesion: 0.19
Nodes (10): NodeDetail, NodeList, PhoneProfile, heard(), PeerStatus, AdminClient, nodePath(), NodeChange (+2 more)

### Community 18 - "adminFlags"
Cohesion: 0.20
Nodes (20): TestPrintEnrollmentInTheLocalZone(), adminFlags(), cleared(), joinNonEmpty(), nodeAction(), nodeName(), nodeState(), printChange() (+12 more)

### Community 19 - "Dir"
Cohesion: 0.15
Nodes (7): Dir, Enrollment, syncDir(), TestWriteAtomicReplacesAndLeavesNoTempFiles(), WriteAtomic(), Keys, Local

### Community 21 - "WGKey"
Cohesion: 0.09
Nodes (40): NewNet(), NewRoutes(), TestRoutesSysctlAndFirewall(), IsStepError(), New(), meshDNS(), newRig(), TestApplyIsIdempotent() (+32 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.12
Nodes (10): Routes, index(), route(), compare(), luidOf(), samePrefixes(), SortPrefixes(), netlinkTable (+2 more)

### Community 23 - "meshctl/main.go"
Cohesion: 0.18
Nodes (16): agentCommands(), commands(), passwords(), runWith(), TestUserCommands(), healthCommands(), find(), isHelp() (+8 more)

### Community 24 - ".nodeView"
Cohesion: 0.13
Nodes (13): TestJudgeHubs(), Health(), phoneHealth(), ago(), bytesize(), TestHelpers(), dashboardData, networkBlock (+5 more)

### Community 25 - "LocalStatus"
Cohesion: 0.10
Nodes (20): statusDetail, Logs, logs(), laptop(), PrefsRequest, PrefsResponse, Rollback, LocalMesh (+12 more)

### Community 26 - "lab"
Cohesion: 0.13
Nodes (16): host, stream, lab, TestContainerNetworks(), deleteNamespaces(), eventually(), lab, inNS() (+8 more)

### Community 27 - "Handler"
Cohesion: 0.10
Nodes (31): TestStatusAndAgentHealthcheck(), TestLogsCommand(), TestUICommand(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe() (+23 more)

### Community 28 - "Session"
Cohesion: 0.13
Nodes (10): CheckServerDomains(), TestCheckServerDomains(), Session, atoi(), before(), ui, versionNumber(), message() (+2 more)

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (30): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+22 more)

### Community 30 - "time.Time"
Cohesion: 0.07
Nodes (18): pendingApply, safeApply, syncBuffer, Failover, replayCache, failures, hitWindow, Lockout (+10 more)

### Community 31 - "Host"
Cohesion: 0.17
Nodes (3): link, exitHost, Host

### Community 32 - "runner"
Cohesion: 0.15
Nodes (5): Deps, runner, kick(), oneLine(), Run()

### Community 33 - "newRig"
Cohesion: 0.19
Nodes (24): TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled(), TestChangesNeedOriginAndCSRF(), TestContainerModePages(), TestDashboardFragmentAndExport() (+16 more)

### Community 34 - "errorf"
Cohesion: 0.20
Nodes (13): ComposeFile, ComposeOptions, CustomRule, MeshChange, RecordRequest, RuleRequest, ServerFirewall, Actor (+5 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.11
Nodes (57): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+49 more)

### Community 37 - "Test catalogue"
Cohesion: 0.10
Nodes (20): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+12 more)

### Community 38 - "runArgs"
Cohesion: 0.12
Nodes (29): TestAgentRunRuntimeSocket(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentRunNeedsAJoin(), TestExportCommand(), TestHealthcheckServer() (+21 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (22): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+14 more)

### Community 40 - "Tx"
Cohesion: 0.16
Nodes (4): boolInt(), Session, Tx, scanUser()

### Community 41 - "store_test.go"
Cohesion: 0.05
Nodes (68): TestReconcilerVersionsOnlyRealChanges(), FromSnapshot(), Document, snapshot(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes(), TestToSnapshotRefusesHostileDocuments(), TestToSnapshotRoundTrips() (+60 more)

### Community 42 - "Agent protocol"
Cohesion: 0.15
Nodes (13): Agent protocol, Apply (agent), Config and events, Container names, Errors, Join, Local API, Mesh view (+5 more)

### Community 43 - "net/http.ResponseWriter"
Cohesion: 0.26
Nodes (6): decode(), reply(), cli(), api, readBody(), api

### Community 44 - "newRig"
Cohesion: 0.19
Nodes (14): rig, newRig(), TestAccountChangesEndSessions(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures(), TestLoginRefusals(), TestLoginUpgradesAnOldHash() (+6 more)

### Community 45 - "planner_test.go"
Cohesion: 0.18
Nodes (28): deliver(), deliverAddr(), encode(), firstDiff(), hasProblem(), lookup(), mustCompute(), node() (+20 more)

### Community 46 - "meshDNS"
Cohesion: 0.17
Nodes (22): desired(), same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects() (+14 more)

### Community 47 - "firewall_test.go"
Cohesion: 0.22
Nodes (22): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestApplyIsIdempotent() (+14 more)

### Community 49 - "CLAUDE.md"
Cohesion: 0.25
Nodes (7): Determinism and goldens, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates

### Community 50 - "Security"
Cohesion: 0.12
Nodes (16): Agent service (systemd), Changes a node asks for, Container-mode agent, DNS, Input and commands, Interfaces the agent did not create, Keys and tokens, NAS container (+8 more)

### Community 52 - "Service"
Cohesion: 0.10
Nodes (30): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), fullExpiry() (+22 more)

### Community 53 - "Client"
Cohesion: 0.18
Nodes (11): Options, apiError(), clip(), Client, Join(), New(), TestJoinIsUnsigned(), TestUnreachableAndMalformedAnswers() (+3 more)

### Community 54 - "nrpt.go"
Cohesion: 0.19
Nodes (15): nrpt, nrptRule, domainName(), fromRules(), nrptKey(), ownedKey(), parseServers(), rulesFor() (+7 more)

### Community 55 - "Container"
Cohesion: 0.15
Nodes (11): fakeRuntime, drain(), runner, Container, hostAddr(), Of(), TestHostAddr(), TestOf() (+3 more)

### Community 56 - "CookieValue"
Cohesion: 0.21
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 57 - "Node"
Cohesion: 0.25
Nodes (9): Peer, Record, ContainerNetwork(), Node, ruleNote(), sortPeers(), sortPrefixes(), plan (+1 more)

### Community 58 - "context.Context"
Cohesion: 0.09
Nodes (15): fakeService, waitApplied(), resetTOTP(), DNS, dial(), dial(), isPipe(), Client (+7 more)

### Community 59 - "DNS"
Cohesion: 0.18
Nodes (5): nameServer, DNS, Want, DNS, MeshName

### Community 60 - "Service"
Cohesion: 0.18
Nodes (6): Registrar, Service, New(), NewRegistrar(), NewRegistrar(), exchange()

### Community 61 - "Tx"
Cohesion: 0.07
Nodes (32): Change, hubMove, hubState, NetworkInfo, NodeAddRequest, NodeSetRequest, PhoneRequest, PlanError (+24 more)

### Community 62 - "Sync"
Cohesion: 0.17
Nodes (12): network, Sysctls, ensureAddr(), Gateway(), Net, Guard(), host(), parse() (+4 more)

### Community 63 - "service_test.go"
Cohesion: 0.08
Nodes (24): hostService(), TestUnitPosture(), SystemdUnit(), applied, Firewall, policy, Firewall, Runner (+16 more)

### Community 64 - "wgdev"
Cohesion: 0.24
Nodes (3): TestUAPISetGolden(), UAPISet(), wgdev

### Community 65 - "UI"
Cohesion: 0.14
Nodes (11): UI, hostOf(), New(), oldest(), validOrNot(), loadPages(), UIToken, LocalToken() (+3 more)

### Community 66 - "netlinkNet"
Cohesion: 0.19
Nodes (6): netlinkNet, blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule()

### Community 67 - "session"
Cohesion: 0.31
Nodes (6): ago(), bytesize(), UI, prefsFlash(), page, session

### Community 68 - "loadMigrations"
Cohesion: 0.27
Nodes (5): Store, Tx, loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 69 - "NodeConfig"
Cohesion: 0.13
Nodes (7): JoinResponse, NodeConfig, Role, withDomains(), Hub(), MeshNode, Self

### Community 70 - "meshHarness"
Cohesion: 0.43
Nodes (5): harness, meshHarness(), refused(), TestCustomDNSRecords(), TestCustomFirewallRules()

### Community 72 - "fail"
Cohesion: 0.08
Nodes (44): check(), Containers, DNS, Firewall, Interface, NodeConfig, validateMeshNames(), newNode() (+36 more)

### Community 74 - "Writer"
Cohesion: 0.29
Nodes (4): New(), New(), SysctlKeys(), Writer

### Community 75 - "mustAffect"
Cohesion: 0.26
Nodes (4): Tx, mustAffect(), scanNode(), scanner

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.12
Nodes (14): resolver, bindBoth(), bindPair(), serve(), Listen(), Listen(), listenSocket(), Listen() (+6 more)

### Community 77 - "container.go"
Cohesion: 0.15
Nodes (23): composeNetwork, containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers() (+15 more)

### Community 78 - "net/http.HandlerFunc"
Cohesion: 0.19
Nodes (9): dnsState(), newRenderer(), ui, sameOrigin(), golden(), TestEveryPageRenders(), TestPageGoldens(), handler (+1 more)

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "nullInt"
Cohesion: 0.20
Nodes (5): Tx, Tx, nullInt(), nullStr(), AuditEntry

### Community 81 - "NewReader"
Cohesion: 0.18
Nodes (12): TestEventsDeliversConfigEvents(), NewReader(), TestSSEReaderBoundsLineLength(), TestSSEReaderJoinsDataLinesAndDefaultsName(), TestSSEReaderReportsTruncatedEvent(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment() (+4 more)

### Community 82 - ".show"
Cohesion: 0.36
Nodes (3): ui, writeHTML(), page

### Community 83 - "IPNet"
Cohesion: 0.10
Nodes (10): IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), lookup(), userspace, NewUserspace(), Link (+2 more)

### Community 84 - ".renderDNS"
Cohesion: 0.36
Nodes (3): ui, dnsData, recordForm

### Community 85 - "deploy_test.go"
Cohesion: 0.10
Nodes (35): parseContainer(), TestContainerComposePosture(), checkPosture(), keys(), mustRead(), sameSet(), service(), TestDevComposePosture() (+27 more)

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
Cohesion: 0.36
Nodes (6): installAgent(), installAgent(), cleanHostname(), JoinOptions, Service, Join()

### Community 93 - ".names"
Cohesion: 0.17
Nodes (13): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), api, LoadNames(), PublishNames() (+5 more)

### Community 94 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 95 - "Tx"
Cohesion: 0.33
Nodes (3): Tx, PeerStat, StatusUpdate

### Community 97 - "Host"
Cohesion: 0.10
Nodes (13): Deps, Rechecker, rig, Sysctl, Engine, Firewall, Result, Sysctls (+5 more)

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
Cohesion: 0.22
Nodes (11): Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList(), TestListFailures() (+3 more)

### Community 102 - "Allocate"
Cohesion: 0.27
Nodes (11): Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateContainers(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan() (+3 more)

### Community 103 - "PhoneConfig"
Cohesion: 0.31
Nodes (9): NewKey(), Render(), aliceKey(), golden(), TestHub(), TestNewKey(), TestProfiles(), TestRenderRefuses() (+1 more)

### Community 104 - "Exporter"
Cohesion: 0.50
Nodes (3): NextExport(), TestNextExport(), Exporter

### Community 105 - "Tx"
Cohesion: 0.28
Nodes (3): DNSRecord, Tx, FirewallRule

### Community 106 - "Run"
Cohesion: 0.12
Nodes (24): Deps, Route, NewNodes(), AdminRoutes(), describe(), api, New(), NewAdmin() (+16 more)

### Community 107 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 108 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 110 - "net/netip.Addr"
Cohesion: 0.18
Nodes (10): nameTable, zone, TestIsPublic(), isPublic(), PublicIPv4(), JoinRequest, pickEndpoint(), TestPickEndpoint() (+2 more)

### Community 113 - "commands.go"
Cohesion: 0.26
Nodes (9): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), shownDefault(), printExit() (+1 more)

### Community 114 - "Arg"
Cohesion: 0.40
Nodes (4): Arg(), FuzzArg(), FuzzName(), TestArg()

### Community 115 - "Input"
Cohesion: 0.32
Nodes (3): Compute(), Input, newPlan()

### Community 116 - "meshctl/docs_test.go"
Cohesion: 0.43
Nodes (6): cell(), renderCommands(), sentence(), TestCommandsDoc(), TestEveryCommandHasHelp(), TestHelpShowsFlagsAsTheDocsDo()

### Community 118 - "NewHub"
Cohesion: 0.50
Nodes (5): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 120 - "serve"
Cohesion: 0.15
Nodes (15): brokenStream, ErrorBody, serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallScriptFailsWithoutChecksums() (+7 more)

### Community 121 - "log/slog.Logger"
Cohesion: 0.08
Nodes (14): Hub, subscription, TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), ConfigEvent, DB (+6 more)

### Community 122 - "printStatus"
Cohesion: 0.50
Nodes (4): TestBytesize(), dnsLine(), bytesize(), printStatus()

### Community 123 - "changed"
Cohesion: 0.19
Nodes (8): ui, changed(), ui, newQR(), firewallData, phoneData, qrCode, ruleForm

### Community 129 - "service_windows.go"
Cohesion: 0.11
Nodes (10): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows() (+2 more)

### Community 134 - "IdentityKey"
Cohesion: 0.13
Nodes (21): newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestNamesRoundTrip(), TestPrefsRoundTrip(), TestStatusRoundTrip() (+13 more)

## Knowledge Gaps
- **165 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `listed`, `inspected` (+160 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 386 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **31 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NodeConfig` connect `NodeConfig` to `runner`, `Host`, `ui/ui_test.go`, `agent/agent_test.go`, `meshHarness`, `PhoneConfig`, `Key`, `planner_test.go`, `net/netip.Addr`, `NodeChange`, `Input`, `Dir`, `Client`, `WGKey`, `log/slog.Logger`, `Tx`, `time.Time`, `Node`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `Net` connect `Net` to `ui/ui_test.go`, `WGKey`, `net/netip.Prefix`, `time.Time`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **Why does `Name()` connect `fail` to `ui/ui_test.go`, `errorf`, `session`, `env`, `container.go`, `adminFlags`, `Arg`, `deploy_test.go`, `.validate`, `Tx`?**
  _High betweenness centrality (0.011) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _165 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08147982336541788 - nodes in this community are weakly interconnected._
- **Should `newHarness` be split into smaller, more focused modules?**
  _Cohesion score 0.09841269841269841 - nodes in this community are weakly interconnected._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.06358543417366946 - nodes in this community are weakly interconnected._