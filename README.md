# Bibleit Go client

A dependency-free Go client for Bibleit Server protocol v1. Import it as:

```go
import bibleit "github.com/bibleit-labs/bibleit-go"
```

This module is released independently from [bibleit-cli](https://github.com/bibleit-labs/bibleit-cli).
It requires Go 1.27.1. The CLI imports this library; the library does not depend
on the CLI or access its profiles. Python, Ruby, and Rust clients have separate repositories.

Publication is paused. Use the local checkout for development; no release
version is available yet. The project-managed service address is currently
`http://localhost:8080`, and will be replaced with the production HTTPS address.
Applications cannot override it through configuration or environment variables.

## Usage

```go
client, err := bibleit.NewClient(apiToken)
if err != nil {
    return err
}
result, err := client.Read(ctx, bibleit.Reference{
    Translation: "web",
    Book: "Song of Songs",
    Chapter: 2,
    Verse: 1,
})
if err != nil {
    return err
}
for _, record := range result.Records {
    fmt.Println(record.Fields["text"])
}
```

`Read`, `Search`, `Ping`, `Identity`, and `ListTranslations` provide convenient
methods. `GetServerInfo`, `Help(ctx, topic)`, `GetAccountInfo`, and
`ListQuotas` return typed results. Use an empty help topic for root discovery;
the server filters discovery by the authenticated principal's authority.

Each typed result retains `Raw`, including original wire records and unknown
fields. Quota entries expose a `QuotaLimit`: zero is a finite limit, while
`Unlimited` is explicit. Counts are non-negative 64-bit integers; malformed
numbers and mismatched record counts fail rather than becoming zero. Account
notification preferences remain opaque Erlang-term text in the raw fields.

Other customer operations use validated command builders:

```go
command, err := bibleit.CreateLiveCommand("Sunday Service")
if err != nil {
    return err
}
result, err := client.Execute(ctx, command)
```

Builders cover account translation add/remove, Live lifecycle and settings,
secrets, and stack operations. They expose no arbitrary protocol constructor.
`Command.String()` exists for transport adapters and may contain a Live secret;
do not log it indiscriminately.

## Behavior and current limits

- Credentials are explicit. The endpoint is the project-owned `ServerEndpoint`
  constant; no constructor or environment override is provided. The client
  does not access CLI profiles, launch a browser, or load API tokens or service
  addresses from environment variables. HTTP transport/proxy behavior is owned by
  `net/http` or the supplied transport.
- The service uses a loopback HTTP placeholder during development. Set the
  production HTTPS constant before release; remote HTTP remains rejected.
- The default HTTP timeout is 15 seconds. Use `NewClientWithHTTPClient` or
  `NewAuthClientWithHTTPClient` with a `*http.Client` to customize transport and
  timeouts. Its value is copied and redirects are disabled; its `Transport` and
  `Jar` remain shared. An explicit custom client controls its own timeout.
- Commands are limited to 8192 bytes and JSON request bodies to 16384 bytes,
  matching the reviewed server. Oversized requests fail before network access.
- All requests accept a context. Operations are never retried automatically.
- Responses retain wire lines and parsed records. V1 field values remain
  strings, and unknown fields are retained in the typed models.
- `*ServerError` retains unknown error codes and metadata. Use `errors.As` and
  `RetryAfter()` for rate-limit information. `*HTTPError` reports HTTP failures
  that do not provide a usable protocol error; JSON error codes and metadata are retained.
- Finite responses are limited to 2 MiB; oversized and incomplete responses
  fail explicitly.
- The HTTP transport rejects subscriptions with `streaming_requires_ssh`.
  OpenSSH execution and subscriptions remain CLI adapters in this increment.
- The server's request tokenizer supports quoted whitespace but not embedded
  double quotes. Builders reject double quotes, CR, LF, and NUL. Backslashes
  are preserved literally in requests; response strings use escaped quotes,
  backslashes and escaped control characters. Response escaping is not a request
  escape syntax. The verified server rejects malformed quote segments, empty
  quoted arguments and attempted escaped embedded quotes with `bad_command`.
  Live secrets are server-generated through CREATE/ROTATE; custom SECRET SET
  remains unsupported and is rejected by the client.
- Browser orchestration and credential persistence remain in the CLI; code
  exchange and revocation use library methods.

Run `go test ./...` from the repository root. Shared response fixtures live in
`contract/fixtures/responses-v1.json`; see `contract/README.md` for their schema.


## Translation metadata and reference catalogue

`GetTranslationInfo(ctx, slug)` returns the short/full name, edition, update
label, rights status/label/URLs, attribution and trademark notice where supplied.
Optional fields stay empty when omitted; unknown fields remain in Raw.

`GetTranslationCatalog(ctx, slug)` returns ordered books and their chapters with
verse counts. `CatalogBook.ID` is the server book identifier, not an array position.
IDs and chapter numbers may be sparse. Names are supplied by the server and can
be localized. Use the returned ID or name when constructing a Reference.
The catalogue describes installed indexed data; metadata availability alone
does not imply installation or enablement in the personal translation library.

The typed parser checks book/chapter counts, parent IDs, duplicate IDs/numbers
and positive chapter/verse values. It retains unknown book/chapter fields.
Raw Execute remains available for callers inspecting newer protocol shapes.

Real-server metadata and catalogue verification pass. The assessment session
fixed an encoder mismatch between atom and binary field names; shared encoder
fixtures now cover binary metadata keys as well.

For example, use the catalogue to inspect the references actually installed:

```go
catalog, err := client.GetTranslationCatalog(ctx, "web")
if err != nil {
    return err
}
for _, book := range catalog.Books {
    for _, chapter := range book.Chapters {
        fmt.Printf("%d %s %d:1-%d\n",
            book.ID, book.Name, chapter.Number, chapter.Verses)
    }
}
```



## Typed identity, reading and Live results

Raw `Identity`, `Read`, `Search` and `Execute` remain available. These additive
methods return domain models and retain Raw/Fields for original records and
future fields:

| Method | Result |
| --- | --- |
| GetIdentity | Actor, display identity/provider, roles, permissions, SSH fingerprint |
| ReadVerses / SearchVerses | Translation, optional response coordinates and verse texts |
| CreateLive / GetLive / SetLive | Live state and optional management metadata |
| ListLives | Ordered Live states |
| GetLiveStats | Revision, duration, stack counts and connection details |
| GetLiveStack | Ordered entries with positions and verse payloads |
| PushLive / PopLive | Verse payload or clear acknowledgement |
| ControlLive | Start/stop state, pause/resume flag, clear/delete event |

`Reading.Book/Chapter/Verse` describe coordinates supplied by the response header.
Chapter/book/search records supply text only; the client does not extract or
invent per-verse coordinates from that text. Search coordinates are normally nil.
Empty searches and stacks return empty slices.

`LiveState.Owned`, `Protected`, `Owner`, `CreatedAt` and other optional fields
use pointers so absence differs from false, zero, or an empty string. Public
Live list/create replies omit management metadata that owner INFO may supply.
Unknown status, role, permission and access strings remain intact.

ControlLive accepts START, STOP, PAUSE, RESUME, CLEAR and DELETE only. Its result
describes the server acknowledgement without a follow-up request. Push/pop
preserve all verse entries, including repeated translation names in a stack.
Decode failures after a mutation do not imply rollback; mutations are not retried.

The typed models reject malformed numeric/boolean fields, mismatched counts,
duplicate Live IDs, inconsistent connection totals and invalid stack positions.
The read example in `examples/read` now uses ReadVerses.


## Typed token metadata

`ListTokens(ctx)` sends `ACCOUNT TOKEN LIST` and returns `TokenList` with the
actor, ordered `TokenMetadata` entries and Raw. Entries expose ID, source,
issuer, label, scopes, Unix-second issue/last-use/expiry timestamps and retirement
status. Optional fields use pointers; absent scopes stay nil while an explicitly
empty scope list means the credential inherits the actor's current permissions.
Scopes describe a restriction, not a guarantee of effective permissions.
Unknown source/permission names and fields remain available.

This is a list of currently active personal credentials, not a credential
history: the server excludes revoked, expired and expired rotation-grace tokens.
The endpoint returns metadata only, never bearer values. Token creation,
rotation and revocation are not added by this increment.

The client validates record counts/types, unique IDs and numeric/boolean values.
Response control-character escapes are decoded without splitting a label into
protocol records. Requests retain their existing, stricter quoting rules.

The assessment fixed the scoped-token encoder to use permission CSV. Real HTTP
verification now covers manual and scoped credentials, restricted-list denial,
labels with escaped controls, and expired/revoked credential exclusion.

## Authorization transport

Use `NewAuthClient()` before a bearer credential
exists. The application owns PKCE verifier generation, browser authorization,
callback-state validation and storage. Then exchange the approved code:

```go
auth, err := bibleit.NewAuthClient()
if err != nil {
    return err
}
credential, err := auth.ExchangeCode(ctx, bibleit.CodeExchange{
    Code: code, Verifier: verifier, DeviceName: "My application",
})
if err != nil {
    return err
}
client, err := bibleit.NewClient(credential.AccessToken)
if err != nil {
    return err
}
// Later, explicitly revoke this credential:
err = client.RevokeCredential(ctx)
```

`AuthorizationResult` preserves actor, identity/provider, token ID, bearer value
and unknown JSON fields. Its Fields also includes the bearer value. Applications
must decide how to store it and avoid logging the result.

Both transports share endpoint validation, a copied HTTP client, default timeout
and redirect rejection. Methods accept contexts, limit response bodies to 2 MiB,
and never retry. Exchange forms are limited to the server's 16384-byte limit.
Omitting DeviceName selects the server default label. HTTP errors retain status,
error code and JSON/protocol metadata; use `errors.As` with `*HTTPError`.

A code is one-use, and even an invalid verifier consumes it on the current server.
RevokeCredential succeeds only on HTTP 204. It does not erase local credentials
or modify the client; another request after revocation is expected to fail.
The CLI keeps credentials when logout fails and clears an already unauthorized
credential or one whose revocation was acknowledged.

Real HTTP checks cover successful exchange/use, replay rejection, bad-verifier
consumption, revocation and post-revocation authentication rejection.


## Contributing and license

This is a public open source project. See [CONTRIBUTING.md](CONTRIBUTING.md)
for development, validation, Conventional Commits and release expectations.
The client is distributed under the [MIT license](LICENSE).


## Fixed service address and local development

Commands, browser-code exchange, and revocation construct requests at the same
project-managed service. The optional HTTP client still configures transport and
timeouts; it does not add a service URL option. Redirects remain blocked.
The server repository stays separate and private. Future Python, Ruby and Rust
clients follow the same fixed-service policy.

Tests use in-memory transports. Real-server integration uses the public constructors and a disposable server
at the fixed endpoint; client URLs and destinations are not overridden. Keep this checkout beside
`bibleit-cli` and use a local workspace while publication is paused. Do not push
commits, tags or releases until the user resumes publication.


## API token and custom HTTP transport

Use the term **API token** for the bearer credential supplied to this client.
It matches the server's scoped, expiring token resources. `NewClient(apiToken)`
returns `(*Client, error)` and rejects invalid credentials before network access.
The API token is kept private; constructor errors do not include its value.

For HTTP transport or timeout customization, use
`NewClientWithHTTPClient(apiToken, httpClient)` or
`NewAuthClientWithHTTPClient(httpClient)`. The supplied client is copied, redirects
are blocked, and the service endpoint remains fixed. No operation is retried.
This keeps the ordinary constructor to a single credential argument, inspired by
[Resend's constructor pattern](https://github.com/resend/resend-go/blob/main/resend.go).

To run real-server checks while the host service is using port 8080, use the CLI's
`python3 ../bibleit-cli/scripts/check_server_container.py`. That runner builds a
frozen source snapshot in an isolated container namespace with its own temporary
database; it does not expose host ports or require client endpoint configuration.
