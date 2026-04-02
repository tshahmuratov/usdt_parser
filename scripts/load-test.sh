#!/usr/bin/env bash
set -euo pipefail

ADDR="${1:-localhost:50051}"
CONCURRENCY="${2:-50}"
DURATION="${3:-60s}"

if ! command -v ghz &>/dev/null; then
  echo "ghz not found. Install with: go install github.com/bojand/ghz/cmd/ghz@latest"
  exit 1
fi

echo "=== gRPC Load Test ==="
echo "Target:      $ADDR"
echo "Concurrency: $CONCURRENCY"
echo "Duration:    $DURATION"
echo ""

ghz --insecure \
    --call rates.v1.RateService/GetRates \
    --data '{"method": {"top_n": {"n": 0}}}' \
    --concurrency "$CONCURRENCY" \
    --duration "$DURATION" \
    "$ADDR"
