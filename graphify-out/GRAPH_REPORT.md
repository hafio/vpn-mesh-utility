# Graph Report - vpn-mesh-utility  (2026-10-22)

## Corpus Check
- 285 files · ~321,112 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3474 nodes · 16917 edges · 129 communities (103 shown, 26 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 1989 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f0effe95`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- .Apply
- Node
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- walk
- `internal/server/api`
- github.com/hafio/vpn-mesh-utility
- serve
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
- ui
- README.md
- Host
- netlinkPolicy
- otherNet
- .Tx
- context.Context
- Compose and stack files
- .BeginEnrollment
- Extending the code
- Client
- commands.go
- Matcher
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
- Run
- File
- otherPolicy
- RenderContainerCompose
- syncBuffer
- PolicyRule
- sync.Mutex
- errorf
- User
- Agent protocol
- .fail
- deploy_test.go
- Join
- fetchwintun/main_test.go
- Security
- AddrEntry
- Kernel
- service_windows.go
- Diff
- Error
- IPv6View
- .queryRow
- .Sync
- userspace
- Signed
- TestQRCode
- serveRuntime
- .settingsView
- .query
- Schedule
- Policy
- exprs
- noKernel
- net/netip.Addr
- NodeConfig
- .Reach
- Ring
- `internal/validate`
- nft_linux.go
- ns
- loadMigrations
- `internal/server` and `internal/server/export`
- io.Reader
- TestMain
- entry
- Error
- NewNet
- NewNet
- NewKernel
- NewKernel
- `internal/proto`
- NewKernel
- Refusal
- StatusError

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 238 edges
2. `ui` - 90 edges
3. ``internal/server/api`` - 76 edges
4. `Session` - 74 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 64 edges
7. `newHarness()` - 55 edges
8. `env` - 48 edges
9. `runner` - 46 edges
10. `api` - 42 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Dev script tasks` --references--> `build()`  [INFERRED]
  docs/developer.md → internal/agent/policy/nft_linux.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
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

## Communities (129 total, 26 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (23): composeNetwork, containerFile, containerService, nasFile, nasHealth, nasService, linkDNS, DNSRecord (+15 more)

### Community 1 - "NodeConfig"
Cohesion: 0.09
Nodes (51): rig, TestRunNeedsAnEnrollment(), NewHost(), NewRoutes(), TestLinksAddressesAndFailures(), TestRoutesSysctlAndFirewall(), IsStepError(), New() (+43 more)

### Community 2 - ".Apply"
Cohesion: 0.11
Nodes (10): dialing, IPv6Host, Rechecker, Add an apply step, forHost(), Engine, Desired(), Host (+2 more)

### Community 3 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.22
Nodes (31): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+23 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (27): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments, 1. Problem (+19 more)

### Community 7 - "AddrEntry"
Cohesion: 0.13
Nodes (18): Service, activeAt(), zoneOf(), checkSegmentation(), Segmentation, host(), mergeServices(), newSegResolver() (+10 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.14
Nodes (16): Rule, Violation, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, Input and commands (+8 more)

### Community 11 - "walk"
Cohesion: 0.83
Nodes (3): admit(), walk(), packet

### Community 12 - "`internal/server/api`"
Cohesion: 0.08
Nodes (50): apiError, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket() (+42 more)

### Community 14 - "serve"
Cohesion: 0.11
Nodes (34): brokenStream, countingDB, harness, joinBody(), postJoin(), serve(), signed(), storeNode() (+26 more)

### Community 15 - "Run"
Cohesion: 0.08
Nodes (32): Deps, Route, route, Add an agent API route, NewUserspace(), NewNodes(), AdminRoutes(), describe() (+24 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.15
Nodes (9): api, decode(), reply(), cli(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+1 more)

### Community 17 - ".do"
Cohesion: 0.09
Nodes (25): NetworkInfo, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, SegmentView, statusDetail (+17 more)

### Community 18 - "New"
Cohesion: 0.29
Nodes (10): Config, Params, Pending, New(), DefaultParams(), HashPassword(), TestHashUsesTheSpecParameters(), TestVerifyPassword() (+2 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (24): `internal/fsutil`, Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip() (+16 more)

### Community 20 - "ServerConfig"
Cohesion: 0.21
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+17 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.03
Nodes (86): Agent (`internal/agent/...`), Exec, TestCleanHostname(), TestDefaultSocket(), TestIsPublic(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine() (+78 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.10
Nodes (25): Routes, index(), route(), compare(), luidOf(), onLink(), ClientIP(), inAny() (+17 more)

### Community 23 - "time.Time"
Cohesion: 0.14
Nodes (9): replayCache, failures, hitWindow, Lockout, LockoutError, Sysctl, sgt(), RateLimiter (+1 more)

### Community 24 - "ui"
Cohesion: 0.14
Nodes (7): Session, listOf(), meshChanged(), atoi(), changed(), message(), actor()

### Community 26 - "lab"
Cohesion: 0.18
Nodes (10): Coverage, host, deleteNamespaces(), eventually(), lab, newLab(), tail(), TestServerMesh() (+2 more)

### Community 27 - "Snapshot"
Cohesion: 0.15
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - ".nodeView"
Cohesion: 0.27
Nodes (6): ago(), dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.15
Nodes (10): Failover, hubMove, hubState, waitApplied(), harness, TestFailoverMovesRelays(), TestFailoverRun(), TestJudgeHubs() (+2 more)

### Community 31 - "commands.go"
Cohesion: 0.24
Nodes (11): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), shownDefault(), cell() (+3 more)

### Community 32 - "runner"
Cohesion: 0.08
Nodes (14): fakeService, pendingApply, runner, safeApply, kick(), oneLine(), Result, dial() (+6 more)

### Community 33 - "newRig"
Cohesion: 0.16
Nodes (33): The web UI's pages (internal/server/ui), `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), group(), bytesize(), TestPhonePage(), code() (+25 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.11
Nodes (58): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+50 more)

### Community 36 - "Input"
Cohesion: 0.12
Nodes (14): PlanError, Record, containers6(), Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable() (+6 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (27): rig, 11. CLI reference, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide() (+19 more)

### Community 38 - "testing.T"
Cohesion: 0.08
Nodes (66): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+58 more)

### Community 39 - "Policy"
Cohesion: 0.15
Nodes (7): Policy, mutate(), Policy, Route, Rule, V6Support, Sync()

### Community 40 - ".exec"
Cohesion: 0.08
Nodes (8): mustAffect(), DNSRecord, joinList(), boolInt(), classify(), nullInt(), nullStr(), StatusUpdate

### Community 41 - "openTest"
Cohesion: 0.12
Nodes (48): `internal/server/store`, TestExportOnceSkipsUnchangedAndKeepsFourteen(), DefaultPolicy(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules() (+40 more)

### Community 42 - ".mesh"
Cohesion: 0.15
Nodes (15): Routes, 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh (+7 more)

### Community 43 - "netlinkNet"
Cohesion: 0.09
Nodes (12): netlinkNet, Status, Owned(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule() (+4 more)

### Community 45 - "planner_test.go"
Cohesion: 0.14
Nodes (43): Add a topology rule, `internal/server/planner`, TestCheckRequested(), Compute(), deliver(), deliverAddr(), encode(), firstDiff() (+35 more)

### Community 46 - "meshDNS"
Cohesion: 0.16
Nodes (23): desired(), same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects() (+15 more)

### Community 47 - "ui"
Cohesion: 0.11
Nodes (10): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), before(), versionNumber(), writeHTML(), sameOrigin() (+2 more)

### Community 48 - "README.md"
Cohesion: 0.17
Nodes (7): `internal/tools/vulnjudge`, Test catalogue, Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "Host"
Cohesion: 0.09
Nodes (17): link, exitHost, Host, peer(), AddrPort(), hexKey(), ParseUAPI(), uapiInt() (+9 more)

### Community 50 - "netlinkPolicy"
Cohesion: 0.20
Nodes (7): Family, netlinkPolicy, dump(), family(), owned(), toRoute(), toRule()

### Community 52 - ".Tx"
Cohesion: 0.15
Nodes (16): DB, addUser(), newPassword(), passwdUser(), resetTOTP(), fullExpiry(), Service, load() (+8 more)

### Community 53 - "context.Context"
Cohesion: 0.10
Nodes (21): Add a web UI page, Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest (+13 more)

### Community 54 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 55 - ".BeginEnrollment"
Cohesion: 0.27
Nodes (5): Sealer, aad(), newTOTPKey(), TestMatchStepAllowsOneStepOfSkew(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 56 - "Extending the code"
Cohesion: 0.22
Nodes (9): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+1 more)

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "commands.go"
Cohesion: 0.20
Nodes (15): agentCommands(), commands(), runWith(), firewallCommands(), healthCommands(), find(), isHelp(), main() (+7 more)

### Community 59 - "Matcher"
Cohesion: 0.29
Nodes (4): serves(), Service, Matcher, matchRule

### Community 60 - "Service"
Cohesion: 0.06
Nodes (28): nameServer, DNSEntry, NameEntry, NameSet, Registrar, resolver, Service, DNS (+20 more)

### Community 61 - "Tx"
Cohesion: 0.08
Nodes (46): Change, ComposeFile, ComposeOptions, PhoneRequest, RecordRequest, Role, addIPv6(), blockOf() (+38 more)

### Community 62 - "Net"
Cohesion: 0.11
Nodes (13): Net, network, Sysctls, ensureAddr(), Gateway(), Net, Rule, Guard() (+5 more)

### Community 63 - "service_test.go"
Cohesion: 0.09
Nodes (18): hostService(), SystemdUnit(), applied, Firewall, policy, Firewall, Runner, atLeast() (+10 more)

### Community 64 - "Container"
Cohesion: 0.20
Nodes (9): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+1 more)

### Community 65 - "Container"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 66 - "adminFlags"
Cohesion: 0.15
Nodes (27): dnsLine(), printExit(), setupFirewallShow(), setupIPv6(), adminFlags(), cleared(), joinNonEmpty(), nodeAction() (+19 more)

### Community 67 - "UI"
Cohesion: 0.11
Nodes (16): hostOf(), New(), oldest(), validOrNot(), ago(), bytesize(), loadPages(), prefsFlash() (+8 more)

### Community 68 - "Registration"
Cohesion: 0.10
Nodes (22): Domain, fakeRegistrar, linkDomain, nrpt, nrptRule, Registration, resolved, domainName() (+14 more)

### Community 69 - "Hub"
Cohesion: 0.16
Nodes (8): Hub, subscription, ConfigEvent, NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 70 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 71 - ".Apply"
Cohesion: 0.16
Nodes (13): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+5 more)

### Community 72 - "Name"
Cohesion: 0.10
Nodes (41): check(), cleanHostname(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), checkDNS(), Endpoint6() (+33 more)

### Community 74 - "Run"
Cohesion: 0.13
Nodes (13): Deps, Deps, Run(), DNS, Firewall, MeshFirewall, Sysctls, NewTable() (+5 more)

### Community 75 - "File"
Cohesion: 0.31
Nodes (4): agentLog(), ttyPassword(), OpenFile(), File

### Community 77 - "RenderContainerCompose"
Cohesion: 0.13
Nodes (21): containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers(), serviceExample() (+13 more)

### Community 79 - "PolicyRule"
Cohesion: 0.33
Nodes (5): Matrix, MatrixCell, PolicyRule, Schedule, SegPolicy

### Community 81 - "errorf"
Cohesion: 0.12
Nodes (19): segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange, meshChange() (+11 more)

### Community 82 - "User"
Cohesion: 0.39
Nodes (3): User, scanUser(), scanner

### Community 83 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+2 more)

### Community 84 - ".fail"
Cohesion: 0.06
Nodes (31): SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+23 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.11
Nodes (42): containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposeHeader(), TestContainerComposePosture(), TestContainerComposeRefusesHostileValues(), TestContainerRendererHoldsItsStructure(), TestHostModeNetworkSetup() (+34 more)

### Community 86 - "Join"
Cohesion: 0.15
Nodes (14): Options, installAgent(), installAgent(), Join(), New(), TestUnreachableAndMalformedAnswers(), JoinOptions, Service (+6 more)

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.16
Nodes (19): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestInstall(), download() (+11 more)

### Community 88 - "Security"
Cohesion: 0.19
Nodes (13): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Security (+5 more)

### Community 89 - "AddrEntry"
Cohesion: 0.15
Nodes (21): addrs(), addrText(), Compile(), contains(), Local, Rule, merged(), portText() (+13 more)

### Community 90 - "Kernel"
Cohesion: 0.21
Nodes (4): watchingFirewall, Kernel, Ruleset, Chain

### Community 92 - "service_windows.go"
Cohesion: 0.12
Nodes (8): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), Windows(), handler, SCM

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 95 - "IPv6View"
Cohesion: 0.40
Nodes (3): IPv6Request, IPv6View, ipv6View()

### Community 96 - ".queryRow"
Cohesion: 0.15
Nodes (4): TestDeleteNodeCascades(), TestTokenLifecycle(), HashToken(), NewToken()

### Community 98 - "userspace"
Cohesion: 0.11
Nodes (6): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace

### Community 100 - "TestQRCode"
Cohesion: 0.50
Nodes (4): newQR(), TestQRCode(), phoneData, qrCode

### Community 101 - "serveRuntime"
Cohesion: 0.21
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 103 - ".query"
Cohesion: 0.10
Nodes (11): compareAddr(), containerAddr(), PublishNames(), TestContainerAddr(), TestPublishNames(), ContainerName, scanAll(), splitList() (+3 more)

### Community 104 - "Schedule"
Cohesion: 0.21
Nodes (7): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), weekday()

### Community 105 - "Policy"
Cohesion: 0.13
Nodes (23): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), compile(), officeHours(), sgt() (+15 more)

### Community 107 - "exprs"
Cohesion: 0.21
Nodes (5): exprs(), ifname(), intervalElems(), TestExprs(), nftOp

### Community 110 - "net/netip.Addr"
Cohesion: 0.12
Nodes (17): nameTable, zone, firstAddr(), isPublic(), isPublic6(), PublicIPv4(), PublicIPv6(), JoinRequest (+9 more)

### Community 111 - "NodeConfig"
Cohesion: 0.07
Nodes (37): Determinism and goldens, Keys and tokens, Shared fixtures -- `internal/testfix`, NewNet(), TestForHostWithoutIPv6(), TestGatewayRules(), TestGuardRules(), TestSyncContainerMode() (+29 more)

### Community 112 - ".Reach"
Cohesion: 0.29
Nodes (6): ReachRequest, flowEnd(), ReachResult, matrixCell, matrixRow, reachView

### Community 114 - "`internal/validate`"
Cohesion: 0.09
Nodes (23): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), LocalToken(), mask() (+15 more)

### Community 116 - "nft_linux.go"
Cohesion: 0.13
Nodes (5): build(), conn(), NewKernel(), nftBatch, nftKernel

### Community 118 - "ns"
Cohesion: 0.17
Nodes (10): End to end -- `test/e2e`, lab, TestContainerNetworks(), inNS(), ns(), lab, TestRelayFailover(), lab (+2 more)

### Community 119 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.11
Nodes (28): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+20 more)

### Community 122 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 134 - "`internal/proto`"
Cohesion: 0.07
Nodes (44): `internal/proto`, newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents(), TestEventsRefusesAnythingButAStream(), TestMeshIsCheckedBeforeUse(), TestNamesRoundTrip() (+36 more)

## Knowledge Gaps
- **5 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `stream`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 335 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **26 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does ``internal/server/api`` connect ``internal/server/api`` to `UI`, `htmx.min.js`, `serveRuntime`, `Hub`, `.query`, `openTest`, `net/netip.Addr`, `serve`, `Run`, `.do`, `README.md`, `deploy_test.go`, `time.Time`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Are the 237 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 237 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _5 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08027999397862412 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `.mesh`, `netlinkNet`, `testing.T`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.09071117561683599 - nodes in this community are weakly interconnected._
- **Why does `at()` connect `htmx.min.js` to ``internal/server/api``?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._