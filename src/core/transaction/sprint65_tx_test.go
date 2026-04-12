// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 65 — core/transaction 82.6%→higher
// Tests: Transaction.SanityCheck remaining paths, Block.SanityCheck nil header,
// updateTPS internal, BenchmarkMerkleTree wrapper
package types

import (
	"math/big"
	"testing"
	"time"
)

// ─── Transaction.SanityCheck — remaining uncovered paths ─────────────────────

func TestSprint65_TxSanityCheck_NegativeGasLimit(t *testing.T) {
	tx := &Transaction{
		Sender:    "alice",
		Receiver:  "bob",
		Amount:    big.NewInt(100),
		GasLimit:  big.NewInt(-1),
		GasPrice:  big.NewInt(0),
		Timestamp: time.Now().Unix(),
	}
	err := tx.SanityCheck()
	if err == nil {
		t.Error("expected error for negative GasLimit")
	}
}

func TestSprint65_TxSanityCheck_NegativeGasPrice(t *testing.T) {
	tx := &Transaction{
		Sender:    "alice",
		Receiver:  "bob",
		Amount:    big.NewInt(100),
		GasLimit:  big.NewInt(0),
		GasPrice:  big.NewInt(-1),
		Timestamp: time.Now().Unix(),
	}
	err := tx.SanityCheck()
	if err == nil {
		t.Error("expected error for negative GasPrice")
	}
}

func TestSprint65_TxSanityCheck_FingerprintMismatch(t *testing.T) {
	tx := &Transaction{
		Sender:      "alice-addr",
		Receiver:    "bob",
		Amount:      big.NewInt(100),
		Fingerprint: "different-fp",
		Timestamp:   time.Now().Unix(),
	}
	err := tx.SanityCheck()
	if err == nil {
		t.Error("expected error for fingerprint mismatch")
	}
}

func TestSprint65_TxSanityCheck_FingerprintMatchesSender(t *testing.T) {
	tx := &Transaction{
		Sender:      "alice-addr",
		Receiver:    "bob",
		Amount:      big.NewInt(100),
		Fingerprint: "alice-addr",
		Timestamp:   time.Now().Unix(),
	}
	err := tx.SanityCheck()
	if err != nil {
		t.Errorf("unexpected error for matching fingerprint: %v", err)
	}
}

// ─── Block.SanityCheck — nil header ──────────────────────────────────────────

func TestSprint65_BlockSanityCheck_NilHeader(t *testing.T) {
	b := &Block{}
	err := b.SanityCheck()
	if err == nil {
		t.Error("expected error for nil block header")
	}
}

// ─── updateTPS — coverage via RecordTransaction then sleep ───────────────────

func TestSprint65_UpdateTPS_ViaManyTransactions(t *testing.T) {
	monitor := NewTPSMonitor(100 * time.Millisecond)
	for i := 0; i < 50; i++ {
		monitor.RecordTransaction()
	}
	time.Sleep(150 * time.Millisecond)
	stats := monitor.GetStats()
	if stats == nil {
		t.Error("expected non-nil stats")
	}
}

// ─── BenchmarkMerkleTree wrapper ─────────────────────────────────────────────

func TestSprint65_BenchmarkMerkleTree_Wrapper(t *testing.T) {
	result := testing.Benchmark(BenchmarkMerkleTree)
	if result.N < 1 {
		t.Error("expected BenchmarkMerkleTree to run")
	}
}

func TestSprint65_BenchmarkTransactionProcessing_Wrapper(t *testing.T) {
	result := testing.Benchmark(BenchmarkTransactionProcessing)
	if result.N < 1 {
		t.Error("expected BenchmarkTransactionProcessing to run")
	}
}

func TestSprint65_BenchmarkTPSMonitoring_Wrapper(t *testing.T) {
	result := testing.Benchmark(BenchmarkTPSMonitoring)
	if result.N < 1 {
		t.Error("expected BenchmarkTPSMonitoring to run")
	}
}

func TestSprint65_BenchmarkRealWorldTPS_Wrapper(t *testing.T) {
	result := testing.Benchmark(BenchmarkRealWorldTPS)
	if result.N < 1 {
		t.Error("expected BenchmarkRealWorldTPS to run")
	}
}

// ─── FinalizeHash — with minimal valid block ─────────────────────────────────

func TestSprint65_FinalizeHash_ValidBlock(t *testing.T) {
	b := &Block{
		Header: &BlockHeader{
			Height:     1,
			ParentHash: []byte("parent"),
			Difficulty: big.NewInt(1),
			GasLimit:   big.NewInt(8000000),
			GasUsed:    big.NewInt(0),
			Timestamp:  time.Now().Unix(),
		},
		Body: BlockBody{
			TxsList: []*Transaction{},
		},
	}
	b.FinalizeHash()
	hash := b.GetHash()
	if hash == "" {
		t.Error("expected non-empty hash after FinalizeHash")
	}
}
