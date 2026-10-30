# Graph Report - vpn-mesh-utility  (2026-10-30)

## Corpus Check
- 337 files · ~419,690 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 18 file(s) not represented in the graph (top: .tmpl 10, (none) 5, .css 2)

## Summary
- 4130 nodes · 22242 edges · 152 communities (119 shown, 33 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2611 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `00e8ff94`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- net/netip.AddrPort
- Policy
- htmx.min.js
- env
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- deploy_test.go
- serve
- github.com/hafio/vpn-mesh-utility
- .queryRow
- fakeServer
- api
- serveRuntime
- netlinkPolicy
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
- key
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
- serve
- PublishNames
- net/http.Request
- CompilePolicy
- Containers
- Sync
- printStatus
- net/netip.Addr
- resolved
- errorf
- rig
- io.Reader
- .renderPolicies
- samples_test.go
- HelloRequest
- io.Writer
- Host
- exprs
- Ruleset
- Host
- .readJSON
- userspace
- ns
- .do
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
- net.Listener
- AuditEntry
- loadMigrations
- ServiceEntry
- rulesFor
- run
- forms.js
- `internal/server` and `internal/server/export`
- .Set
- Agent protocol
- otherPolicy
- NodeInfo
- TestMain
- .SetSegmentation
- entry
- AdminError
- Slot
- Error
- NewNet
- NewNet
- handler
- NewKernel
- Refusal
- NewKernel
- Refusal
- usageError
- StatusError
- AliveRequest
- TestQRCode
- Registration
- Error
- service_windows_test.go
- listText
- stream
- github.com/miekg/dns.Msg
- copy
- Version

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

## Communities (152 total, 33 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.12
Nodes (47): TestProxyLoopAppliesAgain(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), IsStepError(), New(), failed(), isStep() (+39 more)

### Community 2 - "net/netip.AddrPort"
Cohesion: 0.16
Nodes (29): check(), ParseAddr(), newNode(), validateSet(), ToSnapshot(), Endpoint6(), IP(), IPv6() (+21 more)

### Community 3 - "Policy"
Cohesion: 0.14
Nodes (27): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+19 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.14
Nodes (28): checkMachineIDFile(), joinFlags(), machineIDFlag(), meshDir(), meshFlags(), onlyMesh(), sep(), setupAgentRun() (+20 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (35): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+27 more)

### Community 7 - "segResolver"
Cohesion: 0.21
Nodes (10): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), parseSpans(), portRange, segResolver (+2 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.39
Nodes (6): Rule, Violation, Check(), importsOf(), under(), underAny()

### Community 11 - "deploy_test.go"
Cohesion: 0.19
Nodes (24): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+16 more)

### Community 12 - "serve"
Cohesion: 0.17
Nodes (23): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+15 more)

### Community 14 - ".queryRow"
Cohesion: 0.11
Nodes (4): HashToken(), NewToken(), Machine, Refusal

### Community 15 - "fakeServer"
Cohesion: 0.28
Nodes (3): fakeServer, signed(), Definitive()

### Community 16 - "api"
Cohesion: 0.08
Nodes (12): api, hostFamilies(), decodeJSON(), readBody(), writeJSON(), readSums(), errSession(), errWait() (+4 more)

### Community 17 - "serveRuntime"
Cohesion: 0.17
Nodes (16): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+8 more)

### Community 18 - "netlinkPolicy"
Cohesion: 0.20
Nodes (7): Family, netlinkPolicy, dump(), family(), owned(), toRoute(), toRule()

### Community 19 - "Dir"
Cohesion: 0.06
Nodes (42): TestStartNodeRuns(), meshKey(), TestAnswersMustBeSigned(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, Enrollment, NewKeys() (+34 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.19
Nodes (30): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), removable() (+22 more)

### Community 21 - "Run"
Cohesion: 0.13
Nodes (22): TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), Presented(), TestHashes(), cleanHostname(), Join(), meshAddresses() (+14 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (25): Routes, index(), route(), compare(), luidOf(), onLink(), TestULAPrefix(), ULAPrefix() (+17 more)

### Community 24 - "meshHarness"
Cohesion: 0.15
Nodes (25): cli(), TestIPv6DNSRecords(), TestReportedHostFamilies(), harness, TestMeshesAreApartOnTheAgentAPI(), TestMeshesStayApart(), TestNodesServiceInAMesh(), TestSetSettings() (+17 more)

### Community 25 - "Input"
Cohesion: 0.11
Nodes (15): PlanError, Routes, Record, containers6(), Check(), checkContainers(), checkContainers6(), checkNode6() (+7 more)

### Community 26 - "lab"
Cohesion: 0.18
Nodes (7): Coverage, host, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "env"
Cohesion: 0.19
Nodes (42): setupAgentJoin(), setupAgentReset(), socketFlags(), setupVersion(), wantArgs(), setupDNS(), setupExit(), setupExport() (+34 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.13
Nodes (18): Container names, 10. Backups, 11. When the control plane is down, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits (+10 more)

### Community 31 - "NodeConfig"
Cohesion: 0.05
Nodes (25): Deps, statusDetail, dialing, IPv6Host, Reset(), Run(), uiListen(), forHost() (+17 more)

### Community 32 - "context.Context"
Cohesion: 0.10
Nodes (10): runner, Rechecker, Add an apply step, kick(), oneLine(), Result, drain(), RunWindowsService() (+2 more)

### Community 33 - "newRig"
Cohesion: 0.20
Nodes (27): `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages(), TestPhonePage(), newRig() (+19 more)

### Community 34 - "Service"
Cohesion: 0.17
Nodes (8): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), serve()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.09
Nodes (71): rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime(), newFakeServer() (+63 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (125): Agent (`internal/agent/...`), freePort(), TestDefaultSocket(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), pfx(), TestConfigureHasKernelSemantics() (+117 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (32): rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword() (+24 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (51): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+43 more)

### Community 40 - ".exec"
Cohesion: 0.11
Nodes (5): mustAffect(), boolInt(), classify(), changed(), Session

### Community 41 - "`internal/server/store`"
Cohesion: 0.13
Nodes (48): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+40 more)

### Community 42 - "nullInt"
Cohesion: 0.16
Nodes (4): Add a database migration, DNSRecord, nullInt(), nullStr()

### Community 43 - "netlinkNet"
Cohesion: 0.08
Nodes (16): netlinkNet, Status, One image, two modes, blackhole(), dropRoute(), dump(), family(), prefixOf() (+8 more)

### Community 44 - "key"
Cohesion: 0.17
Nodes (17): intervalElems(), TestIntervalElems(), getAnswer(), hexOf(), pfxs(), TestParseUAPI(), TestParseUAPIRejectsMalformedValues(), TestUAPISetGolden() (+9 more)

### Community 45 - "planner_test.go"
Cohesion: 0.15
Nodes (41): Add a topology rule, `internal/server/planner`, Compute(), deliver(), deliverAddr(), encode(), firstDiff(), hasProblem() (+33 more)

### Community 46 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 47 - "Mesh"
Cohesion: 0.29
Nodes (7): SoleMesh(), freeSlot(), PickMesh(), pickMesh(), Mesh, newUUID(), scanMesh()

### Community 48 - "README.md"
Cohesion: 0.13
Nodes (12): Determinism and goldens, graphify, Style, What this repo is, You do not run the gates, Building, Contents, Documentation (+4 more)

### Community 49 - ".AddMesh"
Cohesion: 0.14
Nodes (17): meshSpace, Container-mode agent, checkApart(), clash(), freeCIDR(), meshCIDR(), spaceOf(), spacesOf() (+9 more)

### Community 50 - ".query"
Cohesion: 0.13
Nodes (6): ContainerName, scanAll(), splitList(), PeerStat, scanner, StatusUpdate

### Community 51 - "Test catalogue"
Cohesion: 0.10
Nodes (29): Packages, and which way they may import, Input and commands, Keys and tokens, Contents, `internal/archtest`, `internal/fsutil`, `internal/server/phone`, `internal/tools/vulnjudge` (+21 more)

### Community 52 - "Session"
Cohesion: 0.15
Nodes (23): brokenStream, TestSessionFiles(), HelloResponse, TestSessionShapes(), ValidSecret(), SignSession(), TestVerifyRejectsTampering(), TestEventsStreamPushesChanges() (+15 more)

### Community 53 - "LocalStatus"
Cohesion: 0.09
Nodes (22): Client, Logs, Handler(), logs(), refusal(), New(), laptop(), PrefsRequest (+14 more)

### Community 54 - "`internal/server/api`"
Cohesion: 0.09
Nodes (43): apiError, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket() (+35 more)

### Community 55 - "Open"
Cohesion: 0.15
Nodes (18): emptyService(), TestFirstAdmin(), TestFirstAdminAfterTheCLI(), TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), LegacySegmentation() (+10 more)

### Community 57 - "Client"
Cohesion: 0.13
Nodes (8): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, ValidChallenge(), DroppedName

### Community 58 - "Container"
Cohesion: 0.14
Nodes (5): fakeRuntime, Container, clip(), Source, meshFilter()

### Community 59 - "Enum"
Cohesion: 0.15
Nodes (11): validateMeshNames(), Ref(), TestRef(), Name(), SemVer(), TestName(), TestText(), TestVersionAndSemVer() (+3 more)

### Community 60 - "meshDNS"
Cohesion: 0.18
Nodes (26): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+18 more)

### Community 61 - "`internal/validate`"
Cohesion: 0.09
Nodes (24): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), LocalToken(), mask() (+16 more)

### Community 62 - "Sync"
Cohesion: 0.10
Nodes (38): network, Sysctls, NewNet(), ensureAddr(), Gateway(), gateway(), Net, Rule (+30 more)

### Community 63 - ".Apply"
Cohesion: 0.15
Nodes (14): applied, Firewall, policy, Runner, atLeast(), exitPolicy(), normalize(), relayPolicy() (+6 more)

### Community 64 - "Run"
Cohesion: 0.16
Nodes (17): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+9 more)

### Community 65 - "AddrRange"
Cohesion: 0.21
Nodes (14): addrs(), contains(), ofFamily(), inRanges(), inRange(), AddrEntry(), AddrRange, LastOf() (+6 more)

### Community 66 - ".meshctl"
Cohesion: 0.17
Nodes (15): 8. The web UI, Each node's local UI, The mesh firewall, A node's agent as a Swarm service, Agent service (systemd), Changes a node asks for, Contents, DNS (+7 more)

### Community 67 - "UI"
Cohesion: 0.09
Nodes (18): hostOf(), oldest(), validOrNot(), ago(), bytesize(), loadPages(), prefsFlash(), newRenderer() (+10 more)

### Community 68 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.04
Nodes (74): Options, ticketServer, `internal/proto`, Join(), New(), newClient(), sign(), TestConfigHonoursTheETag() (+66 more)

### Community 70 - "log/slog.Logger"
Cohesion: 0.08
Nodes (15): Change, Hub, subscription, blockingHost, NewUserspace(), NewUserspace(), NewUserspace(), ConfigEvent (+7 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.13
Nodes (19): `internal/tools/fetchwintun`, Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall(), download() (+11 more)

### Community 72 - "serve"
Cohesion: 0.19
Nodes (15): serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallPS1PinsTheChecksum(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums() (+7 more)

### Community 73 - "PublishNames"
Cohesion: 0.12
Nodes (16): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+8 more)

### Community 74 - "net/http.Request"
Cohesion: 0.09
Nodes (18): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), listOf() (+10 more)

### Community 75 - "CompilePolicy"
Cohesion: 0.20
Nodes (12): Matrix, MatrixCell, activeAt(), zoneOf(), CompilePolicy(), newSegResolver(), Overlaps(), Summarize() (+4 more)

### Community 76 - "Containers"
Cohesion: 0.14
Nodes (14): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, 4. Docker Desktop on Windows, A client node on the host network, A host-mode node's containers, Choosing how to run each node, Container mode (+6 more)

### Community 77 - "Sync"
Cohesion: 0.15
Nodes (7): Policy, mutate(), Policy, Route, Rule, V6Support, Sync()

### Community 78 - "printStatus"
Cohesion: 0.40
Nodes (5): TestBytesize(), dnsLine(), ago(), bytesize(), printStatus()

### Community 79 - "net/netip.Addr"
Cohesion: 0.10
Nodes (18): hostAddrs, nameTable, zone, TestIsPublic(), syncProxies(), proxy(), records(), firstAddr() (+10 more)

### Community 80 - "resolved"
Cohesion: 0.35
Nodes (4): linkDomain, resolved, classify(), foreignNetns()

### Community 81 - "errorf"
Cohesion: 0.04
Nodes (68): IPv6Request, IPv6View, PhoneRequest, RecordRequest, segLiteral, segDoc, Role, addIPv6() (+60 more)

### Community 82 - "rig"
Cohesion: 0.21
Nodes (9): code(), cookieOf(), newEmptyRig(), TestEnrolmentQRCodeIsAPNG(), TestFirstAdminFromTheWeb(), TestSignInFlow(), TestWrongEnrolmentCodeKeepsTheSecret(), faultyDB (+1 more)

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - ".renderPolicies"
Cohesion: 0.08
Nodes (24): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), readZoneForm(), ruleRows(), scheduleText(), tabsIn() (+16 more)

### Community 85 - "samples_test.go"
Cohesion: 0.13
Nodes (26): block, SampleParams, Sample, NodeSamples(), docWithSamples(), parseSample(), TestContainersDocSamples(), TestNodeSamples() (+18 more)

### Community 87 - "io.Writer"
Cohesion: 0.10
Nodes (18): agentLog(), TestPrintEnrollmentInTheLocalZone(), printExit(), main(), printMesh(), printContainers(), printEnrollment(), ttyPassword() (+10 more)

### Community 90 - "Host"
Cohesion: 0.07
Nodes (20): link, exitHost, Host, peer(), AddrPort(), hexKey(), ParseUAPI(), uapiInt() (+12 more)

### Community 93 - "exprs"
Cohesion: 0.18
Nodes (7): build(), exprs(), ifname(), intervalAddrElems(), TestExprs(), nftBatch, nftOp

### Community 94 - "Ruleset"
Cohesion: 0.10
Nodes (10): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel (+2 more)

### Community 95 - "Host"
Cohesion: 0.12
Nodes (11): Deps, Sysctls, NewTable(), Table, Sync(), NewTable(), NewTable(), Host (+3 more)

### Community 96 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.13
Nodes (3): defaultInterface(), Link, userspace

### Community 99 - "ns"
Cohesion: 0.10
Nodes (21): Several meshes on one host, End to end -- `test/e2e`, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), inNS() (+13 more)

### Community 100 - ".do"
Cohesion: 0.08
Nodes (23): MeshList, MeshSetRequest, MeshUpdate, NodeAddRequest, NodeDetail, NodeSetRequest, ReachRequest, SegmentView (+15 more)

### Community 101 - ".Run"
Cohesion: 0.33
Nodes (5): TestRunNeedsAnEnrollment(), TestRunGivesTheMeshIPv6(), TestRunLogsTheSetupCode(), TestRunRefusesBadConfig(), TestRunServesAndShutsDown()

### Community 102 - "run"
Cohesion: 0.13
Nodes (28): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+20 more)

### Community 103 - "Extending the code"
Cohesion: 0.14
Nodes (14): Add a command, Add a validator, Add a web UI page, Contents, Dev script tasks, Developer guide, Extending the code, Goldens (+6 more)

### Community 104 - "Schedule"
Cohesion: 0.16
Nodes (10): The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds() (+2 more)

### Community 105 - "time.Time"
Cohesion: 0.07
Nodes (21): pendingApply, safeApply, Failover, hubMove, hubState, replayCache, failures, hitWindow (+13 more)

### Community 106 - "context.Context"
Cohesion: 0.06
Nodes (33): Access, Config, DB, Lockout, Params, Pending, Sealer, newPassword() (+25 more)

### Community 107 - "contents_test.go"
Cohesion: 0.57
Nodes (6): docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents()

### Community 108 - "Read"
Cohesion: 0.39
Nodes (7): source, clean(), first(), fromFile(), sources(), Read(), sources()

### Community 111 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 112 - "NewHub"
Cohesion: 0.40
Nodes (6): NewHub(), closed(), TestHubDeliversChallenges(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 113 - "Engine"
Cohesion: 0.09
Nodes (9): nameServer, syncBuffer, rig, DNS, Firewall, Sysctl, Want, MeshName (+1 more)

### Community 114 - "net.Listener"
Cohesion: 0.20
Nodes (10): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), Listen(), TestServeAndListen(), ListenAdmin() (+2 more)

### Community 116 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 117 - "ServiceEntry"
Cohesion: 0.21
Nodes (6): serves(), Service, ServiceEntry(), TestServiceEntry(), Matcher, matchRule

### Community 118 - "rulesFor"
Cohesion: 0.33
Nodes (12): nrptRule, desired(), domainName(), TestZoneFollowsTheConfig(), fromRules(), keyPart(), nrptKey(), ownedKey() (+4 more)

### Community 119 - "run"
Cohesion: 0.27
Nodes (10): find(), isHelp(), printUsage(), run(), setupUser(), setupUserAdd(), setupUserGrant(), userCommands() (+2 more)

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.10
Nodes (30): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+22 more)

### Community 122 - ".Set"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 123 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Contents, Errors, Join, Local API, Mesh view, Prefs, Sessions (+2 more)

### Community 125 - "NodeInfo"
Cohesion: 0.10
Nodes (22): CopyInfo, NodeList, PhoneProfile, addresses(), heard(), ContainerNetwork, NodeInfo, Health() (+14 more)

### Community 130 - "Slot"
Cohesion: 0.09
Nodes (19): Pipe(), Holder(), Priorities(), SlotOf(), Taken(), TestHolder(), TestSlotOf(), NewKernel() (+11 more)

### Community 142 - "TestQRCode"
Cohesion: 0.67
Nodes (3): newQR(), phoneData, qrCode

### Community 144 - "Registration"
Cohesion: 0.52
Nodes (3): Domain, fakeRegistrar, Registration

### Community 146 - "service_windows_test.go"
Cohesion: 0.38
Nodes (3): executes(), TestHandlerReportsAFailedRun(), TestHandlerStopsTheAgent()

### Community 147 - "listText"
Cohesion: 0.29
Nodes (7): SegObject, anyText(), listText(), objectRows(), objectForm, objectRow, objectsView

### Community 148 - "stream"
Cohesion: 0.33
Nodes (4): dial(), dial(), isPipe(), stream

### Community 150 - "copy"
Cohesion: 0.70
Nodes (4): add(), copy(), done(), select()

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 396 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **33 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Command reference` connect `Command reference` to `README.md`, `.meshctl`, `run`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09815001351251751 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `netlinkNet`, `.mesh`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.11836734693877551 - nodes in this community are weakly interconnected._
- **Why does `Slot` connect `Slot` to `ui/ui_test.go`, `net/netip.AddrPort`, `fakeServer`, `api`, `serveRuntime`, `Dir`, `Run`, `Input`, `.mesh`, `NodeConfig`, `Service`, `testing.T`, `netlinkNet`, `Sync`, `.Apply`, ``internal/proto``, `Sync`, `errorf`, `Host`, `rulesFor`, `.Set`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._