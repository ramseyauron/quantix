// test(PEPPER): Sprint 71 — core/usi 57.3%→higher
// Tests: SignData error paths (empty data, empty keys),
// VerifyUSIMeta remaining error paths (invalid commitment hex, invalid merkle_root hex,
// invalid nonce, invalid sig hex, invalid pubkey hex, wrong file hash, nil meta)
package usi

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// helpers for building minimal USIMeta structs for error-path testing

func sprint71FileHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func sprint71Fingerprint(pkBytes []byte) string {
	h := sha256.Sum256(pkBytes)
	return hex.EncodeToString(h[:])
}

func sprint71BuildMeta(data []byte) *USIMeta {
	pkBytes := []byte{0xaa, 0xbb, 0xcc, 0xdd}
	return &USIMeta{
		Version:      USIMetaVersion,
		FileHash:     sprint71FileHash(data),
		PublicKey:    hex.EncodeToString(pkBytes),
		Fingerprint:  sprint71Fingerprint(pkBytes),
		Signature:    "deadbeef",
		Nonce:        "cafebabe",
		Commitment:   "0102030405060708",
		MerkleRoot:   "0102030405060708090a0b0c0d0e0f10",
		FinalDocHash: sprint71FileHash(data),
	}
}

// ─── SignData — empty data ────────────────────────────────────────────────────

func TestSprint71_SignData_EmptyData(t *testing.T) {
	_, err := SignData(nil, "fp", []byte("sk"), []byte("pk"), nil, nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestSprint71_SignData_EmptyDataBytes(t *testing.T) {
	_, err := SignData([]byte{}, "fp", []byte("sk"), []byte("pk"), nil, nil)
	if err == nil {
		t.Error("expected error for empty data")
	}
}

// ─── SignData — empty keys ────────────────────────────────────────────────────

func TestSprint71_SignData_EmptySkBytes(t *testing.T) {
	_, err := SignData([]byte("hello"), "fp", nil, []byte("pk"), nil, nil)
	if err == nil {
		t.Error("expected error for nil sk")
	}
}

func TestSprint71_SignData_EmptyPkBytes(t *testing.T) {
	_, err := SignData([]byte("hello"), "fp", []byte("sk"), nil, nil, nil)
	if err == nil {
		t.Error("expected error for nil pk")
	}
}

func TestSprint71_SignData_BothKeysEmpty(t *testing.T) {
	_, err := SignData([]byte("hello"), "fp", []byte{}, []byte{}, nil, nil)
	if err == nil {
		t.Error("expected error for empty sk+pk")
	}
}

// ─── VerifyUSIMeta — nil meta ─────────────────────────────────────────────────

func TestSprint71_VerifyUSIMeta_NilMeta(t *testing.T) {
	_, err := VerifyUSIMeta([]byte("data"), nil, nil, nil)
	if err == nil {
		t.Error("expected error for nil meta")
	}
}

// ─── VerifyUSIMeta — wrong file hash ──────────────────────────────────────────

func TestSprint71_VerifyUSIMeta_WrongFileHash(t *testing.T) {
	data := []byte("testdata")
	meta := sprint71BuildMeta(data)
	meta.FileHash = "0000000000000000000000000000000000000000000000000000000000000000"
	_, err := VerifyUSIMeta(data, meta, nil, nil)
	if err == nil {
		t.Error("expected error for wrong file hash")
	}
}

// ─── VerifyUSIMeta — invalid public key hex ───────────────────────────────────

func TestSprint71_VerifyUSIMeta_InvalidPublicKeyHex(t *testing.T) {
	data := []byte("testdata")
	meta := sprint71BuildMeta(data)
	meta.PublicKey = "notvalidhex!!!"
	_, err := VerifyUSIMeta(data, meta, nil, nil)
	if err == nil {
		t.Error("expected error for invalid public key hex")
	}
}

// ─── VerifyUSIMeta — wrong fingerprint ────────────────────────────────────────

func TestSprint71_VerifyUSIMeta_WrongFingerprint(t *testing.T) {
	data := []byte("testdata")
	meta := sprint71BuildMeta(data)
	// Fingerprint doesn't match sha256(publicKey)
	meta.Fingerprint = "0000000000000000000000000000000000000000000000000000000000000000"
	_, err := VerifyUSIMeta(data, meta, nil, nil)
	if err == nil {
		t.Error("expected error for wrong fingerprint")
	}
}

// ─── VerifyUSIMeta — invalid signature hex ────────────────────────────────────

func TestSprint71_VerifyUSIMeta_InvalidSignatureHex(t *testing.T) {
	data := []byte("testdata")
	meta := sprint71BuildMeta(data)
	meta.Signature = "!!notvalidhex"
	_, err := VerifyUSIMeta(data, meta, nil, nil)
	if err == nil {
		t.Error("expected error for invalid signature hex")
	}
}

// ─── VerifyUSIMeta — invalid nonce hex ────────────────────────────────────────

func TestSprint71_VerifyUSIMeta_InvalidNonceHex(t *testing.T) {
	data := []byte("testdata")
	meta := sprint71BuildMeta(data)
	meta.Nonce = "!!notvalidhex"
	_, err := VerifyUSIMeta(data, meta, nil, nil)
	if err == nil {
		t.Error("expected error for invalid nonce hex")
	}
}

// Note: invalid commitment hex and invalid merkle_root hex tests require a real
// KeyManager to deserialize public key (step 3 precedes commitment decode step).
// These paths are covered by sprint21 tests that use a full SphincsManager.
