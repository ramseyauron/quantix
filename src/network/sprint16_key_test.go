package network

// Coverage Sprint 16 — network/key.go: FromNodeID, FromString,
// network/manager.go: NewNetworkKeyManager.

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Key.FromNodeID
// ---------------------------------------------------------------------------

func TestFromNodeID_CopiesValue(t *testing.T) {
	var src Key
	src[0] = 0xAB
	src[1] = 0xCD

	var k Key
	k.FromNodeID(src)

	if k[0] != 0xAB || k[1] != 0xCD {
		t.Errorf("FromNodeID: expected key[0]=0xAB key[1]=0xCD, got [0]=%X [1]=%X", k[0], k[1])
	}
}

func TestFromNodeID_ZeroSource(t *testing.T) {
	var src Key
	var k Key
	k[0] = 0xFF // pre-set to something non-zero

	k.FromNodeID(src)

	if k[0] != 0 {
		t.Errorf("FromNodeID(zero): expected key[0]=0, got %X", k[0])
	}
}

func TestFromNodeID_Roundtrip_AsNodeID(t *testing.T) {
	var src Key
	src[15] = 0x42
	src[31] = 0x99

	var k Key
	k.FromNodeID(src)

	if k != src {
		t.Errorf("FromNodeID roundtrip: key mismatch")
	}
}

// ---------------------------------------------------------------------------
// Key.FromString
// ---------------------------------------------------------------------------

func TestFromString_ValidInput_NoError(t *testing.T) {
	var k Key
	if err := k.FromString("test-node-id-quantix"); err != nil {
		t.Errorf("FromString valid: unexpected error: %v", err)
	}
}

func TestFromString_NonEmptyOutput(t *testing.T) {
	var k Key
	k.FromString("quantix-node")
	// Should produce a non-zero key.
	allZero := true
	for _, b := range k {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("FromString: produced all-zero key for non-empty input")
	}
}

func TestFromString_Deterministic(t *testing.T) {
	var k1, k2 Key
	k1.FromString("deterministic-input-001")
	k2.FromString("deterministic-input-001")
	if k1 != k2 {
		t.Error("FromString: same input should produce same key (deterministic hash)")
	}
}

func TestFromString_DifferentInputs_DifferentKeys(t *testing.T) {
	var k1, k2 Key
	k1.FromString("input-alpha")
	k2.FromString("input-beta")
	if k1 == k2 {
		t.Error("FromString: different inputs should produce different keys")
	}
}

func TestFromString_EmptyInput_NoError(t *testing.T) {
	var k Key
	// Empty string should hash to something (BLAKE3 accepts empty input).
	if err := k.FromString(""); err != nil {
		t.Errorf("FromString empty string: unexpected error: %v", err)
	}
}
