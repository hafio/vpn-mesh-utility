# Graph Report - vpn-mesh-utility  (2026-10-30)

## Corpus Check
- 336 files · ~418,342 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 18 file(s) not represented in the graph (top: .tmpl 10, (none) 5, .css 2)

## Summary
- 4122 nodes · 22225 edges · 145 communities (112 shown, 33 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2607 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `59cd15fb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- net/netip.AddrPort
- IPNet
- htmx.min.js
- env
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- TestDefaultSocket
- serve
- github.com/hafio/vpn-mesh-utility
- .queryRow
- fakeServer
- api
- serveRuntime
- New
- Dir
- Agent (`internal/agent/...`)
- Run
- net/netip.Prefix
- Net
- meshHarness
- Input
- lab
- Snapshot
- env
- Command reference
- .mesh
- NodeConfig
- context.Context
- newRig
- Service
- agent/agent_test.go
- testing.T
- `internal/server/auth`
- testing.T
- syncBuffer
- .exec
- `internal/server/store`
- nullInt
- netlinkNet
- Net
- planner_test.go
- Node
- Mesh
- README.md
- .AddMesh
- .query
- Test catalogue
- Session
- LocalStatus
- `internal/server/api`
- Open
- otherNet
- Client
- Container
- Enum
- meshDNS
- `internal/validate`
- Sync
- .Apply
- Run
- AddrRange
- .meshctl
- UI
- .add
- `internal/proto`
- log/slog.Logger
- fetchwintun/main_test.go
- .as
- PublishNames
- net/http.Request
- Exporter
- TestPageGoldens
- Sync
- io.Writer
- net/netip.Addr
- resolved
- errorf
- rig
- io.Reader
- .renderPolicies
- deploy_test.go
- ServerURL
- io.Writer
- Host
- exprs
- Ruleset
- Host
- .readJSON
- userspace
- ns
- context.Context
- .Run
- run
- Extending the code
- Schedule
- time.Time
- context.Context
- contents_test.go
- Read
- nftKernel
- StepError
- Diff
- NewHub
- Engine
- ReachResult
- AuditEntry
- loadMigrations
- sync.Mutex
- wgdev
- Name
- forms.js
- `internal/server` and `internal/server/export`
- PeerStat
- countingDB
- AgentSession
- NodeInfo
- TestMain
- .SetSegmentation
- entry
- AdminError
- Slot
- Error
- NewNet
- NewNet
- service_windows.go
- NewKernel
- Refusal
- NewKernel
- Refusal
- usageError
- StatusError
- AliveRequest
- TestQRCode
- Error

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 304 edges
2. `ui` - 110 edges
3. ``internal/server/api`` - 101 edges
4. `ServerConfig()` - 89 edges
5. `Session` - 85 edges
6. `Slot` - 82 edges
7. `newHarness()` - 73 edges
8. `NodeConfig` - 69 edges
9. `env` - 64 edges
10. `api` - 58 edges

## Surprising Connections (you probably didn't know these)
- `A node's agent as a Swarm service` --references--> `checkStackSample()`  [INFERRED]
  docs/security.md → deploy/samples_test.go
- `Dev script tasks` --references--> `build()`  [INFERRED]
  docs/developer.md → internal/agent/policy/nft_linux.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the control plane (Dokploy)` --references--> `v()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js

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
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/policy_test.go -> internal/testfix/trust.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/session_test.go -> internal/testfix/trust.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/sign_test.go -> internal/testfix/trust.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/trust_test.go -> internal/testfix/trust.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/sse_test.go -> internal/proto/trust_test.go -> internal/testfix/trust.go -> internal/proto/sse_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/policy_test.go -> internal/testfix/testfix_test.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/session_test.go -> internal/testfix/testfix_test.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/sign_test.go -> internal/testfix/testfix_test.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/trust_test.go -> internal/testfix/testfix_test.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/sse_test.go -> internal/proto/trust_test.go -> internal/testfix/testfix_test.go -> internal/proto/sse_test.go`

## Communities (145 total, 33 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.09
Nodes (58): rig, TestProxyLoopAppliesAgain(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), IsStepError(), New(), failed() (+50 more)

### Community 2 - "net/netip.AddrPort"
Cohesion: 0.19
Nodes (20): check(), newNode(), validateSet(), Endpoint6(), IPv6(), PeerEndpoint(), plain6(), Prefix() (+12 more)

### Community 3 - "IPNet"
Cohesion: 0.10
Nodes (10): Status, One image, two modes, IPNet(), Prefix(), index(), route(), lookup(), netlinkTable (+2 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.15
Nodes (26): checkMachineIDFile(), joinFlags(), meshDir(), meshFlags(), onlyMesh(), sep(), agentRoot(), newLogger() (+18 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem, 2. Goals (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.14
Nodes (17): activeAt(), zoneOf(), checkSegmentation(), CompilePolicy(), hosts(), newSegResolver(), Overlaps(), parseSpans() (+9 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.13
Nodes (17): Rule, Violation, Determinism and goldens, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates (+9 more)

### Community 11 - "TestDefaultSocket"
Cohesion: 0.27
Nodes (10): Pipe(), TestDefaultSocket(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe(), waitPipe() (+2 more)

### Community 12 - "serve"
Cohesion: 0.16
Nodes (31): harness, joinBody(), postJoin(), serve(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+23 more)

### Community 14 - ".queryRow"
Cohesion: 0.11
Nodes (4): HashToken(), NewToken(), Machine, Refusal

### Community 15 - "fakeServer"
Cohesion: 0.13
Nodes (12): fakeServer, signed(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Definitive() (+4 more)

### Community 16 - "api"
Cohesion: 0.11
Nodes (14): api, decode(), reply(), cli(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+6 more)

### Community 17 - "serveRuntime"
Cohesion: 0.20
Nodes (14): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+6 more)

### Community 18 - "New"
Cohesion: 0.36
Nodes (4): New(), sameOrigin(), Deps, handler

### Community 19 - "Dir"
Cohesion: 0.06
Nodes (40): `internal/fsutil`, meshKey(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, Enrollment, NewKeys(), newSeed() (+32 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.19
Nodes (30): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), removable() (+22 more)

### Community 21 - "Run"
Cohesion: 0.09
Nodes (31): Options, TestStartNodeRuns(), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), Join(), New(), TestJoinIsUnsigned() (+23 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (31): Routes, NewTable(), compare(), Table, Sync(), NewTable(), luidOf(), NewTable() (+23 more)

### Community 24 - "meshHarness"
Cohesion: 0.15
Nodes (21): TestIPv6DNSRecords(), harness, TestMeshesStayApart(), TestNodesServiceInAMesh(), TestSetSettings(), harness, meshHarness(), refused() (+13 more)

### Community 25 - "Input"
Cohesion: 0.12
Nodes (13): PlanError, Record, Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable(), Problem (+5 more)

### Community 26 - "lab"
Cohesion: 0.19
Nodes (7): Coverage, eventually(), lab, stream, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.13
Nodes (20): `internal/server/versions`, LegacySegmentation(), SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation() (+12 more)

### Community 28 - "env"
Cohesion: 0.16
Nodes (48): machineIDFlag(), setupAgentJoin(), setupAgentReset(), setupAgentRun(), socketFlags(), setupVersion(), wantArgs(), setupDNS() (+40 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.08
Nodes (30): Agent protocol, Container names, Contents, Errors, Join, Local API, Mesh view, Prefs (+22 more)

### Community 31 - "NodeConfig"
Cohesion: 0.07
Nodes (16): Deps, dialing, IPv6Host, Rechecker, Add an apply step, forHost(), DNS, Engine (+8 more)

### Community 32 - "context.Context"
Cohesion: 0.09
Nodes (13): Deps, runner, statusDetail, kick(), oneLine(), Reset(), Run(), Result (+5 more)

### Community 33 - "newRig"
Cohesion: 0.23
Nodes (25): `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages(), TestPhonePage(), newRig() (+17 more)

### Community 34 - "Service"
Cohesion: 0.08
Nodes (19): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), bindBoth() (+11 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (72): rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime(), newFakeServer() (+64 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (134): Agent (`internal/agent/...`), freePort(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), pfx(), TestConfigureHasKernelSemantics(), TestDNS() (+126 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (30): rig, 11. CLI reference, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants() (+22 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (55): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+47 more)

### Community 40 - ".exec"
Cohesion: 0.11
Nodes (5): mustAffect(), boolInt(), classify(), changed(), Session

### Community 41 - "`internal/server/store`"
Cohesion: 0.13
Nodes (48): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+40 more)

### Community 42 - "nullInt"
Cohesion: 0.12
Nodes (4): DNSRecord, nullInt(), nullStr(), Version

### Community 43 - "netlinkNet"
Cohesion: 0.16
Nodes (7): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), toRule()

### Community 44 - "Net"
Cohesion: 0.14
Nodes (6): ensureAddr(), Net, Rule, syncDropped(), syncProxies(), syncRules()

### Community 45 - "planner_test.go"
Cohesion: 0.15
Nodes (41): Add a topology rule, `internal/server/planner`, TestRangeFor(), Compute(), deliver(), deliverAddr(), encode(), firstDiff() (+33 more)

### Community 46 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 47 - "Mesh"
Cohesion: 0.20
Nodes (6): freeSlot(), Mesh, newUUID(), scanMesh(), scanNode(), scanner

### Community 48 - "README.md"
Cohesion: 0.19
Nodes (7): Building, Contents, Documentation, How it fits together, meshctl, Quick start, Status

### Community 49 - ".AddMesh"
Cohesion: 0.14
Nodes (15): MeshList, MeshSetRequest, meshSpace, MeshUpdate, SettingsRequest, checkApart(), clash(), freeCIDR() (+7 more)

### Community 50 - ".query"
Cohesion: 0.21
Nodes (3): ContainerName, scanAll(), splitList()

### Community 51 - "Test catalogue"
Cohesion: 0.16
Nodes (19): Contents, `internal/server/phone`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue, Hub(), joinSet(), NewKey() (+11 more)

### Community 52 - "Session"
Cohesion: 0.11
Nodes (32): ticketServer, testSession(), newTicketServer(), sessionClient(), TestBye(), TestChallengesAreAnswered(), TestSessionLostAndHello(), TestTicketsRoll() (+24 more)

### Community 53 - "LocalStatus"
Cohesion: 0.09
Nodes (22): Client, Logs, Handler(), logs(), refusal(), ParseAddr(), laptop(), PrefsRequest (+14 more)

### Community 54 - "`internal/server/api`"
Cohesion: 0.09
Nodes (47): apiError, `internal/server/api`, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode() (+39 more)

### Community 55 - "Open"
Cohesion: 0.14
Nodes (17): emptyService(), TestFirstAdmin(), TestFirstAdminAfterTheCLI(), TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), insertAt5() (+9 more)

### Community 57 - "Client"
Cohesion: 0.14
Nodes (8): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, ValidChallenge(), DroppedName

### Community 58 - "Container"
Cohesion: 0.18
Nodes (5): fakeRuntime, Container, clip(), meshFilter(), Runtime

### Community 59 - "Enum"
Cohesion: 0.18
Nodes (9): Ref(), ServiceEntry(), TestRef(), TestServiceEntry(), SemVer(), TestText(), TestVersionAndSemVer(), Text() (+1 more)

### Community 60 - "meshDNS"
Cohesion: 0.19
Nodes (25): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+17 more)

### Community 61 - "`internal/validate`"
Cohesion: 0.09
Nodes (24): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), mask(), POSIXPath() (+16 more)

### Community 62 - "Sync"
Cohesion: 0.15
Nodes (33): NewNet(), TestTeardownFakes(), TestForHostWithoutIPv6(), Gateway(), Guard(), Owned(), parse(), Remove() (+25 more)

### Community 63 - ".Apply"
Cohesion: 0.12
Nodes (15): applied, Firewall, policy, Firewall, Runner, atLeast(), exitPolicy(), normalize() (+7 more)

### Community 64 - "Run"
Cohesion: 0.12
Nodes (22): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+14 more)

### Community 65 - "AddrRange"
Cohesion: 0.08
Nodes (29): addrs(), addrText(), contains(), Rule, merged(), ofFamily(), portText(), services() (+21 more)

### Community 66 - ".meshctl"
Cohesion: 0.17
Nodes (15): 8. The web UI, Each node's local UI, The mesh firewall, A node's agent as a Swarm service, Agent service (systemd), Changes a node asks for, Contents, DNS (+7 more)

### Community 67 - "UI"
Cohesion: 0.09
Nodes (19): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize() (+11 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.06
Nodes (53): `internal/proto`, NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), compile() (+45 more)

### Community 70 - "log/slog.Logger"
Cohesion: 0.09
Nodes (13): Change, Hub, subscription, blockingHost, NewUserspace(), NewUserspace(), NewUserspace(), ConfigEvent (+5 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.20
Nodes (13): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), TestDownload(), TestExtractChecksTheDLL() (+5 more)

### Community 72 - ".as"
Cohesion: 0.21
Nodes (11): brokenStream, ErrorBody, harness, TestDuplicateIsRefused(), TestForeignAndPinning(), TestHelloIsGranted(), TestSessionLapses(), TestStaleAnswerAsksAgain() (+3 more)

### Community 73 - "PublishNames"
Cohesion: 0.14
Nodes (15): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+7 more)

### Community 74 - "net/http.Request"
Cohesion: 0.08
Nodes (18): TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), listOf(), meshChanged(), group() (+10 more)

### Community 75 - "Exporter"
Cohesion: 0.38
Nodes (4): NextExport(), TestExporterRunStopsWithItsContext(), TestNextExport(), Exporter

### Community 76 - "TestPageGoldens"
Cohesion: 0.17
Nodes (13): Add a web UI page, The web UI's pages (internal/server/ui), tabsIn(), dnsState(), newRenderer(), golden(), TestEveryPageRenders(), TestFormsShowWhatApplies() (+5 more)

### Community 77 - "Sync"
Cohesion: 0.07
Nodes (13): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Policy, Route, Rule (+5 more)

### Community 78 - "io.Writer"
Cohesion: 0.16
Nodes (16): TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), printExit(), printMesh(), joinNonEmpty(), nodeState(), printChange() (+8 more)

### Community 79 - "net/netip.Addr"
Cohesion: 0.07
Nodes (28): network, hostAddrs, nameTable, zone, TestIsPublic(), proxied(), gateway(), host() (+20 more)

### Community 80 - "resolved"
Cohesion: 0.13
Nodes (14): Domain, fakeRegistrar, linkDomain, nrpt, Registration, resolved, parseServers(), serverList() (+6 more)

### Community 81 - "errorf"
Cohesion: 0.06
Nodes (57): PhoneRequest, segLiteral, segDoc, Role, blockOf(), checkName(), containerNetwork(), containerSubnet6() (+49 more)

### Community 82 - "rig"
Cohesion: 0.19
Nodes (9): code(), cookieOf(), newEmptyRig(), TestEnrolmentQRCodeIsAPNG(), TestFirstAdminFromTheWeb(), TestSignInFlow(), TestWrongEnrolmentCodeKeepsTheSecret(), faultyDB (+1 more)

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - ".renderPolicies"
Cohesion: 0.10
Nodes (24): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+16 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.06
Nodes (62): block, checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth() (+54 more)

### Community 86 - "ServerURL"
Cohesion: 0.18
Nodes (9): Container-mode agent, HelloRequest, CheckServerDomains(), TestCheckServerDomains(), Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase() (+1 more)

### Community 87 - "io.Writer"
Cohesion: 0.18
Nodes (9): agentLog(), main(), ttyPassword(), OpenFile(), lockFile(), unlockFile(), lockFile(), unlockFile() (+1 more)

### Community 90 - "Host"
Cohesion: 0.10
Nodes (18): link, exitHost, Host, peer(), hexKey(), ParseUAPI(), uapiInt(), compareKeys() (+10 more)

### Community 93 - "exprs"
Cohesion: 0.21
Nodes (6): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), nftOp

### Community 94 - "Ruleset"
Cohesion: 0.10
Nodes (10): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel (+2 more)

### Community 95 - "Host"
Cohesion: 0.31
Nodes (3): Host, SyncAddrs(), SyncLink()

### Community 96 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.11
Nodes (4): adapterGUID(), defaultInterface(), Link, userspace

### Community 99 - "ns"
Cohesion: 0.09
Nodes (22): Several meshes on one host, End to end -- `test/e2e`, host, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces() (+14 more)

### Community 100 - "context.Context"
Cohesion: 0.07
Nodes (21): IPv6Request, IPv6View, NodeAddRequest, NodeSetRequest, ReachRequest, SegmentView, ZoneView, resetTOTP() (+13 more)

### Community 101 - ".Run"
Cohesion: 0.24
Nodes (8): TestRunNeedsAnEnrollment(), TestExporterRunExportsAtNight(), TestRunGivesTheMeshIPv6(), TestRunLogsTheSetupCode(), TestRunRefusesBadConfig(), TestRunServesAndShutsDown(), main(), run()

### Community 102 - "run"
Cohesion: 0.12
Nodes (29): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+21 more)

### Community 103 - "Extending the code"
Cohesion: 0.20
Nodes (10): Add a command, Add a database migration, Add a validator, Contents, Dev script tasks, Developer guide, Extending the code, Goldens (+2 more)

### Community 104 - "Schedule"
Cohesion: 0.13
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "time.Time"
Cohesion: 0.08
Nodes (19): pendingApply, safeApply, Failover, hubMove, hubState, replayCache, failures, hitWindow (+11 more)

### Community 106 - "context.Context"
Cohesion: 0.06
Nodes (36): Access, Config, DB, Lockout, Params, Pending, Sealer, SoleMesh() (+28 more)

### Community 107 - "contents_test.go"
Cohesion: 0.57
Nodes (6): docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents()

### Community 108 - "Read"
Cohesion: 0.38
Nodes (7): source, clean(), first(), fromFile(), sources(), Read(), sources()

### Community 109 - "nftKernel"
Cohesion: 0.29
Nodes (4): build(), conn(), nftBatch, nftKernel

### Community 111 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 112 - "NewHub"
Cohesion: 0.40
Nodes (6): NewHub(), closed(), TestHubDeliversChallenges(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 113 - "Engine"
Cohesion: 0.15
Nodes (6): nameServer, DNS, Want, MeshView, MeshName, MeshNode

### Community 114 - "ReachResult"
Cohesion: 0.50
Nodes (4): ReachResult, matrixCell, matrixRow, reachView

### Community 116 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 117 - "sync.Mutex"
Cohesion: 0.14
Nodes (6): syncBuffer, Sysctls, Sysctl, Sysctls, Sysctl, syncBuffer

### Community 119 - "Name"
Cohesion: 0.19
Nodes (13): RecordRequest, validateMeshNames(), checkGroups(), ToSnapshot(), checkDNS(), IP(), DNSName(), Domains() (+5 more)

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.14
Nodes (25): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+17 more)

### Community 125 - "NodeInfo"
Cohesion: 0.10
Nodes (24): CopyInfo, NodeDetail, NodeList, PhoneProfile, addresses(), heard(), ContainerNetwork, NodeInfo (+16 more)

### Community 130 - "Slot"
Cohesion: 0.07
Nodes (34): nrptRule, uiListen(), TestPolicyHasKernelSemantics(), TestKernelShapes(), desired(), domainName(), TestZoneFollowsTheConfig(), fromRules() (+26 more)

### Community 142 - "TestQRCode"
Cohesion: 0.50
Nodes (4): newQR(), TestQRCode(), phoneData, qrCode

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 395 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **33 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Command reference` connect `Command reference` to `README.md`, `.meshctl`, `run`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09551235336984884 - nodes in this community are weakly interconnected._
- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `Slot`, `archtest.go`, `TestDefaultSocket`, `serveRuntime`, `Dir`, `Agent (`internal/agent/...`)`, `Run`, `.mesh`, `NodeConfig`, `agent/agent_test.go`, `Node`, `Test catalogue`, `Session`, `meshDNS`, `Sync`, `Run`, `AddrRange`, `.meshctl`, `UI`, ``internal/proto``, `net/netip.Addr`, `resolved`, `Host`, `exprs`, `.Run`, `time.Time`, `Engine`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.0912568306010929 - nodes in this community are weakly interconnected._
- **Why does `TestCommandsDoc()` connect `run` to `testing.T`, `Command reference`, `testing.T`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._