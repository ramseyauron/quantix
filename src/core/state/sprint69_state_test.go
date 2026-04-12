// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 69 — core/state 74.4%→higher
// Tests: SetBalance nil amount, GetNonce fresh, IncrementNonce, StateRoot empty,
// PutBatch fallback path (LevelDBAdapter), ListKeysWithPrefix
package database

import (
	"math/big"
	"os"
	"testing"
)

func newTestDB69(t *testing.T) (*DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-state69-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	db, err := NewLevelDB(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("NewLevelDB: %v", err)
	}
	return db, func() {
		_ = db.Close()
		os.RemoveAll(dir)
	}
}

// ─── SetBalance — nil amount treated as zero ──────────────────────────────────

func TestSprint69_SetBalance_NilAmount(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	err := SetBalance(db, "alice69", nil)
	if err != nil {
		t.Fatalf("SetBalance with nil amount: %v", err)
	}

	// Balance should be 0 after nil set
	bal, err := GetBalance(db, "alice69")
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if bal.Sign() != 0 {
		t.Errorf("expected balance 0 after nil SetBalance, got %v", bal)
	}
}

func TestSprint69_SetBalance_PositiveAmount(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	err := SetBalance(db, "bob69", big.NewInt(999))
	if err != nil {
		t.Fatalf("SetBalance: %v", err)
	}

	bal, err := GetBalance(db, "bob69")
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if bal.Cmp(big.NewInt(999)) != 0 {
		t.Errorf("expected 999, got %v", bal)
	}
}

// ─── GetNonce — fresh account returns 0 ──────────────────────────────────────

func TestSprint69_GetNonce_FreshAccount(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	nonce, err := GetNonce(db, "carol69")
	if err != nil {
		t.Fatalf("GetNonce: %v", err)
	}
	if nonce != 0 {
		t.Errorf("expected 0 for fresh account nonce, got %d", nonce)
	}
}

// ─── IncrementNonce — multiple increments ────────────────────────────────────

func TestSprint69_IncrementNonce_Multiple(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	addr := "dave69"
	for i := 1; i <= 3; i++ {
		if err := IncrementNonce(db, addr); err != nil {
			t.Fatalf("IncrementNonce %d: %v", i, err)
		}
	}

	nonce, err := GetNonce(db, addr)
	if err != nil {
		t.Fatalf("GetNonce: %v", err)
	}
	if nonce != 3 {
		t.Errorf("expected nonce 3 after 3 increments, got %d", nonce)
	}
}

// ─── StateRoot — empty db returns non-error ───────────────────────────────────

func TestSprint69_StateRoot_EmptyDB(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	root, err := StateRoot(db)
	if err != nil {
		t.Fatalf("StateRoot on empty DB: %v", err)
	}
	if len(root) == 0 {
		t.Error("expected non-empty StateRoot even for empty DB")
	}
}

func TestSprint69_StateRoot_WithBalance(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	// Set a balance to change state root
	if err := SetBalance(db, "eve69", big.NewInt(5000)); err != nil {
		t.Fatalf("SetBalance: %v", err)
	}

	root1, _ := StateRoot(db)

	// Change balance
	if err := SetBalance(db, "frank69", big.NewInt(1000)); err != nil {
		t.Fatalf("SetBalance: %v", err)
	}

	root2, _ := StateRoot(db)
	if string(root1) == string(root2) {
		t.Error("expected different StateRoot after adding new account")
	}
}

// ─── PutBatch — with LevelDB (real batch path) ───────────────────────────────

func TestSprint69_PutBatch_MultipleFlatEntries(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	batch := map[string][]byte{
		"key-a-69": []byte("value-a"),
		"key-b-69": []byte("value-b"),
		"key-c-69": []byte("value-c"),
	}
	err := db.PutBatch(batch)
	if err != nil {
		t.Fatalf("PutBatch error: %v", err)
	}

	// Verify all keys were stored
	for k, v := range batch {
		got, err := db.Get(k)
		if err != nil {
			t.Errorf("Get(%q): %v", k, err)
			continue
		}
		if string(got) != string(v) {
			t.Errorf("Get(%q) = %q, want %q", k, got, v)
		}
	}
}

// ─── ListKeysWithPrefix — after storing ──────────────────────────────────────

func TestSprint69_ListKeysWithPrefix_AfterStore(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	// Store keys with a specific prefix
	_ = db.Put("sprint69:a", []byte("va"))
	_ = db.Put("sprint69:b", []byte("vb"))
	_ = db.Put("other:x", []byte("vx"))

	keys, err := db.ListKeysWithPrefix("sprint69:")
	if err != nil {
		t.Fatalf("ListKeysWithPrefix error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys with prefix sprint69:, got %d", len(keys))
	}
}

func TestSprint69_ListKeysWithPrefix_EmptyPrefix(t *testing.T) {
	db, cleanup := newTestDB69(t)
	defer cleanup()

	// Empty prefix should return all keys
	_ = db.Put("x69", []byte("1"))
	_ = db.Put("y69", []byte("2"))

	keys, err := db.ListKeysWithPrefix("")
	if err != nil {
		t.Fatalf("ListKeysWithPrefix empty: %v", err)
	}
	if len(keys) < 2 {
		t.Errorf("expected at least 2 keys for empty prefix, got %d", len(keys))
	}
}
