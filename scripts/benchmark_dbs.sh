#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_PATH="${1:-}"

CMD=(
  go test
  -run '^$'
  -bench 'BenchmarkDBS(ComparisonFixtures|ColorFixtures|SchedulesFixtureTest)|BenchmarkAlgorithmsFixtureTest/(dbs-fast|dbs-balanced|dbs-perceptual|clustered-dbs|multilevel-dbs|color-dbs)$'
  -benchmem
  -count=1
)

cd "$ROOT_DIR"

if [[ -n "$OUTPUT_PATH" ]]; then
  mkdir -p "$(dirname "$OUTPUT_PATH")"
  "${CMD[@]}" | tee "$OUTPUT_PATH"
else
  "${CMD[@]}"
fi
