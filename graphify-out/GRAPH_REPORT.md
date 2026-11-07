# Graph Report - vpn-mesh-utility  (2026-11-07)

## Corpus Check
- 337 files · ~429,556 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .tmpl 16, (none) 5, .css 2)

## Summary
- 4173 nodes · 22417 edges · 151 communities (116 shown, 35 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2652 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `a6df718e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- Name
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
- Container
- newHarness
- api
- Segmentation
- Join
- Dir
- Agent (`internal/agent/...`)
- hostid.go
- net/netip.Prefix
- Net
- fakeServer
- contents_test.go
- lab
- Snapshot
- env
- Command reference
- .mesh
- cli
- context.Context
- `internal/server/ui`
- Service
- agent/agent_test.go
- testing.T
- `internal/server/auth`
- testing.T
- Policy
- .exec
- `internal/server/store`
- Run
- netlinkNet
- NodeConfig
- planner_test.go
- Node
- netlinkPolicy
- README.md
- .SetSegmentation
- .query
- Input
- serveRuntime
- LocalStatus
- policy
- samples_test.go
- serve
- Client
- .do
- Ticket
- meshDNS
- IPNet
- Sync
- .Apply
- Run
- net/netip.Addr
- CompilePolicy
- UI
- serve
- `internal/proto`
- Security
- fetchwintun/main_test.go
- embed.go
- rulesFor
- net/http.Request
- otherPolicy
- Containers
- Sync
- NewHub
- exprs
- sync.Mutex
- Node
- Net
- io.Reader
- policyForm
- Open
- .AddMesh
- File
- Host
- Agent protocol
- Ruleset
- .queryRow
- .readJSON
- userspace
- otherNet
- .do
- joinFlags
- meshctl/main.go
- TestSafeApplyStateMachine
- DefaultPolicy
- log/slog.Logger
- Service
- nullInt
- .Reach
- net/netip.AddrPort
- StepError
- TestContainerNetworks
- Key
- TestMeshFirewall
- fail
- net.Listener
- Signed
- 8. The web UI
- service_windows.go
- Diff
- forms.js
- `internal/server` and `internal/server/export`
- Exit
- NewKernel
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
- NewKernel
- Refusal
- .add
- Refusal
- usageError
- StatusError
- AliveRequest
- AuditEntry
- Version
- Error
- nodeRow
- zonesView
- PeerStat
- ServerURL
- ToSnapshot

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 305 edges
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
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the control plane (Dokploy)` --references--> `v()`  [INFERRED]
  docs/operations.md → internal/server/ui/static/htmx.min.js
- `Add a validator` --references--> `mask()`  [INFERRED]
  docs/developer.md → internal/validate/validate.go
- `Several meshes on one host` --references--> `TestTwoMeshes()`  [INFERRED]
  docs/security.md → test/e2e/meshes_test.go

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

## Communities (151 total, 35 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.14
Nodes (45): blockingHost, rig, watchingFirewall, TestProxyLoopAppliesAgain(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), IsStepError() (+37 more)

### Community 2 - "Name"
Cohesion: 0.21
Nodes (13): check(), newNode(), validateSet(), Ref(), ServiceEntry(), Enum(), IntRange(), IPv4() (+5 more)

### Community 3 - "Schedule"
Cohesion: 0.12
Nodes (17): The mesh firewall, Clock(), Window, Schedule(), TestSchedule(), TestTimestamp(), TestWindowHolds(), Timestamp() (+9 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "AddrRange"
Cohesion: 0.12
Nodes (18): addrs(), contains(), ofFamily(), inRanges(), serves(), inRange(), AddrEntry(), AddrRange (+10 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (35): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+27 more)

### Community 7 - "segResolver"
Cohesion: 0.20
Nodes (12): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), newSegResolver(), parseSpans(), ZoneAddresses() (+4 more)

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
Cohesion: 0.15
Nodes (35): checkPosture(), controlEnv(), hereDoc(), keys(), mustRead(), runScript(), sameSet(), service() (+27 more)

### Community 12 - "io.Writer"
Cohesion: 0.14
Nodes (17): TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), printExit(), printMesh(), joinNonEmpty(), nodeState(), printContainers() (+9 more)

### Community 14 - "Container"
Cohesion: 0.12
Nodes (9): fakeRuntime, Container, clip(), Source, meshFilter(), ContainerName, NamesRequest, TestMeshPages() (+1 more)

### Community 15 - "newHarness"
Cohesion: 0.10
Nodes (41): apiError, `internal/server/api`, ErrorBody, ErrorMessage(), adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket() (+33 more)

### Community 16 - "api"
Cohesion: 0.06
Nodes (36): api, MeshSetRequest, RecordRequest, SettingsRequest, 17. Amendments, addIPv6(), Actor, problemStrings() (+28 more)

### Community 17 - "Segmentation"
Cohesion: 0.10
Nodes (13): segLiteral, SegmentView, ZoneView, segDoc, segLiterals(), plannerSegmentation(), PolicyZone(), policyNames() (+5 more)

### Community 18 - "Join"
Cohesion: 0.12
Nodes (25): TestCleanHostname(), TestIsPublic(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), Local(), Presented(), TestHashes() (+17 more)

### Community 19 - "Dir"
Cohesion: 0.05
Nodes (46): enroll(), `internal/fsutil`, meshKey(), TestDefaultSocket(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, Enrollment (+38 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.16
Nodes (33): fakeRunner, TestRunNeedsAnEnrollment(), appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo() (+25 more)

### Community 21 - "hostid.go"
Cohesion: 0.39
Nodes (7): source, clean(), first(), fromFile(), sources(), Read(), sources()

### Community 22 - "net/netip.Prefix"
Cohesion: 0.07
Nodes (34): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+26 more)

### Community 24 - "fakeServer"
Cohesion: 0.24
Nodes (4): fakeServer, signed(), TestSessionShapes(), Definitive()

### Community 25 - "contents_test.go"
Cohesion: 0.57
Nodes (6): docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents()

### Community 26 - "lab"
Cohesion: 0.11
Nodes (22): Coverage, End to end -- `test/e2e`, host, lab, TestDualStack(), deleteNamespaces(), eventually(), lab (+14 more)

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): Add a database migration, `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), fields() (+10 more)

### Community 28 - "env"
Cohesion: 0.13
Nodes (62): checkMachineIDFile(), setupAgentJoin(), setupAgentReset(), setupAgentRun(), socketFlags(), newLogger(), setupVersion(), wantArgs() (+54 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.13
Nodes (18): Container names, 10. Backups, 11. When the control plane is down, 12. Troubleshooting, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh (+10 more)

### Community 31 - "cli"
Cohesion: 0.15
Nodes (25): cli(), TestIPv6DNSRecords(), TestReportedHostFamilies(), harness, TestMeshesAreApartOnTheAgentAPI(), TestMeshesStayApart(), TestNodesServiceInAMesh(), TestSetSettings() (+17 more)

### Community 32 - "context.Context"
Cohesion: 0.09
Nodes (10): runner, kick(), oneLine(), Run(), uiListen(), Result, drain(), RunWindowsService() (+2 more)

### Community 33 - "`internal/server/ui`"
Cohesion: 0.07
Nodes (56): Add a web UI page, The web UI's pages (internal/server/ui), `internal/server/ui`, capped, tabsIn(), TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant() (+48 more)

### Community 34 - "Service"
Cohesion: 0.08
Nodes (13): nameServer, Registrar, resolver, Service, DNS, Want, New(), NewRegistrar() (+5 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (70): rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime(), newFakeServer() (+62 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (124): Agent (`internal/agent/...`), freePort(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), pfx(), TestConfigureHasKernelSemantics(), TestDNS() (+116 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.11
Nodes (32): rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword() (+24 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (55): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+47 more)

### Community 39 - "Policy"
Cohesion: 0.13
Nodes (28): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+20 more)

### Community 40 - ".exec"
Cohesion: 0.10
Nodes (6): newUUID(), mustAffect(), boolInt(), classify(), changed(), Session

### Community 41 - "`internal/server/store`"
Cohesion: 0.13
Nodes (45): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+37 more)

### Community 42 - "Run"
Cohesion: 0.07
Nodes (19): Deps, statusDetail, Deps, dialing, Rechecker, Add an apply step, Reset(), DNS (+11 more)

### Community 43 - "netlinkNet"
Cohesion: 0.13
Nodes (9): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+1 more)

### Community 44 - "NodeConfig"
Cohesion: 0.07
Nodes (35): IPv6Host, Determinism and goldens, `internal/server/phone`, Shared fixtures -- `internal/testfix`, forHost(), proxied(), TestForHostWithoutIPv6(), Desired() (+27 more)

### Community 45 - "planner_test.go"
Cohesion: 0.13
Nodes (43): PlanError, Add a topology rule, `internal/server/planner`, TestCheckRequested(), TestRangeFor(), Problem, Compute(), deliver() (+35 more)

### Community 46 - "Node"
Cohesion: 0.22
Nodes (10): Config and events, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers(), sortPrefixes() (+2 more)

### Community 47 - "netlinkPolicy"
Cohesion: 0.20
Nodes (7): Family, netlinkPolicy, dump(), family(), owned(), toRoute(), toRule()

### Community 48 - "README.md"
Cohesion: 0.15
Nodes (10): Contents, `internal/tools/vulnjudge`, Test catalogue, Building, Contents, Documentation, How it fits together, meshctl (+2 more)

### Community 50 - ".query"
Cohesion: 0.20
Nodes (3): ContainerName, scanAll(), splitList()

### Community 51 - "Input"
Cohesion: 0.15
Nodes (11): Record, Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable(), sortedByID(), Input (+3 more)

### Community 52 - "serveRuntime"
Cohesion: 0.20
Nodes (14): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+6 more)

### Community 53 - "LocalStatus"
Cohesion: 0.09
Nodes (22): Client, Logs, Handler(), logs(), refusal(), oldest(), laptop(), PrefsRequest (+14 more)

### Community 54 - "policy"
Cohesion: 0.23
Nodes (6): policy, Firewall, exitPolicy(), relayPolicy(), wantedPolicies(), Firewall

### Community 55 - "samples_test.go"
Cohesion: 0.12
Nodes (27): block, SampleParams, Sample, NodeSamples(), docWithSamples(), parseSample(), TestContainersDocSamples(), TestNodeSamples() (+19 more)

### Community 56 - "serve"
Cohesion: 0.23
Nodes (13): serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallPS1PinsTheChecksum(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums() (+5 more)

### Community 57 - "Client"
Cohesion: 0.13
Nodes (8): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, ValidChallenge(), DroppedName

### Community 58 - ".do"
Cohesion: 0.10
Nodes (20): MeshList, meshSpace, MeshUpdate, SoleMesh(), checkApart(), clash(), freeCIDR(), freeSlot() (+12 more)

### Community 59 - "Ticket"
Cohesion: 0.19
Nodes (16): brokenStream, SignSession(), TestEventsStreamPushesChanges(), harness, TestDuplicateIsRefused(), TestForeignAndPinning(), TestHelloIsGranted(), TestSessionCallFailures() (+8 more)

### Community 60 - "meshDNS"
Cohesion: 0.18
Nodes (26): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+18 more)

### Community 61 - "IPNet"
Cohesion: 0.11
Nodes (8): Status, One image, two modes, AddrPort(), IPNet(), Prefix(), lookup(), kernel, links

### Community 62 - "Sync"
Cohesion: 0.16
Nodes (31): Sysctls, NewNet(), TestTeardownFakes(), Gateway(), Guard(), Owned(), parse(), SlotOf() (+23 more)

### Community 63 - ".Apply"
Cohesion: 0.22
Nodes (9): applied, Firewall, Runner, atLeast(), normalize(), richRule(), RuleArgv(), FWRule (+1 more)

### Community 64 - "Run"
Cohesion: 0.10
Nodes (24): Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New(), NewAdmin() (+16 more)

### Community 65 - "net/netip.Addr"
Cohesion: 0.10
Nodes (19): network, hostAddrs, nameTable, zone, gateway(), host(), syncProxies(), records() (+11 more)

### Community 66 - "CompilePolicy"
Cohesion: 0.18
Nodes (10): activeAt(), zoneOf(), CompilePolicy(), Overlaps(), Summarize(), meshFirewall(), TestOverlaps(), TestPolicyCompilesDualStack() (+2 more)

### Community 67 - "UI"
Cohesion: 0.11
Nodes (15): hostOf(), New(), validOrNot(), ago(), bytesize(), loadPages(), prefsFlash(), Config (+7 more)

### Community 68 - "serve"
Cohesion: 0.17
Nodes (23): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+15 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.05
Nodes (67): Options, ticketServer, `internal/proto`, Join(), New(), newClient(), sign(), TestAnswersMustBeSigned() (+59 more)

### Community 70 - "Security"
Cohesion: 0.15
Nodes (13): Changes a node asks for, Contents, DNS, Interfaces the agent did not create, Keys and tokens, Security, Several meshes on one host, The host firewall (+5 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.24
Nodes (13): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), TestDownload(), TestExtractChecksTheDLL() (+5 more)

### Community 72 - "embed.go"
Cohesion: 0.26
Nodes (11): ComposeControl(), InstallParams, RenderInstall(), RenderInstallPS1(), renderPS1(), RenderReinstall(), RenderReinstallPS1(), renderShell() (+3 more)

### Community 73 - "rulesFor"
Cohesion: 0.18
Nodes (18): nrpt, nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey() (+10 more)

### Community 74 - "net/http.Request"
Cohesion: 0.07
Nodes (34): writeJSONT(), decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie() (+26 more)

### Community 76 - "Containers"
Cohesion: 0.14
Nodes (14): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, 4. Docker Desktop on Windows, A client node on the host network, A host-mode node's containers, Choosing how to run each node, Container mode (+6 more)

### Community 77 - "Sync"
Cohesion: 0.17
Nodes (6): Policy, mutate(), Policy, Route, Rule, Sync()

### Community 78 - "NewHub"
Cohesion: 0.40
Nodes (6): NewHub(), closed(), TestHubDeliversChallenges(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 79 - "exprs"
Cohesion: 0.09
Nodes (18): Add a command, Add a validator, Contents, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+10 more)

### Community 80 - "sync.Mutex"
Cohesion: 0.11
Nodes (10): syncBuffer, Domain, fakeRegistrar, linkDomain, Registration, resolved, Sysctl, classify() (+2 more)

### Community 81 - "Node"
Cohesion: 0.06
Nodes (49): DNSEntry, NameEntry, NameSet, PhoneRequest, Role, blockOf(), checkGroups(), checkName() (+41 more)

### Community 82 - "Net"
Cohesion: 0.17
Nodes (6): ensureAddr(), Net, Rule, Remove(), syncDropped(), syncRules()

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - "policyForm"
Cohesion: 0.20
Nodes (9): newPolicyForm(), readPolicyForm(), dayBox, policiesView, policyForm, policyRow, policyView, ruleRow (+1 more)

### Community 85 - "Open"
Cohesion: 0.18
Nodes (15): emptyService(), TestFirstAdmin(), TestFirstAdminAfterTheCLI(), TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), insertAt5() (+7 more)

### Community 86 - ".AddMesh"
Cohesion: 0.23
Nodes (8): meshCIDR(), plain6(), Prefix(), Prefix6(), POSIXPath(), Prefix4(), TestPOSIXPath(), TestPrefix4()

### Community 87 - "File"
Cohesion: 0.18
Nodes (9): agentLog(), main(), ttyPassword(), OpenFile(), lockFile(), unlockFile(), lockFile(), unlockFile() (+1 more)

### Community 90 - "Host"
Cohesion: 0.16
Nodes (4): link, exitHost, Host, PeerStats

### Community 93 - "Agent protocol"
Cohesion: 0.18
Nodes (11): Agent protocol, Contents, Errors, Join, Local API, Mesh view, Prefs, Routes (+3 more)

### Community 94 - "Ruleset"
Cohesion: 0.11
Nodes (9): Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel, noKernel (+1 more)

### Community 95 - ".queryRow"
Cohesion: 0.10
Nodes (5): AgentSession, HashToken(), NewToken(), Machine, Refusal

### Community 96 - ".readJSON"
Cohesion: 0.16
Nodes (5): HelloResponse, ValidSecret(), Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.08
Nodes (12): Holder(), Taken(), TestHolder(), adapterPart(), adapterSlot(), NewUserspace(), adapterGUID(), defaultInterface() (+4 more)

### Community 100 - ".do"
Cohesion: 0.07
Nodes (27): CopyInfo, IPv6Request, IPv6View, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile (+19 more)

### Community 101 - "joinFlags"
Cohesion: 0.15
Nodes (23): dataHelp(), dirMesh(), givenDir(), installedDir(), joinedAt(), joinFlags(), machineIDFlag(), meshDir() (+15 more)

### Community 102 - "meshctl/main.go"
Cohesion: 0.11
Nodes (34): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), parsePrefixes() (+26 more)

### Community 103 - "TestSafeApplyStateMachine"
Cohesion: 0.46
Nodes (3): pendingApply, safeApply, TestSafeApplyStateMachine()

### Community 104 - "DefaultPolicy"
Cohesion: 0.39
Nodes (7): DefaultPolicy(), LegacySegmentation(), TestMigrationTurnsRulesIntoPolicies(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), FirewallRule

### Community 105 - "log/slog.Logger"
Cohesion: 0.07
Nodes (21): Change, Deps, Failover, Hub, hubMove, hubState, subscription, hitWindow (+13 more)

### Community 106 - "Service"
Cohesion: 0.06
Nodes (38): replayCache, Access, Config, DB, failures, Lockout, LockoutError, Params (+30 more)

### Community 107 - "nullInt"
Cohesion: 0.18
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 108 - ".Reach"
Cohesion: 0.29
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 109 - "net/netip.AddrPort"
Cohesion: 0.52
Nodes (6): ParseAddr(), Endpoint6(), PeerEndpoint(), TestIPv6Grammar(), Endpoint(), TestEndpoint()

### Community 112 - "Key"
Cohesion: 0.08
Nodes (32): peer(), intervalElems(), TestIntervalElems(), hexKey(), ParseUAPI(), getAnswer(), hexOf(), pfxs() (+24 more)

### Community 113 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

### Community 114 - "fail"
Cohesion: 0.08
Nodes (27): Security invariants, `internal/validate`, TestTimeZone(), TestRef(), TestServiceEntry(), Arg(), Key32(), LocalToken() (+19 more)

### Community 115 - "net.Listener"
Cohesion: 0.14
Nodes (13): bindBoth(), bindPair(), dial(), Listen(), dial(), isPipe(), Listen(), listenSocket() (+5 more)

### Community 117 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

### Community 119 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.09
Nodes (33): `internal/server` and `internal/server/export`, DB, FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport() (+25 more)

### Community 124 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 125 - "Writer"
Cohesion: 0.29
Nodes (3): New(), New(), Writer

### Community 130 - "Slot"
Cohesion: 0.14
Nodes (10): Priorities(), SlotOf(), TestSlotOf(), NewKernel(), NewKernel(), SysctlKeys(), Slot, InterfaceSlot() (+2 more)

### Community 134 - ".Matcher"
Cohesion: 0.29
Nodes (5): Matrix, MatrixCell, PolicyRule, Schedule, SegPolicy

### Community 137 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 146 - "nodeRow"
Cohesion: 0.40
Nodes (5): dashboardData, nodeForm, nodeRow, nodesData, peerAge

### Community 149 - "ServerURL"
Cohesion: 0.18
Nodes (9): Container-mode agent, HelloRequest, CheckServerDomains(), TestCheckServerDomains(), Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase() (+1 more)

### Community 150 - "ToSnapshot"
Cohesion: 0.38
Nodes (8): validateMeshNames(), ToSnapshot(), checkDNS(), IP(), IPv6(), DNSName(), Domains(), Zone()

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 396 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **35 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `AddrRange`, `Test catalogue`, `Join`, `Dir`, `Agent (`internal/agent/...`)`, `lab`, `.mesh`, `Service`, `agent/agent_test.go`, `Policy`, `Run`, `NodeConfig`, `Node`, `README.md`, `serveRuntime`, `meshDNS`, `Sync`, `Run`, ``internal/proto``, `rulesFor`, `exprs`, `Host`, `userspace`, `TestSafeApplyStateMachine`, `Key`, `net.Listener`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Are the 304 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 304 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09768976897689768 - nodes in this community are weakly interconnected._
- **Why does `NodeConfig` connect `NodeConfig` to `ui/ui_test.go`, `WGKey`, `Dir`, `fakeServer`, `cli`, `context.Context`, `agent/agent_test.go`, `Policy`, `Run`, `planner_test.go`, `Node`, `LocalStatus`, `Client`, `Sync`, `Node`, `Agent protocol`, `.do`, `TestSafeApplyStateMachine`, `log/slog.Logger`, `Key`, `Exit`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.14285714285714285 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `IPNet`, `.mesh`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._