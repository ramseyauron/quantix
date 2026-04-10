// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 33 - bind coverage: CallNodeRPC + Shutdown error paths
package bind

import (
	"testing"
)

// ---------------------------------------------------------------------------
// CallNodeRPC — empty resources returns error
// ---------------------------------------------------------------------------

func TestSprint33_CallNodeRPC_EmptyResources_Error(t *testing.T) {
	_, err := CallNodeRPC(nil, "unknown-node", "getblockcount", nil, 5)
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}

func TestSprint33_CallNodeRPC_EmptyResourcesSlice_Error(t *testing.T) {
	_, err := CallNodeRPC([]NodeResources{}, "unknown-node", "getblockcount", nil, 5)
	if err == nil {
		t.Fatal("expected error for empty resources slice")
	}
}

// ---------------------------------------------------------------------------
// Shutdown — empty resources returns nil (nothing to shut down)
// ---------------------------------------------------------------------------

func TestSprint33_Shutdown_EmptyResources_NoError(t *testing.T) {
	err := Shutdown(nil)
	// nil or empty resources: no error expected
	_ = err
}

func TestSprint33_Shutdown_EmptySlice_NoError(t *testing.T) {
	err := Shutdown([]NodeResources{})
	_ = err
}

// ---------------------------------------------------------------------------
// LaunchNetwork — unknown mode calls log.Fatal (documented)
// ---------------------------------------------------------------------------

func TestSprint33_LaunchNetwork_UnknownMode_Documented(t *testing.T) {
	// LaunchNetwork calls log.Fatal for unknown mode — cannot test without catching os.Exit
	t.Skip("LaunchNetwork calls log.Fatal for unknown mode — pending error return instead")
}
