package bibleit

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOrganizationTypedFixtureResults(t *testing.T) {
	body, err := os.ReadFile("contract/fixtures/responses-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Wire string
		Invalid    bool
		ErrorCode  string `json:"error_code"`
	}
	if err := json.Unmarshal(body, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		if !strings.HasPrefix(tc.Name, "organization-") || tc.Invalid || tc.ErrorCode != "" {
			continue
		}
		t.Run(tc.Name, func(t *testing.T) {
			client := discoveryClient(t, map[string]string{"ORGANIZATION LIST": tc.Wire, "ORGANIZATION APPLICATION application-fixture STATUS": tc.Wire})
			if strings.Contains(tc.Name, "directory") {
				value, err := client.ListOrganizations(context.Background(), "")
				if err != nil {
					t.Fatal(err)
				}
				if tc.Name == "organization-directory-pagination-decoding" && value.Next == nil {
					t.Fatal("pagination cursor missing")
				}
				for _, entry := range value.Entries {
					if entry.Kind == "application" && entry.State != "approved" && entry.OrganizationID != nil {
						t.Fatal("application acquired organization ID")
					}
				}
			} else {
				value, err := client.GetOrganizationApplicationStatus(context.Background(), "application-fixture")
				if err != nil {
					t.Fatal(err)
				}
				if tc.Name == "organization-application-deleted-status" && (value.Application.State != "organization_deleted" || value.Application.NextAction != "none" || value.Application.OrganizationID != nil) {
					t.Fatalf("lost deletion state: %+v", value)
				}
			}
		})
	}
}

func TestOrganizationBuildersAndUnknownStates(t *testing.T) {
	for _, bad := range []string{"", "id\nPING", "id\rPING", "id\x00", "bad\"id"} {
		if _, err := OrganizationApplicationStatusCommand(bad); err == nil {
			t.Fatalf("accepted unsafe id %q", bad)
		}
		if bad != "" {
			if _, err := OrganizationListCommand(bad); err == nil {
				t.Fatalf("accepted unsafe cursor %q", bad)
			}
		}
	}
	cmd, err := OrganizationListCommand("a:fixture")
	if err != nil || cmd.String() != "ORGANIZATION LIST a:fixture" {
		t.Fatalf("%s %v", cmd.String(), err)
	}
	cmd, err = OrganizationApplicationStatusCommand("fixture")
	if err != nil || cmd.String() != "ORGANIZATION APPLICATION fixture STATUS" {
		t.Fatalf("%s %v", cmd.String(), err)
	}
	if cmd, err := HelpCommand("organization"); err != nil || cmd.String() != "HELP ORGANIZATION" {
		t.Fatalf("%s %v", cmd.String(), err)
	}
	client := discoveryClient(t, map[string]string{"ORGANIZATION APPLICATION id STATUS": "OK application_id=id name=Example domain=example.org state=future_state next_action=future_action verification_url=/organizations/applications/id future=kept\n"})
	value, err := client.GetOrganizationApplicationStatus(context.Background(), "id")
	if err != nil || value.Application.State != "future_state" || value.Application.Fields["future"] != "kept" {
		t.Fatalf("%+v %v", value, err)
	}
}

func TestOrganizationTypedCountMismatch(t *testing.T) {
	client := discoveryClient(t, map[string]string{"ORGANIZATION LIST": "OK count=1\nEND\n"})
	if _, err := client.ListOrganizations(context.Background(), ""); err == nil {
		t.Fatal("accepted wrong directory count")
	}
}
