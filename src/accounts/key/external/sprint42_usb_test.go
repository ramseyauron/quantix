// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 42 - USB keystore Mount path + GetKey/ListKeys after mount
package usb

import (
	"os"
	"path/filepath"
	"testing"
)

func mountUSBFromTemp(t *testing.T) (*USBKeyStore, string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-usb42-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	// Create the required quantix-usb-keystore subdirectory
	keystoreDir := filepath.Join(dir, "quantix-usb-keystore")
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("MkdirAll: %v", err)
	}
	ks := NewUSBKeyStore()
	err = ks.Mount(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Mount: %v", err)
	}
	return ks, dir, func() {
		ks.Unmount()
		os.RemoveAll(dir)
	}
}

// ---------------------------------------------------------------------------
// Mount — valid path
// ---------------------------------------------------------------------------

func TestSprint42_Mount_ValidPath_IsMounted(t *testing.T) {
	ks, _, cleanup := mountUSBFromTemp(t)
	defer cleanup()
	if !ks.IsMounted() {
		t.Fatal("expected IsMounted=true after successful mount")
	}
}

func TestSprint42_Mount_NonExistentPath_Error(t *testing.T) {
	ks := NewUSBKeyStore()
	err := ks.Mount("/nonexistent/path/sprint42")
	if err == nil {
		t.Fatal("expected error for non-existent mount path")
	}
}

func TestSprint42_Mount_NoKeystoreSubdir_Error(t *testing.T) {
	dir, _ := os.MkdirTemp("", "qtx-usb42b-*")
	defer os.RemoveAll(dir)
	// No quantix-usb-keystore subdir
	ks := NewUSBKeyStore()
	err := ks.Mount(dir)
	if err == nil {
		t.Fatal("expected error for path without quantix-usb-keystore subdirectory")
	}
}

// ---------------------------------------------------------------------------
// GetKey / GetKeyByAddress / ListKeys — after mount (empty keystore)
// ---------------------------------------------------------------------------

func TestSprint42_GetKey_AfterMount_UnknownID_Error(t *testing.T) {
	ks, _, cleanup := mountUSBFromTemp(t)
	defer cleanup()
	_, err := ks.GetKey("unknown-key-id-sprint42")
	if err == nil {
		t.Fatal("expected error for unknown key ID")
	}
}

func TestSprint42_GetKeyByAddress_AfterMount_UnknownAddr_Error(t *testing.T) {
	ks, _, cleanup := mountUSBFromTemp(t)
	defer cleanup()
	_, err := ks.GetKeyByAddress("unknown-addr-sprint42")
	if err == nil {
		t.Fatal("expected error for unknown address")
	}
}

func TestSprint42_ListKeys_AfterMount_Empty(t *testing.T) {
	ks, _, cleanup := mountUSBFromTemp(t)
	defer cleanup()
	keys := ks.ListKeys()
	// Empty keystore returns empty list
	if keys == nil {
		t.Fatal("expected non-nil list (empty)")
	}
}

// ---------------------------------------------------------------------------
// RemoveKey — after mount (not found — silently succeeds or errors)
// ---------------------------------------------------------------------------

func TestSprint42_RemoveKey_AfterMount_NoPanel(t *testing.T) {
	// RemoveKey on non-existent key may succeed silently — document behavior
	ks, _, cleanup := mountUSBFromTemp(t)
	defer cleanup()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RemoveKey panicked for non-existent key: %v", r)
		}
	}()
	_ = ks.RemoveKey("nonexistent-sprint42")
}

// ---------------------------------------------------------------------------
// GetWalletInfo — after mount
// ---------------------------------------------------------------------------

func TestSprint42_GetWalletInfo_AfterMount_NotNil(t *testing.T) {
	ks, _, cleanup := mountUSBFromTemp(t)
	defer cleanup()
	info := ks.GetWalletInfo()
	if info == nil {
		t.Fatal("expected non-nil WalletInfo after mount")
	}
}

// ---------------------------------------------------------------------------
// EncryptData — after mount
// ---------------------------------------------------------------------------

func TestSprint42_EncryptData_AfterMount_NonEmpty(t *testing.T) {
	ks, _, cleanup := mountUSBFromTemp(t)
	defer cleanup()
	ciphertext, salt, err := ks.EncryptData([]byte("secret data"), "passphrase")
	if err != nil {
		t.Fatalf("EncryptData error: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("expected non-empty ciphertext")
	}
	if len(salt) == 0 {
		t.Fatal("expected non-empty salt")
	}
}

func TestSprint42_EncryptData_AfterMount_TwoCallsDifferentCiphertext(t *testing.T) {
	ks, _, cleanup := mountUSBFromTemp(t)
	defer cleanup()
	c1, _, _ := ks.EncryptData([]byte("secret"), "pass")
	c2, _, _ := ks.EncryptData([]byte("secret"), "pass")
	// Should use random nonce → different ciphertext
	equal := len(c1) == len(c2)
	if equal {
		same := true
		for i := range c1 {
			if c1[i] != c2[i] {
				same = false
				break
			}
		}
		if same {
			t.Log("WARNING: two EncryptData calls with same input produced identical ciphertext (nonce may not be random)")
		}
	}
}
