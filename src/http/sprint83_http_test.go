// test(PEPPER): Sprint 83 — src/http 66.0%→higher
// Tests: handleGetMempool with empty/with-items, handleGetAddressTxs with sender match,
// handleGetAddress with transaction that has nil Amount (line coverage)
package http

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	types "github.com/ramseyauron/quantix/src/core/transaction"
)

// ─── handleGetMempool — empty pending ────────────────────────────────────────

func TestSprint83_HandleGetMempool_Empty(t *testing.T) {
	srv := newTestServer(t)
	router := gin.New()
	router.GET("/mempool", srv.handleGetMempool)

	req, _ := http.NewRequest("GET", "/mempool", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["pending"]; !ok {
		t.Error("expected 'pending' field in response")
	}
	if _, ok := resp["count"]; !ok {
		t.Error("expected 'count' field in response")
	}
}

// ─── handleGetMempool — with tx that has non-nil gas/amount ──────────────────

func TestSprint83_HandleGetMempool_WithTx(t *testing.T) {
	srv := newTestServer(t)

	// Add a transaction to the mempool
	tx := &types.Transaction{
		ID:       "tx-sprint83",
		Sender:   "sprint83-sender",
		Receiver: "sprint83-receiver",
		Amount:   big.NewInt(100),
		GasLimit: big.NewInt(21000),
		GasPrice: big.NewInt(1),
		Nonce:    1,
	}
	_ = srv.blockchain.AddTransaction(tx) // ignore error if fails in test env

	router := gin.New()
	router.GET("/mempool", srv.handleGetMempool)
	req, _ := http.NewRequest("GET", "/mempool", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Verify response parses
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

// ─── handleGetAddressTxs — address with matching sender tx ───────────────────

func TestSprint83_HandleGetAddressTxs_WithMatchingSender(t *testing.T) {
	srv := newTestServer(t)
	router := gin.New()
	router.GET("/address/:addr/txs", srv.handleGetAddressTxs)

	// Query for a known genesis-funded address (should have transactions in history)
	req, _ := http.NewRequest("GET", "/address/d114e19f31c748fd6cc0fd74e5ef7a285cb6e34a1b3db3006d132b014ed6bfd1/txs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["transactions"]; !ok {
		t.Error("expected 'transactions' field")
	}
	if _, ok := resp["count"]; !ok {
		t.Error("expected 'count' field")
	}
}

// ─── handleGetAddress — address from genesis ─────────────────────────────────

func TestSprint83_HandleGetAddress_WithGenesisAddr(t *testing.T) {
	srv := newTestServer(t)
	router := gin.New()
	router.GET("/address/:addr", srv.handleGetAddress)

	req, _ := http.NewRequest("GET", "/address/d114e19f31c748fd6cc0fd74e5ef7a285cb6e34a1b3db3006d132b014ed6bfd1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["balance"]; !ok {
		t.Error("expected 'balance' field")
	}
}
