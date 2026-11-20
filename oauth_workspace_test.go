package bibleit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestWorkspaceProbeStrictDecodePreservesOptionalProfile(t *testing.T) {
	const personal = `{"id":"00000000-0000-4000-8000-000000000001","display_name":"Synthetic","handle":null,"personal_account_id":"00000000-0000-4000-8000-000000000002","client_id":"synthetic-client","scopes":["identity.read"],"expires_at":2000000000,"resource":"https://mcp.example.test/mcp","permissions":["mcp.get","mcp.list","mcp.call"]}`
	profile := `,"workspace_permissions":[{"organization_id":"00000000-0000-4000-8000-000000000003","name":"Synthetic workspace","scopes":["broadcast.get"]}]}`
	expanded := strings.TrimSuffix(personal, "}") + profile
	for _, tc := range []struct {
		name, wire string
		count      int
	}{{"personal", personal, 0}, {"workspace", expanded, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			var probe McpOAuthProbe
			decoder := json.NewDecoder(strings.NewReader(tc.wire))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&probe); err != nil {
				t.Fatal(err)
			}
			if len(probe.WorkspacePermissions) != tc.count {
				t.Fatal("profile presence changed")
			}
			if tc.count == 1 {
				p := probe.WorkspacePermissions[0]
				if p.OrganizationID != "00000000-0000-4000-8000-000000000003" || p.Name != "Synthetic workspace" || len(p.Scopes) != 1 || p.Scopes[0] != McpWorkspacePermissionScopesBroadcastGet || !p.Scopes[0].Valid() {
					t.Fatal("typed profile lost")
				}
			}
			out, err := json.Marshal(probe)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(out, []byte("workspace_permissions")) != (tc.count == 1) {
				t.Fatal("optional profile serialization changed")
			}
		})
	}
	var probe McpOAuthProbe
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSuffix(personal, "}") + `,"actor_override":"not-authority"}`))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&probe); err == nil {
		t.Fatal("strict consumer accepted unknown field")
	}
}

func TestWorkspaceResponseScopeValidationStaysSeparateFromRequests(t *testing.T) {
	for _, tc := range []struct {
		name, wire string
		target     any
	}{
		{"metadata", `{"scopes_supported":["identity.read","broadcast.get"]}`, &McpOAuthMetadata{}},
		{"grants", `{"data":[{"id":"00000000000000000000000000000000","client_id":"synthetic-client","client_name":"Synthetic","scopes":["identity.read","broadcast.get"],"expires_at":2000000000}]}`, &OAuthGrantList{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := Response{Response: &http.Response{Body: io.NopCloser(strings.NewReader(tc.wire))}}
			defer r.Body.Close()
			if err := r.Decode(tc.target); err != nil {
				t.Fatal(err)
			}
			switch value := tc.target.(type) {
			case *McpOAuthMetadata:
				if len(value.ScopesSupported) != 2 || value.ScopesSupported[1] != McpOAuthScopeBroadcastGet || !value.ScopesSupported[1].Valid() {
					t.Fatal("metadata scope not preserved or accepted")
				}
			case *OAuthGrantList:
				if len(value.Data) != 1 || len(value.Data[0].Scopes) != 2 || value.Data[0].Scopes[1] != OAuthGrantedScopeBroadcastGet || !value.Data[0].Scopes[1].Valid() {
					t.Fatal("owner grant scope not preserved or accepted")
				}
			}
		})
	}
	if !McpOAuthScopeBroadcastGet.Valid() || !OAuthGrantedScopeBroadcastGet.Valid() {
		t.Fatal("known response scope rejected")
	}
	if OAuthScope("broadcast.get").Valid() {
		t.Fatal("workspace response scope widened ordinary OAuth")
	}
	if McpOAuthScope("broadcast.publish").Valid() || OAuthGrantedScope("future.unknown").Valid() || McpWorkspacePermissionScopes("broadcast.publish").Valid() {
		t.Fatal("strict vocabulary widened")
	}
	for name, operation := range operations {
		if strings.HasPrefix(operation.Path, "/oauth/mcp/") {
			t.Fatal("ordinary REST client exposed MCP issuer bridge", name)
		}
	}
}
