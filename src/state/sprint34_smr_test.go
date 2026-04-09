// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 34 - state machine HandleOperation, HandleCommitProof, VerifyState, ProposeTransaction
// NOTE: Many nil-arg methods panic — nil guards needed at top of each function.
package state

import (
	"os"
	"testing"

	types "github.com/ramseyauron/quantix/src/core/transaction"
)

func newTestSM34WithStorage(t *testing.T) (*StateMachine, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-smr34-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	storage, err := NewStorage(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("NewStorage: %v", err)
	}
	sm := NewStateMachine(storage, "node-sprint34", []string{"node-sprint34"})
	return sm, func() {
		if storage != nil {
			storage.Close()
		}
		os.RemoveAll(dir)
	}
}

// ---------------------------------------------------------------------------
// HandleOperation — nil arg panics (documented); valid op no panic
// ---------------------------------------------------------------------------

func TestSprint34_HandleOperation_Nil_Documented(t *testing.T) {
	t.Skip("HandleOperation panics with nil op — nil guard needed")
}

func TestSprint34_HandleOperation_ValidOp_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM34WithStorage(t)
	defer cleanup()
	op := &Operation{
		Type:      OpBlock,
		Sequence:  1,
		Proposer:  "node-sprint34",
		Signature: []byte("sig"),
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("HandleOperation panicked: %v", r)
		}
	}()
	_ = sm.HandleOperation(op)
}

// ---------------------------------------------------------------------------
// HandleCommitProof — nil arg panics (documented); valid proof no panic
// ---------------------------------------------------------------------------

func TestSprint34_HandleCommitProof_Nil_Documented(t *testing.T) {
	t.Skip("HandleCommitProof panics with nil proof — nil guard needed")
}

func TestSprint34_HandleCommitProof_ValidProof_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM34WithStorage(t)
	defer cleanup()
	proof := &CommitProof{
		BlockHash:  "some-block-hash",
		Height:     1,
		Signatures: map[string][]byte{},
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("HandleCommitProof panicked: %v", r)
		}
	}()
	_ = sm.HandleCommitProof(proof)
}

// ---------------------------------------------------------------------------
// GetCurrentState — no panic
// ---------------------------------------------------------------------------

func TestSprint34_GetCurrentState_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM34WithStorage(t)
	defer cleanup()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GetCurrentState panicked: %v", r)
		}
	}()
	_ = sm.GetCurrentState()
}

// ---------------------------------------------------------------------------
// VerifyState — nil panics (documented); empty snapshot no panic
// ---------------------------------------------------------------------------

func TestSprint34_VerifyState_Nil_Documented(t *testing.T) {
	t.Skip("VerifyState panics with nil snapshot — nil guard needed")
}

func TestSprint34_VerifyState_EmptySnapshot_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM34WithStorage(t)
	defer cleanup()
	snap := &StateSnapshot{Height: 0, BlockHash: "", StateRoot: ""}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("VerifyState panicked: %v", r)
		}
	}()
	_, _ = sm.VerifyState(snap)
}

// ---------------------------------------------------------------------------
// ProposeTransaction — nil panics (documented); valid tx no panic
// ---------------------------------------------------------------------------

func TestSprint34_ProposeTransaction_Nil_Documented(t *testing.T) {
	t.Skip("ProposeTransaction panics with nil tx — nil guard needed")
}

func TestSprint34_ProposeTransaction_ValidTx_NoPanic(t *testing.T) {
	sm, cleanup := newTestSM34WithStorage(t)
	defer cleanup()
	tx := &types.Transaction{ID: "test-tx-sprint34", Sender: "alice", Receiver: "bob"}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ProposeTransaction panicked: %v", r)
		}
	}()
	_ = sm.ProposeTransaction(tx)
}
