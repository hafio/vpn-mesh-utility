# Graph Report - vpn-mesh-utility  (2026-10-24)

## Corpus Check
- 286 files · ~336,639 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3554 nodes · 17320 edges · 111 communities (103 shown, 8 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2054 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `c43aa04e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- Node
- actor
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
- .add
- github.com/hafio/vpn-mesh-utility
- `internal/server/api`
- Run
- net/http.ResponseWriter
- context.Context
- PublishNames
- Dir
- ServerConfig
- Agent (`internal/agent/...`)
- net/netip.Prefix
- Net
- Session
- cli
- lab
- Snapshot
- .mesh
- Command reference
- time.Time
- Host
- runner
- newRig
- Service
- agent/agent_test.go
- Nodes
- `internal/server/auth`
- testing.T
- Policy
- .exec
- openTest
- otherNet
- netlinkNet
- Allocate
- planner_test.go
- Handler
- host
- README.md
- Host
- Run
- time.Duration
- context.Context
- LocalStatus
- Compose and stack files
- Tx
- .BeginEnrollment
- Client
- .routes
- net/netip.Addr
- meshDNS
- Deps
- Sync
- .Apply
- StatusReport
- .query
- Writer
- UI
- Registration
- Hub
- renderer
- Of
- validate.go
- Security
- CookieValue
- .apply
- Exit
- MeshNetworkSetup
- newRig
- 7. Agent
- resolved
- errorf
- conn
- Agent protocol
- .renderPolicies
- deploy_test.go
- service_windows.go
- fetchwintun/main_test.go
- 8. The web UI
- .Sync
- Ruleset
- noKernel
- .Reach
- userspace
- NodeConfig
- sync.Mutex
- serveRuntime
- .queryRow
- Schedule
- testing.T
- nameTable
- exprs
- Policy
- Extending the code
- DNS
- validate_test.go
- ns
- `internal/server` and `internal/server/export`
- NewNet
- Test catalogue

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 250 edges
2. `ui` - 90 edges
3. ``internal/server/api`` - 80 edges
4. `Session` - 74 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 67 edges
7. `newHarness()` - 58 edges
8. `env` - 48 edges
9. `runner` - 47 edges
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

## Communities (111 total, 8 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (32): AdminError, Error, composeNetwork, containerFile, containerService, nasFile, nasHealth, nasService (+24 more)

### Community 1 - "NodeConfig"
Cohesion: 0.16
Nodes (34): rig, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS() (+26 more)

### Community 2 - "Node"
Cohesion: 0.17
Nodes (11): Config and events, 4. Roles and topology, Peer, checkUsable(), ContainerNetwork(), endpoints(), Node, sortPeers() (+3 more)

### Community 3 - "actor"
Cohesion: 0.13
Nodes (9): SegPolicy, SegSchedule, listOf(), meshChanged(), scheduleText(), title(), changed(), actor() (+1 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.07
Nodes (106): Status, TestCreateLinkNeedsPrivilege(), A(), Ae(), an(), at(), B(), Be() (+98 more)

### Community 5 - "env"
Cohesion: 0.05
Nodes (98): watchdog, agentCommands(), agentLog(), joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall() (+90 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (28): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments (+20 more)

### Community 7 - "AddrEntry"
Cohesion: 0.11
Nodes (26): addrs(), contains(), ofFamily(), Service, inRanges(), inRange(), checkSegmentation(), hosts() (+18 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.21
Nodes (12): Rule, Violation, Packages, and which way they may import, Input and commands, `internal/archtest`, Check(), importsOf(), ModuleRoot() (+4 more)

### Community 11 - "net.Listener"
Cohesion: 0.33
Nodes (6): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), serve()

### Community 12 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.07
Nodes (77): brokenStream, countingDB, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput() (+69 more)

### Community 15 - "Run"
Cohesion: 0.15
Nodes (17): NewNodes(), New(), NewAdmin(), newAPI(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries(), NewRateLimiter() (+9 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.14
Nodes (10): api, decode(), reply(), ETag(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+2 more)

### Community 17 - "context.Context"
Cohesion: 0.08
Nodes (18): fakeService, IPv6Request, IPv6View, NodeAddRequest, NodeSetRequest, dial(), dial(), isPipe() (+10 more)

### Community 18 - "PublishNames"
Cohesion: 0.15
Nodes (14): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+6 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (27): `internal/fsutil`, TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious() (+19 more)

### Community 20 - "ServerConfig"
Cohesion: 0.21
Nodes (27): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+19 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.06
Nodes (55): Agent (`internal/agent/...`), DefaultSocket(), TestCleanHostname(), TestDefaultSocket(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx(), TestConfigureHasKernelSemantics() (+47 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.10
Nodes (14): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+6 more)

### Community 24 - "Session"
Cohesion: 0.12
Nodes (9): Session, group(), atoi(), before(), versionNumber(), message(), writeHTML(), settingsData (+1 more)

### Community 25 - "cli"
Cohesion: 0.19
Nodes (19): cli(), TestIPv6DNSRecords(), TestReportedHostFamilies(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness (+11 more)

### Community 26 - "lab"
Cohesion: 0.27
Nodes (6): Coverage, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.10
Nodes (24): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+16 more)

### Community 28 - ".mesh"
Cohesion: 0.08
Nodes (25): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+17 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.07
Nodes (20): replayCache, dialing, failures, hitWindow, Lockout, LockoutError, ago(), serves() (+12 more)

### Community 31 - "Host"
Cohesion: 0.16
Nodes (4): link, exitHost, Host, TestLinksAddressesAndFailures()

### Community 32 - "runner"
Cohesion: 0.13
Nodes (6): runner, kick(), oneLine(), TestOneLine(), TestReportKeepsToTheServersLimits(), drain()

### Community 33 - "newRig"
Cohesion: 0.13
Nodes (38): Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, TestTemplatesHaveNoScript(), TestDashboardHealth(), TestFirewallPages(), TestNodeGroupsOnPages(), newQR() (+30 more)

### Community 34 - "Service"
Cohesion: 0.17
Nodes (8): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), serve()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (58): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+50 more)

### Community 36 - "Nodes"
Cohesion: 0.12
Nodes (19): Change, NetworkInfo, NodeDetail, NodeList, PhoneProfile, PhoneRequest, addresses(), heard() (+11 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.13
Nodes (26): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures() (+18 more)

### Community 38 - "testing.T"
Cohesion: 0.11
Nodes (43): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+35 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (6): mustAffect(), joinList(), boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.11
Nodes (48): `internal/server/store`, TestExportOnceSkipsUnchangedAndKeepsFourteen(), DefaultPolicy(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules() (+40 more)

### Community 43 - "netlinkNet"
Cohesion: 0.11
Nodes (13): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+5 more)

### Community 44 - "Allocate"
Cohesion: 0.24
Nodes (14): Allocate(), Allocate6(), AllocateContainers(), AllocateContainers6(), Block(), Block6(), CheckMeshCIDR(), CheckMeshCIDR6() (+6 more)

### Community 45 - "planner_test.go"
Cohesion: 0.10
Nodes (54): PlanError, Add a topology rule, `internal/server/planner`, containers6(), Check(), checkContainers(), checkContainers6(), checkNode6() (+46 more)

### Community 46 - "Handler"
Cohesion: 0.24
Nodes (7): Client, Logs, Handler(), logs(), refusal(), UIToken, Funcs

### Community 47 - "host"
Cohesion: 0.16
Nodes (11): TestUnitPosture(), SystemdUnit(), Agent service (systemd), Runner, exists(), host(), TestAvailable(), TestInstallRefusesABinaryElsewhere() (+3 more)

### Community 48 - "README.md"
Cohesion: 0.15
Nodes (9): graphify, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together, meshctl (+1 more)

### Community 49 - "Host"
Cohesion: 0.09
Nodes (35): peer(), AddrPort(), TestAddrPort(), hexKey(), ParseUAPI(), getAnswer(), hexOf(), pfxs() (+27 more)

### Community 50 - "Run"
Cohesion: 0.20
Nodes (7): Run(), Listen(), New(), TestServeAndListen(), TestOpen(), Open(), TestOpenRefusesUnknownBackends()

### Community 51 - "time.Duration"
Cohesion: 0.13
Nodes (8): Failover, hubMove, hubState, waitApplied(), Exec, HashToken(), NewToken(), Config

### Community 52 - "context.Context"
Cohesion: 0.11
Nodes (25): Config, DB, Params, Pending, addUser(), newPassword(), passwdUser(), resetTOTP() (+17 more)

### Community 53 - "LocalStatus"
Cohesion: 0.14
Nodes (13): laptop(), PrefsRequest, PrefsResponse, LocalStatus, Config, dashboardData, fakeAgent, Funcs (+5 more)

### Community 54 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 55 - "Tx"
Cohesion: 0.13
Nodes (34): apiError, ComposeFile, ComposeOptions, Role, addIPv6(), blockOf(), checkGroups(), containerSubnet() (+26 more)

### Community 56 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - ".routes"
Cohesion: 0.42
Nodes (7): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 59 - "net/netip.Addr"
Cohesion: 0.09
Nodes (24): hostService(), hostAddrs, TestIsPublic(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), syncProxies(), firstAddr(), Service (+16 more)

### Community 60 - "meshDNS"
Cohesion: 0.25
Nodes (22): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+14 more)

### Community 61 - "Deps"
Cohesion: 0.24
Nodes (6): Deps, Deps, DNS, Firewall, MeshFirewall, Sysctls

### Community 62 - "Sync"
Cohesion: 0.10
Nodes (30): network, Sysctls, NewNet(), ensureAddr(), Gateway(), Net, Rule, Guard() (+22 more)

### Community 63 - ".Apply"
Cohesion: 0.15
Nodes (11): applied, Firewall, policy, Firewall, normalize(), richRule(), RuleArgv(), wantedPolicies() (+3 more)

### Community 64 - "StatusReport"
Cohesion: 0.29
Nodes (7): statusDetail, PeerStatus, Rollback, StatusReport, TestStatusReportValidate(), validStatus(), sameRollback()

### Community 65 - ".query"
Cohesion: 0.14
Nodes (7): ContainerName, DNSRecord, scanAll(), splitList(), AuditEntry, PeerStat, StatusUpdate

### Community 66 - "Writer"
Cohesion: 0.29
Nodes (3): New(), New(), Writer

### Community 67 - "UI"
Cohesion: 0.15
Nodes (10): hostOf(), oldest(), validOrNot(), bytesize(), prefsFlash(), ULAPrefix(), handler, page (+2 more)

### Community 68 - "Registration"
Cohesion: 0.14
Nodes (18): Domain, fakeRegistrar, nrpt, nrptRule, Registration, desired(), domainName(), fromRules() (+10 more)

### Community 69 - "Hub"
Cohesion: 0.14
Nodes (9): Deps, Hub, subscription, NewUserspace(), ConfigEvent, DB, Reconciler, NewReconciler() (+1 more)

### Community 70 - "renderer"
Cohesion: 0.25
Nodes (5): loadPages(), newRenderer(), pages, page, renderer

### Community 71 - "Of"
Cohesion: 0.18
Nodes (10): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+2 more)

### Community 72 - "validate.go"
Cohesion: 0.10
Nodes (44): RecordRequest, check(), ParseAddr(), TestParseAddr(), validateMeshNames(), newNode(), validateSet(), ToSnapshot() (+36 more)

### Community 73 - "Security"
Cohesion: 0.20
Nodes (9): Changes a node asks for, DNS, Interfaces the agent did not create, Security, The host firewall, The local UI (internal/agent/localui), Web UI sign-in (internal/server/auth), What a node sees of the others (+1 more)

### Community 74 - "CookieValue"
Cohesion: 0.21
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), sameOrigin(), handler

### Community 75 - ".apply"
Cohesion: 0.44
Nodes (3): pendingApply, safeApply, TestSafeApplyStateMachine()

### Community 77 - "MeshNetworkSetup"
Cohesion: 0.31
Nodes (9): checkIPv6(), containerHeader(), MeshNetworkSetup(), networkCommands(), resolvers(), serviceExample(), TestHostModeNetworkSetup(), NetworkParams (+1 more)

### Community 78 - "newRig"
Cohesion: 0.20
Nodes (12): capped, golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed() (+4 more)

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.11
Nodes (21): segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange, meshChange() (+13 more)

### Community 82 - "conn"
Cohesion: 0.43
Nodes (3): conn(), hasLegacy(), nftKernel

### Community 83 - "Agent protocol"
Cohesion: 0.15
Nodes (13): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+5 more)

### Community 84 - ".renderPolicies"
Cohesion: 0.09
Nodes (21): anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm(), ruleRows(), dnsState() (+13 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.09
Nodes (49): hostSocket(), renderContainer(), RenderContainerCompose(), containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposeHeader(), TestContainerComposePosture() (+41 more)

### Community 86 - "service_windows.go"
Cohesion: 0.05
Nodes (20): entry, StepError, hostService(), uninstallAgent(), NewNet(), NewKernel(), NewKernel(), commandLine() (+12 more)

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.11
Nodes (22): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+14 more)

### Community 88 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

### Community 90 - "Ruleset"
Cohesion: 0.17
Nodes (7): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, packet

### Community 97 - ".Reach"
Cohesion: 0.13
Nodes (12): ReachRequest, SegmentView, ZoneView, activeAt(), ReachResult, zoneOf(), Overlaps(), Summarize() (+4 more)

### Community 98 - "userspace"
Cohesion: 0.09
Nodes (8): lookup(), NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, links, userspace

### Community 99 - "NodeConfig"
Cohesion: 0.06
Nodes (38): IPv6Host, Rechecker, Determinism and goldens, Add an apply step, Keys and tokens, `internal/server/phone`, Shared fixtures -- `internal/testfix`, forHost() (+30 more)

### Community 100 - "sync.Mutex"
Cohesion: 0.12
Nodes (5): syncBuffer, Sysctl, UAPISet(), Sysctl, wgdev

### Community 101 - "serveRuntime"
Cohesion: 0.13
Nodes (15): TestOpenRuntime(), clip(), Detect(), meshFilter(), New(), Sockets(), inspectOf(), serveRuntime() (+7 more)

### Community 103 - ".queryRow"
Cohesion: 0.13
Nodes (5): scanNode(), User, scanUser(), scanner, Version

### Community 104 - "Schedule"
Cohesion: 0.12
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "testing.T"
Cohesion: 0.06
Nodes (69): Options, `internal/proto`, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents() (+61 more)

### Community 106 - "nameTable"
Cohesion: 0.18
Nodes (4): nameTable, zone, exchange(), records()

### Community 107 - "exprs"
Cohesion: 0.17
Nodes (8): exprs(), ifname(), intervalAddrElems(), intervalElems(), TestExprs(), TestIntervalAddrElems(), TestIntervalElems(), nftOp

### Community 108 - "Policy"
Cohesion: 0.19
Nodes (22): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+14 more)

### Community 111 - "Extending the code"
Cohesion: 0.15
Nodes (11): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+3 more)

### Community 113 - "DNS"
Cohesion: 0.17
Nodes (6): nameServer, DNS, Want, newNameTable(), DNS, MeshName

### Community 114 - "validate_test.go"
Cohesion: 0.08
Nodes (25): Security invariants, `internal/validate`, TestTimeZone(), TestAddrEntry(), TestAddrRangePrefix(), TestRef(), TestServiceEntry(), Arg() (+17 more)

### Community 118 - "ns"
Cohesion: 0.15
Nodes (11): host, lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), newLab(), ns() (+3 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.08
Nodes (34): `internal/server` and `internal/server/export`, NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses(), FromSnapshot(), Document (+26 more)

### Community 138 - "Test catalogue"
Cohesion: 0.16
Nodes (8): End to end -- `test/e2e`, `internal/tools/vulnjudge`, Test catalogue, stream, inNS(), lab, stream, TestMeshFirewall()

## Knowledge Gaps
- **6 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `lab` (+1 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 338 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `NodeConfig`, `Node`, `htmx.min.js`, `env`, `archtest.go`, `Test catalogue`, `Dir`, `ServerConfig`, `.mesh`, `time.Time`, `Host`, `runner`, `newRig`, `agent/agent_test.go`, `Policy`, `netlinkNet`, `host`, `Host`, `Run`, `.routes`, `net/netip.Addr`, `meshDNS`, `Sync`, `Registration`, `Of`, `validate.go`, `.apply`, `newRig`, `resolved`, `service_windows.go`, `fetchwintun/main_test.go`, `NodeConfig`, `serveRuntime`, `testing.T`, `exprs`, `Policy`?**
  _High betweenness centrality (0.036) - this node is a cross-community bridge._
- **Are the 249 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 249 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _6 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.0756839327871129 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect ``internal/server/api`` to `newRig`, `htmx.min.js`, `testing.T`, `openTest`, `Test catalogue`, `Run`, `context.Context`, `PublishNames`, ``internal/server` and `internal/server/export``, `.routes`, `cli`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Should `actor` be split into smaller, more focused modules?**
  _Cohesion score 0.1339031339031339 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `.mesh`, `testing.T`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._