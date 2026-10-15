package responses

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/bsfdsagfadg/vertex/internal/db"
	_ "github.com/glebarez/go-sqlite"
)

func testStore(t *testing.T) *ResponseStore {
	t.Helper()
	sdb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	sdb.SetMaxOpenConns(1)
	if err := db.CreateTables(sdb); err != nil {
		sdb.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { sdb.Close() })
	return NewResponseStore(sdb)
}

func TestResponseStoreIdempotencyAndCAS(t *testing.T) {
	s := testStore(t)
	now := time.Now()
	r := ResponseRecord{ID: "resp_1", Status: "in_progress", Model: "m", Input: json.RawMessage(`[{"type":"message"}]`), CreatedAt: now.Unix(), ExpiresAt: now.Add(time.Hour).Unix()}
	rec := IdempotencyRecord{Endpoint: "/v1/responses", Key: "k", BodyHash: "h", CreatedAt: now.Unix(), ExpiresAt: now.Add(time.Hour).Unix()}
	id, reused, conflict, err := s.CreateIdempotent(context.Background(), r, rec)
	if err != nil || id != r.ID || reused || conflict {
		t.Fatalf("first create: %v %q %v %v", err, id, reused, conflict)
	}
	id, reused, conflict, err = s.CreateIdempotent(context.Background(), ResponseRecord{ID: "resp_2", ExpiresAt: r.ExpiresAt}, rec)
	if err != nil || id != r.ID || !reused || conflict {
		t.Fatalf("reuse: %v %q %v %v", err, id, reused, conflict)
	}
	_, reused, conflict, err = s.CreateIdempotent(context.Background(), ResponseRecord{ID: "resp_3", ExpiresAt: r.ExpiresAt}, IdempotencyRecord{Endpoint: rec.Endpoint, Key: rec.Key, BodyHash: "different", ExpiresAt: rec.ExpiresAt})
	if err != nil || !reused || !conflict {
		t.Fatalf("conflict: %v %v %v", err, reused, conflict)
	}
	r.Status = "completed"
	ok, err := s.UpdateTerminalCAS(context.Background(), r)
	if err != nil || !ok {
		t.Fatalf("CAS completion: %v %v", err, ok)
	}
	ok, err = s.UpdateTerminalCAS(context.Background(), r)
	if err != nil || ok {
		t.Fatalf("second CAS should lose: %v %v", err, ok)
	}
}

func TestResponseStoreEventSequenceReplay(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	seq, err := s.AllocateEventSequence(ctx, "response", "r")
	if err != nil || seq != 1 {
		t.Fatalf("first sequence: %v %d", err, seq)
	}
	if err := s.AppendEvent(ctx, Event{ResourceKind: "response", ResourceID: "r", Sequence: seq, EventID: "e1", EventType: "created", Data: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
	seq, _ = s.AllocateEventSequence(ctx, "response", "r")
	_ = s.AppendEvent(ctx, Event{ResourceKind: "response", ResourceID: "r", Sequence: seq, EventID: "e2", EventType: "completed", Data: json.RawMessage(`{}`)})
	events, err := s.ReplayEvents(ctx, "response", "r", "e1")
	if err != nil || len(events) != 1 || events[0].EventID != "e2" {
		t.Fatalf("replay: %v %#v", err, events)
	}
}
func TestConsumeToolStatesAtomic(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	now := time.Now()
	tool := ToolState{
		ExternalCallID: "call_123",
		ResponseID:     "resp_1",
		State:          []byte("state_data"),
		ExpiresAt:      now.Add(time.Hour).Unix(),
	}
	if err := s.PutToolState(ctx, tool); err != nil {
		t.Fatal(err)
	}
	// First consumption with hash1 should succeed
	if err := s.ConsumeToolStates(ctx, []string{"call_123"}, "hash1", now); err != nil {
		t.Fatalf("first consume should succeed: %v", err)
	}
	// Same consumption with hash1 should succeed (idempotent for same transcript)
	if err := s.ConsumeToolStates(ctx, []string{"call_123"}, "hash1", now); err != nil {
		t.Fatalf("idempotent consume with same hash should succeed: %v", err)
	}
	// Consumption with different hash should fail
	if err := s.ConsumeToolStates(ctx, []string{"call_123"}, "hash2", now); err == nil {
		t.Fatal("consume with different hash should fail")
	}
}
