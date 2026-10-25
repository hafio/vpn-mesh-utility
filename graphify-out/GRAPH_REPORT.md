# Graph Report - vpn-mesh-utility  (2026-10-25)

## Corpus Check
- 290 files · ~341,957 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3606 nodes · 17718 edges · 133 communities (102 shown, 31 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2093 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fef59dab`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- env
- `internal/server/api`
- htmx.min.js
- io.Writer
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
- services
- github.com/hafio/vpn-mesh-utility
- Node
- Hub
- net/http.ResponseWriter
- NodeChange
- commands.go
- Dir
- ServerConfig
- Agent (`internal/agent/...`)
- net/netip.Prefix
- Net
- Session
- Join
- lab
- Snapshot
- Container
- Command reference
- time.Time
- usagef
- runner
- newRig
- Service
- agent/agent_test.go
- File
- `internal/server/auth`
- testing.T
- Policy
- .exec
- openTest
- otherNet
- netlinkNet
- Input
- planner_test.go
- meshctl/main.go
- .mesh
- README.md
- Host
- alloc.go
- log/slog.Logger
- Service
- LocalStatus
- samples_test.go
- Node
- serveRuntime
- Client
- Extending the code
- Test catalogue
- meshDNS
- Of
- Sync
- .Apply
- CompilePolicy
- .queryRow
- setupFirewallApply
- UI
- Registration
- .add
- kernel
- fetchwintun/main.go
- ToSnapshot
- Security
- .show
- Policy
- service_test.go
- Error
- syncBuffer
- AddrRange
- resolved
- errorf
- PublishNames
- 8. The web UI
- .renderPolicies
- deploy_test.go
- service_windows.go
- Ruleset
- exprs
- Diff
- Node
- loadMigrations
- Compare
- userspace
- .Apply
- Agent protocol
- .routes
- .nodeView
- nullInt
- Schedule
- `internal/proto`
- net/netip.Addr
- conn
- AdminError
- .Sync
- StepError
- build
- TestMain
- DNS
- validate_test.go
- noKernel
- entry
- Error
- Version
- NewNet
- Segmentation
- `internal/server` and `internal/server/export`
- .SetSegmentation
- NewKernel
- NewKernel
- Refusal
- PeerStat
- PlanError
- NewKernel
- NewNet
- NewKernel
- StatusError
- ns

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 255 edges
2. `ui` - 89 edges
3. ``internal/server/api`` - 82 edges
4. `Session` - 73 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 67 edges
7. `newHarness()` - 58 edges
8. `env` - 57 edges
9. `runner` - 47 edges
10. ``internal/server/store`` - 46 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `A node's container on the host network (the NAS)` --references--> `TestContainerSamples()`  [INFERRED]
  docs/security.md → deploy/samples_test.go
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

## Communities (133 total, 31 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, DNSRequest (+9 more)

### Community 1 - "NodeConfig"
Cohesion: 0.11
Nodes (44): rig, TestProxyLoopAppliesAgain(), TestRunNeedsAnEnrollment(), NewHost(), NewRoutes(), TestLinksAddressesAndFailures(), IsStepError(), New() (+36 more)

### Community 2 - "env"
Cohesion: 0.26
Nodes (26): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), setupVersion() (+18 more)

### Community 3 - "`internal/server/api`"
Cohesion: 0.06
Nodes (90): brokenStream, countingDB, `internal/server/api`, ErrorBody, cli(), ErrorMessage(), adminCall(), TestAdminAddIssuesAToken() (+82 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (114): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+106 more)

### Community 5 - "io.Writer"
Cohesion: 0.14
Nodes (25): TestPrintEnrollmentInTheLocalZone(), dnsLine(), setupFirewallApply(), setupIPv6(), adminFlags(), cleared(), joinNonEmpty(), nodeAction() (+17 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (33): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem, 2. Goals (+25 more)

### Community 7 - "segResolver"
Cohesion: 0.17
Nodes (13): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), parseSpans(), PrefixRange(), Ref() (+5 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.15
Nodes (15): Rule, Violation, Packages, and which way they may import, Input and commands, `internal/archtest`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue (+7 more)

### Community 11 - "net.Listener"
Cohesion: 0.48
Nodes (5): bindBoth(), bindPair(), Listen(), Listen(), listenSocket()

### Community 12 - "services"
Cohesion: 0.23
Nodes (10): addrText(), Rule, merged(), portText(), services(), admit(), walk(), packet (+2 more)

### Community 14 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 15 - "Hub"
Cohesion: 0.19
Nodes (5): Change, Hub, subscription, ConfigEvent, Result

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.15
Nodes (10): api, decode(), reply(), ETag(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+2 more)

### Community 17 - "NodeChange"
Cohesion: 0.16
Nodes (10): IPv6Request, IPv6View, NodeAddRequest, NodeSetRequest, SegmentView, ZoneView, AdminClient, nodePath() (+2 more)

### Community 18 - "commands.go"
Cohesion: 0.20
Nodes (15): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), parsePrefixes(), printFlags(), shownDefault() (+7 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (23): `internal/fsutil`, Dir, Enrollment, mkdirs(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestJoinRecord(), TestKeysAreCreatedOnceAndKept() (+15 more)

### Community 20 - "ServerConfig"
Cohesion: 0.24
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+17 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.06
Nodes (44): Agent (`internal/agent/...`), TestDefaultSocket(), TestRestartNeedsAConfig(), pfx(), TestConfigureHasKernelSemantics(), TestKernelHoldsOneTable(), TestNetHasKernelSemantics(), TestPolicyHasKernelSemantics() (+36 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.10
Nodes (16): Routes, gateway(), host(), index(), NewTable(), route(), compare(), Table (+8 more)

### Community 24 - "Session"
Cohesion: 0.11
Nodes (12): Session, listOf(), meshChanged(), atoi(), before(), changed(), versionNumber(), message() (+4 more)

### Community 25 - "Join"
Cohesion: 0.13
Nodes (22): TestCleanHostname(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), cleanHostname(), Service, Join(), meshAddresses() (+14 more)

### Community 26 - "lab"
Cohesion: 0.18
Nodes (10): Coverage, End to end -- `test/e2e`, host, eventually(), lab, stream, inNS(), TestServerMesh() (+2 more)

### Community 27 - "Snapshot"
Cohesion: 0.21
Nodes (6): SnapNode, Snapshot, snapNode(), describeDNS(), fields(), fingerprint()

### Community 28 - "Container"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns, meshctl exit (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.06
Nodes (26): pendingApply, safeApply, Failover, hubMove, hubState, replayCache, failures, hitWindow (+18 more)

### Community 31 - "usagef"
Cohesion: 0.19
Nodes (21): newLogger(), checkAgent(), checkControl(), setupHealthStart(), uninstallAgent(), usagef(), fromVar(), logLevel() (+13 more)

### Community 32 - "runner"
Cohesion: 0.07
Nodes (16): Deps, fakeService, runner, kick(), oneLine(), Run(), TestOneLine(), TestReportKeepsToTheServersLimits() (+8 more)

### Community 33 - "newRig"
Cohesion: 0.11
Nodes (43): The web UI's pages (internal/server/ui), `internal/server/ui`, TestDashboardHealth(), TestFirewallPages(), TestNodeGroupsOnPages(), group(), bytesize(), dnsState() (+35 more)

### Community 34 - "Service"
Cohesion: 0.14
Nodes (9): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+1 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (59): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+51 more)

### Community 36 - "File"
Cohesion: 0.15
Nodes (11): agentLog(), printExit(), installAgent(), installAgent(), ttyPassword(), JoinOptions, OpenFile(), TestFileRotates() (+3 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (29): rig, 11. CLI reference, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword() (+21 more)

### Community 38 - "testing.T"
Cohesion: 0.05
Nodes (98): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+90 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.13
Nodes (4): newUUID(), mustAffect(), boolInt(), classify()

### Community 41 - "openTest"
Cohesion: 0.11
Nodes (54): `internal/server/store`, addMesh(), TestMeshes(), TestMeshesAreApart(), TestMigrationMakesMeshOne(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh(), TestServerSettings() (+46 more)

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (13): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+5 more)

### Community 44 - "Input"
Cohesion: 0.15
Nodes (12): Record, containerNetwork(), Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable(), sortedByID() (+4 more)

### Community 45 - "planner_test.go"
Cohesion: 0.12
Nodes (46): Add a topology rule, `internal/server/planner`, TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan(), TestCheckMeshCIDR(), TestCheckRequested(), TestRangeFor(), Problem (+38 more)

### Community 46 - "meshctl/main.go"
Cohesion: 0.20
Nodes (15): agentCommands(), commands(), runWith(), firewallCommands(), healthCommands(), find(), isHelp(), main() (+7 more)

### Community 47 - ".mesh"
Cohesion: 0.18
Nodes (13): 10. Backups, 11. When the control plane is down, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits, 5. Day to day (+5 more)

### Community 48 - "README.md"
Cohesion: 0.13
Nodes (11): Determinism and goldens, graphify, Style, What this repo is, You do not run the gates, TestHashIsPinned(), Building, Documentation (+3 more)

### Community 49 - "Host"
Cohesion: 0.07
Nodes (35): link, exitHost, Host, peer(), AddrPort(), TestAddrPort(), hexKey(), ParseUAPI() (+27 more)

### Community 50 - "alloc.go"
Cohesion: 0.26
Nodes (15): TestULAPrefix(), ULAPrefix(), Allocate(), Allocate6(), AllocateContainers(), AllocateContainers6(), Block(), Block6() (+7 more)

### Community 51 - "log/slog.Logger"
Cohesion: 0.06
Nodes (34): Deps, TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), NewNodes(), New(), NewAdmin() (+26 more)

### Community 52 - "Service"
Cohesion: 0.09
Nodes (29): Config, DB, Lockout, Params, Pending, Sealer, addUser(), newPassword() (+21 more)

### Community 53 - "LocalStatus"
Cohesion: 0.08
Nodes (26): statusDetail, Add a web UI page, Client, Logs, logs(), refusal(), laptop(), TestTemplatesHaveNoScript() (+18 more)

### Community 54 - "samples_test.go"
Cohesion: 0.22
Nodes (11): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+3 more)

### Community 55 - "Node"
Cohesion: 0.11
Nodes (21): NodeList, PhoneProfile, PhoneRequest, addresses(), heard(), checkName(), defaultMTU(), Nodes (+13 more)

### Community 56 - "serveRuntime"
Cohesion: 0.21
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "Extending the code"
Cohesion: 0.25
Nodes (8): Add a command, Add a database migration, Add a validator, Developer guide, Extending the code, Goldens, Releases, Toolchain

### Community 59 - "Test catalogue"
Cohesion: 0.10
Nodes (21): NodeDetail, Keys and tokens, `internal/server/phone`, NodeConfig, ContainerNetwork, Hub(), joinSet(), NewKey() (+13 more)

### Community 60 - "meshDNS"
Cohesion: 0.25
Nodes (22): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+14 more)

### Community 61 - "Of"
Cohesion: 0.18
Nodes (10): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+2 more)

### Community 62 - "Sync"
Cohesion: 0.09
Nodes (30): Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard(), Owned() (+22 more)

### Community 63 - ".Apply"
Cohesion: 0.12
Nodes (14): applied, Firewall, policy, Firewall, Firewall, Runner, atLeast(), normalize() (+6 more)

### Community 64 - "CompilePolicy"
Cohesion: 0.22
Nodes (12): Matrix, MatrixCell, activeAt(), zoneOf(), CompilePolicy(), newSegResolver(), Overlaps(), Summarize() (+4 more)

### Community 65 - ".queryRow"
Cohesion: 0.08
Nodes (10): Mesh, scanMesh(), ContainerName, scanAll(), splitList(), HashToken(), NewToken(), scanUser() (+2 more)

### Community 66 - "setupFirewallApply"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 67 - "UI"
Cohesion: 0.10
Nodes (17): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize() (+9 more)

### Community 68 - "Registration"
Cohesion: 0.14
Nodes (18): Domain, fakeRegistrar, nrpt, nrptRule, Registration, desired(), domainName(), fromRules() (+10 more)

### Community 69 - ".add"
Cohesion: 0.18
Nodes (5): syncBuffer, segLiteral, clip(), Ring, segLiterals()

### Community 70 - "kernel"
Cohesion: 0.16
Nodes (5): Status, TestCreateLinkNeedsPrivilege(), lookup(), kernel, links

### Community 71 - "fetchwintun/main.go"
Cohesion: 0.15
Nodes (20): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+12 more)

### Community 72 - "ToSnapshot"
Cohesion: 0.09
Nodes (48): dialing, check(), ParseAddr(), TestParseAddr(), validateMeshNames(), checkGroups(), newNode(), validateSet() (+40 more)

### Community 73 - "Security"
Cohesion: 0.15
Nodes (13): A node's container on the host network (the NAS), Agent service (systemd), Changes a node asks for, Container-mode agent, DNS, Interfaces the agent did not create, Security, The host firewall (+5 more)

### Community 74 - ".show"
Cohesion: 0.15
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), sameOrigin(), handler

### Community 75 - "Policy"
Cohesion: 0.16
Nodes (24): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily() (+16 more)

### Community 76 - "service_test.go"
Cohesion: 0.23
Nodes (10): hostService(), SystemdUnit(), exists(), Systemd(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere(), TestInstallWritesTheUnitAndStartsIt() (+2 more)

### Community 79 - "AddrRange"
Cohesion: 0.14
Nodes (14): addrs(), contains(), inRanges(), serves(), AddrEntry(), AddrRange, Service, MergeRanges() (+6 more)

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.13
Nodes (18): RecordRequest, segDoc, Actor, PolicyZone(), MeshChange, meshChange(), movePolicy(), place() (+10 more)

### Community 82 - "PublishNames"
Cohesion: 0.17
Nodes (13): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+5 more)

### Community 83 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

### Community 84 - ".renderPolicies"
Cohesion: 0.07
Nodes (24): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+16 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.23
Nodes (21): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+13 more)

### Community 86 - "service_windows.go"
Cohesion: 0.11
Nodes (10): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows() (+2 more)

### Community 87 - "Ruleset"
Cohesion: 0.21
Nodes (4): watchingFirewall, Kernel, Ruleset, Chain

### Community 90 - "exprs"
Cohesion: 0.17
Nodes (8): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), TestIntervalAddrElems(), TestIntervalElems(), nftOp

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 94 - "Node"
Cohesion: 0.12
Nodes (29): apiError, ReachRequest, Role, addIPv6(), blockOf(), containerSubnet(), containerSubnet6(), moveMeshIP() (+21 more)

### Community 96 - "Compare"
Cohesion: 0.30
Nodes (12): `internal/server/versions`, Compare(), describeSegmentation(), names(), setDiff(), sortedKeys(), server(), snap() (+4 more)

### Community 98 - "userspace"
Cohesion: 0.09
Nodes (7): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace, wgdev

### Community 99 - ".Apply"
Cohesion: 0.08
Nodes (15): Deps, IPv6Host, Rechecker, Add an apply step, Sysctl, forHost(), Engine, MeshFirewall (+7 more)

### Community 100 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+2 more)

### Community 101 - ".routes"
Cohesion: 0.42
Nodes (7): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 102 - ".nodeView"
Cohesion: 0.17
Nodes (8): Health(), ago(), writeHTML(), dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 103 - "nullInt"
Cohesion: 0.18
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 104 - "Schedule"
Cohesion: 0.16
Nodes (10): The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds() (+2 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.05
Nodes (59): Options, `internal/proto`, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents() (+51 more)

### Community 106 - "net/netip.Addr"
Cohesion: 0.10
Nodes (20): network, hostAddrs, nameTable, zone, TestIsPublic(), proxied(), records(), firstAddr() (+12 more)

### Community 107 - "conn"
Cohesion: 0.43
Nodes (3): conn(), hasLegacy(), nftKernel

### Community 111 - "build"
Cohesion: 0.40
Nodes (3): Dev script tasks, build(), nftBatch

### Community 113 - "DNS"
Cohesion: 0.17
Nodes (6): nameServer, DNS, Want, newNameTable(), DNS, MeshName

### Community 114 - "validate_test.go"
Cohesion: 0.08
Nodes (26): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), TestRef(), TestServiceEntry(), Arg() (+18 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.10
Nodes (33): `internal/server` and `internal/server/export`, NewAdminClient(), FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport() (+25 more)

### Community 138 - "ns"
Cohesion: 0.13
Nodes (13): lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), newLab(), ns(), tail() (+5 more)

## Knowledge Gaps
- **7 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `harness`, `NO_COLOR`, `lab` (+2 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 349 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **31 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `on()` connect `htmx.min.js` to `testing.T`, `kernel`, `.mesh`?**
  _High betweenness centrality (0.040) - this node is a cross-community bridge._
- **Are the 254 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 254 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `harness` to the rest of the system?**
  _7 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08350084126204196 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect ``internal/server/api`` to `newRig`, `htmx.min.js`, `.routes`, `testing.T`, ``internal/server/auth``, ``internal/proto``, `archtest.go`, `NodeChange`, `alloc.go`, `log/slog.Logger`, `PublishNames`?**
  _High betweenness centrality (0.033) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.11400966183574879 - nodes in this community are weakly interconnected._
- **Why does `at()` connect `htmx.min.js` to ``internal/server/api``?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._