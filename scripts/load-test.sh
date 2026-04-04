#!/usr/bin/env bash
# F.R.I.D.A.Y. Load Test Script — Quantix P3-D1
# Target: 1000 tx/s sustained | PASS threshold: 500 tx/s
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

RPC_URL="${QUANTIX_RPC:-http://localhost:8560}"
DURATION="${LOAD_TEST_DURATION:-60}"
PASS_THRESHOLD=500

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'

log()  { echo -e "${CYAN}[FRIDAY]${NC} $*"; }
ok()   { echo -e "${GREEN}[  OK  ]${NC} $*"; }
warn() { echo -e "${YELLOW}[ WARN ]${NC} $*"; }
fail() { echo -e "${RED}[ FAIL ]${NC} $*"; }

# ── 1. Ensure node is up ─────────────────────────────────────────────────────
log "Checking if Quantix node is live at ${RPC_URL}..."
if ! curl -sf "${RPC_URL}/blockcount" > /dev/null 2>&1; then
  warn "Node not reachable — starting devnet..."
  cd "${ROOT}"
  docker compose up --build -d
  log "Waiting for node to become healthy (up to 60s)..."
  for i in $(seq 1 60); do
    sleep 1
    if curl -sf "${RPC_URL}/blockcount" > /dev/null 2>&1; then
      ok "Node is up after ${i}s"
      break
    fi
    if [ "$i" -eq 60 ]; then
      fail "Node failed to start in 60s. Aborting."
      exit 1
    fi
  done
else
  ok "Node is live."
fi

# ── 2. Snapshot starting state ───────────────────────────────────────────────
START_BLOCK=$(curl -sf "${RPC_URL}/blockcount" 2>/dev/null | tr -d '"' | tr -d ' ' || echo "0")
START_TIME=$(date +%s%3N)  # milliseconds
log "Starting load test: ${DURATION}s, target ${PASS_THRESHOLD}+ tx/s"
log "Start block: ${START_BLOCK}"

# ── 3. Fire transactions ──────────────────────────────────────────────────────
TX_SUBMITTED=0
TX_ERRORS=0
END_TIME=$(( $(date +%s) + DURATION ))

# Build a minimal test transaction payload
TX_PAYLOAD='{"type":"transfer","from":"0x0000000000000000000000000000000000000001","to":"0x0000000000000000000000000000000000000002","amount":"1","fee":"1","nonce":0,"signature":"deadbeef"}'

log "Firing transactions for ${DURATION}s..."
BATCH_SIZE=50  # parallel curl requests per batch

while [ "$(date +%s)" -lt "${END_TIME}" ]; do
  # Launch parallel batch
  for _ in $(seq 1 ${BATCH_SIZE}); do
    curl -sf -X POST "${RPC_URL}/tx" \
      -H "Content-Type: application/json" \
      -d "${TX_PAYLOAD}" > /dev/null 2>&1 && \
      echo "ok" || echo "err" &
  done

  # Collect results
  RESULTS=$(wait; jobs -l 2>/dev/null | wc -l || echo "0")
  TX_SUBMITTED=$(( TX_SUBMITTED + BATCH_SIZE ))
done

# Wait for background jobs
wait

ACTUAL_DURATION_MS=$(( $(date +%s%3N) - START_TIME ))
ACTUAL_DURATION_S=$(( ACTUAL_DURATION_MS / 1000 ))
[ "$ACTUAL_DURATION_S" -lt 1 ] && ACTUAL_DURATION_S=1

# ── 4. Use a proper counter via tmpfile approach ──────────────────────────────
TMPDIR_LOAD=$(mktemp -d)
OK_FILE="${TMPDIR_LOAD}/ok"
ERR_FILE="${TMPDIR_LOAD}/err"
touch "$OK_FILE" "$ERR_FILE"

TX_SUBMITTED=0
TX_ERRORS=0
END_TIME2=$(( $(date +%s) + DURATION ))

log "Phase 2: Accurate measurement run (${DURATION}s)..."
while [ "$(date +%s)" -lt "${END_TIME2}" ]; do
  for _ in $(seq 1 ${BATCH_SIZE}); do
    (
      if curl -sf -X POST "${RPC_URL}/tx" \
           -H "Content-Type: application/json" \
           -d "${TX_PAYLOAD}" > /dev/null 2>&1; then
        echo 1 >> "$OK_FILE"
      else
        echo 1 >> "$ERR_FILE"
      fi
    ) &
  done
  wait
done
wait

TX_SUBMITTED=$(wc -l < "$OK_FILE" 2>/dev/null || echo 0)
TX_ERRORS=$(wc -l < "$ERR_FILE" 2>/dev/null || echo 0)
rm -rf "$TMPDIR_LOAD"

# ── 5. Measure block production ──────────────────────────────────────────────
END_BLOCK=$(curl -sf "${RPC_URL}/blockcount" 2>/dev/null | tr -d '"' | tr -d ' ' || echo "${START_BLOCK}")
BLOCKS_PRODUCED=$(( END_BLOCK - START_BLOCK ))

SUBMITTED_TPS=$(( TX_SUBMITTED / ACTUAL_DURATION_S ))
BLOCK_RATE_PER_MIN=$(( BLOCKS_PRODUCED * 60 / ACTUAL_DURATION_S ))

# Estimate committed tx (via block stats if available)
COMMITTED_TPS=$(curl -sf "${RPC_URL}/stats" 2>/dev/null | \
  grep -o '"tx_per_second":[0-9.]*' | cut -d: -f2 | cut -d. -f1 || \
  echo "${SUBMITTED_TPS}")

# ── 6. Report ─────────────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  QUANTIX LOAD TEST REPORT"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
printf "  Duration:            %ss\n" "${ACTUAL_DURATION_S}"
printf "  TX Submitted:        %s\n" "${TX_SUBMITTED}"
printf "  TX Errors:           %s\n" "${TX_ERRORS}"
printf "  TX/s Submitted:      %s\n" "${SUBMITTED_TPS}"
printf "  TX/s Committed:      %s (estimated)\n" "${COMMITTED_TPS}"
printf "  Blocks Produced:     %s\n" "${BLOCKS_PRODUCED}"
printf "  Block Rate:          %s blocks/min\n" "${BLOCK_RATE_PER_MIN}"
printf "  Target:              %s tx/s\n" "1000"
printf "  Pass Threshold:      %s tx/s\n" "${PASS_THRESHOLD}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ "${SUBMITTED_TPS}" -ge "${PASS_THRESHOLD}" ]; then
  ok "RESULT: ✅ PASS — ${SUBMITTED_TPS} tx/s >= ${PASS_THRESHOLD} tx/s threshold"
  exit 0
else
  fail "RESULT: ❌ FAIL — ${SUBMITTED_TPS} tx/s < ${PASS_THRESHOLD} tx/s threshold"
  exit 1
fi
