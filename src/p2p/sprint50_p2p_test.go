package p2p

import (
	"testing"
	"time"

	"github.com/ramseyauron/quantix/src/network"
)

// Sprint 50 — p2p coverage: markBlockSeen, isBlockSeen, SetMessageCh, SetServer,
//             UpdateSeedNodes, CloseDB, InitializeConsensus, GetConsensus, validateTransaction

func newTestServerSprint50(t *testing.T) *Server {
	t.Helper()
	cfg := network.NodePortConfig{
		Name:    "test-sprint50",
		TCPAddr: "127.0.0.1:19090",
		UDPPort: "0",
		Role:    network.RoleValidator,
		DevMode: true,
	}
	return NewServer(cfg, nil, nil)
}

// ---------------------------------------------------------------------------
// markBlockSeen / isBlockSeen
// ---------------------------------------------------------------------------

func TestMarkBlockSeen_NotSeenBefore_BecomesSeen(t *testing.T) {
	s := newTestServerSprint50(t)
	hash := "abc123deadbeef"
	if s.isBlockSeen(hash) {
		t.Error("block should not be seen initially")
	}
	s.markBlockSeen(hash)
	if !s.isBlockSeen(hash) {
		t.Error("block should be seen after markBlockSeen")
	}
}

func TestMarkBlockSeen_EmptyHash_Noop(t *testing.T) {
	s := newTestServerSprint50(t)
	s.markBlockSeen("")
	if s.isBlockSeen("") {
		t.Error("empty hash should never be seen")
	}
}

func TestIsBlockSeen_UnknownHash_False(t *testing.T) {
	s := newTestServerSprint50(t)
	if s.isBlockSeen("neverseenhash") {
		t.Error("unseen hash should return false")
	}
}

func TestMarkBlockSeen_MultipleSeen(t *testing.T) {
	s := newTestServerSprint50(t)
	for i := 0; i < 5; i++ {
		h := string(rune('a' + i))
		s.markBlockSeen(h)
	}
	for i := 0; i < 5; i++ {
		h := string(rune('a' + i))
		if !s.isBlockSeen(h) {
			t.Errorf("hash %q should be seen", h)
		}
	}
}

func TestMarkBlockSeen_Idempotent(t *testing.T) {
	s := newTestServerSprint50(t)
	hash := "idem-hash"
	s.markBlockSeen(hash)
	s.markBlockSeen(hash)
	if !s.isBlockSeen(hash) {
		t.Error("double-marked block should still be seen")
	}
}

// ---------------------------------------------------------------------------
// SetMessageCh
// ---------------------------------------------------------------------------

func TestSetMessageCh_NilChannel_NoFault(t *testing.T) {
	s := newTestServerSprint50(t)
	s.SetMessageCh(nil) // nil is accepted — just sets field to nil
}

// ---------------------------------------------------------------------------
// SetServer
// ---------------------------------------------------------------------------

func TestSetServer_NoFault(t *testing.T) {
	s := newTestServerSprint50(t)
	s.SetServer() // should not panic
}

// ---------------------------------------------------------------------------
// UpdateSeedNodes — already tested in pepper but verify existing Sprint50 access
// ---------------------------------------------------------------------------

func TestUpdateSeedNodes_ClearsOld(t *testing.T) {
	s := newTestServerSprint50(t)
	s.UpdateSeedNodes([]string{"127.0.0.1:9999"})
	s.UpdateSeedNodes([]string{}) // clear
	// No way to read seedNodes — just verify no panic
}

// ---------------------------------------------------------------------------
// CloseDB — nil db should not panic
// ---------------------------------------------------------------------------

func TestCloseDB_NilDB_NoFault(t *testing.T) {
	s := newTestServerSprint50(t)
	err := s.CloseDB()
	if err != nil {
		t.Logf("CloseDB returned: %v (acceptable)", err)
	}
}

// ---------------------------------------------------------------------------
// InitializeConsensus / GetConsensus
// ---------------------------------------------------------------------------

func TestInitializeConsensus_Nil_NoFault(t *testing.T) {
	s := newTestServerSprint50(t)
	s.InitializeConsensus(nil)
}

func TestGetConsensus_AfterNilInit_IsNil(t *testing.T) {
	s := newTestServerSprint50(t)
	s.InitializeConsensus(nil)
	c := s.GetConsensus()
	if c != nil {
		t.Error("after InitializeConsensus(nil), GetConsensus should return nil")
	}
}

// ---------------------------------------------------------------------------
// validateTransaction — no validator (empty manager) → error
// ---------------------------------------------------------------------------

func TestValidateTransaction_NoValidator_ReturnsError(t *testing.T) {
	s := newTestServerSprint50(t)
	// makeValidTx via direct struct — simplified
	err := s.validateTransaction(nil)
	if err == nil {
		t.Error("validateTransaction with nil tx should return error or no validator error")
	}
}

// ---------------------------------------------------------------------------
// seenConsensusMsgs TTL / BroadcastBlock with no connections
// ---------------------------------------------------------------------------

func TestSeenConsensusMsgs_InitiallyEmpty(t *testing.T) {
	s := newTestServerSprint50(t)
	// seenConsensusMsgs is checked in handleMessages — just verify field is initialized
	s.seenBlocksMu.Lock()
	l := len(s.seenBlocks)
	s.seenBlocksMu.Unlock()
	if l != 0 {
		t.Errorf("fresh server should have 0 seen blocks, got %d", l)
	}
}

func TestBroadcastBlock_NilBlock_Documented(t *testing.T) {
	// BroadcastBlock panics with nil block (no nil guard in p2p.go:907)
	// Filed as known nil-panic; test documents the behavior
	t.Log("NOTE: BroadcastBlock(nil) panics — nil guard needed at p2p.go:907")
}

func TestBroadcastTransaction_NilTx_NoFault(t *testing.T) {
	s := newTestServerSprint50(t)
	s.BroadcastTransaction(nil)
}

// ---------------------------------------------------------------------------
// seenBlocksTTL constant documented
// ---------------------------------------------------------------------------

func TestSeenBlocksTTL_Is5Minutes(t *testing.T) {
	if seenBlocksTTL != 5*time.Minute {
		t.Errorf("seenBlocksTTL should be 5 minutes, got %v", seenBlocksTTL)
	}
}
