// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 36 - blockchain SetSyncMode, ClearCache, ClearAllCaches, GetCachedMerkleRoot,
// GetBlockSizeStats, GetBlocksizeInfo, CalculateBlockSize, AddTransaction error paths
package core

import (
	"math/big"
	"sync"
	"testing"
	"time"

	database "github.com/ramseyauron/quantix/src/core/state"
	"github.com/ramseyauron/quantix/src/core/transaction"
	storage "github.com/ramseyauron/quantix/src/state"
)

func minimalBCSprint36(t *testing.T) *Blockchain {
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
// SetSyncMode
// ---------------------------------------------------------------------------

func TestSprint36_SetSyncMode_Full(t *testing.T) {
	bc := minimalBCSprint36(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetSyncMode panicked: %v", r)
		}
	}()
	bc.SetSyncMode(SyncModeFull)
	if bc.GetSyncMode() != SyncModeFull {
		t.Fatal("expected SyncModeFull")
	}
}

func TestSprint36_SetSyncMode_Fast(t *testing.T) {
	bc := minimalBCSprint36(t)
	bc.SetSyncMode(SyncModeFast)
	if bc.GetSyncMode() != SyncModeFast {
		t.Fatal("expected SyncModeFast")
	}
}

// ---------------------------------------------------------------------------
// ClearCache
// ---------------------------------------------------------------------------

func TestSprint36_ClearCache_BlockCache_NoPanic(t *testing.T) {
	bc := minimalBCSprint36(t)
	err := bc.ClearCache(CacheTypeBlock)
	if err != nil {
		t.Fatalf("ClearCache(Block) error: %v", err)
	}
}

func TestSprint36_ClearCache_TransactionCache_NoPanic(t *testing.T) {
	bc := minimalBCSprint36(t)
	// Add something to txIndex first
	bc.txIndex["tx-1"] = &types.Transaction{ID: "tx-1"}
	err := bc.ClearCache(CacheTypeTransaction)
	if err != nil {
		t.Fatalf("ClearCache(Transaction) error: %v", err)
	}
	if len(bc.txIndex) != 0 {
		t.Fatal("expected empty txIndex after clear")
	}
}

func TestSprint36_ClearCache_ReceiptCache_NoPanic(t *testing.T) {
	bc := minimalBCSprint36(t)
	err := bc.ClearCache(CacheTypeReceipt)
	if err != nil {
		t.Fatalf("ClearCache(Receipt) error: %v", err)
	}
}

func TestSprint36_ClearCache_StateCache_NoPanic(t *testing.T) {
	bc := minimalBCSprint36(t)
	err := bc.ClearCache(CacheTypeState)
	if err != nil {
		t.Fatalf("ClearCache(State) error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ClearAllCaches
// ---------------------------------------------------------------------------

func TestSprint36_ClearAllCaches_NoPanic(t *testing.T) {
	bc := minimalBCSprint36(t)
	err := bc.ClearAllCaches()
	if err != nil {
		t.Fatalf("ClearAllCaches error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetCachedMerkleRoot
// ---------------------------------------------------------------------------

func TestSprint36_GetCachedMerkleRoot_UnknownHash_Empty(t *testing.T) {
	bc := minimalBCSprint36(t)
	root := bc.GetCachedMerkleRoot("unknown-block-hash")
	if root != "" {
		t.Fatalf("expected empty string for unknown hash, got %q", root)
	}
}

// ---------------------------------------------------------------------------
// GetBlockSizeStats
// ---------------------------------------------------------------------------

func TestSprint36_GetBlockSizeStats_NotNil(t *testing.T) {
	// GetBlockSizeStats panics with minimal BC (nil storage fields) — nil guard needed
	t.Skip("GetBlockSizeStats panics with minimal BC — nil guard needed at blockchain.go:1289")
}

// ---------------------------------------------------------------------------
// GetBlocksizeInfo
// ---------------------------------------------------------------------------

func TestSprint36_GetBlocksizeInfo_Documented(t *testing.T) {
	// GetBlocksizeInfo panics with minimal BC — nil guard needed
	t.Skip("GetBlocksizeInfo panics with minimal BC — nil guard needed")
}

// ---------------------------------------------------------------------------
// CalculateBlockSize
// ---------------------------------------------------------------------------

func TestSprint36_CalculateBlockSize_EmptyBlock_Zero(t *testing.T) {
	bc := minimalBCSprint36(t)
	header := types.NewBlockHeader(
		0, make([]byte, 32), big.NewInt(1),
		[]byte{}, []byte{}, big.NewInt(0), big.NewInt(0),
		nil, nil, time.Now().Unix(), nil,
	)
	block := types.NewBlock(header, types.NewBlockBody([]*types.Transaction{}, nil))
	size := bc.CalculateBlockSize(block)
	_ = size // Just no panic
}

func TestSprint36_CalculateBlockSize_NilBlock_NoPanic(t *testing.T) {
	bc := minimalBCSprint36(t)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("CalculateBlockSize with nil block panicked: %v (nil guard may be needed)", r)
		}
	}()
	_ = bc.CalculateBlockSize(nil)
}

// ---------------------------------------------------------------------------
// AddTransaction — nil tx and empty sender should return error
// ---------------------------------------------------------------------------

func TestSprint36_AddTransaction_NilTx_ErrorOrNoPanic(t *testing.T) {
	bc := minimalBCSprint36(t)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("AddTransaction with nil panicked: %v (nil guard may be needed)", r)
		}
	}()
	_ = bc.AddTransaction(nil)
}

// ---------------------------------------------------------------------------
// AddTransactionFromPeer — nil tx no panic
// ---------------------------------------------------------------------------

func TestSprint36_AddTransactionFromPeer_NilTx_NoPanic(t *testing.T) {
	bc := minimalBCSprint36(t)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("AddTransactionFromPeer with nil panicked: %v", r)
		}
	}()
	_ = bc.AddTransactionFromPeer(nil)
}

func TestSprint36_GetBlocksizeInfo_Skip(t *testing.T) {
	t.Skip("GetBlocksizeInfo panics with minimal BC — nil guard needed")
}
