# Graph Report - vpn-mesh-utility  (2026-10-26)

## Corpus Check
- 311 files · ~383,745 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3893 nodes · 20316 edges · 129 communities (103 shown, 26 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2377 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `bcf7c13c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- fail
- `internal/server/api`
- htmx.min.js
- env
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
- commands.go
- github.com/hafio/vpn-mesh-utility
- Node
- otherNet
- net/http.ResponseWriter
- .do
- New
- Dir
- Agent (`internal/agent/...`)
- Agent (`internal/agent/...`)
- net/netip.Prefix
- Net
- .nodeView
- service_test.go
- lab
- Snapshot
- flag.FlagSet
- Command reference
- .mesh
- cli
- context.Context
- newRig
- Service
- agent/agent_test.go
- sync.Mutex
- `internal/server/auth`
- testing.T
- Hub
- .exec
- openTest
- samples_test.go
- netlinkNet
- Input
- planner_test.go
- netlinkPolicy
- serveRuntime
- README.md
- Host
- Tx
- newRig
- Test catalogue
- LocalStatus
- exprs
- errorf
- Exit
- Client
- exprs
- NodeConfig
- meshDNS
- AddrRange
- Sync
- .Apply
- nullInt
- Join
- .meshctl
- UI
- Registration
- Policy
- time.Time
- fetchwintun/main_test.go
- WriteEvent
- .AddMesh
- ui
- Of
- nrpt.go
- Sync
- PublishNames
- net/netip.Addr
- resolved
- Segmentation
- io.Writer
- .Set
- .renderPolicies
- deploy_test.go
- service_windows.go
- File
- otherPolicy
- .add
- Ruleset
- Diff
- server
- userspace
- NewUserspace
- loadMigrations
- io.Reader
- userRun
- Version
- Schedule
- testing.T
- context.Context
- Open
- AdminError
- .SetSegmentation
- StepError
- .InsertMesh
- TestMain
- Engine
- `internal/validate`
- .Reach
- Error
- NewNet
- `internal/server` and `internal/server/export`
- NewNet
- NewKernel
- Refusal
- Slot
- StatusError
- NewKernel
- Run
- entry
- Error
- usageError

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 286 edges
2. `ui` - 104 edges
3. ``internal/server/api`` - 88 edges
4. `Session` - 85 edges
5. `Slot` - 82 edges
6. `ServerConfig()` - 80 edges
7. `NodeConfig` - 67 edges
8. `env` - 66 edges
9. `newHarness()` - 63 edges
10. `runner` - 50 edges

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
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/policy_test.go -> internal/testfix/trust.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/sign_test.go -> internal/testfix/trust.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/trust_test.go -> internal/testfix/trust.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/sse_test.go -> internal/proto/trust_test.go -> internal/testfix/trust.go -> internal/proto/sse_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/policy_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/sign_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/trust_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/sse_test.go -> internal/proto/trust_test.go -> internal/testfix/testfix.go -> internal/proto/sse_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`

## Communities (129 total, 26 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, DNSRequest (+9 more)

### Community 1 - "WGKey"
Cohesion: 0.13
Nodes (46): rig, watchingFirewall, TestProxyLoopAppliesAgain(), TestRunNeedsAnEnrollment(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), TestLinksAddressesAndFailures() (+38 more)

### Community 2 - "fail"
Cohesion: 0.10
Nodes (42): Container-mode agent, check(), ParseAddr(), TestParseAddr(), validateMeshNames(), newNode(), validateSet(), meshCIDR() (+34 more)

### Community 3 - "`internal/server/api`"
Cohesion: 0.07
Nodes (81): countingDB, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket() (+73 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (114): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+106 more)

### Community 5 - "env"
Cohesion: 0.13
Nodes (40): joinFlags(), setupAgentInstall(), setupAgentReset(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+32 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.14
Nodes (17): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), newSegResolver(), Overlaps(), parseSpans() (+9 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.14
Nodes (16): Rule, Violation, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, Input and commands (+8 more)

### Community 11 - "net.Listener"
Cohesion: 0.21
Nodes (9): bindBoth(), bindPair(), dial(), Listen(), dial(), isPipe(), Listen(), listenSocket() (+1 more)

### Community 12 - "commands.go"
Cohesion: 0.13
Nodes (28): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+20 more)

### Community 14 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.13
Nodes (15): api, signed(), decode(), reply(), ETag(), ParseSigned(), Definitive(), SignRefusal() (+7 more)

### Community 17 - ".do"
Cohesion: 0.10
Nodes (20): MeshSetRequest, NodeList, NodeSetRequest, PhoneProfile, SettingsRequest, statusDetail, addresses(), heard() (+12 more)

### Community 18 - "New"
Cohesion: 0.16
Nodes (8): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), New(), sameOrigin(), handler, meshHandler

### Community 19 - "Dir"
Cohesion: 0.06
Nodes (46): `internal/fsutil`, TestJoin(), sign(), TestAnswersMustBeSigned(), TestDataDirIsProtectedOnWindows(), Dir, Enrollment, NewKeys() (+38 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.19
Nodes (32): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+24 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.04
Nodes (83): Options, TestHealthcheckStart(), Agent (`internal/agent/...`), capped, TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine() (+75 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (29): Routes, index(), NewTable(), route(), compare(), Table, Sync(), TestSyncAddsThenRemovesAndIsIdempotent() (+21 more)

### Community 24 - ".nodeView"
Cohesion: 0.12
Nodes (9): Health(), ago(), versionNumber(), writeHTML(), dashboardData, nodeForm, nodeRow, nodesData (+1 more)

### Community 25 - "service_test.go"
Cohesion: 0.21
Nodes (11): hostService(), TestUnitPosture(), SystemdUnit(), exists(), Systemd(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere() (+3 more)

### Community 26 - "lab"
Cohesion: 0.09
Nodes (25): Several meshes on one host, Coverage, End to end -- `test/e2e`, host, lab, TestContainerNetworks(), lab, TestDualStack() (+17 more)

### Community 27 - "Snapshot"
Cohesion: 0.16
Nodes (11): SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields(), fingerprint() (+3 more)

### Community 28 - "flag.FlagSet"
Cohesion: 0.22
Nodes (29): setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), setupMeshAdd(), setupMeshList(), setupMeshRemove(), setupMeshSet() (+21 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (43): Command reference, meshctl agent install, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns (+35 more)

### Community 30 - ".mesh"
Cohesion: 0.19
Nodes (13): 10. Backups, 11. When the control plane is down, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits, 5. Day to day (+5 more)

### Community 31 - "cli"
Cohesion: 0.16
Nodes (24): cli(), TestIPv6DNSRecords(), harness, TestMeshesAreApartOnTheAgentAPI(), TestMeshesStayApart(), TestNodesServiceInAMesh(), TestSetSettings(), TestSoleMesh() (+16 more)

### Community 32 - "context.Context"
Cohesion: 0.07
Nodes (16): Deps, fakeService, runner, Deps, kick(), oneLine(), Reset(), Run() (+8 more)

### Community 33 - "newRig"
Cohesion: 0.09
Nodes (46): brokenStream, Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, TestTemplatesHaveNoScript(), TestDashboardHealth(), TestFirewallPages(), TestNodeGroupsOnPages() (+38 more)

### Community 34 - "Service"
Cohesion: 0.15
Nodes (7): resolver, Service, New(), exchange(), serve(), NewUserspace(), Options

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (63): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+55 more)

### Community 36 - "sync.Mutex"
Cohesion: 0.22
Nodes (3): syncBuffer, Sysctl, syncBuffer

### Community 37 - "`internal/server/auth`"
Cohesion: 0.13
Nodes (28): rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword() (+20 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (55): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+47 more)

### Community 39 - "Hub"
Cohesion: 0.24
Nodes (3): Hub, subscription, ConfigEvent

### Community 40 - ".exec"
Cohesion: 0.09
Nodes (7): mustAffect(), boolInt(), classify(), changed(), Session, PeerStat, StatusUpdate

### Community 41 - "openTest"
Cohesion: 0.09
Nodes (62): `internal/server/store`, TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes() (+54 more)

### Community 42 - "samples_test.go"
Cohesion: 0.18
Nodes (12): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+4 more)

### Community 43 - "netlinkNet"
Cohesion: 0.10
Nodes (16): netlinkNet, Owned(), SlotOf(), TestSlotOf(), blackhole(), dropRoute(), dump(), family() (+8 more)

### Community 44 - "Input"
Cohesion: 0.08
Nodes (27): PlanError, Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs (+19 more)

### Community 45 - "planner_test.go"
Cohesion: 0.14
Nodes (41): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), deliver(), deliverAddr(), encode(), firstDiff() (+33 more)

### Community 46 - "netlinkPolicy"
Cohesion: 0.18
Nodes (9): Family, netlinkPolicy, dump(), family(), owned(), TestKernelShapes(), TestOwnsOnlyItsPriorities(), toRoute() (+1 more)

### Community 47 - "serveRuntime"
Cohesion: 0.18
Nodes (15): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+7 more)

### Community 48 - "README.md"
Cohesion: 0.23
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "Host"
Cohesion: 0.05
Nodes (40): link, exitHost, Host, peer(), AddrPort(), TestAddrPort(), hexKey(), ParseUAPI() (+32 more)

### Community 50 - "Tx"
Cohesion: 0.06
Nodes (13): SoleMesh(), freeSlot(), PickMesh(), pickMesh(), Mesh, scanMesh(), scanNode(), scanAll() (+5 more)

### Community 51 - "newRig"
Cohesion: 0.24
Nodes (11): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+3 more)

### Community 52 - "Test catalogue"
Cohesion: 0.19
Nodes (17): `internal/server/phone`, `internal/tools/vulnjudge`, Test catalogue, Hub(), joinSet(), NewKey(), Render(), aliceKey() (+9 more)

### Community 53 - "LocalStatus"
Cohesion: 0.07
Nodes (22): Client, Logs, logs(), refusal(), laptop(), clip(), meshFilter(), PrefsRequest (+14 more)

### Community 54 - "exprs"
Cohesion: 0.13
Nodes (14): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+6 more)

### Community 55 - "errorf"
Cohesion: 0.08
Nodes (45): apiError, IPv6Request, IPv6View, NodeAddRequest, PhoneRequest, Role, addIPv6(), blockOf() (+37 more)

### Community 57 - "Client"
Cohesion: 0.22
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "exprs"
Cohesion: 0.17
Nodes (8): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), TestIntervalAddrElems(), TestIntervalElems(), nftOp

### Community 59 - "NodeConfig"
Cohesion: 0.07
Nodes (25): NodeDetail, dialing, IPv6Host, Rechecker, Determinism and goldens, Add an apply step, Shared fixtures -- `internal/testfix`, forHost() (+17 more)

### Community 60 - "meshDNS"
Cohesion: 0.23
Nodes (22): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceAnswersOverIPv6() (+14 more)

### Community 61 - "AddrRange"
Cohesion: 0.20
Nodes (13): addrs(), contains(), inRanges(), inRange(), AddrEntry(), AddrRange, LastOf(), MergeRanges() (+5 more)

### Community 62 - "Sync"
Cohesion: 0.09
Nodes (35): Sysctls, NewNet(), Sysctls, ensureAddr(), Gateway(), Net, Rule, Guard() (+27 more)

### Community 63 - ".Apply"
Cohesion: 0.13
Nodes (14): applied, Firewall, policy, Firewall, Runner, exitPolicy(), normalize(), relayPolicy() (+6 more)

### Community 64 - "nullInt"
Cohesion: 0.18
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 65 - "Join"
Cohesion: 0.27
Nodes (8): setupAgentJoin(), hostService(), installAgent(), cleanHostname(), JoinOptions, Service, Join(), JoinOnStart()

### Community 66 - ".meshctl"
Cohesion: 0.19
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens (+6 more)

### Community 67 - "UI"
Cohesion: 0.10
Nodes (17): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize() (+9 more)

### Community 68 - "Registration"
Cohesion: 0.21
Nodes (7): Domain, fakeRegistrar, Registrar, Registration, NewRegistrar(), NewRegistrar(), NewRegistrar()

### Community 69 - "Policy"
Cohesion: 0.13
Nodes (28): NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily(), compile() (+20 more)

### Community 70 - "time.Time"
Cohesion: 0.06
Nodes (29): pendingApply, safeApply, Change, Failover, hubMove, hubState, replayCache, failures (+21 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.17
Nodes (18): `internal/tools/fetchwintun`, Pin, install(), Sum(), TestInstall(), TestExporterRunExportsAtNight(), download(), extract() (+10 more)

### Community 72 - "WriteEvent"
Cohesion: 0.26
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 73 - ".AddMesh"
Cohesion: 0.18
Nodes (14): MeshList, meshSpace, MeshUpdate, ULAPrefix(), checkApart(), clash(), freeCIDR(), MeshAddRequest (+6 more)

### Community 74 - "ui"
Cohesion: 0.09
Nodes (11): Session, listOf(), meshChanged(), atoi(), before(), changed(), message(), actor() (+3 more)

### Community 75 - "Of"
Cohesion: 0.22
Nodes (8): fakeRuntime, Container, Of(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack(), TestOfWithIPv6(), ContainerName

### Community 76 - "nrpt.go"
Cohesion: 0.29
Nodes (13): nrptRule, desired(), domainName(), same(), TestSameRegistration(), fromRules(), keyPart(), nrptKey() (+5 more)

### Community 77 - "Sync"
Cohesion: 0.10
Nodes (17): Policy, mutate(), Desired(), Policy, Route, Rule, V6Support, Sync() (+9 more)

### Community 78 - "PublishNames"
Cohesion: 0.15
Nodes (13): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+5 more)

### Community 79 - "net/netip.Addr"
Cohesion: 0.09
Nodes (22): network, hostAddrs, nameTable, zone, Apply (agent), TestIsPublic(), gateway(), host() (+14 more)

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "Segmentation"
Cohesion: 0.09
Nodes (23): RecordRequest, segLiteral, segDoc, Actor, problemStrings(), segLiterals(), plannerSegmentation(), PolicyZone() (+15 more)

### Community 82 - "io.Writer"
Cohesion: 0.20
Nodes (10): dnsLine(), setupDNS(), printExit(), printUsage(), printMesh(), printEnrollment(), bytesize(), printStatus() (+2 more)

### Community 83 - ".Set"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 84 - ".renderPolicies"
Cohesion: 0.06
Nodes (30): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+22 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.24
Nodes (20): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+12 more)

### Community 86 - "service_windows.go"
Cohesion: 0.12
Nodes (9): uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows(), handler (+1 more)

### Community 87 - "File"
Cohesion: 0.16
Nodes (10): agentLog(), main(), ttyPassword(), OpenFile(), TestFileRotates(), lockFile(), unlockFile(), lockFile() (+2 more)

### Community 93 - ".add"
Cohesion: 0.16
Nodes (11): clip(), Ring, addrText(), Rule, merged(), portText(), services(), ServiceEntry() (+3 more)

### Community 94 - "Ruleset"
Cohesion: 0.11
Nodes (9): Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel, noKernel (+1 more)

### Community 95 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 96 - "server"
Cohesion: 0.57
Nodes (7): `internal/server/versions`, server(), snap(), TestARenameIsAChangeNotAReplacement(), TestCompareEqualSnapshotsIsEmpty(), TestCompareListsEveryKindOfChange(), TestCompareSeesThroughTheMigration()

### Community 98 - "userspace"
Cohesion: 0.06
Nodes (10): Status, TestCreateLinkNeedsPrivilege(), lookup(), NewUserspace(), adapterGUID(), defaultInterface(), Link, kernel (+2 more)

### Community 99 - "NewUserspace"
Cohesion: 0.50
Nodes (3): NewUserspace(), Ensure(), TestEnsureWithoutABundle()

### Community 100 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 101 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 102 - "userRun"
Cohesion: 0.60
Nodes (5): setupUser(), setupUserGrant(), userCommands(), userRun(), userAction

### Community 104 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "testing.T"
Cohesion: 0.06
Nodes (58): `internal/proto`, TestDNS(), next(), executes(), TestHandlerReportsAFailedRun(), TestHandlerStopsTheAgent(), New(), New() (+50 more)

### Community 106 - "context.Context"
Cohesion: 0.07
Nodes (29): Access, Config, DB, Lockout, Params, Pending, Sealer, resetTOTP() (+21 more)

### Community 107 - "Open"
Cohesion: 0.50
Nodes (3): TestOpen(), Open(), TestOpenRefusesUnknownBackends()

### Community 113 - "Engine"
Cohesion: 0.17
Nodes (5): nameServer, DNS, Want, newNameTable(), MeshName

### Community 114 - "`internal/validate`"
Cohesion: 0.06
Nodes (32): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), TestRef(), Arg(), isLoopbackHost() (+24 more)

### Community 115 - ".Reach"
Cohesion: 0.16
Nodes (9): ReachRequest, SegmentView, ZoneView, activeAt(), ReachResult, zoneOf(), matrixCell, matrixRow (+1 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.09
Nodes (34): `internal/server` and `internal/server/export`, DB, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport() (+26 more)

### Community 130 - "Slot"
Cohesion: 0.10
Nodes (20): Pipe(), TestPolicyHasKernelSemantics(), Holder(), Priorities(), SlotOf(), Taken(), TestHolder(), TestSlotOf() (+12 more)

### Community 133 - "Run"
Cohesion: 0.09
Nodes (27): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+19 more)

## Knowledge Gaps
- **8 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `harness`, `NO_COLOR`, `lab` (+3 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 367 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **26 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `WGKey`, `Slot`, `fail`, `Run`, `archtest.go`, `Node`, `Dir`, `Agent (`internal/agent/...`)`, `net/netip.Prefix`, `service_test.go`, `.mesh`, `newRig`, `agent/agent_test.go`, `netlinkNet`, `netlinkPolicy`, `serveRuntime`, `Host`, `newRig`, `Test catalogue`, `exprs`, `exprs`, `NodeConfig`, `meshDNS`, `Sync`, `.meshctl`, `UI`, `Policy`, `time.Time`, `fetchwintun/main_test.go`, `WriteEvent`, `Of`, `nrpt.go`, `Sync`, `net/netip.Addr`, `resolved`, `.Set`, `service_windows.go`, `File`, `userspace`, `NewUserspace`, `testing.T`, `Open`, `Engine`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 285 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 285 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `harness` to the rest of the system?**
  _8 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09329116479375547 - nodes in this community are weakly interconnected._
- **Why does `Command reference` connect `Command reference` to `README.md`, `.meshctl`, `commands.go`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.1346938775510204 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `userspace`, `.mesh`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._