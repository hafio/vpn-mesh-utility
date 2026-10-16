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

`testfix` imports only `internal/proto`, so tests on both sides of the server/agent boundary use
it; `TestModuleBoundaries` forbids production code from importing it.

## `internal/validate`

`validate_test.go`

| Test | Pins |
| --- | --- |
| `TestName`, `TestDNSName`, `TestHostnameFoldsCase`, `TestUsername`, `TestEnum` | name grammars; the shared `hostile` payloads (metacharacters, quotes, newlines, NUL, traversal) are refused |
| `TestIPv4`, `TestPrefix4`, `TestPortAndRange`, `TestEndpoint`, `TestKey32` | address, prefix (host bits rejected with a suggestion), port, endpoint and key shapes |
| `TestTokenMasksTheValue`, `TestLocalTokenMasksTheValue` | a malformed join or local UI token never appears in the error; the error is a `*validate.Error` |
| `TestVersionAndSemVer`, `TestServerURL`, `TestPOSIXPath`, `TestFilePath`, `TestText` | version strings, https-only origins (http only on loopback), artifact paths, DSN-safe paths, control-character-free text |
| `TestArg` | the shell-layer charset accepts firewall-cmd options and rich rules and refuses metacharacters |
| `FuzzArg`, `FuzzName` | nothing accepted contains a byte outside the documented charset |

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
| | `TestValidateAcceptsFixtures`, `TestValidateRejects`, `TestValidateReportsEveryProblem` | the agent-side config gate, which every fixture passes (server, spoke, spoke with an exit): version, addresses, MTU, fwmark, peer names, duplicate peers/keys/prefixes, endpoints, keepalive, exit and its choices, sysctl allowlist, firewall fields, DNS records, upstreams and forward_for |
| `api_test.go` | `TestJoinRequestValidateNormalises`, `TestJoinRequestValidateRejects`, `TestJoinRequestErrorNeverEchoesTheToken` | join payload validation (including `public_ip`) and token masking |
| | `TestETagQuotesTheHash` | the config entity tag |
| | `TestStatusReportValidate` | heartbeat limits (peers, errors), counters, degraded-step allowlist, the DNS states |
| | `TestPrefsRequestValidate` | a prefs request: exit name grammar (metacharacters refused), keepalive and MTU ranges, an empty request refused |
| `mesh_test.go` | `TestMeshViewValidate` | the mesh view the agent caches and shows: names, roles, IPv4 mesh IPs, hub and exit grammar, the health colours, a state without control characters and at most 200 bytes, no negative time or version, at most 65536 nodes |
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
| `TestOpenAppliesMigrations`, `TestOpenIsIdempotent` | schema v1 and default settings; reopening is a no-op |
| `TestOpenRefusesNewerSchema` | a database newer than the binary is refused with an upgrade hint |
| `TestOpenRejectsDSNInjection` | a path carrying `?` cannot rewrite the driver DSN |
| `TestLoadMigrationsValidatesNames` | gaps, zero and malformed migration names fail at startup |
| `TestNodeCRUD`, `TestNodeUniquenessIsAConflict`, `TestNotFound` | node accessors; uniqueness violations map to `ErrConflict` |
| `TestDeleteNodeCascades` | deleting a node removes its tokens, bound records and rules, peer status, and clears spokes' hub references |
| `TestTokenLifecycle` | single use, unknown, expired, reissue revokes the previous token, tokens stored only as hashes |
| `TestTokenConcurrentConsume` | eight concurrent joins with one token: exactly one succeeds |
| `TestVersions`, `TestAudit` | version numbering, listing and lookup; audit rows carry the transaction clock |
| `TestSnapshotRestoreRoundTrip` | restore after drift (node removed, added, edited; setting and rules changed) reproduces the snapshot exactly, keeping ids |
| `TestDecodeSnapshotRefusesUnknownShapes` | unknown fields and formats are refused |
| `TestRecordStatus` | a heartbeat updates the node and replaces (not merges) its peer rows |
| `TestPhoneState` | `LastHandshakeWith` is the latest handshake any node reported with a peer (0 for none); `SetApplied` records a phone's profile without counting as a heartbeat |
| `TestOpenReadOnly` | `export`'s read-only open reads, refuses writes, a missing file (pointing at `--data`) and a schema other than the binary's |
| `TestUsers`, `TestSessions` | account and session accessors; `UseTOTPStep` refuses a step at or before the last one (replay), and a new secret starts the counter over |

## `internal/server/auth`

Tests use a fresh store, `testfix.Clock` and cheap argon2 parameters (one test pins the real
ones).

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
| `TestEditSpoke` | a spoke's hub, exit and keepalive edit; the hub and exit choices offered; refused edits (400/404) change nothing; no token for a disabled node; the computed config on a server's page |
| `TestDNSAndFirewallPages` | the zone with automatic names and the upstreams; a custom record added and removed; a record named like a node (409) and a refusal without CSRF (403) changing nothing; each server's rules with their source; a custom rule added and removed; a bad port (400) and a missing rule (404) coming back on the page with the form kept |
| `TestSpokeComposeDownload` | a spoke's page offers its compose file (a server's does not); the download carries the form's backend and directory as `compose.yml`; a server (409) and a hostile directory (400) come back on the page; no download without a session |
| `TestStoreFailuresAreLoggedNotShown` | with the mesh store failing (`faultyDB`), every page and change answers 500 without the cause, the cause reaches the log, and nothing is written |
| `TestRollbackThatAddsNodesNeedsConfirmation` | the version diff, the confirmation for a membership change, the new version, an unknown version (404), a rollback to the current version changing nothing |
| `TestSettings`, `TestUsersAndAccount`, `TestDashboardFragmentAndExport` | validation of every setting (CIDR only while empty), user management (not on yourself), password change ending the session, the refresh fragment, the export download |
| `TestHTMXIsPinned`, `TestTemplatesHaveNoInlineScriptOrStyle` | the vendored htmx's sha256; no inline script, style or event handler in any template; no static directory listing or path escape |
| `TestPageGoldens` | `testdata/{login,dashboard,node,version}.golden.html` for fixed data, the node's DNS state included (`./scripts/dev.sh regen` creates and updates them) |
| `TestEveryPageRenders`, `TestHelpers` | every template renders; ages (never, a clock ahead of the report shows 0s), byte sizes and the origin check (the dashboard colours are `api`'s `TestDashboardHealth`) |
| `TestPhonePage`, `TestQRCode` (`phone_test.go`) | a phone added from the nodes page (no token, no compose, no peers table, no exit setting); its via-exit profile shown once as an inline SVG QR code and text, `no-store`, the private key never in the log; a server that is no exit (409) refused on the page; a hub change marks the profile out of date; audited `phone.profile` as `user:admin`. The QR path draws exactly the library's dark modules with a 4-module quiet zone, and a too-long profile's error does not quote it |

## `internal/server/versions`

`TestCompareEqualSnapshotsIsEmpty`, `TestCompareListsEveryKindOfChange`,
`TestARenameIsAChangeNotAReplacement`: added, removed and changed nodes by field (keys as
fingerprints), settings, DNS records and firewall rules; nodes are matched by id.

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
| | `TestAdminCompose` | a spoke's compose file with the server's release and origin, both backends and a chosen directory; a server (409), an unknown node (404), a bad name, backend or directory (400) and a development build refused |
| | `TestCustomDNSRecords` (`records_test.go`) | an alias and a fixed record reach every config; an alias of a node not on the mesh changes no config but shows as not live; the DNS page's rows; bad names, both or neither target, a bad IP: 400 before any transaction; a duplicate, a node's name, an IP outside the mesh (409) and an unknown node (404) write nothing; removal; audited `dns.add`/`dns.rm` |
| | `TestCustomFirewallRules` (`records_test.go`) | a rule for every server and one for vps-1 rendered with their notes; every bad field at once (400, no transaction); a spoke (400) and an unknown server (404) refused; removal; audited `firewall.add`/`firewall.rm` |
| | `TestAdminErrorCodes`, `TestAdminSetClearsReferences` | every refusal of every admin write (404, 400, 409) leaves state untouched, a phone's profile included (a server, a bad, unknown or non-exit exit, an unknown field); clearing an endpoint warns |
| `phone_test.go` | `TestPhoneProfile` | a phone added without a token; a mesh-only profile whose private key matches the public key all three servers peer with, the hub's resolver as DNS; a via-exit profile replaces the key and adds `0.0.0.0/0, ::/0` to the exit's peer; no audit row carries a private key; a hub change marks the profile out of date; the servers' reported handshake in `node list` and `node show` |
| | `TestPhoneProfileNeedsAServer` | without a joined server with an endpoint: 409, the phone keeps no key and nothing is audited |
| `health_test.go` | `TestDashboardHealth` | the dashboard colour and label the web UI and the mesh view share: disabled, not joined, no or a silent heartbeat, a rollback, degraded, applying, a late heartbeat; a phone by its profile and its servers' handshakes; a phone is seen by its handshake |
| `reconcile_test.go` | `TestReconcilerVersionsOnlyRealChanges` | a no-op writes no version; a change notifies exactly the nodes whose config changed; the snapshot is stored (amendment A16) |
| | `TestReconcilerRollsBackAFailedMutation`, `TestPlanErrorListsTheProblems`, `TestLoadInputRefusesBrokenSettings` | rollback, the 409 mapping, broken settings |
| | `TestRollbackIsANewVersion` | restoring version N writes a new version recording N, notifies the changed nodes, writes nothing when the state already matches, 404 for an unknown N |
| `guard_test.go` | `TestReplayCache`, `TestReplayCacheRefusesWhenFullOfLiveEntries` | the replay window, refusing rather than evicting |
| `hub_test.go` | `TestHubLatestWins`, `TestHubReplacesKicksAndCloses` | latest-wins delivery; one stream per node; kick and shutdown |
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
`TestToSnapshotRefusesHostileDocuments` (every field validated before an import may restore it).

## `deploy`

| Test | Pins |
| --- | --- |
| `TestUnitPosture` | the unit's `[Service]` section exactly -- any added, removed or widened setting fails |
| `TestInstallGolden`, `TestInstallChangesOnlyTheMarkers`, `TestInstallParses` | `testdata/install.golden.sh`; rendering changes only the four quoted markers; `sh -n` accepts it |
| `TestInstallRefusesHostileValues` | quotes, paths, `$(...)`, newlines, bad checksums are refused before rendering |
| `TestServerComposePosture`, `TestDevComposePosture` | the exact service keys, user, no privileges, no capabilities, read-only root, named volumes only, no widening token anywhere in the file |
| `TestNASComposeGolden` | `testdata/compose.nas-{userspace,kernel}.golden.yml`; empty backend and directory are the userspace variant in `/volume1/docker/meshctl`, and 1.2.3 is the same image as v1.2.3 |
| `TestNASComposePosture` | both variants' exact service keys and values: uid 0, host network, NET_ADMIN and NET_BIND_SERVICE only, the tun device in userspace only, one bind mount; no other widening token |
| `TestNASComposeRefusesHostileValues`, `TestNASRendererHoldsItsStructure` | newlines, `${...}`, `---`, `: `, bad names, origins, backends and development builds refused before rendering; the renderer fed them directly keeps every key, the posture and the comment block, and doubles `$` |

## Agent (`internal/agent/...`)

`internal/agent/agentfake` holds in-memory fakes of the host (`Host` for links, addresses and
WireGuard; `Routes`; `Policy` for exit mode's rules and table routes; `Sysctl`; `Firewall`, which
counts the drift check's `Recheck`; `DNS`, which answers the state a host with or without
systemd-resolved would) that apply changes as the kernel would and count every mutation.
Production code must not import it (`TestModuleBoundaries`). Its own tests
(`TestConfigureHasKernelSemantics`, `TestLinksAddressesAndFailures`, `TestRoutesSysctlAndFirewall`,
`TestPolicyHasKernelSemantics`, `TestDNS`) pin that fidelity -- append versus replace of AllowedIPs,
removals, copies, EEXIST and ENOENT -- because every idempotency test depends on it. `Host.SetFail`
and `Policy.Fail`/`FailOn` inject a failure into one method, also while an agent runs.

| Package | Tests | Pins |
| --- | --- | --- |
| `agent` | `TestRunExplainsRefusals`, `TestIsPublic` | `node_disabled` and `clock_skew` logged with what to do while the cached config stays up; public-address detection |
| | `TestRunFollowsTheServer`, `TestRunPollsWhenTheStreamIsSilent`, `TestRunBootsFromTheCacheWithoutTheServer`, `TestRunKeepsTheOldConfigWhenANewOneFails`, `TestRunNeedsAnEnrollment` | the run loop against a fake server that verifies every signature: first apply and heartbeat, pushed and polled changes, the cache, boot without the server, a refused config reported and not cached |
| | `TestJoin`, `TestJoinValidatesBeforeCallingTheServer`, `TestCleanHostname`, `TestOneLine` | the join request, enrollment, service install or the reason it is skipped, cache dropped on a node change, keys kept; status text always passes the server's validation |
| | `TestRunRefusesAnUnusableDataDir`, `TestRunIgnoresDamagedConfigFiles`, `TestRunReportsUnreadablePeerStatus`, `TestReportKeepsToTheServersLimits`, `TestDefaultSocket` | a bad data dir path, a damaged enrollment, missing keys, an unknown `--wg-backend` or an unusable socket stop Run before the host is touched; a damaged cached or pending config is ignored (the pending one removed, no rollback reported); unreadable peer status is a reported error; a report keeps at most 32 errors |
| | `TestReportNamesTheBackend` | the heartbeat names the backend in use; a fake host stands for the kernel unless the test names userspace |
| | `TestDNSOnAndOffStayOnTheNode`, `TestDNSSettingThatCannotBeSaved` | `POST /v1/dns` reaches the DNS step at once, is reported, is kept in `local.json` across a restart and never reaches the server; the resolver stops with the agent; a setting that cannot be saved is a 500 that applies nothing |
| | `TestDNSFailureIsReported`, `TestDriftCheckPutsBackWhatChanged` | a failing DNS step degrades the node with its error and the tunnel up; the drift check rebuilds a deleted `wg0` without a new config, after the firewall was asked to look again |
| | `TestPrefsNeedTheServer` | `meshctl exit` with the server down is a 502 naming the call, and exit mode stays off |
| | `TestExitThroughTheLocalAPI` | `POST /v1/prefs` reaches the server signed, the agent fetches the new config without an event, exit mode comes on, the status shows the exit and its choices; a server refusal comes back with its status |
| | `TestMeshViewIsCached`, `TestMeshViewIsRefreshed` | the mesh view fetched at start into `mesh.json`; under 10 s old it is shown without asking again, unless a new config arrived; a restarted agent with the server down shows the cached view and the server's reason; refreshed every `MeshEvery` |
| | `TestLogsAreServed` | `GET /v1/logs` reads the ring the agent's log writes to |
| | `TestLocalUIThroughTheAgent`, `TestLocalUIOffOrUnavailable`, `TestRestartNeedsAConfig` | `POST /v1/ui-token`'s link signs a real browser into the UI, whose dashboard shows the node and whose restart drops and re-adds both peers; the status carries the MTU; without a UI, or with its port taken, the agent runs and the link is refused with the reason; a UI address off loopback stops Run before the host is touched; a restart before any config is refused |
| `agent` (`safeapply_test.go`) | `TestSafeApplyStateMachine` | the state machine on a fake clock: a report sent before the apply does not confirm, the window, a pending config replaced keeps the confirmed one as prev, disarm |
| | `TestSafeApplyConfirmsWhenTheServerAnswers`, `TestSafeApplyRollsBackWithoutAnAnswer`, `TestSafeApplyRecoversAPendingConfigAtBoot`, `TestSpokesDoNotArmSafeApply`, `TestSafeApplyDisarmsWhenTheServerGoesBack`, `TestSafeApplyReportsAFailedRollback` | the runner with a 1 s window on a fake clock: confirm moves the cache; no answer puts the previous config back, reports the rollback and never re-applies the rejected version, while the next version is applied and confirmed; a pending config found at boot is rolled back the same way; spokes cache at once; the server going back to the confirmed config disarms without a rollback; a rollback that fails too is reported as an error with the rollback |
| `apply` | `TestApplyIsIdempotent` | the second apply of the same config changes nothing on the host |
| | `TestApplyValidatesBeforeTouchingTheHost`, `TestApplyStopsAtTheFailingStep`, `TestFirewallFailureOnlyDegrades`, `TestRemovedPeerLeavesDeviceAndRoutes`, `TestStatsNamesKnownPeers` | step order and failure modes |
| | `TestRestartRenewsEveryPeer` | restart drops both peers and adds them back (two Configure calls, routes untouched) after rechecking the firewall; a missing wg0 is rebuilt, a failed drop is the peers step's error, a foreign wg0 is refused with nothing changed |
| | `TestDNSFailureOnlyDegrades`, `TestDNSFollowsDNSOffAndExitMode`, `TestReassertPutsBackWhatDrifted` | step 9 gets the mesh IP, the config's DNS, exit mode and `dns off`, and only degrades; `Reassert` rechecks the firewall, changes nothing on an unchanged host and rebuilds a deleted interface; `Close` stops the DNS step |
| | `TestExitModeOnSwitchOff`, `TestExitModeFailureIsFatal` | exit on installs the rules, a switch moves 0.0.0.0/0 between peers without touching them, off removes them, 0.0.0.0/0 never in the main table; a policy failure is fatal; an engine without policy routing refuses exit on |
| `exitmode` | `TestDesiredIsWgQuicksExitRouting`, `TestSyncTurnsExitModeOnAndOff`, `TestSyncReplacesAStaleRuleOfItsOwn`, `TestSyncStopsAtTheFirstFailure`, `TestSyncReportsEachFailure` | the exact rules and routes (IPv4 and IPv6 blackhole), routes before rules on, rules before routes off, idempotent, a stale own rule replaced, exit off on a clean host touches nothing; a failure of each call stops Sync there and counts the changes before it |
| | `TestKernelShapes`, `TestKernelPolicyWithoutPrivilege` (Linux) | the netlink rule and route each one becomes (protocol 201, inverted fwmark, blackhole); unprivileged listing works and every change is refused with what it attempted (skipped as root; the kernel path itself is e2e's) |
| `wg` | `TestDiffChangesOnlyWhatDiffers`, `TestDiffLeavesARoamingDialInPeerAlone`, `TestSyncIsIdempotent`, `TestSyncLinkRefusesAForeignInterface`, `TestSyncSurfacesHostFailures`, `TestParseKey`, `TestDesiredFromServerConfig` | the device diff and the interface steps (amendment A18) |
| | `TestUAPISetGolden`, `TestParseUAPI`, `TestParseUAPIRejectsMalformedValues` | the userspace backend's text protocol: `testdata/uapi-set.golden.txt` for a change of every kind (an empty change renders nothing); a device's "get" answer read back with peers sorted and a zero handshake kept zero, so a re-apply diffs to nothing; malformed values named by line and key, a key's value never quoted |
| | `TestOpenRefusesUnknownBackends`, `TestUserspaceWithoutADevice` (Linux) | `--wg-backend` takes only auto, kernel and userspace; the userspace host with no device: no interface is ours, reads and writes refused, wireguard-go's log mapped to debug and warn (creating the tun is e2e's) |
| | `TestOpen`, `TestCreateLinkNeedsPrivilege` (Linux) | each backend opens unprivileged and is the one reported (auto: what DetectBackend finds; no /dev/net/tun names the device); without CAP_NET_ADMIN either CreateLink fails with the interface and a hint, leaving no interface or device (skipped as root) |
| `dns` | `TestZoneAnswers`, `TestNewZoneRejects` | A records (the name's case kept), NODATA for another type and for an empty non-terminal, NXDOMAIN with an SOA, the apex SOA; other zones REFUSED in split mode, forwarded in exit mode and for a `forward_for` phone; another class REFUSED, two questions FORMERR; a section that does not parse |
| | `TestServiceAnswersAndRegisters` | on a loopback socket against a loopback upstream: the zone over UDP and TCP, REFUSED then forwarded after exit mode came on, SERVFAIL with the upstream gone; the registration (`~mesh`, plus `~.` and default-route in exit mode) set once and left alone on a repeat; `dns off` takes it back and the resolver keeps answering; no DNS section stops both |
| | `TestServiceWithoutAResolverToRegisterWith`, `TestServiceRegistrationFailuresAreReported`, `TestServiceBinding`, `TestSameRegistration` | no registrar or resolved not running: `unavailable` (`off` when turned off) and nothing written; a failing read, set or revert named, with the resolver still answering; a taken port and a foreign address reported before any registration; a resolver that died is started again; registrations compared as sets, `mesh.`/`MESH`/`""`/`.` normalised |
| | `TestClassifyBusErrors`, `TestRegistrarNeedsTheInterface` (Linux) | resolved's absence on the bus is `unavailable`, a refusal names CAP_NET_ADMIN; a missing interface fails before anything reaches resolved |
| `routes` | `TestDesiredSkipsTheDefaultRouteAndDuplicates`, `TestSyncAddsThenRemovesAndIsIdempotent` | routes follow AllowedIPs, never 0.0.0.0/0 |
| `firewall` | `TestFreshHostArgvGolden` | `testdata/fresh.golden.jsonl`: the exact firewall-cmd argv for a fresh host |
| | `TestApplyIsIdempotent`, `TestRemovedRuleIsRemovedBothWays`, `TestNilFirewallRemovesOnlyOurRules` | state-file diffing, runtime and permanent |
| | `TestHostileRulesNeverReachFirewallCmd` | a hostile rule runs no command at all |
| | `TestForwardingPolicies` | `testdata/policies.golden.jsonl`: the relay and exit policies created --permanent and one reload; not asked again while unchanged; a restarted agent finds them and changes nothing; one edited by hand is put back whole; one no longer wanted is deleted |
| | `TestMissingFirewalldIsActionable`, `TestOldFirewalldIsRefused`, `TestPolicyFailureIsReported`, `TestAtLeast`, `TestDamagedStateIsActionable` | failure messages; a failed policy is tried again on the next apply |
| | `TestRecheckPutsBackWhatDrifted`, `TestRecheckAsksAboutARuleSpelledDifferently` | after `Recheck` an unchanged host changes nothing; wg0 moved out of the zone at runtime and a rule removed by hand are put back (runtime binding only, the rule both ways); a rule listed in other words is confirmed with `--query-rich-rule`; without `Recheck` nothing is asked |
| | `TestEachFailingCallStopsTheApply` | each firewall-cmd call of a fresh host with a hand-edited policy, failed in turn: the apply returns that error and runs nothing after it |
| `execrun` | `TestCheckAllowsFirewallArgv`, `TestCheckRefuses`, `TestRunRefusesBeforeExecuting`, `TestCappedKeepsTheHead` | the allowlist and the argv gate |
| | `TestRunCapturesOutputWithAMinimalEnvironment`, `TestRunFoldsStderrIntoTheError`, `TestRunTimesOut`, `TestRunReportsAMissingProgram` | the real exec path against fake `firewall-cmd` / `systemctl` scripts on a temp PATH (not on Windows) |
| `sysctl` | `TestApplyChangesOnlyWhatDiffers`, `TestApplyRefusesUnknownKeysBeforeTouchingAPath`, `TestApplyReportsAMissingParameter` | against a fake `/proc/sys` |
| `state` | `TestMeshCache`, `TestLocalSettings`, `TestKeysAreCreatedOnceAndKept`, `TestLoadKeysRefusesLooseOrDamagedFiles`, `TestEnrollmentRoundTrip`, `TestConfigCacheKeepsThePrevious`, `TestPendingConfigIsKeptApart`, `TestWriteAtomicLeavesNoTempFiles`, `TestOpenValidatesThePath`, `TestUnusableFilesAreReported` | keys, modes, atomic writes; the pending config beside the confirmed one, dropped by a re-join; the mesh view cached without its staleness reason, a damaged or invalid one refused; a file that cannot be read, replaced or removed is an error naming it, and a failed save leaves the cache as it was |
| `client` | `TestConfigHonoursTheETag`, `TestStatusRoundTrip`, `TestPrefsRoundTrip`, `TestMeshIsCheckedBeforeUse`, `TestErrorsCarryTheCode`, `TestEventsDeliversConfigEvents`, `TestEventsRefusesAnythingButAStream`, `TestJoinIsUnsigned`, `TestUnreachableAndMalformedAnswers` | against a server that verifies signatures; a mesh view that does not validate is never returned; a server that is down, a long error page (clipped), a 200 or an event that is not JSON |
| `localapi` | `TestStatusOverTheSocket`, `TestStatusErrorsAreReported`, `TestPrefsAreRelayed`, `TestDNSIsLocal`, `TestMeshView`, `TestLogs`, `TestUIToken`, `TestClientExplainsAMissingAgent`, `TestListenRemovesAStaleSocket`, `TestListenExplainsAnUnusablePath`, `TestServeReportsAClosedListener`, `TestClientExplainsUnexpectedAnswers` | the socket's mode, one agent per socket; a prefs request validated before it leaves the node, the server's refusal passed on, an unreachable server as 502; the mesh view with its staleness reason; the UI link or the reason there is none; `logs` as text, `follow=true` until the reader leaves (nil) or the agent stops (an error), a dropped reader told so, any other `follow` a 400; an unusable socket path; a foreign field, a status-only agent's missing prefs route, another program on the socket |
| `service` | `TestInstallWritesTheUnitAndStartsIt`, `TestInstallRefusesABinaryElsewhere`, `TestAvailable` | the unit and the systemctl argv, against a temp root |
| `netconv` | `TestPrefixRoundTrip`, `TestPrefixUnmapsIPv4`, `TestAddrPort` | net/netip conversions |
| `localui` | `TestParseAddr`, `TestSignIn`, `TestSessionsAndLinksAreBounded` | `--ui-listen` takes loopback ip:port only; a `meshctl ui` link signs in once within 2 minutes, a used, expired, malformed or unknown one is refused, the cookie is HttpOnly, SameSite=Strict and not logged; sign-out ends the session; an idle hour or 12 hours end it; at most 8 links and 8 sessions, the oldest dropped |
| | `TestOnlyTheUIsOwnHostIsServed`, `TestChangesNeedTheUIsOriginAndCSRF` | any Host but 127.0.0.1:7443 or localhost:7443 is a 421 that reaches nothing, a link included; a change without a session, without or with a foreign, opaque or https Origin, without or with another session's CSRF token, or over 16 KiB is refused and never reaches the agent |
| | `TestDashboard`, `TestForms`, `TestLogsPage`, `TestTemplatesHaveNoScript`, `TestServeAndListen` | `testdata/dashboard.golden.html` (a laptop with an exit, its peers, the forms and a stale mesh view) and its headers (CSP, no framing, no-store); a fresh server's page; each form sends exactly its values, a bad value never leaves the page, a refusal shows its status and message, an unreachable server is 502; the log page escapes markup; no script, inline style or handler in any template; a real socket, a port in use, a closed listener |
| `logbuf` | `TestRingKeepsTheLastLines`, `TestFollow`, `TestSlowFollowerIsDropped` | the last N lines, oldest first, a line only once it is whole, a long line cut at 8 KiB; a follower gets the held lines then each new one until it leaves; one that stops reading is dropped after 256 lines and never holds up the log |

## `cmd/meshctl`

| Test | Pins |
| --- | --- |
| `TestCommandsDoc` | docs/commands.md equals the registry's rendering (`./scripts/dev.sh regen` rewrites it) |
| `TestEveryCommandHasHelp`, `TestGroupsNeedASubcommand`, `TestFlagsMayFollowPositionalArguments` | the command tree |
| `TestNodeLifecycle`, `TestNodeCommandsExplainAMissingServer` | every `node` command against a real admin socket, `node compose` with both backends, for a server and with a hostile directory |
| `TestNodePhone` | `node add --role phone` prints the profile command, not a token; `node phone` prints the profile alone on stdout (the note on stderr) and `--json`; `node list` and `node show` of a phone; a server and a non-exit exit refused |
| `TestExportCommand`, `TestServerFlagErrors`, `TestAgentJoinTakesTheTokenFromTheEnvironment`, `TestHealthcheckServer`, `TestStatusAndAgentHealthcheck`, `TestPrintStatusOfALaptop`, `TestBytesize` | the other commands; the token never echoed; `status` of a laptop with nothing applied, an exit and errors |
| `TestNodeShowAnEnrolledServer` | `node show` of a joined server: rows, peers, the computed config, a reported rollback; `node list` shows it rolled back |
| `TestUICommand`, `TestAgentRunUIListen` (`ui_test.go`) | `meshctl ui` prints the link alone on stdout and its lifetime and the ssh hint on stderr; a refusal is exit 1; `agent run --ui-listen` refuses addresses off loopback before the agent starts and takes `off` |
| `TestLogsCommand` | `meshctl logs` prints the agent's ring as it holds it; an argument is a usage error; no agent is exit 1 |
| `TestDNSCommand` (`exit_test.go`) | `meshctl dns` shows the state, `on`/`off` switch it through the agent, usage errors never reach it, no agent is exit 1 |
| `TestExitCommand`, `TestExitReportsAnAgentThatDoesNotApply` (`exit_test.go`) | `meshctl exit` shows the exit and choices, switches and waits for the applied hash, `off` sends an empty exit, `--wait 0`; a refusal is exit 1 with the server's reason; usage errors never reach the agent; an agent that does not apply in time is reported with its error |
| `TestUserCommands` | passwords only from a terminal (a fake in tests), twice and matching, never echoed; add, passwd, reset-totp |
| `TestImportCommand` | import then export round trip; a file with unknown fields refused (Linux only) |

## End to end -- `test/e2e`

Build tags `e2e linux`, root, `./scripts/dev.sh e2e`. Network namespaces on a bridge run the real
binary with kernel WireGuard -- the NAS spoke with the userspace backend -- and the server behind a
TLS proxy (a helper process of the test binary).

`TestServerMesh` checks: join -> two peers and traffic between every pair within 60 s; `node
disable` -> the node leaves both other servers within 30 s and is unreachable; server stopped, an
agent killed and its `wg0` deleted -> the restarted agent rebuilds `wg0` from its cache and reaches
the other server.

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

## `internal/server/planner`

Goldens: `testdata/{servers,full,degraded}.golden.json`, derived by hand from the rules in
`planner.go` (see [developer.md](developer.md#goldens)); the full topology carries custom records
and firewall rules. The fixtures are built by `topology()`
in `planner_test.go` from the testfix reference topology.

| Test | Pins |
| --- | --- |
| `TestComputeGolden` | full output for three servers; the full topology (exit, hubs, phone); the degraded one (disabled laptop, endpointless vps-3) |
| `TestComputeIsDeterministic` | output is independent of input order |
| `TestComputedConfigsPassAgentValidation` | every emitted config passes `proto.NodeConfig.Validate` |
| `TestServerPeersForSpokesAreExactly32` | on a server, a spoke or phone owns only its /32 and has no endpoint or keepalive |
| `TestEveryPairCanTalk` | packet simulation with WireGuard's longest-prefix cryptokey routing: every pair of planned nodes reaches each other both ways (amendment A3) |
| `TestSpokesWithDifferentHubsShareOneRelay`, `TestExitRouting`, `TestRelayAndExitServersForward` | relay choice, exit routing, ip_forward and firewall policy flags |
| `TestUnreachableHubFallsBack`, `TestEndpointlessServerWarns` | fallbacks and their actionable warnings |
| `TestDNS`, `TestFirewallRules` | the zone (every active node and custom record, an alias of a disabled node dropped), upstreams in order, split off only in exit mode, `forward_for` only on the phone's hub; 53 from the mesh on every server, custom rules on the servers they name with their source, one order for rules that differ only in their note |
| `TestCheckRejects`, `TestProblemString` | every blocking constraint (custom records, rules and upstreams included), and that a plan with errors yields no configs |

`alloc_test.go`: `TestRangeFor`, `TestBlockMatchesTheSpecAddressPlan`, `TestAllocateLowestFree`,
`TestCheckRequested`, `TestCheckMeshCIDR` -- the S4 address plan and its limits.

## `internal/fsutil`

`TestWriteAtomicReplacesAndLeavesNoTempFiles`: the atomic write both sides use.

## Coverage

`./scripts/dev.sh cov` totals every package, including the adapters the unit tests cannot reach
by design: `internal/agent/wg/{kernel,link,userspace}_linux.go`,
`internal/agent/routes/netlink_linux.go`, `internal/agent/exitmode/policy_linux.go` (netlink,
wgctrl, wireguard-go's tun) and `internal/agent/dns/resolved_linux.go` (systemd-resolved, which a
test must not change on the host it runs on). Their unprivileged paths (opening a backend, a refused create) are
unit-tested; the rest needs root and the wireguard module, and is exercised by
`./scripts/dev.sh e2e` (`TestServerMesh`, and `TestSpokesAndExit` for exit mode and the userspace
backend).

## `internal/tools/vulnjudge`

Copied from solace-cnt-scripts with its tests and govulncheck JSON fixtures in `testdata/`:
fatal on a fixable called vulnerability, warning on one without a released fix, quiet on
findings the code does not call.
