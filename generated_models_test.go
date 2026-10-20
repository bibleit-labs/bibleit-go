package bibleit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGeneratedModelsPreserveOptionalNullAndSourceTypes(t *testing.T) {
	empty, e := json.Marshal(PatchServiceToken{})
	if e != nil || strings.Contains(string(empty), "expires_at") {
		t.Fatal(string(empty), e)
	}
	explicit, e := json.Marshal(PatchServiceToken{ExpiresAt: json.RawMessage("null")})
	if e != nil || !strings.Contains(string(explicit), `"expires_at":null`) {
		t.Fatal(string(explicit), e)
	}
	live, e := json.Marshal(CreateLive{})
	if e != nil || string(live) != "{}" {
		t.Fatal("unnamed creation contract", string(live), e)
	}
	id := PersonalIdentity{ID: "synthetic-user", PersonalAccountID: "synthetic-account"}
	if _, e = json.Marshal(id); e != nil {
		t.Fatal(e)
	}
	var position int64 = 9007199254740993
	if (LiveStackVerse{Position: position}).Position != position {
		t.Fatal("integer precision lost")
	}
}
func TestGeneratedUnionModelsRoundTrip(t *testing.T) {
	var display ReplaceDisplay
	if e := json.Unmarshal([]byte(`{"type":"empty"}`), &display); e != nil {
		t.Fatal(e)
	}
	variant, e := display.AsReplaceDisplay0()
	if e != nil || variant.Type != "empty" {
		t.Fatal(variant, e)
	}
	raw, e := json.Marshal(display)
	if e != nil || !strings.Contains(string(raw), `"type":"empty"`) {
		t.Fatal(string(raw), e)
	}
	var event LiveEvent
	if e = json.Unmarshal([]byte(`{"schema_version":1,"type":"stream.heartbeat","resource_type":"live","resource_id":"L1","sequence":1,"cursor":"opaque","data":null}`), &event); e != nil {
		t.Fatal(e)
	}
	heartbeat, e := event.AsLiveHeartbeatEvent()
	if e != nil || heartbeat.Cursor != "opaque" {
		t.Fatal(heartbeat, e)
	}
}
