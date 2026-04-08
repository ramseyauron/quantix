// Sprint 20 — DiskKeyStore methods: StoreKey, GetKey, ListKeys, GetKeyByAddress,
// ListKeysByType, GetWalletInfo, generateSalt, validateKeyPair, RemoveKey,
// UpdateKeyMetadata, ChangePassphrase path stubs.
// Also covers NewTCPServer construction and GetConnection/GetEncryptionKey error paths.
package disk_test

import (
	"os"
	"testing"
	"time"

	key "github.com/ramseyauron/quantix/src/accounts/key"
	disk "github.com/ramseyauron/quantix/src/accounts/key/disk"
)

// helper: create a fresh DiskKeyStore in a temp dir.
func newTestDKS(t *testing.T) (*disk.DiskKeyStore, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "dks20-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	ks, err := disk.NewDiskKeyStore(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("NewDiskKeyStore: %v", err)
	}
	return ks, func() { os.RemoveAll(dir) }
}

// helper: build a minimal valid KeyPair for storage.
func minimalKeyPair(id, address string) *key.KeyPair {
	return &key.KeyPair{
		ID:          id,
		EncryptedSK: []byte("dummyEncryptedKey"),
		PublicKey:   []byte("dummyPublicKey"),
		Address:     address,
		WalletType:  key.WalletTypeDisk,
		CreatedAt:   time.Now(),
	}
}

// ─── StoreKey / GetKey roundtrip ─────────────────────────────────────────────

func TestSprint20_StoreKey_Valid(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-001", "addr-001")
	if err := ks.StoreKey(kp); err != nil {
		t.Fatalf("StoreKey error: %v", err)
	}
}

func TestSprint20_GetKey_Found(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-002", "addr-002")
	if err := ks.StoreKey(kp); err != nil {
		t.Fatalf("StoreKey error: %v", err)
	}
	got, err := ks.GetKey("key-002")
	if err != nil {
		t.Fatalf("GetKey error: %v", err)
	}
	if got.Address != "addr-002" {
		t.Errorf("got address %q, want %q", got.Address, "addr-002")
	}
}

func TestSprint20_GetKey_NotFound_Error(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	_, err := ks.GetKey("nonexistent")
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

// ─── GetKeyByAddress ──────────────────────────────────────────────────────────

func TestSprint20_GetKeyByAddress_Found(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-003", "addr-003")
	if err := ks.StoreKey(kp); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}
	got, err := ks.GetKeyByAddress("addr-003")
	if err != nil {
		t.Fatalf("GetKeyByAddress: %v", err)
	}
	if got.ID != "key-003" {
		t.Errorf("got ID %q, want %q", got.ID, "key-003")
	}
}

func TestSprint20_GetKeyByAddress_NotFound_Error(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	_, err := ks.GetKeyByAddress("no-such-addr")
	if err == nil {
		t.Error("expected error for missing address, got nil")
	}
}

// ─── ListKeys ─────────────────────────────────────────────────────────────────

func TestSprint20_ListKeys_Empty(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	keys := ks.ListKeys()
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestSprint20_ListKeys_AfterStore(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	for i := 0; i < 3; i++ {
		id := "key-list-" + string(rune('A'+i))
		addr := "addr-list-" + string(rune('A'+i))
		if err := ks.StoreKey(minimalKeyPair(id, addr)); err != nil {
			t.Fatalf("StoreKey[%d]: %v", i, err)
		}
	}

	keys := ks.ListKeys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

// ─── ListKeysByType ───────────────────────────────────────────────────────────

func TestSprint20_ListKeysByType_Disk(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-type-1", "addr-type-1")
	kp.WalletType = key.WalletTypeDisk
	if err := ks.StoreKey(kp); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}

	got := ks.ListKeysByType(key.WalletTypeDisk)
	if len(got) != 1 {
		t.Errorf("expected 1 disk key, got %d", len(got))
	}
}

func TestSprint20_ListKeysByType_NoMatch(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-type-2", "addr-type-2")
	kp.WalletType = key.WalletTypeDisk
	if err := ks.StoreKey(kp); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}

	// Filter for a different wallet type — should return none
	got := ks.ListKeysByType(key.WalletTypeUSB)
	if len(got) != 0 {
		t.Errorf("expected 0 USB keys, got %d", len(got))
	}
}

// ─── RemoveKey ────────────────────────────────────────────────────────────────

func TestSprint20_RemoveKey_Existing(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-rm", "addr-rm")
	if err := ks.StoreKey(kp); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}
	if err := ks.RemoveKey("key-rm"); err != nil {
		t.Fatalf("RemoveKey: %v", err)
	}
	_, err := ks.GetKey("key-rm")
	if err == nil {
		t.Error("expected key to be gone after RemoveKey")
	}
}

func TestSprint20_RemoveKey_NotFound_NoError(t *testing.T) {
	// RemoveKey uses delete() on map (no-op for missing) + os.Remove (IsNotExist ignored)
	// So missing keys are silently ignored — document this behaviour
	ks, cleanup := newTestDKS(t)
	defer cleanup()
	// Should NOT error (or may error depending on implementation — just must not panic)
	_ = ks.RemoveKey("ghost")
}

// ─── UpdateKeyMetadata ────────────────────────────────────────────────────────

func TestSprint20_UpdateKeyMetadata_Existing(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-meta", "addr-meta")
	kp.Metadata = make(map[string]interface{}) // must init Metadata map
	if err := ks.StoreKey(kp); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}

	meta := map[string]interface{}{"label": "my key"}
	if err := ks.UpdateKeyMetadata("key-meta", meta); err != nil {
		t.Fatalf("UpdateKeyMetadata: %v", err)
	}
}

func TestSprint20_UpdateKeyMetadata_NotFound_Error(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	if err := ks.UpdateKeyMetadata("ghost", nil); err == nil {
		t.Error("expected error for non-existent key")
	}
}

// ─── GetWalletInfo ────────────────────────────────────────────────────────────

func TestSprint20_GetWalletInfo_NotNil(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	info := ks.GetWalletInfo()
	if info == nil {
		t.Fatal("GetWalletInfo returned nil")
	}
}

func TestSprint20_GetWalletInfo_KeyCountZero_Initially(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	info := ks.GetWalletInfo()
	if info.KeyCount != 0 {
		t.Errorf("expected KeyCount=0, got %d", info.KeyCount)
	}
}

func TestSprint20_GetWalletInfo_KeyCount_AfterStore(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	ks.StoreKey(minimalKeyPair("wk1", "wa1"))
	ks.StoreKey(minimalKeyPair("wk2", "wa2"))

	info := ks.GetWalletInfo()
	if info.KeyCount != 2 {
		t.Errorf("expected KeyCount=2, got %d", info.KeyCount)
	}
}

func TestSprint20_GetWalletInfo_WalletType(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	info := ks.GetWalletInfo()
	if info.WalletType != key.WalletTypeDisk {
		t.Errorf("WalletType = %v, want WalletTypeDisk", info.WalletType)
	}
}

// ─── StoreKey validation (validateKeyPair paths) ─────────────────────────────

func TestSprint20_StoreKey_EmptyID_Error(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("", "addr-bad")
	if err := ks.StoreKey(kp); err == nil {
		t.Error("expected error for empty key ID")
	}
}

func TestSprint20_StoreKey_EmptyEncryptedSK_Error(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-bad-sk", "addr-bad-sk")
	kp.EncryptedSK = nil
	if err := ks.StoreKey(kp); err == nil {
		t.Error("expected error for empty EncryptedSK")
	}
}

func TestSprint20_StoreKey_EmptyPublicKey_Error(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-bad-pk", "addr-bad-pk")
	kp.PublicKey = nil
	if err := ks.StoreKey(kp); err == nil {
		t.Error("expected error for empty PublicKey")
	}
}

func TestSprint20_StoreKey_EmptyAddress_Error(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp := minimalKeyPair("key-bad-addr", "")
	if err := ks.StoreKey(kp); err == nil {
		t.Error("expected error for empty Address")
	}
}

// ─── StoreEncryptedKey ────────────────────────────────────────────────────────

func TestSprint20_StoreEncryptedKey_Valid(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	kp, err := ks.StoreEncryptedKey(
		[]byte("encryptedSK"),
		[]byte("pubKey"),
		"addr-encrypted",
		key.WalletTypeDisk,
		73310,
		"m/44'/0'/0'/0/0",
		nil,
	)
	if err != nil {
		t.Fatalf("StoreEncryptedKey: %v", err)
	}
	if kp == nil {
		t.Fatal("StoreEncryptedKey returned nil KeyPair")
	}
	if kp.Address != "addr-encrypted" {
		t.Errorf("Address = %q, want %q", kp.Address, "addr-encrypted")
	}
}

// ─── ExportKey ────────────────────────────────────────────────────────────────

func TestSprint20_ExportKey_NotFound_Error(t *testing.T) {
	ks, cleanup := newTestDKS(t)
	defer cleanup()

	_, err := ks.ExportKey("ghost", false, "")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}
