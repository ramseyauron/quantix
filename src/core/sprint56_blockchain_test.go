// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 56 — src/core blockchain 56.9%→higher
// Tests: ImportBlock non-running, checkConsensusRequirements, getRecentBlocks,
// GenerateTransactionProof, VerifyTransactionInBlock, ClearAllCaches, GetCachedMerkleRoot
package core

import (
	"sync"
	"testing"

	database "github.com/ramseyauron/quantix/src/core/state"
	storage "github.com/ramseyauron/quantix/src/state"
	types "github.com/ramseyauron/quantix/src/core/transaction"
)

func minimalBCSprint56(t *testing.T) *Blockchain {
	t.Helper()
	db, err := database.NewLevelDB(t.TempDir())
	if err != nil {
		t.Fatalf("NewLevelDB: %v", err)
	}
	dir := t.TempDir()
	store, err := storage.NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	store.SetDB(db)
	bc := &Blockchain{
		storage:     store,
		chain:       []*types.Block{},
		txIndex:     make(map[string]*types.Transaction),
		lock:        sync.RWMutex{},
		chainParams: GetDevnetChainParams(),
	}
	t.Cleanup(func() {
		_ = store.Close()
		_ = db.Close()
	})
	return bc
}

// ─── ImportBlock — non-running status → ImportError ──────────────────────────

func TestSprint56_ImportBlock_NotRunning(t *testing.T) {
	bc := minimalBCSprint56(t)
	bc.SetStatus(StatusStopped)
	result := bc.ImportBlock(&types.Block{})
	if result != ImportError {
		t.Errorf("expected ImportError for non-running blockchain, got %v", result)
	}
}

func TestSprint56_ImportBlock_InvalidBlock(t *testing.T) {
	bc := minimalBCSprint56(t)
	bc.SetStatus(StatusRunning)
	// Empty block fails Validate() → ImportInvalid
	result := bc.ImportBlock(&types.Block{})
	if result == ImportedBest {
		t.Error("expected non-ImportedBest for invalid block")
	}
}

// ─── checkConsensusRequirements ──────────────────────────────────────────────

func TestSprint56_CheckConsensusRequirements_NilBlock(t *testing.T) {
	bc := minimalBCSprint56(t)
	result := bc.checkConsensusRequirements(nil)
	if result {
		t.Error("expected false for nil block")
	}
}

func TestSprint56_CheckConsensusRequirements_EmptyHash(t *testing.T) {
	bc := minimalBCSprint56(t)
	b := &types.Block{}
	result := bc.checkConsensusRequirements(b)
	if result {
		t.Error("expected false for block with empty/nil hash")
	}
}

func TestSprint56_CheckConsensusRequirements_WithHash(t *testing.T) {
	bc := minimalBCSprint56(t)
	b := &types.Block{
		Header: &types.BlockHeader{Hash: []byte("0xdeadbeef")},
	}
	result := bc.checkConsensusRequirements(b)
	if !result {
		t.Error("expected true for block with non-empty hash")
	}
}

// ─── getRecentBlocks — empty chain ───────────────────────────────────────────

func TestSprint56_GetRecentBlocks_EmptyChain(t *testing.T) {
	bc := minimalBCSprint56(t)
	blocks := bc.getRecentBlocks(10)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty chain, got %d", len(blocks))
	}
}

func TestSprint56_GetRecentBlocks_ZeroCount(t *testing.T) {
	bc := minimalBCSprint56(t)
	blocks := bc.getRecentBlocks(0)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(blocks))
	}
}

// ─── GenerateTransactionProof — unknown block hash ───────────────────────────

func TestSprint56_GenerateTransactionProof_UnknownHash(t *testing.T) {
	bc := minimalBCSprint56(t)
	_, err := bc.GenerateTransactionProof(&types.Transaction{ID: "tx1"}, "unknown-hash")
	if err == nil {
		t.Error("expected error for unknown block hash")
	}
}

// ─── VerifyTransactionInBlock — unknown block hash → false + error ───────────

func TestSprint56_VerifyTransactionInBlock_UnknownHash(t *testing.T) {
	bc := minimalBCSprint56(t)
	ok, err := bc.VerifyTransactionInBlock(&types.Transaction{ID: "tx1"}, "unknown-hash")
	if ok {
		t.Error("expected false for unknown block hash")
	}
	if err == nil {
		t.Error("expected error for unknown block hash")
	}
}

// ─── ClearAllCaches — no panic on minimal BC ─────────────────────────────────

func TestSprint56_ClearAllCaches_NoPanic(t *testing.T) {
	bc := minimalBCSprint56(t)
	bc.ClearAllCaches()
}

// ─── GetCachedMerkleRoot — unknown hash → empty string ───────────────────────

func TestSprint56_GetCachedMerkleRoot_UnknownHash(t *testing.T) {
	bc := minimalBCSprint56(t)
	result := bc.GetCachedMerkleRoot("unknown-hash-xyz")
	if result != "" {
		t.Errorf("expected empty string for unknown hash, got %q", result)
	}
}

// ─── ImportResultString (method, coverage for switch cases) ─────────────────

func TestSprint56_ImportResultString_Existing(t *testing.T) {
	bc := minimalBCSprint56(t)
	// Cover the ImportedExisting case not covered by Sprint 11
	got := bc.ImportResultString(ImportedExisting)
	if got == "" {
		t.Error("ImportResultString(ImportedExisting) should be non-empty")
	}
}
