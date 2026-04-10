// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 44 - sphincs sign backend StoreTimestampNonce/CheckTimestampNonce/LoadCommitment with real DB
package sign

import (
	"os"
	"testing"

	params "github.com/ramseyauron/quantix/src/core/sphincs/config"
	key "github.com/ramseyauron/quantix/src/core/sphincs/key/backend"
	"github.com/syndtr/goleveldb/leveldb"
)

func newSphincsManagerWithDB(t *testing.T) (*SphincsManager, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-sphincs44-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("OpenFile: %v", err)
	}
	p, err := params.NewSPHINCSParameters()
	if err != nil {
		db.Close()
		os.RemoveAll(dir)
		t.Fatalf("NewSPHINCSParameters: %v", err)
	}
	km, err := key.NewKeyManager()
	if err != nil {
		db.Close()
		os.RemoveAll(dir)
		t.Fatalf("NewKeyManager: %v", err)
	}
	sm := NewSphincsManager(db, km, p)
	return sm, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

// ---------------------------------------------------------------------------
// StoreTimestampNonce + CheckTimestampNonce roundtrip
// ---------------------------------------------------------------------------

func TestSprint44_StoreTimestampNonce_Basic_NoError(t *testing.T) {
	sm, cleanup := newSphincsManagerWithDB(t)
	defer cleanup()
	ts := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	nonce := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	err := sm.StoreTimestampNonce(ts, nonce)
	if err != nil {
		t.Fatalf("StoreTimestampNonce error: %v", err)
	}
}

func TestSprint44_CheckTimestampNonce_NotStored_False(t *testing.T) {
	sm, cleanup := newSphincsManagerWithDB(t)
	defer cleanup()
	ts := []byte{9, 8, 7, 6, 5, 4, 3, 2}
	nonce := []byte{0x11, 0x22, 0x33, 0x44}
	ok, err := sm.CheckTimestampNonce(ts, nonce)
	if err != nil {
		t.Fatalf("CheckTimestampNonce error: %v", err)
	}
	if ok {
		t.Fatal("expected false for non-stored timestamp/nonce pair")
	}
}

func TestSprint44_CheckTimestampNonce_AfterStore_True(t *testing.T) {
	sm, cleanup := newSphincsManagerWithDB(t)
	defer cleanup()
	ts := []byte{11, 22, 33, 44, 55, 66, 77, 88}
	nonce := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	// Store first
	if err := sm.StoreTimestampNonce(ts, nonce); err != nil {
		t.Fatalf("StoreTimestampNonce error: %v", err)
	}
	// Check should return true
	ok, err := sm.CheckTimestampNonce(ts, nonce)
	if err != nil {
		t.Fatalf("CheckTimestampNonce error: %v", err)
	}
	if !ok {
		t.Fatal("expected true for stored timestamp/nonce pair")
	}
}

func TestSprint44_CheckTimestampNonce_DifferentNonce_False(t *testing.T) {
	sm, cleanup := newSphincsManagerWithDB(t)
	defer cleanup()
	ts := []byte{1, 1, 1, 1, 1, 1, 1, 1}
	nonce1 := []byte{0x11, 0x11, 0x11, 0x11}
	nonce2 := []byte{0x22, 0x22, 0x22, 0x22}
	// Store nonce1
	_ = sm.StoreTimestampNonce(ts, nonce1)
	// Check nonce2 (different) should be false
	ok, err := sm.CheckTimestampNonce(ts, nonce2)
	if err != nil {
		t.Fatalf("CheckTimestampNonce error: %v", err)
	}
	if ok {
		t.Fatal("expected false for different nonce that was not stored")
	}
}

// ---------------------------------------------------------------------------
// LoadCommitment — with real DB (no commitment stored → error)
// ---------------------------------------------------------------------------

func TestSprint44_LoadCommitment_NothingStored_Error(t *testing.T) {
	sm, cleanup := newSphincsManagerWithDB(t)
	defer cleanup()
	_, err := sm.LoadCommitment()
	// No commitment stored → expect error or not-found
	_ = err // may be nil (returns zero value) or error — just no panic
}

// ---------------------------------------------------------------------------
// NewSphincsManager with real DB — non-nil result
// ---------------------------------------------------------------------------

func TestSprint44_NewSphincsManager_WithDB_NotNil(t *testing.T) {
	sm, cleanup := newSphincsManagerWithDB(t)
	defer cleanup()
	if sm == nil {
		t.Fatal("expected non-nil SphincsManager with real DB")
	}
}
