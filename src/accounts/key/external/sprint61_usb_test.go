// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 61 — accounts/key/external 55.2%→higher
// Tests: validateKeyPair, saveKeyToUSB via StoreKey, ChangePassphrase not-mounted,
// ExportKey not-mounted, BackupFromDisk not-mounted, StoreKey error paths
package usb

import (
	"os"
	"path/filepath"
	"testing"

	key "github.com/ramseyauron/quantix/src/accounts/key"
)

func mountedKS61(t *testing.T) (*USBKeyStore, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-usb61-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	keystoreDir := filepath.Join(dir, "quantix-usb-keystore")
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("MkdirAll: %v", err)
	}
	ks := NewUSBKeyStore()
	if err := ks.Mount(dir); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Mount: %v", err)
	}
	return ks, func() {
		ks.Unmount()
		os.RemoveAll(dir)
	}
}

// ─── validateKeyPair — error paths ────────────────────────────────────────────

func TestSprint61_ValidateKeyPair_EmptyID(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	kp := &key.KeyPair{ID: "", EncryptedSK: []byte("sk"), PublicKey: []byte("pk"), Address: "addr"}
	err := ks.validateKeyPair(kp)
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestSprint61_ValidateKeyPair_EmptyEncryptedSK(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	kp := &key.KeyPair{ID: "id1", EncryptedSK: nil, PublicKey: []byte("pk"), Address: "addr"}
	err := ks.validateKeyPair(kp)
	if err == nil {
		t.Error("expected error for empty encrypted SK")
	}
}

func TestSprint61_ValidateKeyPair_EmptyPublicKey(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	kp := &key.KeyPair{ID: "id1", EncryptedSK: []byte("sk"), PublicKey: nil, Address: "addr"}
	err := ks.validateKeyPair(kp)
	if err == nil {
		t.Error("expected error for empty public key")
	}
}

func TestSprint61_ValidateKeyPair_EmptyAddress(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	kp := &key.KeyPair{ID: "id1", EncryptedSK: []byte("sk"), PublicKey: []byte("pk"), Address: ""}
	err := ks.validateKeyPair(kp)
	if err == nil {
		t.Error("expected error for empty address")
	}
}

func TestSprint61_ValidateKeyPair_Valid(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	kp := &key.KeyPair{ID: "id1", EncryptedSK: []byte("sk"), PublicKey: []byte("pk"), Address: "addr1"}
	err := ks.validateKeyPair(kp)
	if err != nil {
		t.Errorf("unexpected error for valid key pair: %v", err)
	}
}

// ─── StoreKey — not mounted ────────────────────────────────────────────────────

func TestSprint61_StoreKey_NotMounted_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	kp := &key.KeyPair{ID: "id1", EncryptedSK: []byte("sk"), PublicKey: []byte("pk"), Address: "addr"}
	err := ks.StoreKey(kp)
	if err == nil {
		t.Error("expected error for StoreKey on unmounted device")
	}
}

// ─── StoreKey — mounted, valid key pair (hits saveKeyToUSB) ──────────────────

func TestSprint61_StoreKey_Mounted_Valid(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	kp := &key.KeyPair{
		ID:          "test-id-61",
		EncryptedSK: []byte("encrypted-secret-key-bytes"),
		PublicKey:   []byte("public-key-bytes"),
		Address:     "addr61",
	}
	err := ks.StoreKey(kp)
	if err != nil {
		t.Errorf("StoreKey on mounted device: unexpected error: %v", err)
	}
}

// ─── ChangePassphrase — not mounted ───────────────────────────────────────────

func TestSprint61_ChangePassphrase_NotMounted(t *testing.T) {
	ks := NewUSBKeyStore()
	err := ks.ChangePassphrase("id1", "old", "new")
	if err == nil {
		t.Error("expected error for ChangePassphrase on unmounted device")
	}
}

// ─── ChangePassphrase — key not found ─────────────────────────────────────────

func TestSprint61_ChangePassphrase_KeyNotFound(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	err := ks.ChangePassphrase("nonexistent-id", "old", "new")
	if err == nil {
		t.Error("expected error for ChangePassphrase with unknown key ID")
	}
}

// ─── ExportKey — not mounted ──────────────────────────────────────────────────

func TestSprint61_ExportKey_NotMounted(t *testing.T) {
	ks := NewUSBKeyStore()
	_, err := ks.ExportKey("id1", true, "passphrase")
	if err == nil {
		t.Error("expected error for ExportKey on unmounted device")
	}
}

// ─── ExportKey — key not found ────────────────────────────────────────────────

func TestSprint61_ExportKey_KeyNotFound(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	_, err := ks.ExportKey("nonexistent-id", true, "passphrase")
	if err == nil {
		t.Error("expected error for ExportKey with unknown key ID")
	}
}

// ─── BackupFromDisk — not mounted ─────────────────────────────────────────────

func TestSprint61_BackupFromDisk_NotMounted(t *testing.T) {
	ks := NewUSBKeyStore()
	err := ks.BackupFromDisk(nil, "passphrase")
	if err == nil {
		t.Error("expected error for BackupFromDisk on unmounted device")
	}
}

// ─── RestoreToDisk — not mounted ──────────────────────────────────────────────

func TestSprint61_RestoreToDisk_NotMounted(t *testing.T) {
	ks := NewUSBKeyStore()
	_, err := ks.RestoreToDisk(nil, "passphrase")
	if err == nil {
		t.Error("expected error for RestoreToDisk on unmounted device")
	}
}

// ─── GetKeyByAddress — mounted, not found ────────────────────────────────────

func TestSprint61_GetKeyByAddress_NotFound(t *testing.T) {
	ks, cleanup := mountedKS61(t)
	defer cleanup()
	_, err := ks.GetKeyByAddress("nonexistent-addr")
	if err == nil {
		t.Error("expected error for GetKeyByAddress with unknown address")
	}
}
