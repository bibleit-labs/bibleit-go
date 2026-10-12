# Bibleit client contract: protocol v1

> Public-copy update, 2026-10-12: private implementation references in these historical records are withheld; originals remain in private local storage. Current SSH verifies bibleit.app directly without a legacy host-key alias.

## Latest canonical-domain compatibility snapshot — 11 October 2026

Fresh isolated HTTP/SSH/import and boot checks passed against assessment revision
5969f2f73547010cddd3ae5b0828839cad118f2e. The synchronized baseline records
69 Erlang source, 72 test and 11 native input fingerprints; all Erlang inputs
match that commit. Native remains f6fed37919c848b1aa0208900af992a5e8a507f2.
Unrelated frontend working-tree edits were present, so dirty metadata stays true;
no server checkout was modified by the client harness.

The fixed HTTPS/browser origin is https://bibleit.app; ordinary CLI SSH uses
bibleit.app:22. The historical Fly name is solely an OpenSSH known_hosts alias,
never the network destination, preserving existing verified host trust. A real
OpenSSH regression replaces the saved key and confirms changed-key rejection.
Clients expose no endpoint overrides and continue to reject HTTP redirects.

The disposable container has no external route and resolves the service name only
to loopback. Public constructors retain the fixed URL and verify TLS with a fixture
CA trusted only in the container. Correct name/trust succeeds; wrong hostname and
untrusted CA fail. The server fixture configures the canonical public origin while
exercising client paths and unsets it before other suite teardown. No PROXY preamble,
host certificate installation or production principal/data mutation is involved.

Passing coverage: 34 decoder/18 encoder fixtures, seven security cases, 11 Go HTTP
subtests, real OpenSSH lifecycle/JSON parity and changed-key rejection, 24 auth/key/
protocol/canonical/provider unit tests, one boot configuration case, and 26 import
access/publication cases (11 access, 15 publication).
Import delivery is adapted only inside suite setup; URL policy, authorization,
validation and publication remain real. The unique disposable database is removed.

The earlier assessment closed 25/25 packages for the Fly-hostname deployment.
Actual custom-domain technical acceptance has since passed; see the deployed
qualification below. This isolated check itself does not certify the public
route or browser/provider integrations.
Final packaging/deployment input freeze, released client version and standalone
CLI candidate remain release gates. Earlier Fly-hostname native/candidate results are historical.

## Supplemental alias boot qualification — 11 October 2026

The assessment subsequently repaired the missing BIBLEIT_HTTP_REDIRECT_HOSTS
parser registration at ef63e913ae4d7f3e28a2dc561e869976db667a20. The new direct
boot regression passes independently in a read-only, egress-disabled container:
configured aliases normalize, absent aliases remain empty and wildcards fail.
Each child process clears its environment, uses an empty cwd and synthetic DB
settings, and calls configure() without starting a service or connecting to a DB.

`server-boot-validation.json` records all 69 source/73 test/11 native hashes at
that committed repair, the preceding integration-baseline hash and the sole
runtime difference: the one-line alias parser registration. Native f6fed37 is
unchanged. Unrelated frontend edits retain truthful dirty metadata. The original
`server-baseline.json` deliberately retains the full HTTP/SSH/import run at
5969f2f; those unchanged client checks were not repeated at ef63e91. Both evidence
files are identical in the Go and CLI repositories. The full future integration
runner now includes this direct boot EUnit case alongside the existing checks.

Reproduce the focused qualification using the existing integration tools image:

```sh
python3 scripts/check_server_boot.py --revision ef63e913ae4d7f3e28a2dc561e869976db667a20 --report contract/server-boot-validation.json
```

The assessment reports its full suite green: 337 CT/63 EUnit/71 JavaScript/native
65536/10 bootstrap cases. The subsequent actual public-domain technical
qualification is recorded below; human Google/GitHub sign-in passes. Real
Turnstile backend acceptance remains open.

## Actual public canonical-domain technical qualification — 11 October 2026

Assessment/Fly confirms actual public-DNS acceptance on deployed c586107, runtime
repair ef63e91/native f6fed37 and immutable image
sha256:1140436c3d9e7de1b1f80e88aa208a0a6c06cff4b7ba371e0f5ddcbb5b6ace22.
The deployment verified 403 source hashes and 20 actual-image checks. Two DNS
authorities plus Cloudflare/Google confirm the DNS-only apex/live routes and www
CNAME. Each address family passes 23 HTTPS boundary and 39 alias checks; 11 auth,
10 actual canonical SSH and 15 paired rate checks pass. Public SSH retains
SHA256:MUnCXVIEU/t8AZk0es6XgyeE9c9Ak6LmCZdEmy99ips. Scoped synthetic fixture
counts are zero after cleanup; their credentials/private key were removed.

`canonical-deployment-validation.json` records credential-free source/image,
endpoint, report-hash and cleanup provenance, and links the two independent
client evidence files. It contains no customer credentials. The summary is
identical across Go/CLI and does not change their runtime or publish a release.

The operator confirms Google/GitHub sign-in. Real Turnstile backend token
success/replay remains pending. Visible widget success and two passing
missing/bogus-token backend denials are partial evidence only. Keep customer releases held until the
remaining acceptance and standalone release gates close.

## Repaired native HTML form deployment qualification

Repair source 34d80f811b3ac9179d6afe0966185e2b3517181e/native f6fed37 is deployed
and technically qualified on immutable image
sha256:33b4ca3c13829c1c25f0745037212cc0c14e5e7768c2f78e58206c461ef8d35d,
run 38144206075 (private record retained locally).
403 source hashes and 20 actual-image checks match. Real production requestSubmit
carries canonical Origin/Referer and passes CSRF before missing-token denial;
ordinary HTML uses same-origin referrer policy. Token/callback/Live viewer/
explicit private handler/assets/redirects/non-HTML retain no-referrer, and strict
Origin/CSRF/CSP is retained. Provider traffic is blocked and no account/email is
created. Updated 23 boundary/39 alias cases pass per address family; local final
source passes 338 CT/63 EUnit/71 JS/native/10 bootstrap.

canonical-deployment-validation.json records current repaired report hashes and
source/native/image provenance, retaining earlier auth/SSH/rate/provider evidence
at its tested source. The full client integration/boot snapshots remain unchanged;
those client checks were not unnecessarily repeated. Real Turnstile backend
fresh-token success/replay and actual email delivery/verification remain pending.
Keep customer releases held for that manual gate and standalone packaging. The
separately approved catalog rollout uses server 6ba81e2/native d6e31dc
(workflow 38144732991) is operationally qualified; its focused compatibility and
subsequent UTF-8 repair evidence are recorded below.
The 34d80f8/f6fed37 evidence is point-in-time and does not qualify that later image. No client runtime changes are required.

## Sparse native and localized finite-client qualification

Catalog source6ba81e2/native d6e31dc is independently operationally qualified on
image sha256:7456bfec37b1d2285041387ef6135685dc707748ea82adbff60417e96c5fec47,
run38144732991 (private record retained locally):
405 source hashes/eight build stages/20 actual-image checks, native amd64, matching
published/running digest and clean fixture cleanup. Public boundary23/alias39
checks pass per explicit address family, along with native browser admission.
Existing BPML66-book catalogue and omitted/present Luke17:36/37 semantics pass.

Focused clients then exposed Latin-1 bytes in Portuguese/German catalogue binary
literals. Server repair4243c9a3c1b6e8f8b1f94b93f0880b3aa98464f6 makes those
literals UTF-8; native d6e31dc8eb7c478e2f915e7bd156b047a2cb279e is unchanged.
Server340 CT/66 EUnit/75 JS/native65536/10 bootstrap passes, with all198 language
names and installed66-book catalogue dashboard/protocol agreement across3languages.

Independent fixed-endpoint --catalog-only qualification passes five typed HTTP
groups, both BPML/LUT1912 catalogues and accented named reads, five HTTP/SSH parity
cases, invalid-index publication rejection preserving the prior sparse generation,
and three TLS checks. All69source/75test/11native hashes match committed4243/d6
and after-run inputs; unrelated frontend edits preserve truthful dirty metadata.
server-catalog-validation.json records this focused evidence separately from the
full5969f2f/f6 integration and ef63e91/f6 boot snapshots. All files match across
Go/CLI. No broad protocol/auth/Live suite was unnecessarily repeated.

A catalogue chapter's verses value is its highest source verse identifier, not
a dense row count. An omitted verse returns verse_not_found; chapter reads include
only present rows and preserve source numbering. The strict client UTF-8 decoder
is unchanged. Exact repaired deployment/public qualification is complete at server
4243c9a/native d6e31dc, image
sha256:46670b5fba63859a6892d64c1501122ecdddfd1edc1410555eba7c165143b5c1 (workflow
38147193441): 406 committed hashes, eight successful stages, 20 actual-image startup
checks, published/running digest parity and 185 APK comparisons with zero fixed-advisory
alerts. Refreshed boundary 23/alias 39 checks pass independently on IPv4 and IPv6;
native form CSRF admission and missing-token denial pass. Production BPML has 66 strict
UTF-8 BOOK records and matching Gênesis. The five public report hashes are recorded in
canonical-deployment-validation.json. Real Turnstile backend success/replay and actual
email delivery/verification remain open. No customer tags/assets.

Reproduce with the existing disposable fixture:

```sh
python3 scripts/check_server_container.py --catalog-only --revision 4243c9a3c1b6e8f8b1f94b93f0880b3aa98464f6 --baseline contract/server-catalog-validation.json
```

## Transport

`POST /api/cli/command` takes `Authorization: Bearer <token>` and a JSON body
of the form `{"command":"PING"}`. Responses use the same UTF-8, newline-delimited
text records as SSH exec. Domain errors can arrive with HTTP 200; inspect the
protocol envelope, not only the HTTP status.

HTTP supports finite command replies. `LIVE <id> SUBSCRIBE` is SSH-only.
Browser WebSockets are not yet a bearer-token replacement for client subscriptions.
V1 interactive SSH has FIFO replies without request IDs, with EVENT records
between envelopes. The initial Go decoder is for finite HTTP/SSH-exec replies,
not multiplexed interactive channels.

## Commands and quoting

Commands are constructed from supported operations and individual arguments.
Arguments containing spaces or tabs are double quoted. CR, LF, NUL, and double
quotes are rejected: the current server request tokenizer does not implement
escaped quotes. The server rejects empty quoted arguments, unmatched quotes,
concatenated quote segments and attempted escaped embedded quotes with
`bad_command`. Backslashes in requests are literal and must not be doubled; a
backslash immediately before a closing delimiter remains part of the value.
The server's response encoder does escape quotes, backslashes, LF/CR/tab and ASCII control characters
(`\n`, `\r`, `\t`, `\uNNNN`); clients decode these when parsing field values. Request and response escaping differ in v1.

Customer translation commands use `ACCOUNT TRANSLATION ...`. Global
`TRANSLATION FETCH`, `DELETE`, and catalogue administration are outside this
initial client surface.

The current server accepts `LIVE <id> SECRET <value>` for secret authentication
and `SECRET CREATE|ROTATE|DELETE` for secret management. It explicitly rejects
`SECRET SET <value>`. The CLI rejects `secret set <value>` with an explicit unsupported-operation
message, rather than silently treating management as authentication. Secret authentication is connection-scoped;
a one-command invocation does not establish a persistent subscription session.

## Finite responses

- Single reply: `OK ...` followed by a newline.
- Multiple records: `OK ...`, data records, and an `END` line.
- Error: `ERR <code> [metadata...]` followed by a newline.
- Empty multi-record replies still require END.
- Unknown success fields and error codes must be retained.
- Error metadata can include `retry_after_ms`; clients do not retry mutations.

Current multi-record headers use `count`, `verses`, `results`, `books`,
`commands`, or `connections`. Live push responses use `event=verse` and a numeric
`translations` field. Ordinary translation lists also use `translations`, but
as a comma-separated string in a single-record response. Account translation
list/add/remove also include `count` without an END marker. HTTP and CLI SSH exec
use `DecodeCommandResponse` to apply that command-specific framing rule. Generic
`DecodeResponse` cannot disambiguate count-only headers. Never use a field
alone to infer framing. A future protocol should make envelope framing explicit.

Account information currently includes notification preferences formatted as an
opaque Erlang map. The encoder can wrap this value across physical lines.
The Go parser retains balanced compound fields as text, including line breaks;
it never evaluates them. Newlines outside compound fields remain invalid.
Original logical records and unknown fields remain available in typed results.

## Fixture format

`fixtures/responses-v1.json` is a JSON array. Each case has:

- `name`: stable descriptive name.
- `wire`: complete response, including newlines.
- `records`: expected parsed record types and string fields for success.
- `error_code` and `error_fields`: expected server error and metadata.
- `invalid: true`: malformed or incomplete data that must fail decoding.

Go tests consume these fixtures now. The other language clients should reuse
this file. Fixtures were derived from `[private implementation reference withheld]`
and its protocol tests; 27 request fixtures and 18 response encodings are now verified against compiled
server source with `python3 scripts/check_server.py`. Malformed and future-error
response cases remain client-decoder robustness fixtures. Server code has not been modified by this extraction.

## Server review status (2026-10-10)

The server now has organization workspaces, separate CLI credential quotas,
additional SSH key algorithms, request byte limits, and browser management
WebSockets. The current client fixtures cover finite protocol replies only.
See [the revised plan](../docs/plan.md) for the operation inventory and integration
testing work required before publishing clients.

The new server protocol reference documents `SECRET SET`, while its parser still
rejects it. The CLI now rejects `secret set` until server support exists;
the documentation/parser mismatch still needs a server-side resolution. Browser-only workspace
and event operations are not supported bearer-token client APIs.

## Next increments

The tracked Go core now passes complete isolated HTTP/SSH integration, including
all typed results and authorization transport. See docs/status.md for counts.

1. Resolve the remaining SECRET SET and repeated stats-field discrepancies.
2. Finalize shell completion/destructive-command UX and native platform CI;
   account profiles at the project-managed service and versioned JSON/NDJSON are implemented.
3. Define public bearer-authenticated workspace and event contracts with the server.
4. Validate the core contract with Python, then Ruby and Rust; settle stable
   repository/package paths before publishing independent releases.

## Verification and progress

See [features.md](features.md) for operation/transport support and
[../docs/status.md](../docs/status.md) for item-by-item progress.
`fixtures/requests-v1.json` contains wire strings and expected Erlang decoder
terms in `server_term`; those terms are parsed as data, not evaluated. Response
fixtures with `server_response_term` are checked against the real server encoder.
Source fingerprints are recorded in `server-baseline.json`.

```sh
python3 scripts/check_server.py
python3 scripts/check_server.py --integration
# Include latest assessment import-access/publication cases in the isolated runner:
python3 scripts/check_server_container.py --imports --baseline contract/server-baseline.json
```

Identical repeated response fields are accepted for compatibility with older
Live stats; conflicting repeated values are rejected.

## Translation metadata encoder issue

The assessment session fixed the HTTP 500 from optional metadata: `fields/1`
now accepts both atom and binary field names. The shared metadata encoder
fixture exercises binary keys, and the real metadata/catalogue HTTP checks
pass against a fresh snapshot. This CLI work did not edit server files.


## Token metadata (2026-10-10)

`ACCOUNT TOKEN LIST` requires `token.get` and emits an `OK actor=... count=N`
header, N `TOKEN` records and END. IDs, source, issue timestamp and optional
issuer/label/scopes/last-use/expiry/retirement metadata are retained. No bearer
value is returned. Listings filter inactive credentials and do not expose other
actors' tokens. A scoped credential still needs token.get to list metadata.

The assessment fixed nonempty scope encoding using the permissions CSV
convention. A compiled encoder fixture and real HTTP listing now verify that
representation, with scoped authority checks and active-only metadata filtering.
The assessment also fixed an intermediate SSH Live-deletion closure regression.
The complete integration run passes and the verified baseline is updated.

## Authorization HTTP transport

`POST /api/cli/token` exchanges form fields code, code_verifier and optional
device_name. HTTP 200 returns actor, identity_name, auth_provider, token_id and
access_token. Errors are JSON with HTTP status (e.g. 400 for invalid/reused code
or invalid verifier, 403 for username onboarding, 409 for CLI credential quota).
Codes are one-use; a bad verifier also consumes the code. No automatic retries.

`POST /api/cli/logout` uses the credential's bearer header. HTTP 204 acknowledges
revocation; an already invalid credential returns 401. The Go library retains
error status/code/metadata and leaves local storage decisions to the application.
The CLI retains its saved credential on failed revocation and clears it after 204 or 401.
Both methods share HTTPS/loopback endpoint rules, reject redirects, accept
contexts and bound response bodies. Real HTTP checks verify exchange, replay,
bad-verifier consumption, logout and post-logout rejection.


## Historical fixed-service integration (2026-10-10)

At that checkpoint the simple API-token constructors passed against a real server at the fixed
`http://localhost:8080` endpoint. No client destination override is used. If host
port 8080 is occupied, run `python3 scripts/check_server_container.py` to use an
isolated container namespace and disposable database. The verified baseline
includes the fixed endpoint and current source fingerprints. The server checkout
and developer service are not changed.

The 2026-10-11 refresh supersedes the localhost fixture strategy: run
`python3 scripts/check_server_container.py --imports` for the egress-disabled
HTTPS/SSH fixture at the fixed production hostname. Direct host integration is
rejected. See the latest compatibility snapshot above.


A later font-only deployment is independently recorded at source14628f3/native d6,
workflow38157279015, image [historical deployment image withheld].
Its artifact verifies406 hashes/eight stages/native amd64/cleanup[0,0], with185
APK comparisons and zero fixed alerts; published and running image match. Its
owner confirms public CSS/signup/readiness. This separate proof does not silently
inherit4243's full client/public-boundary qualification. Assessment ledger a40d27d
records human email delivery/verification and fresh-token acceptance; real replay
rejection remains pending an improved human helper retry. No identifiers/secrets
are retained and no customer release is published.

The improved human retry confirms same-body token replay reaches input validation
twice, with CSRF/rate/continuity interference ruled out. Assessment repair is in
progress; genuine fresh-token/replay acceptance must be requalified before release.
Mail delivery/verification remains separately human-confirmed.

Replay guard source d160773 is now committed and locally qualified (344 CT,
68 EUnit,75 JS,10 bootstrap,native65536;focused six unit/four DB-HTTP cases).
It appends migration008 and preserves prior identities. Exact-image deployment
run38162034582 and genuine fresh-token/replay acceptance remain pending; client
runtime and native d6e31dc are unchanged.

Production-image run38162034582 failed before publish/deploy: its SQL release
allowlist omitted008. Local runtime tests did not qualify assembled packaging.
The staging correction covers all eight migrations and rejects omissions; a final
image is still required. Production remains14628f3. Fly requests narrow security-
release approval excluding undeployed Reader changes; real fresh/replay retesting
remains open after that exact corrected rollout.

The packaging correction is committed at3359535, parentd160773, on the exact
security-only branch explicitly approved by the human in Fly. Reader changes are
excluded. Corrected run38163282528 is source/ref-pinned and queued/running; no
qualified image or provider gate closure is yet claimed.


Final provider qualification: exact3359535/native d6 image0bbb449/run38163282528
is qualified and running (408 hashes/eight stages/20 image checks/native amd64/
cleanup[0,0]/digest parity;185 APK comparisons/zero fixed alerts; eight migration
hashes/table metadata verified). Native form/boundary23/alias39 pass each family.
The redacted genuine human report confirms passed/fresh/replayRejected all true,
with stable CSRF/no rate limiting and no helper-created account/email. Actual
email delivery/verification remains separately attributed to the prior14628 flow.
Ledger53239ad closes both provider gates. Six public/human report hashes and the
artifact report hash are recorded in canonical-deployment-validation.json;
earlier failures and independent client snapshots retain their tested provenance.
No customer tag/release or client runtime change is implied.
