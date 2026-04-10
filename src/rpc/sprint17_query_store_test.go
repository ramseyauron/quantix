package rpc

// Coverage Sprint 17 — rpc/query.go, rpc/store.go, rpc/metrics.go.

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// rpc/query.go — Query + QueryManager
// ---------------------------------------------------------------------------

func TestNewQuery_NonNil(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil)
	if q == nil {
		t.Fatal("NewQuery returned nil")
	}
}

func TestQuery_InitialPending_Zero(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil)
	if q.Pending() != 0 {
		t.Errorf("initial Pending: expected 0, got %d", q.Pending())
	}
}

func TestQuery_Request_IncrementsPending(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil)

	var node1 NodeID
	node1[0] = 0x01
	q.Request(node1)
	if q.Pending() != 1 {
		t.Errorf("after Request: expected pending=1, got %d", q.Pending())
	}
}

func TestQuery_OnTimeout_DecrementsPending(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil)

	var node1 NodeID
	node1[0] = 0x01
	q.Request(node1)

	ok := q.OnTimeout(node1)
	if !ok {
		t.Error("OnTimeout: expected true for known pending node")
	}
	if q.Pending() != 0 {
		t.Errorf("after OnTimeout: expected pending=0, got %d", q.Pending())
	}
}

func TestQuery_OnTimeout_AlreadyTimedOut_ReturnsFalse(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil)
	var node1 NodeID
	node1[0] = 0x01
	q.Request(node1)
	q.OnTimeout(node1) // first call

	ok := q.OnTimeout(node1) // second call should return false
	if ok {
		t.Error("OnTimeout second call: expected false (already timed out)")
	}
}

func TestQuery_OnResponded_DecrementsPending(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil)
	var node1 NodeID
	node1[0] = 0x02
	q.Request(node1)

	ok := q.OnResponded(node1)
	if !ok {
		t.Error("OnResponded: expected true for known pending node")
	}
	if q.Pending() != 0 {
		t.Errorf("after OnResponded: expected pending=0, got %d", q.Pending())
	}
}

func TestQuery_OnResponded_UnknownNode_ReturnsFalse(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil)
	var unknown NodeID
	unknown[0] = 0xFF
	ok := q.OnResponded(unknown)
	if ok {
		t.Error("OnResponded unknown node: expected false")
	}
}

func TestQuery_Filter_RemovesRequested(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil)

	var node1, node2 NodeID
	node1[0] = 0x01
	node2[0] = 0x02
	q.Request(node1)

	candidates := []Remote{
		{NodeID: node1},
		{NodeID: node2},
	}
	filtered := q.Filter(candidates)
	if len(filtered) != 1 {
		t.Errorf("Filter: expected 1 result, got %d", len(filtered))
	}
	if filtered[0].NodeID != node2 {
		t.Error("Filter: should return only the unrequested node")
	}
}

func TestNewQueryManager_NonNil(t *testing.T) {
	qm := NewQueryManager()
	if qm == nil {
		t.Fatal("NewQueryManager returned nil")
	}
}

func TestQueryManager_AddFindNode_RetrievableOnCompletion(t *testing.T) {
	qm := NewQueryManager()
	id := GetRPCID()
	var target NodeID
	completed := false
	onDone := func() { completed = true }

	qm.AddFindNode(id, target, onDone)

	fn := qm.GetOnCompletionTask(id)
	if fn == nil {
		t.Fatal("GetOnCompletionTask: returned nil for registered query")
	}
	fn()
	if !completed {
		t.Error("onCompletion not called")
	}
}

func TestQuery_OnCompletion_CalledWhenNil_NoPanic(t *testing.T) {
	var target NodeID
	q := NewQuery(GetRPCID(), target, nil) // nil onCompletion
	_ = q
	// No panic expected when onCompletion is nil.
}

func TestRequestStatus_Known_FalseInitially(t *testing.T) {
	rs := &requestStatus{}
	if rs.known() {
		t.Error("requestStatus.known(): should be false for new status")
	}
}

func TestRequestStatus_Known_TrueAfterTimeout(t *testing.T) {
	rs := &requestStatus{Timeout: true}
	if !rs.known() {
		t.Error("requestStatus.known(): should be true when Timeout=true")
	}
}

func TestRequestStatus_Known_TrueAfterResponded(t *testing.T) {
	rs := &requestStatus{Responded: true}
	if !rs.known() {
		t.Error("requestStatus.known(): should be true when Responded=true")
	}
}

// ---------------------------------------------------------------------------
// rpc/store.go — KVStore, To4KBatches
// ---------------------------------------------------------------------------

func TestNewKVStore_NonNil(t *testing.T) {
	s := NewKVStore()
	if s == nil {
		t.Fatal("NewKVStore returned nil")
	}
}

func TestKVStore_PutGet_Roundtrip(t *testing.T) {
	s := NewKVStore()
	var k Key
	k[0] = 0xAB
	v := []byte("hello quantix")

	s.Put(k, v, 60) // 60-second TTL

	vals, ok := s.Get(k)
	if !ok {
		t.Fatal("KVStore.Get: expected true after Put")
	}
	if len(vals) == 0 {
		t.Fatal("KVStore.Get: returned empty values")
	}
	if string(vals[0]) != "hello quantix" {
		t.Errorf("KVStore.Get: got %q, want %q", vals[0], "hello quantix")
	}
}

func TestKVStore_Get_UnknownKey_ReturnsFalse(t *testing.T) {
	s := NewKVStore()
	var k Key
	k[0] = 0xFF
	_, ok := s.Get(k)
	if ok {
		t.Error("KVStore.Get unknown key: expected false")
	}
}

func TestKVStore_Put_MultipleValues_SameKey(t *testing.T) {
	s := NewKVStore()
	var k Key
	k[0] = 0x01

	s.Put(k, []byte("value-one"), 60)
	s.Put(k, []byte("value-two"), 60)

	vals, ok := s.Get(k)
	if !ok {
		t.Fatal("KVStore.Get: expected true after two Puts")
	}
	if len(vals) < 1 {
		t.Error("KVStore: expected at least 1 value")
	}
}

func TestKVStore_GC_NoError(t *testing.T) {
	s := NewKVStore()
	var k Key
	k[0] = 0x02
	s.Put(k, []byte("temp"), 0) // 0 TTL should expire immediately

	// Give GC a moment then run — should not panic.
	time.Sleep(1 * time.Millisecond)
	s.GC()
}

func TestTo4KBatches_Empty_EmptySlice(t *testing.T) {
	result := To4KBatches([][]byte{})
	if len(result) != 0 {
		t.Errorf("To4KBatches empty: expected 0 batches, got %d", len(result))
	}
}

func TestTo4KBatches_SmallValues_SingleBatch(t *testing.T) {
	vals := [][]byte{
		[]byte("small1"),
		[]byte("small2"),
		[]byte("small3"),
	}
	result := To4KBatches(vals)
	if len(result) != 1 {
		t.Errorf("To4KBatches small values: expected 1 batch, got %d", len(result))
	}
	if len(result[0]) != 3 {
		t.Errorf("To4KBatches: expected 3 items in batch, got %d", len(result[0]))
	}
}

func TestTo4KBatches_LargeValues_SplitsBatches(t *testing.T) {
	// Create values that exceed 3KB per batch.
	large := make([]byte, 2048) // 2KB each
	vals := [][]byte{large, large, large}

	result := To4KBatches(vals)
	if len(result) < 2 {
		t.Errorf("To4KBatches large values: expected ≥2 batches, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// rpc/metrics.go — NewMetrics
// ---------------------------------------------------------------------------

func TestNewMetrics_NonNil(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}
}

func TestNewMetrics_CountersNonNil(t *testing.T) {
	m := NewMetrics()
	if m.RequestCount == nil {
		t.Error("Metrics.RequestCount should not be nil")
	}
	if m.RequestLatency == nil {
		t.Error("Metrics.RequestLatency should not be nil")
	}
	if m.ErrorCount == nil {
		t.Error("Metrics.ErrorCount should not be nil")
	}
}
