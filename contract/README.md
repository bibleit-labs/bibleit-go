# Bibleit client contract: protocol v1

## Sparse native and localized finite-client qualification

Catalog source[private revision withheld]/native [private revision withheld] is independently operationally qualified on
image sha256:[private digest withheld],
[run38144732991]([private service reference withheld]):
405 source hashes/eight build stages/20 actual-image checks, native amd64, matching
published/running digest and clean fixture cleanup. Public boundary23/alias39
checks pass per explicit address family, along with native browser admission.
Existing BPML66-book catalogue and omitted/present Luke17:36/37 semantics pass.

Focused clients then exposed Latin-1 bytes in Portuguese/German catalogue binary
literals. Server repair[private digest withheld] makes those
literals UTF-8; native [private digest withheld] is unchanged.
Server340 CT/66 EUnit/75 JS/native65536/10 bootstrap passes, with all198 language
names and installed66-book catalogue dashboard/protocol agreement across3languages.

Independent fixed-endpoint --catalog-only qualification passes five typed HTTP
groups, both BPML/LUT1912 catalogues and accented named reads, five HTTP/SSH parity
cases, invalid-index publication rejection preserving the prior sparse generation,
and three TLS checks. All69source/75test/11native hashes match committed4243/d6
and after-run inputs; unrelated frontend edits preserve truthful dirty metadata.
server-catalog-validation.json records this focused evidence separately from the
full[private revision withheld]/f6 integration and [private revision withheld]/f6 boot snapshots. All files match across
Go/CLI. No broad protocol/auth/Live suite was unnecessarily repeated.

A catalogue chapter's verses value is its highest source verse identifier, not
a dense row count. An omitted verse returns verse_not_found; chapter reads include
only present rows and preserve source numbering. The strict client UTF-8 decoder
is unchanged. Exact repaired deployment/public qualification is complete at server
[private revision withheld]/native [private revision withheld], image
sha256:[private digest withheld] (workflow
38147193441): 406 committed hashes, eight successful stages, 20 actual-image startup
checks, published/running digest parity and 185 APK comparisons with zero fixed-advisory
alerts. Refreshed boundary 23/alias 39 checks pass independently on IPv4 and IPv6;
native form CSRF admission and missing-token denial pass. Production BPML has 66 strict
UTF-8 BOOK records and matching Gênesis. The five public report hashes are recorded in
canonical-deployment-validation.json. Real Turnstile backend success/replay and actual
email delivery/verification remain open. No customer tags/assets.

Reproduce with the existing disposable fixture:

```sh
python3 scripts/check_server_container.py --catalog-only --revision [private digest withheld] --baseline contract/server-catalog-validation.json
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
this file. Fixtures were derived from `private-service/src/protocol/bibleit_protocol.erl`
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


A later font-only deployment is independently recorded at source[private revision withheld]/native d6,
workflow38157279015, image [private image withheld]
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

Replay guard source [private revision withheld] is now committed and locally qualified (344 CT,
68 EUnit,75 JS,10 bootstrap,native65536;focused six unit/four DB-HTTP cases).
It appends migration008 and preserves prior identities. Exact-image deployment
run38162034582 and genuine fresh-token/replay acceptance remain pending; client
runtime and native [private revision withheld] are unchanged.

Production-image run38162034582 failed before publish/deploy: its SQL release
allowlist omitted008. Local runtime tests did not qualify assembled packaging.
The staging correction covers all eight migrations and rejects omissions; a final
image is still required. Production remains[private revision withheld]. Fly requests narrow security-
release approval excluding undeployed Reader changes; real fresh/replay retesting
remains open after that exact corrected rollout.

The packaging correction is committed at[private revision withheld], parent[private revision withheld], on the exact
security-only branch explicitly approved by the human in Fly. Reader changes are
excluded. Corrected run38163282528 is source/ref-pinned and queued/running; no
qualified image or provider gate closure is yet claimed.


Final provider qualification: exact[private revision withheld]/native d6 image[private revision withheld]/run38163282528
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
