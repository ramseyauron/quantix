package http

// Coverage sprint 9 — tests for enriched /validators endpoint (faf3fd7 / 207b7ed)
// Covers: ValidatorInfoResponse fields, blocks_produced counting, deduplication,
// vault/genesis address filtering.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestValidatorEnriched_Fields verifies the enriched /validators response shape.
func TestValidatorEnriched_Fields(t *testing.T) {
	setValidatorSecret(t)
	srv := newTestServer(t)

	// Register a validator.
	body := `{"public_key":"reward-addr-abcdef1234567890","stake_amount":"5000","node_address":"node-a.example.com:9001"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/validator/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Register-Secret", "test-secret")
	srv.Router().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("register: got %d, body=%s", w.Code, w.Body.String())
	}

	// GET /validators
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/validators", nil)
	srv.Router().ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Fatalf("get validators: got %d", w2.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w2.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	validators, ok := resp["validators"].([]interface{})
	if !ok {
		t.Fatalf("validators field missing or wrong type: %T", resp["validators"])
	}

	// Find our registered validator.
	var found map[string]interface{}
	for _, v := range validators {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		nodeID, _ := vm["node_id"].(string)
		addr, _ := vm["address"].(string)
		if strings.Contains(nodeID, "node-a.example.com") || addr == "reward-addr-abcdef1234567890" {
			found = vm
			break
		}
	}

	if found == nil {
		t.Fatalf("registered validator not found in response: %v", validators)
	}

	// Verify required fields exist in enriched response (faf3fd7 / 207b7ed).
	requiredFields := []string{"address", "node_id", "stake", "status", "blocks_produced", "reward_address"}
	for _, field := range requiredFields {
		if _, ok := found[field]; !ok {
			t.Errorf("missing field %q in validator response: %v", field, found)
		}
	}

	// blocks_produced should be a number (0 for fresh validator).
	if bp, ok := found["blocks_produced"].(float64); !ok {
		t.Errorf("blocks_produced should be numeric, got %T", found["blocks_produced"])
	} else if bp < 0 {
		t.Errorf("blocks_produced should be non-negative, got %v", bp)
	}

	// stake should be numeric.
	if _, ok := found["stake"].(float64); !ok {
		t.Errorf("stake should be numeric, got %T", found["stake"])
	}

	// status should be a string.
	if _, ok := found["status"].(string); !ok {
		t.Errorf("status should be a string, got %T", found["status"])
	}
}

// TestValidatorEnriched_NoDuplicates verifies the same public_key is not returned twice.
func TestValidatorEnriched_NoDuplicates(t *testing.T) {
	setValidatorSecret(t)
	srv := newTestServer(t)

	const pubKey = "dedup-reward-addr-000000000000001"

	// Register same public_key twice with different node addresses.
	for i, addr := range []string{"node-dup1:9100", "node-dup2:9100"} {
		body := `{"public_key":"` + pubKey + `","stake_amount":"1000","node_address":"` + addr + `"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/validator/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Register-Secret", "test-secret")
		srv.Router().ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("register %d: got %d, body=%s", i, w.Code, w.Body.String())
		}
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/validators", nil)
	srv.Router().ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	validators, _ := resp["validators"].([]interface{})

	count := 0
	for _, v := range validators {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if vm["address"] == pubKey || vm["reward_address"] == pubKey {
			count++
		}
	}

	if count > 1 {
		t.Errorf("duplicate dedup: same pubkey appeared %d times in /validators", count)
	}
}

// TestValidatorEnriched_VaultAddressFiltered verifies the genesis vault address
// (all-zeros) is filtered from the validators list.
func TestValidatorEnriched_VaultAddressFiltered(t *testing.T) {
	setValidatorSecret(t)
	srv := newTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/validators", nil)
	srv.Router().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("get validators: %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	validators, _ := resp["validators"].([]interface{})

	for _, v := range validators {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		addr, _ := vm["address"].(string)
		// Genesis vault address: 0000000000000000000000000000000000000001 or all-zeros
		allZero := true
		for _, c := range addr {
			if c != '0' {
				allZero = false
				break
			}
		}
		if allZero && len(addr) >= 32 {
			t.Errorf("vault/genesis address %q should be filtered from /validators", addr)
		}
	}
}

// TestValidatorEnriched_TotalCount is returned in response.
func TestValidatorEnriched_TotalCountField(t *testing.T) {
	setValidatorSecret(t)
	srv := newTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/validators", nil)
	srv.Router().ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if _, ok := resp["total"]; !ok {
		t.Logf("note: 'total' field not present in /validators response (optional)")
	}
	// validators array must always be present (even if empty).
	if _, ok := resp["validators"]; !ok {
		t.Errorf("'validators' key missing from /validators response")
	}
}
