// Sprint 19 — state/smr coverage: pure functions, StateMachine getters, FinalStateInfo helpers
package state

import (
	"os"
	"testing"
)

// newTestStorage19 creates a minimal Storage in a temp dir for testing.
func newTestStorage19(t *testing.T) (*Storage, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-smr19-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	s, err := NewStorage(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("NewStorage: %v", err)
	}
	cleanup := func() {
		if s != nil {
			s.Close()
		}
		os.RemoveAll(dir)
	}
	return s, cleanup
}

// newTestSM19 creates a StateMachine with a real Storage for testing.
func newTestSM19(t *testing.T, nodeID string, validators []string) (*StateMachine, func()) {
	t.Helper()
	st, cleanup := newTestStorage19(t)
	sm := NewStateMachine(st, nodeID, validators)
	return sm, cleanup
}

// ─── calculateQuorumSize ───────────────────────────────────────────────────────

func TestSprint19_CalcQuorum_Zero(t *testing.T) {
	if got := calculateQuorumSize(0); got != 1 {
		t.Errorf("quorum(0) = %d, want 1", got)
	}
}

func TestSprint19_CalcQuorum_Table(t *testing.T) {
	cases := []struct{ n, want int }{
		{1, 1}, {2, 2}, {3, 3}, {4, 3}, {5, 4}, {7, 5}, {10, 7}, {100, 67},
	}
	for _, c := range cases {
		if got := calculateQuorumSize(c.n); got != c.want {
			t.Errorf("calculateQuorumSize(%d) = %d, want %d", c.n, got, c.want)
		}
	}
}

// ─── mapMessageTypeToStatus (package-level func) ──────────────────────────────

func TestSprint19_MapMsgTypeToStatus_AllCases(t *testing.T) {
	cases := map[string]string{
		"proposal": "proposed",
		"prepare":  "prepared",
		"commit":   "committed",
		"timeout":  "view_change",
		"unknown":  "unknown",
		"":         "unknown",
	}
	for input, want := range cases {
		if got := mapMessageTypeToStatus(input); got != want {
			t.Errorf("mapMessageTypeToStatus(%q) = %q, want %q", input, got, want)
		}
	}
}

// ─── mapValidToStatus (package-level func) ───────────────────────────────────

func TestSprint19_MapValidToStatus_True(t *testing.T) {
	if s := mapValidToStatus(true); s != "Valid" {
		t.Errorf("got %q, want \"Valid\"", s)
	}
}

func TestSprint19_MapValidToStatus_False(t *testing.T) {
	if s := mapValidToStatus(false); s != "Invalid" {
		t.Errorf("got %q, want \"Invalid\"", s)
	}
}

// ─── NewStateMachine ──────────────────────────────────────────────────────────

func TestSprint19_NewSM_NotNil(t *testing.T) {
	sm, cleanup := newTestSM19(t, "node-1", nil)
	defer cleanup()
	if sm == nil {
		t.Fatal("NewStateMachine returned nil")
	}
}

func TestSprint19_NewSM_NodeID(t *testing.T) {
	sm, cleanup := newTestSM19(t, "node-abc", nil)
	defer cleanup()
	if sm.nodeID != "node-abc" {
		t.Errorf("nodeID = %q, want %q", sm.nodeID, "node-abc")
	}
}

func TestSprint19_NewSM_EmptyValidators(t *testing.T) {
	sm, cleanup := newTestSM19(t, "node-1", nil)
	defer cleanup()
	if len(sm.validators) != 0 {
		t.Errorf("expected 0 validators, got %d", len(sm.validators))
	}
}

func TestSprint19_NewSM_WithValidators(t *testing.T) {
	sm, cleanup := newTestSM19(t, "v1", []string{"v1", "v2", "v3", "v4"})
	defer cleanup()
	if len(sm.validators) != 4 {
		t.Errorf("expected 4 validators, got %d", len(sm.validators))
	}
}

func TestSprint19_NewSM_QuorumWith4(t *testing.T) {
	sm, cleanup := newTestSM19(t, "v1", []string{"v1", "v2", "v3", "v4"})
	defer cleanup()
	if sm.quorumSize != 3 {
		t.Errorf("quorumSize = %d, want 3", sm.quorumSize)
	}
}

func TestSprint19_IsValidator_True(t *testing.T) {
	sm, cleanup := newTestSM19(t, "v1", []string{"v1", "v2", "v3"})
	defer cleanup()
	if !sm.isValidator() {
		t.Error("v1 should be a validator")
	}
}

func TestSprint19_IsValidator_False(t *testing.T) {
	sm, cleanup := newTestSM19(t, "notInList", []string{"v1", "v2"})
	defer cleanup()
	if sm.isValidator() {
		t.Error("notInList should not be a validator")
	}
}

// ─── GetFinalStates ───────────────────────────────────────────────────────────

func TestSprint19_GetFinalStates_Empty(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	if fs := sm.GetFinalStates(); len(fs) != 0 {
		t.Errorf("expected empty, got %d", len(fs))
	}
}

func TestSprint19_GetFinalStates_Content(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	sm.stateMutex.Lock()
	// BlockHash must be ≥8 chars (PopulatedFinalStates does BlockHash[:8])
	sm.finalStates = []*FinalStateInfo{{
		BlockHash:  "abcdef1234567890",
		MerkleRoot: "mroot",
		Status:     "committed",
		Signature:  "sig",
		Timestamp:  "2026-01-01T00:00:00Z",
	}}
	sm.stateMutex.Unlock()

	got := sm.GetFinalStates()
	if len(got) != 1 || got[0].BlockHash != "abcdef1234567890" {
		t.Errorf("unexpected result: %v", got)
	}
}

// ─── mapValidToSignatureStatus (method) ──────────────────────────────────────

func TestSprint19_MapValidToSigStatus_True(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	if s := sm.mapValidToSignatureStatus(true); s != "Valid" {
		t.Errorf("got %q, want \"Valid\"", s)
	}
}

func TestSprint19_MapValidToSigStatus_False(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	if s := sm.mapValidToSignatureStatus(false); s != "Invalid" {
		t.Errorf("got %q, want \"Invalid\"", s)
	}
}

// ─── determineFinalStateStatus ────────────────────────────────────────────────

func TestSprint19_DetermineStatus_Invalid(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	if s := sm.determineFinalStateStatus("commit", false); s != "invalid" {
		t.Errorf("got %q, want \"invalid\"", s)
	}
}

func TestSprint19_DetermineStatus_ValidTypes(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	cases := map[string]string{
		"proposal": "proposed",
		"prepare":  "prepared",
		"commit":   "committed",
		"timeout":  "view_change",
		"other":    "processed",
	}
	for input, want := range cases {
		got := sm.determineFinalStateStatus(input, true)
		if got != want {
			t.Errorf("determineFinalStateStatus(%q, true) = %q, want %q", input, got, want)
		}
	}
}

// ─── GetCurrentState / GetStateAtHeight ──────────────────────────────────────

func TestSprint19_GetCurrentState_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	_ = sm.GetCurrentState() // should not panic
}

func TestSprint19_GetStateAtHeight_Unknown_Error(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	_, err := sm.GetStateAtHeight(999)
	if err == nil {
		t.Error("expected error for unknown height, got nil")
	}
}

// ─── Start / Stop ─────────────────────────────────────────────────────────────

func TestSprint19_StartStop_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	if err := sm.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}
	if err := sm.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

// ─── SyncFinalStatesNow / DebugFinalStates ────────────────────────────────────

func TestSprint19_SyncFinalStatesNow_NilConsensus_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	sm.SyncFinalStatesNow()
}

func TestSprint19_DebugFinalStates_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	sm.DebugFinalStates()
}

// ─── PopulatedFinalStates ─────────────────────────────────────────────────────

func TestSprint19_PopulatedFinalStates_Nil(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	got := sm.PopulatedFinalStates(nil)
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestSprint19_PopulatedFinalStates_EmptySlice(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	got := sm.PopulatedFinalStates([]*FinalStateInfo{})
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

// ─── ValidateAndFixFinalStates ───────────────────────────────────────────────

func TestSprint19_ValidateAndFix_EmptyFinalStates_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	_ = sm.ValidateAndFixFinalStates()
}

// ─── RepopulateFinalStates ────────────────────────────────────────────────────

func TestSprint19_Repopulate_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	sm.RepopulateFinalStates()
}

// ─── HandleCommitProof (nil safety) ──────────────────────────────────────────

func TestSprint19_HandleCommitProof_ValidProof_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	proof := &CommitProof{
		Height:     1,
		BlockHash:  "abc",
		Signatures: map[string][]byte{},
		View:       0,
		Quorum:     1,
	}
	_ = sm.HandleCommitProof(proof) // error or nil, just no panic
}

// ─── VerifyState ─────────────────────────────────────────────────────────────

func TestSprint19_VerifyState_ValidSnapshot_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM19(t, "n", nil)
	defer cleanup()
	snap := &StateSnapshot{
		Height:    1,
		BlockHash: "abc",
		StateRoot: "root",
	}
	_, _ = sm.VerifyState(snap)
}
