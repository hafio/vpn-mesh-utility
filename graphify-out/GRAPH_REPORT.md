# Graph Report - vpn-mesh-utility  (2026-10-25)

## Corpus Check
- 301 files · ~368,667 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3789 nodes · 19280 edges · 129 communities (105 shown, 24 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2304 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `82bac192`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- env
- `internal/server/api`
- htmx.min.js
- flag.FlagSet
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
- sync.Mutex
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
- NodeInfo
- Join
- lab
- Snapshot
- exprs
- Command reference
- .mesh
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
- `internal/validate`
- README.md
- Host
- Tx
- Hub
- usagef
- LocalStatus
- newRig
- errorf
- serveRuntime
- Client
- Net
- NodeConfig
- meshDNS
- .add
- Sync
- .Apply
- nullInt
- Open
- .meshctl
- UI
- WriteEvent
- AddrRange
- Policy
- fetchwintun/main.go
- CompilePolicy
- Mesh
- New
- Of
- nrpt.go
- ToSnapshot
- fail
- net/netip.Addr
- resolved
- errorf
- net/netip.AddrPort
- Agent protocol
- TestPageGoldens
- deploy_test.go
- stop
- meshctl/docs_test.go
- net/http.Request
- Diff
- Ruleset
- ServerURL
- loadMigrations
- userspace
- AuditEntry
- .SetSegmentation
- Rollback
- objectForm
- Schedule
- `internal/proto`
- context.Context
- AdminError
- StepError
- TestMain
- Engine
- Name
- Error
- .queryRow
- NewNet
- `internal/server` and `internal/server/export`
- .Reach
- NewNet
- NewKernel
- Refusal
- Slot
- StatusError
- NewKernel
- log/slog.Logger
- stream
- ns
- entry
- PlanError
- Error
- io.Writer

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 267 edges
2. `ui` - 104 edges
3. ``internal/server/api`` - 86 edges
4. `Session` - 85 edges
5. `Slot` - 72 edges
6. `ServerConfig()` - 71 edges
7. `NodeConfig` - 67 edges
8. `env` - 65 edges
9. `newHarness()` - 60 edges
10. `runner` - 47 edges

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
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/policy_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 3-file cycle: `internal/proto/names_test.go -> internal/proto/sign_test.go -> internal/testfix/testfix.go -> internal/proto/names_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/containers/containers_test.go -> internal/agent/containers/net_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/exitmode/exitmode_test.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/userspace_windows.go -> internal/agent/exitmode/policy_linux.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/userspace_windows.go -> internal/agent/exitmode/policy_linux_test.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`

## Communities (129 total, 24 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (16): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, DNSRequest (+8 more)

### Community 1 - "WGKey"
Cohesion: 0.16
Nodes (36): rig, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS() (+28 more)

### Community 2 - "env"
Cohesion: 0.23
Nodes (30): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), setupVersion() (+22 more)

### Community 3 - "`internal/server/api`"
Cohesion: 0.06
Nodes (100): countingDB, `internal/server/api`, ErrorBody, cli(), ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput() (+92 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (117): Status, 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names (+109 more)

### Community 5 - "flag.FlagSet"
Cohesion: 0.21
Nodes (23): setupFirewallApply(), setupFirewallShow(), setupIPv6(), setupMeshAdd(), setupMeshRemove(), setupMeshSet(), setupMeshShow(), adminFlags() (+15 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.18
Nodes (14): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), newSegResolver(), parseSpans(), ZoneAddresses() (+6 more)

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
Cohesion: 0.18
Nodes (14): bindBoth(), bindPair(), Listen(), Listen(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly() (+6 more)

### Community 12 - "sync.Mutex"
Cohesion: 0.08
Nodes (14): syncBuffer, Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, Sysctl (+6 more)

### Community 14 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.12
Nodes (10): api, Nodes, hostFamilies(), decodeJSON(), readBody(), writeJSON(), readSums(), containerAddr() (+2 more)

### Community 17 - ".do"
Cohesion: 0.08
Nodes (23): MeshList, MeshSetRequest, meshSpace, MeshUpdate, NodeAddRequest, NodeSetRequest, SegmentView, ZoneView (+15 more)

### Community 18 - "meshctl/main.go"
Cohesion: 0.15
Nodes (21): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+13 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (24): `internal/fsutil`, Dir, Enrollment, mkdirs(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestJoinRecord(), TestKeysAreCreatedOnceAndKept() (+16 more)

### Community 20 - "ServerConfig"
Cohesion: 0.23
Nodes (26): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+18 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.05
Nodes (54): Agent (`internal/agent/...`), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx() (+46 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (28): Routes, NewTable(), compare(), Table, Sync(), NewTable(), luidOf(), NewTable() (+20 more)

### Community 24 - "NodeInfo"
Cohesion: 0.11
Nodes (19): NodeList, PhoneProfile, addresses(), heard(), ContainerNetwork, NodeInfo, Health(), phoneHealth() (+11 more)

### Community 25 - "Join"
Cohesion: 0.31
Nodes (7): installAgent(), hostService(), installAgent(), JoinOptions, Service, Join(), JoinOnStart()

### Community 26 - "lab"
Cohesion: 0.12
Nodes (18): Several meshes on one host, Coverage, End to end -- `test/e2e`, host, lab, TestDualStack(), deleteNamespaces(), eventually() (+10 more)

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "exprs"
Cohesion: 0.06
Nodes (29): Route, route, Add a command, Add a database migration, Add a validator, Add an agent API route, Dev script tasks, Developer guide (+21 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.09
Nodes (25): DNSEntry, NameEntry, NameSet, 10. Backups, 11. When the control plane is down, 12. Troubleshooting, 13. Upgrading, 1. Run the control plane (Dokploy) (+17 more)

### Community 31 - "Installer"
Cohesion: 0.19
Nodes (11): hostService(), SystemdUnit(), Runner, exists(), Systemd(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere() (+3 more)

### Community 32 - "runner"
Cohesion: 0.09
Nodes (12): pendingApply, runner, safeApply, replayCache, failures, LockoutError, kick(), oneLine() (+4 more)

### Community 33 - "newRig"
Cohesion: 0.11
Nodes (39): brokenStream, The web UI's pages (internal/server/ui), `internal/server/ui`, TestDashboardHealth(), TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes() (+31 more)

### Community 34 - "Service"
Cohesion: 0.14
Nodes (9): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+1 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (60): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+52 more)

### Community 36 - "File"
Cohesion: 0.27
Nodes (5): agentLog(), ttyPassword(), OpenFile(), TestFileRotates(), File

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (29): rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword() (+21 more)

### Community 38 - "`cmd/meshctl`"
Cohesion: 0.07
Nodes (70): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+62 more)

### Community 39 - "Sync"
Cohesion: 0.05
Nodes (27): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+19 more)

### Community 40 - ".exec"
Cohesion: 0.11
Nodes (5): mustAffect(), boolInt(), classify(), changed(), Session

### Community 41 - "openTest"
Cohesion: 0.11
Nodes (53): `internal/server/store`, TestReconcilerVersionsOnlyRealChanges(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestMigrationMakesMeshOne(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+45 more)

### Community 42 - "samples_test.go"
Cohesion: 0.20
Nodes (12): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+4 more)

### Community 43 - "netlinkNet"
Cohesion: 0.10
Nodes (15): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), TestKernelShapes(), toRule() (+7 more)

### Community 44 - "Input"
Cohesion: 0.16
Nodes (12): Record, containerNetwork(), Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable(), sortedByID() (+4 more)

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (44): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), Problem, Compute(), deliver(), deliverAddr() (+36 more)

### Community 46 - "Run"
Cohesion: 0.11
Nodes (14): Deps, Deps, Options, Run(), uiListen(), DNS, Firewall, MeshFirewall (+6 more)

### Community 47 - "`internal/validate`"
Cohesion: 0.12
Nodes (18): Security invariants, `internal/validate`, TestTimeZone(), TestAddrEntry(), TestAddrRangePrefix(), Arg(), POSIXPath(), FuzzArg() (+10 more)

### Community 48 - "README.md"
Cohesion: 0.19
Nodes (7): `internal/tools/vulnjudge`, Test catalogue, Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "Host"
Cohesion: 0.05
Nodes (41): link, exitHost, Host, peer(), TestLinksAddressesAndFailures(), AddrPort(), TestAddrPort(), hexKey() (+33 more)

### Community 50 - "Tx"
Cohesion: 0.17
Nodes (5): ContainerName, scanAll(), splitList(), PeerStat, StatusUpdate

### Community 51 - "Hub"
Cohesion: 0.24
Nodes (3): Hub, subscription, ConfigEvent

### Community 52 - "usagef"
Cohesion: 0.18
Nodes (18): newLogger(), parsePrefixes(), checkAgent(), checkControl(), setupHealthStart(), uninstallAgent(), usagef(), setupControl() (+10 more)

### Community 53 - "LocalStatus"
Cohesion: 0.08
Nodes (24): fakeService, Add a web UI page, Client, Logs, Handler(), logs(), refusal(), laptop() (+16 more)

### Community 54 - "newRig"
Cohesion: 0.20
Nodes (12): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+4 more)

### Community 55 - "errorf"
Cohesion: 0.08
Nodes (39): apiError, IPv6Request, IPv6View, PhoneRequest, Role, addIPv6(), blockOf(), checkName() (+31 more)

### Community 56 - "serveRuntime"
Cohesion: 0.13
Nodes (17): TestOpenRuntime(), clip(), Detect(), meshFilter(), New(), Sockets(), inspectOf(), serveRuntime() (+9 more)

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "Net"
Cohesion: 0.21
Nodes (5): ensureAddr(), Net, Rule, syncDropped(), syncRules()

### Community 59 - "NodeConfig"
Cohesion: 0.05
Nodes (42): NodeDetail, dialing, IPv6Host, Rechecker, Determinism and goldens, Add an apply step, Keys and tokens, `internal/server/phone` (+34 more)

### Community 60 - "meshDNS"
Cohesion: 0.17
Nodes (24): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+16 more)

### Community 62 - "Sync"
Cohesion: 0.14
Nodes (28): network, Sysctls, NewNet(), Gateway(), Guard(), Owned(), parse(), SlotOf() (+20 more)

### Community 63 - ".Apply"
Cohesion: 0.14
Nodes (14): applied, Firewall, policy, Firewall, atLeast(), exitPolicy(), normalize(), relayPolicy() (+6 more)

### Community 64 - "nullInt"
Cohesion: 0.12
Nodes (4): DNSRecord, nullInt(), nullStr(), Version

### Community 65 - "Open"
Cohesion: 0.18
Nodes (16): TestJoin(), TestMeshAddresses(), meshAddresses(), isRefusal(), joinOnStart(), startDir(), TestJoinOnStart(), TestJoinOnStartAsksTheServer() (+8 more)

### Community 66 - ".meshctl"
Cohesion: 0.20
Nodes (13): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Security (+5 more)

### Community 67 - "UI"
Cohesion: 0.11
Nodes (17): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize() (+9 more)

### Community 68 - "WriteEvent"
Cohesion: 0.26
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 69 - "AddrRange"
Cohesion: 0.10
Nodes (20): addrs(), addrText(), contains(), Rule, merged(), portText(), services(), inRanges() (+12 more)

### Community 70 - "Policy"
Cohesion: 0.14
Nodes (24): NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily(), compile() (+16 more)

### Community 71 - "fetchwintun/main.go"
Cohesion: 0.15
Nodes (18): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+10 more)

### Community 72 - "CompilePolicy"
Cohesion: 0.15
Nodes (14): Matrix, MatrixCell, Policy, PolicyRule, Schedule, activeAt(), zoneOf(), CompilePolicy() (+6 more)

### Community 73 - "Mesh"
Cohesion: 0.27
Nodes (7): SoleMesh(), freeSlot(), PickMesh(), pickMesh(), Mesh, newUUID(), scanMesh()

### Community 74 - "New"
Cohesion: 0.20
Nodes (7): newRenderer(), New(), sameOrigin(), TestEveryPageRenders(), handler, page, renderer

### Community 75 - "Of"
Cohesion: 0.22
Nodes (8): fakeRuntime, Container, Of(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack(), TestOfWithIPv6(), ContainerName

### Community 76 - "nrpt.go"
Cohesion: 0.18
Nodes (18): nrpt, nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey() (+10 more)

### Community 77 - "ToSnapshot"
Cohesion: 0.23
Nodes (11): SettingsRequest, ToSnapshot(), checkDNS(), IP(), IPv6(), DNSName(), Domains(), IPv4() (+3 more)

### Community 78 - "fail"
Cohesion: 0.20
Nodes (12): plain6(), Prefix(), Prefix6(), TestTimestamp(), Timestamp(), TimeZone(), fail(), Prefix4() (+4 more)

### Community 79 - "net/netip.Addr"
Cohesion: 0.08
Nodes (26): hostAddrs, nameTable, zone, Apply (agent), TestIsPublic(), gateway(), host(), syncProxies() (+18 more)

### Community 80 - "resolved"
Cohesion: 0.19
Nodes (8): Domain, fakeRegistrar, linkDomain, Registration, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.10
Nodes (22): RecordRequest, segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange (+14 more)

### Community 82 - "net/netip.AddrPort"
Cohesion: 0.24
Nodes (11): ParseAddr(), TestParseAddr(), Endpoint6(), PeerEndpoint(), TestIPv6Grammar(), Endpoint(), Key32(), mask() (+3 more)

### Community 83 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+2 more)

### Community 84 - "TestPageGoldens"
Cohesion: 0.08
Nodes (24): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), readZoneForm(), ruleRows(), scheduleText(), tabsIn() (+16 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.23
Nodes (21): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+13 more)

### Community 86 - "stop"
Cohesion: 0.12
Nodes (9): uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows(), handler (+1 more)

### Community 87 - "meshctl/docs_test.go"
Cohesion: 0.25
Nodes (12): cell(), envDoc(), renderCommands(), sentence(), startFlag(), TestCommandsDoc(), TestContainersDocNamesRealVariables(), fromVar() (+4 more)

### Community 90 - "net/http.Request"
Cohesion: 0.08
Nodes (22): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), anyText() (+14 more)

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 94 - "Ruleset"
Cohesion: 0.10
Nodes (10): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel (+2 more)

### Community 95 - "ServerURL"
Cohesion: 0.22
Nodes (8): Container-mode agent, cleanHostname(), CheckServerDomains(), Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase(), TestServerURL()

### Community 96 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 98 - "userspace"
Cohesion: 0.09
Nodes (6): lookup(), NewUserspace(), defaultInterface(), Link, links, userspace

### Community 101 - "Rollback"
Cohesion: 0.67
Nodes (3): statusDetail, Rollback, sameRollback()

### Community 102 - "objectForm"
Cohesion: 0.67
Nodes (3): objectForm, objectRow, objectsView

### Community 104 - "Schedule"
Cohesion: 0.14
Nodes (11): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), weekday(), ServiceEntry() (+3 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.06
Nodes (44): `internal/proto`, newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestNamesRoundTrip(), TestPrefsRoundTrip() (+36 more)

### Community 106 - "context.Context"
Cohesion: 0.07
Nodes (32): Access, Config, DB, Lockout, Params, Pending, Sealer, newPassword() (+24 more)

### Community 113 - "Engine"
Cohesion: 0.21
Nodes (4): nameServer, DNS, Want, MeshName

### Community 114 - "Name"
Cohesion: 0.16
Nodes (14): check(), validateMeshNames(), checkGroups(), newNode(), validateSet(), Ref(), TestRef(), Enum() (+6 more)

### Community 118 - ".queryRow"
Cohesion: 0.13
Nodes (4): scanNode(), HashToken(), scanUser(), scanner

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.07
Nodes (41): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+33 more)

### Community 122 - ".Reach"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 130 - "Slot"
Cohesion: 0.09
Nodes (21): Pipe(), TestPolicyHasKernelSemantics(), Holder(), Priorities(), SlotOf(), Taken(), TestHolder(), TestSlotOf() (+13 more)

### Community 133 - "log/slog.Logger"
Cohesion: 0.06
Nodes (33): Change, Deps, TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), NewNodes(), New() (+25 more)

### Community 135 - "stream"
Cohesion: 0.33
Nodes (4): dial(), dial(), isPipe(), stream

### Community 138 - "ns"
Cohesion: 0.18
Nodes (8): lab, TestContainerNetworks(), ns(), lab, TestRelayFailover(), lab, stream, TestMeshFirewall()

### Community 146 - "io.Writer"
Cohesion: 0.14
Nodes (13): watchdog, dnsLine(), printExit(), printMesh(), printChange(), printEnrollment(), bytesize(), printStatus() (+5 more)

## Knowledge Gaps
- **8 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `harness`, `NO_COLOR`, `lab` (+3 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 360 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `WGKey`, `Slot`, `htmx.min.js`, `log/slog.Logger`, `archtest.go`, `net.Listener`, `Node`, `Dir`, `ServerConfig`, `exprs`, `.mesh`, `Installer`, `runner`, `agent/agent_test.go`, `File`, ``cmd/meshctl``, `Sync`, `netlinkNet`, `Run`, `README.md`, `Host`, `LocalStatus`, `newRig`, `serveRuntime`, `NodeConfig`, `meshDNS`, `Sync`, `Open`, `.meshctl`, `UI`, `WriteEvent`, `AddrRange`, `Policy`, `fetchwintun/main.go`, `Of`, `nrpt.go`, `net/netip.Addr`, `resolved`, `net/netip.AddrPort`, `stop`, ``internal/proto``?**
  _High betweenness centrality (0.034) - this node is a cross-community bridge._
- **Are the 266 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 266 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `harness` to the rest of the system?**
  _8 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09105581519374623 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to ``cmd/meshctl``, `.mesh`?**
  _High betweenness centrality (0.033) - this node is a cross-community bridge._
- **Should ``internal/server/api`` be split into smaller, more focused modules?**
  _Cohesion score 0.05690312738367658 - nodes in this community are weakly interconnected._
- **Why does ``cmd/meshctl`` connect ``cmd/meshctl`` to `Service`, `htmx.min.js`, `README.md`, `meshctl/docs_test.go`, `net/http.Request`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._