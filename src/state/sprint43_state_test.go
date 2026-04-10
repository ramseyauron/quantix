// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 43 - state package pure functions: isHexString, getStringFromMap, getUint64FromMap, isHexEncodedGenesis, GetChainStatePath
package state

import (
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// isHexString
// ---------------------------------------------------------------------------

func TestSprint43_IsHexString_ValidHex_True(t *testing.T) {
	if !isHexString("aabbcc") {
		t.Fatal("expected true for valid hex string")
	}
}

func TestSprint43_IsHexString_Empty_False(t *testing.T) {
	if isHexString("") {
		t.Fatal("expected false for empty string")
	}
}

func TestSprint43_IsHexString_OddLength_False(t *testing.T) {
	if isHexString("abc") {
		t.Fatal("expected false for odd-length string")
	}
}

func TestSprint43_IsHexString_NonHex_False(t *testing.T) {
	if isHexString("gggg") {
		t.Fatal("expected false for non-hex chars")
	}
}

func TestSprint43_IsHexString_AllZeros_True(t *testing.T) {
	if !isHexString("0000000000000000") {
		t.Fatal("expected true for all-zeros hex")
	}
}

func TestSprint43_IsHexString_UpperCase_True(t *testing.T) {
	if !isHexString("AABBCCDD") {
		t.Fatal("expected true for uppercase hex")
	}
}

// ---------------------------------------------------------------------------
// getStringFromMap
// ---------------------------------------------------------------------------

func TestSprint43_GetStringFromMap_ExistingKey(t *testing.T) {
	m := map[string]interface{}{"key": "value"}
	got := getStringFromMap(m, "key")
	if got != "value" {
		t.Fatalf("expected 'value', got %q", got)
	}
}

func TestSprint43_GetStringFromMap_MissingKey_Empty(t *testing.T) {
	m := map[string]interface{}{}
	got := getStringFromMap(m, "missing")
	if got != "" {
		t.Fatalf("expected empty string for missing key, got %q", got)
	}
}

func TestSprint43_GetStringFromMap_NonStringValue_Empty(t *testing.T) {
	m := map[string]interface{}{"key": 42}
	got := getStringFromMap(m, "key")
	if got != "" {
		t.Fatalf("expected empty string for non-string value, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// getUint64FromMap
// ---------------------------------------------------------------------------

func TestSprint43_GetUint64FromMap_Float64Value(t *testing.T) {
	m := map[string]interface{}{"height": float64(100)}
	got := getUint64FromMap(m, "height")
	if got != 100 {
		t.Fatalf("expected 100, got %d", got)
	}
}

func TestSprint43_GetUint64FromMap_StringValue(t *testing.T) {
	m := map[string]interface{}{"nonce": "42"}
	got := getUint64FromMap(m, "nonce")
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestSprint43_GetUint64FromMap_IntValue(t *testing.T) {
	m := map[string]interface{}{"count": 7}
	got := getUint64FromMap(m, "count")
	if got != 7 {
		t.Fatalf("expected 7, got %d", got)
	}
}

func TestSprint43_GetUint64FromMap_Uint64Value(t *testing.T) {
	m := map[string]interface{}{"val": uint64(999)}
	got := getUint64FromMap(m, "val")
	if got != 999 {
		t.Fatalf("expected 999, got %d", got)
	}
}

func TestSprint43_GetUint64FromMap_MissingKey_Zero(t *testing.T) {
	m := map[string]interface{}{}
	got := getUint64FromMap(m, "missing")
	if got != 0 {
		t.Fatalf("expected 0 for missing key, got %d", got)
	}
}

func TestSprint43_GetUint64FromMap_InvalidString_Zero(t *testing.T) {
	m := map[string]interface{}{"val": "not-a-number"}
	got := getUint64FromMap(m, "val")
	if got != 0 {
		t.Fatalf("expected 0 for invalid string, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// isHexEncodedGenesis
// ---------------------------------------------------------------------------

func TestSprint43_IsHexEncodedGenesis_Short_False(t *testing.T) {
	if isHexEncodedGenesis("abc") {
		t.Fatal("expected false for short string")
	}
}

func TestSprint43_IsHexEncodedGenesis_RegularHex_False(t *testing.T) {
	if isHexEncodedGenesis("aabbccddeeff0011") {
		t.Fatal("expected false for regular hex that doesn't encode GENESIS_")
	}
}

// ---------------------------------------------------------------------------
// GetChainStatePath — uses storage dir
// ---------------------------------------------------------------------------

func TestSprint43_GetChainStatePath_NonEmpty(t *testing.T) {
	dir, err := os.MkdirTemp("", "qtx-state43-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)
	s, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	defer s.Close()
	path := s.GetChainStatePath()
	if path == "" {
		t.Fatal("expected non-empty chain state path")
	}
}

// ---------------------------------------------------------------------------
// Storage.decodeHexField — various inputs
// ---------------------------------------------------------------------------

func TestSprint43_DecodeHexField_GenesisPrefix_ReturnsBytes(t *testing.T) {
	dir, err := os.MkdirTemp("", "qtx-state43b-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)
	s, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	defer s.Close()
	result := s.decodeHexField("GENESIS_abc123")
	if string(result) != "GENESIS_abc123" {
		t.Fatalf("expected GENESIS_ string returned as bytes, got %q", result)
	}
}

func TestSprint43_DecodeHexField_ValidHex_Decoded(t *testing.T) {
	dir, err := os.MkdirTemp("", "qtx-state43c-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)
	s, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	defer s.Close()
	// "aabb" decodes to []byte{0xAA, 0xBB}
	result := s.decodeHexField("aabb")
	if len(result) != 2 || result[0] != 0xAA || result[1] != 0xBB {
		t.Fatalf("expected [0xAA, 0xBB], got %v", result)
	}
}

func TestSprint43_DecodeHexField_InvalidHex_ReturnsAsBytes(t *testing.T) {
	dir, err := os.MkdirTemp("", "qtx-state43d-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)
	s, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	defer s.Close()
	result := s.decodeHexField("not-hex-content")
	// Invalid hex returns as raw bytes
	if string(result) != "not-hex-content" {
		t.Fatalf("expected raw bytes for invalid hex, got %q", result)
	}
}
