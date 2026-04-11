// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 59 — rpc 45.7%→higher
// Tests: processSingleRequest error paths (invalid version, empty method, unknown method),
// processBatchRequest empty batch, Codec.Marshal/Unmarshal additional edge cases,
// Message.Unmarshal error paths
package rpc

import (
	"encoding/json"
	"testing"
)

// helper: rpc server with nil blockchain (no genesis, fast)
func newTestRPCServer() *Server {
	return NewServer(nil, nil)
}

// ─── processSingleRequest — invalid JSON-RPC version ──────────────────────────

func TestSprint59_ProcessSingle_InvalidVersion(t *testing.T) {
	s := newTestRPCServer()
	req := JSONRPCRequest{
		JSONRPC: "1.0", // invalid — must be "2.0"
		Method:  "getblockcount",
		ID:      1,
	}
	resp, err := s.handler.processSingleRequest(req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	// Should be a JSON error response
	var r JSONRPCResponse
	if jsonErr := json.Unmarshal(resp, &r); jsonErr != nil {
		t.Fatalf("response not valid JSON: %v", jsonErr)
	}
	if r.Error == nil {
		t.Error("expected error field in response for invalid version")
	}
}

// ─── processSingleRequest — empty method ──────────────────────────────────────

func TestSprint59_ProcessSingle_EmptyMethod(t *testing.T) {
	s := newTestRPCServer()
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "", // empty method
		ID:      1,
	}
	resp, err := s.handler.processSingleRequest(req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	var r JSONRPCResponse
	if jsonErr := json.Unmarshal(resp, &r); jsonErr != nil {
		t.Fatalf("response not valid JSON: %v", jsonErr)
	}
	if r.Error == nil {
		t.Error("expected error field in response for empty method")
	}
}

// ─── processSingleRequest — unknown method ────────────────────────────────────

func TestSprint59_ProcessSingle_UnknownMethod(t *testing.T) {
	s := newTestRPCServer()
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "no_such_method_xyz",
		ID:      1,
	}
	resp, err := s.handler.processSingleRequest(req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	var r JSONRPCResponse
	if jsonErr := json.Unmarshal(resp, &r); jsonErr != nil {
		t.Fatalf("response not valid JSON: %v", jsonErr)
	}
	if r.Error == nil {
		t.Error("expected error field in response for unknown method")
	}
	if r.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("expected ErrCodeMethodNotFound (%d), got %d", ErrCodeMethodNotFound, r.Error.Code)
	}
}

// ─── ProcessRequest — valid JSON-RPC 2.0 dispatches to processSingleRequest ─

func TestSprint59_ProcessRequest_ValidJSON_UnknownMethod(t *testing.T) {
	s := newTestRPCServer()
	data := []byte(`{"jsonrpc":"2.0","method":"unknownmethod","id":1}`)
	resp, err := s.handler.ProcessRequest(data)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	var r JSONRPCResponse
	if jsonErr := json.Unmarshal(resp, &r); jsonErr != nil {
		t.Fatalf("response not valid JSON: %v", jsonErr)
	}
	if r.Error == nil {
		t.Error("expected error in response for unknown method")
	}
}

// ─── processBatchRequest — empty batch response ───────────────────────────────

func TestSprint59_ProcessBatch_InvalidRequests(t *testing.T) {
	s := newTestRPCServer()
	// batch with invalid version — will produce error responses
	reqs := []JSONRPCRequest{
		{JSONRPC: "1.0", Method: "getblockcount", ID: 1},
	}
	resp, err := s.handler.processBatchRequest(reqs)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	// Should be a JSON array response
	if len(resp) == 0 {
		t.Error("expected non-empty response for batch")
	}
}

// ─── ProcessRequest — batch dispatch ──────────────────────────────────────────

func TestSprint59_ProcessRequest_BatchDispatch(t *testing.T) {
	s := newTestRPCServer()
	batch := `[{"jsonrpc":"2.0","method":"unknownA","id":1},{"jsonrpc":"2.0","method":"unknownB","id":2}]`
	resp, err := s.handler.ProcessRequest([]byte(batch))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if len(resp) == 0 {
		t.Error("expected non-empty batch response")
	}
}

// ─── Codec.Unmarshal — edge cases ─────────────────────────────────────────────

func TestSprint59_Codec_Message_Unmarshal_TooShort(t *testing.T) {
	var msg Message
	err := msg.Unmarshal([]byte{0x01, 0x02}) // too short
	if err == nil {
		t.Error("expected error for too-short data in Message.Unmarshal")
	}
}

func TestSprint59_Codec_Remote_Unmarshal_TooShort(t *testing.T) {
	var r Remote
	err := r.Unmarshal([]byte{0x01}) // too short
	if err == nil {
		t.Error("expected error for too-short data in Remote.Unmarshal")
	}
}

// ─── errorResponse — direct call coverage ─────────────────────────────────────

func TestSprint59_ErrorResponse_NonNilOutput(t *testing.T) {
	h := NewJSONRPCHandler(newTestRPCServer())
	resp, err := h.errorResponse(1, ErrCodeInternalError, "test error")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) == 0 {
		t.Error("expected non-empty error response")
	}
	var r JSONRPCResponse
	if jsonErr := json.Unmarshal(resp, &r); jsonErr != nil {
		t.Fatalf("error response is not valid JSON: %v", jsonErr)
	}
	if r.Error == nil || r.Error.Code != ErrCodeInternalError {
		t.Errorf("unexpected error code: %v", r.Error)
	}
}

// ─── HandleRequest via Server ─────────────────────────────────────────────────

func TestSprint59_HandleRequest_UnknownMethod(t *testing.T) {
	s := newTestRPCServer()
	data := []byte(`{"jsonrpc":"2.0","method":"nosuchmethod","id":42}`)
	resp, _ := s.HandleRequest(data)
	if len(resp) == 0 {
		t.Error("expected non-empty response from HandleRequest")
	}
}
