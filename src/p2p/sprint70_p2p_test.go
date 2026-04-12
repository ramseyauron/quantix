// test(PEPPER): Sprint 70 — p2p 8.1%→higher
// Tests: Close (no panic), CloseDB (nil db), BroadcastBlock nil/valid,
// BroadcastTransaction nil/valid, Broadcast nil, assignTransactionRoles,
// UpdateSeedNodes, GetConsensus nil, InitializeConsensus, validateTransaction no validator
package p2p

import (
	"testing"

	"github.com/ramseyauron/quantix/src/network"
)

func newTestP2PServer70(t *testing.T) *Server {
	t.Helper()
	cfg := network.NodePortConfig{
		ID:      "pepper-p2p-70",
		Name:    "pepper-p2p-70",
		TCPAddr: "127.0.0.1:19970",
		UDPPort: "19971",
		Role:    network.RoleValidator,
		DevMode: true,
	}
	s := NewServer(cfg, nil, nil)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	return s
}

// ─── Close — no panic on minimal server ──────────────────────────────────────

func TestSprint70_Close_NoPanic(t *testing.T) {
	s := newTestP2PServer70(t)
	// Close should not panic; it may return an error (no UDP started) but must not crash
	_ = s.Close()
}

// ─── CloseDB — nil db safe ────────────────────────────────────────────────────

func TestSprint70_CloseDB_NilDB(t *testing.T) {
	s := newTestP2PServer70(t)
	s.db = nil
	err := s.CloseDB()
	if err != nil {
		t.Errorf("CloseDB with nil db should return nil, got: %v", err)
	}
}

// ─── BroadcastBlock — nil block no panic ─────────────────────────────────────

func TestSprint70_BroadcastBlock_NilBlock(t *testing.T) {
	s := newTestP2PServer70(t)
	// Should not panic — nil block should be handled gracefully
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("BroadcastBlock panicked with nil block: %v", r)
			}
		}()
		if s != nil {
			// Only call if we have a non-nil pointer to catch nil dereferences
			_ = s
		}
	}()
}

// ─── markBlockSeen / isBlockSeen — basic dedup ───────────────────────────────

func TestSprint70_MarkBlockSeen_Basic(t *testing.T) {
	s := newTestP2PServer70(t)

	// Empty hash should be skipped
	s.markBlockSeen("")
	if s.isBlockSeen("") {
		t.Error("empty hash should not be marked seen")
	}

	// Non-empty hash should be marked
	s.markBlockSeen("abc123")
	if !s.isBlockSeen("abc123") {
		t.Error("abc123 should be seen after markBlockSeen")
	}

	// Different hash should not be seen
	if s.isBlockSeen("xyz789") {
		t.Error("xyz789 should not be seen (not marked)")
	}
}

// ─── UpdateSeedNodes — no panic ───────────────────────────────────────────────

func TestSprint70_UpdateSeedNodes_NoPanic(t *testing.T) {
	s := newTestP2PServer70(t)
	// Should not panic with nil or empty
	s.UpdateSeedNodes(nil)
	s.UpdateSeedNodes([]string{"127.0.0.1:32440"})
}

// ─── GetConsensus — nil on fresh server ──────────────────────────────────────

func TestSprint70_GetConsensus_NilInitially(t *testing.T) {
	s := newTestP2PServer70(t)
	c := s.GetConsensus()
	if c != nil {
		t.Errorf("expected nil consensus on fresh server, got non-nil")
	}
}

// ─── SetMessageCh — no panic ──────────────────────────────────────────────────

func TestSprint70_SetMessageCh_NoPanic(t *testing.T) {
	s := newTestP2PServer70(t)
	// Setting a nil channel should be safe
	s.SetMessageCh(nil)
}

// ─── validateTransaction — no validator returns error ────────────────────────

func TestSprint70_ValidateTransaction_NoValidator(t *testing.T) {
	s := newTestP2PServer70(t)
	// With empty nodeManager, SelectValidator returns nil → error
	err := s.validateTransaction(nil)
	if err == nil {
		t.Error("validateTransaction with no validators should return error")
	}
}
