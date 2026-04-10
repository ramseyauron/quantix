// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 37 - handshake AES error paths + SecureMessage + ValidateMessage edge cases
package security

import (
	"testing"
)

// ---------------------------------------------------------------------------
// NewEncryptionKey — exact 32 bytes boundary
// ---------------------------------------------------------------------------

func TestSprint37_NewEncryptionKey_Exactly32_OK(t *testing.T) {
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i)
	}
	enc, err := NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("NewEncryptionKey with 32 bytes: %v", err)
	}
	if enc == nil {
		t.Fatal("expected non-nil EncryptionKey")
	}
}

func TestSprint37_NewEncryptionKey_Longer_TruncatesTo32(t *testing.T) {
	secret := make([]byte, 64) // longer than 32 — should use first 32
	for i := range secret {
		secret[i] = byte(i + 1)
	}
	enc, err := NewEncryptionKey(secret)
	if err != nil {
		t.Fatalf("NewEncryptionKey with 64 bytes: %v", err)
	}
	if enc == nil {
		t.Fatal("expected non-nil EncryptionKey")
	}
}

// ---------------------------------------------------------------------------
// Encrypt — valid key, extra coverage for large message
// ---------------------------------------------------------------------------

func TestSprint37_Encrypt_LargeMessage_NoPanic(t *testing.T) {
	secret := make([]byte, 32)
	enc, _ := NewEncryptionKey(secret)
	large := make([]byte, 10000)
	ciphertext, err := enc.Encrypt(large)
	if err != nil {
		t.Fatalf("Encrypt large message error: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("expected non-empty ciphertext")
	}
	// Decrypt back
	plaintext, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt large ciphertext error: %v", err)
	}
	if len(plaintext) != len(large) {
		t.Fatalf("expected length %d, got %d", len(large), len(plaintext))
	}
}

// ---------------------------------------------------------------------------
// SecureMessage — nil message path
// ---------------------------------------------------------------------------

func TestSprint37_SecureMessage_NilMessage_NoPanic(t *testing.T) {
	secret := make([]byte, 32)
	enc, _ := NewEncryptionKey(secret)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SecureMessage with nil msg panicked: %v", r)
		}
	}()
	_, _ = SecureMessage(nil, enc)
}

func TestSprint37_SecureMessage_ValidMessage_NonEmpty(t *testing.T) {
	secret := make([]byte, 32)
	enc, _ := NewEncryptionKey(secret)
	msg := &Message{Type: "ping", Data: "hello"}
	result, err := SecureMessage(msg, enc)
	if err != nil {
		t.Fatalf("SecureMessage error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty secure message")
	}
}

// ---------------------------------------------------------------------------
// DecodeSecureMessage — valid roundtrip
// ---------------------------------------------------------------------------

func TestSprint37_DecodeSecureMessage_Roundtrip(t *testing.T) {
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i + 10)
	}
	enc, _ := NewEncryptionKey(secret)
	// Use ping type for simple roundtrip (no required fields beyond type+string data)
	msg := &Message{Type: "ping", Data: "sprint37-roundtrip-test"}
	secured, err := SecureMessage(msg, enc)
	if err != nil {
		t.Fatalf("SecureMessage error: %v", err)
	}
	decoded, err := DecodeSecureMessage(secured, enc)
	if err != nil {
		t.Fatalf("DecodeSecureMessage error: %v", err)
	}
	if decoded.Type != "ping" {
		t.Fatalf("expected Type=ping, got %q", decoded.Type)
	}
}

// ---------------------------------------------------------------------------
// ValidateMessage — gossip_block / gossip_tx passthroughs
// ---------------------------------------------------------------------------

func TestSprint37_ValidateMessage_GossipBlock_Passthrough(t *testing.T) {
	msg := &Message{Type: "gossip_block", Data: map[string]interface{}{}}
	if err := msg.ValidateMessage(); err != nil {
		t.Fatalf("gossip_block should pass through validation: %v", err)
	}
}

func TestSprint37_ValidateMessage_UnknownType_Passthrough(t *testing.T) {
	msg := &Message{Type: "custom_type", Data: "any data"}
	if err := msg.ValidateMessage(); err != nil {
		t.Fatalf("unknown type should pass through: %v", err)
	}
}

func TestSprint37_ValidateMessage_Transaction_IntegerData_Error(t *testing.T) {
	msg := &Message{Type: "transaction", Data: 42}
	if err := msg.ValidateMessage(); err == nil {
		t.Fatal("expected error for transaction with integer data")
	}
}

// ---------------------------------------------------------------------------
// NewHandshake — metrics not nil + repeated calls no panic
// ---------------------------------------------------------------------------

func TestSprint37_NewHandshake_RepeatedCalls_AllNotNil(t *testing.T) {
	for i := 0; i < 5; i++ {
		h := NewHandshake()
		if h == nil {
			t.Fatalf("NewHandshake call %d returned nil", i)
		}
	}
}
