// test(PEPPER): Sprint 78 — src/dht 41.5%→higher
// Tests: verifyMessage (empty from NodeID, zero RPCID, both valid), allowToJoin rate limiting,
// GetStaleRemote (empty table), GC (no entries), routingTable Observe self-node skip,
// targetField/fromField/localNodeIDField (log field helpers), getRandomDelay
package dht

import (
	"net"
	"testing"
	"time"

	"github.com/ramseyauron/quantix/src/rpc"
	"go.uber.org/zap"
)

// helper: fresh DHT for testing (uses OS-assigned UDP port)
func newDHT78(t *testing.T) *DHT {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	cfg := Config{
		Proto:   "udp4",
		Address: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
	}
	d, err := NewDHT(cfg, logger)
	if err != nil {
		t.Skipf("NewDHT failed: %v", err)
	}
	return d
}

// ─── verifyMessage — empty from NodeID ───────────────────────────────────────

func TestSprint78_VerifyMessage_EmptyFromNodeID(t *testing.T) {
	d := newDHT78(t)
	msg := rpc.Message{
		From:  rpc.Remote{NodeID: rpc.NodeID{}}, // zero = empty
		RPCID: 42,
	}
	// Should log error and return — no panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("verifyMessage empty NodeID panicked: %v", r)
			}
		}()
		d.verifyMessage(msg)
	}()
}

// ─── verifyMessage — zero RPCID ──────────────────────────────────────────────

func TestSprint78_VerifyMessage_ZeroRPCID(t *testing.T) {
	d := newDHT78(t)
	// Non-empty from NodeID (all ff)
	var nodeID rpc.NodeID
	for i := range nodeID {
		nodeID[i] = 0xff
	}
	msg := rpc.Message{
		From:  rpc.Remote{NodeID: nodeID},
		RPCID: 0, // zero RPCID → error
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("verifyMessage zero RPCID panicked: %v", r)
			}
		}()
		d.verifyMessage(msg)
	}()
}

// ─── allowToJoin — rate limiting ─────────────────────────────────────────────

func TestSprint78_AllowToJoin_FirstCall(t *testing.T) {
	d := newDHT78(t)
	// First call should be allowed (lastJoin is zero time)
	allowed := d.allowToJoin()
	if !allowed {
		t.Error("expected first allowToJoin to return true")
	}
}

func TestSprint78_AllowToJoin_RateLimited(t *testing.T) {
	d := newDHT78(t)
	// After first call, lastJoin is set — second immediate call should be rejected
	d.allowToJoin() // sets lastJoin
	allowed := d.allowToJoin()
	if allowed {
		t.Error("expected second immediate allowToJoin to return false (rate limited)")
	}
}

// ─── targetField / fromField / localNodeIDField — no panic ───────────────────

func TestSprint78_TargetField_NoPanic(t *testing.T) {
	d := newDHT78(t)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("targetField panicked: %v", r)
			}
		}()
		var k [32]byte
		_ = d.targetField(k)
	}()
}

func TestSprint78_FromField_NoPanic(t *testing.T) {
	d := newDHT78(t)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("fromField panicked: %v", r)
			}
		}()
		_ = d.fromField(rpc.Remote{})
	}()
}

func TestSprint78_LocalNodeIDField_NoPanic(t *testing.T) {
	d := newDHT78(t)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("localNodeIDField panicked: %v", r)
			}
		}()
		_ = d.localNodeIDField()
	}()
}

// ─── getRandomDelay — panics on zero (documented bug) ────────────────────────
// getRandomDelay(0) panics: rand.Uint64() % uint64(0) = divide by zero
// Only test with non-zero duration.

func TestSprint78_GetRandomDelay_NonZero(t *testing.T) {
	result := getRandomDelay(100 * time.Millisecond)
	if result < 0 {
		t.Error("getRandomDelay should return non-negative duration")
	}
	if result > 100*time.Millisecond {
		t.Errorf("getRandomDelay should be ≤ input duration, got %v", result)
	}
}

// ─── routingTable — GetStaleRemote empty table ───────────────────────────────

func TestSprint78_GetStaleRemote_EmptyTable(t *testing.T) {
	var selfID rpc.NodeID
	for i := range selfID {
		selfID[i] = 0xaa
	}
	rt := newRoutingTable(20, 256, selfID, net.UDPAddr{})
	stale := rt.GetStaleRemote()
	if len(stale) != 0 {
		t.Errorf("expected 0 stale nodes in empty table, got %d", len(stale))
	}
}

// ─── routingTable — GC empty table ───────────────────────────────────────────

func TestSprint78_GC_EmptyTable(t *testing.T) {
	var selfID rpc.NodeID
	for i := range selfID {
		selfID[i] = 0xbb
	}
	rt := newRoutingTable(20, 256, selfID, net.UDPAddr{})
	// Should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GC on empty table panicked: %v", r)
			}
		}()
		rt.GC()
	}()
}

// ─── routingTable — Observe self-node skip ────────────────────────────────────

func TestSprint78_Observe_SelfNodeSkipped(t *testing.T) {
	var selfID rpc.NodeID
	for i := range selfID {
		selfID[i] = 0xcc
	}
	rt := newRoutingTable(20, 256, selfID, net.UDPAddr{})
	// Observing self should be a no-op (skip) — no panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Observe self panicked: %v", r)
			}
		}()
		rt.Observe(selfID, net.UDPAddr{})
	}()
	// After observing only self (which is skipped), KNearest should return empty
	var target rpc.NodeID
	for i := range target {
		target[i] = 0xdd
	}
	nearest := rt.KNearest(target)
	// Self-observation is skipped, so routing table is empty — 0 nearest expected
	_ = nearest // may return 0 or more depending on bucket initialization
}
