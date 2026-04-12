// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 57 — sphincs/key/backend 71%→higher
// Tests: GenerateKey nil-params, SerializeKeyPair nil paths,
// DeserializeKeyPair empty pkBytes, DeserializePublicKey nil-params
package key

import (
	"testing"

	params "github.com/ramseyauron/quantix/src/core/sphincs/config"
)

// helper: KeyManager with nil params (to test nil-params paths)
func nilParamsKM() *KeyManager {
	return &KeyManager{Params: nil}
}

// helper: KeyManager with valid params but nil inner Params.Params
func nilInnerParamsKM(t *testing.T) *KeyManager {
	t.Helper()
	p, err := params.NewSPHINCSParameters()
	if err != nil {
		t.Fatalf("NewSPHINCSParameters: %v", err)
	}
	// Simulate nil inner params
	return &KeyManager{Params: &params.SPHINCSParameters{Params: nil}}
	_ = p
	return nil
}

// ─── GenerateKey — nil params paths ───────────────────────────────────────────

func TestSprint57_GenerateKey_NilParams(t *testing.T) {
	km := nilParamsKM()
	_, _, err := km.GenerateKey()
	if err == nil {
		t.Error("expected error for nil Params")
	}
}

func TestSprint57_GenerateKey_NilInnerParams(t *testing.T) {
	km := &KeyManager{Params: &params.SPHINCSParameters{Params: nil}}
	_, _, err := km.GenerateKey()
	if err == nil {
		t.Error("expected error for nil Params.Params")
	}
}

// ─── SerializeKeyPair — nil SK or PK ──────────────────────────────────────────

func TestSprint57_SerializeKeyPair_NilSK(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	_, _, err = km.SerializeKeyPair(nil, nil)
	if err == nil {
		t.Error("expected error for nil SK")
	}
}

func TestSprint57_SerializeKeyPair_NilPK(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	// Generate a real SK but pass nil PK
	sk, _, err2 := km.GenerateKey()
	if err2 != nil {
		t.Fatalf("GenerateKey: %v", err2)
	}
	_, _, err = km.SerializeKeyPair(sk, nil)
	if err == nil {
		t.Error("expected error for nil PK")
	}
}

// ─── DeserializeKeyPair — empty pkBytes ───────────────────────────────────────

func TestSprint57_DeserializeKeyPair_EmptyPKBytes(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	// Generate and serialize a real SK to get valid skBytes
	sk, pk, err2 := km.GenerateKey()
	if err2 != nil {
		t.Fatalf("GenerateKey: %v", err2)
	}
	skBytes, _, err3 := km.SerializeKeyPair(sk, pk)
	if err3 != nil {
		t.Fatalf("SerializeKeyPair: %v", err3)
	}
	// Now pass valid skBytes but empty pkBytes
	_, _, err = km.DeserializeKeyPair(skBytes, []byte{})
	if err == nil {
		t.Error("expected error for empty pkBytes")
	}
}

func TestSprint57_DeserializeKeyPair_NilParams(t *testing.T) {
	km := nilParamsKM()
	_, _, err := km.DeserializeKeyPair([]byte("sk"), []byte("pk"))
	if err == nil {
		t.Error("expected error for nil params")
	}
}

// ─── DeserializePublicKey — nil params ────────────────────────────────────────

func TestSprint57_DeserializePublicKey_NilParams(t *testing.T) {
	km := nilParamsKM()
	_, err := km.DeserializePublicKey([]byte("pkbytes"))
	if err == nil {
		t.Error("expected error for nil params in DeserializePublicKey")
	}
}

// ─── GetSPHINCSParameters — returns Params ────────────────────────────────────

func TestSprint57_GetSPHINCSParameters_NotNil(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	p := km.GetSPHINCSParameters()
	if p == nil {
		t.Error("expected non-nil SPHINCSParameters")
	}
}

func TestSprint57_GetSPHINCSParameters_Nil(t *testing.T) {
	km := nilParamsKM()
	p := km.GetSPHINCSParameters()
	if p != nil {
		t.Error("expected nil for nil-params KeyManager")
	}
}
