// Sprint 26 — RANDAO ValidateState/ValidateVDFParams/GetVDFParams/GetSeed
package consensus

import (
	"math/big"
	"testing"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// newTestRANDAO26 creates a minimal RANDAO for testing.
func newTestRANDAO26() *RANDAO {
	params := VDFParams{
		Discriminant: big.NewInt(-23), // minimal valid discriminant
		T:            100,
	}
	seed := [32]byte{0x01}
	return NewRANDAO(seed, params, "test-validator")
}

// ─── RANDAO.ValidateState ─────────────────────────────────────────────────────

func TestSprint26_RANDAO_ValidateState_FreshRANDAO_Nil(t *testing.T) {
	r := newTestRANDAO26()
	if err := r.ValidateState(); err != nil {
		t.Errorf("ValidateState on fresh RANDAO = %v, want nil", err)
	}
}

// ─── RANDAO.ValidateVDFParams ─────────────────────────────────────────────────

func TestSprint26_RANDAO_ValidateVDFParams_MatchingParams_Nil(t *testing.T) {
	r := newTestRANDAO26()
	same := VDFParams{
		Discriminant: big.NewInt(-23),
		T:            100,
	}
	if err := r.ValidateVDFParams(same); err != nil {
		t.Errorf("ValidateVDFParams matching = %v, want nil", err)
	}
}

func TestSprint26_RANDAO_ValidateVDFParams_WrongDiscriminant_Error(t *testing.T) {
	r := newTestRANDAO26()
	wrong := VDFParams{
		Discriminant: big.NewInt(-47),
		T:            100,
	}
	if err := r.ValidateVDFParams(wrong); err == nil {
		t.Error("expected error for mismatched discriminant")
	}
}

func TestSprint26_RANDAO_ValidateVDFParams_WrongT_Error(t *testing.T) {
	r := newTestRANDAO26()
	wrong := VDFParams{
		Discriminant: big.NewInt(-23),
		T:            999,
	}
	if err := r.ValidateVDFParams(wrong); err == nil {
		t.Error("expected error for mismatched T")
	}
}

// ─── RANDAO.GetVDFParams ──────────────────────────────────────────────────────

func TestSprint26_RANDAO_GetVDFParams_ReturnsParams(t *testing.T) {
	r := newTestRANDAO26()
	p := r.GetVDFParams()
	if p.T != 100 {
		t.Errorf("GetVDFParams T = %d, want 100", p.T)
	}
	if p.Discriminant.Cmp(big.NewInt(-23)) != 0 {
		t.Errorf("GetVDFParams Discriminant = %s, want -23", p.Discriminant)
	}
}

// ─── RANDAO.GetSeed ───────────────────────────────────────────────────────────

func TestSprint26_RANDAO_GetSeed_ReturnsNonZero(t *testing.T) {
	seed := [32]byte{0xAB, 0xCD}
	params := VDFParams{Discriminant: big.NewInt(-23), T: 10}
	r := NewRANDAO(seed, params, "v1")
	got := r.GetSeed(0)
	allZero := true
	for _, b := range got {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("GetSeed returned all-zero seed")
	}
}

func TestSprint26_RANDAO_GetSeed_DifferentSeeds(t *testing.T) {
	params := VDFParams{Discriminant: big.NewInt(-23), T: 10}
	seed1 := [32]byte{0x01}
	seed2 := [32]byte{0x02}
	r1 := NewRANDAO(seed1, params, "v")
	r2 := NewRANDAO(seed2, params, "v")
	got1 := r1.GetSeed(0)
	got2 := r2.GetSeed(0)
	if got1 == got2 {
		t.Error("different genesis seeds should produce different GetSeed results")
	}
}
