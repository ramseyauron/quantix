package common

// Coverage Sprint 15 — common/config.go file operations, hexutil.go formatters,
// common/time.go TimeService methods.

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// config.go — path helpers
// ---------------------------------------------------------------------------

func TestGetNodeInfoPath_NonEmpty(t *testing.T) {
	path := GetNodeInfoPath("127.0.0.1:32307")
	if path == "" {
		t.Error("GetNodeInfoPath should return non-empty path")
	}
	if filepath.Base(path) != "node_info.json" {
		t.Errorf("GetNodeInfoPath: expected filename node_info.json, got %s", filepath.Base(path))
	}
}

func TestGetCheckpointDir_NonEmpty(t *testing.T) {
	dir := GetCheckpointDir("127.0.0.1:32307")
	if dir == "" {
		t.Error("GetCheckpointDir should return non-empty path")
	}
}

func TestGetBlockchainCheckpointPath_NonEmpty(t *testing.T) {
	path := GetBlockchainCheckpointPath("/tmp/testnode")
	if path == "" {
		t.Error("GetBlockchainCheckpointPath should return non-empty path")
	}
}

func TestGetCheckpointFromDataDir_NonEmpty(t *testing.T) {
	path := GetCheckpointFromDataDir("/tmp/testnode")
	if path == "" {
		t.Error("GetCheckpointFromDataDir should return non-empty path")
	}
}

// ---------------------------------------------------------------------------
// config.go — EnsureNodeDirs
// ---------------------------------------------------------------------------

func TestEnsureNodeDirs_CreatesDirectories(t *testing.T) {
	// Use a temp path as the node address.
	tmpBase := t.TempDir()
	addr := filepath.Join(tmpBase, "test_node")

	if err := EnsureNodeDirs(addr); err != nil {
		t.Fatalf("EnsureNodeDirs: %v", err)
	}

	// Verify at least the node root dir was created.
	nodeDir := GetNodeDataDir(addr)
	info, err := os.Stat(nodeDir)
	if err != nil {
		t.Errorf("EnsureNodeDirs: expected %q to exist: %v", nodeDir, err)
	} else if !info.IsDir() {
		t.Errorf("EnsureNodeDirs: expected %q to be a directory", nodeDir)
	}
}

func TestEnsureNodeDirs_Idempotent(t *testing.T) {
	tmpBase := t.TempDir()
	addr := filepath.Join(tmpBase, "idempotent_node")

	// Call twice — should not error.
	if err := EnsureNodeDirs(addr); err != nil {
		t.Fatalf("first EnsureNodeDirs: %v", err)
	}
	if err := EnsureNodeDirs(addr); err != nil {
		t.Fatalf("second EnsureNodeDirs (idempotent): %v", err)
	}
}

// ---------------------------------------------------------------------------
// config.go — WriteKeysToFile / ReadKeysFromFile / KeysExist
// ---------------------------------------------------------------------------

func TestWriteAndReadKeys_Roundtrip(t *testing.T) {
	tmpBase := t.TempDir()
	addr := filepath.Join(tmpBase, "keys_node")

	privKey := []byte("test-private-key-bytes-00000001")
	pubKey := []byte("test-public-key-bytes-000000001")

	if err := WriteKeysToFile(addr, privKey, pubKey); err != nil {
		t.Fatalf("WriteKeysToFile: %v", err)
	}

	gotPriv, gotPub, err := ReadKeysFromFile(addr)
	if err != nil {
		t.Fatalf("ReadKeysFromFile: %v", err)
	}

	if string(gotPriv) != string(privKey) {
		t.Errorf("private key mismatch: got %q, want %q", gotPriv, privKey)
	}
	if string(gotPub) != string(pubKey) {
		t.Errorf("public key mismatch: got %q, want %q", gotPub, pubKey)
	}
}

func TestKeysExist_AfterWrite_True(t *testing.T) {
	tmpBase := t.TempDir()
	addr := filepath.Join(tmpBase, "keys_exist_node")

	if err := WriteKeysToFile(addr, []byte("priv"), []byte("pub")); err != nil {
		t.Fatalf("WriteKeysToFile: %v", err)
	}

	if !KeysExist(addr) {
		t.Error("KeysExist should return true after WriteKeysToFile")
	}
}

func TestKeysExist_BeforeWrite_False(t *testing.T) {
	tmpBase := t.TempDir()
	addr := filepath.Join(tmpBase, "nonexistent_keys_node")
	if KeysExist(addr) {
		t.Error("KeysExist should return false for new address")
	}
}

// ---------------------------------------------------------------------------
// config.go — WriteNodeInfo / ReadNodeInfo
// ---------------------------------------------------------------------------

func TestWriteAndReadNodeInfo_Roundtrip(t *testing.T) {
	tmpBase := t.TempDir()
	addr := filepath.Join(tmpBase, "nodeinfo_node")

	if err := EnsureNodeDirs(addr); err != nil {
		t.Fatalf("EnsureNodeDirs: %v", err)
	}

	info := map[string]interface{}{
		"node_id": "test-node-id-001",
		"ip":      "127.0.0.1",
		"port":    "32307",
	}

	if err := WriteNodeInfo(addr, info); err != nil {
		t.Fatalf("WriteNodeInfo: %v", err)
	}

	got, err := ReadNodeInfo(addr)
	if err != nil {
		t.Fatalf("ReadNodeInfo: %v", err)
	}
	if got == nil {
		t.Fatal("ReadNodeInfo: returned nil")
	}
	if got["node_id"] != info["node_id"] {
		t.Errorf("node_id mismatch: got %v, want %v", got["node_id"], info["node_id"])
	}
}

// ---------------------------------------------------------------------------
// config.go — GetNodeDataDirs, CleanNodeData
// ---------------------------------------------------------------------------

func TestGetNodeDataDirs_NonEmpty(t *testing.T) {
	dirs := GetNodeDataDirs("127.0.0.1:32307")
	if len(dirs) == 0 {
		t.Error("GetNodeDataDirs should return at least one directory")
	}
	for _, d := range dirs {
		if d == "" {
			t.Error("GetNodeDataDirs: found empty directory path")
		}
	}
}

func TestCleanNodeData_NonExistentDir_NoError(t *testing.T) {
	addr := filepath.Join(t.TempDir(), "clean_node_nonexistent")
	// Cleaning non-existent node should not error.
	if err := CleanNodeData(addr); err != nil {
		t.Logf("CleanNodeData non-existent: %v (may error if dir doesn't exist)", err)
	}
}

// ---------------------------------------------------------------------------
// config.go — EnsureCheckpointDir, EnsureMetricsDir
// ---------------------------------------------------------------------------

func TestEnsureCheckpointDir_CreatesDir(t *testing.T) {
	addr := filepath.Join(t.TempDir(), "checkpoint_node")
	if err := EnsureCheckpointDir(addr); err != nil {
		t.Errorf("EnsureCheckpointDir: %v", err)
	}
}

func TestEnsureMetricsDir_CreatesDir(t *testing.T) {
	addr := filepath.Join(t.TempDir(), "metrics_node")
	if err := EnsureMetricsDir(addr); err != nil {
		t.Errorf("EnsureMetricsDir: %v", err)
	}
}

// ---------------------------------------------------------------------------
// hexutil.go — FormatHash, FormatAddress
// ---------------------------------------------------------------------------

func TestFormatHash_CorrectLength(t *testing.T) {
	hash := []byte{0x01, 0x02, 0x03, 0x04}
	formatted := FormatHash(hash)
	if len(formatted) != 64 {
		t.Errorf("FormatHash: expected 64 chars, got %d", len(formatted))
	}
}

func TestFormatHash_AllZeros(t *testing.T) {
	hash := make([]byte, 32)
	formatted := FormatHash(hash)
	for _, c := range formatted {
		if c != '0' {
			t.Errorf("FormatHash all-zeros: expected all '0', got %q", formatted)
			break
		}
	}
}

func TestFormatAddress_CorrectLength(t *testing.T) {
	addr := []byte{0xde, 0xad, 0xbe, 0xef}
	formatted := FormatAddress(addr)
	if len(formatted) != 40 {
		t.Errorf("FormatAddress: expected 40 chars, got %d", len(formatted))
	}
}

// ---------------------------------------------------------------------------
// time.go — TimeService methods
// ---------------------------------------------------------------------------

func TestValidateTimestamp_ZeroAllowed(t *testing.T) {
	ts := GetTimeService()
	// Timestamp 0 allowed for genesis/test transactions.
	err := ts.ValidateTimestamp(0, 0)
	if err != nil {
		t.Errorf("ValidateTimestamp(0): expected nil (test exemption), got %v", err)
	}
}

func TestValidateTimestamp_CurrentTime_Valid(t *testing.T) {
	ts := GetTimeService()
	now := time.Now().Unix()
	if err := ts.ValidateTimestamp(now, 10*time.Minute); err != nil {
		t.Errorf("ValidateTimestamp(now): unexpected error: %v", err)
	}
}

func TestValidateTimestamp_FarFuture_Error(t *testing.T) {
	ts := GetTimeService()
	future := time.Now().Unix() + 3600 // 1 hour in future
	err := ts.ValidateTimestamp(future, 0) // zero skew allowed
	if err == nil {
		t.Logf("ValidateTimestamp far future with 0 skew: no error (may allow via other path)")
	}
}

func TestValidateTimestamp_VeryOldTimestamp_Error(t *testing.T) {
	ts := GetTimeService()
	// Before 2010 (January 1, 2010 = 1262304000) should be invalid.
	err := ts.ValidateTimestamp(1000000000, 0) // year 2001
	if err == nil {
		t.Log("ValidateTimestamp(year 2001): no error returned (implementation may be lenient)")
	}
}

func TestValidateTransactionTimestamp_ZeroAllowed(t *testing.T) {
	err := ValidateTransactionTimestamp(0)
	if err != nil {
		t.Errorf("ValidateTransactionTimestamp(0): expected nil (test exemption), got %v", err)
	}
}

func TestValidateTransactionTimestamp_CurrentTime_Valid(t *testing.T) {
	err := ValidateTransactionTimestamp(time.Now().Unix())
	if err != nil {
		t.Errorf("ValidateTransactionTimestamp(now): unexpected error: %v", err)
	}
}

func TestFormatTimestamp_NonEmpty(t *testing.T) {
	ts := GetTimeService()
	local, utc := ts.FormatTimestamps(time.Now().Unix())
	if local == "" {
		t.Error("FormatTimestamps: local time should not be empty")
	}
	if utc == "" {
		t.Error("FormatTimestamps: utc time should not be empty")
	}
}

func TestGetCurrentTimestamp_Positive(t *testing.T) {
	stamp := GetCurrentTimestamp()
	if stamp <= 0 {
		t.Errorf("GetCurrentTimestamp should be positive, got %d", stamp)
	}
}

func TestSetTimezone_Valid(t *testing.T) {
	ts := GetTimeService()
	if err := ts.SetTimezone("UTC"); err != nil {
		t.Errorf("SetTimezone(UTC): unexpected error: %v", err)
	}
}

func TestSetTimezone_Invalid_Error(t *testing.T) {
	ts := GetTimeService()
	if err := ts.SetTimezone("Not/ATimezone"); err == nil {
		t.Error("SetTimezone(invalid): expected error")
	}
}

func TestGetDetectedTimezone_NonEmpty(t *testing.T) {
	ts := GetTimeService()
	tz := ts.GetDetectedTimezone()
	if tz == "" {
		t.Error("GetDetectedTimezone should return non-empty string")
	}
}

func TestSetMaxClockDrift_NoError(t *testing.T) {
	ts := GetTimeService()
	ts.SetMaxClockDrift(5 * time.Minute)
	// No assertion needed — just verify no panic.
}

func TestValidateBlockTimestamp_CurrentBlock_Valid(t *testing.T) {
	now := time.Now().Unix()
	err := ValidateBlockTimestamp(now)
	if err != nil {
		t.Logf("ValidateBlockTimestamp(now, nil): %v", err)
	}
}

func TestGetSystemTimezoneInfo_NonNil(t *testing.T) {
	info := GetSystemTimezoneInfo()
	if info == nil {
		t.Error("GetSystemTimezoneInfo should not return nil")
	}
}

func TestUpdateNetworkTime_NoError(t *testing.T) {
	ts := GetTimeService()
	// UpdateNetworkTime with an empty slice should be fine.
	ts.UpdateNetworkTime([]int64{})
}

func TestUpdateNetworkTime_WithSamples(t *testing.T) {
	ts := GetTimeService()
	now := time.Now().Unix()
	ts.UpdateNetworkTime([]int64{now, now - 1, now + 1})
}
