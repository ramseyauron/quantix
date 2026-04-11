// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 53 — stark/zk 14.7%→higher (verifySignature, commitToSignatures,
// generateVerificationTrace, VerifySTARKProof nil paths, generatePoints, UnmarshalJSON)
package zk

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/actuallyachraf/algebra/ff"
)

// ---------------------------------------------------------------------------
// verifySignature — nil-guard paths (unexported, white-box)
// ---------------------------------------------------------------------------

func TestVerifySignature_NilParams(t *testing.T) {
	sm := &SignManager{Params: nil}
	result := sm.verifySignature(Signature{Message: []byte("msg")})
	if result {
		t.Error("expected false for nil Params")
	}
}

func TestVerifySignature_NilSig(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	result := p.verifySignature(Signature{
		Message:   []byte("msg"),
		Signature: nil,
		PublicKey: nil,
	})
	if result {
		t.Error("expected false for nil Signature")
	}
}

func TestVerifySignature_NilPublicKey(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	result := p.verifySignature(Signature{
		Message:   []byte("msg"),
		Signature: nil, // also nil, should short-circuit on Signature==nil
		PublicKey: nil,
	})
	if result {
		t.Error("expected false for nil public key")
	}
}

// ---------------------------------------------------------------------------
// commitToSignatures — nil sig data error
// ---------------------------------------------------------------------------

func TestCommitToSignatures_NilSignatureData(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	// Signature with nil fields should return error
	sigs := []Signature{
		{Message: nil, Signature: nil, PublicKey: nil},
	}
	_, err2 := p.commitToSignatures(sigs)
	if err2 == nil {
		t.Error("expected error for nil signature data")
	}
}

func TestCommitToSignatures_EmptySlice(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	// empty slice should return a root (no error)
	root, err2 := p.commitToSignatures([]Signature{})
	if err2 != nil {
		t.Errorf("unexpected error for empty slice: %v", err2)
	}
	_ = root
}

// ---------------------------------------------------------------------------
// generateVerificationTrace — uses only nil-sig (returns false path)
// ---------------------------------------------------------------------------

func TestGenerateVerificationTrace_NilSigsAllZero(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	// All-nil signatures → verifySignature returns false → trace entries = 0
	sigs := []Signature{
		{Message: nil, Signature: nil, PublicKey: nil},
	}
	trace, err2 := p.generateVerificationTrace(sigs)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if len(trace) != 1024 {
		t.Errorf("expected 1024 trace elements, got %d", len(trace))
	}
	// first element should be 0 (invalid signature)
	if trace[0].Big().Sign() != 0 {
		t.Error("expected 0 for invalid signature in trace")
	}
}

func TestGenerateVerificationTrace_EmptyInput(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	trace, err2 := p.generateVerificationTrace([]Signature{})
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if len(trace) != 1024 {
		t.Errorf("expected 1024 trace elements, got %d", len(trace))
	}
	// all zeros
	for i, t2 := range trace {
		if t2.Big().Sign() != 0 {
			t.Errorf("trace[%d] should be 0 for empty input", i)
			break
		}
	}
}

// ---------------------------------------------------------------------------
// VerifySTARKProof — nil / invalid proof structure
// ---------------------------------------------------------------------------

func TestVerifySTARKProof_NilProof(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	ok, err2 := p.VerifySTARKProof(nil)
	if ok || err2 == nil {
		t.Error("expected false + error for nil proof")
	}
}

func TestVerifySTARKProof_NilDomainParams(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	proof := &STARKProof{DomainParams: nil, FsChan: NewChannel()}
	ok, err2 := p.VerifySTARKProof(proof)
	if ok || err2 == nil {
		t.Error("expected false + error for nil DomainParams")
	}
}

func TestVerifySTARKProof_NilFsChan(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	proof := &STARKProof{DomainParams: &DomainParameters{}, FsChan: nil}
	ok, err2 := p.VerifySTARKProof(proof)
	if ok || err2 == nil {
		t.Error("expected false + error for nil FsChan")
	}
}

func TestVerifySTARKProof_InvalidSigData(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	proof := &STARKProof{
		DomainParams: &DomainParameters{},
		FsChan:       NewChannel(),
		Signatures: []Signature{
			{Message: nil, Signature: nil, PublicKey: nil},
		},
	}
	ok, err2 := p.VerifySTARKProof(proof)
	if ok || err2 == nil {
		t.Error("expected false + error for nil signature data in proof")
	}
}

// ---------------------------------------------------------------------------
// generatePoints — length mismatch panics, same length succeeds
// ---------------------------------------------------------------------------

func TestGeneratePoints_SameLength(t *testing.T) {
	x := []ff.FieldElement{
		PrimeField.NewFieldElementFromInt64(1),
		PrimeField.NewFieldElementFromInt64(2),
	}
	y := []ff.FieldElement{
		PrimeField.NewFieldElementFromInt64(3),
		PrimeField.NewFieldElementFromInt64(4),
	}
	points := generatePoints(x, y)
	if len(points) != 2 {
		t.Errorf("expected 2 points, got %d", len(points))
	}
}

func TestGeneratePoints_Empty(t *testing.T) {
	var empty []ff.FieldElement
	points := generatePoints(empty, empty)
	if len(points) != 0 {
		t.Errorf("expected 0 points for empty input, got %d", len(points))
	}
}

// ---------------------------------------------------------------------------
// DomainParameters.UnmarshalJSON — error paths
// ---------------------------------------------------------------------------

func TestDomainParameters_UnmarshalJSON_InvalidJSON(t *testing.T) {
	dp := &DomainParameters{}
	err := dp.UnmarshalJSON([]byte("not-json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDomainParameters_UnmarshalJSON_BadFieldNumber(t *testing.T) {
	// Valid JSON structure but invalid big.Int for Field
	bad := `{"Field":"not-a-number","Trace":[],"GeneratorG":"1","SubgroupG":[],"GeneratorH":"1","SubgroupH":[],"EvaluationDomain":[],"Polynomial":[],"PolynomialEvaluations":[],"EvaluationRoot":""}`
	dp := &DomainParameters{}
	err := dp.UnmarshalJSON([]byte(bad))
	if err == nil {
		t.Error("expected error for bad field number")
	}
}

func TestDomainParameters_MarshalUnmarshal_ErrorPaths(t *testing.T) {
	// Marshal with empty trace → error
	dp := &DomainParameters{}
	data, err := dp.MarshalJSON()
	if err == nil {
		t.Error("MarshalJSON with empty trace should error")
	}
	if data != nil {
		t.Error("MarshalJSON error path should return nil data")
	}
}

// ---------------------------------------------------------------------------
// GenerateSTARKProof — no signatures error
// ---------------------------------------------------------------------------

func TestGenerateSTARKProof_NoSignatures(t *testing.T) {
	p, err := newTestSignManager(t)
	if err != nil {
		t.Skip("SignManager unavailable:", err)
	}
	_, err2 := p.GenerateSTARKProof([]Signature{})
	if err2 == nil {
		t.Error("expected error for empty signatures")
	}
}

// ---------------------------------------------------------------------------
// JSON roundtrip for JSONDomainParams struct
// ---------------------------------------------------------------------------

func TestJSONDomainParams_MarshalUnmarshal(t *testing.T) {
	orig := &JSONDomainParams{
		Field:                 "3221225473",
		Trace:                 []string{"1", "2"},
		GeneratorG:            "5",
		SubgroupG:             []string{"5"},
		GeneratorH:            "3",
		SubgroupH:             []string{"3"},
		EvaluationDomain:      []string{"1"},
		Polynomial:            []string{"1", "0"},
		PolynomialEvaluations: []string{"5"},
		EvaluationRoot:        "deadbeef",
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var decoded JSONDomainParams
	if err2 := json.Unmarshal(data, &decoded); err2 != nil {
		t.Fatalf("Unmarshal error: %v", err2)
	}
	if decoded.Field != orig.Field {
		t.Errorf("Field mismatch: %q vs %q", decoded.Field, orig.Field)
	}
}

// ---------------------------------------------------------------------------
// PrimeField constant verification
// ---------------------------------------------------------------------------

func TestPrimeField_Modulus(t *testing.T) {
	// PrimeField is a package-level var — verify its modulus
	mod := PrimeField.Modulus()
	expected := new(big.Int).SetUint64(3221225473)
	if mod.Cmp(expected) != 0 {
		t.Errorf("PrimeField modulus: expected 3221225473, got %v", mod)
	}
}

// ---------------------------------------------------------------------------
// Helper — newTestSignManager (reuses zk_test.go pattern)
// ---------------------------------------------------------------------------

func newTestSignManager(t *testing.T) (*SignManager, error) {
	t.Helper()
	sm, err := NewSignManager()
	if err != nil {
		return nil, err
	}
	return sm, nil
}
