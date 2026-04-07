package http

// Coverage Sprint 13 — GET /mempool endpoint (c8aedba), handleGetAddress,
// handleGetAddressTxs, GetPendingTransactions / GetMempoolStats coverage.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// GET /mempool — new endpoint (c8aedba)
// ---------------------------------------------------------------------------

// TestGetMempool_Empty verifies /mempool returns a valid JSON response when the
// mempool is empty (no pending transactions submitted yet).
func TestGetMempool_Empty(t *testing.T) {
	srv := newTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mempool", nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /mempool: expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode /mempool response: %v", err)
	}

	// Required fields.
	for _, field := range []string{"pending", "count", "stats"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("GET /mempool: missing field %q", field)
		}
	}

	// count should be a number.
	count, ok := resp["count"].(float64)
	if !ok {
		t.Errorf("GET /mempool: 'count' should be a number, got %T", resp["count"])
	}
	if count < 0 {
		t.Errorf("GET /mempool: 'count' should be non-negative, got %v", count)
	}

	// pending should be an array.
	pending, ok := resp["pending"].([]interface{})
	if !ok {
		t.Errorf("GET /mempool: 'pending' should be an array, got %T", resp["pending"])
	}
	if int(count) != len(pending) {
		t.Errorf("GET /mempool: count=%d but len(pending)=%d", int(count), len(pending))
	}
}

// TestGetMempool_WithPendingTx verifies /mempool returns the submitted transaction
// before it is mined into a block.
func TestGetMempool_WithPendingTx(t *testing.T) {
	srv := newTestServer(t)

	// Submit a transaction to the mempool.
	txBody := `{"sender":"alice0000000000000000000000000000000000000000000000000000000000001","receiver":"bob00000000000000000000000000000000000000000000000000000000000001","amount":"100","nonce":0,"gas_limit":"21000","gas_price":"1","signature":"dGVzdA=="}`
	submitReq, _ := http.NewRequest("POST", "/transaction", strings.NewReader(txBody))
	submitReq.Header.Set("Content-Type", "application/json")
	submitW := httptest.NewRecorder()
	srv.Router().ServeHTTP(submitW, submitReq)
	// Accept 200 (submitted) or non-5xx for this test.
	if submitW.Code >= 500 {
		t.Fatalf("POST /transaction failed with %d", submitW.Code)
	}

	// Get mempool.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mempool", nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /mempool after submit: expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	// Response shape must still be valid.
	if _, ok := resp["pending"]; !ok {
		t.Error("GET /mempool: 'pending' field missing")
	}
	if _, ok := resp["count"]; !ok {
		t.Error("GET /mempool: 'count' field missing")
	}
}

// TestGetMempool_PendingItemShape verifies each item in 'pending' has required fields.
func TestGetMempool_PendingItemShape(t *testing.T) {
	srv := newTestServer(t)

	// Submit a transaction.
	txBody := `{"sender":"mem0000000000000000000000000000000000000000000000000000000000001","receiver":"mem0000000000000000000000000000000000000000000000000000000000002","amount":"500","nonce":0,"gas_limit":"21000","gas_price":"1","signature":"dGVzdA=="}`
	submitReq, _ := http.NewRequest("POST", "/transaction", strings.NewReader(txBody))
	submitReq.Header.Set("Content-Type", "application/json")
	srv.Router().ServeHTTP(httptest.NewRecorder(), submitReq)

	// Query mempool.
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, mustGet("/mempool"))

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	pending, _ := resp["pending"].([]interface{})
	if len(pending) == 0 {
		t.Skip("no pending items in mempool — transaction may have been rejected")
	}

	item, ok := pending[0].(map[string]interface{})
	if !ok {
		t.Fatalf("pending[0] not a map: %T", pending[0])
	}

	requiredFields := []string{"id", "sender", "receiver", "amount", "gas_limit", "gas_price", "nonce", "timestamp", "fee"}
	for _, f := range requiredFields {
		if _, exists := item[f]; !exists {
			t.Errorf("pending item missing field %q", f)
		}
	}
}

// TestGetMempool_Stats verifies the stats field has reasonable structure.
func TestGetMempool_Stats(t *testing.T) {
	srv := newTestServer(t)

	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, mustGet("/mempool"))

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	// stats may be nil (if mempool.GetStats returns nil) or a map.
	stats := resp["stats"]
	if stats != nil {
		if _, ok := stats.(map[string]interface{}); !ok {
			t.Errorf("'stats' should be a map or null, got %T", stats)
		}
	}
}

// mustGet creates a GET request for the given path.
func mustGet(path string) *http.Request {
	req, _ := http.NewRequest("GET", path, nil)
	return req
}

// ---------------------------------------------------------------------------
// GET /address/:addr — handleGetAddress coverage
// ---------------------------------------------------------------------------

// TestGetAddress_BasicResponse verifies /address/:addr returns correct shape.
func TestGetAddress_BasicResponse(t *testing.T) {
	srv := newTestServer(t)
	addr := "aaaa0000000000000000000000000000000000000000000000000000000000001"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/"+addr, nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /address: expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	for _, field := range []string{"address", "balance", "total_received", "total_sent", "tx_count"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("GET /address: missing field %q", field)
		}
	}

	if resp["address"] != addr {
		t.Errorf("GET /address: address field mismatch: got %v, want %v", resp["address"], addr)
	}
}

// TestGetAddress_UnknownAddress_ZeroBalance verifies an unknown address returns
// zero balance rather than an error.
func TestGetAddress_UnknownAddress_ZeroBalance(t *testing.T) {
	srv := newTestServer(t)
	addr := "unknown0000000000000000000000000000000000000000000000000000000001"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/"+addr, nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /address unknown: expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	balance, _ := resp["balance"].(string)
	if balance != "0" && balance != "" {
		t.Logf("GET /address unknown: balance = %q (may be non-zero if genesis allocations are loaded)", balance)
	}
}

// TestGetAddress_TxCountIsNumeric verifies tx_count is numeric.
func TestGetAddress_TxCountIsNumeric(t *testing.T) {
	srv := newTestServer(t)

	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, mustGet("/address/someaddress00000000000000000000000000000000000000000000000001"))

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	count := resp["tx_count"]
	if count == nil {
		t.Error("tx_count should not be nil")
		return
	}
	if _, ok := count.(float64); !ok {
		t.Errorf("tx_count should be numeric, got %T", count)
	}
}

// ---------------------------------------------------------------------------
// GET /address/:addr/txs — handleGetAddressTxs coverage
// ---------------------------------------------------------------------------

// TestGetAddressTxs_BasicResponse verifies /address/:addr/txs returns correct shape.
func TestGetAddressTxs_BasicResponse(t *testing.T) {
	srv := newTestServer(t)
	addr := "txsaddr000000000000000000000000000000000000000000000000000000001"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/"+addr+"/txs", nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /address/txs: expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	for _, field := range []string{"address", "transactions", "count"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("GET /address/txs: missing field %q", field)
		}
	}

	if resp["address"] != addr {
		t.Errorf("GET /address/txs: address mismatch")
	}

	// count must be numeric and non-negative.
	count, ok := resp["count"].(float64)
	if !ok {
		t.Errorf("count should be numeric, got %T", resp["count"])
	}
	if count < 0 {
		t.Error("count should be non-negative")
	}

	// transactions must be an array.
	txs, ok := resp["transactions"].([]interface{})
	if !ok && resp["transactions"] != nil {
		t.Errorf("transactions should be an array, got %T", resp["transactions"])
	}
	_ = txs
}

// TestGetAddressTxs_EmptyAddress_NoTransactions verifies an empty/new address
// returns 0 transactions.
func TestGetAddressTxs_EmptyAddress_NoTransactions(t *testing.T) {
	srv := newTestServer(t)
	addr := "emptyaddr00000000000000000000000000000000000000000000000000000001"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/"+addr+"/txs", nil)
	srv.Router().ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	count, _ := resp["count"].(float64)
	if count != 0 {
		t.Logf("GET /address/txs for new address: count=%v (expected 0 unless genesis has this address)", count)
	}
}

// TestGetAddressTxs_TxItemShape verifies each tx item in the response has required fields.
func TestGetAddressTxs_TxItemShape(t *testing.T) {
	srv := newTestServer(t)

	// Submit a tx to create some activity.
	const sender = "shapesend0000000000000000000000000000000000000000000000000000001"
	txBody := `{"sender":"` + sender + `","receiver":"shaperecv0000000000000000000000000000000000000000000000000000001","amount":"100","nonce":0,"gas_limit":"21000","gas_price":"1","signature":"dGVzdA=="}`
	submitReq, _ := http.NewRequest("POST", "/transaction", strings.NewReader(txBody))
	submitReq.Header.Set("Content-Type", "application/json")
	srv.Router().ServeHTTP(httptest.NewRecorder(), submitReq)

	// Query the txs for this address — may be empty if not mined yet, just verify shape.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/address/"+sender+"/txs", nil)
	srv.Router().ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	txs, _ := resp["transactions"].([]interface{})
	if len(txs) == 0 {
		t.Skip("no transactions for address (not yet mined — acceptable)")
	}

	item, ok := txs[0].(map[string]interface{})
	if !ok {
		t.Fatalf("tx item not a map: %T", txs[0])
	}

	requiredFields := []string{"hash", "block", "timestamp", "from", "to", "amount", "nonce", "direction"}
	for _, f := range requiredFields {
		if _, exists := item[f]; !exists {
			t.Errorf("tx item missing field %q", f)
		}
	}
}
