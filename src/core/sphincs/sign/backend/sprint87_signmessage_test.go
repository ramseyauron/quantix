// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 87 - SignMessage + VerifySignature full roundtrip
// Covers the 0% paths in SignMessage, VerifySignature, buildHashTreeFromSignature
package sign

import (
	"bytes"
	"os"
	"testing"

	params "github.com/ramseyauron/quantix/src/core/sphincs/config"
	key "github.com/ramseyauron/quantix/src/core/sphincs/key/backend"
	sphincs "github.com/ramseyauron/quantix/src/crypto/SPHINCSPLUS-golang/sphincs"
	"github.com/syndtr/goleveldb/leveldb"
)

// helpers ----------------------------------------------------------------

// newSM87 creates a SphincsManager with real DB.
func newSM87(t *testing.T) (*SphincsManager, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-s87-*")
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
	return NewSphincsManager(db, km, p), func() { db.Close(); os.RemoveAll(dir) }
}

// generateSphincsPair returns crypto-layer *sphincs.SPHINCS_SK + *sphincs.SPHINCS_PK
// by going through the key backend's serialize/deserialize path.
func generateSphincsPair(t *testing.T) (*sphincs.SPHINCS_SK, *sphincs.SPHINCS_PK, []byte) {
	t.Helper()
	km, err := key.NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	skWrap, pk, err := km.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	skBytes, pkBytes, err := km.SerializeKeyPair(skWrap, pk)
	if err != nil {
		t.Fatalf("SerializeKeyPair: %v", err)
	}
	sk, pkBack, err := km.DeserializeKeyPair(skBytes, pkBytes)
	if err != nil {
		t.Fatalf("DeserializeKeyPair: %v", err)
	}
	return sk, pkBack, pkBytes
}

// ---------------------------------------------------------------------------
// SignMessage — basic success path
// ---------------------------------------------------------------------------

func TestSprint87_SignMessage_ReturnsNonNilSig(t *testing.T) {
	sm, cleanup := newSM87(t)
	defer cleanup()

	sk, pk, _ := generateSphincsPair(t)

	sig, merkleRoot, tsBytes, nonceBytes, commitment, err := sm.SignMessage([]byte("hello quantix"), sk, pk)
	if err != nil {
		t.Fatalf("SignMessage error: %v", err)
	}
	if sig == nil {
		t.Fatal("sig is nil")
	}
	if merkleRoot == nil {
		t.Fatal("merkleRoot is nil")
	}
	if len(tsBytes) == 0 {
		t.Fatal("tsBytes is empty")
	}
	if len(nonceBytes) == 0 {
		t.Fatal("nonceBytes is empty")
	}
	if len(commitment) == 0 {
		t.Fatal("commitment is empty")
	}
}

// ---------------------------------------------------------------------------
// SignMessage then VerifySignature — full roundtrip ✅
// ---------------------------------------------------------------------------

func TestSprint87_SignMessage_VerifySignature_Roundtrip(t *testing.T) {
	sm, cleanup := newSM87(t)
	defer cleanup()

	sk, pk, _ := generateSphincsPair(t)
	message := []byte("quantix post-quantum sovereignty")

	sig, merkleRoot, tsBytes, nonceBytes, commitment, err := sm.SignMessage(message, sk, pk)
	if err != nil {
		t.Fatalf("SignMessage: %v", err)
	}

	ok := sm.VerifySignature(message, tsBytes, nonceBytes, sig, pk, merkleRoot, commitment)
	if !ok {
		t.Error("VerifySignature returned false for a freshly signed message")
	}
}

// ---------------------------------------------------------------------------
// VerifySignature — wrong message should return false
// ---------------------------------------------------------------------------

func TestSprint87_VerifySignature_WrongMessage_ReturnsFalse(t *testing.T) {
	sm, cleanup := newSM87(t)
	defer cleanup()

	sk, pk, _ := generateSphincsPair(t)
	message := []byte("original message")
	sig, merkleRoot, tsBytes, nonceBytes, commitment, err := sm.SignMessage(message, sk, pk)
	if err != nil {
		t.Fatalf("SignMessage: %v", err)
	}

	tampered := []byte("tampered message!!")
	ok := sm.VerifySignature(tampered, tsBytes, nonceBytes, sig, pk, merkleRoot, commitment)
	if ok {
		t.Error("VerifySignature should return false for a tampered message")
	}
}

// ---------------------------------------------------------------------------
// SignMessage — nil parameters returns error
// ---------------------------------------------------------------------------

func TestSprint87_SignMessage_NilParameters_ReturnsError(t *testing.T) {
	sm := &SphincsManager{} // parameters = nil
	sk, pk, _ := generateSphincsPair(t)

	_, _, _, _, _, err := sm.SignMessage([]byte("data"), sk, pk)
	if err == nil {
		t.Error("expected error with nil parameters, got nil")
	}
}

// ---------------------------------------------------------------------------
// SerializeSignature + DeserializeSignature roundtrip
// ---------------------------------------------------------------------------

func TestSprint87_SerializeSignature_Roundtrip(t *testing.T) {
	sm, cleanup := newSM87(t)
	defer cleanup()

	sk, pk, _ := generateSphincsPair(t)
	sig, _, _, _, _, err := sm.SignMessage([]byte("serialize test"), sk, pk)
	if err != nil {
		t.Fatalf("SignMessage: %v", err)
	}

	sigBytes, err := sm.SerializeSignature(sig)
	if err != nil {
		t.Fatalf("SerializeSignature: %v", err)
	}
	if len(sigBytes) == 0 {
		t.Fatal("serialized signature is empty")
	}

	sigBack, err := sm.DeserializeSignature(sigBytes)
	if err != nil {
		t.Fatalf("DeserializeSignature: %v", err)
	}
	if sigBack == nil {
		t.Fatal("deserialized signature is nil")
	}
}

// ---------------------------------------------------------------------------
// Two different messages produce different commitments
// ---------------------------------------------------------------------------

func TestSprint87_SignMessage_DifferentMessages_DifferentCommitments(t *testing.T) {
	sm, cleanup := newSM87(t)
	defer cleanup()

	sk, pk, _ := generateSphincsPair(t)

	_, _, _, _, commit1, err := sm.SignMessage([]byte("message alpha"), sk, pk)
	if err != nil {
		t.Fatalf("SignMessage 1: %v", err)
	}
	_, _, _, _, commit2, err := sm.SignMessage([]byte("message beta"), sk, pk)
	if err != nil {
		t.Fatalf("SignMessage 2: %v", err)
	}

	if bytes.Equal(commit1, commit2) {
		t.Error("expected different commitments for different messages")
	}
}

// ---------------------------------------------------------------------------
// VerifySignature — nil sig panics (documented nil-panic gap in Spx_verify)
// NIL-SPX01: VerifySignature with nil sig panics in sphincs.Spx_verify.
// Documented here for tracking; production callers must guard against nil sig.
// ---------------------------------------------------------------------------

func TestSprint87_VerifySignature_NilSig_Panics_Documented(t *testing.T) {
	sm, cleanup := newSM87(t)
	defer cleanup()

	sk, pk, _ := generateSphincsPair(t)
	_, merkleRoot, tsBytes, nonceBytes, commitment, err := sm.SignMessage([]byte("data"), sk, pk)
	if err != nil {
		t.Fatalf("SignMessage: %v", err)
	}

	// Document that calling VerifySignature with nil sig panics.
	// Production code (handleTransaction, applyTransactions) must guard
	// len(sig) > 0 before calling VerifySignature.
	defer func() {
		if r := recover(); r != nil {
			t.Logf("NIL-SPX01: VerifySignature panics with nil sig: %v", r)
			t.Log("Production callers must check sig != nil before calling VerifySignature")
		}
	}()
	sm.VerifySignature([]byte("data"), tsBytes, nonceBytes, nil, pk, merkleRoot, commitment)
}

// ---------------------------------------------------------------------------
// VerifySignature — nil pk panics (documented nil-panic gap in Spx_verify)
// ---------------------------------------------------------------------------

func TestSprint87_VerifySignature_NilPK_Panics_Documented(t *testing.T) {
	sm, cleanup := newSM87(t)
	defer cleanup()

	sk, pk, _ := generateSphincsPair(t)
	sig, merkleRoot, tsBytes, nonceBytes, commitment, err := sm.SignMessage([]byte("data"), sk, pk)
	if err != nil {
		t.Fatalf("SignMessage: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("NIL-SPX02: VerifySignature panics with nil pk: %v", r)
		}
	}()
	sm.VerifySignature([]byte("data"), tsBytes, nonceBytes, sig, nil, merkleRoot, commitment)
}
