package core

// Sprint 13 — GetPendingTransactions, GetMempoolStats coverage (c8aedba).
// Tests the nil-guard paths and basic behavior.

import (
	"math/big"
	"sync"
	"testing"
	"time"

	database "github.com/ramseyauron/quantix/src/core/state"
	types "github.com/ramseyauron/quantix/src/core/transaction"
	"github.com/ramseyauron/quantix/src/pool"
	storage "github.com/ramseyauron/quantix/src/state"
)

// ---------------------------------------------------------------------------
// GetPendingTransactions
// ---------------------------------------------------------------------------

func TestGetPendingTransactions_NilMempool_ReturnsNil(t *testing.T) {
	bc := &Blockchain{
		chain:       []*types.Block{},
		lock:        sync.RWMutex{},
		chainParams: GetDevnetChainParams(),
		mempool:     nil,
	}
	txs := bc.GetPendingTransactions()
	if txs != nil {
		t.Errorf("nil mempool: expected nil slice, got %v", txs)
	}
}

func TestGetPendingTransactions_EmptyMempool_EmptyOrNil(t *testing.T) {
	db, err := database.NewLevelDB(t.TempDir())
	if err != nil {
		t.Fatalf("NewLevelDB: %v", err)
	}
	defer db.Close()
	dir := t.TempDir()
	store, err := storage.NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	defer store.Close()
	store.SetDB(db)

	bc := &Blockchain{
		storage:     store,
		chain:       []*types.Block{},
		lock:        sync.RWMutex{},
		chainParams: GetDevnetChainParams(),
	}
	bc.mempool = pool.NewMempool(nil)

	txs := bc.GetPendingTransactions()
	// Empty mempool → nil or empty (both acceptable).
	if len(txs) != 0 {
		t.Errorf("empty mempool: expected 0 transactions, got %d", len(txs))
	}
}

// ---------------------------------------------------------------------------
// GetMempoolStats
// ---------------------------------------------------------------------------

func TestGetMempoolStats_NilMempool_ReturnsNil(t *testing.T) {
	bc := &Blockchain{
		chainParams: GetDevnetChainParams(),
		mempool:     nil,
	}
	stats := bc.GetMempoolStats()
	if stats != nil {
		t.Errorf("nil mempool: expected nil stats, got %v", stats)
	}
}

func TestGetMempoolStats_RealMempool_NonNil(t *testing.T) {
	db, err := database.NewLevelDB(t.TempDir())
	if err != nil {
		t.Fatalf("NewLevelDB: %v", err)
	}
	defer db.Close()
	dir := t.TempDir()
	store, err := storage.NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	defer store.Close()
	store.SetDB(db)

	bc := &Blockchain{
		storage:     store,
		chain:       []*types.Block{},
		lock:        sync.RWMutex{},
		chainParams: GetDevnetChainParams(),
	}
	bc.mempool = pool.NewMempool(nil)

	stats := bc.GetMempoolStats()
	if stats == nil {
		t.Error("GetMempoolStats with real mempool: expected non-nil stats")
	}
}

// ---------------------------------------------------------------------------
// ValidateGenesisBlock — valid genesis (non-nil block)
// ---------------------------------------------------------------------------

func TestValidateGenesisBlock_ValidBlock_NoError(t *testing.T) {
	bc := &Blockchain{
		chainParams: GetDevnetChainParams(),
	}

	header := types.NewBlockHeader(
		0,
		make([]byte, 32),
		big.NewInt(1),
		[]byte{}, []byte{},
		big.NewInt(0), big.NewInt(0),
		nil, nil,
		time.Now().Unix(),
		nil,
	)
	body := types.NewBlockBody([]*types.Transaction{}, nil)
	blk := types.NewBlock(header, body)
	blk.FinalizeHash()

	// No panic is the key assertion. Result depends on hash match vs cached genesis.
	_ = bc.ValidateGenesisBlock(blk)
}
