// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 35 - sphincs key backend additional coverage (nil args, empty bytes)
package key

import (
	"testing"
)

// ---------------------------------------------------------------------------
// SerializeSK — nil and empty fields
// ---------------------------------------------------------------------------

func TestSprint35_SerializeSK_Nil_Error(t *testing.T) {
	var sk *SPHINCS_SK
	_, err := sk.SerializeSK()
	if err == nil {
		t.Fatal("expected error for nil SK")
	}
}

func TestSprint35_SerializeSK_EmptyFields_Error(t *testing.T) {
	sk := &SPHINCS_SK{
		SKseed: []byte{},
		SKprf:  []byte{},
		PKseed: []byte{},
		PKroot: []byte{},
	}
	_, err := sk.SerializeSK()
	if err == nil {
		t.Fatal("expected error for SK with empty fields")
	}
}

// ---------------------------------------------------------------------------
// SerializeKeyPair — nil args
// ---------------------------------------------------------------------------

func TestSprint35_SerializeKeyPair_NilSK_Error(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	_, _, e := km.SerializeKeyPair(nil, nil)
	if e == nil {
		t.Fatal("expected error for nil SK/PK")
	}
}

// ---------------------------------------------------------------------------
// DeserializeKeyPair — empty bytes error
// ---------------------------------------------------------------------------

func TestSprint35_DeserializeKeyPair_EmptySK_Error(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	_, _, e := km.DeserializeKeyPair([]byte{}, []byte("pk"))
	if e == nil {
		t.Fatal("expected error for empty SK bytes")
	}
}

func TestSprint35_DeserializeKeyPair_NilSK_Error(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	_, _, e := km.DeserializeKeyPair(nil, []byte("pk"))
	if e == nil {
		t.Fatal("expected error for nil SK bytes")
	}
}

// ---------------------------------------------------------------------------
// DeserializePublicKey — nil/empty bytes
// ---------------------------------------------------------------------------

func TestSprint35_DeserializePublicKey_NilBytes_Error(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	_, e := km.DeserializePublicKey(nil)
	if e == nil {
		t.Fatal("expected error for nil PK bytes")
	}
}

func TestSprint35_DeserializePublicKey_EmptyBytes_Error(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	_, e := km.DeserializePublicKey([]byte{})
	if e == nil {
		t.Fatal("expected error for empty PK bytes")
	}
}

// ---------------------------------------------------------------------------
// GenerateKey — verify full return value structure
// ---------------------------------------------------------------------------

func TestSprint35_GenerateKey_AllFieldsNonNil(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	sk, pk, e := km.GenerateKey()
	if e != nil {
		t.Fatalf("GenerateKey error: %v", e)
	}
	if sk == nil {
		t.Fatal("expected non-nil SK")
	}
	if pk == nil {
		t.Fatal("expected non-nil PK")
	}
}

func TestSprint35_GenerateKey_ThenSerialize_Roundtrip(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	sk, pk, e := km.GenerateKey()
	if e != nil {
		t.Fatalf("GenerateKey error: %v", e)
	}

	skBytes, pkBytes, e2 := km.SerializeKeyPair(sk, pk)
	if e2 != nil {
		t.Fatalf("SerializeKeyPair error: %v", e2)
	}
	if len(skBytes) == 0 || len(pkBytes) == 0 {
		t.Fatal("expected non-empty serialized bytes")
	}

	// Deserialize back
	_, pk2, e3 := km.DeserializeKeyPair(skBytes, pkBytes)
	if e3 != nil {
		t.Fatalf("DeserializeKeyPair error: %v", e3)
	}
	if pk2 == nil {
		t.Fatal("expected non-nil deserialized PK")
	}
}
