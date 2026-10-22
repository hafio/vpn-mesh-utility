# Test catalogue

Every test file, what it pins, and the fixtures and seams it uses. Update this file in the same
change as the tests. Tests are stdlib-only and table-driven; security and error boundaries
assert that bad input is rejected **and** that the downstream call is not made.

## Shared fixtures -- `internal/testfix`

| Fixture | Use |
| --- | --- |
| `Epoch` | the fixed "now" of every time-insensitive test |
| `WGKey(n)` | deterministic base64 32-byte public key (`n` repeated) |
| `IdentityKey(n)`, `IdentityKeyB64(n)` | deterministic Ed25519 key pair seeded from `n` |
| `Clock` | manually advanced clock (`Now`, `Advance`), safe for concurrent use |
| `VPS1`..`VPS3`, `NAS`, `Laptop`, `Phone`, `Servers`, `MeshCIDR` | the reference topology of spec S1/S4 |
| `Token` | a well-formed (43-char) join token |
| `ServerConfig()` | a valid `proto.NodeConfig` for vps-1 in the three-server mesh |
| `SpokeConfig(exit)` | a valid config for the laptop (hub vps-1 relaying to the NAS, vps-2/vps-3 exit-capable); a non-empty exit puts 0.0.0.0/0 on that peer and turns exit mode on |
| `PhoneConfig()` | the phone's computed config: hub vps-1's peer carries the mesh CIDR, vps-2/vps-3 exit-capable |
| `MeshView()` | the server's mesh view of the reference topology: servers in sync, the NAS applying, the laptop (exit vps-2) silent, the phone idle |
| `Policy()` | a server's compiled mesh firewall (amendment A43): the system rules, the laptop allowed SMB to the NAS and denied the servers, a zone of the spokes allowed within it |

`testfix` imports only `internal/proto`, so tests on both sides of the server/agent boundary use
it; `TestModuleBoundaries` forbids production code from importing it.

## `internal/validate`

`validate_test.go`

| Test | Pins |
| --- | --- |
| `TestName`, `TestDNSName`, `TestHostnameFoldsCase`, `TestUsername`, `TestEnum` | name grammars; the shared `hostile` payloads (metacharacters, quotes, newlines, NUL, traversal) are refused |
| `TestIPv4`, `TestPrefix4`, `TestPortAndRange`, `TestEndpoint`, `TestKey32` | address, prefix (host bits rejected with a suggestion), port, endpoint and key shapes |
| `TestTokenMasksTheValue`, `TestLocalTokenMasksTheValue` | a malformed join or local UI token never appears in the error; the error is a `*validate.Error` |
| `TestZone` | the DNS zone setting: 1-4 lowercase labels, no leading or trailing dot, not `local` or `localhost`; the hostile payloads refused. `TestDNSName` takes a record name without the zone, which is no longer trimmed |
| `TestDomains` | the extra DNS domains (amendment A39): lower case, trailing dot dropped, sorted; at most 8, none twice or inside another, not the zone, inside it or around it, no .arpa but home.arpa, not local; every hostile payload refused |
| `TestPipePath` | the Windows local API pipe: `\\.\pipe\` and a plain name only; remote pipes, file paths, nested names and the hostile payloads refused |
| `TestVersionAndSemVer`, `TestServerURL`, `TestPOSIXPath`, `TestFilePath`, `TestText` | version strings, https-only origins (http only on loopback), artifact paths, DSN-safe paths, control-character-free text |
| `TestArg` | the shell-layer charset accepts firewall-cmd options and rich rules and refuses metacharacters |
| `FuzzArg`, `FuzzName` | nothing accepted contains a byte outside the documented charset |

`segment_test.go` (the mesh firewall's grammar, amendment A43)

| Test | Pins |
| --- | --- |
| `TestAddrEntry`, `TestAddrRangePrefix` | an address, a network (host bits refused) or a range first-last (in order) in canonical form: a range that is a network reads as one, a single address bare; IPv6, lists, references and newlines refused |
| `TestServiceEntry` | tcp or udp with no port, one or a range, and icmp without ports; port 0, 65536, leading zeros or signs, reversed ranges, upper case and lists refused |
| `TestRef` | `kind:name` with a kind the field takes and a Name; a string without a colon is a literal; another kind, an empty or malformed name refused |

## `internal/proto`

| File | Test | Pins |
| --- | --- | --- |
| `sign_test.go` | `TestSignVerifyRoundTrip` | headers written by `Apply` parse and verify |
| | `TestVerifyRejectsTampering` | method, path, query, body, timestamp, nonce and key each break the signature |
| | `TestSignIsUniquePerRequest`, `TestCheckClockNeedsNoKey` | identical requests in one second still differ (the replay cache would refuse the second); the clock check runs before any key lookup |
| | `TestVerifySkewBoundaries` | +-60 s accepted, 61 s refused with `ErrSkew` and an NTP hint |
| | `TestVerifyRejectsMissingKey`, `TestParseSignedRejectsMalformedHeaders` | missing/malformed headers are `ErrUnsigned`; no key is `ErrBadSignature` |
| `config_test.go` | `TestHashIsPinned` | the canonical encoding and its sha256 (changing either re-applies every deployed config) |
| | `TestHashIgnoresVersionOnly` | `Version` is excluded from the hash; desired-state changes are not |
| | `TestValidateAcceptsFixtures`, `TestValidateRejects`, `TestValidateReportsEveryProblem` | the agent-side config gate, which every fixture passes (server, spoke, spoke with an exit, a container-mode spoke, a server with containers): version, addresses, MTU, fwmark, peer names, duplicate peers/keys/prefixes, endpoints, keepalive, exit and its choices, sysctl allowlist, firewall fields (nftables: mesh-side rules only), DNS records, upstreams, forward_for and extra domains (lower case, sorted, not around the zone), the container network's shape (amendment A37), the accept-all mesh-zone rule (mesh zone, a source, no port) and a policy (servers only, its grammar; amendment A43) |
| `api_test.go` | `TestJoinRequestValidateNormalises`, `TestJoinRequestValidateRejects`, `TestJoinRequestErrorNeverEchoesTheToken` | join payload validation (including `public_ip`) and token masking |
| | `TestETagQuotesTheHash` | the config entity tag |
| | `TestStatusReportValidate` | heartbeat limits (peers, errors), counters, degraded-step allowlist, the DNS states |
| | `TestPrefsRequestValidate` | a prefs request: exit name grammar (metacharacters refused), keepalive and MTU ranges, an empty request refused |
| `mesh_test.go` | `TestMeshViewValidate` | the mesh view the agent caches and shows: names, roles, IPv4 mesh IPs, hub and exit grammar, the health colours, a state without control characters and at most 200 bytes, no negative time or version, at most 65536 nodes; the containers' names (name, node and IPv4 grammar, at most 65536) |
| `names_test.go` | `TestNamesRequestNormalises`, `TestNamesRequestRejects` | a container names report (amendment A38): names folded to lower case without a trailing dot, an empty list allowed; a newline, an underscore, an empty name, an unknown kind, a non-IPv4 address, a name listed twice for one address and more than 512 entries refused |
| `policy_test.go` | `TestPolicyDecide` | first match wins and the rest is dropped (`RuleDefault`); traffic with an end outside the mesh is `RuleNotMesh` |
| | `TestPolicyMatcherRefusesBadValues` | the agent-side gate of a policy: mesh, action, name, addresses (inside the mesh), protocols, ports, icmp without ports, at most 512 rules and 1024 entries |
| | `TestPolicyRuleString` | one line per rule, as `meshctl firewall show` and the web UI print it |
| `sse_test.go` | `TestSSERoundTrip` | exact wire format of retry, comment and event frames, and the reader |
| | `TestSSEReaderJoinsDataLinesAndDefaultsName`, `TestSSEReaderReportsTruncatedEvent`, `TestSSEReaderBoundsLineLength` | reader edge cases |
| | `TestSSEWriterRefusesFrameInjection` | names or data with newlines/colons are refused and nothing is written |

## `internal/archtest`

`archtest_test.go`: `TestModuleBoundaries` applies the import rules in CLAUDE.md to every
package for GOOS linux and windows, test files included -- among them the single exec site (no
package under `internal/` or `cmd/` but `internal/agent/execrun` imports `os/exec`) and "fakes
and fixtures are for tests only". `TestCheckerCatchesViolations` runs the same checker over
`testdata/violation/` (a fake module that breaks four rules, has one allowed test-only import and
one sanctioned exec site) so a checker that matched nothing cannot pass.

## `internal/server/store`

`store_test.go` uses a fresh SQLite file under `t.TempDir()` per test and `testfix.Clock`.

| Test | Pins |
| --- | --- |
| `TestOpenAppliesMigrations`, `TestOpenIsIdempotent` | schema v2 and default settings (`dns_zone` = mesh); reopening is a no-op |
| `TestOpenRefusesNewerSchema` | a database newer than the binary is refused with an upgrade hint |
| `TestOpenRejectsDSNInjection` | a path carrying `?` cannot rewrite the driver DSN |
| `TestLoadMigrationsValidatesNames` | gaps, zero and malformed migration names fail at startup |
| `TestNodeCRUD`, `TestNodeUniquenessIsAConflict`, `TestNotFound` | node accessors; uniqueness violations map to `ErrConflict` |
| `TestDeleteNodeCascades` | deleting a node removes its tokens, bound records, group memberships, peer status, container names, and clears spokes' hub references |
| `TestContainerNames` | a node's container names are replaced by each report, which says whether anything changed (a new kind counts); listed by name, node and address; an empty report clears one node's only |
| `TestTokenLifecycle` | single use, unknown, expired, reissue revokes the previous token, tokens stored only as hashes |
| `TestTokenConcurrentConsume` | eight concurrent joins with one token: exactly one succeeds |
| `TestVersions`, `TestAudit` | version numbering, listing and lookup; audit rows carry the transaction clock |
| `TestHubState` | relay failover's state (amendment A40) survives an update of the node's desired fields, stays out of the snapshot, and a missing node is `ErrNotFound` |
| `TestRestoreOfAnOlderSnapshot`, `TestSplitDomains` | a snapshot from before a setting existed puts its default back (the zone to mesh, no extra domains); the extra domains read from their comma-separated setting |
| `TestSnapshotRestoreRoundTrip` | restore after drift (node removed, added, edited; setting and mesh firewall changed) reproduces the snapshot exactly, keeping ids |
| `TestNodeContainerColumns` | a node's mode and container subnet (amendment A37): host and none by default, kept by insert and update, one node per subnet (`ErrConflict`), and a snapshot naming only a container-mode node's mode |
| `TestDecodeSnapshotRefusesUnknownShapes` | unknown fields and formats are refused |
| `TestRecordStatus` | a heartbeat updates the node and replaces (not merges) its peer rows |
| `TestPhoneState` | `LastHandshakeWith` is the latest handshake any node reported with a peer (0 for none); `SetApplied` records a phone's profile without counting as a heartbeat |
| `TestOpenReadOnly` | `export`'s read-only open reads, refuses writes, a missing file (pointing at `--data`) and a schema other than the binary's |
| `TestSegmentationRoundTrip`, `TestSetSegmentationRefusesAStranger` (`segments_test.go`) | the mesh firewall's document (amendment A43): a new mesh holds `allow-all` only; written and read back exactly (groups with their nodes, objects and zones by name, policies in order); a group naming a node that does not exist is `ErrNotFound` and nothing stays |
| `TestSetNodeGroups` | a node's groups replaced, a new group made on the way, a group's nodes listed by name; out of every group, the groups stay |
| `TestSegmentationValidate` | every value's grammar and limit: names once per kind, a group's nodes, entries, references by field (no zone in a zone, no svc: in from), actions, descriptions without control characters, at most 32 references and 256 policies |
| `TestMigrationTurnsRulesIntoPolicies` | migration 0006 on a v5 database turns each custom firewall rule into the same policies `LegacySegmentation` makes (to its server, or `role:server`), `allow-all` last, and drops the table |
| `TestRestoreOfASnapshotWithFirewallRules` | a snapshot from before the mesh firewall restores its rules as policies |
| `TestUsers`, `TestSessions` | account and session accessors; `UseTOTPStep` refuses a step at or before the last one (replay), and a new secret starts the counter over |

## `internal/server/auth`

Tests use a fresh store, `testfix.Clock` and cheap argon2 parameters (one test pins the real
ones; `TestHashUsesTheSpecParameters` hashes at the full cost only without `-race`, so `cov` and
the Windows run do it and the race run checks the parameters alone).

| File | Test | Pins |
| --- | --- | --- |
| `password_test.go` | `TestHashUsesTheSpecParameters`, `TestVerifyPassword`, `TestVerifyRefusesMalformedHashes`, `TestCheckPasswordNeverEchoesThePassword` | argon2id t=3 m=64 MiB p=4 PHC strings, the rehash flag, tampered or foreign hashes refused (including a memory bomb), policy errors that never contain the password |
| `totp_test.go` | `TestMatchStepAllowsOneStepOfSkew`, `TestSealerBindsTheSecretToItsUser`, `TestLoadOrCreateKey`, `TestNewTOTPKeyRebuildsAStoredSecret` | +-1 step, nothing else; secrets sealed per user id and refused under another key or tampered; `secret.key` 0600, 32 bytes, refused when loose |
| `lockout_test.go` | `TestLockoutPerAddress`, `TestFailuresOutsideTheWindowAreForgotten`, `TestLockoutPerUsernameAcrossAddresses`, `TestLockoutStaysBounded` | 5 per address per 15 min, 20 per username per hour across addresses, bounded memory |
| `auth_test.go` | `TestSignInWithEnrolment`, `TestSecondFactorAndReplay`, `TestTheSecondFactorWindowIsFiveMinutes` | the two-step sign-in, enrolment, a reused code refused, the 5 min window |
| | `TestLoginRefusals`, `TestLoginLocksOutAfterFiveFailures`, `TestLoginUpgradesAnOldHash` | one error for unknown, wrong and disabled; lockout even with the right password; hash upgrade on sign-in |
| | `TestSessionExpiry`, `TestAccountChangesEndSessions`, `TestLogout` | 12 h idle, 7 days absolute; password change and TOTP reset end sessions; audit without secrets |
| | `TestAccountManagementValidates`, `TestCookieAndCSRF` | usernames and policy; the `__Host-` cookie's attributes; CSRF by header or form field only |
| | `TestEnrolmentReusesItsSecretAndPeekDoesNotSlide`, `TestChangePassword` | a reloaded enrolment shows the same secret; background requests do not extend a session; a password change checks the current one and ends every session |
| `clientip_test.go` | `TestRateLimiter`, `TestClientIP` | the join limiter; X-Forwarded-For believed only from trusted proxies (right-most untrusted hop) |

## `internal/server/ui`

Tests drive the real handler on a fresh store, a fake clock and cheap argon2 parameters, signing
in through the UI itself (password, then TOTP enrolment).

| Test | Pins |
| --- | --- |
| `TestEveryResponseCarriesTheSecurityHeaders` | CSP, `X-Frame-Options`, `nosniff` on pages, static files and 404s |
| `TestSignInFlow`, `TestBadSignInsAreRefusedAndThrottled`, `TestEnrolmentQRCodeIsAPNG`, `TestWrongEnrolmentCodeKeepsTheSecret` | enrolment at the first sign-in, the code step afterwards (wrong code, missing CSRF, right code), no second enrolment or QR for an enrolled account, lockout after five failures, the QR as a same-origin PNG, a wrong enrolment code showing the same secret |
| `TestPagesNeedAFullSession` | redirects and 401s without a session; htmx refreshes get `HX-Redirect`; a pending session reaches no page |
| `TestChangesNeedOriginAndCSRF` | missing or wrong CSRF token, another origin, no Origin: refused, and nothing written |
| `TestNodeLifecycle` | add (token shown once, form kept on errors), edit, disable, enable, new token, remove with confirmation, audited as `user:admin` |
| `TestEditSpoke` | a spoke's hub, exit and keepalive edit; the hub and exit choices offered; refused edits (400/404) change nothing; a new mesh IP from the form with its note, a taken one refused; no token for a disabled node; the computed config on a server's page |
| `TestDNSPage` | the zone with automatic names and the upstreams; a custom record added and removed; a record named like a node (409) and a refusal without CSRF (403) changing nothing |
| `TestFirewallPages` (`firewall_test.go`) | the mesh firewall built one item at a time (amendment A43): a group, a service and an address object, a zone, a policy at position 1, moved, renamed; the zone's outbound rules and addresses; a reachability check and the zone matrix; refusals that keep the form -- an item still named elsewhere (409), a bad action, position or address (400), a dangling member (409), an unknown kind, a missing policy or zone (404), a service checked without a port -- and a removal without CSRF (403); then everything removed in order |
| `TestNodeGroupsOnPages` | a node's groups from the add form and on its page; changed by the edit form; a bad group name refused (400) |
| `TestContainerModePages` | a container-mode node from the add form: its page names the mode and subnet, offers the compose file per runtime and no exit; a host-mode node's edit form puts its containers on the mesh and takes them off; a container-mode phone is refused; the DNS page lists no names, then box's reported names with how one resolves (`web.mesh`, `web.box.mesh`) and why another is not published |
| `TestSpokeComposeDownload` | a spoke's page offers its compose file (a server's does not); the download carries the form's backend and directory as `compose.yml`; a server (409) and a hostile directory (400) come back on the page; no download without a session |
| `TestStoreFailuresAreLoggedNotShown` | with the mesh store failing (`faultyDB`), every page and change answers 500 without the cause, the cause reaches the log, and nothing is written |
| `TestRollbackThatAddsNodesNeedsConfirmation` | the version diff, the confirmation for a membership change, the new version, an unknown version (404), a rollback to the current version changing nothing |
| `TestSettings`, `TestUsersAndAccount`, `TestDashboardFragmentAndExport` | validation of every setting (CIDR only while empty), user management (not on yourself), password change ending the session, the refresh fragment, the export download; the DNS zone shown, refused when it ends in local, and a new one renaming the DNS page's names; extra domains saved sorted and shown on the DNS page, a bad one and the server's own domain refused (400), one inside the zone refused by the plan (409), kept by a form without the field and cleared by an empty one |
| `TestHTMXIsPinned`, `TestTemplatesHaveNoInlineScriptOrStyle` | the vendored htmx's sha256; no inline script, style or event handler in any template; no static directory listing or path escape |
| `TestPageGoldens` | `testdata/{login,dashboard,node,version,firewall}.golden.html` for fixed data, the node's DNS state included (`./scripts/dev.sh regen` creates and updates them) |
| `TestEveryPageRenders`, `TestHelpers` | every template renders; ages (never, a clock ahead of the report shows 0s), byte sizes, the origin check, and times in the server's time zone (UTC, SGT) (the dashboard colours are `api`'s `TestDashboardHealth`) |
| `TestPhonePage`, `TestQRCode` (`phone_test.go`) | a phone added from the nodes page (no token, no compose, no peers table, no exit setting); its via-exit profile shown once as an inline SVG QR code and text, `no-store`, the private key never in the log; a server that is no exit (409) refused on the page; a hub change marks the profile out of date; audited `phone.profile` as `user:admin`. The QR path draws exactly the library's dark modules with a 4-module quiet zone, and a too-long profile's error does not quote it |

## `internal/server/versions`

`TestCompareEqualSnapshotsIsEmpty`, `TestCompareListsEveryKindOfChange`,
`TestARenameIsAChangeNotAReplacement`: added, removed and changed nodes by field (keys as
fingerprints), settings, DNS records and the mesh firewall (each group, object, zone and policy,
a policy with its place); nodes are matched by id. `TestCompareSeesThroughTheMigration`: a
snapshot from before the mesh firewall compares by the policies its rules became.

## `internal/server/api`

Tests run the real handlers on a fresh store with `testfix.Clock`. `countingDB` wraps the store
and counts transactions, so a refusal can be shown to happen before any database access; its
`fail` makes one numbered transaction fail as a broken disk would.

| File | Test | Pins |
| --- | --- | --- |
| `api_test.go` | `TestSignedRequestRejections` | unsigned, malformed nonce, 61 s skew (with `server_time`), oversized body: refused with zero transactions; unknown node, wrong key, query tampered after signing, disabled, not joined: refused; no version written, no stream opened |
| | `TestReplayIsRefusedBeforeTheDatabase` | a repeated signed request is `replay` with zero transactions |
| | `TestJoinEnrollsAndConsumesTheToken` | keys and endpoint registered, the other servers get an event, the token works once |
| | `TestJoinRefusesHostileInputUpFront` | metacharacters, newlines, quotes, bad keys and endpoints, unknown fields: 400, zero transactions, the token never echoed, and the token still works |
| | `TestRefusedJoinKeepsTheToken` | disabled node, name mismatch, a key another node owns: refused and rolled back |
| | `TestJoinIsRateLimitedPerClient`, `TestPickEndpoint` | 10 joins per IP per minute; amendment A17's endpoint order |
| | `TestConfigAnswers304ForTheAppliedHash` | the served config matches the reconciled hash; `If-None-Match` gives 304 |
| | `TestEventsStreamPushesChanges` | the stream opens with the current version, pushes a change, and a second stream ends the first |
| | `TestStatusRecordsTheHeartbeat`, `TestStatusRefusesAnInvalidReport` | heartbeat storage, unknown peers skipped, a rollback audited once per distinct rollback; invalid reports change nothing |
| | `TestMeshViewCarriesNoSecrets` | any joined node reads every node as the dashboard shows it (vps-1 in sync after its heartbeat, the NAS never heard from, the phone seen through vps-1's handshake); no key, endpoint, hostname, observed address or error text in the answer; unsigned: 401 before any transaction |
| | `TestPrefsChangeTheNodesOwnSettings` | a spoke's exit, keepalive and a server's MTU through the reconciler, audited as `node:<id>`; no version for no change; a server's exit or keepalive, a non-exit-capable or unknown exit, a metacharacter, an empty or foreign-field request: refused with the right code and nothing changed |
| | `TestHealth`, `TestHealthFailsWithoutTheDatabase` | `/healthz`, security headers, 405 for other methods; a dead database is a 500 that leaks nothing |
| | `TestJoinRefusesAPhone` | phones cannot join with an agent, and the refusal keeps the token |
| | `TestStoreFailuresAreInternalErrors`, `TestEventsEndAtAFailedSend` | a store failure at each step of config, events, status, prefs, mesh and join is a 500 that leaks nothing; an event stream ends at the first send that fails (retry hint, current version, keepalive) |
| `admin_test.go` | `TestAdminAddIssuesAToken` | allocation, token and commands in the answer, audited without the token |
| | `TestAdminAddRefusesBadInput`, `TestAdminPlanErrorsChangeNothing` | every refusal is a 4xx; a change that breaks the plan writes nothing |
| | `TestAdminSetDisableEnableRemove`, `TestAdminTokenAndShow` | each write; disable closes the node's stream and drops it from peers |
| | `TestAdminClientOverTheSocket` | `AdminClient` against the handler on a real Unix socket, including error codes |
| | `TestContainerNetworksNeedRoom`, `TestAdminContainerMode` | container networks over the admin API: the lowest free /24 for a container-mode node and for a host-mode one on request, the refusals (no endpoint, an exit, a phone, an unknown mode, turning container mode's off), the compose file per runtime and backend, a hub's published port, a hub without an endpoint (409), and deploy's network names equal proto's |
| | `TestMoveMeshIP` | a new mesh IP (amendment A41): every peer and the DNS name follow in a new version; the notes name a firewall policy holding the old address, a server's Swarm advertise address, a phone's profile; the same address changes nothing; a taken, outside, broadcast or malformed address is 400, one inside a container subnet 409, and nothing moves |
| | `TestAdminCompose` | a spoke's compose file with the server's release and origin, both backends and a chosen directory; a server (409), an unknown node (404), a bad name, backend or directory (400) and a development build refused |
| | `TestCustomDNSRecords` (`records_test.go`) | an alias and a fixed record reach every config; an alias of a node not on the mesh changes no config but shows as not live; the DNS page's rows; bad names, both or neither target, a bad IP: 400 before any transaction; a duplicate, a node's name, an IP outside the mesh (409) and an unknown node (404) write nothing; removal; audited `dns.add`/`dns.rm`; a name given with the zone is refused (409) |
| `segment_test.go` | `TestSegmentView` | the mesh firewall of a new mesh (`allow-all` and the system rules), then each zone's addresses and its inbound and outbound rules |
| | `TestSetSegment` | the document replaced whole: a new version, audited, every server's policy (a spoke has none); a bad value refused before the database (400), a dangling reference refused by the plan (409), a group of strangers (400), none written; the same document again changes no config |
| | `TestEditSegment` | the web UI's edits: groups, objects, a zone and a policy at position 1; renames that follow into every reference; moves and bad positions; a zone still named refused (409), then everything removed in order; unknown kinds (400) and missing items (404); audited |
| | `TestReach`, `TestMatrix` | one flow's verdict by node or address, with the deciding rule (system, policy, none, not mesh traffic); a service without one port and an unknown node refused before the database; the zone matrix's allow and partial cells |
| | `TestNodeGroups`, `TestMoveMeshIPNotesTheFirewall` | groups at add and set, sorted, in the node list; bad names and duplicates refused before the database; a node a zone names cannot be removed until it is out of the zone; a mesh IP move names the zone that held the old address |
| | `TestSegmentAdminRoutes` | `GET`, `PUT` and `/check` of `/admin/v1/segmentation` over the admin handler; a bad document is 400 `bad_request` |
| | `TestAdminErrorCodes`, `TestAdminSetClearsReferences` | every refusal of every admin write (404, 400, 409) leaves state untouched, a phone's profile included (a server, a bad, unknown or non-exit exit, an unknown field); clearing an endpoint warns |
| `phone_test.go` | `TestPhoneProfile` | a phone added without a token; a mesh-only profile whose private key matches the public key all three servers peer with, the hub's resolver as DNS; a via-exit profile replaces the key and adds `0.0.0.0/0, ::/0` to the exit's peer; no audit row carries a private key; a hub change marks the profile out of date; the servers' reported handshake in `node list` and `node show` |
| | `TestPhoneProfileNeedsAServer` | without a joined server with an endpoint: 409, the phone keeps no key and nothing is audited |
| `health_test.go` | `TestDashboardHealth` | the dashboard colour and label the web UI and the mesh view share: disabled, not joined, no or a silent heartbeat, a rollback, degraded, applying, a late heartbeat; a phone by its profile and its servers' handshakes; a phone is seen by its handshake; a hub relay failover judged down says since when, in the server's time zone |
| `reconcile_test.go` | `TestReconcilerVersionsOnlyRealChanges` | a no-op writes no version; a change notifies exactly the nodes whose config changed; the snapshot is stored (amendment A16) |
| | `TestReconcilerRollsBackAFailedMutation`, `TestPlanErrorListsTheProblems`, `TestLoadInputRefusesBrokenSettings` | rollback, the 409 mapping, broken settings |
| | `TestRollbackIsANewVersion` | restoring version N writes a new version recording N, notifies the changed nodes, writes nothing when the state already matches, 404 for an unknown N |
| `guard_test.go` | `TestReplayCache`, `TestReplayCacheRefusesWhenFullOfLiveEntries` | the replay window, refusing rather than evicting |
| `hub_test.go` | `TestHubLatestWins`, `TestHubReplacesKicksAndCloses`, `TestHubNudgesEveryStream` | latest-wins delivery; one stream per node; kick and shutdown; a mesh nudge reaches every stream and nudges collapse into one |
| `failover_test.go` | `TestJudgeHubs`, `TestFailoverMovesRelays`, `TestFailoverRun` | relay failover (amendment A40): a hub is down only when its heartbeat AND its handshakes are stale (never heard of counts; silence before the server started does not), back after a steady heartbeat of 2.5 times the threshold; vps-1 silent moves the NAS <-> laptop pair to vps-2 in a new version audited `relay.failover`, the node list shows it down and warns, a second check changes nothing, healthy-since is recorded without a version, and after 5 minutes the pair returns (`relay.restore`); a disabled server loses its state; a failing store is an error; `Run` judges on its own and does nothing with the threshold 0; `Check` alone takes the wall clock |
| `names_test.go` | `TestContainerAddr`, `TestPublishNames` | a container's address is in the node's /24 and not .0, .1, .2 or .255; what is published (amendment A38): one label in the zone (`web.mesh` is `web`), once per node and address, by name, node and address; a node's name, a custom record's, the zone's and the server URL's host refused, as are dotted names outside an extra domain, an extra domain itself, an address the node's network no longer holds and a disabled node's names; a name inside an extra domain published as it is (amendment A39); what each published name resolves as |
| | `TestCheckServerDomains` | an extra domain may not hold the server URL's host (any case, any port); a server known by its address has none |
| | `TestNamesEndpoint`, `TestNamesEndpointRefuses` | `PUT names` stores what is in the node's network and drops the rest with the reason, says what is not published, nudges every stream only when something changed; the mesh view and the DNS page show them; an empty report clears them; a node without a container network (409), a malformed name (400) and an unsigned request (401, no transaction) store nothing |
| `downloads_test.go` | `TestDownloadsServeOnlyAllowlistedNames`, `TestInstallScriptPinsTheChecksums`, `TestInstallScriptFailsWithoutChecksums` | the download allowlist; install.sh carries origin, version and both checksums; missing checksums or dist dir fail with a reason |
| `docs_test.go` | `TestRoutesAreDocumented` | every public and admin route appears in docs/agent-protocol.md |

## `internal/server` and `internal/server/export`

`server_test.go`: `TestRunServesAndShutsDown` (health, the UI on the same listener, `secret.key`
created, a 0600 admin socket, a second server on the same data dir refused, clean shutdown; Linux
only), `TestRunRefusesBadConfig`,
`TestListenAdminRefusesALongPath`. `import_test.go` (Linux only, flock): `TestImportRestoresAnExport`,
`TestImportRefusesWhileTheServerRuns` (nothing created), `TestImportRefusesInvalidDocumentsUntouched`,
`TestLockIsExclusive`. `exporter_test.go`: `TestNextExport`, `TestExportOnceSkipsUnchangedAndKeepsFourteen`
(0600 files, no file for an unchanged night, the newest 14 kept), `TestExporterRunStopsWithItsContext`.
`export_test.go`: `TestExportIsDeterministicAndByName` (same state, same bytes; references by name;
no secrets), `TestReadRoundTripsAndRefusesUnknownShapes`, `TestToSnapshotRoundTrips`,
`TestToSnapshotRefusesHostileDocuments` (every field validated before an import may restore it, the extra domains and the
mesh firewall included), `TestToSnapshotReadsAnOlderExport` (custom rules become policies above `allow-all`),
`TestSegmentationFile` (`meshctl firewall show --yaml` and `apply`: round trip, unknown fields and bad values refused).
`TestImportRestoresAnExport` imports an export made before the mesh firewall: its custom rule becomes a policy.

## `deploy`

| Test | Pins |
| --- | --- |
| `TestUnitPosture` | the unit's `[Service]` section exactly (`ProtectKernelLogs=yes` included) -- any added, removed or widened setting fails |
| `TestInstallGolden`, `TestInstallChangesOnlyTheMarkers`, `TestInstallParses` | `testdata/install.golden.sh`; rendering changes only the four quoted markers; `sh -n` accepts it |
| `TestInstallRefusesHostileValues` | quotes, paths, `$(...)`, newlines, bad checksums are refused before rendering |
| `TestServerComposePosture`, `TestDevComposePosture` | the exact service keys, user, no privileges, no capabilities, read-only root, named volumes only, no widening token anywhere in the file; the server's environment is `TZ` only |
| `TestServerStackPosture` (`reference_test.go`) | the Swarm stack's exact service keys, user, no capabilities, read-only root, the named volume and a tmpfs only, one replicated task on the labelled node stopped before the next, the external network, the compose file's command and healthcheck; no widening token |
| `TestComposeDoc` (`reference_test.go`) | docs/compose.md equals every compose and stack file as the renderers make them for example nodes (`./scripts/dev.sh regen` rewrites it) |
| `TestNASComposeGolden` | `testdata/compose.nas-{userspace,kernel}.golden.yml`; empty backend and directory are the userspace variant in `/volume1/docker/meshctl`, and 1.2.3 is the same image as v1.2.3 |
| `TestNASComposePosture` | both variants' exact service keys and values: uid 0, host network, NET_ADMIN and NET_BIND_SERVICE only, the tun device in userspace only, one bind mount; no other widening token |
| `TestContainerComposeGolden` | `testdata/compose.container-{hub-docker,podman,swarm}.golden.yml` (amendment A37; swarm without names); the defaults are docker, auto and the named volume |
| `TestContainerComposePosture` | every variant's exact service keys and values: uid 0, NET_ADMIN and NET_BIND_SERVICE, the two sysctls, the tun device unless kernel, one published UDP port on a hub, the egress and external mesh networks, the runtime's socket at `/var/run/docker.sock` read-only (Podman's for podman) or, with `--no-names`, no socket and `--runtime-socket off`; no host network, SELinux label, other capability or static address |
| `TestContainerComposeHeader` | the mesh network's commands per runtime as the M6 trial proved them, the join, the service example (dns: the gateway, or for Swarm the servers; dns_search, ndots:1), and the names: how they resolve, the socket and `--no-names`, Podman's socket rootful and rootless |
| `TestHostModeNetworkSetup` | a host-mode node's network commands per runtime: Docker on the existing bridge (inhibit_ipv4, no masquerade, nat-unprotected), Podman unmanaged, Swarm config-only with the same options; never internal; a Swarm service resolving through the servers, anything else through its node; every value validated |
| `TestContainerComposeRefusesHostileValues`, `TestContainerRendererHoldsItsStructure` | every field validated before rendering; below the validators a hostile value stays a value and the header holds |
| `TestNASComposeRefusesHostileValues`, `TestNASRendererHoldsItsStructure` | newlines, `${...}`, `---`, `: `, bad names, origins, backends and development builds refused before rendering; the renderer fed them directly keeps every key, the posture and the comment block, and doubles `$` |

## Agent (`internal/agent/...`)

`internal/agent/agentfake` holds in-memory fakes of the host (`Host` for links, addresses and
WireGuard; `Routes`; `Policy` for exit mode's rules and table routes; `Sysctl`; `Firewall`, which
counts the drift check's `Recheck`; `DNS`, which answers the state a host with or without
systemd-resolved would; `Kernel`, the mesh firewall's table by fingerprint, each call's failure
injectable) that apply changes as the kernel would and count every mutation.
Production code must not import it (`TestModuleBoundaries`). Its own tests
(`TestNetHasKernelSemantics`, `TestConfigureHasKernelSemantics`, `TestLinksAddressesAndFailures`, `TestRoutesSysctlAndFirewall`,
`TestPolicyHasKernelSemantics`, `TestDNS`, `TestKernelHoldsOneTable`) pin that fidelity -- append versus replace of AllowedIPs,
removals, copies, EEXIST and ENOENT -- because every idempotency test depends on it. `Host.SetFail`
and `Policy.Fail`/`FailOn` inject a failure into one method, also while an agent runs.

| Package | Tests | Pins |
| --- | --- | --- |
| `agent` | `TestContainerNamesAreReported`, `TestContainerNamesNeedAContainerNetwork`, `TestMeshNamesReachTheResolver`, `TestOpenRuntime` | container names (amendment A38): reported once the config has a network, again only when they change, a refused report sent with the next change, unpublished names and a failing runtime logged (the same failure once); nothing listed or sent without a network, a failing event stream logged; the mesh view's names reach the resolver from mesh.json, from the server and at once on a `mesh` event; `--runtime-socket` off, a path, auto |
| `names` | `TestOf`, `TestOfSwarmWithoutAStack`, `TestOfCapsTheNames`, `TestHostAddr` | the names of the containers on the network: compose service, Swarm service without its stack, container name, aliases, hostname and hostname.domainname, each once per address, lower case, by name; the ID defaults, invalid names (a Swarm task's), the gateway's and other networks' addresses left out; at most 512, with the count left out |
| | `TestList`, `TestListFailures`, `TestWatch`, `TestDetect` | against a fake Docker-compatible API on a unix socket: GET only, the mesh network filter, a listed ID that is not one never in a path, a container gone or off the network skipped; a refusal, an oversized or non-JSON answer, a failing inspect and a dead socket name the call; every event line a change, an overlong one an error; the first path that is a socket |
| `agent` | `TestRunExplainsRefusals`, `TestIsPublic` | `node_disabled` and `clock_skew` logged with what to do while the cached config stays up; public-address detection |
| | `TestRunFollowsTheServer`, `TestRunPollsWhenTheStreamIsSilent`, `TestRunBootsFromTheCacheWithoutTheServer`, `TestRunKeepsTheOldConfigWhenANewOneFails`, `TestRunNeedsAnEnrollment` | the run loop against a fake server that verifies every signature: first apply and heartbeat, pushed and polled changes, the cache, boot without the server, a refused config reported and not cached |
| | `TestJoin`, `TestJoinValidatesBeforeCallingTheServer`, `TestCleanHostname`, `TestOneLine` | the join request, enrollment, service install or the reason it is skipped, cache dropped on a node change, keys kept; status text always passes the server's validation |
| | `TestRunRefusesAnUnusableDataDir`, `TestRunIgnoresDamagedConfigFiles`, `TestRunReportsUnreadablePeerStatus`, `TestReportKeepsToTheServersLimits`, `TestDefaultSocket` | a bad data dir path, a damaged enrollment, missing keys, an unknown `--wg-backend` or an unusable socket stop Run before the host is touched, the socket even with a cached and a pending config (neither applied nor removed, no step asked); a damaged cached or pending config is ignored (the pending one removed, no rollback reported); unreadable peer status is a reported error; a report keeps at most 32 errors |
| | `TestReportNamesTheBackend` | the heartbeat names the backend in use; a fake host stands for the kernel unless the test names userspace |
| | `TestDNSOnAndOffStayOnTheNode`, `TestDNSSettingThatCannotBeSaved` | `POST /v1/dns` reaches the DNS step at once, is reported, is kept in `local.json` across a restart and never reaches the server; the resolver stops with the agent; a setting that cannot be saved is a 500 that applies nothing |
| | `TestDNSFailureIsReported`, `TestDriftCheckPutsBackWhatChanged` | a failing DNS step degrades the node with its error and the tunnel up; the drift check rebuilds a deleted `wg0` without a new config, after the firewall was asked to look again |
| | `TestPrefsNeedTheServer` | `meshctl exit` with the server down is a 502 naming the call, and exit mode stays off |
| | `TestExitThroughTheLocalAPI` | `POST /v1/prefs` reaches the server signed, the agent fetches the new config without an event, exit mode comes on, the status shows the exit and its choices; a server refusal comes back with its status |
| | `TestMeshViewIsCached`, `TestMeshViewIsRefreshed` | the mesh view fetched at start into `mesh.json`; under 10 s old it is shown without asking again; a new config has it fetched again at once, and a look right after is answered from that fetch; a restarted agent with the server down shows the cached view and the server's reason; refreshed every `MeshEvery` |
| | `TestLogsAreServed` | `GET /v1/logs` reads the ring the agent's log writes to |
| | `TestLocalUIThroughTheAgent`, `TestLocalUIOffOrUnavailable`, `TestRestartNeedsAConfig` | `POST /v1/ui-token`'s link signs a real browser into the UI, whose dashboard shows the node and whose restart drops and re-adds both peers; the status carries the MTU; without a UI, or with its port taken, the agent runs and the link is refused with the reason; a UI address off loopback stops Run before the host is touched when no config is cached (`agent run` refuses it before Run); a restart before any config is refused |
| `agent` (`safeapply_test.go`) | `TestSafeApplyStateMachine` | the state machine on a fake clock: a report sent before the apply does not confirm, the window, a pending config replaced keeps the confirmed one as prev, disarm |
| | `TestSafeApplyConfirmsWhenTheServerAnswers`, `TestSafeApplyRollsBackWithoutAnAnswer`, `TestSafeApplyRecoversAPendingConfigAtBoot`, `TestSpokesDoNotArmSafeApply`, `TestSafeApplyDisarmsWhenTheServerGoesBack`, `TestSafeApplyReportsAFailedRollback` | the runner with a 1 s window on a fake clock: confirm moves the cache; no answer puts the previous config back, reports the rollback and never re-applies the rejected version, while the next version is applied and confirmed; a pending config found at boot is rolled back the same way; spokes cache at once; the server going back to the confirmed config disarms without a rollback; a rollback that fails too is reported as an error with the rollback |
| `apply` | `TestApplyIsIdempotent` | the second apply of the same config changes nothing on the host |
| | `TestMeshFirewallStep` | the mesh firewall (amendment A43): in the kernel when the firewall step runs, left alone when unchanged, replaced (re-judging flows) when changed; a config without a policy removes it after the firewall step, keeps it while that step fails (degraded firewall only), a failed removal degrades `policy`; a refused table stops the apply before the firewall step; a node without a mesh firewall refuses a policy, and one for another mesh is not applied |
| `policy` | `TestCompileGolden` | `testdata/ruleset.golden.txt`: the table of a server, and of one with containers, in nft's spelling with where each rule comes from |
| | `TestCompileAgreesWithDecide` | for every pair of addresses (servers, spokes, a phone, containers, the mesh's unassigned, outside), protocol and port, forwarded, received and sent: the table admits exactly what `proto.Matcher.Decide` allows, but for traffic within the node and loopback |
| | `TestCompileMerges` | a rule's addresses and ports merged when the config's overlap, touch or nest; tcp without ports is every port |
| | `TestCompileRefuses`, `TestRulesetText` | a bad mesh, an address outside it for the node, bad entries, services, ports, actions; one rule in nft's spelling, the count and the fingerprint |
| | `TestEnforcer` | `Sync` installs once, leaves the same table alone, replaces a changed or foreign one and deletes the mesh flows' conntrack (not on a first install), removes for nil; a failed read, install or removal says which, a failed conntrack deletion does not fail the step |
| | `TestBuildBatch`, `TestIntervalElems`, `TestExprs` (`nft_linux_test.go`) | the netlink batch: the six chains, the base ones accepting by default; every set added before the rule that looks it up; the fingerprint on the first rule only; interval elements as nft sends them (an end at 0 first, none past the top), and each match's expressions |
| | `TestKernelNeedsPrivilege`, `TestNoKernelOffLinux` | without CAP_NET_ADMIN every kernel call is refused and says so (skipped as root, where it would change the host); off Linux there is no table and one is refused |
| | `TestApplyValidatesBeforeTouchingTheHost`, `TestApplyStopsAtTheFailingStep`, `TestContainersStep`, `TestFirewallFailureOnlyDegrades`, `TestRemovedPeerLeavesDeviceAndRoutes`, `TestStatsNamesKnownPeers` | step order and failure modes |
| | `TestRestartRenewsEveryPeer` | restart drops both peers and adds them back (two Configure calls, routes untouched) after rechecking the firewall; a missing wg0 is rebuilt, a failed drop is the peers step's error, a foreign wg0 is refused with nothing changed |
| | `TestDNSFailureOnlyDegrades`, `TestDNSFollowsDNSOffAndExitMode`, `TestReassertPutsBackWhatDrifted` | step 11 gets the mesh IP, the config's DNS, exit mode and `dns off`, and only degrades; `Reassert` rechecks the firewall, changes nothing on an unchanged host and rebuilds a deleted interface; `Close` stops the DNS step |
| | `TestExitModeThroughTheHost` | without policy routing, step 6 hands exit mode to a host that does it itself (Windows): on with an exit, off without, nothing on a repeat, a failure fatal |
| | `TestExitModeOnSwitchOff`, `TestExitModeFailureIsFatal` | exit on installs the rules, a switch moves 0.0.0.0/0 between peers without touching them, off removes them, 0.0.0.0/0 never in the main table; a policy failure is fatal; an engine without policy routing refuses exit on |
| `exitmode` | `TestDesiredIsWgQuicksExitRouting`, `TestSyncTurnsExitModeOnAndOff`, `TestSyncReplacesAStaleRuleOfItsOwn`, `TestSyncStopsAtTheFirstFailure`, `TestSyncReportsEachFailure` | the exact rules and routes (IPv4 and IPv6 blackhole), routes before rules on, rules before routes off, idempotent, a stale own rule replaced, exit off on a clean host touches nothing; a failure of each call stops Sync there and counts the changes before it |
| | `TestKernelShapes`, `TestOwnsOnlyItsPriorities`, `TestKernelPolicyWithoutPrivilege` (Linux) | the netlink rule and route each one becomes (protocol 201, inverted fwmark, blackhole); unprivileged listing works and every change is refused with what it attempted (skipped as root; the kernel path itself is e2e's) |
| `wintun` | `TestInstall`, `TestEnsureWithoutABundle` | the Windows driver is written beside the binary once and only when it matches its pin; one there that matches is kept, a changed one replaced, a bundle that misses its pin refused; a build without the `wintun` tag has nothing bundled and says how to get a binary that has it. The Windows host itself (`wg/userspace_windows.go` with its exit mode, `routes/table_windows.go`, `dns/nrpt_windows.go`) needs an elevated prompt, the bundled DLL and the registry, so it has no unit test: it is checked on a Windows laptop (spec S10) |
| `wg` | `TestDiffChangesOnlyWhatDiffers`, `TestDiffLeavesARoamingDialInPeerAlone`, `TestSyncIsIdempotent`, `TestSyncLinkRefusesAForeignInterface`, `TestSyncSurfacesHostFailures`, `TestParseKey`, `TestDesiredFromServerConfig` | the device diff and the interface steps (amendment A18) |
| | `TestUAPISetGolden`, `TestParseUAPI`, `TestParseUAPIRejectsMalformedValues` | the userspace backend's text protocol: `testdata/uapi-set.golden.txt` for a change of every kind (an empty change renders nothing); a device's "get" answer read back with peers sorted and a zero handshake kept zero, so a re-apply diffs to nothing; malformed values named by line and key, a key's value never quoted |
| | `TestOpenRefusesUnknownBackends`, `TestUserspaceWithoutADevice` (Linux) | `--wg-backend` takes only auto, kernel and userspace; the userspace host with no device: no interface is ours, reads and writes refused, wireguard-go's log mapped to debug and warn (creating the tun is e2e's) |
| | `TestOpen`, `TestCreateLinkNeedsPrivilege` (Linux) | each backend opens unprivileged and is the one reported (auto: what DetectBackend finds; no /dev/net/tun names the device); without CAP_NET_ADMIN either CreateLink fails with the interface and a hint, leaving no interface or device (skipped as root) |
| `dns` | `TestZoneFollowsTheConfig` | another zone in the config answers its own names, refuses the default's, and is the one registered with systemd-resolved and NRPT |
| `dns` | `TestNRPTRules`, `TestNRPTServers` | the Windows registration as NRPT rules: `.mesh`, each extra domain, and `.` in exit mode, each under a key with the agent's prefix (a domain's from its name; the pre-A39 zone key is the agent's too, another program's is not), reading back as the same registration; a search domain has no rule; the server list round-trips and skips junk |
| `dns` | `TestZoneAnswersExtraDomains` | extra domains (amendment A39): a container's name there answered (case kept, TTL 10 s, not authoritative), NOERROR for another type, every other name there forwarded from a split node too, a domain that is not extra still refused, CHAOS refused; the host registers the zone and the domains |
| `dns` | `TestZoneAnswersContainerNames` | container names (amendment A38): `web.mesh` every node's addresses, rotated, `web.nas.mesh` one node's, TTL 10 s; NODATA for another type; a node's record wins; a container may fill an empty non-terminal; dotted and non-IPv4 names, and another node's, are NXDOMAIN |
| `dns` | `TestZoneAnswers`, `TestNewZoneRejects` | A records (the name's case kept), NODATA for another type and for an empty non-terminal, NXDOMAIN with an SOA, the apex SOA; other zones REFUSED in split mode, forwarded in exit mode and for a `forward_for` phone; another class REFUSED, two questions FORMERR; a section that does not parse |
| | `TestServiceAnswersTheContainers`, `TestServiceAnswersAndRegisters` | on a loopback socket against a loopback upstream: the zone over UDP and TCP, REFUSED then forwarded after exit mode came on, SERVFAIL with the upstream gone; the registration (`~mesh`, plus `~.` and default-route in exit mode) set once and left alone on a repeat; `dns off` takes it back and the resolver keeps answering; no DNS section stops both |
| | `TestServiceWithoutAResolverToRegisterWith`, `TestServiceRegistrationFailuresAreReported`, `TestServiceBinding`, `TestSameRegistration` | no registrar or resolved not running: `unavailable` (`off` when turned off) and nothing written; a failing read, set or revert named, with the resolver still answering; a taken port and a foreign address reported before any registration; a resolver that died is started again; registrations compared as sets, `mesh.`/`MESH`/`""`/`.` normalised |
| | `TestClassifyBusErrors`, `TestRegistrarNeedsTheInterface` (Linux) | resolved's absence on the bus is `unavailable`, a refusal names CAP_NET_ADMIN; a missing interface fails before anything reaches resolved |
| `routes` | `TestDesiredSkipsTheDefaultRouteAndDuplicates`, `TestSyncAddsThenRemovesAndIsIdempotent` | routes follow AllowedIPs, never 0.0.0.0/0 |
| `firewall` | `TestFreshHostArgvGolden` | `testdata/fresh.golden.jsonl`: the exact firewall-cmd argv for a fresh host |
| | `TestApplyIsIdempotent`, `TestRemovedRuleIsRemovedBothWays`, `TestNilFirewallRemovesOnlyOurRules` | state-file diffing, runtime and permanent |
| | `TestPingRule` | a ping rule is the rich rule accepting ICMP echo requests from its source, added at runtime and permanently (amendment A42); one with a port or without a source is refused before firewall-cmd (`TestHostileRulesNeverReachFirewallCmd`) |
| | `TestAcceptAllRule` | the mesh zone's accept-all rule (amendment A43) is the rich rule accepting everything from its source, added at runtime and permanently; one without a source, with a port or outside the mesh zone is refused before firewall-cmd |
| | `TestHostileRulesNeverReachFirewallCmd` | a hostile rule runs no command at all |
| | `TestContainersBridgeZone`, `TestForwardingPolicies` | `testdata/policies.golden.jsonl`: the relay and exit policies created --permanent and one reload; not asked again while unchanged; a restarted agent finds them and changes nothing; one edited by hand is put back whole; one no longer wanted is deleted |
| | `TestMissingFirewalldIsActionable`, `TestOldFirewalldIsRefused`, `TestPolicyFailureIsReported`, `TestAtLeast`, `TestDamagedStateIsActionable` | failure messages; a failed policy is tried again on the next apply |
| | `TestRecheckPutsBackWhatDrifted`, `TestRecheckAsksAboutARuleSpelledDifferently` | after `Recheck` an unchanged host changes nothing; wg0 moved out of the zone at runtime and a rule removed by hand are put back (runtime binding only, the rule both ways); a rule listed in other words is confirmed with `--query-rich-rule`; without `Recheck` nothing is asked |
| | `TestEachFailingCallStopsTheApply` | each firewall-cmd call of a fresh host with a hand-edited policy, failed in turn: the apply returns that error and runs nothing after it |
| `execrun` | `TestCheckAllowsFirewallArgv`, `TestCheckRefuses`, `TestRunRefusesBeforeExecuting`, `TestCappedKeepsTheHead` | the allowlist and the argv gate |
| | `TestRunCapturesOutputWithAMinimalEnvironment`, `TestRunFoldsStderrIntoTheError`, `TestRunTimesOut`, `TestRunReportsAMissingProgram` | the real exec path against fake `firewall-cmd` / `systemctl` scripts on a temp PATH (not on Windows) |
| `containers` | `TestGatewayRules`, `TestGuardRules`, `TestSyncContainerMode`, `TestSyncWithoutTheMeshNetwork`, `TestSyncHostMode`, `TestSyncFailures`, `TestKernelShapes`, `TestKernelNetWithoutPrivilege` | the gateway's rules and the host-mode guard exactly as the M6 trial proved them; the gateway's .2, blackhole route before the rules, a stale owned rule removed, idempotence, rules then route taken away; the agent off the mesh network says what to do; host mode makes the bridge, its .1 and sysctls; every failure named; the netlink shapes, and no change without privilege |
| `sysctl` | `TestApplyWithoutProc`, `TestApplyChangesOnlyWhatDiffers`, `TestApplyRefusesUnknownKeysBeforeTouchingAPath`, `TestApplyReportsAMissingParameter`, `TestApplyContainerKeys`, `TestApplyReportsAReadOnlyParameter` | against a fake `/proc/sys`; the container network's keys; a read-only parameter names the compose file |
| `state` | `TestMeshCache`, `TestLocalSettings`, `TestKeysAreCreatedOnceAndKept`, `TestLoadKeysRefusesLooseOrDamagedFiles`, `TestEnrollmentRoundTrip`, `TestConfigCacheKeepsThePrevious`, `TestPendingConfigIsKeptApart`, `TestWriteAtomicLeavesNoTempFiles`, `TestOpenValidatesThePath`, `TestUnusableFilesAreReported` | keys, modes, atomic writes; the pending config beside the confirmed one, dropped by a re-join; the mesh view cached without its staleness reason, a damaged or invalid one refused; a file that cannot be read, replaced or removed is an error naming it, and a failed save leaves the cache as it was |
| `client` | `TestConfigHonoursTheETag`, `TestStatusRoundTrip`, `TestPrefsRoundTrip`, `TestMeshIsCheckedBeforeUse`, `TestNamesRoundTrip`, `TestErrorsCarryTheCode`, `TestEventsDeliversConfigEvents`, `TestEventsRefusesAnythingButAStream`, `TestJoinIsUnsigned`, `TestUnreachableAndMalformedAnswers` | against a server that verifies signatures; a mesh view that does not validate is never returned; `PUT names` and its refusal; `mesh` events reach their callback; a server that is down, a long error page (clipped), a 200 or an event that is not JSON |
| `localapi` | `TestStatusOverTheSocket`, `TestStatusErrorsAreReported`, `TestPrefsAreRelayed`, `TestDNSIsLocal`, `TestMeshView`, `TestLogs`, `TestUIToken`, `TestClientExplainsAMissingAgent`, `TestListenRemovesAStaleSocket`, `TestListenExplainsAnUnusablePath`, `TestServeReportsAClosedListener`, `TestClientExplainsUnexpectedAnswers` | the socket's mode, one agent per socket; a prefs request validated before it leaves the node, the server's refusal passed on, an unreachable server as 502; the mesh view with its staleness reason; the UI link or the reason there is none; `logs` as text, `follow=true` until the reader leaves (nil) or the agent stops (an error), a dropped reader told so, any other `follow` a 400; an unusable socket path; a foreign field, a status-only agent's missing prefs route, another program on the socket |
| `service` | `TestInstallWritesTheUnitAndStartsIt`, `TestInstallRefusesABinaryElsewhere`, `TestAvailable`, `TestNoWindowsServiceHere` | the unit and the systemctl argv, against a temp root; off Windows no process is a Windows service |
| `service` (Windows) | `TestCommandLineQuotesThePaths`, `TestAvailableNeedsElevation`, `TestHandlerStopsTheAgent`, `TestHandlerReportsAFailedRun` | the service's command line quoted as the service manager reads it; installing needs elevation, and a test is no service; Stop and Shutdown cancel the agent's run (Interrogate answered), a failed run exits 1 for the recovery to restart it -- the service manager itself is not driven by tests |
| `localapi` (Windows) | `TestPipeIsForAdministratorsOnly`, `TestStatusOverAPipe` | the pipe's DACL is Administrators and SYSTEM only, and the pipe carries it as shipped: unelevated, the client is refused with ACCESS_DENIED and the elevation hint (the pipe is left unserved, as go-winio's Accept needs the read and write the DACL refuses this process too), elevated it is answered; a second listener and a bad pipe name refused; over a pipe the test user may open, the API answers; a listener that cannot serve (Serve returns anything but a clean stop) fails the test |
| `state` (Windows) | `TestDataDirIsProtectedOnWindows` | the data dir's ACL is protected and a key file's inherited: SYSTEM and Administrators (plus the unelevated creator), never Users, Everyone or Authenticated Users; the creator still reads its key |
| `netconv` | `TestPrefixRoundTrip`, `TestPrefixUnmapsIPv4`, `TestAddrPort` | net/netip conversions |
| `localui` | `TestParseAddr`, `TestSignIn`, `TestSessionsAndLinksAreBounded` | `--ui-listen` takes loopback ip:port only; a `meshctl ui` link signs in once within 2 minutes, a used, expired, malformed or unknown one is refused, the cookie is HttpOnly, SameSite=Strict and not logged; sign-out ends the session; an idle hour or 12 hours end it; at most 8 links and 8 sessions, the oldest dropped |
| | `TestOnlyTheUIsOwnHostIsServed`, `TestChangesNeedTheUIsOriginAndCSRF` | any Host but 127.0.0.1:7443 or localhost:7443 is a 421 that reaches nothing, a link included; a change without a session, without or with a foreign, opaque or https Origin, without or with another session's CSRF token, or over 16 KiB is refused and never reaches the agent |
| | `TestDashboard`, `TestForms`, `TestLogsPage`, `TestTemplatesHaveNoScript`, `TestServeAndListen` | `testdata/dashboard.golden.html` (a laptop with an exit, its peers, the forms and a stale mesh view) and its headers (CSP, no framing, no-store); a fresh server's page; each form sends exactly its values, a bad value never leaves the page, a refusal shows its status and message, an unreachable server is 502; the log page escapes markup; no script, inline style or handler in any template; a real socket, a port in use, a closed listener |
| `logbuf` | `TestFileRotates`, `TestRingKeepsTheLastLines`, `TestFollow`, `TestSlowFollowerIsDropped` | the Windows service's log file rotates to one `.1` copy past its cap and appends after a restart; a rotation that fails is reported and the log goes on in the old file; the last N lines, oldest first, a line only once it is whole, a long line cut at 8 KiB; a follower gets the held lines then each new one until it leaves; one that stops reading is dropped after 256 lines and never holds up the log |

## `cmd/meshctl`

| Test | Pins |
| --- | --- |
| `TestCommandsDoc` | docs/commands.md equals the registry's rendering (`./scripts/dev.sh regen` rewrites it) |
| `TestEveryCommandHasHelp`, `TestHelpShowsFlagsAsTheDocsDo`, `TestGroupsNeedASubcommand`, `TestFlagsMayFollowPositionalArguments` | the command tree; help spells flags as the docs do (`--name`, a one-letter `-f`) with their kind and a non-zero default |
| `TestNodeLifecycle`, `TestNodeCommandsExplainAMissingServer` | every `node` command against a real admin socket (`node set --mesh-ip` with its note), `node compose` with both backends, for a server and with a hostile directory |
| `TestNodePhone` | `node add --role phone` prints the profile command, not a token; `node phone` prints the profile alone on stdout (the note on stderr) and `--json`; `node list` and `node show` of a phone; a server and a non-exit exit refused |
| `TestExportCommand`, `TestServerFlagErrors`, `TestAgentJoinTakesTheTokenFromTheEnvironment`, `TestHealthcheckServer`, `TestStatusAndAgentHealthcheck`, `TestPrintStatusOfALaptop`, `TestBytesize` | the other commands; the token never echoed; `status` of a laptop with nothing applied, an exit and errors; `server --hub-down-after` only 0 or 1s to 1h |
| `TestNodeShowAnEnrolledServer` | `node show` of a joined server: rows, peers, the computed config, a reported rollback; `node list` shows it rolled back |
| `TestAgentRunRuntimeSocket` (`agent_test.go`) | `agent run --runtime-socket` takes auto, off or a clean absolute path; a relative path, `..`, a metacharacter and an empty value are usage errors before the agent starts; `--status-every` only from 1s to 5m |
| `TestUICommand`, `TestAgentRunUIListen` (`ui_test.go`) | `meshctl ui` prints the link alone on stdout and its lifetime and the ssh hint on stderr; a refusal is exit 1; `agent run --ui-listen` refuses addresses off loopback before the agent starts and takes `off` |
| `TestAgentRunNeedsAJoin`, `TestAgentInstallIsForWindows`, `TestAgentLog` | `agent run` before a join says to join first and a bad `--log-level` is a usage error, neither touching the host; off Windows, `agent install` and `agent uninstall` point to install.sh and systemctl and touch nothing (on Windows they would change the host's real service, so the test skips there); the agent logs to stderr, or as the Windows service to `agent.log` in its data dir; an unusable data dir or a directory in the log's place refused |
| `TestLogsCommand` | `meshctl logs` prints the agent's ring as it holds it; an argument is a usage error; no agent is exit 1 |
| `TestDNSCommand` (`exit_test.go`) | `meshctl dns` shows the state, `on`/`off` switch it through the agent, usage errors never reach it, no agent is exit 1; the state names the zone the node serves |
| `TestExitCommand`, `TestExitReportsAnAgentThatDoesNotApply` (`exit_test.go`) | `meshctl exit` shows the exit and choices, switches and waits for the applied hash, `off` sends an empty exit, `--wait 0`; a refusal is exit 1 with the server's reason; usage errors never reach the agent; an agent that does not apply in time is reported with its error |
| `TestPrintEnrollmentInTheLocalZone` (`commands_test.go`) | a join token's expiry in the local time zone; no enrollment, nothing printed |
| `TestUserCommands` | passwords only from a terminal (a fake in tests), twice and matching, never echoed; add, passwd, reset-totp |
| `TestImportCommand` | import then export round trip; a file with unknown fields refused (Linux only) |
| `TestFirewallCommands` (`firewall_test.go`) | `node add --groups`; `firewall show` (policies, the rules every server enforces), `--yaml` and `--json`; `firewall apply` of a file; `firewall check` by node and address with the deciding rule, not mesh traffic, a service without one port; a bad file, a missing one and a dangling reference change nothing; `node show` lists the groups and `node set --groups` replaces them, a bad group refused |

## End to end -- `test/e2e`

Build tags `e2e linux`, root, `./scripts/dev.sh e2e`. Network namespaces on a bridge run the real
binary with kernel WireGuard -- the NAS spoke with the userspace backend -- and the server behind a
TLS proxy (a helper process of the test binary).

`TestServerMesh` checks: join -> two peers and traffic between every pair within 60 s; `node set
vps-2 --mesh-ip` -> vps-2 reports the new address and the others reach it there, not at the old
one; `node disable` -> the node leaves both other servers within 30 s and is unreachable; server stopped, an
agent killed and its `wg0` deleted -> the restarted agent rebuilds `wg0` from its cache and reaches
the other server.

`TestContainerNetworks` (`containers_test.go`, amendment A37) adds box in container mode -- its
agent in a namespace whose second interface is a port of a bridge in a "host" namespace holding
.1, with an attached container -- and pc in host mode with its containers on the mesh, whose
agent makes `meshctl0`. It checks: box's gateway claims .2 and installs its rules; box's container
reaches vps-1 and is reached from vps-2, and resolves `vps-1.mesh` at the gateway; box's host
reaches no mesh address and vps-1 cannot reach the host's .1; box's agent reads a fake runtime
socket whose container is `web` (amendment A38), and vps-2's resolver answers `web.mesh` and
`web.box.mesh`, the gateway `web.mesh`, with the container's address (every other agent runs with
`--runtime-socket off`, so none reads the host's runtime); pc's agent gives `meshctl0` its .1;
vps-1, box's container (relayed by vps-1) and pc's host reach pc's container; a LAN host routing
the subnet through pc does not (the guard); `node set pc --containers=false` takes the guard away
and leaves the bridge.

`TestRelayFailover` (`failover_test.go`, amendment A40) runs the agents with `--status-every 1s` and
the server with `--hub-down-after 3s`: the laptop (hub vps-2) and the NAS (hub vps-3) talk through
vps-2; its agent stops and its wg0 goes, and the pair moves to vps-3 and keeps talking while the
node list shows vps-2 down; its agent comes back, the spokes re-handshake with its new wg0 (their data
to it unanswered for 15 s), and the pair returns to vps-2 and talks through it.

`TestSpokesAndExit` adds a laptop (hub vps-1), a NAS (hub vps-3) and an "internet" namespace that
routes for every node and runs an echo service answering the caller's address; vps-2 and vps-3
are made exit-capable and get the masquerade firewalld would add (`nft`, since the firewall is off
in e2e). It checks: the two spokes ping each other across their different hubs; with exit off the
internet sees the laptop; after `meshctl exit vps-2`, then `vps-3`, it sees that server within
5 s; a TCP stream from the laptop to vps-1's mesh address carries data through both switches;
`meshctl exit off` brings the laptop's own address back and leaves no policy rule behind. Every node
answers the mesh zone: the laptop resolves `nas.mesh` and `vps-3.mesh` at its own resolver, the
NAS `laptop.mesh` at vps-1's; `example.com` is REFUSED; both spokes report DNS `unavailable`, since
an agent in a lab namespace must not register with the host's systemd-resolved. With the server
and the laptop's agent stopped and its `wg0` deleted, the restarted laptop resolves `nas.mesh`
and reaches the NAS from its cache. A phone (hub and exit vps-2) brings its `node phone --exit vps-2`
profile up the way the WireGuard app does (`wgQuick`: wgctrl, address, a route per allowed IPv4
prefix): its DNS is the hub's mesh IP, it pings the NAS through its hub, resolves `nas.mesh` at
the hub and reaches the internet as vps-2; a server's heartbeat then reports its handshake. The
laptop's mesh view (`GET /v1/mesh`) lists all six nodes; after the restart with the server down
it is the cached view with the reason, and `meshctl logs` shows the agent's start. Before that,
`meshctl ui` on the laptop prints a link that signs a browser (dialing inside the laptop's
namespace) into its local UI, whose dashboard lists the mesh; a post without the UI's Origin is a
403. The NAS
runs `--wg-backend userspace`: its `wg0` is a tun with the `meshctl` alias, a UAPI "get" on
`/var/run/wireguard/wg0.sock` (what `wg show` does) lists its three peers, and the NAS, the laptop
and the server's `node show` report `userspace` and `kernel`. The test's sockets are opened inside
the namespaces from the test process (`inNS`).

`TestMeshFirewall` (`firewall_test.go`, amendment A43): every server's namespace holds the table
`ip meshctl` with its fingerprint comment and allow-all; after `meshctl firewall apply` of a policy
denying the laptop TCP to the NAS, an open laptop-NAS connection stops at its next packet (the
relay deleted its connection tracking) and a new one is refused, ping still crosses (a system
rule), `meshctl firewall check` names the policy, and vps-1 reports no error; allow-all alone
lets new connections through again. Every other e2e test runs with the default allow-all table on
every server.

## `internal/server/planner`

Goldens: `testdata/{servers,full,degraded}.golden.json`, derived by hand from the rules in
`planner.go` (see [developer.md](developer.md#goldens)); the full topology carries custom records
and the mesh firewall of `meshFirewall()` (`segment_test.go`). The fixtures are built by `topology()`
in `planner_test.go` from the testfix reference topology.

| Test | Pins |
| --- | --- |
| `TestComputeGolden` | full output for three servers; the full topology (exit, hubs, phone); the degraded one (disabled laptop, endpointless vps-3) |
| `TestComputeIsDeterministic` | output is independent of input order |
| `TestComputedConfigsPassAgentValidation` | every emitted config passes `proto.NodeConfig.Validate` |
| `TestServerPeersForSpokesAreExactly32` | on a server, a spoke or phone owns only its /32 and has no endpoint or keepalive |
| `TestEveryPairCanTalk` | packet simulation with WireGuard's longest-prefix cryptokey routing: every pair of planned nodes reaches each other both ways (amendment A3) |
| `TestContainersTalkToEveryNode` | the same simulation for a container behind each node with a container network: it reaches every other node and is reached from it (amendment A37) |
| `TestContainerNetworks` | each peer routes a container subnet with the node's /32 and nothing more; forward_for (every container network on every resolver), the `containers` section, the sysctls of host and container mode, no firewall section in container mode (the gateway isolates by policy routing), no exit in container mode |
| `TestSpokesWithDifferentHubsShareOneRelay`, `TestExitRouting`, `TestRelayAndExitServersForward` | relay choice, exit routing, ip_forward and firewall policy flags |
| `TestUnreachableHubFallsBack`, `TestEndpointlessServerWarns` | fallbacks and their actionable warnings |
| `TestDNS` | the zone (every active node and custom record, an alias of a disabled node dropped), upstreams in order, split off only in exit mode, `forward_for` only on the phone's hub; the default zone in every config, another set for the mesh in every config; the extra domains in every resolver's section, sorted, and none without them |
| `TestFirewallRules` | every server's mesh zone holds the accept-all rule from the mesh and udp/51820 in the public zone, and the one compiled policy, a container-mode server too; no spoke or phone has a policy (amendment A43) |
| `TestPolicyCompiles`, `TestPolicyMerges` (`segment_test.go`) | the compiled policy, derived by hand: system rules (ping, DNS to the servers, Swarm between them), the enabled policies in order with references resolved (0.0.0.0/0 clipped to the mesh, a group, a role, a zone, a service object), allow-all, each allowing zone; it passes the agent's check; a mesh without a firewall keeps the system rules and is warned; addresses and ports merged (touching ranges joined, a network read as one, every port winning) |
| `TestCheckSegmentation`, `TestSegmentationWarnings` | a reference to no node, group, zone or object, a bad role, a group of strangers, bad literals, services, object entries, intra and actions -- a disabled policy's references too -- block the plan; a container network a node does not have, an address outside the mesh, an empty zone and a policy that matches nothing are warnings, and the empty ones are left out of the policy |
| `TestSummarize`, `TestOverlaps` | the zone matrix's verdicts (allow, partial, deny) with the rules that decide them, without allow-all too; whether a rule side reaches a zone (any reaches every non-empty zone) |
| `TestServerIPs` | a Swarm service's resolvers: the mesh IPs of the servers in peer lists, by id (a server without an endpoint is in none) |
| `TestRelayFailover` | with any one server down, in the full and the containers topologies, every pair still gets through and no pair of spokes relays through the down server; the choice is symmetric; a down hub without a phone on it stops relaying and is warned about; the default hub before the lowest-id server; all down moves nothing |
| `TestCheckRejects`, `TestProblemString` | every blocking constraint (custom records and upstreams included), and that a plan with errors yields no configs; an invalid zone and a record name ending in the zone; an extra domain inside the zone, and two that overlap |

`alloc_test.go`: `TestRangeFor`, `TestBlockMatchesTheSpecAddressPlan`, `TestAllocateLowestFree`, `TestAllocateContainers`,
`TestCheckRequested`, `TestCheckMeshCIDR` -- the S4 address plan and its limits.

## `internal/fsutil`

`TestWriteAtomicReplacesAndLeavesNoTempFiles`: the atomic write both sides use.
`TestWriteAtomicReplacesAFileThatIsOpen`, `TestRenameRetrying`, `TestTransientRename`: on Windows
a reader holding the target open (every `os.Open` shares no delete access) makes the rename fail;
it is retried for up to 2 s, any other failure returns at once.

## Coverage

`./scripts/dev.sh cov` totals every package, including the adapters the unit tests cannot reach
by design: `internal/agent/wg/{kernel,link,userspace}_linux.go`,
`internal/agent/routes/netlink_linux.go`, `internal/agent/exitmode/policy_linux.go`, the privileged half of
`internal/agent/policy/nft_linux.go` (netlink,
wgctrl, wireguard-go's tun) and `internal/agent/dns/resolved_linux.go` (systemd-resolved, which a
test must not change on the host it runs on). Their unprivileged paths (opening a backend, a refused create) are
unit-tested; the rest needs root and the wireguard module, and is exercised by
`./scripts/dev.sh e2e` (`TestServerMesh`, and `TestSpokesAndExit` for exit mode and the userspace
backend).

## `internal/tools/fetchwintun`

`TestFetchRefusesAnythingButThePinnedRelease`, `TestFetchKeepsAMatchingDLL`,
`TestExtractChecksTheDLL`, `TestFetchWritesACheckedDLL`, `TestRun`, `TestDownload`: a download
that is not the pinned zip, or none at all, writes nothing; a DLL already in place that matches is
kept without a download; inside a zip that matches, a DLL that is missing, too large or not the
pinned one is refused; one that matches is written, and an unwritable dir is the error; bad flags
are exit 2, an unpinned download exit 1; the download reads up to its size limit and reports a
failed status, an oversize body and an unreachable server.

## `internal/tools/vulnjudge`

Copied from solace-cnt-scripts with its tests and govulncheck JSON fixtures in `testdata/`:
fatal on a fixable called vulnerability, warning on one without a released fix, quiet on
findings the code does not call.
