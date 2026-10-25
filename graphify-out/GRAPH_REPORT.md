# Graph Report - vpn-mesh-utility  (2026-10-25)

## Corpus Check
- 296 files · ~359,566 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3725 nodes · 18547 edges · 131 communities (100 shown, 31 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2245 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ff3d3065`
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
- ToSnapshot
- github.com/hafio/vpn-mesh-utility
- Node
- otherNet
- net/http.ResponseWriter
- .do
- meshctl/main.go
- Dir
- ServerConfig
- testing.T
- net/netip.Prefix
- Net
- ui
- Join
- lab
- Snapshot
- exprs
- Command reference
- .mesh
- host
- runner
- newRig
- Service
- agent/agent_test.go
- io.Writer
- `internal/server/auth`
- `cmd/meshctl`
- Policy
- .exec
- openTest
- samples_test.go
- netlinkNet
- Input
- planner_test.go
- Engine
- .meshctl
- README.md
- Host
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
- Sysctl
- CookieValue
- Agent protocol
- UI
- WriteEvent
- AddrRange
- Compare
- fetchwintun/main.go
- nullInt
- .AddMesh
- .routes
- Container
- nrpt_windows.go
- validate.go
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
- TestContainerNetworks
- loadMigrations
- userspace
- build
- capped
- noKernel
- Ruleset
- handler
- Schedule
- `internal/proto`
- context.Context
- AuditEntry
- AdminError
- toAPIError
- StepError
- .SetSegmentation
- TestMain
- Engine
- validate_test.go
- Exit
- .relayFor
- Error
- NewNet
- same
- `internal/server` and `internal/server/export`
- NewNet
- NewKernel
- Refusal
- Table
- NewKernel
- StatusError
- NewKernel
- Run
- TestMeshFirewall
- io.Reader

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 255 edges
2. `ui` - 104 edges
3. ``internal/server/api`` - 86 edges
4. `Session` - 85 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 67 edges
7. `env` - 65 edges
8. `newHarness()` - 60 edges
9. `runner` - 47 edges
10. ``internal/server/store`` - 47 edges

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
- 4-file cycle: `internal/agent/agentfake/agentfake.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/uapi_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`
- 4-file cycle: `internal/agent/agentfake/agentfake_test.go -> internal/agent/wg/wg_test.go -> internal/agent/wg/userspace_windows.go -> internal/agent/routes/routes_test.go -> internal/agent/agentfake/agentfake_test.go`

## Communities (131 total, 31 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (18): entry, linkDNS, DNSRecord, FirewallRule, Node, NewKernel(), logsData, inspected (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.16
Nodes (35): rig, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS() (+27 more)

### Community 2 - "env"
Cohesion: 0.13
Nodes (40): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+32 more)

### Community 3 - "`internal/server/api`"
Cohesion: 0.06
Nodes (102): countingDB, `internal/server/api`, ErrorBody, cli(), ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput() (+94 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (117): Status, 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names (+109 more)

### Community 5 - "flag.FlagSet"
Cohesion: 0.24
Nodes (27): setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), setupMeshAdd(), setupMeshList(), setupMeshRemove(), setupMeshSet() (+19 more)

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
Cohesion: 0.17
Nodes (10): bindBoth(), bindPair(), dial(), Listen(), dial(), isPipe(), Listen(), listenSocket() (+2 more)

### Community 12 - "ToSnapshot"
Cohesion: 0.19
Nodes (19): check(), ParseAddr(), TestParseAddr(), newNode(), validateSet(), meshCIDR(), ToSnapshot(), Endpoint6() (+11 more)

### Community 14 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.15
Nodes (9): api, decode(), reply(), hostFamilies(), decodeJSON(), readBody(), writeJSON(), readSums() (+1 more)

### Community 17 - ".do"
Cohesion: 0.07
Nodes (28): MeshSetRequest, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, ReachRequest, SegmentView (+20 more)

### Community 18 - "meshctl/main.go"
Cohesion: 0.12
Nodes (25): agentCommands(), commands(), passwords(), runWith(), TestUserAccessCommands(), firewallCommands(), healthCommands(), find() (+17 more)

### Community 19 - "Dir"
Cohesion: 0.11
Nodes (19): Dir, Enrollment, mkdirs(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestJoinRecord(), TestKeysAreCreatedOnceAndKept(), TestLoadKeysRefusesLooseOrDamagedFiles() (+11 more)

### Community 20 - "ServerConfig"
Cohesion: 0.24
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+17 more)

### Community 21 - "testing.T"
Cohesion: 0.04
Nodes (65): Agent (`internal/agent/...`), TestJoin(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx(), TestConfigureHasKernelSemantics(), TestKernelHoldsOneTable(), TestNetHasKernelSemantics() (+57 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (25): Routes, index(), route(), compare(), luidOf(), onLink(), Allocate(), Allocate6() (+17 more)

### Community 24 - "ui"
Cohesion: 0.11
Nodes (12): Session, listOf(), meshChanged(), objectRows(), title(), atoi(), changed(), message() (+4 more)

### Community 25 - "Join"
Cohesion: 0.20
Nodes (10): hostService(), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), cleanHostname(), Service, Join(), meshAddresses() (+2 more)

### Community 26 - "lab"
Cohesion: 0.15
Nodes (16): Coverage, End to end -- `test/e2e`, host, lab, TestDualStack(), deleteNamespaces(), eventually(), lab (+8 more)

### Community 27 - "Snapshot"
Cohesion: 0.19
Nodes (7): SnapNode, Snapshot, snapNode(), describeDNS(), fields(), fingerprint(), names()

### Community 28 - "exprs"
Cohesion: 0.17
Nodes (8): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), TestIntervalAddrElems(), TestIntervalElems(), nftOp

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.18
Nodes (13): 10. Backups, 11. When the control plane is down, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits, 5. Day to day (+5 more)

### Community 31 - "host"
Cohesion: 0.19
Nodes (9): SystemdUnit(), Runner, exists(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere(), TestInstallWritesTheUnitAndStartsIt(), Installer (+1 more)

### Community 32 - "runner"
Cohesion: 0.07
Nodes (20): Deps, pendingApply, runner, safeApply, replayCache, failures, LockoutError, 12. Troubleshooting (+12 more)

### Community 33 - "newRig"
Cohesion: 0.08
Nodes (49): brokenStream, Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, TestTemplatesHaveNoScript(), TestDashboardHealth(), TestFirewallPages(), TestNodeGroupsOnPages() (+41 more)

### Community 34 - "Service"
Cohesion: 0.14
Nodes (9): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+1 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.09
Nodes (59): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+51 more)

### Community 36 - "io.Writer"
Cohesion: 0.11
Nodes (17): agentLog(), TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), setupDNS(), printMesh(), printEnrollment(), ago() (+9 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (34): Params, rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants() (+26 more)

### Community 38 - "`cmd/meshctl`"
Cohesion: 0.06
Nodes (91): TestAgentRunRuntimeSocket(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin(), TestControlFlagErrors() (+83 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.11
Nodes (5): newUUID(), mustAffect(), boolInt(), classify(), changed()

### Community 41 - "openTest"
Cohesion: 0.11
Nodes (57): `internal/server/store`, addMesh(), TestMeshes(), TestMeshesAreApart(), TestMigrationMakesMeshOne(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh(), TestServerSettings() (+49 more)

### Community 42 - "samples_test.go"
Cohesion: 0.20
Nodes (12): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+4 more)

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (12): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), TestKernelShapes(), toRule() (+4 more)

### Community 44 - "Input"
Cohesion: 0.09
Nodes (25): PlanError, Record, activeAt(), zoneOf(), Check(), checkContainers(), checkContainers6(), checkNode6() (+17 more)

### Community 45 - "planner_test.go"
Cohesion: 0.14
Nodes (40): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), deliver(), deliverAddr(), encode(), firstDiff() (+32 more)

### Community 46 - "Engine"
Cohesion: 0.11
Nodes (10): Deps, dialing, IPv6Host, Rechecker, Add an apply step, DNS, Engine, Firewall (+2 more)

### Community 47 - ".meshctl"
Cohesion: 0.19
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens (+6 more)

### Community 48 - "README.md"
Cohesion: 0.19
Nodes (7): `internal/tools/vulnjudge`, Test catalogue, Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "Host"
Cohesion: 0.06
Nodes (36): link, exitHost, Host, peer(), TestLinksAddressesAndFailures(), hexKey(), ParseUAPI(), getAnswer() (+28 more)

### Community 50 - "Tx"
Cohesion: 0.16
Nodes (5): ContainerName, scanAll(), splitList(), PeerStat, StatusUpdate

### Community 51 - "log/slog.Logger"
Cohesion: 0.06
Nodes (25): Change, Failover, Hub, hubMove, hubState, subscription, hitWindow, waitApplied() (+17 more)

### Community 52 - "PublishNames"
Cohesion: 0.22
Nodes (10): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), PublishNames(), DNSView (+2 more)

### Community 53 - "LocalStatus"
Cohesion: 0.07
Nodes (24): syncBuffer, statusDetail, printExit(), Client, Logs, logs(), refusal(), laptop() (+16 more)

### Community 54 - "newRig"
Cohesion: 0.26
Nodes (11): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+3 more)

### Community 55 - "MeshTx"
Cohesion: 0.08
Nodes (42): IPv6Request, IPv6View, PhoneRequest, Role, addIPv6(), blockOf(), checkGroups(), checkName() (+34 more)

### Community 56 - "serveRuntime"
Cohesion: 0.21
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 57 - "Client"
Cohesion: 0.11
Nodes (14): Options, apiError(), clip(), Client, Join(), New(), TestJoinIsUnsigned(), TestUnreachableAndMalformedAnswers() (+6 more)

### Community 58 - "Net"
Cohesion: 0.15
Nodes (6): ensureAddr(), Net, Rule, syncDropped(), syncProxies(), syncRules()

### Community 59 - "NodeConfig"
Cohesion: 0.09
Nodes (28): Determinism and goldens, `internal/server/phone`, Shared fixtures -- `internal/testfix`, forHost(), proxied(), TestForHostWithoutIPv6(), NodeConfig, goldenConfig() (+20 more)

### Community 60 - "meshDNS"
Cohesion: 0.27
Nodes (20): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceAnswersOverIPv6() (+12 more)

### Community 61 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 62 - "Sync"
Cohesion: 0.19
Nodes (24): Sysctls, NewNet(), Gateway(), Guard(), Owned(), parse(), Sync(), TestGatewayRules() (+16 more)

### Community 63 - ".Apply"
Cohesion: 0.16
Nodes (11): applied, Firewall, policy, Firewall, normalize(), richRule(), RuleArgv(), wantedPolicies() (+3 more)

### Community 64 - "Sysctl"
Cohesion: 0.33
Nodes (3): Sysctl, Sysctls, Sysctl

### Community 65 - "CookieValue"
Cohesion: 0.18
Nodes (7): TestCookieAndCSRF(), ClientIP(), inAny(), TestClientIP(), ClearCookie(), CookieValue(), SetCookie()

### Community 66 - "Agent protocol"
Cohesion: 0.15
Nodes (13): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+5 more)

### Community 67 - "UI"
Cohesion: 0.11
Nodes (16): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize() (+8 more)

### Community 68 - "WriteEvent"
Cohesion: 0.26
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 69 - "AddrRange"
Cohesion: 0.12
Nodes (17): addrs(), contains(), ofFamily(), inRanges(), serves(), inRange(), AddrEntry(), AddrRange (+9 more)

### Community 70 - "Compare"
Cohesion: 0.35
Nodes (11): `internal/server/versions`, Compare(), describeSegmentation(), setDiff(), sortedKeys(), server(), snap(), TestARenameIsAChangeNotAReplacement() (+3 more)

### Community 71 - "fetchwintun/main.go"
Cohesion: 0.15
Nodes (18): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+10 more)

### Community 72 - "nullInt"
Cohesion: 0.13
Nodes (4): DNSRecord, nullInt(), nullStr(), Version

### Community 73 - ".AddMesh"
Cohesion: 0.07
Nodes (24): MeshList, meshSpace, MeshUpdate, SoleMesh(), ULAPrefix(), checkApart(), clash(), freeCIDR() (+16 more)

### Community 74 - ".routes"
Cohesion: 0.42
Nodes (7): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 75 - "Container"
Cohesion: 0.36
Nodes (3): fakeRuntime, Container, ContainerName

### Community 76 - "nrpt_windows.go"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 77 - "validate.go"
Cohesion: 0.12
Nodes (25): RecordRequest, Container-mode agent, validateMeshNames(), CheckServerDomains(), checkDNS(), IP(), DNSName(), Domains() (+17 more)

### Community 78 - "Extending the code"
Cohesion: 0.25
Nodes (8): Add a command, Add a database migration, Add a validator, Developer guide, Extending the code, Goldens, Releases, Toolchain

### Community 79 - "net/netip.Addr"
Cohesion: 0.10
Nodes (18): network, hostAddrs, nameTable, zone, TestIsPublic(), proxy(), records(), firstAddr() (+10 more)

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.10
Nodes (21): segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange, meshChange() (+13 more)

### Community 82 - "conn"
Cohesion: 0.43
Nodes (3): conn(), hasLegacy(), nftKernel

### Community 84 - ".renderPolicies"
Cohesion: 0.08
Nodes (22): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), readPolicyForm(), readZoneForm(), ruleRows() (+14 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.23
Nodes (21): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+13 more)

### Community 86 - "stop"
Cohesion: 0.18
Nodes (8): hostService(), uninstallAgent(), commandLine(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows(), SCM

### Community 87 - "meshctl/docs_test.go"
Cohesion: 0.16
Nodes (18): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), shownDefault(), cell() (+10 more)

### Community 90 - "ui"
Cohesion: 0.07
Nodes (19): ago(), before(), versionNumber(), staticHandler(), writeHTML(), headers(), New(), sameOrigin() (+11 more)

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 98 - "userspace"
Cohesion: 0.09
Nodes (8): lookup(), NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, links, userspace

### Community 99 - "build"
Cohesion: 0.40
Nodes (3): Dev script tasks, build(), nftBatch

### Community 102 - "Ruleset"
Cohesion: 0.20
Nodes (7): Ruleset, admit(), walk(), Chain, Enforcer, Kernel, packet

### Community 104 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.05
Nodes (68): `internal/proto`, newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestNamesRoundTrip(), TestPrefsRoundTrip() (+60 more)

### Community 106 - "context.Context"
Cohesion: 0.06
Nodes (31): fakeService, Access, Config, DB, Lockout, Pending, Sealer, resetTOTP() (+23 more)

### Community 113 - "Engine"
Cohesion: 0.15
Nodes (6): nameServer, DNS, Want, newNameTable(), DNS, MeshName

### Community 114 - "validate_test.go"
Cohesion: 0.12
Nodes (28): `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), FuzzArg(), FuzzName(), TestArg() (+20 more)

### Community 120 - "same"
Cohesion: 0.20
Nodes (13): Domain, fakeRegistrar, nrptRule, Registration, desired(), domainName(), same(), TestSameRegistration() (+5 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.09
Nodes (34): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+26 more)

### Community 126 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 133 - "Run"
Cohesion: 0.10
Nodes (21): Deps, NewNodes(), New(), NewAdmin(), newAPI(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries() (+13 more)

### Community 138 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

### Community 146 - "io.Reader"
Cohesion: 0.25
Nodes (5): watchdog, setupImport(), Read(), ReadSegmentation(), TestSegmentationFile()

## Knowledge Gaps
- **7 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `harness`, `NO_COLOR`, `lab` (+2 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 361 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **31 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `env`, `htmx.min.js`, `archtest.go`, `ToSnapshot`, `Node`, `Dir`, `ServerConfig`, `Join`, `exprs`, `.mesh`, `host`, `runner`, `newRig`, `agent/agent_test.go`, `io.Writer`, ``cmd/meshctl``, `Policy`, `netlinkNet`, `Engine`, `README.md`, `Host`, `log/slog.Logger`, `newRig`, `serveRuntime`, `Client`, `NodeConfig`, `meshDNS`, `Sync`, `UI`, `WriteEvent`, `AddrRange`, `fetchwintun/main.go`, `.routes`, `nrpt_windows.go`, `net/netip.Addr`, `resolved`, `stop`, `capped`, ``internal/proto``, `Engine`, `same`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 254 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 254 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `harness` to the rest of the system?**
  _7 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08244232628816135 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to ``cmd/meshctl``, `.mesh`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Should `env` be split into smaller, more focused modules?**
  _Cohesion score 0.1282051282051282 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect ``internal/server/api`` to `runner`, `newRig`, `htmx.min.js`, `Run`, `.routes`, `README.md`, `.do`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._