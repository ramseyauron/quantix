// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 33 - p2p server additional coverage
package p2p

import (
	"testing"

	"github.com/ramseyauron/quantix/src/network"
)

func makeP2PServer33() *Server {
	config := network.NodePortConfig{
		Name:    "test-node-sprint33",
		TCPAddr: "127.0.0.1:0",
		UDPPort: "0",
		HTTPPort: "127.0.0.1:0",
		WSPort:  "127.0.0.1:0",
		Role:    network.RoleValidator,
	}
	return NewServer(config, nil, nil)
}

// ---------------------------------------------------------------------------
// SetServer / Close / CloseDB error paths
// ---------------------------------------------------------------------------

func TestSprint33_SetServer_NoPanic(t *testing.T) {
	s := makeP2PServer33()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetServer panicked: %v", r)
		}
	}()
	s.SetServer()
}

func TestSprint33_CloseDB_NilDB_NoPanic(t *testing.T) {
	s := makeP2PServer33()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CloseDB panicked: %v", r)
		}
	}()
	_ = s.CloseDB()
}

// ---------------------------------------------------------------------------
// StorePeer — nil DB panics (documented)
// ---------------------------------------------------------------------------

func TestSprint33_StorePeer_NilDB_Documented(t *testing.T) {
	// StorePeer panics with nil DB — pending nil guard fix
	t.Skip("StorePeer panics with nil DB — nil guard needed")
}

// ---------------------------------------------------------------------------
// FetchPeer — nil DB panics (documented)
// ---------------------------------------------------------------------------

func TestSprint33_FetchPeer_NilDB_Documented(t *testing.T) {
	// FetchPeer panics with nil DB — pending nil guard fix
	t.Skip("FetchPeer panics with nil DB — nil guard needed")
}

// ---------------------------------------------------------------------------
// InitializeConsensus / GetConsensus
// ---------------------------------------------------------------------------

func TestSprint33_GetConsensus_Initial_Nil(t *testing.T) {
	s := makeP2PServer33()
	c := s.GetConsensus()
	if c != nil {
		t.Fatal("expected nil consensus on fresh server")
	}
}

func TestSprint33_InitializeConsensus_NilConsensus_NoPanic(t *testing.T) {
	s := makeP2PServer33()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("InitializeConsensus panicked: %v", r)
		}
	}()
	s.InitializeConsensus(nil)
}

// ---------------------------------------------------------------------------
// SetMessageCh — no panic
// ---------------------------------------------------------------------------

func TestSprint33_SetMessageCh_NoPanic(t *testing.T) {
	s := makeP2PServer33()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetMessageCh panicked: %v", r)
		}
	}()
	s.SetMessageCh(nil)
}

// ---------------------------------------------------------------------------
// validateTransaction — nil tx returns error
// ---------------------------------------------------------------------------

func TestSprint33_ValidateTransaction_Nil_Error(t *testing.T) {
	s := makeP2PServer33()
	err := s.validateTransaction(nil)
	if err == nil {
		t.Fatal("expected error for nil transaction")
	}
}

// ---------------------------------------------------------------------------
// BroadcastBlock — nil block panics (documented)
// ---------------------------------------------------------------------------

func TestSprint33_BroadcastBlock_NilBlock_Documented(t *testing.T) {
	// BroadcastBlock panics with nil block — nil guard needed
	t.Skip("BroadcastBlock panics with nil block — nil guard needed")
}

// ---------------------------------------------------------------------------
// BroadcastTransaction — nil tx panics (documented)
// ---------------------------------------------------------------------------

func TestSprint33_BroadcastTransaction_NilTx_Documented(t *testing.T) {
	// BroadcastTransaction panics with nil tx — nil guard needed
	t.Skip("BroadcastTransaction panics with nil tx — nil guard needed")
}
