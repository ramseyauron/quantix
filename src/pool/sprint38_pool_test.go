// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 38 - pool GetMemoryUsage + calculateAverageTxSize
package pool

import (
	"testing"
)

// ---------------------------------------------------------------------------
// GetMemoryUsage
// ---------------------------------------------------------------------------

func TestSprint38_GetMemoryUsage_NotNil(t *testing.T) {
	mp := NewMempool(nil)
	usage := mp.GetMemoryUsage()
	if usage == nil {
		t.Fatal("expected non-nil memory usage map")
	}
}

func TestSprint38_GetMemoryUsage_HasFields(t *testing.T) {
	mp := NewMempool(nil)
	usage := mp.GetMemoryUsage()
	fields := []string{"current_bytes", "max_bytes", "available_bytes", "utilization_percent", "average_tx_size"}
	for _, f := range fields {
		if _, ok := usage[f]; !ok {
			t.Fatalf("expected field %q in memory usage", f)
		}
	}
}

func TestSprint38_GetMemoryUsage_EmptyPool_UtilizationZero(t *testing.T) {
	mp := NewMempool(nil)
	usage := mp.GetMemoryUsage()
	utilization, ok := usage["utilization_percent"].(float64)
	if !ok {
		t.Fatal("expected float64 utilization_percent")
	}
	if utilization != 0.0 {
		t.Fatalf("expected 0.0 utilization for empty pool, got %f", utilization)
	}
}

// ---------------------------------------------------------------------------
// calculateAverageTxSize (via GetMemoryUsage)
// ---------------------------------------------------------------------------

func TestSprint38_CalculateAverageTxSize_EmptyPool_Returns256(t *testing.T) {
	mp := NewMempool(nil)
	avg := mp.calculateAverageTxSize()
	if avg != 256 {
		t.Fatalf("expected 256 for empty pool, got %d", avg)
	}
}

// ---------------------------------------------------------------------------
// cleanupExpiredTransactions (via Clear + pool operations)
// ---------------------------------------------------------------------------

func TestSprint38_CleanupExpiredTransactions_EmptyPool_NoPanic(t *testing.T) {
	mp := NewMempool(nil)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("cleanupExpiredTransactions panicked: %v", r)
		}
	}()
	mp.cleanupExpiredTransactions()
}
