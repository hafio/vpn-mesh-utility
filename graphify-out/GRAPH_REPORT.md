# Graph Report - vpn-mesh-utility  (2026-10-26)

## Corpus Check
- 329 files · ~406,137 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 10 file(s) not represented in the graph (top: (none) 5, .tmpl 2, .css 2)

## Summary
- 4051 nodes · 21797 edges · 141 communities (114 shown, 27 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2560 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f2f215fa`
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
- net/netip.AddrPort
- commands.go
- github.com/hafio/vpn-mesh-utility
- Session
- otherNet
- net/http.ResponseWriter
- Node
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
- openTest
- Exit
- netlinkNet
- Net
- planner_test.go
- Registration
- `internal/server/api`
- README.md
- wg.go
- .query
- sync.Mutex
- serve
- LocalStatus
- storeNode
- errorf
- time.Time
- Client
- exprs
- Engine
- meshDNS
- `internal/validate`
- Sync
- .Apply
- Run
- kernel
- .meshctl
- UI
- .add
- `internal/proto`
- time.Duration
- fetchwintun/main_test.go
- net/http.Request
- .AddMesh
- ui
- Container
- rulesFor
- Sync
- .readJSON
- net/netip.Addr
- resolved
- Segmentation
- printStatus
- io.Reader
- .renderPolicies
- deploy_test.go
- Open
- File
- Host
- services
- Ruleset
- Diff
- samples_test.go
- userspace
- ns
- context.Context
- fail
- run
- sync.Mutex
- Schedule
- NodeInfo
- context.Context
- StatusError
- AdminError
- Containers
- StepError
- Prefix6
- TestMain
- Engine
- .apply
- .Reach
- Test catalogue
- .events
- wgdev
- IP
- DefaultPolicy
- `internal/server` and `internal/server/export`
- Writer
- Deps
- run
- ServerURL
- DNSView
- Segmentation
- newQR
- PlanError
- Slot
- Error
- NewNet
- NewKernel
- service_windows.go
- Refusal
- Refusal
- usageError
- AliveRequest
- entry
- Error

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 304 edges
2. `ui` - 106 edges
3. ``internal/server/api`` - 101 edges
4. `ServerConfig()` - 89 edges
5. `Session` - 85 edges
6. `Slot` - 82 edges
7. `newHarness()` - 73 edges
8. `NodeConfig` - 69 edges
9. `env` - 64 edges
10. `api` - 58 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Packages, and which way they may import` --references--> `TestModuleBoundaries()`  [INFERRED]
  CLAUDE.md → internal/archtest/archtest_test.go
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

## Communities (141 total, 27 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (24): linkDNS, DNSRecord, FirewallRule, Node, TestDNS(), NewNet(), NewPolicy(), NewPolicy() (+16 more)

### Community 1 - "WGKey"
Cohesion: 0.14
Nodes (46): blockingHost, rig, watchingFirewall, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New() (+38 more)

### Community 2 - "Name"
Cohesion: 0.21
Nodes (18): check(), newNode(), validateSet(), ToSnapshot(), Endpoint6(), IPv6(), PeerEndpoint(), TestIPv6Grammar() (+10 more)

### Community 3 - "newHarness"
Cohesion: 0.10
Nodes (32): adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing(), TestAdminSetClearsReferences() (+24 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.13
Nodes (43): checkMachineIDFile(), joinFlags(), machineIDFlag(), meshDir(), meshFlags(), onlyMesh(), sep(), setupAgentJoin() (+35 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.14
Nodes (19): Service, checkSegmentation(), CompilePolicy(), hosts(), mergePorts(), mergeServices(), newSegResolver(), Overlaps() (+11 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.39
Nodes (6): Rule, Violation, Check(), importsOf(), under(), underAny()

### Community 11 - "net/netip.AddrPort"
Cohesion: 0.15
Nodes (12): dialing, bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), Listen(), ParseAddr() (+4 more)

### Community 12 - "commands.go"
Cohesion: 0.20
Nodes (15): flagName(), cell(), envDoc(), renderCommands(), sentence(), startFlag(), TestCommandsDoc(), TestContainersDocNamesRealVariables() (+7 more)

### Community 14 - "Session"
Cohesion: 0.11
Nodes (36): Options, ticketServer, Join(), New(), testSession(), TestUnreachableAndMalformedAnswers(), newTicketServer(), sessionClient() (+28 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.12
Nodes (13): api, decode(), reply(), cli(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+5 more)

### Community 17 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 18 - "New"
Cohesion: 0.10
Nodes (17): Add a web UI page, newPolicyForm(), tabsIn(), dnsState(), newRenderer(), New(), sameOrigin(), golden() (+9 more)

### Community 19 - "Dir"
Cohesion: 0.05
Nodes (49): TestStartNodeRuns(), `internal/fsutil`, meshKey(), TestDefaultSocket(), TestJoin(), Dir, Enrollment, NewKeys() (+41 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.20
Nodes (29): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), removable() (+21 more)

### Community 21 - "Run"
Cohesion: 0.11
Nodes (24): hostname(), Run(), TestIsPublic(), TestOpenRuntime(), uiListen(), Presented(), cleanHostname(), firstAddr() (+16 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (26): Routes, NewTable(), compare(), Table, Sync(), NewTable(), luidOf(), NewTable() (+18 more)

### Community 24 - ".nodeView"
Cohesion: 0.21
Nodes (7): ago(), writeHTML(), dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 25 - "Input"
Cohesion: 0.15
Nodes (13): Record, containers6(), Check(), checkContainers(), checkContainers6(), checkDNS(), checkNode6(), checkUsable() (+5 more)

### Community 26 - "lab"
Cohesion: 0.16
Nodes (10): Coverage, lab, TestDualStack(), eventually(), lab, inNS(), newLab(), TestServerMesh() (+2 more)

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "flag.FlagSet"
Cohesion: 0.18
Nodes (33): TestPrintEnrollmentInTheLocalZone(), setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), printMesh(), setupMeshAdd(), setupMeshList() (+25 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (41): Command reference, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit, meshctl export (+33 more)

### Community 30 - ".mesh"
Cohesion: 0.09
Nodes (26): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+18 more)

### Community 31 - "NodeConfig"
Cohesion: 0.05
Nodes (48): IPv6Host, Determinism and goldens, Input and commands, Keys and tokens, `internal/archtest`, `internal/server/phone`, Shared fixtures -- `internal/testfix`, forHost() (+40 more)

### Community 32 - "context.Context"
Cohesion: 0.12
Nodes (4): runner, kick(), oneLine(), drain()

### Community 33 - "newRig"
Cohesion: 0.07
Nodes (52): The web UI's pages (internal/server/ui), `internal/server/ui`, capped, clip(), meshFilter(), New(), inspectOf(), serveRuntime() (+44 more)

### Community 34 - "Service"
Cohesion: 0.22
Nodes (4): resolver, Service, exchange(), serve()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.09
Nodes (74): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), newFakeRuntime() (+66 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (142): Agent (`internal/agent/...`), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), pfx() (+134 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (35): Params, rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants() (+27 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (52): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+44 more)

### Community 39 - "log/slog.Logger"
Cohesion: 0.11
Nodes (10): Hub, subscription, New(), NewUserspace(), NewUserspace(), ConfigEvent, Reconciler, NewReconciler() (+2 more)

### Community 40 - ".exec"
Cohesion: 0.04
Nodes (16): Add a database migration, errWait(), mustAffect(), DNSRecord, joinList(), AgentSession, boolInt(), classify() (+8 more)

### Community 41 - "openTest"
Cohesion: 0.13
Nodes (44): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+36 more)

### Community 43 - "netlinkNet"
Cohesion: 0.17
Nodes (8): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), TestKernelShapes(), toRule()

### Community 44 - "Net"
Cohesion: 0.14
Nodes (7): ensureAddr(), Net, Rule, Remove(), syncDropped(), syncProxies(), syncRules()

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (44): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), Problem, Compute(), deliver(), deliverAddr() (+36 more)

### Community 46 - "Registration"
Cohesion: 0.17
Nodes (9): Domain, fakeRegistrar, Registrar, Registration, NewRegistrar(), NewRegistrar(), NewRegistrar(), serverZone() (+1 more)

### Community 47 - "`internal/server/api`"
Cohesion: 0.10
Nodes (39): apiError, `internal/server/api`, ErrorMessage(), toAPIError(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries(), NewHub() (+31 more)

### Community 48 - "README.md"
Cohesion: 0.14
Nodes (10): graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together (+2 more)

### Community 49 - "wg.go"
Cohesion: 0.14
Nodes (18): peer(), hexKey(), ParseUAPI(), uapiInt(), compareKeys(), Diff(), Change, Device (+10 more)

### Community 50 - ".query"
Cohesion: 0.13
Nodes (6): ContainerName, scanAll(), splitList(), AuditEntry, PeerStat, StatusUpdate

### Community 51 - "sync.Mutex"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 52 - "serve"
Cohesion: 0.12
Nodes (24): brokenStream, ErrorBody, serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallPS1PinsTheChecksum() (+16 more)

### Community 53 - "LocalStatus"
Cohesion: 0.09
Nodes (23): statusDetail, Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest (+15 more)

### Community 54 - "storeNode"
Cohesion: 0.17
Nodes (23): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+15 more)

### Community 55 - "errorf"
Cohesion: 0.07
Nodes (41): NameSet, NodeAddRequest, PhoneRequest, Role, blockOf(), checkGroups(), checkName(), containerNetwork() (+33 more)

### Community 56 - "time.Time"
Cohesion: 0.12
Nodes (13): replayCache, failures, LockoutError, ago(), serves(), sgt(), Health(), phoneHealth() (+5 more)

### Community 57 - "Client"
Cohesion: 0.12
Nodes (10): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, HelloRequest, HelloResponse (+2 more)

### Community 58 - "exprs"
Cohesion: 0.09
Nodes (17): Add a command, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases, Toolchain (+9 more)

### Community 59 - "Engine"
Cohesion: 0.21
Nodes (5): Rechecker, Add an apply step, Engine, Result, PeerStatus

### Community 60 - "meshDNS"
Cohesion: 0.20
Nodes (25): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+17 more)

### Community 61 - "`internal/validate`"
Cohesion: 0.11
Nodes (22): Security invariants, `internal/validate`, addrs(), TestTimeZone(), AddrEntry(), MergeRanges(), TestAddrEntry(), TestAddrRangePrefix() (+14 more)

### Community 62 - "Sync"
Cohesion: 0.16
Nodes (30): Sysctls, NewNet(), Gateway(), Guard(), Owned(), parse(), SlotOf(), Sync() (+22 more)

### Community 63 - ".Apply"
Cohesion: 0.12
Nodes (15): applied, Firewall, policy, Firewall, Runner, atLeast(), exitPolicy(), normalize() (+7 more)

### Community 64 - "Run"
Cohesion: 0.15
Nodes (18): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+10 more)

### Community 65 - "kernel"
Cohesion: 0.11
Nodes (9): Status, IPNet(), Prefix(), index(), route(), lookup(), netlinkTable, kernel (+1 more)

### Community 66 - ".meshctl"
Cohesion: 0.18
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Security (+6 more)

### Community 67 - "UI"
Cohesion: 0.12
Nodes (14): hostOf(), New(), oldest(), validOrNot(), bytesize(), loadPages(), prefsFlash(), UIToken (+6 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.07
Nodes (51): `internal/proto`, TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile(), dualPolicy(), officeHours() (+43 more)

### Community 70 - "time.Duration"
Cohesion: 0.10
Nodes (12): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, TestJudgeHubs(), RateLimiter (+4 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.21
Nodes (12): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), TestDownload(), TestExtractChecksTheDLL() (+4 more)

### Community 72 - "net/http.Request"
Cohesion: 0.14
Nodes (9): fakeServer, signed(), Definitive(), SignRefusal(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie() (+1 more)

### Community 73 - ".AddMesh"
Cohesion: 0.10
Nodes (23): Change, MeshList, meshSpace, SoleMesh(), ULAPrefix(), checkApart(), clash(), freeCIDR() (+15 more)

### Community 74 - "ui"
Cohesion: 0.09
Nodes (11): Session, listOf(), meshChanged(), title(), atoi(), before(), changed(), versionNumber() (+3 more)

### Community 75 - "Container"
Cohesion: 0.31
Nodes (4): fakeRuntime, Container, ContainerName, NamesRequest

### Community 76 - "rulesFor"
Cohesion: 0.18
Nodes (18): nrpt, nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey() (+10 more)

### Community 77 - "Sync"
Cohesion: 0.06
Nodes (20): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+12 more)

### Community 78 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 79 - "net/netip.Addr"
Cohesion: 0.08
Nodes (26): network, hostAddrs, nameTable, zone, gateway(), host(), proxy(), records() (+18 more)

### Community 80 - "resolved"
Cohesion: 0.35
Nodes (4): linkDomain, resolved, classify(), foreignNetns()

### Community 81 - "Segmentation"
Cohesion: 0.09
Nodes (31): MeshSetRequest, RecordRequest, segLiteral, SettingsRequest, segDoc, addIPv6(), Actor, moveMeshIP6() (+23 more)

### Community 82 - "printStatus"
Cohesion: 0.40
Nodes (5): TestBytesize(), dnsLine(), setupDNS(), bytesize(), printStatus()

### Community 83 - "io.Reader"
Cohesion: 0.29
Nodes (4): watchdog, setupImport(), Read(), ReadSegmentation()

### Community 84 - ".renderPolicies"
Cohesion: 0.08
Nodes (22): SegPolicy, SegSchedule, anyText(), listText(), objectRows(), readPolicyForm(), readZoneForm(), ruleRows() (+14 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.19
Nodes (24): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+16 more)

### Community 86 - "Open"
Cohesion: 0.15
Nodes (20): TestExportOnceFails(), TestExportOnceWritesEachMesh(), Import(), document(), linuxOnly(), TestImportIntoANamedMesh(), TestImportRefusesInvalidDocumentsUntouched(), TestImportRefusesWhileTheServerRuns() (+12 more)

### Community 87 - "File"
Cohesion: 0.21
Nodes (7): agentLog(), OpenFile(), lockFile(), unlockFile(), lockFile(), unlockFile(), File

### Community 90 - "Host"
Cohesion: 0.16
Nodes (3): link, exitHost, Host

### Community 93 - "services"
Cohesion: 0.32
Nodes (7): addrText(), Rule, merged(), portText(), services(), Range, service

### Community 94 - "Ruleset"
Cohesion: 0.11
Nodes (9): Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel, noKernel (+1 more)

### Community 95 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 96 - "samples_test.go"
Cohesion: 0.19
Nodes (14): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), mountsMachineID(), sampleVolume() (+6 more)

### Community 98 - "userspace"
Cohesion: 0.10
Nodes (8): adapterPart(), adapterSlot(), NewUserspace(), adapterGUID(), defaultInterface(), exitRoutes(), Link, userspace

### Community 99 - "ns"
Cohesion: 0.16
Nodes (10): lab, TestContainerNetworks(), deleteNamespaces(), ns(), tail(), lab, TestRelayFailover(), lab (+2 more)

### Community 100 - "context.Context"
Cohesion: 0.06
Nodes (24): IPv6Request, IPv6View, MeshUpdate, NodeSetRequest, DNS, Firewall, MeshFirewall, dial() (+16 more)

### Community 101 - "fail"
Cohesion: 0.14
Nodes (16): TestTimestamp(), Timestamp(), TimeZone(), fail(), LocalToken(), mask(), PipePath(), POSIXPath() (+8 more)

### Community 102 - "run"
Cohesion: 0.13
Nodes (27): agentCommands(), argSynopsis(), commands(), newFlags(), parseArgs(), parseFlags(), printFlags(), setupVersion() (+19 more)

### Community 103 - "sync.Mutex"
Cohesion: 0.15
Nodes (6): syncBuffer, Deps, Sysctl, Sysctls, Sysctl, syncBuffer

### Community 104 - "Schedule"
Cohesion: 0.15
Nodes (11): Matrix, MatrixCell, PolicyRule, Schedule, Clock(), Window, Schedule(), TestSchedule() (+3 more)

### Community 105 - "NodeInfo"
Cohesion: 0.18
Nodes (11): CopyInfo, NodeDetail, NodeList, PhoneProfile, addresses(), heard(), ContainerNetwork, NodeInfo (+3 more)

### Community 106 - "context.Context"
Cohesion: 0.07
Nodes (25): Access, Config, DB, Lockout, Pending, Sealer, resetTOTP(), fullExpiry() (+17 more)

### Community 109 - "Containers"
Cohesion: 0.18
Nodes (11): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+3 more)

### Community 111 - "Prefix6"
Cohesion: 0.24
Nodes (8): plain6(), Prefix(), Prefix6(), ServiceEntry(), TestServiceEntry(), Prefix4(), TestPrefix4(), Interface

### Community 113 - "Engine"
Cohesion: 0.15
Nodes (6): nameServer, DNS, Want, MeshView, MeshName, MeshNode

### Community 114 - ".apply"
Cohesion: 0.38
Nodes (4): pendingApply, safeApply, 12. Troubleshooting, TestSafeApplyStateMachine()

### Community 115 - ".Reach"
Cohesion: 0.18
Nodes (9): ReachRequest, SegmentView, ZoneView, activeAt(), ReachResult, zoneOf(), matrixCell, matrixRow (+1 more)

### Community 116 - "Test catalogue"
Cohesion: 0.15
Nodes (10): Several meshes on one host, End to end -- `test/e2e`, `internal/tools/vulnjudge`, Test catalogue, host, lab, stream, TestMeshFirewall() (+2 more)

### Community 117 - ".events"
Cohesion: 0.27
Nodes (7): TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 119 - "IP"
Cohesion: 0.28
Nodes (8): validateMeshNames(), IP(), DNSName(), Domains(), TestDNSName(), TestDomains(), TestZone(), Zone()

### Community 120 - "DefaultPolicy"
Cohesion: 0.36
Nodes (8): DefaultPolicy(), LegacySegmentation(), TestMigrationTurnsRulesIntoPolicies(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSetSegmentationRefusesAStranger(), FirewallRule

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.12
Nodes (22): `internal/server` and `internal/server/export`, DB, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport() (+14 more)

### Community 122 - "Writer"
Cohesion: 0.29
Nodes (3): New(), New(), Writer

### Community 123 - "Deps"
Cohesion: 0.23
Nodes (9): Deps, source, Reset(), clean(), first(), fromFile(), sources(), Read() (+1 more)

### Community 124 - "run"
Cohesion: 0.33
Nodes (5): TestExporterRunExportsAtNight(), TestRunGivesTheMeshIPv6(), main(), run(), TestRun()

### Community 125 - "ServerURL"
Cohesion: 0.33
Nodes (5): Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase(), TestServerURL()

### Community 126 - "DNSView"
Cohesion: 0.40
Nodes (5): DNSEntry, NameEntry, DNSView, dnsData, recordForm

### Community 128 - "newQR"
Cohesion: 0.67
Nodes (3): newQR(), phoneData, qrCode

### Community 130 - "Slot"
Cohesion: 0.14
Nodes (10): Pipe(), Priorities(), TestSlotOf(), NewKernel(), NewKernel(), SysctlKeys(), Slot, InterfaceSlot() (+2 more)

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 379 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **27 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `Session`, `Node`, `Dir`, `Agent (`internal/agent/...`)`, `Run`, `.mesh`, `NodeConfig`, `newRig`, `agent/agent_test.go`, `sync.Mutex`, `time.Time`, `exprs`, `Engine`, `meshDNS`, `Sync`, `Run`, `.meshctl`, ``internal/proto``, `rulesFor`, `Host`, `Engine`, `.apply`, `Test catalogue`?**
  _High betweenness centrality (0.034) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.0922263771034086 - nodes in this community are weakly interconnected._
- **Why does `Slot` connect `Slot` to `ui/ui_test.go`, `Name`, `Session`, `net/http.ResponseWriter`, `Dir`, `Run`, `Input`, `.mesh`, `newRig`, `testing.T`, `netlinkNet`, `Net`, `Registration`, `LocalStatus`, `errorf`, `Sync`, `.Apply`, `net/http.Request`, `rulesFor`, `Sync`, `net/netip.Addr`, `userspace`, `sync.Mutex`, `Writer`, `Deps`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.13725490196078433 - nodes in this community are weakly interconnected._
- **Why does `runner` connect `context.Context` to `ui/ui_test.go`, `newRig`, `log/slog.Logger`, `sync.Mutex`, `Engine`, `net/netip.AddrPort`, `.apply`, `Dir`, `Run`, `LocalStatus`, `time.Time`, `Client`, `Deps`, `NodeConfig`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._