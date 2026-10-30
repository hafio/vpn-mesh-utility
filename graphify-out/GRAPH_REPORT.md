# Graph Report - vpn-mesh-utility  (2026-10-30)

## Corpus Check
- 331 files · ~409,576 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 10 file(s) not represented in the graph (top: (none) 5, .tmpl 2, .css 2)

## Summary
- 4072 nodes · 21991 edges · 142 communities (112 shown, 30 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2584 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `a2fc8eee`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- Name
- newHarness
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
- otherNet
- net/http.ResponseWriter
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
- testing.T
- `internal/server/auth`
- testing.T
- log/slog.Logger
- .exec
- `internal/server/store`
- nullInt
- netlinkNet
- Net
- planner_test.go
- Node
- `internal/server/api`
- README.md
- Host
- .query
- Test catalogue
- serve
- LocalStatus
- storeNode
- errorf
- time.Time
- Client
- Runtime
- Engine
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
- time.Duration
- fetchwintun/main_test.go
- .setup
- .AddMesh
- ui
- Exporter
- rulesFor
- Sync
- .readJSON
- net/netip.Addr
- resolved
- Segmentation
- HashPassword
- io.Reader
- .renderPolicies
- deploy_test.go
- Open
- io.Writer
- Host
- exprs
- Ruleset
- .BeginEnrollment
- Agent protocol
- userspace
- ns
- .do
- stream
- run
- Extending the code
- Schedule
- hostid.go
- context.Context
- Allocate
- DefaultPolicy
- nftKernel
- StepError
- Diff
- .Sync
- Engine
- IPv6View
- .Reach
- loadMigrations
- Sysctl
- wgdev
- .settings
- DNSView
- `internal/server` and `internal/server/export`
- noKernel
- build
- walk
- Segmentation
- TestMain
- Exit
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
- `A node's container on the host network (the NAS)` --references--> `TestContainerSamples()`  [INFERRED]
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

## Communities (142 total, 30 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (19): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+11 more)

### Community 1 - "WGKey"
Cohesion: 0.13
Nodes (43): rig, TestProxyLoopAppliesAgain(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), IsStepError(), New(), failed() (+35 more)

### Community 2 - "Name"
Cohesion: 0.13
Nodes (36): check(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), Endpoint6(), IP(), IPv6() (+28 more)

### Community 3 - "newHarness"
Cohesion: 0.12
Nodes (29): adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing(), TestAdminSetClearsReferences() (+21 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (116): Status, 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names (+108 more)

### Community 5 - "env"
Cohesion: 0.12
Nodes (50): checkMachineIDFile(), joinFlags(), meshDir(), meshFlags(), onlyMesh(), sep(), setupAgentJoin(), setupAgentReset() (+42 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.15
Nodes (15): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), parseSpans(), PrefixRange(), Ref() (+7 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.13
Nodes (17): Rule, Violation, graphify, Packages, and which way they may import, Security invariants, Style, What this repo is, You do not run the gates (+9 more)

### Community 11 - "net.Listener"
Cohesion: 0.32
Nodes (6): Listen(), Listen(), listenSocket(), ListenAdmin(), serve(), TestListenAdminRefusesALongPath()

### Community 12 - "Policy"
Cohesion: 0.13
Nodes (28): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+20 more)

### Community 14 - ".queryRow"
Cohesion: 0.10
Nodes (5): AgentSession, HashToken(), NewToken(), Machine, Refusal

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.10
Nodes (15): api, fakeServer, signed(), decode(), reply(), Definitive(), hostFamilies(), decodeJSON() (+7 more)

### Community 17 - "serveRuntime"
Cohesion: 0.18
Nodes (15): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+7 more)

### Community 18 - "New"
Cohesion: 0.19
Nodes (7): newRenderer(), New(), sameOrigin(), handler, meshHandler, page, renderer

### Community 19 - "Dir"
Cohesion: 0.06
Nodes (40): meshKey(), TestDefaultSocket(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, Enrollment, NewKeys(), newSeed() (+32 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.16
Nodes (33): fakeRunner, TestRunNeedsAnEnrollment(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+25 more)

### Community 21 - "Run"
Cohesion: 0.12
Nodes (26): TestCleanHostname(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), Local(), Presented(), TestHashes(), cleanHostname() (+18 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (22): network, Routes, syncDropped(), index(), NewTable(), route(), compare(), Table (+14 more)

### Community 24 - ".nodeView"
Cohesion: 0.12
Nodes (13): CopyInfo, CopyLine(), Health(), phoneHealth(), ago(), bytesize(), writeHTML(), TestHelpers() (+5 more)

### Community 25 - "Input"
Cohesion: 0.09
Nodes (26): PlanError, Record, activeAt(), zoneOf(), containers6(), Check(), checkContainers(), checkContainers6() (+18 more)

### Community 26 - "lab"
Cohesion: 0.20
Nodes (6): Coverage, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "flag.FlagSet"
Cohesion: 0.13
Nodes (37): machineIDFlag(), TestBytesize(), printExit(), setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), printMesh() (+29 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (41): Command reference, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit, meshctl export (+33 more)

### Community 30 - ".mesh"
Cohesion: 0.13
Nodes (17): Container names, 10. Backups, 11. When the control plane is down, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits (+9 more)

### Community 31 - "NodeConfig"
Cohesion: 0.06
Nodes (36): IPv6Host, Determinism and goldens, forHost(), proxied(), TestForHostWithoutIPv6(), intervalElems(), TestIntervalElems(), Desired() (+28 more)

### Community 32 - "context.Context"
Cohesion: 0.07
Nodes (15): Deps, runner, Deps, kick(), oneLine(), Reset(), Run(), uiListen() (+7 more)

### Community 33 - "newRig"
Cohesion: 0.10
Nodes (42): The web UI's pages (internal/server/ui), `internal/server/ui`, capped, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages() (+34 more)

### Community 34 - "Service"
Cohesion: 0.11
Nodes (12): dialing, Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar() (+4 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.09
Nodes (69): rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime(), newFakeServer() (+61 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (135): Agent (`internal/agent/...`), `internal/fsutil`, freePort(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), pfx(), TestConfigureHasKernelSemantics() (+127 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (30): rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword() (+22 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (52): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+44 more)

### Community 39 - "log/slog.Logger"
Cohesion: 0.13
Nodes (10): Hub, subscription, blockingHost, ConfigEvent, NewHub(), closed(), TestHubDeliversChallenges(), TestHubLatestWins() (+2 more)

### Community 40 - ".exec"
Cohesion: 0.08
Nodes (8): newUUID(), mustAffect(), joinList(), boolInt(), classify(), changed(), Session, StatusUpdate

### Community 41 - "`internal/server/store`"
Cohesion: 0.13
Nodes (44): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+36 more)

### Community 42 - "nullInt"
Cohesion: 0.12
Nodes (4): DNSRecord, nullInt(), nullStr(), Version

### Community 43 - "netlinkNet"
Cohesion: 0.09
Nodes (13): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+5 more)

### Community 44 - "Net"
Cohesion: 0.16
Nodes (6): ensureAddr(), Net, Rule, Remove(), syncProxies(), syncRules()

### Community 45 - "planner_test.go"
Cohesion: 0.14
Nodes (40): Add a topology rule, `internal/server/planner`, TestCheckRequested(), deliver(), deliverAddr(), encode(), firstDiff(), hasProblem() (+32 more)

### Community 46 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 47 - "`internal/server/api`"
Cohesion: 0.12
Nodes (36): apiError, `internal/server/api`, cli(), ErrorMessage(), toAPIError(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries() (+28 more)

### Community 48 - "README.md"
Cohesion: 0.23
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "Host"
Cohesion: 0.31
Nodes (3): Host, SyncAddrs(), SyncLink()

### Community 50 - ".query"
Cohesion: 0.13
Nodes (5): ContainerName, scanAll(), splitList(), AuditEntry, PeerStat

### Community 51 - "Test catalogue"
Cohesion: 0.17
Nodes (18): `internal/server/phone`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue, Hub(), joinSet(), NewKey(), Render() (+10 more)

### Community 52 - "serve"
Cohesion: 0.18
Nodes (21): brokenStream, TestSessionFiles(), ErrorBody, TestSessionShapes(), SignSession(), TestVerifyRejectsTampering(), TestEventsStreamPushesChanges(), harness (+13 more)

### Community 53 - "LocalStatus"
Cohesion: 0.10
Nodes (21): Add a web UI page, Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest (+13 more)

### Community 54 - "storeNode"
Cohesion: 0.11
Nodes (39): countingDB, harness, joinBody(), postJoin(), serve(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash() (+31 more)

### Community 55 - "errorf"
Cohesion: 0.06
Nodes (50): hubMove, hubState, NameSet, NodeAddRequest, PhoneRequest, Role, blockOf(), checkGroups() (+42 more)

### Community 56 - "time.Time"
Cohesion: 0.08
Nodes (15): syncBuffer, replayCache, hitWindow, LockoutError, serves(), CanonicalRequest(), Signed, RateLimiter (+7 more)

### Community 57 - "Client"
Cohesion: 0.11
Nodes (11): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, HelloRequest, HelloResponse (+3 more)

### Community 58 - "Runtime"
Cohesion: 0.15
Nodes (7): fakeRuntime, Container, clip(), meshFilter(), ContainerName, NamesRequest, Runtime

### Community 59 - "Engine"
Cohesion: 0.21
Nodes (5): Rechecker, Add an apply step, Engine, Result, PeerStatus

### Community 60 - "meshDNS"
Cohesion: 0.20
Nodes (21): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters() (+13 more)

### Community 61 - "`internal/validate`"
Cohesion: 0.12
Nodes (29): `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), FuzzArg(), FuzzName(), TestArg() (+21 more)

### Community 62 - "Sync"
Cohesion: 0.16
Nodes (31): Sysctls, NewNet(), TestTeardownFakes(), Gateway(), Guard(), Owned(), parse(), SlotOf() (+23 more)

### Community 63 - ".Apply"
Cohesion: 0.12
Nodes (15): applied, Firewall, policy, Firewall, Runner, atLeast(), exitPolicy(), normalize() (+7 more)

### Community 64 - "Run"
Cohesion: 0.08
Nodes (28): Deps, Route, route, Add an agent API route, NewUserspace(), NewUserspace(), NewNodes(), AdminRoutes() (+20 more)

### Community 65 - "AddrRange"
Cohesion: 0.20
Nodes (13): addrs(), contains(), ofFamily(), inRanges(), inRange(), AddrEntry(), AddrRange, LastOf() (+5 more)

### Community 66 - ".meshctl"
Cohesion: 0.17
Nodes (15): 8. The web UI, Each node's local UI, The mesh firewall, A node's container on the host network (the NAS), Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create (+7 more)

### Community 67 - "UI"
Cohesion: 0.09
Nodes (20): hostOf(), Listen(), New(), oldest(), ParseAddr(), TestServeAndListen(), validOrNot(), ago() (+12 more)

### Community 68 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.05
Nodes (65): Options, ticketServer, `internal/proto`, Join(), New(), newClient(), sign(), TestAnswersMustBeSigned() (+57 more)

### Community 70 - "time.Duration"
Cohesion: 0.13
Nodes (9): pendingApply, safeApply, Failover, waitApplied(), 12. Troubleshooting, Exec, TestSafeApplyStateMachine(), TestJudgeHubs() (+1 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.24
Nodes (13): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), TestDownload(), TestExtractChecksTheDLL() (+5 more)

### Community 72 - ".setup"
Cohesion: 0.19
Nodes (5): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), group()

### Community 73 - ".AddMesh"
Cohesion: 0.10
Nodes (23): Change, MeshList, meshSpace, MeshUpdate, SoleMesh(), checkApart(), clash(), freeCIDR() (+15 more)

### Community 74 - "ui"
Cohesion: 0.09
Nodes (12): Session, listOf(), meshChanged(), title(), atoi(), before(), changed(), versionNumber() (+4 more)

### Community 75 - "Exporter"
Cohesion: 0.16
Nodes (11): DB, NextExport(), TestExporterRunExportsAtNight(), TestExporterRunStopsWithItsContext(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestNextExport(), TestRunGivesTheMeshIPv6(), TestRunLogsTheSetupCode() (+3 more)

### Community 76 - "rulesFor"
Cohesion: 0.18
Nodes (18): nrpt, nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey() (+10 more)

### Community 77 - "Sync"
Cohesion: 0.07
Nodes (16): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Policy, Route, Rule (+8 more)

### Community 78 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 79 - "net/netip.Addr"
Cohesion: 0.09
Nodes (24): hostAddrs, nameTable, zone, Apply (agent), TestIsPublic(), gateway(), host(), exchange() (+16 more)

### Community 80 - "resolved"
Cohesion: 0.20
Nodes (8): Domain, fakeRegistrar, linkDomain, Registration, resolved, TestServiceRevert(), classify(), foreignNetns()

### Community 81 - "Segmentation"
Cohesion: 0.10
Nodes (25): RecordRequest, segLiteral, segDoc, addIPv6(), Actor, problemStrings(), segLiterals(), errorf() (+17 more)

### Community 82 - "HashPassword"
Cohesion: 0.36
Nodes (9): Config, DB, Params, DefaultParams(), HashPassword(), TestHashUsesTheSpecParameters(), TestVerifyPassword(), TestVerifyRefusesMalformedHashes() (+1 more)

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - ".renderPolicies"
Cohesion: 0.06
Nodes (30): SegObject, SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm() (+22 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.11
Nodes (34): block, checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth() (+26 more)

### Community 86 - "Open"
Cohesion: 0.18
Nodes (11): TestExportOnceFails(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), insertAt5(), openAt(), Store, Open(), TestOpenIsIdempotent() (+3 more)

### Community 87 - "io.Writer"
Cohesion: 0.18
Nodes (9): agentLog(), main(), ttyPassword(), OpenFile(), lockFile(), unlockFile(), lockFile(), unlockFile() (+1 more)

### Community 90 - "Host"
Cohesion: 0.09
Nodes (16): link, exitHost, Host, peer(), hexKey(), ParseUAPI(), uapiInt(), compareKeys() (+8 more)

### Community 93 - "exprs"
Cohesion: 0.24
Nodes (5): exprs(), ifname(), intervalAddrElems(), TestExprs(), nftOp

### Community 94 - "Ruleset"
Cohesion: 0.19
Nodes (4): watchingFirewall, Kernel, Ruleset, Chain

### Community 95 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 96 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Errors, Join, Local API, Mesh view, Prefs, Routes, Sessions (+2 more)

### Community 98 - "userspace"
Cohesion: 0.10
Nodes (6): adapterPart(), NewUserspace(), adapterGUID(), defaultInterface(), Link, userspace

### Community 99 - "ns"
Cohesion: 0.09
Nodes (22): Several meshes on one host, End to end -- `test/e2e`, host, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces() (+14 more)

### Community 100 - ".do"
Cohesion: 0.09
Nodes (22): MeshSetRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, SegmentView, SettingsRequest, statusDetail (+14 more)

### Community 101 - "stream"
Cohesion: 0.33
Nodes (4): dial(), dial(), isPipe(), stream

### Community 102 - "run"
Cohesion: 0.10
Nodes (36): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+28 more)

### Community 103 - "Extending the code"
Cohesion: 0.22
Nodes (9): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+1 more)

### Community 104 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "hostid.go"
Cohesion: 0.39
Nodes (7): source, clean(), first(), fromFile(), sources(), Read(), sources()

### Community 106 - "context.Context"
Cohesion: 0.09
Nodes (21): Access, failures, Lockout, Pending, resetTOTP(), fullExpiry(), Service, insertUser() (+13 more)

### Community 107 - "Allocate"
Cohesion: 0.36
Nodes (9): Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), Range, TestAllocateContainers(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan() (+1 more)

### Community 108 - "DefaultPolicy"
Cohesion: 0.36
Nodes (8): DefaultPolicy(), LegacySegmentation(), TestMigrationTurnsRulesIntoPolicies(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSetSegmentationRefusesAStranger(), FirewallRule

### Community 111 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 113 - "Engine"
Cohesion: 0.17
Nodes (4): nameServer, DNS, Want, MeshName

### Community 114 - "IPv6View"
Cohesion: 0.38
Nodes (3): IPv6Request, IPv6View, ipv6View()

### Community 115 - ".Reach"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 116 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 117 - "Sysctl"
Cohesion: 0.33
Nodes (3): Sysctl, Sysctls, Sysctl

### Community 119 - ".settings"
Cohesion: 0.25
Nodes (8): Container-mode agent, meshCIDR(), CheckServerDomains(), TestCheckServerDomains(), checkDNS(), DNSName(), Domains(), Zone()

### Community 120 - "DNSView"
Cohesion: 0.40
Nodes (5): DNSEntry, NameEntry, DNSView, dnsData, recordForm

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.14
Nodes (25): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+17 more)

### Community 124 - "walk"
Cohesion: 0.83
Nodes (3): admit(), walk(), packet

### Community 130 - "Slot"
Cohesion: 0.11
Nodes (16): Pipe(), Holder(), Priorities(), Taken(), TestHolder(), TestSlotOf(), NewKernel(), NewKernel() (+8 more)

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 382 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **30 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `Slot`, `archtest.go`, `Policy`, `serveRuntime`, `Dir`, `Agent (`internal/agent/...`)`, `Run`, `.mesh`, `NodeConfig`, `agent/agent_test.go`, `Node`, `Test catalogue`, `serve`, `time.Time`, `Engine`, `meshDNS`, `Sync`, `Run`, `.meshctl`, `UI`, ``internal/proto``, `time.Duration`, `rulesFor`, `net/netip.Addr`, `resolved`, `Host`, `exprs`, `Engine`?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.0957020612193026 - nodes in this community are weakly interconnected._
- **Why does `Slot` connect `Slot` to `ui/ui_test.go`, `Name`, `net/http.ResponseWriter`, `serveRuntime`, `Dir`, `Run`, `Input`, `.mesh`, `context.Context`, `Service`, `testing.T`, `netlinkNet`, `Net`, `errorf`, `Sync`, `.Apply`, ``internal/proto``, `rulesFor`, `Sync`, `net/netip.Addr`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.13429951690821257 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `.mesh`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._