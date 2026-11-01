# Graph Report - vpn-mesh-utility  (2026-11-01)

## Corpus Check
- 337 files · ~428,236 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .tmpl 16, (none) 5, .css 2)

## Summary
- 4168 nodes · 22393 edges · 140 communities (111 shown, 29 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2645 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1c74ce19`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- ToSnapshot
- Schedule
- htmx.min.js
- Compile
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- Test catalogue
- deploy_test.go
- io.Writer
- github.com/hafio/vpn-mesh-utility
- serveRuntime
- fakeServer
- api
- Runtime
- Exit
- Dir
- Agent (`internal/agent/...`)
- Run
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
- .setup
- netlinkNet
- NodeConfig
- planner_test.go
- Node
- .AddMesh
- README.md
- errorf
- .query
- lab
- netlinkPolicy
- LocalStatus
- `internal/server/api`
- samples_test.go
- newRig
- Client
- context.Context
- inNS
- meshDNS
- kernel
- Sync
- .Apply
- Run
- net/netip.Addr
- .add
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
- 7. Agent
- nftKernel
- sync.Mutex
- Node
- .Set
- io.Reader
- TestPageGoldens
- wgdev
- otherPolicy
- File
- Host
- exprs
- Ruleset
- .Sync
- .readJSON
- userspace
- TestOneRunningCopy
- .nodeView
- net.Listener
- run
- Registration
- TestMeshFirewall
- sync.Mutex
- Service
- contents_test.go
- meshctl
- 8. Server
- StepError
- Developer guide
- Of
- TestContainerNetworks
- fail
- ServerURL
- build
- copy
- noKernel
- walk
- forms.js
- `internal/server` and `internal/server/export`
- Open
- TestMain
- entry
- AdminError
- Slot
- Error
- NewNet
- NewNet
- NewKernel
- Refusal
- NewKernel
- Refusal
- usageError
- StatusError
- AliveRequest
- Error
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
- `Packages, and which way they may import` --references--> `TestModuleBoundaries()`  [INFERRED]
  CLAUDE.md → internal/archtest/archtest_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go

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

## Communities (140 total, 29 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.13
Nodes (48): blockingHost, rig, watchingFirewall, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New() (+40 more)

### Community 2 - "ToSnapshot"
Cohesion: 0.13
Nodes (29): check(), ParseAddr(), TestParseAddr(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), Endpoint6() (+21 more)

### Community 3 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "Compile"
Cohesion: 0.17
Nodes (15): TestKernelNeedsPrivilege(), addrs(), Compile(), contains(), ofFamily(), inRanges(), inRange(), AddrEntry() (+7 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.09
Nodes (22): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+14 more)

### Community 7 - "segResolver"
Cohesion: 0.17
Nodes (13): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), parseSpans(), PrefixRange(), Ref() (+5 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "Test catalogue"
Cohesion: 0.24
Nodes (10): Rule, Violation, `internal/archtest`, Check(), importsOf(), ModuleRoot(), TestCheckerCatchesViolations(), TestModuleBoundaries() (+2 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.15
Nodes (34): checkPosture(), controlEnv(), hereDoc(), keys(), mustRead(), runScript(), sameSet(), service() (+26 more)

### Community 12 - "io.Writer"
Cohesion: 0.17
Nodes (35): machineIDFlag(), TestBytesize(), dnsLine(), setupDNS(), setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6() (+27 more)

### Community 14 - "serveRuntime"
Cohesion: 0.20
Nodes (14): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+6 more)

### Community 15 - "fakeServer"
Cohesion: 0.13
Nodes (13): fakeServer, TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Definitive() (+5 more)

### Community 16 - "api"
Cohesion: 0.13
Nodes (9): api, writeJSONT(), decode(), reply(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+1 more)

### Community 17 - "Runtime"
Cohesion: 0.36
Nodes (3): clip(), meshFilter(), Runtime

### Community 19 - "Dir"
Cohesion: 0.05
Nodes (45): enroll(), `internal/fsutil`, signed(), TestDefaultSocket(), sign(), TestAnswersMustBeSigned(), TestDataDirIsProtectedOnWindows(), TestWipe() (+37 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.18
Nodes (31): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+23 more)

### Community 21 - "Run"
Cohesion: 0.10
Nodes (32): source, ranHere(), TestCleanHostname(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), clean(), first() (+24 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.06
Nodes (30): otherNet, Routes, NewTable(), compare(), Table, Sync(), prefixes(), TestSyncAddsThenRemovesAndIsIdempotent() (+22 more)

### Community 24 - "Policy"
Cohesion: 0.13
Nodes (23): NewPolicy(), NewPolicy(), NewPolicy(), compile(), dualPolicy(), officeHours(), sgt(), TestCompileAgreesWithDecide() (+15 more)

### Community 25 - "meshctl/docs_test.go"
Cohesion: 0.22
Nodes (15): flagName(), cell(), envDoc(), renderCommands(), sentence(), startFlag(), TestCommandsDoc(), TestContainersDocNamesRealVariables() (+7 more)

### Community 26 - "TestSpokesAndExit"
Cohesion: 0.29
Nodes (5): Coverage, eventually(), TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.10
Nodes (24): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+16 more)

### Community 28 - "env"
Cohesion: 0.12
Nodes (51): checkMachineIDFile(), dataHelp(), dirMesh(), givenDir(), installedDir(), joinedAt(), joinFlags(), meshDir() (+43 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.12
Nodes (19): Container names, 10. Backups, 11. When the control plane is down, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits (+11 more)

### Community 31 - "DNS"
Cohesion: 0.14
Nodes (6): nameServer, DNS, Want, newNameTable(), DNS, MeshName

### Community 32 - "context.Context"
Cohesion: 0.08
Nodes (16): Deps, pendingApply, runner, safeApply, 12. Troubleshooting, kick(), meshKey(), oneLine() (+8 more)

### Community 33 - "newRig"
Cohesion: 0.11
Nodes (45): The web UI's pages (internal/server/ui), `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages(), newQR() (+37 more)

### Community 34 - "Service"
Cohesion: 0.15
Nodes (7): resolver, Service, New(), exchange(), serve(), NewUserspace(), Options

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (72): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), newFakeRuntime() (+64 more)

### Community 36 - "testing.T"
Cohesion: 0.04
Nodes (77): TestHealthcheckStart(), Agent (`internal/agent/...`), Exec, TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx(), TestConfigureHasKernelSemantics(), TestKernelHoldsOneTable() (+69 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.08
Nodes (55): Params, rig, `internal/server/auth`, TestDNS(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream() (+47 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (51): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+43 more)

### Community 39 - "New"
Cohesion: 0.33
Nodes (4): New(), sameOrigin(), handler, meshHandler

### Community 40 - ".exec"
Cohesion: 0.04
Nodes (16): mustAffect(), DNSRecord, AgentSession, boolInt(), classify(), nullInt(), nullStr(), HashToken() (+8 more)

### Community 41 - "`internal/server/store`"
Cohesion: 0.10
Nodes (57): `internal/server/store`, TestReconcilerVersionsOnlyRealChanges(), TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestMigrationMakesMeshOne() (+49 more)

### Community 42 - ".setup"
Cohesion: 0.18
Nodes (5): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), group()

### Community 43 - "netlinkNet"
Cohesion: 0.08
Nodes (17): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), TestKernelShapes(), toRule() (+9 more)

### Community 44 - "NodeConfig"
Cohesion: 0.05
Nodes (46): NodeDetail, dialing, IPv6Host, Rechecker, Agent protocol, Contents, Errors, Join (+38 more)

### Community 45 - "planner_test.go"
Cohesion: 0.07
Nodes (65): PlanError, Add a topology rule, `internal/server/planner`, peer(), Policy, containerNetwork(), activeAt(), zoneOf() (+57 more)

### Community 46 - "Node"
Cohesion: 0.19
Nodes (11): Config and events, 4. Roles and topology, Peer, checkUsable(), ContainerNetwork(), endpoints(), Node, sortPeers() (+3 more)

### Community 47 - ".AddMesh"
Cohesion: 0.10
Nodes (25): DNSEntry, MeshList, meshSpace, NameEntry, NameSet, ULAPrefix(), checkApart(), clash() (+17 more)

### Community 48 - "README.md"
Cohesion: 0.21
Nodes (5): graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates

### Community 49 - "errorf"
Cohesion: 0.07
Nodes (27): RecordRequest, segLiteral, SegmentView, SettingsRequest, ZoneView, segDoc, Actor, segLiterals() (+19 more)

### Community 50 - ".query"
Cohesion: 0.09
Nodes (11): freeSlot(), pickMesh(), Mesh, newUUID(), scanMesh(), ContainerName, scanNode(), scanAll() (+3 more)

### Community 51 - "lab"
Cohesion: 0.23
Nodes (4): deleteNamespaces(), lab, newLab(), tail()

### Community 52 - "netlinkPolicy"
Cohesion: 0.18
Nodes (9): Family, netlinkPolicy, dump(), family(), owned(), TestKernelShapes(), TestOwnsOnlyItsPriorities(), toRoute() (+1 more)

### Community 53 - "LocalStatus"
Cohesion: 0.08
Nodes (25): statusDetail, Add a web UI page, Client, Logs, logs(), refusal(), laptop(), TestTemplatesHaveNoScript() (+17 more)

### Community 54 - "`internal/server/api`"
Cohesion: 0.06
Nodes (81): `internal/server/api`, cli(), ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode() (+73 more)

### Community 55 - "samples_test.go"
Cohesion: 0.32
Nodes (11): block, checkComposeSample(), checkKubeSample(), checkRunSample(), checkStackSample(), dig(), docBlocks(), mountsMachineID() (+3 more)

### Community 56 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 57 - "Client"
Cohesion: 0.10
Nodes (11): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, HelloResponse, ValidChallenge() (+3 more)

### Community 58 - "context.Context"
Cohesion: 0.04
Nodes (31): IPv6Request, IPv6View, MeshSetRequest, MeshUpdate, NodeAddRequest, NodeSetRequest, ReachRequest, DNS (+23 more)

### Community 59 - "inNS"
Cohesion: 0.20
Nodes (7): End to end -- `test/e2e`, host, lab, TestDualStack(), inNS(), lab, TestTwoMeshes()

### Community 60 - "meshDNS"
Cohesion: 0.23
Nodes (23): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+15 more)

### Community 61 - "kernel"
Cohesion: 0.22
Nodes (4): Status, One image, two modes, TestCreateLinkNeedsPrivilege(), kernel

### Community 62 - "Sync"
Cohesion: 0.09
Nodes (37): Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard(), Owned() (+29 more)

### Community 63 - ".Apply"
Cohesion: 0.13
Nodes (14): applied, Firewall, policy, Firewall, Runner, exitPolicy(), normalize(), relayPolicy() (+6 more)

### Community 64 - "Run"
Cohesion: 0.10
Nodes (25): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+17 more)

### Community 65 - "net/netip.Addr"
Cohesion: 0.07
Nodes (28): network, hostAddrs, nameTable, zone, TestIsPublic(), gateway(), host(), proxy() (+20 more)

### Community 66 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 67 - "UI"
Cohesion: 0.10
Nodes (18): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize() (+10 more)

### Community 68 - "serve"
Cohesion: 0.11
Nodes (38): brokenStream, countingDB, ErrorBody, harness, joinBody(), postJoin(), serve(), signed() (+30 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.06
Nodes (69): Determinism and goldens, Options, ticketServer, `internal/proto`, Join(), New(), TestJoinIsUnsigned(), testSession() (+61 more)

### Community 70 - "Security"
Cohesion: 0.14
Nodes (14): A node's agent as a Swarm service, Changes a node asks for, Container-mode agent, Contents, DNS, Interfaces the agent did not create, Security, Several meshes on one host (+6 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.14
Nodes (20): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+12 more)

### Community 72 - "embed.go"
Cohesion: 0.41
Nodes (10): InstallParams, RenderInstall(), RenderInstallPS1(), renderPS1(), RenderReinstall(), RenderReinstallPS1(), renderShell(), RenderUninstall() (+2 more)

### Community 73 - "rulesFor"
Cohesion: 0.36
Nodes (11): nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey(), rootKey() (+3 more)

### Community 74 - "net/http.Request"
Cohesion: 0.09
Nodes (11): Session, listOf(), meshChanged(), atoi(), before(), changed(), versionNumber(), message() (+3 more)

### Community 75 - "samples_test.go"
Cohesion: 0.20
Nodes (15): SampleParams, Sample, NodeSamples(), docWithSamples(), parseSample(), TestContainersDocSamples(), TestNodeSamples(), TestSampleImage() (+7 more)

### Community 76 - "Containers"
Cohesion: 0.14
Nodes (14): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, 4. Docker Desktop on Windows, A client node on the host network, A host-mode node's containers, Choosing how to run each node, Container mode (+6 more)

### Community 77 - "Sync"
Cohesion: 0.10
Nodes (19): Policy, mutate(), Desired(), Policy, Route, Rule, Priorities(), SlotOf() (+11 more)

### Community 78 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "sync.Mutex"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "Node"
Cohesion: 0.07
Nodes (50): apiError, CopyInfo, NodeList, PhoneProfile, PhoneRequest, addresses(), heard(), Role (+42 more)

### Community 82 - ".Set"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - "TestPageGoldens"
Cohesion: 0.07
Nodes (30): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+22 more)

### Community 87 - "File"
Cohesion: 0.11
Nodes (18): agentLog(), TestPrintEnrollmentInTheLocalZone(), printExit(), main(), printMesh(), printChange(), printEnrollment(), ttyPassword() (+10 more)

### Community 90 - "Host"
Cohesion: 0.05
Nodes (39): link, exitHost, Host, TestLinksAddressesAndFailures(), AddrPort(), TestAddrPort(), hexKey(), ParseUAPI() (+31 more)

### Community 93 - "exprs"
Cohesion: 0.17
Nodes (8): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), TestIntervalAddrElems(), TestIntervalElems(), nftOp

### Community 94 - "Ruleset"
Cohesion: 0.24
Nodes (3): Kernel, Ruleset, Chain

### Community 96 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.10
Nodes (5): NewUserspace(), adapterGUID(), defaultInterface(), Link, userspace

### Community 99 - "TestOneRunningCopy"
Cohesion: 0.50
Nodes (3): lab, staleTicket(), TestOneRunningCopy()

### Community 100 - ".nodeView"
Cohesion: 0.13
Nodes (12): Health(), phoneHealth(), SharesLine(), ago(), bytesize(), writeHTML(), TestHelpers(), dashboardData (+4 more)

### Community 101 - "net.Listener"
Cohesion: 0.33
Nodes (6): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), serve()

### Community 102 - "run"
Cohesion: 0.11
Nodes (28): agentCommands(), argSynopsis(), commands(), newFlags(), parseArgs(), parseFlags(), printFlags(), setupVersion() (+20 more)

### Community 103 - "Registration"
Cohesion: 0.21
Nodes (7): Domain, fakeRegistrar, Registrar, Registration, NewRegistrar(), NewRegistrar(), NewRegistrar()

### Community 104 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

### Community 105 - "sync.Mutex"
Cohesion: 0.04
Nodes (26): syncBuffer, Change, Failover, Hub, hubMove, hubState, subscription, Deps (+18 more)

### Community 106 - "Service"
Cohesion: 0.05
Nodes (36): replayCache, Access, Config, DB, failures, Lockout, LockoutError, Pending (+28 more)

### Community 107 - "contents_test.go"
Cohesion: 0.57
Nodes (6): docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents()

### Community 108 - "meshctl"
Cohesion: 0.29
Nodes (7): Building, Contents, Documentation, How it fits together, meshctl, Quick start, Status

### Community 109 - "8. Server"
Cohesion: 0.33
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 111 - "Developer guide"
Cohesion: 0.33
Nodes (6): Contents, Dev script tasks, Developer guide, Goldens, Releases, Toolchain

### Community 112 - "Of"
Cohesion: 0.20
Nodes (9): fakeRuntime, Container, Of(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack(), TestOfWithIPv6(), ContainerName (+1 more)

### Community 113 - "TestContainerNetworks"
Cohesion: 0.33
Nodes (5): lab, TestContainerNetworks(), ns(), lab, TestRelayFailover()

### Community 114 - "fail"
Cohesion: 0.07
Nodes (40): Security invariants, `internal/validate`, checkDNS(), TestTimestamp(), TestTimeZone(), Timestamp(), TimeZone(), TestRef() (+32 more)

### Community 115 - "ServerURL"
Cohesion: 0.25
Nodes (6): HelloRequest, Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase(), TestServerURL()

### Community 116 - "build"
Cohesion: 0.50
Nodes (3): build(), TestBuildBatch(), nftBatch

### Community 117 - "copy"
Cohesion: 0.70
Nodes (4): add(), copy(), done(), select()

### Community 119 - "walk"
Cohesion: 0.83
Nodes (3): admit(), walk(), packet

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.07
Nodes (45): Contents, `internal/server` and `internal/server/export`, `internal/tools/vulnjudge`, Test catalogue, TestRead(), FromSnapshot(), Document, dualSnapshot() (+37 more)

### Community 122 - "Open"
Cohesion: 0.50
Nodes (3): TestOpen(), Open(), TestOpenRefusesUnknownBackends()

### Community 130 - "Slot"
Cohesion: 0.11
Nodes (16): Pipe(), TestPolicyHasKernelSemantics(), Holder(), Taken(), TestHolder(), NewKernel(), NewKernel(), adapterPart() (+8 more)

### Community 151 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 396 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **29 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `Slot`, `ToSnapshot`, `Compile`, `Test catalogue`, `serveRuntime`, `fakeServer`, `Dir`, `Agent (`internal/agent/...`)`, `Run`, `net/netip.Prefix`, `Policy`, `TestSpokesAndExit`, `.mesh`, `DNS`, `context.Context`, `agent/agent_test.go`, ``internal/server/auth``, `netlinkNet`, `NodeConfig`, `Node`, `netlinkPolicy`, `LocalStatus`, `newRig`, `meshDNS`, `kernel`, `Sync`, `Run`, `net/netip.Addr`, `UI`, ``internal/proto``, `fetchwintun/main_test.go`, `rulesFor`, `Sync`, `sync.Mutex`, `.Set`, `File`, `Host`, `exprs`, `Service`, `Of`, `build`, ``internal/server` and `internal/server/export``, `Open`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09674295881192432 - nodes in this community are weakly interconnected._
- **Why does `NodeConfig` connect `NodeConfig` to `ui/ui_test.go`, `WGKey`, `ToSnapshot`, `fakeServer`, `Exit`, `Dir`, `Policy`, `context.Context`, `agent/agent_test.go`, `planner_test.go`, `Node`, `LocalStatus`, ``internal/server/api``, `Client`, `context.Context`, `Sync`, ``internal/proto``, `Node`, `Host`, `sync.Mutex`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.13062409288824384 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `kernel`, `.mesh`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._