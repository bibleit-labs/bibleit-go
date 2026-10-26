package bibleit

import (
	"context"
	"fmt"
)

// Event is a v1 snapshot, changed snapshot, or heartbeat. Cursor is opaque and
// is never an ETag. Data is nil on heartbeats. Sequence can remain unchanged when
// Studio design changes. Callers should reconnect without a cursor to resync.
type Event struct {
	SchemaVersion int                `json:"schema_version"`
	Type          string             `json:"type"`
	ResourceType  string             `json:"resource_type"`
	ResourceID    string             `json:"resource_id"`
	Sequence      int64              `json:"sequence"`
	Cursor        string             `json:"cursor"`
	Data          *LiveEventSnapshot `json:"data"`
}

// PollLiveEvents performs exactly one bounded long poll; it never retries.
// Run writes independently of the outstanding poll and honor rate limits.
func (c *Client) PollLiveEvents(ctx context.Context, id, after string, wait int) (Event, error) {
	waitSeconds := int64(wait)
	params := GetLiveEventsParams{Wait: &waitSeconds}
	if after != "" {
		params.After = &after
	}
	return c.PollLiveEventsWithOptions(ctx, id, &params)
}

// PollLiveEventsWithOptions performs one long poll with a default wait of 25
// seconds and no cursor. Set After to resume, or Wait to explicitly choose a
// wait (including zero). The supplied parameters are never changed.
func (c *Client) PollLiveEventsWithOptions(ctx context.Context, id string, options ...*GetLiveEventsParams) (Event, error) {
	if len(options) > 1 {
		return Event{}, fmt.Errorf("at most one event options value is supported")
	}
	wait := int64(25)
	params := GetLiveEventsParams{}
	if len(options) == 1 && options[0] != nil {
		params = *options[0]
	}
	if params.Wait != nil {
		wait = *params.Wait
	}
	if wait < 0 || wait > 25 {
		return Event{}, fmt.Errorf("event wait must be between 0 and 25 seconds")
	}
	params.Wait = &wait
	if params.After != nil && *params.After == "" {
		params.After = nil
	}
	res, err := c.GetLiveEvents(ctx, id, &Options{Params: params})
	if err != nil {
		return Event{}, err
	}
	defer res.Body.Close()
	var event Event
	if err = res.Decode(&event); err != nil {
		return Event{}, err
	}
	if event.SchemaVersion != 1 {
		return Event{}, fmt.Errorf("unsupported Live event schema version %d", event.SchemaVersion)
	}
	switch event.Type {
	case "live.snapshot", "live.changed", "stream.heartbeat":
	default:
		return Event{}, fmt.Errorf("unsupported Live event type %q", event.Type)
	}
	if event.ResourceType != "live" || event.ResourceID != id || event.Cursor == "" {
		return Event{}, fmt.Errorf("invalid Live event resource or cursor")
	}
	if (event.Type == "stream.heartbeat") != (event.Data == nil) {
		return Event{}, fmt.Errorf("invalid Live event data")
	}
	return event, nil
}
