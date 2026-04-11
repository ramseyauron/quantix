package core

import (
	"math/big"
	"testing"
)

// Sprint 51 — core allocations.go coverage

// ---------------------------------------------------------------------------
// NewGenesisAllocation / NewGenesisAllocationQTX / NewFounderAlloc etc.
// ---------------------------------------------------------------------------

func Sprint51_TestNewGenesisAllocation_Fields(t *testing.T) {
	alloc := NewGenesisAllocation("test-addr", big.NewInt(1000), "test-label")
	if alloc == nil {
		t.Fatal("NewGenesisAllocation returned nil")
	}
	if alloc.Address != "test-addr" {
		t.Errorf("Address: got %q, want %q", alloc.Address, "test-addr")
	}
	if alloc.Label != "test-label" {
		t.Errorf("Label: got %q, want %q", alloc.Label, "test-label")
	}
	if alloc.BalanceNQTX.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("BalanceNQTX: got %s, want 1000", alloc.BalanceNQTX)
	}
}

func Sprint51_TestNewGenesisAllocationQTX_Converts(t *testing.T) {
	// 1 QTX = 1e18 nQTX
	alloc := NewGenesisAllocationQTX("1111111111111111111111111111111111111111", 1, "qtx-label")
	if alloc == nil {
		t.Fatal("NewGenesisAllocationQTX returned nil")
	}
	// Balance should be at least 1 (and typically 1e18)
	if alloc.BalanceNQTX.Sign() <= 0 {
		t.Error("BalanceNQTX should be positive")
	}
}

func Sprint51_TestNewFounderAlloc_NotNil(t *testing.T) {
	alloc := NewFounderAlloc("founder-addr", 1000)
	if alloc == nil {
		t.Fatal("NewFounderAlloc returned nil")
	}
	if alloc.Address != "founder-addr" {
		t.Errorf("Address: got %q", alloc.Address)
	}
}

func TestNewReserveAlloc_NotNil(t *testing.T) {
	alloc := NewReserveAlloc("reserve-addr", 500)
	if alloc == nil {
		t.Fatal("NewReserveAlloc returned nil")
	}
}

func TestNewTreasuryAlloc_NotNil(t *testing.T) {
	alloc := NewTreasuryAlloc("treasury-addr", 200)
	if alloc == nil {
		t.Fatal("NewTreasuryAlloc returned nil")
	}
}

func TestNewCommunityAlloc_NotNil(t *testing.T) {
	alloc := NewCommunityAlloc("community-addr", 100)
	if alloc == nil {
		t.Fatal("NewCommunityAlloc returned nil")
	}
}

func TestNewValidatorAlloc_NotNil(t *testing.T) {
	alloc := NewValidatorAlloc("validator-addr", 50)
	if alloc == nil {
		t.Fatal("NewValidatorAlloc returned nil")
	}
}

// ---------------------------------------------------------------------------
// DefaultGenesisAllocations
// ---------------------------------------------------------------------------

func Sprint51_TestDefaultGenesisAllocations_NotEmpty(t *testing.T) {
	allocs := DefaultGenesisAllocations()
	if len(allocs) == 0 {
		t.Error("DefaultGenesisAllocations should return non-empty slice")
	}
}

func Sprint51_TestDefaultGenesisAllocations_AllHaveAddress(t *testing.T) {
	allocs := DefaultGenesisAllocations()
	for i, a := range allocs {
		if a.Address == "" {
			t.Errorf("alloc[%d] has empty address", i)
		}
	}
}

// ---------------------------------------------------------------------------
// SummariseAllocations
// ---------------------------------------------------------------------------

func Sprint51_TestSummariseAllocations_NonNil(t *testing.T) {
	allocs := []*GenesisAllocation{
		NewGenesisAllocation("addr1", big.NewInt(100), "a"),
		NewGenesisAllocation("addr2", big.NewInt(200), "b"),
	}
	summary := SummariseAllocations(allocs)
	if summary == nil {
		t.Fatal("SummariseAllocations returned nil")
	}
}

func Sprint51_TestSummariseAllocations_Empty(t *testing.T) {
	summary := SummariseAllocations([]*GenesisAllocation{})
	if summary == nil {
		t.Fatal("SummariseAllocations(empty) returned nil")
	}
}

// ---------------------------------------------------------------------------
// LogAllocationSummary — just verify no panic
// ---------------------------------------------------------------------------

func Sprint51_TestLogAllocationSummary_NoPanic(t *testing.T) {
	allocs := DefaultGenesisAllocations()
	LogAllocationSummary(allocs)
}

// ---------------------------------------------------------------------------
// NewAllocationSet + AllocationSet methods
// ---------------------------------------------------------------------------

func Sprint51_TestNewAllocationSet_Valid(t *testing.T) {
	allocs := []*GenesisAllocation{
		NewGenesisAllocation("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", big.NewInt(100), "a"),
		NewGenesisAllocation("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", big.NewInt(200), "b"),
	}
	set, err := NewAllocationSet(allocs)
	if err != nil {
		t.Fatalf("NewAllocationSet: %v", err)
	}
	if set == nil {
		t.Fatal("expected non-nil AllocationSet")
	}
}

func TestAllocationSet_Len(t *testing.T) {
	allocs := []*GenesisAllocation{
		NewGenesisAllocation("1111111111111111111111111111111111111111", big.NewInt(100), "x"),
		NewGenesisAllocation("2222222222222222222222222222222222222222", big.NewInt(200), "y"),
		NewGenesisAllocation("3333333333333333333333333333333333333333", big.NewInt(300), "z"),
	}
	set, err := NewAllocationSet(allocs)
	if err != nil {
		t.Fatalf("NewAllocationSet: %v", err)
	}
	if set.Len() != 3 {
		t.Errorf("Len: got %d, want 3", set.Len())
	}
}

func TestAllocationSet_Contains_True(t *testing.T) {
	allocs := []*GenesisAllocation{
		NewGenesisAllocation("dddddddddddddddddddddddddddddddddddddddd", big.NewInt(100), ""),
	}
	set, _ := NewAllocationSet(allocs)
	if !set.Contains("dddddddddddddddddddddddddddddddddddddddd") {
		t.Error("Contains should return true for known address")
	}
}

func TestAllocationSet_Contains_False(t *testing.T) {
	allocs := []*GenesisAllocation{
		NewGenesisAllocation("5555555555555555555555555555555555555555", big.NewInt(100), ""),
	}
	set, _ := NewAllocationSet(allocs)
	if set.Contains("cccccccccccccccccccccccccccccccccccccccc") {
		t.Error("Contains should return false for unknown address")
	}
}

func TestAllocationSet_TotalSupplyNSPX(t *testing.T) {
	allocs := []*GenesisAllocation{
		NewGenesisAllocation("a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1", big.NewInt(100), ""),
		NewGenesisAllocation("a2a2a2a2a2a2a2a2a2a2a2a2a2a2a2a2a2a2a2a2", big.NewInt(200), ""),
	}
	set, _ := NewAllocationSet(allocs)
	total := set.TotalSupplyNSPX()
	if total == nil || total.Sign() <= 0 {
		t.Error("TotalSupplyNSPX should be positive")
	}
}

func TestAllocationSet_TotalSupplySPX(t *testing.T) {
	allocs := []*GenesisAllocation{
		NewGenesisAllocation("a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1", big.NewInt(1_000_000_000_000_000_000), ""),
	}
	set, _ := NewAllocationSet(allocs)
	total := set.TotalSupplySPX()
	if total == nil || total.Sign() < 0 {
		t.Error("TotalSupplySPX should be non-negative")
	}
}

func Sprint51_TestAllocationSet_All_Length(t *testing.T) {
	allocs := []*GenesisAllocation{
		NewGenesisAllocation("e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1", big.NewInt(100), ""),
		NewGenesisAllocation("e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2", big.NewInt(100), ""),
	}
	set, _ := NewAllocationSet(allocs)
	all := set.All()
	if len(all) != 2 {
		t.Errorf("All: got %d, want 2", len(all))
	}
}
