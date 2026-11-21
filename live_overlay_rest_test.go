package bibleit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestLiveOverlayRESTNullAndStaleWrites(t *testing.T) {
	calls := 0
	client, e := NewClient("svc_fixture", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/v1/lives/L1/overlay" {
			t.Fatal(r.URL.Path)
		}
		if r.Method == "GET" {
			v := response(200, `{"live_id":"L1","sequence":7,"overlay":null}`)
			v.Header.Set("ETag", `"live-7"`)
			return v, nil
		}
		if r.Method != "PUT" || r.Header.Get("If-Match") != `"live-7"` {
			t.Fatal(r.Method, r.Header)
		}
		b, _ := io.ReadAll(r.Body)
		var data map[string]any
		if json.Unmarshal(b, &data) != nil || len(data) != 1 || data["design_id"] != nil {
			t.Fatal(string(b))
		}
		return response(412, `{"error":"precondition_failed"}`), nil
	})})
	if e != nil {
		t.Fatal(e)
	}
	r, e := client.GetLiveOverlay(context.Background(), "L1")
	if e != nil {
		t.Fatal(e)
	}
	var state LiveOverlayState
	if e = r.Decode(&state); e != nil {
		t.Fatal(e)
	}
	r.Body.Close()
	if state.LiveID != "L1" || state.Sequence != 7 || state.Overlay != nil {
		t.Fatalf("%+v", state)
	}
	_, e = client.SetLiveOverlay(context.Background(), "L1", &Options{IfMatch: `"live-7"`, JSON: LiveOverlayInput{DesignID: nil}})
	if e == nil {
		t.Fatal("stale write succeeded")
	}
	if calls != 2 {
		t.Fatal("stale mutation was retried", calls)
	}
}
