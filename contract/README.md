# Bibleit client contract: protocol v1

## Latest fixed-endpoint compatibility snapshot — 11 October 2026

Fresh isolated HTTP/SSH/import checks passed against clean server
8923d8675b2e129bdd868e307dd89c599235f25a, which includes deployed code
7442010eebcfb1a2a9105db312ebaeae5f789003. Native libbibleit remains
f6fed37919c848b1aa0208900af992a5e8a507f2. The synchronized baseline records
68 Erlang source, 70 test and 11 native build-input fingerprints. Native code is
rebuilt from the frozen snapshot; concurrent copy-input changes are rejected.

The fixed HTTPS origin is https://bibleit-server.fly.dev; ordinary CLI SSH uses
bibleit-server.fly.dev:22. Public constructors use the production URL unchanged.
The disposable container has no external network route and maps the hostname
exclusively to loopback. TLS verification uses a fixture CA only in the container;
correct trust/name succeeds, while wrong name and untrusted CA fail. No host
certificate installation, customer endpoint override or PROXY preamble is used.

Passing coverage: 34 decoder/18 encoder fixtures, seven server security cases,
11 Go HTTP subtests, OpenSSH subscription/pause/reconnect/delete closure and
single-credential JSON parity, 18 auth/key/protocol tests, and 26 import-access/
publication cases (11 access, 15 publication). Approved import delivery is adapted
only inside suite setup; URL policy, authorization, validation and publication
remain real. Authorized unreviewed/private-source import rejection passes.
The disposable database is removed; no production principal/data is modified.

The assessment separately records 25/25 scoped packages complete and qualifies
actual HTTPS/SSH/Fly routing. This isolated client check does not certify a new
public-edge deployment, browser/provider login or bearer management WebSocket API.
A released client version and standalone CLI candidate remain release gates.

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
this file. Fixtures were derived from `bibleit-server/src/protocol/bibleit_protocol.erl`
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
The CLI retains its profile on failed revocation and clears it after 204 or 401.
Both methods share HTTPS/loopback endpoint rules, reject redirects, accept
contexts and bound response bodies. Real HTTP checks verify exchange, replay,
bad-verifier consumption, logout and post-logout rejection.


## Fixed-service integration (2026-10-10)

The simple API-token constructors now pass against a real server at the fixed
`http://localhost:8080` endpoint. No client destination override is used. If host
port 8080 is occupied, run `python3 scripts/check_server_container.py` to use an
isolated container namespace and disposable database. The verified baseline
includes the fixed endpoint and current source fingerprints. The server checkout
and developer service are not changed.
