# Customer operation and transport inventory

Reviewed 2026-10-09 against the server working tree. See `server-baseline.json`
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
| Custom secret assignment | Reference documents SECRET SET; parser rejects it | Intended owner/update authority | CLI rejects it explicitly; server contract unresolved |
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

## Outstanding server discrepancies

`SECRET SET` documentation differs from the parser and remains server-owned.
Request quoting also has no embedded-quote escape syntax: a synthetic
`LIVE CREATE "He said \"hello\""` decodes to `He said \hello\`, silently
changing the intended value. Go builders reject embedded quotes to prevent this
loss. Both protocol issues were reproduced and reported to the assessment on
2026-10-10. Changing literal-backslash semantics requires an explicit compatible
contract before client restrictions can be relaxed.

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
`[private digest withheld]` plus assessment changes. Both full
runs passed; subsequent translation/registry edits remain outside that snapshot.
The runner isolates the fixed localhost:8080 endpoint and removes its disposable
database. No production endpoint or browser-only API coverage is implied.

Latest assessment validation (2026-10-10), CLI harness `26de771` and client
`5138077`, passes the full HTTP/SSH gate plus 10 import-access and 14 publication
cases. The immutable publication module, concurrent archive installation and
retained-handle changes are included. Source (63) and test (57) fingerprints in
`server-baseline.json` matched the server working tree at completion. The optional
`--imports` runner executes actual suite callbacks/cases with local fixtures,
not the full Common Test campaign; later server edits need a fresh snapshot.
