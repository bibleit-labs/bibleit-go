package bibleit

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestOAuthGrantResponseScopesRemainSeparateFromRequestScopes(t *testing.T) {
	const wire = `{"data":[{"id":"00000000000000000000000000000000","client_id":"synthetic-response-client","client_name":"Synthetic response only","scopes":["identity.read","live.queue.stage","live.rename"],"expires_at":1}]}`
	calls := 0
	client, err := NewClient("svc_synthetic", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodGet || r.URL.String() != Endpoint+"/me/oauth-grants" {
			t.Fatal("unexpected grant-list endpoint")
		}
		return response(http.StatusOK, wire), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ListOAuthGrants(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer result.Body.Close()
	var grants OAuthGrantList
	if err := result.Decode(&grants); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(grants.Data) != 1 || len(grants.Data[0].Scopes) != 3 {
		t.Fatal("grant response lost scopes or retried")
	}
	for _, scope := range grants.Data[0].Scopes {
		if !scope.Valid() {
			t.Fatalf("known granted scope rejected: %s", scope)
		}
	}
	for _, scope := range []string{"live.queue.stage", "live.rename"} {
		if OAuthScope(scope).Valid() || !McpOAuthScope(scope).Valid() {
			t.Fatal("MCP scope crossed the ordinary request boundary")
		}
	}
	if OAuthGrantedScope("future.unknown").Valid() || OAuthScope("future.unknown").Valid() {
		t.Fatal("closed validation accepted an unknown scope")
	}
	encoded, err := json.Marshal(grants)
	var expected, actual any
	if err != nil || json.Unmarshal([]byte(wire), &expected) != nil || json.Unmarshal(encoded, &actual) != nil || !reflect.DeepEqual(expected, actual) {
		t.Fatal("response-only scope type changed the grant JSON representation", err)
	}
}
