// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 32 - transport Stop/Disconnect/ConnectTCP error paths
package transport_test

import (
	"testing"

	security "github.com/ramseyauron/quantix/src/handshake"
	"github.com/ramseyauron/quantix/src/network"
	"github.com/ramseyauron/quantix/src/transport"
)

// ---------------------------------------------------------------------------
// TCPServer.Stop — nil listener should return nil, not panic
// ---------------------------------------------------------------------------

func TestSprint32_TCPServer_Stop_NilListener_ReturnsNil(t *testing.T) {
	s := transport.NewTCPServer(":0", nil, nil, nil)
	err := s.Stop()
	if err != nil {
		t.Fatalf("expected nil error when stopping server with nil listener, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// WebSocketServer.Stop — nil server should not panic
// ---------------------------------------------------------------------------

func TestSprint32_WebSocketServer_Stop_NilServer_NoPanel(t *testing.T) {
	ws := transport.NewWebSocketServer(":0", nil, nil)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("WebSocketServer.Stop panicked: %v", r)
		}
	}()
	_ = ws.Stop()
}

// ---------------------------------------------------------------------------
// ConnectTCP — empty address returns error immediately
// ---------------------------------------------------------------------------

func TestSprint32_ConnectTCP_EmptyAddress_ReturnsError(t *testing.T) {
	_, err := transport.ConnectTCP("", nil)
	if err == nil {
		t.Fatal("expected error for empty address")
	}
}

func TestSprint32_ConnectTCP_UnreachableAddress_ReturnsError(t *testing.T) {
	_, err := transport.ConnectTCP("127.0.0.1:19999", nil)
	if err == nil {
		t.Fatal("expected error for unreachable address")
	}
}

// ---------------------------------------------------------------------------
// DisconnectNode — node with no active connection returns error
// ---------------------------------------------------------------------------

func TestSprint32_DisconnectNode_NoConnection_ReturnsError(t *testing.T) {
	node := &network.Node{
		IP:      "127.0.0.1",
		Port:    "9999",
		Address: "127.0.0.1:9999",
	}
	err := transport.DisconnectNode(node)
	if err == nil {
		t.Fatal("expected error when no active connection exists")
	}
}

func TestSprint32_DisconnectNode_EmptyIPPort_ReturnsError(t *testing.T) {
	node := &network.Node{
		IP:   "",
		Port: "",
	}
	err := transport.DisconnectNode(node)
	if err == nil {
		t.Fatal("expected error for node with empty IP and Port")
	}
}

// ---------------------------------------------------------------------------
// SendMessage — unknown address returns error (no connection, dial fails fast)
// ---------------------------------------------------------------------------

func TestSprint32_SendMessage_UnknownAddress_ReturnsError(t *testing.T) {
	msg := &security.Message{Type: "test"}
	err := transport.SendMessage("127.0.0.1:19998", msg)
	if err == nil {
		t.Fatal("expected error for unreachable address")
	}
}
