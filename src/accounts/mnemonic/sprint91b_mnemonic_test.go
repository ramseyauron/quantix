// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 91b — accounts/mnemonic 35.6%→higher
// Covers: FetchFileList (error paths), SelectAndLoadTxtFile (error paths), NewMnemonic (more paths)
package sips3

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// FetchFileList — 404 returns error
// ---------------------------------------------------------------------------

func TestSprint91b_FetchFileList_404_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := FetchFileList(srv.URL)
	if err == nil {
		t.Error("expected error for 404 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// FetchFileList — invalid JSON returns error
// ---------------------------------------------------------------------------

func TestSprint91b_FetchFileList_InvalidJSON_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	_, err := FetchFileList(srv.URL)
	if err == nil {
		t.Error("expected error for invalid JSON response, got nil")
	}
}

// ---------------------------------------------------------------------------
// FetchFileList — valid JSON returns file list
// ---------------------------------------------------------------------------

func TestSprint91b_FetchFileList_ValidJSON_ReturnsList(t *testing.T) {
	files := []GitHubFile{
		{Name: "words.txt", Path: "words.txt", Type: "file"},
		{Name: "readme.md", Path: "readme.md", Type: "file"},
	}
	data, _ := json.Marshal(files)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	result, err := FetchFileList(srv.URL)
	if err != nil {
		t.Fatalf("FetchFileList error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 files, got %d", len(result))
	}
	if result[0].Name != "words.txt" {
		t.Errorf("first file name = %q, want words.txt", result[0].Name)
	}
}

// ---------------------------------------------------------------------------
// FetchFileList — unreachable URL returns error
// ---------------------------------------------------------------------------

func TestSprint91b_FetchFileList_UnreachableURL_ReturnsError(t *testing.T) {
	_, err := FetchFileList("http://127.0.0.1:0/nonexistent")
	if err == nil {
		t.Error("expected error for unreachable URL, got nil")
	}
}

// ---------------------------------------------------------------------------
// SelectAndLoadTxtFile — URL returns only non-txt files → error
// ---------------------------------------------------------------------------

func TestSprint91b_SelectAndLoadTxtFile_NoTxtFiles_ReturnsError(t *testing.T) {
	files := []GitHubFile{
		{Name: "readme.md", Path: "readme.md", Type: "file"},
	}
	data, _ := json.Marshal(files)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	_, err := SelectAndLoadTxtFile(srv.URL)
	if err == nil {
		t.Error("expected error when no .txt files in list, got nil")
	}
}

// ---------------------------------------------------------------------------
// SelectAndLoadTxtFile — fetch fails → error
// ---------------------------------------------------------------------------

func TestSprint91b_SelectAndLoadTxtFile_FetchFails_ReturnsError(t *testing.T) {
	_, err := SelectAndLoadTxtFile("http://127.0.0.1:0/nonexistent")
	if err == nil {
		t.Error("expected error for unreachable URL, got nil")
	}
}

// ---------------------------------------------------------------------------
// NewMnemonic — all invalid entropy values return error
// ---------------------------------------------------------------------------

func TestSprint91b_NewMnemonic_InvalidEntropy_127_Error(t *testing.T) {
	_, _, err := NewMnemonic(127)
	if err == nil {
		t.Error("expected error for entropy=127")
	}
}

func TestSprint91b_NewMnemonic_InvalidEntropy_0_Error(t *testing.T) {
	_, _, err := NewMnemonic(0)
	if err == nil {
		t.Error("expected error for entropy=0")
	}
}

func TestSprint91b_NewMnemonic_InvalidEntropy_512_Error(t *testing.T) {
	_, _, err := NewMnemonic(512)
	if err == nil {
		t.Error("expected error for entropy=512")
	}
}

func TestSprint91b_NewMnemonic_InvalidEntropy_255_Error(t *testing.T) {
	_, _, err := NewMnemonic(255)
	if err == nil {
		t.Error("expected error for entropy=255")
	}
}
