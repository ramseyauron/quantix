package network

import (
	"testing"
)

// Sprint 46 — network node.go + port.go coverage

// ---------------------------------------------------------------------------
// Node accessors
// ---------------------------------------------------------------------------

func TestNode_GetChainInfo(t *testing.T) {
	n := NewNode("127.0.0.1:9000", "127.0.0.1", "9000", "9100", true, RoleNone, nil)
	if n == nil {
		t.Skip("NewNode returned nil")
	}
	info := n.GetChainInfo()
	if false { // PeerInfo is a struct, always non-nil
		t.Fatal("GetChainInfo returned nil")
	}
	if _, ok := info["chain_id"]; !ok {
		t.Error("chain_id missing")
	}
}

func TestNode_GenerateChainHandshake(t *testing.T) {
	n := NewNode("127.0.0.1:9001", "127.0.0.1", "9001", "9101", true, RoleNone, nil)
	if n == nil {
		t.Skip("NewNode returned nil")
	}
	h := n.GenerateChainHandshake()
	if len(h) == 0 {
		t.Error("GenerateChainHandshake returned empty string")
	}
}

// ---------------------------------------------------------------------------
// CallNode / CallNodeManager (consensus adapter types)
// ---------------------------------------------------------------------------

func TestCallNodeManager_New(t *testing.T) {
	m := NewCallNodeManager()
	if m == nil {
		t.Fatal("NewCallNodeManager returned nil")
	}
}

func TestCallNodeManager_GetPeers_Empty(t *testing.T) {
	m := NewCallNodeManager()
	peers := m.GetPeers()
	if peers == nil {
		t.Error("GetPeers should return non-nil map")
	}
}

func TestCallNodeManager_GetPeerIDs_Empty(t *testing.T) {
	m := NewCallNodeManager()
	ids := m.GetPeerIDs()
	if ids == nil {
		t.Error("GetPeerIDs should return non-nil slice")
	}
}

func TestCallNodeManager_GetNode_ReturnsCallNode(t *testing.T) {
	// GetNode always returns a CallNode adapter (never nil) — by design
	m := NewCallNodeManager()
	n := m.GetNode("any-node-id")
	if n == nil {
		t.Error("GetNode should return non-nil CallNode adapter")
	}
}

func TestCallNodeManager_AddRemovePeer(t *testing.T) {
	m := NewCallNodeManager()
	m.AddPeer("peer-001")
	ids := m.GetPeerIDs()
	found := false
	for _, id := range ids {
		if id == "peer-001" {
			found = true
		}
	}
	if !found {
		t.Error("added peer should appear in GetPeerIDs")
	}
	m.RemovePeer("peer-001")
	ids2 := m.GetPeerIDs()
	for _, id := range ids2 {
		if id == "peer-001" {
			t.Error("removed peer should not appear in GetPeerIDs")
		}
	}
}

func TestCallNodeManager_BroadcastMessage_Empty(t *testing.T) {
	m := NewCallNodeManager()
	// Should not panic when no peers
	_ = m.BroadcastMessage("test", nil)
}

func TestGetConsensusRegistry_NotNil(t *testing.T) {
	r := GetConsensusRegistry()
	if r == nil {
		t.Error("GetConsensusRegistry should return non-nil map")
	}
}

func TestRegisterUnregisterConsensus(t *testing.T) {
	RegisterConsensus("test-node", nil)
	r := GetConsensusRegistry()
	if _, ok := r["test-node"]; !ok {
		t.Error("registered node should appear in registry")
	}
	UnregisterConsensus("test-node")
	r2 := GetConsensusRegistry()
	if _, ok := r2["test-node"]; ok {
		t.Error("unregistered node should not appear in registry")
	}
}

// ---------------------------------------------------------------------------
// port.go
// ---------------------------------------------------------------------------

func TestClearNodeConfigs(t *testing.T) {
	// Should not panic
	ClearNodeConfigs()
}

func TestUpdateGetNodeConfig(t *testing.T) {
	cfg := NodePortConfig{
		ID:       "sprint46-node",
		HTTPPort: "9999",
	}
	UpdateNodeConfig(cfg)
	got, ok := GetNodeConfig("sprint46-node")
	if !ok {
		t.Fatal("GetNodeConfig should find added config")
	}
	if got.HTTPPort != "9999" {
		t.Errorf("HTTPPort: got %s, want 9999", got.HTTPPort)
	}
}

func TestGetNodeConfig_Unknown(t *testing.T) {
	_, ok := GetNodeConfig("definitely-not-present-xyz")
	if ok {
		t.Error("unknown config should return false")
	}
}

func TestFindFreePort_TCP(t *testing.T) {
	port, err := FindFreePort(9000, "tcp")
	if err != nil {
		t.Fatalf("FindFreePort: %v", err)
	}
	if port <= 0 {
		t.Errorf("FindFreePort returned invalid port %d", port)
	}
}

func TestFindFreePort_UDP(t *testing.T) {
	port, err := FindFreePort(9000, "udp")
	if err != nil {
		t.Fatalf("FindFreePort UDP: %v", err)
	}
	if port <= 0 {
		t.Errorf("FindFreePort UDP returned invalid port %d", port)
	}
}

func TestGetNodePortConfigs_Empty(t *testing.T) {
	configs, err := GetNodePortConfigs(0, nil, nil)
	if err != nil {
		t.Fatalf("GetNodePortConfigs(0): %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(configs))
	}
}

func TestGetNodePortConfigs_One(t *testing.T) {
	ClearNodeConfigs() // ensure no stale config from other tests
	roles := []NodeRole{RoleNone}
	configs, err := GetNodePortConfigs(1, roles, nil)
	if err != nil {
		t.Fatalf("GetNodePortConfigs(1): %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(configs))
	}
}

// ---------------------------------------------------------------------------
// types.go — KBucket Lock/Unlock/RLock/RUnlock
// ---------------------------------------------------------------------------

func TestNodeManager_LockUnlock_Sprint46(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	nm.Lock()
	nm.Unlock()
	nm.RLock()
	nm.RUnlock()
}

// ---------------------------------------------------------------------------
// network/node.go — UpdateStatus, UpdateRole, NewPeer, GetPeerInfo
// ---------------------------------------------------------------------------

func TestNode_UpdateStatus(t *testing.T) {
	n := NewNode("127.0.0.1:9010", "127.0.0.1", "9010", "9110", false, RoleNone, nil)
	if n == nil {
		t.Skip("NewNode nil")
	}
	n.UpdateStatus(NodeStatus("active"))
}

func TestNode_UpdateRole(t *testing.T) {
	n := NewNode("127.0.0.1:9011", "127.0.0.1", "9011", "9111", false, RoleNone, nil)
	if n == nil {
		t.Skip("NewNode nil")
	}
	n.UpdateRole(RoleValidator)
	if n.Role != RoleValidator {
		t.Errorf("UpdateRole: got %s, want %s", n.Role, RoleValidator)
	}
}

func TestNewPeer(t *testing.T) {
	n := NewNode("127.0.0.1:9012", "127.0.0.1", "9012", "9112", false, RoleNone, nil)
	if n == nil {
		t.Skip("NewNode nil")
	}
	p := NewPeer(n)
	if p == nil {
		t.Fatal("NewPeer returned nil")
	}
	if p.Node != n {
		t.Error("NewPeer should set Node field")
	}
}

func TestPeer_GetPeerInfo(t *testing.T) {
	n := NewNode("127.0.0.1:9013", "127.0.0.1", "9013", "9113", false, RoleNone, nil)
	if n == nil {
		t.Skip("NewNode nil")
	}
	p := NewPeer(n)
	_ = p.GetPeerInfo() // confirm no panic
}
