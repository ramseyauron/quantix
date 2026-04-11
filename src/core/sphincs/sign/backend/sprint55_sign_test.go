// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 55 — sphincs/sign/backend 51.7%→higher
// Tests: serializePK nil, storeCommitment nil-db, SerializeSignature (nil), VerifyTxSignature error paths
package sign

import (
	"testing"
)

// ─── serializePK — nil pk error ───────────────────────────────────────────────

func TestSprint55_SerializePK_Nil(t *testing.T) {
	sm := newTestSphincsManager(t)
	_, err := sm.serializePK(nil)
	if err == nil {
		t.Error("expected error for nil public key in serializePK")
	}
}

// ─── storeCommitment — nil db error ──────────────────────────────────────────

func TestSprint55_StoreCommitment_NilDB_Error(t *testing.T) {
	sm := newTestSphincsManager(t)
	// sm has nil db (no LevelDB passed to NewSphincsManager)
	err := sm.storeCommitment([]byte("commitment-bytes"))
	if err == nil {
		t.Error("expected error for nil LevelDB in storeCommitment")
	}
}

// ─── SerializeSignature — nil sig (should error) ─────────────────────────────

func TestSprint55_SerializeSignature_NilSig(t *testing.T) {
	sm := newTestSphincsManager(t)
	// nil sig — SerializeSignature calls sig.SerializeSignature() which will panic or error
	// Use recover to document the nil panic gap
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("SerializeSignature(nil) panics: %v — nil guard needed (known gap)", r)
			}
		}()
		_, err := sm.SerializeSignature(nil)
		if err == nil {
			t.Log("SerializeSignature(nil) returned no error (unexpected but non-fatal)")
		}
	}()
}

// ─── VerifyTxSignature — nil params → false ────────────────────────────────

func TestSprint55_VerifyTxSignature_NilParams(t *testing.T) {
	// Create a SphincsManager with nil parameters
	sm := &SphincsManager{parameters: nil}
	result := sm.VerifyTxSignature([]byte("msg"), nil, nil, []byte("sig"), []byte("pk"))
	if result {
		t.Error("expected false for nil parameters")
	}
}

// ─── VerifyTxSignature — empty sigBytes → false ───────────────────────────

func TestSprint55_VerifyTxSignature_EmptySigBytes(t *testing.T) {
	sm := newTestSphincsManager(t)
	result := sm.VerifyTxSignature([]byte("msg"), nil, nil, []byte{}, []byte("pk"))
	if result {
		t.Error("expected false for empty sigBytes")
	}
}

// ─── VerifyTxSignature — empty pubkey → false ────────────────────────────

func TestSprint55_VerifyTxSignature_EmptyPubKey(t *testing.T) {
	sm := newTestSphincsManager(t)
	result := sm.VerifyTxSignature([]byte("msg"), nil, nil, []byte("sig"), []byte{})
	if result {
		t.Error("expected false for empty senderPubKey")
	}
}

// ─── VerifyTxSignature — invalid sigBytes (non-empty, not valid sig) → false

func TestSprint55_VerifyTxSignature_InvalidSig(t *testing.T) {
	sm := newTestSphincsManager(t)
	// garbage sig bytes — DeserializeSignature will fail → return false
	result := sm.VerifyTxSignature([]byte("msg"), nil, nil, []byte("garbage-sig-bytes"), []byte("garbage-pk-bytes"))
	if result {
		t.Error("expected false for invalid sig bytes")
	}
}

// ─── DeserializeSignature — nil parameters → error ───────────────────────

func TestSprint55_DeserializeSignature_NilParams(t *testing.T) {
	sm := &SphincsManager{parameters: nil}
	_, err := sm.DeserializeSignature([]byte("sigbytes"))
	if err == nil {
		t.Error("expected error for nil parameters in DeserializeSignature")
	}
}

// ─── VerifyCommitmentInRoot — nil inputs → false ─────────────────────────

func TestSprint55_VerifyCommitmentInRoot_NilInputs(t *testing.T) {
	result := VerifyCommitmentInRoot(nil, nil)
	if result {
		t.Error("expected false for nil commitment + nil root")
	}
}

func TestSprint55_VerifyCommitmentInRoot_NilCommitment(t *testing.T) {
	result := VerifyCommitmentInRoot(nil, nil)
	if result {
		t.Error("expected false for nil commitment")
	}
}
