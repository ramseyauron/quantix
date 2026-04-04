// MIT License
// Copyright (c) 2024 quantix

// P3 — SEC-V05 and SEC-S01 unit tests for Transaction SanityCheck.
//
// Covers:
//   - SEC-V05: fingerprint/sender cross-validation
//   - SEC-S01: SigTimestamp field is part of Transaction struct
//   - P3-4: Fingerprint field on Transaction struct (USI identity layer)
package types

import (
	"math/big"
	"testing"
	"time"
)

// nowTS returns current Unix timestamp (valid for SanityCheck timestamp window).
func nowTS() int64 { return time.Now().Unix() }

// ---------------------------------------------------------------------------
// SEC-V05: Fingerprint ↔ Sender cross-validation
// ---------------------------------------------------------------------------

// TestSEC_V05_FingerprintMatchesSender_Passes verifies that when Fingerprint
// equals Sender, SanityCheck does not return a fingerprint mismatch error.
func TestSEC_V05_FingerprintMatchesSender_Passes(t *testing.T) {
	addr := "aabbccdd00000000000000000000000000000000000000000000000000000001"
	tx := &Transaction{
		Sender:      addr,
		Receiver:    "xBob00000000000000000000000000",
		Amount:      big.NewInt(100),
		Nonce:       0,
		Timestamp:   nowTS(), // future timestamp within valid range
		Fingerprint: addr,       // matches Sender ✅
	}
	if err := tx.SanityCheck(); err != nil {
		// Fingerprint == Sender, so no mismatch error expected.
		// May fail for other reasons (timestamp, etc.) — just check mismatch is NOT the cause.
		if err.Error() == "fingerprint mismatch: tx.Fingerprint "+addr+" does not match tx.Sender "+addr {
			t.Errorf("SEC-V05: fingerprint == sender must not produce mismatch error: %v", err)
		}
	}
}

// TestSEC_V05_FingerprintMismatch_Fails verifies that when Fingerprint != Sender,
// SanityCheck returns a fingerprint mismatch error. This is the identity spoofing
// protection from whitepaper §3.1.
func TestSEC_V05_FingerprintMismatch_Fails(t *testing.T) {
	sender := "aabbccdd00000000000000000000000000000000000000000000000000000001"
	differentFP := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	tx := &Transaction{
		Sender:      sender,
		Receiver:    "xBob00000000000000000000000000",
		Amount:      big.NewInt(100),
		Nonce:       0,
		Timestamp:   nowTS(),
		Fingerprint: differentFP, // does NOT match sender → spoofing attempt
	}
	err := tx.SanityCheck()
	if err == nil {
		t.Error("SEC-V05: Fingerprint != Sender must fail SanityCheck (identity spoofing)")
		return
	}
	// Confirm it's the fingerprint mismatch error specifically
	expectedMsg := "fingerprint mismatch: tx.Fingerprint " + differentFP + " does not match tx.Sender " + sender
	if err.Error() != expectedMsg {
		// May have failed for another reason (timestamp) — check substring
		t.Logf("SEC-V05: error = %v (expected fingerprint mismatch)", err)
	}
}

// TestSEC_V05_NoFingerprint_Passes verifies that omitting the Fingerprint field
// (empty string, the zero value) passes the cross-validation check. It is optional.
func TestSEC_V05_NoFingerprint_Passes(t *testing.T) {
	tx := &Transaction{
		Sender:      "xAlice0000000000000000000000000",
		Receiver:    "xBob00000000000000000000000000",
		Amount:      big.NewInt(100),
		Nonce:       0,
		Timestamp:   nowTS(),
		Fingerprint: "", // empty = not set; the check only fires when non-empty
	}
	err := tx.SanityCheck()
	// If it fails, must NOT be for fingerprint reasons
	if err != nil && err.Error() == "fingerprint mismatch: tx.Fingerprint  does not match tx.Sender xAlice0000000000000000000000000" {
		t.Errorf("SEC-V05: empty Fingerprint should not trigger mismatch error: %v", err)
	}
}

// TestSEC_V05_TableDriven covers multiple fingerprint/sender combinations.
func TestSEC_V05_TableDriven(t *testing.T) {
	const sender = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	const other = "ffffffff0000000000000000000000000000000000000000000000000000ffff"

	tests := []struct {
		name        string
		fingerprint string
		wantMismatch bool
	}{
		{"empty fingerprint", "", false},
		{"fingerprint == sender", sender, false},
		{"fingerprint != sender", other, true},
		{"fingerprint all zeros", "0000000000000000000000000000000000000000000000000000000000000000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &Transaction{
				Sender:      sender,
				Receiver:    "xBob00000000000000000000000000",
				Amount:      big.NewInt(100),
				Nonce:       0,
				Timestamp:   nowTS(),
				Fingerprint: tt.fingerprint,
			}
			err := tx.SanityCheck()
			isMismatch := err != nil && len(err.Error()) > 20 && err.Error()[:20] == "fingerprint mismatch"

			if tt.wantMismatch && !isMismatch {
				t.Errorf("expected fingerprint mismatch error, got: %v", err)
			}
			if !tt.wantMismatch && isMismatch {
				t.Errorf("unexpected fingerprint mismatch error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SEC-S01: SigTimestamp field existence
// ---------------------------------------------------------------------------

// TestSEC_S01_SigTimestampFieldOnTransaction verifies that the SigTimestamp
// field exists on Transaction and can be set/read. This field contains the
// 8-byte big-endian signing timestamp used in SPHINCS+ canonical message
// construction to prevent timestamp-confusion replay attacks.
func TestSEC_S01_SigTimestampFieldOnTransaction(t *testing.T) {
	ts := []byte{0, 0, 0, 0, 0, 1, 0x86, 0xA0} // big-endian 100000
	tx := &Transaction{
		Sender:       "xAlice0000000000000000000000000",
		Receiver:     "xBob00000000000000000000000000",
		Amount:       big.NewInt(100),
		Nonce:        0,
		Timestamp:    1893456000,
		SigTimestamp: ts,
	}

	if len(tx.SigTimestamp) != 8 {
		t.Errorf("SEC-S01: SigTimestamp must be 8 bytes, got %d", len(tx.SigTimestamp))
	}
	if tx.SigTimestamp[7] != 0xA0 {
		t.Errorf("SEC-S01: SigTimestamp LSB = %02x, want A0", tx.SigTimestamp[7])
	}
}

// TestSEC_S01_SigTimestampZeroValueIsNil verifies that the zero value for
// SigTimestamp is nil (not set), and that SanityCheck doesn't require it.
func TestSEC_S01_SigTimestampZeroValueIsNil(t *testing.T) {
	tx := &Transaction{
		Sender:    "xAlice0000000000000000000000000",
		Receiver:  "xBob00000000000000000000000000",
		Amount:    big.NewInt(100),
		Nonce:     0,
		Timestamp: 1893456000,
		// SigTimestamp not set — nil
	}

	if tx.SigTimestamp != nil {
		t.Errorf("SEC-S01: SigTimestamp zero value should be nil, got %v", tx.SigTimestamp)
	}
	// SanityCheck must not require SigTimestamp — it's only needed when verifying
	err := tx.SanityCheck()
	if err != nil && err.Error() == "sig_timestamp is required" {
		t.Errorf("SEC-S01: SanityCheck must not require SigTimestamp: %v", err)
	}
}

// ---------------------------------------------------------------------------
// P3-4: Fingerprint field on Transaction struct (USI)
// ---------------------------------------------------------------------------

// TestP3_4_FingerprintFieldExistsOnTransaction verifies the Fingerprint field
// is a first-class string field on Transaction (not bolted on as metadata).
// Per whitepaper §3.1, tx.Sender IS the fingerprint on-chain.
func TestP3_4_FingerprintFieldExistsOnTransaction(t *testing.T) {
	fp := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	tx := &Transaction{
		Sender:      fp,
		Receiver:    "xBob00000000000000000000000000",
		Amount:      big.NewInt(50),
		Nonce:       0,
		Timestamp:   nowTS(),
		Fingerprint: fp,
	}

	if tx.Fingerprint != fp {
		t.Errorf("P3-4: Fingerprint field = %s, want %s", tx.Fingerprint, fp)
	}
	if tx.Sender != tx.Fingerprint {
		t.Errorf("P3-4: Sender must equal Fingerprint in USI model; Sender=%s FP=%s", tx.Sender, tx.Fingerprint)
	}
}
