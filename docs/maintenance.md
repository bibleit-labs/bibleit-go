# Maintaining the Bibleit Go client

Normal contributor setup and offline tests are in [CONTRIBUTING.md](../CONTRIBUTING.md).
They use in-memory transports and do not require private service source or an account.

## Compatibility validation

Keep the [protocol reference](../contract/README.md), [feature coverage](../contract/features.md)
and shared fixtures synchronized with client behavior. SDK native CI covers Linux,
macOS and Windows; package installation checks are maintained by the CLI and tap.
Passing protocol fixtures does not establish production or provider behavior.

Local service integration procedures are maintained in private operations
documentation and are not part of public contributor setup.
[Historical compatibility records](maintenance/compatibility-history.md) preserve
prior audit, integration and deployment narratives moved out of the contract README.
Their commands and release gates describe past revisions and are not current instructions.

## Releases and audits

SDK and CLI versions are independent. The CLI pins a tagged SDK release, while
other consumers choose when to update. A branch push is not a stable release.
Keep API compatibility, supported operations, Go requirements and fixture changes
documented in the changelog. Obtain owner approval before release tags or repository
visibility changes; keep published tags immutable.

The Publish Go package documentation workflow runs on version-tag pushes and
published GitHub releases. It
downloads the published version through proxy.golang.org with checksum
verification, requests pkg.go.dev indexing, and verifies the versioned
documentation page. No publishing credentials are required. For an existing
release, dispatch `.github/workflows/pkg-go-dev.yml` with its `version` input.
Indexing is asynchronous; a failed availability check can be rerun without
changing the release tag. See [pkg.go.dev's indexing process](https://pkg.go.dev/about#adding-a-package).

The [cross-repository launch record](https://github.com/bibleit-labs/bibleit-cli/blob/main/docs/public-launch.md),
[security audit](https://github.com/bibleit-labs/bibleit-cli/blob/main/docs/security-audit.md)
and [licensing audit](https://github.com/bibleit-labs/bibleit-cli/blob/main/docs/licensing-audit.md)
record the reviewed publication scope. Recheck relevant rights and sensitive material
when source, dependencies or fixtures change. Keep raw credentials, private service
data and unredacted reports outside Git; use [SECURITY.md](../SECURITY.md) for reporting.

Retain the MIT notice when redistributing SDK source and applicable Go/dependency
notices in compiled binaries. Software licensing does not grant translation rights,
service access or quotas. None of these maintainer checks changes private service visibility.
