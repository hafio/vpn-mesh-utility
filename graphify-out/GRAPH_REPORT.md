# Graph Report - vpn-mesh-utility  (2026-10-22)

## Corpus Check
- 277 files · ~296,479 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3356 nodes · 16054 edges · 131 communities (110 shown, 21 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 1860 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f574af99`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- storeNode
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
- `internal/server/api`
- Run
- net/http.ResponseWriter
- context.Context
- action
- Dir
- ServerConfig
- Agent (`internal/agent/...`)
- net/netip.Prefix
- meshctl/main.go
- .nodeView
- Role
- lab
- Snapshot
- Session
- Command reference
- time.Time
- Key
- runner
- newRig
- TestMeshFirewall
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
- time.Time
- Security
- otherNet
- Service
- context.Context
- nrpt.go
- key
- Diff
- Client
- Open
- DNS
- Service
- Tx
- Sync
- service_test.go
- Container
- Container
- Compose and stack files
- UI
- Registration
- log/slog.Logger
- Agent protocol
- .Apply
- fail
- StepError
- .Apply
- TestMain
- net/netip.AddrPort
- MeshNetworkSetup
- newRig
- Installer
- .query
- errorf
- resolved
- Error
- .fail
- deploy_test.go
- Run
- fetchwintun/main_test.go
- Policy
- AddrEntry
- NewNet
- service_windows.go
- route
- NewKernel
- Refusal
- .queryRow
- sync.Mutex
- userspace
- File
- kernel
- serveRuntime
- Extending the code
- PublishNames
- Schedule
- Add a web UI page
- NewHub
- nft_linux.go
- WriteEvent
- JoinRequest
- net/netip.Addr
- NodeConfig
- Matcher
- HashPassword
- `internal/validate`
- .BeginEnrollment
- .Segment
- NewNet
- TestContainerNetworks
- Signed
- NewKernel
- `internal/server` and `internal/server/export`
- TestQRCode
- Error
- stream
- AdminError
- entry
- .relayFor
- NewKernel
- `internal/proto`
- StatusError

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 230 edges
2. `ui` - 88 edges
3. `Session` - 73 edges
4. ``internal/server/api`` - 70 edges
5. `ServerConfig()` - 64 edges
6. `NodeConfig` - 59 edges
7. `newHarness()` - 52 edges
8. `env` - 47 edges
9. `runner` - 46 edges
10. `api` - 40 edges

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

## Communities (131 total, 21 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (23): composeNetwork, containerFile, containerService, nasFile, nasHealth, nasService, linkDNS, DNSRecord (+15 more)

### Community 1 - "NodeConfig"
Cohesion: 0.17
Nodes (33): rig, watchingFirewall, NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS() (+25 more)

### Community 2 - "storeNode"
Cohesion: 0.15
Nodes (27): brokenStream, countingDB, ErrorBody, harness, joinBody(), postJoin(), serve(), signed() (+19 more)

### Community 3 - "Node"
Cohesion: 0.25
Nodes (9): Config and events, 4. Roles and topology, Peer, Record, ContainerNetwork(), Node, sortPeers(), sortPrefixes() (+1 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.21
Nodes (32): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+24 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (35): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments, 1. Problem (+27 more)

### Community 7 - "AddrEntry"
Cohesion: 0.16
Nodes (17): services(), Service, checkSegmentation(), host(), mergeServices(), newSegResolver(), Overlaps(), parseSpans() (+9 more)

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
Cohesion: 0.11
Nodes (9): Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel, noKernel (+1 more)

### Community 12 - "meshHarness"
Cohesion: 0.18
Nodes (16): ErrorMessage(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness, officeFirewall(), TestEditSegment() (+8 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.13
Nodes (33): `internal/server/api`, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose(), TestAdminContainerMode(), TestAdminErrorCodes() (+25 more)

### Community 15 - "Run"
Cohesion: 0.14
Nodes (18): NewNodes(), New(), NewAdmin(), newAPI(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallScriptPinsTheChecksums(), newReplayCache() (+10 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.15
Nodes (12): api, apiError, cli(), decodeJSON(), errorf(), readBody(), toAPIError(), writeJSON() (+4 more)

### Community 17 - "context.Context"
Cohesion: 0.11
Nodes (14): fakeService, dial(), dial(), isPipe(), Source, RunWindowsService(), AdminClient, nodePath() (+6 more)

### Community 18 - "action"
Cohesion: 0.17
Nodes (24): setupFirewallApply(), setupFirewallShow(), adminFlags(), cleared(), heard(), joinNonEmpty(), nodeAction(), nodeName() (+16 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (24): `internal/fsutil`, Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip() (+16 more)

### Community 20 - "ServerConfig"
Cohesion: 0.21
Nodes (24): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+16 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.04
Nodes (76): Agent (`internal/agent/...`), TestDefaultSocket(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), NewNet(), pfx() (+68 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.10
Nodes (15): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+7 more)

### Community 23 - "meshctl/main.go"
Cohesion: 0.14
Nodes (22): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+14 more)

### Community 24 - ".nodeView"
Cohesion: 0.24
Nodes (8): ago(), bytesize(), TestHelpers(), dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 25 - "Role"
Cohesion: 0.27
Nodes (11): Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateContainers(), TestAllocateLowestFree(), TestBlockMatchesTheSpecAddressPlan() (+3 more)

### Community 26 - "lab"
Cohesion: 0.17
Nodes (13): End to end -- `test/e2e`, host, deleteNamespaces(), eventually(), lab, inNS(), newLab(), ns() (+5 more)

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation(), fields() (+10 more)

### Community 28 - "Session"
Cohesion: 0.11
Nodes (11): Session, group(), atoi(), before(), changed(), versionNumber(), message(), writeHTML() (+3 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (33): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+25 more)

### Community 30 - "time.Time"
Cohesion: 0.14
Nodes (9): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, TestJudgeHubs(), RateLimiter (+1 more)

### Community 31 - "Key"
Cohesion: 0.09
Nodes (16): link, exitHost, Host, peer(), hexKey(), ParseUAPI(), uapiInt(), compareKeys() (+8 more)

### Community 32 - "runner"
Cohesion: 0.12
Nodes (7): pendingApply, runner, safeApply, kick(), oneLine(), drain(), TestSafeApplyStateMachine()

### Community 33 - "newRig"
Cohesion: 0.21
Nodes (26): `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), TestPhonePage(), code(), cookieOf(), newRig(), TestBadSignInsAreRefusedAndThrottled() (+18 more)

### Community 34 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.11
Nodes (58): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+50 more)

### Community 36 - "Input"
Cohesion: 0.13
Nodes (25): PlanError, `internal/server/planner`, TestCheckRequested(), Check(), checkContainers(), Problem, sortedByID(), Compute() (+17 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (28): rig, 11. CLI reference, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword() (+20 more)

### Community 38 - "testing.T"
Cohesion: 0.07
Nodes (74): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+66 more)

### Community 39 - "Policy"
Cohesion: 0.07
Nodes (15): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Policy, Route, Rule (+7 more)

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (6): mustAffect(), joinList(), boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.15
Nodes (38): `internal/server/store`, DefaultPolicy(), LegacySegmentation(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate() (+30 more)

### Community 42 - ".mesh"
Cohesion: 0.13
Nodes (17): Routes, 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh (+9 more)

### Community 43 - "netlinkNet"
Cohesion: 0.16
Nodes (7): netlinkNet, Owned(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule()

### Community 45 - "planner_test.go"
Cohesion: 0.21
Nodes (23): Add a topology rule, deliver(), deliverAddr(), encode(), firstDiff(), lookup(), mustCompute(), node() (+15 more)

### Community 46 - "meshDNS"
Cohesion: 0.21
Nodes (22): desired(), same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects() (+14 more)

### Community 47 - "net/http.ResponseWriter"
Cohesion: 0.15
Nodes (8): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), sameOrigin(), handler

### Community 48 - "README.md"
Cohesion: 0.14
Nodes (9): graphify, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together, meshctl (+1 more)

### Community 49 - "time.Time"
Cohesion: 0.13
Nodes (15): replayCache, failures, Lockout, LockoutError, ago(), sgt(), Health(), phoneHealth() (+7 more)

### Community 50 - "Security"
Cohesion: 0.12
Nodes (20): TestUnitPosture(), SystemdUnit(), 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS (+12 more)

### Community 52 - "Service"
Cohesion: 0.14
Nodes (14): Pending, addUser(), newPassword(), passwdUser(), resetTOTP(), Service, session(), CheckPassword() (+6 more)

### Community 53 - "context.Context"
Cohesion: 0.08
Nodes (25): Client, Logs, logs(), refusal(), New(), ParseAddr(), laptop(), loadPages() (+17 more)

### Community 54 - "nrpt.go"
Cohesion: 0.20
Nodes (14): nrpt, nrptRule, domainName(), fromRules(), nrptKey(), ownedKey(), parseServers(), rulesFor() (+6 more)

### Community 55 - "key"
Cohesion: 0.22
Nodes (16): getAnswer(), hexOf(), pfxs(), TestParseUAPI(), TestParseUAPIRejectsMalformedValues(), TestUAPISetGolden(), uapiChange(), Desired() (+8 more)

### Community 56 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "Open"
Cohesion: 0.18
Nodes (9): loadMigrations(), openAt(), Store, Open(), TestLoadMigrationsValidatesNames(), TestOpenIsIdempotent(), TestOpenRefusesNewerSchema(), TestOpenRejectsDSNInjection() (+1 more)

### Community 59 - "DNS"
Cohesion: 0.20
Nodes (5): nameServer, DNS, Want, newNameTable(), MeshName

### Community 60 - "Service"
Cohesion: 0.14
Nodes (9): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+1 more)

### Community 61 - "Tx"
Cohesion: 0.07
Nodes (44): Change, ComposeFile, ComposeOptions, NetworkInfo, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest (+36 more)

### Community 62 - "Sync"
Cohesion: 0.11
Nodes (13): Net, network, Sysctls, ensureAddr(), Gateway(), Net, Rule, Guard() (+5 more)

### Community 63 - "service_test.go"
Cohesion: 0.14
Nodes (12): applied, Firewall, policy, Firewall, atLeast(), normalize(), richRule(), RuleArgv() (+4 more)

### Community 64 - "Container"
Cohesion: 0.20
Nodes (9): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+1 more)

### Community 65 - "Container"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 66 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 67 - "UI"
Cohesion: 0.16
Nodes (10): hostOf(), oldest(), validOrNot(), bytesize(), prefsFlash(), LocalToken(), handler, page (+2 more)

### Community 68 - "Registration"
Cohesion: 0.33
Nodes (5): Domain, fakeRegistrar, Registration, serverZone(), TestServerZone()

### Community 69 - "log/slog.Logger"
Cohesion: 0.14
Nodes (11): Deps, Hub, subscription, NewUserspace(), ConfigEvent, DB, Reconciler, NewReconciler() (+3 more)

### Community 70 - "Agent protocol"
Cohesion: 0.20
Nodes (10): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+2 more)

### Community 71 - ".Apply"
Cohesion: 0.20
Nodes (11): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+3 more)

### Community 72 - "fail"
Cohesion: 0.10
Nodes (37): check(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), checkDNS(), TimeZone(), Ref() (+29 more)

### Community 74 - ".Apply"
Cohesion: 0.13
Nodes (9): Rechecker, Add an apply step, Engine, Result, Desired(), Host, SyncAddr(), SyncLink() (+1 more)

### Community 76 - "net/netip.AddrPort"
Cohesion: 0.20
Nodes (9): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), Listen(), TestServeAndListen(), AddrPort() (+1 more)

### Community 77 - "MeshNetworkSetup"
Cohesion: 0.36
Nodes (8): containerHeader(), MeshNetworkSetup(), networkCommands(), resolvers(), serviceExample(), TestHostModeNetworkSetup(), NetworkParams, NetworkSetup

### Community 78 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 79 - "Installer"
Cohesion: 0.20
Nodes (6): hostService(), Runner, Service, exists(), Systemd(), Installer

### Community 80 - ".query"
Cohesion: 0.13
Nodes (7): ContainerName, DNSRecord, scanAll(), splitList(), AuditEntry, PeerStat, StatusUpdate

### Community 81 - "errorf"
Cohesion: 0.10
Nodes (24): RecordRequest, segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange (+16 more)

### Community 82 - "resolved"
Cohesion: 0.35
Nodes (4): linkDomain, resolved, classify(), foreignNetns()

### Community 84 - ".fail"
Cohesion: 0.07
Nodes (27): SegPolicy, SegSchedule, anyText(), listOf(), listText(), meshChanged(), newPolicyForm(), objectRows() (+19 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.08
Nodes (50): hostSocket(), renderContainer(), RenderContainerCompose(), containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposeHeader(), TestContainerComposePosture() (+42 more)

### Community 86 - "Run"
Cohesion: 0.15
Nodes (8): Deps, Deps, Sysctl, DNS, Firewall, MeshFirewall, Sysctls, Sysctl

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.15
Nodes (18): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), download(), extract() (+10 more)

### Community 88 - "Policy"
Cohesion: 0.14
Nodes (25): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+17 more)

### Community 89 - "AddrEntry"
Cohesion: 0.10
Nodes (21): clip(), Ring, addrs(), addrText(), contains(), Rule, merged(), portText() (+13 more)

### Community 92 - "service_windows.go"
Cohesion: 0.12
Nodes (8): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), Windows(), handler, SCM

### Community 93 - "route"
Cohesion: 0.42
Nodes (7): Route, route, Add an agent API route, AdminRoutes(), describe(), Routes(), TestRoutesAreDocumented()

### Community 96 - ".queryRow"
Cohesion: 0.14
Nodes (3): HashToken(), NewToken(), Version

### Community 97 - "sync.Mutex"
Cohesion: 0.17
Nodes (3): syncBuffer, capped, syncBuffer

### Community 98 - "userspace"
Cohesion: 0.12
Nodes (6): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace

### Community 99 - "File"
Cohesion: 0.17
Nodes (10): agentLog(), dnsLine(), printExit(), bytesize(), printStatus(), ttyPassword(), OpenFile(), Write() (+2 more)

### Community 100 - "kernel"
Cohesion: 0.18
Nodes (5): Status, IPNet(), lookup(), kernel, links

### Community 101 - "serveRuntime"
Cohesion: 0.14
Nodes (16): Run(), TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect() (+8 more)

### Community 102 - "Extending the code"
Cohesion: 0.22
Nodes (9): Add a command, Add a database migration, Add a validator, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+1 more)

### Community 103 - "PublishNames"
Cohesion: 0.20
Nodes (11): DNSEntry, NameEntry, NameSet, compareAddr(), LoadNames(), PublishNames(), TestPublishNames(), DNSView (+3 more)

### Community 104 - "Schedule"
Cohesion: 0.14
Nodes (12): Matrix, MatrixCell, The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule() (+4 more)

### Community 105 - "Add a web UI page"
Cohesion: 0.18
Nodes (10): Add a web UI page, The web UI's pages (internal/server/ui), dnsState(), newRenderer(), golden(), TestEveryPageRenders(), TestHTMXIsPinned(), TestPageGoldens() (+2 more)

### Community 106 - "NewHub"
Cohesion: 0.50
Nodes (5): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 107 - "nft_linux.go"
Cohesion: 0.09
Nodes (10): build(), conn(), exprs(), ifname(), intervalElems(), NewKernel(), TestExprs(), nftBatch (+2 more)

### Community 108 - "WriteEvent"
Cohesion: 0.25
Nodes (7): TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Event, Reader

### Community 109 - "JoinRequest"
Cohesion: 0.67
Nodes (3): JoinRequest, pickEndpoint(), TestPickEndpoint()

### Community 110 - "net/netip.Addr"
Cohesion: 0.31
Nodes (3): nameTable, zone, Record

### Community 111 - "NodeConfig"
Cohesion: 0.11
Nodes (12): Shared fixtures -- `internal/testfix`, NodeConfig, withDomains(), Hub(), Render(), golden(), TestHub(), TestProfiles() (+4 more)

### Community 112 - "Matcher"
Cohesion: 0.29
Nodes (4): serves(), Service, Matcher, matchRule

### Community 113 - "HashPassword"
Cohesion: 0.36
Nodes (9): Config, DB, Params, DefaultParams(), HashPassword(), TestHashUsesTheSpecParameters(), TestVerifyPassword(), TestVerifyRefusesMalformedHashes() (+1 more)

### Community 114 - "`internal/validate`"
Cohesion: 0.11
Nodes (19): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), TestServiceEntry(), Arg(), PipePath() (+11 more)

### Community 115 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 116 - ".Segment"
Cohesion: 0.40
Nodes (4): SegmentView, ZoneView, activeAt(), zoneOf()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.08
Nodes (32): watchdog, `internal/server` and `internal/server/export`, FromSnapshot(), Document, Read(), ReadSegmentation(), legacy(), snapshot() (+24 more)

### Community 122 - "TestQRCode"
Cohesion: 0.50
Nodes (4): newQR(), TestQRCode(), phoneData, qrCode

### Community 134 - "`internal/proto`"
Cohesion: 0.06
Nodes (55): Determinism and goldens, Options, `internal/proto`, TestCleanHostname(), TestIsPublic(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), Join() (+47 more)

## Knowledge Gaps
- **5 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `stream`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 328 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does ``internal/server/api`` connect ``internal/server/api`` to `storeNode`, `htmx.min.js`, `serveRuntime`, `PublishNames`, `NewHub`, `archtest.go`, `meshHarness`, `JoinRequest`, `Run`, `context.Context`, `time.Time`, `deploy_test.go`, `net/netip.Prefix`, `route`, `time.Time`?**
  _High betweenness centrality (0.046) - this node is a cross-community bridge._
- **Are the 229 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 229 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _5 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07880694177112185 - nodes in this community are weakly interconnected._
- **Why does `at()` connect `htmx.min.js` to ``internal/server/api``?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Should `storeNode` be split into smaller, more focused modules?**
  _Cohesion score 0.1484480431848853 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `.mesh`, `kernel`, `testing.T`?**
  _High betweenness centrality (0.040) - this node is a cross-community bridge._