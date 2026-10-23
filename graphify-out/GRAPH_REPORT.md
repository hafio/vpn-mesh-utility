# Graph Report - vpn-mesh-utility  (2026-10-23)

## Corpus Check
- 286 files · ~334,723 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 5, .css 2, .tmpl 1)

## Summary
- 3546 nodes · 17267 edges · 128 communities (104 shown, 24 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2039 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d735612b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- NodeConfig
- Node
- Service
- htmx.min.js
- env
- meshctl — Design Spec
- AddrEntry
- dev.sh
- dev.ps1
- archtest.go
- net.Listener
- AddrRange
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
- AdminError
- lab
- Snapshot
- .nodeView
- Command reference
- time.Time
- Net
- runner
- newRig
- Service
- agent/agent_test.go
- Installer
- `internal/server/auth`
- testing.T
- Policy
- .exec
- openTest
- kernel
- netlinkNet
- IdentityKey
- planner_test.go
- Operations
- Registration
- README.md
- Host
- Input
- otherNet
- .Tx
- context.Context
- Compose and stack files
- Tx
- cli
- Client
- IPv4
- net/netip.Addr
- meshDNS
- Join
- Sync
- service_test.go
- NodeInfo
- .query
- .Apply
- UI
- Registration
- log/slog.Logger
- Add a web UI page
- Container
- ToSnapshot
- StepError
- CookieValue
- newClient
- Name
- MeshNetworkSetup
- newRig
- .Validate
- resolved
- errorf
- Service
- Agent protocol
- .renderPolicies
- deploy_test.go
- stop
- fetchwintun/main_test.go
- Security
- sync.Mutex
- Ruleset
- service_windows.go
- service_windows_test.go
- Error
- Runtime
- Table
- .BeginEnrollment
- userspace
- Test catalogue
- Firewall
- serveRuntime
- policy
- .queryRow
- Schedule
- Policy
- exprs
- PolicyRule
- NodeConfig
- execrun_test.go
- DNS
- `internal/validate`
- NewKernel
- ns
- `internal/server` and `internal/server/export`
- TestMain
- entry
- Segmentation
- Error
- NewNet
- NewNet
- NewKernel
- NewKernel
- `internal/proto`
- NewKernel
- Refusal
- StatusError
- stream

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 246 edges
2. `ui` - 90 edges
3. ``internal/server/api`` - 79 edges
4. `Session` - 74 edges
5. `ServerConfig()` - 68 edges
6. `NodeConfig` - 67 edges
7. `newHarness()` - 57 edges
8. `env` - 48 edges
9. `runner` - 46 edges
10. `api` - 42 edges

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

## Communities (128 total, 24 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.08
Nodes (24): composeNetwork, containerFile, containerService, nasFile, nasHealth, nasService, linkDNS, DNSRecord (+16 more)

### Community 1 - "NodeConfig"
Cohesion: 0.16
Nodes (34): rig, NewHost(), NewRoutes(), IsStepError(), New(), isStep(), meshDNS(), newRig() (+26 more)

### Community 2 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 3 - "Service"
Cohesion: 0.15
Nodes (17): Config, DB, Params, Sealer, addUser(), newPassword(), passwdUser(), resetTOTP() (+9 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "env"
Cohesion: 0.05
Nodes (95): watchdog, agentCommands(), agentLog(), joinFlags(), setupAgentInstall(), setupAgentJoin(), setupAgentRun(), setupAgentUninstall() (+87 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (33): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem, 2. Goals (+25 more)

### Community 7 - "AddrEntry"
Cohesion: 0.16
Nodes (14): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), newSegResolver(), Overlaps(), parseSpans() (+6 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.22
Nodes (11): Rule, Violation, Input and commands, `internal/archtest`, Check(), importsOf(), ModuleRoot(), TestCheckerCatchesViolations() (+3 more)

### Community 11 - "net.Listener"
Cohesion: 0.47
Nodes (5): Listen(), isPipe(), Listen(), listenSocket(), Listen()

### Community 12 - "AddrRange"
Cohesion: 0.14
Nodes (17): addrs(), addrText(), contains(), Rule, merged(), ofFamily(), portText(), services() (+9 more)

### Community 14 - "`internal/server/api`"
Cohesion: 0.07
Nodes (70): apiError, countingDB, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput() (+62 more)

### Community 15 - "Run"
Cohesion: 0.13
Nodes (19): Deps, NewNodes(), New(), NewAdmin(), newAPI(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries() (+11 more)

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.14
Nodes (11): api, Routes, decode(), reply(), cli(), hostFamilies(), decodeJSON(), readBody() (+3 more)

### Community 17 - "context.Context"
Cohesion: 0.07
Nodes (24): fakeService, IPv6Request, IPv6View, NetworkInfo, NodeAddRequest, NodeSetRequest, ReachRequest, dial() (+16 more)

### Community 18 - "PublishNames"
Cohesion: 0.15
Nodes (11): DNSEntry, NameEntry, NameSet, compareAddr(), PublishNames(), TestPublishNames(), DNSView, ContainerName (+3 more)

### Community 19 - "Dir"
Cohesion: 0.10
Nodes (20): Dir, Enrollment, Open(), mkdirs(), open(), TestConfigCacheKeepsThePrevious(), TestEnrollmentRoundTrip(), TestKeysAreCreatedOnceAndKept() (+12 more)

### Community 20 - "ServerConfig"
Cohesion: 0.24
Nodes (25): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), TestAcceptAllRule() (+17 more)

### Community 21 - "Agent (`internal/agent/...`)"
Cohesion: 0.07
Nodes (45): Agent (`internal/agent/...`), TestCleanHostname(), TestDefaultSocket(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine(), TestReportKeepsToTheServersLimits() (+37 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.10
Nodes (25): Routes, index(), route(), compare(), luidOf(), onLink(), TestULAPrefix(), ULAPrefix() (+17 more)

### Community 24 - "Session"
Cohesion: 0.10
Nodes (12): Session, listOf(), meshChanged(), title(), atoi(), before(), changed(), versionNumber() (+4 more)

### Community 26 - "lab"
Cohesion: 0.20
Nodes (8): End to end -- `test/e2e`, host, eventually(), lab, inNS(), TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.10
Nodes (23): `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), fields(), fingerprint() (+15 more)

### Community 28 - ".nodeView"
Cohesion: 0.13
Nodes (11): Health(), group(), ago(), bytesize(), writeHTML(), TestHelpers(), dashboardData, nodeForm (+3 more)

### Community 29 - "Command reference"
Cohesion: 0.06
Nodes (34): Command reference, meshctl agent install, meshctl agent join, meshctl agent run, meshctl agent uninstall, meshctl dns, meshctl exit, meshctl export (+26 more)

### Community 30 - "time.Time"
Cohesion: 0.07
Nodes (23): pendingApply, safeApply, Failover, hubMove, hubState, replayCache, failures, hitWindow (+15 more)

### Community 31 - "Net"
Cohesion: 0.16
Nodes (6): ensureAddr(), Net, Rule, syncDropped(), syncProxies(), syncRules()

### Community 32 - "runner"
Cohesion: 0.11
Nodes (9): nameServer, runner, kick(), oneLine(), Result, drain(), ETag(), TestETagQuotesTheHash() (+1 more)

### Community 33 - "newRig"
Cohesion: 0.14
Nodes (35): brokenStream, The web UI's pages (internal/server/ui), `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), newQR(), TestPhonePage(), TestQRCode() (+27 more)

### Community 34 - "Service"
Cohesion: 0.15
Nodes (8): resolver, Service, bindBoth(), bindPair(), exchange(), serve(), ParseAddr(), TestParseAddr()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.10
Nodes (58): fakeServer, rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir() (+50 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (29): rig, 11. CLI reference, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestChangePassword() (+21 more)

### Community 38 - "testing.T"
Cohesion: 0.08
Nodes (64): TestAgentRunRuntimeSocket(), passwords(), shortDir(), startAdmin(), TestAgentInstallIsForWindows(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentRunNeedsAJoin() (+56 more)

### Community 39 - "Policy"
Cohesion: 0.06
Nodes (26): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Desired(), Policy, Route (+18 more)

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (6): mustAffect(), joinList(), boolInt(), classify(), nullInt(), nullStr()

### Community 41 - "openTest"
Cohesion: 0.12
Nodes (45): `internal/server/store`, DefaultPolicy(), LegacySegmentation(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), TestRestoreOfASnapshotWithFirewallRules(), testSegmentation() (+37 more)

### Community 42 - "kernel"
Cohesion: 0.16
Nodes (5): Status, TestCreateLinkNeedsPrivilege(), lookup(), kernel, links

### Community 43 - "netlinkNet"
Cohesion: 0.12
Nodes (13): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+5 more)

### Community 44 - "IdentityKey"
Cohesion: 0.21
Nodes (12): CanonicalRequest(), ParseSigned(), Sign(), TestCheckClockNeedsNoKey(), TestParseSignedRejectsMalformedHeaders(), TestSignIsUniquePerRequest(), TestSignVerifyRoundTrip(), TestVerifyRejectsMissingKey() (+4 more)

### Community 45 - "planner_test.go"
Cohesion: 0.15
Nodes (38): Add a topology rule, `internal/server/planner`, deliver(), deliverAddr(), encode(), firstDiff(), hasProblem(), lookup() (+30 more)

### Community 46 - "Operations"
Cohesion: 0.14
Nodes (14): 10. Backups, 11. When the server is down, 12. Troubleshooting, 13. Upgrading, 1. Run the server (Dokploy), 2. Enroll the servers, 3. Swarm over the mesh, 4. Spokes and exits (+6 more)

### Community 47 - "Registration"
Cohesion: 0.20
Nodes (8): Domain, fakeRegistrar, Registrar, Registration, New(), NewRegistrar(), NewRegistrar(), NewRegistrar()

### Community 48 - "README.md"
Cohesion: 0.13
Nodes (10): graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, Building, Documentation, How it fits together (+2 more)

### Community 49 - "Host"
Cohesion: 0.05
Nodes (40): link, exitHost, Agent service (systemd), Host, peer(), TestLinksAddressesAndFailures(), hexKey(), ParseUAPI() (+32 more)

### Community 50 - "Input"
Cohesion: 0.10
Nodes (23): PlanError, Record, activeAt(), zoneOf(), containers6(), Check(), checkContainers(), checkContainers6() (+15 more)

### Community 52 - ".Tx"
Cohesion: 0.20
Nodes (9): Lockout, Pending, fullExpiry(), load(), session(), hashCookie(), newSecret(), matchStep() (+1 more)

### Community 53 - "context.Context"
Cohesion: 0.10
Nodes (18): Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest, PrefsResponse (+10 more)

### Community 54 - "Compose and stack files"
Cohesion: 0.20
Nodes (10): A host-mode node's containers, Client: host network (NAS), Compose and stack files, Container mode, Container mode: a client on Docker, Container mode: a Docker Swarm node, Container mode: a server on Docker, Container mode: Podman (+2 more)

### Community 55 - "Tx"
Cohesion: 0.08
Nodes (41): Change, ComposeFile, ComposeOptions, PhoneRequest, Role, blockOf(), checkGroups(), checkName() (+33 more)

### Community 56 - "cli"
Cohesion: 0.18
Nodes (17): TestIPv6DNSRecords(), harness, meshHarness(), refused(), TestCustomDNSRecords(), harness, officeFirewall(), TestEditSegment() (+9 more)

### Community 57 - "Client"
Cohesion: 0.27
Nodes (6): apiError(), clip(), Client, StatusResponse, NamesResponse, DroppedName

### Community 58 - "IPv4"
Cohesion: 0.24
Nodes (10): validateMeshNames(), checkDNS(), DNSName(), Domains(), IPv4(), TestDNSName(), TestDomains(), TestIPv4() (+2 more)

### Community 59 - "net/netip.Addr"
Cohesion: 0.09
Nodes (25): network, hostAddrs, nameTable, zone, Apply (agent), TestIsPublic(), gateway(), host() (+17 more)

### Community 60 - "meshDNS"
Cohesion: 0.25
Nodes (22): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+14 more)

### Community 61 - "Join"
Cohesion: 0.22
Nodes (8): Options, Join(), New(), TestJoinIsUnsigned(), TestUnreachableAndMalformedAnswers(), WriteEvent(), Event, Reader

### Community 62 - "Sync"
Cohesion: 0.14
Nodes (26): Sysctls, Sysctl, NewNet(), Gateway(), Guard(), Owned(), parse(), Sync() (+18 more)

### Community 63 - "service_test.go"
Cohesion: 0.29
Nodes (7): applied, Firewall, normalize(), richRule(), RuleArgv(), FWRule, sortRules()

### Community 64 - "NodeInfo"
Cohesion: 0.14
Nodes (13): NodeDetail, NodeList, PhoneProfile, statusDetail, addresses(), heard(), Rollback, NodeInfo (+5 more)

### Community 65 - ".query"
Cohesion: 0.16
Nodes (6): scanAll(), splitList(), AuditEntry, PeerStat, StatusUpdate, Version

### Community 66 - ".Apply"
Cohesion: 0.16
Nodes (13): New(), New(), fakeProc(), TestApplyChangesOnlyWhatDiffers(), TestApplyContainerKeys(), TestApplyRefusesUnknownKeysBeforeTouchingAPath(), TestApplyReportsAMissingParameter(), TestApplyReportsAReadOnlyParameter() (+5 more)

### Community 67 - "UI"
Cohesion: 0.12
Nodes (15): hostOf(), New(), oldest(), TestServeAndListen(), bytesize(), loadPages(), prefsFlash(), LocalToken() (+7 more)

### Community 68 - "Registration"
Cohesion: 0.17
Nodes (15): nrpt, nrptRule, desired(), domainName(), fromRules(), nrptKey(), ownedKey(), parseServers() (+7 more)

### Community 69 - "log/slog.Logger"
Cohesion: 0.07
Nodes (20): Deps, Hub, subscription, Deps, Run(), DNS, Firewall, MeshFirewall (+12 more)

### Community 70 - "Add a web UI page"
Cohesion: 0.13
Nodes (14): Add a command, Add a database migration, Add a validator, Add a web UI page, Extending the code, TestTemplatesHaveNoScript(), newPolicyForm(), dnsState() (+6 more)

### Community 71 - "Container"
Cohesion: 0.20
Nodes (9): fakeRuntime, Container, hostAddr(), Of(), TestHostAddr(), TestOf(), TestOfCapsTheNames(), TestOfSwarmWithoutAStack() (+1 more)

### Community 72 - "ToSnapshot"
Cohesion: 0.23
Nodes (15): checkIPv6(), check(), newNode(), validateSet(), ToSnapshot(), Endpoint6(), IPv6(), plain6() (+7 more)

### Community 74 - "CookieValue"
Cohesion: 0.16
Nodes (6): TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), sameOrigin(), handler

### Community 75 - "newClient"
Cohesion: 0.27
Nodes (9): newClient(), TestErrorsCarryTheCode(), TestEventsDeliversConfigEvents(), TestEventsRefusesAnythingButAStream(), TestNamesRoundTrip(), TestPrefsRoundTrip(), TestStatusRoundTrip(), verifying() (+1 more)

### Community 76 - "Name"
Cohesion: 0.12
Nodes (13): validOrNot(), Ref(), ServiceEntry(), TestRef(), TestServiceEntry(), Name(), SemVer(), TestName() (+5 more)

### Community 77 - "MeshNetworkSetup"
Cohesion: 0.36
Nodes (8): containerHeader(), MeshNetworkSetup(), networkCommands(), resolvers(), serviceExample(), TestHostModeNetworkSetup(), NetworkParams, NetworkSetup

### Community 78 - "newRig"
Cohesion: 0.26
Nodes (11): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+3 more)

### Community 79 - ".Validate"
Cohesion: 0.25
Nodes (7): Hostname(), Key32(), mask(), TestHostnameFoldsCase(), TestKey32(), TestTokenMasksTheValue(), Token()

### Community 80 - "resolved"
Cohesion: 0.31
Nodes (5): linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors()

### Community 81 - "errorf"
Cohesion: 0.09
Nodes (30): RecordRequest, segLiteral, SegmentView, ZoneView, segDoc, addIPv6(), Actor, segLiterals() (+22 more)

### Community 82 - "Service"
Cohesion: 0.22
Nodes (7): hostService(), hostService(), uninstallAgent(), Service, Systemd(), TestAvailableNeedsElevation(), Windows()

### Community 83 - "Agent protocol"
Cohesion: 0.16
Nodes (16): Route, route, Agent protocol, Container names, Errors, Join, Local API, Mesh view (+8 more)

### Community 84 - ".renderPolicies"
Cohesion: 0.09
Nodes (22): SegPolicy, SegSchedule, anyText(), listText(), objectRows(), readPolicyForm(), readZoneForm(), ruleRows() (+14 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.08
Nodes (50): hostSocket(), renderContainer(), RenderContainerCompose(), containerParams(), parseContainer(), TestContainerComposeGolden(), TestContainerComposeHeader(), TestContainerComposePosture() (+42 more)

### Community 87 - "fetchwintun/main_test.go"
Cohesion: 0.11
Nodes (22): `internal/tools/fetchwintun`, NewUserspace(), Ensure(), Pin, install(), Sum(), TestEnsureWithoutABundle(), TestInstall() (+14 more)

### Community 88 - "Security"
Cohesion: 0.15
Nodes (15): 8. The web UI, Each node's local UI, The mesh firewall, Changes a node asks for, DNS, Interfaces the agent did not create, Keys and tokens, Security (+7 more)

### Community 89 - "sync.Mutex"
Cohesion: 0.22
Nodes (3): syncBuffer, clip(), Ring

### Community 90 - "Ruleset"
Cohesion: 0.10
Nodes (10): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel (+2 more)

### Community 93 - "service_windows_test.go"
Cohesion: 0.32
Nodes (5): commandLine(), executes(), TestCommandLineQuotesThePaths(), TestHandlerReportsAFailedRun(), TestHandlerStopsTheAgent()

### Community 95 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 96 - "Table"
Cohesion: 0.36
Nodes (5): NewTable(), Table, Sync(), NewTable(), NewTable()

### Community 97 - ".BeginEnrollment"
Cohesion: 0.43
Nodes (3): aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 98 - "userspace"
Cohesion: 0.11
Nodes (6): NewUserspace(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace

### Community 99 - "Test catalogue"
Cohesion: 0.22
Nodes (14): Coverage, `internal/server/phone`, `internal/tools/vulnjudge`, Test catalogue, joinSet(), NewKey(), Render(), aliceKey() (+6 more)

### Community 100 - "Firewall"
Cohesion: 0.40
Nodes (3): Firewall, wantedPolicies(), Firewall

### Community 101 - "serveRuntime"
Cohesion: 0.21
Nodes (12): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), TestDetect(), TestList() (+4 more)

### Community 103 - ".queryRow"
Cohesion: 0.12
Nodes (6): scanNode(), HashToken(), NewToken(), User, scanUser(), scanner

### Community 104 - "Schedule"
Cohesion: 0.12
Nodes (15): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestTimestamp(), TestWindowHolds(), Timestamp() (+7 more)

### Community 105 - "Policy"
Cohesion: 0.13
Nodes (28): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+20 more)

### Community 107 - "exprs"
Cohesion: 0.08
Nodes (18): Dev script tasks, Developer guide, Goldens, Releases, Toolchain, build(), conn(), exprs() (+10 more)

### Community 108 - "PolicyRule"
Cohesion: 0.33
Nodes (5): Matrix, MatrixCell, PolicyRule, Schedule, SegPolicy

### Community 111 - "NodeConfig"
Cohesion: 0.07
Nodes (28): dialing, IPv6Host, Rechecker, Determinism and goldens, Add an apply step, Shared fixtures -- `internal/testfix`, forHost(), Engine (+20 more)

### Community 112 - "execrun_test.go"
Cohesion: 0.09
Nodes (20): TestUnitPosture(), SystemdUnit(), capped, allowed(), Check(), fakeProgram(), TestCappedKeepsTheHead(), TestCheckAllowsFirewallArgv() (+12 more)

### Community 113 - "DNS"
Cohesion: 0.20
Nodes (5): DNS, Want, newNameTable(), DNS, MeshName

### Community 114 - "`internal/validate`"
Cohesion: 0.13
Nodes (19): Security invariants, `internal/validate`, IP(), PeerEndpoint(), TestIPv6Grammar(), TestTimeZone(), AddrEntry(), TestAddrEntry() (+11 more)

### Community 118 - "ns"
Cohesion: 0.13
Nodes (13): lab, TestContainerNetworks(), lab, TestDualStack(), deleteNamespaces(), newLab(), ns(), tail() (+5 more)

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.08
Nodes (36): `internal/server` and `internal/server/export`, NewHub(), closed(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses(), FromSnapshot(), Document (+28 more)

### Community 134 - "`internal/proto`"
Cohesion: 0.14
Nodes (20): `internal/proto`, TestJoinRequestErrorNeverEchoesTheToken(), TestJoinRequestValidateNormalises(), TestJoinRequestValidateRejects(), TestPrefsRequestValidate(), TestStatusReportValidate(), validJoin(), validStatus() (+12 more)

## Knowledge Gaps
- **6 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR`, `lab`, `lab` (+1 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 338 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does ``internal/server/api`` connect ``internal/server/api`` to `Test catalogue`, `htmx.min.js`, `serveRuntime`, ``internal/server/auth``, `Run`, `context.Context`, `PublishNames`, `Agent protocol`, `deploy_test.go`, `net/netip.Prefix`, `cli`, ``internal/server` and `internal/server/export``, `net/netip.Addr`, `time.Time`?**
  _High betweenness centrality (0.033) - this node is a cross-community bridge._
- **Are the 245 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.apply()` and `.SetFail()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 245 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `harness`, `NO_COLOR` to the rest of the system?**
  _6 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.0807705903145988 - nodes in this community are weakly interconnected._
- **Why does `Agent (`internal/agent/...`)` connect `Agent (`internal/agent/...`)` to `ui/ui_test.go`, `NodeConfig`, `Node`, `env`, `archtest.go`, `net/http.ResponseWriter`, `Dir`, `ServerConfig`, `time.Time`, `runner`, `Service`, `agent/agent_test.go`, `testing.T`, `Policy`, `kernel`, `netlinkNet`, `Host`, `net/netip.Addr`, `meshDNS`, `Join`, `Sync`, `.Apply`, `UI`, `Registration`, `log/slog.Logger`, `Add a web UI page`, `Container`, `newClient`, `newRig`, `resolved`, `Service`, `Agent protocol`, `fetchwintun/main_test.go`, `service_windows_test.go`, `Test catalogue`, `serveRuntime`, `Policy`, `exprs`, `NodeConfig`, `execrun_test.go`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.08159073935772965 - nodes in this community are weakly interconnected._
- **Why does `at()` connect `htmx.min.js` to ``internal/server/api``?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._