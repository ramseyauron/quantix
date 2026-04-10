// Sprint 26b — p2p PeerManager: UpdatePeerScore, DisconnectPeer, BanPeer
// Note: BanPeer and UpdatePeerScore (negative path) both eventually call
// DisconnectPeer → pm.server.nodeManager.RemovePeer which panics with nil server.
// Tests are scoped to safe paths only.
package p2p

import (
	"testing"
)

// ─── PeerManager UpdatePeerScore ──────────────────────────────────────────────

func TestSprint26_PeerManager_UpdatePeerScore_UnknownPeer_NoPanic(t *testing.T) {
	pm := NewPeerManager(nil, 0)
	// Unknown peer → silently ignored (no panic)
	pm.UpdatePeerScore("ghost-peer", 10)
}

func TestSprint26_PeerManager_UpdatePeerScore_PositiveDelta_Accumulates(t *testing.T) {
	pm := NewPeerManager(nil, 0)
	pm.scores["p1"] = 50
	pm.peers["p1"] = nil
	pm.UpdatePeerScore("p1", 5)
	if pm.scores["p1"] != 55 {
		t.Errorf("score = %d, want 55", pm.scores["p1"])
	}
}

func TestSprint26_PeerManager_UpdatePeerScore_CapsAtHundred(t *testing.T) {
	pm := NewPeerManager(nil, 0)
	pm.scores["p2"] = 95
	pm.peers["p2"] = nil
	pm.UpdatePeerScore("p2", 20)
	if pm.scores["p2"] != 100 {
		t.Errorf("score = %d, want 100 (capped)", pm.scores["p2"])
	}
}

// ─── PeerManager DisconnectPeer ───────────────────────────────────────────────

func TestSprint26_PeerManager_DisconnectPeer_UnknownPeer_Error(t *testing.T) {
	pm := NewPeerManager(nil, 0)
	err := pm.DisconnectPeer("ghost-peer")
	if err == nil {
		t.Error("expected error for disconnecting unknown peer")
	}
}
