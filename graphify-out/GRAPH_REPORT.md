# Graph Report - vpn-mesh-utility  (2026-10-21)

## Corpus Check
- 249 files · ~251,079 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3013 nodes · 12973 edges · 135 communities (108 shown, 27 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 744 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f5644a18`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- net/http.Handler
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
- net/http.ResponseWriter
- context.Context
- action
- Dir
- ServerConfig
- WGKey
- net/netip.Prefix
- meshctl/main.go
- .nodeView
- LocalStatus
- lab
- Snapshot
- Session
- Command reference
- time.Time
- Host
- runner
- newRig
- .Firewall
- agent/agent_test.go
- Net
- Test catalogue
- runArgs
- Policy
- Tx
- openTest
- Agent protocol
- netlinkNet
- NodeConfig
- planner_test.go
- meshDNS
- CookieValue
- CLAUDE.md
- Security
- otherNet
- Service
- Client
- nrpt.go
- Container
- net/http.Request
- Deps
- Runtime
- DNS
- log/slog.Logger
- Tx
- Sync
- service_test.go
- wgdev
- UI
- WriteEvent
- session
- loadMigrations
- Hub
- meshHarness
- otherPolicy
- fail
- StepError
- NodeInfo
- mustAffect
- net/netip.AddrPort
- container.go
- Ring
- 7. Agent
- nullInt
- errorf
- .show
- kernel
- changed
- deploy_test.go
- Host
- fetchwintun/main_test.go
- FromSnapshot
- 6. Control plane
- 8. Server
- .Seen
- Join
- PublishNames
- resolved
- Diff
- Tx
- Reconciler
- userspace
- File
- meshctl
- serveRuntime
- meshctl/docs_test.go
- PhoneConfig
- Run
- Tx
- Deps
- Registration
- New
- NewKernel
- net/netip.Addr
- .phoneProfile
- NewNet
- commands.go
- Arg
- .Restore
- printStatus
- Open
- Import
- .validate
- Document
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

## Communities (135 total, 27 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, Exit, Interface, logsData, inspected (+9 more)

### Community 1 - "newRig"
Cohesion: 0.32
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 2 - "net/http.Handler"
Cohesion: 0.14
Nodes (14): DB, dnsState(), newRenderer(), staticHandler(), ui, headers(), New(), sameOrigin() (+6 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (119): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+111 more)

### Community 5 - "env"
Cohesion: 0.20
Nodes (28): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+20 more)

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
Cohesion: 0.14
Nodes (26): peer(), hexKey(), ParseUAPI(), getAnswer(), hexOf(), pfxs(), TestParseUAPI(), TestParseUAPIRejectsMalformedValues() (+18 more)

### Community 14 - "newHarness"
Cohesion: 0.07
Nodes (61): ErrorBody, ErrorMessage(), NewNodes(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose() (+53 more)

### Community 15 - "testing.T"
Cohesion: 0.03
Nodes (94): rig, TestCleanHostname(), TestDefaultSocket(), TestJoinValidatesBeforeCallingTheServer(), TestOneLine(), TestOpenRuntime(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig() (+86 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.17
Nodes (9): decode(), reply(), cli(), api, api, decodeJSON(), readBody(), writeJSON() (+1 more)

### Community 17 - "context.Context"
Cohesion: 0.12
Nodes (14): fakeService, NetworkInfo, NodeAddRequest, NodeSetRequest, dial(), dial(), isPipe(), Source (+6 more)

### Community 18 - "action"
Cohesion: 0.21
Nodes (21): TestPrintEnrollmentInTheLocalZone(), adminFlags(), cleared(), joinNonEmpty(), nodeAction(), nodeName(), nodeState(), printChange() (+13 more)

### Community 19 - "Dir"
Cohesion: 0.11
Nodes (14): TestJoin(), Dir, Enrollment, Open(), open(), TestOpenValidatesThePath(), renameRetrying(), syncDir() (+6 more)

### Community 20 - "ServerConfig"
Cohesion: 0.18
Nodes (29): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+21 more)

### Community 21 - "WGKey"
Cohesion: 0.10
Nodes (36): rig, TestRunNeedsAnEnrollment(), NewHost(), NewRoutes(), TestConfigureHasKernelSemantics(), TestLinksAddressesAndFailures(), TestRoutesSysctlAndFirewall(), IsStepError() (+28 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.10
Nodes (19): Routes, NewTable(), compare(), Table, Sync(), NewTable(), luidOf(), NewTable() (+11 more)

### Community 23 - "meshctl/main.go"
Cohesion: 0.18
Nodes (16): agentCommands(), commands(), passwords(), runWith(), TestUserCommands(), healthCommands(), find(), isHelp() (+8 more)

### Community 24 - ".nodeView"
Cohesion: 0.15
Nodes (11): TestJudgeHubs(), Health(), phoneHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeForm (+3 more)

### Community 25 - "LocalStatus"
Cohesion: 0.09
Nodes (19): Client, Logs, logs(), refusal(), laptop(), PrefsRequest, PrefsResponse, LocalMesh (+11 more)

### Community 26 - "lab"
Cohesion: 0.13
Nodes (16): host, stream, lab, TestContainerNetworks(), deleteNamespaces(), eventually(), lab, inNS() (+8 more)

### Community 27 - "Snapshot"
Cohesion: 0.18
Nodes (15): SnapNode, Snapshot, Compare(), describeDNS(), describeRules(), fields(), fingerprint(), names() (+7 more)

### Community 28 - "Session"
Cohesion: 0.13
Nodes (10): CheckServerDomains(), TestCheckServerDomains(), Session, atoi(), before(), ui, versionNumber(), message() (+2 more)

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (30): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+22 more)

### Community 30 - "time.Time"
Cohesion: 0.06
Nodes (24): pendingApply, safeApply, syncBuffer, Failover, replayCache, failures, hitWindow, Lockout (+16 more)

### Community 31 - "Host"
Cohesion: 0.16
Nodes (3): link, exitHost, Host

### Community 32 - "runner"
Cohesion: 0.17
Nodes (4): runner, kick(), oneLine(), Result

### Community 33 - "newRig"
Cohesion: 0.11
Nodes (29): brokenStream, capped, NewSealer(), TestSealerBindsTheSecretToItsUser(), TestPhonePage(), code(), cookieOf(), newRig() (+21 more)

### Community 34 - ".Firewall"
Cohesion: 0.33
Nodes (5): CustomRule, ServerFirewall, FirewallView, firewallData, ruleForm

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (56): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+48 more)

### Community 36 - "Net"
Cohesion: 0.15
Nodes (5): Net, ensureAddr(), Net, Rule, syncRules()

### Community 37 - "Test catalogue"
Cohesion: 0.10
Nodes (20): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+12 more)

### Community 38 - "runArgs"
Cohesion: 0.06
Nodes (62): TestAgentRunRuntimeSocket(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin(), TestExportCommand() (+54 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (22): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+14 more)

### Community 40 - "Tx"
Cohesion: 0.21
Nodes (3): boolInt(), Session, Tx

### Community 41 - "openTest"
Cohesion: 0.11
Nodes (32): countingDB, TestSafeApplyStateMachine(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), Store, Open(), insertTopology(), node(), openTest() (+24 more)

### Community 42 - "Agent protocol"
Cohesion: 0.15
Nodes (13): Agent protocol, Apply (agent), Config and events, Container names, Errors, Join, Local API, Mesh view (+5 more)

### Community 43 - "netlinkNet"
Cohesion: 0.10
Nodes (15): netlinkNet, Owned(), TestGatewayRules(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule() (+7 more)

### Community 44 - "NodeConfig"
Cohesion: 0.11
Nodes (7): NodeDetail, Rechecker, Engine, Desired(), PeerStatus, NodeConfig, withDomains()

### Community 45 - "planner_test.go"
Cohesion: 0.06
Nodes (51): PlanError, Sysctl, New(), New(), Peer, Record, Sysctl, SysctlKeys() (+43 more)

### Community 46 - "meshDNS"
Cohesion: 0.17
Nodes (21): desired(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters() (+13 more)

### Community 47 - "CookieValue"
Cohesion: 0.23
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 49 - "CLAUDE.md"
Cohesion: 0.25
Nodes (7): Determinism and goldens, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates

### Community 50 - "Security"
Cohesion: 0.12
Nodes (16): Agent service (systemd), Changes a node asks for, Container-mode agent, DNS, Input and commands, Interfaces the agent did not create, Keys and tokens, NAS container (+8 more)

### Community 52 - "Service"
Cohesion: 0.08
Nodes (31): Config, DB, Params, Pending, Sealer, addUser(), newPassword(), passwdUser() (+23 more)

### Community 53 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 54 - "nrpt.go"
Cohesion: 0.18
Nodes (16): nrpt, nrptRule, domainName(), same(), TestSameRegistration(), fromRules(), nrptKey(), ownedKey() (+8 more)

### Community 55 - "Container"
Cohesion: 0.15
Nodes (11): fakeRuntime, drain(), runner, Container, hostAddr(), Of(), TestHostAddr(), TestOf() (+3 more)

### Community 56 - "net/http.Request"
Cohesion: 0.16
Nodes (6): ETag(), TestETagQuotesTheHash(), api, readSums(), ClientIP(), TestClientIP()

### Community 57 - "Deps"
Cohesion: 0.16
Nodes (11): Deps, Deps, Options, Run(), DNS, Firewall, Sysctls, Join() (+3 more)

### Community 58 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 59 - "DNS"
Cohesion: 0.22
Nodes (4): nameServer, DNS, Want, MeshName

### Community 60 - "log/slog.Logger"
Cohesion: 0.14
Nodes (9): Registrar, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange(), NewUserspace() (+1 more)

### Community 61 - "Tx"
Cohesion: 0.08
Nodes (32): hubMove, hubState, PhoneRequest, JoinResponse, Role, checkName(), containerSubnet(), defaultMTU() (+24 more)

### Community 62 - "Sync"
Cohesion: 0.19
Nodes (17): Sysctls, NewNet(), pfx(), TestNetHasKernelSemantics(), Gateway(), Guard(), parse(), Sync() (+9 more)

### Community 63 - "service_test.go"
Cohesion: 0.08
Nodes (23): hostService(), TestUnitPosture(), SystemdUnit(), applied, Firewall, policy, Firewall, Runner (+15 more)

### Community 64 - "wgdev"
Cohesion: 0.11
Nodes (8): lookup(), TestUAPISetGolden(), UAPISet(), userspace, NewUserspace(), Link, links, wgdev

### Community 65 - "UI"
Cohesion: 0.21
Nodes (7): UI, hostOf(), oldest(), validOrNot(), LocalToken(), TestLocalTokenMasksTheValue(), handler

### Community 66 - "WriteEvent"
Cohesion: 0.26
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 67 - "session"
Cohesion: 0.34
Nodes (5): bytesize(), UI, prefsFlash(), page, session

### Community 68 - "loadMigrations"
Cohesion: 0.27
Nodes (5): Store, Tx, loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 69 - "Hub"
Cohesion: 0.24
Nodes (3): Hub, subscription, ConfigEvent

### Community 70 - "meshHarness"
Cohesion: 0.43
Nodes (5): harness, meshHarness(), refused(), TestCustomDNSRecords(), TestCustomFirewallRules()

### Community 72 - "fail"
Cohesion: 0.09
Nodes (39): check(), Containers, DNS, Firewall, Interface, NodeConfig, validateMeshNames(), newNode() (+31 more)

### Community 74 - "NodeInfo"
Cohesion: 0.18
Nodes (10): NodeList, PhoneProfile, statusDetail, heard(), Rollback, NodeInfo, sameRollback(), phoneState() (+2 more)

### Community 75 - "mustAffect"
Cohesion: 0.23
Nodes (4): Tx, mustAffect(), classify(), nullStr()

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.16
Nodes (11): resolver, bindBoth(), bindPair(), serve(), Listen(), Listen(), listenSocket(), ParseAddr() (+3 more)

### Community 77 - "container.go"
Cohesion: 0.15
Nodes (25): composeNetwork, containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers() (+17 more)

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "nullInt"
Cohesion: 0.16
Nodes (6): Tx, DNSRecord, Tx, nullInt(), AuditEntry, FirewallRule

### Community 81 - "errorf"
Cohesion: 0.19
Nodes (13): apiError, ComposeFile, ComposeOptions, MeshChange, RecordRequest, RuleRequest, errorf(), invalid() (+5 more)

### Community 82 - ".show"
Cohesion: 0.36
Nodes (3): ui, writeHTML(), page

### Community 84 - "changed"
Cohesion: 0.25
Nodes (3): ui, ui, changed()

### Community 85 - "deploy_test.go"
Cohesion: 0.11
Nodes (33): checkPosture(), keys(), mustRead(), sameSet(), service(), TestDevComposePosture(), TestInstallChangesOnlyTheMarkers(), TestInstallGolden() (+25 more)

### Community 86 - "Host"
Cohesion: 0.36
Nodes (3): Host, SyncAddr(), SyncLink()

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.13
Nodes (19): NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall(), download() (+11 more)

### Community 88 - "FromSnapshot"
Cohesion: 0.36
Nodes (7): FromSnapshot(), snapshot(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes(), TestToSnapshotRefusesHostileDocuments(), TestToSnapshotRoundTrips(), counter

### Community 89 - "6. Control plane"
Cohesion: 0.40
Nodes (5): 6.1 Enrollment, 6.2 Request authentication, 6.3 Config sync, 6.4 Safe apply, 6. Control plane

### Community 90 - "8. Server"
Cohesion: 0.40
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 92 - "Join"
Cohesion: 0.36
Nodes (6): installAgent(), installAgent(), cleanHostname(), JoinOptions, Service, Join()

### Community 93 - "PublishNames"
Cohesion: 0.12
Nodes (17): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), api, LoadNames(), PublishNames() (+9 more)

### Community 94 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 95 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 97 - "Reconciler"
Cohesion: 0.48
Nodes (4): Change, Reconciler, Result, NewReconciler()

### Community 98 - "userspace"
Cohesion: 0.19
Nodes (4): defaultInterface(), exitRoutes(), userspace, unspecified()

### Community 99 - "File"
Cohesion: 0.27
Nodes (5): agentLog(), ttyPassword(), OpenFile(), TestFileRotates(), File

### Community 100 - "meshctl"
Cohesion: 0.40
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 101 - "serveRuntime"
Cohesion: 0.22
Nodes (11): Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList(), TestListFailures() (+3 more)

### Community 102 - "meshctl/docs_test.go"
Cohesion: 0.43
Nodes (6): cell(), renderCommands(), sentence(), TestCommandsDoc(), TestEveryCommandHasHelp(), TestHelpShowsFlagsAsTheDocsDo()

### Community 103 - "PhoneConfig"
Cohesion: 0.24
Nodes (11): Hub(), NewKey(), Render(), aliceKey(), golden(), TestHub(), TestNewKey(), TestProfiles() (+3 more)

### Community 104 - "Run"
Cohesion: 0.29
Nodes (6): LoadOrCreateKey(), TestLoadOrCreateKey(), hostOf(), Routes(), Run(), serve()

### Community 105 - "Tx"
Cohesion: 0.33
Nodes (3): Tx, PeerStat, StatusUpdate

### Community 106 - "Deps"
Cohesion: 0.22
Nodes (13): Deps, Route, AdminRoutes(), describe(), api, New(), NewAdmin(), newAPI() (+5 more)

### Community 107 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 108 - "New"
Cohesion: 0.33
Nodes (5): Listen(), New(), TestServeAndListen(), loadPages(), pages

### Community 110 - "net/netip.Addr"
Cohesion: 0.15
Nodes (12): network, nameTable, zone, TestIsPublic(), host(), isPublic(), PublicIPv4(), JoinRequest (+4 more)

### Community 111 - ".phoneProfile"
Cohesion: 0.33
Nodes (5): ui, newQR(), TestQRCode(), phoneData, qrCode

### Community 113 - "commands.go"
Cohesion: 0.26
Nodes (9): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), setupVersion(), shownDefault() (+1 more)

### Community 114 - "Arg"
Cohesion: 0.40
Nodes (4): Arg(), FuzzArg(), FuzzName(), TestArg()

### Community 116 - "printStatus"
Cohesion: 0.50
Nodes (4): TestBytesize(), dnsLine(), bytesize(), printStatus()

### Community 117 - "Open"
Cohesion: 0.50
Nodes (3): TestOpen(), Open(), TestOpenRefusesUnknownBackends()

### Community 118 - "Import"
Cohesion: 0.24
Nodes (12): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses(), Import(), document(), linuxOnly() (+4 more)

### Community 121 - "Document"
Cohesion: 0.17
Nodes (6): watchdog, Document, Read(), Write(), Config, ImportConfig

### Community 129 - "service_windows.go"
Cohesion: 0.11
Nodes (10): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows() (+2 more)

### Community 134 - "IdentityKey"
Cohesion: 0.12
Nodes (22): newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestNamesRoundTrip(), TestPrefsRoundTrip(), TestStatusRoundTrip() (+14 more)

## Knowledge Gaps
- **165 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `listed`, `inspected` (+160 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 388 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **27 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Service` connect `log/slog.Logger` to `ui/ui_test.go`, `net/netip.AddrPort`, `meshDNS`, `net/netip.Addr`, `time.Time`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `Dir` connect `Dir` to `runner`, `ui/ui_test.go`, `agent/agent_test.go`, `NodeConfig`, `testing.T`, `LocalStatus`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `fakeServer` connect `agent/agent_test.go` to `WriteEvent`, `Hub`, `NodeConfig`, `newHarness`, `testing.T`, `net/http.Request`, `LocalStatus`, `time.Time`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _165 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07932234678768532 - nodes in this community are weakly interconnected._
- **Should `net/http.Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.14285714285714285 - nodes in this community are weakly interconnected._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.06358543417366946 - nodes in this community are weakly interconnected._