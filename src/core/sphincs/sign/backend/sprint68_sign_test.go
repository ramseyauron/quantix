// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 68 — sphincs/sign/backend 59.1%→higher
// Tests: VerifyCommitmentInRoot non-nil different/same roots, storeCommitment nil-db,
// VerifyTxSignature garbage sig, generateNonce uniqueness
package sign

import (
	"testing"

	"github.com/holiman/uint256"
	hashtree "github.com/ramseyauron/quantix/src/core/hashtree"
)

// ─── VerifyCommitmentInRoot — non-nil different hashes ───────────────────────

func TestSprint68_VerifyCommitmentInRoot_DifferentHashes(t *testing.T) {
	root1 := &hashtree.HashTreeNode{Hash: uint256.NewInt(0).SetBytes32(make([]byte, 32))}
	b := make([]byte, 32)
	b[0] = 0xff
	root2 := &hashtree.HashTreeNode{Hash: uint256.NewInt(0).SetBytes32(b)}

	result := VerifyCommitmentInRoot(root1, root2)
	if result {
		t.Error("expected false for different root hashes")
	}
}

func TestSprint68_VerifyCommitmentInRoot_SameHash(t *testing.T) {
	hashBytes := make([]byte, 32)
	for i := range hashBytes {
		hashBytes[i] = byte(i + 1)
	}
	root1 := &hashtree.HashTreeNode{Hash: uint256.NewInt(0).SetBytes32(hashBytes)}
	root2 := &hashtree.HashTreeNode{Hash: uint256.NewInt(0).SetBytes32(hashBytes)}

	result := VerifyCommitmentInRoot(root1, root2)
	if !result {
		t.Error("expected true for identical root hashes")
	}
}

func TestSprint68_VerifyCommitmentInRoot_OneNil(t *testing.T) {
	root := &hashtree.HashTreeNode{Hash: uint256.NewInt(0)}
	result := VerifyCommitmentInRoot(root, nil)
	if result {
		t.Error("expected false when one root is nil")
	}
}

// ─── storeCommitment — nil db (already tested in sprint55, now confirm path)

func TestSprint68_StoreCommitment_NilDB_ConfirmError(t *testing.T) {
	sm := newTestSphincsManager(t)
	// sm has nil db — storeCommitment must return error
	err := sm.storeCommitment([]byte("commitment-data-sprint68"))
	if err == nil {
		t.Error("expected error for storeCommitment with nil DB")
	}
}

// ─── VerifyTxSignature — garbage sig that fails DeserializeSignature ─────────

func TestSprint68_VerifyTxSignature_GarbageSigBytes(t *testing.T) {
	sm := newTestSphincsManager(t)
	result := sm.VerifyTxSignature(
		[]byte("message"),
		[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		[]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18},
		[]byte("garbage-sig-not-valid-sphincs-signature-data-bytes"),
		[]byte("garbage-pk-not-valid-sphincs-public-key-data-bytes"),
	)
	if result {
		t.Error("expected false for garbage sig/pk bytes")
	}
}

// ─── generateNonce — uniqueness ───────────────────────────────────────────────

func TestSprint68_GenerateNonce_UniqueAcrossCalls(t *testing.T) {
	n1, err1 := generateNonce()
	n2, err2 := generateNonce()
	if err1 != nil || err2 != nil {
		t.Fatalf("generateNonce errors: %v, %v", err1, err2)
	}
	if string(n1) == string(n2) {
		t.Error("generateNonce returned same nonce twice")
	}
}

func TestSprint68_GenerateNonce_Length(t *testing.T) {
	n, err := generateNonce()
	if err != nil {
		t.Fatalf("generateNonce error: %v", err)
	}
	if len(n) == 0 {
		t.Error("expected non-empty nonce")
	}
}
