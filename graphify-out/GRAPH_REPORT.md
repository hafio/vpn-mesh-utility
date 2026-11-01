# Graph Report - vpn-mesh-utility  (2026-11-01)

## Corpus Check
- 337 files · ~428,750 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .tmpl 16, (none) 5, .css 2)

## Summary
- 4171 nodes · 22405 edges · 155 communities (124 shown, 31 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2650 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `34ba1c66`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- ToSnapshot
- Schedule
- htmx.min.js
- AddrRange
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- Test catalogue
- deploy_test.go
- io.Writer
- github.com/hafio/vpn-mesh-utility
- serveRuntime
- newHarness
- api
- Segmentation
- Run
- Dir
- Agent (`internal/agent/...`)
- hostid.go
- net/netip.Prefix
- Net
- fakeServer
- contents_test.go
- TestSpokesAndExit
- Snapshot
- env
- Command reference
- .mesh
- DNS
- context.Context
- newRig
- Service
- agent/agent_test.go
- testing.T
- `internal/server/auth`
- testing.T
- New
- .exec
- `internal/server/store`
- Run
- netlinkNet
- NodeConfig
- planner_test.go
- Node
- PublishNames
- README.md
- boolInt
- .query
- Input
- time.Time
- LocalStatus
- `internal/proto`
- samples_test.go
- newRig
- Client
- .do
- .as
- meshDNS
- IPNet
- Sync
- .Apply
- Run
- net/netip.Addr
- Container
- UI
- serve
- Session
- Security
- fetchwintun/main_test.go
- embed.go
- rulesFor
- net/http.Request
- samples_test.go
- Containers
- Sync
- Hub
- nftKernel
- sync.Mutex
- Node
- Net
- io.Reader
- policyForm
- Open
- log/slog.Logger
- File
- Host
- exprs
- Ruleset
- .queryRow
- .readJSON
- userspace
- otherNet
- .do
- joinFlags
- meshctl/main.go
- Registration
- TestPageGoldens
- time.Duration
- Service
- nullInt
- rig
- TestControlComposePosture
- StepError
- Developer guide
- key
- ns
- fail
- net/netip.AddrPort
- Name
- copy
- service_windows.go
- Diff
- forms.js
- `internal/server` and `internal/server/export`
- Host
- syncBuffer
- loadMigrations
- Writer
- TestMain
- wgdev
- entry
- AdminError
- Slot
- Error
- NewNet
- NewNet
- .Matcher
- TestDefaultSocket
- Refusal
- logbuf.go
- Refusal
- TestOneRunningCopy
- StatusError
- AliveRequest
- AuditEntry
- Version
- Error
- nodeRow
- TestQRCode
- PeerStat
- ServerURL
- Zone
- stream
- service_other.go
- droppedLogs
- PlanError

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 304 edges
2. `ui` - 110 edges
3. ``internal/server/api`` - 101 edges
4. `ServerConfig()` - 89 edges
5. `Session` - 85 edges
6. `Slot` - 82 edges
7. `env` - 73 edges
8. `newHarness()` - 73 edges
9. `NodeConfig` - 69 edges
10. `api` - 62 edges

## Surprising Connections (you probably didn't know these)
- `Dev script tasks` --references--> `build()`  [INFERRED]
  docs/developer.md → internal/agent/policy/nft_linux.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the control plane (Dokploy)` --references--> `v()`  [INFERRED]
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

## Communities (155 total, 31 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (19): linkDNS, DNSRecord, FirewallRule, Node, TestDNS(), logsData, inspected, listed (+11 more)

### Community 1 - "WGKey"
Cohesion: 0.14
Nodes (45): blockingHost, rig, watchingFirewall, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New() (+37 more)

### Community 2 - "ToSnapshot"
Cohesion: 0.22
Nodes (19): check(), newNode(), validateSet(), ToSnapshot(), Endpoint6(), IP(), IPv6(), PeerEndpoint() (+11 more)

### Community 3 - "Schedule"
Cohesion: 0.14
Nodes (14): Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), TimeZone(), weekday(), Ref() (+6 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "AddrRange"
Cohesion: 0.09
Nodes (25): addrs(), addrText(), contains(), Rule, merged(), ofFamily(), portText(), services() (+17 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem, 2. Goals (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.14
Nodes (18): Service, checkSegmentation(), CompilePolicy(), hosts(), mergePorts(), mergeServices(), newSegResolver(), Overlaps() (+10 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "Test catalogue"
Cohesion: 0.15
Nodes (15): Rule, Violation, graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, `internal/archtest` (+7 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.22
Nodes (23): hereDoc(), runScript(), TestInstallChangesOnlyTheMarkers(), TestInstallChecksTheDataDir(), TestInstallDataDropIn(), TestInstallDataRecord(), TestInstallGolden(), TestInstallParses() (+15 more)

### Community 12 - "io.Writer"
Cohesion: 0.13
Nodes (36): TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), printExit(), setupFirewallApply(), setupFirewallShow(), setupIPv6(), printMesh() (+28 more)

### Community 14 - "serveRuntime"
Cohesion: 0.15
Nodes (13): clip(), meshFilter(), New(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestList(), TestListAnotherSlotsNetwork() (+5 more)

### Community 15 - "newHarness"
Cohesion: 0.07
Nodes (63): apiError, `internal/server/api`, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode() (+55 more)

### Community 16 - "api"
Cohesion: 0.09
Nodes (12): api, cli(), hostFamilies(), decodeJSON(), readBody(), writeJSON(), readSums(), errSession() (+4 more)

### Community 17 - "Segmentation"
Cohesion: 0.07
Nodes (30): MeshSetRequest, RecordRequest, segLiteral, SettingsRequest, segDoc, Actor, segLiterals(), plannerSegmentation() (+22 more)

### Community 18 - "Run"
Cohesion: 0.07
Nodes (40): enroll(), hostname(), meshKey(), Run(), TestIsPublic(), TestJoin(), TestOpenRuntime(), uiListen() (+32 more)

### Community 19 - "Dir"
Cohesion: 0.07
Nodes (31): `internal/fsutil`, Dir, NewKeys(), newSeed(), newWGKey(), mkdirs(), TestCachesAreCheckedWithThePinnedKey(), TestConfigCacheKeepsThePrevious() (+23 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.15
Nodes (34): fakeRunner, TestRunNeedsAnEnrollment(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+26 more)

### Community 21 - "hostid.go"
Cohesion: 0.39
Nodes (7): source, clean(), first(), fromFile(), sources(), Read(), sources()

### Community 22 - "net/netip.Prefix"
Cohesion: 0.07
Nodes (31): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+23 more)

### Community 24 - "fakeServer"
Cohesion: 0.11
Nodes (21): fakeServer, signed(), newClient(), sign(), TestAnswersMustBeSigned(), TestConfigHonoursTheETag(), TestNamesRoundTrip(), TestPrefsRoundTrip() (+13 more)

### Community 25 - "contents_test.go"
Cohesion: 0.57
Nodes (6): docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents()

### Community 26 - "TestSpokesAndExit"
Cohesion: 0.21
Nodes (6): Coverage, eventually(), lab, TestServerMesh(), TestSpokesAndExit(), uapiGet()

### Community 27 - "Snapshot"
Cohesion: 0.13
Nodes (20): `internal/server/versions`, LegacySegmentation(), SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation() (+12 more)

### Community 28 - "env"
Cohesion: 0.13
Nodes (47): checkMachineIDFile(), givenDir(), setupAgentJoin(), setupAgentReset(), setupAgentRun(), socketFlags(), newLogger(), parsePrefixes() (+39 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.08
Nodes (30): Agent protocol, Container names, Contents, Errors, Join, Local API, Mesh view, Prefs (+22 more)

### Community 31 - "DNS"
Cohesion: 0.17
Nodes (4): nameServer, DNS, Want, MeshName

### Community 32 - "context.Context"
Cohesion: 0.10
Nodes (8): pendingApply, runner, safeApply, kick(), oneLine(), Result, drain(), TestSafeApplyStateMachine()

### Community 33 - "newRig"
Cohesion: 0.20
Nodes (29): `internal/server/ui`, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages(), TestPhonePage(), newRig() (+21 more)

### Community 34 - "Service"
Cohesion: 0.12
Nodes (10): Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar(), exchange() (+2 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (76): rig, enroll(), enrollAs(), eventually(), freePort(), joinedTo(), newDir(), newFakeRuntime() (+68 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (150): Agent (`internal/agent/...`), TestCleanHostname(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), pfx() (+142 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (30): rig, 11. CLI reference, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants() (+22 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (52): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+44 more)

### Community 39 - "New"
Cohesion: 0.42
Nodes (3): New(), sameOrigin(), handler

### Community 40 - ".exec"
Cohesion: 0.12
Nodes (4): newUUID(), mustAffect(), classify(), changed()

### Community 41 - "`internal/server/store`"
Cohesion: 0.12
Nodes (47): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+39 more)

### Community 42 - "Run"
Cohesion: 0.09
Nodes (12): Deps, Deps, dialing, Rechecker, Add an apply step, Reset(), DNS, Engine (+4 more)

### Community 43 - "netlinkNet"
Cohesion: 0.14
Nodes (10): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), TestKernelShapes(), toRule() (+2 more)

### Community 44 - "NodeConfig"
Cohesion: 0.06
Nodes (37): IPv6Host, Determinism and goldens, Contents, `internal/server/phone`, `internal/tools/vulnjudge`, Shared fixtures -- `internal/testfix`, Test catalogue, forHost() (+29 more)

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (44): Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), Problem, Compute(), deliver(), deliverAddr() (+36 more)

### Community 46 - "Node"
Cohesion: 0.21
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 47 - "PublishNames"
Cohesion: 0.18
Nodes (12): DNSEntry, NameEntry, NameSet, compareAddr(), containerAddr(), LoadNames(), notContainerAddr(), PublishNames() (+4 more)

### Community 48 - "README.md"
Cohesion: 0.19
Nodes (7): Building, Contents, Documentation, How it fits together, meshctl, Quick start, Status

### Community 49 - "boolInt"
Cohesion: 0.17
Nodes (3): joinList(), boolInt(), Session

### Community 50 - ".query"
Cohesion: 0.18
Nodes (3): ContainerName, scanAll(), splitList()

### Community 51 - "Input"
Cohesion: 0.15
Nodes (13): Record, containers6(), Check(), checkContainers(), checkContainers6(), checkDNS(), checkNode6(), checkUsable() (+5 more)

### Community 52 - "time.Time"
Cohesion: 0.12
Nodes (14): replayCache, Config, failures, Lockout, LockoutError, Sealer, fullExpiry(), load() (+6 more)

### Community 53 - "LocalStatus"
Cohesion: 0.09
Nodes (22): statusDetail, Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest (+14 more)

### Community 54 - "`internal/proto`"
Cohesion: 0.09
Nodes (27): `internal/proto`, TestEventsDeliversConfigEvents(), JoinRequest, TestJoinRequestErrorNeverEchoesTheToken(), TestJoinRequestValidateNormalises(), TestJoinRequestValidateRejects(), TestPrefsRequestValidate(), validJoin() (+19 more)

### Community 55 - "samples_test.go"
Cohesion: 0.18
Nodes (15): block, checkComposeSample(), checkKubeSample(), checkRunSample(), checkStackSample(), dig(), docBlocks(), mountsMachineID() (+7 more)

### Community 56 - "newRig"
Cohesion: 0.36
Nodes (10): golden(), newRig(), TestChangesNeedTheUIsOriginAndCSRF(), TestDashboard(), TestForms(), TestLogsPage(), TestOnlyTheUIsOwnHostIsServed(), TestSessionsAndLinksAreBounded() (+2 more)

### Community 57 - "Client"
Cohesion: 0.14
Nodes (8): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, ValidChallenge(), DroppedName

### Community 58 - ".do"
Cohesion: 0.10
Nodes (21): MeshList, meshSpace, MeshUpdate, SoleMesh(), checkApart(), clash(), freeCIDR(), freeSlot() (+13 more)

### Community 59 - ".as"
Cohesion: 0.32
Nodes (8): ErrorBody, harness, TestDuplicateIsRefused(), TestForeignAndPinning(), TestHelloIsGranted(), TestSessionLapses(), TestStaleAnswerAsksAgain(), TestUnansweredHolderLapses()

### Community 60 - "meshDNS"
Cohesion: 0.22
Nodes (24): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+16 more)

### Community 61 - "IPNet"
Cohesion: 0.15
Nodes (5): Status, One image, two modes, lookup(), kernel, links

### Community 62 - "Sync"
Cohesion: 0.16
Nodes (31): Sysctls, NewNet(), Gateway(), Guard(), Owned(), parse(), Remove(), SlotOf() (+23 more)

### Community 63 - ".Apply"
Cohesion: 0.12
Nodes (15): applied, Firewall, policy, Firewall, Runner, atLeast(), exitPolicy(), normalize() (+7 more)

### Community 64 - "Run"
Cohesion: 0.10
Nodes (23): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+15 more)

### Community 65 - "net/netip.Addr"
Cohesion: 0.10
Nodes (18): network, hostAddrs, nameTable, zone, Apply (agent), gateway(), host(), proxy() (+10 more)

### Community 66 - "Container"
Cohesion: 0.27
Nodes (3): fakeRuntime, Container, Source

### Community 67 - "UI"
Cohesion: 0.10
Nodes (17): hostOf(), New(), oldest(), validOrNot(), ago(), bytesize(), loadPages(), prefsFlash() (+9 more)

### Community 68 - "serve"
Cohesion: 0.12
Nodes (36): brokenStream, countingDB, harness, joinBody(), postJoin(), serve(), signed(), storeNode() (+28 more)

### Community 69 - "Session"
Cohesion: 0.11
Nodes (40): Options, ticketServer, Join(), New(), testSession(), TestUnreachableAndMalformedAnswers(), newTicketServer(), sessionClient() (+32 more)

### Community 70 - "Security"
Cohesion: 0.17
Nodes (15): 8. The web UI, Each node's local UI, The mesh firewall, Changes a node asks for, Contents, DNS, Interfaces the agent did not create, Keys and tokens (+7 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.18
Nodes (15): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), main(), run() (+7 more)

### Community 72 - "embed.go"
Cohesion: 0.41
Nodes (10): InstallParams, RenderInstall(), RenderInstallPS1(), renderPS1(), RenderReinstall(), RenderReinstallPS1(), renderShell(), RenderUninstall() (+2 more)

### Community 73 - "rulesFor"
Cohesion: 0.36
Nodes (11): nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey(), rootKey() (+3 more)

### Community 74 - "net/http.Request"
Cohesion: 0.09
Nodes (18): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), listOf() (+10 more)

### Community 75 - "samples_test.go"
Cohesion: 0.22
Nodes (13): SampleParams, Sample, NodeSamples(), parseSample(), TestNodeSamples(), TestSampleImage(), TestSamplesRefuseHostileValues(), TestWindowsSamples() (+5 more)

### Community 76 - "Containers"
Cohesion: 0.12
Nodes (16): docWithSamples(), TestContainersDocSamples(), 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, 4. Docker Desktop on Windows, A client node on the host network, A host-mode node's containers (+8 more)

### Community 77 - "Sync"
Cohesion: 0.07
Nodes (16): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Policy, Route, Rule (+8 more)

### Community 78 - "Hub"
Cohesion: 0.14
Nodes (9): Hub, subscription, ConfigEvent, NewHub(), closed(), TestHubDeliversChallenges(), TestHubLatestWins(), TestHubNudgesEveryStream() (+1 more)

### Community 79 - "nftKernel"
Cohesion: 0.29
Nodes (4): build(), conn(), nftBatch, nftKernel

### Community 80 - "sync.Mutex"
Cohesion: 0.16
Nodes (7): syncBuffer, linkDomain, resolved, Sysctl, classify(), foreignNetns(), syncBuffer

### Community 81 - "Node"
Cohesion: 0.07
Nodes (49): IPv6Request, IPv6View, PhoneRequest, Role, addIPv6(), blockOf(), checkGroups(), checkName() (+41 more)

### Community 82 - "Net"
Cohesion: 0.14
Nodes (6): ensureAddr(), Net, Rule, syncDropped(), syncProxies(), syncRules()

### Community 83 - "io.Reader"
Cohesion: 0.29
Nodes (4): watchdog, Read(), ReadSegmentation(), Config

### Community 84 - "policyForm"
Cohesion: 0.10
Nodes (18): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), readZoneForm(), ruleRows(), scheduleText(), title() (+10 more)

### Community 85 - "Open"
Cohesion: 0.15
Nodes (17): emptyService(), TestFirstAdmin(), TestFirstAdminAfterTheCLI(), TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), insertAt5() (+9 more)

### Community 86 - "log/slog.Logger"
Cohesion: 0.09
Nodes (20): Change, NewUserspace(), NewUserspace(), Reconciler, Result, NewReconciler(), LoadOrCreateKey(), TestLoadOrCreateKey() (+12 more)

### Community 87 - "File"
Cohesion: 0.19
Nodes (8): agentLog(), ttyPassword(), OpenFile(), lockFile(), unlockFile(), lockFile(), unlockFile(), File

### Community 90 - "Host"
Cohesion: 0.08
Nodes (17): link, exitHost, Host, peer(), AddrPort(), hexKey(), ParseUAPI(), uapiInt() (+9 more)

### Community 93 - "exprs"
Cohesion: 0.24
Nodes (5): exprs(), ifname(), intervalAddrElems(), TestExprs(), nftOp

### Community 94 - "Ruleset"
Cohesion: 0.11
Nodes (9): Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel, noKernel (+1 more)

### Community 95 - ".queryRow"
Cohesion: 0.12
Nodes (3): HashToken(), Machine, Refusal

### Community 96 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.08
Nodes (12): Holder(), Taken(), TestHolder(), adapterPart(), adapterSlot(), NewUserspace(), adapterGUID(), defaultInterface() (+4 more)

### Community 100 - ".do"
Cohesion: 0.07
Nodes (32): CopyInfo, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, ReachRequest, SegmentView (+24 more)

### Community 101 - "joinFlags"
Cohesion: 0.23
Nodes (17): dataHelp(), dirMesh(), installedDir(), joinedAt(), joinFlags(), machineIDFlag(), meshDir(), meshFlags() (+9 more)

### Community 102 - "meshctl/main.go"
Cohesion: 0.10
Nodes (34): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+26 more)

### Community 103 - "Registration"
Cohesion: 0.20
Nodes (10): Domain, fakeRegistrar, nrpt, Registration, parseServers(), serverList(), TestNRPTServers(), deleteRule() (+2 more)

### Community 104 - "TestPageGoldens"
Cohesion: 0.18
Nodes (12): Add a web UI page, The web UI's pages (internal/server/ui), tabsIn(), dnsState(), newRenderer(), golden(), TestEveryPageRenders(), TestFormsShowWhatApplies() (+4 more)

### Community 105 - "time.Duration"
Cohesion: 0.12
Nodes (11): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, TestJudgeHubs(), RateLimiter (+3 more)

### Community 106 - "Service"
Cohesion: 0.08
Nodes (27): Access, DB, Params, Pending, newPassword(), passwdUser(), resetTOTP(), dial() (+19 more)

### Community 107 - "nullInt"
Cohesion: 0.18
Nodes (4): Add a database migration, DNSRecord, nullInt(), nullStr()

### Community 108 - "rig"
Cohesion: 0.27
Nodes (7): code(), cookieOf(), newEmptyRig(), TestFirstAdminFromTheWeb(), TestSignInFlow(), faultyDB, rig

### Community 109 - "TestControlComposePosture"
Cohesion: 0.27
Nodes (12): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+4 more)

### Community 111 - "Developer guide"
Cohesion: 0.22
Nodes (9): Add a command, Add a validator, Contents, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+1 more)

### Community 112 - "key"
Cohesion: 0.20
Nodes (16): intervalElems(), getAnswer(), hexOf(), pfxs(), TestParseUAPI(), TestParseUAPIRejectsMalformedValues(), uapiChange(), Desired() (+8 more)

### Community 113 - "ns"
Cohesion: 0.11
Nodes (17): Several meshes on one host, End to end -- `test/e2e`, host, lab, TestContainerNetworks(), lab, TestDualStack(), inNS() (+9 more)

### Community 114 - "fail"
Cohesion: 0.09
Nodes (24): Security invariants, `internal/validate`, TestTimestamp(), TestTimeZone(), Timestamp(), Arg(), Key32(), mask() (+16 more)

### Community 115 - "net/netip.AddrPort"
Cohesion: 0.18
Nodes (10): bindBoth(), bindPair(), Listen(), isPipe(), Listen(), listenSocket(), Listen(), ParseAddr() (+2 more)

### Community 116 - "Name"
Cohesion: 0.23
Nodes (8): validateMeshNames(), Enum(), Name(), TestEnum(), TestName(), TestText(), Text(), Version()

### Community 117 - "copy"
Cohesion: 0.70
Nodes (4): add(), copy(), done(), select()

### Community 119 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.11
Nodes (28): `internal/server` and `internal/server/export`, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+20 more)

### Community 122 - "Host"
Cohesion: 0.31
Nodes (3): Host, SyncAddrs(), SyncLink()

### Community 124 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 125 - "Writer"
Cohesion: 0.29
Nodes (3): New(), New(), Writer

### Community 130 - "Slot"
Cohesion: 0.13
Nodes (11): Pipe(), Priorities(), TestSlotOf(), NewKernel(), NewKernel(), SysctlKeys(), Slot, InterfaceSlot() (+3 more)

### Community 134 - ".Matcher"
Cohesion: 0.25
Nodes (5): Matrix, MatrixCell, PolicyRule, Schedule, SegPolicy

### Community 135 - "TestDefaultSocket"
Cohesion: 0.31
Nodes (9): TestDefaultSocket(), listenPipe(), pipeName(), servePipe(), TestPipeIsForAdministratorsOnly(), TestStatusOverAPipe(), waitPipe(), PipePath() (+1 more)

### Community 136 - "Refusal"
Cohesion: 0.19
Nodes (3): Refusal, NewKernel(), NewKernel()

### Community 139 - "TestOneRunningCopy"
Cohesion: 0.29
Nodes (5): deleteNamespaces(), tail(), lab, staleTicket(), TestOneRunningCopy()

### Community 146 - "nodeRow"
Cohesion: 0.33
Nodes (5): dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 147 - "TestQRCode"
Cohesion: 0.50
Nodes (4): newQR(), TestQRCode(), phoneData, qrCode

### Community 149 - "ServerURL"
Cohesion: 0.25
Nodes (6): HelloRequest, Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase(), TestServerURL()

### Community 150 - "Zone"
Cohesion: 0.33
Nodes (6): DNSName(), Domains(), TestDNSName(), TestDomains(), TestZone(), Zone()

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 396 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **31 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `AddrRange`, `TestDefaultSocket`, `Test catalogue`, `serveRuntime`, `Run`, `Dir`, `Agent (`internal/agent/...`)`, `fakeServer`, `.mesh`, `DNS`, `context.Context`, `agent/agent_test.go`, `Run`, `NodeConfig`, `Node`, ``internal/proto``, `newRig`, `meshDNS`, `Sync`, `Run`, `Session`, `Security`, `rulesFor`, `Host`, `exprs`, `userspace`, `Registration`, `key`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09754221411645017 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `IPNet`, `.mesh`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.14204081632653062 - nodes in this community are weakly interconnected._
- **Why does `ui` connect `net/http.Request` to `ui/ui_test.go`, `UI`, `.do`, `Security`, `New`, `Service`, `nodeRow`, `.do`, `.mesh`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._