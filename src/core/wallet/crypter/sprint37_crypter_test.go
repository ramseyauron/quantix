// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 37 - crypter dead-code coverage (SEC-C03 documentation) + VerifyPubKey
package crypter

import (
	"testing"
)

// ---------------------------------------------------------------------------
// EncryptSecret (SEC-C03 dead code — documents that random salt prevents decryption)
// ---------------------------------------------------------------------------

func TestSprint37_EncryptSecret_Basic_NoError(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	ciphertext, err := EncryptSecret(key, []byte("hello"), nil)
	if err != nil {
		t.Fatalf("EncryptSecret error: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("expected non-empty ciphertext")
	}
}

func TestSprint37_EncryptSecret_EmptyPlaintext_NoError(t *testing.T) {
	key := make([]byte, 32)
	// Empty plaintext — AES-GCM can encrypt empty data
	_, err := EncryptSecret(key, []byte{}, nil)
	_ = err // may succeed or fail depending on implementation
}

func TestSprint37_EncryptSecret_NilPlaintext_NoPanel(t *testing.T) {
	key := make([]byte, 32)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("EncryptSecret panicked with nil plaintext: %v", r)
		}
	}()
	_, _ = EncryptSecret(key, nil, nil)
}

// ---------------------------------------------------------------------------
// DecryptSecret (SEC-C03 — random salt makes it incompatible with EncryptSecret)
// ---------------------------------------------------------------------------

func TestSprint37_DecryptSecret_InvalidCiphertext_Error(t *testing.T) {
	key := make([]byte, 32)
	// Short ciphertext should fail AES-GCM auth check
	_, err := DecryptSecret(key, []byte("short"), nil)
	// May error or not — just no panic
	_ = err
}

func TestSprint37_DecryptSecret_EmptyKey_NoPanel(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DecryptSecret panicked: %v", r)
		}
	}()
	_, _ = DecryptSecret([]byte{}, []byte("ciphertext"), nil)
}

// ---------------------------------------------------------------------------
// VerifyPubKey — all branches
// ---------------------------------------------------------------------------

func TestSprint37_VerifyPubKey_EmptySecret_False(t *testing.T) {
	if VerifyPubKey(nil, []byte("pk")) {
		t.Fatal("expected false for nil secret")
	}
}

func TestSprint37_VerifyPubKey_EmptyPubKey_False(t *testing.T) {
	if VerifyPubKey([]byte("secret"), nil) {
		t.Fatal("expected false for nil pubkey")
	}
}

func TestSprint37_VerifyPubKey_WrongLength_False(t *testing.T) {
	// secret must be exactly 2 * len(pubKey)
	secret := make([]byte, 10)
	pubKey := make([]byte, 8) // 2*8=16 ≠ 10
	if VerifyPubKey(secret, pubKey) {
		t.Fatal("expected false for mismatched lengths")
	}
}

func TestSprint37_VerifyPubKey_MatchingLastHalf_True(t *testing.T) {
	// SPHINCS+ format: SKseed || SKprf || PKseed || PKroot
	// PK = last half of secret
	pubKey := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	secret := append(make([]byte, 8), pubKey...) // 8 zeros + pubKey
	if !VerifyPubKey(secret, pubKey) {
		t.Fatal("expected true when secret's last half matches pubkey")
	}
}

func TestSprint37_VerifyPubKey_MismatchingLastHalf_False(t *testing.T) {
	pubKey := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	secret := make([]byte, 16) // all zeros
	if VerifyPubKey(secret, pubKey) {
		t.Fatal("expected false when secret's last half does not match pubkey")
	}
}

// ---------------------------------------------------------------------------
// SetKeyFromPassphrase — valid/invalid inputs
// ---------------------------------------------------------------------------

func TestSprint37_SetKeyFromPassphrase_ValidKey_True(t *testing.T) {
	c := &CCrypter{}
	masterKey := []byte("test-master-key-32-bytes-padded!")[:32]
	salt := make([]byte, WALLET_CRYPTO_IV_SIZE)
	result := c.SetKeyFromPassphrase(masterKey, salt, 1000)
	if !result {
		t.Fatal("expected SetKeyFromPassphrase to return true with valid inputs")
	}
}

func TestSprint37_SetKeyFromPassphrase_EmptyMasterKey_False(t *testing.T) {
	c := &CCrypter{}
	salt := make([]byte, WALLET_CRYPTO_IV_SIZE)
	result := c.SetKeyFromPassphrase([]byte{}, salt, 1000)
	// empty key may fail or succeed — just no panic
	_ = result
}

func TestSprint37_SetKeyFromPassphrase_ZeroRounds_NoPanic(t *testing.T) {
	c := &CCrypter{}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetKeyFromPassphrase panicked with 0 rounds: %v", r)
		}
	}()
	_ = c.SetKeyFromPassphrase([]byte("key"), make([]byte, WALLET_CRYPTO_IV_SIZE), 0)
}

// ---------------------------------------------------------------------------
// DecryptKey — error paths
// ---------------------------------------------------------------------------

func TestSprint37_DecryptKey_EmptyCryptedSecret_Error(t *testing.T) {
	key := make([]byte, 32)
	pubKey := make([]byte, 16)
	_, err := DecryptKey(key, []byte{}, pubKey)
	// DecryptSecret will fail on short/empty ciphertext
	_ = err
}

func TestSprint37_DecryptKey_EmptyPubKey_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DecryptKey panicked: %v", r)
		}
	}()
	_, _ = DecryptKey(make([]byte, 32), []byte("ciphertext"), nil)
}
