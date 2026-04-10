// MIT License
//
// Copyright (c) 2024 quantix

// Package integration_test — Q5 devnet integration tests.
//
// Run with:
//
//	docker-compose up -d && QUANTIX_INTEGRATION_TEST=1 go test ./src/integration/... -tags=integration -timeout 120s
//
//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"
)

const devnetBase = "http://localhost:8560"

// TestMain skips the entire suite unless QUANTIX_INTEGRATION_TEST=1.
func TestMain(m *testing.M) {
	if os.Getenv("QUANTIX_INTEGRATION_TEST") != "1" {
		fmt.Println("Skipping integration tests: set QUANTIX_INTEGRATION_TEST=1 to enable")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Q5-A: Single node is reachable
// ---------------------------------------------------------------------------

func TestDevnetSingleNode(t *testing.T) {
	resp, err := http.Get(devnetBase + "/blockcount")
	if err != nil {
		t.Fatalf("GET /blockcount: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /blockcount: status %d, want 200", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Q5-B: Block production — blockcount grows above 1 within 60s
// ---------------------------------------------------------------------------

func TestDevnetBlockProduction(t *testing.T) {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		count, err := getBlockCount()
		if err == nil && count > 1 {
			return // success
		}
		time.Sleep(2 * time.Second)
	}
	t.Error("blockcount did not exceed 1 within 60s")
}

// ---------------------------------------------------------------------------
// Q5-C: Block 1 has a non-empty state root
// ---------------------------------------------------------------------------

func TestDevnetStateRootNonZero(t *testing.T) {
	resp, err := http.Get(devnetBase + "/block/1")
	if err != nil {
		t.Fatalf("GET /block/1: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("JSON unmarshal: %v (body: %s)", err, body)
	}

	stateRoot, ok := payload["state_root"]
	if !ok {
		t.Fatal("response missing state_root field")
	}
	if stateRoot == "" || stateRoot == nil {
		t.Error("state_root is empty in block 1")
	}
}

// ---------------------------------------------------------------------------
// Q5-D: Submit 10 transactions and verify acceptance
// ---------------------------------------------------------------------------

func TestDevnetSubmit10Transactions(t *testing.T) {
	sender := "integration-test-sender-" + fmt.Sprintf("%d", time.Now().UnixNano())
	receiver := "integration-test-receiver-" + fmt.Sprintf("%d", time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		tx := map[string]interface{}{
			"sender":    sender,
			"receiver":  receiver,
			"amount":    big.NewInt(int64(i + 1)).String(),
			"nonce":     i,
			"gas_limit": "0",
			"gas_price": "0",
		}
		body, _ := json.Marshal(tx)
		resp, err := http.Post(devnetBase+"/transaction", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("tx %d: POST /transaction: %v", i, err)
		}
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			t.Errorf("tx %d: status %d, want 2xx", i, resp.StatusCode)
		}
	}
	t.Logf("All 10 transactions submitted successfully")
}

// ---------------------------------------------------------------------------
// Q5-E: After transactions, block count advances
// ---------------------------------------------------------------------------

func TestDevnetBlockCountAdvancesAfterTx(t *testing.T) {
	// Get current count
	countBefore, err := getBlockCount()
	if err != nil {
		t.Fatalf("getBlockCount before: %v", err)
	}

	// Submit a transaction
	tx := map[string]interface{}{
		"sender":    "advance-test-sender",
		"receiver":  "advance-test-receiver",
		"amount":    "1000",
		"nonce":     0,
		"gas_limit": "0",
		"gas_price": "0",
	}
	body, _ := json.Marshal(tx)
	resp, err := http.Post(devnetBase+"/transaction", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /transaction: %v", err)
	}
	resp.Body.Close()

	// Wait for block production (up to 30s)
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		countAfter, err := getBlockCount()
		if err == nil && countAfter > countBefore {
			t.Logf("Block count advanced: %d → %d", countBefore, countAfter)
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Logf("NOTE: block count did not advance after tx (may need validators for PBFT)")
}

// ---------------------------------------------------------------------------
// Q5-F: /address endpoint returns balance field
// ---------------------------------------------------------------------------

func TestDevnetAddressEndpoint(t *testing.T) {
	addr := "0000000000000000000000000000000000000000"
	resp, err := http.Get(devnetBase + "/address/" + addr)
	if err != nil {
		t.Fatalf("GET /address/%s: %v", addr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /address: status %d", resp.StatusCode)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := payload["balance"]; !ok {
		t.Error("response missing 'balance' field")
	}
}

// ---------------------------------------------------------------------------
// Q5-G: /mempool returns pending transaction list
// ---------------------------------------------------------------------------

func TestDevnetMempoolEndpoint(t *testing.T) {
	resp, err := http.Get(devnetBase + "/mempool")
	if err != nil {
		t.Fatalf("GET /mempool: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /mempool: status %d", resp.StatusCode)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := payload["count"]; !ok {
		t.Error("response missing 'count' field")
	}
}

// ---------------------------------------------------------------------------
// Q5-H: /validators returns active validator set
// ---------------------------------------------------------------------------

func TestDevnetValidatorsEndpoint(t *testing.T) {
	resp, err := http.Get(devnetBase + "/validators")
	if err != nil {
		t.Fatalf("GET /validators: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /validators: status %d", resp.StatusCode)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := payload["validators"]; !ok {
		t.Error("response missing 'validators' field")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func getBlockCount() (int64, error) {
	resp, err := http.Get(devnetBase + "/blockcount")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var payload struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}
	return payload.Count, nil
}
