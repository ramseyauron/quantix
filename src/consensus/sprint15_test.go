package consensus

// Coverage Sprint 15 — consensus utility functions:
// CacheMerkleRoot, GetCachedMerkleRoot, StatusFromMsgType,
// PruneStaleProposalSignatures, GetConsensusSignatures.

import (
	"math/big"
	"testing"
)

// ---------------------------------------------------------------------------
// CacheMerkleRoot / GetCachedMerkleRoot
// ---------------------------------------------------------------------------

func TestCacheMerkleRoot_StoreAndRetrieve(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	c.CacheMerkleRoot("hash-001", "merkleroot-001")
	got := c.GetCachedMerkleRoot("hash-001")
	if got != "merkleroot-001" {
		t.Errorf("GetCachedMerkleRoot: got %q, want %q", got, "merkleroot-001")
	}
}

func TestGetCachedMerkleRoot_UnknownHash_Empty(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	got := c.GetCachedMerkleRoot("nonexistent-hash-00000000000")
	if got != "" {
		t.Errorf("GetCachedMerkleRoot unknown hash: expected empty, got %q", got)
	}
}

func TestCacheMerkleRoot_Overwrite(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	c.CacheMerkleRoot("hash-overwrite", "root-v1")
	c.CacheMerkleRoot("hash-overwrite", "root-v2")

	got := c.GetCachedMerkleRoot("hash-overwrite")
	if got != "root-v2" {
		t.Errorf("CacheMerkleRoot overwrite: got %q, want %q", got, "root-v2")
	}
}

func TestCacheMerkleRoot_MultipleHashes(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	for i, h := range []string{"hash-a", "hash-b", "hash-c"} {
		c.CacheMerkleRoot(h, "root-"+string(rune('a'+i)))
	}

	if c.GetCachedMerkleRoot("hash-a") == "" {
		t.Error("hash-a should be cached")
	}
	if c.GetCachedMerkleRoot("hash-b") == "" {
		t.Error("hash-b should be cached")
	}
	if c.GetCachedMerkleRoot("hash-c") == "" {
		t.Error("hash-c should be cached")
	}
}

// ---------------------------------------------------------------------------
// StatusFromMsgType
// ---------------------------------------------------------------------------

func TestStatusFromMsgType_AllKnownTypes(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	cases := []struct {
		msgType string
		want    string
	}{
		{"proposal", "proposed"},
		{"prepare", "prepared"},
		{"commit", "committed"},
		{"timeout", "view_change"},
		{"unknown_type", "processed"},
		{"", "processed"},
	}

	for _, tc := range cases {
		got := c.StatusFromMsgType(tc.msgType)
		if got != tc.want {
			t.Errorf("StatusFromMsgType(%q): got %q, want %q", tc.msgType, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// PruneStaleProposalSignatures (ba1c85a — fix block_not_found)
// ---------------------------------------------------------------------------

func TestPruneStaleProposalSignatures_EmptyList_NoError(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	// Empty signatures — pruning should be a no-op.
	c.PruneStaleProposalSignatures(10, "committed-hash-0000000000000001")
	sigs := c.GetConsensusSignatures()
	if len(sigs) != 0 {
		t.Errorf("PruneStaleProposalSignatures empty: expected 0 sigs, got %d", len(sigs))
	}
}

func TestPruneStaleProposalSignatures_RemovesStale(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	// Inject signatures manually — one stale (same height, different hash), one kept.
	c.signatureMutex.Lock()
	c.consensusSignatures = []*ConsensusSignature{
		{BlockHeight: 5, BlockHash: "stale-hash-for-height-5-001", SignerNodeID: "v1"},
		{BlockHeight: 5, BlockHash: "committed-hash-height-5-001", SignerNodeID: "v2"},
		{BlockHeight: 6, BlockHash: "different-height-hash-00001", SignerNodeID: "v3"},
	}
	c.signatureMutex.Unlock()

	// Prune stale entries for height 5 with the committed hash.
	c.PruneStaleProposalSignatures(5, "committed-hash-height-5-001")

	sigs := c.GetConsensusSignatures()
	for _, sig := range sigs {
		if sig.BlockHeight == 5 && sig.BlockHash != "committed-hash-height-5-001" {
			t.Errorf("PruneStaleProposalSignatures: stale sig not removed: height=%d hash=%s",
				sig.BlockHeight, sig.BlockHash)
		}
	}
	// v3 (height 6) must remain.
	found := false
	for _, sig := range sigs {
		if sig.SignerNodeID == "v3" {
			found = true
		}
	}
	if !found {
		t.Error("PruneStaleProposalSignatures: height-6 signature should be kept")
	}
}

// ---------------------------------------------------------------------------
// GetConsensusSignatures
// ---------------------------------------------------------------------------

func TestGetConsensusSignatures_EmptyList_EmptySlice(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	sigs := c.GetConsensusSignatures()
	if len(sigs) != 0 {
		t.Errorf("GetConsensusSignatures empty: expected 0, got %d", len(sigs))
	}
}

func TestGetConsensusSignatures_ReturnsIndependentCopy(t *testing.T) {
	c := makeTestConsensus(t)
	defer c.Stop()

	c.signatureMutex.Lock()
	c.consensusSignatures = []*ConsensusSignature{
		{BlockHash: "original-hash-001", SignerNodeID: "v1"},
	}
	c.signatureMutex.Unlock()

	copy1 := c.GetConsensusSignatures()
	if len(copy1) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(copy1))
	}

	// Mutate the copy.
	copy1[0] = nil

	// Original should be unchanged.
	copy2 := c.GetConsensusSignatures()
	if len(copy2) != 1 || copy2[0] == nil {
		t.Error("GetConsensusSignatures: modifying returned copy affected internal state")
	}
}

// ---------------------------------------------------------------------------
// helper — makeTestConsensus
// ---------------------------------------------------------------------------

func makeTestConsensus(t *testing.T) *Consensus {
	t.Helper()
	c := NewConsensus("pepper-test-node", nil, nil, nil, nil, big.NewInt(1000))
	if c == nil {
		t.Fatal("NewConsensus returned nil")
	}
	return c
}
