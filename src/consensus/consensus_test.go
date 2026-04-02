// MIT License
//
// Copyright (c) 2024 quantix
//
// go/src/consensus/consensus_test.go
package consensus

import (
	"math"
	"math/big"
	"testing"
)

// ---------------------------------------------------------------------------
// Q3-A: Quorum calculation
// ---------------------------------------------------------------------------

// quorumFor returns the minimum quorum needed for N total nodes using 2/3+1 BFT rule.
// This matches CalculateMinQuorumSize with quorumFraction = 2/3.
func quorumFor(n int) int {
	return int(math.Ceil(float64(n) * (2.0 / 3.0)))
}

func TestQuorumCalculation(t *testing.T) {
	tests := []struct {
		name          string
		totalNodes    int
		wantMinQuorum int
	}{
		{"N=4 need 3", 4, 3},
		{"N=7 need 5", 7, 5},
		{"N=10 need 7", 10, 7},
		{"N=1 need 1", 1, 1},
		{"N=3 need 2", 3, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qv := NewQuorumVerifier(tc.totalNodes, 0, 2.0/3.0)
			got := qv.CalculateMinQuorumSize()
			if got != tc.wantMinQuorum {
				t.Errorf("N=%d: want quorum %d, got %d", tc.totalNodes, tc.wantMinQuorum, got)
			}
		})
	}
}

func TestQuorumVerifierSafety(t *testing.T) {
	tests := []struct {
		name        string
		totalNodes  int
		faultyNodes int
		qFraction   float64
		wantSafe    bool
	}{
		// For N=4, f=1: VerifySafety checks f < N/3 (integer division): 1 < 1 = false
		{"4 nodes 1 faulty 2/3 quorum", 4, 1, 2.0 / 3.0, false},
		// For N=6, f=1: 1 < 2 = true
		{"6 nodes 1 faulty 2/3 quorum", 6, 1, 2.0 / 3.0, true},
		// Faulty >= total/3 breaks safety
		{"3 nodes 2 faulty — unsafe", 3, 2, 2.0 / 3.0, false},
		// Quorum too low
		{"4 nodes 0 faulty 0.5 quorum — unsafe", 4, 0, 0.5, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qv := NewQuorumVerifier(tc.totalNodes, tc.faultyNodes, tc.qFraction)
			got := qv.VerifySafety()
			if got != tc.wantSafe {
				t.Errorf("VerifySafety = %v, want %v", got, tc.wantSafe)
			}
		})
	}
}

func TestQuorumCalculatorMaxFaulty(t *testing.T) {
	tests := []struct {
		name       string
		n          int
		wantFaulty int
	}{
		{"N=4", 4, 1},
		{"N=7", 7, 2},
		{"N=10", 10, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qc := NewQuorumCalculator(2.0 / 3.0)
			got := qc.CalculateMaxFaulty(tc.n)
			if got != tc.wantFaulty {
				t.Errorf("N=%d: want maxFaulty %d, got %d", tc.n, tc.wantFaulty, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Q3-B: Slashing — ValidatorSet stake and ejection
// ---------------------------------------------------------------------------

func newValidatorSet(minStakeQTX int64) *ValidatorSet {
	minStake := new(big.Int).Mul(big.NewInt(minStakeQTX), big.NewInt(1e9)) // nQTX
	return NewValidatorSet(minStake)
}

func TestSlashValidatorStakeReduction(t *testing.T) {
	vs := newValidatorSet(1000)
	err := vs.AddValidator("val-1", 2000) // 2000 QTX stake
	if err != nil {
		t.Fatalf("AddValidator failed: %v", err)
	}

	v := vs.validators["val-1"]
	initialStake := new(big.Int).Set(v.StakeAmount)

	// Slash 100 basis points = 1%
	vs.SlashValidator("val-1", "missed VDF submission", 100)

	expectedPenalty := new(big.Int).Div(
		new(big.Int).Mul(initialStake, big.NewInt(100)),
		big.NewInt(10000),
	)
	expectedRemaining := new(big.Int).Sub(initialStake, expectedPenalty)

	if v.StakeAmount.Cmp(expectedRemaining) != 0 {
		t.Errorf("after slash: want stake %s, got %s", expectedRemaining, v.StakeAmount)
	}
}

func TestSlashValidatorEjection(t *testing.T) {
	// AddValidator multiplies stakeSPX by denom.QTX (1e18).
	// Set minStake = 90% of fullStake so a 50% slash drops below min.
	stakeQTX := int64(1000)
	oneQTX := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	fullStake := new(big.Int).Mul(big.NewInt(stakeQTX), oneQTX)
	// minStake = 90% of fullStake
	minStake := new(big.Int).Mul(fullStake, big.NewInt(90))
	minStake.Div(minStake, big.NewInt(100))
	vs := NewValidatorSet(minStake)

	err := vs.AddValidator("val-ejected", uint64(stakeQTX))
	if err != nil {
		t.Fatalf("AddValidator failed: %v", err)
	}
	if vs.validators["val-ejected"].IsSlashed {
		t.Fatal("validator should not be slashed initially")
	}

	// Slash 50% — drops below minimum
	vs.SlashValidator("val-ejected", "double-sign", 5000)

	if !vs.validators["val-ejected"].IsSlashed {
		t.Error("validator should be marked as slashed after dropping below minimum stake")
	}
}

func TestSlashUnknownValidatorIsNoop(t *testing.T) {
	vs := newValidatorSet(1000)
	// Should not panic on unknown validator
	vs.SlashValidator("unknown-val", "test", 100)
}

func TestSlashValidatorDoubleSignDetection(t *testing.T) {
	// Double-sign detection: if a validator appears to have signed two different
	// blocks at the same height, slash them. This test verifies the slashing
	// mechanism works correctly for a double-sign scenario.
	vs := newValidatorSet(1000)
	if err := vs.AddValidator("double-signer", 5000); err != nil {
		t.Fatalf("AddValidator: %v", err)
	}

	v := vs.validators["double-signer"]
	stakeBefore := new(big.Int).Set(v.StakeAmount)

	// Simulate double-sign slash (same penalty as missed VDF but triggered on equivocation)
	vs.SlashValidator("double-signer", "double-sign detected", SlashBps)

	if v.StakeAmount.Cmp(stakeBefore) >= 0 {
		t.Error("stake should have decreased after double-sign slash")
	}
}

// ---------------------------------------------------------------------------
// Q3-C: RANDAO — SyncState with valid and invalid proofs
// ---------------------------------------------------------------------------

// mockVDF is a deterministic stub used in tests: Eval always returns (1,1) and
// Verify always returns true (or false when toggled).
type mockVDF struct {
	shouldVerify bool
}

func (m *mockVDF) Eval(params VDFParams, x *big.Int) (y, proof *big.Int, err error) {
	return big.NewInt(1), big.NewInt(1), nil
}

func (m *mockVDF) Verify(params VDFParams, x, y, proof *big.Int) bool {
	return m.shouldVerify
}

// newTestRANDAO creates a RANDAO with a deterministic mock VDF.
func newTestRANDAO(verify bool) *RANDAO {
	disc := big.NewInt(-23)
	params := VDFParams{Discriminant: disc, T: 1, Lambda: 128}
	var seed [32]byte
	copy(seed[:], []byte("test-genesis-seed-12345678901234"))

	r := &RANDAO{
		mix:             seed,
		reveals:         make(map[uint64][][32]byte),
		submissions:     make(map[uint64]map[string]*VDFSubmission),
		missed:          make(map[uint64]map[string]bool),
		epochFinalized:  make(map[uint64]bool),
		params:          params,
		impl:            &mockVDF{shouldVerify: verify},
		cache:           NewVDFCache(),
		consecutiveFailures: make(map[string]int),
	}
	return r
}

func TestRANDAOSyncWithSameMixIsNoop(t *testing.T) {
	r := newTestRANDAO(true)
	originalMix := r.mix

	// Syncing with same mix should succeed immediately without touching state
	err := r.SyncState(originalMix, nil)
	if err != nil {
		t.Errorf("SyncState with same mix should succeed, got: %v", err)
	}
	if r.mix != originalMix {
		t.Error("mix should not change when syncing same mix")
	}
}

func TestRANDAOSyncWithValidProofAccepted(t *testing.T) {
	r := newTestRANDAO(true)
	originalMix := r.mix

	// Different peer mix
	var peerMix [32]byte
	copy(peerMix[:], []byte("peer-mix-value-for-testing-12345"))

	// Build a valid peer submission whose Input matches the RANDAO seed for epoch 1
	peerSeed := r.GetSeed(1)
	peerSub := &VDFSubmission{
		Epoch:       1,
		SlotInEpoch: 5,
		ValidatorID: "peer-val-1",
		Input:       peerSeed,
		Output:      big.NewInt(1),
		Proof:       big.NewInt(1),
	}
	peerSubmissions := map[uint64]map[string]*VDFSubmission{
		1: {"peer-val-1": peerSub},
	}

	err := r.SyncState(peerMix, peerSubmissions)
	if err != nil {
		t.Errorf("SyncState with valid proof should succeed, got: %v", err)
	}
	if r.mix == originalMix {
		t.Error("mix should have updated to peer mix after valid sync")
	}
}

func TestRANDAOSyncWithInvalidProofRejected(t *testing.T) {
	r := newTestRANDAO(false) // mock returns false for Verify

	var peerMix [32]byte
	copy(peerMix[:], []byte("peer-mix-value-for-testing-12345"))

	peerSeed := r.GetSeed(1)
	peerSub := &VDFSubmission{
		Epoch:       1,
		SlotInEpoch: 5,
		ValidatorID: "bad-val",
		Input:       peerSeed,
		Output:      big.NewInt(999),
		Proof:       big.NewInt(999),
	}
	peerSubmissions := map[uint64]map[string]*VDFSubmission{
		1: {"bad-val": peerSub},
	}

	originalMix := r.mix
	err := r.SyncState(peerMix, peerSubmissions)
	if err == nil {
		t.Error("SyncState with invalid VDF proof should be rejected")
	}
	if r.mix != originalMix {
		t.Error("mix should not change when SyncState is rejected")
	}
}

func TestRANDAOSyncWithInputMismatchRejected(t *testing.T) {
	r := newTestRANDAO(true)

	var peerMix [32]byte
	copy(peerMix[:], []byte("peer-mix-value-for-testing-12345"))

	// Wrong input (not matching the seed for epoch 1)
	var wrongInput [32]byte
	copy(wrongInput[:], []byte("wrong-input-00000000000000000000"))

	peerSub := &VDFSubmission{
		Epoch:       1,
		SlotInEpoch: 5,
		ValidatorID: "peer-val-1",
		Input:       wrongInput, // mismatch
		Output:      big.NewInt(1),
		Proof:       big.NewInt(1),
	}
	peerSubmissions := map[uint64]map[string]*VDFSubmission{
		1: {"peer-val-1": peerSub},
	}

	originalMix := r.mix
	err := r.SyncState(peerMix, peerSubmissions)
	if err == nil {
		t.Error("SyncState with mismatched input should be rejected")
	}
	if r.mix != originalMix {
		t.Error("mix should not change when SyncState is rejected due to input mismatch")
	}
}

func TestRANDAOSyncWithNilSubmissionRejected(t *testing.T) {
	r := newTestRANDAO(true)

	var peerMix [32]byte
	copy(peerMix[:], []byte("peer-mix-value-for-testing-12345"))

	peerSubmissions := map[uint64]map[string]*VDFSubmission{
		1: {"nil-val": nil}, // nil submission
	}

	err := r.SyncState(peerMix, peerSubmissions)
	if err == nil {
		t.Error("SyncState with nil submission should be rejected")
	}
}
