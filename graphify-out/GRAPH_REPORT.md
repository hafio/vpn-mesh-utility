# Graph Report - vpn-mesh-utility  (2026-10-30)

## Corpus Check
- 332 files · ~410,643 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 10 file(s) not represented in the graph (top: (none) 5, .tmpl 2, .css 2)

## Summary
- 4079 nodes · 22004 edges · 146 communities (114 shown, 32 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2586 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `480ca8b2`
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
- net.Listener
- Policy
- github.com/hafio/vpn-mesh-utility
- .queryRow
- net/http.ResponseWriter
- api
- serveRuntime
- New
- Dir
- Agent (`internal/agent/...`)
- Run
- net/netip.Prefix
- Net
- .nodeView
- Input
- lab
- Snapshot
- flag.FlagSet
- Command reference
- .mesh
- NodeConfig
- context.Context
- newRig
- Service
- agent/agent_test.go
- Agent (`internal/agent/...`)
- `internal/server/auth`
- testing.T
- newRig
- .exec
- `internal/server/store`
- nullInt
- netlinkNet
- Net
- planner_test.go
- Node
- errorf
- README.md
- samples_test.go
- .query
- Test catalogue
- Session
- LocalStatus
- `internal/server/api`
- MeshTx
- Matcher
- Client
- Of
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
- net/http.Request
- PublishNames
- ui
- Exporter
- rulesFor
- Sync
- io.Writer
- net/netip.Addr
- resolved
- Segmentation
- .Set
- io.Reader
- .renderPolicies
- deploy_test.go
- ServerURL
- io.Writer
- Host
- exprs
- Ruleset
- .routes
- Agent protocol
- userspace
- ns
- .do
- stream
- run
- Extending the code
- Schedule
- TestSafeApplyStateMachine
- context.Context
- Allocate
- Table
- nftKernel
- StepError
- Diff
- .Sync
- Engine
- .tx
- userRun
- loadMigrations
- Run
- wgdev
- Name
- forms.js
- `internal/server` and `internal/server/export`
- noKernel
- build
- Version
- LoadInput
- TestMain
- .SetSegmentation
- entry
- AdminError
- Slot
- Error
- NewNet
- NewNet
- service_windows_test.go
- NewKernel
- Refusal
- NewKernel
- Refusal
- usageError
- StatusError
- AliveRequest
- TestQRCode
- 17. Amendments
- PlanError
- Error

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 304 edges
2. `ui` - 109 edges
3. ``internal/server/api`` - 101 edges
4. `ServerConfig()` - 89 edges
5. `Session` - 85 edges
6. `Slot` - 82 edges
7. `newHarness()` - 73 edges
8. `NodeConfig` - 69 edges
9. `env` - 64 edges
10. `api` - 58 edges

## Surprising Connections (you probably didn't know these)
- `Dev script tasks` --references--> `build()`  [INFERRED]
  docs/developer.md → internal/agent/policy/nft_linux.go
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

## Communities (146 total, 32 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (19): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+11 more)

### Community 1 - "WGKey"
Cohesion: 0.15
Nodes (44): blockingHost, rig, watchingFirewall, TestProxyLoopAppliesAgain(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), IsStepError() (+36 more)

### Community 2 - "net/netip.AddrPort"
Cohesion: 0.18
Nodes (18): check(), ParseAddr(), TestParseAddr(), newNode(), validateSet(), Endpoint6(), IPv6(), PeerEndpoint() (+10 more)

### Community 3 - "IPNet"
Cohesion: 0.10
Nodes (12): Status, IPNet(), Prefix(), TestPrefixRoundTrip(), TestPrefixUnmapsIPv4(), index(), route(), TestCreateLinkNeedsPrivilege() (+4 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (114): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+106 more)

### Community 5 - "env"
Cohesion: 0.13
Nodes (44): checkMachineIDFile(), joinFlags(), machineIDFlag(), meshDir(), meshFlags(), onlyMesh(), sep(), setupAgentJoin() (+36 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (26): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+18 more)

### Community 7 - "segResolver"
Cohesion: 0.13
Nodes (18): Service, activeAt(), zoneOf(), checkSegmentation(), hosts(), mergePorts(), mergeServices(), newSegResolver() (+10 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.13
Nodes (17): Rule, Violation, Determinism and goldens, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates (+9 more)

### Community 11 - "net.Listener"
Cohesion: 0.21
Nodes (12): bindBoth(), bindPair(), Listen(), Listen(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly() (+4 more)

### Community 12 - "Policy"
Cohesion: 0.17
Nodes (23): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+15 more)

### Community 14 - ".queryRow"
Cohesion: 0.10
Nodes (5): AgentSession, HashToken(), NewToken(), Machine, Refusal

### Community 15 - "net/http.ResponseWriter"
Cohesion: 0.20
Nodes (6): fakeServer, signed(), decode(), reply(), readSums(), writeHTML()

### Community 16 - "api"
Cohesion: 0.14
Nodes (3): api, Nodes, hostFamilies()

### Community 17 - "serveRuntime"
Cohesion: 0.13
Nodes (17): TestOpenRuntime(), clip(), Detect(), meshFilter(), New(), Sockets(), inspectOf(), serveRuntime() (+9 more)

### Community 18 - "New"
Cohesion: 0.13
Nodes (12): loadPages(), newRenderer(), staticHandler(), headers(), New(), sameOrigin(), TestEveryPageRenders(), TestFormsShowWhatApplies() (+4 more)

### Community 19 - "Dir"
Cohesion: 0.06
Nodes (31): TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, NewKeys(), newSeed(), newWGKey(), mkdirs(), TestCachesAreCheckedWithThePinnedKey() (+23 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.18
Nodes (31): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+23 more)

### Community 21 - "Run"
Cohesion: 0.08
Nodes (37): source, TestCleanHostname(), TestIsPublic(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), clean(), first() (+29 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (12): otherNet, Routes, compare(), luidOf(), onLink(), Allocate6(), AllocateContainers6(), Block6() (+4 more)

### Community 24 - ".nodeView"
Cohesion: 0.16
Nodes (9): group(), ago(), bytesize(), TestHelpers(), dashboardData, nodeForm, nodeRow, nodesData (+1 more)

### Community 25 - "Input"
Cohesion: 0.13
Nodes (13): Routes, Record, containers6(), Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable() (+5 more)

### Community 26 - "lab"
Cohesion: 0.18
Nodes (7): Coverage, host, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.15
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "flag.FlagSet"
Cohesion: 0.23
Nodes (28): setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), setupMeshAdd(), setupMeshList(), setupMeshRemove(), setupMeshSet() (+20 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (41): Command reference, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit, meshctl export (+33 more)

### Community 30 - ".mesh"
Cohesion: 0.11
Nodes (20): 10. Backups, 11. When the control plane is down, 12. Troubleshooting, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 5. Day to day (+12 more)

### Community 31 - "NodeConfig"
Cohesion: 0.08
Nodes (22): dialing, IPv6Host, Rechecker, Add an apply step, Shared fixtures -- `internal/testfix`, forHost(), Engine, Result (+14 more)

### Community 32 - "context.Context"
Cohesion: 0.08
Nodes (10): runner, kick(), oneLine(), TestOneLine(), TestReportKeepsToTheServersLimits(), DNS, drain(), RunWindowsService() (+2 more)

### Community 33 - "newRig"
Cohesion: 0.14
Nodes (36): The web UI's pages (internal/server/ui), `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages(), TestPhonePage() (+28 more)

### Community 34 - "Service"
Cohesion: 0.12
Nodes (11): Registrar, resolver, Service, 4. Spokes and exits, Container mode, New(), NewRegistrar(), NewRegistrar() (+3 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (72): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), newFakeRuntime() (+64 more)

### Community 36 - "Agent (`internal/agent/...`)"
Cohesion: 0.05
Nodes (63): Agent (`internal/agent/...`), capped, TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx(), TestConfigureHasKernelSemantics(), TestKernelHoldsOneTable(), TestNetHasKernelSemantics() (+55 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (32): rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword() (+24 more)

### Community 38 - "testing.T"
Cohesion: 0.08
Nodes (57): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+49 more)

### Community 39 - "newRig"
Cohesion: 0.24
Nodes (11): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+3 more)

### Community 40 - ".exec"
Cohesion: 0.09
Nodes (7): newUUID(), mustAffect(), boolInt(), classify(), changed(), Session, StatusUpdate

### Community 41 - "`internal/server/store`"
Cohesion: 0.09
Nodes (61): `internal/server/store`, TestReconcilerVersionsOnlyRealChanges(), TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestMigrationMakesMeshOne() (+53 more)

### Community 42 - "nullInt"
Cohesion: 0.16
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 43 - "netlinkNet"
Cohesion: 0.14
Nodes (9): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+1 more)

### Community 44 - "Net"
Cohesion: 0.23
Nodes (5): ensureAddr(), Net, Rule, syncDropped(), syncRules()

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (46): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), Problem, Compute(), deliver(), deliverAddr() (+38 more)

### Community 46 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 47 - "errorf"
Cohesion: 0.09
Nodes (30): apiError, Role, Signed, addIPv6(), blockOf(), containerSubnet6(), moveMeshIP6(), setContainers6() (+22 more)

### Community 48 - "README.md"
Cohesion: 0.23
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "samples_test.go"
Cohesion: 0.25
Nodes (11): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), mountsMachineID(), sampleVolume() (+3 more)

### Community 50 - ".query"
Cohesion: 0.10
Nodes (12): SoleMesh(), freeSlot(), PickMesh(), pickMesh(), Mesh, scanMesh(), ContainerName, scanAll() (+4 more)

### Community 51 - "Test catalogue"
Cohesion: 0.19
Nodes (17): `internal/server/phone`, `internal/tools/vulnjudge`, Test catalogue, Hub(), joinSet(), NewKey(), Render(), aliceKey() (+9 more)

### Community 52 - "Session"
Cohesion: 0.09
Nodes (47): Options, ticketServer, Join(), New(), TestJoinIsUnsigned(), testSession(), TestUnreachableAndMalformedAnswers(), newTicketServer() (+39 more)

### Community 53 - "LocalStatus"
Cohesion: 0.08
Nodes (27): Add a web UI page, Client, Logs, Handler(), logs(), refusal(), Listen(), New() (+19 more)

### Community 54 - "`internal/server/api`"
Cohesion: 0.05
Nodes (110): brokenStream, countingDB, `internal/server/api`, ErrorBody, cli(), ErrorMessage(), adminCall(), TestAdminAddIssuesAToken() (+102 more)

### Community 55 - "MeshTx"
Cohesion: 0.15
Nodes (15): NodeAddRequest, NodeSetRequest, segLiteral, checkGroups(), checkName(), containerSubnet(), defaultMTU(), moveMeshIP() (+7 more)

### Community 56 - "Matcher"
Cohesion: 0.29
Nodes (4): serves(), Service, Matcher, matchRule

### Community 57 - "Client"
Cohesion: 0.13
Nodes (9): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, HelloResponse, ValidChallenge() (+1 more)

### Community 58 - "Of"
Cohesion: 0.15
Nodes (10): fakeRuntime, Container, Of(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack(), TestOfWithIPv6(), Source (+2 more)

### Community 59 - "Enum"
Cohesion: 0.21
Nodes (8): Ref(), ServiceEntry(), TestRef(), TestServiceEntry(), Enum(), TestEnum(), TestText(), Text()

### Community 60 - "meshDNS"
Cohesion: 0.20
Nodes (25): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+17 more)

### Community 61 - "`internal/validate`"
Cohesion: 0.08
Nodes (36): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), TimeZone(), TestAddrEntry(), TestAddrRangePrefix() (+28 more)

### Community 62 - "Sync"
Cohesion: 0.12
Nodes (33): network, Sysctls, NewNet(), TestTeardownFakes(), Gateway(), Guard(), Owned(), parse() (+25 more)

### Community 63 - ".Apply"
Cohesion: 0.13
Nodes (14): applied, Firewall, policy, Firewall, Runner, exitPolicy(), normalize(), relayPolicy() (+6 more)

### Community 64 - "Run"
Cohesion: 0.09
Nodes (24): Deps, NewNodes(), New(), NewAdmin(), newAPI(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries() (+16 more)

### Community 65 - "AddrRange"
Cohesion: 0.12
Nodes (22): addrs(), addrText(), contains(), Rule, merged(), ofFamily(), portText(), services() (+14 more)

### Community 66 - ".meshctl"
Cohesion: 0.19
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens (+6 more)

### Community 67 - "UI"
Cohesion: 0.14
Nodes (11): hostOf(), oldest(), validOrNot(), ago(), bytesize(), prefsFlash(), ULAPrefix(), handler (+3 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.06
Nodes (66): `internal/fsutil`, `internal/proto`, TestDNS(), newClient(), sign(), TestAnswersMustBeSigned(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode() (+58 more)

### Community 70 - "log/slog.Logger"
Cohesion: 0.06
Nodes (23): Change, Failover, Hub, hubMove, hubState, subscription, hitWindow, waitApplied() (+15 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.14
Nodes (18): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+10 more)

### Community 72 - "net/http.Request"
Cohesion: 0.16
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), message(), page

### Community 73 - "PublishNames"
Cohesion: 0.14
Nodes (16): DNSEntry, meshSpace, NameEntry, NameSet, checkApart(), clash(), freeCIDR(), spaceOf() (+8 more)

### Community 74 - "ui"
Cohesion: 0.12
Nodes (9): Session, listOf(), atoi(), before(), changed(), versionNumber(), actor(), settingsData (+1 more)

### Community 75 - "Exporter"
Cohesion: 0.28
Nodes (5): DB, NextExport(), TestExporterRunStopsWithItsContext(), TestNextExport(), Exporter

### Community 76 - "rulesFor"
Cohesion: 0.36
Nodes (11): nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey(), rootKey() (+3 more)

### Community 77 - "Sync"
Cohesion: 0.05
Nodes (27): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+19 more)

### Community 78 - "io.Writer"
Cohesion: 0.18
Nodes (11): dnsLine(), setupDNS(), printExit(), printUsage(), printMesh(), printEnrollment(), ago(), bytesize() (+3 more)

### Community 79 - "net/netip.Addr"
Cohesion: 0.09
Nodes (17): hostAddrs, nameTable, zone, Apply (agent), gateway(), host(), syncProxies(), records() (+9 more)

### Community 80 - "resolved"
Cohesion: 0.20
Nodes (8): Domain, fakeRegistrar, linkDomain, Registration, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "Segmentation"
Cohesion: 0.10
Nodes (19): RecordRequest, segDoc, Actor, problemStrings(), PolicyZone(), MeshChange, meshChange(), movePolicy() (+11 more)

### Community 82 - ".Set"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 83 - "io.Reader"
Cohesion: 0.29
Nodes (4): watchdog, Read(), ReadSegmentation(), TestSegmentationFile()

### Community 84 - ".renderPolicies"
Cohesion: 0.06
Nodes (31): SegPolicy, SegSchedule, anyText(), listText(), meshChanged(), newPolicyForm(), objectRows(), readPolicyForm() (+23 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.19
Nodes (24): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+16 more)

### Community 86 - "ServerURL"
Cohesion: 0.20
Nodes (8): Container-mode agent, HelloRequest, CheckServerDomains(), Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase(), TestServerURL()

### Community 87 - "io.Writer"
Cohesion: 0.16
Nodes (10): agentLog(), main(), ttyPassword(), OpenFile(), TestFileRotates(), lockFile(), unlockFile(), lockFile() (+2 more)

### Community 90 - "Host"
Cohesion: 0.05
Nodes (40): link, exitHost, Host, peer(), TestLinksAddressesAndFailures(), AddrPort(), TestAddrPort(), hexKey() (+32 more)

### Community 93 - "exprs"
Cohesion: 0.17
Nodes (8): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), TestIntervalAddrElems(), TestIntervalElems(), nftOp

### Community 94 - "Ruleset"
Cohesion: 0.24
Nodes (3): Kernel, Ruleset, Chain

### Community 95 - ".routes"
Cohesion: 0.42
Nodes (7): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 96 - "Agent protocol"
Cohesion: 0.22
Nodes (9): Agent protocol, Errors, Join, Local API, Mesh view, Prefs, Sessions, Signed requests (+1 more)

### Community 98 - "userspace"
Cohesion: 0.11
Nodes (5): Holder(), defaultInterface(), unspecified(), Link, userspace

### Community 99 - "ns"
Cohesion: 0.10
Nodes (21): Several meshes on one host, End to end -- `test/e2e`, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), inNS() (+13 more)

### Community 100 - ".do"
Cohesion: 0.11
Nodes (15): MeshList, MeshSetRequest, MeshUpdate, SegmentView, SettingsRequest, ZoneView, AdminClient, nodePath() (+7 more)

### Community 101 - "stream"
Cohesion: 0.33
Nodes (4): dial(), dial(), isPipe(), stream

### Community 102 - "run"
Cohesion: 0.12
Nodes (29): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+21 more)

### Community 103 - "Extending the code"
Cohesion: 0.22
Nodes (9): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+1 more)

### Community 104 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "TestSafeApplyStateMachine"
Cohesion: 0.46
Nodes (3): pendingApply, safeApply, TestSafeApplyStateMachine()

### Community 106 - "context.Context"
Cohesion: 0.05
Nodes (38): replayCache, Access, Config, DB, failures, Lockout, LockoutError, Params (+30 more)

### Community 107 - "Allocate"
Cohesion: 0.36
Nodes (9): Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), Range, TestAllocateContainers(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan() (+1 more)

### Community 108 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 111 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 113 - "Engine"
Cohesion: 0.10
Nodes (8): nameServer, syncBuffer, DNS, Sysctl, Want, MeshName, MeshNames(), syncBuffer

### Community 114 - ".tx"
Cohesion: 0.13
Nodes (8): IPv6Request, IPv6View, ReachRequest, ipv6View(), ReachResult, matrixCell, matrixRow, reachView

### Community 115 - "userRun"
Cohesion: 0.47
Nodes (6): setupUser(), setupUserAdd(), setupUserGrant(), userCommands(), userRun(), userAction

### Community 116 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 117 - "Run"
Cohesion: 0.12
Nodes (13): Deps, Deps, Container names, hostname(), meshKey(), Reset(), Run(), uiListen() (+5 more)

### Community 119 - "Name"
Cohesion: 0.20
Nodes (12): validateMeshNames(), meshCIDR(), ToSnapshot(), checkDNS(), IP(), DNSName(), Domains(), IPv4() (+4 more)

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.09
Nodes (36): `internal/server` and `internal/server/export`, TestRead(), FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport() (+28 more)

### Community 125 - "LoadInput"
Cohesion: 0.08
Nodes (28): CopyInfo, NodeDetail, NodeList, PhoneProfile, PhoneRequest, statusDetail, addresses(), heard() (+20 more)

### Community 130 - "Slot"
Cohesion: 0.11
Nodes (16): TestPolicyHasKernelSemantics(), Priorities(), SlotOf(), TestSlotOf(), NewKernel(), NewKernel(), adapterPart(), adapterSlot() (+8 more)

### Community 134 - "service_windows_test.go"
Cohesion: 0.12
Nodes (8): Name(), TestNoWindowsServiceHere(), RunWindowsService(), executes(), TestHandlerReportsAFailedRun(), TestHandlerStopsTheAgent(), TestNotAService(), handler

### Community 142 - "TestQRCode"
Cohesion: 0.50
Nodes (4): newQR(), TestQRCode(), phoneData, qrCode

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 382 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **32 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `WGKey`, `Slot`, `net/netip.AddrPort`, `IPNet`, `env`, `service_windows_test.go`, `archtest.go`, `net.Listener`, `Policy`, `serveRuntime`, `Dir`, `Agent (`internal/agent/...`)`, `Run`, `.mesh`, `NodeConfig`, `context.Context`, `agent/agent_test.go`, `newRig`, `Node`, `Test catalogue`, `Session`, `LocalStatus`, `Matcher`, `Of`, `meshDNS`, `Sync`, `.meshctl`, ``internal/proto``, `log/slog.Logger`, `fetchwintun/main_test.go`, `rulesFor`, `Sync`, `net/netip.Addr`, `resolved`, `.Set`, `io.Writer`, `Host`, `exprs`, `.routes`, `TestSafeApplyStateMachine`, `Engine`, ``internal/server` and `internal/server/export``?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09547437556929862 - nodes in this community are weakly interconnected._
- **Why does `Slot` connect `Slot` to `ui/ui_test.go`, `net/netip.AddrPort`, `env`, `17. Amendments`, `api`, `net/http.ResponseWriter`, `serveRuntime`, `Dir`, `Run`, `Input`, `.mesh`, `Service`, `Agent (`internal/agent/...`)`, `netlinkNet`, `Net`, `Session`, `LocalStatus`, `Sync`, `.Apply`, `rulesFor`, `Sync`, `net/netip.Addr`, `.Set`, `userspace`, `Run`, `LoadInput`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Should `IPNet` be split into smaller, more focused modules?**
  _Cohesion score 0.09788359788359788 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect ``internal/server/api`` to `Run`, `.do`, ``internal/proto``, `htmx.min.js`, ``internal/server/auth``, ``internal/server/store``, `errorf`, `Test catalogue`, `Session`, `.routes`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._