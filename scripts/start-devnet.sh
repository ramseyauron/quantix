#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "🚀 Starting Quantix devnet..."
cd "${ROOT}"

mkdir -p data/devnet
docker compose up --build -d

echo ""
echo "✅ Devnet is up!"
echo "   RPC API  : http://localhost:8560"
echo "   WebSocket: ws://localhost:8700"
echo "   P2P      : :32307"
echo ""
echo "Logs: ./scripts/logs.sh devnet"
echo "Stop: docker compose down"
