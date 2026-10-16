# Graph Report - vpn-mesh-utility  (2026-10-17)

## Corpus Check
- 204 files · ~194,071 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 8 file(s) not represented in the graph (top: (none) 4, .css 2, .tmpl 1)

## Summary
- 2569 nodes · 10729 edges · 125 communities (100 shown, 25 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 592 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7f29e6cc`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- store_test.go
- NodeConfig
- htmx.min.js
- io.Writer
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
- UI
- TestNASComposePosture
- Engine
- Dir
- IPNet
- WGKey
- net/netip.Prefix
- WriteEvent
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
- TestServiceAnswersAndRegisters
- Agent protocol
- net/http.ResponseWriter
- newRig
- Snapshot
- Import
- ServerConfig
- CLAUDE.md
- Security
- Host
- Service
- Client
- Node
- env
- CookieValue
- Key
- NodeChange
- net/http.Request
- log/slog.Logger
- Hub
- Ring
- ServerURL
- Registration
- Run
- Deps
- session
- execrun_test.go
- Arg
- nullInt
- otherPolicy
- fail
- StepError
- sync.Mutex
- userspace
- net.Listener
- loadMigrations
- Change
- 7. Agent
- FromSnapshot
- Diff
- page
- main.go
- .validate
- Error
- Refusal
- fakeServer
- changed
- 6. Control plane
- 8. Server
- .Seen
- net/http.Handler
- time.Duration
- errorf
- .install
- NewNodes
- route
- .DNS
- commands.go
- meshctl
- AdminError
- Allocate
- mustAffect
- Document
- Tx
- newDir
- File
- Join
- renderer
- safeRig
- PhoneConfig
- Exporter
- validJoin
- Signed
- fakeService
- TestMain
- droppedLogs
- entry
- Error
- DefaultSocket
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

## Communities (125 total, 25 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (18): nasFile, nasHealth, nasService, linkDNS, DNSRecord, FirewallRule, Node, Exit (+10 more)

### Community 1 - "newRig"
Cohesion: 0.16
Nodes (12): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+4 more)

### Community 2 - "store_test.go"
Cohesion: 0.16
Nodes (30): TestSafeApplyStateMachine(), TestReconcilerVersionsOnlyRealChanges(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), DecodeSnapshot(), Open(), insertTopology(), node(), openTest() (+22 more)

### Community 3 - "NodeConfig"
Cohesion: 0.06
Nodes (48): Change, PlanError, zone, TestIsPublic(), soa(), isPublic(), PublicIPv4(), NodeConfig (+40 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (118): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+110 more)

### Community 5 - "io.Writer"
Cohesion: 0.18
Nodes (22): TestBytesize(), dnsLine(), printExit(), adminFlags(), cleared(), joinNonEmpty(), nodeAction(), nodeName() (+14 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.11
Nodes (18): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments (+10 more)

### Community 7 - "Extending the code"
Cohesion: 0.15
Nodes (13): Add a command, Add a database migration, Add a topology rule, Add a validator, Add a web UI page, Add an agent API route, Add an apply step, Dev script tasks (+5 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (29): finish(), has_linux_docker(), log_init(), main(), NO_COLOR, dev.sh script, build_image(), build_one() (+21 more)

### Community 9 - "dev.ps1"
Cohesion: 0.17
Nodes (30): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Invoke-TrivyImage(), Build-Image(), Build-One(), Cap() (+22 more)

### Community 10 - "archtest.go"
Cohesion: 0.27
Nodes (9): Rule, Violation, Check(), importsOf(), ModuleRoot(), TestCheckerCatchesViolations(), TestModuleBoundaries(), under() (+1 more)

### Community 11 - "RenderNASCompose"
Cohesion: 0.28
Nodes (9): comment(), nasHeader(), renderNAS(), RenderNASCompose(), NASParams, POSIXPath(), SemVer(), TestPOSIXPath() (+1 more)

### Community 14 - "newHarness"
Cohesion: 0.10
Nodes (46): countingDB, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing(), TestAdminSetClearsReferences() (+38 more)

### Community 15 - ".Apply"
Cohesion: 0.08
Nodes (22): CustomRule, ServerFirewall, hostService(), applied, Firewall, policy, Firewall, Runner (+14 more)

### Community 16 - "UI"
Cohesion: 0.21
Nodes (7): UI, hostOf(), oldest(), validOrNot(), LocalToken(), TestLocalTokenMasksTheValue(), handler

### Community 17 - "TestNASComposePosture"
Cohesion: 0.22
Nodes (14): checkPosture(), keys(), mustRead(), sameSet(), service(), TestDevComposePosture(), TestServerComposePosture(), ComposeServer() (+6 more)

### Community 18 - "Engine"
Cohesion: 0.21
Nodes (6): Deps, Rechecker, Engine, Result, Sysctls, New()

### Community 19 - "Dir"
Cohesion: 0.13
Nodes (7): Dir, Enrollment, syncDir(), TestWriteAtomicReplacesAndLeavesNoTempFiles(), WriteAtomic(), Keys, Local

### Community 20 - "IPNet"
Cohesion: 0.15
Nodes (7): IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), lookup(), kernel, links

### Community 21 - "WGKey"
Cohesion: 0.09
Nodes (39): TestImportCommand(), IsStepError(), meshDNS(), newRig(), TestApplyIsIdempotent(), TestApplyStopsAtTheFailingStep(), TestApplyValidatesBeforeTouchingTheHost(), TestDNSFailureOnlyDegrades() (+31 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.12
Nodes (15): Routes, NewTable(), compare(), Desired(), Table, Sync(), prefixes(), TestDesiredSkipsTheDefaultRouteAndDuplicates() (+7 more)

### Community 23 - "WriteEvent"
Cohesion: 0.26
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 24 - ".nodeView"
Cohesion: 0.15
Nodes (11): Health(), phoneHealth(), TestDashboardHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeForm (+3 more)

### Community 25 - "LocalStatus"
Cohesion: 0.08
Nodes (21): Client, Logs, logs(), refusal(), Listen(), New(), laptop(), TestServeAndListen() (+13 more)

### Community 26 - "lab"
Cohesion: 0.16
Nodes (13): host, lab, stream, deleteNamespaces(), eventually(), inNS(), meshNames(), newLab() (+5 more)

### Community 27 - "service_windows.go"
Cohesion: 0.08
Nodes (15): hostService(), uninstallAgent(), RunWindowsService(), TestNoWindowsServiceHere(), commandLine(), RunWindowsService(), stop(), executes() (+7 more)

### Community 28 - "Session"
Cohesion: 0.18
Nodes (6): Session, atoi(), ui, message(), actor(), settingsData

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (29): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+21 more)

### Community 30 - "time.Time"
Cohesion: 0.13
Nodes (8): pendingApply, safeApply, replayCache, failures, Lockout, LockoutError, record(), Clock

### Community 31 - "Host"
Cohesion: 0.14
Nodes (5): link, rig, Host, TestLinksAddressesAndFailures(), Link

### Community 32 - "context.Context"
Cohesion: 0.12
Nodes (11): Deps, runner, kick(), oneLine(), Run(), TestOneLine(), DNS, Firewall (+3 more)

### Community 33 - "newRig"
Cohesion: 0.12
Nodes (27): brokenStream, Reconciler, NewReconciler(), Store, TestPhonePage(), code(), cookieOf(), newRig() (+19 more)

### Community 34 - "Tx"
Cohesion: 0.10
Nodes (18): NodeAddRequest, NodeSetRequest, PhoneRequest, checkName(), defaultMTU(), Nodes, nodeByName(), nodeInfo() (+10 more)

### Community 35 - "agent_test.go"
Cohesion: 0.18
Nodes (34): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), start(), startLogged() (+26 more)

### Community 36 - "testing.T"
Cohesion: 0.05
Nodes (85): passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentRunNeedsAJoin(), TestExportCommand(), TestHealthcheckServer() (+77 more)

### Community 37 - "Test catalogue"
Cohesion: 0.11
Nodes (19): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+11 more)

### Community 38 - "IdentityKey"
Cohesion: 0.11
Nodes (26): Options, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream(), TestJoinIsUnsigned() (+18 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (20): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+12 more)

### Community 40 - "Tx"
Cohesion: 0.18
Nodes (4): boolInt(), Session, Tx, scanUser()

### Community 41 - "TestServiceAnswersAndRegisters"
Cohesion: 0.15
Nodes (18): domainName(), same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects() (+10 more)

### Community 42 - "Agent protocol"
Cohesion: 0.17
Nodes (12): Agent protocol, Apply (agent), Config and events, Errors, Join, Local API, Mesh view, Prefs (+4 more)

### Community 43 - "net/http.ResponseWriter"
Cohesion: 0.18
Nodes (9): decode(), reply(), cli(), api, api, decodeJSON(), readBody(), writeJSON() (+1 more)

### Community 44 - "newRig"
Cohesion: 0.18
Nodes (15): rig, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures() (+7 more)

### Community 45 - "Snapshot"
Cohesion: 0.13
Nodes (17): SnapNode, Snapshot, Tx, snapNode(), Compare(), describeDNS(), describeRules(), fields() (+9 more)

### Community 46 - "Import"
Cohesion: 0.25
Nodes (11): NewHub(), closed(), TestHubLatestWins(), TestHubReplacesKicksAndCloses(), Import(), document(), linuxOnly(), TestImportRefusesInvalidDocumentsUntouched() (+3 more)

### Community 47 - "ServerConfig"
Cohesion: 0.21
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestApplyIsIdempotent() (+17 more)

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
Cohesion: 0.08
Nodes (32): Config, DB, Params, Pending, Sealer, addUser(), newPassword(), passwdUser() (+24 more)

### Community 53 - "Client"
Cohesion: 0.36
Nodes (4): apiError(), clip(), Client, StatusResponse

### Community 54 - "Node"
Cohesion: 0.27
Nodes (6): reachable(), Node, Tx, scanNode(), nullStr(), scanner

### Community 55 - "env"
Cohesion: 0.27
Nodes (24): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), setupVersion() (+16 more)

### Community 56 - "CookieValue"
Cohesion: 0.23
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 57 - "Key"
Cohesion: 0.15
Nodes (16): peer(), pfx(), TestConfigureHasKernelSemantics(), ParseAddr(), TestParseAddr(), AddrPort(), TestAddrPort(), hexKey() (+8 more)

### Community 58 - "NodeChange"
Cohesion: 0.12
Nodes (17): ComposeFile, NodeDetail, NodeList, PhoneProfile, statusDetail, heard(), PeerStatus, Rollback (+9 more)

### Community 59 - "net/http.Request"
Cohesion: 0.17
Nodes (6): ClientIP(), inAny(), TestClientIP(), before(), stamp(), versionNumber()

### Community 60 - "log/slog.Logger"
Cohesion: 0.12
Nodes (11): Registrar, Service, New(), NewRegistrar(), NewRegistrar(), exchange(), TestOpen(), Open() (+3 more)

### Community 61 - "Hub"
Cohesion: 0.31
Nodes (3): Hub, subscription, ConfigEvent

### Community 63 - "ServerURL"
Cohesion: 0.29
Nodes (10): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams, isLoopbackHost() (+2 more)

### Community 64 - "Registration"
Cohesion: 0.18
Nodes (9): Domain, fakeRegistrar, linkDomain, Registration, resolved, desired(), classify(), foreignNetns() (+1 more)

### Community 65 - "Run"
Cohesion: 0.18
Nodes (9): LoadOrCreateKey(), NewSealer(), TestLoadOrCreateKey(), TestSealerBindsTheSecretToItsUser(), hostOf(), Routes(), Run(), serve() (+1 more)

### Community 66 - "Deps"
Cohesion: 0.22
Nodes (13): Deps, Route, AdminRoutes(), describe(), api, New(), NewAdmin(), newAPI() (+5 more)

### Community 67 - "session"
Cohesion: 0.31
Nodes (6): ago(), bytesize(), UI, prefsFlash(), page, session

### Community 68 - "execrun_test.go"
Cohesion: 0.12
Nodes (18): TestUnitPosture(), SystemdUnit(), Check(), fakeProgram(), TestCappedKeepsTheHead(), TestCheckAllowsFirewallArgv(), TestCheckRefuses(), TestRunCapturesOutputWithAMinimalEnvironment() (+10 more)

### Community 69 - "Arg"
Cohesion: 0.40
Nodes (4): Arg(), FuzzArg(), FuzzName(), TestArg()

### Community 70 - "nullInt"
Cohesion: 0.16
Nodes (6): Tx, Tx, nullInt(), AuditEntry, DNSRecord, FirewallRule

### Community 72 - "fail"
Cohesion: 0.11
Nodes (32): check(), DNS, Firewall, Interface, NodeConfig, newNode(), validateSet(), ToSnapshot() (+24 more)

### Community 74 - "sync.Mutex"
Cohesion: 0.12
Nodes (7): syncBuffer, DNS, Sysctl, Want, New(), Sysctl, Writer

### Community 76 - "net.Listener"
Cohesion: 0.14
Nodes (13): resolver, bindPair(), serve(), Listen(), Listen(), listenPipe(), pipeName(), servePipe() (+5 more)

### Community 77 - "loadMigrations"
Cohesion: 0.31
Nodes (4): Store, Tx, loadMigrations(), migration

### Community 78 - "Change"
Cohesion: 0.18
Nodes (7): pfxs(), TestUAPISetGolden(), uapiChange(), UAPISet(), Change, samePrefixes(), SortPrefixes()

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

### Community 83 - "main.go"
Cohesion: 0.21
Nodes (14): agentCommands(), commands(), runWith(), healthCommands(), find(), isHelp(), main(), printUsage() (+6 more)

### Community 87 - "fakeServer"
Cohesion: 0.17
Nodes (11): fakeServer, newFakeServer(), TestMeshViewIsRefreshed(), TestMeshCache(), ErrorBody, StatusReport, TestStatusReportValidate(), validStatus() (+3 more)

### Community 88 - "changed"
Cohesion: 0.16
Nodes (7): ui, ui, changed(), ui, newQR(), phoneData, qrCode

### Community 89 - "6. Control plane"
Cohesion: 0.40
Nodes (5): 6.1 Enrollment, 6.2 Request authentication, 6.3 Config sync, 6.4 Safe apply, 6. Control plane

### Community 90 - "8. Server"
Cohesion: 0.40
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 92 - "net/http.Handler"
Cohesion: 0.30
Nodes (7): staticHandler(), ui, headers(), New(), sameOrigin(), Deps, handler

### Community 93 - "time.Duration"
Cohesion: 0.16
Nodes (9): hitWindow, waitApplied(), Exec, RateLimiter, NewRateLimiter(), TestRateLimiter(), Tx, HashToken() (+1 more)

### Community 94 - "errorf"
Cohesion: 0.23
Nodes (11): apiError, MeshChange, RecordRequest, RuleRequest, Actor, errorf(), invalid(), Nodes (+3 more)

### Community 96 - "NewNodes"
Cohesion: 0.19
Nodes (11): ErrorMessage(), NewNodes(), TestAdminCompose(), toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion(), harness, meshHarness() (+3 more)

### Community 97 - "route"
Cohesion: 0.43
Nodes (3): index(), route(), netlinkTable

### Community 98 - ".DNS"
Cohesion: 0.40
Nodes (4): DNSEntry, DNSView, dnsData, recordForm

### Community 99 - "commands.go"
Cohesion: 0.21
Nodes (13): argSynopsis(), newFlags(), newLogger(), parseArgs(), parseFlags(), parsePrefixes(), cell(), renderCommands() (+5 more)

### Community 100 - "meshctl"
Cohesion: 0.40
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 102 - "Allocate"
Cohesion: 0.25
Nodes (9): Allocate(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan(), TestCheckMeshCIDR(), TestRangeFor() (+1 more)

### Community 103 - "mustAffect"
Cohesion: 0.16
Nodes (5): mustAffect(), Tx, classify(), PeerStat, StatusUpdate

### Community 104 - "Document"
Cohesion: 0.17
Nodes (6): watchdog, Document, Read(), Write(), Config, ImportConfig

### Community 106 - "newDir"
Cohesion: 0.20
Nodes (12): newDir(), shortTemp(), TestJoin(), TestRunNeedsAnEnrollment(), TestRunRefusesAnUnusableDataDir(), NewHost(), Open(), mkdirs() (+4 more)

### Community 107 - "File"
Cohesion: 0.25
Nodes (6): serviceLog(), TestServiceLog(), ttyPassword(), File, OpenFile(), TestFileRotates()

### Community 108 - "Join"
Cohesion: 0.27
Nodes (8): installAgent(), installAgent(), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), cleanHostname(), JoinOptions, Service, Join()

### Community 109 - "renderer"
Cohesion: 0.20
Nodes (8): loadPages(), dnsState(), newRenderer(), golden(), TestEveryPageRenders(), TestPageGoldens(), pages, renderer

### Community 110 - "safeRig"
Cohesion: 0.38
Nodes (9): pendingExists(), safeRig(), TestSafeApplyConfirmsWhenTheServerAnswers(), TestSafeApplyDisarmsWhenTheServerGoesBack(), TestSafeApplyRecoversAPendingConfigAtBoot(), TestSafeApplyReportsAFailedRollback(), TestSafeApplyRollsBackWithoutAnAnswer(), TestSpokesDoNotArmSafeApply() (+1 more)

### Community 111 - "PhoneConfig"
Cohesion: 0.31
Nodes (9): NewKey(), Render(), aliceKey(), golden(), TestHub(), TestNewKey(), TestProfiles(), TestRenderRefuses() (+1 more)

### Community 112 - "Exporter"
Cohesion: 0.33
Nodes (4): DB, NextExport(), TestNextExport(), Exporter

### Community 113 - "validJoin"
Cohesion: 0.29
Nodes (7): JoinRequest, TestJoinRequestErrorNeverEchoesTheToken(), TestJoinRequestValidateNormalises(), TestJoinRequestValidateRejects(), validJoin(), pickEndpoint(), TestPickEndpoint()

## Knowledge Gaps
- **157 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `Interface`, `Exit` (+152 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 359 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **25 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Dir` connect `Dir` to `context.Context`, `ui/ui_test.go`, `agent_test.go`, `newDir`, `safeRig`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `NodeConfig` connect `NodeConfig` to `context.Context`, `NewNodes`, `ui/ui_test.go`, `agent_test.go`, `Tx`, `safeRig`, `PhoneConfig`, `ServerConfig`, `Engine`, `Dir`, `Client`, `net/netip.Prefix`, `fakeServer`, `WGKey`, `NodeChange`, `Hub`, `time.Time`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Host` connect `Host` to `ui/ui_test.go`, `agent_test.go`, `sync.Mutex`, `newDir`, `Change`, `Key`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _157 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08114678094378869 - nodes in this community are weakly interconnected._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.05627705627705628 - nodes in this community are weakly interconnected._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.06452072354365475 - nodes in this community are weakly interconnected._