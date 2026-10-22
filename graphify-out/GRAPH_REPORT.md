# Graph Report - vpn-mesh-utility  (2026-10-22)

## Corpus Check
- 272 files · ~286,350 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3298 nodes · 15688 edges · 129 communities (111 shown, 18 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 1817 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `4d908cdb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- Run
- time.Time
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- Ruleset
- execrun_test.go
- github.com/hafio/vpn-mesh-utility
- `internal/server/api`
- Agent (`internal/agent/...`)
- net/http.ResponseWriter
- context.Context
- action
- Dir
- ServerConfig
- WGKey
- net/netip.Prefix
- meshctl/main.go
- .nodeView
- LocalStatus
- lab
- Snapshot
- Session
- Command reference
- time.Duration
- Host
- runner
- newRig
- ns
- agent/agent_test.go
- Net
- `internal/server/auth`
- runArgs
- Policy
- .exec
- openTest
- .mesh
- netlinkNet
- NodeConfig
- planner_test.go
- meshDNS
- .show
- README.md
- CLAUDE.md
- Security
- otherNet
- Service
- Client
- TestNRPTRules
- Role
- Input
- .Segment
- Runtime
- DNS
- Service
- Tx
- Sync
- service_test.go
- Policy
- Container
- Compose and stack files
- UI
- Registration
- Hub
- meshHarness
- otherPolicy
- Name
- StepError
- Clock
- Installer
- net/netip.AddrPort
- container.go
- ReachResult
- NodeInfo
- .query
- errorf
- resolved
- links
- .fail
- deploy_test.go
- Host
- fetchwintun/main_test.go
- exprs
- Range
- NodeConfig
- Join
- PublishNames
- log/slog.Logger
- .BeginEnrollment
- .queryRow
- ServerURL
- userspace
- File
- kernel
- serveRuntime
- build
- WriteEvent
- New
- Ring
- Run
- conn
- io.Reader
- TestCompileAgreesWithDecide
- net/netip.Addr
- noKernel
- 8. The web UI
- commands.go
- validate_test.go
- stream
- Exit
- NewNet
- NewHub
- NewKernel
- NewKernel
- `internal/server` and `internal/server/export`
- NewKernel
- Error
- AdminError
- service_windows.go
- `internal/proto`
- StatusError
- Zone

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 223 edges
2. `ui` - 87 edges
3. `Session` - 72 edges
4. ``internal/server/api`` - 68 edges
5. `ServerConfig()` - 62 edges
6. `NodeConfig` - 59 edges
7. `newHarness()` - 52 edges
8. `env` - 47 edges
9. `runner` - 44 edges
10. `api` - 40 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Packages, and which way they may import` --references--> `TestModuleBoundaries()`  [INFERRED]
  CLAUDE.md → internal/archtest/archtest_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the server (Dokploy)` --references--> `v()`  [INFERRED]
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

## Communities (129 total, 18 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.07
Nodes (29): Error, composeNetwork, containerFile, containerService, nasFile, nasHealth, nasService, linkDNS (+21 more)

### Community 1 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 2 - "Run"
Cohesion: 0.10
Nodes (15): Deps, Deps, Rechecker, Add a command, Add a database migration, Add a web UI page, Add an apply step, Extending the code (+7 more)

### Community 3 - "time.Time"
Cohesion: 0.14
Nodes (10): replayCache, failures, Lockout, LockoutError, CanonicalRequest(), Health(), phoneHealth(), record() (+2 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.23
Nodes (30): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+22 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.08
Nodes (26): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+18 more)

### Community 7 - "AddrEntry"
Cohesion: 0.16
Nodes (18): Service, checkSegmentation(), host(), mergeServices(), Overlaps(), parseSpans(), spanOf(), Summarize() (+10 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.22
Nodes (11): Rule, Violation, Input and commands, `internal/archtest`, Check(), importsOf(), ModuleRoot(), TestCheckerCatchesViolations() (+3 more)

### Community 11 - "Ruleset"
Cohesion: 0.15
Nodes (6): watchingFirewall, Kernel, Ruleset, Chain, Enforcer, Kernel

### Community 12 - "execrun_test.go"
Cohesion: 0.12
Nodes (13): capped, allowed(), Check(), fakeProgram(), TestCappedKeepsTheHead(), TestCheckAllowsFirewallArgv(), TestCheckRefuses(), TestRunCapturesOutputWithAMinimalEnvironment() (+5 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.09
Nodes (61): countingDB, `internal/server/api`, ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose() (+53 more)

### Community 15 - "Agent (`internal/agent/...`)"
Cohesion: 0.05
Nodes (53): Agent (`internal/agent/...`), DefaultSocket(), oneLine(), TestCleanHostname(), TestDefaultSocket(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig() (+45 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.14
Nodes (14): api, apiError, decode(), reply(), cli(), decodeJSON(), errorf(), readBody() (+6 more)

### Community 17 - "context.Context"
Cohesion: 0.09
Nodes (16): fakeService, NetworkInfo, NodeAddRequest, NodeSetRequest, ReachRequest, dial(), dial(), isPipe() (+8 more)

### Community 18 - "action"
Cohesion: 0.12
Nodes (30): TestPrintEnrollmentInTheLocalZone(), dnsLine(), printExit(), setupFirewallShow(), adminFlags(), cleared(), heard(), joinNonEmpty() (+22 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (25): `internal/fsutil`, TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious() (+17 more)

### Community 20 - "ServerConfig"
Cohesion: 0.21
Nodes (26): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+18 more)

### Community 21 - "WGKey"
Cohesion: 0.10
Nodes (40): rig, Keys and tokens, TestRunNeedsAnEnrollment(), NewHost(), NewRoutes(), TestLinksAddressesAndFailures(), IsStepError(), New() (+32 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.10
Nodes (15): network, Routes, host(), NewTable(), compare(), Desired(), Table, Sync() (+7 more)

### Community 23 - "meshctl/main.go"
Cohesion: 0.20
Nodes (15): agentCommands(), commands(), runWith(), firewallCommands(), healthCommands(), find(), isHelp(), main() (+7 more)

### Community 24 - ".nodeView"
Cohesion: 0.24
Nodes (8): ago(), bytesize(), TestHelpers(), dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 25 - "LocalStatus"
Cohesion: 0.09
Nodes (21): Client, Logs, Handler(), logs(), refusal(), Listen(), laptop(), PrefsRequest (+13 more)

### Community 26 - "lab"
Cohesion: 0.21
Nodes (8): End to end -- `test/e2e`, host, eventually(), lab, inNS(), TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.09
Nodes (24): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+16 more)

### Community 28 - "Session"
Cohesion: 0.11
Nodes (11): Session, group(), atoi(), before(), changed(), versionNumber(), message(), writeHTML() (+3 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (33): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+25 more)

### Community 30 - "time.Duration"
Cohesion: 0.14
Nodes (10): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, RateLimiter, NewRateLimiter() (+2 more)

### Community 31 - "Host"
Cohesion: 0.06
Nodes (37): link, exitHost, Agent service (systemd), Host, peer(), AddrPort(), TestAddrPort(), hexKey() (+29 more)

### Community 32 - "runner"
Cohesion: 0.12
Nodes (8): pendingApply, runner, safeApply, 12. Troubleshooting, kick(), Result, drain(), TestSafeApplyStateMachine()

### Community 33 - "newRig"
Cohesion: 0.13
Nodes (35): brokenStream, The web UI's pages (internal/server/ui), `internal/server/ui`, TestDashboardHealth(), TestFirewallPages(), TestNodeGroupsOnPages(), newQR(), TestPhonePage() (+27 more)

### Community 34 - "ns"
Cohesion: 0.15
Nodes (11): lab, TestContainerNetworks(), deleteNamespaces(), newLab(), ns(), tail(), lab, TestRelayFailover() (+3 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.12
Nodes (46): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+38 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.14
Nodes (25): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures() (+17 more)

### Community 38 - "runArgs"
Cohesion: 0.09
Nodes (56): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+48 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (22): Family, netlinkPolicy, Policy, mutate(), Desired(), Policy, Route, Rule (+14 more)

### Community 40 - ".exec"
Cohesion: 0.07
Nodes (10): mustAffect(), DNSRecord, joinList(), boolInt(), classify(), nullInt(), nullStr(), Session (+2 more)

### Community 41 - "openTest"
Cohesion: 0.14
Nodes (42): `internal/server/store`, DefaultPolicy(), LegacySegmentation(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation(), TestSegmentationRoundTrip() (+34 more)

### Community 42 - ".mesh"
Cohesion: 0.07
Nodes (32): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+24 more)

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (13): netlinkNet, blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule(), IPNet(), Prefix() (+5 more)

### Community 44 - "NodeConfig"
Cohesion: 0.25
Nodes (9): Config and events, 4. Roles and topology, Peer, Record, ContainerNetwork(), Node, sortPeers(), sortPrefixes() (+1 more)

### Community 45 - "planner_test.go"
Cohesion: 0.16
Nodes (35): Add a topology rule, `internal/server/planner`, TestCheckRequested(), deliver(), deliverAddr(), encode(), firstDiff(), hasProblem() (+27 more)

### Community 46 - "meshDNS"
Cohesion: 0.25
Nodes (19): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceAnswersTheContainers() (+11 more)

### Community 47 - ".show"
Cohesion: 0.21
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), sameOrigin(), handler

### Community 48 - "README.md"
Cohesion: 0.17
Nodes (8): Coverage, `internal/tools/vulnjudge`, Test catalogue, Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "CLAUDE.md"
Cohesion: 0.25
Nodes (7): Determinism and goldens, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, TestHashIsPinned()

### Community 50 - "Security"
Cohesion: 0.18
Nodes (10): Changes a node asks for, DNS, Interfaces the agent did not create, Security, The host firewall, The local UI (internal/agent/localui), The mesh firewall, Web UI sign-in (internal/server/auth) (+2 more)

### Community 52 - "Service"
Cohesion: 0.14
Nodes (15): DB, Pending, addUser(), newPassword(), passwdUser(), resetTOTP(), fullExpiry(), Service (+7 more)

### Community 53 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 54 - "TestNRPTRules"
Cohesion: 0.17
Nodes (16): nrpt, nrptRule, domainName(), same(), TestSameRegistration(), fromRules(), nrptKey(), ownedKey() (+8 more)

### Community 55 - "Role"
Cohesion: 0.27
Nodes (11): Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateContainers(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan() (+3 more)

### Community 56 - "Input"
Cohesion: 0.19
Nodes (14): PlanError, Check(), checkContainers(), Problem, sortedByID(), Compute(), Input, newPlan() (+6 more)

### Community 57 - ".Segment"
Cohesion: 0.25
Nodes (5): Matrix, MatrixCell, SegmentView, ZoneView, PolicyRule

### Community 58 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 59 - "DNS"
Cohesion: 0.15
Nodes (8): nameServer, DNS, desired(), Want, newNameTable(), TestMeshViewValidate(), MeshName, MeshNames()

### Community 61 - "Tx"
Cohesion: 0.11
Nodes (28): Change, ComposeFile, ComposeOptions, PhoneRequest, checkGroups(), checkName(), containerSubnet(), defaultMTU() (+20 more)

### Community 62 - "Sync"
Cohesion: 0.13
Nodes (21): Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard(), Owned() (+13 more)

### Community 63 - "service_test.go"
Cohesion: 0.15
Nodes (11): applied, Firewall, policy, Firewall, normalize(), richRule(), RuleArgv(), wantedPolicies() (+3 more)

### Community 64 - "Policy"
Cohesion: 0.18
Nodes (16): NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile(), Local, compile(), TestCompileGolden() (+8 more)

### Community 65 - "Container"
Cohesion: 0.20
Nodes (9): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+1 more)

### Community 66 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 67 - "UI"
Cohesion: 0.12
Nodes (16): hostOf(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize(), loadPages() (+8 more)

### Community 68 - "Registration"
Cohesion: 0.22
Nodes (7): Domain, fakeRegistrar, Registrar, Registration, NewRegistrar(), NewRegistrar(), NewRegistrar()

### Community 69 - "Hub"
Cohesion: 0.21
Nodes (5): Hub, subscription, ConfigEvent, Reconciler, NewReconciler()

### Community 70 - "meshHarness"
Cohesion: 0.18
Nodes (15): ErrorMessage(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness, officeFirewall(), TestEditSegment() (+7 more)

### Community 72 - "Name"
Cohesion: 0.13
Nodes (29): check(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), Ref(), ServiceEntry(), TestRef() (+21 more)

### Community 73 - "StepError"
Cohesion: 0.14
Nodes (4): entry, StepError, NewNet(), NewKernel()

### Community 74 - "Clock"
Cohesion: 0.25
Nodes (11): Shared fixtures -- `internal/testfix`, pendingExists(), safeRig(), TestSafeApplyConfirmsWhenTheServerAnswers(), TestSafeApplyDisarmsWhenTheServerGoesBack(), TestSafeApplyRecoversAPendingConfigAtBoot(), TestSafeApplyReportsAFailedRollback(), TestSafeApplyRollsBackWithoutAnAnswer() (+3 more)

### Community 75 - "Installer"
Cohesion: 0.19
Nodes (10): TestUnitPosture(), SystemdUnit(), Runner, exists(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere(), TestInstallWritesTheUnitAndStartsIt() (+2 more)

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.24
Nodes (9): resolver, bindBoth(), bindPair(), serve(), Listen(), Listen(), listenSocket(), ParseAddr() (+1 more)

### Community 77 - "container.go"
Cohesion: 0.15
Nodes (19): containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers(), serviceExample() (+11 more)

### Community 78 - "ReachResult"
Cohesion: 0.50
Nodes (4): ReachResult, matrixCell, matrixRow, reachView

### Community 79 - "NodeInfo"
Cohesion: 0.15
Nodes (11): NodeDetail, NodeList, PhoneProfile, statusDetail, PeerStatus, Rollback, NodeInfo, sameRollback() (+3 more)

### Community 80 - ".query"
Cohesion: 0.18
Nodes (5): ContainerName, scanAll(), splitList(), AuditEntry, PeerStat

### Community 81 - "errorf"
Cohesion: 0.10
Nodes (24): RecordRequest, segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), MeshChange, meshChange() (+16 more)

### Community 82 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 83 - "links"
Cohesion: 0.24
Nodes (3): lookup(), Link, links

### Community 84 - ".fail"
Cohesion: 0.07
Nodes (29): anyText(), listOf(), listText(), meshChanged(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+21 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.14
Nodes (30): containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposePosture(), TestContainerComposeRefusesHostileValues(), TestContainerRendererHoldsItsStructure(), checkPosture(), keys() (+22 more)

### Community 86 - "Host"
Cohesion: 0.36
Nodes (3): Host, SyncAddr(), SyncLink()

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.15
Nodes (20): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+12 more)

### Community 88 - "exprs"
Cohesion: 0.19
Nodes (6): exprs(), ifname(), intervalElems(), TestExprs(), TestIntervalElems(), nftOp

### Community 89 - "Range"
Cohesion: 0.24
Nodes (11): addrs(), addrText(), contains(), Rule, merged(), portText(), prefixRange(), services() (+3 more)

### Community 90 - "NodeConfig"
Cohesion: 0.18
Nodes (6): NodeConfig, Role, Render(), Profile, MeshNode, Self

### Community 92 - "Join"
Cohesion: 0.27
Nodes (8): hostService(), installAgent(), installAgent(), TestJoinValidatesBeforeCallingTheServer(), JoinOptions, Service, Join(), Systemd()

### Community 93 - "PublishNames"
Cohesion: 0.18
Nodes (12): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), PublishNames(), DNSView (+4 more)

### Community 94 - "log/slog.Logger"
Cohesion: 0.18
Nodes (7): New(), TestOpen(), Open(), TestOpenRefusesUnknownBackends(), NewUserspace(), NewUserspace(), Options

### Community 95 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 96 - ".queryRow"
Cohesion: 0.13
Nodes (6): scanNode(), HashToken(), NewToken(), User, scanUser(), scanner

### Community 97 - "ServerURL"
Cohesion: 0.29
Nodes (10): TestInstallChangesOnlyTheMarkers(), TestInstallGolden(), TestInstallParses(), TestInstallRefusesHostileValues(), testParams(), RenderInstall(), InstallParams, isLoopbackHost() (+2 more)

### Community 98 - "userspace"
Cohesion: 0.16
Nodes (4): defaultInterface(), exitRoutes(), unspecified(), userspace

### Community 99 - "File"
Cohesion: 0.27
Nodes (5): agentLog(), ttyPassword(), OpenFile(), TestFileRotates(), File

### Community 100 - "kernel"
Cohesion: 0.20
Nodes (5): Status, 4. Spokes and exits, Container mode, TestCreateLinkNeedsPrivilege(), kernel

### Community 101 - "serveRuntime"
Cohesion: 0.12
Nodes (14): syncBuffer, TestOpenRuntime(), Sysctl, Detect(), New(), Sockets(), inspectOf(), serveRuntime() (+6 more)

### Community 102 - "build"
Cohesion: 0.20
Nodes (8): Dev script tasks, Developer guide, Goldens, Releases, Toolchain, build(), TestBuildBatch(), nftBatch

### Community 103 - "WriteEvent"
Cohesion: 0.27
Nodes (8): TestEventsDeliversConfigEvents(), TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 104 - "New"
Cohesion: 0.42
Nodes (9): Config, Params, New(), DefaultParams(), HashPassword(), TestHashUsesTheSpecParameters(), TestVerifyPassword(), TestVerifyRefusesMalformedHashes() (+1 more)

### Community 106 - "Run"
Cohesion: 0.11
Nodes (24): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+16 more)

### Community 108 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 109 - "TestCompileAgreesWithDecide"
Cohesion: 0.60
Nodes (4): admit(), TestCompileAgreesWithDecide(), walk(), packet

### Community 110 - "net/netip.Addr"
Cohesion: 0.12
Nodes (14): nameTable, zone, TestIsPublic(), isPublic(), PublicIPv4(), inRanges(), parseAddrs(), serves() (+6 more)

### Community 112 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

### Community 113 - "commands.go"
Cohesion: 0.23
Nodes (11): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), shownDefault(), cell() (+3 more)

### Community 114 - "validate_test.go"
Cohesion: 0.11
Nodes (19): Security invariants, Add a validator, `internal/validate`, Arg(), mask(), PipePath(), FuzzArg(), FuzzName() (+11 more)

### Community 118 - "NewHub"
Cohesion: 0.50
Nodes (5): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.10
Nodes (30): `internal/server` and `internal/server/export`, DB, FromSnapshot(), Document, legacy(), snapshot(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes() (+22 more)

### Community 129 - "service_windows.go"
Cohesion: 0.09
Nodes (13): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), executes(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths() (+5 more)

### Community 134 - "`internal/proto`"
Cohesion: 0.07
Nodes (47): Options, `internal/proto`, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsRefusesAnythingButAStream() (+39 more)

### Community 137 - "Zone"
Cohesion: 0.32
Nodes (5): checkDNS(), Domains(), TestDomains(), TestZone(), Zone()

## Knowledge Gaps
- **5 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `stream`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 330 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `on()` connect `htmx.min.js` to `.mesh`, `kernel`, `runArgs`?**
  _High betweenness centrality (0.076) - this node is a cross-community bridge._
- **Are the 222 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 222 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _5 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07142959001782531 - nodes in this community are weakly interconnected._
- **Why does ``cmd/meshctl`` connect `runArgs` to `htmx.min.js`, `README.md`, `commands.go`, `action`, `Service`?**
  _High betweenness centrality (0.058) - this node is a cross-community bridge._
- **Should `Run` be split into smaller, more focused modules?**
  _Cohesion score 0.09523809523809523 - nodes in this community are weakly interconnected._
- **Why does `Command reference` connect `Command reference` to `README.md`, `commands.go`, `lab`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._