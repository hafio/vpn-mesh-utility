# Graph Report - vpn-mesh-utility  (2026-10-25)

## Corpus Check
- 293 files · ~351,326 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3673 nodes · 18232 edges · 146 communities (114 shown, 32 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2189 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b4eb30b8`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
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
- meshHarness
- github.com/hafio/vpn-mesh-utility
- Node
- Open
- net/http.ResponseWriter
- .do
- commands.go
- Dir
- ServerConfig
- testing.T
- net/netip.Prefix
- Net
- Session
- Join
- lab
- Snapshot
- Container
- Command reference
- time.Time
- storeNode
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
- otherNet
- netlinkNet
- .reachesAServer
- planner_test.go
- meshctl/main.go
- .mesh
- README.md
- Host
- Mesh
- Hub
- Service
- LocalStatus
- samples_test.go
- MeshTx
- serveRuntime
- Client
- Net
- Test catalogue
- meshDNS
- sync.Mutex
- Sync
- .Apply
- Exit
- .queryRow
- Nodes
- UI
- Registration
- .add
- CookieValue
- fetchwintun/main.go
- fail
- Segmentation
- WriteEvent
- Policy
- nrpt_windows.go
- Error
- kernel
- net/netip.Addr
- resolved
- errorf
- PublishNames
- 7. Agent
- .renderPolicies
- deploy_test.go
- stop
- Ruleset
- Extending the code
- Diff
- toAPIError
- loadMigrations
- TestRunGivesTheMeshIPv6
- userspace
- serve
- nullInt
- .routes
- .nodeView
- Role
- Schedule
- `internal/proto`
- nameTable
- newRig
- AdminError
- wgdev
- StepError
- renderer
- TestMain
- Engine
- validate_test.go
- DefaultPolicy
- entry
- Error
- Version
- NewNet
- same
- `internal/server` and `internal/server/export`
- Host
- NewKernel
- NewKernel
- Refusal
- PeerStat
- TestFailoverMovesRelays
- NewKernel
- Matcher
- NewKernel
- StatusError
- boolInt
- Run
- Security
- .Apply
- .Reach
- CheckMeshCIDR6
- ns
- printStatus
- github.com/miekg/dns.Msg
- TestMigrationTurnsRulesIntoPolicies
- fakeService
- Container-mode agent
- NewNet
- .InsertMesh

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 255 edges
2. `ui` - 89 edges
3. ``internal/server/api`` - 86 edges
4. `Session` - 73 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 67 edges
7. `env` - 62 edges
8. `newHarness()` - 60 edges
9. `runner` - 47 edges
10. `api` - 46 edges

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

## Communities (146 total, 32 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, DNSRequest (+9 more)

### Community 1 - "NodeConfig"
Cohesion: 0.06
Nodes (69): IPv6Host, rig, Determinism and goldens, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), forHost(), IsStepError() (+61 more)

### Community 2 - "env"
Cohesion: 0.14
Nodes (37): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+29 more)

### Community 3 - "`internal/server/api`"
Cohesion: 0.11
Nodes (31): ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing() (+23 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (114): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+106 more)

### Community 5 - "flag.FlagSet"
Cohesion: 0.23
Nodes (28): setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), setupMeshAdd(), setupMeshList(), setupMeshRemove(), setupMeshSet() (+20 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (27): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+19 more)

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
Cohesion: 0.14
Nodes (16): Rule, Violation, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, Input and commands (+8 more)

### Community 11 - "net.Listener"
Cohesion: 0.33
Nodes (6): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), serve()

### Community 12 - "meshHarness"
Cohesion: 0.16
Nodes (26): `internal/server/api`, ErrorMessage(), TestIPv6DNSRecords(), harness, TestMeshAdmin(), TestMeshesAreApartOnTheAgentAPI(), TestMeshesStayApart(), TestNodesServiceInAMesh() (+18 more)

### Community 14 - "Node"
Cohesion: 0.24
Nodes (7): Peer, endpoints(), Node, sortPeers(), sortPrefixes(), fam, plan

### Community 15 - "Open"
Cohesion: 0.16
Nodes (22): `internal/server` and `internal/server/export`, TestToSnapshotReadsAnOlderExport(), TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), Import(), document(), linuxOnly() (+14 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.16
Nodes (9): api, decode(), reply(), cli(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+1 more)

### Community 17 - ".do"
Cohesion: 0.13
Nodes (11): IPv6Request, IPv6View, MeshAddRequest, MeshSetRequest, MeshUpdate, SettingsRequest, AdminClient, nodePath() (+3 more)

### Community 18 - "commands.go"
Cohesion: 0.16
Nodes (20): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), setupVersion(), shownDefault() (+12 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (23): `internal/fsutil`, Dir, Enrollment, mkdirs(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestJoinRecord(), TestKeysAreCreatedOnceAndKept() (+15 more)

### Community 20 - "ServerConfig"
Cohesion: 0.21
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+17 more)

### Community 21 - "testing.T"
Cohesion: 0.03
Nodes (86): Agent (`internal/agent/...`), TestDefaultSocket(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx(), TestConfigureHasKernelSemantics() (+78 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.11
Nodes (14): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+6 more)

### Community 24 - "Session"
Cohesion: 0.11
Nodes (11): Session, listOf(), meshChanged(), atoi(), before(), changed(), versionNumber(), message() (+3 more)

### Community 25 - "Join"
Cohesion: 0.13
Nodes (24): TestCleanHostname(), TestIsPublic(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), cleanHostname(), firstAddr(), Join() (+16 more)

### Community 26 - "lab"
Cohesion: 0.27
Nodes (6): Coverage, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (19): `internal/server/versions`, LegacySegmentation(), SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), fields() (+11 more)

### Community 28 - "Container"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (39): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns, meshctl exit (+31 more)

### Community 30 - "time.Time"
Cohesion: 0.05
Nodes (33): pendingApply, safeApply, Failover, hubMove, hubState, replayCache, Deps, failures (+25 more)

### Community 31 - "storeNode"
Cohesion: 0.19
Nodes (21): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+13 more)

### Community 32 - "runner"
Cohesion: 0.10
Nodes (11): Deps, runner, kick(), oneLine(), Run(), DNS, Result, drain() (+3 more)

### Community 33 - "newRig"
Cohesion: 0.10
Nodes (41): Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, capped, TestFirewallPages(), TestNodeGroupsOnPages(), newQR(), TestPhonePage() (+33 more)

### Community 34 - "Service"
Cohesion: 0.15
Nodes (10): Registrar, resolver, Service, 4. Spokes and exits, Container mode, New(), NewRegistrar(), NewRegistrar() (+2 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.11
Nodes (60): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+52 more)

### Community 36 - "io.Writer"
Cohesion: 0.13
Nodes (13): agentLog(), TestPrintEnrollmentInTheLocalZone(), printExit(), setupExit(), printMesh(), printEnrollment(), ttyPassword(), OpenFile() (+5 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (27): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures() (+19 more)

### Community 38 - "`cmd/meshctl`"
Cohesion: 0.07
Nodes (81): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+73 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (18): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+10 more)

### Community 41 - "openTest"
Cohesion: 0.15
Nodes (35): `internal/server/store`, addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh(), TestServerSettings(), TestTokenMesh() (+27 more)

### Community 43 - "netlinkNet"
Cohesion: 0.13
Nodes (10): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+2 more)

### Community 45 - "planner_test.go"
Cohesion: 0.09
Nodes (57): Add a topology rule, `internal/server/planner`, activeAt(), zoneOf(), Check(), checkContainers(), checkContainers6(), sortedByID() (+49 more)

### Community 46 - "meshctl/main.go"
Cohesion: 0.16
Nodes (18): agentCommands(), commands(), runWith(), firewallCommands(), healthCommands(), find(), isHelp(), main() (+10 more)

### Community 47 - ".mesh"
Cohesion: 0.07
Nodes (32): Agent protocol, Apply (agent), Config and events, Container names, Errors, Join, Local API, Mesh view (+24 more)

### Community 48 - "README.md"
Cohesion: 0.23
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "Host"
Cohesion: 0.09
Nodes (17): link, exitHost, Host, peer(), AddrPort(), hexKey(), ParseUAPI(), uapiInt() (+9 more)

### Community 50 - "Mesh"
Cohesion: 0.09
Nodes (21): MeshList, meshSpace, SoleMesh(), checkApart(), clash(), freeCIDR(), freeSlot(), MeshInfo (+13 more)

### Community 51 - "Hub"
Cohesion: 0.16
Nodes (8): Hub, subscription, ConfigEvent, NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 52 - "Service"
Cohesion: 0.09
Nodes (31): Config, DB, Lockout, Params, Pending, Sealer, addUser(), newPassword() (+23 more)

### Community 53 - "LocalStatus"
Cohesion: 0.07
Nodes (25): statusDetail, Client, Logs, logs(), refusal(), Listen(), New(), laptop() (+17 more)

### Community 54 - "samples_test.go"
Cohesion: 0.27
Nodes (10): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+2 more)

### Community 55 - "MeshTx"
Cohesion: 0.09
Nodes (35): NodeAddRequest, NodeSetRequest, segLiteral, checkGroups(), containerNetwork(), containerSubnet(), containerSubnet6(), defaultMTU() (+27 more)

### Community 56 - "serveRuntime"
Cohesion: 0.19
Nodes (13): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+5 more)

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "Net"
Cohesion: 0.20
Nodes (4): ensureAddr(), Net, syncDropped(), syncRules()

### Community 59 - "Test catalogue"
Cohesion: 0.16
Nodes (19): Keys and tokens, `internal/server/phone`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue, Hub(), joinSet(), NewKey() (+11 more)

### Community 60 - "meshDNS"
Cohesion: 0.24
Nodes (22): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceAnswersOverIPv6() (+14 more)

### Community 61 - "sync.Mutex"
Cohesion: 0.12
Nodes (7): fakeRuntime, syncBuffer, Sysctl, Sysctls, Container, Sysctl, syncBuffer

### Community 62 - "Sync"
Cohesion: 0.19
Nodes (24): Sysctls, NewNet(), Gateway(), Guard(), Owned(), parse(), Sync(), TestGatewayRules() (+16 more)

### Community 63 - ".Apply"
Cohesion: 0.15
Nodes (12): applied, Firewall, policy, Firewall, atLeast(), normalize(), richRule(), RuleArgv() (+4 more)

### Community 65 - ".queryRow"
Cohesion: 0.11
Nodes (4): HashToken(), NewToken(), scanUser(), AuditEntry

### Community 66 - "Nodes"
Cohesion: 0.11
Nodes (15): NodeDetail, NodeList, PhoneProfile, addresses(), heard(), PeerStatus, Nodes, nodeInfo() (+7 more)

### Community 67 - "UI"
Cohesion: 0.15
Nodes (11): hostOf(), oldest(), validOrNot(), ago(), bytesize(), prefsFlash(), LocalToken(), handler (+3 more)

### Community 68 - "Registration"
Cohesion: 0.33
Nodes (5): Domain, fakeRegistrar, Registration, serverZone(), TestServerZone()

### Community 69 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 70 - "CookieValue"
Cohesion: 0.14
Nodes (9): TestCookieAndCSRF(), ClientIP(), inAny(), TestClientIP(), ClearCookie(), CookieValue(), SetCookie(), sameOrigin() (+1 more)

### Community 71 - "fetchwintun/main.go"
Cohesion: 0.14
Nodes (20): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestInstall(), TestExporterRunExportsAtNight() (+12 more)

### Community 72 - "fail"
Cohesion: 0.13
Nodes (33): check(), ParseAddr(), validateMeshNames(), newNode(), validateSet(), meshCIDR(), ToSnapshot(), checkDNS() (+25 more)

### Community 73 - "Segmentation"
Cohesion: 0.11
Nodes (11): SegmentView, ZoneView, segDoc, plannerSegmentation(), PolicyZone(), policyNames(), Segmentation, SegGroup (+3 more)

### Community 74 - "WriteEvent"
Cohesion: 0.25
Nodes (7): TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 75 - "Policy"
Cohesion: 0.17
Nodes (23): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+15 more)

### Community 76 - "nrpt_windows.go"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 78 - "kernel"
Cohesion: 0.13
Nodes (5): Status, lookup(), Link, kernel, links

### Community 79 - "net/netip.Addr"
Cohesion: 0.08
Nodes (28): network, hostAddrs, syncProxies(), isPublic(), isPublic6(), addrs(), contains(), ofFamily() (+20 more)

### Community 80 - "resolved"
Cohesion: 0.35
Nodes (4): linkDomain, resolved, classify(), foreignNetns()

### Community 81 - "errorf"
Cohesion: 0.13
Nodes (20): PhoneRequest, RecordRequest, addIPv6(), checkName(), Actor, nodeByName(), problemStrings(), errorf() (+12 more)

### Community 82 - "PublishNames"
Cohesion: 0.21
Nodes (10): DNSEntry, NameEntry, NameSet, compareAddr(), LoadNames(), PublishNames(), TestPublishNames(), DNSView (+2 more)

### Community 83 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 84 - ".renderPolicies"
Cohesion: 0.07
Nodes (24): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+16 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.20
Nodes (23): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+15 more)

### Community 86 - "stop"
Cohesion: 0.08
Nodes (14): hostService(), hostService(), uninstallAgent(), Runner, Service, exists(), Systemd(), commandLine() (+6 more)

### Community 87 - "Ruleset"
Cohesion: 0.05
Nodes (22): watchingFirewall, Dev script tasks, Kernel, build(), conn(), exprs(), hasLegacy(), ifname() (+14 more)

### Community 90 - "Extending the code"
Cohesion: 0.29
Nodes (7): Add a command, Add a validator, Developer guide, Extending the code, Goldens, Releases, Toolchain

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 94 - "toAPIError"
Cohesion: 0.40
Nodes (4): apiError, toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion()

### Community 95 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 96 - "TestRunGivesTheMeshIPv6"
Cohesion: 0.22
Nodes (5): dial(), dial(), isPipe(), TestRunGivesTheMeshIPv6(), stream

### Community 98 - "userspace"
Cohesion: 0.13
Nodes (5): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), userspace

### Community 99 - "serve"
Cohesion: 0.19
Nodes (13): brokenStream, serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums() (+5 more)

### Community 100 - "nullInt"
Cohesion: 0.16
Nodes (4): Add a database migration, DNSRecord, nullInt(), nullStr()

### Community 101 - ".routes"
Cohesion: 0.36
Nodes (8): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented(), TestReportedHostFamilies()

### Community 102 - ".nodeView"
Cohesion: 0.13
Nodes (13): TestJudgeHubs(), group(), ago(), bytesize(), dnsState(), phoneState(), writeHTML(), TestHelpers() (+5 more)

### Community 103 - "Role"
Cohesion: 0.20
Nodes (15): Role, blockOf(), Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), Range, RangeFor() (+7 more)

### Community 104 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.06
Nodes (49): Options, `internal/proto`, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents() (+41 more)

### Community 106 - "nameTable"
Cohesion: 0.26
Nodes (3): nameTable, zone, records()

### Community 107 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 111 - "renderer"
Cohesion: 0.29
Nodes (4): loadPages(), pages, page, renderer

### Community 113 - "Engine"
Cohesion: 0.10
Nodes (8): nameServer, dialing, Rechecker, Add an apply step, DNS, Engine, Want, MeshName

### Community 114 - "validate_test.go"
Cohesion: 0.06
Nodes (34): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), TestRef(), TestServiceEntry(), Arg() (+26 more)

### Community 115 - "DefaultPolicy"
Cohesion: 0.29
Nodes (10): DefaultPolicy(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSetNodeGroups(), TestSetSegmentationRefusesAStranger(), DecodeSnapshot() (+2 more)

### Community 120 - "same"
Cohesion: 0.29
Nodes (10): nrptRule, desired(), domainName(), same(), TestSameRegistration(), fromRules(), nrptKey(), ownedKey() (+2 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.12
Nodes (16): watchdog, FromSnapshot(), Document, Read(), ReadSegmentation(), dualSnapshot(), legacy(), snapshot() (+8 more)

### Community 122 - "Host"
Cohesion: 0.36
Nodes (3): Host, SyncAddrs(), SyncLink()

### Community 127 - "TestFailoverMovesRelays"
Cohesion: 0.50
Nodes (3): harness, TestFailoverMovesRelays(), TestFailoverRun()

### Community 129 - "Matcher"
Cohesion: 0.29
Nodes (4): serves(), Service, Matcher, matchRule

### Community 133 - "Run"
Cohesion: 0.20
Nodes (14): Deps, NewNodes(), New(), NewAdmin(), newAPI(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries() (+6 more)

### Community 134 - "Security"
Cohesion: 0.22
Nodes (9): Changes a node asks for, DNS, Interfaces the agent did not create, Security, The host firewall, The local UI (internal/agent/localui), Web UI sign-in (internal/server/auth), What a node sees of the others (+1 more)

### Community 135 - ".Apply"
Cohesion: 0.32
Nodes (4): Change, PlanError, Result, Problem

### Community 136 - ".Reach"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 137 - "CheckMeshCIDR6"
Cohesion: 0.48
Nodes (7): TestULAPrefix(), ULAPrefix(), Allocate6(), AllocateContainers6(), Block6(), CheckMeshCIDR6(), TestIPv6AddressPlan()

### Community 138 - "ns"
Cohesion: 0.11
Nodes (16): End to end -- `test/e2e`, host, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), inNS() (+8 more)

### Community 139 - "printStatus"
Cohesion: 0.40
Nodes (5): dnsLine(), setupDNS(), bytesize(), printStatus(), setupStatus()

### Community 141 - "TestMigrationTurnsRulesIntoPolicies"
Cohesion: 0.67
Nodes (4): TestMigrationMakesMeshOne(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies()

### Community 143 - "Container-mode agent"
Cohesion: 0.67
Nodes (3): Container-mode agent, CheckServerDomains(), TestCheckServerDomains()

## Knowledge Gaps
- **7 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `harness`, `NO_COLOR`, `lab` (+2 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 357 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **32 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `on()` connect `htmx.min.js` to ``cmd/meshctl``, `kernel`, `.mesh`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 254 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 254 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `harness` to the rest of the system?**
  _7 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08602025205798791 - nodes in this community are weakly interconnected._
- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `NodeConfig`, `Matcher`, `archtest.go`, `Node`, `Dir`, `ServerConfig`, `Join`, `time.Time`, `agent/agent_test.go`, ``cmd/meshctl``, `.mesh`, `Host`, `serveRuntime`, `Test catalogue`, `meshDNS`, `Sync`, `fetchwintun/main.go`, `Policy`, `nrpt_windows.go`, `Ruleset`, `.routes`, ``internal/proto``, `newRig`, `Engine`, `same`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.059940059940059943 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect `meshHarness` to ``internal/server/api``, `htmx.min.js`, `Run`, `CheckMeshCIDR6`, `Container-mode agent`, `Open`, `.do`, `storeNode`, `Hub`, `MeshTx`, `serveRuntime`, `Test catalogue`, `net/netip.Addr`, `PublishNames`, `toAPIError`, `serve`, `.routes`, `.nodeView`, `TestFailoverMovesRelays`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._