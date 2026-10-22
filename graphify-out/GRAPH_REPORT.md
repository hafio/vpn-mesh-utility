# Graph Report - vpn-mesh-utility  (2026-10-22)

## Corpus Check
- 285 files · ~318,591 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3464 nodes · 16866 edges · 135 communities (107 shown, 28 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 1986 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6741fef5`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- Check
- Node
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- Ruleset
- meshHarness
- github.com/hafio/vpn-mesh-utility
- api/api_test.go
- Run
- net/http.ResponseWriter
- .do
- New
- Dir
- ServerConfig
- Agent (`internal/agent/...`)
- net/netip.Prefix
- time.Time
- ui
- AdminError
- lab
- Snapshot
- .nodeView
- Command reference
- time.Time
- commands.go
- runner
- newRig
- stream
- agent/agent_test.go
- Input
- `internal/server/auth`
- testing.T
- Policy
- .exec
- openTest
- .mesh
- netlinkNet
- wgdev
- planner_test.go
- meshDNS
- net/http.ResponseWriter
- README.md
- Host
- .renderVersion
- otherNet
- .Tx
- context.Context
- TestSafeApplyStateMachine
- .BeginEnrollment
- Extending the code
- Client
- commands.go
- DNS
- Service
- Tx
- Net
- service_test.go
- Container
- Container
- adminFlags
- UI
- Registration
- Hub
- 7. Agent
- .Apply
- Name
- StepError
- Deps
- io.Writer
- net/netip.AddrPort
- MeshNetworkSetup
- newRig
- Policy
- sync.Mutex
- errorf
- resolved
- Agent protocol
- .fail
- deploy_test.go
- Join
- fetchwintun/main_test.go
- Security
- AddrEntry
- Kernel
- stop
- Sync
- validate.go
- WriteEvent
- .queryRow
- .Sync
- userspace
- CLAUDE.md
- IPNet
- serveRuntime
- Add a web UI page
- .query
- Schedule
- build
- TestContainerNetworks
- exprs
- noKernel
- handler
- net/netip.Addr
- PhoneConfig
- .Reach
- Ring
- `internal/validate`
- luidOf
- conn
- Open
- TestMeshFirewall
- loadMigrations
- .relayFor
- `internal/server` and `internal/server/export`
- io.Reader
- TestMain
- entry
- Error
- NewNet
- NewNet
- NewKernel
- NewKernel
- NewKernel
- `internal/proto`
- NewKernel
- Refusal
- StatusError

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 235 edges
2. `ui` - 90 edges
3. ``internal/server/api`` - 76 edges
4. `Session` - 74 edges
5. `ServerConfig()` - 67 edges
6. `NodeConfig` - 63 edges
7. `newHarness()` - 55 edges
8. `env` - 48 edges
9. `runner` - 46 edges
10. `api` - 42 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the server (Dokploy)` --references--> `v()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js
- `The mesh firewall` --references--> `Schedule()`  [INFERRED]
  docs/security.md → internal/validate/schedule.go

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

## Communities (135 total, 28 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (23): composeNetwork, containerFile, containerService, nasFile, nasHealth, nasService, linkDNS, DNSRecord (+15 more)

### Community 1 - "NodeConfig"
Cohesion: 0.11
Nodes (40): rig, Determinism and goldens, NewRoutes(), TestRoutesSysctlAndFirewall(), IsStepError(), New(), isStep(), meshDNS() (+32 more)

### Community 2 - "Check"
Cohesion: 0.14
Nodes (13): SystemdUnit(), allowed(), Check(), Runner, TestCheckAllowsFirewallArgv(), TestCheckRefuses(), exists(), host() (+5 more)

### Community 3 - "Node"
Cohesion: 0.20
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.21
Nodes (27): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+19 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (28): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments (+20 more)

### Community 7 - "AddrEntry"
Cohesion: 0.14
Nodes (21): Service, inRanges(), parseAddrs(), inRange(), checkSegmentation(), host(), mergeServices(), newSegResolver() (+13 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.16
Nodes (15): Rule, Violation, Packages, and which way they may import, Input and commands, Coverage, `internal/archtest`, `internal/tools/vulnjudge`, Test catalogue (+7 more)

### Community 11 - "Ruleset"
Cohesion: 0.23
Nodes (6): MeshFirewall, Ruleset, admit(), walk(), Chain, packet

### Community 12 - "meshHarness"
Cohesion: 0.19
Nodes (16): TestIPv6DNSRecords(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness, officeFirewall(), TestEditSegment() (+8 more)

### Community 14 - "api/api_test.go"
Cohesion: 0.07
Nodes (77): countingDB, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket() (+69 more)

### Community 15 - "Run"
Cohesion: 0.09
Nodes (27): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+19 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.18
Nodes (5): api, decode(), reply(), cli(), readSums()

### Community 17 - ".do"
Cohesion: 0.08
Nodes (26): IPv6Request, IPv6View, NetworkInfo, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile (+18 more)

### Community 18 - "New"
Cohesion: 0.42
Nodes (9): Config, Params, New(), DefaultParams(), HashPassword(), TestHashUsesTheSpecParameters(), TestVerifyPassword(), TestVerifyRefusesMalformedHashes() (+1 more)

### Community 19 - "Dir"
Cohesion: 0.07
Nodes (28): `internal/fsutil`, TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious() (+20 more)

### Community 20 - "ServerConfig"
Cohesion: 0.21
Nodes (26): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+18 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.05
Nodes (58): Agent (`internal/agent/...`), TestCleanHostname(), TestDefaultSocket(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig() (+50 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.11
Nodes (23): Routes, NewTable(), compare(), Table, Sync(), NewTable(), NewTable(), ULAPrefix() (+15 more)

### Community 23 - "time.Time"
Cohesion: 0.14
Nodes (10): replayCache, failures, Lockout, LockoutError, serves(), record(), Service, Matcher (+2 more)

### Community 24 - "ui"
Cohesion: 0.12
Nodes (11): Session, listOf(), meshChanged(), atoi(), changed(), message(), actor(), dnsData (+3 more)

### Community 26 - "lab"
Cohesion: 0.17
Nodes (13): End to end -- `test/e2e`, host, deleteNamespaces(), eventually(), lab, inNS(), newLab(), ns() (+5 more)

### Community 27 - "Snapshot"
Cohesion: 0.10
Nodes (24): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+16 more)

### Community 28 - ".nodeView"
Cohesion: 0.13
Nodes (11): ago(), bytesize(), writeHTML(), TestHelpers(), dashboardData, networkBlock, nodeData, nodeForm (+3 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.12
Nodes (11): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, RateLimiter, NewRateLimiter() (+3 more)

### Community 31 - "commands.go"
Cohesion: 0.21
Nodes (12): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), setupVersion(), shownDefault() (+4 more)

### Community 32 - "runner"
Cohesion: 0.06
Nodes (23): Deps, fakeService, runner, Rechecker, resetTOTP(), Add an apply step, kick(), oneLine() (+15 more)

### Community 33 - "newRig"
Cohesion: 0.13
Nodes (36): brokenStream, The web UI's pages (internal/server/ui), `internal/server/ui`, TestDashboardHealth(), TestFirewallPages(), TestNodeGroupsOnPages(), newQR(), TestPhonePage() (+28 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (58): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+50 more)

### Community 36 - "Input"
Cohesion: 0.16
Nodes (12): Record, containers6(), Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable(), sortedByID() (+4 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.14
Nodes (25): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures() (+17 more)

### Community 38 - "testing.T"
Cohesion: 0.10
Nodes (51): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+43 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (25): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+17 more)

### Community 40 - ".exec"
Cohesion: 0.09
Nodes (7): mustAffect(), DNSRecord, joinList(), boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.12
Nodes (51): `internal/server/store`, TestExportOnceSkipsUnchangedAndKeepsFourteen(), DefaultPolicy(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules() (+43 more)

### Community 42 - ".mesh"
Cohesion: 0.12
Nodes (18): DNSEntry, NameEntry, NameSet, 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy) (+10 more)

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (14): netlinkNet, Owned(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule(), IPNet() (+6 more)

### Community 44 - "wgdev"
Cohesion: 0.24
Nodes (3): TestUAPISetGolden(), UAPISet(), wgdev

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (43): PlanError, Add a topology rule, `internal/server/planner`, Problem, Compute(), deliver(), deliverAddr(), encode() (+35 more)

### Community 46 - "meshDNS"
Cohesion: 0.14
Nodes (21): desired(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters() (+13 more)

### Community 47 - "net/http.ResponseWriter"
Cohesion: 0.12
Nodes (8): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), group(), sameOrigin(), handler, page

### Community 48 - "README.md"
Cohesion: 0.10
Nodes (15): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+7 more)

### Community 49 - "Host"
Cohesion: 0.05
Nodes (43): link, exitHost, IPv6Host, Agent service (systemd), TestRunNeedsAnEnrollment(), Host, NewHost(), peer() (+35 more)

### Community 52 - ".Tx"
Cohesion: 0.14
Nodes (15): DB, Pending, addUser(), newPassword(), passwdUser(), fullExpiry(), Service, load() (+7 more)

### Community 53 - "context.Context"
Cohesion: 0.08
Nodes (25): Client, Logs, Handler(), logs(), refusal(), Listen(), New(), laptop() (+17 more)

### Community 54 - "TestSafeApplyStateMachine"
Cohesion: 0.46
Nodes (3): pendingApply, safeApply, TestSafeApplyStateMachine()

### Community 55 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 56 - "Extending the code"
Cohesion: 0.25
Nodes (8): Add a command, Add a database migration, Add a validator, Developer guide, Extending the code, Goldens, Releases, Toolchain

### Community 57 - "Client"
Cohesion: 0.18
Nodes (11): Options, apiError(), clip(), Client, Join(), New(), TestJoinIsUnsigned(), TestUnreachableAndMalformedAnswers() (+3 more)

### Community 58 - "commands.go"
Cohesion: 0.20
Nodes (15): agentCommands(), commands(), runWith(), firewallCommands(), healthCommands(), find(), isHelp(), main() (+7 more)

### Community 59 - "DNS"
Cohesion: 0.18
Nodes (5): nameServer, DNS, Want, DNS, MeshName

### Community 60 - "Service"
Cohesion: 0.14
Nodes (10): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), serve() (+2 more)

### Community 61 - "Tx"
Cohesion: 0.08
Nodes (48): apiError, Change, ComposeFile, ComposeOptions, PhoneRequest, Role, addIPv6(), blockOf() (+40 more)

### Community 62 - "Net"
Cohesion: 0.13
Nodes (5): Net, ensureAddr(), Net, Rule, syncRules()

### Community 63 - "service_test.go"
Cohesion: 0.14
Nodes (11): applied, Firewall, policy, Firewall, normalize(), richRule(), RuleArgv(), wantedPolicies() (+3 more)

### Community 64 - "Container"
Cohesion: 0.20
Nodes (9): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+1 more)

### Community 65 - "Container"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 66 - "adminFlags"
Cohesion: 0.20
Nodes (25): dnsLine(), setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), adminFlags(), cleared(), joinNonEmpty() (+17 more)

### Community 67 - "UI"
Cohesion: 0.15
Nodes (10): hostOf(), oldest(), validOrNot(), ago(), bytesize(), prefsFlash(), handler, page (+2 more)

### Community 68 - "Registration"
Cohesion: 0.16
Nodes (16): Domain, fakeRegistrar, nrpt, nrptRule, Registration, domainName(), same(), TestSameRegistration() (+8 more)

### Community 69 - "Hub"
Cohesion: 0.24
Nodes (3): Hub, subscription, ConfigEvent

### Community 70 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 71 - ".Apply"
Cohesion: 0.16
Nodes (13): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+5 more)

### Community 72 - "Name"
Cohesion: 0.12
Nodes (29): RecordRequest, check(), ParseAddr(), TestParseAddr(), validateMeshNames(), newNode(), validateSet(), ToSnapshot() (+21 more)

### Community 74 - "Deps"
Cohesion: 0.33
Nodes (4): Deps, Firewall, Sysctls, Sysctl

### Community 75 - "io.Writer"
Cohesion: 0.16
Nodes (11): agentLog(), printExit(), printChange(), printEnrollment(), ttyPassword(), OpenFile(), TestFileRotates(), Write() (+3 more)

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.48
Nodes (5): bindBoth(), bindPair(), Listen(), Listen(), listenSocket()

### Community 77 - "MeshNetworkSetup"
Cohesion: 0.36
Nodes (8): containerHeader(), MeshNetworkSetup(), networkCommands(), resolvers(), serviceExample(), TestHostModeNetworkSetup(), NetworkParams, NetworkSetup

### Community 78 - "newRig"
Cohesion: 0.17
Nodes (13): capped, TestCappedKeepsTheHead(), golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage() (+5 more)

### Community 79 - "Policy"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, Policy, PolicyRule, Schedule, activeAt(), zoneOf(), Overlaps() (+4 more)

### Community 81 - "errorf"
Cohesion: 0.10
Nodes (21): segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange, meshChange() (+13 more)

### Community 82 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 83 - "Agent protocol"
Cohesion: 0.18
Nodes (11): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+3 more)

### Community 84 - ".fail"
Cohesion: 0.09
Nodes (23): SegPolicy, SegSchedule, anyText(), listText(), objectRows(), readPolicyForm(), readZoneForm(), ruleRows() (+15 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.09
Nodes (49): hostSocket(), renderContainer(), RenderContainerCompose(), containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposeHeader(), TestContainerComposePosture() (+41 more)

### Community 86 - "Join"
Cohesion: 0.16
Nodes (14): hostService(), installAgent(), hostService(), installAgent(), uninstallAgent(), cleanHostname(), JoinOptions, Service (+6 more)

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.14
Nodes (20): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+12 more)

### Community 88 - "Security"
Cohesion: 0.17
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens, Security (+6 more)

### Community 89 - "AddrEntry"
Cohesion: 0.27
Nodes (12): addrs(), addrText(), Compile(), contains(), Rule, merged(), portText(), prefixRange() (+4 more)

### Community 92 - "stop"
Cohesion: 0.27
Nodes (4): commandLine(), stop(), TestCommandLineQuotesThePaths(), SCM

### Community 93 - "Sync"
Cohesion: 0.21
Nodes (19): Sysctls, NewNet(), pfx(), TestNetHasKernelSemantics(), Gateway(), Guard(), parse(), Sync() (+11 more)

### Community 94 - "validate.go"
Cohesion: 0.19
Nodes (16): checkDNS(), Arg(), DNSName(), Domains(), fail(), Hostname(), isLoopbackHost(), Key32() (+8 more)

### Community 95 - "WriteEvent"
Cohesion: 0.27
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 96 - ".queryRow"
Cohesion: 0.15
Nodes (4): scanNode(), User, scanUser(), scanner

### Community 98 - "userspace"
Cohesion: 0.12
Nodes (5): NewUserspace(), defaultInterface(), exitRoutes(), Link, userspace

### Community 99 - "CLAUDE.md"
Cohesion: 0.33
Nodes (5): graphify, Security invariants, Style, What this repo is, You do not run the gates

### Community 100 - "IPNet"
Cohesion: 0.14
Nodes (7): Status, 4. Spokes and exits, Container mode, TestCreateLinkNeedsPrivilege(), lookup(), kernel, links

### Community 101 - "serveRuntime"
Cohesion: 0.23
Nodes (11): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+3 more)

### Community 102 - "Add a web UI page"
Cohesion: 0.17
Nodes (11): Add a web UI page, TestTemplatesHaveNoScript(), loadPages(), newPolicyForm(), dnsState(), newRenderer(), golden(), TestEveryPageRenders() (+3 more)

### Community 103 - ".query"
Cohesion: 0.10
Nodes (10): compareAddr(), containerAddr(), PublishNames(), ContainerName, scanAll(), splitList(), AuditEntry, PeerStat (+2 more)

### Community 104 - "Schedule"
Cohesion: 0.26
Nodes (6): Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), weekday()

### Community 105 - "build"
Cohesion: 0.33
Nodes (4): Dev script tasks, build(), TestBuildBatch(), nftBatch

### Community 107 - "exprs"
Cohesion: 0.19
Nodes (6): exprs(), ifname(), intervalElems(), TestExprs(), TestIntervalElems(), nftOp

### Community 110 - "net/netip.Addr"
Cohesion: 0.11
Nodes (22): network, nameTable, zone, TestIsPublic(), host(), parseServers(), serverList(), TestNRPTServers() (+14 more)

### Community 111 - "PhoneConfig"
Cohesion: 0.19
Nodes (12): Shared fixtures -- `internal/testfix`, Hub(), NewKey(), Render(), aliceKey(), golden(), TestHub(), TestNewKey() (+4 more)

### Community 112 - ".Reach"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 114 - "`internal/validate`"
Cohesion: 0.13
Nodes (26): `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), FuzzArg(), FuzzName(), TestArg(), TestDNSName() (+18 more)

### Community 115 - "luidOf"
Cohesion: 0.53
Nodes (3): luidOf(), onLink(), winTable

### Community 117 - "Open"
Cohesion: 0.50
Nodes (3): TestOpen(), Open(), TestOpenRefusesUnknownBackends()

### Community 118 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.08
Nodes (37): `internal/server` and `internal/server/export`, NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses(), FromSnapshot(), Document (+29 more)

### Community 122 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 134 - "`internal/proto`"
Cohesion: 0.06
Nodes (51): `internal/proto`, NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Local, compile(), officeHours() (+43 more)

## Knowledge Gaps
- **5 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `stream`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 332 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **28 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does ``internal/server/api`` connect `api/api_test.go` to `newRig`, `htmx.min.js`, `openTest`, `archtest.go`, `meshHarness`, `Run`, `.do`, ``internal/server` and `internal/server/export``?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Are the 234 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 234 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _5 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07523906467290792 - nodes in this community are weakly interconnected._
- **Why does `at()` connect `htmx.min.js` to `api/api_test.go`?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.11265969802555169 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `.mesh`, `IPNet`, `testing.T`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._