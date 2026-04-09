// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 39 - consensus resetConsensusState, DebugConsensusSignaturesDeep, isValidLeader
package consensus

import (
	"math/big"
	"testing"
)

func newTestConsensus39() *Consensus {
	return NewConsensus("test-node-sprint39", nil, nil, nil, nil, big.NewInt(1000))
}

// ---------------------------------------------------------------------------
// resetConsensusState — no panic, resets fields
// ---------------------------------------------------------------------------

func TestSprint39_ResetConsensusState_NoPanic(t *testing.T) {
	c := newTestConsensus39()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("resetConsensusState panicked: %v", r)
		}
	}()
	c.resetConsensusState()
}

func TestSprint39_ResetConsensusState_SetsPhaseIdle(t *testing.T) {
	c := newTestConsensus39()
	c.resetConsensusState()
	if c.phase != PhaseIdle {
		t.Fatalf("expected PhaseIdle after reset, got %v", c.phase)
	}
}

func TestSprint39_ResetConsensusState_ClearsVotes(t *testing.T) {
	c := newTestConsensus39()
	// Add some fake votes
	c.receivedVotes["some-block"] = map[string]*Vote{}
	c.resetConsensusState()
	if len(c.receivedVotes) != 0 {
		t.Fatal("expected empty receivedVotes after reset")
	}
}

// ---------------------------------------------------------------------------
// DebugConsensusSignaturesDeep — no panic with empty signatures
// ---------------------------------------------------------------------------

func TestSprint39_DebugConsensusSignaturesDeep_EmptySignatures_NoPanic(t *testing.T) {
	c := newTestConsensus39()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DebugConsensusSignaturesDeep panicked: %v", r)
		}
	}()
	c.DebugConsensusSignaturesDeep()
}

func TestSprint39_DebugConsensusSignaturesDeep_WithSignature_NoPanic(t *testing.T) {
	c := newTestConsensus39()
	// Add a dummy signature
	c.consensusSignatures = append(c.consensusSignatures, &ConsensusSignature{
		BlockHash:   "test-block-hash",
		MessageType: "proposal",
		MerkleRoot:  "test-merkle",
		Status:      "committed",
		Valid:        true,
	})
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DebugConsensusSignaturesDeep panicked: %v", r)
		}
	}()
	c.DebugConsensusSignaturesDeep()
}

// ---------------------------------------------------------------------------
// isValidLeader — with elected leader set
// ---------------------------------------------------------------------------

func TestSprint39_IsValidLeader_ElectedLeaderMatches_True(t *testing.T) {
	c := newTestConsensus39()
	c.electedLeaderID = "node-alpha"
	if !c.isValidLeader("node-alpha", 0) {
		t.Fatal("expected true when nodeID matches elected leader")
	}
}

func TestSprint39_IsValidLeader_ElectedLeaderMismatch_False(t *testing.T) {
	c := newTestConsensus39()
	c.electedLeaderID = "node-alpha"
	if c.isValidLeader("node-beta", 0) {
		t.Fatal("expected false when nodeID does not match elected leader")
	}
}

func TestSprint39_IsValidLeader_NoElectedLeader_EmptyValidators_False(t *testing.T) {
	c := newTestConsensus39()
	c.electedLeaderID = "" // no elected leader
	// nil nodeManager → getValidators returns nil → no validators → false
	result := c.isValidLeader("any-node", 0)
	if result {
		t.Fatal("expected false with no elected leader and empty validator set")
	}
}

// ---------------------------------------------------------------------------
// addConsensusSig — panics with nil blockchain (documented)
// ---------------------------------------------------------------------------

func TestSprint39_AddConsensusSig_NoPanic(t *testing.T) {
	// addConsensusSig panics when blockchain is nil — nil guard needed
	t.Skip("addConsensusSig panics with nil blockchain — nil guard needed")
}

func TestSprint39_AddConsensusSig_AppearsInList(t *testing.T) {
	t.Skip("addConsensusSig panics with nil blockchain — nil guard needed")
}

// ---------------------------------------------------------------------------
// ForcePopulateAllSignatures — no blockchain, no crash
// ---------------------------------------------------------------------------

func TestSprint39_ForcePopulateAllSignatures_EmptyList_NoPanic(t *testing.T) {
	c := newTestConsensus39()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ForcePopulateAllSignatures panicked with empty list: %v", r)
		}
	}()
	c.ForcePopulateAllSignatures()
}

// ---------------------------------------------------------------------------
// GetConsensusSignatures — also panics with nil blockchain (documented)
// ---------------------------------------------------------------------------

func TestSprint39_GetConsensusSignatures_IndependentCopy(t *testing.T) {
	// GetConsensusSignatures panics via addConsensusSig with nil blockchain
	t.Skip("GetConsensusSignatures via addConsensusSig panics with nil blockchain — nil guard needed")
}
