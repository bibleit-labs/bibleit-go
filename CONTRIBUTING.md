# Contributing to bibleit-go

The SDK is MIT licensed. Open issues and focused pull requests for contributions;
discuss larger API changes before implementation. Report
vulnerabilities through [SECURITY.md](SECURITY.md), and redact credentials/account
or private server data from issues, source, examples and fixtures.

## Development

Use Go **1.27.1**. The module is github.com/bibleit-labs/bibleit-go and uses
a pinned oapi-codegen runtime dependency; generation tooling is isolated in
`tools/go.mod`. A CLI/server checkout is not required:

```sh
GOWORK=off go test -race -count=1 ./...
GOWORK=off go vet ./...
GOWORK=off go build ./...
```

Normal tests use in-memory transports and synthetic protocol fixtures; optional
server integration cases are skipped unless explicitly enabled. Native CI covers
Linux, macOS and Windows. Review links, commands and content for docs-only changes.

Use named Go types for domain models, nested objects and decoder payloads rather
than inline structs. Keep generated model conventions in the generation config
and templates so regeneration preserves them.

Credentials are explicit, and the service address is project-managed. There are
no constructor/environment endpoint overrides. Preserve HTTPS verification,
context cancellation and the no-automatic-retry behavior. Browser orchestration,
credential storage, terminal presentation and OpenSSH integration belong in the
CLI; the library must not import it. Add meaningful tests for changed protocol
behavior, including malformed responses and unknown fields when applicable.

[Contract coverage](contract/features.md) and [shared fixtures](contract/README.md)
record supported behavior and tested baselines. Synchronize fixture changes
explicitly with the CLI. Passing fixture tests does not establish
production/provider behavior.

## Commits and reviews

Use Conventional Commits, for example `fix(protocol): reject incomplete envelopes`
or `docs: clarify typed results`. Use `!` or BREAKING CHANGE for incompatible
changes, and branches without a codex prefix. Explain the resulting behavior and
relevant validation in pull requests.

See the [maintainer guide](docs/maintenance.md) for independent SDK releases,
audit records and historical compatibility evidence.

Retain the MIT copyright/permission notice and applicable THIRD_PARTY_NOTICES
when redistributing SDK source or compiled applications.
The software grant does not cover translations obtained through the service,
account access or service quotas. Applications distributing compiled binaries
must also retain applicable Go/dependency notices.
