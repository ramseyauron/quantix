// Sprint 26c — core/state NewLevelDB, Put/Get/Delete/Has coverage
package database

import (
	"os"
	"testing"
)

// ─── NewLevelDB ───────────────────────────────────────────────────────────────

func TestSprint26_NewLevelDB_ValidPath_NotNil(t *testing.T) {
	dir, _ := os.MkdirTemp("", "qtx-ldb26-*")
	defer os.RemoveAll(dir)

	db, err := NewLevelDB(dir + "/testdb")
	if err != nil {
		t.Fatalf("NewLevelDB error: %v", err)
	}
	if db == nil {
		t.Fatal("NewLevelDB returned nil")
	}
	db.Close()
}

func TestSprint26_NewLevelDB_Reopen_SameData(t *testing.T) {
	dir, _ := os.MkdirTemp("", "qtx-ldb26-reopen-*")
	defer os.RemoveAll(dir)
	path := dir + "/testdb"

	db1, err := NewLevelDB(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	db1.Put("key1", []byte("value1"))
	db1.Close()

	db2, err := NewLevelDB(path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer db2.Close()

	val, err := db2.Get("key1")
	if err != nil {
		t.Fatalf("Get after reopen: %v", err)
	}
	if string(val) != "value1" {
		t.Errorf("value = %q, want %q", val, "value1")
	}
}

func TestSprint26_NewLevelDB_Has_TrueAfterPut(t *testing.T) {
	dir, _ := os.MkdirTemp("", "qtx-ldb26-has-*")
	defer os.RemoveAll(dir)

	db, err := NewLevelDB(dir + "/testdb")
	if err != nil {
		t.Fatalf("NewLevelDB: %v", err)
	}
	defer db.Close()

	db.Put("existkey", []byte("v"))
	has, err := db.Has("existkey")
	if err != nil {
		t.Fatalf("Has: %v", err)
	}
	if !has {
		t.Error("expected Has=true after Put")
	}
}

func TestSprint26_NewLevelDB_Has_FalseForMissing(t *testing.T) {
	dir, _ := os.MkdirTemp("", "qtx-ldb26-hasmiss-*")
	defer os.RemoveAll(dir)

	db, err := NewLevelDB(dir + "/testdb")
	if err != nil {
		t.Fatalf("NewLevelDB: %v", err)
	}
	defer db.Close()

	has, err := db.Has("nosuchkey")
	if err != nil {
		t.Fatalf("Has: %v", err)
	}
	if has {
		t.Error("expected Has=false for missing key")
	}
}

func TestSprint26_NewLevelDB_Delete_AfterPut(t *testing.T) {
	dir, _ := os.MkdirTemp("", "qtx-ldb26-del-*")
	defer os.RemoveAll(dir)

	db, err := NewLevelDB(dir + "/testdb")
	if err != nil {
		t.Fatalf("NewLevelDB: %v", err)
	}
	defer db.Close()

	db.Put("delkey", []byte("v"))
	db.Delete("delkey")
	has, _ := db.Has("delkey")
	if has {
		t.Error("key should not exist after Delete")
	}
}
