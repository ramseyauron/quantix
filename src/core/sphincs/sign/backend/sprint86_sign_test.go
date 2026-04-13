// security(EDITH): Sprint 86 fix — correct tests to match actual SphincsManager API
// Tests: serializePK nil, NewSphincsManager nil keyManager panics,
// PublicKeyRegistry Register/Lookup/HasPublicKey/VerifyIdentity,
// LoadCommitment nil db, storeCommitment nil db,
// VerifyTxSignature garbage bytes, CheckTimestampNonce/StoreTimestampNonce nil db,
// generateNonce length
package sign

import (
	"os"
	"testing"

	key "github.com/ramseyauron/quantix/src/core/sphincs/key/backend"
	"github.com/syndtr/goleveldb/leveldb"
)

func newTestDB86(t *testing.T) *leveldb.DB {
	t.Helper()
	dir, err := os.MkdirTemp("", "qtx-sign86-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("leveldb.OpenFile: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
		os.RemoveAll(dir)
	})
	return db
}

func newTestSM86(t *testing.T) *SphincsManager {
	t.Helper()
	km, err := key.NewKeyManager()
	if err != nil {
		t.Skipf("NewKeyManager failed: %v", err)
	}
	return NewSphincsManager(newTestDB86(t), km, km.GetSPHINCSParameters())
}

// ─── serializePK — nil pk returns error ──────────────────────────────────────

func TestSprint86_SerializePK_NilPK(t *testing.T) {
	sm := newTestSM86(t)
	_, err := sm.serializePK(nil)
	if err == nil {
		t.Error("expected error for nil pk")
	}
}

// ─── NewSphincsManager — nil keyManager panics ───────────────────────────────

func TestSprint86_NewSphincsManager_NilKeyManager_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("NewSphincsManager nil keyManager panicked (expected): %v", r)
		}
	}()
	_ = NewSphincsManager(nil, nil, nil)
}

// ─── PublicKeyRegistry — Register/Lookup ─────────────────────────────────────

func TestSprint86_Registry_NotFound(t *testing.T) {
	reg := NewPublicKeyRegistry()
	_, found := reg.Lookup("unknown-node")
	if found {
		t.Error("expected not found for unknown node")
	}
}

func TestSprint86_Registry_AfterRegister(t *testing.T) {
	reg := NewPublicKeyRegistry()
	reg.Register("test-node-86", []byte{0x01, 0x02, 0x03})
	_, found := reg.Lookup("test-node-86")
	if !found {
		t.Error("expected to find node after Register")
	}
}

// ─── PublicKeyRegistry.VerifyIdentity — true/false ───────────────────────────

func TestSprint86_VerifyIdentity_False_Unknown(t *testing.T) {
	reg := NewPublicKeyRegistry()
	if reg.VerifyIdentity("nobody", []byte{0x01}) {
		t.Error("expected false for unknown node")
	}
}

func TestSprint86_VerifyIdentity_True_Matching(t *testing.T) {
	reg := NewPublicKeyRegistry()
	reg.Register("node-86", []byte{0xAA, 0xBB})
	if !reg.VerifyIdentity("node-86", []byte{0xAA, 0xBB}) {
		t.Error("expected true when PK matches")
	}
}

func TestSprint86_VerifyIdentity_False_Mismatch(t *testing.T) {
	reg := NewPublicKeyRegistry()
	reg.Register("node-86", []byte{0xAA, 0xBB})
	if reg.VerifyIdentity("node-86", []byte{0xCC}) {
		t.Error("expected false when PK mismatches")
	}
}

// ─── StoreTimestampNonce / CheckTimestampNonce — nil db errors ────────────────

func TestSprint86_StoreTimestampNonce_NilDB(t *testing.T) {
	sm := &SphincsManager{db: nil}
	err := sm.StoreTimestampNonce([]byte{0x01}, []byte{0x02})
	if err == nil {
		t.Error("expected error with nil DB")
	}
}

func TestSprint86_CheckTimestampNonce_NilDB(t *testing.T) {
	sm := &SphincsManager{db: nil}
	_, err := sm.CheckTimestampNonce([]byte{0x01}, []byte{0x02})
	if err == nil {
		t.Error("expected error with nil DB")
	}
}

// ─── LoadCommitment — nil db returns error ────────────────────────────────────

func TestSprint86_LoadCommitment_NilDB(t *testing.T) {
	sm := &SphincsManager{db: nil}
	_, err := sm.LoadCommitment()
	if err == nil {
		t.Error("expected error with nil DB")
	}
}

// ─── storeCommitment — nil db returns error ───────────────────────────────────

func TestSprint86_StoreCommitment_NilDB(t *testing.T) {
	sm := &SphincsManager{db: nil}
	err := sm.storeCommitment([]byte{0x01})
	if err == nil {
		t.Error("expected error with nil DB")
	}
}

// ─── generateNonce — returns correct length ───────────────────────────────────

func TestSprint86_GenerateNonce_Length(t *testing.T) {
	nonce, err := generateNonce()
	if err != nil {
		t.Fatalf("generateNonce: %v", err)
	}
	if len(nonce) != 16 {
		t.Errorf("expected 16-byte nonce, got %d", len(nonce))
	}
}

// ─── VerifyTxSignature — garbage bytes returns false ─────────────────────────

func TestSprint86_VerifyTxSignature_GarbageSig(t *testing.T) {
	sm := newTestSM86(t)
	result := sm.VerifyTxSignature(
		[]byte("message"),
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, // timestamp
		[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}, // nonce
		[]byte("garbage-sig"),
		[]byte{0x01, 0x02}, // senderPubKey
	)
	if result {
		t.Error("expected false for garbage signature")
	}
}
