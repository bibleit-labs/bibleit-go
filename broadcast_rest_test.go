package bibleit

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestBroadcastRESTSignedBytesAndRevisions(t *testing.T) {
	payload := []byte("{\n  \"name\": \"exact bytes\"\n}\n")
	original := bytes.Clone(payload)
	intent, e := NewBroadcastWriteIntentInput(BroadcastSave, broadcastTestID, 7, payload)
	if e != nil {
		t.Fatal(e)
	}
	h := sha256.Sum256(original)
	if intent.PayloadSha256 != hex.EncodeToString(h[:]) || intent.PayloadBytes != int64(len(original)) || intent.TargetID == nil || *intent.TargetID != broadcastTestID {
		t.Fatalf("invalid intent: %+v", intent)
	}
	options, e := BroadcastWriteOptions("signed-ticket", 7, payload)
	if e != nil {
		t.Fatal(e)
	}
	payload[0] = 'x' // The pending HTTP request must own its immutable snapshot.
	calls := 0
	client, _ := NewClient("bt_test", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		calls++
		got, e := io.ReadAll(r.Body)
		if e != nil || !bytes.Equal(got, original) {
			t.Fatalf("signed bytes changed: %q %v", got, e)
		}
		if r.Method != "PUT" || r.URL.String() != Endpoint+"/organizations/"+broadcastTestOrg+"/broadcast/setups/"+broadcastTestID || r.Header.Get("If-Match") != `"setup-7"` || r.Header.Get("Idempotency-Key") != "signed-ticket" || r.Header.Get("Content-Type") != "application/json" || r.ContentLength != int64(len(original)) {
			t.Fatal(r.Method, r.URL, r.Header, r.ContentLength)
		}
		return response(409, `{"type":"about:blank","title":"Uncertain","status":409,"code":"uncertain_outcome","request_id":"r1"}`), nil
	})})
	_, e = client.ReplaceBroadcastSetup(context.Background(), broadcastTestOrg, broadcastTestID, options)
	var problem *ProblemError
	if !errors.As(e, &problem) || problem.Code != "uncertain_outcome" || calls != 1 {
		t.Fatal("ambiguous write retried or lost outcome", calls, e)
	}
}
func TestBroadcastRESTNullableChannel(t *testing.T) {
	client, _ := NewClient("bt_test", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) { return response(200, "null"), nil })})
	response, e := client.GetBroadcastChannel(context.Background(), broadcastTestOrg)
	if e != nil {
		t.Fatal(e)
	}
	defer response.Body.Close()
	var channel BroadcastChannelResult
	if e = response.Decode(&channel); e != nil || channel != nil {
		t.Fatalf("null connection lost: %+v %v", channel, e)
	}
}
func TestBroadcastRESTIntentAndBodyValidation(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("[]"), []byte("null"), []byte("{invalid}"), bytes.Repeat([]byte{'x'}, MaxBroadcastDocumentBytes+1)} {
		if _, e := NewBroadcastWriteIntentInput(BroadcastCreate, "", 0, data); e == nil {
			t.Fatalf("accepted invalid intent payload %q", data[:min(10, len(data))])
		}
		if _, e := BroadcastWriteOptions("ticket", 0, data); e == nil {
			t.Fatal("accepted invalid write body")
		}
	}
	if _, e := NewBroadcastWriteIntentInput(BroadcastSave, broadcastTestID, 0, []byte("{}")); e == nil {
		t.Fatal("accepted stale zero revision")
	}
	options, e := BroadcastWriteOptions("ticket", 0, []byte("{}"))
	if e != nil || options.IfMatch != "" || options.JSON != nil {
		t.Fatal(options, e)
	}
	var empty BroadcastEmptyInput
	body, e := json.Marshal(empty)
	if e != nil || string(body) != "{}" {
		t.Fatal(string(body), e)
	}
}
