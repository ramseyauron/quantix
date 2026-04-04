// PEPPER Sprint 2 — usi package coverage push for VerifyUSIMeta error paths
package usi

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// makeValidMeta creates a minimal USIMeta with correct file_hash and fingerprint
// for a given data and pubkey, used to test partial validation paths.
func makeValidFileHashMeta(data []byte, pubKeyBytes []byte) *USIMeta {
	fileHashArr := sha256.Sum256(data)
	fileHash := hex.EncodeToString(fileHashArr[:])

	fpArr := sha256.Sum256(pubKeyBytes)
	fp := hex.EncodeToString(fpArr[:])
	pubKeyHex := hex.EncodeToString(pubKeyBytes)

	return &USIMeta{
		Version:     USIMetaVersion,
		Fingerprint: fp,
		PublicKey:   pubKeyHex,
		FileHash:    fileHash,
		Signature:   hex.EncodeToString([]byte("fakesig")),
		Nonce:       hex.EncodeToString([]byte("fakenonce")),
		Commitment:  hex.EncodeToString([]byte("fakecommit")),
		MerkleRoot:  hex.EncodeToString(make([]byte, 32)),
		Timestamp:   1000000,
	}
}

// ── VerifyUSIMeta nil meta ────────────────────────────────────────────────────

func TestVerifyUSIMeta_NilMeta_ReturnsFalse(t *testing.T) {
	ok, err := VerifyUSIMeta([]byte("data"), nil, nil, nil)
	if ok {
		t.Error("expected false for nil meta")
	}
	if err == nil {
		t.Error("expected non-nil error for nil meta")
	}
}

// ── VerifyUSIMeta file_hash mismatch ─────────────────────────────────────────

func TestVerifyUSIMeta_FileHashMismatch_ReturnsFalse(t *testing.T) {
	data := []byte("original data")
	meta := makeValidFileHashMeta(data, []byte("pubkey"))
	meta.FileHash = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for file_hash mismatch")
	}
	if err == nil {
		t.Error("expected error for file_hash mismatch")
	}
}

// ── VerifyUSIMeta invalid public_key hex ──────────────────────────────────────

func TestVerifyUSIMeta_InvalidPubKeyHex_ReturnsFalse(t *testing.T) {
	data := []byte("test data")
	fileHashArr := sha256.Sum256(data)
	meta := &USIMeta{
		Version:   USIMetaVersion,
		FileHash:  hex.EncodeToString(fileHashArr[:]),
		PublicKey: "NOT_VALID_HEX!!", // invalid hex
	}

	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for invalid pubkey hex")
	}
	if err == nil {
		t.Error("expected error for invalid pubkey hex")
	}
}

// ── VerifyUSIMeta fingerprint mismatch ───────────────────────────────────────

func TestVerifyUSIMeta_FingerprintMismatch_ReturnsFalse(t *testing.T) {
	data := []byte("hello usi")
	pubKeyBytes := []byte("some-public-key-bytes")
	meta := makeValidFileHashMeta(data, pubKeyBytes)
	meta.Fingerprint = "wrongfingerprintwrongfingerprintwrongfingerprintwrongfingerprint"

	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for fingerprint mismatch")
	}
	if err == nil {
		t.Error("expected error for fingerprint mismatch")
	}
}

// ── VerifyUSIMeta invalid signature hex ──────────────────────────────────────

func TestVerifyUSIMeta_InvalidSigHex_ReturnsFalse(t *testing.T) {
	data := []byte("test sig hex")
	pubKeyBytes := []byte("test-pk")
	meta := makeValidFileHashMeta(data, pubKeyBytes)
	meta.Signature = "ZZ_NOT_HEX" // invalid hex

	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for invalid signature hex")
	}
	if err == nil {
		t.Error("expected error for invalid signature hex")
	}
}

// ── VerifyUSIMeta invalid nonce hex ──────────────────────────────────────────

func TestVerifyUSIMeta_InvalidNonceHex_ReturnsFalse(t *testing.T) {
	data := []byte("test nonce hex")
	pubKeyBytes := []byte("nonce-pk")
	meta := makeValidFileHashMeta(data, pubKeyBytes)
	meta.Nonce = "ZZ_NOT_HEX_NONCE"

	ok, err := VerifyUSIMeta(data, meta, nil, nil)
	if ok {
		t.Error("expected false for invalid nonce hex")
	}
	if err == nil {
		t.Error("expected error for invalid nonce hex")
	}
}
