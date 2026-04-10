// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 40 - common time.go SetTimeSource, FormatTimestamp, PrintTimeServiceInfo, getTimeSourceString
package common

import (
	"testing"
)

// ---------------------------------------------------------------------------
// SetTimeSource
// ---------------------------------------------------------------------------

func TestSprint40_SetTimeSource_LocalSystem_NoPanic(t *testing.T) {
	ts := GetTimeService()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetTimeSource panicked: %v", r)
		}
	}()
	ts.SetTimeSource(LocalSystem)
}

func TestSprint40_SetTimeSource_NetworkConsensus_NoPanic(t *testing.T) {
	ts := GetTimeService()
	ts.SetTimeSource(NetworkConsensus)
	// Verify it was set
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if ts.timeSource != NetworkConsensus {
		t.Fatal("expected NetworkConsensus time source")
	}
}

func TestSprint40_SetTimeSource_Hybrid_NoPanic(t *testing.T) {
	ts := GetTimeService()
	ts.SetTimeSource(Hybrid)
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if ts.timeSource != Hybrid {
		t.Fatal("expected Hybrid time source")
	}
	// Reset to local
	ts.mu.RUnlock()
	ts.SetTimeSource(LocalSystem)
	ts.mu.RLock()
}

// ---------------------------------------------------------------------------
// FormatTimestamp
// ---------------------------------------------------------------------------

func TestSprint40_FormatTimestamp_NonEmpty(t *testing.T) {
	local, utc := FormatTimestamp(1700000000)
	if local == "" {
		t.Fatal("expected non-empty local time string")
	}
	if utc == "" {
		t.Fatal("expected non-empty UTC time string")
	}
}

func TestSprint40_FormatTimestamp_Zero(t *testing.T) {
	local, utc := FormatTimestamp(0)
	_ = local
	_ = utc
}

// ---------------------------------------------------------------------------
// PrintTimeServiceInfo
// ---------------------------------------------------------------------------

func TestSprint40_PrintTimeServiceInfo_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PrintTimeServiceInfo panicked: %v", r)
		}
	}()
	PrintTimeServiceInfo()
}

// ---------------------------------------------------------------------------
// getTimeSourceString — all branches
// ---------------------------------------------------------------------------

func TestSprint40_GetTimeSourceString_LocalSystem(t *testing.T) {
	ts := GetTimeService()
	ts.SetTimeSource(LocalSystem)
	s := ts.getTimeSourceString()
	if s == "" {
		t.Fatal("expected non-empty time source string")
	}
}

func TestSprint40_GetTimeSourceString_NetworkConsensus(t *testing.T) {
	ts := GetTimeService()
	ts.SetTimeSource(NetworkConsensus)
	s := ts.getTimeSourceString()
	if s == "" {
		t.Fatal("expected non-empty time source string for NetworkConsensus")
	}
}

func TestSprint40_GetTimeSourceString_Hybrid(t *testing.T) {
	ts := GetTimeService()
	ts.SetTimeSource(Hybrid)
	s := ts.getTimeSourceString()
	if s == "" {
		t.Fatal("expected non-empty time source string for Hybrid")
	}
	ts.SetTimeSource(LocalSystem) // reset
}

// ---------------------------------------------------------------------------
// ValidateNonceFormat — 16-char hex nonce
// ---------------------------------------------------------------------------

func TestSprint40_ValidateNonceFormat_Valid16HexChars_NoError(t *testing.T) {
	// nonce must be exactly 16 chars (8 bytes hex)
	err := ValidateNonceFormat("aabbccddeeff0011")
	if err != nil {
		t.Fatalf("expected nil error for valid 16-char hex nonce, got: %v", err)
	}
}

func TestSprint40_ValidateNonceFormat_TooShort_Error(t *testing.T) {
	err := ValidateNonceFormat("aabb")
	if err == nil {
		t.Fatal("expected error for short nonce")
	}
}

func TestSprint40_ValidateNonceFormat_NonHex_Error(t *testing.T) {
	// 16 chars but not valid hex
	err := ValidateNonceFormat("gggggggggggggggg")
	if err == nil {
		t.Fatal("expected error for non-hex nonce")
	}
}

func TestSprint40_ValidateNonceFormat_TooLong_Error(t *testing.T) {
	err := ValidateNonceFormat("aabbccddeeff001122") // 18 chars
	if err == nil {
		t.Fatal("expected error for too-long nonce")
	}
}

// ---------------------------------------------------------------------------
// NonceToBytes — valid 16-char hex nonce
// ---------------------------------------------------------------------------

func TestSprint40_NonceToBytes_Valid16Chars(t *testing.T) {
	nonce := "aabbccddeeff0011"
	b, err := NonceToBytes(nonce)
	if err != nil {
		t.Fatalf("NonceToBytes error: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty bytes")
	}
}
