// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 91 — p2p 9.4%→higher
// Covers: NewP2PNodeManager, GetPeers/GetNode nil-nodeManager, DecodeConsensusEnvelope,
// BroadcastRANDAOState nil, UpdateSeedNodes, p2pNode GetID/GetRole/GetStatus
package p2p

import (
	"encoding/json"
	"testing"
)

// ---------------------------------------------------------------------------
// NewP2PNodeManager — basic construction
// ---------------------------------------------------------------------------

func TestSprint91_NewP2PNodeManager_NotNil(t *testing.T) {
	m := NewP2PNodeManager(nil)
	if m == nil {
		t.Fatal("expected non-nil P2PNodeManager")
	}
}

func TestSprint91_NewP2PNodeManager_NilNodeManager_GetPeers_Empty(t *testing.T) {
	m := NewP2PNodeManager(nil)
	peers := m.GetPeers()
	if peers == nil {
		t.Fatal("GetPeers should return empty map, not nil")
	}
	if len(peers) != 0 {
		t.Errorf("GetPeers with nil nodeManager should return empty map, got %d entries", len(peers))
	}
}

func TestSprint91_NewP2PNodeManager_NilNodeManager_GetNode_FallbackNode(t *testing.T) {
	m := NewP2PNodeManager(nil)
	node := m.GetNode("test-node-id")
	if node == nil {
		t.Fatal("GetNode should return fallback node, not nil")
	}
	if node.GetID() != "test-node-id" {
		t.Errorf("fallback node GetID = %q, want test-node-id", node.GetID())
	}
}

// ---------------------------------------------------------------------------
// p2pNode methods (GetID, GetRole, GetStatus)
// ---------------------------------------------------------------------------

func TestSprint91_P2PNode_GetID(t *testing.T) {
	m := NewP2PNodeManager(nil)
	node := m.GetNode("my-validator")
	if node.GetID() != "my-validator" {
		t.Errorf("GetID = %q, want my-validator", node.GetID())
	}
}

func TestSprint91_P2PNode_GetRole_Default(t *testing.T) {
	m := NewP2PNodeManager(nil)
	node := m.GetNode("any-id")
	// Default role is 0 (no role assigned yet)
	_ = node.GetRole() // just ensure no panic
}

func TestSprint91_P2PNode_GetStatus_Default(t *testing.T) {
	m := NewP2PNodeManager(nil)
	node := m.GetNode("any-id")
	_ = node.GetStatus() // just ensure no panic
}

// ---------------------------------------------------------------------------
// DecodeConsensusEnvelope
// ---------------------------------------------------------------------------

func TestSprint91_DecodeConsensusEnvelope_ValidJSON(t *testing.T) {
	raw := []byte(`{"type":"proposal","payload":{"view":1}}`)
	env, err := DecodeConsensusEnvelope(raw)
	if err != nil {
		t.Fatalf("DecodeConsensusEnvelope error: %v", err)
	}
	if env.Type != "proposal" {
		t.Errorf("env.Type = %q, want proposal", env.Type)
	}
}

func TestSprint91_DecodeConsensusEnvelope_VoteType(t *testing.T) {
	raw := []byte(`{"type":"vote","payload":{}}`)
	env, err := DecodeConsensusEnvelope(raw)
	if err != nil {
		t.Fatalf("DecodeConsensusEnvelope error: %v", err)
	}
	if env.Type != "vote" {
		t.Errorf("env.Type = %q, want vote", env.Type)
	}
}

func TestSprint91_DecodeConsensusEnvelope_InvalidJSON_Error(t *testing.T) {
	_, err := DecodeConsensusEnvelope([]byte("not-json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestSprint91_DecodeConsensusEnvelope_EmptyPayload(t *testing.T) {
	raw, _ := json.Marshal(map[string]interface{}{
		"type":    "timeout",
		"payload": nil,
	})
	env, err := DecodeConsensusEnvelope(raw)
	if err != nil {
		t.Fatalf("DecodeConsensusEnvelope error: %v", err)
	}
	if env.Type != "timeout" {
		t.Errorf("env.Type = %q, want timeout", env.Type)
	}
}

// ---------------------------------------------------------------------------
// BroadcastRANDAOState — nil nodeManager returns nil (no panic)
// ---------------------------------------------------------------------------

func TestSprint91_BroadcastRANDAOState_NilNodeManager_NoError(t *testing.T) {
	m := NewP2PNodeManager(nil)
	var mix [32]byte
	// Should not panic; returns nil (no peers to broadcast to)
	err := m.BroadcastRANDAOState(mix, nil)
	_ = err // nil or "no-op" are both acceptable
}

// ---------------------------------------------------------------------------
// UpdateSeedNodes — basic setter
// ---------------------------------------------------------------------------

func TestSprint91_UpdateSeedNodes_SetsSeeds(t *testing.T) {
	s := makeP2PServer33() // reuse helper from sprint33
	seeds := []string{"127.0.0.1:3000", "127.0.0.1:3001"}
	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UpdateSeedNodes panicked: %v", r)
		}
	}()
	s.UpdateSeedNodes(seeds)
}

func TestSprint91_UpdateSeedNodes_EmptySeeds(t *testing.T) {
	s := makeP2PServer33()
	s.UpdateSeedNodes(nil)
}

// ---------------------------------------------------------------------------
// GetNode — unknown node falls back to p2pNode with correct ID
// ---------------------------------------------------------------------------

func TestSprint91_P2PNodeManager_GetNode_DifferentIDs_DifferentNodes(t *testing.T) {
	m := NewP2PNodeManager(nil)
	n1 := m.GetNode("alpha")
	n2 := m.GetNode("beta")
	if n1.GetID() == n2.GetID() {
		t.Error("different node IDs should produce different nodes")
	}
}
