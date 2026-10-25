# Graph Report - vpn-mesh-utility  (2026-10-26)

## Corpus Check
- 311 files · ~377,608 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3852 nodes · 20121 edges · 139 communities (109 shown, 30 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2354 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7dbfcec0`
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
- .nodeView
- service_test.go
- lab
- Snapshot
- build
- Command reference
- .mesh
- `internal/server/api`
- context.Context
- newRig
- Service
- agent/agent_test.go
- sync.Mutex
- `internal/server/auth`
- testing.T
- netlinkPolicy
- .exec
- openTest
- samples_test.go
- netlinkNet
- Input
- planner_test.go
- Run
- usagef
- README.md
- Host
- Tx
- newRig
- .Reach
- LocalStatus
- exprs
- errorf
- .SecondFactor
- Client
- Net
- PhoneConfig
- meshDNS
- .add
- Sync
- .Apply
- nullInt
- Allocate
- .meshctl
- UI
- .Apply
- Policy
- time.Time
- fetchwintun/main_test.go
- .SetExit
- .AddMesh
- .show
- Of
- nrpt.go
- Sync
- PublishNames
- net/netip.Addr
- resolved
- errorf
- kernel
- Agent protocol
- TestPageGoldens
- deploy_test.go
- service_windows.go
- File
- net/http.Request
- log/slog.Logger
- Ruleset
- .rollback
- WriteEvent
- userspace
- otherPolicy
- .BeginEnrollment
- NewReader
- .Sync
- IPv6View
- Schedule
- `internal/proto`
- context.Context
- wgdev
- AdminError
- Signed
- StepError
- officeHours
- TestMain
- Engine
- fail
- Segmentation
- Registration
- Error
- .queryRow
- NewNet
- policy
- `internal/server` and `internal/server/export`
- TestContainerNetworks
- NewNet
- NewKernel
- Refusal
- tabsIn
- zonesView
- Slot
- StatusError
- NewKernel
- Run
- stream
- .relayFor
- TestMeshFirewall
- entry
- Error
- usageError
- io.Reader

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 272 edges
2. `ui` - 104 edges
3. ``internal/server/api`` - 88 edges
4. `Session` - 85 edges
5. `Slot` - 79 edges
6. `ServerConfig()` - 76 edges
7. `NodeConfig` - 66 edges
8. `env` - 65 edges
9. `newHarness()` - 63 edges
10. `runner` - 49 edges

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

## Communities (139 total, 30 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, DNSRequest (+9 more)

### Community 1 - "WGKey"
Cohesion: 0.07
Nodes (50): dialing, Rechecker, rig, Determinism and goldens, Add an apply step, TestProxyLoopAppliesAgain(), TestRunNeedsTheDataDirToItself(), NewHost() (+42 more)

### Community 2 - "env"
Cohesion: 0.23
Nodes (30): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), setupVersion() (+22 more)

### Community 3 - "`internal/server/api`"
Cohesion: 0.07
Nodes (76): apiError, brokenStream, countingDB, `internal/server/api`, ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput() (+68 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (114): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+106 more)

### Community 5 - "flag.FlagSet"
Cohesion: 0.12
Nodes (35): dnsLine(), printExit(), setupFirewallApply(), setupFirewallShow(), setupIPv6(), printUsage(), printMesh(), setupMeshAdd() (+27 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.16
Nodes (14): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), Overlaps(), parseSpans(), TestOverlaps() (+6 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.18
Nodes (13): Rule, Violation, Input and commands, `internal/archtest`, `internal/tools/vulnjudge`, Test catalogue, Check(), importsOf() (+5 more)

### Community 11 - "net.Listener"
Cohesion: 0.21
Nodes (9): Listen(), Listen(), listenSocket(), Listen(), New(), ParseAddr(), TestParseAddr(), TestServeAndListen() (+1 more)

### Community 12 - "sync.Mutex"
Cohesion: 0.12
Nodes (12): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, TestJudgeHubs(), RateLimiter (+4 more)

### Community 14 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.12
Nodes (15): api, signed(), decode(), reply(), ETag(), ParseSigned(), Definitive(), SignRefusal() (+7 more)

### Community 17 - ".do"
Cohesion: 0.11
Nodes (18): MeshSetRequest, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, SettingsRequest, statusDetail (+10 more)

### Community 18 - "meshctl/main.go"
Cohesion: 0.12
Nodes (29): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+21 more)

### Community 19 - "Dir"
Cohesion: 0.06
Nodes (43): `internal/fsutil`, sign(), TestAnswersMustBeSigned(), TestConfigHonoursTheETag(), TestDataDirIsProtectedOnWindows(), Dir, Enrollment, NewKeys() (+35 more)

### Community 20 - "ServerConfig"
Cohesion: 0.21
Nodes (27): Shared fixtures -- `internal/testfix`, fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+19 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.04
Nodes (75): Options, Agent (`internal/agent/...`), TestCleanHostname(), TestIsPublic(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestRestartNeedsAConfig() (+67 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.11
Nodes (14): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+6 more)

### Community 24 - ".nodeView"
Cohesion: 0.13
Nodes (11): Health(), group(), ago(), bytesize(), writeHTML(), TestHelpers(), dashboardData, nodeForm (+3 more)

### Community 25 - "service_test.go"
Cohesion: 0.13
Nodes (14): hostService(), hostService(), TestUnitPosture(), SystemdUnit(), Runner, Service, exists(), Systemd() (+6 more)

### Community 26 - "lab"
Cohesion: 0.13
Nodes (19): Several meshes on one host, Coverage, End to end -- `test/e2e`, host, lab, TestDualStack(), deleteNamespaces(), eventually() (+11 more)

### Community 27 - "Snapshot"
Cohesion: 0.07
Nodes (33): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+25 more)

### Community 28 - "build"
Cohesion: 0.29
Nodes (4): build(), conn(), nftBatch, nftKernel

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.17
Nodes (14): 10. Backups, 11. When the control plane is down, 12. Troubleshooting, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits (+6 more)

### Community 31 - "`internal/server/api`"
Cohesion: 0.14
Nodes (26): cli(), ErrorMessage(), TestIPv6DNSRecords(), TestReportedHostFamilies(), harness, TestMeshAdmin(), TestMeshesStayApart(), TestNodesServiceInAMesh() (+18 more)

### Community 32 - "context.Context"
Cohesion: 0.08
Nodes (15): fakeService, pendingApply, runner, safeApply, kick(), oneLine(), TestOneLine(), TestReportKeepsToTheServersLimits() (+7 more)

### Community 33 - "newRig"
Cohesion: 0.10
Nodes (45): Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, TestTemplatesHaveNoScript(), TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes() (+37 more)

### Community 34 - "Service"
Cohesion: 0.17
Nodes (6): resolver, Service, bindBoth(), bindPair(), exchange(), serve()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (62): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+54 more)

### Community 36 - "sync.Mutex"
Cohesion: 0.22
Nodes (3): syncBuffer, Sysctl, syncBuffer

### Community 37 - "`internal/server/auth`"
Cohesion: 0.13
Nodes (27): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide() (+19 more)

### Community 38 - "testing.T"
Cohesion: 0.06
Nodes (88): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+80 more)

### Community 39 - "netlinkPolicy"
Cohesion: 0.18
Nodes (9): Family, netlinkPolicy, dump(), family(), owned(), TestKernelShapes(), TestOwnsOnlyItsPriorities(), toRoute() (+1 more)

### Community 40 - ".exec"
Cohesion: 0.08
Nodes (9): newUUID(), mustAffect(), joinList(), boolInt(), classify(), changed(), Session, PeerStat (+1 more)

### Community 41 - "openTest"
Cohesion: 0.10
Nodes (57): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestMigrationMakesMeshOne(), TestNodeIDsAreNeverReused() (+49 more)

### Community 42 - "samples_test.go"
Cohesion: 0.18
Nodes (12): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+4 more)

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (13): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+5 more)

### Community 44 - "Input"
Cohesion: 0.11
Nodes (24): PlanError, Record, Policy, containers6(), Check(), checkContainers(), checkContainers6(), checkDNS() (+16 more)

### Community 45 - "planner_test.go"
Cohesion: 0.14
Nodes (41): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), deliver(), deliverAddr(), encode(), firstDiff() (+33 more)

### Community 46 - "Run"
Cohesion: 0.13
Nodes (10): Deps, Deps, Run(), uiListen(), DNS, Firewall, MeshFirewall, Sysctls (+2 more)

### Community 47 - "usagef"
Cohesion: 0.15
Nodes (21): newLogger(), parsePrefixes(), checkAgent(), checkControl(), setupHealthStart(), installAgent(), uninstallAgent(), installAgent() (+13 more)

### Community 48 - "README.md"
Cohesion: 0.14
Nodes (10): graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together (+2 more)

### Community 49 - "Host"
Cohesion: 0.05
Nodes (44): link, exitHost, IPv6Host, Host, peer(), TestLinksAddressesAndFailures(), forHost(), proxied() (+36 more)

### Community 50 - "Tx"
Cohesion: 0.11
Nodes (10): SoleMesh(), PickMesh(), pickMesh(), Mesh, scanMesh(), ContainerName, scanAll(), splitList() (+2 more)

### Community 51 - "newRig"
Cohesion: 0.17
Nodes (13): capped, TestCappedKeepsTheHead(), golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage() (+5 more)

### Community 52 - ".Reach"
Cohesion: 0.14
Nodes (11): ReachRequest, SegmentView, ZoneView, activeAt(), flowEnd(), flowEnds(), ReachResult, zoneOf() (+3 more)

### Community 53 - "LocalStatus"
Cohesion: 0.10
Nodes (20): Client, Logs, logs(), refusal(), laptop(), PrefsRequest, PrefsResponse, LocalMesh (+12 more)

### Community 54 - "exprs"
Cohesion: 0.17
Nodes (8): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), TestIntervalAddrElems(), TestIntervalElems(), nftOp

### Community 55 - "errorf"
Cohesion: 0.09
Nodes (33): Role, addIPv6(), blockOf(), checkGroups(), containerNetwork(), containerSubnet(), containerSubnet6(), defaultMTU() (+25 more)

### Community 56 - ".SecondFactor"
Cohesion: 0.26
Nodes (6): Lockout, Pending, fullExpiry(), load(), session(), matchStep()

### Community 57 - "Client"
Cohesion: 0.22
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "Net"
Cohesion: 0.13
Nodes (9): ensureAddr(), Net, Rule, Owned(), SlotOf(), syncDropped(), syncProxies(), syncRules() (+1 more)

### Community 59 - "PhoneConfig"
Cohesion: 0.23
Nodes (15): `internal/server/phone`, Hub(), joinSet(), NewKey(), Render(), aliceKey(), dualPhone(), golden() (+7 more)

### Community 60 - "meshDNS"
Cohesion: 0.25
Nodes (22): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+14 more)

### Community 61 - ".add"
Cohesion: 0.10
Nodes (21): clip(), Ring, addrs(), addrText(), contains(), Rule, merged(), portText() (+13 more)

### Community 62 - "Sync"
Cohesion: 0.19
Nodes (25): Sysctls, NewNet(), Gateway(), Guard(), parse(), Sync(), TestGatewayRules(), TestGatewayRulesWithIPv6() (+17 more)

### Community 63 - ".Apply"
Cohesion: 0.20
Nodes (9): applied, Firewall, Firewall, normalize(), richRule(), RuleArgv(), Firewall, FWRule (+1 more)

### Community 64 - "nullInt"
Cohesion: 0.12
Nodes (4): DNSRecord, nullInt(), nullStr(), Version

### Community 65 - "Allocate"
Cohesion: 0.22
Nodes (15): TestULAPrefix(), Allocate(), Allocate6(), AllocateContainers(), AllocateContainers6(), Block(), Block6(), CheckMeshCIDR() (+7 more)

### Community 66 - ".meshctl"
Cohesion: 0.19
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens (+6 more)

### Community 67 - "UI"
Cohesion: 0.12
Nodes (13): hostOf(), oldest(), validOrNot(), ago(), bytesize(), prefsFlash(), LocalToken(), handler (+5 more)

### Community 68 - ".Apply"
Cohesion: 0.16
Nodes (13): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+5 more)

### Community 69 - "Policy"
Cohesion: 0.14
Nodes (24): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily() (+16 more)

### Community 70 - "time.Time"
Cohesion: 0.14
Nodes (10): replayCache, failures, LockoutError, serves(), phoneHealth(), record(), Clock, Service (+2 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.06
Nodes (38): `internal/tools/fetchwintun`, TestOpenRuntime(), clip(), Detect(), meshFilter(), New(), Sockets(), inspectOf() (+30 more)

### Community 72 - ".SetExit"
Cohesion: 0.20
Nodes (9): Holder(), Taken(), TestHolder(), adapterPart(), adapterSlot(), TestAdapterGUID(), adapterGUID(), exitRoutes() (+1 more)

### Community 73 - ".AddMesh"
Cohesion: 0.09
Nodes (21): Change, Hub, MeshList, meshSpace, MeshUpdate, subscription, ConfigEvent, ULAPrefix() (+13 more)

### Community 74 - ".show"
Cohesion: 0.11
Nodes (9): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), New(), sameOrigin(), handler, meshHandler (+1 more)

### Community 75 - "Of"
Cohesion: 0.36
Nodes (3): fakeRuntime, Container, ContainerName

### Community 76 - "nrpt.go"
Cohesion: 0.18
Nodes (18): nrpt, nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey() (+10 more)

### Community 77 - "Sync"
Cohesion: 0.10
Nodes (17): Policy, mutate(), Desired(), Policy, Route, Rule, V6Support, Sync() (+9 more)

### Community 78 - "PublishNames"
Cohesion: 0.16
Nodes (13): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+5 more)

### Community 79 - "net/netip.Addr"
Cohesion: 0.11
Nodes (18): network, hostAddrs, nameTable, zone, gateway(), host(), records(), hostAddr() (+10 more)

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.09
Nodes (26): PhoneRequest, RecordRequest, segLiteral, segDoc, checkName(), Actor, problemStrings(), segLiterals() (+18 more)

### Community 82 - "kernel"
Cohesion: 0.16
Nodes (5): Status, TestCreateLinkNeedsPrivilege(), lookup(), kernel, links

### Community 83 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+2 more)

### Community 84 - "TestPageGoldens"
Cohesion: 0.13
Nodes (12): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), scheduleText(), dayBox, policiesView, policyForm (+4 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.24
Nodes (20): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+12 more)

### Community 86 - "service_windows.go"
Cohesion: 0.11
Nodes (11): uninstallAgent(), main(), ttyPassword(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths() (+3 more)

### Community 87 - "File"
Cohesion: 0.19
Nodes (8): agentLog(), OpenFile(), TestFileRotates(), lockFile(), unlockFile(), lockFile(), unlockFile(), File

### Community 90 - "net/http.Request"
Cohesion: 0.08
Nodes (19): Session, anyText(), listOf(), listText(), meshChanged(), objectRows(), readZoneForm(), ruleRows() (+11 more)

### Community 93 - "log/slog.Logger"
Cohesion: 0.22
Nodes (6): New(), TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), Options

### Community 94 - "Ruleset"
Cohesion: 0.12
Nodes (8): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, noKernel, packet

### Community 96 - "WriteEvent"
Cohesion: 0.46
Nodes (6): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry()

### Community 98 - "userspace"
Cohesion: 0.12
Nodes (4): NewUserspace(), defaultInterface(), Link, userspace

### Community 100 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 101 - "NewReader"
Cohesion: 0.25
Nodes (6): NewReader(), TestSSEReaderBoundsLineLength(), TestSSEReaderJoinsDataLinesAndDefaultsName(), TestSSEReaderReportsTruncatedEvent(), Event, Reader

### Community 103 - "IPv6View"
Cohesion: 0.38
Nodes (3): IPv6Request, IPv6View, ipv6View()

### Community 104 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.13
Nodes (24): `internal/proto`, TestETagQuotesTheHash(), TestJoinRequestErrorNeverEchoesTheToken(), TestJoinRequestValidateNormalises(), TestJoinRequestValidateRejects(), TestPrefsRequestValidate(), TestStatusReportValidate(), validJoin() (+16 more)

### Community 106 - "context.Context"
Cohesion: 0.10
Nodes (22): Access, Config, DB, Params, resetTOTP(), DB, Service, keepAdmin() (+14 more)

### Community 111 - "officeHours"
Cohesion: 0.50
Nodes (4): officeHours(), sgt(), TestPolicyRuleString(), TestPolicySchedules()

### Community 113 - "Engine"
Cohesion: 0.13
Nodes (8): nameServer, DNS, Want, newNameTable(), DNS, TestMeshViewValidate(), MeshName, MeshNames()

### Community 114 - "fail"
Cohesion: 0.05
Nodes (73): Security invariants, Container-mode agent, `internal/validate`, check(), validateMeshNames(), newNode(), validateSet(), meshCIDR() (+65 more)

### Community 116 - "Registration"
Cohesion: 0.22
Nodes (7): Domain, fakeRegistrar, Registrar, Registration, NewRegistrar(), NewRegistrar(), NewRegistrar()

### Community 118 - ".queryRow"
Cohesion: 0.14
Nodes (4): HashToken(), NewToken(), User, scanUser()

### Community 120 - "policy"
Cohesion: 0.54
Nodes (4): policy, exitPolicy(), relayPolicy(), wantedPolicies()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.08
Nodes (40): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+32 more)

### Community 126 - "tabsIn"
Cohesion: 0.67
Nodes (3): tabsIn(), fwPage, fwTab

### Community 130 - "Slot"
Cohesion: 0.13
Nodes (12): Pipe(), TestPolicyHasKernelSemantics(), Priorities(), SlotOf(), TestSlotOf(), NewKernel(), NewKernel(), SysctlKeys() (+4 more)

### Community 133 - "Run"
Cohesion: 0.08
Nodes (31): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+23 more)

### Community 135 - "stream"
Cohesion: 0.33
Nodes (4): dial(), dial(), isPipe(), stream

### Community 138 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

### Community 146 - "io.Reader"
Cohesion: 0.22
Nodes (6): watchdog, Read(), ReadSegmentation(), TestSegmentationFile(), Write(), WriteSegmentation()

## Knowledge Gaps
- **8 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `harness`, `NO_COLOR`, `lab` (+3 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 363 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **30 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `on()` connect `htmx.min.js` to `testing.T`, `kernel`, `.mesh`?**
  _High betweenness centrality (0.034) - this node is a cross-community bridge._
- **Are the 271 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 271 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `harness` to the rest of the system?**
  _8 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09489229765013055 - nodes in this community are weakly interconnected._
- **Why does `Command reference` connect `Command reference` to `README.md`, `meshctl/main.go`, `.meshctl`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.07412587412587412 - nodes in this community are weakly interconnected._
- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `WGKey`, `Slot`, `Run`, `archtest.go`, `net.Listener`, `Node`, `Dir`, `ServerConfig`, `service_test.go`, `.mesh`, `context.Context`, `newRig`, `agent/agent_test.go`, `testing.T`, `netlinkPolicy`, `netlinkNet`, `Host`, `newRig`, `exprs`, `meshDNS`, `Sync`, `.meshctl`, `.Apply`, `Policy`, `time.Time`, `fetchwintun/main_test.go`, `.SetExit`, `nrpt.go`, `Sync`, `net/netip.Addr`, `resolved`, `kernel`, `service_windows.go`, `File`, `log/slog.Logger`, `WriteEvent`, `Engine`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._