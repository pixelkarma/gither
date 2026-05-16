#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INPUT_SPEC="${1:-$ROOT_DIR/images}"
OUTPUT_DIR="${2:-$ROOT_DIR/examples-out}"
BIN_MODE="${GITHER_RUN_MODE:-run}"
VERBOSE_FLAG=()
INPUT_FILES=()

if [[ "${GITHER_VERBOSE:-0}" == "1" ]]; then
  VERBOSE_FLAG=(-verbose)
fi

collect_inputs() {
  local spec="$1"
  if [[ -f "$spec" ]]; then
    INPUT_FILES=("$spec")
    return
  fi
  if [[ ! -d "$spec" ]]; then
    echo "input path not found: $spec" >&2
    exit 1
  fi
  while IFS= read -r path; do
    INPUT_FILES+=("$path")
  done < <(find "$spec" -maxdepth 1 -type f \( -iname '*.png' -o -iname '*.jpg' -o -iname '*.jpeg' -o -iname '*.webp' \) | sort)
  if [[ ${#INPUT_FILES[@]} -eq 0 ]]; then
    echo "no input images found in: $spec" >&2
    exit 1
  fi
}

image_label() {
  local path="$1"
  local name
  name="$(basename "$path")"
  name="${name%.*}"
  printf '%s' "$name"
}

run_gither() {
  local input_path="$1"
  local out_name="$2"
  shift 2
  local cmd=()

  echo "rendering $out_name"
  if [[ "$BIN_MODE" == "bin" ]]; then
    cmd=("$ROOT_DIR/dist/gither" -in "$input_path" -out "$OUTPUT_DIR/$out_name")
  else
    cmd=(go run ./cmd/gither -in "$input_path" -out "$OUTPUT_DIR/$out_name")
  fi
  if [[ ${#VERBOSE_FLAG[@]} -gt 0 ]]; then
    cmd+=("${VERBOSE_FLAG[@]}")
  fi
  cmd+=("$@")
  "${cmd[@]}"
}

render_for_image() {
  local input_path="$1"
  local image_name="$2"

  run_gither "$input_path" "gither-${image_name}-dbs-preview-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule preview
  run_gither "$input_path" "gither-${image_name}-dbs-balanced-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule balanced
  run_gither "$input_path" "gither-${image_name}-dbs-hq-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule hq
  run_gither "$input_path" "gither-${image_name}-clustered-dbs-gray2.png" -algorithm clustered-dbs -quantizer gray-levels -levels 2 -dbs-schedule balanced -dbs-cluster-strength 0.18
  run_gither "$input_path" "gither-${image_name}-multilevel-dbs-gray4.png" -algorithm multilevel-dbs -quantizer gray-levels -levels 4 -dbs-schedule balanced
  run_gither "$input_path" "gither-${image_name}-color-dbs-auto6.png" -algorithm color-dbs -quantizer palette -palette auto -palette-colors 6 -dbs-schedule balanced
}

collect_inputs "$INPUT_SPEC"
mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_DIR"/*.png "$OUTPUT_DIR"/*.jpg "$OUTPUT_DIR"/*.jpeg

for input_path in "${INPUT_FILES[@]}"; do
  render_for_image "$input_path" "$(image_label "$input_path")"
done

echo "wrote DBS outputs to $OUTPUT_DIR"
