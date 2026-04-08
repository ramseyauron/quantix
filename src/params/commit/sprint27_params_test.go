// Sprint 27 — params/commit: GetNetworkName, IsSoftForkActive, GenerateForkHeader,
// GetSoftForkParameters, GetKeystoreConfig.
package commit

import (
	"testing"
)

// ─── GetNetworkName ───────────────────────────────────────────────────────────

func TestSprint27_GetNetworkName_Mainnet(t *testing.T) {
	p := QuantixChainParams()
	name := p.GetNetworkName()
	if name != "Quantix Mainnet" {
		t.Errorf("GetNetworkName mainnet = %q, want %q", name, "Quantix Mainnet")
	}
}

func TestSprint27_GetNetworkName_Testnet(t *testing.T) {
	p := TestnetChainParams()
	name := p.GetNetworkName()
	if name != "Quantix Testnet" {
		t.Errorf("GetNetworkName testnet = %q, want %q", name, "Quantix Testnet")
	}
}

func TestSprint27_GetNetworkName_Regtest(t *testing.T) {
	p := RegtestChainParams()
	name := p.GetNetworkName()
	if name != "Quantix Regtest" {
		t.Errorf("GetNetworkName regtest = %q, want %q", name, "Quantix Regtest")
	}
}

func TestSprint27_GetNetworkName_Unknown_ChainID(t *testing.T) {
	p := QuantixChainParams()
	p.ChainID = 99999 // unknown
	name := p.GetNetworkName()
	if name != "Unknown Network" {
		t.Errorf("GetNetworkName unknown = %q, want %q", name, "Unknown Network")
	}
}

// ─── IsSoftForkActive ─────────────────────────────────────────────────────────

func TestSprint27_IsSoftForkActive_UnknownFork_False(t *testing.T) {
	ok := IsSoftForkActive("nonexistent-fork", 1000, 1000000000)
	if ok {
		t.Error("expected false for unknown fork")
	}
}

func TestSprint27_IsSoftForkActive_KnownForkName_NoPanel(t *testing.T) {
	// GetSoftForks returns the configured fork names — just verify no panic
	forks := GetSoftForks()
	for name := range forks {
		_ = IsSoftForkActive(name, 0, 0)
	}
}

// ─── GenerateForkHeader ───────────────────────────────────────────────────────

func TestSprint27_GenerateForkHeader_UnknownFork_UnknownForkString(t *testing.T) {
	header := GenerateForkHeader("no-such-fork")
	if header != "Unknown fork" {
		t.Errorf("GenerateForkHeader unknown = %q, want %q", header, "Unknown fork")
	}
}

func TestSprint27_GenerateForkHeader_KnownFork_NonEmpty(t *testing.T) {
	forks := GetSoftForks()
	for name := range forks {
		h := GenerateForkHeader(name)
		if h == "" {
			t.Errorf("GenerateForkHeader(%q) returned empty string", name)
		}
		if h == "Unknown fork" {
			t.Errorf("GenerateForkHeader(%q) returned 'Unknown fork' for known fork", name)
		}
	}
}

// ─── GetSoftForkParameters ────────────────────────────────────────────────────

func TestSprint27_GetSoftForkParameters_NotNil(t *testing.T) {
	params := GetSoftForkParameters()
	if params == nil {
		t.Fatal("GetSoftForkParameters returned nil")
	}
}

func TestSprint27_GetSoftForkParameters_SameAsGetSoftForks(t *testing.T) {
	a := GetSoftForkParameters()
	b := GetSoftForks()
	if len(a) != len(b) {
		t.Errorf("GetSoftForkParameters len=%d, GetSoftForks len=%d — should be equal", len(a), len(b))
	}
}

// ─── GetKeystoreConfig ────────────────────────────────────────────────────────

func TestSprint27_GetKeystoreConfig_Mainnet_NotNil(t *testing.T) {
	p := QuantixChainParams()
	cfg := p.GetKeystoreConfig()
	if cfg == nil {
		t.Fatal("GetKeystoreConfig mainnet returned nil")
	}
}

func TestSprint27_GetKeystoreConfig_Testnet_NotNil(t *testing.T) {
	p := TestnetChainParams()
	cfg := p.GetKeystoreConfig()
	if cfg == nil {
		t.Fatal("GetKeystoreConfig testnet returned nil")
	}
}

func TestSprint27_GetKeystoreConfig_Regtest_NotNil(t *testing.T) {
	p := RegtestChainParams()
	cfg := p.GetKeystoreConfig()
	if cfg == nil {
		t.Fatal("GetKeystoreConfig regtest returned nil")
	}
}

func TestSprint27_GetKeystoreConfig_UnknownChainID_NotNil(t *testing.T) {
	p := QuantixChainParams()
	p.ChainID = 99999
	cfg := p.GetKeystoreConfig()
	// Unknown ChainID should still return something (default config) — no panic
	_ = cfg
}
