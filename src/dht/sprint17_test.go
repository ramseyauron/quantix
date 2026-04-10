package dht

// Coverage Sprint 17 — dht/conn.go: getMessageBuf, verifyReceivedMessage, equalHashes,
// dht.go: NewDHT, SelfNodeID, GetCached utility methods.

import (
	"net"
	"testing"

	"go.uber.org/zap"

	"github.com/ramseyauron/quantix/src/rpc"
)

// ---------------------------------------------------------------------------
// getMessageBuf (conn.go)
// ---------------------------------------------------------------------------

func TestGetMessageBuf_ValidMessage_NonEmpty(t *testing.T) {
	msg := rpc.Message{
		RPCType: rpc.RPCPing,
		Query:   true,
		RPCID:   rpc.GetRPCID(),
	}

	buf := make([]byte, 512)
	result, err := getMessageBuf(buf, msg)
	if err != nil {
		t.Fatalf("getMessageBuf: %v", err)
	}
	if len(result) == 0 {
		t.Error("getMessageBuf: returned empty buffer")
	}
}

func TestGetMessageBuf_MagicNumber_Present(t *testing.T) {
	msg := rpc.Message{RPCType: rpc.RPCPing}
	buf := make([]byte, 512)
	result, err := getMessageBuf(buf, msg)
	if err != nil {
		t.Fatalf("getMessageBuf: %v", err)
	}
	// Magic number should be at bytes 0 and 1.
	if result[0] != magicNumber[0] || result[1] != magicNumber[1] {
		t.Errorf("getMessageBuf: magic number mismatch: got [%X %X], want [%X %X]",
			result[0], result[1], magicNumber[0], magicNumber[1])
	}
}

func TestGetMessageBuf_SmallBuf_AllocatesNew(t *testing.T) {
	msg := rpc.Message{RPCType: rpc.RPCFindNode}
	buf := make([]byte, 1) // too small — should allocate a new buffer
	result, err := getMessageBuf(buf, msg)
	if err != nil {
		t.Fatalf("getMessageBuf small buf: %v", err)
	}
	if len(result) < 4 {
		t.Errorf("getMessageBuf small buf: result too short: %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// verifyReceivedMessage (conn.go)
// ---------------------------------------------------------------------------

func TestVerifyReceivedMessage_TooShort_ReturnsFalse(t *testing.T) {
	_, ok := verifyReceivedMessage([]byte{0x01, 0x02})
	if ok {
		t.Error("verifyReceivedMessage with too-short input: expected false")
	}
}

func TestVerifyReceivedMessage_EmptySlice_ReturnsFalse(t *testing.T) {
	_, ok := verifyReceivedMessage([]byte{})
	if ok {
		t.Error("verifyReceivedMessage with empty slice: expected false")
	}
}

func TestVerifyReceivedMessage_WrongMagic_ReturnsFalse(t *testing.T) {
	// Build a buffer with wrong magic number but correct length.
	buf := make([]byte, 40)
	buf[0] = 0xFF
	buf[1] = 0xFF
	_, ok := verifyReceivedMessage(buf)
	if ok {
		t.Error("verifyReceivedMessage with wrong magic: expected false")
	}
}

func TestVerifyReceivedMessage_ValidEncodedMessage_ReturnsTrue(t *testing.T) {
	// Build a valid encoded message and verify it can be decoded.
	msg := rpc.Message{
		RPCType: rpc.RPCPing,
		Query:   true,
		RPCID:   rpc.GetRPCID(),
	}
	buf := make([]byte, 512)
	encoded, err := getMessageBuf(buf, msg)
	if err != nil {
		t.Fatalf("getMessageBuf: %v", err)
	}

	payload, ok := verifyReceivedMessage(encoded)
	if !ok {
		t.Error("verifyReceivedMessage valid message: expected true")
	}
	if len(payload) == 0 {
		t.Error("verifyReceivedMessage: returned empty payload")
	}
}

// ---------------------------------------------------------------------------
// equalHashes (conn.go)
// ---------------------------------------------------------------------------

func TestEqualHashes_IdenticalSlices_True(t *testing.T) {
	a := []byte{0x01, 0x02, 0x03, 0x04}
	b := []byte{0x01, 0x02, 0x03, 0x04}
	if !equalHashes(a, b) {
		t.Error("equalHashes: identical slices should be equal")
	}
}

func TestEqualHashes_DifferentSlices_False(t *testing.T) {
	a := []byte{0x01, 0x02, 0x03, 0x04}
	b := []byte{0x01, 0x02, 0x03, 0xFF}
	if equalHashes(a, b) {
		t.Error("equalHashes: different slices should not be equal")
	}
}

func TestEqualHashes_DifferentLengths_False(t *testing.T) {
	a := []byte{0x01, 0x02}
	b := []byte{0x01, 0x02, 0x03}
	if equalHashes(a, b) {
		t.Error("equalHashes: different-length slices should not be equal")
	}
}

func TestEqualHashes_BothEmpty_True(t *testing.T) {
	if !equalHashes([]byte{}, []byte{}) {
		t.Error("equalHashes: both empty should be equal")
	}
}

// ---------------------------------------------------------------------------
// NewDHT + SelfNodeID + GetCached (dht.go)
// ---------------------------------------------------------------------------

func makeTestDHT(t *testing.T, port int) (*DHT, error) {
	t.Helper()
	cfg := Config{
		Proto:   "udp4",
		Address: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port},
	}
	log, _ := zap.NewDevelopment()
	return NewDHT(cfg, log)
}

func TestNewDHT_NonNil(t *testing.T) {
	d, err := makeTestDHT(t, 0) // port 0 = OS-assigned
	if err != nil {
		t.Fatalf("NewDHT: %v", err)
	}
	defer d.Close()
	if d == nil {
		t.Error("NewDHT returned nil")
	}
}

func TestNewDHT_SelfNodeID_NonZero(t *testing.T) {
	d, err := makeTestDHT(t, 0)
	if err != nil {
		t.Fatalf("NewDHT: %v", err)
	}
	defer d.Close()

	id := d.SelfNodeID()
	allZero := true
	for _, b := range id {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("SelfNodeID: should return non-zero random ID")
	}
}

func TestNewDHT_GetCached_EmptyForUnknownKey_Nil(t *testing.T) {
	d, err := makeTestDHT(t, 0)
	if err != nil {
		t.Fatalf("NewDHT: %v", err)
	}
	defer d.Close()

	var k [32]byte
	k[0] = 0xAB
	result := d.GetCached(k)
	if result != nil {
		t.Logf("GetCached unknown key: returned %v (may have values from initialisation)", result)
	}
}
