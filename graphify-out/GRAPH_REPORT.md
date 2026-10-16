# Graph Report - vpn-mesh-utility  (2026-10-16)

## Corpus Check
- 190 files · ~188,427 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 7 file(s) not represented in the graph (top: (none) 3, .css 2, .tmpl 1)

## Summary
- 2487 nodes · 10391 edges · 105 communities (89 shown, 16 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 580 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- store_test.go
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
- UI
- testing.T
- Engine
- NodeConfig
- IPNet
- WGKey
- net/netip.Prefix
- NewReader
- .nodeView
- LocalStatus
- lab
- logbuf_test.go
- Session
- Command reference
- time.Time
- Host
- runner
- newRig
- errorf
- agent_test.go
- runArgs
- Test catalogue
- Run
- Policy
- Tx
- dns_test.go
- Agent protocol
- net/http.Request
- newRig
- Snapshot
- Open
- ServerConfig
- CLAUDE.md
- Security
- Host
- Service
- Join
- Node
- log/slog.Logger
- CookieValue
- wg.go
- context.Context
- net/http.ResponseWriter
- Service
- Hub
- logbuf.go
- RenderInstall
- resolved
- Run
- Deps
- session
- Registration
- validate_test.go
- nullInt
- otherPolicy
- validate.go
- StepError
- DNS
- userspace
- resolver
- loadMigrations
- TestParseUAPI
- 7. Agent
- FromSnapshot
- Diff
- page
- Sealer
- .validate
- Error
- Refusal
- .Firewall
- changed
- 6. Control plane
- 8. Server
- .Seen
- net/http.Handler
- time.Duration
- DNSName
- .install
- net/netip.Addr
- route
- .DNS
- .ServeDNS
- meshctl
- AdminError
- Allocate
- mustAffect
- Tx

## God Nodes (most connected - your core abstractions)
1. `ServerConfig()` - 57 edges
2. `Session` - 55 edges
3. `NodeConfig` - 49 edges
4. `newHarness()` - 45 edges
5. `runner` - 39 edges
6. `ui` - 39 edges
7. `env` - 36 edges
8. `Dir` - 31 edges
9. `lab` - 31 edges
10. `TestSpokesAndExit()` - 30 edges

## Surprising Connections (you probably didn't know these)
- `Status` --references--> `on()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/static/htmx.min.js
- ``cmd/meshctl`` --references--> `on()`  [INFERRED]
  docs/test.md → internal/server/ui/static/htmx.min.js
- `1. Run the server (Dokploy)` --references--> `v()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js
- `6. DNS` --references--> `on()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js
- `TestRunRefusesAnUnusableDataDir()` --calls--> `run()`  [INFERRED]
  internal/agent/agent_test.go → cmd/meshctl/main.go

## Import Cycles
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 3-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`

## Communities (105 total, 16 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (27): Error, nasFile, nasHealth, nasService, linkDNS, DNSRecord, FirewallRule, Node (+19 more)

### Community 1 - "newRig"
Cohesion: 0.12
Nodes (16): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+8 more)

### Community 2 - "store_test.go"
Cohesion: 0.24
Nodes (21): insertTopology(), node(), openTest(), TestAudit(), TestDeleteNodeCascades(), TestNodeCRUD(), TestNodeUniquenessIsAConflict(), TestNotFound() (+13 more)

### Community 3 - "planner_test.go"
Cohesion: 0.08
Nodes (43): PlanError, JoinResponse, Peer, Record, Role, plannerNode(), Check(), checkDNS() (+35 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (118): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+110 more)

### Community 5 - "env"
Cohesion: 0.07
Nodes (74): watchdog, agentCommands(), setupAgentJoin(), setupAgentRun(), socketFlags(), argSynopsis(), commands(), defaultAgentData() (+66 more)

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

### Community 14 - "api/api_test.go"
Cohesion: 0.05
Nodes (76): apiError, brokenStream, countingDB, ErrorBody, JoinRequest, CanonicalRequest(), ParseSigned(), Sign() (+68 more)

### Community 15 - ".Apply"
Cohesion: 0.08
Nodes (24): TestUnitPosture(), SystemdUnit(), applied, Firewall, policy, Firewall, Runner, atLeast() (+16 more)

### Community 16 - "UI"
Cohesion: 0.14
Nodes (10): UI, hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), loadPages() (+2 more)

### Community 17 - "testing.T"
Cohesion: 0.04
Nodes (74): checkPosture(), keys(), mustRead(), sameSet(), service(), TestDevComposePosture(), TestServerComposePosture(), ComposeServer() (+66 more)

### Community 18 - "Engine"
Cohesion: 0.12
Nodes (13): NodeDetail, NodeList, PhoneProfile, statusDetail, Rechecker, heard(), Engine, Result (+5 more)

### Community 19 - "NodeConfig"
Cohesion: 0.07
Nodes (19): Dir, Enrollment, syncDir(), TestWriteAtomicReplacesAndLeavesNoTempFiles(), WriteAtomic(), NodeConfig, Hub(), NewKey() (+11 more)

### Community 20 - "IPNet"
Cohesion: 0.13
Nodes (8): IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), lookup(), Link, kernel, links

### Community 21 - "WGKey"
Cohesion: 0.08
Nodes (40): rig, TestRunNeedsAnEnrollment(), NewHost(), NewRoutes(), TestConfigureHasKernelSemantics(), TestLinksAddressesAndFailures(), TestRoutesSysctlAndFirewall(), IsStepError() (+32 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.12
Nodes (14): Routes, pfx(), NewTable(), compare(), Desired(), Table, Sync(), prefixes() (+6 more)

### Community 23 - "NewReader"
Cohesion: 0.18
Nodes (12): TestEventsDeliversConfigEvents(), NewReader(), TestSSEReaderBoundsLineLength(), TestSSEReaderJoinsDataLinesAndDefaultsName(), TestSSEReaderReportsTruncatedEvent(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment() (+4 more)

### Community 24 - ".nodeView"
Cohesion: 0.13
Nodes (12): Health(), phoneHealth(), TestDashboardHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeData (+4 more)

### Community 25 - "LocalStatus"
Cohesion: 0.09
Nodes (19): waitApplied(), Client, refusal(), laptop(), PrefsRequest, PrefsResponse, LocalMesh, LocalStatus (+11 more)

### Community 26 - "lab"
Cohesion: 0.16
Nodes (12): host, lab, stream, deleteNamespaces(), eventually(), inNS(), newLab(), ns() (+4 more)

### Community 27 - "logbuf_test.go"
Cohesion: 0.47
Nodes (5): New(), next(), TestFollow(), TestRingKeepsTheLastLines(), TestSlowFollowerIsDropped()

### Community 28 - "Session"
Cohesion: 0.14
Nodes (9): Session, atoi(), before(), ui, stamp(), versionNumber(), message(), actor() (+1 more)

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (27): Command reference, meshctl agent join, meshctl agent run, meshctl dns, meshctl exit, meshctl export, meshctl healthcheck agent, meshctl healthcheck server (+19 more)

### Community 30 - "time.Time"
Cohesion: 0.06
Nodes (17): pendingApply, safeApply, syncBuffer, replayCache, failures, hitWindow, Lockout, LockoutError (+9 more)

### Community 32 - "runner"
Cohesion: 0.17
Nodes (3): runner, kick(), oneLine()

### Community 33 - "newRig"
Cohesion: 0.19
Nodes (23): TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled(), TestChangesNeedOriginAndCSRF(), TestDashboardFragmentAndExport(), TestDNSAndFirewallPages() (+15 more)

### Community 34 - "errorf"
Cohesion: 0.12
Nodes (22): MeshChange, NodeAddRequest, NodeSetRequest, PhoneRequest, RecordRequest, RuleRequest, checkName(), Actor (+14 more)

### Community 35 - "agent_test.go"
Cohesion: 0.12
Nodes (47): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+39 more)

### Community 36 - "runArgs"
Cohesion: 0.08
Nodes (49): passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentRunNeedsAJoin(), TestExportCommand(), TestHealthcheckServer() (+41 more)

### Community 37 - "Test catalogue"
Cohesion: 0.11
Nodes (19): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+11 more)

### Community 38 - "Run"
Cohesion: 0.12
Nodes (12): Deps, Deps, Run(), Sysctl, DNS, Firewall, Sysctls, Logs (+4 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (20): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+12 more)

### Community 40 - "Tx"
Cohesion: 0.18
Nodes (4): boolInt(), Session, Tx, scanUser()

### Community 41 - "dns_test.go"
Cohesion: 0.28
Nodes (15): domainName(), same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects() (+7 more)

### Community 42 - "Agent protocol"
Cohesion: 0.17
Nodes (12): Agent protocol, Apply (agent), Config and events, Errors, Join, Local API, Mesh view, Prefs (+4 more)

### Community 43 - "net/http.Request"
Cohesion: 0.32
Nodes (3): cli(), api, api

### Community 44 - "newRig"
Cohesion: 0.18
Nodes (15): rig, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures(), TestLoginRefusals() (+7 more)

### Community 45 - "Snapshot"
Cohesion: 0.13
Nodes (17): SnapNode, Snapshot, Tx, snapNode(), Compare(), describeDNS(), describeRules(), fields() (+9 more)

### Community 46 - "Open"
Cohesion: 0.16
Nodes (13): Document, Import(), document(), linuxOnly(), TestImportRefusesInvalidDocumentsUntouched(), TestImportRefusesWhileTheServerRuns(), TestImportRestoresAnExport(), TestLockIsExclusive() (+5 more)

### Community 47 - "ServerConfig"
Cohesion: 0.21
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestApplyIsIdempotent() (+17 more)

### Community 49 - "CLAUDE.md"
Cohesion: 0.25
Nodes (7): Determinism and goldens, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates

### Community 50 - "Security"
Cohesion: 0.13
Nodes (14): Agent service (systemd), Changes a node asks for, DNS, Input and commands, Interfaces the agent did not create, Keys and tokens, NAS container, Security (+6 more)

### Community 51 - "Host"
Cohesion: 0.36
Nodes (3): Host, SyncAddr(), SyncLink()

### Community 52 - "Service"
Cohesion: 0.10
Nodes (29): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), fullExpiry() (+21 more)

### Community 53 - "Join"
Cohesion: 0.15
Nodes (13): JoinOptions, Service, Options, apiError(), clip(), Client, Join(), New() (+5 more)

### Community 54 - "Node"
Cohesion: 0.27
Nodes (6): reachable(), Node, Tx, scanNode(), nullStr(), scanner

### Community 55 - "log/slog.Logger"
Cohesion: 0.15
Nodes (12): Change, TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), Reconciler, Result, NewReconciler() (+4 more)

### Community 56 - "CookieValue"
Cohesion: 0.23
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 57 - "wg.go"
Cohesion: 0.14
Nodes (22): peer(), ParseAddr(), TestParseAddr(), AddrPort(), TestAddrPort(), hexKey(), ParseUAPI(), uapiInt() (+14 more)

### Community 58 - "context.Context"
Cohesion: 0.20
Nodes (8): fakeService, ComposeFile, resetTOTP(), AdminClient, nodePath(), Enrollment, NodeChange, droppedLogs

### Community 59 - "net/http.ResponseWriter"
Cohesion: 0.16
Nodes (11): decode(), reply(), ETag(), TestETagQuotesTheHash(), api, decodeJSON(), readBody(), writeJSON() (+3 more)

### Community 60 - "Service"
Cohesion: 0.25
Nodes (5): Registrar, Service, New(), NewRegistrar(), NewRegistrar()

### Community 61 - "Hub"
Cohesion: 0.31
Nodes (3): Hub, subscription, ConfigEvent

### Community 63 - "RenderInstall"
Cohesion: 0.48
Nodes (7): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams

### Community 64 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 65 - "Run"
Cohesion: 0.12
Nodes (14): NewHub(), closed(), TestHubLatestWins(), TestHubReplacesKicksAndCloses(), LoadOrCreateKey(), TestLoadOrCreateKey(), hostOf(), ListenAdmin() (+6 more)

### Community 66 - "Deps"
Cohesion: 0.18
Nodes (15): Deps, Route, AdminRoutes(), describe(), api, New(), NewAdmin(), newAPI() (+7 more)

### Community 67 - "session"
Cohesion: 0.34
Nodes (5): bytesize(), UI, prefsFlash(), page, session

### Community 68 - "Registration"
Cohesion: 0.39
Nodes (4): Domain, fakeRegistrar, Registration, desired()

### Community 69 - "validate_test.go"
Cohesion: 0.14
Nodes (14): Arg(), FuzzArg(), FuzzName(), TestArg(), TestEndpoint(), TestEnum(), TestHostnameFoldsCase(), TestIPv4() (+6 more)

### Community 70 - "nullInt"
Cohesion: 0.16
Nodes (6): Tx, Tx, nullInt(), AuditEntry, DNSRecord, FirewallRule

### Community 72 - "validate.go"
Cohesion: 0.15
Nodes (26): check(), Firewall, Interface, NodeConfig, newNode(), validateSet(), ToSnapshot(), Endpoint() (+18 more)

### Community 73 - "StepError"
Cohesion: 0.11
Nodes (5): entry, StepError, NewPolicy(), NewKernel(), NewKernel()

### Community 74 - "DNS"
Cohesion: 0.22
Nodes (5): DNS, Want, newZone(), parent(), DNS

### Community 77 - "loadMigrations"
Cohesion: 0.27
Nodes (5): Store, Tx, loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 78 - "TestParseUAPI"
Cohesion: 0.22
Nodes (8): getAnswer(), hexOf(), pfxs(), TestParseUAPI(), TestParseUAPIRejectsMalformedValues(), TestUAPISetGolden(), uapiChange(), UAPISet()

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

### Community 83 - "Sealer"
Cohesion: 0.33
Nodes (4): Sealer, aad(), NewSealer(), TestSealerBindsTheSecretToItsUser()

### Community 87 - ".Firewall"
Cohesion: 0.33
Nodes (5): CustomRule, ServerFirewall, FirewallView, firewallData, ruleForm

### Community 88 - "changed"
Cohesion: 0.21
Nodes (4): ui, ui, changed(), ui

### Community 89 - "6. Control plane"
Cohesion: 0.40
Nodes (5): 6.1 Enrollment, 6.2 Request authentication, 6.3 Config sync, 6.4 Safe apply, 6. Control plane

### Community 90 - "8. Server"
Cohesion: 0.40
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 92 - "net/http.Handler"
Cohesion: 0.16
Nodes (12): dnsState(), newRenderer(), staticHandler(), ui, headers(), New(), sameOrigin(), golden() (+4 more)

### Community 93 - "time.Duration"
Cohesion: 0.40
Nodes (3): Tx, HashToken(), NewToken()

### Community 94 - "DNSName"
Cohesion: 0.50
Nodes (3): DNS, DNSName(), TestDNSName()

### Community 96 - "net/netip.Addr"
Cohesion: 0.31
Nodes (6): zone, TestIsPublic(), soa(), isPublic(), PublicIPv4(), Record

### Community 97 - "route"
Cohesion: 0.36
Nodes (3): index(), route(), netlinkTable

### Community 98 - ".DNS"
Cohesion: 0.40
Nodes (4): DNSEntry, DNSView, dnsData, recordForm

### Community 100 - "meshctl"
Cohesion: 0.40
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 102 - "Allocate"
Cohesion: 0.25
Nodes (9): Allocate(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan(), TestCheckMeshCIDR(), TestRangeFor() (+1 more)

### Community 103 - "mustAffect"
Cohesion: 0.16
Nodes (5): mustAffect(), Tx, classify(), PeerStat, StatusUpdate

## Knowledge Gaps
- **154 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `Interface`, `Exit` (+149 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 347 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **16 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NodeConfig` connect `NodeConfig` to `runner`, `ui/ui_test.go`, `errorf`, `agent_test.go`, `planner_test.go`, `Run`, `api/api_test.go`, `ServerConfig`, `Engine`, `Join`, `net/netip.Prefix`, `WGKey`, `wg.go`, `Hub`, `time.Time`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `runner` connect `runner` to `ui/ui_test.go`, `newRig`, `Run`, `UI`, `Engine`, `NodeConfig`, `Join`, `log/slog.Logger`, `LocalStatus`, `time.Time`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `Host` connect `Host` to `ui/ui_test.go`, `agent_test.go`, `WGKey`, `wg.go`, `time.Time`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _154 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07956812901462348 - nodes in this community are weakly interconnected._
- **Should `newRig` be split into smaller, more focused modules?**
  _Cohesion score 0.1164021164021164 - nodes in this community are weakly interconnected._
- **Should `planner_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07540983606557378 - nodes in this community are weakly interconnected._