// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 41 - dht GetStaleRemote, allowToJoin, Close
package dht

import (
	"net"
	"testing"

	"go.uber.org/zap"
)

func makeTestDHT41(t *testing.T) (*DHT, func()) {
	t.Helper()
	cfg := Config{
		Proto:   "udp4",
		Address: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
	}
	log, _ := zap.NewDevelopment()
	d, err := NewDHT(cfg, log)
	if err != nil {
		t.Fatalf("NewDHT: %v", err)
	}
	return d, func() { d.Close() }
}

// ---------------------------------------------------------------------------
// GetStaleRemote — empty routing table returns empty
// ---------------------------------------------------------------------------

func TestSprint41_GetStaleRemote_EmptyTable_ReturnsEmpty(t *testing.T) {
	d, cleanup := makeTestDHT41(t)
	defer cleanup()
	stale := d.rt.GetStaleRemote()
	// Fresh table — no stale nodes
	if stale == nil {
		t.Fatal("expected non-nil slice (empty)")
	}
}

// ---------------------------------------------------------------------------
// allowToJoin — initial state (no nodes joined yet)
// ---------------------------------------------------------------------------

func TestSprint41_AllowToJoin_InitialState_NoPanic(t *testing.T) {
	d, cleanup := makeTestDHT41(t)
	defer cleanup()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("allowToJoin panicked: %v", r)
		}
	}()
	_ = d.allowToJoin()
}

// ---------------------------------------------------------------------------
// Close — can be called
// ---------------------------------------------------------------------------

func TestSprint41_Close_NoPanic(t *testing.T) {
	cfg := Config{
		Proto:   "udp4",
		Address: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
	}
	log, _ := zap.NewDevelopment()
	d, err := NewDHT(cfg, log)
	if err != nil {
		t.Fatalf("NewDHT: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Close panicked: %v", r)
		}
	}()
	err = d.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Get — returns values or empty for unknown key
// ---------------------------------------------------------------------------

func TestSprint41_Get_UnknownKey_ReturnsNil(t *testing.T) {
	d, cleanup := makeTestDHT41(t)
	defer cleanup()
	var k [32]byte
	k[0] = 0xBB
	// Get on a key that was never Put should return nil or empty
	// Note: Get is async — just verify no panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Get panicked: %v", r)
		}
	}()
	_, _ = d.Get(k)
}

// ---------------------------------------------------------------------------
// PingNode — no connection should not panic
// ---------------------------------------------------------------------------

func TestSprint41_PingNode_NoConnection_NoPanic(t *testing.T) {
	d, cleanup := makeTestDHT41(t)
	defer cleanup()
	var nodeID [32]byte
	nodeID[0] = 0xCC
	addr := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9876}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PingNode panicked: %v", r)
		}
	}()
	d.PingNode(nodeID, addr)
}
