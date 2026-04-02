#!/usr/bin/env bash
set -euo pipefail

# logs.sh — Tail logs for running Quantix nodes
# Usage: ./scripts/logs.sh [devnet|testnet|monitoring] [--follow]

MODE="${1:-devnet}"
FOLLOW="${2:-}"
TAIL_FLAG=()
[[ "${FOLLOW}" == "--follow" || "${FOLLOW}" == "-f" ]] && TAIL_FLAG=("-f")

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT}"

case "${MODE}" in
  devnet)
    docker compose logs --tail=50 "${TAIL_FLAG[@]}" quantix-devnet
    ;;
  testnet)
    docker compose -f docker-compose-testnet.yml logs --tail=50 "${TAIL_FLAG[@]}"
    ;;
  monitoring)
    docker compose -f docker-compose-monitoring.yml logs --tail=50 "${TAIL_FLAG[@]}"
    ;;
  *)
    echo "Usage: $0 [devnet|testnet|monitoring] [--follow]"
    exit 1
    ;;
esac
