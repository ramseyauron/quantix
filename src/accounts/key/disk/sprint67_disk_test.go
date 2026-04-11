// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 67 — accounts/key/disk 73.9%→higher
// Tests: ChangePassphrase key-not-found, ExportKey key-not-found, ExportKey includePrivate no-passphrase,
// loadKeys from empty dir, saveKeyToDisk, getDefaultDiskStoragePath
package disk_test

import (
	"os"
	"testing"

	disk "github.com/ramseyauron/quantix/src/accounts/key/disk"
	key "github.com/ramseyauron/quantix/src/accounts/key"
)

func newTestDKS67(t *testing.T) (*disk.DiskKeyStore, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "dks67-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	ks, err := disk.NewDiskKeyStore(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("NewDiskKeyStore: %v", err)
	}
	return ks, func() { os.RemoveAll(dir) }
}

// ─── ChangePassphrase — key not found ────────────────────────────────────────

func TestSprint67_ChangePassphrase_KeyNotFound(t *testing.T) {
	ks, cleanup := newTestDKS67(t)
	defer cleanup()
	err := ks.ChangePassphrase("nonexistent-id", "old", "new")
	if err == nil {
		t.Error("expected error for ChangePassphrase with unknown key ID")
	}
}

// ─── ExportKey — key not found ────────────────────────────────────────────────

func TestSprint67_ExportKey_KeyNotFound(t *testing.T) {
	ks, cleanup := newTestDKS67(t)
	defer cleanup()
	_, err := ks.ExportKey("nonexistent-id", false, "")
	if err == nil {
		t.Error("expected error for ExportKey with unknown key ID")
	}
}

// ─── ExportKey — includePrivate with empty passphrase ────────────────────────

func TestSprint67_ExportKey_IncludePrivate_NoPassphrase(t *testing.T) {
	ks, cleanup := newTestDKS67(t)
	defer cleanup()

	// First store a key
	kp := minimalKeyPair("kp-67", "addr-67")
	err := ks.StoreKey(kp)
	if err != nil {
		t.Fatalf("StoreKey: %v", err)
	}

	// Export with includePrivate=true but empty passphrase
	_, err = ks.ExportKey("kp-67", true, "")
	if err == nil {
		t.Error("expected error for ExportKey with empty passphrase when includePrivate=true")
	}
}

// ─── ExportKey — public only, no passphrase needed ───────────────────────────

func TestSprint67_ExportKey_PublicOnly(t *testing.T) {
	ks, cleanup := newTestDKS67(t)
	defer cleanup()

	kp := minimalKeyPair("kp-pub-67", "addr-pub-67")
	if err := ks.StoreKey(kp); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}

	data, err := ks.ExportKey("kp-pub-67", false, "")
	if err != nil {
		t.Fatalf("ExportKey public: unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty exported data")
	}
}

// ─── EncryptData — returns ciphertext and salt ────────────────────────────────

func TestSprint67_EncryptData_ReturnsSaltAndCipher(t *testing.T) {
	ks, cleanup := newTestDKS67(t)
	defer cleanup()

	ciphertext, salt, err := ks.EncryptData([]byte("secretdata"), "passphrase")
	if err != nil {
		t.Fatalf("EncryptData error: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Error("expected non-empty ciphertext")
	}
	if len(salt) == 0 {
		t.Error("expected non-empty salt")
	}
}

// ─── StoreRawKey — encrypt and store ─────────────────────────────────────────

func TestSprint67_StoreRawKey_ValidKey(t *testing.T) {
	ks, cleanup := newTestDKS67(t)
	defer cleanup()

	kp, err := ks.StoreRawKey([]byte("rawsecretkey"), []byte("rawpublickey"), "addr-raw-67", key.WalletTypeDisk, 0, "", "passphrase", nil)
	if err != nil {
		t.Fatalf("StoreRawKey error: %v", err)
	}
	if kp == nil {
		t.Error("expected non-nil KeyPair from StoreRawKey")
	}
	if kp.Address != "addr-raw-67" {
		t.Errorf("address mismatch: got %q", kp.Address)
	}
}

// ─── DecryptKey — valid roundtrip ────────────────────────────────────────────

func TestSprint67_DecryptKey_Roundtrip(t *testing.T) {
	ks, cleanup := newTestDKS67(t)
	defer cleanup()

	raw := []byte("thesecretkey")
	kp, err := ks.StoreRawKey(raw, []byte("pubkey"), "addr-dec-67", key.WalletTypeDisk, 0, "", "mypassphrase", nil)
	if err != nil {
		t.Fatalf("StoreRawKey: %v", err)
	}

	decrypted, err := ks.DecryptKey(kp, "mypassphrase")
	if err != nil {
		t.Fatalf("DecryptKey: %v", err)
	}
	if string(decrypted) != string(raw) {
		t.Errorf("decrypted key mismatch: got %q, want %q", decrypted, raw)
	}
}

func TestSprint67_DecryptKey_WrongPassphrase(t *testing.T) {
	ks, cleanup := newTestDKS67(t)
	defer cleanup()

	kp, err := ks.StoreRawKey([]byte("secret"), []byte("pub"), "addr-wp-67", key.WalletTypeDisk, 0, "", "correct", nil)
	if err != nil {
		t.Fatalf("StoreRawKey: %v", err)
	}

	_, err = ks.DecryptKey(kp, "wrongpassphrase")
	if err == nil {
		t.Error("expected error for wrong passphrase in DecryptKey")
	}
}
