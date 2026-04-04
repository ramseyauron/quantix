// PEPPER Sprint2 — common package coverage push
// Targets: config.go paths, hexutil.go, types.go
package common

import (
	"os"
	"testing"
)

// ── Config path functions ─────────────────────────────────────────────────────

const testAddr = "127.0.0.1:8080"

func TestGetStateDBPath(t *testing.T) {
	path := GetStateDBPath(testAddr)
	if path == "" {
		t.Error("GetStateDBPath should return non-empty string")
	}
}

func TestGetBlockchainDataDir(t *testing.T) {
	dir := GetBlockchainDataDir(testAddr)
	if dir == "" {
		t.Error("GetBlockchainDataDir should return non-empty string")
	}
}

func TestGetBlocksDir(t *testing.T) {
	dir := GetBlocksDir(testAddr)
	if dir == "" {
		t.Error("GetBlocksDir should return non-empty string")
	}
}

func TestGetIndexDir(t *testing.T) {
	dir := GetIndexDir(testAddr)
	if dir == "" {
		t.Error("GetIndexDir should return non-empty string")
	}
}

func TestGetStateDir(t *testing.T) {
	dir := GetStateDir(testAddr)
	if dir == "" {
		t.Error("GetStateDir should return non-empty string")
	}
}

func TestGetMetricsDir(t *testing.T) {
	dir := GetMetricsDir(testAddr)
	if dir == "" {
		t.Error("GetMetricsDir should return non-empty string")
	}
}

func TestGetChainStatePath(t *testing.T) {
	path := GetChainStatePath(testAddr)
	if path == "" {
		t.Error("GetChainStatePath should return non-empty string")
	}
}

func TestGetCheckpointPath(t *testing.T) {
	path := GetCheckpointPath(testAddr)
	if path == "" {
		t.Error("GetCheckpointPath should return non-empty string")
	}
}

func TestGetTPSMetricsPath(t *testing.T) {
	path := GetTPSMetricsPath(testAddr)
	if path == "" {
		t.Error("GetTPSMetricsPath should return non-empty string")
	}
}

func TestGetKeysDataDir(t *testing.T) {
	dir := GetKeysDataDir(testAddr)
	if dir == "" {
		t.Error("GetKeysDataDir should return non-empty string")
	}
}

func TestGetPrivateKeyPath(t *testing.T) {
	path := GetPrivateKeyPath(testAddr)
	if path == "" {
		t.Error("GetPrivateKeyPath should return non-empty string")
	}
}

func TestGetPublicKeyPath(t *testing.T) {
	path := GetPublicKeyPath(testAddr)
	if path == "" {
		t.Error("GetPublicKeyPath should return non-empty string")
	}
}

// ── WriteJSONToFile ───────────────────────────────────────────────────────────

func TestWriteJSONToFile_Valid(t *testing.T) {
	dir := t.TempDir()
	DataDir = dir
	defer func() { DataDir = "data" }()

	data := map[string]interface{}{"key": "value", "num": 42}
	if err := WriteJSONToFile(data, "test.json"); err != nil {
		t.Fatalf("WriteJSONToFile: %v", err)
	}
	// Verify file was created
	path := dir + "/output/test.json"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to be created at output/test.json")
	}
}

// ── hexutil.go ────────────────────────────────────────────────────────────────

func TestFormatNonce_RoundTrip(t *testing.T) {
	for _, n := range []uint64{0, 1, 255, 65535, 1000000} {
		s := FormatNonce(n)
		if s == "" {
			t.Errorf("FormatNonce(%d) returned empty string", n)
		}
		parsed, err := ParseNonce(s)
		if err != nil {
			t.Errorf("ParseNonce(%q): %v", s, err)
		}
		if parsed != n {
			t.Errorf("FormatNonce/ParseNonce round-trip: want %d, got %d", n, parsed)
		}
	}
}

func TestFormatNonce32(t *testing.T) {
	s := FormatNonce32(42)
	if s == "" {
		t.Error("FormatNonce32 returned empty string")
	}
}

func TestParseNonce_InvalidStr(t *testing.T) {
	_, err := ParseNonce("not_a_number_xyz")
	if err == nil {
		t.Error("expected error for invalid nonce string")
	}
}

func TestGenerateRandomNonce_NotEmpty(t *testing.T) {
	n1 := GenerateRandomNonce()
	if n1 == "" {
		t.Error("GenerateRandomNonce returned empty string")
	}
}

func TestGenerateRandomNonceUint64_NoPanic(t *testing.T) {
	_ = GenerateRandomNonceUint64()
}

// ── types.go ──────────────────────────────────────────────────────────────────

func TestSpxHash_MultipleInputs(t *testing.T) {
	inputs := [][]byte{
		[]byte("pepper sprint2 test 1"),
		[]byte("pepper sprint2 test 2"),
		{0xFF, 0x00, 0xAB},
	}
	for _, data := range inputs {
		hash := SpxHash(data)
		if len(hash) == 0 {
			t.Errorf("SpxHash(%q) returned empty", data)
		}
	}
}
