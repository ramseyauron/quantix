// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 64 — stark/zk 30.8%→higher
// Tests: MarshalJSON valid DomainParameters, UnmarshalJSON roundtrip,
// verifySignature valid params + nil sig, commitToSignatures empty slice
package zk

import (
	"math/big"
	"testing"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/poly"
)

// makeDomainParams creates a minimal valid DomainParameters for testing
func makeDomainParams() *DomainParameters {
	g := PrimeField.NewFieldElementFromInt64(5)
	h := PrimeField.NewFieldElementFromInt64(3)

	trace := []ff.FieldElement{
		PrimeField.NewFieldElementFromInt64(1),
		PrimeField.NewFieldElementFromInt64(0),
	}

	subG := []ff.FieldElement{g}
	subH := []ff.FieldElement{h}
	evalDomain := []ff.FieldElement{PrimeField.NewFieldElementFromInt64(2)}

	coeffs := []ff.FieldElement{
		PrimeField.NewFieldElementFromInt64(1),
		PrimeField.NewFieldElementFromInt64(2),
	}
	polynomial := poly.NewPolynomial(coeffs)

	return &DomainParameters{
		Trace:                 trace,
		GeneratorG:            g,
		SubgroupG:             subG,
		GeneratorH:            h,
		SubgroupH:             subH,
		EvaluationDomain:      evalDomain,
		Polynomial:            polynomial,
		PolynomialEvaluations: []*big.Int{big.NewInt(5)},
		EvaluationRoot:        []byte{0xde, 0xad, 0xbe, 0xef},
	}
}

// ─── MarshalJSON — valid DomainParameters ─────────────────────────────────────

func TestSprint64_MarshalJSON_ValidParams(t *testing.T) {
	dp := makeDomainParams()
	data, err := dp.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty marshaled data")
	}
}

func TestSprint64_MarshalJSON_NonEmptyOutput(t *testing.T) {
	dp := makeDomainParams()
	data, err := dp.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	// Should contain field elements as strings
	if len(data) < 10 {
		t.Errorf("marshaled data too short: %d bytes", len(data))
	}
}

// ─── UnmarshalJSON — roundtrip via MarshalJSON → UnmarshalJSON ────────────────

func TestSprint64_UnmarshalJSON_Roundtrip(t *testing.T) {
	dp := makeDomainParams()
	marshaled, err := dp.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var dp2 DomainParameters
	if err := dp2.UnmarshalJSON(marshaled); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	// Verify trace length preserved
	if len(dp2.Trace) != len(dp.Trace) {
		t.Errorf("trace length mismatch: got %d, want %d", len(dp2.Trace), len(dp.Trace))
	}

	// Verify evaluation root preserved
	if len(dp2.EvaluationRoot) == 0 {
		t.Error("expected non-empty EvaluationRoot after unmarshal")
	}
}

func TestSprint64_UnmarshalJSON_SubgroupPreserved(t *testing.T) {
	dp := makeDomainParams()
	marshaled, err := dp.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var dp2 DomainParameters
	if err := dp2.UnmarshalJSON(marshaled); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if len(dp2.SubgroupG) != len(dp.SubgroupG) {
		t.Errorf("SubgroupG length mismatch: got %d, want %d", len(dp2.SubgroupG), len(dp.SubgroupG))
	}
}

// ─── verifySignature — valid manager, nil sig returns false ──────────────────

func TestSprint64_VerifySignature_NilSig_ReturnsFalse(t *testing.T) {
	sm, err := NewSignManager()
	if err != nil {
		t.Fatalf("NewSignManager: %v", err)
	}
	result := sm.verifySignature(Signature{
		Message:   []byte("msg"),
		Signature: nil,
		PublicKey: nil,
	})
	if result {
		t.Error("expected false for nil signature")
	}
}

// ─── commitToSignatures — valid but nil-field sigs return error ───────────────

func TestSprint64_CommitToSignatures_NilFieldSig_Error(t *testing.T) {
	sm, err := NewSignManager()
	if err != nil {
		t.Fatalf("NewSignManager: %v", err)
	}
	sigs := []Signature{
		{Message: []byte("msg"), Signature: nil, PublicKey: nil},
	}
	_, err2 := sm.commitToSignatures(sigs)
	if err2 == nil {
		t.Error("expected error for nil-field signature in commitToSignatures")
	}
}

// ─── GenElems — determinism ────────────────────────────────────────────────────

func TestSprint64_GenElems_Deterministic(t *testing.T) {
	r1 := GenElems(PrimeFieldGen, 8)
	r2 := GenElems(PrimeFieldGen, 8)
	if len(r1) != len(r2) {
		t.Fatal("GenElems results have different lengths")
	}
	for i := range r1 {
		if r1[i].Big().Cmp(r2[i].Big()) != 0 {
			t.Errorf("GenElems element %d differs", i)
		}
	}
}

func TestSprint64_GenElems_FirstIsOne(t *testing.T) {
	// g^0 = 1 for any generator
	elems := GenElems(PrimeFieldGen, 4)
	if len(elems) < 1 {
		t.Fatal("expected at least 1 element")
	}
	// g^0 should equal 1
	one := PrimeField.NewFieldElementFromInt64(1)
	if elems[0].Big().Cmp(one.Big()) != 0 {
		t.Errorf("GenElems[0] should be 1 (g^0), got %v", elems[0].Big())
	}
}

// ─── generatePoints — mismatch panics, verify docs ───────────────────────────

func TestSprint64_GeneratePoints_LengthMismatch_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for mismatched x and y lengths")
		}
	}()
	x := []ff.FieldElement{PrimeField.NewFieldElementFromInt64(1)}
	y := []ff.FieldElement{
		PrimeField.NewFieldElementFromInt64(1),
		PrimeField.NewFieldElementFromInt64(2),
	}
	generatePoints(x, y) // should panic
}
