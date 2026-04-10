// Sprint 20b — transport package: NewTCPServer construction, globalServer error paths,
// NewWebSocketServer construction, parsePort via ValidateIP.
package transport_test

import (
	"testing"

	security "github.com/ramseyauron/quantix/src/handshake"
	"github.com/ramseyauron/quantix/src/transport"
)

// ─── NewTCPServer construction ────────────────────────────────────────────────

func TestSprint20_NewTCPServer_NotNil(t *testing.T) {
	messageCh := make(chan *security.Message, 1)
	readyCh := make(chan struct{}, 1)
	srv := transport.NewTCPServer("127.0.0.1:0", messageCh, nil, readyCh)
	if srv == nil {
		t.Fatal("NewTCPServer returned nil")
	}
}

func TestSprint20_NewTCPServer_NoReadyCh(t *testing.T) {
	messageCh := make(chan *security.Message, 1)
	srv := transport.NewTCPServer("127.0.0.1:0", messageCh, nil, nil)
	if srv == nil {
		t.Fatal("NewTCPServer with nil readyCh returned nil")
	}
}

// ─── GetConnection error path ─────────────────────────────────────────────────

func TestSprint20_GetConnection_Unknown_Error(t *testing.T) {
	_, err := transport.GetConnection("127.0.0.1:19999")
	if err == nil {
		t.Error("expected error for unknown connection, got nil")
	}
}

// ─── GetEncryptionKey error path ──────────────────────────────────────────────

func TestSprint20_GetEncryptionKey_Unknown_Error(t *testing.T) {
	_, err := transport.GetEncryptionKey("127.0.0.1:29999")
	if err == nil {
		t.Error("expected error for unknown encryption key, got nil")
	}
}

// ─── NewWebSocketServer construction ─────────────────────────────────────────

func TestSprint20_NewWebSocketServer_NotNil(t *testing.T) {
	messageCh := make(chan *security.Message, 1)
	srv := transport.NewWebSocketServer("127.0.0.1:0", messageCh, nil)
	if srv == nil {
		t.Fatal("NewWebSocketServer returned nil")
	}
}

// ─── BroadcastToAll with no connections ──────────────────────────────────────

func TestSprint20_BroadcastToAll_NoConnections_EmptyErrors(t *testing.T) {
	msg := &security.Message{Type: "ping"}
	errs := transport.BroadcastToAll(msg)
	// No active connections → no errors (empty slice)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}
