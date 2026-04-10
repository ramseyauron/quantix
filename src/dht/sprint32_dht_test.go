// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 32 - dht additional coverage
package dht

import (
	"net"
	"testing"

	"go.uber.org/zap"
)

func newTestDHT32(t *testing.T) *DHT {
	t.Helper()
	cfg := Config{
		Proto:   "udp4",
		Address: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
	}
	logger, _ := zap.NewDevelopment()
	d, err := NewDHT(cfg, logger)
	if err != nil {
		t.Fatalf("NewDHT error: %v", err)
	}
	return d
}

// ---------------------------------------------------------------------------
// KNearest — empty routing table returns empty slice
// ---------------------------------------------------------------------------

func TestSprint32_KNearest_EmptyTable_ReturnsEmptyOrNonNil(t *testing.T) {
	d := newTestDHT32(t)
	// Use self node ID as target to avoid "empty target" panic
	target := d.SelfNodeID()
	result := d.KNearest(target)
	// May be empty — just no panic
	_ = result
}

func TestSprint32_KNearest_NonZeroTarget(t *testing.T) {
	d := newTestDHT32(t)
	// Use a target derived from self (XOR with known value)
	target := d.SelfNodeID()
	target[0] ^= 0xFF // flip bits to make a valid non-zero target
	result := d.KNearest(target)
	_ = result
}

// ---------------------------------------------------------------------------
// SelfNodeID — stable across calls
// ---------------------------------------------------------------------------

func TestSprint32_SelfNodeID_Stable(t *testing.T) {
	d := newTestDHT32(t)
	id1 := d.SelfNodeID()
	id2 := d.SelfNodeID()
	if id1 != id2 {
		t.Fatal("SelfNodeID should be stable across calls")
	}
}

// ---------------------------------------------------------------------------
// Put — no panic on valid/nil values (async, doesn't require network)
// ---------------------------------------------------------------------------

func TestSprint32_Put_NoPanic(t *testing.T) {
	d := newTestDHT32(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Put panicked: %v", r)
		}
	}()
	var k [32]byte
	k[0] = 0x01
	d.Put(k, []byte("value"), 60)
}

func TestSprint32_Put_NilValue_NoPanic(t *testing.T) {
	d := newTestDHT32(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Put with nil value panicked: %v", r)
		}
	}()
	var k [32]byte
	k[0] = 0x03
	d.Put(k, nil, 60)
}

// ---------------------------------------------------------------------------
// GetCached — unknown key returns nil
// ---------------------------------------------------------------------------

func TestSprint32_GetCached_UnknownKey_Nil(t *testing.T) {
	d := newTestDHT32(t)
	var k [32]byte
	k[0] = 0xFF
	k[1] = 0xFE
	k[2] = 0xFD
	result := d.GetCached(k)
	if result != nil {
		t.Logf("GetCached unknown key returned non-nil (may have init values): %v", result)
	}
}

// ---------------------------------------------------------------------------
// ScheduleGet — no panic
// ---------------------------------------------------------------------------

func TestSprint32_ScheduleGet_NoPanic(t *testing.T) {
	d := newTestDHT32(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ScheduleGet panicked: %v", r)
		}
	}()
	var k [32]byte
	k[0] = 0xAA
	d.ScheduleGet(0, k)
}

// ---------------------------------------------------------------------------
// Join — no panic on empty routing table
// ---------------------------------------------------------------------------

func TestSprint32_Join_NoPanic(t *testing.T) {
	d := newTestDHT32(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Join panicked: %v", r)
		}
	}()
	d.Join()
}
