// MIT License
// Copyright (c) 2024 quantix

// Q14 — Tests for src/common package (previously 0% coverage)
// Covers: hexutil (Address, ValidateAddress, Bytes2Hex, Hex2Bytes, FormatNonce,
//         ParseNonce, NonceToBytes, BytesToNonce, GenerateRandomNonce,
//         IsValidHexString, ParseBigInt, FormatBigInt),
//         types (SpxHash), and TimeService
package common

import (
	"math/big"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// SpxHash
// ---------------------------------------------------------------------------

func TestSpxHash_NonEmpty(t *testing.T) {
	h := SpxHash([]byte("hello"))
	if len(h) == 0 {
		t.Fatal("SpxHash should return non-empty bytes")
	}
}

func TestSpxHash_Deterministic(t *testing.T) {
	data := []byte("quantix-test-vector")
	h1 := SpxHash(data)
	h2 := SpxHash(data)
	if string(h1) != string(h2) {
		t.Error("SpxHash must be deterministic")
	}
}

func TestSpxHash_DifferentInputsDifferentOutput(t *testing.T) {
	h1 := SpxHash([]byte("a"))
	h2 := SpxHash([]byte("b"))
	if string(h1) == string(h2) {
		t.Error("SpxHash: different inputs must produce different outputs")
	}
}

func TestSpxHash_EmptyInput(t *testing.T) {
	h := SpxHash([]byte{})
	if len(h) == 0 {
		t.Fatal("SpxHash of empty input should still return hash bytes")
	}
}

// ---------------------------------------------------------------------------
// Bytes2Hex / Hex2Bytes
// ---------------------------------------------------------------------------

func TestBytes2Hex_RoundTrip(t *testing.T) {
	input := []byte{0xde, 0xad, 0xbe, 0xef}
	hex := Bytes2Hex(input)
	if hex != "deadbeef" {
		t.Errorf("Bytes2Hex: got %q want deadbeef", hex)
	}
}

func TestHex2Bytes_Valid(t *testing.T) {
	b, err := Hex2Bytes("deadbeef")
	if err != nil {
		t.Fatalf("Hex2Bytes: %v", err)
	}
	if len(b) != 4 || b[0] != 0xde || b[3] != 0xef {
		t.Errorf("Hex2Bytes: unexpected result %v", b)
	}
}

func TestHex2Bytes_Invalid(t *testing.T) {
	_, err := Hex2Bytes("not-hex!")
	if err == nil {
		t.Error("expected error for invalid hex string")
	}
}

func TestBytes2Hex_Hex2Bytes_RoundTrip(t *testing.T) {
	original := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
	hex := Bytes2Hex(original)
	back, err := Hex2Bytes(hex)
	if err != nil {
		t.Fatalf("Hex2Bytes round-trip: %v", err)
	}
	if string(original) != string(back) {
		t.Errorf("round-trip mismatch: %v vs %v", original, back)
	}
}

// ---------------------------------------------------------------------------
// FormatNonce / ParseNonce / NonceToBytes / BytesToNonce
// ---------------------------------------------------------------------------

func TestFormatNonce_Length(t *testing.T) {
	n := FormatNonce(12345)
	// FormatNonce returns a 16-char hex string
	if len(n) != 16 {
		t.Errorf("FormatNonce: expected 16 chars, got %d (%q)", len(n), n)
	}
}

func TestFormatNonce32_Length(t *testing.T) {
	n := FormatNonce32(99)
	if len(n) != 32 {
		t.Errorf("FormatNonce32: expected 32 chars, got %d (%q)", len(n), n)
	}
}

func TestParseNonce_Valid(t *testing.T) {
	formatted := FormatNonce(42)
	n, err := ParseNonce(formatted)
	if err != nil {
		t.Fatalf("ParseNonce: %v", err)
	}
	if n != 42 {
		t.Errorf("ParseNonce: got %d want 42", n)
	}
}

func TestParseNonce_Invalid(t *testing.T) {
	_, err := ParseNonce("ZZZZ")
	if err == nil {
		t.Error("expected error for non-hex nonce")
	}
}

func TestNonceToBytes_ValidRoundTrip(t *testing.T) {
	nonce := FormatNonce(1234567890)
	b, err := NonceToBytes(nonce)
	if err != nil {
		t.Fatalf("NonceToBytes: %v", err)
	}
	if len(b) != 8 {
		t.Errorf("expected 8 bytes, got %d", len(b))
	}
	back, err := BytesToNonce(b)
	if err != nil {
		t.Fatalf("BytesToNonce: %v", err)
	}
	if back != nonce {
		t.Errorf("nonce round-trip: got %q want %q", back, nonce)
	}
}

func TestBytesToNonce_WrongLength(t *testing.T) {
	_, err := BytesToNonce([]byte{0x01, 0x02}) // only 2 bytes
	if err == nil {
		t.Error("expected error for non-8-byte input")
	}
}

func TestZeroNonce(t *testing.T) {
	z := ZeroNonce()
	if len(z) != 16 {
		t.Errorf("ZeroNonce: expected 16 chars, got %d", len(z))
	}
	for _, c := range z {
		if c != '0' {
			t.Errorf("ZeroNonce: expected all zeros, got %q", z)
			break
		}
	}
}

func TestMaxNonce(t *testing.T) {
	m := MaxNonce()
	if len(m) != 16 {
		t.Errorf("MaxNonce: expected 16 chars, got %d", len(m))
	}
	// Max nonce should be all 'f's
	for _, c := range m {
		if c != 'f' {
			t.Errorf("MaxNonce: expected all 'f', got %q", m)
			break
		}
	}
}

// ---------------------------------------------------------------------------
// GenerateRandomNonce / GenerateRandomNonceUint64
// ---------------------------------------------------------------------------

func TestGenerateRandomNonce_Length(t *testing.T) {
	n := GenerateRandomNonce()
	if len(n) != 16 {
		t.Errorf("GenerateRandomNonce: expected 16 chars, got %d (%q)", len(n), n)
	}
}

func TestGenerateRandomNonce_IsHex(t *testing.T) {
	n := GenerateRandomNonce()
	if !IsValidHexString(n) {
		t.Errorf("GenerateRandomNonce: not valid hex: %q", n)
	}
}

func TestGenerateRandomNonce_DifferentEachTime(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		n := GenerateRandomNonce()
		if seen[n] {
			t.Errorf("GenerateRandomNonce produced duplicate: %q", n)
		}
		seen[n] = true
	}
}

func TestGenerateRandomNonceUint64_NonZero(t *testing.T) {
	seen := make(map[uint64]bool)
	allZero := true
	for i := 0; i < 10; i++ {
		n := GenerateRandomNonceUint64()
		if n != 0 {
			allZero = false
		}
		seen[n] = true
	}
	if allZero {
		t.Error("GenerateRandomNonceUint64: all values were zero")
	}
}

// ---------------------------------------------------------------------------
// IsValidHexString
// ---------------------------------------------------------------------------

func TestIsValidHexString_Valid(t *testing.T) {
	cases := []string{"", "deadbeef", "0123456789abcdef", "ABCDEF"}
	for _, c := range cases {
		if !IsValidHexString(c) {
			t.Errorf("IsValidHexString(%q) = false, want true", c)
		}
	}
}

func TestIsValidHexString_Invalid(t *testing.T) {
	cases := []string{"xyz", "0xdeadbeef", "hello!", "12 34"}
	for _, c := range cases {
		if IsValidHexString(c) {
			t.Errorf("IsValidHexString(%q) = true, want false", c)
		}
	}
}

// ---------------------------------------------------------------------------
// ParseBigInt / FormatBigInt
// ---------------------------------------------------------------------------

func TestParseBigInt_Valid(t *testing.T) {
	b, err := ParseBigInt("ff")
	if err != nil {
		t.Fatalf("ParseBigInt: %v", err)
	}
	if b.Cmp(big.NewInt(255)) != 0 {
		t.Errorf("ParseBigInt(ff): got %s want 255", b)
	}
}

func TestParseBigInt_Invalid(t *testing.T) {
	_, err := ParseBigInt("not-hex")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestFormatBigInt_PadsCorrectly(t *testing.T) {
	s := FormatBigInt(big.NewInt(255), 4) // should be 0xff padded to 4 hex chars
	if len(s) < 4 {
		t.Errorf("FormatBigInt padding: expected at least 4 chars, got %q", s)
	}
}

// ---------------------------------------------------------------------------
// Address
// ---------------------------------------------------------------------------

func TestAddress_Valid32Bytes(t *testing.T) {
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i)
	}
	addr, err := Address(hash)
	if err != nil {
		t.Fatalf("Address: %v", err)
	}
	if addr == "" {
		t.Error("Address: should return non-empty string")
	}
	// Should start with 0x
	if !strings.HasPrefix(addr, "0x") {
		t.Errorf("Address: expected 0x prefix, got %q", addr)
	}
}

func TestAddress_WrongLength_Error(t *testing.T) {
	_, err := Address([]byte{0x01, 0x02})
	if err == nil {
		t.Error("expected error for non-32-byte hash")
	}
}

// ---------------------------------------------------------------------------
// ValidateAddress
// ---------------------------------------------------------------------------

func TestValidateAddress_MissingPrefix_Error(t *testing.T) {
	_, err := ValidateAddress("deadbeef00000000000000000000000000000000")
	if err == nil {
		t.Error("expected error for address without 0x prefix")
	}
}

func TestValidateAddress_TooShort_Error(t *testing.T) {
	_, err := ValidateAddress("0x1234")
	if err == nil {
		t.Error("expected error for too-short address")
	}
}

func TestValidateAddress_Valid40Hex(t *testing.T) {
	// ValidateAddress requires EIP-55 checksum — use Address() to generate a valid one
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i + 1)
	}
	addr, err := Address(hash)
	if err != nil {
		t.Fatalf("Address: %v", err)
	}
	ok, err := ValidateAddress(addr)
	if err != nil {
		t.Fatalf("ValidateAddress of generated address: %v", err)
	}
	if !ok {
		t.Error("expected Address()-generated address to be valid")
	}
}

// ---------------------------------------------------------------------------
// GetNodeIdentifier / GetNodePortFromAddress / path helpers
// ---------------------------------------------------------------------------

func TestGetNodeIdentifier_NonEmpty(t *testing.T) {
	id := GetNodeIdentifier("10.0.0.1:32307")
	if id == "" {
		t.Error("GetNodeIdentifier should return non-empty string")
	}
}

func TestGetNodePortFromAddress_ExtractsPort(t *testing.T) {
	port := GetNodePortFromAddress("10.0.0.1:8560")
	if port != "8560" {
		t.Errorf("GetNodePortFromAddress: got %q want 8560", port)
	}
}

func TestGetNodeDataDir_NonEmpty(t *testing.T) {
	dir := GetNodeDataDir("node-1")
	if dir == "" {
		t.Error("GetNodeDataDir should return non-empty string")
	}
}

func TestGetLevelDBPath_ContainsAddress(t *testing.T) {
	path := GetLevelDBPath("mynode")
	if !strings.Contains(path, "mynode") {
		t.Errorf("GetLevelDBPath should contain address, got %q", path)
	}
}

// ---------------------------------------------------------------------------
// TimeService
// ---------------------------------------------------------------------------

func TestGetTimeService_NotNil(t *testing.T) {
	ts := GetTimeService()
	if ts == nil {
		t.Fatal("GetTimeService should return non-nil")
	}
}

func TestTimeService_NowUnix_Positive(t *testing.T) {
	ts := GetTimeService()
	now := ts.NowUnix()
	if now <= 0 {
		t.Errorf("NowUnix should be positive, got %d", now)
	}
}

func TestTimeService_Now_NonZero(t *testing.T) {
	ts := GetTimeService()
	if ts.Now().IsZero() {
		t.Error("TimeService.Now() should not be zero")
	}
}

func TestTimeService_GetCurrentTimeInfo_NonNil(t *testing.T) {
	ts := GetTimeService()
	info := ts.GetCurrentTimeInfo()
	if info == nil {
		t.Error("GetCurrentTimeInfo should return non-nil")
	}
}

func TestTimeService_FormatTimestamps_NonEmpty(t *testing.T) {
	ts := GetTimeService()
	local, utc := ts.FormatTimestamps(1704067200) // 2024-01-01 00:00:00 UTC
	if local == "" || utc == "" {
		t.Errorf("FormatTimestamps returned empty strings: local=%q utc=%q", local, utc)
	}
}

func TestTimeService_GetRelativeTime_NonEmpty(t *testing.T) {
	ts := GetTimeService()
	rel := ts.GetRelativeTime(ts.NowUnix() - 3600) // 1 hour ago
	if rel == "" {
		t.Error("GetRelativeTime should return non-empty string")
	}
}

// ---------------------------------------------------------------------------
// BytesToHexWithPrefix / HexToBytesWithoutPrefix
// ---------------------------------------------------------------------------

func TestBytesToHexWithPrefix_HasPrefix(t *testing.T) {
	s := BytesToHexWithPrefix([]byte{0xab, 0xcd})
	if !strings.HasPrefix(s, "0x") {
		t.Errorf("expected 0x prefix, got %q", s)
	}
}

func TestHexToBytesWithoutPrefix_Valid(t *testing.T) {
	b, err := HexToBytesWithoutPrefix("0xabcd")
	if err != nil {
		t.Fatalf("HexToBytesWithoutPrefix: %v", err)
	}
	if len(b) != 2 || b[0] != 0xab || b[1] != 0xcd {
		t.Errorf("unexpected result: %v", b)
	}
}

func TestHexToBytesWithoutPrefix_NoPrefix(t *testing.T) {
	b, err := HexToBytesWithoutPrefix("abcd")
	if err != nil {
		t.Fatalf("HexToBytesWithoutPrefix without prefix: %v", err)
	}
	if len(b) != 2 {
		t.Errorf("expected 2 bytes, got %d", len(b))
	}
}
