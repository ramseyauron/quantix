package rpc

import (
	"encoding/json"
	"testing"
)

// Sprint 50 — rpc json handler: errorResponse, mapRPCTypeToMethod, RPCType.String,
//             processBatchRequest, processSingleRequest error paths

// ---------------------------------------------------------------------------
// errorResponse — pure JSON, no server needed
// ---------------------------------------------------------------------------

func TestErrorResponse_ValidID(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	data, err := h.errorResponse(1, ErrCodeParseError, "parse error")
	if err != nil {
		t.Fatalf("errorResponse: %v", err)
	}
	var resp JSONRPCResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error field in response")
	}
	if resp.Error.Code != ErrCodeParseError {
		t.Errorf("code: got %d, want %d", resp.Error.Code, ErrCodeParseError)
	}
}

func TestErrorResponse_NilID(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	data, err := h.errorResponse(nil, ErrCodeMethodNotFound, "not found")
	if err != nil {
		t.Fatalf("errorResponse nil ID: %v", err)
	}
	if len(data) == 0 {
		t.Error("errorResponse nil ID should produce non-empty JSON")
	}
}

func TestErrorResponse_StringID(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	data, err := h.errorResponse("req-1", ErrCodeInvalidParams, "bad params")
	if err != nil {
		t.Fatalf("errorResponse string ID: %v", err)
	}
	var resp JSONRPCResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error.Message != "bad params" {
		t.Errorf("message: got %q, want %q", resp.Error.Message, "bad params")
	}
}

// ---------------------------------------------------------------------------
// mapRPCTypeToMethod — pure switch, no server needed
// ---------------------------------------------------------------------------

func TestMapRPCTypeToMethod_AllKnownTypes(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	knownTypes := []RPCType{
		RPCGetBlockCount, RPCGetBestBlockHash, RPCGetBlock, RPCGetBlocks,
		RPCSendRawTransaction, RPCGetTransaction, RPCPing, RPCJoin,
		RPCFindNode, RPCGet, RPCStore, RPCGetBlockByNumber, RPCGetBlockHash,
		RPCGetDifficulty, RPCGetChainTip, RPCGetNetworkInfo, RPCGetMiningInfo,
		RPCEstimateFee, RPCGetMemPoolInfo, RPCValidateAddress, RPCVerifyMessage,
		RPCGetRawTransaction,
	}
	for _, rt := range knownTypes {
		method, err := h.mapRPCTypeToMethod(rt)
		if err != nil {
			t.Errorf("mapRPCTypeToMethod(%d) error: %v", rt, err)
		}
		if method == "" {
			t.Errorf("mapRPCTypeToMethod(%d) returned empty method", rt)
		}
	}
}

func TestMapRPCTypeToMethod_Unknown_Error(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	_, err := h.mapRPCTypeToMethod(RPCType(127))
	if err == nil {
		t.Error("unknown RPCType should return error")
	}
}

// ---------------------------------------------------------------------------
// RPCType.String — pure switch, no server needed
// ---------------------------------------------------------------------------

func TestRPCType_String_AllKnown(t *testing.T) {
	types := []RPCType{
		RPCGetBlockCount, RPCGetBestBlockHash, RPCGetBlock, RPCGetBlocks,
		RPCSendRawTransaction, RPCGetTransaction, RPCPing, RPCJoin,
		RPCFindNode, RPCGet, RPCStore, RPCGetBlockByNumber, RPCGetBlockHash,
		RPCGetDifficulty, RPCGetChainTip, RPCGetNetworkInfo, RPCGetMiningInfo,
		RPCEstimateFee, RPCGetMemPoolInfo, RPCValidateAddress, RPCVerifyMessage,
		RPCGetRawTransaction,
	}
	for _, rt := range types {
		s := rt.String()
		if s == "" {
			t.Errorf("RPCType(%d).String() returned empty string", rt)
		}
	}
}

func TestRPCType_String_Unknown(t *testing.T) {
	s := RPCType(127).String()
	if s == "" {
		t.Error("unknown RPCType.String() should return non-empty fallback")
	}
}

// ---------------------------------------------------------------------------
// ProcessRequest — errorResponse path (neither single nor batch JSON-RPC,
//                  and not a binary Message)
// ---------------------------------------------------------------------------

func TestProcessRequest_ParseError_Path(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	// Send garbage that's neither binary Message nor JSON — goes to errorResponse
	resp, err := h.ProcessRequest([]byte("not-json-nor-binary"))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	var errResp JSONRPCResponse
	if err := json.Unmarshal(resp, &errResp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if errResp.Error == nil {
		t.Error("expected JSON-RPC error in response")
	}
}

func TestProcessRequest_BatchEmptyArray_ErrorResponse(t *testing.T) {
	h := NewJSONRPCHandler(nil)
	// Empty JSON array — batch with 0 elements
	resp, err := h.ProcessRequest([]byte("[]"))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	// Should return some kind of response (may be empty or error)
	_ = resp
}
