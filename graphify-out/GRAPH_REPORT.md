# Graph Report - vpn-mesh-utility  (2026-11-07)

## Corpus Check
- 337 files · ~430,036 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .tmpl 16, (none) 5, .css 2)

## Summary
- 4174 nodes · 22424 edges · 153 communities (120 shown, 33 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2655 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `e83150a4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- flag.FlagSet
- Schedule
- htmx.min.js
- time.Time
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- Test catalogue
- deploy_test.go
- io.Writer
- github.com/hafio/vpn-mesh-utility
- Container
- newHarness
- api
- objectRows
- Join
- Dir
- Agent (`internal/agent/...`)
- hostid.go
- net/netip.Prefix
- Net
- fakeServer
- contents_test.go
- ns
- Snapshot
- env
- Command reference
- .mesh
- cli
- context.Context
- `internal/server/ui`
- Service
- agent/agent_test.go
- testing.T
- `internal/server/auth`
- testing.T
- Policy
- .exec
- `internal/server/store`
- NodeConfig
- netlinkNet
- Test catalogue
- planner_test.go
- Node
- netlinkPolicy
- README.md
- lab
- .query
- Input
- serveRuntime
- LocalStatus
- Run
- samples_test.go
- serve
- Client
- .do
- Ticket
- meshDNS
- IPNet
- Sync
- .Apply
- Run
- net/netip.Addr
- FromSnapshot
- UI
- serve
- `internal/proto`
- Security
- fetchwintun/main_test.go
- embed.go
- rulesFor
- net/http.Request
- otherPolicy
- Containers
- Sync
- log/slog.Logger
- exprs
- resolved
- context.Context
- Net
- New
- policyForm
- meshctl/docs_test.go
- DNS
- File
- Host
- Agent protocol
- Ruleset
- .queryRow
- .readJSON
- userspace
- otherNet
- .nodeView
- `deploy`
- meshctl/main.go
- TestSafeApplyStateMachine
- DefaultPolicy
- time.Duration
- Service
- nullInt
- .Reach
- samples_test.go
- StepError
- Handler
- key
- Host
- Name
- net.Listener
- .BeginEnrollment
- Extending the code
- service_windows.go
- Diff
- forms.js
- `internal/server` and `internal/server/export`
- TestDefaultSocket
- NewKernel
- Store
- nftKernel
- TestMain
- Table
- entry
- AdminError
- Slot
- Error
- NewNet
- NewNet
- .Sync
- NewKernel
- Refusal
- .add
- Refusal
- usageError
- StatusError
- AliveRequest
- sync.Mutex
- rig
- Error
- nodeRow
- zonesView
- stream
- noKernel
- build
- logs
- newQR

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 305 edges
2. `ui` - 110 edges
3. ``internal/server/api`` - 101 edges
4. `ServerConfig()` - 89 edges
5. `Session` - 85 edges
6. `Slot` - 82 edges
7. `env` - 73 edges
8. `newHarness()` - 73 edges
9. `NodeConfig` - 69 edges
10. `api` - 62 edges

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

## Communities (153 total, 33 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.09
Nodes (56): TestProxyLoopAppliesAgain(), TestRunNeedsAnEnrollment(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), TestLinksAddressesAndFailures(), TestRoutesSysctlAndFirewall(), TestTeardownFakes() (+48 more)

### Community 2 - "flag.FlagSet"
Cohesion: 0.22
Nodes (29): setupVersion(), setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), setupMeshAdd(), setupMeshList(), setupMeshRemove() (+21 more)

### Community 3 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "time.Time"
Cohesion: 0.10
Nodes (14): replayCache, failures, Lockout, LockoutError, serves(), CanonicalRequest(), Signed, DB (+6 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (35): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+27 more)

### Community 7 - "segResolver"
Cohesion: 0.12
Nodes (19): Service, checkSegmentation(), Segmentation, hosts(), mergePorts(), mergeServices(), newSegResolver(), Overlaps() (+11 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "Test catalogue"
Cohesion: 0.14
Nodes (16): Rule, Violation, Determinism and goldens, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates (+8 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.19
Nodes (21): checkPosture(), controlEnv(), hereDoc(), keys(), mustRead(), runScript(), sameSet(), service() (+13 more)

### Community 12 - "io.Writer"
Cohesion: 0.12
Nodes (15): watchdog, TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), setupDNS(), printExit(), printMesh(), printChange() (+7 more)

### Community 14 - "Container"
Cohesion: 0.31
Nodes (4): fakeRuntime, Container, ContainerName, NamesRequest

### Community 15 - "newHarness"
Cohesion: 0.11
Nodes (39): apiError, `internal/server/api`, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode() (+31 more)

### Community 16 - "api"
Cohesion: 0.07
Nodes (18): api, 17. Amendments, Nodes, hostFamilies(), decodeJSON(), readBody(), writeJSON(), readSums() (+10 more)

### Community 17 - "objectRows"
Cohesion: 0.40
Nodes (5): SegObject, objectRows(), objectForm, objectRow, objectsView

### Community 18 - "Join"
Cohesion: 0.13
Nodes (25): TestCleanHostname(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), Local(), Presented(), TestHashes(), cleanHostname() (+17 more)

### Community 19 - "Dir"
Cohesion: 0.05
Nodes (48): `internal/fsutil`, meshKey(), sign(), TestAnswersMustBeSigned(), TestConfigHonoursTheETag(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir (+40 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.19
Nodes (30): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), removable() (+22 more)

### Community 21 - "hostid.go"
Cohesion: 0.39
Nodes (7): source, clean(), first(), fromFile(), sources(), Read(), sources()

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (25): Routes, ensureAddr(), index(), route(), compare(), luidOf(), onLink(), TestULAPrefix() (+17 more)

### Community 24 - "fakeServer"
Cohesion: 0.13
Nodes (9): Change, fakeServer, signed(), ETag(), ConfigEvent, HelloResponse, Definitive(), SignRefusal() (+1 more)

### Community 25 - "contents_test.go"
Cohesion: 0.57
Nodes (6): docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents()

### Community 26 - "ns"
Cohesion: 0.09
Nodes (22): Several meshes on one host, End to end -- `test/e2e`, host, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces() (+14 more)

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "env"
Cohesion: 0.10
Nodes (54): agentLog(), checkMachineIDFile(), dataHelp(), dirMesh(), givenDir(), installedDir(), joinedAt(), joinFlags() (+46 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.06
Nodes (36): DNSEntry, NameEntry, NameSet, Container names, 10. Backups, 11. When the control plane is down, 12. Troubleshooting, 13. Upgrading (+28 more)

### Community 31 - "cli"
Cohesion: 0.14
Nodes (26): cli(), TestIPv6DNSRecords(), TestReportedHostFamilies(), harness, TestMeshesAreApartOnTheAgentAPI(), TestMeshesStayApart(), TestNodesServiceInAMesh(), TestSetSettings() (+18 more)

### Community 32 - "context.Context"
Cohesion: 0.13
Nodes (5): runner, kick(), oneLine(), Result, drain()

### Community 33 - "`internal/server/ui`"
Cohesion: 0.07
Nodes (55): Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, capped, tabsIn(), TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant() (+47 more)

### Community 34 - "Service"
Cohesion: 0.14
Nodes (9): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+1 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (71): rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime(), newFakeServer() (+63 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (119): Agent (`internal/agent/...`), freePort(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), pfx(), TestConfigureHasKernelSemantics(), TestDNS() (+111 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (30): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide() (+22 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (54): TestAgentRunRuntimeSocket(), enroll(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+46 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (51): `internal/proto`, NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local (+43 more)

### Community 40 - ".exec"
Cohesion: 0.07
Nodes (9): newUUID(), mustAffect(), joinList(), boolInt(), classify(), changed(), Session, PeerStat (+1 more)

### Community 41 - "`internal/server/store`"
Cohesion: 0.13
Nodes (45): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+37 more)

### Community 42 - "NodeConfig"
Cohesion: 0.09
Nodes (14): NodeDetail, dialing, IPv6Host, Rechecker, Add an apply step, forHost(), Engine, proxied() (+6 more)

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (11): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+3 more)

### Community 44 - "Test catalogue"
Cohesion: 0.16
Nodes (19): Contents, `internal/server/phone`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue, Hub(), joinSet(), NewKey() (+11 more)

### Community 45 - "planner_test.go"
Cohesion: 0.12
Nodes (47): PlanError, Add a topology rule, `internal/server/planner`, TestCheckRequested(), Problem, Compute(), deliver(), deliverAddr() (+39 more)

### Community 46 - "Node"
Cohesion: 0.20
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 47 - "netlinkPolicy"
Cohesion: 0.19
Nodes (8): Family, netlinkPolicy, SlotOf(), dump(), family(), owned(), toRoute(), toRule()

### Community 48 - "README.md"
Cohesion: 0.19
Nodes (7): Building, Contents, Documentation, How it fits together, meshctl, Quick start, Status

### Community 49 - "lab"
Cohesion: 0.23
Nodes (6): Coverage, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 50 - ".query"
Cohesion: 0.15
Nodes (4): ContainerName, scanAll(), splitList(), AuditEntry

### Community 51 - "Input"
Cohesion: 0.16
Nodes (13): Record, containers6(), Check(), checkContainers(), checkContainers6(), checkDNS(), checkNode6(), checkUsable() (+5 more)

### Community 52 - "serveRuntime"
Cohesion: 0.17
Nodes (16): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+8 more)

### Community 53 - "LocalStatus"
Cohesion: 0.13
Nodes (16): statusDetail, PrefsRequest, PrefsResponse, Rollback, LocalMesh, LocalStatus, sameRollback(), Funcs (+8 more)

### Community 54 - "Run"
Cohesion: 0.12
Nodes (9): Deps, Deps, Reset(), Run(), uiListen(), DNS, Firewall, MeshFirewall (+1 more)

### Community 55 - "samples_test.go"
Cohesion: 0.22
Nodes (14): SampleParams, Sample, NodeSamples(), docWithSamples(), parseSample(), TestContainersDocSamples(), TestNodeSamples(), TestSampleImage() (+6 more)

### Community 56 - "serve"
Cohesion: 0.23
Nodes (13): serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallPS1PinsTheChecksum(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums() (+5 more)

### Community 57 - "Client"
Cohesion: 0.10
Nodes (10): Session, apiError(), clip(), Client, clip(), meshFilter(), StatusResponse, NamesResponse (+2 more)

### Community 58 - ".do"
Cohesion: 0.10
Nodes (22): MeshList, meshSpace, SoleMesh(), checkApart(), clash(), freeCIDR(), freeSlot(), MeshAddRequest (+14 more)

### Community 59 - "Ticket"
Cohesion: 0.15
Nodes (23): brokenStream, TestSessionFiles(), ErrorBody, TestSessionShapes(), ValidSecret(), SignSession(), TestVerifyRejectsTampering(), TestEventsStreamPushesChanges() (+15 more)

### Community 60 - "meshDNS"
Cohesion: 0.17
Nodes (27): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+19 more)

### Community 61 - "IPNet"
Cohesion: 0.15
Nodes (5): Status, One image, two modes, lookup(), kernel, links

### Community 62 - "Sync"
Cohesion: 0.15
Nodes (33): Sysctls, NewNet(), TestContainersStepWithIPv6(), TestForHostWithoutIPv6(), Gateway(), Guard(), Owned(), parse() (+25 more)

### Community 63 - ".Apply"
Cohesion: 0.12
Nodes (15): applied, Firewall, policy, Firewall, Runner, atLeast(), exitPolicy(), normalize() (+7 more)

### Community 64 - "Run"
Cohesion: 0.14
Nodes (20): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+12 more)

### Community 65 - "net/netip.Addr"
Cohesion: 0.06
Nodes (38): network, hostAddrs, nameTable, zone, Apply (agent), TestIsPublic(), gateway(), host() (+30 more)

### Community 66 - "FromSnapshot"
Cohesion: 0.18
Nodes (13): FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes() (+5 more)

### Community 67 - "UI"
Cohesion: 0.08
Nodes (24): asset(), hostOf(), New(), oldest(), validOrNot(), ago(), bytesize(), loadPages() (+16 more)

### Community 68 - "serve"
Cohesion: 0.16
Nodes (24): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+16 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.09
Nodes (40): Options, ticketServer, Join(), New(), newClient(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents(), TestEventsRefusesAnythingButAStream() (+32 more)

### Community 70 - "Security"
Cohesion: 0.13
Nodes (19): 8. The web UI, Each node's local UI, The mesh firewall, A node's agent as a Swarm service, A node's container on the host network (the NAS), Changes a node asks for, Container-mode agent, Contents (+11 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.24
Nodes (13): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), TestDownload(), TestExtractChecksTheDLL() (+5 more)

### Community 72 - "embed.go"
Cohesion: 0.29
Nodes (10): ComposeControl(), InstallParams, RenderInstall(), RenderInstallPS1(), renderPS1(), RenderReinstall(), RenderReinstallPS1(), renderShell() (+2 more)

### Community 73 - "rulesFor"
Cohesion: 0.16
Nodes (18): Domain, fakeRegistrar, nrpt, nrptRule, Registration, desired(), domainName(), fromRules() (+10 more)

### Community 74 - "net/http.Request"
Cohesion: 0.08
Nodes (24): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), anyText() (+16 more)

### Community 76 - "Containers"
Cohesion: 0.14
Nodes (14): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, 4. Docker Desktop on Windows, A client node on the host network, A host-mode node's containers, Choosing how to run each node, Container mode (+6 more)

### Community 77 - "Sync"
Cohesion: 0.15
Nodes (7): Policy, mutate(), Policy, Route, Rule, V6Support, Sync()

### Community 78 - "log/slog.Logger"
Cohesion: 0.09
Nodes (15): Hub, subscription, blockingHost, NewUserspace(), NewUserspace(), NewHub(), closed(), TestHubDeliversChallenges() (+7 more)

### Community 79 - "exprs"
Cohesion: 0.24
Nodes (5): exprs(), ifname(), intervalAddrElems(), TestExprs(), nftOp

### Community 80 - "resolved"
Cohesion: 0.35
Nodes (4): linkDomain, resolved, classify(), foreignNetns()

### Community 81 - "context.Context"
Cohesion: 0.04
Nodes (83): IPv6Request, IPv6View, MeshSetRequest, MeshUpdate, NodeAddRequest, NodeList, NodeSetRequest, PhoneProfile (+75 more)

### Community 82 - "Net"
Cohesion: 0.16
Nodes (5): Net, Rule, syncDropped(), syncProxies(), syncRules()

### Community 83 - "New"
Cohesion: 0.21
Nodes (12): Config, Params, Pending, New(), CheckPassword(), DefaultParams(), HashPassword(), TestCheckPasswordNeverEchoesThePassword() (+4 more)

### Community 84 - "policyForm"
Cohesion: 0.12
Nodes (13): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), scheduleText(), title(), dayBox, policiesView (+5 more)

### Community 85 - "meshctl/docs_test.go"
Cohesion: 0.21
Nodes (14): cell(), envDoc(), renderCommands(), sentence(), startFlag(), TestCommandsDoc(), TestContainersDocNamesRealVariables(), TestEveryCommandHasHelp() (+6 more)

### Community 86 - "DNS"
Cohesion: 0.17
Nodes (4): nameServer, DNS, Want, MeshName

### Community 87 - "File"
Cohesion: 0.19
Nodes (8): main(), ttyPassword(), OpenFile(), lockFile(), unlockFile(), lockFile(), unlockFile(), File

### Community 90 - "Host"
Cohesion: 0.07
Nodes (19): link, exitHost, Host, peer(), AddrPort(), hexKey(), ParseUAPI(), uapiInt() (+11 more)

### Community 93 - "Agent protocol"
Cohesion: 0.18
Nodes (11): Agent protocol, Contents, Errors, Join, Local API, Mesh view, Prefs, Routes (+3 more)

### Community 94 - "Ruleset"
Cohesion: 0.16
Nodes (7): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, packet

### Community 95 - ".queryRow"
Cohesion: 0.11
Nodes (4): HashToken(), NewToken(), Machine, Refusal

### Community 96 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.08
Nodes (12): Holder(), Taken(), TestHolder(), adapterPart(), adapterSlot(), NewUserspace(), adapterGUID(), defaultInterface() (+4 more)

### Community 100 - ".nodeView"
Cohesion: 0.19
Nodes (11): CopyInfo, ContainerNetwork, CopyLine(), Health(), phoneHealth(), SharesLine(), ago(), bytesize() (+3 more)

### Community 101 - "`deploy`"
Cohesion: 0.27
Nodes (15): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallPS1Parses(), TestInstallRefusesHostileValues(), testParams(), TestPS1ScriptsCheckTheirVariables(), TestPS1ScriptsKeepToTheirBlock() (+7 more)

### Community 102 - "meshctl/main.go"
Cohesion: 0.12
Nodes (28): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+20 more)

### Community 103 - "TestSafeApplyStateMachine"
Cohesion: 0.44
Nodes (3): pendingApply, safeApply, TestSafeApplyStateMachine()

### Community 104 - "DefaultPolicy"
Cohesion: 0.47
Nodes (6): DefaultPolicy(), testSegmentation(), TestSegmentationValidate(), DecodeSnapshot(), TestDecodeSnapshotRefusesUnknownShapes(), TestSnapshotRestoreRoundTrip()

### Community 105 - "time.Duration"
Cohesion: 0.14
Nodes (10): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, TestJudgeHubs(), RateLimiter (+2 more)

### Community 106 - "Service"
Cohesion: 0.10
Nodes (17): Access, DB, resetTOTP(), fullExpiry(), Service, insertUser(), keepAdmin(), load() (+9 more)

### Community 107 - "nullInt"
Cohesion: 0.12
Nodes (4): DNSRecord, nullInt(), nullStr(), Version

### Community 108 - ".Reach"
Cohesion: 0.50
Nodes (4): ReachResult, matrixCell, matrixRow, reachView

### Community 109 - "samples_test.go"
Cohesion: 0.32
Nodes (11): block, checkComposeSample(), checkKubeSample(), checkRunSample(), checkStackSample(), dig(), docBlocks(), mountsMachineID() (+3 more)

### Community 111 - "Handler"
Cohesion: 0.40
Nodes (3): Client, Handler(), refusal()

### Community 112 - "key"
Cohesion: 0.17
Nodes (18): intervalElems(), TestIntervalElems(), getAnswer(), hexOf(), pfxs(), TestParseUAPI(), TestParseUAPIRejectsMalformedValues(), TestUAPISetGolden() (+10 more)

### Community 113 - "Host"
Cohesion: 0.31
Nodes (3): Host, SyncAddrs(), SyncLink()

### Community 114 - "Name"
Cohesion: 0.06
Nodes (70): Security invariants, `internal/validate`, check(), ParseAddr(), validateMeshNames(), HelloRequest, newNode(), validateSet() (+62 more)

### Community 115 - "net.Listener"
Cohesion: 0.23
Nodes (9): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), Listen(), ListenAdmin(), serve() (+1 more)

### Community 116 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 117 - "Extending the code"
Cohesion: 0.20
Nodes (10): Add a command, Add a database migration, Add a validator, Contents, Dev script tasks, Developer guide, Extending the code, Goldens (+2 more)

### Community 119 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.13
Nodes (26): `internal/server` and `internal/server/export`, NextExport(), TestExporterRunExportsAtNight(), TestExporterRunStopsWithItsContext(), TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestNextExport() (+18 more)

### Community 122 - "TestDefaultSocket"
Cohesion: 0.31
Nodes (9): TestDefaultSocket(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe(), waitPipe(), PipePath() (+1 more)

### Community 124 - "Store"
Cohesion: 0.16
Nodes (10): TestMigrationMakesMeshOne(), loadMigrations(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), Store, TestLoadMigrationsValidatesNames() (+2 more)

### Community 127 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 130 - "Slot"
Cohesion: 0.15
Nodes (9): Priorities(), TestSlotOf(), NewKernel(), NewKernel(), SysctlKeys(), Slot, InterfaceSlot(), TestInterfaceSlot() (+1 more)

### Community 137 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 144 - "rig"
Cohesion: 0.29
Nodes (4): rig, Sysctl, Sysctls, Sysctl

### Community 146 - "nodeRow"
Cohesion: 0.40
Nodes (5): dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 148 - "stream"
Cohesion: 0.29
Nodes (4): dial(), dial(), isPipe(), stream

### Community 152 - "newQR"
Cohesion: 0.67
Nodes (3): newQR(), phoneData, qrCode

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 396 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **33 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `time.Time`, `Test catalogue`, `Join`, `Dir`, `Agent (`internal/agent/...`)`, `.mesh`, `agent/agent_test.go`, `Policy`, `NodeConfig`, `Test catalogue`, `Node`, `serveRuntime`, `Ticket`, `meshDNS`, `Sync`, `Run`, `net/netip.Addr`, ``internal/proto``, `Security`, `rulesFor`, `exprs`, `DNS`, `Host`, `userspace`, `TestSafeApplyStateMachine`, `key`, `TestDefaultSocket`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Are the 304 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 304 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09802347503496929 - nodes in this community are weakly interconnected._
- **Why does `NodeConfig` connect `NodeConfig` to `ui/ui_test.go`, `WGKey`, `Test catalogue`, `Dir`, `fakeServer`, `.mesh`, `cli`, `context.Context`, `agent/agent_test.go`, `Policy`, `Test catalogue`, `planner_test.go`, `Node`, `Run`, `Client`, `Sync`, `context.Context`, `Agent protocol`, `TestSafeApplyStateMachine`, `key`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.08644067796610169 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `IPNet`, `.mesh`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._