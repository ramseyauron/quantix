// MIT License
//
// Copyright (c) 2024 quantix

// P.E.P.P.E.R. SEC-A01 tests — dual-format address validation.
// Covers 91488ad (SEC-A01 ValidateAddress fix) and 62ef33c (genesis allocation
// 64-char address support).
package core

import (
	"math/big"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newMinimalBC returns a Blockchain with devnet chain params, no storage.
func newMinimalBC(t *testing.T) *Blockchain {
	t.Helper()
	return &Blockchain{chainParams: GetDevnetChainParams()}
}

// ---------------------------------------------------------------------------
// SEC-A01: ValidateAddress dual-format support
// ---------------------------------------------------------------------------

// TestSECA01_ValidateAddress_40Char verifies that legacy 40-char hex addresses pass.
func TestSECA01_ValidateAddress_40Char(t *testing.T) {
	bc := newMinimalBC(t)
	addr40 := strings.Repeat("a", 40) // "aaa...a" — 40 hex chars
	if !bc.ValidateAddress(addr40) {
		t.Errorf("40-char hex address should be valid (SEC-A01): %s", addr40)
	}
}

// TestSECA01_ValidateAddress_64Char verifies that 64-char SPHINCS+ fingerprints pass.
func TestSECA01_ValidateAddress_64Char(t *testing.T) {
	bc := newMinimalBC(t)
	// Realistic 64-char SPHINCS+ fingerprint (SHA-256 hex)
	addr64 := "a3f2b1c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2"
	if !bc.ValidateAddress(addr64) {
		t.Errorf("64-char hex address should be valid (SEC-A01): %s", addr64)
	}
}

// TestSECA01_ValidateAddress_AllZero_40Char verifies GenesisVaultAddress (40-char all-zeros) is valid.
func TestSECA01_ValidateAddress_AllZero_40Char(t *testing.T) {
	bc := newMinimalBC(t)
	// GenesisVaultAddress is 40 hex chars
	if !bc.ValidateAddress(GenesisVaultAddress) {
		t.Errorf("GenesisVaultAddress should be valid: %s", GenesisVaultAddress)
	}
}

// TestSECA01_ValidateAddress_39Char_Rejected verifies 39-char addresses fail.
func TestSECA01_ValidateAddress_39Char_Rejected(t *testing.T) {
	bc := newMinimalBC(t)
	if bc.ValidateAddress(strings.Repeat("a", 39)) {
		t.Error("39-char address should be invalid")
	}
}

// TestSECA01_ValidateAddress_41Char_Rejected verifies 41-char addresses fail.
func TestSECA01_ValidateAddress_41Char_Rejected(t *testing.T) {
	bc := newMinimalBC(t)
	if bc.ValidateAddress(strings.Repeat("a", 41)) {
		t.Error("41-char address should be invalid")
	}
}

// TestSECA01_ValidateAddress_63Char_Rejected verifies 63-char addresses fail.
func TestSECA01_ValidateAddress_63Char_Rejected(t *testing.T) {
	bc := newMinimalBC(t)
	if bc.ValidateAddress(strings.Repeat("a", 63)) {
		t.Error("63-char address should be invalid")
	}
}

// TestSECA01_ValidateAddress_65Char_Rejected verifies 65-char addresses fail.
func TestSECA01_ValidateAddress_65Char_Rejected(t *testing.T) {
	bc := newMinimalBC(t)
	if bc.ValidateAddress(strings.Repeat("a", 65)) {
		t.Error("65-char address should be invalid")
	}
}

// TestSECA01_ValidateAddress_NonHex_Rejected verifies non-hex chars fail.
func TestSECA01_ValidateAddress_NonHex_Rejected(t *testing.T) {
	bc := newMinimalBC(t)
	// 40 chars but contains 'z' — not valid hex
	addr := strings.Repeat("z", 40)
	if bc.ValidateAddress(addr) {
		t.Error("non-hex address should be invalid")
	}
}

// TestSECA01_ValidateAddress_MixedCase_Accepted verifies case-insensitive hex is accepted.
func TestSECA01_ValidateAddress_MixedCase_Accepted(t *testing.T) {
	bc := newMinimalBC(t)
	// hex.DecodeString accepts both upper and lowercase
	addr := "AABBCCDDEEFF00112233445566778899AABBCCDD" // 40 uppercase hex
	if !bc.ValidateAddress(addr) {
		t.Errorf("uppercase hex address should be valid: %s", addr)
	}
}

// TestSECA01_ValidateAddress_Empty_Rejected verifies empty string fails.
func TestSECA01_ValidateAddress_Empty_Rejected(t *testing.T) {
	bc := newMinimalBC(t)
	if bc.ValidateAddress("") {
		t.Error("empty address should be invalid")
	}
}

// TestSECA01_ValidateAddress_TableDriven runs a comprehensive table of cases.
func TestSECA01_ValidateAddress_TableDriven(t *testing.T) {
	bc := newMinimalBC(t)
	cases := []struct {
		name  string
		addr  string
		valid bool
	}{
		{"40-char zeros", strings.Repeat("0", 40), true},
		{"64-char zeros", strings.Repeat("0", 64), true},
		{"40-char mixed hex", "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0", true},
		{"64-char SHA-256 style", "3ba7ec8ada83c5b82cfa3e5a7c97e4e5a6d5b9f0a1e2c4b8d7f9a3b5c6d8e0f2", true},
		{"39-char", strings.Repeat("a", 39), false},
		{"41-char", strings.Repeat("a", 41), false},
		{"63-char", strings.Repeat("a", 63), false},
		{"65-char", strings.Repeat("a", 65), false},
		{"32-char", strings.Repeat("a", 32), false},
		{"empty", "", false},
		{"non-hex 40-char", strings.Repeat("g", 40), false},
		{"non-hex 64-char", strings.Repeat("x", 64), false},
		{"spaces in 40-char", "a1b2 3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := bc.ValidateAddress(tc.addr)
			if got != tc.valid {
				t.Errorf("ValidateAddress(%q) = %v, want %v", tc.addr, got, tc.valid)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GenesisAllocation.validate() — 64-char address support (62ef33c)
// ---------------------------------------------------------------------------

// TestGenesisAllocation_Validate_64CharAddress verifies 64-char SPHINCS+ addresses pass.
func TestGenesisAllocation_Validate_64CharAddress(t *testing.T) {
	addr64 := "a3f2b1c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2"
	alloc := NewGenesisAllocation(addr64, big.NewInt(1_000_000_000_000_000_000), "TestWallet")
	if err := alloc.validate(); err != nil {
		t.Errorf("64-char SPHINCS+ address should pass allocation validation: %v", err)
	}
}

// TestGenesisAllocation_Validate_40CharAddress verifies legacy 40-char addresses still pass.
func TestGenesisAllocation_Validate_40CharAddress(t *testing.T) {
	addr40 := strings.Repeat("b", 40)
	alloc := NewGenesisAllocation(addr40, big.NewInt(1e18), "Legacy")
	if err := alloc.validate(); err != nil {
		t.Errorf("40-char address should pass allocation validation: %v", err)
	}
}

// TestGenesisAllocation_Validate_WrongLength_Fails verifies non-40/64-char addresses fail.
func TestGenesisAllocation_Validate_WrongLength_Fails(t *testing.T) {
	for _, length := range []int{0, 20, 39, 41, 63, 65, 128} {
		addr := strings.Repeat("a", length)
		alloc := NewGenesisAllocation(addr, big.NewInt(1e18), "Bad")
		if err := alloc.validate(); err == nil {
			t.Errorf("address of length %d should fail validation", length)
		}
	}
}

// TestGenesisAllocation_Validate_NilBalance_Fails verifies nil balance is rejected.
func TestGenesisAllocation_Validate_NilBalance_Fails(t *testing.T) {
	addr := strings.Repeat("a", 40)
	alloc := &GenesisAllocation{Address: addr, BalanceNQTX: nil, Label: "Test"}
	if err := alloc.validate(); err == nil {
		t.Error("nil balance should fail validation")
	}
}

// TestGenesisAllocation_Validate_NegativeBalance_Fails verifies negative balance is rejected.
func TestGenesisAllocation_Validate_NegativeBalance_Fails(t *testing.T) {
	addr := strings.Repeat("a", 40)
	alloc := NewGenesisAllocation(addr, big.NewInt(-1), "Negative")
	if err := alloc.validate(); err == nil {
		t.Error("negative balance should fail validation")
	}
}

// TestGenesisAllocation_Validate_ZeroBalance_Passes verifies zero balance is allowed
// (allocations with zero balance are valid, though uncommon).
func TestGenesisAllocation_Validate_ZeroBalance_Passes(t *testing.T) {
	addr := strings.Repeat("a", 40)
	alloc := NewGenesisAllocation(addr, big.NewInt(0), "ZeroTest")
	if err := alloc.validate(); err != nil {
		t.Errorf("zero balance should pass validation: %v", err)
	}
}

// TestGenesisAllocation_Validate_Nil_Fails verifies nil allocation is handled.
func TestGenesisAllocation_Validate_Nil_Fails(t *testing.T) {
	var alloc *GenesisAllocation
	if err := alloc.validate(); err == nil {
		t.Error("nil GenesisAllocation should fail validation")
	}
}

// TestDefaultGenesisAllocations_AllValidate verifies every default genesis
// allocation passes the validation function with the new dual-format rules.
func TestDefaultGenesisAllocations_AllValidate(t *testing.T) {
	allocs := DefaultGenesisAllocations()
	for _, alloc := range allocs {
		if err := alloc.validate(); err != nil {
			t.Errorf("default allocation %q (%s): validate() failed: %v",
				alloc.Address, alloc.Label, err)
		}
	}
}
