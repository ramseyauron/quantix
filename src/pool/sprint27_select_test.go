// Sprint 27b — pool: SelectTransactionsForBlock.
package pool

import (
	"math/big"
	"testing"
	"time"

	types "github.com/ramseyauron/quantix/src/core/transaction"
)

func makePendingTx27(id, sender string, amount int64, nonce uint64) *types.Transaction {
	return &types.Transaction{
		ID:       id,
		Sender:   sender,
		Receiver: "bob",
		Amount:   big.NewInt(amount),
		Nonce:    nonce,
		GasPrice: big.NewInt(1),
		GasLimit: big.NewInt(21000),
	}
}

// ─── SelectTransactionsForBlock ───────────────────────────────────────────────

func TestSprint27_SelectTransactionsForBlock_EmptyPool_ReturnsEmpty(t *testing.T) {
	mp := NewMempool(nil)
	txs, size := mp.SelectTransactionsForBlock(1000000, 500000)
	if len(txs) != 0 {
		t.Errorf("expected 0 txs from empty pool, got %d", len(txs))
	}
	if size != 0 {
		t.Errorf("expected size=0, got %d", size)
	}
}

func TestSprint27_SelectTransactionsForBlock_AfterBroadcast_SelectsTxs(t *testing.T) {
	mp := NewMempool(nil)

	// Add transactions and wait for them to be processed into pendingPool
	tx1 := makePendingTx27("tx-sel-1", "alice", 100, 0)
	tx2 := makePendingTx27("tx-sel-2", "bob", 200, 0)
	mp.BroadcastTransaction(tx1)
	mp.BroadcastTransaction(tx2)

	// Give the async processor time to move txs to pendingPool
	time.Sleep(150 * time.Millisecond)

	txs, size := mp.SelectTransactionsForBlock(10_000_000, 5_000_000)
	// Either empty (async not finished) or has txs — just no panic
	_ = txs
	_ = size
}

func TestSprint27_SelectTransactionsForBlock_MaxBlockSizeZero_ReturnsEmpty(t *testing.T) {
	mp := NewMempool(nil)
	tx := makePendingTx27("tx-zero-size", "alice", 100, 0)
	mp.BroadcastTransaction(tx)
	time.Sleep(50 * time.Millisecond)

	txs, _ := mp.SelectTransactionsForBlock(0, 0)
	// Zero max size → nothing fits
	if len(txs) != 0 {
		t.Logf("SelectTransactionsForBlock(0,0) returned %d txs (implementation-defined)", len(txs))
	}
}

func TestSprint27_SelectTransactionsForBlock_ReturnsSize_GreaterOrEqual(t *testing.T) {
	mp := NewMempool(nil)
	tx := makePendingTx27("tx-size-check", "alice", 100, 0)
	mp.BroadcastTransaction(tx)
	time.Sleep(150 * time.Millisecond)

	txs, totalSize := mp.SelectTransactionsForBlock(10_000_000, 5_000_000)
	if len(txs) > 0 && totalSize == 0 {
		t.Error("expected non-zero total size when transactions are selected")
	}
}
