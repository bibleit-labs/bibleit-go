package bibleit

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestOAuthProtocolSeparationAndNoRetries(t *testing.T) {
	calls := 0
	c, _ := NewOAuthClient(&http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.String() != OAuthEndpoint+"/oauth/token" || r.Header.Get("Authorization") != "" || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Fatal(r)
		}
		b, _ := io.ReadAll(r.Body)
		f, _ := url.ParseQuery(string(b))
		if f.Get("refresh_token") != "refresh-secret" || f.Get("grant_type") != "refresh_token" || f.Get("client_id") != "registered-client" {
			t.Fatal(f)
		}
		return response(400, `{"error":"invalid_grant"}`), nil
	})})
	_, err := c.Refresh(context.Background(), "registered-client", "refresh-secret", nil)
	var p *OAuthProtocolError
	if !errors.As(err, &p) || p.Code != "invalid_grant" || calls != 1 {
		t.Fatal(err, calls)
	}
	u, err := AuthorizationURL("registered-client", "https://client.example/callback", "state", strings.Repeat("a", 43), []string{"identity.read"})
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(u)
	if parsed.Host != "bibleit.app" || parsed.Query().Get("code_challenge_method") != "S256" || parsed.Query().Get("state") != "state" {
		t.Fatal(u)
	}
}
func TestLiveEventSchemaVersion(t *testing.T) {
	for _, version := range []string{"1", "2"} {
		c, _ := NewClient("bt_test", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
			if r.URL.Query().Get("wait") != "25" {
				t.Fatal(r.URL)
			}
			return response(200, `{"schema_version":`+version+`,"type":"stream.heartbeat","resource_type":"live","resource_id":"live-1","sequence":1,"cursor":"v1:test","data":null}`), nil
		})})
		e, err := c.PollLiveEvents(context.Background(), "live-1", "", 25)
		if version == "1" && (err != nil || e.Type != "stream.heartbeat") {
			t.Fatal(e, err)
		}
		if version == "2" && err == nil {
			t.Fatal("accepted unknown event schema")
		}
	}
}
