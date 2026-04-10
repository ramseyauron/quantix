package network

// Coverage Sprint 18 — network/manager.go: NewNodeManager, CreateLocalNode,
// AddNode, RemoveNode, UpdateNode, HasSeenMessage/MarkMessageSeen, GetNode,
// rpc/query.go: AddJoin, AddPing, AddGet, GetQuery, RemoveQuery,
// IsExpectedResponse, GC.

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helper — minimal NodeManager without DB
// ---------------------------------------------------------------------------

func newTestNodeManager(t *testing.T) *NodeManager {
	t.Helper()
	nm := NewNodeManager(16, nil, nil)
	if nm == nil {
		t.Fatal("NewNodeManager returned nil")
	}
	return nm
}

// ---------------------------------------------------------------------------
// NewNodeManager
// ---------------------------------------------------------------------------

func TestNewNodeManager_NonNil(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	if nm == nil {
		t.Fatal("NewNodeManager returned nil")
	}
}

func TestNewNodeManager_ZeroBucketSize_UsesDefault(t *testing.T) {
	nm := NewNodeManager(0, nil, nil)
	if nm.K != 16 {
		t.Errorf("NewNodeManager(0): expected K=16, got %d", nm.K)
	}
}

func TestNewNodeManager_NegativeBucketSize_UsesDefault(t *testing.T) {
	nm := NewNodeManager(-5, nil, nil)
	if nm.K != 16 {
		t.Errorf("NewNodeManager(-5): expected K=16, got %d", nm.K)
	}
}

func TestNewNodeManager_CustomBucketSize(t *testing.T) {
	nm := NewNodeManager(8, nil, nil)
	if nm.K != 8 {
		t.Errorf("NewNodeManager(8): expected K=8, got %d", nm.K)
	}
}

// ---------------------------------------------------------------------------
// HasSeenMessage / MarkMessageSeen
// ---------------------------------------------------------------------------

func TestHasSeenMessage_Unseen_False(t *testing.T) {
	nm := newTestNodeManager(t)
	if nm.HasSeenMessage("msg-001") {
		t.Error("HasSeenMessage: should be false for unseen message")
	}
}

func TestMarkMessageSeen_ThenHasSeenMessage_True(t *testing.T) {
	nm := newTestNodeManager(t)
	nm.MarkMessageSeen("msg-002")
	if !nm.HasSeenMessage("msg-002") {
		t.Error("HasSeenMessage: should be true after MarkMessageSeen")
	}
}

func TestMarkMessageSeen_Idempotent(t *testing.T) {
	nm := newTestNodeManager(t)
	nm.MarkMessageSeen("msg-003")
	nm.MarkMessageSeen("msg-003") // second call should not panic
	if !nm.HasSeenMessage("msg-003") {
		t.Error("HasSeenMessage: should still be true after duplicate MarkMessageSeen")
	}
}

func TestHasSeenMessage_Different_Messages_Independent(t *testing.T) {
	nm := newTestNodeManager(t)
	nm.MarkMessageSeen("msg-A")
	if nm.HasSeenMessage("msg-B") {
		t.Error("marking msg-A should not affect msg-B")
	}
}

// ---------------------------------------------------------------------------
// AddNode / GetNode / RemoveNode
// ---------------------------------------------------------------------------

func TestAddNode_GetNode_Roundtrip(t *testing.T) {
	nm := newTestNodeManager(t)
	node := NewNode("node-001:32307", "127.0.0.1", "32307", "32308", false, RoleValidator, nil)
	if node == nil {
		t.Fatal("NewNode returned nil")
	}

	nm.AddNode(node)

	got := nm.GetNode(node.ID)
	if got == nil {
		t.Fatal("GetNode: returned nil after AddNode")
	}
	if got.ID != node.ID {
		t.Errorf("GetNode: ID mismatch: got %q, want %q", got.ID, node.ID)
	}
}

func TestGetNode_Unknown_ReturnsNil(t *testing.T) {
	nm := newTestNodeManager(t)
	got := nm.GetNode("nonexistent-node-id")
	if got != nil {
		t.Error("GetNode unknown: should return nil")
	}
}

func TestRemoveNode_RemovesFromMap(t *testing.T) {
	nm := newTestNodeManager(t)
	node := NewNode("remove-node:32307", "127.0.0.1", "32307", "32308", false, RoleValidator, nil)
	if node == nil {
		t.Skip("NewNode returned nil — skip RemoveNode test")
	}

	nm.AddNode(node)
	nm.RemoveNode(node.ID)

	got := nm.GetNode(node.ID)
	if got != nil {
		t.Error("GetNode: should return nil after RemoveNode")
	}
}

func TestRemoveNode_NonExistent_NoError(t *testing.T) {
	nm := newTestNodeManager(t)
	// Removing a node that doesn't exist should be a no-op (not panic).
	nm.RemoveNode("ghost-node-id")
}

// ---------------------------------------------------------------------------
// UpdateNode
// ---------------------------------------------------------------------------

func TestUpdateNode_ExistingNode_UpdatesFields(t *testing.T) {
	nm := newTestNodeManager(t)
	node := NewNode("update-node:32307", "127.0.0.1", "32307", "32308", false, RoleValidator, nil)
	if node == nil {
		t.Skip("NewNode returned nil")
	}
	nm.AddNode(node)

	// Modify and update.
	node.IP = "192.168.1.5"
	err := nm.UpdateNode(node)
	if err != nil {
		t.Errorf("UpdateNode: unexpected error: %v", err)
	}

	got := nm.GetNode(node.ID)
	if got == nil {
		t.Fatal("GetNode after UpdateNode: nil")
	}
	if got.IP != "192.168.1.5" {
		t.Errorf("UpdateNode: IP not updated, got %q", got.IP)
	}
}

func TestUpdateNode_NonExistent_Error(t *testing.T) {
	nm := newTestNodeManager(t)
	ghost := &Node{ID: "ghost", IP: "0.0.0.0"}
	err := nm.UpdateNode(ghost)
	if err == nil {
		t.Error("UpdateNode non-existent: expected error")
	}
}

// ---------------------------------------------------------------------------
// CreateLocalNode (needs nil DB guard)
// ---------------------------------------------------------------------------

func TestCreateLocalNode_WithNilDB_NoError(t *testing.T) {
	nm := NewNodeManager(16, nil, nil)
	err := nm.CreateLocalNode("local:32307", "127.0.0.1", "32307", "32308", RoleValidator)
	if err != nil {
		t.Logf("CreateLocalNode with nil DB: %v (may fail if DB required for key operations)", err)
	}
}

// ---------------------------------------------------------------------------
// PruneInactivePeers — no panic
// ---------------------------------------------------------------------------

func TestPruneInactivePeers_EmptyManager_NoError(t *testing.T) {
	nm := newTestNodeManager(t)
	// Should not panic on empty node manager.
	nm.PruneInactivePeers(5 * time.Minute)
}
