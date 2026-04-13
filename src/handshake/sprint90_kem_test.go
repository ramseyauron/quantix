// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 90 — handshake 51.6%→higher
// Tests: PerformKEM full roundtrip (initiator+responder), PerformHandshake full roundtrip
package security_test

import (
	"net"
	"testing"

	security "github.com/ramseyauron/quantix/src/handshake"
)

// ---------------------------------------------------------------------------
// PerformKEM — full roundtrip via net.Pipe
// ---------------------------------------------------------------------------

func TestSprint90_PerformKEM_FullRoundtrip_DerivesSameKey(t *testing.T) {
	// Create an in-memory network connection pair
	initiatorConn, responderConn, err := netPipe90(t)
	if err != nil {
		t.Fatalf("net.Pipe: %v", err)
	}
	defer initiatorConn.Close()
	defer responderConn.Close()

	type kemResult struct {
		key *security.EncryptionKey
		err error
	}

	initiatorCh := make(chan kemResult, 1)
	responderCh := make(chan kemResult, 1)

	go func() {
		key, err := security.PerformKEM(initiatorConn, true)
		initiatorCh <- kemResult{key, err}
	}()
	go func() {
		key, err := security.PerformKEM(responderConn, false)
		responderCh <- kemResult{key, err}
	}()

	iRes := <-initiatorCh
	rRes := <-responderCh

	if iRes.err != nil {
		t.Fatalf("initiator PerformKEM error: %v", iRes.err)
	}
	if rRes.err != nil {
		t.Fatalf("responder PerformKEM error: %v", rRes.err)
	}
	if iRes.key == nil {
		t.Fatal("initiator key is nil")
	}
	if rRes.key == nil {
		t.Fatal("responder key is nil")
	}

	// Both sides must derive the same AES-256 key
	iBytes := iRes.key.SharedSecret
	rBytes := rRes.key.SharedSecret
	if len(iBytes) == 0 {
		t.Errorf("initiator key is empty, got len=%d", len(iBytes))
	}
	for i, b := range iBytes {
		if b != rBytes[i] {
			t.Errorf("key mismatch at byte %d: initiator=%02x responder=%02x", i, b, rBytes[i])
			break
		}
	}
}

// ---------------------------------------------------------------------------
// PerformKEM — two independent rounds produce different keys
// ---------------------------------------------------------------------------

func TestSprint90_PerformKEM_TwoRounds_DifferentKeys(t *testing.T) {
	key1 := performKEMOnce(t)
	key2 := performKEMOnce(t)

	for i := range key1 {
		if key1[i] != key2[i] {
			return // different — expected
		}
	}
	t.Error("two KEM rounds produced identical keys (extremely unlikely if random)")
}

func performKEMOnce(t *testing.T) []byte {
	t.Helper()
	c1, c2, err := netPipe90(t)
	if err != nil {
		t.Fatalf("net.Pipe: %v", err)
	}
	defer c1.Close()
	defer c2.Close()

	type result struct {
		key *security.EncryptionKey
		err error
	}
	ch := make(chan result, 1)
	go func() {
		k, e := security.PerformKEM(c2, false)
		ch <- result{k, e}
	}()
	key, err := security.PerformKEM(c1, true)
	<-ch
	if err != nil {
		t.Fatalf("PerformKEM: %v", err)
	}
	return key.SharedSecret
}

// ---------------------------------------------------------------------------
// PerformHandshake — full roundtrip via net.Pipe
// ---------------------------------------------------------------------------

func TestSprint90_PerformHandshake_FullRoundtrip(t *testing.T) {
	c1, c2, err := netPipe90(t)
	if err != nil {
		t.Fatalf("net.Pipe: %v", err)
	}
	defer c1.Close()
	defer c2.Close()

	h := security.NewHandshake()

	type result struct {
		key *security.EncryptionKey
		err error
	}
	respCh := make(chan result, 1)
	go func() {
		k, e := h.PerformHandshake(c2, "p2p", false)
		respCh <- result{k, e}
	}()

	initKey, initErr := h.PerformHandshake(c1, "p2p", true)
	respRes := <-respCh

	if initErr != nil {
		t.Fatalf("initiator PerformHandshake error: %v", initErr)
	}
	if respRes.err != nil {
		t.Fatalf("responder PerformHandshake error: %v", respRes.err)
	}
	if initKey == nil {
		t.Fatal("initiator key is nil")
	}
	if respRes.key == nil {
		t.Fatal("responder key is nil")
	}

	// Keys should match
	iBytes := initKey.SharedSecret
	rBytes := respRes.key.SharedSecret
	for i, b := range iBytes {
		if b != rBytes[i] {
			t.Errorf("handshake key mismatch at byte %d", i)
			break
		}
	}
}

// ---------------------------------------------------------------------------
// PerformHandshake — nil conn returns error
// ---------------------------------------------------------------------------

func TestSprint90_PerformHandshake_NilConn_ReturnsError(t *testing.T) {
	h := security.NewHandshake()
	_, err := h.PerformHandshake(nil, "test", true)
	if err == nil {
		t.Error("expected error with nil conn, got nil")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func netPipe90(t *testing.T) (net.Conn, net.Conn, error) {
	t.Helper()
	c1, c2 := net.Pipe()
	return c1, c2, nil
}
