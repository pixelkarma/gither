#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INPUT_PATH="${1:-$ROOT_DIR/images/test.png}"
OUTPUT_DIR="${2:-$ROOT_DIR/examples-out}"
BIN_MODE="${GITHER_RUN_MODE:-run}"
VERBOSE_FLAG=()

if [[ "${GITHER_VERBOSE:-0}" == "1" ]]; then
  VERBOSE_FLAG=(-verbose)
fi

if [[ ! -f "$INPUT_PATH" ]]; then
  echo "input image not found: $INPUT_PATH" >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_DIR"/*.png "$OUTPUT_DIR"/*.jpg "$OUTPUT_DIR"/*.jpeg

run_gither() {
  local out_name="$1"
  shift
  local cmd=()

  echo "rendering $out_name"
  if [[ "$BIN_MODE" == "bin" ]]; then
    cmd=("$ROOT_DIR/gither" -in "$INPUT_PATH" -out "$OUTPUT_DIR/$out_name")
  else
    cmd=(go run ./cmd/gither -in "$INPUT_PATH" -out "$OUTPUT_DIR/$out_name")
  fi
  if [[ ${#VERBOSE_FLAG[@]} -gt 0 ]]; then
    cmd+=("${VERBOSE_FLAG[@]}")
  fi
  cmd+=("$@")
  "${cmd[@]}"
}

run_gither "cat-dbs-preview-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule preview
run_gither "cat-dbs-balanced-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule balanced
run_gither "cat-dbs-hq-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule hq
run_gither "cat-clustered-dbs-gray2.png" -algorithm clustered-dbs -quantizer gray-levels -levels 2 -dbs-schedule balanced -dbs-cluster-strength 0.18
run_gither "cat-multilevel-dbs-gray4.png" -algorithm multilevel-dbs -quantizer gray-levels -levels 4 -dbs-schedule balanced
run_gither "cat-color-dbs-auto6.png" -algorithm color-dbs -quantizer palette -palette auto -palette-colors 6 -dbs-schedule balanced

echo "wrote DBS outputs to $OUTPUT_DIR"
