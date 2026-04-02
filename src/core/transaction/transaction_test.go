// MIT License
//
// Copyright (c) 2024 quantix
//
// go/src/core/transaction/transaction_test.go
package types

import (
	"bytes"
	"math/big"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newValidTx(nonce uint64) *Transaction {
	return &Transaction{
		ID:        "tx-" + string(rune('A'+nonce)),
		Sender:    "xSenderAddress1234567890ABCDEF",
		Receiver:  "xReceiverAddress1234567890ABC",
		Amount:    big.NewInt(100),
		GasLimit:  big.NewInt(21000),
		GasPrice:  big.NewInt(1),
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
	}
}

func newValidBlock(height uint64, parentHash []byte) *Block {
	header := NewBlockHeader(
		height,
		parentHash,
		big.NewInt(1),
		[]byte{},
		[]byte{},
		big.NewInt(8000000),
		big.NewInt(0),
		nil,
		nil,
		time.Now().Unix(),
		nil,
	)
	body := NewBlockBody(nil, nil)
	b := NewBlock(header, body)
	b.FinalizeHash()
	return b
}

// ---------------------------------------------------------------------------
// Q2-A: Block creation and hash stability
// ---------------------------------------------------------------------------

func TestBlockCreation(t *testing.T) {
	t.Run("genesis block has height 0", func(t *testing.T) {
		b := newValidBlock(0, nil)
		if b.Header.Height != 0 {
			t.Errorf("expected height 0, got %d", b.Header.Height)
		}
	})

	t.Run("regular block has correct height", func(t *testing.T) {
		parent := newValidBlock(0, nil)
		parentHashBytes := []byte(parent.GetHash())
		b := newValidBlock(1, parentHashBytes)
		if b.Header.Height != 1 {
			t.Errorf("expected height 1, got %d", b.Header.Height)
		}
	})

	t.Run("block hash is non-empty after finalize", func(t *testing.T) {
		b := newValidBlock(1, make([]byte, 32))
		if len(b.GetHash()) == 0 {
			t.Error("expected non-empty hash after FinalizeHash")
		}
	})
}

func TestBlockHashStability(t *testing.T) {
	tests := []struct {
		name   string
		height uint64
	}{
		{"genesis stability", 0},
		{"height-1 stability", 1},
		{"height-100 stability", 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parentHash := make([]byte, 32)
			ts := time.Now().Unix()

			header1 := NewBlockHeader(tc.height, parentHash, big.NewInt(1), []byte{}, []byte{}, big.NewInt(8e6), big.NewInt(0), nil, nil, ts, nil)
			b1 := NewBlock(header1, NewBlockBody(nil, nil))
			b1.FinalizeHash()

			header2 := NewBlockHeader(tc.height, parentHash, big.NewInt(1), []byte{}, []byte{}, big.NewInt(8e6), big.NewInt(0), nil, nil, ts, nil)
			b2 := NewBlock(header2, NewBlockBody(nil, nil))
			b2.FinalizeHash()

			if b1.GetHash() != b2.GetHash() {
				t.Errorf("same inputs produced different hashes:\n  h1=%s\n  h2=%s", b1.GetHash(), b2.GetHash())
			}
		})
	}
}

func TestBlockHashChangesWithDifferentInputs(t *testing.T) {
	ts := time.Now().Unix()
	parentHash := make([]byte, 32)

	header1 := NewBlockHeader(1, parentHash, big.NewInt(1), []byte{}, []byte{}, big.NewInt(8e6), big.NewInt(0), nil, nil, ts, nil)
	b1 := NewBlock(header1, NewBlockBody(nil, nil))
	b1.FinalizeHash()

	// Different difficulty
	header2 := NewBlockHeader(1, parentHash, big.NewInt(2), []byte{}, []byte{}, big.NewInt(8e6), big.NewInt(0), nil, nil, ts, nil)
	b2 := NewBlock(header2, NewBlockBody(nil, nil))
	b2.FinalizeHash()

	if b1.GetHash() == b2.GetHash() {
		t.Error("blocks with different difficulties produced the same hash")
	}
}

// ---------------------------------------------------------------------------
// Q2-B: Transaction validation (SanityCheck)
// ---------------------------------------------------------------------------

func TestTransactionSanityCheck(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Transaction)
		wantErr bool
	}{
		{
			name:    "valid transaction",
			mutate:  func(tx *Transaction) {},
			wantErr: false,
		},
		{
			name: "zero amount",
			mutate: func(tx *Transaction) {
				tx.Amount = big.NewInt(0)
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			mutate: func(tx *Transaction) {
				tx.Amount = big.NewInt(-1)
			},
			wantErr: true,
		},
		{
			name: "missing sender",
			mutate: func(tx *Transaction) {
				tx.Sender = ""
			},
			wantErr: true,
		},
		{
			name: "missing receiver",
			mutate: func(tx *Transaction) {
				tx.Receiver = ""
			},
			wantErr: true,
		},
		{
			name: "negative gas limit",
			mutate: func(tx *Transaction) {
				tx.GasLimit = big.NewInt(-1)
			},
			wantErr: true,
		},
		{
			name: "negative gas price",
			mutate: func(tx *Transaction) {
				tx.GasPrice = big.NewInt(-1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := newValidTx(0)
			tc.mutate(tx)
			err := tx.SanityCheck()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Q2-C: Merkle root consistency
// ---------------------------------------------------------------------------

func TestMerkleRootConsistency(t *testing.T) {
	t.Run("same txs same order yields same root", func(t *testing.T) {
		txs := []*Transaction{newValidTx(0), newValidTx(1), newValidTx(2)}

		root1 := CalculateMerkleRoot(txs)
		root2 := CalculateMerkleRoot(txs)

		if !bytes.Equal(root1, root2) {
			t.Error("same transactions in same order produced different Merkle roots")
		}
	})

	t.Run("empty tx list yields stable root", func(t *testing.T) {
		root1 := CalculateMerkleRoot(nil)
		root2 := CalculateMerkleRoot(nil)
		if !bytes.Equal(root1, root2) {
			t.Error("empty tx list produced different roots across calls")
		}
	})

	t.Run("different order yields different root", func(t *testing.T) {
		tx0 := newValidTx(0)
		tx1 := newValidTx(1)

		root1 := CalculateMerkleRoot([]*Transaction{tx0, tx1})
		root2 := CalculateMerkleRoot([]*Transaction{tx1, tx0})

		if bytes.Equal(root1, root2) {
			t.Error("different tx order should produce different Merkle roots")
		}
	})

	t.Run("adding tx changes root", func(t *testing.T) {
		txs1 := []*Transaction{newValidTx(0)}
		txs2 := []*Transaction{newValidTx(0), newValidTx(1)}

		root1 := CalculateMerkleRoot(txs1)
		root2 := CalculateMerkleRoot(txs2)

		if bytes.Equal(root1, root2) {
			t.Error("adding a transaction should change the Merkle root")
		}
	})

	t.Run("block TxsRoot reflects transactions", func(t *testing.T) {
		tx := newValidTx(0)
		b := newValidBlock(1, make([]byte, 32))
		b.AddTxs(tx)
		b.FinalizeHash()

		// After FinalizeHash, TxsRoot must equal CalculateMerkleRoot of the txs
		expectedRoot := CalculateMerkleRoot([]*Transaction{tx})
		if !bytes.Equal(b.Header.TxsRoot, expectedRoot) {
			t.Errorf("TxsRoot mismatch: got %x, want %x", b.Header.TxsRoot, expectedRoot)
		}
	})
}

// ---------------------------------------------------------------------------
// Q2-D: Block-level ValidateTxsRoot
// ---------------------------------------------------------------------------

func TestValidateTxsRoot(t *testing.T) {
	t.Run("valid block passes", func(t *testing.T) {
		b := newValidBlock(1, make([]byte, 32))
		b.AddTxs(newValidTx(0))
		b.FinalizeHash()

		if err := b.ValidateTxsRoot(); err != nil {
			t.Errorf("expected valid TxsRoot, got: %v", err)
		}
	})

	t.Run("tampered TxsRoot detected", func(t *testing.T) {
		b := newValidBlock(1, make([]byte, 32))
		b.AddTxs(newValidTx(0))
		b.FinalizeHash()

		// Corrupt TxsRoot
		b.Header.TxsRoot = []byte("tampered_root_0000000000000000")
		if err := b.ValidateTxsRoot(); err == nil {
			t.Error("expected error for tampered TxsRoot, got nil")
		}
	})
}
