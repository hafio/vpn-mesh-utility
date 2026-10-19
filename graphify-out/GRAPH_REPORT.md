# Graph Report - vpn-mesh-utility  (2026-10-19)

## Corpus Check
- 204 files · ~195,593 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .css 2, .tmpl 1)

## Summary
- 2575 nodes · 10759 edges · 105 communities (83 shown, 22 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 595 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `c5325888`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- Open
- planner_test.go
- htmx.min.js
- env
- meshctl — Design Spec
- Extending the code
- dev.sh
- dev.ps1
- archtest.go
- RenderNASCompose
- Tx
- github.com/hafio/vpn-mesh-utility
- newHarness
- .Apply
- Run
- TestNASComposePosture
- nullInt
- Dir
- userspace
- WGKey
- net/netip.Prefix
- Registration
- .nodeView
- LocalStatus
- lab
- service_windows.go
- Session
- Command reference
- time.Time
- Host
- context.Context
- newRig
- Tx
- agent_test.go
- testing.T
- Test catalogue
- IdentityKey
- Policy
- Tx
- dns_test.go
- Agent protocol
- net/http.Request
- newRig
- Snapshot
- runArgs
- newFirewall
- CLAUDE.md
- Security
- Host
- Service
- Client
- Node
- .renderFirewall
- net/http.ResponseWriter
- Sealer
- NodeChange
- newQR
- Service
- log/slog.Logger
- Ring
- RenderInstall
- resolved
- .exec
- session
- validate_test.go
- Tx
- otherPolicy
- Name
- StepError
- Run
- net/netip.AddrPort
- 7. Agent
- FromSnapshot
- Diff
- page
- .validate
- validate.go
- Refusal
- .renderDNS
- 6. Control plane
- 8. Server
- .Seen
- .AddRule
- errorf
- toAPIError
- meshctl
- AdminError
- Node
- Tx
- Tx
- Join
- safeRig
- NodeConfig
- net/netip.Addr
- fakeService
- TestMain
- entry
- Error
- NewPolicy
- NewPolicy
- NewKernel
- NewKernel

## God Nodes (most connected - your core abstractions)
1. `ServerConfig()` - 57 edges
2. `Session` - 55 edges
3. `NodeConfig` - 49 edges
4. `newHarness()` - 45 edges
5. `env` - 43 edges
6. `runner` - 39 edges
7. `ui` - 39 edges
8. `Dir` - 31 edges
9. `lab` - 31 edges
10. `runArgs()` - 30 edges

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
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`

## Communities (105 total, 22 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (18): nasFile, nasHealth, nasService, linkDNS, DNSRecord, FirewallRule, Node, Exit (+10 more)

### Community 1 - "newRig"
Cohesion: 0.16
Nodes (12): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+4 more)

### Community 2 - "Open"
Cohesion: 0.10
Nodes (36): TestSafeApplyStateMachine(), Document, TestExportOnceSkipsUnchangedAndKeepsFourteen(), Import(), document(), linuxOnly(), TestImportRefusesInvalidDocumentsUntouched(), TestImportRefusesWhileTheServerRuns() (+28 more)

### Community 3 - "planner_test.go"
Cohesion: 0.12
Nodes (30): PlanError, Check(), checkDNS(), Problem, sortedByID(), Compute(), Input, newPlan() (+22 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (118): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+110 more)

### Community 5 - "env"
Cohesion: 0.06
Nodes (81): watchdog, agentCommands(), agentLog(), joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall() (+73 more)

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
Cohesion: 0.16
Nodes (33): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+25 more)

### Community 10 - "archtest.go"
Cohesion: 0.27
Nodes (9): Rule, Violation, Check(), importsOf(), ModuleRoot(), TestCheckerCatchesViolations(), TestModuleBoundaries(), under() (+1 more)

### Community 11 - "RenderNASCompose"
Cohesion: 0.60
Nodes (5): comment(), nasHeader(), renderNAS(), RenderNASCompose(), NASParams

### Community 14 - "newHarness"
Cohesion: 0.10
Nodes (47): adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing(), TestAdminSetClearsReferences() (+39 more)

### Community 15 - ".Apply"
Cohesion: 0.10
Nodes (18): CustomRule, ServerFirewall, applied, Firewall, policy, Firewall, atLeast(), normalize() (+10 more)

### Community 16 - "Run"
Cohesion: 0.07
Nodes (30): Deps, Route, UI, hostOf(), oldest(), validOrNot(), loadPages(), NewNodes() (+22 more)

### Community 17 - "TestNASComposePosture"
Cohesion: 0.33
Nodes (10): checkPosture(), keys(), sameSet(), service(), TestDevComposePosture(), TestServerComposePosture(), ComposeServer(), parseNAS() (+2 more)

### Community 18 - "nullInt"
Cohesion: 0.18
Nodes (5): Tx, Tx, nullInt(), nullStr(), AuditEntry

### Community 19 - "Dir"
Cohesion: 0.12
Nodes (10): Dir, Enrollment, Open(), open(), TestOpenValidatesThePath(), syncDir(), TestWriteAtomicReplacesAndLeavesNoTempFiles(), WriteAtomic() (+2 more)

### Community 20 - "userspace"
Cohesion: 0.08
Nodes (12): IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), lookup(), TestUAPISetGolden(), UAPISet(), NewUserspace() (+4 more)

### Community 21 - "WGKey"
Cohesion: 0.10
Nodes (33): rig, TestRunNeedsAnEnrollment(), NewHost(), NewRoutes(), pfx(), TestConfigureHasKernelSemantics(), TestLinksAddressesAndFailures(), IsStepError() (+25 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (21): Routes, index(), NewTable(), route(), compare(), Desired(), Table, Sync() (+13 more)

### Community 23 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 24 - ".nodeView"
Cohesion: 0.13
Nodes (12): Health(), phoneHealth(), TestDashboardHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeData (+4 more)

### Community 25 - "LocalStatus"
Cohesion: 0.10
Nodes (20): Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest, PrefsResponse (+12 more)

### Community 26 - "lab"
Cohesion: 0.16
Nodes (12): host, lab, stream, deleteNamespaces(), eventually(), inNS(), newLab(), ns() (+4 more)

### Community 27 - "service_windows.go"
Cohesion: 0.08
Nodes (15): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows() (+7 more)

### Community 28 - "Session"
Cohesion: 0.13
Nodes (11): Session, atoi(), before(), changed(), ui, stamp(), versionNumber(), ui (+3 more)

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (29): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+21 more)

### Community 30 - "time.Time"
Cohesion: 0.08
Nodes (19): pendingApply, safeApply, syncBuffer, replayCache, failures, hitWindow, Lockout, LockoutError (+11 more)

### Community 31 - "Host"
Cohesion: 0.08
Nodes (29): link, Host, peer(), hexKey(), ParseUAPI(), getAnswer(), hexOf(), pfxs() (+21 more)

### Community 32 - "context.Context"
Cohesion: 0.11
Nodes (9): runner, Rechecker, kick(), oneLine(), TestOneLine(), Engine, Result, RunWindowsService() (+1 more)

### Community 33 - "newRig"
Cohesion: 0.13
Nodes (26): brokenStream, countingDB, Store, TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled() (+18 more)

### Community 34 - "Tx"
Cohesion: 0.11
Nodes (17): DNSEntry, PhoneRequest, checkName(), defaultMTU(), Actor, Nodes, nodeByName(), nodeInfo() (+9 more)

### Community 35 - "agent_test.go"
Cohesion: 0.12
Nodes (48): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+40 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (96): TestStatusAndAgentHealthcheck(), TestLogsCommand(), TestUICommand(), mustRead(), nasParams(), TestNASComposeGolden(), TestNASComposeRefusesHostileValues(), TestReportKeepsToTheServersLimits() (+88 more)

### Community 37 - "Test catalogue"
Cohesion: 0.11
Nodes (19): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+11 more)

### Community 38 - "IdentityKey"
Cohesion: 0.07
Nodes (36): New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestPrefsRoundTrip() (+28 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (20): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+12 more)

### Community 40 - "Tx"
Cohesion: 0.17
Nodes (4): mustAffect(), boolInt(), Tx, scanUser()

### Community 41 - "dns_test.go"
Cohesion: 0.20
Nodes (18): domainName(), same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects() (+10 more)

### Community 42 - "Agent protocol"
Cohesion: 0.17
Nodes (12): Agent protocol, Apply (agent), Config and events, Errors, Join, Local API, Mesh view, Prefs (+4 more)

### Community 43 - "net/http.Request"
Cohesion: 0.32
Nodes (3): cli(), api, api

### Community 44 - "newRig"
Cohesion: 0.16
Nodes (17): rig, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures(), TestLoginRefusals() (+9 more)

### Community 45 - "Snapshot"
Cohesion: 0.13
Nodes (19): TestReconcilerVersionsOnlyRealChanges(), DecodeSnapshot(), SnapNode, Snapshot, snapNode(), TestDecodeSnapshotRefusesUnknownShapes(), Compare(), describeDNS() (+11 more)

### Community 46 - "runArgs"
Cohesion: 0.09
Nodes (37): passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+29 more)

### Community 47 - "newFirewall"
Cohesion: 0.08
Nodes (37): hostService(), TestUnitPosture(), SystemdUnit(), Exec, fakeRunner, allowed(), Check(), Runner (+29 more)

### Community 49 - "CLAUDE.md"
Cohesion: 0.25
Nodes (7): Determinism and goldens, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates

### Community 50 - "Security"
Cohesion: 0.12
Nodes (15): Agent service (systemd), Changes a node asks for, DNS, Input and commands, Interfaces the agent did not create, Keys and tokens, NAS container, Security (+7 more)

### Community 51 - "Host"
Cohesion: 0.31
Nodes (4): Host, SyncAddr(), SyncLink(), TestSyncLinkRefusesAForeignInterface()

### Community 52 - "Service"
Cohesion: 0.10
Nodes (29): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), resetTOTP() (+21 more)

### Community 53 - "Client"
Cohesion: 0.32
Nodes (4): apiError(), clip(), Client, StatusResponse

### Community 54 - "Node"
Cohesion: 0.29
Nodes (5): reachable(), Node, Tx, scanNode(), scanner

### Community 56 - "net/http.ResponseWriter"
Cohesion: 0.17
Nodes (8): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 58 - "NodeChange"
Cohesion: 0.14
Nodes (16): ComposeFile, NodeAddRequest, NodeList, NodeSetRequest, PhoneProfile, statusDetail, heard(), Rollback (+8 more)

### Community 59 - "newQR"
Cohesion: 0.50
Nodes (4): newQR(), TestQRCode(), phoneData, qrCode

### Community 60 - "Service"
Cohesion: 0.13
Nodes (9): Registrar, Service, DNS, desired(), Want, New(), NewRegistrar(), NewRegistrar() (+1 more)

### Community 61 - "log/slog.Logger"
Cohesion: 0.09
Nodes (15): Change, Hub, subscription, NewUserspace(), ConfigEvent, DB, Reconciler, Result (+7 more)

### Community 63 - "RenderInstall"
Cohesion: 0.48
Nodes (7): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams

### Community 64 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 67 - "session"
Cohesion: 0.34
Nodes (5): bytesize(), UI, prefsFlash(), page, session

### Community 69 - "validate_test.go"
Cohesion: 0.12
Nodes (17): Arg(), FuzzArg(), FuzzName(), TestArg(), TestEndpoint(), TestEnum(), TestFilePath(), TestHostnameFoldsCase() (+9 more)

### Community 70 - "Tx"
Cohesion: 0.28
Nodes (3): Tx, DNSRecord, FirewallRule

### Community 72 - "Name"
Cohesion: 0.22
Nodes (15): check(), Firewall, Interface, NodeConfig, newNode(), validateSet(), ToSnapshot(), Endpoint() (+7 more)

### Community 74 - "Run"
Cohesion: 0.11
Nodes (13): Deps, Deps, Run(), Sysctl, DNS, Firewall, Sysctls, New() (+5 more)

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.09
Nodes (17): resolver, bindPair(), serve(), dial(), Listen(), dial(), isPipe(), Listen() (+9 more)

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "FromSnapshot"
Cohesion: 0.36
Nodes (7): FromSnapshot(), snapshot(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes(), TestToSnapshotRefusesHostileDocuments(), TestToSnapshotRoundTrips(), counter

### Community 81 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 82 - "page"
Cohesion: 0.36
Nodes (3): ui, writeHTML(), page

### Community 85 - "validate.go"
Cohesion: 0.16
Nodes (19): TestCleanHostname(), cleanHostname(), fail(), FilePath(), Hostname(), isLoopbackHost(), Key32(), LocalToken() (+11 more)

### Community 88 - ".renderDNS"
Cohesion: 0.36
Nodes (3): ui, dnsData, recordForm

### Community 89 - "6. Control plane"
Cohesion: 0.40
Nodes (5): 6.1 Enrollment, 6.2 Request authentication, 6.3 Config sync, 6.4 Safe apply, 6. Control plane

### Community 90 - "8. Server"
Cohesion: 0.40
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 94 - ".AddRule"
Cohesion: 0.20
Nodes (10): MeshChange, RecordRequest, RuleRequest, DNS, Nodes, meshChange(), DNSName(), IPv4() (+2 more)

### Community 95 - "errorf"
Cohesion: 0.20
Nodes (8): api, decodeJSON(), errorf(), readBody(), writeJSON(), api, readSums(), ClientIP()

### Community 96 - "toAPIError"
Cohesion: 0.20
Nodes (10): apiError, ErrorMessage(), toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion(), harness, meshHarness(), refused() (+2 more)

### Community 100 - "meshctl"
Cohesion: 0.40
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 102 - "Node"
Cohesion: 0.18
Nodes (14): Peer, Record, Role, plannerNode(), RangeFor(), TestRangeFor(), Node, ruleNote() (+6 more)

### Community 103 - "Tx"
Cohesion: 0.33
Nodes (3): Tx, PeerStat, StatusUpdate

### Community 108 - "Join"
Cohesion: 0.24
Nodes (9): Options, installAgent(), TestJoinValidatesBeforeCallingTheServer(), Join(), TestJoinIsUnsigned(), JoinOptions, Service, Join() (+1 more)

### Community 110 - "safeRig"
Cohesion: 0.38
Nodes (9): pendingExists(), safeRig(), TestSafeApplyConfirmsWhenTheServerAnswers(), TestSafeApplyDisarmsWhenTheServerGoesBack(), TestSafeApplyRecoversAPendingConfigAtBoot(), TestSafeApplyReportsAFailedRollback(), TestSafeApplyRollsBackWithoutAnAnswer(), TestSpokesDoNotArmSafeApply() (+1 more)

### Community 111 - "NodeConfig"
Cohesion: 0.14
Nodes (7): NodeDetail, PeerStatus, NodeConfig, Hub(), Render(), TestHub(), Profile

### Community 113 - "net/netip.Addr"
Cohesion: 0.17
Nodes (12): zone, TestIsPublic(), soa(), isPublic(), PublicIPv4(), JoinRequest, pickEndpoint(), TestPickEndpoint() (+4 more)

## Knowledge Gaps
- **157 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `Interface`, `Exit` (+152 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 358 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **22 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Dir` connect `Dir` to `context.Context`, `ui/ui_test.go`, `agent_test.go`, `testing.T`, `safeRig`, `LocalStatus`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Why does `NodeConfig` connect `NodeConfig` to `context.Context`, `toAPIError`, `ui/ui_test.go`, `agent_test.go`, `Tx`, `planner_test.go`, `Node`, `Run`, `safeRig`, `Dir`, `Client`, `net/netip.Prefix`, `WGKey`, `log/slog.Logger`, `time.Time`, `Host`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `runner` connect `context.Context` to `ui/ui_test.go`, `newRig`, `Run`, `net/netip.AddrPort`, `NodeConfig`, `Dir`, `Client`, `LocalStatus`, `NodeChange`, `log/slog.Logger`, `time.Time`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _157 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07889671304305451 - nodes in this community are weakly interconnected._
- **Should `Open` be split into smaller, more focused modules?**
  _Cohesion score 0.09872241579558652 - nodes in this community are weakly interconnected._
- **Should `planner_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11740890688259109 - nodes in this community are weakly interconnected._