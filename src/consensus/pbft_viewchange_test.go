// PEPPER — SEC-P2P03 PBFT view-change sticky-proposal regression tests (823a19e)
// Tests for GetLastPreparedBlock accessor and related PBFT state management.
package consensus

import (
	"testing"
)

// ── GetLastPreparedBlock ─────────────────────────────────────────────────────

func TestGetLastPreparedBlock_InitiallyNilZero(t *testing.T) {
	// Fresh consensus should have no prepared block
	c := &Consensus{}
	blk, height := c.GetLastPreparedBlock()
	if blk != nil {
		t.Error("GetLastPreparedBlock: initial block should be nil")
	}
	if height != 0 {
		t.Errorf("GetLastPreparedBlock: initial height should be 0, got %d", height)
	}
}

func TestGetLastPreparedBlock_ReturnsStoredValues(t *testing.T) {
	c := &Consensus{}
	mock := &mockBlock{height: 5, hash: "abc123def456abc123def456abc123de"}
	c.lastPreparedBlock = mock
	c.lastPreparedHeight = 5

	blk, height := c.GetLastPreparedBlock()
	if blk == nil {
		t.Error("GetLastPreparedBlock: should return stored block")
	}
	if height != 5 {
		t.Errorf("GetLastPreparedBlock: want height 5, got %d", height)
	}
	if blk.GetHash() != mock.hash {
		t.Errorf("GetLastPreparedBlock: wrong block returned")
	}
}

func TestGetLastPreparedBlock_NilAfterClear(t *testing.T) {
	c := &Consensus{}
	c.lastPreparedBlock = &mockBlock{height: 3, hash: "deadbeefdeadbeefdeadbeefdeadbeef"}
	c.lastPreparedHeight = 3

	// Simulate clear (done after commit)
	c.lastPreparedBlock = nil
	c.lastPreparedHeight = 0

	blk, height := c.GetLastPreparedBlock()
	if blk != nil {
		t.Error("GetLastPreparedBlock: should be nil after clear")
	}
	if height != 0 {
		t.Errorf("GetLastPreparedBlock: height should be 0 after clear, got %d", height)
	}
}

func TestGetLastPreparedBlock_HeightMatchesBlock(t *testing.T) {
	c := &Consensus{}
	for _, h := range []uint64{1, 10, 100, 9999} {
		c.lastPreparedBlock = &mockBlock{height: h, hash: "0000000000000000000000000000000a"}
		c.lastPreparedHeight = h
		_, gotHeight := c.GetLastPreparedBlock()
		if gotHeight != h {
			t.Errorf("height mismatch: stored %d got %d", h, gotHeight)
		}
	}
}

// ── ViewChange block hash stability (823a19e behavior) ──────────────────────

// TestLastPreparedBlock_UsedOnViewChange documents the design intent:
// after a view change, the new leader should prefer lastPreparedBlock over
// creating a fresh block. This is tested at the state level (not requiring
// full PBFT execution).
func TestLastPreparedBlock_PersistAcrossViewChange(t *testing.T) {
	c := &Consensus{currentView: 1}
	prepared := &mockBlock{height: 7, hash: "7777777777777777777777777777777a"}
	c.lastPreparedBlock = prepared
	c.lastPreparedHeight = 7

	// Simulate view change: view increments, lastPreparedBlock preserved
	c.currentView = 2

	blk, height := c.GetLastPreparedBlock()
	if blk == nil {
		t.Error("lastPreparedBlock should be preserved across view change")
	}
	if height != 7 {
		t.Errorf("height should be 7 across view change, got %d", height)
	}
	if blk.GetHash() != prepared.hash {
		t.Error("block hash should be stable across view change")
	}
}
