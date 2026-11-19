package bibleit

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// SSHBroadcastValue is a bounded lifecycle projection. Value is null for a
// disconnected channel. Unknown JSON fields remain available in Value; decode
// into a named domain model with Decode when the REST model is appropriate.
type SSHBroadcastValue struct {
	RecordType string
	Value      json.RawMessage
	Fields     map[string]string
}

func (v SSHBroadcastValue) Decode(destination any) error { return json.Unmarshal(v.Value, destination) }

// DecodeSSHBroadcastValue validates the advertised framing before exposing JSON.
// It never includes provider credentials omitted by the server projection.
func DecodeSSHBroadcastValue(record Record) (SSHBroadcastValue, error) {
	f := record.Fields
	validRecord := record.Type == "ok" || (record.Type == "youtube_event" && f["record_type"] == "event") || (record.Type == "publisher" && f["record_type"] == "publisher")
	if !validRecord {
		return SSHBroadcastValue{}, errors.New("invalid broadcast lifecycle record")
	}
	if f["document_format"] != "broadcast-lifecycle-json-v1" {
		return SSHBroadcastValue{}, errors.New("unsupported broadcast lifecycle format")
	}
	switch f["record_type"] {
	case "channel", "event", "operation", "publisher", "preflight":
	default:
		return SSHBroadcastValue{}, errors.New("unsupported broadcast lifecycle record type")
	}
	size, e := broadcastResponseNumber(f, "bytes", 0, MaxBroadcastDocumentBytes)
	if e != nil {
		return SSHBroadcastValue{}, e
	}
	b, e := decodeBroadcastBase64(f["data_base64"], size)
	if e != nil {
		return SSHBroadcastValue{}, e
	}
	var envelope map[string]json.RawMessage
	if e = json.Unmarshal(b, &envelope); e != nil || envelope["value"] == nil {
		return SSHBroadcastValue{}, errors.New("invalid broadcast lifecycle envelope")
	}
	return SSHBroadcastValue{RecordType: f["record_type"], Value: envelope["value"], Fields: copyBroadcastFields(f)}, nil
}

// SSHBroadcastSetupMetadata binds document pages to one immutable revision.
type SSHBroadcastSetupMetadata struct {
	OrganizationID  string
	ID              string
	Name            string
	Revision        int64
	DocumentVersion int64
	DocumentBytes   int
	DocumentSHA256  string
	ChunkBytes      int
	Fields          map[string]string
}

func DecodeSSHBroadcastSetupMetadata(record Record) (SSHBroadcastSetupMetadata, error) {
	f := record.Fields
	if record.Type != "ok" {
		return SSHBroadcastSetupMetadata{}, errors.New("invalid broadcast setup record")
	}
	if f["document_format"] != "studio-setup-json-v1" || !broadcastUUID.MatchString(f["organization_id"]) || !broadcastUUID.MatchString(f["id"]) || !broadcastDigest.MatchString(f["document_sha256"]) {
		return SSHBroadcastSetupMetadata{}, errors.New("invalid broadcast setup metadata")
	}
	revision, e := broadcastResponseNumber(f, "revision", 1, 1<<63-1)
	if e != nil {
		return SSHBroadcastSetupMetadata{}, e
	}
	version, e := broadcastResponseNumber(f, "document_version", 1, 1<<63-1)
	if e != nil {
		return SSHBroadcastSetupMetadata{}, e
	}
	size, e := broadcastResponseNumber(f, "document_bytes", 2, MaxBroadcastDocumentBytes)
	if e != nil {
		return SSHBroadcastSetupMetadata{}, e
	}
	chunk, e := broadcastResponseNumber(f, "chunk_bytes", 1, BroadcastChunkBytes)
	if e != nil {
		return SSHBroadcastSetupMetadata{}, e
	}
	return SSHBroadcastSetupMetadata{OrganizationID: f["organization_id"], ID: f["id"], Name: f["name"], Revision: revision, DocumentVersion: version, DocumentBytes: int(size), DocumentSHA256: f["document_sha256"], ChunkBytes: int(chunk), Fields: copyBroadcastFields(f)}, nil
}

// SSHBroadcastDocumentChunk is one revision-bound page, never an implicit retry.
type SSHBroadcastDocumentChunk struct {
	Metadata   SSHBroadcastSetupMetadata
	Offset     int
	NextOffset int
	EOF        bool
	Data       []byte
}

func DecodeSSHBroadcastDocumentChunk(record Record) (SSHBroadcastDocumentChunk, error) {
	m, e := DecodeSSHBroadcastSetupMetadata(record)
	if e != nil {
		return SSHBroadcastDocumentChunk{}, e
	}
	f := record.Fields
	offset, e := broadcastResponseNumber(f, "offset", 0, int64(m.DocumentBytes))
	if e != nil {
		return SSHBroadcastDocumentChunk{}, e
	}
	next, e := broadcastResponseNumber(f, "next_offset", offset, int64(m.DocumentBytes))
	if e != nil {
		return SSHBroadcastDocumentChunk{}, e
	}
	size, e := broadcastResponseNumber(f, "bytes", 0, int64(m.ChunkBytes))
	if e != nil {
		return SSHBroadcastDocumentChunk{}, e
	}
	eof, e := broadcastResponseBool(f, "eof")
	if e != nil {
		return SSHBroadcastDocumentChunk{}, e
	}
	if next != offset+size || eof != (next == int64(m.DocumentBytes)) || (offset%int64(m.ChunkBytes) != 0 && offset != int64(m.DocumentBytes)) || size != min(int64(m.ChunkBytes), int64(m.DocumentBytes)-offset) {
		return SSHBroadcastDocumentChunk{}, errors.New("inconsistent broadcast document offsets")
	}
	data, e := decodeBroadcastBase64(f["data_base64"], size)
	if e != nil {
		return SSHBroadcastDocumentChunk{}, e
	}
	return SSHBroadcastDocumentChunk{Metadata: m, Offset: int(offset), NextOffset: int(next), EOF: eof, Data: data}, nil
}

// VerifyBroadcastDocument checks the exact exported bytes against Setup metadata.
func VerifyBroadcastDocument(metadata SSHBroadcastSetupMetadata, data []byte) error {
	h := sha256.Sum256(data)
	if len(data) != metadata.DocumentBytes || hex.EncodeToString(h[:]) != metadata.DocumentSHA256 {
		return errors.New("broadcast document checksum mismatch")
	}
	var document map[string]json.RawMessage
	if json.Unmarshal(data, &document) != nil || document == nil {
		return errors.New("invalid broadcast document JSON")
	}
	return nil
}

// SSHBroadcastWriteReceipt preserves both known and future durable outcomes.
// Never treat an unknown state or not_seen as permission to mint a replacement.
type SSHBroadcastWriteReceipt struct {
	RequestID       string
	ResultState     string
	Terminal        bool
	Applied         bool
	Replay          bool
	Operation       BroadcastWriteOperation
	InputHash       string
	ID              string
	Revision        int64
	DeletedRevision int64
	ErrorCode       string
	Fields          map[string]string
}

func DecodeSSHBroadcastWriteReceipt(record Record) (SSHBroadcastWriteReceipt, error) {
	f := record.Fields
	if record.Type != "ok" {
		return SSHBroadcastWriteReceipt{}, errors.New("invalid broadcast receipt record")
	}
	if !broadcastTicket(f["request_id"]) || f["result_state"] == "" {
		return SSHBroadcastWriteReceipt{}, errors.New("invalid broadcast receipt")
	}
	terminal, e := broadcastResponseBool(f, "terminal")
	if e != nil {
		return SSHBroadcastWriteReceipt{}, e
	}
	r := SSHBroadcastWriteReceipt{RequestID: f["request_id"], ResultState: f["result_state"], Terminal: terminal, Operation: BroadcastWriteOperation(f["operation"]), InputHash: f["input_hash"], ID: f["id"], ErrorCode: f["error_code"], Fields: copyBroadcastFields(f)}
	if _, ok := f["applied"]; ok {
		r.Applied, e = broadcastResponseBool(f, "applied")
		if e != nil {
			return r, e
		}
	}
	if _, ok := f["replay"]; ok {
		r.Replay, e = broadcastResponseBool(f, "replay")
		if e != nil {
			return r, e
		}
	}
	if _, ok := f["revision"]; ok {
		r.Revision, e = broadcastResponseNumber(f, "revision", 1, 1<<63-1)
		if e != nil {
			return r, e
		}
	}
	if _, ok := f["deleted_revision"]; ok {
		r.DeletedRevision, e = broadcastResponseNumber(f, "deleted_revision", 1, 1<<63-1)
		if e != nil {
			return r, e
		}
	}
	if r.ID != "" && !broadcastUUID.MatchString(r.ID) {
		return r, errors.New("invalid broadcast receipt identifier")
	}
	if r.InputHash != "" && !broadcastDigest.MatchString(r.InputHash) {
		return r, errors.New("invalid broadcast receipt input hash")
	}
	if r.Applied && !r.Terminal {
		return r, errors.New("nonterminal broadcast receipt cannot be applied")
	}
	return r, nil
}
func broadcastResponseNumber(fields map[string]string, key string, min, max int64) (int64, error) {
	v := fields[key]
	if !broadcastNumber(v, min, max) {
		return 0, fmt.Errorf("invalid broadcast response %s", key)
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n, nil
}
func broadcastResponseBool(fields map[string]string, key string) (bool, error) {
	switch fields[key] {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid broadcast response %s", key)
	}
}
func decodeBroadcastBase64(encoded string, size int64) ([]byte, error) {
	if len(encoded) > base64.StdEncoding.EncodedLen(MaxBroadcastDocumentBytes) {
		return nil, errors.New("broadcast response exceeds size limit")
	}
	data, e := base64.StdEncoding.Strict().DecodeString(encoded)
	if e != nil || int64(len(data)) != size || base64.StdEncoding.EncodeToString(data) != encoded {
		return nil, errors.New("invalid broadcast response bytes")
	}
	return data, nil
}
func copyBroadcastFields(fields map[string]string) map[string]string {
	out := make(map[string]string, len(fields))
	for k, v := range fields {
		out[k] = v
	}
	return out
}
