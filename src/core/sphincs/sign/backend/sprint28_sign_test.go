// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 28 - sphincs sign backend + utils coverage
package sign

import (
	"bytes"
	"testing"

	params "github.com/ramseyauron/quantix/src/core/sphincs/config"
	key "github.com/ramseyauron/quantix/src/core/sphincs/key/backend"
)

// helper: create a real SphincsManager with real params/keymanager (no DB)
func newTestSphincsManager(t *testing.T) *SphincsManager {
	t.Helper()
	p, err := params.NewSPHINCSParameters()
	if err != nil {
		t.Fatalf("NewSPHINCSParameters: %v", err)
	}
	km, err := key.NewKeyManager()
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	return NewSphincsManager(nil, km, p)
}

// ---------------------------------------------------------------------------
// NewSphincsManager
// ---------------------------------------------------------------------------

func TestNewSphincsManager_NotNil(t *testing.T) {
	sm := newTestSphincsManager(t)
	if sm == nil {
		t.Fatal("expected non-nil SphincsManager")
	}
}

// ---------------------------------------------------------------------------
// buildMessageWithTimestampAndNonce
// ---------------------------------------------------------------------------

func TestBuildMessageWithTimestampAndNonce_NonEmpty(t *testing.T) {
	ts := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	nonce := []byte{0xAB, 0xCD}
	msg := []byte("hello quantix")
	out := buildMessageWithTimestampAndNonce(ts, nonce, msg)
	if len(out) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestBuildMessageWithTimestampAndNonce_BindsAllInputs(t *testing.T) {
	ts := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	nonce := []byte{0xAB, 0xCD}
	out1 := buildMessageWithTimestampAndNonce(ts, nonce, []byte("hello"))
	out2 := buildMessageWithTimestampAndNonce(ts, nonce, []byte("world"))
	if bytes.Equal(out1, out2) {
		t.Fatal("different messages should produce different outputs")
	}
}

func TestBuildMessageWithTimestampAndNonce_NilInputs_NoPanel(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic with nil inputs: %v", r)
		}
	}()
	buildMessageWithTimestampAndNonce(nil, nil, nil)
}

// ---------------------------------------------------------------------------
// buildSigParts
// ---------------------------------------------------------------------------

func TestBuildSigParts_NonEmpty(t *testing.T) {
	parts := buildSigParts([]byte("fakesig"), []byte("fakecommit"))
	if len(parts) == 0 {
		t.Fatal("expected non-empty parts")
	}
}

func TestBuildSigParts_CommitmentInParts(t *testing.T) {
	sig := make([]byte, 100) // non-trivial size for chunking
	commit := []byte("fakecommit")
	parts := buildSigParts(sig, commit)
	// parts[0] = commit || sig_chunk0 — commitment at prefix
	if !bytes.HasPrefix(parts[0], commit) {
		t.Fatal("commitment bytes not at prefix of parts[0]")
	}
	// parts[4] = CommitmentLeaf(commit) — non-empty
	if len(parts[4]) == 0 {
		t.Fatal("expected non-empty commitment leaf in parts[4]")
	}
}

// ---------------------------------------------------------------------------
// generateNonce + generateTimestamp
// ---------------------------------------------------------------------------

func TestGenerateNonce_Length16(t *testing.T) {
	n, err := generateNonce()
	if err != nil {
		t.Fatalf("generateNonce error: %v", err)
	}
	if len(n) != 16 {
		t.Fatalf("expected 16-byte nonce, got %d", len(n))
	}
}

func TestGenerateNonce_Unique(t *testing.T) {
	n1, _ := generateNonce()
	n2, _ := generateNonce()
	if bytes.Equal(n1, n2) {
		t.Fatal("two nonces should not be equal")
	}
}

func TestGenerateTimestamp_Length8(t *testing.T) {
	ts := generateTimestamp()
	if len(ts) != 8 {
		t.Fatalf("expected 8-byte timestamp, got %d", len(ts))
	}
}

func TestGenerateTimestamp_NonZero(t *testing.T) {
	ts := generateTimestamp()
	allZero := true
	for _, b := range ts {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Fatal("timestamp should not be all zeros")
	}
}

// ---------------------------------------------------------------------------
// PublicKeyRegistry
// ---------------------------------------------------------------------------

func TestPublicKeyRegistry_NotNil(t *testing.T) {
	r := NewPublicKeyRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestPublicKeyRegistry_RegisterAndLookup(t *testing.T) {
	r := NewPublicKeyRegistry()
	pk := []byte{1, 2, 3, 4}
	r.Register("node-1", pk)
	got, found := r.Lookup("node-1")
	if !found {
		t.Fatal("expected to find registered key")
	}
	if !bytes.Equal(got, pk) {
		t.Fatal("retrieved key does not match registered key")
	}
}

func TestPublicKeyRegistry_Lookup_Unknown_False(t *testing.T) {
	r := NewPublicKeyRegistry()
	_, found := r.Lookup("unknown-node")
	if found {
		t.Fatal("should not find unknown node")
	}
}

func TestPublicKeyRegistry_Register_NilPK_Handled(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic on nil pk register: %v", r)
		}
	}()
	r := NewPublicKeyRegistry()
	r.Register("node-nil", nil)
}

func TestPublicKeyRegistry_VerifyIdentity_Unknown_False(t *testing.T) {
	r := NewPublicKeyRegistry()
	if r.VerifyIdentity("unknown", []byte{1, 2, 3}) {
		t.Fatal("should return false for unknown node")
	}
}

func TestPublicKeyRegistry_VerifyIdentity_Match_True(t *testing.T) {
	r := NewPublicKeyRegistry()
	pk := []byte{0xAA, 0xBB, 0xCC}
	r.Register("node-A", pk)
	if !r.VerifyIdentity("node-A", pk) {
		t.Fatal("matching PK should return true")
	}
}

func TestPublicKeyRegistry_VerifyIdentity_Mismatch_False(t *testing.T) {
	r := NewPublicKeyRegistry()
	r.Register("node-A", []byte{0xAA, 0xBB, 0xCC})
	if r.VerifyIdentity("node-A", []byte{0x11, 0x22, 0x33}) {
		t.Fatal("mismatched PK should return false")
	}
}

func TestPublicKeyRegistry_MultipleNodes(t *testing.T) {
	r := NewPublicKeyRegistry()
	for i := 0; i < 5; i++ {
		nodeID := string(rune('A' + i))
		pk := []byte{byte(i), byte(i + 1)}
		r.Register(nodeID, pk)
	}
	for i := 0; i < 5; i++ {
		nodeID := string(rune('A' + i))
		pk := []byte{byte(i), byte(i + 1)}
		got, found := r.Lookup(nodeID)
		if !found {
			t.Fatalf("expected to find node %s", nodeID)
		}
		if !bytes.Equal(got, pk) {
			t.Fatalf("node %s key mismatch", nodeID)
		}
	}
}

// ---------------------------------------------------------------------------
// StoreTimestampNonce / CheckTimestampNonce (nil DB — should error gracefully)
// ---------------------------------------------------------------------------

func TestStoreTimestampNonce_NilDB_NoPanel(t *testing.T) {
	sm := newTestSphincsManager(t)
	err := sm.StoreTimestampNonce([]byte{1, 2, 3}, []byte{4, 5, 6})
	_ = err // nil DB may error or not — just no panic
}

func TestCheckTimestampNonce_NilDB_NoPanel(t *testing.T) {
	sm := newTestSphincsManager(t)
	_, err := sm.CheckTimestampNonce([]byte{1, 2, 3}, []byte{4, 5, 6})
	_ = err
}

// ---------------------------------------------------------------------------
// LoadCommitment (nil DB)
// ---------------------------------------------------------------------------

func TestLoadCommitment_NilDB_NoPanel(t *testing.T) {
	sm := newTestSphincsManager(t)
	_, err := sm.LoadCommitment()
	_ = err
}

// ---------------------------------------------------------------------------
// DeserializeSignature error paths
// ---------------------------------------------------------------------------

func TestDeserializeSignature_InvalidBytes_Error(t *testing.T) {
	sm := newTestSphincsManager(t)
	_, err := sm.DeserializeSignature([]byte("not a valid sphincs sig"))
	if err == nil {
		t.Fatal("expected error deserializing invalid sig bytes")
	}
}

func TestDeserializeSignature_EmptyBytes_Error(t *testing.T) {
	sm := newTestSphincsManager(t)
	_, err := sm.DeserializeSignature([]byte{})
	if err == nil {
		t.Fatal("expected error deserializing empty bytes")
	}
}
