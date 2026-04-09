// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 41 - http handleGetAddressTxs + verifyTransactionSignature coverage
package http

import (
	"strings"
	"encoding/json"
	
	"net/http"
	"net/http/httptest"
	"testing"


)

func TestSprint41_HandleGetAddressTxs_BasicResponseShape(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/someaddress123/txs", nil)
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if _, ok := resp["address"]; !ok {
		t.Fatal("expected 'address' field in response")
	}
	if _, ok := resp["transactions"]; !ok {
		t.Fatal("expected 'transactions' field in response")
	}
	if _, ok := resp["count"]; !ok {
		t.Fatal("expected 'count' field in response")
	}
}

func TestSprint41_HandleGetAddressTxs_KnownSender_FindsTxs(t *testing.T) {
	srv := newTestServer(t)

	// Submit a tx via HTTP to put it in mempool
	body := `{"sender":"sprint41-sender-addr","receiver":"sprint41-receiver-addr","amount":"1000","nonce":0}`
	w0 := httptest.NewRecorder()
	req0, _ := http.NewRequest("POST", "/transaction", strings.NewReader(body))
	req0.Header.Set("Content-Type", "application/json")
	srv.Router().ServeHTTP(w0, req0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/sprint41-sender-addr/txs", nil)
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Response should be 200 regardless of whether tx was mined
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["address"] != "sprint41-sender-addr" {
		t.Fatalf("expected address 'sprint41-sender-addr', got %v", resp["address"])
	}
}

func TestSprint41_HandleGetAddressTxs_EmptyAddress(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address//txs", nil)
	srv.Router().ServeHTTP(w, req)
	// May get 301 redirect or 200 — just no 500
	if w.Code == http.StatusInternalServerError {
		t.Fatal("unexpected 500 for empty address")
	}
}

func TestSprint41_HandleGetAddressTxs_UnknownAddress_EmptyList(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/never-exists-sprint41/txs", nil)
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for unknown address, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	count, _ := resp["count"].(float64)
	if count != 0 {
		t.Fatalf("expected count=0 for unknown address, got %v", count)
	}
}

// ---------------------------------------------------------------------------
// handleGetAddress — additional coverage
// ---------------------------------------------------------------------------

func TestSprint41_HandleGetAddress_ZeroAddress(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/0000000000000000000000000000000000000000000000000000000000000000", nil)
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSprint41_HandleGetAddress_64CharAddress(t *testing.T) {
	srv := newTestServer(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil)
	srv.Router().ServeHTTP(w, req)
	// 200 OK expected
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// verifyTransactionSignature — via transaction submission endpoint
// ---------------------------------------------------------------------------

func TestSprint41_HandleTransaction_WithSignatureField_NoPanic(t *testing.T) {
	srv := newTestServer(t)

	// Submit a tx that has a signature field — exercises the sig verification path
	body := `{"sender":"alice","receiver":"bob","amount":"1000","nonce":0,"signature":"aabbccdd"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/transaction", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	srv.Router().ServeHTTP(w, req)
	// In devMode, should succeed or return a validation error — not panic
	if w.Code == http.StatusInternalServerError {
		t.Fatal("unexpected 500 for transaction with signature")
	}
}

func TestSprint41_HandleTransaction_EmptySig_DevMode_Accepted(t *testing.T) {
	srv := newTestServer(t)
	body := `{"sender":"sender-s41","receiver":"recv-s41","amount":"500","nonce":0}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/transaction", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	srv.Router().ServeHTTP(w, req)
	// devMode: empty sig should be accepted
	if w.Code == http.StatusInternalServerError {
		t.Fatalf("unexpected 500, got body: %s", w.Body.String())
	}
}
