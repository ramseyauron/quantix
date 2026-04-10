package rpc

// Coverage Sprint 16 — rpc/codec.go: Codec utility methods, NodeID.String(),
// Remote.MarshalSize/Marshal/Unmarshal, Message.MarshalSize/Marshal/Unmarshal.

import (
	"net"
	"testing"
)

// ---------------------------------------------------------------------------
// NodeID.String
// ---------------------------------------------------------------------------

func TestNodeID_String_Length(t *testing.T) {
	var id NodeID
	s := id.String()
	if len(s) != 64 {
		t.Errorf("NodeID.String() length: expected 64, got %d", len(s))
	}
}

func TestNodeID_String_AllZeros_IsZeroString(t *testing.T) {
	var id NodeID
	s := id.String()
	for _, c := range s {
		if c != '0' {
			t.Errorf("all-zero NodeID.String() should be all '0', got %q", s)
			break
		}
	}
}

func TestNodeID_String_NonZero_NonEmpty(t *testing.T) {
	var id NodeID
	id[0] = 0xde
	id[1] = 0xad
	s := id.String()
	if s == "" {
		t.Error("NodeID.String() should not be empty")
	}
	if s[:4] != "dead" {
		t.Errorf("NodeID.String(): expected to start with 'dead', got %q", s[:4])
	}
}

// ---------------------------------------------------------------------------
// Codec — uint64/uint32/uint16 round-trips
// ---------------------------------------------------------------------------

func TestCodec_Uint64_Roundtrip(t *testing.T) {
	c := Codec{}
	buf := make([]byte, 8)
	values := []uint64{0, 1, 255, 65535, 1<<32 - 1, 1<<63 - 1, ^uint64(0)}
	for _, v := range values {
		c.PutUint64(buf, v)
		got := c.Uint64(buf)
		if got != v {
			t.Errorf("Codec uint64 roundtrip: %d → put → get → %d", v, got)
		}
	}
}

func TestCodec_Uint32_Roundtrip(t *testing.T) {
	c := Codec{}
	buf := make([]byte, 4)
	values := []uint32{0, 1, 255, 65535, ^uint32(0)}
	for _, v := range values {
		c.PutUint32(buf, v)
		got := c.Uint32(buf)
		if got != v {
			t.Errorf("Codec uint32 roundtrip: %d → put → get → %d", v, got)
		}
	}
}

func TestCodec_Uint16_Roundtrip(t *testing.T) {
	c := Codec{}
	buf := make([]byte, 2)
	values := []uint16{0, 1, 255, ^uint16(0)}
	for _, v := range values {
		c.PutUint16(buf, v)
		got := c.Uint16(buf)
		if got != v {
			t.Errorf("Codec uint16 roundtrip: %d → put → get → %d", v, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Remote.MarshalSize / Marshal / Unmarshal
// ---------------------------------------------------------------------------

func makeTestRemote() Remote {
	var id NodeID
	id[0] = 0x11
	id[1] = 0x22
	return Remote{
		NodeID: id,
		Address: net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 32307,
			Zone: "",
		},
	}
}

func TestRemote_MarshalSize_Positive(t *testing.T) {
	r := makeTestRemote()
	sz := r.MarshalSize()
	if sz <= 0 {
		t.Errorf("Remote.MarshalSize() should be positive, got %d", sz)
	}
}

func TestRemote_MarshalUnmarshal_Roundtrip(t *testing.T) {
	orig := makeTestRemote()
	buf := make([]byte, orig.MarshalSize())

	remaining, err := orig.Marshal(buf)
	if err != nil {
		t.Fatalf("Remote.Marshal: %v", err)
	}
	_ = remaining

	var got Remote
	err = got.Unmarshal(buf)
	if err != nil {
		t.Fatalf("Remote.Unmarshal: %v", err)
	}

	if got.NodeID != orig.NodeID {
		t.Errorf("Remote roundtrip: NodeID mismatch\norig: %v\ngot:  %v", orig.NodeID, got.NodeID)
	}
	if got.Address.Port != orig.Address.Port {
		t.Errorf("Remote roundtrip: Port mismatch: orig=%d got=%d", orig.Address.Port, got.Address.Port)
	}
}

func TestRemote_Marshal_BufferTooSmall_Error(t *testing.T) {
	r := makeTestRemote()
	buf := make([]byte, 1) // too small
	_, err := r.Marshal(buf)
	if err == nil {
		t.Error("Remote.Marshal with tiny buffer: expected error")
	}
}

func TestRemote_Unmarshal_BufferTooSmall_Error(t *testing.T) {
	var r Remote
	err := r.Unmarshal([]byte{0x01}) // too small
	if err == nil {
		t.Error("Remote.Unmarshal with tiny buffer: expected error")
	}
}

// ---------------------------------------------------------------------------
// Message.MarshalSize / Marshal / Unmarshal
// ---------------------------------------------------------------------------

func makeTestMessage() Message {
	var id NodeID
	id[3] = 0xAB

	return Message{
		RPCType: RPCFindNode,
		Query:   true,
		TTL:     30,
		Target:  id,
		RPCID:   12345,
		From:    makeTestRemote(),
	}
}

func TestMessage_MarshalSize_Positive(t *testing.T) {
	m := makeTestMessage()
	sz := m.MarshalSize()
	if sz <= 0 {
		t.Errorf("Message.MarshalSize() should be positive, got %d", sz)
	}
}

func TestMessage_MarshalUnmarshal_Roundtrip(t *testing.T) {
	orig := makeTestMessage()
	buf := make([]byte, orig.MarshalSize())

	remaining, err := orig.Marshal(buf)
	if err != nil {
		t.Fatalf("Message.Marshal: %v", err)
	}
	_ = remaining

	var got Message
	err = got.Unmarshal(buf)
	if err != nil {
		t.Fatalf("Message.Unmarshal: %v", err)
	}

	if got.RPCType != orig.RPCType {
		t.Errorf("Message roundtrip: RPCType mismatch: orig=%v got=%v", orig.RPCType, got.RPCType)
	}
	if got.RPCID != orig.RPCID {
		t.Errorf("Message roundtrip: RPCID mismatch: orig=%v got=%v", orig.RPCID, got.RPCID)
	}
	if got.TTL != orig.TTL {
		t.Errorf("Message roundtrip: TTL mismatch: orig=%v got=%v", orig.TTL, got.TTL)
	}
	if got.Target != orig.Target {
		t.Errorf("Message roundtrip: Target mismatch")
	}
}

func TestMessage_Marshal_BufferTooSmall_Error(t *testing.T) {
	m := makeTestMessage()
	buf := make([]byte, 2) // far too small
	_, err := m.Marshal(buf)
	if err == nil {
		t.Error("Message.Marshal with tiny buffer: expected error")
	}
}

func TestMessage_Unmarshal_TooSmall_Error(t *testing.T) {
	var m Message
	err := m.Unmarshal([]byte{0x01})
	if err == nil {
		t.Error("Message.Unmarshal with tiny buffer: expected error")
	}
}
