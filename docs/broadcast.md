# Broadcast Studio SSH controls

Broadcast controls use an explicit workspace UUID and the operator's registered
personal SSH identity. The server checks current workspace permissions and
approval for every operation. SDK helpers do not grant admission, create a
browser capture, authorize Google access, or restart a publisher.

Negotiate `BroadcastReadCapability`, `BroadcastWriteCapability` and
`BroadcastLifecycleCapability` from `ServerInfoCommand` before using the matching
surface. Writes are unavailable when the server has no configured signing key
and recovery epoch. Older servers can continue using existing SDK commands.

| Surface | SSH constructors |
| --- | --- |
| Setup list and metadata | `ListBroadcastSetupsCommand`, `GetBroadcastSetupCommand` |
| Revision-bound document export | `BroadcastSetupDocumentCommand` |
| Immutable mutation intent | `BroadcastWriteIntentCommand` |
| Upload lifecycle | `BroadcastWriteBeginCommand`, `BroadcastWriteInspectCommand`, `BroadcastWriteChunkCommand`, `BroadcastWriteCommitCommand`, `BroadcastWriteAbortCommand` |
| Durable mutation outcome | `BroadcastWriteStatusCommand` |
| Profiles, channel and readiness | `BroadcastProfilesCommand`, `BroadcastChannelCommand`, `BroadcastPreflightCommand` |
| Event list, lookup and provider operation | `BroadcastEventsCommand`, `GetBroadcastEventCommand`, `GetBroadcastOperationCommand` |
| Create, attach and transitions | `CreateBroadcastEventCommand`, `AttachBroadcastEventCommand`, `TransitionBroadcastEventCommand` |
| Publisher lookup and stop | `BroadcastPublishersCommand`, `GetBroadcastPublisherCommand`, `StopBroadcastPublisherCommand` |
| Browser takeover requirement | `BroadcastTakeoverCommand` |

`BroadcastCommand(workspace, args...)` provides the same validated grammar for
transport adapters. `SSHClient.Broadcast` calls the caller-supplied executor
exactly once. Invalid arguments and cancelled contexts fail before execution.
The SDK never opens OpenSSH, stores tickets, or retries automatically.

## Setup writes

Prepare the exact UTF-8 JSON bytes first. `BroadcastWriteIntentCommand` computes
the byte count and SHA-256 without changing whitespace. Create uses an empty
target and revision zero; save, duplicate and delete require a target UUID and
positive revision. Duplicate uploads exactly `{"name":"Copy name"}`; delete
uploads exactly `{}`. The shared server validates the Setup document.

Retain the returned signed request ticket and original bytes securely before
uploading. The ticket is bounded to 768 bytes and is not a bearer permission.
Begin, inspect, chunk, commit and abort must execute over the **same persistent
SSH shell without a PTY**. Intent and status may use independent exec channels.
Applications own this transport lifetime and host-key verification.

A document is at most 1 MiB and a chunk at most `BroadcastChunkBytes` (3,072 raw
bytes). Offsets are contiguous. Only an identical retry of the most recently
accepted chunk is allowed. A lost commit reply is an uncertain outcome: inspect
`BroadcastWriteStatusCommand` using the original ticket. Never invent a fresh
intent automatically, including when status says `not_seen`. An unknown outcome,
expired receipt or changed recovery epoch requires explicit reconciliation.

`DecodeSSHBroadcastWriteReceipt` retains unknown states and fields. Clients must
not interpret an unfamiliar state as success. A terminal acknowledgement of a
queued YouTube operation also does not prove provider completion: poll that
operation's original request UUID.

## Decoding and export integrity

Use `DecodeSSHBroadcastSetupMetadata` and
`DecodeSSHBroadcastDocumentChunk` to reject invalid offsets, mismatched lengths,
unsupported document formats and malformed base64. Keep the same Setup ID,
revision and digest across all pages, assemble them in order, then call
`VerifyBroadcastDocument` to check the full length, SHA-256 and JSON object.
These helpers do not silently fetch newer revisions or retry a failed export.

`DecodeSSHBroadcastValue` validates the bounded lifecycle JSON envelope and
retains its `value` as `json.RawMessage`. `SSHBroadcastValue.Decode` can decode a
named model while the raw value retains unknown future fields. A disconnected
channel is `null`. Browser takeover returns `browser_required`; publishing still
requires Studio's browser-owned feed and readiness checks.

Synthetic command/decoder tests establish framing and library behavior only;
they do not establish YouTube media delivery or production readiness.

## REST parity

The resource `Client` exposes all 21 broadcast control-plane operations under
`/organizations/{organization_id}/broadcast`. Use personal `bt_` credentials
with the appropriate `broadcast.get`, `broadcast.create`, `broadcast.manage` or
`broadcast.publish` ceiling, in addition to current workspace authority.
Organization service and delegated OAuth credentials are not supported by this
broadcast contract. A token scope never overrides workspace membership, approval
or the configured publisher admission and capacity limits.

Setup writes use the same signed intent and durable receipt semantics as SSH.
Prepare a `BroadcastWriteIntentInput` with `NewBroadcastWriteIntentInput`, send
it to `CreateBroadcastWriteRequest` using `Options.JSON`, and decode the returned
`BroadcastWriteIntent`. Retain its `RequestID` and the original payload before
sending a mutation. Use `BroadcastWriteOptions` for the mutation body:

```go
input, err := bibleit.NewBroadcastWriteIntentInput(
    bibleit.BroadcastSave, setupID, expectedRevision, payload)
if err != nil { return err }
intentResponse, err := client.CreateBroadcastWriteRequest(ctx, workspaceID,
    &bibleit.Options{JSON: input})
if err != nil { return err }
var intent bibleit.BroadcastWriteIntent
err = intentResponse.Decode(&intent)
intentResponse.Body.Close()
if err != nil { return err }
// Persist intent.RequestID and payload privately before the mutation.
options, err := bibleit.BroadcastWriteOptions(intent.RequestID, expectedRevision, payload)
if err != nil { return err }
result, err := client.ReplaceBroadcastSetup(ctx, workspaceID, setupID, options)
if err != nil { return err }
defer result.Body.Close()
var receipt bibleit.BroadcastWriteReceipt
if err := result.Decode(&receipt); err != nil { return err }
```

`Options.JSON`, including a `json.RawMessage`, can change whitespace during
marshaling. Signed mutation bodies must use the exact raw bytes: the helper
creates an independent `Options.Body` reader and sets the length and media type.
For an explicit retry, create fresh options with the original ticket and bytes.
On an ambiguous result, call `GetBroadcastWriteRequest` with that same ticket;
the SDK never replaces it automatically.

REST event create/attach/transition use a retained client-generated request UUID
in `Options.IdempotencyKey`. Transitions require `Options.IfMatch` formatted as
`"event-<revision>"`; publisher stop requires `"publisher-<revision>"`. The
metadata, lifecycle outcome and browser requirements match the SSH surface.
`BroadcastChannelResult` is a pointer alias so a disconnected channel remains
nil when decoding JSON `null`, rather than becoming an apparently valid empty
channel. `RequestBroadcastTakeover` returns the server's `browser_required`
problem after authorization; it does not start browser capture.
