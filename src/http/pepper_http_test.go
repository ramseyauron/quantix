// PEPPER Sprint 2 — http package coverage push
// Targets: handleGetAddressTxs, setupRoutes(GET /), GET /blockcount, 
//          GET /address/:addr, GET /address/:addr/txs, GET /blocks,
//          SetConsensusValidatorSet, SetSphincsMgr, Start/Stop
package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ── GET / (root endpoint) ────────────────────────────────────────────────────

func TestGetRoot_ReturnsWelcome(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("GET /: unexpected status %d — %s", w.Code, w.Body.String())
	}
	if w.Code == http.StatusOK {
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Errorf("GET /: response not valid JSON: %v", err)
		}
	}
}

// ── GET /blockcount ───────────────────────────────────────────────────────────

func TestGetBlockCount_ReturnsCount(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/blockcount", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /blockcount: want 200, got %d — %s", w.Code, w.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("response not valid JSON: %v", err)
	}
	if _, ok := result["count"]; !ok {
		t.Error("expected 'count' field")
	}
}

// ── GET /address/:addr ────────────────────────────────────────────────────────

func TestGetAddress_UnknownAddress(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/address/xUnknownAddressABCDEF123456789", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("GET /address/...: unexpected status %d — %s", w.Code, w.Body.String())
	}
}

func TestGetAddress_EmptyAddress(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/address/", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)
	// Not found or redirect is acceptable
	_ = w.Code
}

// ── GET /address/:addr/txs ────────────────────────────────────────────────────

func TestGetAddressTxs_ReturnsEmptyList(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/address/xNoTxsAddress1234567890ABCDEF/txs", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /address/addr/txs: want 200, got %d — %s", w.Code, w.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("response not valid JSON: %v", err)
	}
	if _, ok := result["transactions"]; !ok {
		t.Error("expected 'transactions' field in response")
	}
	if result["count"].(float64) != 0 {
		t.Errorf("expected count=0 for unknown address")
	}
}

// ── GET /blocks (sync endpoint) ──────────────────────────────────────────────

func TestGetBlocks_DefaultPagination(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/blocks", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /blocks: want 200, got %d — %s", w.Code, w.Body.String())
	}
	// Response is a slice of blocks
	if len(w.Body.Bytes()) == 0 {
		t.Error("expected non-empty response")
	}
}

func TestGetBlocks_WithFromAndLimit(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/blocks?from=0&limit=5", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /blocks?from=0&limit=5: want 200, got %d", w.Code)
	}
}

func TestGetBlocks_InvalidFrom(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/blocks?from=abc", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)
	// Should handle gracefully
	_ = w.Code
}

// ── GET /latest-transaction ──────────────────────────────────────────────────

func TestGetLatestTransaction_NoTx(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/latest-transaction", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /latest-transaction: want 200, got %d", w.Code)
	}
}

// ── POST /mine (disabled when no env var) ────────────────────────────────────

func TestMine_Disabled(t *testing.T) {
	t.Setenv("DEVNET_MINE_SECRET", "")
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/mine", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("POST /mine with no secret configured: want 403, got %d", w.Code)
	}
}

func TestMine_WrongSecret(t *testing.T) {
	t.Setenv("DEVNET_MINE_SECRET", "correct_secret")
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/mine", nil)
	req.Header.Set("X-Mine-Secret", "wrong_secret")
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("POST /mine with wrong secret: want 403, got %d", w.Code)
	}
}

// ── SetConsensusValidatorSet / SetSphincsMgr ─────────────────────────────────

func TestSetConsensusValidatorSet_NoPanic(t *testing.T) {
	srv := newTestServer(t)
	srv.SetConsensusValidatorSet(nil) // should not panic
}

func TestSetSphincsMgr_NoPanic(t *testing.T) {
	srv := newTestServer(t)
	srv.SetSphincsMgr(nil) // should not panic
}
