// MIT License
// Copyright (c) 2024 quantix

// P3-Q1: PEPPER targeted tests for src/network — Key, NodeManager, types coverage.
package network

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Key tests
// ---------------------------------------------------------------------------

func TestKey_IsEmpty_Zero(t *testing.T) {
	var k Key
	if !k.IsEmpty() {
		t.Error("zero key should be empty")
	}
}

func TestKey_IsEmpty_NonZero(t *testing.T) {
	var k Key
	k[31] = 1
	if k.IsEmpty() {
		t.Error("non-zero key should not be empty")
	}
}

func TestKey_Short_Format(t *testing.T) {
	var k Key
	k[30] = 0xAB
	k[31] = 0xCD
	s := k.Short()
	if len(s) == 0 {
		t.Error("Short() should return non-empty string")
	}
}

func TestKey_String_IsHex(t *testing.T) {
	var k Key
	k[0] = 0xFF
	s := k.String()
	if len(s) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("String() expected 64 chars, got %d", len(s))
	}
}

func TestKey_Equal_SameKey(t *testing.T) {
	var a, b Key
	a[0] = 0x55
	b[0] = 0x55
	if !a.Equal(b) {
		t.Error("equal keys should be equal")
	}
}

func TestKey_Equal_DifferentKey(t *testing.T) {
	var a, b Key
	a[0] = 0x01
	b[0] = 0x02
	if a.Equal(b) {
		t.Error("different keys should not be equal")
	}
}

func TestKey_Less(t *testing.T) {
	var a, b Key
	a[0] = 0x01
	b[0] = 0x02
	if !a.Less(b) {
		t.Error("a should be less than b")
	}
	if b.Less(a) {
		t.Error("b should not be less than a")
	}
}

func TestKey_CommonPrefixLength_SameKey(t *testing.T) {
	var k Key
	k[0] = 0xAB
	cpl := k.CommonPrefixLength(k)
	if cpl != 256 {
		t.Errorf("same key CPL should be 256, got %d", cpl)
	}
}

func TestKey_CommonPrefixLength_DifferHighBit(t *testing.T) {
	var a, b Key
	a[0] = 0x80 // 1000 0000
	b[0] = 0x00 // 0000 0000
	cpl := a.CommonPrefixLength(b)
	if cpl != 0 {
		t.Errorf("expected 0 common prefix, got %d", cpl)
	}
}

func TestKey_Distance_XOR(t *testing.T) {
	var a, b Key
	a[0] = 0x0F
	b[0] = 0x01
	var d Key
	d.Distance(a, b)
	if d[0] != 0x0E {
		t.Errorf("expected XOR 0x0E, got 0x%02x", d[0])
	}
}

func TestKey_FromString_HashesInput(t *testing.T) {
	var k Key
	if err := k.FromString("hello-world"); err != nil {
		t.Fatalf("FromString: %v", err)
	}
	// Should produce a non-zero key
	if k.IsEmpty() {
		t.Error("expected non-zero key from FromString")
	}
}

func TestKey_FromString_Deterministic(t *testing.T) {
	var a, b Key
	_ = a.FromString("same-input")
	_ = b.FromString("same-input")
	if !a.Equal(b) {
		t.Error("FromString should be deterministic")
	}
}

func TestGenerateKademliaID_Deterministic(t *testing.T) {
	a := GenerateKademliaID("test-input")
	b := GenerateKademliaID("test-input")
	if a != b {
		t.Error("GenerateKademliaID should be deterministic")
	}
}

func TestGenerateKademliaID_DifferentInputs(t *testing.T) {
	a := GenerateKademliaID("alice")
	b := GenerateKademliaID("bob")
	if a == b {
		t.Error("different inputs should produce different IDs")
	}
}

func TestGetRandomNodeID_NonZero(t *testing.T) {
	id := GetRandomNodeID()
	var zero NodeID
	if id == zero {
		t.Error("random node ID should not be zero")
	}
}

// ---------------------------------------------------------------------------
// NodeManager basic tests (no DB)
// ---------------------------------------------------------------------------

func TestNewNodeManager_NotNil(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	if nm == nil {
		t.Fatal("NewNodeManager returned nil")
	}
}

func TestNodeManager_GetChainInfo(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	info := nm.GetChainInfo()
	if info["chain_id"] != 7331 {
		t.Errorf("expected chain_id=7331, got %v", info["chain_id"])
	}
	if info["symbol"] != "QTX" {
		t.Errorf("expected symbol=QTX, got %v", info["symbol"])
	}
}

func TestNodeManager_GenerateNodeIdentification(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	id := nm.GenerateNodeIdentification("node-test")
	if len(id) == 0 {
		t.Error("expected non-empty identification string")
	}
}

func TestNodeManager_ValidateChainCompatibility_Match(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	compatible := nm.ValidateChainCompatibility(map[string]interface{}{
		"chain_id": 7331,
	})
	if !compatible {
		t.Error("matching chain_id should be compatible")
	}
}

func TestNodeManager_ValidateChainCompatibility_Mismatch(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	compatible := nm.ValidateChainCompatibility(map[string]interface{}{
		"chain_id": 9999,
	})
	if compatible {
		t.Error("mismatched chain_id should not be compatible")
	}
}

func TestNodeManager_HasSeenMessage_MarkSeen(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	if nm.HasSeenMessage("msg-1") {
		t.Error("message should not be seen before marking")
	}
	nm.MarkMessageSeen("msg-1")
	if !nm.HasSeenMessage("msg-1") {
		t.Error("message should be seen after marking")
	}
}

func TestNodeManager_AddGetRemoveNode(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	node := &Node{
		ID:       "node-test-1",
		Address:  "127.0.0.1:9001",
		Status:   NodeStatusActive,
		LastSeen: time.Now(),
	}
	nm.AddNode(node)
	got := nm.GetNode("node-test-1")
	if got == nil {
		t.Fatal("expected to find added node")
	}
	nm.RemoveNode("node-test-1")
	got = nm.GetNode("node-test-1")
	if got != nil {
		t.Error("expected node to be removed")
	}
}

func TestNodeManager_GetPeers_Empty(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	peers := nm.GetPeers()
	if peers == nil {
		t.Error("GetPeers should return non-nil map")
	}
}

func TestNodeManager_LockUnlock(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	nm.Lock()
	nm.Unlock()
	nm.RLock()
	nm.RUnlock()
}

func TestNodeManager_CalculateDistance(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	var a, b NodeID
	a[0] = 0x0F
	b[0] = 0x01
	dist := nm.CalculateDistance(a, b)
	if dist[0] != 0x0E {
		t.Errorf("expected 0x0E, got 0x%02x", dist[0])
	}
}

func TestNodeManager_SelectValidator_NoPeers(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	// No panic expected
	v := nm.SelectValidator()
	_ = v
}
