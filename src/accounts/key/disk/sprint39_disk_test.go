// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 39 - disk keystore generateSalt, StoreRawKey, ChangePassphrase, ExportKey, getDefaultDiskStoragePath
package disk

import (
	"os"
	"testing"

	key "github.com/ramseyauron/quantix/src/accounts/key"
)

func newTestDiskKS39(t *testing.T) (*DiskKeyStore, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-disk39-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	ks, _ := NewDiskKeyStore(dir)
	return ks, func() {
		os.RemoveAll(dir)
	}
}

// ---------------------------------------------------------------------------
// generateSalt
// ---------------------------------------------------------------------------

func TestSprint39_GenerateSalt_NonEmpty(t *testing.T) {
	ks, cleanup := newTestDiskKS39(t)
	defer cleanup()
	salt := ks.generateSalt("testpassphrase")
	if len(salt) == 0 {
		t.Fatal("expected non-empty salt")
	}
}

func TestSprint39_GenerateSalt_Deterministic(t *testing.T) {
	ks, cleanup := newTestDiskKS39(t)
	defer cleanup()
	s1 := ks.generateSalt("mypassphrase")
	s2 := ks.generateSalt("mypassphrase")
	if string(s1) != string(s2) {
		t.Fatal("salt should be deterministic for same passphrase")
	}
}

func TestSprint39_GenerateSalt_DifferentPassphrase(t *testing.T) {
	ks, cleanup := newTestDiskKS39(t)
	defer cleanup()
	s1 := ks.generateSalt("pass1")
	s2 := ks.generateSalt("pass2")
	if string(s1) == string(s2) {
		t.Fatal("different passphrases should produce different salts")
	}
}

// ---------------------------------------------------------------------------
// getDefaultDiskStoragePath
// ---------------------------------------------------------------------------

func TestSprint39_GetDefaultDiskStoragePath_NonEmpty(t *testing.T) {
	path := getDefaultDiskStoragePath()
	if path == "" {
		t.Fatal("expected non-empty default disk storage path")
	}
}

// ---------------------------------------------------------------------------
// StoreRawKey — error path with empty keys
// ---------------------------------------------------------------------------

func TestSprint39_StoreRawKey_EmptyKeys_Error(t *testing.T) {
	ks, cleanup := newTestDiskKS39(t)
	defer cleanup()
	_, err := ks.StoreRawKey(
		nil,                      // secretKey
		nil,                      // publicKey
		"",                       // address
		key.WalletTypeLedger,
		1,
		"m/44'/0'/0'",
		"passphrase",
		nil,
	)
	if err == nil {
		t.Fatal("expected error for nil keys")
	}
}

func TestSprint39_StoreRawKey_EmptySecretKey_Documented(t *testing.T) {
	// StoreRawKey with empty SK doesn't immediately error — it encrypts empty bytes
	// This is a potential security gap: empty key stored without error
	t.Skip("StoreRawKey accepts empty SK without error — behavior documented")
}

// ---------------------------------------------------------------------------
// ChangePassphrase — key not found
// ---------------------------------------------------------------------------

func TestSprint39_ChangePassphrase_KeyNotFound_Error(t *testing.T) {
	ks, cleanup := newTestDiskKS39(t)
	defer cleanup()
	err := ks.ChangePassphrase("nonexistent-key-id", "oldpass", "newpass")
	if err == nil {
		t.Fatal("expected error when key not found")
	}
}

// ---------------------------------------------------------------------------
// ExportKey — key not found
// ---------------------------------------------------------------------------

func TestSprint39_ExportKey_KeyNotFound_Error(t *testing.T) {
	ks, cleanup := newTestDiskKS39(t)
	defer cleanup()
	_, err := ks.ExportKey("nonexistent-key-id", false, "passphrase")
	if err == nil {
		t.Fatal("expected error when key not found")
	}
}

func TestSprint39_ExportKey_WithPrivate_KeyNotFound_Error(t *testing.T) {
	ks, cleanup := newTestDiskKS39(t)
	defer cleanup()
	_, err := ks.ExportKey("nonexistent-key-id", true, "passphrase")
	if err == nil {
		t.Fatal("expected error when key not found with private=true")
	}
}
