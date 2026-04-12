// test(PEPPER): Sprint 77 — src/bind 2.3%→higher
// Tests: Shutdown empty resources, Shutdown nil fields, CallNodeRPC empty resources,
// CallNodeRPC node-not-found, ParseRoles edge cases already covered, NodeSetupConfig fields
package bind

import (
	"testing"
)

// ─── Shutdown — empty resources (no-op) ──────────────────────────────────────

func TestSprint77_Shutdown_EmptyResources(t *testing.T) {
	err := Shutdown([]NodeResources{})
	if err != nil {
		t.Errorf("Shutdown empty: expected nil error, got %v", err)
	}
}

// ─── Shutdown — nil resources slice ──────────────────────────────────────────

func TestSprint77_Shutdown_NilResources(t *testing.T) {
	err := Shutdown(nil)
	if err != nil {
		t.Errorf("Shutdown nil: expected nil error, got %v", err)
	}
}

// ─── Shutdown — resource with all-nil servers ─────────────────────────────────

func TestSprint77_Shutdown_NilServers(t *testing.T) {
	resources := []NodeResources{
		{
			// All server fields are nil — Shutdown should skip them gracefully
		},
	}
	err := Shutdown(resources)
	if err != nil {
		t.Errorf("Shutdown nil servers: expected nil error, got %v", err)
	}
}

// ─── CallNodeRPC — empty resources (node not found) ──────────────────────────

func TestSprint77_CallNodeRPC_EmptyResources(t *testing.T) {
	_, err := CallNodeRPC([]NodeResources{}, "missing-node", "ping", nil, 10)
	if err == nil {
		t.Error("expected error for node not found in empty resources")
	}
}

func TestSprint77_CallNodeRPC_NilResources(t *testing.T) {
	_, err := CallNodeRPC(nil, "missing-node", "ping", nil, 10)
	if err == nil {
		t.Error("expected error for node not found in nil resources")
	}
}
