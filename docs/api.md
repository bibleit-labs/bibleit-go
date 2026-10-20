# Resource client

Use `github.com/bibleit-labs/bibleit-go` for new integrations. It pins
`https://api.bibleit.app/v1`, uses the pinned oapi-codegen runtime, and includes all
88 REST operation bindings and 105 schema models from the canonical OpenAPI
contract. Browser/consent and OAuth protocol endpoints remain outside this
bearer REST client. SSH command helpers require an explicit SSH executor.

```go
client, err := bibleit.NewClient(token)
if err != nil { return err }
response, err := client.GetPersonalIdentity(ctx, nil)
if err != nil { return err }
defer response.Body.Close()
var identity bibleit.PersonalIdentity
if err := response.Decode(&identity); err != nil { return err }
fmt.Println(identity.ID, identity.PersonalAccountID)
```

Supply finite personal `bt_`, registered delegated `oa_`, or organization `svc_`
tokens explicitly. The server enforces scopes and ownership; these credentials
are not interchangeable. The client does not load environment variables, open
browsers or persist tokens. An empty token permits public OpenAPI discovery.

Every generated method accepts explicit path identifiers and `*bibleit.Options`.
JSON bodies can use generated schema models or ordinary Go values. Optional
nullable request fields use `json.RawMessage` to distinguish omitted (`nil`),
explicit null (`json.RawMessage("null")`) and zero. Union schemas expose generated `As...`, `From...` and `Merge...` helpers;
choose the concrete variant using the contract discriminator.

```go
read, err := client.GetLive(ctx, liveID, nil)
if err != nil { return err }
etag := read.Header.Get("ETag")
read.Body.Close()
updated, err := client.UpdateLivePlayback(ctx, liveID, &bibleit.Options{
    IfMatch: etag,
    JSON: map[string]any{"paused": true},
})
if err != nil { return err }
updated.Body.Close()
```

Response exposes `StatusCode`, `Header` and streamed `Body`; callers close every
successful body. `Decode` bounds JSON at 4 MiB. Stream media/imports with
`Options.Body`, `ContentType` and exact `ContentLength`. Stream binary responses
with `io.Copy`, rather than `Decode`. One-use export downloads require an empty
JSON object and cannot be repeated after consumption. Conditional writes forward
exact `IfMatch` validators. `IdempotencyKey` is opt-in for supported creations;
retain it across manual retries. `DestinationToken` sends the additional secret
in `X-Destination-Service-Token` for workspace moves. Never put secrets in queries.

`ProblemError` preserves HTTP status, stable code, request ID, optional server
`Detail` and response headers, with a seconds-based `RetryAfter` helper. Its `Error()`
text omits untrusted server detail; applications choose how to present it. All HTTP redirects are blocked,
cookie jars are disabled, and no operation is automatically retried. Transport
injection retains the fixed host. The default timeout is 40 seconds; supplied
clients retain their own timeout and must allow the 25-second long-poll wait plus
latency. Mutations must not queue behind an outstanding poll.

Collections retain `NextCursor`. Explicitly pass it as `Options.Query["after"]`
with a limit of 1–100 until nil; search instead uses `HasMore` without cursors.
`PollLiveEvents` makes one request, decodes a typed envelope and rejects unsupported
event schema versions. Snapshot cursors are opaque and are never ETags. Poll
again with the returned cursor; reconnect without one to resynchronize. Events
are current snapshots, not a durable or exactly-once event log.

## Whole-book pagination

`ReadPersonalBook` and `ReadBook` make one request and always return a bounded
page: 50 verses by default, or a `limit` of 1–500. Every book response includes
`next_cursor`, an opaque string or `null` on the final page. Pass the returned
cursor as `after` to continue. Keep the same account, translation and book;
page size may change. The server also caps serialized pages at 3 MiB, so a page
may contain fewer verses than its item limit. There is no unbounded whole-book
API response.

```go
response, err := client.ReadPersonalBook(ctx, accountID, "web", "19", &bibleit.Options{
    Query: url.Values{"limit": {"100"}},
})
if err != nil { return err }
defer response.Body.Close()
var page bibleit.ScriptureBookRead
if err := response.Decode(&page); err != nil { return err }
// Use page.NextCursor as Query["after"] for the next request.
```

Use the lazy Go iterators to traverse pages without managing response bodies:

```go
for page, err := range client.ReadPersonalBookPages(ctx, accountID, "web", "19", nil) {
    if err != nil { return err }
    for _, verse := range page.Verses {
        fmt.Println(verse.Chapter, verse.Verse, verse.Text)
    }
}
```

`ReadBookPages` is the organization-scoped counterpart. Both iterators default
to 50 verses per page, copy their options, close bodies before yielding, and
stop when the caller breaks or cancels the context. They reject cursor loops,
unordered or duplicate verse coordinates, and inconsistent translation metadata
with `ErrInvalidBookPage`. An iterator error means the traversal is incomplete;
discard accumulated verses if you need an atomic complete-book result.

Every page is authorized independently. A changed edition returns HTTP 409
`stale_cursor`; restart the read to avoid combining editions. Invalid cursors
return HTTP 422 `invalid_cursor`; cursors for a different account, translation or
book return `cursor_mismatch`. Errors are returned without automatic retries or
silently restarting the traversal. A single source verse that cannot fit returns
HTTP 413 `verse_too_large`, without silently skipping it. These cursors are
unrelated to Live event cursors and ETags.

REST operations accept resource identifiers and structured bodies directly;
this package does not translate commands or produce command-protocol responses.
The SSH command helpers require `NewSSHClient` with a caller-owned
SSH executor. Credential clients cannot execute commands. Browser login and
logout remain independent supported flows.

## Registered OAuth clients

`NewOAuthClient` returns a separate client pinned to `https://bibleit.app`.
`AuthorizationURL` builds the registered callback/state/S256 consent URL;
`Metadata`, `ExchangeCode`, `Refresh` and `Revoke` implement the public-client
protocol. These requests never send REST bearer credentials or browser cookies.
Applications validate callback state and securely store issued tokens. Code and
refresh exchange each make one attempt: reusing a consumed credential may revoke
the whole grant. Existing official CLI login remains supported separately.

## Updating the contract

Copy the reviewed `priv/openapi/bibleit-v1.json` from bibleit-server to
`contract/openapi.json`, then run from the module root:

```sh
go generate .
git diff --exit-code -- models.gen.go operations.gen.go
go test -race ./...
go vet ./...
```

Only operations marked with the `rest` server are generated. Protocol/browser
operations remain on the main website and are deliberately excluded. Generated
operation coverage is checked against the pinned contract in tests. Schema
constraints and authorization are enforced by the server; generated Go structs
do not replace server validation.

Generation uses `oapi-codegen` **v2.8.0** through a pinned Go `tool` dependency in
`tools/go.mod`. The tool supports the canonical OpenAPI 3.1 document directly;
no downgrade or Python generator is used. Consumers compile the checked-in Go
files and do not need the tool module. Run generation twice to verify determinism.

The models use the standard generator, including nullable fields, enums and union
helpers. A small supported client template retains the public `*Client` methods
and operation directory; the handwritten transport still enforces fixed hosts,
blocked redirects/cookies, bounded decoding and no automatic retries. A separate
OpenAPI Overlay preserves string UUID identifiers, int64 coordinates/timestamps,
optional/null `RawMessage` fields and the optional string Live name. This changes
Go representation only, not the canonical server wire schema. `GetOpenApi` is a
compatibility alias for the generated Go-initialism spelling `GetOpenAPI`.
