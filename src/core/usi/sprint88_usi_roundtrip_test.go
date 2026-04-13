// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 88 - USI SignData + VerifyUSIMeta full roundtrip
// Covers SignData success path and VerifyUSIMeta end-to-end (57.3% → higher)
package usi_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	params "github.com/ramseyauron/quantix/src/core/sphincs/config"
	key "github.com/ramseyauron/quantix/src/core/sphincs/key/backend"
	sign "github.com/ramseyauron/quantix/src/core/sphincs/sign/backend"
	"github.com/ramseyauron/quantix/src/core/usi"
	"github.com/syndtr/goleveldb/leveldb"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type usiTestEnv struct {
	sm      *sign.SphincsManager
	km      *key.KeyManager
	skBytes []byte
	pkBytes []byte
	fp      string
	cleanup func()
}

func newUSIEnv(t *testing.T) *usiTestEnv {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-usi88-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("leveldb.OpenFile: %v", err)
	}
	p, err := params.NewSPHINCSParameters()
	if err != nil {
		db.Close(); os.RemoveAll(dir)
		t.Fatalf("NewSPHINCSParameters: %v", err)
	}
	km, err := key.NewKeyManager()
	if err != nil {
		db.Close(); os.RemoveAll(dir)
		t.Fatalf("NewKeyManager: %v", err)
	}
	sm := sign.NewSphincsManager(db, km, p)

	// Generate keypair + serialize
	skWrap, pk, err := km.GenerateKey()
	if err != nil {
		db.Close(); os.RemoveAll(dir)
		t.Fatalf("GenerateKey: %v", err)
	}
	skBytes, pkBytes, err := km.SerializeKeyPair(skWrap, pk)
	if err != nil {
		db.Close(); os.RemoveAll(dir)
		t.Fatalf("SerializeKeyPair: %v", err)
	}

	// Fingerprint = SHA-256(pkBytes)
	fpArr := sha256.Sum256(pkBytes)
	fp := hex.EncodeToString(fpArr[:])

	return &usiTestEnv{
		sm: sm, km: km, skBytes: skBytes, pkBytes: pkBytes, fp: fp,
		cleanup: func() { db.Close(); os.RemoveAll(dir) },
	}
}

// ---------------------------------------------------------------------------
// Sprint 88 Tests
// ---------------------------------------------------------------------------

// TestSprint88_SignData_ReturnsNonNilMeta tests the happy path of SignData.
func TestSprint88_SignData_ReturnsNonNilMeta(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	data := []byte("quantix sovereign document content")
	meta, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData error: %v", err)
	}
	if meta == nil {
		t.Fatal("expected non-nil USIMeta")
	}
}

// TestSprint88_SignData_MetaFieldsPopulated checks all meta fields are populated.
func TestSprint88_SignData_MetaFieldsPopulated(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	data := []byte("document for field population check")
	meta, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData error: %v", err)
	}

	if meta.Version == "" {
		t.Error("Version is empty")
	}
	if meta.Fingerprint != env.fp {
		t.Errorf("Fingerprint = %q, want %q", meta.Fingerprint, env.fp)
	}
	if meta.PublicKey == "" {
		t.Error("PublicKey is empty")
	}
	if len(meta.FileHash) != 64 {
		t.Errorf("FileHash len = %d, want 64", len(meta.FileHash))
	}
	if len(meta.FinalDocHash) != 64 {
		t.Errorf("FinalDocHash len = %d, want 64", len(meta.FinalDocHash))
	}
	if meta.Timestamp == 0 {
		t.Error("Timestamp is zero")
	}
	if meta.Nonce == "" {
		t.Error("Nonce is empty")
	}
	if meta.Commitment == "" {
		t.Error("Commitment is empty")
	}
	if meta.Signature == "" {
		t.Error("Signature is empty")
	}
	if meta.SignedAt == "" {
		t.Error("SignedAt is empty")
	}
}

// TestSprint88_SignData_FileHash_MatchesSHA256Data verifies file_hash = SHA-256(data).
func TestSprint88_SignData_FileHash_MatchesSHA256Data(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	data := []byte("data integrity test content")
	meta, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData error: %v", err)
	}

	expected := sha256.Sum256(data)
	expectedHex := hex.EncodeToString(expected[:])
	if meta.FileHash != expectedHex {
		t.Errorf("FileHash = %q, want %q", meta.FileHash, expectedHex)
	}
}

// TestSprint88_SignData_Fingerprint_MatchesPKHash verifies fingerprint = SHA-256(pubkey).
func TestSprint88_SignData_Fingerprint_MatchesPKHash(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	data := []byte("fingerprint binding test")
	meta, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData error: %v", err)
	}

	pkBytes, _ := hex.DecodeString(meta.PublicKey)
	fpArr := sha256.Sum256(pkBytes)
	expectedFP := hex.EncodeToString(fpArr[:])
	if meta.Fingerprint != expectedFP {
		t.Errorf("Fingerprint = %q, want SHA-256(pubkey) = %q", meta.Fingerprint, expectedFP)
	}
}

// TestSprint88_VerifyUSIMeta_ValidMeta_ReturnsTrue tests the end-to-end roundtrip.
func TestSprint88_VerifyUSIMeta_ValidMeta_ReturnsTrue(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	data := []byte("the truth is signed on-chain")
	meta, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData error: %v", err)
	}

	ok, err := usi.VerifyUSIMeta(data, meta, env.sm, env.km)
	if err != nil {
		t.Fatalf("VerifyUSIMeta error: %v", err)
	}
	if !ok {
		t.Error("VerifyUSIMeta returned false for a freshly signed document")
	}
}

// TestSprint88_VerifyUSIMeta_TamperedData_ReturnsFalse ensures data tampering is detected.
func TestSprint88_VerifyUSIMeta_TamperedData_ReturnsFalse(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	data := []byte("original document content")
	meta, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData error: %v", err)
	}

	tampered := []byte("tampered document content!!")
	ok, err := usi.VerifyUSIMeta(tampered, meta, env.sm, env.km)
	// Should fail — file_hash mismatch
	if ok {
		t.Error("VerifyUSIMeta should return false for tampered data")
	}
	if err == nil {
		t.Error("expected error for tampered data, got nil")
	}
}

// TestSprint88_SignData_Deterministic_FileHash shows file_hash is stable across calls.
func TestSprint88_SignData_Deterministic_FileHash(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	data := []byte("determinism check")
	meta1, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData 1: %v", err)
	}
	meta2, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData 2: %v", err)
	}

	// FileHash should be identical (deterministic SHA-256 of same data)
	if meta1.FileHash != meta2.FileHash {
		t.Error("FileHash should be deterministic for same data")
	}
	// Signatures should differ (different timestamp+nonce per call)
	if meta1.Signature == meta2.Signature {
		t.Log("Note: signatures are identical — expected if same timestamp/nonce (rare)")
	}
}

// TestSprint88_SignData_DifferentData_DifferentFileHash ensures different data → different hash.
func TestSprint88_SignData_DifferentData_DifferentFileHash(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	meta1, _ := usi.SignData([]byte("document alpha"), env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	meta2, _ := usi.SignData([]byte("document beta"), env.fp, env.skBytes, env.pkBytes, env.sm, env.km)

	if meta1.FileHash == meta2.FileHash {
		t.Error("different data should produce different FileHash")
	}
}

// TestSprint88_VerifyUSIMeta_EmptyData_ReturnsFalse tests empty data rejection.
func TestSprint88_VerifyUSIMeta_EmptyData_ReturnsFalse(t *testing.T) {
	env := newUSIEnv(t)
	defer env.cleanup()

	data := []byte("some content")
	meta, err := usi.SignData(data, env.fp, env.skBytes, env.pkBytes, env.sm, env.km)
	if err != nil {
		t.Fatalf("SignData error: %v", err)
	}

	// Verify with empty data — file_hash won't match
	ok, err := usi.VerifyUSIMeta([]byte{}, meta, env.sm, env.km)
	if ok {
		t.Error("expected false for empty data against non-empty meta")
	}
	_ = err
}
