# Graph Report - vpn-mesh-utility  (2026-11-01)

## Corpus Check
- 337 files · ~425,559 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .tmpl 16, (none) 5, .css 2)

## Summary
- 4154 nodes · 22339 edges · 141 communities (111 shown, 30 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2637 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `46e73584`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- ToSnapshot
- Schedule
- htmx.min.js
- joinFlags
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- Test catalogue
- deploy_test.go
- io.Writer
- github.com/hafio/vpn-mesh-utility
- DNSView
- fakeServer
- api
- Runtime
- Name
- Dir
- Agent (`internal/agent/...`)
- Run
- net/netip.Prefix
- Net
- Policy
- Input
- lab
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
- nullInt
- netlinkNet
- NodeConfig
- planner_test.go
- Node
- .AddMesh
- README.md
- errorf
- .query
- .queryRow
- netlinkPolicy
- LocalStatus
- `internal/server/api`
- Open
- Host
- Client
- context.Context
- CompilePolicy
- meshDNS
- PhoneConfig
- Sync
- .Apply
- Run
- net/netip.Addr
- .add
- UI
- IdentityKeyB64
- `internal/proto`
- Security
- fetchwintun/main_test.go
- embed.go
- rulesFor
- net/http.Request
- samples_test.go
- Containers
- Sync
- .settings
- nameTable
- sync.Mutex
- Node
- .Set
- io.Reader
- TestPageGoldens
- Deps
- otherPolicy
- File
- Host
- exprs
- Ruleset
- Host
- .readJSON
- userspace
- TestOneRunningCopy
- .do
- net.Listener
- run
- Registration
- DefaultPolicy
- time.Time
- Service
- contents_test.go
- Read
- Writer
- StepError
- Diff
- NamesRequest
- TestContainerNetworks
- fail
- ServerURL
- Version
- PeerStat
- AddrPort
- forms.js
- `internal/server` and `internal/server/export`
- Agent protocol
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
- Error
- 8. The web UI

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
10. `api` - 62 edges

## Surprising Connections (you probably didn't know these)
- `A node's agent as a Swarm service` --references--> `checkStackSample()`  [INFERRED]
  docs/security.md → deploy/samples_test.go
- `Prefs` --references--> `version()`  [INFERRED]
  docs/agent-protocol.md → internal/agent/safeapply_test.go
- `Errors` --references--> `message()`  [INFERRED]
  docs/agent-protocol.md → internal/server/ui/render.go
- `1. Run the control plane (Dokploy)` --references--> `v()`  [INFERRED]
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

## Communities (141 total, 30 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.10
Nodes (18): linkDNS, DNSRecord, FirewallRule, Node, logsData, inspected, listed, ChallengeEvent (+10 more)

### Community 1 - "WGKey"
Cohesion: 0.16
Nodes (42): blockingHost, watchingFirewall, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), TestRoutesSysctlAndFirewall(), IsStepError(), New() (+34 more)

### Community 2 - "ToSnapshot"
Cohesion: 0.16
Nodes (24): dialing, check(), ParseAddr(), TestParseAddr(), newNode(), validateSet(), ToSnapshot(), Endpoint6() (+16 more)

### Community 3 - "Schedule"
Cohesion: 0.16
Nodes (9): The mesh firewall, Schedule, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds(), weekday() (+1 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.08
Nodes (103): A(), Ae(), an(), at(), B(), Be(), bn(), bt() (+95 more)

### Community 5 - "joinFlags"
Cohesion: 0.36
Nodes (10): joinFlags(), meshDir(), meshFlags(), onlyMesh(), sep(), agentRoot(), TestMeshFlags(), TestOnlyMeshFollowsALink() (+2 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.07
Nodes (27): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+19 more)

### Community 7 - "segResolver"
Cohesion: 0.18
Nodes (12): Service, checkSegmentation(), hosts(), mergePorts(), mergeServices(), parseSpans(), PrefixRange(), ServiceEntry() (+4 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "Test catalogue"
Cohesion: 0.17
Nodes (14): Rule, Violation, Packages, and which way they may import, Contents, `internal/archtest`, `internal/tools/vulnjudge`, Test catalogue, Check() (+6 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.17
Nodes (31): checkPosture(), controlEnv(), keys(), mustRead(), runScript(), sameSet(), service(), startHealth() (+23 more)

### Community 12 - "io.Writer"
Cohesion: 0.12
Nodes (34): TestBytesize(), TestPrintEnrollmentInTheLocalZone(), dnsLine(), printExit(), setupFirewallShow(), setupIPv6(), printMesh(), setupMeshAdd() (+26 more)

### Community 14 - "DNSView"
Cohesion: 0.40
Nodes (5): DNSEntry, NameEntry, DNSView, dnsData, recordForm

### Community 15 - "fakeServer"
Cohesion: 0.07
Nodes (19): Change, Hub, subscription, fakeServer, signed(), TestEventsDeliversConfigEvents(), writeJSONT(), ConfigEvent (+11 more)

### Community 16 - "api"
Cohesion: 0.12
Nodes (5): api, 17. Amendments, cli(), readSums(), ExitRouter

### Community 17 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 18 - "Name"
Cohesion: 0.18
Nodes (11): validateMeshNames(), checkGroups(), Ref(), TestRef(), Enum(), Name(), TestEnum(), TestName() (+3 more)

### Community 19 - "Dir"
Cohesion: 0.05
Nodes (44): `internal/fsutil`, meshKey(), sign(), TestAnswersMustBeSigned(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, Enrollment (+36 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.20
Nodes (29): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), removable() (+21 more)

### Community 21 - "Run"
Cohesion: 0.14
Nodes (23): TestCleanHostname(), TestJoin(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), Local(), Presented(), TestHashes(), cleanHostname() (+15 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.04
Nodes (49): otherNet, Routes, pfx(), TestNetHasKernelSemantics(), index(), NewTable(), route(), compare() (+41 more)

### Community 24 - "Policy"
Cohesion: 0.12
Nodes (27): NewPolicy(), NewPolicy(), NewPolicy(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily(), compile() (+19 more)

### Community 25 - "Input"
Cohesion: 0.13
Nodes (14): PlanError, Record, Check(), checkContainers(), checkContainers6(), checkNode6(), checkUsable(), Problem (+6 more)

### Community 26 - "lab"
Cohesion: 0.11
Nodes (21): Coverage, End to end -- `test/e2e`, host, lab, TestDualStack(), deleteNamespaces(), eventually(), lab (+13 more)

### Community 27 - "Snapshot"
Cohesion: 0.13
Nodes (19): Add a database migration, `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), describeSegmentation() (+11 more)

### Community 28 - "env"
Cohesion: 0.16
Nodes (46): checkMachineIDFile(), machineIDFlag(), setupAgentJoin(), setupAgentReset(), setupAgentRun(), socketFlags(), newLogger(), parsePrefixes() (+38 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (42): Command reference, Contents, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit (+34 more)

### Community 30 - ".mesh"
Cohesion: 0.11
Nodes (20): 10. Backups, 11. When the control plane is down, 13. Upgrading, 1. Run the control plane (Dokploy), 2. Enroll the server nodes, 3. Swarm over the mesh, 5. Day to day, 6. DNS (+12 more)

### Community 31 - "DNS"
Cohesion: 0.14
Nodes (6): nameServer, DNS, Want, newNameTable(), DNS, MeshName

### Community 32 - "context.Context"
Cohesion: 0.09
Nodes (13): pendingApply, runner, safeApply, 12. Troubleshooting, kick(), oneLine(), Run(), TestOneLine() (+5 more)

### Community 33 - "newRig"
Cohesion: 0.08
Nodes (51): The web UI's pages (internal/server/ui), `internal/server/ui`, capped, TestFirewallPages(), TestNodeGroupsOnPages(), group(), TestAccessByGrant(), TestHomeWithoutMeshes() (+43 more)

### Community 34 - "Service"
Cohesion: 0.14
Nodes (11): Registrar, resolver, Service, 4. Spokes and exits, Container mode, Start, stop and remove an agent, New(), NewRegistrar() (+3 more)

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (71): rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime(), newFakeServer() (+63 more)

### Community 36 - "testing.T"
Cohesion: 0.04
Nodes (116): Agent (`internal/agent/...`), freePort(), TestDefaultSocket(), TestOpenRuntime(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), TestDNS(), TestKernelHoldsOneTable() (+108 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.12
Nodes (33): Params, rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants() (+25 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (53): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+45 more)

### Community 39 - "New"
Cohesion: 0.16
Nodes (10): newRenderer(), staticHandler(), headers(), New(), sameOrigin(), TestEveryPageRenders(), TestFormsShowWhatApplies(), handler (+2 more)

### Community 40 - ".exec"
Cohesion: 0.11
Nodes (4): mustAffect(), boolInt(), classify(), changed()

### Community 41 - "`internal/server/store`"
Cohesion: 0.14
Nodes (43): `internal/server/store`, TestKeyHistory(), TestRestoreKeepsWhatTheNodeBrought(), addMesh(), TestMeshes(), TestMeshesAreApart(), TestNodeIDsAreNeverReused(), TestRecordsStayInTheirMesh() (+35 more)

### Community 42 - "nullInt"
Cohesion: 0.18
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 43 - "netlinkNet"
Cohesion: 0.11
Nodes (13): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+5 more)

### Community 44 - "NodeConfig"
Cohesion: 0.09
Nodes (18): IPv6Host, Rechecker, Determinism and goldens, Add an apply step, forHost(), Engine, proxied(), PeerStatus (+10 more)

### Community 45 - "planner_test.go"
Cohesion: 0.15
Nodes (40): Add a topology rule, `internal/server/planner`, deliver(), deliverAddr(), encode(), firstDiff(), hasProblem(), lookup() (+32 more)

### Community 46 - "Node"
Cohesion: 0.21
Nodes (11): Config and events, Routes, 4. Roles and topology, Peer, ContainerNetwork(), endpoints(), Node, sortPeers() (+3 more)

### Community 47 - ".AddMesh"
Cohesion: 0.09
Nodes (24): MeshList, meshSpace, MeshUpdate, SoleMesh(), ULAPrefix(), checkApart(), clash(), freeCIDR() (+16 more)

### Community 48 - "README.md"
Cohesion: 0.13
Nodes (11): graphify, Style, What this repo is, You do not run the gates, Building, Contents, Documentation, How it fits together (+3 more)

### Community 49 - "errorf"
Cohesion: 0.09
Nodes (25): segLiteral, segDoc, Actor, segLiterals(), plannerSegmentation(), PolicyZone(), MeshChange, meshChange() (+17 more)

### Community 50 - ".query"
Cohesion: 0.16
Nodes (4): ContainerName, scanAll(), splitList(), AuditEntry

### Community 51 - ".queryRow"
Cohesion: 0.09
Nodes (5): AgentSession, HashToken(), NewToken(), Machine, Refusal

### Community 52 - "netlinkPolicy"
Cohesion: 0.18
Nodes (9): Family, netlinkPolicy, dump(), family(), owned(), TestKernelShapes(), TestOwnsOnlyItsPriorities(), toRoute() (+1 more)

### Community 53 - "LocalStatus"
Cohesion: 0.08
Nodes (27): statusDetail, Add a web UI page, Client, Logs, Handler(), logs(), refusal(), oldest() (+19 more)

### Community 54 - "`internal/server/api`"
Cohesion: 0.05
Nodes (107): countingDB, `internal/server/api`, ErrorBody, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode() (+99 more)

### Community 55 - "Open"
Cohesion: 0.12
Nodes (18): emptyService(), TestFirstAdmin(), TestFirstAdminAfterTheCLI(), TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), TestMigrationMakesMeshOne(), insertAt5() (+10 more)

### Community 56 - "Host"
Cohesion: 0.15
Nodes (4): link, exitHost, Host, TestLinksAddressesAndFailures()

### Community 57 - "Client"
Cohesion: 0.12
Nodes (10): Session, apiError(), clip(), Client, StatusResponse, NamesResponse, HelloRequest, HelloResponse (+2 more)

### Community 58 - "context.Context"
Cohesion: 0.06
Nodes (17): Deps, fakeRuntime, Reset(), DNS, Firewall, Result, dial(), dial() (+9 more)

### Community 59 - "CompilePolicy"
Cohesion: 0.19
Nodes (14): Matrix, MatrixCell, Policy, PolicyRule, activeAt(), zoneOf(), CompilePolicy(), newSegResolver() (+6 more)

### Community 60 - "meshDNS"
Cohesion: 0.24
Nodes (23): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestSameRegistration() (+15 more)

### Community 61 - "PhoneConfig"
Cohesion: 0.18
Nodes (17): Keys and tokens, `internal/server/phone`, Shared fixtures -- `internal/testfix`, Hub(), joinSet(), NewKey(), Render(), aliceKey() (+9 more)

### Community 62 - "Sync"
Cohesion: 0.09
Nodes (37): Sysctls, NewNet(), TestTeardownFakes(), ensureAddr(), Gateway(), Net, Rule, Guard() (+29 more)

### Community 63 - ".Apply"
Cohesion: 0.13
Nodes (14): applied, Firewall, policy, Firewall, Runner, exitPolicy(), normalize(), relayPolicy() (+6 more)

### Community 64 - "Run"
Cohesion: 0.10
Nodes (26): Deps, Route, route, Add an agent API route, ValidChallenge(), NewNodes(), AdminRoutes(), describe() (+18 more)

### Community 65 - "net/netip.Addr"
Cohesion: 0.07
Nodes (29): network, hostAddrs, Apply (agent), TestIsPublic(), gateway(), host(), syncProxies(), firstAddr() (+21 more)

### Community 66 - ".add"
Cohesion: 0.18
Nodes (9): clip(), Ring, addrText(), Rule, merged(), portText(), services(), Range (+1 more)

### Community 67 - "UI"
Cohesion: 0.12
Nodes (14): hostOf(), New(), TestServeAndListen(), validOrNot(), ago(), bytesize(), loadPages(), prefsFlash() (+6 more)

### Community 68 - "IdentityKeyB64"
Cohesion: 0.21
Nodes (12): Document, Import(), document(), linuxOnly(), TestImportIntoANamedMesh(), TestImportRefusesInvalidDocumentsUntouched(), TestImportRefusesWhileTheServerRuns(), TestImportRestoresAnExport() (+4 more)

### Community 69 - "`internal/proto`"
Cohesion: 0.05
Nodes (76): brokenStream, Options, ticketServer, `internal/proto`, ranHere(), Join(), New(), newClient() (+68 more)

### Community 70 - "Security"
Cohesion: 0.14
Nodes (14): A node's agent as a Swarm service, Changes a node asks for, Container-mode agent, Contents, DNS, Interfaces the agent did not create, Security, Several meshes on one host (+6 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.18
Nodes (15): `internal/tools/fetchwintun`, Pin, install(), Sum(), TestInstall(), download(), extract(), fetch() (+7 more)

### Community 72 - "embed.go"
Cohesion: 0.36
Nodes (11): ComposeControl(), InstallParams, RenderInstall(), RenderInstallPS1(), renderPS1(), RenderReinstall(), RenderReinstallPS1(), renderShell() (+3 more)

### Community 73 - "rulesFor"
Cohesion: 0.36
Nodes (11): nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey(), rootKey() (+3 more)

### Community 74 - "net/http.Request"
Cohesion: 0.08
Nodes (21): decode(), reply(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), Session, SetCookie(), anyText() (+13 more)

### Community 75 - "samples_test.go"
Cohesion: 0.13
Nodes (26): block, SampleParams, Sample, NodeSamples(), docWithSamples(), parseSample(), TestContainersDocSamples(), TestNodeSamples() (+18 more)

### Community 76 - "Containers"
Cohesion: 0.14
Nodes (14): 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, 4. Docker Desktop on Windows, A client node on the host network, A host-mode node's containers, Choosing how to run each node, Container mode (+6 more)

### Community 77 - "Sync"
Cohesion: 0.10
Nodes (17): Policy, mutate(), Desired(), Policy, Route, Rule, V6Support, Sync() (+9 more)

### Community 78 - ".settings"
Cohesion: 0.29
Nodes (7): checkDNS(), DNSName(), Domains(), TestDNSName(), TestDomains(), TestZone(), Zone()

### Community 79 - "nameTable"
Cohesion: 0.18
Nodes (4): nameTable, zone, exchange(), records()

### Community 80 - "sync.Mutex"
Cohesion: 0.17
Nodes (7): syncBuffer, linkDomain, resolved, classify(), foreignNetns(), TestClassifyBusErrors(), syncBuffer

### Community 81 - "Node"
Cohesion: 0.05
Nodes (52): apiError, IPv6Request, IPv6View, NameSet, PhoneRequest, RecordRequest, checkName(), containerNetwork() (+44 more)

### Community 82 - ".Set"
Cohesion: 0.31
Nodes (7): nrpt, parseServers(), serverList(), TestNRPTServers(), deleteRule(), flush(), ownedKeys()

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - "TestPageGoldens"
Cohesion: 0.08
Nodes (24): SegPolicy, SegSchedule, newPolicyForm(), readPolicyForm(), readZoneForm(), ruleRows(), scheduleText(), tabsIn() (+16 more)

### Community 85 - "Deps"
Cohesion: 0.22
Nodes (6): Deps, rig, Sysctl, MeshFirewall, Sysctls, Sysctl

### Community 87 - "File"
Cohesion: 0.16
Nodes (10): agentLog(), main(), ttyPassword(), OpenFile(), TestFileRotates(), lockFile(), unlockFile(), lockFile() (+2 more)

### Community 90 - "Host"
Cohesion: 0.08
Nodes (32): peer(), TestConfigureHasKernelSemantics(), hexKey(), ParseUAPI(), getAnswer(), hexOf(), pfxs(), TestParseUAPI() (+24 more)

### Community 93 - "exprs"
Cohesion: 0.07
Nodes (22): Add a command, Add a validator, Contents, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+14 more)

### Community 94 - "Ruleset"
Cohesion: 0.11
Nodes (9): Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel, noKernel (+1 more)

### Community 95 - "Host"
Cohesion: 0.22
Nodes (5): Host, SyncAddrs(), SyncLink(), TestSyncLinkRefusesAForeignInterface(), TestSyncSurfacesHostFailures()

### Community 96 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 98 - "userspace"
Cohesion: 0.07
Nodes (9): Status, One image, two modes, lookup(), adapterGUID(), defaultInterface(), Link, kernel, links (+1 more)

### Community 99 - "TestOneRunningCopy"
Cohesion: 0.40
Nodes (4): tail(), lab, staleTicket(), TestOneRunningCopy()

### Community 100 - ".do"
Cohesion: 0.06
Nodes (36): CopyInfo, MeshSetRequest, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile, ReachRequest (+28 more)

### Community 101 - "net.Listener"
Cohesion: 0.33
Nodes (6): bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), serve()

### Community 102 - "run"
Cohesion: 0.11
Nodes (32): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+24 more)

### Community 103 - "Registration"
Cohesion: 0.46
Nodes (3): Domain, fakeRegistrar, Registration

### Community 104 - "DefaultPolicy"
Cohesion: 0.39
Nodes (7): DefaultPolicy(), LegacySegmentation(), testSegmentation(), TestSegmentationRoundTrip(), TestSegmentationValidate(), TestSetSegmentationRefusesAStranger(), FirewallRule

### Community 105 - "time.Time"
Cohesion: 0.08
Nodes (18): Failover, hubMove, hubState, hitWindow, waitApplied(), Exec, NewUserspace(), NewUserspace() (+10 more)

### Community 106 - "Service"
Cohesion: 0.06
Nodes (34): replayCache, Access, Config, DB, failures, Lockout, LockoutError, Pending (+26 more)

### Community 107 - "contents_test.go"
Cohesion: 0.57
Nodes (6): docFiles(), heading(), slug(), TestDocsOpenWithContents(), TestWithContents(), withContents()

### Community 108 - "Read"
Cohesion: 0.33
Nodes (8): source, clean(), first(), fromFile(), sources(), Read(), TestFirst(), sources()

### Community 109 - "Writer"
Cohesion: 0.29
Nodes (3): New(), New(), Writer

### Community 111 - "Diff"
Cohesion: 0.29
Nodes (6): Diff, versionData, versionRow, versionsData, FieldChange, NodeDiff

### Community 112 - "NamesRequest"
Cohesion: 0.29
Nodes (5): ContainerName, NamesRequest, TestNamesRequestNormalises(), TestNamesRequestRejects(), validNames()

### Community 114 - "fail"
Cohesion: 0.08
Nodes (34): Security invariants, `internal/validate`, addrs(), TestTimestamp(), TestTimeZone(), Timestamp(), TimeZone(), AddrEntry() (+26 more)

### Community 115 - "ServerURL"
Cohesion: 0.33
Nodes (5): Hostname(), isLoopbackHost(), ServerURL(), TestHostnameFoldsCase(), TestServerURL()

### Community 120 - "forms.js"
Cohesion: 0.60
Nodes (5): holds(), update(), updateAll(), valueOf(), wanted()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.11
Nodes (26): `internal/server` and `internal/server/export`, TestRead(), FromSnapshot(), dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName() (+18 more)

### Community 123 - "Agent protocol"
Cohesion: 0.17
Nodes (12): Agent protocol, Container names, Contents, Errors, Join, Local API, Mesh view, Prefs (+4 more)

### Community 130 - "Slot"
Cohesion: 0.10
Nodes (20): Pipe(), TestPolicyHasKernelSemantics(), Holder(), Priorities(), SlotOf(), Taken(), TestHolder(), TestSlotOf() (+12 more)

### Community 151 - "8. The web UI"
Cohesion: 0.67
Nodes (3): 8. The web UI, Each node's local UI, The mesh firewall

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 396 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **30 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `Slot`, `ToSnapshot`, `Test catalogue`, `fakeServer`, `Dir`, `Agent (`internal/agent/...`)`, `Run`, `net/netip.Prefix`, `Policy`, `lab`, `.mesh`, `DNS`, `context.Context`, `agent/agent_test.go`, `netlinkNet`, `NodeConfig`, `Node`, `netlinkPolicy`, `LocalStatus`, `Host`, `meshDNS`, `Sync`, `Run`, `net/netip.Addr`, `UI`, ``internal/proto``, `fetchwintun/main_test.go`, `rulesFor`, `Sync`, `sync.Mutex`, `.Set`, `File`, `Host`, `exprs`, `Host`, `time.Time`, `Read`, `AddrPort`, ``internal/server` and `internal/server/export``?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09660596425302308 - nodes in this community are weakly interconnected._
- **Why does `Slot` connect `Slot` to `ui/ui_test.go`, `ToSnapshot`, `fakeServer`, `api`, `Dir`, `Run`, `Input`, `.mesh`, `context.Context`, `Service`, `testing.T`, `netlinkNet`, `LocalStatus`, `Sync`, `.Apply`, `net/netip.Addr`, ``internal/proto``, `rulesFor`, `Sync`, `Node`, `.Set`, `Deps`, `Host`, `Writer`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Should `htmx.min.js` be split into smaller, more focused modules?**
  _Cohesion score 0.08159073935772965 - nodes in this community are weakly interconnected._
- **Why does `on()` connect `htmx.min.js` to `testing.T`, `userspace`, `.mesh`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._