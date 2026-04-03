// MIT License
// Copyright (c) 2024 quantix

// Q21 — SEC-P02 regression tests: NodeAddress validation in RegisterValidator
// Tests that validateNodeAddress prevents LevelDB key injection
package core

import (
	"testing"
)

// TestSEC_P02_ValidatorRegister_InjectionPatterns tests that malicious
// node addresses are rejected before being used as LevelDB keys.
func TestSEC_P02_ValidatorRegister_InjectionPatterns(t *testing.T) {
	bc := newPhase2BC(t)

	maliciousAddresses := []struct {
		name string
		addr string
	}{
		{"null byte", "node\x00injected"},
		{"LF injection", "node\nkey"},
		{"tab injection", "node\tkey"},
		{"path traversal", "../../../etc/passwd"},
		{"semicolon", "node;DROP TABLE"},
		{"empty string", ""},
		// Note: "val:injected" contains a colon which IS valid per SEC-P02 regex
		// but the test ensures other dangerous patterns are blocked
	}

	for _, tc := range maliciousAddresses {
		t.Run(tc.name, func(t *testing.T) {
			err := bc.RegisterValidator(&ValidatorRegistration{
				PublicKey:   "pubkey-abc",
				StakeAmount: "1000",
				NodeAddress: tc.addr,
			})
			if err == nil {
				t.Errorf("SEC-P02: malicious node_address %q should be rejected", tc.addr)
			}
		})
	}
}

func TestSEC_P02_ValidatorRegister_ValidAddresses_Accepted(t *testing.T) {
	bc := newPhase2BC(t)

	validAddresses := []struct {
		name string
		addr string
	}{
		{"IP:port", "10.0.0.1:32307"},
		{"hostname:port", "node1.quantix.local:8080"},
		{"alphanumeric only", "validator01"},
		{"with hyphen", "val-node-1"},
		{"with underscore", "val_node_1"},
	}

	for _, tc := range validAddresses {
		t.Run(tc.name, func(t *testing.T) {
			err := bc.RegisterValidator(&ValidatorRegistration{
				PublicKey:   "pubkey-" + tc.name,
				StakeAmount: "1000",
				NodeAddress: tc.addr,
			})
			if err != nil {
				t.Errorf("SEC-P02: valid node_address %q should be accepted: %v", tc.addr, err)
			}
		})
	}
}

func TestSEC_P02_NodeAddress_TooLong_Rejected(t *testing.T) {
	bc := newPhase2BC(t)

	// 254 chars (exceeds 253 max)
	longAddr := ""
	for i := 0; i < 254; i++ {
		longAddr += "a"
	}

	err := bc.RegisterValidator(&ValidatorRegistration{
		PublicKey:   "pubkey",
		StakeAmount: "1000",
		NodeAddress: longAddr,
	})
	if err == nil {
		t.Error("SEC-P02: address exceeding 253 chars should be rejected")
	}
}
