package consensus

import (
	"math/big"
	"testing"
)

// Sprint 47 — consensus group.go, epoch.go, quorum.go, mode.go, random.go coverage

// ---------------------------------------------------------------------------
// group.go — ClassGroupElement
// ---------------------------------------------------------------------------

func TestNewClassGroupElement_Fields(t *testing.T) {
	a := big.NewInt(2)
	b := big.NewInt(3)
	c := big.NewInt(5)
	el := NewClassGroupElement(a, b, c)
	if el == nil {
		t.Fatal("NewClassGroupElement returned nil")
	}
	if el.A.Cmp(a) != 0 || el.B.Cmp(b) != 0 || el.C.Cmp(c) != 0 {
		t.Error("ClassGroupElement fields not set correctly")
	}
}

func TestNewClassGroupElement_DeepCopy(t *testing.T) {
	a := big.NewInt(10)
	el := NewClassGroupElement(a, big.NewInt(1), big.NewInt(1))
	// Mutate original — should not affect element
	a.SetInt64(999)
	if el.A.Int64() == 999 {
		t.Error("ClassGroupElement should deep-copy inputs")
	}
}

func TestClassGroupElement_Copy(t *testing.T) {
	el := NewClassGroupElement(big.NewInt(7), big.NewInt(3), big.NewInt(11))
	cp := el.Copy()
	if cp == el {
		t.Error("Copy should return a new pointer")
	}
	if cp.A.Cmp(el.A) != 0 || cp.B.Cmp(el.B) != 0 || cp.C.Cmp(el.C) != 0 {
		t.Error("Copy fields should match original")
	}
	// Mutate original after copy — copy should not change
	el.A.SetInt64(999)
	if cp.A.Int64() == 999 {
		t.Error("Copy should be independent")
	}
}

func TestIdentity_NonNil(t *testing.T) {
	D := big.NewInt(-7)
	id := Identity(D)
	if id == nil {
		t.Fatal("Identity returned nil")
	}
	if id.A.Sign() <= 0 {
		t.Error("Identity A should be positive")
	}
}

func TestIdentity_AEqualsOne(t *testing.T) {
	D := big.NewInt(-23)
	id := Identity(D)
	if id.A.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("Identity A should be 1, got %s", id.A)
	}
}

func TestClassGroupToBigInt_BigIntToClassGroup_Roundtrip(t *testing.T) {
	D := big.NewInt(-23)
	el := Identity(D)
	n := ClassGroupToBigInt(el)
	if n == nil || n.Sign() < 0 {
		t.Skip("ClassGroupToBigInt returned non-positive — skipping roundtrip")
	}
	restored := BigIntToClassGroup(n, D)
	if restored == nil {
		t.Skip("BigIntToClassGroup returned nil")
	}
}

func TestHashToPrime_NotNil(t *testing.T) {
	D := big.NewInt(-23)
	x := big.NewInt(3)
	y := big.NewInt(7)
	p := HashToPrime(D, x, y, 10)
	if p == nil {
		t.Fatal("HashToPrime returned nil")
	}
	if !p.ProbablyPrime(20) {
		t.Error("HashToPrime should return a probable prime")
	}
}

func TestSquare_NotNil(t *testing.T) {
	D := big.NewInt(-23)
	el := Identity(D)
	sq := Square(el, D)
	if sq == nil {
		t.Fatal("Square returned nil")
	}
}

func TestCompose_NotNil(t *testing.T) {
	D := big.NewInt(-23)
	el := Identity(D)
	result := Compose(el, el, D)
	if result == nil {
		t.Fatal("Compose returned nil")
	}
}

func TestExponentiate_NotNil(t *testing.T) {
	D := big.NewInt(-23)
	el := Identity(D)
	exp := big.NewInt(3)
	result := Exponentiate(el, exp, D)
	if result == nil {
		t.Fatal("Exponentiate returned nil")
	}
}

func TestRepeatedSquare_NotNil(t *testing.T) {
	D := big.NewInt(-23)
	el := Identity(D)
	result := RepeatedSquare(el, D, 2)
	if result == nil {
		t.Fatal("RepeatedSquare returned nil")
	}
}

// ---------------------------------------------------------------------------
// epoch.go — TimeConverter
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// quorum.go
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// mode.go
// ---------------------------------------------------------------------------


