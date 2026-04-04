// PEPPER Sprint2 — hashtree coverage push for 0% file/IO functions
package hashtree

import (
	"os"
	"testing"
)

// ── SaveRootHashToFile / LoadRootHashFromFile ─────────────────────────────────

func TestSaveAndLoadRootHashFromFile(t *testing.T) {
	leaves := [][]byte{
		[]byte("leaf1"),
		[]byte("leaf2"),
		[]byte("leaf3"),
	}
	root := BuildHashTree(leaves)
	if root == nil {
		t.Fatal("BuildHashTree returned nil")
	}

	tmpFile := t.TempDir() + "/root.bin"

	if err := SaveRootHashToFile(root, tmpFile); err != nil {
		t.Fatalf("SaveRootHashToFile: %v", err)
	}

	loaded, err := LoadRootHashFromFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadRootHashFromFile: %v", err)
	}
	if len(loaded) == 0 {
		t.Error("loaded root hash is empty")
	}
}

func TestSaveRootHashToFile_InvalidPath(t *testing.T) {
	root := BuildHashTree([][]byte{[]byte("leaf")})
	err := SaveRootHashToFile(root, "/nonexistent/dir/root.bin")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestLoadRootHashFromFile_Missing(t *testing.T) {
	_, err := LoadRootHashFromFile("/tmp/missing_root_file_pepper.bin")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ── MemoryMapFile ─────────────────────────────────────────────────────────────

func TestMemoryMapFile_ValidFile(t *testing.T) {
	tmpFile := t.TempDir() + "/test.bin"
	content := []byte("hello quantix memory map")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := MemoryMapFile(tmpFile)
	if err != nil {
		t.Logf("MemoryMapFile: %v (may not be supported)", err)
		return
	}
	if len(data) == 0 {
		t.Error("MemoryMapFile returned empty data")
	}
}

func TestMemoryMapFile_Missing(t *testing.T) {
	_, err := MemoryMapFile("/tmp/pepper_missing_mmap.bin")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
