// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 39 - core/state DB.Has, LevelDBAdapter.Has, PutBatch large, ListKeysWithPrefix
package database

import (
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func newTestStatDB39(t *testing.T) (*DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-state39-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	db, err := NewLevelDB(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("NewLevelDB: %v", err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

// ---------------------------------------------------------------------------
// DB.Has — string key interface
// ---------------------------------------------------------------------------

func TestSprint39_DB_Has_AfterPut_True(t *testing.T) {
	db, cleanup := newTestStatDB39(t)
	defer cleanup()
	if err := db.Put("sprint39-key", []byte("value")); err != nil {
		t.Fatalf("Put error: %v", err)
	}
	ok, err := db.Has("sprint39-key")
	if err != nil {
		t.Fatalf("Has error: %v", err)
	}
	if !ok {
		t.Fatal("expected Has to return true for existing key")
	}
}

func TestSprint39_DB_Has_Missing_False(t *testing.T) {
	db, cleanup := newTestStatDB39(t)
	defer cleanup()
	ok, err := db.Has("sprint39-never-inserted")
	if err != nil {
		t.Fatalf("Has error: %v", err)
	}
	if ok {
		t.Fatal("expected Has to return false for missing key")
	}
}

func TestSprint39_DB_Has_AfterDelete_False(t *testing.T) {
	db, cleanup := newTestStatDB39(t)
	defer cleanup()
	_ = db.Put("delete-me", []byte("val"))
	_ = db.Delete("delete-me")
	ok, err := db.Has("delete-me")
	if err != nil {
		t.Fatalf("Has error: %v", err)
	}
	if ok {
		t.Fatal("expected Has to return false after delete")
	}
}

// ---------------------------------------------------------------------------
// LevelDBAdapter.Has — direct test via raw leveldb
// ---------------------------------------------------------------------------

func TestSprint39_LevelDBAdapter_Has_ExistingKey_True(t *testing.T) {
	dir, err := os.MkdirTemp("", "qtx-adapter39-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	rawDB, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer rawDB.Close()

	// Put a key via the raw db
	if err := rawDB.Put([]byte("adapter-key-39"), []byte("value"), nil); err != nil {
		t.Fatalf("rawDB.Put: %v", err)
	}

	// Now test LevelDBAdapter.Has
	adapter := &LevelDBAdapter{db: rawDB}
	ok, err := adapter.Has([]byte("adapter-key-39"), nil)
	if err != nil {
		t.Fatalf("LevelDBAdapter.Has error: %v", err)
	}
	if !ok {
		t.Fatal("expected LevelDBAdapter.Has to return true for existing key")
	}
}

func TestSprint39_LevelDBAdapter_Has_MissingKey_False(t *testing.T) {
	dir, err := os.MkdirTemp("", "qtx-adapter39b-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	rawDB, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer rawDB.Close()

	adapter := &LevelDBAdapter{db: rawDB}
	ok, err := adapter.Has([]byte("never-inserted-adapter"), nil)
	if err != nil {
		t.Fatalf("LevelDBAdapter.Has error: %v", err)
	}
	if ok {
		t.Fatal("expected false for missing key")
	}
}

// ---------------------------------------------------------------------------
// PutBatch — large batch
// ---------------------------------------------------------------------------

func TestSprint39_PutBatch_LargeMap_NoError(t *testing.T) {
	db, cleanup := newTestStatDB39(t)
	defer cleanup()
	batch := make(map[string][]byte)
	for i := 0; i < 50; i++ {
		batch[string([]byte{byte(i + 1)})] = []byte("value")
	}
	if err := db.PutBatch(batch); err != nil {
		t.Fatalf("PutBatch error: %v", err)
	}
}

func TestSprint39_PutBatch_ThenHas_True(t *testing.T) {
	db, cleanup := newTestStatDB39(t)
	defer cleanup()
	_ = db.PutBatch(map[string][]byte{"batch-key-39": []byte("val")})
	ok, err := db.Has("batch-key-39")
	if err != nil {
		t.Fatalf("Has error: %v", err)
	}
	if !ok {
		t.Fatal("expected batch-key to exist after PutBatch")
	}
}

// ---------------------------------------------------------------------------
// ListKeysWithPrefix
// ---------------------------------------------------------------------------

func TestSprint39_ListKeysWithPrefix_AfterPut_FindsKey(t *testing.T) {
	db, cleanup := newTestStatDB39(t)
	defer cleanup()
	_ = db.Put("pfx39:test-key", []byte("value"))
	keys, err := db.ListKeysWithPrefix("pfx39:")
	if err != nil {
		t.Fatalf("ListKeysWithPrefix error: %v", err)
	}
	found := false
	for _, k := range keys {
		if k == "pfx39:test-key" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find 'pfx39:test-key' in results, got: %v", keys)
	}
}

func TestSprint39_ListKeysWithPrefix_NoMatch_Empty(t *testing.T) {
	db, cleanup := newTestStatDB39(t)
	defer cleanup()
	keys, err := db.ListKeysWithPrefix("neverexistsprefix9876:")
	if err != nil {
		t.Fatalf("ListKeysWithPrefix error: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected empty, got %d keys", len(keys))
	}
}
