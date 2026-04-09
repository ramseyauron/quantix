// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 40 - validatorkey persist + LoadOrCreate additional paths
package validatorkey

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// persist — direct function test
// ---------------------------------------------------------------------------

func TestSprint40_Persist_Basic_NoError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-vk.json")
	vk := &ValidatorKey{
		Address:       "test-address-sprint40",
		PrivateKeyHex: "deadbeef",
		PublicKeyHex:  "cafebabe",
	}
	if err := persist(path, vk); err != nil {
		t.Fatalf("persist error: %v", err)
	}
	// Verify file was written
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after persist: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty file after persist")
	}
}

func TestSprint40_Persist_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip-vk.json")
	vk := &ValidatorKey{
		Address:       "roundtrip-address",
		PrivateKeyHex: "sk-bytes",
		PublicKeyHex:  "pk-bytes",
	}
	if err := persist(path, vk); err != nil {
		t.Fatalf("persist error: %v", err)
	}
	data, _ := os.ReadFile(path)
	var vk2 ValidatorKey
	if err := json.Unmarshal(data, &vk2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if vk2.Address != vk.Address {
		t.Fatalf("expected Address=%q, got %q", vk.Address, vk2.Address)
	}
}

func TestSprint40_Persist_CreatesDirectory_NoError(t *testing.T) {
	dir := t.TempDir()
	// Path with nested subdirectory that doesn't exist yet
	path := filepath.Join(dir, "subdir", "nested", "vk.json")
	vk := &ValidatorKey{Address: "test"}
	if err := persist(path, vk); err != nil {
		t.Fatalf("persist with nested dir error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// LoadOrCreate — existing key (from file)
// ---------------------------------------------------------------------------

func TestSprint40_LoadOrCreate_ExistingValidKey_Loads(t *testing.T) {
	dir := t.TempDir()
	// Create a validator key file first
	path := filepath.Join(dir, keyFileName)
	vk := &ValidatorKey{
		Address:       "existing-address-sprint40",
		PrivateKeyHex: "existing-sk",
		PublicKeyHex:  "existing-pk",
	}
	if err := persist(path, vk); err != nil {
		t.Fatalf("persist error: %v", err)
	}

	// Now load it
	loaded, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("LoadOrCreate error: %v", err)
	}
	if loaded.Address != "existing-address-sprint40" {
		t.Fatalf("expected address 'existing-address-sprint40', got %q", loaded.Address)
	}
}

func TestSprint40_LoadOrCreate_ExistingInvalidJSON_GeneratesNew(t *testing.T) {
	dir := t.TempDir()
	// Write invalid JSON
	path := filepath.Join(dir, keyFileName)
	if err := os.WriteFile(path, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Should generate new key (invalid JSON → generates fresh)
	vk, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("LoadOrCreate with invalid JSON: %v", err)
	}
	if vk.Address == "" {
		t.Fatal("expected non-empty address in generated key")
	}
}

func TestSprint40_LoadOrCreate_ExistingEmptyAddress_GeneratesNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, keyFileName)
	// Valid JSON but empty address (should trigger regeneration)
	vk := &ValidatorKey{Address: "", PrivateKeyHex: "sk", PublicKeyHex: "pk"}
	data, _ := json.MarshalIndent(vk, "", "  ")
	os.WriteFile(path, data, 0600)

	loaded, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if loaded.Address == "" {
		t.Fatal("expected non-empty address after regeneration")
	}
}
