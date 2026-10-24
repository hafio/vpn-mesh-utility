# Graph Report - vpn-mesh-utility  (2026-10-25)

## Corpus Check
- 286 files · ~334,473 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3555 nodes · 17378 edges · 130 communities (104 shown, 26 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2051 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b23c1f39`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- Node
- storeNode
- htmx.min.js
- env
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
- services
- github.com/hafio/vpn-mesh-utility
- `internal/server/api`
- log/slog.Logger
- net/http.ResponseWriter
- NodeChange
- PublishNames
- Dir
- ServerConfig
- Agent (`internal/agent/...`)
- net/netip.Prefix
- Net
- Session
- execrun_test.go
- lab
- Snapshot
- .Phone
- Command reference
- time.Time
- Net
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
- cli
- planner_test.go
- servePipe
- .mesh
- README.md
- Host
- Open
- Run
- context.Context
- LocalStatus
- samples_test.go
- Tx
- .BeginEnrollment
- Client
- Extending the code
- Test catalogue
- meshDNS
- serve
- Sync
- .Apply
- New
- .query
- .Apply
- UI
- Registration
- Service
- User
- fetchwintun/main_test.go
- ToSnapshot
- Security
- .show
- CompilePolicy
- Name
- ServerURL
- newRig
- Matcher
- resolved
- errorf
- .Matcher
- NodeConfig
- .renderPolicies
- deploy_test.go
- service_windows.go
- Registration
- DNSName
- Installer
- Ruleset
- NewHub
- Diff
- .Reach
- loadMigrations
- .renderVersion
- Input
- userspace
- .Apply
- wgdev
- Segmentation
- .nodeView
- .queryRow
- Schedule
- `internal/proto`
- net/netip.Addr
- stream
- AdminError
- toAPIError
- StepError
- PlanError
- TestMain
- DNS
- `internal/validate`
- .relayFor
- entry
- Error
- TestContainerNetworks
- NewNet
- NewKernel
- `internal/server` and `internal/server/export`
- NewKernel
- NewKernel
- NewKernel
- Refusal
- StatusError
- Error
- NewNet
- TestMeshFirewall

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 256 edges
2. `ui` - 89 edges
3. ``internal/server/api`` - 79 edges
4. `Session` - 73 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 67 edges
7. `env` - 57 edges
8. `newHarness()` - 57 edges
9. `runner` - 47 edges
10. `Tx` - 41 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
- `Dev script tasks` --references--> `build()`  [INFERRED]
  docs/developer.md → internal/agent/policy/nft_linux.go
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

## Communities (130 total, 26 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (17): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, DNSRequest (+9 more)

### Community 1 - "NodeConfig"
Cohesion: 0.14
Nodes (38): rig, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS() (+30 more)

### Community 2 - "Node"
Cohesion: 0.19
Nodes (13): network, Config and events, 4. Roles and topology, parse(), Containers, Peer, ContainerNetwork(), endpoints() (+5 more)

### Community 3 - "storeNode"
Cohesion: 0.19
Nodes (22): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+14 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (116): Status, 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names (+108 more)

### Community 5 - "env"
Cohesion: 0.05
Nodes (111): watchdog, agentCommands(), agentLog(), joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall() (+103 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.19
Nodes (11): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), parseSpans(), PrefixRange(), portRange (+3 more)

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

### Community 12 - "services"
Cohesion: 0.32
Nodes (7): addrText(), Rule, merged(), portText(), services(), Range, service

### Community 14 - "`internal/server/api`"
Cohesion: 0.13
Nodes (32): `internal/server/api`, ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes() (+24 more)

### Community 15 - "log/slog.Logger"
Cohesion: 0.07
Nodes (19): Change, Failover, Hub, hubMove, hubState, subscription, NewUserspace(), NewUserspace() (+11 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.15
Nodes (10): api, decode(), reply(), ETag(), hostFamilies(), decodeJSON(), readBody(), writeJSON() (+2 more)

### Community 17 - "NodeChange"
Cohesion: 0.24
Nodes (7): NodeSetRequest, SegmentView, ZoneView, AdminClient, nodePath(), Enrollment, NodeChange

### Community 18 - "PublishNames"
Cohesion: 0.20
Nodes (11): DNSEntry, NameEntry, NameSet, compareAddr(), LoadNames(), PublishNames(), TestPublishNames(), DNSView (+3 more)

### Community 19 - "Dir"
Cohesion: 0.09
Nodes (23): `internal/fsutil`, Dir, Enrollment, mkdirs(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestJoinRecord(), TestKeysAreCreatedOnceAndKept() (+15 more)

### Community 20 - "ServerConfig"
Cohesion: 0.20
Nodes (26): fakeRunner, atLeast(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+18 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.04
Nodes (76): Options, Agent (`internal/agent/...`), TestCleanHostname(), TestDefaultSocket(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine() (+68 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (27): Routes, NewTable(), compare(), Table, Sync(), NewTable(), luidOf(), NewTable() (+19 more)

### Community 24 - "Session"
Cohesion: 0.13
Nodes (10): Session, listOf(), meshChanged(), title(), atoi(), changed(), message(), actor() (+2 more)

### Community 25 - "execrun_test.go"
Cohesion: 0.12
Nodes (18): capped, Exec, allowed(), Check(), fakeProgram(), TestCappedKeepsTheHead(), TestCheckAllowsFirewallArgv(), TestCheckRefuses() (+10 more)

### Community 26 - "lab"
Cohesion: 0.15
Nodes (16): Coverage, End to end -- `test/e2e`, host, lab, TestDualStack(), deleteNamespaces(), eventually(), lab (+8 more)

### Community 27 - "Snapshot"
Cohesion: 0.13
Nodes (20): `internal/server/versions`, LegacySegmentation(), SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation() (+12 more)

### Community 28 - ".Phone"
Cohesion: 0.09
Nodes (26): NodeDetail, NodeList, PhoneProfile, PhoneRequest, statusDetail, addresses(), heard(), Add a web UI page (+18 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl control, meshctl dns, meshctl exit (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.14
Nodes (8): syncBuffer, replayCache, failures, hitWindow, Lockout, LockoutError, RateLimiter, record()

### Community 31 - "Net"
Cohesion: 0.12
Nodes (10): Sysctls, Sysctl, Sysctls, ensureAddr(), Net, Rule, Sync(), syncDropped() (+2 more)

### Community 32 - "runner"
Cohesion: 0.09
Nodes (11): Deps, pendingApply, runner, safeApply, kick(), oneLine(), Run(), Result (+3 more)

### Community 33 - "newRig"
Cohesion: 0.13
Nodes (35): The web UI's pages (internal/server/ui), `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), newQR(), TestPhonePage(), TestQRCode(), code() (+27 more)

### Community 34 - "Service"
Cohesion: 0.14
Nodes (9): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+1 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.11
Nodes (60): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+52 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.14
Nodes (25): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(), TestLoginLocksOutAfterFiveFailures() (+17 more)

### Community 38 - "testing.T"
Cohesion: 0.08
Nodes (71): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+63 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.09
Nodes (7): mustAffect(), DNSRecord, joinList(), boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.15
Nodes (38): `internal/server/store`, DefaultPolicy(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSegScheduleString(), TestSetNodeGroups() (+30 more)

### Community 43 - "netlinkNet"
Cohesion: 0.10
Nodes (16): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+8 more)

### Community 44 - "cli"
Cohesion: 0.19
Nodes (19): cli(), TestIPv6DNSRecords(), TestReportedHostFamilies(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness (+11 more)

### Community 45 - "planner_test.go"
Cohesion: 0.14
Nodes (41): `internal/server/planner`, TestCheckRequested(), TestRangeFor(), Problem, Compute(), deliver(), deliverAddr(), encode() (+33 more)

### Community 46 - "servePipe"
Cohesion: 0.53
Nodes (6): listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe(), waitPipe()

### Community 47 - ".mesh"
Cohesion: 0.16
Nodes (14): 10. Backups, 11. When the control plane is down, 12. Troubleshooting, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits (+6 more)

### Community 48 - "README.md"
Cohesion: 0.23
Nodes (5): Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "Host"
Cohesion: 0.07
Nodes (32): link, exitHost, Host, peer(), AddrPort(), TestAddrPort(), hexKey(), ParseUAPI() (+24 more)

### Community 50 - "Open"
Cohesion: 0.26
Nodes (10): TestExportOnceSkipsUnchangedAndKeepsFourteen(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), Store, Open(), TestOpenIsIdempotent(), TestOpenRefusesNewerSchema() (+2 more)

### Community 51 - "Run"
Cohesion: 0.13
Nodes (20): Deps, NewNodes(), New(), NewAdmin(), newAPI(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries() (+12 more)

### Community 52 - "context.Context"
Cohesion: 0.22
Nodes (8): Pending, fullExpiry(), load(), session(), hashCookie(), newSecret(), matchStep(), Session

### Community 53 - "LocalStatus"
Cohesion: 0.07
Nodes (29): fakeService, waitApplied(), DNS, dial(), dial(), isPipe(), Client, Logs (+21 more)

### Community 54 - "samples_test.go"
Cohesion: 0.16
Nodes (15): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), sampleVolume(), shellCommands() (+7 more)

### Community 55 - "Tx"
Cohesion: 0.09
Nodes (35): IPv6Request, IPv6View, NodeAddRequest, Role, addIPv6(), blockOf(), checkGroups(), containerNetwork() (+27 more)

### Community 56 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "Extending the code"
Cohesion: 0.16
Nodes (16): Route, route, Add a command, Add a database migration, Add a validator, Add an agent API route, Dev script tasks, Developer guide (+8 more)

### Community 59 - "Test catalogue"
Cohesion: 0.17
Nodes (18): `internal/server/phone`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue, Hub(), joinSet(), NewKey(), Render() (+10 more)

### Community 60 - "meshDNS"
Cohesion: 0.28
Nodes (20): answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters(), TestServiceAnswersOverIPv6() (+12 more)

### Community 61 - "serve"
Cohesion: 0.19
Nodes (13): brokenStream, serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums() (+5 more)

### Community 62 - "Sync"
Cohesion: 0.16
Nodes (22): NewNet(), TestForHostWithoutIPv6(), Gateway(), Guard(), Owned(), TestGatewayRules(), TestGatewayRulesWithIPv6(), TestGuardRules() (+14 more)

### Community 63 - ".Apply"
Cohesion: 0.14
Nodes (12): applied, Firewall, policy, Firewall, Runner, normalize(), richRule(), RuleArgv() (+4 more)

### Community 64 - "New"
Cohesion: 0.42
Nodes (9): Config, Params, New(), DefaultParams(), HashPassword(), TestHashUsesTheSpecParameters(), TestVerifyPassword(), TestVerifyRefusesMalformedHashes() (+1 more)

### Community 65 - ".query"
Cohesion: 0.15
Nodes (6): ContainerName, scanAll(), splitList(), AuditEntry, PeerStat, StatusUpdate

### Community 66 - ".Apply"
Cohesion: 0.16
Nodes (13): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+5 more)

### Community 67 - "UI"
Cohesion: 0.11
Nodes (17): hostOf(), New(), oldest(), TestServeAndListen(), validOrNot(), ago(), bytesize(), loadPages() (+9 more)

### Community 68 - "Registration"
Cohesion: 0.17
Nodes (16): nrpt, nrptRule, domainName(), same(), TestSameRegistration(), fromRules(), nrptKey(), ownedKey() (+8 more)

### Community 69 - "Service"
Cohesion: 0.20
Nodes (10): DB, addUser(), newPassword(), passwdUser(), resetTOTP(), Service, CheckPassword(), TestCheckPasswordNeverEchoesThePassword() (+2 more)

### Community 70 - "User"
Cohesion: 0.39
Nodes (3): User, scanUser(), scanner

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.05
Nodes (38): fakeRuntime, `internal/tools/fetchwintun`, TestOpenRuntime(), clip(), Ring, Container, clip(), Detect() (+30 more)

### Community 72 - "ToSnapshot"
Cohesion: 0.18
Nodes (21): check(), ParseAddr(), TestParseAddr(), newNode(), validateSet(), ToSnapshot(), Endpoint6(), IP() (+13 more)

### Community 73 - "Security"
Cohesion: 0.18
Nodes (14): 8. The web UI, Each node's local UI, The mesh firewall, Agent service (systemd), Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens (+6 more)

### Community 74 - ".show"
Cohesion: 0.11
Nodes (9): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), group(), before(), sameOrigin(), handler (+1 more)

### Community 75 - "CompilePolicy"
Cohesion: 0.26
Nodes (11): Policy, activeAt(), zoneOf(), CompilePolicy(), newSegResolver(), Overlaps(), Summarize(), TestOverlaps() (+3 more)

### Community 76 - "Name"
Cohesion: 0.14
Nodes (13): validateMeshNames(), Ref(), ServiceEntry(), TestRef(), TestServiceEntry(), Enum(), Name(), TestEnum() (+5 more)

### Community 77 - "ServerURL"
Cohesion: 0.33
Nodes (5): Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase(), TestServerURL()

### Community 78 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 79 - "Matcher"
Cohesion: 0.25
Nodes (4): serves(), Service, Matcher, matchRule

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.12
Nodes (23): RecordRequest, segLiteral, segDoc, Actor, segLiterals(), errorf(), invalid(), PolicyZone() (+15 more)

### Community 82 - ".Matcher"
Cohesion: 0.25
Nodes (5): Matrix, MatrixCell, PolicyRule, Schedule, SegPolicy

### Community 83 - "NodeConfig"
Cohesion: 0.11
Nodes (13): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Signed requests (+5 more)

### Community 84 - ".renderPolicies"
Cohesion: 0.06
Nodes (30): SegObject, SegPolicy, SegSchedule, anyText(), listText(), newPolicyForm(), objectRows(), readPolicyForm() (+22 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.21
Nodes (22): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+14 more)

### Community 86 - "service_windows.go"
Cohesion: 0.11
Nodes (10): hostService(), uninstallAgent(), commandLine(), RunWindowsService(), stop(), TestAvailableNeedsElevation(), TestCommandLineQuotesThePaths(), Windows() (+2 more)

### Community 87 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 88 - "DNSName"
Cohesion: 0.29
Nodes (7): checkDNS(), DNSName(), Domains(), TestDNSName(), TestDomains(), TestZone(), Zone()

### Community 90 - "Ruleset"
Cohesion: 0.05
Nodes (21): watchingFirewall, Kernel, build(), conn(), exprs(), hasLegacy(), ifname(), intervalAddrElems() (+13 more)

### Community 92 - "NewHub"
Cohesion: 0.50
Nodes (5): NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 93 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 94 - ".Reach"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 95 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 97 - "Input"
Cohesion: 0.13
Nodes (14): Routes, Record, containers6(), Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable() (+6 more)

### Community 98 - "userspace"
Cohesion: 0.09
Nodes (8): lookup(), NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, links, userspace

### Community 99 - ".Apply"
Cohesion: 0.09
Nodes (14): Deps, dialing, IPv6Host, Rechecker, Add an apply step, forHost(), Engine, Firewall (+6 more)

### Community 100 - "wgdev"
Cohesion: 0.24
Nodes (3): TestUAPISetGolden(), UAPISet(), wgdev

### Community 102 - ".nodeView"
Cohesion: 0.15
Nodes (10): Health(), ago(), bytesize(), writeHTML(), TestHelpers(), dashboardData, nodeForm, nodeRow (+2 more)

### Community 103 - ".queryRow"
Cohesion: 0.13
Nodes (3): HashToken(), NewToken(), Version

### Community 104 - "Schedule"
Cohesion: 0.11
Nodes (18): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestTimestamp(), TestWindowHolds(), Timestamp() (+10 more)

### Community 105 - "`internal/proto`"
Cohesion: 0.04
Nodes (70): Determinism and goldens, `internal/proto`, TestEventsDeliversConfigEvents(), NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile() (+62 more)

### Community 106 - "net/netip.Addr"
Cohesion: 0.06
Nodes (34): hostService(), hostAddrs, nameTable, zone, Apply (agent), TestIsPublic(), gateway(), host() (+26 more)

### Community 109 - "toAPIError"
Cohesion: 0.33
Nodes (5): apiError, ErrorMessage(), toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion()

### Community 113 - "DNS"
Cohesion: 0.14
Nodes (7): nameServer, DNS, desired(), Want, newNameTable(), DNS, MeshName

### Community 114 - "`internal/validate`"
Cohesion: 0.12
Nodes (20): Security invariants, `internal/validate`, addrs(), TestTimeZone(), AddrEntry(), MergeRanges(), TestAddrEntry(), TestAddrRangePrefix() (+12 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.11
Nodes (28): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+20 more)

### Community 138 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

## Knowledge Gaps
- **6 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `lab` (+1 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 338 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **26 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `on()` connect `htmx.min.js` to `testing.T`, `.mesh`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Are the 255 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 255 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _6 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08279062804102104 - nodes in this community are weakly interconnected._
- **Why does ``internal/server/api`` connect ``internal/server/api`` to `storeNode`, `htmx.min.js`, `fetchwintun/main_test.go`, `openTest`, `net/netip.Addr`, `cli`, `toAPIError`, `log/slog.Logger`, `NodeChange`, `PublishNames`, `Run`, `Tx`, `net/netip.Prefix`, `samples_test.go`, `Extending the code`, `Test catalogue`, `NewHub`, `serve`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Should `NodeConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.14358974358974358 - nodes in this community are weakly interconnected._
- **Why does `Net` connect `Net` to `ui/ui_test.go`, `net/netip.Addr`, `net/netip.Prefix`, `Sync`, `time.Time`, `Net`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._