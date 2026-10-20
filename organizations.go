package bibleit

import (
	"context"
	"fmt"
)

// OrganizationDirectoryEntry represents either a membership or a private application.
// State and verification strings remain open to future server values.
type OrganizationDirectoryEntry struct {
	Kind, ID, ApplicationID, Handle, Name, Domain, Role      string
	Verification, Status, State, NextAction, VerificationURL string
	OrganizationID                                           *string
	Fields                                                   map[string]string
}
type OrganizationDirectory struct {
	Entries []OrganizationDirectoryEntry
	Next    *string
	Raw     CommandResponse
}
type OrganizationApplicationResult struct {
	Application OrganizationDirectoryEntry
	Raw         CommandResponse
}

// OrganizationListCommand returns one page; pass the previous Next cursor unchanged.
func OrganizationListCommand(cursor string) (Command, error) {
	if cursor == "" {
		return command("ORGANIZATION LIST")
	}
	return command("ORGANIZATION LIST", cursor)
}
func OrganizationApplicationStatusCommand(id string) (Command, error) {
	value, err := command("ORGANIZATION APPLICATION", id)
	if err != nil {
		return Command{}, err
	}
	return command(value.wire + " STATUS")
}
func organizationEntry(kind string, fields map[string]string) (OrganizationDirectoryEntry, error) {
	keys := []string{"id", "handle", "name", "role", "verification", "status"}
	if kind == "application" {
		keys = []string{"application_id", "name", "domain", "state", "next_action", "verification_url"}
	}
	for _, key := range keys {
		if _, err := required(fields, key); err != nil {
			return OrganizationDirectoryEntry{}, err
		}
	}
	return OrganizationDirectoryEntry{Kind: kind, ID: fields["id"], ApplicationID: fields["application_id"], Handle: fields["handle"], Name: fields["name"], Domain: fields["domain"], Role: fields["role"], Verification: fields["verification"], Status: fields["status"], State: fields["state"], NextAction: fields["next_action"], VerificationURL: fields["verification_url"], OrganizationID: optionalString(fields, "organization_id"), Fields: fields}, nil
}
func (c *SSHClient) ListOrganizations(ctx context.Context, cursor string) (OrganizationDirectory, error) {
	cmd, err := OrganizationListCommand(cursor)
	if err != nil {
		return OrganizationDirectory{}, err
	}
	result, err := c.Execute(ctx, cmd)
	if err != nil {
		return OrganizationDirectory{}, err
	}
	if len(result.Records) == 0 || result.Records[0].Type != "ok" {
		return OrganizationDirectory{}, fmt.Errorf("expected an OK directory response")
	}
	count, err := number(result.Records[0].Fields, "count")
	if err != nil {
		return OrganizationDirectory{}, err
	}
	if count != int64(len(result.Records)-1) {
		return OrganizationDirectory{}, fmt.Errorf("directory count does not match records")
	}
	value := OrganizationDirectory{Entries: []OrganizationDirectoryEntry{}, Next: optionalString(result.Records[0].Fields, "next"), Raw: result}
	for _, record := range result.Records[1:] {
		if record.Type != "organization" && record.Type != "application" {
			return OrganizationDirectory{}, fmt.Errorf("unexpected directory record %s", record.Type)
		}
		entry, err := organizationEntry(record.Type, record.Fields)
		if err != nil {
			return OrganizationDirectory{}, err
		}
		value.Entries = append(value.Entries, entry)
	}
	return value, nil
}
func (c *SSHClient) GetOrganizationApplicationStatus(ctx context.Context, id string) (OrganizationApplicationResult, error) {
	cmd, err := OrganizationApplicationStatusCommand(id)
	if err != nil {
		return OrganizationApplicationResult{}, err
	}
	result, err := c.Execute(ctx, cmd)
	if err != nil {
		return OrganizationApplicationResult{}, err
	}
	fields, err := single(result)
	if err != nil {
		return OrganizationApplicationResult{}, err
	}
	entry, err := organizationEntry("application", fields)
	return OrganizationApplicationResult{Application: entry, Raw: result}, err
}
