package bibleit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type trip func(*http.Request) (*http.Response, error)

func (f trip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
}
func TestContractCoverage(t *testing.T) {
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(contract, &spec); err != nil {
		t.Fatal(err)
	}
	n := 0
	for path, methods := range spec.Paths {
		for method, raw := range methods {
			var op struct {
				ID     string `json:"operationId"`
				Server string `json:"x-fern-server-name"`
			}
			_ = json.Unmarshal(raw, &op)
			if op.Server != "rest" {
				continue
			}
			n++
			got, ok := operations[op.ID]
			if !ok || got.Path != path || got.Method != strings.ToUpper(method) {
				t.Fatalf("contract mismatch: %s", op.ID)
			}
		}
	}
	if n != len(operations) {
		t.Fatal(n, len(operations))
	}
}
func TestEveryOperationPinned(t *testing.T) {
	for name, op := range Operations() {
		t.Run(name, func(t *testing.T) {
			params := map[string]string{}
			path := op.Path
			for strings.Contains(path, "{") {
				i := strings.Index(path, "{")
				j := strings.Index(path, "}")
				key := path[i+1 : j]
				params[key] = "resource"
				path = strings.ReplaceAll(path, "{"+key+"}", "resource")
			}
			client, _ := NewClient("svc_test", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
				if r.URL.String() != Endpoint+path || r.Method != op.Method || r.Header.Get("Authorization") != "Bearer svc_test" {
					t.Fatal(r)
				}
				return response(204, ""), nil
			})})
			res, err := client.Call(context.Background(), name, params, nil)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()
		})
	}
}
func TestPreconditionsAndNoRetries(t *testing.T) {
	calls := 0
	client, _ := NewClient("svc_test", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("If-Match") != `"live:7"` || r.Header.Get("Idempotency-Key") != "create-1" || r.Header.Get("X-Destination-Service-Token") != "svc_dest" {
			t.Fatal(r.Header)
		}
		res := response(412, `{"type":"about:blank","title":"Stale","code":"precondition_failed","status":412,"request_id":"abc"}`)
		res.Header.Set("Retry-After", "2")
		return res, nil
	})})
	_, err := client.Call(context.Background(), "createLiveQueueItem", map[string]string{"live_id": "test"}, &Options{JSON: map[string]any{}, IfMatch: `"live:7"`, IdempotencyKey: "create-1", DestinationToken: "svc_dest"})
	var p *ProblemError
	if !errors.As(err, &p) || p.Status != 412 || p.Code != "precondition_failed" || p.RequestID != "abc" || p.RetryAfter() != 2*time.Second || calls != 1 {
		t.Fatal(err, calls)
	}
}
func TestNoRedirectOrPathEscape(t *testing.T) {
	c, _ := NewClient("bt_test", nil)
	if c.http.Timeout < 30*time.Second {
		t.Fatal("long polling needs timeout margin")
	}
	if c.http.CheckRedirect(nil, nil) != http.ErrUseLastResponse {
		t.Fatal("redirect enabled")
	}
	c.http.Transport = trip(func(*http.Request) (*http.Response, error) {
		t.Fatal("unexpected network call")
		return nil, fmt.Errorf("unexpected")
	})
	for _, params := range []map[string]string{nil, {"live_id": ".."}, {"live_id": "../me"}, {"live_id": "https://evil.example"}} {
		if _, err := c.GetLive(context.Background(), params["live_id"], nil); err == nil {
			t.Fatal(params)
		}
		if _, err := c.Call(context.Background(), "getLive", params, nil); err == nil {
			t.Fatal(params)
		}
	}
}
func TestBinaryAndJSON(t *testing.T) {
	c, _ := NewClient("svc_test", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		if string(b) != "raw bytes" || r.ContentLength != 9 || r.Header.Get("Content-Type") != "image/png" || r.URL.Query().Get("kind") != "image" {
			t.Fatal(r, string(b))
		}
		return response(200, "download"), nil
	})})
	r, err := c.UploadPresentationAsset(context.Background(), "deck", &Options{Body: strings.NewReader("raw bytes"), ContentLength: 9, ContentType: "image/png", Query: map[string][]string{"kind": {"image"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	if string(b) != "download" {
		t.Fatal(string(b))
	}
}

func TestNewClientOptionalTransport(t *testing.T) {
	c, err := NewClient("")
	if err != nil || c.http.Timeout != 40*time.Second {
		t.Fatal("default client", err)
	}
	if _, err := NewClient("", nil, nil); err == nil {
		t.Fatal("accepted multiple transports")
	}
}
