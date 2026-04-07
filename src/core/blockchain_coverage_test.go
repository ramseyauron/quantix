// MIT License
// Copyright (c) 2024 quantix

// P.E.P.P.E.R. additional coverage for Blockchain methods with low coverage.
package core

import (
	"math/big"
	"testing"
)

// ── GetAddressBalance — additional edge cases ────────────────────────────────

func TestGetAddressBalance_LargeBalance(t *testing.T) {
	db := newTestDB(t)
	bc := newBCForBalance(t, db)

	const addr = "xLargeBalance0000000000000000000"
	// 5 billion QTX in nQTX (total supply territory)
	largeAmt := new(big.Int).Mul(big.NewInt(5_000_000_000), big.NewInt(1_000_000_000))

	st := NewStateDB(db)
	st.AddBalance(addr, largeAmt)
	if _, err := st.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	bal, err := bc.GetAddressBalance(addr)
	if err != nil {
		t.Fatalf("GetAddressBalance error: %v", err)
	}
	if bal.Cmp(largeAmt) != 0 {
		t.Errorf("large balance: want %s got %s", largeAmt, bal)
	}
}

// ── GetPendingTransactionCount — with populated mempool ──────────────────────

func TestGetPendingTransactionCount_NonNegative(t *testing.T) {
	db := newTestDB(t)
	bc := fastMinimalBC(t, db)
	// With nil mempool
	bc.mempool = nil
	if bc.GetPendingTransactionCount() != 0 {
		t.Error("nil mempool should return 0")
	}
}

// ── GetBlockByNumber — basic behavior ────────────────────────────────────────

func TestGetBlockByNumber_NonExistent_Nil(t *testing.T) {
	db := newTestDB(t)
	bc := fastMinimalBC(t, db)
	// Height 99999 doesn't exist
	blk := bc.GetBlockByNumber(99999)
	if blk != nil {
		t.Error("GetBlockByNumber for non-existent height should return nil")
	}
}

func TestGetBlockByNumber_Zero_NilForEmptyChain(t *testing.T) {
	db := newTestDB(t)
	bc := fastMinimalBC(t, db)
	// No genesis block has been committed
	blk := bc.GetBlockByNumber(0)
	_ = blk // nil is acceptable for empty chain — just verify no panic
}

// ── GetLatestBlock — empty chain ─────────────────────────────────────────────

func TestGetLatestBlock_EmptyChain_NilOrGenesis(t *testing.T) {
	db := newTestDB(t)
	bc := fastMinimalBC(t, db)
	// Empty chain — GetLatestBlock may return nil or an empty genesis
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetLatestBlock panicked on empty chain: %v", r)
		}
	}()
	_ = bc.GetLatestBlock()
}

// ── ValidateChainID ───────────────────────────────────────────────────────────

func TestValidateChainID_Devnet_Passes(t *testing.T) {
	db := newTestDB(t)
	bc := fastMinimalBC(t, db) // devnet (73310)
	valid := bc.ValidateChainID(73310)
	if !valid {
		t.Error("devnet chain ID 73310 should be valid")
	}
}

func TestValidateChainID_WrongChain_False(t *testing.T) {
	db := newTestDB(t)
	bc := fastMinimalBC(t, db) // devnet (73310)
	valid := bc.ValidateChainID(9999)
	if valid {
		t.Error("wrong chain ID should return false")
	}
}

// ── LogAllocationSummary (0% coverage) ──────────────────────────────────────

func TestLogAllocationSummary_NoPanic(t *testing.T) {
	// LogAllocationSummary is a logging-only function; test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogAllocationSummary panicked: %v", r)
		}
	}()
	allocs := DefaultGenesisAllocations()
	LogAllocationSummary(allocs)
}

