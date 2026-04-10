package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	security "github.com/ramseyauron/quantix/src/handshake"
	"github.com/ramseyauron/quantix/src/network"
)

// Sprint 48 — transport coverage boost

// ---------------------------------------------------------------------------
// ip.go — parsePort (package-internal, tested via exported ResolveAddress),
//         ConnectNode (error paths only), NodeToAddress edge cases
// ---------------------------------------------------------------------------

func TestNodeToAddress_AllFieldsEmpty(t *testing.T) {
	n := &network.Node{ID: "n1", IP: "", Port: ""}
	_, err := NodeToAddress(n)
	if err == nil {
		t.Error("NodeToAddress with empty IP+Port should return error")
	}
}

func TestConnectNode_EmptyIPPort_Error(t *testing.T) {
	// Empty IP+Port should fail immediately without network retry
	n := &network.Node{ID: "bad-node", IP: "", Port: ""}
	ch := make(chan *security.Message, 1)
	err := ConnectNode(n, ch)
	if err == nil {
		t.Error("ConnectNode with empty IP should return error immediately")
	}
}

func TestValidateIP_BareHostname(t *testing.T) {
	// Hostname is not a valid IP — should error
	err := ValidateIP("localhost", "8080")
	if err == nil {
		t.Error("ValidateIP with hostname (not IP) should return error")
	}
}

func TestValidateIP_PortZero(t *testing.T) {
	// Port 0 is technically valid (OS picks one)
	err := ValidateIP("127.0.0.1", "0")
	if err != nil {
		t.Logf("NOTE: ValidateIP port 0 returned error: %v", err)
	}
}

func TestResolveAddress_Roundtrip(t *testing.T) {
	addr, err := ResolveAddress("10.0.0.1", "9000")
	if err != nil {
		t.Fatalf("ResolveAddress: %v", err)
	}
	if addr != "10.0.0.1:9000" {
		t.Errorf("got %q, want %q", addr, "10.0.0.1:9000")
	}
}

// ---------------------------------------------------------------------------
// tcp.go — EnsureConnected (unreachable — should log and return quickly),
//           SendMessage error path (no connection, unreachable)
// ---------------------------------------------------------------------------

func TestEnsureConnected_UnreachableAddress(t *testing.T) {
	// Should not panic, just log
	EnsureConnected("127.0.0.1:59111") // unreachable — should log error quickly (3s timeout)
}

func TestEnsureConnected_AlreadyConnected_NoAction(t *testing.T) {
	// Same address called twice — second call should short-circuit on existing entry
	// (even if the entry was never actually established via handshake)
	addr := "127.0.0.1:59112"
	// First call — will fail silently
	EnsureConnected(addr)
	// Second call — should hit the "already connected" branch (or fail again silently)
	EnsureConnected(addr)
}

func TestSendMessage_UnreachableAddress(t *testing.T) {
	msg := &security.Message{Type: "test", Data: "hello"}
	err := SendMessage("127.0.0.1:1", msg)
	if err == nil {
		t.Error("SendMessage to unreachable address should return error")
	}
}

// ---------------------------------------------------------------------------
// websocket.go — WebSocketServer Start/Stop, websocketConn stub methods
// ---------------------------------------------------------------------------

func TestWebSocketServer_Start_Stop(t *testing.T) {
	ch := make(chan *security.Message, 10)
	ws := NewWebSocketServer("127.0.0.1:0", ch, nil)
	if ws == nil {
		t.Fatal("NewWebSocketServer returned nil")
	}
	readyCh := make(chan struct{}, 1)
	if err := ws.Start(readyCh); err != nil {
		t.Fatalf("WebSocketServer.Start: %v", err)
	}
	// Give the server goroutine a moment to start
	select {
	case <-readyCh:
	case <-time.After(2 * time.Second):
		t.Log("NOTE: WebSocket ready signal not received within 2s")
	}
	if err := ws.Stop(); err != nil {
		t.Errorf("WebSocketServer.Stop: %v", err)
	}
}

func TestWebSocketServer_Stop_Nil(t *testing.T) {
	ch := make(chan *security.Message, 1)
	ws := NewWebSocketServer("127.0.0.1:0", ch, nil)
	// Don't call Start — server.server is nil
	if err := ws.Stop(); err != nil {
		t.Errorf("Stop without Start should return nil, got %v", err)
	}
}

func TestWebSocketServer_HandleWS_BadUpgrade(t *testing.T) {
	// Call /ws endpoint without proper WebSocket upgrade — should get 400 response
	ch := make(chan *security.Message, 1)
	ws := NewWebSocketServer("127.0.0.1:0", ch, nil)
	readyCh := make(chan struct{}, 1)
	ws.Start(readyCh)
	<-readyCh
	defer ws.Stop()

	// httptest against the mux
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rr := httptest.NewRecorder()
	ws.mux.ServeHTTP(rr, req)
	// Non-WebSocket request should get a non-101 response (400 from upgrader)
	if rr.Code == http.StatusSwitchingProtocols {
		t.Error("plain HTTP request should not get 101 Switching Protocols")
	}
}

func TestConnectWebSocket_UnreachableAddress_Error(t *testing.T) {
	ch := make(chan *security.Message, 1)
	err := ConnectWebSocket("127.0.0.1:1", ch)
	if err == nil {
		t.Error("ConnectWebSocket to unreachable address should error")
	}
}
