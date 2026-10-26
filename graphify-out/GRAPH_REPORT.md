# Graph Report - vpn-mesh-utility  (2026-10-26)

## Corpus Check
- 329 files · ~406,288 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 10 file(s) not represented in the graph (top: (none) 5, .tmpl 2, .css 2)

## Summary
- 4052 nodes · 21801 edges · 123 communities (103 shown, 20 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 2562 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `53efcec8`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ui/ui_test.go
- WGKey
- Name
- newHarness
- htmx.min.js
- env
- meshctl — Design Spec
- segResolver
- dev.sh
- dev.ps1
- archtest.go
- net/netip.AddrPort
- Policy
- github.com/hafio/vpn-mesh-utility
- .queryRow
- otherNet
- net/http.ResponseWriter
- serveRuntime
- New
- Dir
- Agent (`internal/agent/...`)
- Run
- net/netip.Prefix
- Net
- .nodeView
- serve
- lab
- Snapshot
- flag.FlagSet
- Command reference
- .mesh
- NodeConfig
- context.Context
- newRig
- Service
- agent/agent_test.go
- testing.T
- `internal/server/auth`
- testing.T
- log/slog.Logger
- .exec
- commands_test.go
- nullInt
- netlinkNet
- Net
- planner_test.go
- Registration
- `internal/server/api`
- README.md
- wg.go
- .query
- boolInt
- serve
- LocalStatus
- storeNode
- errorf
- ServiceEntry
- Client
- Runtime
- Engine
- meshDNS
- `internal/validate`
- Sync
- .Apply
- Run
- links
- .meshctl
- UI
- Container
- `internal/proto`
- time.Time
- fetchwintun/main_test.go
- net/http.Request
- .AddMesh
- ui
- Exporter
- rulesFor
- Sync
- .readJSON
- net/netip.Addr
- resolved
- Segmentation
- HashPassword
- io.Reader
- .renderPolicies
- deploy_test.go
- Open
- io.Writer
- Host
- services
- Ruleset
- TestContainerNetworks
- samples_test.go
- userspace
- TestOneRunningCopy
- .do
- stream
- run
- syncBuffer
- Schedule
- NewHub
- context.Context
- userRun
- Version
- PeerStat
- StepError
- Machine
- .relayFor
- Engine
- HelloRequest
- TestMeshFirewall
- wgdev
- IP
- `internal/server` and `internal/server/export`
- run
- Slot
- service_windows.go
- Refusal

## God Nodes (most connected - your core abstractions)
1. `Agent (`internal/agent/...`)` - 304 edges
2. `ui` - 106 edges
3. ``internal/server/api`` - 101 edges
4. `ServerConfig()` - 89 edges
5. `Session` - 85 edges
6. `Slot` - 82 edges
7. `newHarness()` - 73 edges
8. `NodeConfig` - 69 edges
9. `env` - 64 edges
10. `api` - 58 edges

## Surprising Connections (you probably didn't know these)
- `Add a command` --references--> `commands()`  [INFERRED]
  docs/developer.md → cmd/meshctl/commands.go
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

## Communities (123 total, 20 thin omitted)

### Community 0 - "ui/ui_test.go"
Cohesion: 0.09
Nodes (35): entry, Error, linkDNS, DNSRecord, FirewallRule, Node, source, apiError() (+27 more)

### Community 1 - "WGKey"
Cohesion: 0.13
Nodes (43): rig, TestProxyLoopAppliesAgain(), NewHost(), NewRoutes(), IsStepError(), New(), failed(), isStep() (+35 more)

### Community 2 - "Name"
Cohesion: 0.15
Nodes (30): check(), validateMeshNames(), newNode(), validateSet(), ToSnapshot(), Endpoint6(), IP(), IPv6() (+22 more)

### Community 3 - "newHarness"
Cohesion: 0.11
Nodes (37): apiError, `internal/server/api`, adminCall(), TestAdminAddIssuesAToken(), TestAdminAddRefusesBadInput(), TestAdminClientOverTheSocket(), TestAdminContainerMode(), TestAdminErrorCodes() (+29 more)

### Community 4 - "htmx.min.js"
Cohesion: 0.06
Nodes (116): Status, 1. The mesh network, once per host, 2. The agent, 3. A service on the mesh, A client node on the host network, A host-mode node's containers, Container mode, Container names (+108 more)

### Community 5 - "env"
Cohesion: 0.13
Nodes (43): checkMachineIDFile(), joinFlags(), machineIDFlag(), meshDir(), meshFlags(), onlyMesh(), sep(), setupAgentJoin() (+35 more)

### Community 6 - "meshctl — Design Spec"
Cohesion: 0.06
Nodes (34): 10. Requirements, 11. CLI reference, 12. Deployment, 13. Build, 14. Swarm notes (informational, not managed by meshctl), 15. Milestones, 16. Open questions, 1. Problem (+26 more)

### Community 7 - "segResolver"
Cohesion: 0.19
Nodes (12): checkSegmentation(), hosts(), newSegResolver(), Overlaps(), parseSpans(), Summarize(), TestOverlaps(), ZoneAddresses() (+4 more)

### Community 8 - "dev.sh"
Cohesion: 0.14
Nodes (31): finish(), has_linux_docker(), log_init(), logline(), main(), NO_COLOR, dev.sh script, build_image() (+23 more)

### Community 9 - "dev.ps1"
Cohesion: 0.15
Nodes (34): Get-HostArch(), Get-HostOs(), Get-Log(), Get-Now(), Get-Text(), Invoke-TrivyImage(), Build-Image(), Build-One() (+26 more)

### Community 10 - "archtest.go"
Cohesion: 0.33
Nodes (7): Rule, Violation, Check(), importsOf(), ModuleRoot(), under(), underAny()

### Community 11 - "net/netip.AddrPort"
Cohesion: 0.21
Nodes (10): dialing, bindBoth(), bindPair(), Listen(), Listen(), listenSocket(), Listen(), New() (+2 more)

### Community 12 - "Policy"
Cohesion: 0.13
Nodes (28): NewPolicy(), NewPolicy(), NewPolicy(), TestBuildBatch(), TestKernelNeedsPrivilege(), Compile(), Local, ofFamily() (+20 more)

### Community 14 - ".queryRow"
Cohesion: 0.11
Nodes (4): AgentSession, HashToken(), NewToken(), Refusal

### Community 16 - "net/http.ResponseWriter"
Cohesion: 0.11
Nodes (12): api, decode(), reply(), hostFamilies(), decodeJSON(), readBody(), writeJSON(), readSums() (+4 more)

### Community 17 - "serveRuntime"
Cohesion: 0.18
Nodes (15): TestOpenRuntime(), Detect(), New(), Sockets(), inspectOf(), serveRuntime(), serveRuntimeOn(), TestDetect() (+7 more)

### Community 18 - "New"
Cohesion: 0.17
Nodes (8): loadPages(), newRenderer(), New(), sameOrigin(), pages, handler, page, renderer

### Community 19 - "Dir"
Cohesion: 0.05
Nodes (42): `internal/fsutil`, meshKey(), TestDataDirIsProtectedOnWindows(), TestWipe(), Dir, Enrollment, NewKeys(), newSeed() (+34 more)

### Community 20 - "Agent (`internal/agent/...`)"
Cohesion: 0.19
Nodes (30): fakeRunner, appliedHost(), encodeCalls(), freshHost(), golden(), newFirewall(), policyInfo(), removable() (+22 more)

### Community 21 - "Run"
Cohesion: 0.10
Nodes (28): TestStartNodeRuns(), hostname(), TestCleanHostname(), TestIsPublic(), TestJoinValidatesBeforeCallingTheServer(), TestMeshAddresses(), Presented(), TestHashes() (+20 more)

### Community 22 - "net/netip.Prefix"
Cohesion: 0.08
Nodes (28): Routes, syncDropped(), index(), NewTable(), route(), compare(), Table, Sync() (+20 more)

### Community 24 - ".nodeView"
Cohesion: 0.20
Nodes (9): CopyLine(), Health(), ago(), writeHTML(), dashboardData, nodeForm, nodeRow, nodesData (+1 more)

### Community 25 - "serve"
Cohesion: 0.23
Nodes (13): serve(), TestHealth(), TestHealthFailsWithoutTheDatabase(), distHarness(), TestDownloadsServeOnlyAllowlistedNames(), TestInstallPS1PinsTheChecksum(), TestInstallScriptFailsWithoutChecksums(), TestInstallScriptPinsTheChecksums() (+5 more)

### Community 26 - "lab"
Cohesion: 0.15
Nodes (15): End to end -- `test/e2e`, host, lab, TestDualStack(), eventually(), lab, inNS(), newLab() (+7 more)

### Community 27 - "Snapshot"
Cohesion: 0.14
Nodes (18): Add a database migration, `internal/server/versions`, SnapNode, Snapshot, snapNode(), Compare(), describeDNS(), fields() (+10 more)

### Community 28 - "flag.FlagSet"
Cohesion: 0.17
Nodes (34): TestBytesize(), dnsLine(), setupDNS(), setupFirewallApply(), setupFirewallCheck(), setupFirewallShow(), setupIPv6(), setupMeshAdd() (+26 more)

### Community 29 - "Command reference"
Cohesion: 0.05
Nodes (41): Command reference, meshctl agent join, meshctl agent reset, meshctl agent run, meshctl control, meshctl dns, meshctl exit, meshctl export (+33 more)

### Community 30 - ".mesh"
Cohesion: 0.08
Nodes (26): Agent protocol, Container names, Errors, Join, Local API, Mesh view, Prefs, Routes (+18 more)

### Community 31 - "NodeConfig"
Cohesion: 0.06
Nodes (37): IPv6Host, Determinism and goldens, Input and commands, `internal/archtest`, `internal/server/phone`, Shared fixtures -- `internal/testfix`, forHost(), proxied() (+29 more)

### Community 32 - "context.Context"
Cohesion: 0.09
Nodes (11): runner, Rechecker, Add an apply step, kick(), oneLine(), Result, drain(), RunWindowsService() (+3 more)

### Community 33 - "newRig"
Cohesion: 0.05
Nodes (59): Add a command, Add a validator, Add a web UI page, Dev script tasks, Developer guide, Extending the code, Goldens, Releases (+51 more)

### Community 34 - "Service"
Cohesion: 0.22
Nodes (4): resolver, Service, exchange(), serve()

### Community 35 - "agent/agent_test.go"
Cohesion: 0.08
Nodes (75): rig, enroll(), enrollAs(), eventually(), joinedTo(), newDir(), newFakeRuntime(), newFakeServer() (+67 more)

### Community 36 - "testing.T"
Cohesion: 0.03
Nodes (136): Agent (`internal/agent/...`), freePort(), TestOneLine(), TestReportKeepsToTheServersLimits(), TestRestartNeedsAConfig(), TestRunNeedsAnEnrollment(), pfx(), TestConfigureHasKernelSemantics() (+128 more)

### Community 37 - "`internal/server/auth`"
Cohesion: 0.13
Nodes (27): rig, `internal/server/auth`, New(), newRig(), TestAccountChangesEndSessions(), TestAccountManagementValidates(), TestAdminsAndGrants(), TestChangePassword() (+19 more)

### Community 38 - "testing.T"
Cohesion: 0.09
Nodes (52): TestAgentRunRuntimeSocket(), passwords(), runWith(), shortDir(), startAdmin(), TestAgentJoinTakesTheTokenFromTheEnvironment(), TestAgentLog(), TestAgentReset() (+44 more)

### Community 39 - "log/slog.Logger"
Cohesion: 0.13
Nodes (9): Hub, subscription, blockingHost, New(), NewUserspace(), ConfigEvent, Reconciler, NewReconciler() (+1 more)

### Community 40 - ".exec"
Cohesion: 0.12
Nodes (4): newUUID(), mustAffect(), classify(), changed()

### Community 41 - "commands_test.go"
Cohesion: 0.08
Nodes (104): AdminError, DNSEntry, Matrix, MatrixCell, NameEntry, NameSet, `internal/server/store`, compareAddr() (+96 more)

### Community 42 - "nullInt"
Cohesion: 0.18
Nodes (3): DNSRecord, nullInt(), nullStr()

### Community 43 - "netlinkNet"
Cohesion: 0.11
Nodes (10): netlinkNet, blackhole(), dropRoute(), dump(), family(), prefixOf(), proxy(), TestKernelShapes() (+2 more)

### Community 44 - "Net"
Cohesion: 0.16
Nodes (6): ensureAddr(), Net, Rule, Remove(), syncProxies(), syncRules()

### Community 45 - "planner_test.go"
Cohesion: 0.06
Nodes (69): PlanError, Config and events, Add a topology rule, 4. Roles and topology, `internal/server/planner`, Peer, Record, TestCheckRequested() (+61 more)

### Community 46 - "Registration"
Cohesion: 0.19
Nodes (8): Domain, fakeRegistrar, Registrar, Registration, TestServiceRevert(), NewRegistrar(), NewRegistrar(), NewRegistrar()

### Community 47 - "`internal/server/api`"
Cohesion: 0.14
Nodes (27): cli(), ErrorMessage(), TestIPv6DNSRecords(), TestReportedHostFamilies(), harness, TestMeshAdmin(), TestMeshesAreApartOnTheAgentAPI(), TestMeshesStayApart() (+19 more)

### Community 48 - "README.md"
Cohesion: 0.18
Nodes (8): Coverage, `internal/tools/vulnjudge`, Test catalogue, Building, Documentation, How it fits together, meshctl, Quick start

### Community 49 - "wg.go"
Cohesion: 0.07
Nodes (42): Deps, Keys and tokens, Reset(), Run(), uiListen(), peer(), AddrPort(), TestAddrPort() (+34 more)

### Community 50 - ".query"
Cohesion: 0.10
Nodes (7): scanMesh(), ContainerName, scanNode(), scanAll(), splitList(), AuditEntry, scanner

### Community 51 - "boolInt"
Cohesion: 0.17
Nodes (3): joinList(), boolInt(), Session

### Community 52 - "serve"
Cohesion: 0.13
Nodes (26): brokenStream, Pending, TestSessionFiles(), ErrorBody, HelloResponse, TestSessionShapes(), ValidSecret(), SignSession() (+18 more)

### Community 53 - "LocalStatus"
Cohesion: 0.12
Nodes (16): Client, Logs, Handler(), logs(), refusal(), PrefsRequest, PrefsResponse, LocalMesh (+8 more)

### Community 54 - "storeNode"
Cohesion: 0.18
Nodes (23): countingDB, harness, joinBody(), postJoin(), signed(), storeNode(), TestConfigAnswers304ForTheAppliedHash(), TestEventsEndAtAFailedSend() (+15 more)

### Community 55 - "errorf"
Cohesion: 0.08
Nodes (44): PhoneRequest, Role, addIPv6(), blockOf(), checkGroups(), checkName(), containerNetwork(), containerSubnet() (+36 more)

### Community 56 - "ServiceEntry"
Cohesion: 0.13
Nodes (12): Service, serves(), mergePorts(), mergeServices(), Service, Ref(), ServiceEntry(), TestRef() (+4 more)

### Community 57 - "Client"
Cohesion: 0.16
Nodes (4): Session, Client, StatusResponse, ValidChallenge()

### Community 58 - "Runtime"
Cohesion: 0.31
Nodes (3): clip(), meshFilter(), Runtime

### Community 59 - "Engine"
Cohesion: 0.10
Nodes (11): statusDetail, Deps, DNS, Engine, Firewall, MeshFirewall, PeerStatus, Rollback (+3 more)

### Community 60 - "meshDNS"
Cohesion: 0.19
Nodes (21): same(), answerOf(), ask(), meshDNS(), newRegistrar(), query(), TestNewZoneRejects(), TestServiceAnswersAndRegisters() (+13 more)

### Community 61 - "`internal/validate`"
Cohesion: 0.07
Nodes (33): Security invariants, `internal/validate`, addrs(), TestTimestamp(), TestTimeZone(), Timestamp(), AddrEntry(), MergeRanges() (+25 more)

### Community 62 - "Sync"
Cohesion: 0.12
Nodes (34): Sysctls, Sysctl, NewNet(), TestTeardownFakes(), Sysctls, Gateway(), Guard(), Owned() (+26 more)

### Community 63 - ".Apply"
Cohesion: 0.12
Nodes (15): applied, Firewall, policy, Firewall, Runner, atLeast(), exitPolicy(), normalize() (+7 more)

### Community 64 - "Run"
Cohesion: 0.09
Nodes (28): Deps, Route, route, Add an agent API route, NewNodes(), AdminRoutes(), describe(), New() (+20 more)

### Community 66 - ".meshctl"
Cohesion: 0.12
Nodes (19): graphify, Packages, and which way they may import, Style, What this repo is, You do not run the gates, 8. The web UI, Each node's local UI, The mesh firewall (+11 more)

### Community 67 - "UI"
Cohesion: 0.13
Nodes (13): hostOf(), oldest(), validOrNot(), ago(), bytesize(), prefsFlash(), UIToken, TestULAPrefix() (+5 more)

### Community 68 - "Container"
Cohesion: 0.14
Nodes (7): fakeRuntime, segLiteral, clip(), Ring, Container, Source, segLiterals()

### Community 69 - "`internal/proto`"
Cohesion: 0.05
Nodes (65): Options, ticketServer, `internal/proto`, Join(), New(), newClient(), sign(), TestAnswersMustBeSigned() (+57 more)

### Community 70 - "time.Time"
Cohesion: 0.06
Nodes (26): pendingApply, safeApply, Change, Failover, hubMove, hubState, replayCache, failures (+18 more)

### Community 71 - "fetchwintun/main_test.go"
Cohesion: 0.24
Nodes (13): `internal/tools/fetchwintun`, Pin, Sum(), download(), extract(), fetch(), TestDownload(), TestExtractChecksTheDLL() (+5 more)

### Community 72 - "net/http.Request"
Cohesion: 0.16
Nodes (9): fakeServer, signed(), Definitive(), SignRefusal(), TestCookieAndCSRF(), ClearCookie(), CookieValue(), SetCookie() (+1 more)

### Community 73 - ".AddMesh"
Cohesion: 0.09
Nodes (23): MeshList, MeshSetRequest, meshSpace, MeshUpdate, SettingsRequest, SoleMesh(), checkApart(), clash() (+15 more)

### Community 74 - "ui"
Cohesion: 0.10
Nodes (9): Session, atoi(), before(), changed(), versionNumber(), message(), actor(), meshHandler (+1 more)

### Community 75 - "Exporter"
Cohesion: 0.32
Nodes (5): DB, NextExport(), TestExporterRunStopsWithItsContext(), TestNextExport(), Exporter

### Community 76 - "rulesFor"
Cohesion: 0.18
Nodes (18): nrpt, nrptRule, desired(), domainName(), fromRules(), keyPart(), nrptKey(), ownedKey() (+10 more)

### Community 77 - "Sync"
Cohesion: 0.06
Nodes (19): Family, netlinkPolicy, otherPolicy, Policy, mutate(), Policy, Route, Rule (+11 more)

### Community 78 - ".readJSON"
Cohesion: 0.22
Nodes (3): Host, Refused, Session

### Community 79 - "net/netip.Addr"
Cohesion: 0.08
Nodes (27): network, hostAddrs, nameTable, zone, Apply (agent), gateway(), host(), records() (+19 more)

### Community 80 - "resolved"
Cohesion: 0.35
Nodes (4): linkDomain, resolved, classify(), foreignNetns()

### Community 81 - "Segmentation"
Cohesion: 0.10
Nodes (21): RecordRequest, segDoc, Actor, plannerSegmentation(), PolicyZone(), MeshChange, meshChange(), movePolicy() (+13 more)

### Community 82 - "HashPassword"
Cohesion: 0.52
Nodes (7): Params, DefaultParams(), HashPassword(), TestHashUsesTheSpecParameters(), TestVerifyPassword(), TestVerifyRefusesMalformedHashes(), VerifyPassword()

### Community 83 - "io.Reader"
Cohesion: 0.33
Nodes (3): watchdog, Read(), ReadSegmentation()

### Community 84 - ".renderPolicies"
Cohesion: 0.07
Nodes (27): anyText(), listOf(), listText(), meshChanged(), newPolicyForm(), objectRows(), readPolicyForm(), readZoneForm() (+19 more)

### Community 85 - "deploy_test.go"
Cohesion: 0.19
Nodes (24): checkPosture(), controlEnv(), keys(), mustRead(), sameSet(), service(), startHealth(), TestControlComposePosture() (+16 more)

### Community 86 - "Open"
Cohesion: 0.28
Nodes (15): `internal/server` and `internal/server/export`, TestExportOnceFails(), TestExportOnceSkipsUnchangedAndKeepsFourteen(), TestExportOnceWritesEachMesh(), Import(), document(), linuxOnly(), TestImportIntoANamedMesh() (+7 more)

### Community 87 - "io.Writer"
Cohesion: 0.10
Nodes (19): agentLog(), TestPrintEnrollmentInTheLocalZone(), printExit(), main(), printMesh(), printChange(), printEnrollment(), ttyPassword() (+11 more)

### Community 90 - "Host"
Cohesion: 0.15
Nodes (3): link, exitHost, Host

### Community 93 - "services"
Cohesion: 0.32
Nodes (7): addrText(), Rule, merged(), portText(), services(), Range, service

### Community 94 - "Ruleset"
Cohesion: 0.10
Nodes (10): watchingFirewall, Kernel, Ruleset, admit(), walk(), Chain, Enforcer, Kernel (+2 more)

### Community 96 - "samples_test.go"
Cohesion: 0.19
Nodes (14): block, checkComposeSample(), checkKubeSample(), checkRunSample(), dig(), docBlocks(), mountsMachineID(), sampleVolume() (+6 more)

### Community 98 - "userspace"
Cohesion: 0.08
Nodes (12): luidOf(), onLink(), adapterPart(), adapterSlot(), NewUserspace(), adapterGUID(), defaultInterface(), exitRoutes() (+4 more)

### Community 99 - "TestOneRunningCopy"
Cohesion: 0.29
Nodes (5): deleteNamespaces(), tail(), lab, staleTicket(), TestOneRunningCopy()

### Community 100 - ".do"
Cohesion: 0.08
Nodes (22): CopyInfo, IPv6Request, IPv6View, NodeAddRequest, NodeDetail, NodeList, NodeSetRequest, PhoneProfile (+14 more)

### Community 101 - "stream"
Cohesion: 0.33
Nodes (4): dial(), dial(), isPipe(), stream

### Community 102 - "run"
Cohesion: 0.11
Nodes (32): agentCommands(), argSynopsis(), commands(), flagName(), newFlags(), parseArgs(), parseFlags(), printFlags() (+24 more)

### Community 104 - "Schedule"
Cohesion: 0.16
Nodes (10): The mesh firewall, PolicyRule, Schedule, Clock(), Window, Schedule(), TestSchedule(), TestWindowHolds() (+2 more)

### Community 105 - "NewHub"
Cohesion: 0.40
Nodes (6): NewHub(), closed(), TestHubDeliversChallenges(), TestHubLatestWins(), TestHubNudgesEveryStream(), TestHubReplacesKicksAndCloses()

### Community 106 - "context.Context"
Cohesion: 0.07
Nodes (23): Access, Config, DB, Sealer, resetTOTP(), fullExpiry(), Service, keepAdmin() (+15 more)

### Community 107 - "userRun"
Cohesion: 0.60
Nodes (5): setupUser(), setupUserGrant(), userCommands(), userRun(), userAction

### Community 113 - "Engine"
Cohesion: 0.15
Nodes (6): nameServer, DNS, Want, MeshView, MeshName, MeshNode

### Community 116 - "TestMeshFirewall"
Cohesion: 0.38
Nodes (3): lab, stream, TestMeshFirewall()

### Community 119 - "IP"
Cohesion: 0.31
Nodes (7): checkDNS(), DNSName(), Domains(), TestDNSName(), TestDomains(), TestZone(), Zone()

### Community 121 - "`internal/server` and `internal/server/export`"
Cohesion: 0.16
Nodes (14): FromSnapshot(), Document, dualSnapshot(), legacy(), snapshot(), TestDualStackExport(), TestExportIsDeterministicAndByName(), TestReadRoundTripsAndRefusesUnknownShapes() (+6 more)

### Community 124 - "run"
Cohesion: 0.40
Nodes (4): TestExporterRunExportsAtNight(), TestRunGivesTheMeshIPv6(), main(), run()

### Community 130 - "Slot"
Cohesion: 0.14
Nodes (10): Pipe(), Priorities(), TestSlotOf(), NewKernel(), NewKernel(), SysctlKeys(), Slot, InterfaceSlot() (+2 more)

## Knowledge Gaps
- **9 isolated node(s):** `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness`, `harness`, `NO_COLOR` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 379 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Agent (`internal/agent/...`)` connect `testing.T` to `WGKey`, `Policy`, `serveRuntime`, `Dir`, `Agent (`internal/agent/...`)`, `Run`, `.mesh`, `NodeConfig`, `context.Context`, `newRig`, `agent/agent_test.go`, `planner_test.go`, `Registration`, `README.md`, `wg.go`, `serve`, `ServiceEntry`, `meshDNS`, `Sync`, `Run`, `.meshctl`, ``internal/proto``, `time.Time`, `rulesFor`, `Sync`, `Host`, `Engine`?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **Are the 303 inferred relationships involving `Agent (`internal/agent/...`)` (e.g. with `.SetFail()` and `.mesh()`) actually correct?**
  _`Agent (`internal/agent/...`)` has 303 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/hafio/vpn-mesh-utility`, `fakeServer`, `harness` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ui/ui_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09060971244481565 - nodes in this community are weakly interconnected._
- **Why does `Slot` connect `Slot` to `ui/ui_test.go`, `Name`, `net/http.ResponseWriter`, `serveRuntime`, `Dir`, `Run`, `.mesh`, `testing.T`, `netlinkNet`, `Net`, `planner_test.go`, `Registration`, `wg.go`, `LocalStatus`, `errorf`, `Engine`, `Sync`, `.Apply`, ``internal/proto``, `net/http.Request`, `rulesFor`, `Sync`, `net/netip.Addr`, `userspace`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Should `WGKey` be split into smaller, more focused modules?**
  _Cohesion score 0.13333333333333333 - nodes in this community are weakly interconnected._
- **Why does `ui` connect `ui` to `.meshctl`, `net/http.Request`, `.AddMesh`, `commands_test.go`, `context.Context`, `New`, `.renderPolicies`, `.nodeView`, `.mesh`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._