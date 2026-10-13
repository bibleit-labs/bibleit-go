# Changelog

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
