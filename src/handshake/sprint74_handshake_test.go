// test(PEPPER): Sprint 74 — handshake 49.7%→higher
// Tests: ValidateMessage remaining branches (peer_info valid/invalid, unknown type),
// NewEncryptionKey edge cases (31-byte secret), Encrypt nil key, DecodeSecureMessage
// short ciphertext, Encode/DecodeMessage roundtrip additional cases
package security

import (
	"encoding/json"
	"strings"
	"testing"
)

// ─── ValidateMessage — peer_info valid map ────────────────────────────────────

func TestSprint74_ValidateMessage_PeerInfoValidMap(t *testing.T) {
	m := &Message{
		Type: "peer_info",
		Data: map[string]interface{}{"node_id": "abc", "address": "127.0.0.1:1234"},
	}
	if err := m.ValidateMessage(); err != nil {
		t.Errorf("expected valid peer_info map, got: %v", err)
	}
}

func TestSprint74_ValidateMessage_PeerInfoInvalidType(t *testing.T) {
	m := &Message{
		Type: "peer_info",
		Data: 42, // invalid
	}
	if err := m.ValidateMessage(); err == nil {
		t.Error("expected error for non-map/non-PeerInfo data in peer_info")
	}
}

// ─── ValidateMessage — unknown type passes through ───────────────────────────

func TestSprint74_ValidateMessage_UnknownTypePasses(t *testing.T) {
	m := &Message{
		Type: "gossip_block",
		Data: map[string]interface{}{"hash": "abc"},
	}
	if err := m.ValidateMessage(); err != nil {
		t.Errorf("expected unknown type to pass through, got: %v", err)
	}
}

func TestSprint74_ValidateMessage_GossipTxPasses(t *testing.T) {
	m := &Message{
		Type: "gossip_tx",
		Data: "some_serialized_tx",
	}
	if err := m.ValidateMessage(); err != nil {
		t.Errorf("expected gossip_tx to pass through, got: %v", err)
	}
}

// ─── ValidateMessage — consensus_msg passes through ──────────────────────────

func TestSprint74_ValidateMessage_ConsensusMsgPasses(t *testing.T) {
	m := &Message{
		Type: "consensus_msg",
		Data: map[string]interface{}{"type": "proposal"},
	}
	if err := m.ValidateMessage(); err != nil {
		t.Errorf("expected consensus_msg to pass through, got: %v", err)
	}
}

// ─── NewEncryptionKey — exactly 31 bytes (too short) ─────────────────────────

func TestSprint74_NewEncryptionKey_TooShort(t *testing.T) {
	_, err := NewEncryptionKey(make([]byte, 31))
	if err == nil {
		t.Error("expected error for 31-byte shared secret")
	}
}

// ─── Encrypt — nil AESGCM ────────────────────────────────────────────────────

func TestSprint74_Encrypt_NilKey(t *testing.T) {
	enc := &EncryptionKey{} // AESGCM is nil
	_, err := enc.Encrypt([]byte("plaintext"))
	if err == nil {
		t.Error("expected error for nil AESGCM encrypt")
	}
}

// ─── Decrypt — nil AESGCM ────────────────────────────────────────────────────

func TestSprint74_Decrypt_NilKey(t *testing.T) {
	enc := &EncryptionKey{} // AESGCM is nil
	_, err := enc.Decrypt([]byte("ciphertext"))
	if err == nil {
		t.Error("expected error for nil AESGCM decrypt")
	}
}

// ─── SecureMessage + DecodeSecureMessage roundtrip ───────────────────────────

func TestSprint74_SecureMessage_Roundtrip(t *testing.T) {
	key, err := NewEncryptionKey(make([]byte, 32))
	if err != nil {
		t.Fatalf("NewEncryptionKey: %v", err)
	}
	msg := &Message{Type: "ping", Data: "node-test"}
	encoded, err := SecureMessage(msg, key)
	if err != nil {
		t.Fatalf("SecureMessage: %v", err)
	}
	decoded, err := DecodeSecureMessage(encoded, key)
	if err != nil {
		t.Fatalf("DecodeSecureMessage: %v", err)
	}
	if decoded.Type != "ping" {
		t.Errorf("expected ping, got %q", decoded.Type)
	}
}

// ─── DecodeSecureMessage — too short ciphertext ──────────────────────────────

func TestSprint74_DecodeSecureMessage_TooShort(t *testing.T) {
	key, err := NewEncryptionKey(make([]byte, 32))
	if err != nil {
		t.Fatalf("NewEncryptionKey: %v", err)
	}
	_, err = DecodeSecureMessage([]byte("short"), key)
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

// ─── Encode / DecodeMessage roundtrip ────────────────────────────────────────

func TestSprint74_Encode_DecodeMessage_Roundtrip(t *testing.T) {
	original := &Message{
		Type: "version",
		Data: map[string]interface{}{
			"version":      "1.0",
			"node_id":      "abc123",
			"chain_id":     "73310",
			"block_height": float64(5),
			"nonce":        "deadbeef",
		},
	}
	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := DecodeMessage(encoded)
	if err != nil {
		t.Fatalf("DecodeMessage: %v", err)
	}
	if decoded.Type != "version" {
		t.Errorf("expected version, got %q", decoded.Type)
	}
}

// ─── DecodeMessage — invalid JSON ────────────────────────────────────────────

func TestSprint74_DecodeMessage_InvalidJSON(t *testing.T) {
	_, err := DecodeMessage([]byte("not-json!!!"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ─── ValidateMessage — version missing chain_id ───────────────────────────────

func TestSprint74_ValidateMessage_VersionMissingChainID(t *testing.T) {
	m := &Message{
		Type: "version",
		Data: map[string]interface{}{
			"version":      "1.0",
			"node_id":      "abc",
			"block_height": float64(1),
			"nonce":        "ff",
			// chain_id missing
		},
	}
	if err := m.ValidateMessage(); err == nil {
		t.Error("expected error for missing chain_id")
	}
}

// ─── ValidateMessage — version missing block_height ──────────────────────────

func TestSprint74_ValidateMessage_VersionMissingBlockHeight(t *testing.T) {
	m := &Message{
		Type: "version",
		Data: map[string]interface{}{
			"version":  "1.0",
			"node_id":  "abc",
			"chain_id": "73310",
			"nonce":    "ff",
			// block_height missing
		},
	}
	if err := m.ValidateMessage(); err == nil {
		t.Error("expected error for missing block_height")
	}
}

// ─── ValidateMessage — version missing nonce ──────────────────────────────────

func TestSprint74_ValidateMessage_VersionMissingNonce(t *testing.T) {
	m := &Message{
		Type: "version",
		Data: map[string]interface{}{
			"version":      "1.0",
			"node_id":      "abc",
			"chain_id":     "73310",
			"block_height": float64(1),
			// nonce missing
		},
	}
	if err := m.ValidateMessage(); err == nil {
		t.Error("expected error for missing nonce")
	}
}

// ─── ValidateMessage — version not-a-map ─────────────────────────────────────

func TestSprint74_ValidateMessage_VersionNotMap(t *testing.T) {
	m := &Message{
		Type: "version",
		Data: "not-a-map",
	}
	if err := m.ValidateMessage(); err == nil {
		t.Error("expected error for version data that is not a map")
	}
}

// ─── Encode JSON structure ─────────────────────────────────────────────────────

func TestSprint74_Encode_ContainsType(t *testing.T) {
	m := &Message{Type: "ping", Data: "hello"}
	encoded, err := m.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !strings.Contains(string(encoded), "ping") {
		t.Error("encoded message should contain type 'ping'")
	}
	// Verify it's valid JSON
	var out map[string]interface{}
	if err := json.Unmarshal(encoded, &out); err != nil {
		t.Errorf("encoded message is not valid JSON: %v", err)
	}
}
