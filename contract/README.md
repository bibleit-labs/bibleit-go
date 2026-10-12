# Bibleit client contract: protocol v1

This reference describes the client wire format and shared fixtures. See
[feature coverage](features.md) for supported operations and transports.

## Transport

`POST /api/cli/command` takes `Authorization: Bearer <token>` and a JSON body
of the form `{"command":"PING"}`. Responses use the same UTF-8, newline-delimited
text records as SSH exec. Domain errors can arrive with HTTP 200; inspect the
protocol envelope, not only the HTTP status.

HTTP supports finite command replies. `LIVE <id> SUBSCRIBE` is SSH-only.
V1 interactive SSH has FIFO replies without request IDs, with EVENT records
between envelopes. The Go decoder is for finite HTTP/SSH-exec replies,
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
client surface.

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
alone to infer framing.

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

Tests consume [response fixtures](fixtures/responses-v1.json) and
[request fixtures](fixtures/requests-v1.json) as data. Malformed and future-error
cases exercise decoder robustness. The request fixtures' `server_term` and
response fixtures' `server_response_term` fields are serialized expectations;
clients never evaluate them.

Identical repeated response fields are accepted for compatibility with older
Live stats; conflicting repeated values are rejected.

## Token metadata

`ACCOUNT TOKEN LIST` requires `token.get` and emits an `OK actor=... count=N`
header, N `TOKEN` records and END. IDs, source, issue timestamp and optional
issuer/label/scopes/last-use/expiry/retirement metadata are retained. No bearer
value is returned. Listings filter inactive credentials and do not expose other
actors' tokens. A scoped credential still needs token.get to list metadata.

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
Both methods use HTTPS, reject redirects, accept contexts and bound response bodies.

For contributor checks, see [CONTRIBUTING.md](../CONTRIBUTING.md).
Historical integration and release records are available through the
[maintainer guide](../docs/maintenance.md).
