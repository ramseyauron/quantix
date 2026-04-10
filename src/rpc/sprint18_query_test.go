package rpc

// Coverage Sprint 18 — rpc/query.go remaining: AddJoin, AddPing, AddGet,
// GetQuery, RemoveQuery, IsExpectedResponse, GC.

import (
	"testing"
)

// ---------------------------------------------------------------------------
// QueryManager — AddJoin, AddPing, AddGet
// ---------------------------------------------------------------------------

func TestQueryManager_AddJoin_NoError(t *testing.T) {
	qm := NewQueryManager()
	id := GetRPCID()
	qm.AddJoin(id)
	// Verify IsExpectedResponse recognizes it.
	var sender NodeID
	msg := Message{
		RPCType: RPCJoin,
		Query:   false,
		RPCID:   id,
		From:    Remote{NodeID: sender},
	}
	if !qm.IsExpectedResponse(msg) {
		t.Error("IsExpectedResponse: should return true for added join")
	}
}

func TestQueryManager_AddPing_NoError(t *testing.T) {
	qm := NewQueryManager()
	id := GetRPCID()
	var target NodeID
	target[0] = 0x11
	qm.AddPing(id, target)

	// Verify IsExpectedResponse recognizes ping from target.
	msg := Message{
		RPCType: RPCPing,
		Query:   false,
		RPCID:   id,
		From:    Remote{NodeID: target},
	}
	if !qm.IsExpectedResponse(msg) {
		t.Error("IsExpectedResponse: should return true for expected ping response")
	}
}

func TestQueryManager_AddGet_NoError(t *testing.T) {
	qm := NewQueryManager()
	id := GetRPCID()
	qm.AddGet(id)

	var sender NodeID
	msg := Message{
		RPCType: RPCGet,
		Query:   false,
		RPCID:   id,
		From:    Remote{NodeID: sender},
	}
	if !qm.IsExpectedResponse(msg) {
		t.Error("IsExpectedResponse: should return true for added get")
	}
}

// ---------------------------------------------------------------------------
// GetQuery / RemoveQuery
// ---------------------------------------------------------------------------

func TestQueryManager_GetQuery_Found(t *testing.T) {
	qm := NewQueryManager()
	id := GetRPCID()
	var target NodeID
	qm.AddFindNode(id, target, nil)

	q, ok := qm.GetQuery(id)
	if !ok {
		t.Error("GetQuery: expected true for added query")
	}
	if q == nil {
		t.Error("GetQuery: returned nil query")
	}
}

func TestQueryManager_GetQuery_NotFound(t *testing.T) {
	qm := NewQueryManager()
	_, ok := qm.GetQuery(GetRPCID())
	if ok {
		t.Error("GetQuery: expected false for non-existent query")
	}
}

func TestQueryManager_RemoveQuery_Removes(t *testing.T) {
	qm := NewQueryManager()
	id := GetRPCID()
	var target NodeID
	qm.AddFindNode(id, target, nil)

	qm.RemoveQuery(id)

	_, ok := qm.GetQuery(id)
	if ok {
		t.Error("GetQuery: should return false after RemoveQuery")
	}
}

func TestQueryManager_RemoveQuery_NonExistent_NoError(t *testing.T) {
	qm := NewQueryManager()
	// Should not panic.
	qm.RemoveQuery(GetRPCID())
}

// ---------------------------------------------------------------------------
// IsExpectedResponse
// ---------------------------------------------------------------------------

func TestIsExpectedResponse_QueryMessage_False(t *testing.T) {
	qm := NewQueryManager()
	msg := Message{Query: true, RPCType: RPCPing}
	if qm.IsExpectedResponse(msg) {
		t.Error("IsExpectedResponse: should return false for query (not response)")
	}
}

func TestIsExpectedResponse_UnknownRPCID_False(t *testing.T) {
	qm := NewQueryManager()
	msg := Message{
		Query:   false,
		RPCType: RPCFindNode,
		RPCID:   GetRPCID(), // not registered
	}
	if qm.IsExpectedResponse(msg) {
		t.Error("IsExpectedResponse: should return false for unregistered RPCID")
	}
}

func TestIsExpectedResponse_UnknownType_False(t *testing.T) {
	qm := NewQueryManager()
	msg := Message{
		Query:   false,
		RPCType: RPCType(99), // unknown type
		RPCID:   GetRPCID(),
	}
	if qm.IsExpectedResponse(msg) {
		t.Error("IsExpectedResponse: should return false for unknown RPCType")
	}
}

func TestIsExpectedResponse_StoreType_Recognized(t *testing.T) {
	qm := NewQueryManager()
	id := GetRPCID()
	qm.AddGet(id)

	var sender NodeID
	msg := Message{
		RPCType: RPCStore,
		Query:   false,
		RPCID:   id,
		From:    Remote{NodeID: sender},
	}
	if !qm.IsExpectedResponse(msg) {
		t.Error("IsExpectedResponse: RPCStore should use same path as RPCGet")
	}
}

// ---------------------------------------------------------------------------
// GC — no panic on empty / populated manager
// ---------------------------------------------------------------------------

func TestQueryManager_GC_Empty_NoError(t *testing.T) {
	qm := NewQueryManager()
	qm.GC() // should not panic on empty manager
}

func TestQueryManager_GC_WithEntries_NoError(t *testing.T) {
	qm := NewQueryManager()
	id1, id2, id3 := GetRPCID(), GetRPCID(), GetRPCID()
	var target NodeID

	qm.AddFindNode(id1, target, nil)
	qm.AddJoin(id2)
	qm.AddPing(id3, target)

	qm.GC() // should not panic
}
