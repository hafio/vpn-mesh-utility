# Graph Report - vpn-mesh-utility  (2026-10-25)

## Corpus Check
- 295 files · ~353,251 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3691 nodes · 18390 edges · 137 communities (115 shown, 22 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2210 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `8d080e1c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- env
- `internal/server/api`
- htmx.min.js
- flag.FlagSet
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- net/netip.AddrPort
- meshHarness
- github.com/hafio/vpn-mesh-utility
- Node
- otherNet
- net/http.ResponseWriter
- .do
- commands.go
- Dir
- ServerConfig
- testing.T
- net/netip.Prefix
- Net
- Session
- Join
- lab
- Snapshot
- Container
- Command reference
- time.Time
- harness
- runner
- newRig
- Service
- agent/agent_test.go
- File
- `internal/server/auth`
- `cmd/meshctl`
- Policy
- .exec
- openTest
- samples_test.go
- netlinkNet
- Input
- planner_test.go
- TestPageGoldens
- .meshctl
- README.md
- Host
- Tx
- log/slog.Logger
- .Tx
- LocalStatus
- newRig
- MeshTx
- serveRuntime
- Client
- Net
- Test catalogue
- meshDNS
- sync.Mutex
- Sync
- .Apply
- AddrEntry
- .queryRow
- .nodeView
- UI
- NewHub
- services
- New
- fetchwintun/main.go
- fail
- .AddMesh
- service_test.go
- ServerURL
- nrpt_windows.go
- Error
- kernel
- net/netip.Addr
- resolved
- errorf
- Node
- Container
- .renderPolicies
- deploy_test.go
- stop
- Ruleset
- WriteEvent
- Diff
- toAPIError
- loadMigrations
- TestRunGivesTheMeshIPv6
- userspace
- serve
- Version
- PolicyRule
- exprs
- Exporter
- Schedule
- `internal/proto`
- Service
- .mesh
- AdminError
- wgdev
- StepError
- .SetSegmentation
- TestMain
- Engine
- validate_test.go
- Exit
- entry
- Error
- .add
- NewNet
- same
- `internal/server` and `internal/server/export`
- Refusal
- Table
- TestFailoverMovesRelays
- NewKernel
- time.Time
- NewKernel
- StatusError
- Run
- .BeginEnrollment
- Writer
- ns
- servePipe
- Open
- Name
- io.Reader

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 255 edges
2. `ui` - 100 edges
3. ``internal/server/api`` - 86 edges
4. `Session` - 80 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 67 edges
7. `env` - 62 edges
8. `newHarness()` - 60 edges
9. `runner` - 47 edges
10. `api` - 46 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the control plane (Dokploy)` --references--> `v()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js
- `Add a validator` --references--> `mask()`  [INFERRED]
  docs/developer.md → internal/validate/validate.go

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
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/policy_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/sign_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`

## Communities (137 total, 22 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (16): linkDNS, DNSRecord, FirewallRule, Node, TestDNS(), logsData, inspected, listed (+8 more)

### Community 1 - "NodeConfig"
Cohesion: 0.07
Nodes (52): IPv6Host, rig, Determinism and goldens, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), forHost(), Engine (+44 more)

### Community 2 - "env"
Cohesion: 0.14
Nodes (48): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+40 more)

### Community 3 - "`internal/server/api`"
Cohesion: 0.13
Nodes (28): ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing() (+20 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (114): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+106 more)

### Community 5 - "flag.FlagSet"
Cohesion: 0.12
Nodes (35): TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), printExit(), setupFirewallShow(), setupIPv6(), printMesh(), setupMeshAdd() (+27 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (33): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem, 2. Goals (+25 more)

### Community 7 - "segResolver"
Cohesion: 0.13
Nodes (20): Service, checkSegmentation(), CompilePolicy(), hosts(), mergePorts(), mergeServices(), newSegResolver(), Overlaps() (+12 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.18
Nodes (14): Rule, Violation, Packages, and which way they may import, Input and commands, `internal/archtest`, `internal/tools/vulnjudge`, Test catalogue, Check() (+6 more)

### Community 11 - "net/netip.AddrPort"
Cohesion: 0.11
Nodes (14): dialing, bindBoth(), bindPair(), exchange(), Listen(), Listen(), listenSocket(), Listen() (+6 more)

### Community 12 - "meshHarness"
Cohesion: 0.16
Nodes (29): `internal/server/api`, cli(), ErrorMessage(), TestContainerNetworksWithIPv6(), TestDefaultIPv6(), TestIPv6DNSRecords(), harness, TestMeshAdmin() (+21 more)

### Community 14 - "Node"
Cohesion: 0.20
Nodes (11): Config and events, 4. Roles and topology, Peer, SysctlKeys(), ContainerNetwork(), endpoints(), Node, sortPeers() (+3 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.14
Nodes (10): api, decode(), reply(), ETag(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+2 more)

### Community 17 - ".do"
Cohesion: 0.12
Nodes (11): IPv6Request, IPv6View, MeshSetRequest, SegmentView, SettingsRequest, ZoneView, AdminClient, nodePath() (+3 more)

### Community 18 - "commands.go"
Cohesion: 0.10
Nodes (33): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+25 more)

### Community 19 - "Dir"
Cohesion: 0.10
Nodes (22): `internal/fsutil`, Dir, mkdirs(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestJoinRecord(), TestKeysAreCreatedOnceAndKept(), TestLoadKeysRefusesLooseOrDamagedFiles() (+14 more)

### Community 20 - "ServerConfig"
Cohesion: 0.22
Nodes (24): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+16 more)

### Community 21 - "testing.T"
Cohesion: 0.04
Nodes (124): Agent (`internal/agent/...`), TestCleanHostname(), TestDefaultSocket(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig() (+116 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (27): Routes, syncDropped(), index(), route(), compare(), luidOf(), onLink(), TestULAPrefix() (+19 more)

### Community 24 - "Session"
Cohesion: 0.09
Nodes (13): Session, listOf(), meshChanged(), group(), atoi(), before(), changed(), versionNumber() (+5 more)

### Community 25 - "Join"
Cohesion: 0.17
Nodes (17): Options, TestIsPublic(), TestJoin(), Join(), New(), TestUnreachableAndMalformedAnswers(), cleanHostname(), firstAddr() (+9 more)

### Community 26 - "lab"
Cohesion: 0.18
Nodes (10): Coverage, End to end -- `test/e2e`, host, eventually(), lab, inNS(), tail(), TestServerMesh() (+2 more)

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "Container"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (39): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns, meshctl exit (+31 more)

### Community 30 - "time.Time"
Cohesion: 0.15
Nodes (10): Failover, hubMove, hubState, hitWindow, waitApplied(), TestJudgeHubs(), RateLimiter, NewRateLimiter() (+2 more)

### Community 31 - "harness"
Cohesion: 0.28
Nodes (11): countingDB, harness, joinBody(), postJoin(), TestJoinEnrollsAndConsumesTheToken(), TestJoinIsRateLimitedPerClient(), TestJoinRefusesAPhone(), TestJoinRefusesHostileInputUpFront() (+3 more)

### Community 32 - "runner"
Cohesion: 0.08
Nodes (15): Deps, fakeService, pendingApply, runner, safeApply, 12. Troubleshooting, kick(), oneLine() (+7 more)

### Community 33 - "newRig"
Cohesion: 0.12
Nodes (35): `internal/server/ui`, capped, TestFirewallPages(), TestNodeGroupsOnPages(), TestMeshPages(), newQR(), TestPhonePage(), TestQRCode() (+27 more)

### Community 34 - "Service"
Cohesion: 0.17
Nodes (8): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), serve()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.11
Nodes (60): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+52 more)

### Community 36 - "File"
Cohesion: 0.31
Nodes (4): agentLog(), ttyPassword(), OpenFile(), File

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (28): rig, 11. CLI reference, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword() (+20 more)

### Community 38 - "`cmd/meshctl`"
Cohesion: 0.09
Nodes (53): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+45 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (18): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+10 more)

### Community 40 - ".exec"
Cohesion: 0.08
Nodes (8): Add a database migration, newUUID(), mustAffect(), DNSRecord, boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.13
Nodes (44): `internal/server/store`, addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh(), TestServerSettings(), TestTokenMesh() (+36 more)

### Community 42 - "samples_test.go"
Cohesion: 0.20
Nodes (12): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+4 more)

### Community 43 - "netlinkNet"
Cohesion: 0.14
Nodes (9): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+1 more)

### Community 44 - "Input"
Cohesion: 0.11
Nodes (14): PlanError, Add a topology rule, Record, Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable() (+6 more)

### Community 45 - "planner_test.go"
Cohesion: 0.16
Nodes (39): `internal/server/planner`, TestCheckRequested(), Compute(), deliver(), deliverAddr(), encode(), firstDiff(), hasProblem() (+31 more)

### Community 46 - "TestPageGoldens"
Cohesion: 0.14
Nodes (13): Add a web UI page, The web UI's pages (internal/server/ui), tabsIn(), dnsState(), newRenderer(), golden(), TestEveryPageRenders(), TestPageGoldens() (+5 more)

### Community 47 - ".meshctl"
Cohesion: 0.18
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens (+6 more)

### Community 48 - "README.md"
Cohesion: 0.15
Nodes (9): graphify, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together, meshctl (+1 more)

### Community 49 - "Host"
Cohesion: 0.07
Nodes (35): link, exitHost, Host, peer(), AddrPort(), hexKey(), ParseUAPI(), getAnswer() (+27 more)

### Community 50 - "Tx"
Cohesion: 0.13
Nodes (6): ContainerName, scanAll(), splitList(), AuditEntry, PeerStat, StatusUpdate

### Community 51 - "log/slog.Logger"
Cohesion: 0.13
Nodes (8): Hub, subscription, NewUserspace(), ConfigEvent, Reconciler, NewReconciler(), Config, Options

### Community 52 - ".Tx"
Cohesion: 0.23
Nodes (8): fullExpiry(), load(), session(), hashCookie(), newSecret(), matchStep(), Session, User

### Community 53 - "LocalStatus"
Cohesion: 0.09
Nodes (23): statusDetail, Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest (+15 more)

### Community 54 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 55 - "MeshTx"
Cohesion: 0.10
Nodes (20): NodeDetail, NodeList, NodeSetRequest, PhoneProfile, PhoneRequest, checkName(), Nodes, nodeByName() (+12 more)

### Community 56 - "serveRuntime"
Cohesion: 0.20
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "Net"
Cohesion: 0.18
Nodes (5): ensureAddr(), Net, Rule, syncProxies(), syncRules()

### Community 59 - "Test catalogue"
Cohesion: 0.21
Nodes (16): `internal/server/phone`, Shared fixtures -- `internal/testfix`, Hub(), joinSet(), NewKey(), Render(), aliceKey(), dualPhone() (+8 more)

### Community 60 - "meshDNS"
Cohesion: 0.24
Nodes (22): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceAnswersOverIPv6() (+14 more)

### Community 61 - "sync.Mutex"
Cohesion: 0.17
Nodes (5): syncBuffer, Sysctl, Sysctls, Sysctl, syncBuffer

### Community 62 - "Sync"
Cohesion: 0.22
Nodes (24): Sysctls, NewNet(), Gateway(), Guard(), Owned(), parse(), Sync(), TestGatewayRules() (+16 more)

### Community 63 - ".Apply"
Cohesion: 0.08
Nodes (19): hostService(), SystemdUnit(), applied, Firewall, policy, Firewall, Runner, atLeast() (+11 more)

### Community 64 - "AddrEntry"
Cohesion: 0.20
Nodes (9): AddrEntry(), Ref(), ServiceEntry(), TestAddrEntry(), TestAddrRangePrefix(), TestRef(), TestServiceEntry(), TestText() (+1 more)

### Community 65 - ".queryRow"
Cohesion: 0.11
Nodes (10): SoleMesh(), freeSlot(), PickMesh(), pickMesh(), Mesh, scanMesh(), Tx, HashToken() (+2 more)

### Community 66 - ".nodeView"
Cohesion: 0.14
Nodes (10): Health(), ago(), bytesize(), writeHTML(), TestHelpers(), dashboardData, nodeForm, nodeRow (+2 more)

### Community 67 - "UI"
Cohesion: 0.15
Nodes (10): hostOf(), oldest(), validOrNot(), bytesize(), prefsFlash(), LocalToken(), handler, page (+2 more)

### Community 68 - "NewHub"
Cohesion: 0.50
Nodes (5): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 69 - "services"
Cohesion: 0.32
Nodes (7): addrText(), Rule, merged(), portText(), services(), Range, service

### Community 70 - "New"
Cohesion: 0.15
Nodes (9): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), New(), sameOrigin(), Deps, handler (+1 more)

### Community 71 - "fetchwintun/main.go"
Cohesion: 0.13
Nodes (20): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestInstall(), TestExporterRunExportsAtNight() (+12 more)

### Community 72 - "fail"
Cohesion: 0.20
Nodes (17): check(), newNode(), validateSet(), meshCIDR(), Endpoint6(), IPv6(), PeerEndpoint(), plain6() (+9 more)

### Community 73 - ".AddMesh"
Cohesion: 0.15
Nodes (16): Change, MeshList, meshSpace, MeshUpdate, ULAPrefix(), checkApart(), clash(), freeCIDR() (+8 more)

### Community 74 - "service_test.go"
Cohesion: 0.25
Nodes (8): Exec, allowed(), Check(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere(), TestInstallWritesTheUnitAndStartsIt(), recorder

### Community 75 - "ServerURL"
Cohesion: 0.22
Nodes (8): Container-mode agent, CheckServerDomains(), TestCheckServerDomains(), Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase(), TestServerURL()

### Community 76 - "nrpt_windows.go"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 78 - "kernel"
Cohesion: 0.15
Nodes (5): Status, IPNet(), lookup(), kernel, links

### Community 79 - "net/netip.Addr"
Cohesion: 0.08
Nodes (30): network, hostAddrs, nameTable, zone, Apply (agent), gateway(), host(), records() (+22 more)

### Community 80 - "resolved"
Cohesion: 0.35
Nodes (4): linkDomain, resolved, classify(), foreignNetns()

### Community 81 - "errorf"
Cohesion: 0.08
Nodes (37): NodeAddRequest, RecordRequest, segDoc, Enrollment, Role, addIPv6(), blockOf(), checkGroups() (+29 more)

### Community 82 - "Node"
Cohesion: 0.06
Nodes (37): DNSEntry, NameEntry, NameSet, ReachRequest, segLiteral, containerNetwork(), containerSubnet(), containerSubnet6() (+29 more)

### Community 83 - "Container"
Cohesion: 0.36
Nodes (3): fakeRuntime, Container, ContainerName

### Community 84 - ".renderPolicies"
Cohesion: 0.07
Nodes (24): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+16 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.23
Nodes (21): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+13 more)

### Community 86 - "stop"
Cohesion: 0.12
Nodes (8): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), Windows(), handler, SCM

### Community 87 - "Ruleset"
Cohesion: 0.10
Nodes (10): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel (+2 more)

### Community 90 - "WriteEvent"
Cohesion: 0.46
Nodes (6): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry()

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 94 - "toAPIError"
Cohesion: 0.40
Nodes (4): apiError, toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion()

### Community 95 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 96 - "TestRunGivesTheMeshIPv6"
Cohesion: 0.22
Nodes (5): dial(), dial(), isPipe(), TestRunGivesTheMeshIPv6(), stream

### Community 98 - "userspace"
Cohesion: 0.11
Nodes (5): NewUserspace(), defaultInterface(), exitRoutes(), Link, userspace

### Community 99 - "serve"
Cohesion: 0.17
Nodes (21): brokenStream, serve(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend(), TestEventsStreamPushesChanges(), TestHealthFailsWithoutTheDatabase() (+13 more)

### Community 101 - "PolicyRule"
Cohesion: 0.29
Nodes (5): Matrix, MatrixCell, PolicyRule, Schedule, SegPolicy

### Community 102 - "exprs"
Cohesion: 0.07
Nodes (21): Rechecker, Add a command, Add a validator, Add an apply step, Dev script tasks, Developer guide, Extending the code, Goldens (+13 more)

### Community 103 - "Exporter"
Cohesion: 0.32
Nodes (5): DB, NextExport(), TestExporterRunStopsWithItsContext(), TestNextExport(), Exporter

### Community 104 - "Schedule"
Cohesion: 0.16
Nodes (12): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), TimeZone(), weekday() (+4 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.06
Nodes (55): `internal/proto`, NewPolicy(), NewPolicy(), NewPolicy(), Compile(), Local, compile(), dualPolicy() (+47 more)

### Community 106 - "Service"
Cohesion: 0.15
Nodes (18): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), resetTOTP() (+10 more)

### Community 107 - ".mesh"
Cohesion: 0.10
Nodes (23): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+15 more)

### Community 110 - "StepError"
Cohesion: 0.14
Nodes (4): StepError, NewNet(), NewKernel(), NewKernel()

### Community 113 - "Engine"
Cohesion: 0.19
Nodes (4): nameServer, DNS, Want, MeshName

### Community 114 - "validate_test.go"
Cohesion: 0.10
Nodes (21): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), PipePath(), POSIXPath() (+13 more)

### Community 118 - ".add"
Cohesion: 0.20
Nodes (7): clip(), Ring, Key32(), mask(), TestKey32(), TestTokenMasksTheValue(), Token()

### Community 120 - "same"
Cohesion: 0.20
Nodes (13): Domain, fakeRegistrar, nrptRule, Registration, desired(), domainName(), same(), TestSameRegistration() (+5 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.11
Nodes (28): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+20 more)

### Community 126 - "Table"
Cohesion: 0.21
Nodes (8): Deps, Firewall, MeshFirewall, NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 127 - "TestFailoverMovesRelays"
Cohesion: 0.50
Nodes (3): harness, TestFailoverMovesRelays(), TestFailoverRun()

### Community 129 - "time.Time"
Cohesion: 0.11
Nodes (12): replayCache, failures, Lockout, LockoutError, ago(), serves(), phoneHealth(), record() (+4 more)

### Community 133 - "Run"
Cohesion: 0.11
Nodes (24): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+16 more)

### Community 134 - ".BeginEnrollment"
Cohesion: 0.27
Nodes (5): Sealer, aad(), newTOTPKey(), TestMatchStepAllowsOneStepOfSkew(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 137 - "Writer"
Cohesion: 0.29
Nodes (3): New(), New(), Writer

### Community 138 - "ns"
Cohesion: 0.14
Nodes (12): lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), newLab(), ns(), lab (+4 more)

### Community 139 - "servePipe"
Cohesion: 0.53
Nodes (6): listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe(), waitPipe()

### Community 141 - "Open"
Cohesion: 0.16
Nodes (16): TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies() (+8 more)

### Community 143 - "Name"
Cohesion: 0.24
Nodes (12): validateMeshNames(), ToSnapshot(), checkDNS(), IP(), DNSName(), Domains(), IntRange(), IPv4() (+4 more)

### Community 146 - "io.Reader"
Cohesion: 0.25
Nodes (5): watchdog, Read(), ReadSegmentation(), TestSegmentationFile(), WriteSegmentation()

## Knowledge Gaps
- **7 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `harness`, `NO_COLOR`, `lab` (+2 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 358 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **22 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `on()` connect `htmx.min.js` to ``cmd/meshctl``, `.mesh`, `kernel`?**
  _High betweenness centrality (0.050) - this node is a cross-community bridge._
- **Are the 254 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 254 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `harness` to the rest of the system?**
  _7 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08762291197725389 - nodes in this community are weakly interconnected._
- **Why does ``cmd/meshctl`` connect ``cmd/meshctl`` to `Service`, `htmx.min.js`, `flag.FlagSet`, `archtest.go`, `commands.go`, `Session`?**
  _High betweenness centrality (0.040) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.07033248081841433 - nodes in this community are weakly interconnected._
- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `NodeConfig`, `time.Time`, `Run`, `archtest.go`, `servePipe`, `net/netip.AddrPort`, `Node`, `Dir`, `ServerConfig`, `Join`, `runner`, `agent/agent_test.go`, `Host`, `newRig`, `serveRuntime`, `meshDNS`, `Sync`, `fetchwintun/main.go`, `service_test.go`, `nrpt_windows.go`, `WriteEvent`, `exprs`, ``internal/proto``, `.mesh`, `Engine`, `same`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._