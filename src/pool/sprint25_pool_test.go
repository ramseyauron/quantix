// Sprint 25b — pool: RetryFailedTransactions, CalculateTransactionSize
package pool

import (
	"math/big"
	"testing"

	types "github.com/ramseyauron/quantix/src/core/transaction"
)

func makeTx25(id, sender, receiver string, amount int64, nonce uint64) *types.Transaction {
	return &types.Transaction{
		ID:       id,
		Sender:   sender,
		Receiver: receiver,
		Amount:   big.NewInt(amount),
		Nonce:    nonce,
		GasPrice: big.NewInt(1),
		GasLimit: big.NewInt(21000),
	}
}

// ─── RetryFailedTransactions ─────────────────────────────────────────────────

func TestSprint25_RetryFailedTransactions_EmptyPool_ReturnsZero(t *testing.T) {
	mp := NewMempool(nil)
	n := mp.RetryFailedTransactions(3)
	if n != 0 {
		t.Errorf("expected 0 retried from empty pool, got %d", n)
	}
}

func TestSprint25_RetryFailedTransactions_NoInvalidTxs_ReturnsZero(t *testing.T) {
	mp := NewMempool(nil)
	tx := makeTx25("tx1", "alice", "bob", 100, 0)
	mp.BroadcastTransaction(tx)
	// tx is in broadcastPool (valid), not invalidPool
	n := mp.RetryFailedTransactions(3)
	if n != 0 {
		t.Errorf("expected 0 retried (no invalid txs), got %d", n)
	}
}

func TestSprint25_RetryFailedTransactions_MaxRetries_Zero_ReturnsZero(t *testing.T) {
	mp := NewMempool(nil)
	n := mp.RetryFailedTransactions(0)
	if n != 0 {
		t.Errorf("expected 0 for maxRetries=0, got %d", n)
	}
}

// ─── CalculateTransactionSize ────────────────────────────────────────────────

func TestSprint25_CalculateTransactionSize_MinimalTx_Positive(t *testing.T) {
	mp := NewMempool(nil)
	tx := makeTx25("tx-size", "alice", "bob", 100, 1)
	size := mp.CalculateTransactionSize(tx)
	if size == 0 {
		t.Error("expected non-zero size for non-empty tx")
	}
}

func TestSprint25_CalculateTransactionSize_LargerTx_LargerSize(t *testing.T) {
	mp := NewMempool(nil)

	small := makeTx25("s", "a", "b", 1, 0)
	large := makeTx25("large-id-with-more-bytes", "alice-with-long-name", "bob-recipient", 999999, 99)
	large.SigTimestamp = []byte("timestamp-bytes-for-large-tx")

	sSmall := mp.CalculateTransactionSize(small)
	sLarge := mp.CalculateTransactionSize(large)
	if sLarge <= sSmall {
		t.Errorf("expected larger tx to have larger size: small=%d large=%d", sSmall, sLarge)
	}
}

func TestSprint25_CalculateTransactionSize_NilTx_Documented(t *testing.T) {
	// CalculateTransactionSize panics on nil tx (no nil guard).
	// This test documents the behavior — callers must not pass nil.
	t.Log("CalculateTransactionSize(nil) panics — no nil guard (documented gap)")
}
