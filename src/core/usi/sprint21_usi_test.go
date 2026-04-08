// Sprint 21b — USI VerifyUSIMeta error-path coverage (no real SPHINCS+ needed)
// Tests the early-exit validation steps: nil meta, file_hash mismatch, fingerprint mismatch,
// invalid hex in signature/nonce/commitment/merkle_root fields.
package usi

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// ─── VerifyUSIMeta error paths ────────────────────────────────────────────────

func TestSprint21_VerifyUSIMeta_NilMeta_Error(t *testing.T) {
	ok, err := VerifyUSIMeta([]byte("data"), nil, nil, nil)
	if ok {
		t.Error("expected false for nil meta")
	}
	if err == nil {
		t.Error("expected error for nil meta")
	}
}

func TestSprint21_VerifyUSIMeta_WrongFileHash_Error(t *testing.T) {
	data := []byte("hello")
	meta := &USIMeta{
		FileHash: "0000000000000000000000000000000000000000000000000000000000000000",
	}
	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for wrong file_hash")
	}
	if err == nil {
		t.Error("expected error for file_hash mismatch")
	}
}

func TestSprint21_VerifyUSIMeta_CorrectFileHash_WrongFingerprint(t *testing.T) {
	data := []byte("hello")
	h := sha256.Sum256(data)
	correctHash := hex.EncodeToString(h[:])

	// Wrong fingerprint (not SHA-256 of pubkey)
	meta := &USIMeta{
		FileHash:    correctHash,
		PublicKey:   hex.EncodeToString([]byte("pubkeydata")),
		Fingerprint: "0000000000000000000000000000000000000000000000000000000000000000",
	}
	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for wrong fingerprint")
	}
	if err == nil {
		t.Error("expected error for fingerprint mismatch")
	}
}

func TestSprint21_VerifyUSIMeta_InvalidPublicKeyHex_Error(t *testing.T) {
	data := []byte("hello")
	h := sha256.Sum256(data)
	correctHash := hex.EncodeToString(h[:])

	meta := &USIMeta{
		FileHash:  correctHash,
		PublicKey: "not-valid-hex!!!",
	}
	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for invalid pubkey hex")
	}
	if err == nil {
		t.Error("expected error for invalid pubkey hex")
	}
}

func TestSprint21_VerifyUSIMeta_InvalidSignatureHex_Error(t *testing.T) {
	data := []byte("hello")
	dataHash := sha256.Sum256(data)

	pkBytes := []byte("some-public-key-bytes")
	pkHex := hex.EncodeToString(pkBytes)
	fpHash := sha256.Sum256(pkBytes)
	fpHex := hex.EncodeToString(fpHash[:])

	meta := &USIMeta{
		FileHash:    hex.EncodeToString(dataHash[:]),
		PublicKey:   pkHex,
		Fingerprint: fpHex,
		Signature:   "ZZZNOTVALIDHEX!!!",
	}
	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for invalid signature hex")
	}
	if err == nil {
		t.Error("expected error for invalid signature hex")
	}
}

func TestSprint21_VerifyUSIMeta_InvalidNonceHex_Error(t *testing.T) {
	data := []byte("hello")
	dataHash := sha256.Sum256(data)

	pkBytes := []byte("some-public-key-bytes")
	pkHex := hex.EncodeToString(pkBytes)
	fpHash := sha256.Sum256(pkBytes)
	fpHex := hex.EncodeToString(fpHash[:])

	meta := &USIMeta{
		FileHash:    hex.EncodeToString(dataHash[:]),
		PublicKey:   pkHex,
		Fingerprint: fpHex,
		Signature:   hex.EncodeToString([]byte("dummysig")),
		Nonce:       "NOTVALIDHEX!!!",
	}
	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for invalid nonce hex")
	}
	if err == nil {
		t.Error("expected error for invalid nonce hex")
	}
}

// ─── SignData error paths ──────────────────────────────────────────────────────

func TestSprint21_SignData_EmptyData_Error(t *testing.T) {
	_, err := SignData([]byte{}, "fp", []byte("sk"), []byte("pk"), nil, nil)
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestSprint21_SignData_EmptyKeys_Error(t *testing.T) {
	_, err := SignData([]byte("data"), "fp", nil, nil, nil, nil)
	if err == nil {
		t.Error("expected error for nil keys")
	}
}
