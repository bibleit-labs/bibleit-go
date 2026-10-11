# Customer operation and transport inventory

Reviewed 2026-10-11 against the server working tree. See `server-baseline.json`
for source fingerprints and `docs/status.md` for verification progress. Protocol
availability is distinct from client implementation and integration coverage.
Authority is always decided by the server; permission-aware HELP is the discovery
source. Browser presence alone does not imply a supported bearer API.

| Operation | Server surface | Authentication/authority | Current Go client / CLI |
| --- | --- | --- | --- |
| Ping / server info / identity | PING / SERVER INFO / AUTH INFO | Bearer HTTP or registered SSH key; server/identity discovery permissions | Typed identity and server info available; real HTTP checks pass |
| Permission-aware discovery | HELP [topic] | Bearer or SSH; help.get | Typed Go Help and builder; CLI server help; real HTTP verification passes |
| Personal account / quotas | ACCOUNT INFO / ACCOUNT QUOTA LIST | Bearer or SSH; current principal's account | Typed Go account/quota methods and builders; CLI account info/quotas; real HTTP verification passes |
| Personal translation library | ACCOUNT TRANSLATION LIST/ADD/REMOVE | Bearer or SSH; account library authority and effective quotas | Both available; fixtures verify list grammar |
| Read / search | READ / SEARCH | Bearer or SSH; enabled library and translation.read/search | Raw and typed reading/search available; real verse/chapter/book/search and empty search verified |
| Translation metadata / reference catalogue | TRANSLATION INFO/CATALOG | Bearer or SSH; translation.get | Typed Go methods/builders and CLI available; real metadata/catalogue verification passes |
| Live create/list/info/stats | LIVE CREATE/LIST/<id> INFO/STATS | Bearer or SSH; Live permissions/access; stats owner-only | Typed state/list/stats available; Go HTTP integration covers lifecycle and stats |
| Live start/stop/pause/resume/clear/delete | LIVE <id> action | Bearer or SSH; mutation authority | Typed control acknowledgements; real start/stop/pause/resume/clear/delete verified |
| Live settings and stack | LIVE <id> SET / STACK actions | Bearer or SSH; Live update/get authority, library and running-state constraints | Typed settings/stack/push/pop results; real integration passes |
| Secret generation/removal | LIVE <id> SECRET CREATE/ROTATE/DELETE | Bearer or SSH; owner and update authority | Builders and CLI available; decoder fixtures cover create/rotate |
| Secret connection authentication | LIVE <id> SECRET <value> | SSH connection state; HTTP discards resulting connection state | Builder/CLI available; cannot authorize a later fresh SSH connection |
| Custom secret assignment | SECRET SET explicitly unsupported; generated CREATE/ROTATE secrets only | No supported custom-assignment operation | Client and CLI reject it explicitly; documentation matches parser |
| Audience subscription | LIVE <id> SUBSCRIBE | Persistent SSH; audience permission or secret authorization on same connection | CLI SSH path available; HTTP rejects; actual OpenSSH verse/pause/disconnect/reconnect/closure integration passes |
| Personal manual tokens / SSH keys | ACCOUNT TOKEN / ACCOUNT KEY | Bearer or SSH; personal credential authority/quotas | Typed ListTokens and CLI account tokens verified; key and token mutation methods remain planned |
| Browser CLI authorization | /cli/auth, /api/cli/token, /api/cli/logout | PKCE approval and bearer revocation; username onboarding required | CLI uses library ExchangeCode/RevokeCredential; real HTTP exchange/replay/revocation checks and 4 server PKCE tests pass |
| Organization selection and administration | Browser organization routes/domain services | Active scoped membership; browser session/CSRF | No verified public workspace-selection command; planned server contract |
| Live workspace movement, collaboration, reorder/removal | Dashboard control/invitation endpoints | Browser session/CSRF and current access | No client wrapping of browser APIs; public contract needed |
| Audience WebSocket | /ws?live=<id> | Browser account/audience cookie | No bearer client transport |
| Management WebSocket | /dashboard/api/lives/<id>/events | Browser session, Origin check, current management access | Separate from audience; no bearer client transport |
| Organization Live design | /organizations/<handle>/live-design | Scoped design permission; browser session/CSRF; revision conflicts | Browser feature; public client surface deferred |
| Operator RBAC / translation installation | AUTH ACTOR/ROLE and TRANSLATION FETCH/DELETE | Server administration authority | Outside initial customer CLI scope |

## Shared HTTP constraints

- `/api/cli/command` carries a JSON command string; finite replies are protocol v1.
- 8192 command bytes and 16384 JSON-body bytes; measure UTF-8 bytes, not characters.
- Domain errors can use HTTP 200. Rate limiting uses HTTP 429 plus ERR metadata;
  malformed/oversized requests may return JSON errors with HTTP 400/413.
- CLI login credentials have a separate active cap of 10; manual tokens use plan
  quotas. Scopes may narrow authority and tokens can expire or be revoked.
- No mutation retries. A timeout does not establish whether a mutation committed.

## Resolved protocol discrepancies and retained limits

The assessment corrected the SECRET SET reference on 2026-10-11: secrets are
server-generated by CREATE/ROTATE; custom SET is unsupported. The client and CLI
retain their explicit rejection, so management never becomes authentication.

The v1 tokenizer now rejects malformed quote segments, attempted escaped embedded
quotes and empty quoted arguments with `bad_command`, preserving existing literal
and trailing backslashes. Embedded quotes remain unrepresentable in v1; Go
builders continue rejecting them. This resolves silent input alteration without
changing the request grammar to reinterpret existing backslashes. Shared fixtures
cover the accepted boundaries and rejected inputs.

The assessment fixed duplicate `LIVE STATS` connection counts on 2026-10-10;
the encoder now emits one count derived from subscriber records. Canonical
fixtures use a single field. The client retains compatibility with older identical
duplicates, deduplicates their output order and rejects conflicting values.


The assessment fixed token scope serialization using permission CSV. Real token
metadata integration passes, including permission denial and inactive-token filtering.

The assessment also fixed an intermediate normal Live-deletion regression during
stream authorization changes. The real OpenSSH subscription once again receives
EVENT closed; the complete integration run passes.

Fresh local integration (2026-10-10), using client `9332d35` and CLI `045fd20`,
passes all 27 decoder/18 encoder fixtures, 7 targeted security cases, 11 HTTP
subtests, real OpenSSH subscription/profile/output checks and 8 auth/key EUnit
tests. `server-baseline.json` records the tested server snapshot at HEAD
`1a266cc8bf6aa726c4890b78ddf95a3aadda0ba0` plus assessment changes. Both full
runs passed; subsequent translation/registry edits remain outside that snapshot.
At that historical checkpoint the runner isolated the localhost:8080 endpoint and removes its disposable
database. No production endpoint or browser-only API coverage is implied.

Latest assessment validation (2026-10-10), CLI harness `26de771` and client
`5138077`, passes the full HTTP/SSH gate plus 10 import-access and 14 publication
cases. The immutable publication module, concurrent archive installation and
retained-handle changes are included. Source (63) and test (57) fingerprints in
`server-baseline.json` matched the server working tree at completion. The optional
`--imports` runner executes actual suite callbacks/cases with local fixtures,
not the full Common Test campaign; later server edits need a fresh snapshot.

Protocol fix validation (2026-10-11 local date) passes 34 request decoder and
18 response encoder fixtures, 7 targeted security cases, all 11 HTTP subtests,
real OpenSSH lifecycle/profile/output checks, and 17 auth/key/protocol EUnit tests.
The nine new server request tests cover malformed quoting and generated-secret
commands. All 10 import-access and 15 publication cases pass. The synchronized
baseline records 63 source and 58 test fingerprints and UTC review date; it matched
the assessment working tree at validation completion. The assessment subsequently confirmed its full `make test` on 2026-10-11:
248 Common Test cases, 48 Erlang unit tests, 71 JavaScript tests and the native
C test, with no skips. This server-owned evidence is recorded in
`../bibleit-server/docs/developers/security-test-results.md`; it is separate
from the client integration run and confirms the same protocol contract.

The 2026-10-11 single-configuration CLI integration refresh passed the full tracked
HTTP/SSH/auth/import gate: 34 decoder/18 encoder fixtures, 11 HTTP subtests, real
SSH lifecycle and single-credential output parity, 17 auth/key/protocol tests and
25 import cases. The synchronized snapshot now records 63 source and 59 test
fingerprints, including further assessment plan-support changes. Snapshotting
a test is not a claim that the client harness executes every server test.

## Fixed production origin and final assessment compatibility — 11 October 2026

ServerEndpoint is now https://bibleit-server.fly.dev. The CLI uses ordinary SSH
at bibleit-server.fly.dev:22. Credential-only constructors and redirect blocking
remain unchanged; default HTTPS certificate verification is enforced.
Fresh integration at clean server 8923d86/native f6fed37 passes 34 decoder/18
encoder fixtures, seven security cases, 11 HTTP subtests, real OpenSSH lifecycle
and JSON parity, 18 auth/key/protocol tests and 26 import cases. Baselines record
68 source/70 test/11 native fingerprints. The fixture has no egress, maps the
service name to loopback and trusts its CA only inside the container; three TLS
boundary checks pass. Assessment separately closes 25/25 scoped packages and
qualifies the actual deployed edge. This does not add browser-only API coverage
or publish a client version.

## Canonical-domain candidate compatibility — 11 October 2026

The confirmed fixed origin is now https://bibleit.app. CLI SSH connects to
bibleit.app:22 and preserves existing verified service trust with the historical
Fly-hostname known_hosts alias. Fresh isolated HTTP/SSH/import passes at server
0c0da9f plus assessment canonical-host candidate changes/native f6fed37, including
actual changed saved-key rejection. Synchronized baselines record 69 source,
71 test and 11 native hashes. This working-tree candidate does not certify a final
committed source or actual custom-domain DNS/TLS/SSH/authentication deployment;
those qualifications remain with assessment/Fly before customer release.

Final assessed source 5969f2f passes independent fixed-endpoint integration and
boot qualification. All 69 source/72 test hashes match that commit, native
f6fed37 retains 11 verified inputs. Coverage additionally includes the canonical
hostname/default/normalization/stale-host boot case and 24 auth/key/protocol/
canonical/provider unit tests. Unrelated frontend edits keep dirty metadata true.
Actual public custom-domain and human-widget qualification remain pending;
no client runtime changes or customer publication occur in this metadata refresh.

## Supplemental direct alias boot repair — 11 October 2026

Assessment repair ef63e91 registers BIBLEIT_HTTP_REDIRECT_HOSTS in the real
dotenv parser. Independent direct boot qualification passes normalized configured
aliases, absent aliases and wildcard rejection, using the committed server
regression in cleared-env child processes with empty cwd and no external route.
No database connection or service is started. All 69 source/73 test hashes match
that committed repair; the 11 native inputs remain f6fed37. The sole runtime
difference from the previous full client run is the one-line parser registration.

server-boot-validation.json records supplemental provenance and links the retained
5969f2f full HTTP/SSH/import baseline by SHA-256. Those unchanged client checks
were not repeated. Assessment-owned repaired-source broad results pass 337 CT/63
EUnit/71 JavaScript/native 65536/10 bootstrap. Actual public-domain qualification
and human-widget acceptance remain pending; no client runtime or release changes.

## Actual canonical public deployment acceptance — 11 October 2026

Actual public-DNS technical qualification passes on source c586107, runtime
repair ef63e91/native f6fed37 and immutable image
sha256:1140436c3d9e7de1b1f80e88aa208a0a6c06cff4b7ba371e0f5ddcbb5b6ace22.
403 source hashes and 20 actual-image checks matched. HTTPS boundary23/alias39
checks pass on each address family, plus auth11/canonical SSH10/paired rate15.
The actual public SSH fingerprint is unchanged and scoped synthetic fixture
counts are zero after cleanup. Credential-free canonical-deployment-validation.json
links deployed source/image and redacted report hashes to the retained client
integration and supplemental boot evidence. No client runtime changed.

The operator confirms real Google/GitHub OAuth sign-in. The legacy Live
certificate is removed while all three new-server certificates remain Issued;
legacy app/data/volumes and Pages remain intact. Real Turnstile backend
fresh-token success/replay is the remaining domain acceptance gate; visible
widget success and two fail-closed backend denials are only partial evidence.
Customer release tags/assets remain held, with standalone dependency/packaging
validation and release decisions still required.

## Subsequent native browser form qualification blocker

Real Chromium requestSubmit reproduces Origin:null under the global no-referrer
policy; strict CSRF rejects native signup/login/CLI approval/settings forms.
Earlier fetch/explicit-Origin probes did not cover actual native serialization.
The assessment is validating same-origin policy for ordinary HTML while retaining
no-referrer on sensitive URLs/assets/redirects/non-HTML and strict Origin/CSRF
rejection. Preserve earlier technical provenance, but hold final qualification
freeze until repaired-source/browser-regression/full-suite/rollout acceptance.
Real Turnstile backend success/replay remains pending separately. No Go client
protocol/runtime change is needed; customer release remains held.

## Qualified native form repair deployment

Signup repair source34d80f8/nativef6fed37 is technically qualified on image
sha256:33b4ca3c13829c1c25f0745037212cc0c14e5e7768c2f78e58206c461ef8d35d,
run38144206075:403 source hashes/20 image checks match. Production requestSubmit
carries canonical Origin/Referer and passes CSRF before missing-token denial.
Provider traffic is blocked, no account/email is created, and response policy
is same-origin. Updated boundary23/alias39 checks pass per address family;
source338CT/63EUnit/71JS/native/10bootstrap passes. Redacted reports carry exact
source/native/image provenance. Earlier auth/SSH/rate results remain linked at
their tested source; unchanged client protocol tests were not repeated.

Real Turnstile backend fresh-token success/replay and actual email delivery/
verification remain manual gates. Separate catalog rollout server6ba81e2/native
d6e31dc (workflow38144732991) has human approval and independent ownership; its
deployment/catalog qualification is pending. This point-in-time signup repair
qualification must not be claimed for that newer source/native/image. No client
runtime changes or customer tags/assets occur in this metadata update.
