package bibleit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// NewBroadcastWriteIntentInput computes the signed intent input from exact JSON
// bytes, using the same operation/target/revision bounds as the SSH constructor.
// Retain these bytes and the resulting ticket before submitting a mutation.
func NewBroadcastWriteIntentInput(operation BroadcastWriteOperation, target string, revision int64, payload []byte) (BroadcastWriteIntentInput, error) {
	if _, e := BroadcastWriteIntentCommand("00000000-0000-0000-0000-000000000000", operation, target, revision, payload); e != nil {
		return BroadcastWriteIntentInput{}, e
	}
	if !broadcastJSONObject(payload) {
		return BroadcastWriteIntentInput{}, errors.New("broadcast payload must be a JSON object")
	}
	if operation == BroadcastCreate {
		target = "new"
	}
	h := sha256.Sum256(payload)
	return BroadcastWriteIntentInput{Operation: BroadcastWriteIntentInputOperation(operation), TargetID: &target, ExpectedRevision: revision, PayloadBytes: int64(len(payload)), PayloadSha256: hex.EncodeToString(h[:])}, nil
}

// BroadcastWriteOptions creates a new reader over a private copy of the exact
// signed payload. Do not use Options.JSON for signed Setup writes: marshaling a
// RawMessage can compact whitespace and change the checksum. Zero revision is
// only appropriate for create; all other Setup mutations require a positive one.
// A retry must retain the original ticket and bytes, never mint a new intent.
func BroadcastWriteOptions(ticket string, revision int64, payload []byte) (*Options, error) {
	if !broadcastTicket(ticket) || revision < 0 || len(payload) < 2 || len(payload) > MaxBroadcastDocumentBytes || !broadcastJSONObject(payload) {
		return nil, errors.New("invalid signed broadcast write")
	}
	body := bytes.Clone(payload)
	options := &Options{Body: bytes.NewReader(body), ContentType: "application/json", ContentLength: int64(len(body)), IdempotencyKey: ticket}
	if revision > 0 {
		options.IfMatch = fmt.Sprintf(`"setup-%d"`, revision)
	}
	return options, nil
}
func broadcastJSONObject(payload []byte) bool {
	b := bytes.TrimSpace(payload)
	return len(b) > 0 && b[0] == '{' && json.Valid(b)
}
