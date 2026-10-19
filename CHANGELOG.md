# Changelog

## Unreleased

## v0.2.0 — 2026-10-20

- Add bounded whole-book cursor pagination (default 100, maximum 500) and lazy
  personal/organization Go page iterators for complete-book traversal.
  Detect unsafe traversal and preserve typed invalid, mismatched and stale cursor errors.

- Add personal quota/token metadata, organization directory/application status,
  personal library writes, whole-book reads and installed catalogue resources.
- Add conditional Live settings, statistics, retained-stack and audience-secret
  resource operations and atomic workspace bulk deletion.
- Keep `CreateLive.Name` source-compatible while allowing an omitted name.
- Preserve explicit credential scopes, ETags, native JSON and no mutation retries.
- Replace the Python REST generator with pinned `oapi-codegen` v2.8.0; generate
  OpenAPI 3.1 models and 88 resource-operation bindings from the finalized contract.
- Expose generated enum and union helpers, including typed Live event snapshots.
  Callers accessing former raw union fields should use the generated helpers.
- Preserve legacy SSH Scripture text while exposing optional localized content,
  references and complete source coordinates; reject partial coordinate metadata.
- Reject negative and overflowing retained-stack pop counts; preserve zero as
  the SDK's existing default-count shorthand.


## v0.1.0 — 2026-10-19

- Add the versioned `rest` package with 66 REST bindings, 83 schema models,
  the pinned canonical contract and a reproducible dependency-free generator.
- Preserve Problem Details and response metadata, stream binary transfers,
  forward ETags/idempotency keys, and decode bounded versioned Live event polls.
- Add a separate registered OAuth client for S256 consent URLs, metadata, code
  exchange, refresh rotation and revocation on the website origin.
- Remove the command-style HTTP transport. Command helpers require an explicit
  SSH executor; bearer credential clients only support CLI token revocation.
- Migrate examples and integration tests to REST; retain SSH decoding coverage.

## v0.0.4 — 2026-10-14

- Export `ErrIncompleteTerm` so streaming clients can buffer a multiline Live
  presentation field and retry `ParseRecord` without mistaking it for a failed
  subscription. Existing finite-response decoding remains unchanged.

## v0.0.3 — 2026-10-13

- Remove internal deployment and infrastructure records from public source and history.
- Retract v0.0.1 and v0.0.2 because their cached module archives retain those records.
- Retain the v0.0.2 public API and runtime behavior.

## v0.0.2 — 2026-10-13

- Add read-only organization directory and application status command builders and typed results, pagination cursors, and organization HELP. Retain future states and terminal organization_deleted without an organization ID.
- Organization membership is invitation-only; verification requests remain browser-managed and approval remains in Bibleit admin.

## v0.0.1 — 2026-10-11

Initial Go client release for Bibleit, requiring Go 1.27.1.

- Explicit API-token construction with `bibleit.NewClient(apiToken)` and the fixed
  HTTPS origin `https://bibleit.app`; no endpoint overrides or credential loading.
- Typed identity, API-token metadata, translations and reference catalogues,
  reading/search and Live results, retaining raw records and unknown fields.
- Browser-auth code exchange and credential revocation methods.
- Strict finite-response decoding, contextual requests, bounded request/response
  sizes, disabled redirects and explicit rate-limit/error metadata.

HTTP subscriptions require SSH; OpenSSH execution, streaming, browser opening
and credential persistence belong to the CLI. Requests support quoted whitespace
but reject embedded double quotes and control characters. Live secrets are
server-generated; custom SECRET SET is unsupported. Catalogue verse counts are
the highest source identifier, so omitted verse numbers remain gaps. See README
for the full behavior and current limits.

Historical service deployment evidence is maintained privately. Public client
compatibility fixtures remain in contract/.

This historical release is retracted. Use v0.0.3 or newer from the public module proxy.
