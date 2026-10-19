# Graph Report - vpn-mesh-utility  (2026-10-20)

## Corpus Check
- 213 files · ~199,179 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 2625 nodes · 10935 edges · 124 communities (99 shown, 25 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 601 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7599f617`
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
- api/api_test.go
- .Apply
- Run
- ServerURL
- Tx
- Dir
- IPNet
- newRig
- net/netip.Prefix
- Host
- .nodeView
- LocalStatus
- lab
- service_windows.go
- Session
- Command reference
- time.Time
- Key
- context.Context
- newRig
- Tx
- agent_test.go
- testing.T
- Test catalogue
- IdentityKey
- Policy
- Tx
- Registration
- Agent protocol
- net/http.ResponseWriter
- newRig
- Snapshot
- runArgs
- ServerConfig
- CLAUDE.md
- Security
- Host
- Service
- Client
- mustAffect
- UI
- CookieValue
- Sealer
- NodeChange
- commands.go
- Service
- log/slog.Logger
- Ring
- host
- wgdev
- .change
- io.Writer
- session
- loadMigrations
- Arg
- Tx
- otherPolicy
- fail
- StepError
- Run
- newHarness
- net/netip.AddrPort
- route
- servePipe
- 7. Agent
- WGKey
- Diff
- .show
- New
- .validate
- NewReader
- Refusal
- fetch
- changed
- 6. Control plane
- 8. Server
- .Seen
- Route
- Firewall
- .AddRule
- errorf
- toAPIError
- Error
- userspace
- serve
- meshctl
- AdminError
- net/http.Request
- File
- net.Listener
- Tx
- Change
- Allocate
- Join
- nullInt
- Tx
- NodeConfig
- .install
- .syncPolicies
- Check
- newReplayCache
- TestMain
- Node
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

## Communities (124 total, 25 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.07
Nodes (18): nasFile, nasHealth, nasService, linkDNS, DNSRecord, FirewallRule, Node, Exit (+10 more)

### Community 1 - "newRig"
Cohesion: 0.13
Nodes (15): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+7 more)

### Community 2 - "Open"
Cohesion: 0.10
Nodes (33): Document, Import(), document(), linuxOnly(), TestImportRefusesInvalidDocumentsUntouched(), TestImportRefusesWhileTheServerRuns(), TestImportRestoresAnExport(), TestLockIsExclusive() (+25 more)

### Community 3 - "planner_test.go"
Cohesion: 0.05
Nodes (60): PlanError, zone, TestIsPublic(), answerOf(), ask(), meshDNS(), newRegistrar(), query() (+52 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (118): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+110 more)

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

### Community 11 - "RenderNASCompose"
Cohesion: 0.24
Nodes (11): comment(), nasHeader(), renderNAS(), RenderNASCompose(), nasParams(), TestNASComposeRefusesHostileValues(), NASParams, POSIXPath() (+3 more)

### Community 14 - "api/api_test.go"
Cohesion: 0.23
Nodes (19): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+11 more)

### Community 15 - ".Apply"
Cohesion: 0.28
Nodes (8): applied, Firewall, atLeast(), normalize(), richRule(), RuleArgv(), FWRule, sortRules()

### Community 16 - "Run"
Cohesion: 0.19
Nodes (14): Deps, api, New(), NewAdmin(), newAPI(), LoadOrCreateKey(), TestLoadOrCreateKey(), hostOf() (+6 more)

### Community 17 - "ServerURL"
Cohesion: 0.29
Nodes (10): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams, isLoopbackHost() (+2 more)

### Community 19 - "Dir"
Cohesion: 0.13
Nodes (7): Dir, Enrollment, syncDir(), TestWriteAtomicReplacesAndLeavesNoTempFiles(), WriteAtomic(), Keys, Local

### Community 20 - "IPNet"
Cohesion: 0.15
Nodes (7): IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), lookup(), kernel, links

### Community 21 - "newRig"
Cohesion: 0.09
Nodes (36): rig, TestRunNeedsAnEnrollment(), NewHost(), TestConfigureHasKernelSemantics(), TestLinksAddressesAndFailures(), IsStepError(), New(), meshDNS() (+28 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.12
Nodes (13): Routes, pfx(), NewTable(), compare(), Table, Sync(), NewTable(), luidOf() (+5 more)

### Community 24 - ".nodeView"
Cohesion: 0.15
Nodes (11): Health(), phoneHealth(), TestDashboardHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeForm (+3 more)

### Community 25 - "LocalStatus"
Cohesion: 0.09
Nodes (22): Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest, PrefsResponse (+14 more)

### Community 26 - "lab"
Cohesion: 0.14
Nodes (15): host, lab, stream, dial(), dial(), isPipe(), deleteNamespaces(), eventually() (+7 more)

### Community 27 - "service_windows.go"
Cohesion: 0.12
Nodes (9): uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows(), handler (+1 more)

### Community 28 - "Session"
Cohesion: 0.17
Nodes (7): Session, atoi(), ui, ui, message(), actor(), settingsData

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (29): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+21 more)

### Community 30 - "time.Time"
Cohesion: 0.07
Nodes (18): pendingApply, safeApply, syncBuffer, replayCache, failures, hitWindow, Lockout, LockoutError (+10 more)

### Community 31 - "Key"
Cohesion: 0.22
Nodes (11): peer(), hexKey(), ParseUAPI(), uapiInt(), compareKeys(), Device, Key, Peer (+3 more)

### Community 32 - "context.Context"
Cohesion: 0.14
Nodes (6): fakeService, runner, kick(), oneLine(), RunWindowsService(), droppedLogs

### Community 33 - "newRig"
Cohesion: 0.15
Nodes (27): NewHub(), closed(), TestHubLatestWins(), TestHubReplacesKicksAndCloses(), TestPhonePage(), code(), cookieOf(), newRig() (+19 more)

### Community 34 - "Tx"
Cohesion: 0.11
Nodes (18): PhoneRequest, checkName(), defaultMTU(), Actor, Nodes, NewNodes(), nodeByName(), nodeInfo() (+10 more)

### Community 35 - "agent_test.go"
Cohesion: 0.13
Nodes (47): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), newFakeServer() (+39 more)

### Community 36 - "testing.T"
Cohesion: 0.04
Nodes (81): checkPosture(), keys(), mustRead(), sameSet(), service(), TestDevComposePosture(), TestServerComposePosture(), ComposeServer() (+73 more)

### Community 37 - "Test catalogue"
Cohesion: 0.10
Nodes (20): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+12 more)

### Community 38 - "IdentityKey"
Cohesion: 0.12
Nodes (22): newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestPrefsRoundTrip(), TestStatusRoundTrip(), verifying() (+14 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (20): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+12 more)

### Community 40 - "Tx"
Cohesion: 0.18
Nodes (4): boolInt(), Session, Tx, scanUser()

### Community 41 - "Registration"
Cohesion: 0.16
Nodes (11): Domain, fakeRegistrar, linkDomain, Registration, resolved, domainName(), same(), TestSameRegistration() (+3 more)

### Community 42 - "Agent protocol"
Cohesion: 0.17
Nodes (12): Agent protocol, Apply (agent), Config and events, Errors, Join, Local API, Mesh view, Prefs (+4 more)

### Community 43 - "net/http.ResponseWriter"
Cohesion: 0.29
Nodes (5): decode(), reply(), cli(), api, api

### Community 44 - "newRig"
Cohesion: 0.19
Nodes (14): rig, newRig(), TestAccountChangesEndSessions(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures(), TestLoginRefusals(), TestLoginUpgradesAnOldHash() (+6 more)

### Community 45 - "Snapshot"
Cohesion: 0.12
Nodes (19): TestReconcilerVersionsOnlyRealChanges(), DecodeSnapshot(), SnapNode, Snapshot, snapNode(), TestDecodeSnapshotRefusesUnknownShapes(), Compare(), describeDNS() (+11 more)

### Community 46 - "runArgs"
Cohesion: 0.06
Nodes (52): passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+44 more)

### Community 47 - "ServerConfig"
Cohesion: 0.20
Nodes (26): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestApplyIsIdempotent() (+18 more)

### Community 49 - "CLAUDE.md"
Cohesion: 0.25
Nodes (7): Determinism and goldens, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates

### Community 50 - "Security"
Cohesion: 0.12
Nodes (15): Agent service (systemd), Changes a node asks for, DNS, Input and commands, Interfaces the agent did not create, Keys and tokens, NAS container, Security (+7 more)

### Community 51 - "Host"
Cohesion: 0.36
Nodes (3): Host, SyncAddr(), SyncLink()

### Community 52 - "Service"
Cohesion: 0.09
Nodes (31): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), resetTOTP() (+23 more)

### Community 53 - "Client"
Cohesion: 0.13
Nodes (15): statusDetail, Options, apiError(), clip(), Client, Join(), New(), TestJoinIsUnsigned() (+7 more)

### Community 54 - "mustAffect"
Cohesion: 0.16
Nodes (5): mustAffect(), Tx, classify(), PeerStat, StatusUpdate

### Community 55 - "UI"
Cohesion: 0.35
Nodes (4): UI, oldest(), LocalToken(), TestLocalTokenMasksTheValue()

### Community 56 - "CookieValue"
Cohesion: 0.23
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 57 - "Sealer"
Cohesion: 0.33
Nodes (4): Sealer, aad(), NewSealer(), TestSealerBindsTheSecretToItsUser()

### Community 58 - "NodeChange"
Cohesion: 0.16
Nodes (13): ComposeFile, NodeAddRequest, NodeList, NodeSetRequest, PhoneProfile, heard(), AdminClient, nodePath() (+5 more)

### Community 59 - "commands.go"
Cohesion: 0.14
Nodes (22): agentCommands(), argSynopsis(), commands(), newFlags(), parseArgs(), parseFlags(), cell(), renderCommands() (+14 more)

### Community 60 - "Service"
Cohesion: 0.13
Nodes (8): Registrar, Service, DNS, desired(), Want, New(), NewRegistrar(), NewRegistrar()

### Community 61 - "log/slog.Logger"
Cohesion: 0.09
Nodes (17): Change, Hub, subscription, TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), ConfigEvent (+9 more)

### Community 63 - "host"
Cohesion: 0.15
Nodes (12): hostService(), TestUnitPosture(), SystemdUnit(), Runner, exists(), Systemd(), host(), TestAvailable() (+4 more)

### Community 64 - "wgdev"
Cohesion: 0.17
Nodes (4): userspace, NewUserspace(), Link, wgdev

### Community 65 - ".change"
Cohesion: 0.22
Nodes (6): hostOf(), validOrNot(), ui, sameOrigin(), handler, handler

### Community 66 - "io.Writer"
Cohesion: 0.17
Nodes (23): TestBytesize(), dnsLine(), printExit(), adminFlags(), cleared(), joinNonEmpty(), nodeAction(), nodeName() (+15 more)

### Community 67 - "session"
Cohesion: 0.31
Nodes (6): ago(), bytesize(), UI, prefsFlash(), page, session

### Community 68 - "loadMigrations"
Cohesion: 0.27
Nodes (5): Store, Tx, loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 69 - "Arg"
Cohesion: 0.40
Nodes (4): Arg(), FuzzArg(), FuzzName(), TestArg()

### Community 70 - "Tx"
Cohesion: 0.28
Nodes (3): Tx, DNSRecord, FirewallRule

### Community 72 - "fail"
Cohesion: 0.11
Nodes (32): check(), DNS, Firewall, Interface, NodeConfig, newNode(), validateSet(), ToSnapshot() (+24 more)

### Community 74 - "Run"
Cohesion: 0.14
Nodes (10): Deps, Deps, Run(), Sysctl, DNS, Firewall, Sysctls, New() (+2 more)

### Community 75 - "newHarness"
Cohesion: 0.15
Nodes (20): ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing(), TestAdminSetClearsReferences() (+12 more)

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.17
Nodes (7): resolver, bindPair(), serve(), ParseAddr(), TestParseAddr(), AddrPort(), TestAddrPort()

### Community 77 - "route"
Cohesion: 0.36
Nodes (3): index(), route(), netlinkTable

### Community 78 - "servePipe"
Cohesion: 0.36
Nodes (8): listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe(), waitPipe(), PipePath(), TestPipePath()

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "WGKey"
Cohesion: 0.11
Nodes (24): TestImportCommand(), TestNodePhone(), TestNodeShowAnEnrolledServer(), JoinRequest, TestJoinRequestErrorNeverEchoesTheToken(), TestJoinRequestValidateNormalises(), TestJoinRequestValidateRejects(), validJoin() (+16 more)

### Community 81 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 82 - ".show"
Cohesion: 0.36
Nodes (3): ui, writeHTML(), page

### Community 83 - "New"
Cohesion: 0.15
Nodes (11): Listen(), New(), TestServeAndListen(), loadPages(), dnsState(), newRenderer(), golden(), TestEveryPageRenders() (+3 more)

### Community 85 - "NewReader"
Cohesion: 0.13
Nodes (14): watchdog, TestEventsDeliversConfigEvents(), NewReader(), TestSSEReaderBoundsLineLength(), TestSSEReaderJoinsDataLinesAndDefaultsName(), TestSSEReaderReportsTruncatedEvent(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection() (+6 more)

### Community 87 - "fetch"
Cohesion: 0.19
Nodes (14): NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall(), extract() (+6 more)

### Community 88 - "changed"
Cohesion: 0.25
Nodes (3): ui, ui, changed()

### Community 89 - "6. Control plane"
Cohesion: 0.40
Nodes (5): 6.1 Enrollment, 6.2 Request authentication, 6.3 Config sync, 6.4 Safe apply, 6. Control plane

### Community 90 - "8. Server"
Cohesion: 0.40
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 92 - "Route"
Cohesion: 0.60
Nodes (5): Route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 93 - "Firewall"
Cohesion: 0.20
Nodes (8): CustomRule, ServerFirewall, Firewall, wantedPolicies(), Firewall, FirewallView, firewallData, ruleForm

### Community 94 - ".AddRule"
Cohesion: 0.16
Nodes (11): DNSEntry, MeshChange, RecordRequest, RuleRequest, DNSView, Nodes, meshChange(), DNSName() (+3 more)

### Community 95 - "errorf"
Cohesion: 0.30
Nodes (9): Role, MeshView, api, decodeJSON(), errorf(), invalid(), readBody(), writeJSON() (+1 more)

### Community 96 - "toAPIError"
Cohesion: 0.18
Nodes (11): apiError, ErrorMessage(), TestAdminCompose(), toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion(), harness, meshHarness() (+3 more)

### Community 99 - "serve"
Cohesion: 0.24
Nodes (8): brokenStream, serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums()

### Community 100 - "meshctl"
Cohesion: 0.40
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 102 - "net/http.Request"
Cohesion: 0.18
Nodes (4): fakeServer, before(), stamp(), versionNumber()

### Community 103 - "File"
Cohesion: 0.27
Nodes (5): agentLog(), ttyPassword(), OpenFile(), TestFileRotates(), File

### Community 104 - "net.Listener"
Cohesion: 0.24
Nodes (7): Listen(), Listen(), listenSocket(), ListenAdmin(), serve(), TestListenAdminRefusesALongPath(), TestRunServesAndShutsDown()

### Community 106 - "Change"
Cohesion: 0.22
Nodes (6): pfxs(), TestUAPISetGolden(), uapiChange(), UAPISet(), Change, PeerChange

### Community 107 - "Allocate"
Cohesion: 0.25
Nodes (9): Allocate(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan(), TestCheckMeshCIDR(), TestRangeFor() (+1 more)

### Community 108 - "Join"
Cohesion: 0.27
Nodes (8): installAgent(), hostService(), installAgent(), cleanHostname(), JoinOptions, Service, Join(), JoinResponse

### Community 109 - "nullInt"
Cohesion: 0.32
Nodes (3): Tx, nullInt(), nullStr()

### Community 110 - "Tx"
Cohesion: 0.43
Nodes (3): Tx, scanNode(), scanner

### Community 111 - "NodeConfig"
Cohesion: 0.12
Nodes (10): NodeDetail, Rechecker, Engine, Result, Desired(), PeerStatus, NodeConfig, Hub() (+2 more)

### Community 114 - "Check"
Cohesion: 0.67
Nodes (3): Check(), checkDNS(), sortedByID()

### Community 115 - "newReplayCache"
Cohesion: 0.67
Nodes (3): newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries()

## Knowledge Gaps
- **158 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `Interface`, `Exit` (+153 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 361 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **25 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NodeConfig` connect `NodeConfig` to `context.Context`, `toAPIError`, `ui/ui_test.go`, `agent_test.go`, `testing.T`, `Tx`, `net/http.Request`, `planner_test.go`, `Run`, `ServerConfig`, `Dir`, `Client`, `newRig`, `LocalStatus`, `log/slog.Logger`, `time.Time`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Why does `runner` connect `context.Context` to `ui/ui_test.go`, `newRig`, `Run`, `NodeConfig`, `Dir`, `Client`, `LocalStatus`, `log/slog.Logger`, `time.Time`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `Host` connect `Host` to `ui/ui_test.go`, `agent_test.go`, `newRig`, `time.Time`, `Key`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _158 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07333171987415019 - nodes in this community are weakly interconnected._
- **Should `newRig` be split into smaller, more focused modules?**
  _Cohesion score 0.13333333333333333 - nodes in this community are weakly interconnected._
- **Should `Open` be split into smaller, more focused modules?**
  _Cohesion score 0.09872241579558652 - nodes in this community are weakly interconnected._