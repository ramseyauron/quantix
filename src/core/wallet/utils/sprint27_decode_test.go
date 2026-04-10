// Sprint 27c — core/chain_maker LoadChainCheckpoint + wallet/utils DecodeBase32.
// Chain maker tests are in core package; wallet/utils tests in utils_test package.
package utils_test

import (
	"bytes"
	"encoding/base32"
	"testing"

	utils "github.com/ramseyauron/quantix/src/core/wallet/utils"
)

// ─── DecodeBase32 ─────────────────────────────────────────────────────────────

func TestSprint27_DecodeBase32_ValidInput_Roundtrip(t *testing.T) {
	original := []byte("quantix-wallet-test-data")
	// EncodeBase32 uses NoPadding, but DecodeBase32 uses StdEncoding (needs padding)
	encoded := utils.EncodeBase32(original)
	// Add padding for StdEncoding
	padded := encoded
	if pad := len(padded) % 8; pad != 0 {
		for i := 0; i < 8-pad; i++ {
			padded += "="
		}
	}
	decoded, err := utils.DecodeBase32(padded)
	if err != nil {
		t.Fatalf("DecodeBase32: %v", err)
	}
	if !bytes.Equal(decoded, original) {
		t.Errorf("roundtrip failed: got %x, want %x", decoded, original)
	}
}

func TestSprint27_DecodeBase32_InvalidInput_Error(t *testing.T) {
	_, err := utils.DecodeBase32("not-valid-base32-!!!")
	if err == nil {
		t.Error("expected error for invalid base32 input")
	}
}

func TestSprint27_DecodeBase32_EmptyString_ReturnsEmpty(t *testing.T) {
	decoded, err := utils.DecodeBase32("")
	if err != nil {
		t.Fatalf("DecodeBase32 empty: %v", err)
	}
	if len(decoded) != 0 {
		t.Errorf("expected empty decoded, got %d bytes", len(decoded))
	}
}

func TestSprint27_DecodeBase32_StdEncodingDirect(t *testing.T) {
	// Verify DecodeBase32 uses StdEncoding by testing a known encoding
	known := []byte("hello")
	encoded := base32.StdEncoding.EncodeToString(known)
	decoded, err := utils.DecodeBase32(encoded)
	if err != nil {
		t.Fatalf("DecodeBase32 known: %v", err)
	}
	if !bytes.Equal(decoded, known) {
		t.Errorf("got %x, want %x", decoded, known)
	}
}
