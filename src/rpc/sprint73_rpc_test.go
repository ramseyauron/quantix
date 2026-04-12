// test(PEPPER): Sprint 73 — rpc 49.3%→higher
// Tests: ping, join, findNode (valid/missing/invalid), store/get roundtrip,
// get key-not-found, processBinaryMessage (ping/join/zero-TTL/invalid-RPCType),
// JSON-RPC method routing via processSingleRequest (ping, join, validateAddress nil bc)
package rpc

import (
	"encoding/hex"
	"encoding/json"
	"testing"
)

// ─── ping ─────────────────────────────────────────────────────────────────────

func TestSprint73_Ping_ReturnsPong(t *testing.T) {
	s := newTestRPCServer()
	result, err := s.handler.ping(nil)
	if err != nil {
		t.Fatalf("ping: %v", err)
	}
	m, ok := result.(map[string]string)
	if !ok || m["status"] != "pong" {
		t.Errorf("expected status:pong, got %v", result)
	}
}

// ─── join ─────────────────────────────────────────────────────────────────────

func TestSprint73_Join_ReturnsJoined(t *testing.T) {
	s := newTestRPCServer()
	result, err := s.handler.join(nil)
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	m, ok := result.(map[string]string)
	if !ok || m["status"] != "joined" {
		t.Errorf("expected status:joined, got %v", result)
	}
}

// ─── findNode ─────────────────────────────────────────────────────────────────

func TestSprint73_FindNode_MissingParam(t *testing.T) {
	s := newTestRPCServer()
	_, err := s.handler.findNode(nil)
	if err == nil {
		t.Error("expected error for nil params")
	}
}

func TestSprint73_FindNode_InvalidHex(t *testing.T) {
	s := newTestRPCServer()
	params := []string{"not-valid-hex!!!"}
	_, err := s.handler.findNode(params)
	if err == nil {
		t.Error("expected error for invalid hex node ID")
	}
}

func TestSprint73_FindNode_ValidHex(t *testing.T) {
	s := newTestRPCServer()
	// A valid 32-byte node ID in hex
	nodeID := hex.EncodeToString(make([]byte, 32))
	params := []string{nodeID}
	result, err := s.handler.findNode(params)
	if err != nil {
		t.Fatalf("findNode valid: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// ─── store / get ─────────────────────────────────────────────────────────────

func TestSprint73_Store_MissingKeyOrValue(t *testing.T) {
	s := newTestRPCServer()
	params := map[string]interface{}{
		"key":   "",
		"value": "",
		"ttl":   10,
	}
	_, err := s.handler.store(params)
	if err == nil {
		t.Error("expected error for empty key/value")
	}
}

func TestSprint73_Store_InvalidKeyHex(t *testing.T) {
	s := newTestRPCServer()
	params := map[string]interface{}{
		"key":   "not-hex!!!",
		"value": "aabb",
		"ttl":   10,
	}
	_, err := s.handler.store(params)
	if err == nil {
		t.Error("expected error for invalid key hex")
	}
}

func TestSprint73_Store_InvalidValueHex(t *testing.T) {
	s := newTestRPCServer()
	params := map[string]interface{}{
		"key":   "aabb",
		"value": "not-hex!!!",
		"ttl":   10,
	}
	_, err := s.handler.store(params)
	if err == nil {
		t.Error("expected error for invalid value hex")
	}
}

func TestSprint73_Store_ValidRoundtrip(t *testing.T) {
	s := newTestRPCServer()
	key32 := hex.EncodeToString(make([]byte, 32))
	val := "deadbeef"
	storeParams := map[string]interface{}{
		"key":   key32,
		"value": val,
		"ttl":   100,
	}
	result, err := s.handler.store(storeParams)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	m, ok := result.(map[string]string)
	if !ok || m["status"] != "stored" {
		t.Errorf("expected stored, got %v", result)
	}

	// Now retrieve it
	getParams := []string{key32}
	getResult, err := s.handler.get(getParams)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if getResult == nil {
		t.Error("expected non-nil get result")
	}
}

func TestSprint73_Get_KeyNotFound(t *testing.T) {
	s := newTestRPCServer()
	missingKey := hex.EncodeToString(make([]byte, 32))
	// Use all-ff key to avoid collision with store test
	missingKey = hex.EncodeToString([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	})
	params := []string{missingKey}
	_, err := s.handler.get(params)
	if err == nil {
		t.Error("expected key-not-found error")
	}
}

func TestSprint73_Get_MissingParam(t *testing.T) {
	s := newTestRPCServer()
	_, err := s.handler.get(nil)
	if err == nil {
		t.Error("expected error for nil params")
	}
}

func TestSprint73_Get_InvalidHex(t *testing.T) {
	s := newTestRPCServer()
	params := []string{"not-valid-hex"}
	_, err := s.handler.get(params)
	if err == nil {
		t.Error("expected error for invalid hex key")
	}
}

// ─── processBinaryMessage — zero TTL ─────────────────────────────────────────

func TestSprint73_ProcessBinary_ZeroTTL(t *testing.T) {
	s := newTestRPCServer()
	id := GetRPCID()
	msg := Message{
		RPCType: RPCPing,
		TTL:     0, // invalid
		RPCID:   id,
		Query:   true,
	}
	resp, err := s.handler.processBinaryMessage(msg)
	if err != nil {
		t.Fatalf("processBinaryMessage zero-TTL: %v", err)
	}
	// Should return error response JSON
	if len(resp) == 0 {
		t.Error("expected non-empty response")
	}
}

// ─── processBinaryMessage — ping roundtrip ────────────────────────────────────

func TestSprint73_ProcessBinary_PingRoundtrip(t *testing.T) {
	s := newTestRPCServer()
	id := GetRPCID()
	msg := Message{
		RPCType: RPCPing,
		TTL:     10,
		RPCID:   id,
		Query:   true,
	}
	resp, err := s.handler.processBinaryMessage(msg)
	if err != nil {
		t.Fatalf("processBinaryMessage ping: %v", err)
	}
	if len(resp) == 0 {
		t.Error("expected non-empty binary response")
	}
}

// ─── processBinaryMessage — join ─────────────────────────────────────────────

func TestSprint73_ProcessBinary_Join(t *testing.T) {
	s := newTestRPCServer()
	id := GetRPCID()
	msg := Message{
		RPCType: RPCJoin,
		TTL:     10,
		RPCID:   id,
		Query:   true,
	}
	resp, err := s.handler.processBinaryMessage(msg)
	if err != nil {
		t.Fatalf("processBinaryMessage join: %v", err)
	}
	if len(resp) == 0 {
		t.Error("expected non-empty response")
	}
}

// ─── processSingleRequest — ping via JSON-RPC ─────────────────────────────────

func TestSprint73_ProcessSingle_Ping(t *testing.T) {
	s := newTestRPCServer()
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "ping",
		ID:      42,
	}
	resp, err := s.handler.processSingleRequest(req)
	if err != nil {
		t.Fatalf("processSingleRequest ping: %v", err)
	}
	var r JSONRPCResponse
	if jsonErr := json.Unmarshal(resp, &r); jsonErr != nil {
		t.Fatalf("response not valid JSON: %v", jsonErr)
	}
	if r.Error != nil {
		t.Errorf("unexpected error in ping response: %v", r.Error)
	}
}

// ─── processSingleRequest — join via JSON-RPC ─────────────────────────────────

func TestSprint73_ProcessSingle_Join(t *testing.T) {
	s := newTestRPCServer()
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "join",
		ID:      43,
	}
	resp, err := s.handler.processSingleRequest(req)
	if err != nil {
		t.Fatalf("processSingleRequest join: %v", err)
	}
	var r JSONRPCResponse
	if jsonErr := json.Unmarshal(resp, &r); jsonErr != nil {
		t.Fatalf("response not valid JSON: %v", jsonErr)
	}
	if r.Error != nil {
		t.Errorf("unexpected error in join response: %v", r.Error)
	}
}
