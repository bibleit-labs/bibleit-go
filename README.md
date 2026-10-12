# Bibleit Go client

A dependency-free Go client for reading and searching Bible translations,
managing Lives, and inspecting your Bibleit account. It connects to the project-managed service at `https://bibleit.app`.

The client is released independently from the [Bibleit CLI](https://github.com/bibleit-labs/bibleit-cli).
Use the Go toolchain specified in [go.mod](go.mod). The CLI imports this library;
the library does not depend
on CLI configuration or credential storage.

## Installation

The repository is currently private: configure Git access with an authorized
GitHub account and mark the module private before downloading it.

```sh
export GOPRIVATE='github.com/bibleit-labs/*'
go get github.com/bibleit-labs/bibleit-go
```

`GOPRIVATE` controls module proxy/checksum lookup; it does not authenticate Git.
A sibling CLI checkout or Go workspace is not required.
After public access is enabled, ordinary `go get` can use the public module path.

## Quickstart

Supply an **API token** to `bibleit.NewClient(apiToken)`. The application chooses
where to obtain the token; the library does not read environment variables.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    bibleit "github.com/bibleit-labs/bibleit-go"
)

func main() {
    client, err := bibleit.NewClient(os.Getenv("BIBLEIT_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    reading, err := client.ReadVerses(context.Background(), bibleit.Reference{
        Translation: "web",
        Book:        "John",
        Chapter:     3,
        Verse:       16,
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, verse := range reading.Verses {
        fmt.Println(verse.Text)
    }
}
```

Use a translation enabled for your account; `web` is an example.
The translation catalogue supplies book identifiers and names, which may be localized.
Omit `Verse` to read a chapter, or both `Chapter` and `Verse` to read a book.

The constructor returns `(*Client, error)` and validates the API token before
network access. Constructor errors do not include its value.

## API overview

| Methods | Purpose |
| --- | --- |
| `Ping`, `GetServerInfo`, `Help` | Inspect service availability and capabilities |
| `GetIdentity`, `GetAccountInfo`, `ListQuotas` | Inspect identity, account details, and limits |
| `ListTokens` | Inspect active personal token metadata |
| `ListTranslations`, `GetTranslationInfo`, `GetTranslationCatalog` | Discover translations and their references |
| `ReadVerses`, `SearchVerses` | Return typed readings and search results |
| `CreateLive`, `ListLives`, `GetLive`, `SetLive` | Create, discover, and configure Lives |
| `GetLiveStats`, `GetLiveStack`, `PushLive`, `PopLive`, `ControlLive` | Inspect and control Lives and their stacks |

`Read`, `Search`, and `Identity` return protocol responses for callers that need
records directly. `Execute` accepts validated command builders for additional
operations such as translation add/remove and Live secret management. There is
no arbitrary protocol-string constructor. `Command.String()` is intended for
transport adapters and may contain a Live secret; avoid logging it.

Typed results preserve `Raw` records and unknown fields. Parsers reject malformed
numbers/booleans, conflicting duplicates, inconsistent counts, and invalid stack
positions instead of silently accepting them. Unknown status, role, permission,
and error-code strings remain available.

### Translation metadata and catalogue

`GetTranslationInfo(ctx, slug)` returns names, edition, update label, rights,
attribution, and trademark details when supplied.
`GetTranslationCatalog(ctx, slug)` returns ordered books and chapters.
Book IDs are server identifiers, not array positions; IDs and chapter numbers may be sparse.
Metadata availability does not guarantee installation or account enablement.

A catalogue chapter's `Verses` is the highest source verse identifier, not a dense
count. Editions may omit verse numbers: an omitted individual verse returns
`verse_not_found`, while a chapter read contains the available rows.
Source numbering is preserved.

### Readings and Lives

`Reading.Book`, `Chapter`, and `Verse` reflect coordinates supplied by the response
header. Chapter/book reads and search results contain verse text without invented
per-verse coordinates. Search coordinates are normally nil. Empty searches and
stacks return empty slices.

Optional Live management fields use pointers so absence differs from false, zero,
or an empty string. Public list/create replies may omit owner-only metadata.
`ControlLive` accepts START, STOP, PAUSE, RESUME, CLEAR, and DELETE and returns the
acknowledgement without a follow-up request. Push/pop retain all verse entries,
including repeated translation names.

### Token metadata

`ListTokens(ctx)` returns active personal credential metadata: IDs, sources,
issuers, labels, scopes, timestamps, and retirement status when supplied.
It returns no bearer values and is not a credential history; revoked and expired
tokens are excluded. Missing scopes remain nil; an explicitly empty list means
the credential inherits the actor's current permissions. Scopes restrict access
and do not guarantee effective permissions.

## Behavior and current limits

- The service endpoint is fixed. Constructors and environment variables cannot
  override it. HTTPS certificate verification stays enabled and redirects are blocked.
- All requests accept a context. The default HTTP timeout is 15 seconds.
  Operations are never retried automatically; a decoding failure after a mutation
  does not imply rollback.
- Use `NewClientWithHTTPClient(apiToken, httpClient)` or
  `NewAuthClientWithHTTPClient(httpClient)` to customize transport and timeout.
  The supplied `http.Client` value is copied; its `Transport` and `Jar` remain
  shared. A custom client controls its own timeout, but redirects remain blocked.
- Commands are limited to 8192 bytes, JSON request bodies to 16384 bytes, and
  finite responses to 2 MiB. Oversized or incomplete responses fail explicitly.
- Request arguments support quoted whitespace, but embedded double quotes, CR,
  LF, and NUL are rejected. Backslashes are literal in requests. Escaped response
  strings do not define a request escape syntax.
- Live secrets are generated by the server through CREATE/ROTATE. Custom SECRET
  SET is unsupported and rejected by the client.
- HTTP subscriptions return `streaming_requires_ssh`. OpenSSH execution and
  subscriptions are provided by the CLI, not this HTTP library.
- Browser authorization and credential storage belong to the application.
  The library provides code exchange and credential revocation.

Use `errors.As` with `*bibleit.ServerError` for protocol errors and `RetryAfter()`
for rate-limit metadata. `*bibleit.HTTPError` represents HTTP failures without a
usable protocol error and preserves available JSON error codes/metadata.

See the [protocol contract](contract/README.md) and
[feature coverage](contract/features.md) for compatibility details.

## Browser authorization

Use `NewAuthClient()` before an API token exists. Your application owns PKCE
verifier generation, browser authorization, callback-state validation, and storage.
Exchange an approved code, then construct an authenticated client:

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
```

`AuthorizationResult` and its `Fields` contain the bearer value; avoid logging them.
An authorization code is single-use, including when an invalid verifier consumes it.
Omitting `DeviceName` uses the server's default label.

Call `client.RevokeCredential(ctx)` to revoke the credential. Only HTTP 204 is
accepted as success. Revocation does not erase local storage or modify the client;
subsequent authenticated requests with that credential are expected to fail.

## Development

From this repository, using the toolchain in go.mod:

```sh
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

Unit tests use in-memory transports. Shared protocol fixtures live in
`contract/fixtures/responses-v1.json`. Native CI covers Linux, macOS, and Windows.
The CLI repository owns isolated real-server integration checks; the server
repository is separate and private.

See the [changelog](CHANGELOG.md) and [contributing guide](CONTRIBUTING.md).
Contributions use Conventional Commits. Licensed under [MIT](LICENSE).

The MIT software license does not grant rights to Bible translations returned by
the service, account access or quotas. Report vulnerabilities privately through
[SECURITY.md](SECURITY.md).
