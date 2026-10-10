# Contributing to bibleit-go

This is a public MIT-licensed Go client maintained in `bibleit-labs`.
Focused bug reports, documentation improvements, and pull requests are welcome.
Discuss larger API changes in an issue first. Do not include credentials or
private server data in issues, source, or test fixtures.

## Development and validation

Use Go **1.27.1**. The module path is `github.com/bibleit-labs/bibleit-go`.
Run `go test -race -count=1 ./...`, `go vet ./...`, and `go build ./...`
for executable changes. CI validates Windows, Ubuntu, and macOS. Review links
and content for documentation-only changes.

Keep credentials explicit and the service endpoint project-managed. Do not
add constructor, environment or config endpoint overrides. CLI profiles, terminal presentation,
browser orchestration, and process adapters belong in `bibleit-cli`; this
library must not import the CLI. Add tests for meaningful protocol behavior,
including unknown fields and malformed responses when relevant.

Protocol fixtures in `contract/` are versioned snapshots. Synchronize changes
explicitly with the CLI and other client repositories. The CLI's
`../bibleit-cli/scripts/check_server.py --integration --go-client .` can
exercise the real-server tests from this checkout when its required server,
Erlang and Docker dependencies are available. Standard unit tests skip those
live-server cases. State the tested server baseline rather than claiming
coverage of all server APIs.

## Commits and pull requests

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):
`type(scope): description`, with an optional scope. For example:

- `feat(tokens): expose typed token metadata`
- `fix(protocol): reject incomplete response envelopes`
- `docs: explain client configuration`
- `ci: validate native Windows tests`

Use `!` or a `BREAKING CHANGE:` footer for incompatible changes. Describe the
customer-visible outcome and relevant validation in pull requests. Do not use
`codex/` as a branch prefix.

## Releases

Release this module independently from the CLI and other language clients.
Keep its supported operations, minimum Go version, and tested server baseline
in the documentation. Use semantic-version tags for reviewed releases and
explain compatibility changes in release notes. Consumers, including the CLI,
choose when to update their pinned dependency; branch publication alone is
not a stable API release.

Publication is paused by the user. Keep commits and release work local until
publication is explicitly resumed. `ServerEndpoint` is a loopback development
placeholder; set the real production HTTPS address before shipping a release.

The standard constructor is `NewClient(apiToken)` with a validation error return.
Use `NewClientWithHTTPClient` for transport customization. Real-server integration
now exercises those public constructors at the fixed endpoint; it does not change
client destinations. The CLI's container harness provides an isolated namespace
when the developer's port 8080 is in use.
