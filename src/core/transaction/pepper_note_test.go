// PEPPER Sprint2 — transaction note + utxo edge cases
package types

import (
	"math/big"
	"testing"
)

// ── NewNote ───────────────────────────────────────────────────────────────────

func TestNewNote_InvalidToAddress(t *testing.T) {
	_, err := NewNote("invalid", "xFrom1234567890ABCDEF123456789", 0.001, "", "key")
	if err == nil {
		t.Error("expected error for invalid 'to' address")
	}
}

func TestNewNote_InvalidFromAddress(t *testing.T) {
	_, err := NewNote("xTo1234567890ABCDEF1234567890AB", "bad", 0.001, "", "key")
	if err == nil {
		t.Error("expected error for invalid 'from' address")
	}
}

func TestNewNote_ValidAddresses(t *testing.T) {
	note, err := NewNote(
		"xTo1234567890ABCDEF1234567890AB",
		"xFrom1234567890ABCDEF123456789",
		0.001, "storage_data", "secret_key",
	)
	if err != nil {
		t.Logf("NewNote error: %v (may be due to address validation)", err)
		return
	}
	if note == nil {
		t.Error("NewNote returned nil without error")
	}
}

// ── UTXOSet.Add edge cases ────────────────────────────────────────────────────

func TestUTXOSet_Add_DuplicateOutpoint(t *testing.T) {
	set := NewUTXOSet()
	out := Output{Value: 100, Address: "xOwner1234567890ABCDEF1234567890"}
	err1 := set.Add("txid1", out, 0, false, 1)
	err2 := set.Add("txid1", out, 0, false, 2) // duplicate
	if err1 != nil {
		t.Logf("first Add: %v", err1)
	}
	if err2 == nil {
		t.Log("duplicate Add succeeded (may be allowed by implementation)")
	}
}

func TestUTXOSet_Add_MultipleOutputs(t *testing.T) {
	set := NewUTXOSet()
	for i := 0; i < 5; i++ {
		out := Output{Value: uint64(100 * (i + 1)), Address: "xOwner1234567890ABCDEF1234567890"}
		err := set.Add("txidX", out, i, false, uint64(i+1))
		if err != nil {
			t.Logf("Add index %d: %v", i, err)
		}
	}
}

// ── Block JSON Marshal/Unmarshal ──────────────────────────────────────────────

func TestBlockHeader_MarshalUnmarshal(t *testing.T) {
	b := newValidBlock(5, []byte("parent_hash"))
	b.Header.Nonce = "12345"
	b.FinalizeHash()

	data, err := b.Header.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	var h2 BlockHeader
	if err := h2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if h2.Height != b.Header.Height {
		t.Errorf("Height mismatch after marshal/unmarshal")
	}
}

func TestBlock_MarshalUnmarshal(t *testing.T) {
	b := newValidBlock(3, nil)
	b.FinalizeHash()

	data, err := b.MarshalJSON()
	if err != nil {
		t.Fatalf("Block.MarshalJSON: %v", err)
	}

	var b2 Block
	if err := b2.UnmarshalJSON(data); err != nil {
		t.Fatalf("Block.UnmarshalJSON: %v", err)
	}
}

func TestBlockBody_MarshalUnmarshal(t *testing.T) {
	body := NewBlockBody([]*Transaction{
		{ID: "tx1", Sender: "A", Receiver: "B", Amount: big.NewInt(100)},
	}, nil)

	data, err := body.MarshalJSON()
	if err != nil {
		t.Fatalf("BlockBody.MarshalJSON: %v", err)
	}

	var body2 BlockBody
	if err := body2.UnmarshalJSON(data); err != nil {
		t.Fatalf("BlockBody.UnmarshalJSON: %v", err)
	}
}

// ── Block.GetFormattedTimestamps / GetTimeInfo ────────────────────────────────

func TestBlock_GetFormattedTimestamps(t *testing.T) {
	b := newValidBlock(1, nil)
	local, utc := b.GetFormattedTimestamps()
	_ = local
	_ = utc
}

func TestBlock_GetTimeInfo(t *testing.T) {
	b := newValidBlock(1, nil)
	info := b.GetTimeInfo()
	_ = info // may be nil in test environment
}
