package bibleit

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
)

const (
	BroadcastReadCapability      = "broadcast.read.v1"
	BroadcastWriteCapability     = "broadcast.write.v1"
	BroadcastLifecycleCapability = "broadcast.lifecycle.v1"
	MaxBroadcastDocumentBytes    = 1048576
	BroadcastChunkBytes          = 3072
)

var broadcastUUID = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
var broadcastProviderID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
var broadcastDigest = regexp.MustCompile(`^[a-f0-9]{64}$`)
var broadcastTicketEncoding = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// BroadcastWriteOperation binds the immutable intent to a Setup mutation.
type BroadcastWriteOperation string

const (
	BroadcastCreate    BroadcastWriteOperation = "create"
	BroadcastSave      BroadcastWriteOperation = "save"
	BroadcastDuplicate BroadcastWriteOperation = "duplicate"
	BroadcastDelete    BroadcastWriteOperation = "delete"
)

// BroadcastTransition is a revision-aware YouTube lifecycle action.
type BroadcastTransition string

const (
	BroadcastTesting  BroadcastTransition = "test"
	BroadcastGoLive   BroadcastTransition = "go-live"
	BroadcastComplete BroadcastTransition = "end"
)

// BroadcastCommand validates the complete public SSH broadcast grammar. It does
// not negotiate capabilities or start capture. Write upload commands require a
// caller-owned persistent no-PTY shell; never retry with a new request ticket.
func BroadcastCommand(organizationID string, args ...string) (Command, error) {
	if !broadcastUUID.MatchString(organizationID) || !validBroadcastArgs(args) {
		return Command{}, errors.New("invalid broadcast command")
	}
	c, err := command("BROADCAST", append([]string{organizationID}, args...)...)
	if err == nil && len(c.wire) > MaxCommandBytes {
		return Command{}, &RequestSizeError{Part: "command", Bytes: len(c.wire), Limit: MaxCommandBytes}
	}
	return c, err
}
func broadcastNumber(s string, min, max int64) bool {
	if len(s) == 0 || len(s) > 19 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	n, e := strconv.ParseInt(s, 10, 64)
	return e == nil && n >= min && n <= max
}
func broadcastTicket(s string) bool {
	return len(s) > 0 && len(s) <= 768 && broadcastTicketEncoding.MatchString(s)
}
func validBroadcastArgs(a []string) bool {
	if len(a) == 0 {
		return false
	}
	switch a[0] {
	case "profiles", "channel", "preflight", "events", "publishers", "takeover":
		return len(a) == 1
	case "setup":
		if len(a) >= 2 && a[1] == "list" {
			return len(a) == 2 || (len(a) == 3 && broadcastNumber(a[2], 1, 99)) || (len(a) == 4 && broadcastNumber(a[2], 1, 99) && broadcastUUID.MatchString(a[3]))
		}
		return len(a) >= 3 && broadcastUUID.MatchString(a[1]) && ((len(a) == 3 && a[2] == "get") || (len(a) == 5 && a[2] == "document" && broadcastNumber(a[3], 1, 1<<63-1) && broadcastNumber(a[4], 0, MaxBroadcastDocumentBytes)))
	case "write":
		if len(a) < 3 {
			return false
		}
		switch a[1] {
		case "intent":
			if len(a) != 7 || !broadcastNumber(a[4], 0, 1<<63-1) || !broadcastNumber(a[5], 2, MaxBroadcastDocumentBytes) || !broadcastDigest.MatchString(a[6]) {
				return false
			}
			switch a[2] {
			case "create":
				return a[3] == "new" && a[4] == "0"
			case "save", "duplicate", "delete":
				return broadcastUUID.MatchString(a[3]) && broadcastNumber(a[4], 1, 1<<63-1)
			}
		case "begin", "status":
			return len(a) == 3 && broadcastTicket(a[2])
		case "inspect", "abort":
			return len(a) == 3 && broadcastUUID.MatchString(a[2])
		case "commit":
			return len(a) == 4 && broadcastTicket(a[2]) && broadcastUUID.MatchString(a[3])
		case "chunk":
			if len(a) != 5 || !broadcastUUID.MatchString(a[2]) || !broadcastNumber(a[3], 0, MaxBroadcastDocumentBytes) || len(a[4]) > 4096 {
				return false
			}
			b, e := base64.StdEncoding.Strict().DecodeString(a[4])
			return e == nil && len(b) > 0 && len(b) <= BroadcastChunkBytes && base64.StdEncoding.EncodeToString(b) == a[4]
		}
	case "event":
		if len(a) == 4 && a[1] == "create" && broadcastUUID.MatchString(a[2]) && len(a[3]) <= 4096 {
			b, e := base64.StdEncoding.DecodeString(a[3])
			var m map[string]json.RawMessage
			return e == nil && len(b) <= BroadcastChunkBytes && json.Unmarshal(b, &m) == nil && m != nil
		}
		if len(a) == 4 && a[1] == "attach" {
			return broadcastUUID.MatchString(a[2]) && broadcastProviderID.MatchString(a[3])
		}
		if len(a) < 3 || !broadcastUUID.MatchString(a[1]) {
			return false
		}
		if len(a) == 3 {
			return a[2] == "get"
		}
		return len(a) == 5 && (a[2] == "test" || a[2] == "go-live" || a[2] == "end") && broadcastNumber(a[3], 1, 1<<63-1) && broadcastUUID.MatchString(a[4])
	case "operation":
		return len(a) == 3 && broadcastUUID.MatchString(a[1]) && a[2] == "get"
	case "publisher":
		return len(a) >= 3 && broadcastUUID.MatchString(a[1]) && ((len(a) == 3 && a[2] == "get") || (len(a) == 4 && a[2] == "stop" && broadcastNumber(a[3], 1, 1<<63-1)))
	}
	return false
}

// Broadcast executes exactly one command without retries. The executor owns
// capability negotiation and a persistent shell for bounded upload commands.
func (c *SSHClient) Broadcast(ctx context.Context, organizationID string, args ...string) (CommandResponse, error) {
	cmd, e := BroadcastCommand(organizationID, args...)
	if e != nil {
		return CommandResponse{}, e
	}
	return c.Execute(ctx, cmd)
}
func ListBroadcastSetupsCommand(org string, limit int, after string) (Command, error) {
	a := []string{"setup", "list"}
	if limit != 0 || after != "" {
		if limit == 0 {
			limit = 25
		}
		a = append(a, strconv.Itoa(limit))
	}
	if after != "" {
		a = append(a, after)
	}
	return BroadcastCommand(org, a...)
}
func GetBroadcastSetupCommand(org, id string) (Command, error) {
	return BroadcastCommand(org, "setup", id, "get")
}
func BroadcastSetupDocumentCommand(org, id string, revision int64, offset int) (Command, error) {
	return BroadcastCommand(org, "setup", id, "document", strconv.FormatInt(revision, 10), strconv.Itoa(offset))
}
func BroadcastWriteIntentCommand(org string, op BroadcastWriteOperation, target string, revision int64, payload []byte) (Command, error) {
	if op == BroadcastCreate && target == "" {
		target = "new"
	}
	h := sha256.Sum256(payload)
	return BroadcastCommand(org, "write", "intent", string(op), target, strconv.FormatInt(revision, 10), strconv.Itoa(len(payload)), hex.EncodeToString(h[:]))
}
func BroadcastWriteBeginCommand(org, ticket string) (Command, error) {
	return BroadcastCommand(org, "write", "begin", ticket)
}
func BroadcastWriteStatusCommand(org, ticket string) (Command, error) {
	return BroadcastCommand(org, "write", "status", ticket)
}
func BroadcastWriteInspectCommand(org, transfer string) (Command, error) {
	return BroadcastCommand(org, "write", "inspect", transfer)
}
func BroadcastWriteAbortCommand(org, transfer string) (Command, error) {
	return BroadcastCommand(org, "write", "abort", transfer)
}
func BroadcastWriteCommitCommand(org, ticket, transfer string) (Command, error) {
	return BroadcastCommand(org, "write", "commit", ticket, transfer)
}
func BroadcastWriteChunkCommand(org, transfer string, offset int, data []byte) (Command, error) {
	return BroadcastCommand(org, "write", "chunk", transfer, strconv.Itoa(offset), base64.StdEncoding.EncodeToString(data))
}
func BroadcastProfilesCommand(org string) (Command, error) { return BroadcastCommand(org, "profiles") }
func BroadcastChannelCommand(org string) (Command, error)  { return BroadcastCommand(org, "channel") }
func BroadcastPreflightCommand(org string) (Command, error) {
	return BroadcastCommand(org, "preflight")
}
func BroadcastEventsCommand(org string) (Command, error) { return BroadcastCommand(org, "events") }
func BroadcastPublishersCommand(org string) (Command, error) {
	return BroadcastCommand(org, "publishers")
}
func BroadcastTakeoverCommand(org string) (Command, error) { return BroadcastCommand(org, "takeover") }
func GetBroadcastEventCommand(org, id string) (Command, error) {
	return BroadcastCommand(org, "event", id, "get")
}
func GetBroadcastOperationCommand(org, id string) (Command, error) {
	return BroadcastCommand(org, "operation", id, "get")
}
func GetBroadcastPublisherCommand(org, id string) (Command, error) {
	return BroadcastCommand(org, "publisher", id, "get")
}

// CreateBroadcastEventCommand preserves the supplied JSON bytes. Metadata is
// validated by the shared server validator; acceptance is not provider completion.
func CreateBroadcastEventCommand(org, request string, metadata json.RawMessage) (Command, error) {
	return BroadcastCommand(org, "event", "create", request, base64.StdEncoding.EncodeToString(metadata))
}
func AttachBroadcastEventCommand(org, request, provider string) (Command, error) {
	return BroadcastCommand(org, "event", "attach", request, provider)
}
func TransitionBroadcastEventCommand(org, id string, action BroadcastTransition, revision int64, request string) (Command, error) {
	return BroadcastCommand(org, "event", id, string(action), strconv.FormatInt(revision, 10), request)
}
func StopBroadcastPublisherCommand(org, id string, revision int64) (Command, error) {
	return BroadcastCommand(org, "publisher", id, "stop", strconv.FormatInt(revision, 10))
}
