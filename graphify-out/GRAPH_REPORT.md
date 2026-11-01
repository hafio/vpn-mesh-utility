# Graph Report - vpn-mesh-utility  (2026-11-01)

## Corpus Check
- 337 files · ~428,629 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .tmpl 16, (none) 5, .css 2)

## Summary
- 4170 nodes · 22400 edges · 152 communities (115 shown, 37 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2648 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `812796f9`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- ToSnapshot
- Schedule
- htmx.min.js
- AddrRange
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- Test catalogue
- deploy_test.go
- io.Writer
- github.com/hafio/vpn-mesh-utility
- serveRuntime
- newHarness
- api
- Runtime
- Exit
- Dir
- Agent (`internal/agent/...`)
- hostid.go
- net/netip.Prefix
- Net
- Policy
- meshctl/docs_test.go
- TestSpokesAndExit
- Snapshot
- env
- Command reference
- .mesh
- DNS
- context.Context
- newRig
- Service
- agent/agent_test.go
- testing.T
- `internal/server/auth`
- testing.T
- New
- .exec
- `internal/server/store`
- Run
- netlinkNet
- NodeConfig
- planner_test.go
- Node
- PublishNames
- README.md
- .SetSegmentation
- .query
- Input
- netlinkPolicy
- LocalStatus
- `internal/server/api`
- samples_test.go
- newRig
- Client
- .do
- Session
- meshDNS
- IPNet
- Sync
- .Apply
- Run
- net/netip.Addr
- sync.Mutex
- UI
- serve
- `internal/proto`
- Security
- fetchwintun/main_test.go
- embed.go
- rulesFor
- net/http.Request
- samples_test.go
- Containers
- Sync
- Hub
- nftKernel
- sync.Mutex
- Node
- Net
- io.Reader
- policyForm
- Open
- otherPolicy
- File
- Host
- exprs
- Ruleset
- .queryRow
- .readJSON
- userspace
- otherNet
- NodeInfo
- Allocate
- run
- run
- TestPageGoldens
- time.Duration
- Service
- nullInt
- rig
- TestControlComposePosture
- StepError
- Developer guide
- Agent protocol
- ns
- fail
- DefaultPolicy
- TestSafeApplyStateMachine
- copy
- service_windows.go
- Diff
- forms.js
- `internal/server` and `internal/server/export`
- .Reach
- syncBuffer
- loadMigrations
- Writer
- TestMain
- Firewall
- entry
- AdminError
- Slot
- Error
- NewNet
- NewNet
- Signed
- NewKernel
- Refusal
- NewKernel
- Refusal
- usageError
- StatusError
- AliveRequest
- AuditEntry
- Version
- Error
- nodeRow
- TestQRCode
- PeerStat
- Machine
- zonesView
- 8. The web UI

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 304 edges
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

## Communities (152 total, 37 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.15
Nodes (45): blockingHost, rig, watchingFirewall, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New() (+37 more)

### Community 2 - "ToSnapshot"
Cohesion: 0.09
Nodes (45): check(), ParseAddr(), validateMeshNames(), HelloRequest, newNode(), validateSet(), ToSnapshot(), checkDNS() (+37 more)

### Community 3 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "AddrRange"
Cohesion: 0.12
Nodes (16): addrText(), contains(), Rule, merged(), ofFamily(), portText(), services(), inRanges() (+8 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem, 2. Goals (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.11
Nodes (21): addrs(), Service, serves(), checkSegmentation(), hosts(), mergePorts(), mergeServices(), Overlaps() (+13 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "Test catalogue"
Cohesion: 0.11
Nodes (20): Rule, Violation, Determinism and goldens, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates (+12 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.23
Nodes (22): hereDoc(), runScript(), TestInstallChangesOnlyTheMarkers(), TestInstallChecksTheDataDir(), TestInstallDataDropIn(), TestInstallDataRecord(), TestInstallGolden(), TestInstallParses() (+14 more)

### Community 12 - "io.Writer"
Cohesion: 0.16
Nodes (36): setupVersion(), TestBytesize(), dnsLine(), setupDNS(), setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6() (+28 more)

### Community 14 - "serveRuntime"
Cohesion: 0.18
Nodes (15): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+7 more)

### Community 15 - "newHarness"
Cohesion: 0.10
Nodes (32): adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing(), TestAdminSetClearsReferences() (+24 more)

### Community 16 - "api"
Cohesion: 0.08
Nodes (17): api, signed(), writeJSONT(), decode(), reply(), ETag(), ParseSigned(), WriteRetry() (+9 more)

### Community 19 - "Dir"
Cohesion: 0.04
Nodes (56): enroll(), `internal/fsutil`, meshKey(), TestJoin(), Local(), Presented(), TestHashes(), cleanHostname() (+48 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.20
Nodes (29): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), removable() (+21 more)

### Community 21 - "hostid.go"
Cohesion: 0.39
Nodes (7): source, clean(), first(), fromFile(), sources(), Read(), sources()

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (17): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+9 more)

### Community 24 - "Policy"
Cohesion: 0.18
Nodes (21): NewPolicy(), NewPolicy(), NewPolicy(), Compile(), compile(), dualPolicy(), officeHours(), sgt() (+13 more)

### Community 25 - "meshctl/docs_test.go"
Cohesion: 0.15
Nodes (20): flagName(), docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents(), cell() (+12 more)

### Community 26 - "TestSpokesAndExit"
Cohesion: 0.19
Nodes (7): Coverage, host, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "env"
Cohesion: 0.12
Nodes (51): checkMachineIDFile(), dataHelp(), dirMesh(), givenDir(), installedDir(), joinedAt(), joinFlags(), machineIDFlag() (+43 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.11
Nodes (20): Container names, 10. Backups, 11. When the control plane is down, 12. Troubleshooting, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh (+12 more)

### Community 31 - "DNS"
Cohesion: 0.17
Nodes (4): nameServer, DNS, Want, MeshName

### Community 32 - "context.Context"
Cohesion: 0.10
Nodes (9): runner, Rechecker, Add an apply step, kick(), oneLine(), Result, drain(), RunWindowsService() (+1 more)

### Community 33 - "newRig"
Cohesion: 0.22
Nodes (26): `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages(), TestPhonePage(), newRig() (+18 more)

### Community 34 - "Service"
Cohesion: 0.07
Nodes (23): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), bindBoth() (+15 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.07
Nodes (80): rig, enroll(), enrollAs(), eventually(), freePort(), fakeServer, joinedTo(), newDir() (+72 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (138): Agent (`internal/agent/...`), TestCleanHostname(), TestDefaultSocket(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig() (+130 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.10
Nodes (39): Config, DB, Params, rig, 11. CLI reference, `internal/server/auth`, New(), newRig() (+31 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (54): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+46 more)

### Community 39 - "New"
Cohesion: 0.36
Nodes (4): New(), sameOrigin(), Deps, handler

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (5): newUUID(), mustAffect(), boolInt(), classify(), changed()

### Community 41 - "`internal/server/store`"
Cohesion: 0.14
Nodes (43): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+35 more)

### Community 42 - "Run"
Cohesion: 0.09
Nodes (12): Deps, Deps, dialing, Reset(), Run(), uiListen(), DNS, Engine (+4 more)

### Community 43 - "netlinkNet"
Cohesion: 0.13
Nodes (10): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+2 more)

### Community 44 - "NodeConfig"
Cohesion: 0.06
Nodes (46): IPv6Host, Keys and tokens, `internal/server/phone`, forHost(), TestForHostWithoutIPv6(), Desired(), prefixes(), TestDesiredSkipsTheDefaultRoutesAndDuplicates() (+38 more)

### Community 45 - "planner_test.go"
Cohesion: 0.15
Nodes (39): Add a topology rule, `internal/server/planner`, deliver(), deliverAddr(), encode(), firstDiff(), hasProblem(), lookup() (+31 more)

### Community 46 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 47 - "PublishNames"
Cohesion: 0.16
Nodes (13): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+5 more)

### Community 48 - "README.md"
Cohesion: 0.19
Nodes (7): Building, Contents, Documentation, How it fits together, meshctl, Quick start, Status

### Community 50 - ".query"
Cohesion: 0.22
Nodes (3): scanAll(), splitList(), scanner

### Community 51 - "Input"
Cohesion: 0.11
Nodes (23): PlanError, Record, activeAt(), zoneOf(), Check(), checkContainers(), checkContainers6(), checkNode6() (+15 more)

### Community 52 - "netlinkPolicy"
Cohesion: 0.18
Nodes (9): Family, netlinkPolicy, SlotOf(), dump(), family(), owned(), TestKernelShapes(), toRoute() (+1 more)

### Community 53 - "LocalStatus"
Cohesion: 0.07
Nodes (27): statusDetail, Client, Logs, logs(), refusal(), Listen(), New(), oldest() (+19 more)

### Community 54 - "`internal/server/api`"
Cohesion: 0.12
Nodes (36): apiError, `internal/server/api`, cli(), ErrorMessage(), toAPIError(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries() (+28 more)

### Community 55 - "samples_test.go"
Cohesion: 0.32
Nodes (11): block, checkComposeSample(), checkKubeSample(), checkRunSample(), checkStackSample(), dig(), docBlocks(), mountsMachineID() (+3 more)

### Community 56 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 57 - "Client"
Cohesion: 0.13
Nodes (8): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, ValidChallenge(), DroppedName

### Community 58 - ".do"
Cohesion: 0.07
Nodes (29): MeshList, meshSpace, MeshUpdate, NodeAddRequest, NodeSetRequest, SegmentView, ZoneView, SoleMesh() (+21 more)

### Community 59 - "Session"
Cohesion: 0.16
Nodes (22): brokenStream, TestSessionFiles(), ErrorBody, HelloResponse, ValidSecret(), SignSession(), TestVerifyRejectsTampering(), TestEventsStreamPushesChanges() (+14 more)

### Community 60 - "meshDNS"
Cohesion: 0.19
Nodes (21): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestServiceAnswersAndRegisters(), TestServiceAnswersOverIPv6(), TestServiceAnswersTheContainers() (+13 more)

### Community 61 - "IPNet"
Cohesion: 0.13
Nodes (6): Status, One image, two modes, IPNet(), lookup(), kernel, links

### Community 62 - "Sync"
Cohesion: 0.18
Nodes (28): Sysctls, NewNet(), Gateway(), Guard(), Owned(), parse(), SlotOf(), Sync() (+20 more)

### Community 63 - ".Apply"
Cohesion: 0.16
Nodes (13): applied, Firewall, policy, Runner, atLeast(), exitPolicy(), normalize(), relayPolicy() (+5 more)

### Community 64 - "Run"
Cohesion: 0.12
Nodes (23): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+15 more)

### Community 65 - "net/netip.Addr"
Cohesion: 0.08
Nodes (27): network, hostAddrs, nameTable, zone, TestIsPublic(), proxied(), gateway(), host() (+19 more)

### Community 66 - "sync.Mutex"
Cohesion: 0.09
Nodes (9): fakeRuntime, syncBuffer, Sysctl, clip(), Ring, Container, Source, Sysctl (+1 more)

### Community 67 - "UI"
Cohesion: 0.13
Nodes (12): hostOf(), validOrNot(), bytesize(), loadPages(), prefsFlash(), TestULAPrefix(), ULAPrefix(), handler (+4 more)

### Community 68 - "serve"
Cohesion: 0.13
Nodes (33): countingDB, harness, joinBody(), postJoin(), serve(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash() (+25 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.05
Nodes (68): Options, ticketServer, `internal/proto`, Join(), New(), newClient(), sign(), TestAnswersMustBeSigned() (+60 more)

### Community 70 - "Security"
Cohesion: 0.13
Nodes (15): A node's agent as a Swarm service, Changes a node asks for, Container-mode agent, Contents, DNS, Interfaces the agent did not create, Security, Several meshes on one host (+7 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.20
Nodes (13): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), TestDownload(), TestExtractChecksTheDLL() (+5 more)

### Community 72 - "embed.go"
Cohesion: 0.41
Nodes (10): InstallParams, RenderInstall(), RenderInstallPS1(), renderPS1(), RenderReinstall(), RenderReinstallPS1(), renderShell(), RenderUninstall() (+2 more)

### Community 73 - "rulesFor"
Cohesion: 0.17
Nodes (19): nrpt, nrptRule, desired(), domainName(), same(), fromRules(), keyPart(), nrptKey() (+11 more)

### Community 74 - "net/http.Request"
Cohesion: 0.07
Nodes (27): TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), anyText(), listOf(), listText() (+19 more)

### Community 75 - "samples_test.go"
Cohesion: 0.20
Nodes (15): SampleParams, Sample, NodeSamples(), docWithSamples(), parseSample(), TestContainersDocSamples(), TestNodeSamples(), TestSampleImage() (+7 more)

### Community 76 - "Containers"
Cohesion: 0.14
Nodes (14): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, 4. Docker Desktop on Windows, A client node on the host network, A host-mode node's containers, Choosing how to run each node, Container mode (+6 more)

### Community 77 - "Sync"
Cohesion: 0.16
Nodes (8): Policy, mutate(), Desired(), Policy, Route, Rule, Sync(), TestSyncWithoutIPv6PolicyRouting()

### Community 78 - "Hub"
Cohesion: 0.12
Nodes (13): Change, Hub, subscription, ConfigEvent, NewHub(), closed(), TestHubDeliversChallenges(), TestHubLatestWins() (+5 more)

### Community 79 - "nftKernel"
Cohesion: 0.29
Nodes (4): build(), conn(), nftBatch, nftKernel

### Community 80 - "sync.Mutex"
Cohesion: 0.20
Nodes (8): Domain, fakeRegistrar, linkDomain, Registration, resolved, TestServiceRevert(), classify(), foreignNetns()

### Community 81 - "Node"
Cohesion: 0.04
Nodes (67): IPv6Request, IPv6View, MeshSetRequest, PhoneRequest, RecordRequest, segLiteral, SettingsRequest, segDoc (+59 more)

### Community 82 - "Net"
Cohesion: 0.14
Nodes (7): ensureAddr(), Net, Rule, Remove(), syncDropped(), syncProxies(), syncRules()

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - "policyForm"
Cohesion: 0.12
Nodes (14): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), ruleRows(), scheduleText(), title(), dayBox (+6 more)

### Community 85 - "Open"
Cohesion: 0.13
Nodes (17): emptyService(), TestFirstAdmin(), TestFirstAdminAfterTheCLI(), TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), insertAt5() (+9 more)

### Community 87 - "File"
Cohesion: 0.12
Nodes (16): agentLog(), TestPrintEnrollmentInTheLocalZone(), printExit(), main(), printMesh(), printChange(), printEnrollment(), ttyPassword() (+8 more)

### Community 90 - "Host"
Cohesion: 0.07
Nodes (19): link, exitHost, Host, peer(), AddrPort(), hexKey(), ParseUAPI(), uapiInt() (+11 more)

### Community 93 - "exprs"
Cohesion: 0.21
Nodes (6): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), nftOp

### Community 94 - "Ruleset"
Cohesion: 0.11
Nodes (9): Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel, noKernel (+1 more)

### Community 95 - ".queryRow"
Cohesion: 0.11
Nodes (4): AgentSession, HashToken(), NewToken(), Refusal

### Community 96 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.10
Nodes (6): adapterPart(), NewUserspace(), adapterGUID(), defaultInterface(), Link, userspace

### Community 100 - "NodeInfo"
Cohesion: 0.12
Nodes (21): CopyInfo, NodeDetail, NodeList, PhoneProfile, addresses(), heard(), ContainerNetwork, NodeInfo (+13 more)

### Community 101 - "Allocate"
Cohesion: 0.19
Nodes (17): meshCIDR(), Allocate(), Allocate6(), AllocateContainers(), AllocateContainers6(), Block(), Block6(), CheckMeshCIDR() (+9 more)

### Community 102 - "run"
Cohesion: 0.13
Nodes (26): agentCommands(), argSynopsis(), commands(), newFlags(), parseArgs(), parseFlags(), printFlags(), shownDefault() (+18 more)

### Community 103 - "run"
Cohesion: 0.16
Nodes (10): DB, NextExport(), TestExporterRunExportsAtNight(), TestExporterRunStopsWithItsContext(), TestNextExport(), TestRunGivesTheMeshIPv6(), TestRunLogsTheSetupCode(), main() (+2 more)

### Community 104 - "TestPageGoldens"
Cohesion: 0.17
Nodes (13): Add a web UI page, The web UI's pages (internal/server/ui), tabsIn(), dnsState(), newRenderer(), golden(), TestEveryPageRenders(), TestFormsShowWhatApplies() (+5 more)

### Community 105 - "time.Duration"
Cohesion: 0.12
Nodes (11): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, TestJudgeHubs(), RateLimiter (+3 more)

### Community 106 - "Service"
Cohesion: 0.06
Nodes (31): replayCache, Access, failures, Lockout, LockoutError, Pending, Sealer, resetTOTP() (+23 more)

### Community 107 - "nullInt"
Cohesion: 0.18
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 108 - "rig"
Cohesion: 0.19
Nodes (9): code(), cookieOf(), newEmptyRig(), TestEnrolmentQRCodeIsAPNG(), TestFirstAdminFromTheWeb(), TestSignInFlow(), TestWrongEnrolmentCodeKeepsTheSecret(), faultyDB (+1 more)

### Community 109 - "TestControlComposePosture"
Cohesion: 0.27
Nodes (12): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+4 more)

### Community 111 - "Developer guide"
Cohesion: 0.20
Nodes (10): Add a command, Add a database migration, Add a validator, Contents, Dev script tasks, Developer guide, Extending the code, Goldens (+2 more)

### Community 112 - "Agent protocol"
Cohesion: 0.17
Nodes (12): Agent protocol, Apply (agent), Contents, Errors, Join, Local API, Mesh view, Prefs (+4 more)

### Community 113 - "ns"
Cohesion: 0.10
Nodes (20): End to end -- `test/e2e`, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), inNS(), newLab() (+12 more)

### Community 114 - "fail"
Cohesion: 0.07
Nodes (31): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), TestAddrEntry(), TestAddrRangePrefix(), TestRef() (+23 more)

### Community 115 - "DefaultPolicy"
Cohesion: 0.36
Nodes (8): DefaultPolicy(), LegacySegmentation(), TestMigrationTurnsRulesIntoPolicies(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSetSegmentationRefusesAStranger(), FirewallRule

### Community 116 - "TestSafeApplyStateMachine"
Cohesion: 0.46
Nodes (3): pendingApply, safeApply, TestSafeApplyStateMachine()

### Community 117 - "copy"
Cohesion: 0.70
Nodes (4): add(), copy(), done(), select()

### Community 119 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.13
Nodes (25): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+17 more)

### Community 122 - ".Reach"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 124 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 125 - "Writer"
Cohesion: 0.29
Nodes (3): New(), New(), Writer

### Community 130 - "Slot"
Cohesion: 0.11
Nodes (15): Pipe(), Holder(), Priorities(), Taken(), TestHolder(), TestSlotOf(), NewKernel(), NewKernel() (+7 more)

### Community 146 - "nodeRow"
Cohesion: 0.40
Nodes (5): dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 147 - "TestQRCode"
Cohesion: 0.50
Nodes (4): newQR(), TestQRCode(), phoneData, qrCode

### Community 151 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 396 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **37 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `Slot`, `AddrRange`, `Test catalogue`, `serveRuntime`, `Dir`, `Agent (`internal/agent/...`)`, `Policy`, `TestSpokesAndExit`, `.mesh`, `DNS`, `context.Context`, `agent/agent_test.go`, `NodeConfig`, `Node`, `newRig`, `Session`, `meshDNS`, `Sync`, `Run`, `net/netip.Addr`, ``internal/proto``, `rulesFor`, `sync.Mutex`, `Host`, `exprs`, `TestSafeApplyStateMachine`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09639723860204229 - nodes in this community are weakly interconnected._
- **Why does `NodeConfig` connect `NodeConfig` to `ui/ui_test.go`, `WGKey`, `Test catalogue`, `Exit`, `Dir`, `Policy`, `context.Context`, `agent/agent_test.go`, `Run`, `planner_test.go`, `Node`, `Input`, `LocalStatus`, ``internal/server/api``, `Client`, `Sync`, `net/netip.Addr`, `Hub`, `Node`, `NodeInfo`, `Agent protocol`, `TestSafeApplyStateMachine`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.14710884353741496 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `IPNet`, `.mesh`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._