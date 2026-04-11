// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 58 — handshake 48.4%→higher
// Tests: PerformKEM nil-conn, NewEncryptionKey short secret, ValidateMessage fail path,
// SecureMessage + DecodeSecureMessage roundtrip and error paths
package security_test

import (
	"testing"

	security "github.com/ramseyauron/quantix/src/handshake"
)

// ─── PerformKEM — nil connection → error ──────────────────────────────────────

func TestSprint58_PerformKEM_NilConn(t *testing.T) {
	_, err := security.PerformKEM(nil, true)
	if err == nil {
		t.Error("expected error for nil connection")
	}
}

func TestSprint58_PerformKEM_NilConn_Responder(t *testing.T) {
	_, err := security.PerformKEM(nil, false)
	if err == nil {
		t.Error("expected error for nil connection (responder)")
	}
}

// ─── NewEncryptionKey — short shared secret ───────────────────────────────────

func TestSprint58_NewEncryptionKey_TooShort(t *testing.T) {
	_, err := security.NewEncryptionKey([]byte("short"))
	if err == nil {
		t.Error("expected error for shared secret < 32 bytes")
	}
}

func TestSprint58_NewEncryptionKey_Exactly32Bytes(t *testing.T) {
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i + 1)
	}
	ek, err := security.NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("unexpected error for 32-byte secret: %v", err)
	}
	if ek == nil {
		t.Error("expected non-nil EncryptionKey for 32-byte secret")
	}
}

func TestSprint58_NewEncryptionKey_MoreThan32Bytes(t *testing.T) {
	secret := make([]byte, 64)
	for i := range secret {
		secret[i] = byte(i)
	}
	ek, err := security.NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("unexpected error for 64-byte secret: %v", err)
	}
	if ek == nil {
		t.Error("expected non-nil EncryptionKey for 64-byte secret")
	}
}

// ─── Encrypt / Decrypt roundtrip ─────────────────────────────────────────────

func TestSprint58_Encrypt_Decrypt_Roundtrip(t *testing.T) {
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i + 7)
	}
	ek, err := security.NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("NewEncryptionKey: %v", err)
	}

	plaintext := []byte("hello quantix post-quantum blockchain")
	ciphertext, err := ek.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Error("expected non-empty ciphertext")
	}

	decrypted, err := ek.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypt mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestSprint58_Decrypt_TooShortCiphertext(t *testing.T) {
	secret := make([]byte, 32)
	ek, _ := security.NewEncryptionKey(secret)
	_, err := ek.Decrypt([]byte("too-short"))
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

// ─── SecureMessage — nil EncryptionKey ────────────────────────────────────────

func TestSprint58_SecureMessage_NilKey_Sprint58(t *testing.T) {
	msg := &security.Message{Type: "ping", Data: "hello"}
	_, err := security.SecureMessage(msg, nil)
	if err == nil {
		t.Error("expected error for nil encryption key")
	}
}

// ─── DecodeSecureMessage — nil key ────────────────────────────────────────────

func TestSprint58_DecodeSecureMessage_NilKey(t *testing.T) {
	_, err := security.DecodeSecureMessage([]byte("data"), nil)
	if err == nil {
		t.Error("expected error for nil encryption key")
	}
}

// ─── SecureMessage + DecodeSecureMessage — full roundtrip ─────────────────────

func TestSprint58_SecureMessage_Roundtrip(t *testing.T) {
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i + 3)
	}
	ek, err := security.NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("NewEncryptionKey: %v", err)
	}

	msg := &security.Message{Type: "ping", Data: "test-data"}
	encrypted, err := security.SecureMessage(msg, ek)
	if err != nil {
		t.Fatalf("SecureMessage: %v", err)
	}

	decoded, err := security.DecodeSecureMessage(encrypted, ek)
	if err != nil {
		t.Fatalf("DecodeSecureMessage: %v", err)
	}
	if decoded.Type != "ping" {
		t.Errorf("decoded type: got %q, want %q", decoded.Type, "ping")
	}
}

// ─── DecodeSecureMessage — invalid ciphertext ────────────────────────────────

func TestSprint58_DecodeSecureMessage_InvalidCiphertext(t *testing.T) {
	secret := make([]byte, 32)
	ek, _ := security.NewEncryptionKey(secret)
	_, err := security.DecodeSecureMessage([]byte("garbage-not-valid-cipher"), ek)
	if err == nil {
		t.Error("expected error for invalid ciphertext")
	}
}
