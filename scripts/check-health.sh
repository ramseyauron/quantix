#!/usr/bin/env bash
set -euo pipefail

# check-health.sh — Check /blockcount on each running Quantix node
# Usage: ./scripts/check-health.sh [devnet|testnet]

MODE="${1:-devnet}"

check_node() {
  local name="$1"
  local port="$2"
  local url="http://localhost:${port}/blockcount"

  printf "  %-14s " "${name}:"
  local resp
  if resp=$(curl -sf --max-time 3 "${url}" 2>/dev/null); then
    echo "✅ OK  — block count: ${resp}  (${url})"
  else
    echo "❌ UNREACHABLE  (${url})"
  fi
}

echo ""
echo "🔍 Quantix Health Check — ${MODE}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ "${MODE}" == "devnet" ]]; then
  check_node "quantix-devnet" 8560
elif [[ "${MODE}" == "testnet" ]]; then
  check_node "validator-0" 8560
  check_node "validator-1" 8561
  check_node "validator-2" 8562
  check_node "validator-3" 8563
else
  echo "Unknown mode '${MODE}'. Use 'devnet' or 'testnet'."
  exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
