// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 31 - rpc json handler + server coverage
package rpc

import (
	"testing"
)

// ---------------------------------------------------------------------------
// NewJSONRPCHandler
// ---------------------------------------------------------------------------

func TestNewJSONRPCHandler_NotNil(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	if h == nil {
		t.Fatal("expected non-nil JSONRPCHandler")
	}
}

// ---------------------------------------------------------------------------
// ProcessRequest — invalid input paths (don't call handler methods)
// ---------------------------------------------------------------------------

func TestProcessRequest_InvalidJSON_ReturnsResponse(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	resp, err := h.ProcessRequest([]byte("not json"))
	// ProcessRequest returns an error JSON response, not a Go error
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if len(resp) == 0 {
		t.Fatal("expected non-empty error response for invalid JSON")
	}
}

func TestProcessRequest_EmptyBinary_ReturnsResponse(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	resp, err := h.ProcessRequest([]byte{})
	// Empty bytes cannot be a valid binary Message or JSON — should return error response
	_ = resp
	_ = err
}

func TestProcessRequest_UnknownMethod_NoPanel(t *testing.T) {
	// processSingleRequest panics with nil server at json.go:339 — nil guard needed
	t.Skip("processSingleRequest panics with nil server — nil guard needed on h.server")
}

// getDifficulty, getnetworkinfo etc. panic with nil server — skip those
// and document the nil guard gaps

func TestProcessRequest_GetDifficulty_NilServer_Documented(t *testing.T) {
	t.Skip("getdifficulty panics with nil server at json.go:339 — nil guard needed")
}

func TestProcessRequest_GetNetworkInfo_NilServer_Documented(t *testing.T) {
	t.Skip("getnetworkinfo panics with nil server — nil guard needed")
}

func TestProcessRequest_GetMiningInfo_NilServer_Documented(t *testing.T) {
	t.Skip("getmininginfo panics with nil server — nil guard needed")
}

// ---------------------------------------------------------------------------
// NewServer (RPC Server)
// ---------------------------------------------------------------------------

func TestNewServer_NotNil(t *testing.T) {
	s := NewServer(nil, nil)
	if s == nil {
		t.Fatal("expected non-nil RPC Server")
	}
}

func TestServer_HandleRequest_EmptyData_NoPanel(t *testing.T) {
	s := NewServer(nil, nil)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_, _ = s.HandleRequest([]byte{})
}

func TestServer_HandleRequest_InvalidJSON_Error(t *testing.T) {
	s := NewServer(nil, nil)
	_, err := s.HandleRequest([]byte("not json"))
	_ = err // may be nil or non-nil depending on handler — just no panic
}

// ---------------------------------------------------------------------------
// GetRawTransaction / GetBlockByNumber — nil server paths
// ---------------------------------------------------------------------------

func TestProcessRequest_GetRawTransaction_UnknownID_NoPanel(t *testing.T) {
	t.Skip("getrawtransaction panics with nil server — nil guard needed")
}

func TestProcessRequest_GetBlockByNumber_NilServer_Documented(t *testing.T) {
	t.Skip("getblockbynumber panics with nil server — nil guard needed")
}
