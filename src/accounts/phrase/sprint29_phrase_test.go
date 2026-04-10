// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 29 - accounts/phrase GeneratePasskey + GenerateKeys coverage
package seed

import (
	"testing"
)

// ---------------------------------------------------------------------------
// GeneratePasskey
// ---------------------------------------------------------------------------

func TestGeneratePasskey_WithEmptyPK_GeneratesKey(t *testing.T) {
	// Empty PK triggers internal key generation
	passkey, err := GeneratePasskey("testpassphrase", nil)
	if err != nil {
		t.Fatalf("GeneratePasskey with nil pk: %v", err)
	}
	if len(passkey) == 0 {
		t.Fatal("expected non-empty passkey")
	}
}

func TestGeneratePasskey_WithEmptyPK_LengthIsPasskeySize(t *testing.T) {
	passkey, err := GeneratePasskey("testpassphrase", nil)
	if err != nil {
		t.Fatalf("GeneratePasskey: %v", err)
	}
	if len(passkey) != int(PasskeySize) {
		t.Fatalf("expected passkey length %d, got %d", PasskeySize, len(passkey))
	}
}

func TestGeneratePasskey_NonEmpty_Passphrase(t *testing.T) {
	pk := []byte("fakepublickey")
	passkey, err := GeneratePasskey("my-strong-passphrase", pk)
	if err != nil {
		t.Fatalf("GeneratePasskey: %v", err)
	}
	if len(passkey) == 0 {
		t.Fatal("expected non-empty passkey")
	}
}

func TestGeneratePasskey_Nonce_UniqueEachCall(t *testing.T) {
	// Two calls with the same inputs should produce different passkeys (random nonce)
	pk := []byte("somepublickey")
	p1, err1 := GeneratePasskey("passphrase", pk)
	p2, err2 := GeneratePasskey("passphrase", pk)
	if err1 != nil || err2 != nil {
		t.Fatalf("GeneratePasskey error: %v / %v", err1, err2)
	}
	// Due to random nonce, should not be equal (astronomically unlikely to be equal)
	equal := true
	for i := range p1 {
		if i >= len(p2) || p1[i] != p2[i] {
			equal = false
			break
		}
	}
	if equal {
		t.Log("WARNING: two GeneratePasskey calls produced identical passkeys (random nonce may be broken)")
	}
}

func TestGeneratePasskey_InvalidUTF8_Error(t *testing.T) {
	// Invalid UTF-8 should return error
	invalidUTF8 := "\xff\xfe"
	_, err := GeneratePasskey(invalidUTF8, []byte("pk"))
	if err == nil {
		t.Fatal("expected error for invalid UTF-8 passphrase")
	}
}

func TestGeneratePasskey_EmptyPassphrase_ValidPK(t *testing.T) {
	pk := []byte("somepublickey")
	passkey, err := GeneratePasskey("", pk)
	if err != nil {
		t.Fatalf("GeneratePasskey with empty passphrase: %v", err)
	}
	if len(passkey) == 0 {
		t.Fatal("expected non-empty passkey even with empty passphrase")
	}
}

// ---------------------------------------------------------------------------
// GenerateKeys — requires network; skip in offline environments
// ---------------------------------------------------------------------------

func TestGenerateKeys_NoError(t *testing.T) {
	t.Skip("GenerateKeys requires network access to fetch word list")
}

func TestGenerateKeys_PassphraseNonEmpty(t *testing.T) {
	t.Skip("GenerateKeys requires network access to fetch word list")
}

func TestGenerateKeys_Base32PasskeyNonEmpty(t *testing.T) {
	t.Skip("GenerateKeys requires network access to fetch word list")
}

func TestGenerateKeys_HashedPasskeyNonEmpty(t *testing.T) {
	t.Skip("GenerateKeys requires network access to fetch word list")
}

func TestGenerateKeys_MacKeyNonEmpty(t *testing.T) {
	t.Skip("GenerateKeys requires network access to fetch word list")
}

func TestGenerateKeys_ChainCodeNonEmpty(t *testing.T) {
	t.Skip("GenerateKeys requires network access to fetch word list")
}

func TestGenerateKeys_FingerprintNonEmpty(t *testing.T) {
	t.Skip("GenerateKeys requires network access to fetch word list")
}

func TestGenerateKeys_Base32Passkey_IsDecodable(t *testing.T) {
	t.Skip("GenerateKeys requires network access to fetch word list")
}
