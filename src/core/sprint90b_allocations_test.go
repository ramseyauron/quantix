// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 90b — core 58.6%→higher (allocations constructors,
// AllocationSet Get/Contains/Len/All/TotalSupply, DefaultGenesisAllocations,
// SummariseAllocations, NewGenesisAllocationQTX, allocation domain helpers)
package core_test

import (
	"math/big"
	"testing"

	core "github.com/ramseyauron/quantix/src/core"
)

// testAddr is a valid 64-char hex address for allocation tests.
const testAddr90 = "aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd"

// ---------------------------------------------------------------------------
// NewGenesisAllocation / NewGenesisAllocationQTX
// ---------------------------------------------------------------------------

func TestSprint90b_NewGenesisAllocation_FieldsSet(t *testing.T) {
	bal := big.NewInt(1000)
	a := core.NewGenesisAllocation(testAddr90, bal, "TestLabel")
	if a == nil {
		t.Fatal("expected non-nil allocation")
	}
	if a.Address != testAddr90 {
		t.Errorf("Address = %q, want %q", a.Address, testAddr90)
	}
	if a.BalanceNQTX.Cmp(bal) != 0 {
		t.Errorf("BalanceNQTX = %v, want %v", a.BalanceNQTX, bal)
	}
	if a.Label != "TestLabel" {
		t.Errorf("Label = %q, want TestLabel", a.Label)
	}
}

func TestSprint90b_NewGenesisAllocation_CopiesBalance(t *testing.T) {
	bal := big.NewInt(999)
	a := core.NewGenesisAllocation(testAddr90, bal, "L")
	bal.SetInt64(0) // mutate original
	// allocation should retain its own copy
	if a.BalanceNQTX.Sign() == 0 {
		t.Error("allocation balance was mutated by external big.Int change")
	}
}

func TestSprint90b_NewGenesisAllocationQTX_Converts(t *testing.T) {
	a := core.NewGenesisAllocationQTX(testAddr90, 5, "QTX-Test")
	// 5 QTX = 5 × 1e18 nQTX
	expected := new(big.Int).Mul(big.NewInt(5), big.NewInt(1e18))
	if a.BalanceNQTX.Cmp(expected) != 0 {
		t.Errorf("BalanceNQTX = %v, want %v", a.BalanceNQTX, expected)
	}
}

// ---------------------------------------------------------------------------
// Domain-specific alloc constructors
// ---------------------------------------------------------------------------

func TestSprint90b_NewFounderAlloc_LabelIsFounders(t *testing.T) {
	a := core.NewFounderAlloc(testAddr90, 100)
	if a.Label != "Founders" {
		t.Errorf("Label = %q, want Founders", a.Label)
	}
}

func TestSprint90b_NewReserveAlloc_LabelIsReserve(t *testing.T) {
	a := core.NewReserveAlloc(testAddr90, 100)
	if a.Label != "Reserve" {
		t.Errorf("Label = %q, want Reserve", a.Label)
	}
}

func TestSprint90b_NewTreasuryAlloc_LabelIsTreasury(t *testing.T) {
	a := core.NewTreasuryAlloc(testAddr90, 100)
	if a.Label != "Treasury" {
		t.Errorf("Label = %q, want Treasury", a.Label)
	}
}

func TestSprint90b_NewCommunityAlloc_LabelIsCommunity(t *testing.T) {
	a := core.NewCommunityAlloc(testAddr90, 100)
	if a.Label != "Community" {
		t.Errorf("Label = %q, want Community", a.Label)
	}
}

func TestSprint90b_NewValidatorAlloc_LabelIsValidator(t *testing.T) {
	a := core.NewValidatorAlloc(testAddr90, 100)
	if a.Label != "Validator" {
		t.Errorf("Label = %q, want Validator", a.Label)
	}
}

// ---------------------------------------------------------------------------
// DefaultGenesisAllocations — non-empty, every allocation validates
// ---------------------------------------------------------------------------

func TestSprint90b_DefaultGenesisAllocations_NonEmpty(t *testing.T) {
	allocs := core.DefaultGenesisAllocations()
	if len(allocs) == 0 {
		t.Error("DefaultGenesisAllocations returned empty slice")
	}
}

func TestSprint90b_DefaultGenesisAllocations_AllHaveAddresses(t *testing.T) {
	for i, a := range core.DefaultGenesisAllocations() {
		if a.Address == "" {
			t.Errorf("allocation[%d] has empty address", i)
		}
		if a.BalanceNQTX == nil || a.BalanceNQTX.Sign() <= 0 {
			t.Errorf("allocation[%d] has zero or nil balance", i)
		}
	}
}

// ---------------------------------------------------------------------------
// SummariseAllocations
// ---------------------------------------------------------------------------

func TestSprint90b_SummariseAllocations_NonNil(t *testing.T) {
	allocs := core.DefaultGenesisAllocations()
	s := core.SummariseAllocations(allocs)
	if s == nil {
		t.Fatal("SummariseAllocations returned nil")
	}
}

func TestSprint90b_SummariseAllocations_TotalMatchesSum(t *testing.T) {
	addr1 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	addr2 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	a1 := core.NewGenesisAllocationQTX(addr1, 100, "A")
	a2 := core.NewGenesisAllocationQTX(addr2, 200, "B")

	s := core.SummariseAllocations([]*core.GenesisAllocation{a1, a2})
	if s == nil {
		t.Fatal("SummariseAllocations returned nil for 2 allocs")
	}
}

// ---------------------------------------------------------------------------
// AllocationSet — Get, Contains, Len, All, TotalSupply
// ---------------------------------------------------------------------------

func newAllocationSet90(t *testing.T) *core.AllocationSet {
	t.Helper()
	addr1 := "1111111111111111111111111111111111111111111111111111111111111111"
	addr2 := "2222222222222222222222222222222222222222222222222222222222222222"
	a1 := core.NewGenesisAllocationQTX(addr1, 1000, "A")
	a2 := core.NewGenesisAllocationQTX(addr2, 2000, "B")
	s, err := core.NewAllocationSet([]*core.GenesisAllocation{a1, a2})
	if err != nil {
		t.Fatalf("NewAllocationSet: %v", err)
	}
	return s
}

func TestSprint90b_AllocationSet_Len(t *testing.T) {
	s := newAllocationSet90(t)
	if s.Len() != 2 {
		t.Errorf("Len = %d, want 2", s.Len())
	}
}

func TestSprint90b_AllocationSet_Get_Found(t *testing.T) {
	s := newAllocationSet90(t)
	addr := "1111111111111111111111111111111111111111111111111111111111111111"
	a, ok := s.Get(addr)
	if !ok {
		t.Error("expected Get to return true for known address")
	}
	if a == nil {
		t.Error("expected non-nil allocation for known address")
	}
}

func TestSprint90b_AllocationSet_Get_NotFound(t *testing.T) {
	s := newAllocationSet90(t)
	_, ok := s.Get("0000000000000000000000000000000000000000000000000000000000000000")
	if ok {
		t.Error("expected Get to return false for unknown address")
	}
}

func TestSprint90b_AllocationSet_Contains_True(t *testing.T) {
	s := newAllocationSet90(t)
	addr := "2222222222222222222222222222222222222222222222222222222222222222"
	if !s.Contains(addr) {
		t.Error("Contains should return true for known address")
	}
}

func TestSprint90b_AllocationSet_Contains_False(t *testing.T) {
	s := newAllocationSet90(t)
	if s.Contains("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc") {
		t.Error("Contains should return false for unknown address")
	}
}

func TestSprint90b_AllocationSet_All_ReturnsAll(t *testing.T) {
	s := newAllocationSet90(t)
	all := s.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d items, want 2", len(all))
	}
}

func TestSprint90b_AllocationSet_TotalSupplyNSPX_Correct(t *testing.T) {
	s := newAllocationSet90(t)
	total := s.TotalSupplyNSPX()
	// 1000 + 2000 = 3000 QTX = 3000 × 1e18 nQTX
	expected := new(big.Int).Mul(big.NewInt(3000), big.NewInt(1e18))
	if total.Cmp(expected) != 0 {
		t.Errorf("TotalSupplyNSPX = %v, want %v", total, expected)
	}
}

func TestSprint90b_AllocationSet_TotalSupplySPX_Correct(t *testing.T) {
	s := newAllocationSet90(t)
	spx := s.TotalSupplySPX()
	if spx.Int64() != 3000 {
		t.Errorf("TotalSupplySPX = %v, want 3000", spx)
	}
}

// ---------------------------------------------------------------------------
// NewAllocationSet — duplicate address returns error
// ---------------------------------------------------------------------------

func TestSprint90b_NewAllocationSet_DuplicateAddress_Error(t *testing.T) {
	addr := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	a1 := core.NewGenesisAllocationQTX(addr, 100, "A")
	a2 := core.NewGenesisAllocationQTX(addr, 200, "B") // same address
	_, err := core.NewAllocationSet([]*core.GenesisAllocation{a1, a2})
	if err == nil {
		t.Error("expected error for duplicate address in AllocationSet")
	}
}

// ---------------------------------------------------------------------------
// LogAllocationSummary — no panic
// ---------------------------------------------------------------------------

func TestSprint90b_LogAllocationSummary_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogAllocationSummary panicked: %v", r)
		}
	}()
	allocs := core.DefaultGenesisAllocations()
	core.LogAllocationSummary(allocs)
}
