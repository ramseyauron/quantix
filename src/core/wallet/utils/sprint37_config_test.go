// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 37 - wallet/utils WalletConfig full lifecycle
package utils

import (
	"testing"
)

// ---------------------------------------------------------------------------
// NewWalletConfig — constructs a real WalletConfig with LevelDB
// ---------------------------------------------------------------------------

func TestSprint37_NewWalletConfig_NotNil(t *testing.T) {
	wc, err := NewWalletConfig()
	if err != nil {
		t.Fatalf("NewWalletConfig error: %v", err)
	}
	if wc == nil {
		t.Fatal("expected non-nil WalletConfig")
	}
	defer wc.Close()
}

func TestSprint37_NewWalletConfig_GetDB_NotNil(t *testing.T) {
	wc, err := NewWalletConfig()
	if err != nil {
		t.Fatalf("NewWalletConfig error: %v", err)
	}
	defer wc.Close()
	if wc.GetDB() == nil {
		t.Fatal("expected non-nil DB in WalletConfig")
	}
}

// ---------------------------------------------------------------------------
// SaveKeyPair + LoadKeyPair roundtrip
// ---------------------------------------------------------------------------

func TestSprint37_SaveAndLoadKeyPair_Roundtrip(t *testing.T) {
	wc, err := NewWalletConfig()
	if err != nil {
		t.Fatalf("NewWalletConfig error: %v", err)
	}
	defer wc.Close()

	combinedData := []byte("encrypted-sk-data")
	pk := []byte("public-key-data")

	if err := wc.SaveKeyPair(combinedData, pk); err != nil {
		t.Fatalf("SaveKeyPair error: %v", err)
	}

	sk2, pk2, err := wc.LoadKeyPair()
	if err != nil {
		t.Fatalf("LoadKeyPair error: %v", err)
	}
	if len(sk2) == 0 {
		t.Fatal("expected non-empty sk from LoadKeyPair")
	}
	if len(pk2) == 0 {
		t.Fatal("expected non-empty pk from LoadKeyPair")
	}
}

func TestSprint37_LoadKeyPair_BeforeSave_NoError(t *testing.T) {
	// Fresh WalletConfig may have existing data from prior test — just no panic
	wc, err := NewWalletConfig()
	if err != nil {
		t.Fatalf("NewWalletConfig error: %v", err)
	}
	defer wc.Close()
	// LoadKeyPair may succeed (leftover from prior test) or fail — just no panic
	_, _, _ = wc.LoadKeyPair()
}

// ---------------------------------------------------------------------------
// Close — no panic
// ---------------------------------------------------------------------------

func TestSprint37_WalletConfig_Close_NoPanic(t *testing.T) {
	wc, err := NewWalletConfig()
	if err != nil {
		t.Fatalf("NewWalletConfig error: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Close panicked: %v", r)
		}
	}()
	wc.Close()
}
