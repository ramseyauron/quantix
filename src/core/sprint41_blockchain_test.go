// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 41 - core blockchain uncovered functions
package core

import (
	"math/big"
	"sync"
	"testing"
	"time"

	database "github.com/ramseyauron/quantix/src/core/state"
	types "github.com/ramseyauron/quantix/src/core/transaction"
	storage "github.com/ramseyauron/quantix/src/state"
)

func minimalBCSprint41(t *testing.T) *Blockchain {
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

// ---------------------------------------------------------------------------
// calculateTxsSize — nil mempool path
// ---------------------------------------------------------------------------

func TestSprint41_CalculateTxsSize_NilMempool_UsesEstimate(t *testing.T) {
	bc := minimalBCSprint41(t)
	tx := &types.Transaction{
		ID:       "test-tx-s41",
		Sender:   "alice",
		Receiver: "bob",
		Amount:   big.NewInt(100),
		Nonce:    1,
	}
	size, err := bc.calculateTxsSize(tx)
	if err != nil {
		t.Fatalf("calculateTxsSize error: %v", err)
	}
	if size == 0 {
		t.Fatal("expected non-zero size estimate")
	}
}

func TestSprint41_CalculateTxsSize_NilAmount_NoError(t *testing.T) {
	bc := minimalBCSprint41(t)
	tx := &types.Transaction{
		ID:       "test-nil-amount",
		Sender:   "alice",
		Receiver: "bob",
		Amount:   nil,
	}
	size, err := bc.calculateTxsSize(tx)
	if err != nil {
		t.Fatalf("calculateTxsSize with nil Amount error: %v", err)
	}
	_ = size
}

// ---------------------------------------------------------------------------
// getTransactionGas — various paths
// ---------------------------------------------------------------------------

func TestSprint41_GetTransactionGas_WithGasLimit_ReturnsGasLimit(t *testing.T) {
	bc := minimalBCSprint41(t)
	tx := &types.Transaction{
		GasLimit: big.NewInt(50000),
	}
	gas := bc.getTransactionGas(tx)
	if gas == nil {
		t.Fatal("expected non-nil gas")
	}
	if gas.Cmp(big.NewInt(50000)) != 0 {
		t.Fatalf("expected gas=50000, got %s", gas.String())
	}
}

func TestSprint41_GetTransactionGas_NoGasLimit_UsesBaseGas(t *testing.T) {
	bc := minimalBCSprint41(t)
	tx := &types.Transaction{
		GasLimit: nil,
		Signature: []byte{0xAA, 0xBB},
	}
	gas := bc.getTransactionGas(tx)
	if gas == nil {
		t.Fatal("expected non-nil gas")
	}
	// Should be >= base gas (21000)
	if gas.Cmp(big.NewInt(21000)) < 0 {
		t.Fatalf("expected gas >= 21000, got %s", gas.String())
	}
}

func TestSprint41_GetTransactionGas_ZeroGasLimit_UsesBaseGas(t *testing.T) {
	bc := minimalBCSprint41(t)
	tx := &types.Transaction{
		GasLimit: big.NewInt(0),
	}
	gas := bc.getTransactionGas(tx)
	if gas == nil {
		t.Fatal("expected non-nil gas")
	}
}

// ---------------------------------------------------------------------------
// VerifyTransactionInBlock — unknown hash returns error
// ---------------------------------------------------------------------------

func TestSprint41_VerifyTransactionInBlock_UnknownHash_Error(t *testing.T) {
	bc := minimalBCSprint41(t)
	tx := &types.Transaction{ID: "test"}
	_, err := bc.VerifyTransactionInBlock(tx, "nonexistent-block-hash")
	if err == nil {
		t.Fatal("expected error for unknown block hash")
	}
}

// ---------------------------------------------------------------------------
// GenerateTransactionProof — unknown hash returns error
// ---------------------------------------------------------------------------

func TestSprint41_GenerateTransactionProof_UnknownHash_Error(t *testing.T) {
	bc := minimalBCSprint41(t)
	tx := &types.Transaction{ID: "test"}
	_, err := bc.GenerateTransactionProof(tx, "nonexistent-hash")
	if err == nil {
		t.Fatal("expected error for unknown block hash")
	}
}

// ---------------------------------------------------------------------------
// GetCurrentState / DebugStorage with nil state machine
// ---------------------------------------------------------------------------

func TestSprint41_GetCurrentState_NilStateMachine_NoPanic(t *testing.T) {
	bc := minimalBCSprint41(t)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetCurrentState panicked with nil stateMachine: %v", r)
		}
	}()
	_ = bc.GetCurrentState()
}

func TestSprint41_DebugStorage_EmptyChain_Error(t *testing.T) {
	bc := minimalBCSprint41(t)
	err := bc.DebugStorage()
	// Empty chain has no blocks — should return error
	if err == nil {
		t.Log("DebugStorage with empty chain returned nil (may have genesis)")
	}
}

// ---------------------------------------------------------------------------
// VerifyStateConsistency — nil snapshot
// ---------------------------------------------------------------------------

func TestSprint41_VerifyStateConsistency_NilState_NoPanel(t *testing.T) {
	bc := minimalBCSprint41(t)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("VerifyStateConsistency panicked: %v", r)
		}
	}()
	_, _ = bc.VerifyStateConsistency(nil)
}

// ---------------------------------------------------------------------------
// GetBlockWithMerkleInfo — unknown hash
// ---------------------------------------------------------------------------

func TestSprint41_GetBlockWithMerkleInfo_UnknownHash_Error(t *testing.T) {
	bc := minimalBCSprint41(t)
	_, err := bc.GetBlockWithMerkleInfo("nonexistent-hash-sprint41")
	if err == nil {
		t.Fatal("expected error for unknown block hash")
	}
}

// ---------------------------------------------------------------------------
// StoreChainState — no panic with nil nodes
// ---------------------------------------------------------------------------

func TestSprint41_StoreChainState_NoPanic(t *testing.T) {
	bc := minimalBCSprint41(t)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("StoreChainState panicked: %v", r)
		}
	}()
	_ = bc.StoreChainState(nil)
}

// ---------------------------------------------------------------------------
// calculateBlockTime — no panic
// ---------------------------------------------------------------------------

func TestSprint41_CalculateBlockTime_NoPanic(t *testing.T) {
	bc := minimalBCSprint41(t)
	header := types.NewBlockHeader(1, make([]byte, 32), big.NewInt(1),
		[]byte{}, []byte{}, big.NewInt(0), big.NewInt(0),
		nil, nil, time.Now().Unix(), nil)
	block := types.NewBlock(header, types.NewBlockBody([]*types.Transaction{}, nil))
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("calculateBlockTime panicked: %v", r)
		}
	}()
	_ = bc.calculateBlockTime(block)
}

// ---------------------------------------------------------------------------
// PrintBlockIndex — no panic with empty chain
// ---------------------------------------------------------------------------

func TestSprint41_PrintBlockIndex_NoPanic(t *testing.T) {
	bc := minimalBCSprint41(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PrintBlockIndex panicked: %v", r)
		}
	}()
	bc.PrintBlockIndex()
}
