# Graph Report - vpn-mesh-utility  (2026-10-30)

## Corpus Check
- 337 files · ~420,966 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 18 file(s) not represented in the graph (top: .tmpl 10, (none) 5, .css 2)

## Summary
- 4133 nodes · 22259 edges · 153 communities (120 shown, 33 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2620 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ec165ffb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- Name
- Policy
- htmx.min.js
- joinFlags
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- Test catalogue
- deploy_test.go
- serve
- github.com/hafio/vpn-mesh-utility
- .queryRow
- fakeServer
- api
- serveRuntime
- netlinkPolicy
- Dir
- Agent (`internal/agent/...`)
- Run
- net/netip.Prefix
- Net
- meshHarness
- .relay
- lab
- Snapshot
- env
- Command reference
- .mesh
- Engine
- context.Context
- newRig
- log/slog.Logger
- agent/agent_test.go
- testing.T
- `internal/server/auth`
- testing.T
- New
- .exec
- `internal/server/store`
- nullInt
- netlinkNet
- NodeConfig
- planner_test.go
- .server
- Tx
- README.md
- errorf
- .query
- Net
- Session
- LocalStatus
- `internal/server/api`
- Open
- otherNet
- Client
- sync.Mutex
- Allocate
- meshDNS
- User
- Sync
- .Apply
- Run
- AddrRange
- .meshctl
- UI
- kernel
- `internal/proto`
- log/slog.Logger
- fetchwintun/main_test.go
- serve
- meshctl/docs_test.go
- net/http.Request
- samples_test.go
- Containers
- Sync
- New
- net/netip.Addr
- Registration
- Node
- Node
- Import
- .renderPolicies
- samples_render_test.go
- .BeginEnrollment
- io.Writer
- Host
- exprs
- Ruleset
- Host
- .readJSON
- userspace
- TestMeshFirewall
- .do
- 7. Agent
- run
- build
- .Sync
- time.Time
- Service
- contents_test.go
- Read
- nftKernel
- StepError
- Diff
- NewHub
- Engine
- .Reach
- renderer
- loadMigrations
- 8. Server
- StatusReport
- TestOneRunningCopy
- forms.js
- `internal/server` and `internal/server/export`
- TestContainerNetworks
- Agent protocol
- otherPolicy
- NodeInfo
- TestMain
- .SetSegmentation
- entry
- AdminError
- Slot
- Error
- NewNet
- NewNet
- service_windows.go
- NewKernel
- Refusal
- NewKernel
- Refusal
- usageError
- StatusError
- AliveRequest
- TestQRCode
- DefaultPolicy
- Error
- toAPIError
- noKernel
- stream
- TestFailoverMovesRelays
- walk
- 8. The web UI
- .relayFor

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 304 edges
2. `ui` - 110 edges
3. ``internal/server/api`` - 101 edges
4. `ServerConfig()` - 89 edges
5. `Session` - 85 edges
6. `Slot` - 82 edges
7. `newHarness()` - 73 edges
8. `NodeConfig` - 69 edges
9. `env` - 64 edges
10. `api` - 58 edges

## Surprising Connections (you probably didn't know these)
- `A node's agent as a Swarm service` --references--> `checkStackSample()`  [INFERRED]
  docs/security.md → deploy/samples_test.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Packages, and which way they may import` --references--> `TestModuleBoundaries()`  [INFERRED]
  CLAUDE.md → internal/archtest/archtest_test.go
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

## Communities (153 total, 33 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.13
Nodes (43): TestProxyLoopAppliesAgain(), TestRunNeedsTheDataDirToItself(), NewHost(), NewRoutes(), IsStepError(), New(), failed(), isStep() (+35 more)

### Community 2 - "Name"
Cohesion: 0.06
Nodes (71): Security invariants, Add a validator, `internal/validate`, check(), validateMeshNames(), newNode(), validateSet(), ToSnapshot() (+63 more)

### Community 3 - "Policy"
Cohesion: 0.14
Nodes (27): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, compile() (+19 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "joinFlags"
Cohesion: 0.42
Nodes (9): joinFlags(), machineIDFlag(), meshDir(), meshFlags(), onlyMesh(), sep(), agentRoot(), TestMeshFlags() (+1 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.09
Nodes (22): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+14 more)

### Community 7 - "segResolver"
Cohesion: 0.07
Nodes (27): Matrix, MatrixCell, SegmentView, ZoneView, The mesh firewall, PolicyRule, Schedule, plannerSegmentation() (+19 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "Test catalogue"
Cohesion: 0.16
Nodes (15): Rule, Violation, Input and commands, Contents, Coverage, `internal/archtest`, `internal/tools/vulnjudge`, Test catalogue (+7 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.18
Nodes (27): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+19 more)

### Community 12 - "serve"
Cohesion: 0.16
Nodes (24): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+16 more)

### Community 14 - ".queryRow"
Cohesion: 0.11
Nodes (4): AgentSession, HashToken(), Machine, Refusal

### Community 15 - "fakeServer"
Cohesion: 0.13
Nodes (11): fakeServer, TestSSERoundTrip(), TestSSEWriterRefusesFrameInjection(), WriteComment(), WriteEvent(), WriteRetry(), Definitive(), refusalMessage() (+3 more)

### Community 16 - "api"
Cohesion: 0.11
Nodes (13): api, decode(), reply(), hostFamilies(), decodeJSON(), readBody(), writeJSON(), readSums() (+5 more)

### Community 17 - "serveRuntime"
Cohesion: 0.13
Nodes (15): TestOpenRuntime(), clip(), Detect(), meshFilter(), New(), Sockets(), inspectOf(), serveRuntime() (+7 more)

### Community 18 - "netlinkPolicy"
Cohesion: 0.20
Nodes (7): Family, netlinkPolicy, dump(), family(), owned(), toRoute(), toRule()

### Community 19 - "Dir"
Cohesion: 0.06
Nodes (37): meshKey(), signed(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, Enrollment, NewKeys(), newSeed() (+29 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.19
Nodes (30): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), removable() (+22 more)

### Community 21 - "Run"
Cohesion: 0.10
Nodes (30): TestStartNodeRuns(), ranHere(), TestCleanHostname(), TestIsPublic(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), Local() (+22 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.10
Nodes (14): Routes, index(), NewTable(), route(), compare(), Table, Sync(), NewTable() (+6 more)

### Community 24 - "meshHarness"
Cohesion: 0.15
Nodes (32): `internal/server/api`, cli(), ErrorMessage(), newReplayCache(), TestReplayCache(), TestReplayCacheRefusesWhenFullOfLiveEntries(), TestContainerNetworksWithIPv6(), TestDefaultIPv6() (+24 more)

### Community 25 - ".relay"
Cohesion: 0.33
Nodes (3): checkUsable(), endpoints(), fam

### Community 26 - "lab"
Cohesion: 0.14
Nodes (16): End to end -- `test/e2e`, host, lab, TestDualStack(), deleteNamespaces(), eventually(), lab, inNS() (+8 more)

### Community 27 - "Snapshot"
Cohesion: 0.11
Nodes (23): `internal/server/versions`, TestReconcilerVersionsOnlyRealChanges(), LegacySegmentation(), DecodeSnapshot(), SnapNode, Snapshot, snapNode(), TestDecodeSnapshotRefusesUnknownShapes() (+15 more)

### Community 28 - "env"
Cohesion: 0.12
Nodes (66): checkMachineIDFile(), setupAgentJoin(), setupAgentReset(), setupAgentRun(), socketFlags(), newLogger(), parsePrefixes(), setupVersion() (+58 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.13
Nodes (18): Container names, 10. Backups, 11. When the control plane is down, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 4. Spokes and exits (+10 more)

### Community 31 - "Engine"
Cohesion: 0.08
Nodes (14): Deps, Deps, Rechecker, rig, Add an apply step, Reset(), Sysctl, DNS (+6 more)

### Community 32 - "context.Context"
Cohesion: 0.10
Nodes (8): runner, kick(), oneLine(), Run(), Result, drain(), RunWindowsService(), droppedLogs

### Community 33 - "newRig"
Cohesion: 0.09
Nodes (44): The web UI's pages (internal/server/ui), `internal/server/ui`, capped, TestFirewallPages(), TestNodeGroupsOnPages(), TestAccessByGrant(), TestHomeWithoutMeshes(), TestMeshPages() (+36 more)

### Community 34 - "log/slog.Logger"
Cohesion: 0.06
Nodes (26): dialing, Registrar, resolver, Service, New(), NewRegistrar(), NewRegistrar(), NewRegistrar() (+18 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (71): rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime(), newFakeServer() (+63 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (141): Agent (`internal/agent/...`), `internal/fsutil`, freePort(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx() (+133 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.13
Nodes (28): rig, `internal/server/auth`, newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword(), TestEnrolmentReusesItsSecretAndPeekDoesNotSlide() (+20 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (54): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+46 more)

### Community 39 - "New"
Cohesion: 0.10
Nodes (15): Add a command, Add a web UI page, Extending the code, TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie(), group() (+7 more)

### Community 40 - ".exec"
Cohesion: 0.08
Nodes (9): Add a database migration, newUUID(), mustAffect(), boolInt(), classify(), nullStr(), changed(), Session (+1 more)

### Community 41 - "`internal/server/store`"
Cohesion: 0.14
Nodes (42): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+34 more)

### Community 42 - "nullInt"
Cohesion: 0.15
Nodes (3): DNSRecord, nullInt(), Version

### Community 43 - "netlinkNet"
Cohesion: 0.14
Nodes (9): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), toRule(), IPNet() (+1 more)

### Community 44 - "NodeConfig"
Cohesion: 0.05
Nodes (53): IPv6Host, Determinism and goldens, Keys and tokens, `internal/server/phone`, Shared fixtures -- `internal/testfix`, forHost(), proxied(), TestForHostWithoutIPv6() (+45 more)

### Community 45 - "planner_test.go"
Cohesion: 0.09
Nodes (56): PlanError, Add a topology rule, `internal/server/planner`, TestCheckRequested(), Check(), checkContainers(), checkContainers6(), checkNode6() (+48 more)

### Community 46 - ".server"
Cohesion: 0.30
Nodes (5): 4. Roles and topology, Peer, sortPeers(), sortPrefixes(), plan

### Community 47 - "Tx"
Cohesion: 0.09
Nodes (22): MeshList, meshSpace, MeshUpdate, SoleMesh(), checkApart(), clash(), freeCIDR(), freeSlot() (+14 more)

### Community 48 - "README.md"
Cohesion: 0.13
Nodes (12): graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, Building, Contents, Documentation (+4 more)

### Community 49 - "errorf"
Cohesion: 0.08
Nodes (30): MeshSetRequest, RecordRequest, segLiteral, SettingsRequest, Container-mode agent, segDoc, addIPv6(), Actor (+22 more)

### Community 50 - ".query"
Cohesion: 0.13
Nodes (5): ContainerName, scanAll(), splitList(), AuditEntry, PeerStat

### Community 51 - "Net"
Cohesion: 0.16
Nodes (5): ensureAddr(), Net, Rule, syncDropped(), syncRules()

### Community 52 - "Session"
Cohesion: 0.14
Nodes (22): brokenStream, TestSessionFiles(), ErrorBody, HelloRequest, HelloResponse, TestSessionShapes(), ValidSecret(), TestEventsStreamPushesChanges() (+14 more)

### Community 53 - "LocalStatus"
Cohesion: 0.10
Nodes (20): Client, Logs, Handler(), logs(), refusal(), laptop(), PrefsRequest, PrefsResponse (+12 more)

### Community 54 - "`internal/server/api`"
Cohesion: 0.15
Nodes (25): adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes(), TestAdminPlanErrorsChangeNothing(), TestAdminSetClearsReferences() (+17 more)

### Community 55 - "Open"
Cohesion: 0.17
Nodes (12): TestExportOnceFails(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), insertAt5(), openAt(), TestMigrationTurnsRulesIntoPolicies(), Store, Open() (+4 more)

### Community 57 - "Client"
Cohesion: 0.13
Nodes (8): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, ValidChallenge(), DroppedName

### Community 58 - "sync.Mutex"
Cohesion: 0.10
Nodes (10): fakeRuntime, syncBuffer, clip(), Ring, Container, hostAddr(), Of(), Source (+2 more)

### Community 59 - "Allocate"
Cohesion: 0.16
Nodes (19): TestULAPrefix(), ULAPrefix(), meshCIDR(), Allocate(), Allocate6(), AllocateContainers(), AllocateContainers6(), Block() (+11 more)

### Community 60 - "meshDNS"
Cohesion: 0.16
Nodes (25): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+17 more)

### Community 61 - "User"
Cohesion: 0.20
Nodes (7): Access, Pending, insertUser(), User, scanUser(), TestUsername(), Username()

### Community 62 - "Sync"
Cohesion: 0.15
Nodes (32): Sysctls, NewNet(), TestTeardownFakes(), Gateway(), Guard(), Owned(), parse(), Remove() (+24 more)

### Community 63 - ".Apply"
Cohesion: 0.12
Nodes (15): applied, Firewall, policy, Firewall, Runner, atLeast(), exitPolicy(), normalize() (+7 more)

### Community 64 - "Run"
Cohesion: 0.15
Nodes (18): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+10 more)

### Community 65 - "AddrRange"
Cohesion: 0.08
Nodes (29): addrs(), addrText(), contains(), Rule, merged(), ofFamily(), portText(), services() (+21 more)

### Community 66 - ".meshctl"
Cohesion: 0.15
Nodes (13): A node's agent as a Swarm service, A node's container on the host network (the NAS), Changes a node asks for, Contents, DNS, Interfaces the agent did not create, Security, Several meshes on one host (+5 more)

### Community 67 - "UI"
Cohesion: 0.14
Nodes (13): hostOf(), oldest(), validOrNot(), ago(), bytesize(), prefsFlash(), UIToken, LocalToken() (+5 more)

### Community 68 - "kernel"
Cohesion: 0.15
Nodes (5): Status, One image, two modes, lookup(), kernel, links

### Community 69 - "`internal/proto`"
Cohesion: 0.06
Nodes (58): Options, ticketServer, `internal/proto`, Join(), New(), newClient(), sign(), TestAnswersMustBeSigned() (+50 more)

### Community 70 - "log/slog.Logger"
Cohesion: 0.15
Nodes (8): Change, Hub, subscription, blockingHost, ConfigEvent, Reconciler, Result, NewReconciler()

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.19
Nodes (15): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), main(), run() (+7 more)

### Community 72 - "serve"
Cohesion: 0.19
Nodes (15): serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallPS1PinsTheChecksum(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums() (+7 more)

### Community 73 - "meshctl/docs_test.go"
Cohesion: 0.26
Nodes (13): flagName(), cell(), envDoc(), renderCommands(), sentence(), startFlag(), TestCommandsDoc(), TestContainersDocNamesRealVariables() (+5 more)

### Community 74 - "net/http.Request"
Cohesion: 0.10
Nodes (12): Session, meshChanged(), atoi(), before(), changed(), versionNumber(), message(), writeHTML() (+4 more)

### Community 75 - "samples_test.go"
Cohesion: 0.29
Nodes (11): block, checkComposeSample(), checkKubeSample(), checkRunSample(), checkStackSample(), dig(), docBlocks(), mountsMachineID() (+3 more)

### Community 76 - "Containers"
Cohesion: 0.14
Nodes (14): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, 4. Docker Desktop on Windows, A client node on the host network, A host-mode node's containers, Choosing how to run each node, Container mode (+6 more)

### Community 77 - "Sync"
Cohesion: 0.13
Nodes (6): Policy, mutate(), Policy, Route, Rule, V6Support

### Community 78 - "New"
Cohesion: 0.24
Nodes (12): Config, Params, New(), DefaultParams(), HashPassword(), TestHashUsesTheSpecParameters(), TestVerifyPassword(), TestVerifyRefusesMalformedHashes() (+4 more)

### Community 79 - "net/netip.Addr"
Cohesion: 0.08
Nodes (22): network, hostAddrs, nameTable, zone, syncProxies(), proxy(), records(), isPublic() (+14 more)

### Community 80 - "Registration"
Cohesion: 0.13
Nodes (14): Domain, fakeRegistrar, linkDomain, nrpt, Registration, resolved, parseServers(), serverList() (+6 more)

### Community 81 - "Node"
Cohesion: 0.07
Nodes (48): DNSEntry, NameEntry, NameSet, PhoneRequest, Role, blockOf(), checkGroups(), checkName() (+40 more)

### Community 82 - "Node"
Cohesion: 0.36
Nodes (3): Config and events, ContainerNetwork(), Node

### Community 83 - "Import"
Cohesion: 0.15
Nodes (13): watchdog, Document, Read(), ReadSegmentation(), Import(), document(), linuxOnly(), TestImportIntoANamedMesh() (+5 more)

### Community 84 - ".renderPolicies"
Cohesion: 0.07
Nodes (31): SegPolicy, SegSchedule, anyText(), listOf(), listText(), newPolicyForm(), objectRows(), readPolicyForm() (+23 more)

### Community 85 - "samples_render_test.go"
Cohesion: 0.24
Nodes (14): SampleParams, Sample, NodeSamples(), docWithSamples(), parseSample(), TestContainersDocSamples(), TestNodeSamples(), TestSampleImage() (+6 more)

### Community 86 - ".BeginEnrollment"
Cohesion: 0.31
Nodes (4): Sealer, aad(), newTOTPKey(), TestNewTOTPKeyRebuildsAStoredSecret()

### Community 87 - "io.Writer"
Cohesion: 0.09
Nodes (21): agentLog(), TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), printExit(), main(), printMesh(), printEnrollment() (+13 more)

### Community 90 - "Host"
Cohesion: 0.07
Nodes (18): link, exitHost, Host, peer(), hexKey(), ParseUAPI(), uapiInt(), UAPISet() (+10 more)

### Community 93 - "exprs"
Cohesion: 0.24
Nodes (5): exprs(), ifname(), intervalAddrElems(), TestExprs(), nftOp

### Community 94 - "Ruleset"
Cohesion: 0.19
Nodes (4): watchingFirewall, Kernel, Ruleset, Chain

### Community 95 - "Host"
Cohesion: 0.31
Nodes (3): Host, SyncAddrs(), SyncLink()

### Community 96 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.10
Nodes (7): NewUserspace(), adapterGUID(), defaultInterface(), exitRoutes(), unspecified(), Link, userspace

### Community 99 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

### Community 100 - ".do"
Cohesion: 0.10
Nodes (20): CopyInfo, IPv6Request, IPv6View, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile (+12 more)

### Community 101 - "7. Agent"
Cohesion: 0.25
Nodes (8): 7.1 WireGuard backend, 7.2 Routes, 7.3 Exit switching (spokes), 7.4 DNS, 7.5 Firewall rendering, 7.6 Local UI, 7.7 Platform packaging, 7. Agent

### Community 102 - "run"
Cohesion: 0.15
Nodes (23): agentCommands(), argSynopsis(), commands(), newFlags(), parseArgs(), parseFlags(), printFlags(), shownDefault() (+15 more)

### Community 103 - "build"
Cohesion: 0.22
Nodes (8): Contents, Dev script tasks, Developer guide, Goldens, Releases, Toolchain, build(), nftBatch

### Community 105 - "time.Time"
Cohesion: 0.06
Nodes (27): pendingApply, safeApply, Failover, hubMove, hubState, replayCache, failures, hitWindow (+19 more)

### Community 106 - "Service"
Cohesion: 0.12
Nodes (13): DB, Lockout, resetTOTP(), fullExpiry(), Service, keepAdmin(), load(), session() (+5 more)

### Community 107 - "contents_test.go"
Cohesion: 0.57
Nodes (6): docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents()

### Community 108 - "Read"
Cohesion: 0.39
Nodes (7): source, clean(), first(), fromFile(), sources(), Read(), sources()

### Community 111 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 112 - "NewHub"
Cohesion: 0.40
Nodes (6): NewHub(), closed(), TestHubDeliversChallenges(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 113 - "Engine"
Cohesion: 0.14
Nodes (6): nameServer, DNS, Want, MeshView, MeshName, MeshNode

### Community 114 - ".Reach"
Cohesion: 0.33
Nodes (5): ReachRequest, ReachResult, matrixCell, matrixRow, reachView

### Community 115 - "renderer"
Cohesion: 0.29
Nodes (4): loadPages(), pages, page, renderer

### Community 116 - "loadMigrations"
Cohesion: 0.33
Nodes (3): loadMigrations(), TestLoadMigrationsValidatesNames(), migration

### Community 117 - "8. Server"
Cohesion: 0.33
Nodes (5): 8.1 Storage, 8.2 Data model, 8.3 Web UI, 8.4 Authentication, 8. Server

### Community 118 - "StatusReport"
Cohesion: 0.33
Nodes (6): statusDetail, Rollback, StatusReport, TestStatusReportValidate(), validStatus(), sameRollback()

### Community 119 - "TestOneRunningCopy"
Cohesion: 0.40
Nodes (4): tail(), lab, staleTicket(), TestOneRunningCopy()

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.13
Nodes (22): `internal/server` and `internal/server/export`, DB, FromSnapshot(), dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+14 more)

### Community 123 - "Agent protocol"
Cohesion: 0.14
Nodes (14): Agent protocol, Apply (agent), Contents, Errors, Join, Local API, Mesh view, Prefs (+6 more)

### Community 125 - "NodeInfo"
Cohesion: 0.16
Nodes (12): Health(), phoneHealth(), SharesLine(), TestDashboardHealth(), ago(), bytesize(), TestHelpers(), dashboardData (+4 more)

### Community 130 - "Slot"
Cohesion: 0.08
Nodes (29): nrptRule, Pipe(), uiListen(), TestPolicyHasKernelSemantics(), TestKernelShapes(), desired(), domainName(), fromRules() (+21 more)

### Community 142 - "TestQRCode"
Cohesion: 0.67
Nodes (3): newQR(), phoneData, qrCode

### Community 144 - "DefaultPolicy"
Cohesion: 0.60
Nodes (5): DefaultPolicy(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSetSegmentationRefusesAStranger()

### Community 146 - "toAPIError"
Cohesion: 0.40
Nodes (4): apiError, toAPIError(), TestPlanErrorListsTheProblems(), TestRollbackIsANewVersion()

### Community 148 - "stream"
Cohesion: 0.33
Nodes (4): dial(), dial(), isPipe(), stream

### Community 150 - "walk"
Cohesion: 0.83
Nodes (3): admit(), walk(), packet

### Community 151 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 396 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **33 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Slot` connect `Slot` to `ui/ui_test.go`, `Name`, `api`, `Dir`, `Run`, `.mesh`, `Engine`, `log/slog.Logger`, `testing.T`, `NodeConfig`, `planner_test.go`, `Net`, `Sync`, `.Apply`, ``internal/proto``, `Registration`, `Node`, `userspace`, `Agent protocol`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.0968930765102272 - nodes in this community are weakly interconnected._
- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `Slot`, `Policy`, `Test catalogue`, `serveRuntime`, `Dir`, `Agent (`internal/agent/...`)`, `Run`, `lab`, `.mesh`, `Engine`, `log/slog.Logger`, `agent/agent_test.go`, `NodeConfig`, `Session`, `meshDNS`, `Sync`, `Run`, `AddrRange`, ``internal/proto``, `Registration`, `Node`, `Host`, `exprs`, `time.Time`, `Engine`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.13333333333333333 - nodes in this community are weakly interconnected._
- **Why does `Command reference` connect `Command reference` to `README.md`, `meshctl/docs_test.go`, `lab`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._