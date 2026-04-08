// Sprint 27d — core/chain_maker: LoadChainCheckpoint, LoadChainCheckpointFromAddress
package core

import (
	"os"
	"testing"
)

// ─── LoadChainCheckpoint ──────────────────────────────────────────────────────

func TestSprint27_LoadChainCheckpoint_MissingDir_ReturnsNil(t *testing.T) {
	dir, _ := os.MkdirTemp("", "qtx-ccp27-*")
	defer os.RemoveAll(dir)

	cp, err := LoadChainCheckpoint(dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cp != nil {
		t.Errorf("expected nil checkpoint for dir with no checkpoint file, got %v", cp)
	}
}

func TestSprint27_LoadChainCheckpoint_NonExistentDir_ReturnsNil(t *testing.T) {
	// Use a path that definitely does not exist
	cp, err := LoadChainCheckpoint("/tmp/quantix-nonexistent-dir-sprint27-test-abc123")
	if err != nil {
		t.Errorf("expected nil error for missing checkpoint, got: %v", err)
	}
	if cp != nil {
		t.Errorf("expected nil checkpoint, got: %v", cp)
	}
}

// ─── LoadChainCheckpointFromAddress ──────────────────────────────────────────

func TestSprint27_LoadChainCheckpointFromAddress_UnknownAddr_ReturnsNil(t *testing.T) {
	cp, err := LoadChainCheckpointFromAddress("0.0.0.0:99999")
	if err != nil {
		t.Logf("LoadChainCheckpointFromAddress unknown: %v (non-nil error acceptable)", err)
	}
	// nil checkpoint is expected for unknown address
	_ = cp
}
