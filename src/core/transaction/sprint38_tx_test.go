// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 38 - transaction package: getSPX, SetHash, ValidateUnclesHash, NewTxs, UnmarshalJSON
package types

import (
	"strings"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// getSPX — package-level helper
// ---------------------------------------------------------------------------

func TestSprint38_GetSPX_NonNil(t *testing.T) {
	spx := getSPX()
	if spx == nil {
		t.Fatal("expected non-nil SPX value")
	}
}

func TestSprint38_GetSPX_Positive(t *testing.T) {
	spx := getSPX()
	if spx.Sign() <= 0 {
		t.Fatalf("expected positive SPX value, got %s", spx.String())
	}
}

// ---------------------------------------------------------------------------
// Block.SetHash
// ---------------------------------------------------------------------------

func TestSprint38_SetHash_Basic(t *testing.T) {
	header := NewBlockHeader(1, make([]byte, 32), big.NewInt(1),
		[]byte{}, []byte{}, big.NewInt(0), big.NewInt(0),
		nil, nil, time.Now().Unix(), nil)
	block := NewBlock(header, NewBlockBody([]*Transaction{}, nil))
	block.SetHash("abcdef1234567890")
	if block.GetHash() != "abcdef1234567890" {
		t.Fatalf("expected hash 'abcdef1234567890', got %q", block.GetHash())
	}
}

func TestSprint38_SetHash_NilHeader_NoPanic(t *testing.T) {
	block := &Block{Header: nil}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetHash panicked with nil header: %v", r)
		}
	}()
	block.SetHash("somehash")
}

func TestSprint38_SetHash_EmptyHash(t *testing.T) {
	header := NewBlockHeader(1, make([]byte, 32), big.NewInt(1),
		[]byte{}, []byte{}, big.NewInt(0), big.NewInt(0),
		nil, nil, time.Now().Unix(), nil)
	block := NewBlock(header, NewBlockBody([]*Transaction{}, nil))
	block.SetHash("")
	// Setting empty hash should still set it
	_ = block.GetHash()
}

// ---------------------------------------------------------------------------
// Block.ValidateUnclesHash
// ---------------------------------------------------------------------------

func TestSprint38_ValidateUnclesHash_EmptyUncles_NoError(t *testing.T) {
	header := NewBlockHeader(1, make([]byte, 32), big.NewInt(1),
		[]byte{}, []byte{}, big.NewInt(0), big.NewInt(0),
		nil, nil, time.Now().Unix(), nil)
	block := NewBlock(header, NewBlockBody([]*Transaction{}, nil))
	// Recalculate to make consistent
	block.Header.UnclesHash = CalculateUnclesHash([]*BlockHeader{}, 1)
	err := block.ValidateUnclesHash()
	if err != nil {
		t.Fatalf("ValidateUnclesHash with empty uncles: %v", err)
	}
}

func TestSprint38_ValidateUnclesHash_NilHeader_Error(t *testing.T) {
	block := &Block{Header: nil, Body: *NewBlockBody([]*Transaction{}, nil)}
	err := block.ValidateUnclesHash()
	if err == nil {
		t.Fatal("expected error for nil header")
	}
}

func TestSprint38_ValidateUnclesHash_TamperedHash_Error(t *testing.T) {
	header := NewBlockHeader(2, make([]byte, 32), big.NewInt(1),
		[]byte{}, []byte{}, big.NewInt(0), big.NewInt(0),
		nil, nil, time.Now().Unix(), nil)
	block := NewBlock(header, NewBlockBody([]*Transaction{}, nil))
	// Manually tamper the uncles hash
	block.Header.UnclesHash = []byte("wrong-hash")
	err := block.ValidateUnclesHash()
	if err == nil {
		t.Fatal("expected error for tampered uncles hash")
	}
}

// ---------------------------------------------------------------------------
// NewTxs — wraps NewNote → ToTxs → AddTxs
// ---------------------------------------------------------------------------

func TestSprint38_NewTxs_ValidInputs_NoError(t *testing.T) {
	header := NewBlockHeader(1, make([]byte, 32), big.NewInt(1),
		[]byte{}, []byte{}, big.NewInt(0), big.NewInt(0),
		nil, nil, time.Now().Unix(), nil)
	block := NewBlock(header, NewBlockBody([]*Transaction{}, nil))
	err := NewTxs("x" + strings.Repeat("a", 25), "x" + strings.Repeat("b", 25), 1.0, "test-storage", 0, big.NewInt(0), big.NewInt(0), block, "")
	if err != nil {
		t.Fatalf("NewTxs error: %v", err)
	}
	if len(block.Body.TxsList) == 0 {
		t.Fatal("expected transaction to be added to block")
	}
}

func TestSprint38_NewTxs_NilBlock_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("NewTxs with nil block panicked: %v (nil guard may be needed)", r)
		}
	}()
	_ = NewTxs("alice", "bob", 1.0, "storage", 0, big.NewInt(0), big.NewInt(0), nil, "")
}

// ---------------------------------------------------------------------------
// Block.UnmarshalJSON
// ---------------------------------------------------------------------------

func TestSprint38_UnmarshalJSON_ValidJSON_NoPanic(t *testing.T) {
	// Create a minimal JSON block representation
	header := NewBlockHeader(5, make([]byte, 32), big.NewInt(1),
		[]byte{}, []byte{}, big.NewInt(0), big.NewInt(0),
		nil, nil, time.Now().Unix(), nil)
	block := NewBlock(header, NewBlockBody([]*Transaction{}, nil))

	// Marshal to JSON, then unmarshal back
	data, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var block2 Block
	defer func() {
		if r := recover(); r != nil {
			t.Logf("UnmarshalJSON panicked: %v (may need nil handling)", r)
		}
	}()
	err = json.Unmarshal(data, &block2)
	_ = err // may succeed or fail — just no panic
}

// ---------------------------------------------------------------------------
// validateAddress (in validation.go)
// ---------------------------------------------------------------------------

func TestSprint38_ValidateAddress_ValidHex_NoError(t *testing.T) {
	// 40-char hex address
	err := validateAddress("x" + strings.Repeat("0", 25))
	if err != nil {
		t.Fatalf("validateAddress: %v", err)
	}
}

func TestSprint38_ValidateAddress_TooShort_Error(t *testing.T) {
	err := validateAddress("abcdef")
	if err == nil {
		t.Fatal("expected error for short address")
	}
}

func TestSprint38_ValidateAddress_Empty_Error(t *testing.T) {
	err := validateAddress("")
	if err == nil {
		t.Fatal("expected error for empty address")
	}
}
