# Bibleit Go client

A Go SDK for Bibleit REST resources, registered OAuth clients,
and SSH command construction/decoding. HTTP resources use the pinned
`https://api.bibleit.app/v1` service. Browser authentication uses
`https://bibleit.app`. There is no command-style HTTP transport.

The module is released independently from the [Bibleit CLI](https://github.com/bibleit-labs/bibleit-cli).
Use the Go version in [go.mod](go.mod). The SDK does not read CLI configuration,
open browsers, store credentials, or depend on the CLI.

## Installation

```sh
go get github.com/bibleit-labs/bibleit-go
```

## Resource client

Import `github.com/bibleit-labs/bibleit-go`. The client includes all 88 REST
operations, generated schemas, response metadata, conditional writes, binary
streams, versioned Live event polling and lazy whole-book page iterators. See [REST usage](docs/api.md) and
the [runnable reading example](examples/read/main.go).

```go
client, err := bibleit.NewClient(token)
if err != nil { return err }
response, err := client.GetPersonalIdentity(ctx)
if err != nil { return err }
defer response.Body.Close()
var identity bibleit.PersonalIdentity
if err := response.Decode(&identity); err != nil { return err }
fmt.Println(identity.ID, identity.PersonalAccountID)
```

Credentials are explicit: finite personal `bt_`, delegated `oa_`, or organization
`svc_` tokens. The server enforces their separate scopes and resource boundaries.
`ProblemError` exposes HTTP status, stable code, request ID and response headers.
Operations never retry automatically; redirects are blocked and TLS verification
is enabled. The host cannot be overridden through configuration or environment.

## SSH protocol helpers

The `bibleit` package contains validated command builders, response parsers
and typed domain helpers for SSH. It does not open SSH connections. Applications
supply their own SSH executor to `NewSSHClient`:

```go
client, err := bibleit.NewSSHClient(func(ctx context.Context, cmd bibleit.Command) (bibleit.CommandResponse, error) {
    // Execute cmd.String() over your verified SSH connection, then decode stdout.
    return bibleit.DecodeCommandResponse(cmd, stdout)
})
```

The executor owns SSH authentication, host-key verification and connection
lifetime. The CLI uses OpenSSH to `bibleit.app:22`. Persistent Live subscriptions
use the streaming protocol directly, rather than the finite executor.
[Protocol fixtures](contract/README.md) document quoting and response framing.
No command-to-REST adapter is provided. Bearer credential clients cannot execute
commands. `NewClient` returns a resource client; `NewSSHClient` returns an
`SSHClient` with command and domain helpers. SSH decoders return `CommandResponse`;
SSH Live collection results use `SSHLiveList`, `SSHLiveStack` and `SSHLiveStatistics`.

## Browser login and registered OAuth

`NewAuthClient` / `NewAuthClientWithHTTPClient` exchange an approved CLI browser
code with `ExchangeCode`. The application validates callback state, supplies the
PKCE verifier and securely stores the result. Revoke that browser-issued CLI
credential by constructing a resource client with `NewClient(token)` and calling
`client.RevokeCredential(ctx)`. The token exchange and logout endpoints remain supported.

For registered delegated OAuth clients use the separate `bibleit.NewOAuthClient`:
S256 consent URLs, metadata, code exchange, refresh rotation and revocation stay
on the website origin. Codes and refresh tokens are single-use. The SDK never
retries an ambiguous exchange or refresh attempt.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), and the
[REST contract regeneration instructions](docs/api.md#updating-the-contract).

```sh
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

Licensed under [MIT](LICENSE).

Generated-code and dependency redistribution notices are retained in
[THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES), alongside the project MIT license.
