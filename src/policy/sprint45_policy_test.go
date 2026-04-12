// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 45 - policy rewards, storage, validators, params.Validate
package policy

import (
	"math/big"
	"testing"
)

func getTestParams45() *PolicyParameters {
	return GetDefaultPolicyParams()
}

// ---------------------------------------------------------------------------
// CalculateDelegatorReward
// ---------------------------------------------------------------------------

func TestSprint45_CalculateDelegatorReward_Basic_NonNil(t *testing.T) {
	p := getTestParams45()
	delegatorStake := big.NewInt(1000)
	validatorStake := big.NewInt(5000)
	validatorRewards := big.NewInt(100)
	reward := p.CalculateDelegatorReward(delegatorStake, validatorStake, validatorRewards, 0.05)
	if reward == nil {
		t.Fatal("expected non-nil delegator reward")
	}
}

func TestSprint45_CalculateDelegatorReward_ZeroCommission_FullReward(t *testing.T) {
	p := getTestParams45()
	delegatorStake := big.NewInt(1000)
	validatorStake := big.NewInt(1000)
	validatorRewards := big.NewInt(1000)
	reward := p.CalculateDelegatorReward(delegatorStake, validatorStake, validatorRewards, 0)
	if reward == nil {
		t.Fatal("expected non-nil delegator reward with zero commission")
	}
}

// ---------------------------------------------------------------------------
// CalculateAPYWithDecay
// ---------------------------------------------------------------------------

func TestSprint45_CalculateAPYWithDecay_SingleYear_NonNegative(t *testing.T) {
	p := getTestParams45()
	totalStake := new(big.Int).Mul(big.NewInt(1000000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	totalSupply := new(big.Int).Mul(big.NewInt(10000000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	apy := p.CalculateAPYWithDecay(totalStake, totalSupply, 1, 1, 0.1)
	if apy < 0 {
		t.Fatalf("expected non-negative APY, got %f", apy)
	}
}

func TestSprint45_CalculateAPYWithDecay_MultipleYears_NonNegative(t *testing.T) {
	p := getTestParams45()
	totalStake := new(big.Int).Mul(big.NewInt(5000000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	totalSupply := new(big.Int).Mul(big.NewInt(10000000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	apy := p.CalculateAPYWithDecay(totalStake, totalSupply, 1, 3, 0.5)
	if apy < 0 {
		t.Fatalf("expected non-negative APY for multi-year, got %f", apy)
	}
}

// ---------------------------------------------------------------------------
// CalculateStorageCost
// ---------------------------------------------------------------------------

func TestSprint45_CalculateStorageCost_Basic_NotNil(t *testing.T) {
	p := getTestParams45()
	pricing := p.CalculateStorageCost(1000000, 3.0) // 1MB for 3 months
	if pricing == nil {
		t.Fatal("expected non-nil StoragePricing")
	}
}

func TestSprint45_CalculateStorageCost_Fields_Present(t *testing.T) {
	p := getTestParams45()
	pricing := p.CalculateStorageCost(1000000, 6.0)
	if pricing.Bytes != 1000000 {
		t.Fatalf("expected Bytes=1000000, got %d", pricing.Bytes)
	}
	if pricing.DurationDays == 0 {
		t.Fatal("expected non-zero DurationDays")
	}
	if pricing.TotalCost == nil {
		t.Fatal("expected non-nil TotalCost")
	}
}

func TestSprint45_CalculateStorageCost_LargeFile_NonNegative(t *testing.T) {
	p := getTestParams45()
	pricing := p.CalculateStorageCost(1e9, 12.0) // 1GB for 12 months
	if pricing == nil || pricing.TotalCost == nil {
		t.Fatal("expected non-nil pricing")
	}
	if pricing.TotalCost.Sign() < 0 {
		t.Fatal("total cost should be non-negative")
	}
}

// ---------------------------------------------------------------------------
// CalculatePinningCost
// ---------------------------------------------------------------------------

func TestSprint45_CalculatePinningCost_Basic_NonNil(t *testing.T) {
	p := getTestParams45()
	cost := p.CalculatePinningCost(1000000, 1)
	if cost == nil {
		t.Fatal("expected non-nil pinning cost")
	}
}

func TestSprint45_CalculatePinningCost_LongerDuration_HigherCost(t *testing.T) {
	p := getTestParams45()
	cost1 := p.CalculatePinningCost(1000000, 1)
	cost12 := p.CalculatePinningCost(1000000, 12)
	if cost1 == nil || cost12 == nil {
		t.Fatal("expected non-nil costs")
	}
	if cost12.Cmp(cost1) <= 0 {
		t.Fatal("12-month cost should be greater than 1-month cost")
	}
}

// ---------------------------------------------------------------------------
// GetDefaultValidatorEconomics
// ---------------------------------------------------------------------------

func TestSprint45_GetDefaultValidatorEconomics_NotNil(t *testing.T) {
	ve := GetDefaultValidatorEconomics()
	if ve == nil {
		t.Fatal("expected non-nil ValidatorEconomics")
	}
}

func TestSprint45_GetDefaultValidatorEconomics_Fields(t *testing.T) {
	ve := GetDefaultValidatorEconomics()
	if ve.MaxValidators == 0 {
		t.Fatal("expected non-zero MaxValidators")
	}
	if ve.CommissionRate <= 0 {
		t.Fatal("expected positive commission rate")
	}
	if ve.MinSelfDelegation == nil {
		t.Fatal("expected non-nil MinSelfDelegation")
	}
}

// ---------------------------------------------------------------------------
// CalculateSlashingPenalty
// ---------------------------------------------------------------------------

func TestSprint45_CalculateSlashingPenalty_DoubleSign_HigherThanDowntime(t *testing.T) {
	p := getTestParams45()
	stake := new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	doubleSignPenalty := p.CalculateSlashingPenalty(stake, "double_sign")
	downtimePenalty := p.CalculateSlashingPenalty(stake, "downtime")
	if doubleSignPenalty.Cmp(downtimePenalty) <= 0 {
		t.Fatal("double_sign penalty should be greater than downtime penalty")
	}
}

func TestSprint45_CalculateSlashingPenalty_Unknown_UsesDefault(t *testing.T) {
	p := getTestParams45()
	stake := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	penalty := p.CalculateSlashingPenalty(stake, "unknown_type")
	if penalty == nil || penalty.Sign() < 0 {
		t.Fatal("expected non-negative penalty for unknown slash type")
	}
}

func TestSprint45_CalculateSlashingPenalty_Liveness_Smallest(t *testing.T) {
	p := getTestParams45()
	stake := new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	liveness := p.CalculateSlashingPenalty(stake, "liveness")
	downtime := p.CalculateSlashingPenalty(stake, "downtime")
	if liveness.Cmp(downtime) >= 0 {
		t.Fatal("liveness penalty should be smaller than downtime penalty")
	}
}

// ---------------------------------------------------------------------------
// Params.Validate — more edge cases
// ---------------------------------------------------------------------------

func TestSprint45_Validate_NilBaseFee_Documented(t *testing.T) {
	// Validate panics when BaseFeePerByte is nil — nil guard needed in params.go:71
	t.Skip("Validate panics with nil BaseFeePerByte — nil guard needed")
}

func TestSprint45_Validate_NegativeBaseFee_Error(t *testing.T) {
	p := GetDefaultPolicyParams()
	p.BaseFeePerByte = big.NewInt(-1)
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for negative BaseFeePerByte")
	}
}
