package zk

import (
	"testing"
)

// Sprint 49 — stark/zk air.go coverage (white-box, package zk)

// ---------------------------------------------------------------------------
// bytesEqual (unexported)
// ---------------------------------------------------------------------------

func TestBytesEqual_Same(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 2, 3}
	if !bytesEqual(a, b) {
		t.Error("identical slices should be equal")
	}
}

func TestBytesEqual_Different(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 2, 4}
	if bytesEqual(a, b) {
		t.Error("different slices should not be equal")
	}
}

func TestBytesEqual_DifferentLength(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 2}
	if bytesEqual(a, b) {
		t.Error("different length slices should not be equal")
	}
}

func TestBytesEqual_BothNil(t *testing.T) {
	if !bytesEqual(nil, nil) {
		t.Error("both nil slices should be equal")
	}
}

func TestBytesEqual_OneNil(t *testing.T) {
	if bytesEqual([]byte{1}, nil) {
		t.Error("one nil slice should not equal non-nil")
	}
}

// ---------------------------------------------------------------------------
// NewChannel / Channel.Send
// ---------------------------------------------------------------------------

func TestNewChannel_NotNilInternal(t *testing.T) {
	ch := NewChannel()
	if ch == nil {
		t.Fatal("NewChannel returned nil")
	}
}

func TestChannel_Send_NonEmpty(t *testing.T) {
	ch := NewChannel()
	data := []byte("test-data-for-channel")
	// Send should not panic
	ch.Send(data)
}

func TestChannel_Send_Empty(t *testing.T) {
	ch := NewChannel()
	ch.Send([]byte{})
}

func TestChannel_Send_Nil(t *testing.T) {
	ch := NewChannel()
	ch.Send(nil) // should not panic
}

// ---------------------------------------------------------------------------
// MarshalJSON — empty trace error path
// ---------------------------------------------------------------------------

func TestDomainParameters_MarshalJSON_EmptyTrace_Error(t *testing.T) {
	dp := &DomainParameters{
		// Empty Trace — should return error
	}
	_, err := dp.MarshalJSON()
	if err == nil {
		t.Error("MarshalJSON with empty trace should return error")
	}
}

// ---------------------------------------------------------------------------
// GenElems — basic usage
// ---------------------------------------------------------------------------

func TestGenElems_OrderZero(t *testing.T) {
	// Order 0 should return empty slice
	result := GenElems(PrimeFieldGen, 0)
	if len(result) != 0 {
		t.Errorf("GenElems order 0: expected empty, got %d", len(result))
	}
}

func TestGenElems_OrderSmall(t *testing.T) {
	// Small order (e.g. 4) — just verify no panic and correct length
	result := GenElems(PrimeFieldGen, 4)
	if len(result) != 4 {
		t.Errorf("GenElems order 4: expected 4 elements, got %d", len(result))
	}
}
