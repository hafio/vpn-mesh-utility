# Graph Report - vpn-mesh-utility  (2026-10-22)

## Corpus Check
- 265 files · ~278,180 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3194 nodes · 15292 edges · 114 communities (99 shown, 15 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 1788 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `a0bc05b6`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- newRig
- .ExportOnce
- entry
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- api/api_test.go
- serve
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
- context.Context
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
- Registration
- Role
- .Apply
- .Segment
- Runtime
- sync.Mutex
- log/slog.Logger
- Tx
- Sync
- service_test.go
- wgdev
- .apply
- Compose and stack files
- UI
- loadMigrations
- Hub
- meshHarness
- otherPolicy
- Name
- StepError
- Open
- Installer
- .renderVersion
- container.go
- .Reach
- 7. Agent
- .query
- errorf
- RenderNASCompose
- IPNet
- .fail
- deploy_test.go
- Host
- fetchwintun/main_test.go
- install
- JoinRequest
- Join
- PublishNames
- Diff
- .queryRow
- userspace
- File
- serveRuntime
- Run
- NewKernel
- net/netip.Addr
- newQR
- commands.go
- validate_test.go
- NewHub
- `internal/server` and `internal/server/export`
- AdminError
- Error
- service_windows.go
- NewNet
- `internal/proto`
- Refusal
- StatusError
- validate.go

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 211 edges
2. `ui` - 87 edges
3. `Session` - 72 edges
4. ``internal/server/api`` - 68 edges
5. `ServerConfig()` - 61 edges
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

## Communities (114 total, 15 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (24): hostService(), linkDNS, DNSRecord, FirewallRule, Node, NewNet(), RunWindowsService(), NewKernel() (+16 more)

### Community 1 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 2 - ".ExportOnce"
Cohesion: 0.29
Nodes (5): DB, NextExport(), TestExporterRunStopsWithItsContext(), TestNextExport(), Exporter

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.20
Nodes (35): joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall(), socketFlags(), defaultAgentData(), newLogger() (+27 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (28): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 17. Amendments (+20 more)

### Community 7 - "AddrEntry"
Cohesion: 0.13
Nodes (22): Service, checkSegmentation(), host(), mergeServices(), Overlaps(), parseSpans(), spanOf(), Summarize() (+14 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.18
Nodes (14): Rule, Violation, Packages, and which way they may import, Input and commands, `internal/archtest`, `internal/tools/vulnjudge`, Test catalogue, Check() (+6 more)

### Community 11 - "api/api_test.go"
Cohesion: 0.22
Nodes (22): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+14 more)

### Community 12 - "serve"
Cohesion: 0.18
Nodes (13): brokenStream, ErrorBody, serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallScriptFailsWithoutChecksums() (+5 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.13
Nodes (30): apiError, `internal/server/api`, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminCompose(), TestAdminContainerMode() (+22 more)

### Community 15 - "Agent (`internal/agent/...`)"
Cohesion: 0.04
Nodes (74): Agent (`internal/agent/...`), TestCleanHostname(), TestDefaultSocket(), TestIsPublic(), TestJoinValidatesBeforeCallingTheServer(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig() (+66 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.16
Nodes (10): api, decode(), reply(), ETag(), cli(), decodeJSON(), readBody(), writeJSON() (+2 more)

### Community 17 - "context.Context"
Cohesion: 0.11
Nodes (19): NetworkInfo, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, heard(), nodeState() (+11 more)

### Community 18 - "action"
Cohesion: 0.17
Nodes (18): TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), cleared(), joinNonEmpty(), nodeName(), printChange(), printEnrollment() (+10 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (25): `internal/fsutil`, TestJoin(), Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious() (+17 more)

### Community 20 - "ServerConfig"
Cohesion: 0.24
Nodes (24): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestApplyIsIdempotent() (+16 more)

### Community 21 - "WGKey"
Cohesion: 0.06
Nodes (58): Keys and tokens, TestRunNeedsAnEnrollment(), NewHost(), NewRoutes(), peer(), TestLinksAddressesAndFailures(), TestRoutesSysctlAndFirewall(), IsStepError() (+50 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.11
Nodes (13): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+5 more)

### Community 23 - "meshctl/main.go"
Cohesion: 0.20
Nodes (15): agentCommands(), commands(), runWith(), firewallCommands(), healthCommands(), find(), isHelp(), main() (+7 more)

### Community 24 - ".nodeView"
Cohesion: 0.18
Nodes (8): Health(), ago(), writeHTML(), dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 25 - "LocalStatus"
Cohesion: 0.09
Nodes (23): statusDetail, printExit(), Client, Logs, Handler(), logs(), refusal(), laptop() (+15 more)

### Community 26 - "lab"
Cohesion: 0.17
Nodes (10): Coverage, End to end -- `test/e2e`, host, stream, eventually(), lab, inNS(), TestServerMesh() (+2 more)

### Community 27 - "Snapshot"
Cohesion: 0.12
Nodes (20): `internal/server/versions`, LegacySegmentation(), SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation() (+12 more)

### Community 28 - "Session"
Cohesion: 0.11
Nodes (10): Session, listOf(), meshChanged(), title(), atoi(), changed(), message(), actor() (+2 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (33): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+25 more)

### Community 30 - "time.Duration"
Cohesion: 0.13
Nodes (11): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, TestJudgeHubs(), RateLimiter (+3 more)

### Community 31 - "Host"
Cohesion: 0.09
Nodes (16): link, exitHost, Host, AddrPort(), hexKey(), ParseUAPI(), uapiInt(), compareKeys() (+8 more)

### Community 32 - "context.Context"
Cohesion: 0.07
Nodes (19): Deps, fakeService, runner, Deps, Rechecker, Add an apply step, kick(), oneLine() (+11 more)

### Community 33 - "newRig"
Cohesion: 0.08
Nodes (49): Add a command, Add a database migration, Add a validator, Add a web UI page, Dev script tasks, Developer guide, Extending the code, Goldens (+41 more)

### Community 34 - "ns"
Cohesion: 0.22
Nodes (8): lab, TestContainerNetworks(), deleteNamespaces(), newLab(), ns(), tail(), lab, TestRelayFailover()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (57): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+49 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (34): Params, rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword() (+26 more)

### Community 38 - "runArgs"
Cohesion: 0.07
Nodes (71): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+63 more)

### Community 39 - "Policy"
Cohesion: 0.09
Nodes (13): Family, netlinkPolicy, Policy, mutate(), Policy, Route, Rule, Sync() (+5 more)

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (6): mustAffect(), joinList(), boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.19
Nodes (35): `internal/server/store`, DefaultPolicy(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSetNodeGroups(), TestSetSegmentationRefusesAStranger() (+27 more)

### Community 42 - ".mesh"
Cohesion: 0.09
Nodes (24): Agent protocol, Apply (agent), Container names, Errors, Join, Local API, Mesh view, Prefs (+16 more)

### Community 43 - "netlinkNet"
Cohesion: 0.16
Nodes (7): netlinkNet, Owned(), blackhole(), dump(), prefixOf(), TestKernelShapes(), toRule()

### Community 44 - "NodeConfig"
Cohesion: 0.15
Nodes (14): Config and events, 4. Roles and topology, NodeConfig, Peer, Record, withDomains(), Hub(), ContainerNetwork() (+6 more)

### Community 45 - "planner_test.go"
Cohesion: 0.10
Nodes (49): PlanError, Add a topology rule, `internal/server/planner`, Check(), checkContainers(), checkDNS(), Problem, sortedByID() (+41 more)

### Community 46 - "meshDNS"
Cohesion: 0.19
Nodes (20): desired(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters() (+12 more)

### Community 47 - ".show"
Cohesion: 0.11
Nodes (8): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), sameOrigin(), handler, page, renderer

### Community 48 - "README.md"
Cohesion: 0.20
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "CLAUDE.md"
Cohesion: 0.29
Nodes (6): Determinism and goldens, graphify, Security invariants, Style, What this repo is, You do not run the gates

### Community 50 - "Security"
Cohesion: 0.17
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Security (+6 more)

### Community 52 - "Service"
Cohesion: 0.08
Nodes (28): replayCache, Config, DB, failures, Lockout, LockoutError, Pending, Sealer (+20 more)

### Community 53 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 54 - "Registration"
Cohesion: 0.11
Nodes (19): Domain, fakeRegistrar, linkDomain, nrpt, nrptRule, Registration, resolved, domainName() (+11 more)

### Community 55 - "Role"
Cohesion: 0.18
Nodes (15): JoinResponse, Role, Allocate(), AllocateContainers(), Block(), CheckMeshCIDR(), RangeFor(), TestAllocateContainers() (+7 more)

### Community 56 - ".Apply"
Cohesion: 0.20
Nodes (11): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+3 more)

### Community 57 - ".Segment"
Cohesion: 0.18
Nodes (6): Matrix, MatrixCell, SegmentView, ZoneView, Policy, PolicyRule

### Community 58 - "Runtime"
Cohesion: 0.18
Nodes (5): fakeRuntime, Container, clip(), meshFilter(), Runtime

### Community 59 - "sync.Mutex"
Cohesion: 0.08
Nodes (12): nameServer, syncBuffer, rig, Shared fixtures -- `internal/testfix`, DNS, Sysctl, Want, newNameTable() (+4 more)

### Community 60 - "log/slog.Logger"
Cohesion: 0.07
Nodes (21): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), bindBoth() (+13 more)

### Community 61 - "Tx"
Cohesion: 0.10
Nodes (27): Change, ComposeFile, ComposeOptions, PhoneRequest, checkName(), containerSubnet(), defaultMTU(), Nodes (+19 more)

### Community 62 - "Sync"
Cohesion: 0.16
Nodes (12): network, Sysctls, ensureAddr(), Gateway(), Net, Rule, Guard(), host() (+4 more)

### Community 63 - "service_test.go"
Cohesion: 0.14
Nodes (12): applied, Firewall, policy, Firewall, atLeast(), normalize(), richRule(), RuleArgv() (+4 more)

### Community 65 - ".apply"
Cohesion: 0.38
Nodes (4): pendingApply, safeApply, 12. Troubleshooting, TestSafeApplyStateMachine()

### Community 66 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 67 - "UI"
Cohesion: 0.12
Nodes (15): hostOf(), New(), oldest(), validOrNot(), ago(), bytesize(), loadPages(), prefsFlash() (+7 more)

### Community 68 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 69 - "Hub"
Cohesion: 0.21
Nodes (5): Hub, subscription, ConfigEvent, Reconciler, NewReconciler()

### Community 70 - "meshHarness"
Cohesion: 0.18
Nodes (15): ErrorMessage(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness, officeFirewall(), TestEditSegment() (+7 more)

### Community 72 - "Name"
Cohesion: 0.18
Nodes (16): check(), validateMeshNames(), checkGroups(), newNode(), validateSet(), ToSnapshot(), Endpoint(), Enum() (+8 more)

### Community 74 - "Open"
Cohesion: 0.29
Nodes (7): TestExportOnceSkipsUnchangedAndKeepsFourteen(), openAt(), TestMigrationTurnsRulesIntoPolicies(), Store, Open(), TestOpenRefusesNewerSchema(), NewClock()

### Community 75 - "Installer"
Cohesion: 0.28
Nodes (5): SystemdUnit(), Runner, exists(), Systemd(), Installer

### Community 77 - "container.go"
Cohesion: 0.16
Nodes (22): composeNetwork, containerHeader(), hostSocket(), MeshNetworkSetup(), networkCommands(), renderContainer(), RenderContainerCompose(), resolvers() (+14 more)

### Community 78 - ".Reach"
Cohesion: 0.29
Nodes (6): ReachRequest, flowEnd(), ReachResult, matrixCell, matrixRow, reachView

### Community 79 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 80 - ".query"
Cohesion: 0.11
Nodes (12): compareAddr(), containerAddr(), PublishNames(), TestContainerAddr(), TestPublishNames(), ContainerName, DNSRecord, scanAll() (+4 more)

### Community 81 - "errorf"
Cohesion: 0.11
Nodes (23): RecordRequest, segLiteral, segDoc, Actor, segLiterals(), errorf(), invalid(), plannerSegmentation() (+15 more)

### Community 82 - "RenderNASCompose"
Cohesion: 0.47
Nodes (6): comment(), nasHeader(), renderNAS(), RenderNASCompose(), NASParams, POSIXPath()

### Community 83 - "IPNet"
Cohesion: 0.18
Nodes (5): Status, IPNet(), lookup(), kernel, links

### Community 84 - ".fail"
Cohesion: 0.12
Nodes (20): SegObject, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm(), ruleRows() (+12 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.13
Nodes (35): containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposePosture(), TestContainerComposeRefusesHostileValues(), TestContainerRendererHoldsItsStructure(), checkPosture(), keys() (+27 more)

### Community 86 - "Host"
Cohesion: 0.36
Nodes (3): Host, SyncAddr(), SyncLink()

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.20
Nodes (15): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), main(), run() (+7 more)

### Community 88 - "install"
Cohesion: 0.33
Nodes (5): NewUserspace(), Ensure(), install(), TestEnsureWithoutABundle(), TestInstall()

### Community 89 - "JoinRequest"
Cohesion: 0.67
Nodes (3): JoinRequest, pickEndpoint(), TestPickEndpoint()

### Community 92 - "Join"
Cohesion: 0.31
Nodes (6): installAgent(), installAgent(), cleanHostname(), JoinOptions, Service, Join()

### Community 93 - "PublishNames"
Cohesion: 0.18
Nodes (10): DNSEntry, NameEntry, NameSet, LoadNames(), DNSView, SplitDomains(), Tx, TestSplitDomains() (+2 more)

### Community 95 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 96 - ".queryRow"
Cohesion: 0.14
Nodes (3): HashToken(), NewToken(), Version

### Community 98 - "userspace"
Cohesion: 0.12
Nodes (6): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace

### Community 99 - "File"
Cohesion: 0.31
Nodes (4): agentLog(), ttyPassword(), OpenFile(), File

### Community 101 - "serveRuntime"
Cohesion: 0.21
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 106 - "Run"
Cohesion: 0.11
Nodes (24): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+16 more)

### Community 110 - "net/netip.Addr"
Cohesion: 0.10
Nodes (17): nameTable, zone, parseServers(), serverList(), TestNRPTServers(), isPublic(), PublicIPv4(), inRanges() (+9 more)

### Community 111 - "newQR"
Cohesion: 0.67
Nodes (3): newQR(), phoneData, qrCode

### Community 113 - "commands.go"
Cohesion: 0.24
Nodes (11): argSynopsis(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags(), shownDefault(), cell() (+3 more)

### Community 114 - "validate_test.go"
Cohesion: 0.15
Nodes (24): `internal/validate`, Arg(), FuzzArg(), FuzzName(), TestArg(), TestDNSName(), TestDomains(), TestEndpoint() (+16 more)

### Community 118 - "NewHub"
Cohesion: 0.50
Nodes (5): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.09
Nodes (31): watchdog, setupImport(), `internal/server` and `internal/server/export`, FromSnapshot(), Document, Read(), ReadSegmentation(), legacy() (+23 more)

### Community 129 - "service_windows.go"
Cohesion: 0.12
Nodes (8): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), Windows(), handler, SCM

### Community 134 - "`internal/proto`"
Cohesion: 0.05
Nodes (63): Options, `internal/proto`, Join(), New(), newClient(), TestConfigHonoursTheETag(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents() (+55 more)

### Community 137 - "validate.go"
Cohesion: 0.19
Nodes (16): DNSName(), Domains(), fail(), Hostname(), isLoopbackHost(), LocalToken(), mask(), PipePath() (+8 more)

## Knowledge Gaps
- **4 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 313 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **15 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `on()` connect `htmx.min.js` to `.mesh`, `IPNet`, `runArgs`?**
  _High betweenness centrality (0.068) - this node is a cross-community bridge._
- **Are the 210 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 210 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _4 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07697689527077677 - nodes in this community are weakly interconnected._
- **Why does ``cmd/meshctl`` connect `runArgs` to `htmx.min.js`, `archtest.go`, `commands.go`, `action`, `log/slog.Logger`?**
  _High betweenness centrality (0.056) - this node is a cross-community bridge._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.08159073935772965 - nodes in this community are weakly interconnected._
- **Why does `Command reference` connect `Command reference` to `README.md`, `commands.go`, `Security`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._