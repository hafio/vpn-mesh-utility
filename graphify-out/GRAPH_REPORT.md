# Graph Report - vpn-mesh-utility  (2026-10-20)

## Corpus Check
- 213 files · ~199,262 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 2626 nodes · 10941 edges · 108 communities (91 shown, 17 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 601 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7b2b7d87`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- agent_test.go
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
- TestServiceAnswersAndRegisters
- Tx
- github.com/hafio/vpn-mesh-utility
- newHarness
- Registration
- Run
- RenderInstall
- Hub
- Dir
- IPNet
- WGKey
- net/netip.Prefix
- net/netip.Addr
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
- joinedTo
- testing.T
- Test catalogue
- IdentityKey
- Policy
- Tx
- resolved
- Agent protocol
- .reply
- newRig
- Snapshot
- DNS
- ServerConfig
- CLAUDE.md
- Security
- Host
- Service
- Client
- mustAffect
- Node
- net/http.ResponseWriter
- Sealer
- NodeChange
- Table
- Service
- log/slog.Logger
- Ring
- execrun_test.go
- wgdev
- UI
- .Firewall
- session
- loadMigrations
- validate_test.go
- nullInt
- otherPolicy
- validate.go
- StepError
- Sysctl
- .phoneProfile
- net/netip.AddrPort
- route
- fakeService
- 7. Agent
- FromSnapshot
- Diff
- .show
- Compute
- .validate
- WriteEvent
- fetchwintun/main.go
- changed
- 6. Control plane
- 8. Server
- .Seen
- RenderNASCompose
- errorf
- net/http.Request
- NewNodes
- userspace
- meshctl
- AdminError
- PhoneConfig
- Tx
- Allocate
- Node
- NodeConfig
- .install
- TestMain
- Error
- NewPolicy
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

## Communities (108 total, 17 thin omitted)

### Community 0 - "agent_test.go"
Cohesion: 0.06
Nodes (53): entry, nasFile, nasHealth, nasService, linkDNS, DNSRecord, FirewallRule, Node (+45 more)

### Community 1 - "newRig"
Cohesion: 0.18
Nodes (14): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+6 more)

### Community 2 - "Open"
Cohesion: 0.12
Nodes (34): NewHub(), closed(), TestHubLatestWins(), TestHubReplacesKicksAndCloses(), Import(), document(), linuxOnly(), TestImportRefusesInvalidDocumentsUntouched() (+26 more)

### Community 3 - "planner_test.go"
Cohesion: 0.21
Nodes (22): deliver(), encode(), firstDiff(), hasProblem(), lookup(), mustCompute(), node(), TestCheckRejects() (+14 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (118): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+110 more)

### Community 5 - "env"
Cohesion: 0.05
Nodes (86): watchdog, agentCommands(), agentLog(), joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall() (+78 more)

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
Cohesion: 0.26
Nodes (9): Rule, Violation, Check(), importsOf(), ModuleRoot(), TestCheckerCatchesViolations(), TestModuleBoundaries(), under() (+1 more)

### Community 11 - "TestServiceAnswersAndRegisters"
Cohesion: 0.19
Nodes (15): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceBinding() (+7 more)

### Community 14 - "newHarness"
Cohesion: 0.10
Nodes (46): countingDB, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing(), TestAdminSetClearsReferences() (+38 more)

### Community 15 - "Registration"
Cohesion: 0.29
Nodes (6): Domain, fakeRegistrar, Registration, domainName(), same(), TestSameRegistration()

### Community 16 - "Run"
Cohesion: 0.09
Nodes (30): Deps, Route, AdminRoutes(), describe(), api, New(), NewAdmin(), newAPI() (+22 more)

### Community 17 - "RenderInstall"
Cohesion: 0.48
Nodes (7): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams

### Community 18 - "Hub"
Cohesion: 0.31
Nodes (3): Hub, subscription, ConfigEvent

### Community 19 - "Dir"
Cohesion: 0.15
Nodes (7): Dir, Enrollment, syncDir(), TestWriteAtomicReplacesAndLeavesNoTempFiles(), WriteAtomic(), Keys, Local

### Community 20 - "IPNet"
Cohesion: 0.15
Nodes (7): IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), lookup(), kernel, links

### Community 21 - "WGKey"
Cohesion: 0.12
Nodes (27): rig, NewRoutes(), TestRoutesSysctlAndFirewall(), IsStepError(), New(), meshDNS(), newRig(), TestApplyIsIdempotent() (+19 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.20
Nodes (5): Routes, compare(), luidOf(), unsupported, winTable

### Community 23 - "net/netip.Addr"
Cohesion: 0.22
Nodes (9): zone, TestIsPublic(), soa(), isPublic(), PublicIPv4(), inAny(), CheckRequested(), TestCheckRequested() (+1 more)

### Community 24 - ".nodeView"
Cohesion: 0.15
Nodes (11): Health(), phoneHealth(), TestDashboardHealth(), ago(), bytesize(), TestHelpers(), dashboardData, nodeForm (+3 more)

### Community 25 - "LocalStatus"
Cohesion: 0.08
Nodes (22): statusDetail, Client, Logs, logs(), refusal(), laptop(), PrefsRequest, PrefsResponse (+14 more)

### Community 26 - "lab"
Cohesion: 0.16
Nodes (12): host, lab, stream, deleteNamespaces(), eventually(), inNS(), newLab(), ns() (+4 more)

### Community 27 - "service_windows.go"
Cohesion: 0.09
Nodes (13): hostService(), hostService(), uninstallAgent(), Service, Systemd(), commandLine(), RunWindowsService(), stop() (+5 more)

### Community 28 - "Session"
Cohesion: 0.17
Nodes (8): Session, atoi(), before(), ui, stamp(), message(), actor(), settingsData

### Community 29 - "Command reference"
Cohesion: 0.07
Nodes (29): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+21 more)

### Community 30 - "time.Time"
Cohesion: 0.08
Nodes (17): pendingApply, safeApply, syncBuffer, replayCache, failures, hitWindow, Lockout, LockoutError (+9 more)

### Community 31 - "Host"
Cohesion: 0.15
Nodes (3): link, Host, Link

### Community 32 - "context.Context"
Cohesion: 0.09
Nodes (16): Deps, runner, Deps, resetTOTP(), kick(), oneLine(), Run(), TestOneLine() (+8 more)

### Community 33 - "newRig"
Cohesion: 0.18
Nodes (23): brokenStream, TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled(), TestChangesNeedOriginAndCSRF(), TestDashboardFragmentAndExport() (+15 more)

### Community 34 - "Tx"
Cohesion: 0.13
Nodes (14): PhoneRequest, checkName(), defaultMTU(), Actor, Nodes, nodeByName(), nodeInfo(), problemStrings() (+6 more)

### Community 35 - "joinedTo"
Cohesion: 0.10
Nodes (52): fakeServer, rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeServer() (+44 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (115): passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+107 more)

### Community 37 - "Test catalogue"
Cohesion: 0.10
Nodes (20): Agent (`internal/agent/...`), `cmd/meshctl`, Coverage, `deploy`, End to end -- `test/e2e`, `internal/archtest`, `internal/fsutil`, `internal/proto` (+12 more)

### Community 38 - "IdentityKey"
Cohesion: 0.07
Nodes (37): Options, TestCleanHostname(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), Join(), New(), newClient(), TestConfigHonoursTheETag() (+29 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (20): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+12 more)

### Community 40 - "Tx"
Cohesion: 0.18
Nodes (4): boolInt(), Session, Tx, scanUser()

### Community 41 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 42 - "Agent protocol"
Cohesion: 0.17
Nodes (12): Agent protocol, Apply (agent), Config and events, Errors, Join, Local API, Mesh view, Prefs (+4 more)

### Community 43 - ".reply"
Cohesion: 0.29
Nodes (3): cli(), api, api

### Community 44 - "newRig"
Cohesion: 0.19
Nodes (14): rig, newRig(), TestAccountChangesEndSessions(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures(), TestLoginRefusals(), TestLoginUpgradesAnOldHash() (+6 more)

### Community 45 - "Snapshot"
Cohesion: 0.13
Nodes (17): SnapNode, Snapshot, Tx, snapNode(), Compare(), describeDNS(), describeRules(), fields() (+9 more)

### Community 46 - "DNS"
Cohesion: 0.32
Nodes (3): DNS, desired(), Want

### Community 47 - "ServerConfig"
Cohesion: 0.19
Nodes (27): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+19 more)

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
Cohesion: 0.10
Nodes (30): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), fullExpiry() (+22 more)

### Community 53 - "Client"
Cohesion: 0.36
Nodes (4): apiError(), clip(), Client, StatusResponse

### Community 54 - "mustAffect"
Cohesion: 0.16
Nodes (5): mustAffect(), Tx, classify(), PeerStat, StatusUpdate

### Community 55 - "Node"
Cohesion: 0.21
Nodes (12): Peer, Record, Role, plannerNode(), Node, ruleNote(), sortPeers(), sortPrefixes() (+4 more)

### Community 56 - "net/http.ResponseWriter"
Cohesion: 0.17
Nodes (8): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), ui, group()

### Community 57 - "Sealer"
Cohesion: 0.33
Nodes (4): Sealer, aad(), NewSealer(), TestSealerBindsTheSecretToItsUser()

### Community 58 - "NodeChange"
Cohesion: 0.15
Nodes (14): ComposeFile, NodeAddRequest, NodeList, NodeSetRequest, PhoneProfile, heard(), AdminClient, NewAdminClient() (+6 more)

### Community 59 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 60 - "Service"
Cohesion: 0.19
Nodes (6): Registrar, Service, New(), NewRegistrar(), NewRegistrar(), exchange()

### Community 61 - "log/slog.Logger"
Cohesion: 0.08
Nodes (17): Change, TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), DB, Reconciler, Result (+9 more)

### Community 63 - "execrun_test.go"
Cohesion: 0.05
Nodes (35): TestUnitPosture(), SystemdUnit(), capped, Exec, applied, Firewall, policy, Firewall (+27 more)

### Community 64 - "wgdev"
Cohesion: 0.22
Nodes (3): userspace, NewUserspace(), wgdev

### Community 65 - "UI"
Cohesion: 0.14
Nodes (11): UI, hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), loadPages() (+3 more)

### Community 66 - ".Firewall"
Cohesion: 0.33
Nodes (5): CustomRule, ServerFirewall, FirewallView, firewallData, ruleForm

### Community 67 - "session"
Cohesion: 0.34
Nodes (5): bytesize(), UI, prefsFlash(), page, session

### Community 68 - "loadMigrations"
Cohesion: 0.27
Nodes (5): Store, Tx, loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 69 - "validate_test.go"
Cohesion: 0.10
Nodes (19): Arg(), POSIXPath(), FuzzArg(), FuzzName(), TestArg(), TestDNSName(), TestEndpoint(), TestEnum() (+11 more)

### Community 70 - "nullInt"
Cohesion: 0.16
Nodes (6): Tx, Tx, nullInt(), AuditEntry, DNSRecord, FirewallRule

### Community 72 - "validate.go"
Cohesion: 0.12
Nodes (33): check(), DNS, Firewall, Interface, NodeConfig, newNode(), validateSet(), ToSnapshot() (+25 more)

### Community 74 - "Sysctl"
Cohesion: 0.33
Nodes (4): Sysctl, New(), Sysctl, Writer

### Community 75 - ".phoneProfile"
Cohesion: 0.33
Nodes (5): ui, newQR(), TestQRCode(), phoneData, qrCode

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.11
Nodes (15): resolver, bindBoth(), bindPair(), serve(), Listen(), Listen(), listenSocket(), ParseAddr() (+7 more)

### Community 77 - "route"
Cohesion: 0.36
Nodes (3): index(), route(), netlinkTable

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "FromSnapshot"
Cohesion: 0.36
Nodes (7): FromSnapshot(), snapshot(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes(), TestToSnapshotRefusesHostileDocuments(), TestToSnapshotRoundTrips(), counter

### Community 81 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 82 - ".show"
Cohesion: 0.36
Nodes (3): ui, writeHTML(), page

### Community 83 - "Compute"
Cohesion: 0.20
Nodes (8): PlanError, Check(), checkDNS(), Problem, sortedByID(), Compute(), Input, newPlan()

### Community 85 - "WriteEvent"
Cohesion: 0.26
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 87 - "fetchwintun/main.go"
Cohesion: 0.14
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

### Community 92 - "RenderNASCompose"
Cohesion: 0.18
Nodes (19): checkPosture(), keys(), mustRead(), sameSet(), service(), TestDevComposePosture(), TestServerComposePosture(), ComposeServer() (+11 more)

### Community 94 - "errorf"
Cohesion: 0.21
Nodes (11): DNSEntry, MeshChange, RecordRequest, RuleRequest, errorf(), invalid(), DNSView, Nodes (+3 more)

### Community 95 - "net/http.Request"
Cohesion: 0.17
Nodes (9): ETag(), TestETagQuotesTheHash(), api, decodeJSON(), readBody(), writeJSON(), ClientIP(), TestClientIP() (+1 more)

### Community 96 - "NewNodes"
Cohesion: 0.16
Nodes (12): apiError, ErrorMessage(), NewNodes(), TestAdminCompose(), toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion(), harness (+4 more)

### Community 100 - "meshctl"
Cohesion: 0.40
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 104 - "PhoneConfig"
Cohesion: 0.24
Nodes (11): Hub(), NewKey(), Render(), aliceKey(), golden(), TestHub(), TestNewKey(), TestProfiles() (+3 more)

### Community 107 - "Allocate"
Cohesion: 0.25
Nodes (9): Allocate(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan(), TestCheckMeshCIDR(), TestRangeFor() (+1 more)

### Community 110 - "Node"
Cohesion: 0.27
Nodes (6): reachable(), Node, Tx, scanNode(), nullStr(), scanner

### Community 111 - "NodeConfig"
Cohesion: 0.12
Nodes (7): NodeDetail, Rechecker, Engine, Result, Desired(), PeerStatus, NodeConfig

## Knowledge Gaps
- **158 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData`, `Interface`, `Exit` (+153 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 360 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **17 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NodeConfig` connect `NodeConfig` to `context.Context`, `NewNodes`, `agent_test.go`, `joinedTo`, `Tx`, `planner_test.go`, `PhoneConfig`, `ServerConfig`, `Hub`, `Compute`, `Dir`, `Client`, `WGKey`, `Node`, `time.Time`?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **Why does `Host` connect `Host` to `agent_test.go`, `joinedTo`, `testing.T`, `WGKey`, `time.Time`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Dir` connect `Dir` to `context.Context`, `agent_test.go`, `joinedTo`, `testing.T`, `IdentityKey`, `NodeConfig`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **What connects `github.com/hafio/vpn-mesh-utility`, `linkDNS`, `logsData` to the rest of the system?**
  _158 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `agent_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.061480642318965675 - nodes in this community are weakly interconnected._
- **Should `Open` be split into smaller, more focused modules?**
  _Cohesion score 0.11711711711711711 - nodes in this community are weakly interconnected._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.06452072354365475 - nodes in this community are weakly interconnected._