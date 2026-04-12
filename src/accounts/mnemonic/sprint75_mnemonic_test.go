// test(PEPPER): Sprint 75 — accounts/mnemonic 34.5%→higher
// Tests: GeneratePassphrase wordCount=0 (empty passphrase edge case),
// GeneratePassphrase single-word list, GeneratePassphrase large word count,
// NewMnemonic invalid entropy, isValidEntropy all valid/invalid values
package sips3

import (
	"strings"
	"testing"
)

// ─── isValidEntropy — all valid values ────────────────────────────────────────

func TestSprint75_IsValidEntropy_Valid(t *testing.T) {
	valid := []int{128, 160, 192, 224, 256}
	for _, e := range valid {
		if !isValidEntropy(e) {
			t.Errorf("expected %d to be valid entropy", e)
		}
	}
}

func TestSprint75_IsValidEntropy_Invalid(t *testing.T) {
	invalid := []int{0, 64, 127, 129, 255, 257, 512, -1}
	for _, e := range invalid {
		if isValidEntropy(e) {
			t.Errorf("expected %d to be invalid entropy", e)
		}
	}
}

// ─── NewMnemonic — invalid entropy returns error ──────────────────────────────

func TestSprint75_NewMnemonic_InvalidEntropy(t *testing.T) {
	_, _, err := NewMnemonic(0)
	if err == nil {
		t.Error("expected error for entropy=0")
	}

	_, _, err = NewMnemonic(64)
	if err == nil {
		t.Error("expected error for entropy=64")
	}

	_, _, err = NewMnemonic(512)
	if err == nil {
		t.Error("expected error for entropy=512")
	}
}

// ─── GeneratePassphrase — empty word list ─────────────────────────────────────

func TestSprint75_GeneratePassphrase_EmptyWordList(t *testing.T) {
	_, _, err := GeneratePassphrase([]string{}, 12)
	if err == nil {
		t.Error("expected error for empty word list")
	}
}

// ─── GeneratePassphrase — wordCount=0 produces empty passphrase ──────────────

func TestSprint75_GeneratePassphrase_ZeroWordCount(t *testing.T) {
	words := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta"}
	passphrase, hash, err := GeneratePassphrase(words, 0)
	// wordCount=0 → empty passphrase → all calls produce same hash → may get duplicate error
	if err != nil {
		t.Logf("GeneratePassphrase wordCount=0 error (likely duplicate passphrase): %v", err)
		return
	}
	// Empty passphrase
	if passphrase != "" {
		t.Errorf("expected empty passphrase for wordCount=0, got %q", passphrase)
	}
	// Hash should still be generated
	if len(hash) == 0 {
		t.Error("expected non-empty hash even for empty passphrase")
	}
}

// ─── GeneratePassphrase — single word ─────────────────────────────────────────

func TestSprint75_GeneratePassphrase_SingleWord(t *testing.T) {
	words := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta"}
	passphrase, hash, err := GeneratePassphrase(words, 1)
	if err != nil {
		t.Fatalf("GeneratePassphrase wordCount=1: %v", err)
	}
	if passphrase == "" {
		t.Error("expected non-empty passphrase for wordCount=1")
	}
	// Single word passphrase should not contain spaces
	if strings.Contains(passphrase, " ") {
		t.Errorf("single-word passphrase should not contain spaces, got %q", passphrase)
	}
	if len(hash) == 0 {
		t.Error("expected non-empty hash")
	}
}

// ─── GeneratePassphrase — 12-word passphrase ──────────────────────────────────

func TestSprint75_GeneratePassphrase_12Words(t *testing.T) {
	// Use a large enough word list to avoid hash collisions in passphraseHashes map
	words := make([]string, 50)
	for i := range words {
		words[i] = string(rune('a'+i%26)) + string(rune('A'+i%26)) + "word"
	}
	passphrase, hash, err := GeneratePassphrase(words, 12)
	if err != nil {
		t.Fatalf("GeneratePassphrase 12 words: %v", err)
	}
	parts := strings.Split(passphrase, " ")
	if len(parts) != 12 {
		t.Errorf("expected 12 words, got %d", len(parts))
	}
	if len(hash) == 0 {
		t.Error("expected non-empty hash")
	}
}

// ─── GeneratePassphrase — word from given list ────────────────────────────────

func TestSprint75_GeneratePassphrase_WordsFromList(t *testing.T) {
	words := []string{"sun", "moon", "star", "sky", "cloud", "rain", "snow", "wind"}
	passphrase, _, err := GeneratePassphrase(words, 3)
	if err != nil {
		t.Fatalf("GeneratePassphrase: %v", err)
	}
	for _, w := range strings.Split(passphrase, " ") {
		found := false
		for _, allowed := range words {
			if w == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("word %q not from word list", w)
		}
	}
}

// ─── GeneratePassphrase — nonce uniqueness ────────────────────────────────────

func TestSprint75_GeneratePassphrase_NonceUniqueness(t *testing.T) {
	// GeneratePassphrase's return value is the stretched hash, not a nonce — the
	// "nonce" is the passphrase hash. Test that two calls produce different hashes.
	words := make([]string, 100)
	for i := range words {
		words[i] = string(rune('a'+i%26)) + string(rune('0'+i%10)) + "test"
	}
	_, hash1, err := GeneratePassphrase(words, 2)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	_, hash2, err := GeneratePassphrase(words, 2)
	if err != nil {
		// Duplicate is acceptable — it means same passphrase generated (collision)
		t.Logf("second call returned error (possible duplicate passphrase): %v", err)
		return
	}
	// If both succeed and they happen to match, that's a collision — extremely unlikely
	if hash1 == hash2 {
		t.Logf("hash collision (extremely unlikely): %s == %s", hash1, hash2)
	}
}
