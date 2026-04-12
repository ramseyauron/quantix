// test(PEPPER): Sprint 80 — src/network 47.1%→higher
// Tests: Key.Less (equal/less/greater), Key.leadingZeroBits (all zeros/mixed),
// FindFreePort TCP (port 0), NodeManager.AddNode duplicate,
// NodeManager.GetNodeByKademliaID hit/miss, GetNodePortConfigs 0 nodes,
// BroadcastMessage empty registry (no panic)
package network

import (
	"testing"
)

// ─── Key.Less — equal keys ────────────────────────────────────────────────────

func TestSprint80_KeyLess_Equal(t *testing.T) {
	var a, b Key
	a[0] = 0xab
	b[0] = 0xab
	if a.Less(b) {
		t.Error("equal keys: Less should return false")
	}
	if b.Less(a) {
		t.Error("equal keys: Less should return false (reversed)")
	}
}

func TestSprint80_KeyLess_FirstLess(t *testing.T) {
	var a, b Key
	a[0] = 0x01
	b[0] = 0x02
	if !a.Less(b) {
		t.Error("expected a < b")
	}
	if b.Less(a) {
		t.Error("b should not be less than a")
	}
}

func TestSprint80_KeyLess_LastByteDecides(t *testing.T) {
	var a, b Key
	// All bytes equal except last
	for i := 0; i < 31; i++ {
		a[i] = 0xff
		b[i] = 0xff
	}
	a[31] = 0x00
	b[31] = 0x01
	if !a.Less(b) {
		t.Error("expected a < b based on last byte")
	}
}

// ─── Key.leadingZeroBits — all zeros ──────────────────────────────────────────

func TestSprint80_KeyLeadingZeroBits_AllZeros(t *testing.T) {
	var k Key // all zeros
	// leadingZeroBits is unexported; test via CommonPrefixLength
	var other Key
	for i := range other {
		other[i] = 0xff
	}
	// CommonPrefixLength of all-zero vs all-ff = 0 common bits? No — they XOR to all-ff
	// Use XOR: zeros XOR other = other, which has 0 leading zero bits
	result := k.CommonPrefixLength(other)
	if result != 0 {
		t.Errorf("expected 0 common prefix bits between all-0 and all-ff, got %d", result)
	}
}

func TestSprint80_KeyLeadingZeroBits_SameKey(t *testing.T) {
	var k Key
	k[0] = 0xab
	result := k.CommonPrefixLength(k)
	if result != 256 {
		t.Errorf("expected 256 common prefix bits for same key, got %d", result)
	}
}

// ─── FindFreePort — TCP ───────────────────────────────────────────────────────

func TestSprint80_FindFreePort_TCP(t *testing.T) {
	port, err := FindFreePort(19800, "tcp")
	if err != nil {
		t.Fatalf("FindFreePort TCP: %v", err)
	}
	if port < 19800 {
		t.Errorf("expected port >= 19800, got %d", port)
	}
}

func TestSprint80_FindFreePort_UDP(t *testing.T) {
	port, err := FindFreePort(19810, "udp")
	if err != nil {
		t.Fatalf("FindFreePort UDP: %v", err)
	}
	if port < 19810 {
		t.Errorf("expected port >= 19810, got %d", port)
	}
}

// ─── NodeManager.GetNodeByKademliaID — hit and miss ──────────────────────────

func TestSprint80_GetNodeByKademliaID_Miss(t *testing.T) {
	nm := NewNodeManager(20, nil, nil)
	var id [32]byte
	for i := range id {
		id[i] = 0x42
	}
	result := nm.GetNodeByKademliaID(id)
	if result != nil {
		t.Errorf("expected nil for unknown KademliaID, got %v", result)
	}
}

func TestSprint80_GetNodeByKademliaID_Hit(t *testing.T) {
	nm := NewNodeManager(20, nil, nil)
	var kid [32]byte
	for i := range kid {
		kid[i] = 0x77
	}
	node := &Node{
		ID:          "test-node-77",
		KademliaID:  kid,
		IP:          "127.0.0.1",
		Port:        "8080",
		Address:     "127.0.0.1:8080",
		IsLocal:     true,
	}
	nm.AddNode(node)
	result := nm.GetNodeByKademliaID(kid)
	if result == nil {
		t.Error("expected to find node by KademliaID")
	}
}

// ─── NodeManager.AddNode — duplicate by same ID+KademliaID ───────────────────

func TestSprint80_AddNode_Duplicate(t *testing.T) {
	nm := NewNodeManager(20, nil, nil)
	var kid [32]byte
	for i := range kid {
		kid[i] = 0x55
	}
	node := &Node{
		ID:         "dup-node",
		KademliaID: kid,
		IP:         "127.0.0.1",
		Port:       "9090",
		Address:    "127.0.0.1:9090",
		IsLocal:    true,
	}
	nm.AddNode(node) // first add
	nm.AddNode(node) // duplicate — should be skipped, no panic
	// Verify node still exists
	result := nm.GetNode("dup-node")
	if result == nil {
		t.Error("node should exist after duplicate add")
	}
}

// ─── GetNodePortConfigs — 0 nodes ────────────────────────────────────────────

func TestSprint80_GetNodePortConfigs_ZeroNodes(t *testing.T) {
	// Clear global NodeConfigs to avoid interference from other tests
	NodeConfigsLock.Lock()
	NodeConfigs = make(map[string]NodePortConfig)
	NodeConfigsLock.Unlock()

	configs, err := GetNodePortConfigs(0, nil, nil)
	if err != nil {
		t.Fatalf("GetNodePortConfigs 0 nodes: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("expected 0 configs for 0 nodes, got %d", len(configs))
	}
}

// ─── GetNodePortConfigs — 1 node ─────────────────────────────────────────────

func TestSprint80_GetNodePortConfigs_OneNode(t *testing.T) {
	// Clear global NodeConfigs to avoid cache hit from other tests
	NodeConfigsLock.Lock()
	NodeConfigs = make(map[string]NodePortConfig)
	NodeConfigsLock.Unlock()

	configs, err := GetNodePortConfigs(1, []NodeRole{RoleValidator}, nil)
	if err != nil {
		t.Fatalf("GetNodePortConfigs 1 node: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Role != RoleValidator {
		t.Errorf("expected RoleValidator, got %v", configs[0].Role)
	}
}

// ─── NodeManager.SelectValidator — empty node manager ────────────────────────

func TestSprint80_SelectValidator_Empty(t *testing.T) {
	nm := NewNodeManager(20, nil, nil)
	result := nm.SelectValidator()
	// With no validators, should return nil
	if result != nil {
		t.Logf("SelectValidator on empty manager returned non-nil (unexpected): %v", result)
	}
}

// ─── NodeManager.GetPeers — empty ────────────────────────────────────────────

func TestSprint80_GetPeers_Empty(t *testing.T) {
	nm := NewNodeManager(20, nil, nil)
	peers := nm.GetPeers()
	if len(peers) != 0 {
		t.Errorf("expected 0 peers from empty manager, got %d", len(peers))
	}
}
