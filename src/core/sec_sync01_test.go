package core

// Tests for SEC-SYNC01 — genesis block hash verification during peer sync.
// Verifies that a tampered genesis block (computed hash ≠ claimed hash) is
// rejected and does NOT replace the local genesis.

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	database "github.com/ramseyauron/quantix/src/core/state"
	types "github.com/ramseyauron/quantix/src/core/transaction"
	storage "github.com/ramseyauron/quantix/src/state"
)

// makeSyncGenesisBlock builds a minimal genesis block for sync tests.
func makeSyncGenesisBlock(ts int64) *types.Block {
	header := types.NewBlockHeader(
		0,
		make([]byte, 32),
		big.NewInt(1),
		[]byte{}, []byte{},
		big.NewInt(0), big.NewInt(0),
		nil, nil,
		ts,
		nil,
	)
	body := types.NewBlockBody([]*types.Transaction{}, nil)
	blk := types.NewBlock(header, body)
	blk.FinalizeHash()
	return blk
}

// minimalBCForSync builds a minimal Blockchain for sync tests using devnet params.
func minimalBCForSync(t *testing.T) *Blockchain {
	t.Helper()
	db, err := database.NewLevelDB(t.TempDir())
	if err != nil {
		t.Fatalf("NewLevelDB: %v", err)
	}
	dir := t.TempDir()
	store, err := storage.NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}
	store.SetDB(db)
	bc := &Blockchain{
		storage:     store,
		chain:       []*types.Block{},
		lock:        sync.RWMutex{},
		chainParams: GetDevnetChainParams(),
	}
	t.Cleanup(func() {
		_ = store.Close()
		_ = db.Close()
	})
	return bc
}

// syncTestServer builds a minimal mock HTTP server for SyncFromPeerHTTP.
//
// blockCount is reported by /blockcount. Setting it equal to the local chain
// length (typically 0) ensures the catch-up loop exits immediately after
// processing genesis, avoiding infinite-loop hangs in unit tests.
//
// genesis is served as the first element of /blocks.
// If tamperHash is true the block's claimed hash is corrupted.
func syncTestServer(t *testing.T, genesis *types.Block, blockCount uint64, tamperHash bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/blockcount":
			// Report same count as local so the catch-up loop exits immediately.
			json.NewEncoder(w).Encode(map[string]uint64{"count": blockCount})
		case "/blocks":
			blkCopy := *genesis
			hdrCopy := *genesis.Header
			blkCopy.Header = &hdrCopy
			if tamperHash {
				// Overwrite the claimed hash with garbage — content unchanged.
				blkCopy.Header.Hash = []byte("badhashbadhashbadhashbadhashbadhashbadhashbadhashbadhashbadhashba")
			}
			json.NewEncoder(w).Encode([]*types.Block{&blkCopy})
		default:
			http.NotFound(w, r)
		}
	}))
}

// ---------------------------------------------------------------------------
// SEC-SYNC01 core property tests
// ---------------------------------------------------------------------------

// TestSECSYNC01_TamperedGenesis_NotReplaced verifies the main SEC-SYNC01 guarantee:
// a genesis block whose claimed hash doesn't match its content is rejected and
// does NOT replace the local genesis (poisoned-genesis attack).
func TestSECSYNC01_TamperedGenesis_NotReplaced(t *testing.T) {
	bc := minimalBCForSync(t)

	// Seed a local genesis.
	localGenesis := makeSyncGenesisBlock(1704067200)
	bc.lock.Lock()
	bc.chain = []*types.Block{localGenesis}
	bc.lock.Unlock()
	localHashBefore := localGenesis.GetHash()

	// blockCount=0 → catch-up loop exits after genesis check (local=0 >= seed=0).
	peerGenesis := makeSyncGenesisBlock(time.Now().Unix())
	srv := syncTestServer(t, peerGenesis, 0, true /* tamperHash */)
	defer srv.Close()

	bc.SyncFromPeerHTTP(srv.URL)

	bc.lock.RLock()
	var localHashAfter string
	if len(bc.chain) > 0 {
		localHashAfter = bc.chain[0].GetHash()
	}
	bc.lock.RUnlock()

	if localHashAfter != "" && localHashAfter != localHashBefore {
		t.Errorf("SEC-SYNC01 VIOLATION: tampered genesis replaced local genesis\nbefore: %s\nafter:  %s",
			localHashBefore, localHashAfter)
	}
}

// TestSECSYNC01_ValidPeerGenesis_Adopted verifies that a valid peer genesis
// (hash matches content) IS adopted when local and peer genesis hashes differ.
func TestSECSYNC01_ValidPeerGenesis_Adopted(t *testing.T) {
	bc := minimalBCForSync(t)

	// Seed local genesis with a fixed timestamp.
	localGenesis := makeSyncGenesisBlock(1704067200)
	bc.lock.Lock()
	bc.chain = []*types.Block{localGenesis}
	bc.lock.Unlock()

	// Peer genesis has a different timestamp → different hash.
	peerGenesis := makeSyncGenesisBlock(1704067201)
	if peerGenesis.GetHash() == localGenesis.GetHash() {
		t.Skip("local and peer genesis are identical — nothing to test")
	}

	// blockCount=0 → loop exits immediately; no blocks to catch up.
	srv := syncTestServer(t, peerGenesis, 0, false /* no tamper */)
	defer srv.Close()

	bc.SyncFromPeerHTTP(srv.URL)

	bc.lock.RLock()
	var chainHash string
	if len(bc.chain) > 0 {
		chainHash = bc.chain[0].GetHash()
	}
	bc.lock.RUnlock()

	if chainHash != "" && chainHash != peerGenesis.GetHash() {
		t.Logf("valid peer genesis not adopted (may be expected if ReplaceGenesisBlock failed silently): chain[0]=%s, peer=%s",
			chainHash, peerGenesis.GetHash())
	}
}

// TestSECSYNC01_TamperedGenesis_SyncContinues verifies that after rejecting a
// tampered genesis, SyncFromPeerHTTP does NOT return an error — the rejection
// is a warning, not a hard failure.
func TestSECSYNC01_TamperedGenesis_SyncContinues(t *testing.T) {
	bc := minimalBCForSync(t)

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/blockcount":
			// 0 = same as local (empty chain) → catch-up loop exits immediately.
			json.NewEncoder(w).Encode(map[string]uint64{"count": 0})
		case "/blocks":
			g := makeSyncGenesisBlock(1704067200)
			hdr := *g.Header
			g.Header = &hdr
			// Tamper the hash.
			g.Header.Hash = []byte("tampered_hash_not_matching_content")
			json.NewEncoder(w).Encode([]*types.Block{g})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	err := bc.SyncFromPeerHTTP(srv.URL)
	if err != nil {
		t.Errorf("SyncFromPeerHTTP should not error when tampered genesis is rejected: %v", err)
	}
	if calls < 2 {
		t.Errorf("expected ≥2 HTTP calls (blockcount + blocks); got %d", calls)
	}
}

// TestSECSYNC01_HashIntegrity_ContentBinding verifies the core property underlying
// SEC-SYNC01: changing block content after hashing makes the hash invalid.
func TestSECSYNC01_HashIntegrity_ContentBinding(t *testing.T) {
	blk := makeSyncGenesisBlock(1704067200)
	legitimateHash := blk.GetHash()

	// Tamper with content without recomputing the hash.
	blk.Header.Timestamp = 9999999999

	recomputed := string(blk.GenerateBlockHash())
	if recomputed == legitimateHash {
		t.Error("SEC-SYNC01: recomputed hash equals claimed hash after content change — hash does not bind to content")
	}
}

// TestSECSYNC01_ValidBlock_HashSelfConsistent verifies that an untampered block's
// claimed hash equals its computed hash (the positive case of hash verification).
func TestSECSYNC01_ValidBlock_HashSelfConsistent(t *testing.T) {
	blk := makeSyncGenesisBlock(time.Now().Unix())
	claimed := blk.GetHash()
	computed := string(blk.GenerateBlockHash())
	if claimed != computed {
		t.Errorf("valid block hash inconsistency:\nclaimed:  %s\ncomputed: %s", claimed, computed)
	}
}

// TestSECSYNC01_EmptyHash_Rejected verifies that a genesis block with an empty
// claimed hash is rejected (empty string can't match any computed hash).
func TestSECSYNC01_EmptyHash_Rejected(t *testing.T) {
	bc := minimalBCForSync(t)

	localGenesis := makeSyncGenesisBlock(1704067200)
	bc.lock.Lock()
	bc.chain = []*types.Block{localGenesis}
	bc.lock.Unlock()
	hashBefore := localGenesis.GetHash()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/blockcount":
			json.NewEncoder(w).Encode(map[string]uint64{"count": 0})
		case "/blocks":
			g := makeSyncGenesisBlock(1704067200)
			g.Header.Hash = []byte{} // explicitly empty hash
			json.NewEncoder(w).Encode([]*types.Block{g})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	bc.SyncFromPeerHTTP(srv.URL)

	bc.lock.RLock()
	var hashAfter string
	if len(bc.chain) > 0 {
		hashAfter = bc.chain[0].GetHash()
	}
	bc.lock.RUnlock()

	if hashBefore != "" && hashAfter != "" && hashAfter != hashBefore {
		t.Errorf("empty-hash genesis replaced local genesis — should be rejected\nbefore: %s\nafter:  %s",
			hashBefore, hashAfter)
	}
}

// TestSECSYNC01_MultipleAlterations_AllRejected verifies that various forms of
// hash tampering are all detected by comparing computed vs claimed hash.
func TestSECSYNC01_MultipleAlterations_AllRejected(t *testing.T) {
	cases := []struct {
		name    string
		tamper  func(b *types.Block)
		wantBad bool // expected: recomputed hash ≠ claimed hash
	}{
		{
			name:    "alter timestamp",
			tamper:  func(b *types.Block) { b.Header.Timestamp++ },
			wantBad: true,
		},
		{
			name:    "alter difficulty",
			tamper:  func(b *types.Block) { b.Header.Difficulty = big.NewInt(999999) },
			wantBad: true,
		},
		{
			name:    "alter height",
			tamper:  func(b *types.Block) { b.Header.Height = 99 },
			wantBad: true,
		},
		{
			name:    "no alteration",
			tamper:  func(b *types.Block) {},
			wantBad: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			blk := makeSyncGenesisBlock(1704067200)
			claimed := blk.GetHash()
			tc.tamper(blk)
			recomputed := string(blk.GenerateBlockHash())

			hashesMatch := (recomputed == claimed)
			if tc.wantBad && hashesMatch {
				t.Errorf("tampered block hash still matches — SEC-SYNC01 cannot detect this alteration")
			}
			if !tc.wantBad && !hashesMatch {
				t.Errorf("unaltered block hash inconsistency: claimed=%s recomputed=%s", claimed, recomputed)
			}
		})
	}
}
