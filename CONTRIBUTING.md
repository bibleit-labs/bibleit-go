# Contributing to bibleit-go

The SDK is MIT licensed and the repository currently remains private. Authorized
contributors can open issues and focused pull requests; discuss larger API changes
before implementation. Public contributions await approved launch. Report
vulnerabilities through [SECURITY.md](SECURITY.md), and redact credentials/account
or private server data from issues, source, examples and fixtures.

## Development

Use Go **1.27.1**. The module is github.com/bibleit-labs/bibleit-go and has no
external Go module dependencies. A CLI/server checkout is not required:

```sh
GOWORK=off go test -race -count=1 ./...
GOWORK=off go vet ./...
GOWORK=off go build ./...
```

Normal tests use in-memory transports and synthetic protocol fixtures; optional
server integration cases are skipped unless explicitly enabled. Native CI covers
Linux, macOS and Windows. Review links, commands and content for docs-only changes.

Credentials are explicit, and the service address is project-managed. There are
no constructor/environment endpoint overrides. Preserve HTTPS verification,
context cancellation and the no-automatic-retry behavior. Browser orchestration,
credential storage, terminal presentation and OpenSSH integration belong in the
CLI; the library must not import it. Add meaningful tests for changed protocol
behavior, including malformed responses and unknown fields when applicable.

[Contract coverage](contract/features.md) and [shared fixtures](contract/README.md)
record supported behavior and tested baselines. Synchronize fixture changes
explicitly with the CLI. Authorized maintainers may use the CLI's isolated
container harness for private server integration; see the CLI native-validation
guide. Passing fixture tests does not establish production/provider behavior.

## Commits and releases

Use Conventional Commits, for example `fix(protocol): reject incomplete envelopes`
or `docs: clarify typed results`. Use `!` or BREAKING CHANGE for incompatible
changes, and branches without a codex prefix. Explain the resulting behavior and
relevant validation in pull requests.

SDK v0.0.1 is tagged privately and is the CLI's pinned dependency. SDK and CLI
release versions are independent. Consumers choose when to update; a branch
push is not a stable release. Keep supported operations, Go version, tested server
baseline and compatibility changes documented. Release tags and repository
visibility changes require approval. CLI v0.0.2 installer checks do not change
the SDK version; pending Windows installer/Scoop checks are not SDK test failures.

Retain the MIT copyright/permission notice when redistributing SDK source.
The software grant does not cover translations obtained through the service,
account access or service quotas. Applications distributing compiled binaries
must also retain applicable Go/dependency notices.
