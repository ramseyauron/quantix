// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 35 - core/usi vault CreateVault/OpenVault roundtrip + deriveKey
package usi

import (
	"testing"
)

// ---------------------------------------------------------------------------
// deriveKey — internal function tests
// ---------------------------------------------------------------------------

func TestSprint35_DeriveKey_Basic_NonNil(t *testing.T) {
	k, err := deriveKey("passphrase", "creatorFP", []string{"recipient1"})
	if err != nil {
		t.Fatalf("deriveKey error: %v", err)
	}
	if len(k) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(k))
	}
}

func TestSprint35_DeriveKey_Deterministic(t *testing.T) {
	k1, err1 := deriveKey("pass", "fp", []string{"r1", "r2"})
	k2, err2 := deriveKey("pass", "fp", []string{"r1", "r2"})
	if err1 != nil || err2 != nil {
		t.Fatalf("deriveKey errors: %v / %v", err1, err2)
	}
	for i := range k1 {
		if k1[i] != k2[i] {
			t.Fatal("deriveKey should be deterministic")
		}
	}
}

func TestSprint35_DeriveKey_RecipientsOrderIndependent(t *testing.T) {
	k1, _ := deriveKey("pass", "fp", []string{"alice", "bob"})
	k2, _ := deriveKey("pass", "fp", []string{"bob", "alice"})
	for i := range k1 {
		if k1[i] != k2[i] {
			t.Fatal("deriveKey should be order-independent (SEC-V03)")
		}
	}
}

func TestSprint35_DeriveKey_DifferentPassphrase_DifferentKey(t *testing.T) {
	k1, _ := deriveKey("passA", "fp", []string{"r1"})
	k2, _ := deriveKey("passB", "fp", []string{"r1"})
	same := true
	for i := range k1 {
		if k1[i] != k2[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("different passphrases should produce different keys")
	}
}

// ---------------------------------------------------------------------------
// CreateVault — error paths
// ---------------------------------------------------------------------------

func TestSprint35_CreateVault_EmptyData_Error(t *testing.T) {
	_, err := CreateVault(nil, "creatorFP", []string{"recipient"}, "pass")
	if err == nil {
		t.Fatal("expected error for nil data")
	}
}

func TestSprint35_CreateVault_EmptyData_Slice_Error(t *testing.T) {
	_, err := CreateVault([]byte{}, "creatorFP", []string{"recipient"}, "pass")
	if err == nil {
		t.Fatal("expected error for empty data slice")
	}
}

func TestSprint35_CreateVault_EmptyCreatorFP_Error(t *testing.T) {
	_, err := CreateVault([]byte("data"), "", []string{"recipient"}, "pass")
	if err == nil {
		t.Fatal("expected error for empty creator fingerprint")
	}
}

func TestSprint35_CreateVault_EmptyRecipients_Error(t *testing.T) {
	_, err := CreateVault([]byte("data"), "creatorFP", nil, "pass")
	if err == nil {
		t.Fatal("expected error for nil recipients")
	}
}

func TestSprint35_CreateVault_EmptyPassphrase_Error(t *testing.T) {
	_, err := CreateVault([]byte("data"), "creatorFP", []string{"recipient"}, "")
	if err == nil {
		t.Fatal("expected error for empty passphrase")
	}
}

// ---------------------------------------------------------------------------
// CreateVault + OpenVault — roundtrip
// ---------------------------------------------------------------------------

func TestSprint35_CreateOpenVault_Roundtrip(t *testing.T) {
	data := []byte("hello quantix vault")
	creatorFP := "creator-fp-test"
	recipientFPs := []string{creatorFP, "other-recipient"}
	passphrase := "test-passphrase"

	vault, err := CreateVault(data, creatorFP, recipientFPs, passphrase)
	if err != nil {
		t.Fatalf("CreateVault error: %v", err)
	}
	if vault == nil {
		t.Fatal("expected non-nil vault")
	}

	// Open as creator
	decrypted, err := OpenVault(vault, creatorFP, passphrase)
	if err != nil {
		t.Fatalf("OpenVault error: %v", err)
	}
	if string(decrypted) != string(data) {
		t.Fatalf("decrypted data mismatch: got %q, want %q", decrypted, data)
	}
}

func TestSprint35_CreateOpenVault_WrongPassphrase_Error(t *testing.T) {
	vault, _ := CreateVault([]byte("secret"), "fp1", []string{"fp1"}, "correct")
	_, err := OpenVault(vault, "fp1", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
}

func TestSprint35_CreateOpenVault_UnauthorizedRecipient_Error(t *testing.T) {
	vault, _ := CreateVault([]byte("secret"), "fp1", []string{"fp1"}, "pass")
	_, err := OpenVault(vault, "fp-unauthorized", "pass")
	if err == nil {
		t.Fatal("expected error for unauthorized recipient")
	}
}

func TestSprint35_OpenVault_NilVault_Error(t *testing.T) {
	_, err := OpenVault(nil, "fp1", "pass")
	if err == nil {
		t.Fatal("expected error for nil vault")
	}
}

func TestSprint35_CreateVault_NonNilResult_Fields(t *testing.T) {
	vault, err := CreateVault([]byte("data"), "fp1", []string{"fp1"}, "pass")
	if err != nil {
		t.Fatalf("CreateVault error: %v", err)
	}
	if vault.Creator != "fp1" {
		t.Fatalf("expected Creator=fp1, got %q", vault.Creator)
	}
	if vault.EncData == "" {
		t.Fatal("expected non-empty EncData")
	}
	if vault.Version == "" {
		t.Fatal("expected non-empty Version")
	}
}
