// PEPPER Sprint 2 — coverage push for 0% coverage areas
// Targets: state.go, serialize.go, quorum.go, selection.go, epoch.go, mode.go, random.go
package consensus

import (
	"math/big"
	"testing"
	"time"
)

// mockBlock satisfies the Block interface for state tests.
type mockBlock struct {
	height   uint64
	hash     string
	prevHash string
}

func (m *mockBlock) GetHeight() uint64              { return m.height }
func (m *mockBlock) GetHash() string                { return m.hash }
func (m *mockBlock) GetPrevHash() string            { return m.prevHash }
func (m *mockBlock) GetTimestamp() int64            { return 0 }
func (m *mockBlock) Validate() error                { return nil }
func (m *mockBlock) GetDifficulty() *big.Int        { return big.NewInt(1) }
func (m *mockBlock) GetCurrentNonce() (uint64, error) { return 0, nil }

// ── ConsensusState ──────────────────────────────────────────────────────────

func TestConsensusState_NewAndGetters(t *testing.T) {
	cs := NewConsensusState()
	if cs == nil {
		t.Fatal("NewConsensusState returned nil")
	}
	if cs.GetCurrentView() != 0 {
		t.Errorf("initial view: want 0, got %d", cs.GetCurrentView())
	}
	if cs.GetCurrentHeight() != 0 {
		t.Errorf("initial height: want 0, got %d", cs.GetCurrentHeight())
	}
	if cs.GetPhase() != PhaseIdle {
		t.Errorf("initial phase: want PhaseIdle, got %d", cs.GetPhase())
	}
	if cs.GetLockedBlock() != nil {
		t.Error("initial locked block should be nil")
	}
	if cs.GetPreparedBlock() != nil {
		t.Error("initial prepared block should be nil")
	}
}

func TestConsensusState_SetGetView(t *testing.T) {
	cs := NewConsensusState()
	cs.SetCurrentView(42)
	if v := cs.GetCurrentView(); v != 42 {
		t.Errorf("SetCurrentView: want 42, got %d", v)
	}
}

func TestConsensusState_SetGetHeight(t *testing.T) {
	cs := NewConsensusState()
	cs.SetCurrentHeight(100)
	if h := cs.GetCurrentHeight(); h != 100 {
		t.Errorf("SetCurrentHeight: want 100, got %d", h)
	}
}

func TestConsensusState_SetGetPhase(t *testing.T) {
	cs := NewConsensusState()
	for _, phase := range []ConsensusPhase{PhasePrePrepared, PhasePrepared, PhaseCommitted, PhaseViewChanging, PhaseIdle} {
		cs.SetPhase(phase)
		if got := cs.GetPhase(); got != phase {
			t.Errorf("SetPhase(%d): got %d", phase, got)
		}
	}
}

func TestConsensusState_LockedBlock(t *testing.T) {
	cs := NewConsensusState()
	mb := &mockBlock{height: 5, hash: "lockedhash"}
	cs.SetLockedBlock(mb)
	got := cs.GetLockedBlock()
	if got == nil {
		t.Fatal("GetLockedBlock returned nil after set")
	}
	if got.GetHash() != "lockedhash" {
		t.Errorf("wrong locked block hash: %s", got.GetHash())
	}
}

func TestConsensusState_PreparedBlock(t *testing.T) {
	cs := NewConsensusState()
	mb := &mockBlock{height: 3, hash: "preparedhash"}
	cs.SetPreparedBlock(mb, 7)
	if cs.GetPreparedBlock() == nil {
		t.Fatal("GetPreparedBlock returned nil after set")
	}
	if cs.GetPreparedBlock().GetHash() != "preparedhash" {
		t.Error("wrong prepared block hash")
	}
}

func TestConsensusState_ResetForNewView(t *testing.T) {
	cs := NewConsensusState()
	mb := &mockBlock{height: 1, hash: "locked"}
	cs.SetLockedBlock(mb)
	cs.SetPreparedBlock(&mockBlock{height: 2, hash: "prepared"}, 3)
	cs.SetPhase(PhaseCommitted)

	cs.ResetForNewView(10)

	if cs.GetCurrentView() != 10 {
		t.Errorf("after reset: view want 10, got %d", cs.GetCurrentView())
	}
	if cs.GetPhase() != PhaseIdle {
		t.Errorf("after reset: phase want PhaseIdle, got %d", cs.GetPhase())
	}
	if cs.GetPreparedBlock() != nil {
		t.Error("after reset: prepared block should be nil")
	}
	// lockedBlock is preserved across view changes
	if cs.GetLockedBlock() == nil {
		t.Error("after reset: locked block should be preserved")
	}
}

func TestConsensusState_ResetForNewViewEnhanced(t *testing.T) {
	cs := NewConsensusState()
	cs.SetPhase(PhasePrepared)
	cs.SetPreparedBlock(&mockBlock{height: 1, hash: "h1"}, 2)

	cs.ResetForNewViewEnhanced(5, []string{"node1", "node2"})

	if cs.GetCurrentView() != 5 {
		t.Errorf("enhanced reset: view want 5, got %d", cs.GetCurrentView())
	}
	if cs.GetPhase() != PhaseIdle {
		t.Error("enhanced reset: phase should be idle")
	}
}

func TestConsensusState_CanChangeView(t *testing.T) {
	cs := NewConsensusState()
	// Freshly created, lastViewChange is zero — should be able to change view
	if !cs.CanChangeView() {
		t.Error("fresh ConsensusState should be able to change view")
	}
}

// ── Serialize (SignedMessage) ────────────────────────────────────────────────

func TestSignedMessage_SerializeDeserialize(t *testing.T) {
	sm := &SignedMessage{
		Signature:  []byte("sig_bytes_here"),
		Timestamp:  []byte{0, 0, 0, 0, 0, 0, 0, 1},
		Nonce:      []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		Commitment: []byte("commitment_hash_32_bytes_padded!!"),
		Data:       []byte("hello quantix"),
	}

	data, err := sm.Serialize()
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Serialize returned empty bytes")
	}

	sm2, err := DeserializeSignedMessage(data)
	if err != nil {
		t.Fatalf("DeserializeSignedMessage: %v", err)
	}
	if string(sm2.Data) != string(sm.Data) {
		t.Errorf("Data mismatch: want %q, got %q", sm.Data, sm2.Data)
	}
}

func TestSignedMessage_SerializeTooShort(t *testing.T) {
	_, err := DeserializeSignedMessage([]byte{1, 2})
	if err == nil {
		t.Error("expected error for too-short input")
	}
}

func TestSignedMessage_SerializeNilMerkleRoot(t *testing.T) {
	sm := &SignedMessage{
		Signature:  []byte("sig"),
		Timestamp:  []byte{1},
		Nonce:      []byte{2},
		MerkleRoot: nil,
		Commitment: []byte("cmt"),
		Data:       []byte("data"),
	}
	data, err := sm.Serialize()
	if err != nil {
		t.Fatalf("Serialize with nil MerkleRoot: %v", err)
	}
	if len(data) == 0 {
		t.Error("serialized empty")
	}
}

// ── QuorumVerifier / QuorumCalculator ───────────────────────────────────────

func TestQuorumVerifier_VerifySafety_True(t *testing.T) {
	qv := NewQuorumVerifier(10, 2, 2.0/3.0)
	if !qv.VerifySafety() {
		t.Error("expected safety to hold for 10 nodes, 2 faulty, 2/3 quorum")
	}
}

func TestQuorumVerifier_VerifySafety_False_TooManyFaulty(t *testing.T) {
	qv := NewQuorumVerifier(10, 4, 2.0/3.0)
	if qv.VerifySafety() {
		t.Error("safety should not hold when faulty >= totalNodes/3")
	}
}

func TestQuorumVerifier_VerifyQuorumIntersection(t *testing.T) {
	qv := NewQuorumVerifier(10, 2, 2.0/3.0)
	if !qv.VerifyQuorumIntersection() {
		t.Error("quorum intersection should hold for standard BFT setup")
	}
}

func TestQuorumVerifier_CalculateMinQuorumSize(t *testing.T) {
	qv := NewQuorumVerifier(10, 3, 2.0/3.0)
	sz := qv.CalculateMinQuorumSize()
	if sz < 1 {
		t.Errorf("min quorum size must be at least 1, got %d", sz)
	}
	// ceil(10 * 2/3) = ceil(6.67) = 7
	if sz != 7 {
		t.Errorf("CalculateMinQuorumSize: want 7, got %d", sz)
	}
}

func TestQuorumVerifier_CalculateMinQuorumSizeZeroNodes(t *testing.T) {
	qv := NewQuorumVerifier(0, 0, 2.0/3.0)
	sz := qv.CalculateMinQuorumSize()
	if sz < 1 {
		t.Errorf("min quorum size for 0 nodes should be at least 1, got %d", sz)
	}
}

func TestCalculateOptimalQuorumFraction_SmallFaulty(t *testing.T) {
	f := CalculateOptimalQuorumFraction(1, 10)
	if f < 2.0/3.0 {
		t.Errorf("fraction should be at least 2/3, got %f", f)
	}
}

func TestCalculateOptimalQuorumFraction_ZeroNodes(t *testing.T) {
	f := CalculateOptimalQuorumFraction(0, 0)
	if f != 2.0/3.0 {
		t.Errorf("zero nodes: want 2/3, got %f", f)
	}
}

func TestQuorumCalculator_VerifyQuorumIntersection(t *testing.T) {
	qc := NewQuorumCalculator(2.0 / 3.0)
	if !qc.VerifyQuorumIntersection(10, 2) {
		t.Error("intersection should hold for standard BFT setup")
	}
}

func TestQuorumCalculator_CalculateMaxFaulty(t *testing.T) {
	qc := NewQuorumCalculator(2.0 / 3.0)
	// (1 - 2/3) * 12 = 4
	got := qc.CalculateMaxFaulty(12)
	if got != 4 {
		t.Errorf("CalculateMaxFaulty(12): want 4, got %d", got)
	}
}

func TestQuorumCalculator_CalculateMaxFaulty_ZeroNodes(t *testing.T) {
	qc := NewQuorumCalculator(2.0 / 3.0)
	got := qc.CalculateMaxFaulty(0)
	if got < 0 {
		t.Error("max faulty should be >= 0")
	}
}

// ── ConsensusMode ────────────────────────────────────────────────────────────

func TestConsensusModeString(t *testing.T) {
	tests := []struct {
		mode ConsensusMode
		want string
	}{
		{DEVNET_SOLO, "DEVNET_SOLO"},
		{PBFT, "PBFT"},
		{ConsensusMode(99), "UNKNOWN"},
	}
	for _, tc := range tests {
		if got := tc.mode.String(); got != tc.want {
			t.Errorf("ConsensusMode(%d).String() = %q, want %q", tc.mode, got, tc.want)
		}
	}
}

// ── StakeWeightedSelector ────────────────────────────────────────────────────

func makeVSWithValidators(t *testing.T, count int) *ValidatorSet {
	t.Helper()
	minStake := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e9))
	vs := NewValidatorSet(minStake)
	for i := 0; i < count; i++ {
		id := "node" + string(rune('A'+i))
		stake := uint64(1000 + i*100) // in QTX units used by RegisterValidator
		err := vs.RegisterValidator(id, "pubhex"+id, stake, "127.0.0.1:3000")
		if err != nil {
			t.Fatalf("RegisterValidator %d: %v", i, err)
		}
	}
	return vs
}

func TestStakeWeightedSelector_NewNotNil(t *testing.T) {
	vs := makeVSWithValidators(t, 3)
	sel := NewStakeWeightedSelector(vs)
	if sel == nil {
		t.Fatal("NewStakeWeightedSelector returned nil")
	}
}

func TestStakeWeightedSelector_SelectProposer(t *testing.T) {
	vs := makeVSWithValidators(t, 4)
	sel := NewStakeWeightedSelector(vs)
	var seed [32]byte
	copy(seed[:], "deterministic_seed_for_testing!!")
	p := sel.SelectProposer(1, seed)
	if p == nil {
		t.Error("SelectProposer returned nil with validators present")
	}
}

func TestStakeWeightedSelector_SelectProposerNoValidators(t *testing.T) {
	minStake := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e9))
	vs := NewValidatorSet(minStake)
	sel := NewStakeWeightedSelector(vs)
	var seed [32]byte
	p := sel.SelectProposer(1, seed)
	if p != nil {
		t.Error("SelectProposer with no validators should return nil")
	}
}

func TestStakeWeightedSelector_SelectCommittee(t *testing.T) {
	vs := makeVSWithValidators(t, 5)
	sel := NewStakeWeightedSelector(vs)
	var seed [32]byte
	copy(seed[:], "committee_seed_bytes_____________")
	committee := sel.SelectCommittee(1, seed, 3)
	if len(committee) != 3 {
		t.Errorf("SelectCommittee(size=3): want 3 members, got %d", len(committee))
	}
}

func TestStakeWeightedSelector_SelectCommitteeLargerThanValidators(t *testing.T) {
	vs := makeVSWithValidators(t, 2)
	sel := NewStakeWeightedSelector(vs)
	var seed [32]byte
	committee := sel.SelectCommittee(1, seed, 10)
	if len(committee) > 2 {
		t.Error("committee size should not exceed available validators")
	}
}

// ── TimeConverter (epoch.go) ─────────────────────────────────────────────────

func TestTimeConverter_SlotToTime(t *testing.T) {
	genesis := time.Now().Add(-100 * SlotDuration)
	tc := NewTimeConverter(genesis)
	slot0Time := tc.SlotToTime(0)
	if !slot0Time.Equal(genesis) {
		t.Errorf("SlotToTime(0) should equal genesis")
	}
	slot1Time := tc.SlotToTime(1)
	diff := slot1Time.Sub(genesis)
	if diff != SlotDuration {
		t.Errorf("SlotToTime(1) - genesis should be %v, got %v", SlotDuration, diff)
	}
}

func TestTimeConverter_CurrentSlot(t *testing.T) {
	genesis := time.Now().Add(-5 * SlotDuration)
	tc := NewTimeConverter(genesis)
	slot := tc.CurrentSlot()
	if slot < 4 || slot > 6 {
		t.Errorf("CurrentSlot: expected ~5, got %d", slot)
	}
}

func TestTimeConverter_CurrentEpoch(t *testing.T) {
	genesis := time.Now().Add(-time.Duration(SlotsPerEpoch+1) * SlotDuration)
	tc := NewTimeConverter(genesis)
	epoch := tc.CurrentEpoch()
	if epoch < 1 {
		t.Errorf("CurrentEpoch: expected >= 1, got %d", epoch)
	}
}

func TestTimeConverter_IsNewEpoch(t *testing.T) {
	genesis := time.Now().Add(-time.Duration(SlotsPerEpoch+5) * SlotDuration)
	tc := NewTimeConverter(genesis)
	if !tc.IsNewEpoch(0) {
		t.Error("IsNewEpoch(0) should be true when in epoch > 0")
	}
	currentEpoch := tc.CurrentEpoch()
	if tc.IsNewEpoch(currentEpoch) {
		t.Error("IsNewEpoch(current) should be false")
	}
}

// ── VDFCache (random.go) ─────────────────────────────────────────────────────

func TestVDFCache_MarkAndIsVerified(t *testing.T) {
	vc := NewVDFCache()
	if vc.IsVerified(1, "nodeA") {
		t.Error("should not be verified before MarkVerified")
	}
	vc.MarkVerified(1, "nodeA")
	if !vc.IsVerified(1, "nodeA") {
		t.Error("should be verified after MarkVerified")
	}
}

func TestVDFCache_Prune(t *testing.T) {
	vc := NewVDFCache()
	vc.MarkVerified(1, "nodeA")
	vc.MarkVerified(2, "nodeB")
	vc.MarkVerified(5, "nodeC")
	vc.Prune(3)
	if vc.IsVerified(1, "nodeA") {
		t.Error("epoch 1 should have been pruned")
	}
	if vc.IsVerified(2, "nodeB") {
		t.Error("epoch 2 should have been pruned")
	}
	if !vc.IsVerified(5, "nodeC") {
		t.Error("epoch 5 should remain after pruning epoch < 3")
	}
}
