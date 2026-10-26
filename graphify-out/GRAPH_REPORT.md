# Graph Report - vpn-mesh-utility  (2026-10-26)

## Corpus Check
- 332 files · ~405,382 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 4081 nodes · 21877 edges · 123 communities (105 shown, 18 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2535 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `941d3b65`
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
- `internal/proto`
- otherNet
- net/http.ResponseWriter
- Node
- New
- Dir
- Agent (`internal/agent/...`)
- Join
- net/netip.Prefix
- Net
- nodeRow
- Input
- lab
- Snapshot
- flag.FlagSet
- Command reference
- .mesh
- PhoneConfig
- context.Context
- newRig
- Service
- agent/agent_test.go
- Agent (`internal/agent/...`)
- `internal/server/auth`
- testing.T
- log/slog.Logger
- .exec
- openTest
- NodeConfig
- netlinkNet
- Net
- planner_test.go
- Registration
- serveRuntime
- README.md
- Host
- .query
- sync.Mutex
- .queryRow
- LocalStatus
- .Set
- errorf
- Matcher
- Client
- exprs
- Engine
- meshDNS
- AddrEntry
- Sync
- .Apply
- nullInt
- kernel
- .meshctl
- UI
- .add
- Policy
- time.Time
- fetchwintun/main_test.go
- fakeServer
- Tx
- ui
- Of
- rulesFor
- Sync
- .readJSON
- net/netip.Addr
- resolved
- Segmentation
- io.Writer
- io.Reader
- .renderPolicies
- deploy_test.go
- service_windows.go
- File
- JoinRequest
- services
- Ruleset
- Diff
- ValidSecret
- userspace
- TestOneRunningCopy
- loadMigrations
- Version
- run
- Table
- Schedule
- 8. The web UI
- context.Context
- StatusError
- AdminError
- .SetSegmentation
- StepError
- TestMain
- Engine
- .Reach
- ns
- `internal/server` and `internal/server/export`
- stream
- fromFile
- Slot
- AuditEntry
- handler
- Refusal
- entry

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 308 edges
2. `ui` - 104 edges
3. ``internal/server/api`` - 99 edges
4. `ServerConfig()` - 89 edges
5. `Session` - 85 edges
6. `Slot` - 82 edges
7. `newHarness()` - 72 edges
8. `NodeConfig` - 69 edges
9. `env` - 66 edges
10. `api` - 56 edges

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

## Communities (123 total, 18 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.12
Nodes (47): rig, TestProxyLoopAppliesAgain(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), IsStepError(), New(), failed() (+39 more)

### Community 2 - "fail"
Cohesion: 0.06
Nodes (72): Security invariants, Add a validator, Container-mode agent, `internal/validate`, check(), validateMeshNames(), newNode(), validateSet() (+64 more)

### Community 3 - "`internal/server/api`"
Cohesion: 0.05
Nodes (104): countingDB, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket() (+96 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (114): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names, Containers (+106 more)

### Community 5 - "env"
Cohesion: 0.13
Nodes (42): checkMachineIDFile(), joinFlags(), machineIDFlag(), setupAgentInstall(), setupAgentJoin(), setupAgentReset(), setupAgentRun(), setupAgentUninstall() (+34 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.10
Nodes (23): Matrix, MatrixCell, Policy, PolicyRule, Schedule, Service, activeAt(), zoneOf() (+15 more)

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
Cohesion: 0.26
Nodes (10): Listen(), Listen(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe(), waitPipe() (+2 more)

### Community 12 - "commands.go"
Cohesion: 0.16
Nodes (20): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), setupVersion(), shownDefault() (+12 more)

### Community 14 - "`internal/proto`"
Cohesion: 0.04
Nodes (89): brokenStream, Determinism and goldens, Options, ticketServer, `internal/proto`, Join(), New(), newClient() (+81 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.10
Nodes (16): api, apiError, cli(), hostFamilies(), decodeJSON(), errorf(), invalid(), readBody() (+8 more)

### Community 17 - "Node"
Cohesion: 0.20
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 18 - "New"
Cohesion: 0.15
Nodes (10): DB, newRenderer(), staticHandler(), headers(), New(), sameOrigin(), Deps, handler (+2 more)

### Community 19 - "Dir"
Cohesion: 0.06
Nodes (40): `internal/fsutil`, meshKey(), sign(), TestAnswersMustBeSigned(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, Enrollment (+32 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.18
Nodes (31): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+23 more)

### Community 21 - "Join"
Cohesion: 0.08
Nodes (38): hostService(), hostService(), DefaultSocket(), Pipe(), TestDefaultSocket(), TestIsPublic(), TestJoin(), TestJoinValidatesBeforeCallingTheServer() (+30 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.09
Nodes (26): Routes, index(), route(), compare(), luidOf(), onLink(), meshCIDR(), Allocate() (+18 more)

### Community 24 - "nodeRow"
Cohesion: 0.40
Nodes (5): dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 25 - "Input"
Cohesion: 0.13
Nodes (14): PlanError, Record, Check(), checkContainers(), checkContainers6(), checkDNS(), checkNode6(), checkUsable() (+6 more)

### Community 26 - "lab"
Cohesion: 0.23
Nodes (6): Coverage, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "flag.FlagSet"
Cohesion: 0.23
Nodes (28): setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), setupMeshAdd(), setupMeshList(), setupMeshRemove(), setupMeshSet() (+20 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (43): Command reference, meshctl agent install, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns (+35 more)

### Community 30 - ".mesh"
Cohesion: 0.08
Nodes (27): DNSEntry, NameEntry, NameSet, Routes, 10. Backups, 11. When the control plane is down, 12. Troubleshooting, 13. Upgrading (+19 more)

### Community 31 - "PhoneConfig"
Cohesion: 0.21
Nodes (16): Keys and tokens, `internal/server/phone`, Hub(), joinSet(), NewKey(), Render(), aliceKey(), dualPhone() (+8 more)

### Community 32 - "context.Context"
Cohesion: 0.09
Nodes (11): Deps, runner, kick(), oneLine(), Reset(), Run(), TestOneLine(), TestReportKeepsToTheServersLimits() (+3 more)

### Community 33 - "newRig"
Cohesion: 0.12
Nodes (40): Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, TestTemplatesHaveNoScript(), TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes() (+32 more)

### Community 34 - "Service"
Cohesion: 0.15
Nodes (9): dialing, resolver, Service, bindBoth(), bindPair(), exchange(), serve(), ParseAddr() (+1 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (74): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), newFakeRuntime() (+66 more)

### Community 36 - "Agent (`internal/agent/...`)"
Cohesion: 0.04
Nodes (66): Agent (`internal/agent/...`), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx(), TestConfigureHasKernelSemantics(), TestKernelHoldsOneTable(), TestNetHasKernelSemantics(), TestRoutesSysctlAndFirewall() (+58 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.15
Nodes (25): `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide() (+17 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (62): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog() (+54 more)

### Community 39 - "log/slog.Logger"
Cohesion: 0.05
Nodes (37): Change, Deps, Hub, subscription, blockingHost, TestOpen(), Open(), TestOpenRefusesUnknownBackends() (+29 more)

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (6): newUUID(), mustAffect(), boolInt(), classify(), changed(), Session

### Community 41 - "openTest"
Cohesion: 0.09
Nodes (61): rig, `internal/server/store`, TestReconcilerVersionsOnlyRealChanges(), TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart() (+53 more)

### Community 42 - "NodeConfig"
Cohesion: 0.08
Nodes (18): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Sessions (+10 more)

### Community 43 - "netlinkNet"
Cohesion: 0.13
Nodes (12): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), TestKernelShapes(), toRule() (+4 more)

### Community 44 - "Net"
Cohesion: 0.14
Nodes (6): ensureAddr(), Net, Rule, syncDropped(), syncProxies(), syncRules()

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (46): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), Compute(), deliver(), deliverAddr(), encode() (+38 more)

### Community 46 - "Registration"
Cohesion: 0.19
Nodes (8): Domain, fakeRegistrar, Registrar, Registration, New(), NewRegistrar(), NewRegistrar(), NewRegistrar()

### Community 47 - "serveRuntime"
Cohesion: 0.13
Nodes (17): TestOpenRuntime(), clip(), Detect(), meshFilter(), New(), Sockets(), inspectOf(), serveRuntime() (+9 more)

### Community 48 - "README.md"
Cohesion: 0.23
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "Host"
Cohesion: 0.05
Nodes (44): link, Deps, exitHost, Host, Sysctl, peer(), TestLinksAddressesAndFailures(), Sysctls (+36 more)

### Community 50 - ".query"
Cohesion: 0.14
Nodes (5): ContainerName, scanAll(), splitList(), PeerStat, StatusUpdate

### Community 51 - "sync.Mutex"
Cohesion: 0.20
Nodes (12): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+4 more)

### Community 52 - ".queryRow"
Cohesion: 0.10
Nodes (5): AgentSession, HashToken(), NewToken(), Machine, Refusal

### Community 53 - "LocalStatus"
Cohesion: 0.08
Nodes (27): statusDetail, dnsLine(), printExit(), Client, Logs, Handler(), logs(), refusal() (+19 more)

### Community 54 - ".Set"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 55 - "errorf"
Cohesion: 0.06
Nodes (50): CopyInfo, IPv6Request, IPv6View, NodeDetail, NodeList, PhoneProfile, PhoneRequest, addresses() (+42 more)

### Community 56 - "Matcher"
Cohesion: 0.29
Nodes (4): serves(), Service, Matcher, matchRule

### Community 57 - "Client"
Cohesion: 0.13
Nodes (8): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, ValidChallenge(), DroppedName

### Community 58 - "exprs"
Cohesion: 0.07
Nodes (27): Route, route, Add a command, Add an agent API route, Dev script tasks, Developer guide, Extending the code, Goldens (+19 more)

### Community 59 - "Engine"
Cohesion: 0.10
Nodes (8): IPv6Host, Rechecker, Add an apply step, DNS, Engine, Firewall, MeshFirewall, PeerStatus

### Community 60 - "meshDNS"
Cohesion: 0.15
Nodes (24): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+16 more)

### Community 61 - "AddrEntry"
Cohesion: 0.33
Nodes (6): addrs(), AddrEntry(), MergeRanges(), TestAddrEntry(), TestAddrRangePrefix(), TestMergeRanges()

### Community 62 - "Sync"
Cohesion: 0.14
Nodes (34): Sysctls, NewNet(), TestTeardownFakes(), forHost(), TestForHostWithoutIPv6(), Gateway(), Guard(), Owned() (+26 more)

### Community 63 - ".Apply"
Cohesion: 0.14
Nodes (13): applied, Firewall, policy, Firewall, exitPolicy(), normalize(), relayPolicy(), richRule() (+5 more)

### Community 64 - "nullInt"
Cohesion: 0.16
Nodes (4): Add a database migration, DNSRecord, nullInt(), nullStr()

### Community 65 - "kernel"
Cohesion: 0.15
Nodes (5): Status, TestCreateLinkNeedsPrivilege(), lookup(), kernel, links

### Community 66 - ".meshctl"
Cohesion: 0.18
Nodes (11): A node's container on the host network (the NAS), Changes a node asks for, DNS, Interfaces the agent did not create, Security, Several meshes on one host, The host firewall, The local UI (internal/agent/localui) (+3 more)

### Community 67 - "UI"
Cohesion: 0.11
Nodes (15): hostOf(), Listen(), New(), oldest(), TestServeAndListen(), validOrNot(), bytesize(), loadPages() (+7 more)

### Community 69 - "Policy"
Cohesion: 0.15
Nodes (23): NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily(), compile() (+15 more)

### Community 70 - "time.Time"
Cohesion: 0.06
Nodes (23): pendingApply, safeApply, syncBuffer, Failover, hubMove, hubState, replayCache, failures (+15 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.14
Nodes (20): `internal/tools/fetchwintun`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue, NewUserspace(), Ensure(), Pin, install() (+12 more)

### Community 72 - "fakeServer"
Cohesion: 0.28
Nodes (3): fakeServer, signed(), Definitive()

### Community 73 - "Tx"
Cohesion: 0.06
Nodes (34): MeshList, MeshSetRequest, meshSpace, MeshUpdate, NodeAddRequest, NodeSetRequest, SegmentView, SettingsRequest (+26 more)

### Community 74 - "ui"
Cohesion: 0.09
Nodes (20): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), anyText() (+12 more)

### Community 75 - "Of"
Cohesion: 0.13
Nodes (12): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+4 more)

### Community 76 - "rulesFor"
Cohesion: 0.33
Nodes (12): nrptRule, desired(), domainName(), TestZoneFollowsTheConfig(), fromRules(), keyPart(), nrptKey(), ownedKey() (+4 more)

### Community 77 - "Sync"
Cohesion: 0.05
Nodes (28): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+20 more)

### Community 78 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 79 - "net/netip.Addr"
Cohesion: 0.09
Nodes (22): network, hostAddrs, nameTable, zone, Apply (agent), proxied(), gateway(), host() (+14 more)

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "Segmentation"
Cohesion: 0.09
Nodes (26): RecordRequest, segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange (+18 more)

### Community 82 - "io.Writer"
Cohesion: 0.25
Nodes (9): TestBytesize(), TestPrintEnrollmentInTheLocalZone(), printMesh(), printChange(), printEnrollment(), bytesize(), printStatus(), Write() (+1 more)

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - ".renderPolicies"
Cohesion: 0.06
Nodes (29): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), readZoneForm(), ruleRows(), scheduleText(), tabsIn() (+21 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.08
Nodes (41): block, checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth() (+33 more)

### Community 86 - "service_windows.go"
Cohesion: 0.16
Nodes (9): uninstallAgent(), main(), ttyPassword(), commandLine(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows() (+1 more)

### Community 87 - "File"
Cohesion: 0.19
Nodes (8): agentLog(), OpenFile(), TestFileRotates(), lockFile(), unlockFile(), lockFile(), unlockFile(), File

### Community 90 - "JoinRequest"
Cohesion: 0.33
Nodes (6): JoinRequest, pickEndpoint(), pickEndpoint6(), public6(), TestPickEndpoint(), TestPickEndpoint6()

### Community 93 - "services"
Cohesion: 0.32
Nodes (7): addrText(), Rule, merged(), portText(), services(), Range, service

### Community 94 - "Ruleset"
Cohesion: 0.10
Nodes (10): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel (+2 more)

### Community 95 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 96 - "ValidSecret"
Cohesion: 0.40
Nodes (3): HelloRequest, HelloResponse, ValidSecret()

### Community 98 - "userspace"
Cohesion: 0.08
Nodes (7): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace, wgdev

### Community 99 - "TestOneRunningCopy"
Cohesion: 0.33
Nodes (4): tail(), lab, staleTicket(), TestOneRunningCopy()

### Community 100 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 102 - "run"
Cohesion: 0.18
Nodes (19): agentCommands(), commands(), firewallCommands(), healthCommands(), find(), isHelp(), printUsage(), run() (+11 more)

### Community 103 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 104 - "Schedule"
Cohesion: 0.24
Nodes (7): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), weekday()

### Community 105 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

### Community 106 - "context.Context"
Cohesion: 0.06
Nodes (35): fakeService, Access, Config, DB, Params, Pending, Sealer, resetTOTP() (+27 more)

### Community 112 - "TestMain"
Cohesion: 0.05
Nodes (11): Error, NewNet(), NewNet(), NewKernel(), NewKernel(), Refusal, usageError, AliveRequest (+3 more)

### Community 113 - "Engine"
Cohesion: 0.19
Nodes (4): nameServer, DNS, Want, MeshName

### Community 115 - ".Reach"
Cohesion: 0.25
Nodes (7): ReachRequest, flowEnd(), flowEnds(), ReachResult, matrixCell, matrixRow, reachView

### Community 116 - "ns"
Cohesion: 0.11
Nodes (17): End to end -- `test/e2e`, host, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), inNS() (+9 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.07
Nodes (43): `internal/server` and `internal/server/export`, TestRead(), FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport() (+35 more)

### Community 122 - "stream"
Cohesion: 0.33
Nodes (4): dial(), dial(), isPipe(), stream

### Community 123 - "fromFile"
Cohesion: 0.32
Nodes (8): source, clean(), first(), fromFile(), sources(), Read(), TestFirst(), sources()

### Community 130 - "Slot"
Cohesion: 0.11
Nodes (15): TestPolicyHasKernelSemantics(), Priorities(), SlotOf(), TestSlotOf(), NewKernel(), NewKernel(), adapterPart(), adapterSlot() (+7 more)

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 385 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `WGKey`, `Slot`, `archtest.go`, `net.Listener`, ``internal/proto``, `Node`, `Dir`, `Agent (`internal/agent/...`)`, `Join`, `lab`, `.mesh`, `context.Context`, `newRig`, `Service`, `agent/agent_test.go`, `testing.T`, `log/slog.Logger`, `NodeConfig`, `netlinkNet`, `serveRuntime`, `Host`, `sync.Mutex`, `.Set`, `Matcher`, `exprs`, `Engine`, `meshDNS`, `Sync`, `kernel`, `UI`, `Policy`, `time.Time`, `fetchwintun/main_test.go`, `Of`, `rulesFor`, `Sync`, `resolved`, `deploy_test.go`, `service_windows.go`, `File`, `Engine`, ``internal/server` and `internal/server/export``, `fromFile`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Are the 307 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 307 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08949997769748874 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `kernel`, `.mesh`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.11918367346938775 - nodes in this community are weakly interconnected._
- **Why does `otherPolicy` connect `Sync` to `ui/ui_test.go`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._