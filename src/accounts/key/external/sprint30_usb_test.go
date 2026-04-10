// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 30 - accounts/key/external additional coverage
package usb

import (
	"testing"

	key "github.com/ramseyauron/quantix/src/accounts/key"
)

func TestStoreKey_NotMounted_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	kp := &key.KeyPair{ID: "test-kp"}
	err := ks.StoreKey(kp)
	if err == nil {
		t.Fatal("expected error when USB not mounted")
	}
}

func TestStoreEncryptedKey_NotMounted_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	_, err := ks.StoreEncryptedKey(
		[]byte("encsk"),
		[]byte("pk"),
		"0x1234",
		key.WalletTypeLedger,
		1,
		"m/44'/0'/0'",
		nil,
	)
	if err == nil {
		t.Fatal("expected error when USB not mounted")
	}
}

func TestChangePassphrase_NotMounted_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	err := ks.ChangePassphrase("key-id", "old", "new")
	if err == nil {
		t.Fatal("expected error when USB not mounted")
	}
}

func TestExportKey_NotMounted_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	_, err := ks.ExportKey("key-id", false, "passphrase")
	if err == nil {
		t.Fatal("expected error when USB not mounted")
	}
}

func TestDecryptKey_NotMounted_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	kp := &key.KeyPair{ID: "test-kp"}
	_, err := ks.DecryptKey(kp, "passphrase")
	if err == nil {
		t.Fatal("expected error when USB not mounted")
	}
}

func TestBackupFromDisk_NotMounted_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	// mock disk store that returns empty list
	mock := &mockDiskStore{}
	err := ks.BackupFromDisk(mock, "passphrase")
	if err == nil {
		t.Fatal("expected error when USB not mounted")
	}
}

func TestRestoreToDisk_NotMounted_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	mock := &mockDiskStore{}
	_, err := ks.RestoreToDisk(mock, "passphrase")
	if err == nil {
		t.Fatal("expected error when USB not mounted")
	}
}

func TestInitializeUSB_InvalidPath_NoPanel(t *testing.T) {
	ks := NewUSBKeyStore()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_ = ks.InitializeUSB("/tmp/nonexistent-usb-path-sprint30")
}

func TestGenerateKeyID_NonEmpty(t *testing.T) {
	ks := NewUSBKeyStore()
	id := ks.generateKeyID()
	if id == "" {
		t.Fatal("expected non-empty key ID")
	}
}

func TestGenerateKeyID_Unique(t *testing.T) {
	ks := NewUSBKeyStore()
	id1 := ks.generateKeyID()
	id2 := ks.generateKeyID()
	if id1 == id2 {
		t.Fatal("expected unique key IDs")
	}
}

func TestGenerateSalt_NonEmpty(t *testing.T) {
	ks := NewUSBKeyStore()
	salt := ks.generateSalt("mypassphrase")
	if len(salt) == 0 {
		t.Fatal("expected non-empty salt")
	}
}

func TestGenerateSalt_Deterministic(t *testing.T) {
	ks := NewUSBKeyStore()
	s1 := ks.generateSalt("testpassphrase")
	s2 := ks.generateSalt("testpassphrase")
	if string(s1) != string(s2) {
		t.Fatal("expected deterministic salt for same passphrase")
	}
}

func TestGenerateSalt_DifferentPassphrase_DifferentSalt(t *testing.T) {
	ks := NewUSBKeyStore()
	s1 := ks.generateSalt("passphrase1")
	s2 := ks.generateSalt("passphrase2")
	if string(s1) == string(s2) {
		t.Fatal("different passphrases should produce different salts")
	}
}

func TestGetWalletInfo_Fields(t *testing.T) {
	ks := NewUSBKeyStore()
	info := ks.GetWalletInfo()
	if info == nil {
		t.Fatal("expected non-nil WalletInfo")
	}
}

// mockDiskStore satisfies the interface required by BackupFromDisk/RestoreToDisk
type mockDiskStore struct{}

func (m *mockDiskStore) ListKeys() []*key.KeyPair { return nil }
func (m *mockDiskStore) StoreKey(kp *key.KeyPair) error { return nil }
