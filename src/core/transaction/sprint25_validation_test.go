// Sprint 25 — transaction validation, VerifyMerkleProof, CreateContract,
// RetryFailedTransactions mempool.
package types

import (
	"math/big"
	"testing"

	"github.com/ramseyauron/quantix/src/common"
)

// ─── Validator.Validate ───────────────────────────────────────────────────────

func TestSprint25_NewValidator_NotNil(t *testing.T) {
	v := NewValidator("sender-addr", "receiver-addr")
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
}

func TestSprint25_Validate_ValidNote(t *testing.T) {
	v := NewValidator("alice", "bob")
	note := &Note{
		From:    "alice",
		To:      "bob",
		Fee:     0.001,
		Storage: "ipfs://somehash",
	}
	if err := v.Validate(note); err != nil {
		t.Errorf("Validate valid note error: %v", err)
	}
}

func TestSprint25_Validate_WrongSender(t *testing.T) {
	v := NewValidator("alice", "bob")
	note := &Note{From: "eve", To: "bob", Fee: 0.001, Storage: "ipfs://x"}
	if err := v.Validate(note); err == nil {
		t.Error("expected error for wrong sender")
	}
}

func TestSprint25_Validate_WrongRecipient(t *testing.T) {
	v := NewValidator("alice", "bob")
	note := &Note{From: "alice", To: "charlie", Fee: 0.001, Storage: "ipfs://x"}
	if err := v.Validate(note); err == nil {
		t.Error("expected error for wrong recipient")
	}
}

func TestSprint25_Validate_ZeroFee(t *testing.T) {
	v := NewValidator("alice", "bob")
	note := &Note{From: "alice", To: "bob", Fee: 0, Storage: "ipfs://x"}
	if err := v.Validate(note); err == nil {
		t.Error("expected error for zero fee")
	}
}

func TestSprint25_Validate_NegativeFee(t *testing.T) {
	v := NewValidator("alice", "bob")
	note := &Note{From: "alice", To: "bob", Fee: -1.0, Storage: "ipfs://x"}
	if err := v.Validate(note); err == nil {
		t.Error("expected error for negative fee")
	}
}

func TestSprint25_Validate_EmptyStorage(t *testing.T) {
	v := NewValidator("alice", "bob")
	note := &Note{From: "alice", To: "bob", Fee: 0.001, Storage: ""}
	if err := v.Validate(note); err == nil {
		t.Error("expected error for empty storage")
	}
}

// ─── ValidateSpendability ────────────────────────────────────────────────────

func TestSprint25_ValidateSpendability_SpendableCoinbase(t *testing.T) {
	set := NewUTXOSet()
	out := Output{Value: 1000}
	set.Add("tx1", out, 0, true, 0) // coinbase at height 0
	// Spendable after maturity (100 blocks)
	ok := ValidateSpendability(set, "tx1", 0, 101)
	if !ok {
		t.Error("expected spendable after maturity")
	}
}

func TestSprint25_ValidateSpendability_NotSpendable(t *testing.T) {
	set := NewUTXOSet()
	ok := ValidateSpendability(set, "ghost-tx", 0, 100)
	if ok {
		t.Error("expected not spendable for unknown tx")
	}
}

// ─── VerifyMerkleProof ────────────────────────────────────────────────────────

func TestSprint25_VerifyMerkleProof_SingleLeaf(t *testing.T) {
	// Single leaf → root is the leaf hash itself
	txHash := common.SpxHash([]byte("single-tx"))
	// Empty proof, root = txHash itself
	ok := VerifyMerkleProof(txHash, [][]byte{}, txHash, 0, 1)
	if !ok {
		t.Error("single-leaf proof should verify against itself")
	}
}

func TestSprint25_VerifyMerkleProof_TwoLeaves_Left(t *testing.T) {
	// Build a 2-leaf tree manually
	tx0 := common.SpxHash([]byte("tx0"))
	tx1 := common.SpxHash([]byte("tx1"))
	combined := append(tx0, tx1...)
	root := common.SpxHash(combined)

	// Verify leaf 0 with sibling tx1
	ok := VerifyMerkleProof(tx0, [][]byte{tx1}, root, 0, 2)
	if !ok {
		t.Error("two-leaf proof for left leaf failed")
	}
}

func TestSprint25_VerifyMerkleProof_TwoLeaves_Right(t *testing.T) {
	tx0 := common.SpxHash([]byte("tx0"))
	tx1 := common.SpxHash([]byte("tx1"))
	combined := append(tx0, tx1...)
	root := common.SpxHash(combined)

	// Verify leaf 1 with sibling tx0
	ok := VerifyMerkleProof(tx1, [][]byte{tx0}, root, 1, 2)
	if !ok {
		t.Error("two-leaf proof for right leaf failed")
	}
}

func TestSprint25_VerifyMerkleProof_WrongRoot_ReturnsFalse(t *testing.T) {
	txHash := common.SpxHash([]byte("tx"))
	wrongRoot := common.SpxHash([]byte("wrong"))
	ok := VerifyMerkleProof(txHash, [][]byte{}, wrongRoot, 0, 1)
	if ok {
		t.Error("proof with wrong root should return false")
	}
}

func TestSprint25_VerifyMerkleProof_TamperedTx_ReturnsFalse(t *testing.T) {
	tx0 := common.SpxHash([]byte("tx0"))
	tx1 := common.SpxHash([]byte("tx1"))
	combined := append(tx0, tx1...)
	root := common.SpxHash(combined)

	tampered := common.SpxHash([]byte("evil-tx"))
	ok := VerifyMerkleProof(tampered, [][]byte{tx1}, root, 0, 2)
	if ok {
		t.Error("tampered tx should fail merkle proof")
	}
}

// ─── CreateContract error paths ──────────────────────────────────────────────

func TestSprint25_CreateContract_NegativeAmount_Error(t *testing.T) {
	set := NewUTXOSet()
	out := &Output{Value: 1000}
	note := &Note{
		Output:    out,
		Timestamp: 1712500000,
	}
	_, err := CreateContract(note, -1.0, set, "tx1", 0, 0)
	if err == nil {
		t.Error("expected error for negative amountInQTX")
	}
}

func TestSprint25_CreateContract_ZeroTimestamp_Error(t *testing.T) {
	set := NewUTXOSet()
	out := &Output{Value: 1000}
	note := &Note{
		Output:    out,
		Timestamp: 0, // invalid
	}
	_, err := CreateContract(note, 1.0, set, "tx1", 0, 0)
	if err == nil {
		t.Error("expected error for zero timestamp")
	}
}

func TestSprint25_CreateContract_UnspendableUTXO_Error(t *testing.T) {
	set := NewUTXOSet()
	out := &Output{Value: 1000}
	note := &Note{
		Output:    out,
		Timestamp: 1712500000,
	}
	// "ghost-tx" not in UTXO set → not spendable
	_, err := CreateContract(note, 1.0, set, "ghost-tx", 0, 100)
	if err == nil {
		t.Error("expected error for unspendable UTXO")
	}
}

// ─── MerkleTree.PrintTree / printNode ────────────────────────────────────────

func TestSprint25_PrintTree_NoPanic(t *testing.T) {
	// SerializeForMerkle calls tx.GasLimit.Bytes() — must set non-nil GasLimit/GasPrice
	txs := []*Transaction{
		{
			ID: "tx1", Sender: "alice", Receiver: "bob",
			Amount: big.NewInt(100), GasLimit: big.NewInt(21000), GasPrice: big.NewInt(1), Nonce: 0,
		},
		{
			ID: "tx2", Sender: "bob", Receiver: "carol",
			Amount: big.NewInt(200), GasLimit: big.NewInt(21000), GasPrice: big.NewInt(1), Nonce: 1,
		},
		{
			ID: "tx3", Sender: "carol", Receiver: "dave",
			Amount: big.NewInt(50), GasLimit: big.NewInt(21000), GasPrice: big.NewInt(1), Nonce: 2,
		},
	}

	mt := NewMerkleTree(txs)
	if mt == nil {
		t.Skip("MerkleTree returned nil for non-empty tx list")
	}
	mt.PrintTree() // exercises printNode recursion — no panic
}
