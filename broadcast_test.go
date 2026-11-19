package bibleit

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
)

const broadcastTestOrg = "11111111-1111-1111-1111-111111111111"
const broadcastTestID = "22222222-2222-2222-2222-222222222222"

func TestBroadcastCommandContract(t *testing.T) {
	tests := [][]string{
		{"profiles"}, {"channel"}, {"preflight"}, {"events"}, {"publishers"}, {"takeover"},
		{"setup", "list"}, {"setup", "list", "99", broadcastTestID}, {"setup", broadcastTestID, "get"}, {"setup", broadcastTestID, "document", "1", "0"},
		{"write", "intent", "create", "new", "0", "2", strings.Repeat("a", 64)},
		{"write", "intent", "save", broadcastTestID, "1", "2", strings.Repeat("a", 64)},
		{"write", "intent", "duplicate", broadcastTestID, "1", "2", strings.Repeat("a", 64)},
		{"write", "intent", "delete", broadcastTestID, "1", "2", strings.Repeat("a", 64)},
		{"write", "begin", "ticket"}, {"write", "status", "ticket"}, {"write", "inspect", broadcastTestID}, {"write", "abort", broadcastTestID}, {"write", "commit", "ticket", broadcastTestID}, {"write", "chunk", broadcastTestID, "0", "e30="},
		{"event", "create", broadcastTestID, "e30="}, {"event", "attach", broadcastTestID, "provider_ID-1"}, {"event", broadcastTestID, "get"},
		{"event", broadcastTestID, "test", "1", broadcastTestID}, {"event", broadcastTestID, "go-live", "1", broadcastTestID}, {"event", broadcastTestID, "end", "1", broadcastTestID},
		{"operation", broadcastTestID, "get"}, {"publisher", broadcastTestID, "get"}, {"publisher", broadcastTestID, "stop", "1"},
	}
	for _, a := range tests {
		cmd, e := BroadcastCommand(broadcastTestOrg, a...)
		if e != nil {
			t.Errorf("%v: %v", a, e)
			continue
		}
		if cmd.String() != "BROADCAST "+broadcastTestOrg+" "+strings.Join(a, " ") {
			t.Errorf("wire %q", cmd.String())
		}
		if cmd.Streaming() {
			t.Error("broadcast command marked subscription")
		}
	}
}
func TestBroadcastRejectsMalformedAndUnsafeCommands(t *testing.T) {
	bad := [][]string{{}, {"start"}, {"setup", "list", "0"}, {"setup", "list", "100"}, {"setup", "list", "1", "not-uuid"}, {"setup", broadcastTestID, "document", "0", "0"}, {"setup", broadcastTestID, "document", "9223372036854775808", "0"}, {"setup", broadcastTestID, "document", "1", "1048577"},
		{"write", "intent", "create", broadcastTestID, "0", "2", strings.Repeat("a", 64)}, {"write", "intent", "save", broadcastTestID, "0", "2", strings.Repeat("a", 64)}, {"write", "intent", "save", broadcastTestID, "1", "1048577", strings.Repeat("a", 64)},
		{"write", "begin", strings.Repeat("t", 769)}, {"write", "begin", "ticket\nPING"}, {"write", "chunk", broadcastTestID, "0", ""}, {"write", "chunk", broadcastTestID, "0", "e31="}, {"write", "chunk", broadcastTestID, "0", base64.StdEncoding.EncodeToString(make([]byte, 3073))},
		{"event", "attach", broadcastTestID, "https://youtube.com/watch?v=x"}, {"event", "create", broadcastTestID, base64.StdEncoding.EncodeToString([]byte("null"))}, {"event", "create", broadcastTestID, base64.StdEncoding.EncodeToString([]byte("[]"))}, {"publisher", broadcastTestID, "stop", "-1"}, {"event", broadcastTestID, "end", "1", "new"}}
	for _, a := range bad {
		if _, e := BroadcastCommand(broadcastTestOrg, a...); e == nil {
			t.Errorf("accepted %v", a)
		}
	}
	if _, e := BroadcastCommand(strings.ToUpper(broadcastTestID)+"x", "channel"); e == nil {
		t.Fatal("invalid workspace accepted")
	}
}
func TestBroadcastTypedConstructorsAndExactIntent(t *testing.T) {
	payload := []byte("{\n \"name\": \"test\"\n}")
	hash := sha256.Sum256(payload)
	cmd, e := BroadcastWriteIntentCommand(broadcastTestOrg, BroadcastCreate, "", 0, payload)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.HasSuffix(cmd.String(), " "+strconv.Itoa(len(payload))+" "+hex.EncodeToString(hash[:])) {
		t.Fatalf("intent changed original payload: %s", cmd.String())
	}
	chunk, e := BroadcastWriteChunkCommand(broadcastTestOrg, broadcastTestID, 0, bytes.Repeat([]byte{'x'}, 3072))
	if e != nil || len(chunk.String()) > MaxCommandBytes {
		t.Fatalf("maximum chunk: %v", e)
	}
	if _, e = BroadcastWriteChunkCommand(broadcastTestOrg, broadcastTestID, 0, make([]byte, 3073)); e == nil {
		t.Fatal("oversize chunk accepted")
	}
	event, e := CreateBroadcastEventCommand(broadcastTestOrg, broadcastTestID, json.RawMessage(`{"title":"With spaces"}`))
	if e != nil || !strings.Contains(event.String(), base64.StdEncoding.EncodeToString([]byte(`{"title":"With spaces"}`))) {
		t.Fatal("event payload changed", e)
	}
}
func TestSSHBroadcastDoesNotRetryAndHonorsCancellation(t *testing.T) {
	calls := 0
	failure := errors.New("uncertain_outcome")
	client, _ := NewSSHClient(func(_ context.Context, c Command) (CommandResponse, error) {
		calls++
		return CommandResponse{}, failure
	})
	if _, e := client.Broadcast(context.Background(), broadcastTestOrg, "write", "commit", "ticket", broadcastTestID); !errors.Is(e, failure) || calls != 1 {
		t.Fatal("commit was retried or error changed", calls, e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := client.Broadcast(ctx, broadcastTestOrg, "channel"); !errors.Is(e, context.Canceled) || calls != 1 {
		t.Fatal("cancelled command sent", calls, e)
	}
	if _, e := client.Broadcast(context.Background(), broadcastTestOrg, "start"); e == nil || calls != 1 {
		t.Fatal("invalid command sent")
	}
}
func TestBroadcastLifecycleEnvelope(t *testing.T) {
	data := []byte(`{"value":{"id":"x","future":{"safe":true}}}`)
	record := Record{Type: "ok", Fields: map[string]string{"document_format": "broadcast-lifecycle-json-v1", "record_type": "event", "bytes": strconv.Itoa(len(data)), "data_base64": base64.StdEncoding.EncodeToString(data), "future_field": "retained"}}
	got, e := DecodeSSHBroadcastValue(record)
	if e != nil || got.Fields["future_field"] != "retained" || !bytes.Contains(got.Value, []byte("future")) {
		t.Fatal(got, e)
	}
	var decoded map[string]json.RawMessage
	if e = got.Decode(&decoded); e != nil || decoded["future"] == nil {
		t.Fatal(e)
	}
	record.Fields["bytes"] = "1"
	if _, e = DecodeSSHBroadcastValue(record); e == nil {
		t.Fatal("mismatched byte length accepted")
	}
	record.Fields["bytes"] = strconv.Itoa(len(data))
	record.Fields["data_base64"] = "e30="
	if _, e = DecodeSSHBroadcastValue(record); e == nil {
		t.Fatal("missing envelope value accepted")
	}
}
func TestBroadcastDocumentIntegrity(t *testing.T) {
	data := []byte(`{"name":"Demo","version":1}`)
	hash := sha256.Sum256(data)
	record := Record{Type: "ok", Fields: map[string]string{"organization_id": broadcastTestOrg, "id": broadcastTestID, "revision": "7", "document_format": "studio-setup-json-v1", "document_version": "1", "document_bytes": strconv.Itoa(len(data)), "document_sha256": hex.EncodeToString(hash[:]), "chunk_bytes": "3072", "offset": "0", "next_offset": strconv.Itoa(len(data)), "bytes": strconv.Itoa(len(data)), "eof": "true", "data_base64": base64.StdEncoding.EncodeToString(data)}}
	got, e := DecodeSSHBroadcastDocumentChunk(record)
	if e != nil {
		t.Fatal(e)
	}
	if e = VerifyBroadcastDocument(got.Metadata, got.Data); e != nil {
		t.Fatal(e)
	}
	if e = VerifyBroadcastDocument(got.Metadata, append([]byte(nil), data[1:]...)); e == nil {
		t.Fatal("truncation accepted")
	}
	record.Fields["eof"] = "false"
	if _, e = DecodeSSHBroadcastDocumentChunk(record); e == nil {
		t.Fatal("inconsistent EOF accepted")
	}
}
func TestBroadcastReceiptsPreserveUnknownOutcomes(t *testing.T) {
	record := Record{Type: "ok", Fields: map[string]string{"request_id": "opaque-ticket", "result_state": "future_reconciliation", "terminal": "false", "future": "retained"}}
	got, e := DecodeSSHBroadcastWriteReceipt(record)
	if e != nil || got.Terminal || got.ResultState != "future_reconciliation" || got.Fields["future"] != "retained" {
		t.Fatal(got, e)
	}
	record.Fields["applied"] = "true"
	if _, e = DecodeSSHBroadcastWriteReceipt(record); e == nil {
		t.Fatal("nonterminal applied outcome accepted")
	}
	record.Fields["terminal"] = "true"
	record.Fields["revision"] = "9223372036854775808"
	if _, e = DecodeSSHBroadcastWriteReceipt(record); e == nil {
		t.Fatal("overflow revision accepted")
	}
}

func TestBroadcastTicketsUseCanonicalURLEncoding(t *testing.T) {
	for _, ticket := range []string{"ticket;PING", "$(id)", "`id`", "ticket=", "ticket+", "ticket/", "ticket\\name", "ticket\x01"} {
		if _, err := BroadcastWriteStatusCommand(broadcastTestOrg, ticket); err == nil {
			t.Errorf("accepted malformed ticket %q", ticket)
		}
		if _, err := BroadcastWriteOptions(ticket, 0, []byte("{}")); err == nil {
			t.Errorf("accepted malformed REST ticket %q", ticket)
		}
	}
	if _, err := BroadcastWriteStatusCommand(broadcastTestOrg, "Az09_signed-ticket"); err != nil {
		t.Fatal(err)
	}
}
