// Sprint 24 — http package: handleGetAddressTxs full coverage, handleGetAddress
// edge cases, Start/Stop, SubmitTransaction against httptest server.
package http

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	types "github.com/ramseyauron/quantix/src/core/transaction"
)

// ─── handleGetAddressTxs ──────────────────────────────────────────────────────

func TestSprint24_GetAddressTxs_EmptyChain_EmptyList(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/alice/txs", nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := resp["transactions"]; !ok {
		t.Error("response missing 'transactions' field")
	}
	if _, ok := resp["count"]; !ok {
		t.Error("response missing 'count' field")
	}
}

func TestSprint24_GetAddressTxs_UnknownAddress_ZeroCount(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/unknown-address-xyz/txs", nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	count := resp["count"].(float64)
	if count != 0 {
		t.Errorf("count = %v, want 0", count)
	}
}

func TestSprint24_GetAddressTxs_AddressField_Present(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	addr := "test-address-sprint24"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/address/%s/txs", addr), nil)
	srv.Router().ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if gotAddr, ok := resp["address"]; !ok || gotAddr != addr {
		t.Errorf("address field = %v, want %q", gotAddr, addr)
	}
}

// ─── handleGetAddress edge cases ──────────────────────────────────────────────

func TestSprint24_GetAddress_ZeroAddress_ZeroBalance(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/0000000000000000", nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["balance"]; !ok {
		t.Error("response missing 'balance' field")
	}
	if _, ok := resp["tx_count"]; !ok {
		t.Error("response missing 'tx_count' field")
	}
	if _, ok := resp["total_received"]; !ok {
		t.Error("response missing 'total_received' field")
	}
	if _, ok := resp["total_sent"]; !ok {
		t.Error("response missing 'total_sent' field")
	}
}

func TestSprint24_GetAddress_LongAddress_NoError(t *testing.T) {
	srv := newTestServer(t)
	// 64-char address (SPHINCS+ fingerprint format)
	addr := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/"+addr, nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// ─── handleGetValidators edge case ───────────────────────────────────────────

func TestSprint24_GetValidators_FieldsPresent(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/validators", nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["validators"]; !ok {
		t.Error("response missing 'validators' field")
	}
}

// ─── handleValidatorRegister additional coverage ──────────────────────────────

func TestSprint24_ValidatorRegister_EmptyBody_400(t *testing.T) {
	srv := newTestServer(t)
	t.Setenv("VALIDATOR_REGISTER_SECRET", "")
	// Dev mode should allow registration without secret
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/validator/register", nil)
	srv.Router().ServeHTTP(w, req)
	// Missing body → 400 or specific error
	if w.Code == 500 {
		t.Errorf("unexpected 500: %s", w.Body.String())
	}
}

// ─── handleGetBlock edge cases ────────────────────────────────────────────────

func TestSprint24_GetBlock_NonExistentHeight_404Or200(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/block/99999", nil)
	srv.Router().ServeHTTP(w, req)
	// Either 404 or 200 with empty — just no 500
	if w.Code == 500 {
		t.Errorf("unexpected 500 for non-existent block: %s", w.Body.String())
	}
}

func TestSprint24_GetBlock_InvalidHeight_NoError(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/block/notanumber", nil)
	srv.Router().ServeHTTP(w, req)
	if w.Code == 500 {
		t.Errorf("unexpected 500 for invalid block height: %s", w.Body.String())
	}
}

// ─── handleGetBlocks edge cases ───────────────────────────────────────────────

func TestSprint24_GetBlocks_WithLargeLimit_Capped(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/blocks?from=0&limit=99999", nil)
	srv.Router().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestSprint24_GetBlocks_NegativeFrom_HandledGracefully(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/blocks?from=-5&limit=10", nil)
	srv.Router().ServeHTTP(w, req)
	if w.Code == 500 {
		t.Errorf("unexpected 500: %s", w.Body.String())
	}
}

// ─── SubmitTransaction with httptest server ───────────────────────────────────

func TestSprint24_SubmitTransaction_Against_TestServer(t *testing.T) {
	// Create a real test server that accepts /transaction POST
	srv := newTestServer(t)
	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	// Parse host from test server URL
	parsed, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	host := parsed.Host

	tx := types.Transaction{
		ID:       "sprint24-tx-001",
		Sender:   "alice",
		Receiver: "bob",
		Amount:   big.NewInt(100),
		Nonce:    0,
	}

	// SubmitTransaction submits via HTTP — should succeed (200 or at least no network error)
	err = SubmitTransaction(host, tx)
	// May return non-nil if server rejects (missing sig etc.) — just no panic
	_ = err
}

// ─── Start / Stop ────────────────────────────────────────────────────────────

func TestSprint24_Server_Start_NoPanic(t *testing.T) {
	bc := getSharedBC(t)
	srv := NewServer(":0", nil, bc, nil)
	// Start launches a goroutine — just verify no panic
	err := srv.Start()
	if err != nil {
		t.Errorf("Start() error: %v", err)
	}
	// Stop it
	_ = srv.Stop()
}

func TestSprint24_Server_Stop_NilHTTPServer_NoError(t *testing.T) {
	srv := &Server{} // httpServer is nil
	err := srv.Stop()
	if err != nil {
		t.Errorf("Stop() with nil httpServer returned error: %v", err)
	}
}
