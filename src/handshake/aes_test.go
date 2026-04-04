// MIT License
// Copyright (c) 2024 quantix

// P3-Q1: PEPPER targeted tests for src/handshake — AES, encryption, SecureMessage coverage.
package security

import (
	"bytes"
	"testing"
)

// ---------------------------------------------------------------------------
// NewEncryptionKey
// ---------------------------------------------------------------------------

func TestNewEncryptionKey_Valid32Bytes(t *testing.T) {
	secret := bytes.Repeat([]byte{0xAA}, 32)
	key, err := NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("NewEncryptionKey: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
	if key.AESGCM == nil {
		t.Error("expected AESGCM to be set")
	}
}

func TestNewEncryptionKey_TooShort(t *testing.T) {
	secret := bytes.Repeat([]byte{0x01}, 16) // only 16 bytes
	_, err := NewEncryptionKey(secret)
	if err == nil {
		t.Error("expected error for short secret")
	}
}

func TestNewEncryptionKey_Exactly32Bytes(t *testing.T) {
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i)
	}
	key, err := NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(key.SharedSecret, secret) {
		t.Error("SharedSecret should match input")
	}
}

func TestNewEncryptionKey_LongerSecret(t *testing.T) {
	secret := bytes.Repeat([]byte{0x55}, 64)
	key, err := NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

// ---------------------------------------------------------------------------
// Encrypt / Decrypt round-trip
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	secret := bytes.Repeat([]byte{0xBB}, 32)
	key, _ := NewEncryptionKey(secret)

	plaintext := []byte("hello quantix P3-Q1")
	ct, err := key.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if bytes.Equal(ct, plaintext) {
		t.Error("ciphertext should not equal plaintext")
	}

	pt, err := key.Decrypt(ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Errorf("decrypted %q != original %q", pt, plaintext)
	}
}

func TestEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	secret := bytes.Repeat([]byte{0xCC}, 32)
	key, _ := NewEncryptionKey(secret)

	ct, err := key.Encrypt([]byte{})
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}
	pt, err := key.Decrypt(ct)
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}
	if len(pt) != 0 {
		t.Errorf("expected empty plaintext, got %v", pt)
	}
}

func TestDecrypt_TamperedCiphertext_Fails(t *testing.T) {
	secret := bytes.Repeat([]byte{0xDD}, 32)
	key, _ := NewEncryptionKey(secret)

	ct, _ := key.Encrypt([]byte("important data"))
	// Tamper last byte
	ct[len(ct)-1] ^= 0xFF
	_, err := key.Decrypt(ct)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestDecrypt_TooShort_Fails(t *testing.T) {
	secret := bytes.Repeat([]byte{0xEE}, 32)
	key, _ := NewEncryptionKey(secret)

	_, err := key.Decrypt([]byte{0x01, 0x02}) // shorter than nonce
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

func TestEncrypt_NilKey_Fails(t *testing.T) {
	var key *EncryptionKey
	_, err := key.Encrypt([]byte("data"))
	if err == nil {
		t.Error("expected error when key is nil")
	}
}

func TestDecrypt_NilKey_Fails(t *testing.T) {
	var key *EncryptionKey
	_, err := key.Decrypt([]byte("data"))
	if err == nil {
		t.Error("expected error when key is nil")
	}
}

// ---------------------------------------------------------------------------
// SecureMessage / DecodeSecureMessage
// ---------------------------------------------------------------------------

func TestSecureMessage_Roundtrip(t *testing.T) {
	secret := bytes.Repeat([]byte{0xAB}, 32)
	key, _ := NewEncryptionKey(secret)

	msg := &Message{Type: "ping", Data: "secure-test-node"}
	ct, err := SecureMessage(msg, key)
	if err != nil {
		t.Fatalf("SecureMessage: %v", err)
	}

	decoded, err := DecodeSecureMessage(ct, key)
	if err != nil {
		t.Fatalf("DecodeSecureMessage: %v", err)
	}
	if decoded.Type != "ping" {
		t.Errorf("expected ping, got %q", decoded.Type)
	}
}

func TestDecodeSecureMessage_NilKey_Fails(t *testing.T) {
	_, err := DecodeSecureMessage([]byte("anything"), nil)
	if err == nil {
		t.Error("expected error for nil key")
	}
}

func TestDecodeSecureMessage_GarbageData_Fails(t *testing.T) {
	secret := bytes.Repeat([]byte{0xFF}, 32)
	key, _ := NewEncryptionKey(secret)

	_, err := DecodeSecureMessage([]byte("not-encrypted-data-but-long-enough-to-pass-nonce-check-xxxx"), key)
	if err == nil {
		t.Error("expected error for garbage data")
	}
}
