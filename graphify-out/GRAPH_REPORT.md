# Graph Report - vpn-mesh-utility  (2026-10-25)

## Corpus Check
- 299 files · ~366,684 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3773 nodes · 19043 edges · 143 communities (112 shown, 31 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2286 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d380b38e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- env
- newHarness
- htmx.min.js
- flag.FlagSet
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
- time.Time
- github.com/hafio/vpn-mesh-utility
- Node
- otherNet
- net/http.ResponseWriter
- .do
- meshctl/main.go
- Dir
- ServerConfig
- Agent (`internal/agent/...`)
- net/netip.Prefix
- Net
- ui
- Join
- lab
- Snapshot
- exprs
- Command reference
- .meshctl
- Installer
- runner
- newRig
- Service
- agent/agent_test.go
- File
- `internal/server/auth`
- `cmd/meshctl`
- Sync
- .exec
- openTest
- samples_test.go
- netlinkNet
- Input
- planner_test.go
- Run
- `internal/server/api`
- README.md
- Key
- Tx
- log/slog.Logger
- PublishNames
- LocalStatus
- newRig
- MeshTx
- serveRuntime
- Client
- Net
- NodeConfig
- meshDNS
- .add
- Sync
- .Apply
- Host
- net/http.Request
- serve
- UI
- WriteEvent
- AddrRange
- Policy
- fetchwintun/main.go
- CompilePolicy
- .AddMesh
- storeNode
- sync.Mutex
- nrpt.go
- .settings
- Extending the code
- net/netip.Addr
- resolved
- errorf
- conn
- wgdev
- .renderPolicies
- deploy_test.go
- stop
- meshctl/docs_test.go
- ui
- Diff
- Kernel
- netlinkPolicy
- loadMigrations
- userspace
- build
- .apply
- noKernel
- .Sync
- Allocate
- Schedule
- `internal/proto`
- context.Context
- IPNet
- AdminError
- toAPIError
- StepError
- Open
- TestMain
- Engine
- fail
- Exit
- otherPolicy
- Error
- .queryRow
- NewNet
- Registration
- `internal/server` and `internal/server/export`
- .Reach
- NewNet
- NewKernel
- Refusal
- Table
- Runtime
- policy
- Firewall
- Slot
- StatusError
- NewKernel
- Run
- TestFailoverMovesRelays
- TestRunGivesTheMeshIPv6
- NewHub
- newQR
- ns
- entry
- PlanError
- Error
- io.Reader

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 263 edges
2. `ui` - 104 edges
3. ``internal/server/api`` - 86 edges
4. `Session` - 85 edges
5. `ServerConfig()` - 71 edges
6. `NodeConfig` - 67 edges
7. `env` - 65 edges
8. `newHarness()` - 60 edges
9. `Slot` - 53 edges
10. `runner` - 47 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
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

## Communities (143 total, 31 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (16): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, DNSRequest (+8 more)

### Community 1 - "WGKey"
Cohesion: 0.16
Nodes (36): TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS(), newRig() (+28 more)

### Community 2 - "env"
Cohesion: 0.17
Nodes (42): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+34 more)

### Community 3 - "newHarness"
Cohesion: 0.12
Nodes (30): ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing() (+22 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (114): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+106 more)

### Community 5 - "flag.FlagSet"
Cohesion: 0.13
Nodes (33): dnsLine(), printExit(), setupFirewallApply(), setupFirewallShow(), setupIPv6(), printMesh(), setupMeshAdd(), setupMeshRemove() (+25 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (33): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem, 2. Goals (+25 more)

### Community 7 - "segResolver"
Cohesion: 0.16
Nodes (16): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), parseSpans(), AddrEntry(), PrefixRange() (+8 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.33
Nodes (7): Rule, Violation, Check(), importsOf(), ModuleRoot(), under(), underAny()

### Community 11 - "net.Listener"
Cohesion: 0.33
Nodes (6): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), serve()

### Community 12 - "time.Time"
Cohesion: 0.07
Nodes (22): pendingApply, safeApply, Failover, hubMove, hubState, replayCache, failures, hitWindow (+14 more)

### Community 14 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.16
Nodes (10): api, decode(), reply(), cli(), Nodes, hostFamilies(), decodeJSON(), readBody() (+2 more)

### Community 17 - ".do"
Cohesion: 0.09
Nodes (24): NodeList, NodeSetRequest, PhoneProfile, PhoneRequest, addresses(), heard(), checkName(), nodeByName() (+16 more)

### Community 18 - "meshctl/main.go"
Cohesion: 0.14
Nodes (23): agentCommands(), commands(), passwords(), runWith(), TestUserAccessCommands(), TestUserCommands(), firewallCommands(), healthCommands() (+15 more)

### Community 19 - "Dir"
Cohesion: 0.07
Nodes (36): `internal/fsutil`, isRefusal(), JoinOnStart(), joinOnStart(), startDir(), TestJoinOnStart(), TestJoinOnStartAsksTheServer(), TestJoinOnStartCannotRemember() (+28 more)

### Community 20 - "ServerConfig"
Cohesion: 0.16
Nodes (31): Input and commands, `internal/archtest`, Shared fixtures -- `internal/testfix`, fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden() (+23 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.03
Nodes (89): Agent (`internal/agent/...`), Exec, TestCleanHostname(), TestDefaultSocket(), TestIsPublic(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine() (+81 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (14): Routes, index(), route(), compare(), luidOf(), onLink(), Host, samePrefixes() (+6 more)

### Community 24 - "ui"
Cohesion: 0.12
Nodes (7): Session, listOf(), meshChanged(), atoi(), changed(), message(), actor()

### Community 25 - "Join"
Cohesion: 0.29
Nodes (8): installAgent(), installAgent(), cleanHostname(), JoinOptions, Service, Join(), meshAddresses(), JoinResponse

### Community 26 - "lab"
Cohesion: 0.13
Nodes (17): Several meshes on one host, Coverage, End to end -- `test/e2e`, host, lab, TestDualStack(), deleteNamespaces(), eventually() (+9 more)

### Community 27 - "Snapshot"
Cohesion: 0.13
Nodes (20): `internal/server/versions`, DecodeSnapshot(), SnapNode, Snapshot, snapNode(), TestDecodeSnapshotRefusesUnknownShapes(), Compare(), describeDNS() (+12 more)

### Community 28 - "exprs"
Cohesion: 0.24
Nodes (5): exprs(), ifname(), intervalAddrElems(), TestExprs(), nftOp

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".meshctl"
Cohesion: 0.07
Nodes (36): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+28 more)

### Community 31 - "Installer"
Cohesion: 0.28
Nodes (5): hostService(), Runner, exists(), Systemd(), Installer

### Community 32 - "runner"
Cohesion: 0.11
Nodes (10): fakeService, runner, kick(), oneLine(), dial(), dial(), isPipe(), drain() (+2 more)

### Community 33 - "newRig"
Cohesion: 0.12
Nodes (35): The web UI's pages (internal/server/ui), `internal/server/ui`, capped, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages() (+27 more)

### Community 34 - "Service"
Cohesion: 0.14
Nodes (9): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+1 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (63): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+55 more)

### Community 36 - "File"
Cohesion: 0.31
Nodes (4): agentLog(), ttyPassword(), OpenFile(), File

### Community 37 - "`internal/server/auth`"
Cohesion: 0.10
Nodes (38): Config, DB, Params, rig, 11. CLI reference, `internal/server/auth`, New(), newRig() (+30 more)

### Community 38 - "`cmd/meshctl`"
Cohesion: 0.06
Nodes (85): TestAgentRunRuntimeSocket(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin(), TestBytesize() (+77 more)

### Community 39 - "Sync"
Cohesion: 0.17
Nodes (6): Policy, mutate(), Policy, Route, Rule, Sync()

### Community 40 - ".exec"
Cohesion: 0.06
Nodes (11): Add a database migration, newUUID(), mustAffect(), DNSRecord, joinList(), boolInt(), classify(), nullInt() (+3 more)

### Community 41 - "openTest"
Cohesion: 0.15
Nodes (42): `internal/server/store`, addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh(), TestServerSettings(), TestTokenMesh() (+34 more)

### Community 42 - "samples_test.go"
Cohesion: 0.20
Nodes (12): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+4 more)

### Community 43 - "netlinkNet"
Cohesion: 0.13
Nodes (11): netlinkNet, Owned(), SlotOf(), TestSlotOf(), blackhole(), dropRoute(), dump(), family() (+3 more)

### Community 44 - "Input"
Cohesion: 0.16
Nodes (11): Record, Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable(), sortedByID(), Input (+3 more)

### Community 45 - "planner_test.go"
Cohesion: 0.14
Nodes (42): Add a topology rule, `internal/server/planner`, Problem, Compute(), deliver(), deliverAddr(), encode(), firstDiff() (+34 more)

### Community 46 - "Run"
Cohesion: 0.12
Nodes (10): Deps, Deps, Run(), uiListen(), DNS, Firewall, MeshFirewall, Sysctls (+2 more)

### Community 47 - "`internal/server/api`"
Cohesion: 0.16
Nodes (26): `internal/server/api`, ErrorMessage(), TestIPv6DNSRecords(), harness, TestMeshAdmin(), TestMeshesAreApartOnTheAgentAPI(), TestMeshesStayApart(), TestNodesServiceInAMesh() (+18 more)

### Community 48 - "README.md"
Cohesion: 0.13
Nodes (12): graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, `internal/tools/vulnjudge`, Test catalogue, Building (+4 more)

### Community 49 - "Key"
Cohesion: 0.10
Nodes (31): peer(), AddrPort(), TestAddrPort(), intervalElems(), TestIntervalElems(), hexKey(), ParseUAPI(), getAnswer() (+23 more)

### Community 50 - "Tx"
Cohesion: 0.14
Nodes (6): ContainerName, scanAll(), splitList(), AuditEntry, PeerStat, StatusUpdate

### Community 51 - "log/slog.Logger"
Cohesion: 0.10
Nodes (15): Change, Deps, Hub, subscription, NewUserspace(), DB, Reconciler, Result (+7 more)

### Community 52 - "PublishNames"
Cohesion: 0.21
Nodes (10): DNSEntry, NameEntry, NameSet, compareAddr(), LoadNames(), PublishNames(), TestPublishNames(), DNSView (+2 more)

### Community 53 - "LocalStatus"
Cohesion: 0.08
Nodes (23): statusDetail, Client, Logs, logs(), refusal(), laptop(), PrefsRequest, PrefsResponse (+15 more)

### Community 54 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 55 - "MeshTx"
Cohesion: 0.08
Nodes (36): NodeAddRequest, segLiteral, Role, blockOf(), containerNetwork(), containerSubnet(), containerSubnet6(), defaultMTU() (+28 more)

### Community 56 - "serveRuntime"
Cohesion: 0.18
Nodes (15): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+7 more)

### Community 57 - "Client"
Cohesion: 0.20
Nodes (9): Options, apiError(), clip(), Client, Join(), New(), StatusResponse, NamesResponse (+1 more)

### Community 58 - "Net"
Cohesion: 0.18
Nodes (5): ensureAddr(), Net, Rule, syncDropped(), syncRules()

### Community 59 - "NodeConfig"
Cohesion: 0.06
Nodes (38): NodeDetail, dialing, IPv6Host, Determinism and goldens, Keys and tokens, `internal/server/phone`, forHost(), Engine (+30 more)

### Community 60 - "meshDNS"
Cohesion: 0.25
Nodes (22): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+14 more)

### Community 61 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 62 - "Sync"
Cohesion: 0.20
Nodes (24): Sysctls, NewNet(), Gateway(), Guard(), parse(), Sync(), TestGatewayRules(), TestGatewayRulesWithIPv6() (+16 more)

### Community 63 - ".Apply"
Cohesion: 0.26
Nodes (8): applied, Firewall, atLeast(), normalize(), richRule(), RuleArgv(), FWRule, sortRules()

### Community 64 - "Host"
Cohesion: 0.13
Nodes (5): link, exitHost, rig, Host, Sysctl

### Community 65 - "net/http.Request"
Cohesion: 0.20
Nodes (4): ClientIP(), inAny(), TestClientIP(), settingsData

### Community 66 - "serve"
Cohesion: 0.14
Nodes (21): brokenStream, serve(), signed(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend(), TestEventsStreamPushesChanges(), TestHealth(), TestHealthFailsWithoutTheDatabase() (+13 more)

### Community 67 - "UI"
Cohesion: 0.12
Nodes (15): hostOf(), New(), oldest(), validOrNot(), ago(), bytesize(), loadPages(), prefsFlash() (+7 more)

### Community 68 - "WriteEvent"
Cohesion: 0.25
Nodes (7): TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 69 - "AddrRange"
Cohesion: 0.13
Nodes (14): addrs(), contains(), ofFamily(), inRanges(), serves(), inRange(), AddrRange, Service (+6 more)

### Community 70 - "Policy"
Cohesion: 0.14
Nodes (27): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+19 more)

### Community 71 - "fetchwintun/main.go"
Cohesion: 0.22
Nodes (16): `internal/tools/fetchwintun`, Pin, Sum(), TestExporterRunExportsAtNight(), download(), extract(), fetch(), main() (+8 more)

### Community 72 - "CompilePolicy"
Cohesion: 0.13
Nodes (17): Matrix, MatrixCell, SegmentView, ZoneView, PolicyRule, Schedule, activeAt(), zoneOf() (+9 more)

### Community 73 - ".AddMesh"
Cohesion: 0.12
Nodes (20): MeshList, meshSpace, SoleMesh(), checkApart(), clash(), freeCIDR(), freeSlot(), MeshAddRequest (+12 more)

### Community 74 - "storeNode"
Cohesion: 0.26
Nodes (15): countingDB, harness, joinBody(), postJoin(), storeNode(), TestJoinEnrollsAndConsumesTheToken(), TestJoinRefusesAPhone(), TestJoinRefusesHostileInputUpFront() (+7 more)

### Community 75 - "sync.Mutex"
Cohesion: 0.13
Nodes (8): fakeRuntime, syncBuffer, Container, hostAddr(), Of(), Source, ContainerName, syncBuffer

### Community 76 - "nrpt.go"
Cohesion: 0.19
Nodes (15): nrpt, nrptRule, desired(), domainName(), fromRules(), nrptKey(), ownedKey(), parseServers() (+7 more)

### Community 77 - ".settings"
Cohesion: 0.17
Nodes (8): MeshSetRequest, MeshUpdate, SettingsRequest, Container-mode agent, problemStrings(), meshPath(), CheckServerDomains(), TestCheckServerDomains()

### Community 78 - "Extending the code"
Cohesion: 0.18
Nodes (10): Rechecker, Add a command, Add a validator, Add an apply step, Dev script tasks, Developer guide, Extending the code, Goldens (+2 more)

### Community 79 - "net/netip.Addr"
Cohesion: 0.08
Nodes (23): network, hostAddrs, nameTable, zone, Apply (agent), gateway(), host(), syncProxies() (+15 more)

### Community 80 - "resolved"
Cohesion: 0.35
Nodes (4): linkDomain, resolved, classify(), foreignNetns()

### Community 81 - "errorf"
Cohesion: 0.13
Nodes (20): RecordRequest, segDoc, addIPv6(), Actor, errorf(), invalid(), MeshChange, meshChange() (+12 more)

### Community 84 - ".renderPolicies"
Cohesion: 0.06
Nodes (32): Add a web UI page, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm(), ruleRows() (+24 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.21
Nodes (22): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+14 more)

### Community 86 - "stop"
Cohesion: 0.12
Nodes (8): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), Windows(), handler, SCM

### Community 87 - "meshctl/docs_test.go"
Cohesion: 0.12
Nodes (24): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), parsePrefixes(), printFlags(), shownDefault() (+16 more)

### Community 90 - "ui"
Cohesion: 0.07
Nodes (21): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), group(), ago(), before(), bytesize() (+13 more)

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 94 - "Kernel"
Cohesion: 0.21
Nodes (4): watchingFirewall, Kernel, Ruleset, Chain

### Community 95 - "netlinkPolicy"
Cohesion: 0.20
Nodes (7): Family, netlinkPolicy, dump(), family(), owned(), toRoute(), toRule()

### Community 96 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 98 - "userspace"
Cohesion: 0.11
Nodes (6): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace

### Community 100 - ".apply"
Cohesion: 0.19
Nodes (3): IPv6Request, IPv6View, ipv6View()

### Community 103 - "Allocate"
Cohesion: 0.18
Nodes (18): TestULAPrefix(), ULAPrefix(), Allocate(), Allocate6(), AllocateContainers(), AllocateContainers6(), Block(), Block6() (+10 more)

### Community 104 - "Schedule"
Cohesion: 0.15
Nodes (10): The mesh firewall, SegPolicy, SegSchedule, scheduleText(), Clock(), Window, Schedule(), TestSchedule() (+2 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.07
Nodes (42): `internal/proto`, newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestNamesRoundTrip() (+34 more)

### Community 106 - "context.Context"
Cohesion: 0.08
Nodes (24): Access, Lockout, Pending, Sealer, newPassword(), passwdUser(), resetTOTP(), fullExpiry() (+16 more)

### Community 107 - "IPNet"
Cohesion: 0.18
Nodes (5): Status, IPNet(), lookup(), kernel, links

### Community 109 - "toAPIError"
Cohesion: 0.40
Nodes (4): apiError, toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion()

### Community 111 - "Open"
Cohesion: 0.21
Nodes (11): TestMigrationMakesMeshOne(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), Store, Open(), TestOpenIsIdempotent() (+3 more)

### Community 113 - "Engine"
Cohesion: 0.15
Nodes (6): nameServer, DNS, Want, newNameTable(), DNS, MeshName

### Community 114 - "fail"
Cohesion: 0.06
Nodes (68): Security invariants, `internal/validate`, check(), ParseAddr(), validateMeshNames(), checkGroups(), newNode(), validateSet() (+60 more)

### Community 120 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.11
Nodes (29): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+21 more)

### Community 122 - ".Reach"
Cohesion: 0.20
Nodes (9): ReachRequest, flowEnd(), flowEnds(), ReachResult, TestTimestamp(), Timestamp(), matrixCell, matrixRow (+1 more)

### Community 126 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 127 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 128 - "policy"
Cohesion: 0.54
Nodes (4): policy, exitPolicy(), relayPolicy(), wantedPolicies()

### Community 130 - "Slot"
Cohesion: 0.15
Nodes (8): Priorities(), SlotOf(), TestSlotOf(), NewKernel(), NewKernel(), SysctlKeys(), Slot, TestSlotNames()

### Community 133 - "Run"
Cohesion: 0.13
Nodes (20): Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New(), NewAdmin() (+12 more)

### Community 134 - "TestFailoverMovesRelays"
Cohesion: 0.50
Nodes (3): harness, TestFailoverMovesRelays(), TestFailoverRun()

### Community 136 - "NewHub"
Cohesion: 0.50
Nodes (5): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 137 - "newQR"
Cohesion: 0.67
Nodes (3): newQR(), phoneData, qrCode

### Community 138 - "ns"
Cohesion: 0.18
Nodes (8): lab, TestContainerNetworks(), ns(), lab, TestRelayFailover(), lab, stream, TestMeshFirewall()

### Community 146 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

## Knowledge Gaps
- **8 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `harness`, `NO_COLOR`, `lab` (+3 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 360 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **31 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `WGKey`, `Run`, `time.Time`, `Node`, `Dir`, `ServerConfig`, `exprs`, `.meshctl`, `agent/agent_test.go`, ``cmd/meshctl``, `README.md`, `Key`, `newRig`, `serveRuntime`, `NodeConfig`, `meshDNS`, `Sync`, `Host`, `AddrRange`, `Policy`, `nrpt.go`, ``internal/proto``, `Engine`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Are the 262 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 262 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `harness` to the rest of the system?**
  _8 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08929078014184397 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect ``internal/server/api`` to `serve`, `newHarness`, `htmx.min.js`, `Run`, `TestFailoverMovesRelays`, `Allocate`, `NewHub`, `storeNode`, `time.Time`, `.settings`, `toAPIError`, `net/netip.Addr`, `README.md`, `.do`, `PublishNames`, `MeshTx`, `serveRuntime`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Should `newHarness` be split into smaller, more focused modules?**
  _Cohesion score 0.11742424242424243 - nodes in this community are weakly interconnected._
- **Why does `at()` connect `htmx.min.js` to ``internal/server/api``?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._