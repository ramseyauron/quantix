// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 89b — consensus 43.9%→higher
// Covers: SetConfiguredTotalNodes, logModeTransition (via ActiveConsensusMode), mode thresholds
package consensus_test

import (
	"math/big"
	"testing"

	"github.com/ramseyauron/quantix/src/consensus"
)

// newConsensus89 creates a Consensus for testing pure functions.
func newConsensus89(t *testing.T) *consensus.Consensus {
	t.Helper()
	return consensus.NewConsensus("node-89", nil, nil, nil, nil, big.NewInt(1000))
}

// ---------------------------------------------------------------------------
// SetConfiguredTotalNodes
// ---------------------------------------------------------------------------

func TestSprint89b_SetConfiguredTotalNodes_Basic(t *testing.T) {
	c := newConsensus89(t)
	c.SetConfiguredTotalNodes(4) // no panic
}

func TestSprint89b_SetConfiguredTotalNodes_Zero(t *testing.T) {
	c := newConsensus89(t)
	c.SetConfiguredTotalNodes(0)
}

func TestSprint89b_SetConfiguredTotalNodes_Large(t *testing.T) {
	c := newConsensus89(t)
	c.SetConfiguredTotalNodes(100)
}

// ---------------------------------------------------------------------------
// ActiveConsensusMode — uses getTotalNodes → configuredTotalNodes
// Also exercises logModeTransition indirectly
// ---------------------------------------------------------------------------

func TestSprint89b_ActiveConsensusMode_BelowThreshold(t *testing.T) {
	c := newConsensus89(t)
	c.SetConfiguredTotalNodes(2)
	mode := c.ActiveConsensusMode()
	if mode != consensus.DEVNET_SOLO {
		t.Errorf("expected DEVNET_SOLO with 2 nodes, got %v", mode)
	}
}

func TestSprint89b_ActiveConsensusMode_AtThreshold(t *testing.T) {
	c := newConsensus89(t)
	c.SetConfiguredTotalNodes(consensus.MinPBFTValidators)
	mode := c.ActiveConsensusMode()
	if mode != consensus.PBFT {
		t.Errorf("expected PBFT with %d nodes, got %v", consensus.MinPBFTValidators, mode)
	}
}

func TestSprint89b_ActiveConsensusMode_AboveThreshold(t *testing.T) {
	c := newConsensus89(t)
	c.SetConfiguredTotalNodes(10)
	mode := c.ActiveConsensusMode()
	if mode != consensus.PBFT {
		t.Errorf("expected PBFT with 10 nodes, got %v", mode)
	}
}

// ---------------------------------------------------------------------------
// Package-level GetConsensusMode (not a method)
// ---------------------------------------------------------------------------

func TestSprint89b_GetConsensusMode_BelowMin(t *testing.T) {
	mode := consensus.GetConsensusMode(1)
	if mode != consensus.DEVNET_SOLO {
		t.Errorf("expected DEVNET_SOLO for count=1, got %v", mode)
	}
}

func TestSprint89b_GetConsensusMode_AtMin(t *testing.T) {
	mode := consensus.GetConsensusMode(consensus.MinPBFTValidators)
	if mode != consensus.PBFT {
		t.Errorf("expected PBFT for count=%d, got %v", consensus.MinPBFTValidators, mode)
	}
}

// ---------------------------------------------------------------------------
// GetConsensusState after SetConfiguredTotalNodes
// ---------------------------------------------------------------------------

func TestSprint89b_GetConsensusState_WithConfiguredNodes_NonEmpty(t *testing.T) {
	c := newConsensus89(t)
	c.SetConfiguredTotalNodes(4)
	s := c.GetConsensusState()
	if s == "" {
		t.Error("GetConsensusState returned empty string")
	}
}

// ---------------------------------------------------------------------------
// ConsensusMode string representation
// ---------------------------------------------------------------------------

func TestSprint89b_ConsensusModeString_DEVNET_SOLO(t *testing.T) {
	if consensus.DEVNET_SOLO.String() == "" {
		t.Error("DEVNET_SOLO.String() is empty")
	}
}

func TestSprint89b_ConsensusModeString_PBFT(t *testing.T) {
	if consensus.PBFT.String() == "" {
		t.Error("PBFT.String() is empty")
	}
}

// ---------------------------------------------------------------------------
// ForcePopulateAllSignatures — no panic
// ---------------------------------------------------------------------------

func TestSprint89b_ForcePopulateAllSignatures_NoPanic(t *testing.T) {
	c := newConsensus89(t)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ForcePopulateAllSignatures panicked: %v", r)
		}
	}()
	c.ForcePopulateAllSignatures()
}

// ---------------------------------------------------------------------------
// State intact after SetConfiguredTotalNodes
// ---------------------------------------------------------------------------

func TestSprint89b_StateIntact_AfterSetConfiguredTotalNodes(t *testing.T) {
	c := newConsensus89(t)
	c.SetConfiguredTotalNodes(5)

	if v := c.GetCurrentView(); v != 0 {
		t.Errorf("GetCurrentView = %d, want 0", v)
	}
	if h := c.GetCurrentHeight(); h != 0 {
		t.Errorf("GetCurrentHeight = %d, want 0", h)
	}
}
