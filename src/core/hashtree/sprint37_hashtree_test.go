// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 37 - hashtree DB operations coverage
package hashtree

import (
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func openTestDB(t *testing.T) (*leveldb.DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-ht37-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("OpenFile: %v", err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

// ---------------------------------------------------------------------------
// SaveLeavesToDB + FetchLeafFromDB roundtrip
// ---------------------------------------------------------------------------

func TestSprint37_SaveLeavesToDB_Basic_NoError(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	leaves := [][]byte{[]byte("leaf0"), []byte("leaf1"), []byte("leaf2")}
	err := SaveLeavesToDB(db, leaves)
	if err != nil {
		t.Fatalf("SaveLeavesToDB error: %v", err)
	}
}

func TestSprint37_FetchLeafFromDB_AfterSave(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	leaves := [][]byte{[]byte("leaf0"), []byte("leaf1")}
	_ = SaveLeavesToDB(db, leaves)

	got, err := FetchLeafFromDB(db, "leaf-0")
	if err != nil {
		t.Fatalf("FetchLeafFromDB error: %v", err)
	}
	if string(got) != "leaf0" {
		t.Fatalf("expected 'leaf0', got %q", got)
	}
}

func TestSprint37_FetchLeafFromDB_UnknownKey_Error(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	_, err := FetchLeafFromDB(db, "nonexistent-key")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestSprint37_SaveLeavesToDB_EmptySlice_NoError(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	err := SaveLeavesToDB(db, [][]byte{})
	if err != nil {
		t.Fatalf("SaveLeavesToDB with empty slice error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PruneOldLeaves
// ---------------------------------------------------------------------------

func TestSprint37_PruneOldLeaves_AfterSave_NoError(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	leaves := [][]byte{[]byte("leaf0"), []byte("leaf1"), []byte("leaf2")}
	_ = SaveLeavesToDB(db, leaves)
	err := PruneOldLeaves(db, 3)
	if err != nil {
		t.Fatalf("PruneOldLeaves error: %v", err)
	}
}

func TestSprint37_PruneOldLeaves_EmptyDB_NoError(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	// Pruning non-existent keys should be fine (ErrNotFound is ignored)
	err := PruneOldLeaves(db, 5)
	if err != nil {
		t.Fatalf("PruneOldLeaves on empty DB error: %v", err)
	}
}

func TestSprint37_PruneOldLeaves_Zero_NoError(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	err := PruneOldLeaves(db, 0)
	if err != nil {
		t.Fatalf("PruneOldLeaves(0) error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// SaveLeavesBatchToDB
// ---------------------------------------------------------------------------

func TestSprint37_SaveLeavesBatchToDB_Basic_NoError(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	err := SaveLeavesBatchToDB(db, leaves)
	if err != nil {
		t.Fatalf("SaveLeavesBatchToDB error: %v", err)
	}
}

func TestSprint37_SaveLeavesBatchToDB_Empty_NoError(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	err := SaveLeavesBatchToDB(db, nil)
	if err != nil {
		t.Fatalf("SaveLeavesBatchToDB nil error: %v", err)
	}
}

func TestSprint37_SaveLeavesBatchToDB_ThenFetch(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	leaves := [][]byte{[]byte("batch-leaf")}
	_ = SaveLeavesBatchToDB(db, leaves)
	got, err := FetchLeafFromDB(db, "leaf-0")
	if err != nil {
		t.Fatalf("FetchLeafFromDB after batch save: %v", err)
	}
	if string(got) != "batch-leaf" {
		t.Fatalf("expected 'batch-leaf', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// FetchLeafConcurrent
// ---------------------------------------------------------------------------

func TestSprint37_FetchLeafConcurrent_AfterSave(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	_ = SaveLeavesToDB(db, [][]byte{[]byte("concurrent")})
	got, err := FetchLeafConcurrent(db, "leaf-0")
	if err != nil {
		t.Fatalf("FetchLeafConcurrent error: %v", err)
	}
	if string(got) != "concurrent" {
		t.Fatalf("expected 'concurrent', got %q", got)
	}
}

func TestSprint37_FetchLeafConcurrent_UnknownKey_Error(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	_, err := FetchLeafConcurrent(db, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown key in concurrent fetch")
	}
}

// ---------------------------------------------------------------------------
// setMaxFileSize (package-internal)
// ---------------------------------------------------------------------------

func TestSprint37_SetMaxFileSize_PositiveValue_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("setMaxFileSize panicked: %v", r)
		}
	}()
	setMaxFileSize(2)
}

func TestSprint37_SetMaxFileSize_ZeroValue_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("setMaxFileSize panicked with 0: %v", r)
		}
	}()
	setMaxFileSize(0)
}

func TestSprint37_SetMaxFileSize_NegativeValue_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("setMaxFileSize panicked with negative: %v", r)
		}
	}()
	setMaxFileSize(-1)
}
