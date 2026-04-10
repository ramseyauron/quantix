// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 32 - wallet/utils WalletConfig coverage
package utils

import (
	"testing"
)

// ---------------------------------------------------------------------------
// SaveKeyPair — nil argument checks (no DB needed)
// ---------------------------------------------------------------------------

func TestSprint32_SaveKeyPair_NilCombinedData_Error(t *testing.T) {
	// Use a WalletConfig with nil DB — nil-arg check fires before DB access
	wc := &WalletConfig{db: nil}
	err := wc.SaveKeyPair(nil, []byte("pk"))
	if err == nil {
		t.Fatal("expected error for nil combined data")
	}
}

func TestSprint32_SaveKeyPair_NilPK_Error(t *testing.T) {
	wc := &WalletConfig{db: nil}
	err := wc.SaveKeyPair([]byte("combined"), nil)
	if err == nil {
		t.Fatal("expected error for nil public key")
	}
}

func TestSprint32_SaveKeyPair_BothNil_Error(t *testing.T) {
	wc := &WalletConfig{db: nil}
	err := wc.SaveKeyPair(nil, nil)
	if err == nil {
		t.Fatal("expected error for both nil arguments")
	}
}

// ---------------------------------------------------------------------------
// GetDB — returns the stored db (even if nil)
// ---------------------------------------------------------------------------

func TestSprint32_GetDB_Nil(t *testing.T) {
	wc := &WalletConfig{db: nil}
	db := wc.GetDB()
	if db != nil {
		t.Fatal("expected nil DB from WalletConfig with nil db")
	}
}

// ---------------------------------------------------------------------------
// Recovery — error path with invalid passkey
// ---------------------------------------------------------------------------

func TestSprint32_Recovery_InvalidPasskey_Error(t *testing.T) {
	_, err := Recovery([]byte("message"), []string{"participant1"}, 1, "invalidpasskey!@#", "passphrase")
	if err == nil {
		t.Fatal("expected error for invalid passkey in Recovery")
	}
}

func TestSprint32_Recovery_EmptyPasskey_Error(t *testing.T) {
	_, err := Recovery([]byte("message"), []string{"participant1"}, 1, "", "passphrase")
	if err == nil {
		t.Fatal("expected error for empty passkey in Recovery")
	}
}

func TestSprint32_Recovery_EmptyParticipants_Error(t *testing.T) {
	_, err := Recovery([]byte("message"), []string{}, 0, "somepasskey", "passphrase")
	if err == nil {
		t.Fatal("expected error for empty participants list")
	}
}
