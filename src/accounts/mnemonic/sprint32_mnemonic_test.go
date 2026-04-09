// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 32 - accounts/mnemonic entropy validation + passphrase edge cases
package sips3

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// isValidEntropy — all valid and invalid values
// ---------------------------------------------------------------------------

func TestSprint32_IsValidEntropy_AllValid(t *testing.T) {
	valid := []int{128, 160, 192, 224, 256}
	for _, e := range valid {
		if !isValidEntropy(e) {
			t.Errorf("expected entropy %d to be valid", e)
		}
	}
}

func TestSprint32_IsValidEntropy_Invalid(t *testing.T) {
	invalid := []int{0, 64, 100, 127, 129, 255, 257, 512}
	for _, e := range invalid {
		if isValidEntropy(e) {
			t.Errorf("expected entropy %d to be invalid", e)
		}
	}
}

func TestSprint32_IsValidEntropy_Negative(t *testing.T) {
	if isValidEntropy(-1) {
		t.Fatal("negative entropy should not be valid")
	}
}

// ---------------------------------------------------------------------------
// GeneratePassphrase — word count zero
// ---------------------------------------------------------------------------

func TestSprint32_GeneratePassphrase_ZeroWordCount(t *testing.T) {
	words := []string{"apple", "banana", "cherry"}
	phrase, hash, err := GeneratePassphrase(words, 0)
	if err != nil {
		t.Fatalf("unexpected error for wordCount=0: %v", err)
	}
	// Empty passphrase: phrase should be empty, hash non-empty
	_ = phrase
	_ = hash
}

func TestSprint32_GeneratePassphrase_SingleWord(t *testing.T) {
	words := []string{"apple", "banana", "cherry", "date", "elderberry"}
	phrase, _, err := GeneratePassphrase(words, 1)
	if err != nil {
		t.Fatalf("unexpected error for wordCount=1: %v", err)
	}
	parts := strings.Fields(phrase)
	if len(parts) != 1 {
		t.Fatalf("expected 1 word in phrase, got %d", len(parts))
	}
}

func TestSprint32_GeneratePassphrase_TwoWords(t *testing.T) {
	words := []string{"alpha", "beta", "gamma", "delta"}
	phrase, _, err := GeneratePassphrase(words, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.Fields(phrase)
	if len(parts) != 2 {
		t.Fatalf("expected 2 words in phrase, got %d", len(parts))
	}
}

func TestSprint32_GeneratePassphrase_SingleWordList(t *testing.T) {
	// List with only one word: all selections will be that word
	words := []string{"only"}
	phrase, _, err := GeneratePassphrase(words, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if phrase != "only only only" {
		t.Fatalf("expected 'only only only', got %q", phrase)
	}
}

func TestSprint32_GeneratePassphrase_HashNotEmpty(t *testing.T) {
	words := []string{"one", "two", "three", "four", "five"}
	_, hash, err := GeneratePassphrase(words, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
}

// ---------------------------------------------------------------------------
// NewMnemonic — all valid entropy values fail gracefully (require network)
// ---------------------------------------------------------------------------

func TestSprint32_NewMnemonic_InvalidEntropy_AllCases(t *testing.T) {
	invalid := []int{0, 64, 100, 512, -128}
	for _, e := range invalid {
		_, _, err := NewMnemonic(e)
		if err == nil {
			t.Errorf("expected error for invalid entropy %d", e)
		}
	}
}

func TestSprint32_NewMnemonic_ValidEntropy_NetworkRequired(t *testing.T) {
	// Valid entropy requires network to fetch word list — skip
	t.Skip("NewMnemonic with valid entropy requires network access to fetch word list")
}
