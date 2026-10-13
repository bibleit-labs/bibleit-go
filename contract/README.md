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

## Organization verification foundation (additive contract)

The organization foundation adds read-only commands over bearer HTTP and SSH:

- `ORGANIZATION LIST [cursor]`: the authenticated principal's organization memberships
  and private applications, as a mixed directory.
- `ORGANIZATION APPLICATION <application-id> STATUS`: status of one application
  owned by the authenticated principal.

Both require an active authenticated personal principal and `server.get`.
`organization.create` is not required to inspect existing requests. These reads
remain available when new application submission is disabled. Authenticated
`SERVER INFO` advertises `organization_status`; permission-aware
`HELP ORGANIZATION` describes the two commands. An old server without the
capability does not support these reads. SDK command builders and typed directory/status methods support these reads.
The CLI exposes `organization list [cursor]` and
`organization application <id> status`; it discovers `organization_status`
before sending a read and gives upgrade/browser guidance when absent.

A directory reply uses `OK count=N`, then `ORGANIZATION` and/or `APPLICATION`
records, then `END`. A page contains at most 50 records. A nonfinal page has
`next=<cursor>` on the OK header; the final page omits `next`. Pass the returned
cursor unchanged to retrieve the following page. Cursors are opaque private
ordering keys; clients must not parse them or treat them as authority. Empty
directories still require `END`. Invalid cursors return `ERR invalid_cursor`.
Organization fields are `id`, `handle`, `name`, `role`, `verification`, and `status`.
`verification` is `unverified`, `approved`, or `revoked`; operational `status`
is `active` or `disabled`. Organizations default to `unverified` and become
approved only through Bibleit admin checklist review. Membership presence does
not override the approval gate.

Application fields are `application_id`, `name`, `domain`, `state`,
`next_action`, and `verification_url`; `organization_id` is absent until approval.
An application never carries an operational organization identity before approval.
The status command returns one `OK` record with the same application fields,
without `count` or `END`. State strings and unknown future fields are retained.
Verification URLs are safe browser navigation links, without email proof tokens.
No contacts, challenge secrets, private reviewer notes, or another principal's
applications are exposed. Final approval happens in Bibleit admin after evidence
and the reviewer checklist are complete; this slice has no public mutations.

`ERR unauthorized` covers missing authentication; `ERR forbidden` covers missing
read authority. Unknown and foreign application IDs both return `ERR not_found`
to avoid disclosing private applications. Directory and application status reads
succeed for pending requests; an attempted operating action can instead return
`ERR organization_pending_verification`, preserving that server code and existing
forbidden exit category 4. `organization_suspended` uses the same exit category.
Capability discovery and responses do not establish permission to operate an
organization.

The response format stays protocol v1. Generic decoders retain the new record
kinds and string fields, and CLI JSON remains schema version 1. Do not change
existing `LIVE` command names or personal Bible/verse Live behavior. Public
application submission, challenge mutations, quota commands, and Studio
presentations require separately agreed extensions. Membership is invitation-only;
there is no request-to-join API. Applications are read-only: corrections require
canceling in the browser and making a new request with fresh email/DNS proof.
The terminal state `organization_deleted` has `next_action=none` and no
`organization_id`, while retaining the application history navigation link.
Dedicated browser/JSON submission can return HTTP 202; command/SSH replies keep
finite protocol framing rather than substituting a JSON body.

Foundation and Studio deployment target a fresh database reset, without legacy
organization grace or personal-presentation transfer compatibility. The reset
must happen only against an explicitly identified deployment target; these
contract changes do not wipe any database. Coordinate the fresh schema and
additive public read contract before SDK/CLI releases. Keep personal Bible/verse
Lives available after the rollout; presentations belong to approved organizations.
