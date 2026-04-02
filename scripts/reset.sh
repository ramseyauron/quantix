#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

MODE="${1:-all}"  # all | devnet | testnet

echo "⚠️  Quantix Reset — stopping containers and wiping chain data"
echo "   Mode: ${MODE}"
read -rp "Are you sure? (yes/N): " confirm
[[ "${confirm}" == "yes" ]] || { echo "Aborted."; exit 0; }

cd "${ROOT}"

if [[ "${MODE}" == "all" || "${MODE}" == "devnet" ]]; then
  echo "🛑 Stopping devnet..."
  docker compose down -v --remove-orphans 2>/dev/null || true
  rm -rf data/devnet
  echo "   ✅ Devnet wiped"
fi

if [[ "${MODE}" == "all" || "${MODE}" == "testnet" ]]; then
  echo "🛑 Stopping testnet..."
  docker compose -f docker-compose-testnet.yml down -v --remove-orphans 2>/dev/null || true
  rm -rf data/validator-{0,1,2,3}
  echo "   ✅ Testnet wiped"
fi

echo ""
echo "✅ Reset complete. Run start-devnet.sh or start-testnet.sh to start fresh."
