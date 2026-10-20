# Graph Report - vpn-mesh-utility  (2026-10-20)

## Corpus Check
- 220 files · ~204,146 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 2674 nodes · 11105 edges · 113 communities (101 shown, 12 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 620 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `acbddf2f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- store_test.go
- Name
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
- serve
- Handler
- Run
- RenderInstall
- storeNode
- Dir
- IPNet
- WGKey
- net/netip.Prefix
- newHarness
- .nodeView
- LocalStatus
- lab
- service_windows.go
- Session
- Command reference
- time.Time
- Host
- runner
- newRig
- .Phone
- agent_test.go
- testing.T
- Test catalogue
- IdentityKey
- Policy
- Tx
- Snapshot
- Agent protocol
- net/http.Request
- newRig
- NodeConfig
- dns_test.go
- ServerConfig
- CLAUDE.md
- Security
- Host
- Service
- Client
- nrpt_windows.go
- mustAffect
- CookieValue
- Node
- context.Context
- DNS
- log/slog.Logger
- Tx
- Ring
- .Apply
- wgdev
- UI
- NodeInfo
- session
- loadMigrations
- validate_test.go
- meshHarness
- otherPolicy
- validate.go
- apply.go
- Writer
- host
- net/netip.AddrPort
- net/http.Handler
- Reconciler
- 7. Agent
- Registration
- FromSnapshot
- .show
- Diff
- syncBuffer
- NewReader
- .renderDNS
- fetch
- errorf
- 6. Control plane
- 8. Server
- .Seen
- luidOf
- RenderNASCompose
- resolved
- net/http.ResponseWriter
- Hub
- uapiChange
- userspace
- route
- meshctl
- newQR
- Open
- Allocate
- PhoneConfig
- Tx
- agentfake.go
- go_pkg_net_netip
- nullInt
- Engine
- .install
- .renderFirewall
- agent/agent.go

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
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 5-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 5-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`

## Communities (113 total, 12 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (31): AdminError, DNSEntry, Error, nasFile, nasHealth, nasService, DNSRecord, FirewallRule (+23 more)

### Community 1 - "newRig"
Cohesion: 0.22
Nodes (13): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+5 more)

### Community 2 - "store_test.go"
Cohesion: 0.23
Nodes (22): insertTopology(), node(), openTest(), TestAudit(), TestDeleteNodeCascades(), TestNodeCRUD(), TestNodeUniquenessIsAConflict(), TestNotFound() (+14 more)

### Community 3 - "Name"
Cohesion: 0.14
Nodes (11): Exit, Check(), checkDNS(), sortedByID(), Name(), SemVer(), TestName(), TestText() (+3 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (118): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+110 more)

### Community 5 - "env"
Cohesion: 0.05
Nodes (88): watchdog, agentCommands(), agentLog(), joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall() (+80 more)

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
Cohesion: 0.11
Nodes (29): TestRunNeedsAnEnrollment(), NewHost(), peer(), TestConfigureHasKernelSemantics(), TestLinksAddressesAndFailures(), hexKey(), ParseUAPI(), getAnswer() (+21 more)

### Community 14 - "serve"
Cohesion: 0.18
Nodes (19): countingDB, harness, serve(), signed(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend(), TestEventsStreamPushesChanges(), TestHealth() (+11 more)

### Community 15 - "Handler"
Cohesion: 0.11
Nodes (28): TestStatusAndAgentHealthcheck(), TestLogsCommand(), TestUICommand(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe() (+20 more)

### Community 16 - "Run"
Cohesion: 0.12
Nodes (20): Deps, Route, NewNodes(), AdminRoutes(), describe(), api, New(), NewAdmin() (+12 more)

### Community 17 - "RenderInstall"
Cohesion: 0.48
Nodes (7): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams

### Community 18 - "storeNode"
Cohesion: 0.44
Nodes (10): joinBody(), postJoin(), storeNode(), TestJoinEnrollsAndConsumesTheToken(), TestJoinRefusesAPhone(), TestJoinRefusesHostileInputUpFront(), TestRefusedJoinKeepsTheToken(), TestStoreFailuresAreInternalErrors() (+2 more)

### Community 19 - "Dir"
Cohesion: 0.11
Nodes (15): TestCleanHostname(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), cleanHostname(), Join(), Dir, Enrollment, Open() (+7 more)

### Community 20 - "IPNet"
Cohesion: 0.15
Nodes (7): IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), lookup(), kernel, links

### Community 21 - "WGKey"
Cohesion: 0.11
Nodes (30): rig, NewRoutes(), TestRoutesSysctlAndFirewall(), IsStepError(), New(), meshDNS(), newRig(), TestApplyIsIdempotent() (+22 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.11
Nodes (15): Routes, pfx(), NewTable(), compare(), Table, Sync(), NewTable(), NewTable() (+7 more)

### Community 23 - "newHarness"
Cohesion: 0.12
Nodes (24): apiError, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose(), TestAdminErrorCodes() (+16 more)

### Community 24 - ".nodeView"
Cohesion: 0.13
Nodes (12): Health(), phoneHealth(), TestDashboardHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeData (+4 more)

### Community 25 - "LocalStatus"
Cohesion: 0.09
Nodes (16): Client, Logs, logs(), refusal(), laptop(), PrefsRequest, PrefsResponse, LocalMesh (+8 more)

### Community 26 - "lab"
Cohesion: 0.16
Nodes (12): host, lab, stream, deleteNamespaces(), eventually(), inNS(), newLab(), ns() (+4 more)

### Community 27 - "service_windows.go"
Cohesion: 0.07
Nodes (18): hostService(), hostService(), uninstallAgent(), Service, RunWindowsService(), TestNoWindowsServiceHere(), Systemd(), commandLine() (+10 more)

### Community 28 - "Session"
Cohesion: 0.13
Nodes (11): Session, atoi(), before(), changed(), ui, stamp(), versionNumber(), ui (+3 more)

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (29): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+21 more)

### Community 30 - "time.Time"
Cohesion: 0.08
Nodes (17): pendingApply, safeApply, syncBuffer, replayCache, failures, hitWindow, Lockout, LockoutError (+9 more)

### Community 31 - "Host"
Cohesion: 0.17
Nodes (3): link, exitHost, Host

### Community 32 - "runner"
Cohesion: 0.14
Nodes (5): Deps, runner, kick(), oneLine(), Run()

### Community 33 - "newRig"
Cohesion: 0.18
Nodes (23): brokenStream, TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled(), TestChangesNeedOriginAndCSRF(), TestDashboardFragmentAndExport() (+15 more)

### Community 34 - ".Phone"
Cohesion: 0.16
Nodes (14): NodeAddRequest, NodeSetRequest, PhoneRequest, checkName(), Actor, Nodes, nodeByName(), nodeInfo() (+6 more)

### Community 35 - "agent_test.go"
Cohesion: 0.11
Nodes (53): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+45 more)

### Community 36 - "testing.T"
Cohesion: 0.04
Nodes (83): shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin(), TestExportCommand(), TestHealthcheckServer() (+75 more)

### Community 37 - "Test catalogue"
Cohesion: 0.10
Nodes (20): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+12 more)

### Community 38 - "IdentityKey"
Cohesion: 0.21
Nodes (12): CanonicalRequest(), ParseSigned(), Sign(), TestCheckClockNeedsNoKey(), TestParseSignedRejectsMalformedHeaders(), TestSignIsUniquePerRequest(), TestSignVerifyRoundTrip(), TestVerifyRejectsMissingKey() (+4 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (20): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+12 more)

### Community 40 - "Tx"
Cohesion: 0.18
Nodes (4): boolInt(), Session, Tx, scanUser()

### Community 41 - "Snapshot"
Cohesion: 0.13
Nodes (17): SnapNode, Snapshot, Tx, snapNode(), Compare(), describeDNS(), describeRules(), fields() (+9 more)

### Community 42 - "Agent protocol"
Cohesion: 0.17
Nodes (12): Agent protocol, Apply (agent), Config and events, Errors, Join, Local API, Mesh view, Prefs (+4 more)

### Community 43 - "net/http.Request"
Cohesion: 0.25
Nodes (5): cli(), api, api, ClientIP(), TestClientIP()

### Community 44 - "newRig"
Cohesion: 0.12
Nodes (19): rig, Sealer, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures() (+11 more)

### Community 45 - "NodeConfig"
Cohesion: 0.05
Nodes (50): PlanError, zone, TestIsPublic(), isPublic(), PublicIPv4(), JoinRequest, JoinResponse, NodeConfig (+42 more)

### Community 46 - "dns_test.go"
Cohesion: 0.25
Nodes (16): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceBinding() (+8 more)

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
Cohesion: 0.24
Nodes (4): Host, Link, SyncAddr(), SyncLink()

### Community 52 - "Service"
Cohesion: 0.11
Nodes (27): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), fullExpiry() (+19 more)

### Community 53 - "Client"
Cohesion: 0.21
Nodes (9): Options, apiError(), clip(), Client, Join(), New(), TestJoinIsUnsigned(), TestUnreachableAndMalformedAnswers() (+1 more)

### Community 54 - "nrpt_windows.go"
Cohesion: 0.18
Nodes (11): nrpt, nrptRule, domainName(), fromRules(), parseServers(), rulesFor(), serverList(), TestNRPTServers() (+3 more)

### Community 55 - "mustAffect"
Cohesion: 0.16
Nodes (5): mustAffect(), Tx, classify(), PeerStat, StatusUpdate

### Community 56 - "CookieValue"
Cohesion: 0.23
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 57 - "Node"
Cohesion: 0.27
Nodes (6): reachable(), Node, Tx, scanNode(), nullStr(), scanner

### Community 58 - "context.Context"
Cohesion: 0.15
Nodes (12): fakeService, ComposeFile, resetTOTP(), dial(), dial(), isPipe(), AdminClient, NewAdminClient() (+4 more)

### Community 59 - "DNS"
Cohesion: 0.22
Nodes (6): DNS, desired(), Want, same(), TestSameRegistration(), TestNRPTRules()

### Community 60 - "log/slog.Logger"
Cohesion: 0.15
Nodes (8): Registrar, Service, New(), NewRegistrar(), NewRegistrar(), exchange(), NewUserspace(), Options

### Community 61 - "Tx"
Cohesion: 0.12
Nodes (7): defaultMTU(), DB, NextExport(), TestNextExport(), Tx, Exporter, Deps

### Community 63 - ".Apply"
Cohesion: 0.11
Nodes (17): CustomRule, ServerFirewall, applied, Firewall, policy, Firewall, atLeast(), normalize() (+9 more)

### Community 64 - "wgdev"
Cohesion: 0.22
Nodes (3): userspace, NewUserspace(), wgdev

### Community 65 - "UI"
Cohesion: 0.25
Nodes (6): UI, New(), oldest(), loadPages(), Config, pages

### Community 66 - "NodeInfo"
Cohesion: 0.18
Nodes (10): NodeDetail, NodeList, PhoneProfile, statusDetail, heard(), PeerStatus, Rollback, NodeInfo (+2 more)

### Community 67 - "session"
Cohesion: 0.31
Nodes (6): ago(), bytesize(), UI, prefsFlash(), page, session

### Community 68 - "loadMigrations"
Cohesion: 0.27
Nodes (5): Store, Tx, loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 69 - "validate_test.go"
Cohesion: 0.11
Nodes (18): Arg(), POSIXPath(), FuzzArg(), FuzzName(), TestArg(), TestDNSName(), TestEndpoint(), TestHostnameFoldsCase() (+10 more)

### Community 70 - "meshHarness"
Cohesion: 0.43
Nodes (5): harness, meshHarness(), refused(), TestCustomDNSRecords(), TestCustomFirewallRules()

### Community 72 - "validate.go"
Cohesion: 0.13
Nodes (29): check(), DNS, Firewall, Interface, NodeConfig, newNode(), validateSet(), ToSnapshot() (+21 more)

### Community 73 - "apply.go"
Cohesion: 0.15
Nodes (6): Rechecker, StepError, NewPolicy(), Desired(), prefixes(), TestDesiredSkipsTheDefaultRouteAndDuplicates()

### Community 74 - "Writer"
Cohesion: 0.20
Nodes (5): Sysctl, New(), New(), Sysctl, Writer

### Community 75 - "host"
Cohesion: 0.17
Nodes (10): TestUnitPosture(), SystemdUnit(), Runner, exists(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere(), TestInstallWritesTheUnitAndStartsIt() (+2 more)

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.15
Nodes (12): resolver, bindBoth(), bindPair(), serve(), Listen(), Listen(), listenSocket(), Listen() (+4 more)

### Community 77 - "net/http.Handler"
Cohesion: 0.13
Nodes (15): hostOf(), validOrNot(), dnsState(), newRenderer(), staticHandler(), ui, headers(), New() (+7 more)

### Community 78 - "Reconciler"
Cohesion: 0.29
Nodes (7): Change, Reconciler, Result, NewReconciler(), TestReconcilerVersionsOnlyRealChanges(), DecodeSnapshot(), TestDecodeSnapshotRefusesUnknownShapes()

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 81 - "FromSnapshot"
Cohesion: 0.36
Nodes (7): FromSnapshot(), snapshot(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes(), TestToSnapshotRefusesHostileDocuments(), TestToSnapshotRoundTrips(), counter

### Community 82 - ".show"
Cohesion: 0.36
Nodes (3): ui, writeHTML(), page

### Community 83 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 85 - "NewReader"
Cohesion: 0.18
Nodes (12): TestEventsDeliversConfigEvents(), NewReader(), TestSSEReaderBoundsLineLength(), TestSSEReaderJoinsDataLinesAndDefaultsName(), TestSSEReaderReportsTruncatedEvent(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment() (+4 more)

### Community 87 - "fetch"
Cohesion: 0.16
Nodes (17): NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall(), extract() (+9 more)

### Community 88 - "errorf"
Cohesion: 0.29
Nodes (7): MeshChange, RecordRequest, RuleRequest, errorf(), invalid(), Nodes, meshChange()

### Community 89 - "6. Control plane"
Cohesion: 0.40
Nodes (5): 6.1 Enrollment, 6.2 Request authentication, 6.3 Config sync, 6.4 Safe apply, 6. Control plane

### Community 90 - "8. Server"
Cohesion: 0.40
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 93 - "RenderNASCompose"
Cohesion: 0.18
Nodes (19): checkPosture(), keys(), mustRead(), sameSet(), service(), TestDevComposePosture(), TestServerComposePosture(), ComposeServer() (+11 more)

### Community 94 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 95 - "net/http.ResponseWriter"
Cohesion: 0.19
Nodes (8): decode(), reply(), ETag(), TestETagQuotesTheHash(), api, decodeJSON(), readBody(), writeJSON()

### Community 96 - "Hub"
Cohesion: 0.18
Nodes (7): Hub, subscription, ConfigEvent, NewHub(), closed(), TestHubLatestWins(), TestHubReplacesKicksAndCloses()

### Community 97 - "uapiChange"
Cohesion: 0.40
Nodes (4): pfxs(), TestUAPISetGolden(), uapiChange(), UAPISet()

### Community 98 - "userspace"
Cohesion: 0.19
Nodes (4): defaultInterface(), exitRoutes(), userspace, unspecified()

### Community 99 - "route"
Cohesion: 0.36
Nodes (3): index(), route(), netlinkTable

### Community 100 - "meshctl"
Cohesion: 0.40
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 101 - "newQR"
Cohesion: 0.67
Nodes (3): newQR(), phoneData, qrCode

### Community 102 - "Open"
Cohesion: 0.15
Nodes (13): Document, Import(), document(), linuxOnly(), TestImportRefusesInvalidDocumentsUntouched(), TestImportRefusesWhileTheServerRuns(), TestImportRestoresAnExport(), TestLockIsExclusive() (+5 more)

### Community 103 - "Allocate"
Cohesion: 0.25
Nodes (9): Allocate(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan(), TestCheckMeshCIDR(), TestRangeFor() (+1 more)

### Community 104 - "PhoneConfig"
Cohesion: 0.31
Nodes (9): NewKey(), Render(), aliceKey(), golden(), TestHub(), TestNewKey(), TestProfiles(), TestRenderRefuses() (+1 more)

### Community 107 - "agentfake.go"
Cohesion: 0.16
Nodes (6): entry, TestDNS(), TestPolicyHasKernelSemantics(), TestRegistrarNeedsTheInterface(), NewPolicy(), NewPolicy()

### Community 109 - "go_pkg_net_netip"
Cohesion: 0.17
Nodes (4): linkDNS, AddrPort(), TestAddrPort(), NewKernel()

### Community 110 - "nullInt"
Cohesion: 0.16
Nodes (6): Tx, Tx, nullInt(), AuditEntry, DNSRecord, FirewallRule

### Community 111 - "Engine"
Cohesion: 0.18
Nodes (6): Deps, DNS, Engine, Firewall, Result, Sysctls

### Community 118 - "agent/agent.go"
Cohesion: 0.16
Nodes (8): TestDataDirIsProtectedOnWindows(), TestCreateLinkNeedsPrivilege(), TestOpen(), Open(), NewKernel(), TestOpenRefusesUnknownBackends(), TestUserspaceWithoutADevice(), ExitRouter

## Knowledge Gaps
- **158 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `Interface`, `Exit` (+153 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 366 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **12 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NodeConfig` connect `NodeConfig` to `Hub`, `runner`, `NodeInfo`, `agent_test.go`, `ui/ui_test.go`, `.Phone`, `meshHarness`, `PhoneConfig`, `apply.go`, `Key`, `Engine`, `ServerConfig`, `Dir`, `Client`, `WGKey`, `time.Time`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Why does `Host` connect `Host` to `agent_test.go`, `agentfake.go`, `Key`, `WGKey`, `time.Time`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Dir` connect `Dir` to `runner`, `ui/ui_test.go`, `agent_test.go`, `testing.T`, `NodeConfig`, `LocalStatus`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _158 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08611575777852125 - nodes in this community are weakly interconnected._
- **Should `Name` be split into smaller, more focused modules?**
  _Cohesion score 0.13970588235294118 - nodes in this community are weakly interconnected._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.06452072354365475 - nodes in this community are weakly interconnected._