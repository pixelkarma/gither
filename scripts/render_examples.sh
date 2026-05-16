#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INPUT_PATH="${1:-$ROOT_DIR/images/cat.png}"
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

  echo "rendering $out_name"
  if [[ "$BIN_MODE" == "bin" ]]; then
    "$ROOT_DIR/gither" -in "$INPUT_PATH" -out "$OUTPUT_DIR/$out_name" "${VERBOSE_FLAG[@]}" "$@"
  else
    go run ./cmd/gither -in "$INPUT_PATH" -out "$OUTPUT_DIR/$out_name" "${VERBOSE_FLAG[@]}" "$@"
  fi
}

# Ordered families.
run_gither "cat-bayer-2x2-rgb4.png" -algorithm bayer-2x2 -quantizer rgb-levels -levels 4
run_gither "cat-bayer-4x4-rgb4.png" -algorithm bayer-4x4 -quantizer rgb-levels -levels 4
run_gither "cat-bayer-8x8-rgb4.png" -algorithm bayer-8x8 -quantizer rgb-levels -levels 4
run_gither "cat-bayer-16x16-rgb4.png" -algorithm bayer-16x16 -quantizer rgb-levels -levels 4
run_gither "cat-cluster-dot-4x4-palette-auto6.png" -algorithm cluster-dot-4x4 -quantizer palette -palette auto -palette-colors 6
run_gither "cat-cluster-dot-8x8-palette-auto6.png" -algorithm cluster-dot-8x8 -quantizer palette -palette auto -palette-colors 6
run_gither "cat-cluster-dot-16x16-palette-auto6.png" -algorithm cluster-dot-16x16 -quantizer palette -palette auto -palette-colors 6
run_gither "cat-space-filling-16x16-rgb4.png" -algorithm space-filling-16x16 -quantizer rgb-levels -levels 4
run_gither "cat-void-and-cluster-64x64-rgb4.png" -algorithm void-and-cluster-64x64 -quantizer rgb-levels -levels 4
run_gither "cat-blue-noise-multitone-64x64-gray2.png" -algorithm blue-noise-multitone-64x64 -quantizer gray-levels -levels 2

# Yliluoma palette-ordered variants.
run_gither "cat-yliluoma-1-auto6.png" -algorithm yliluoma-1 -quantizer palette -palette auto -palette-colors 6
run_gither "cat-yliluoma-2-auto6.png" -algorithm yliluoma-2 -quantizer palette -palette auto -palette-colors 6
run_gither "cat-yliluoma-3-auto6.png" -algorithm yliluoma-3 -quantizer palette -palette auto -palette-colors 6

# Classic diffusion kernels.
run_gither "cat-floyd-steinberg-palette-auto6.png" -algorithm floyd-steinberg -quantizer palette -palette auto -palette-colors 6
run_gither "cat-false-floyd-steinberg-rgb4.png" -algorithm false-floyd-steinberg -quantizer rgb-levels -levels 4
run_gither "cat-jjn-rgb4.png" -algorithm jjn -quantizer rgb-levels -levels 4
run_gither "cat-stucki-rgb4.png" -algorithm stucki -quantizer rgb-levels -levels 4
run_gither "cat-burkes-rgb4.png" -algorithm burkes -quantizer rgb-levels -levels 4
run_gither "cat-sierra-rgb4.png" -algorithm sierra -quantizer rgb-levels -levels 4
run_gither "cat-two-row-sierra-rgb4.png" -algorithm two-row-sierra -quantizer rgb-levels -levels 4
run_gither "cat-sierra-lite-rgb4.png" -algorithm sierra-lite -quantizer rgb-levels -levels 4
run_gither "cat-stevenson-arce-rgb4.png" -algorithm stevenson-arce -quantizer rgb-levels -levels 4
run_gither "cat-atkinson-rgb4.png" -algorithm atkinson -quantizer rgb-levels -levels 4

# Variable diffusion.
run_gither "cat-ostromoukhov-gray2.png" -algorithm ostromoukhov -quantizer gray-levels -levels 2
run_gither "cat-zhou-fang-gray2.png" -algorithm zhou-fang -quantizer gray-levels -levels 2 -seed 7
run_gither "cat-balanced-variable-gray2.png" -algorithm balanced-variable -quantizer gray-levels -levels 2
run_gither "cat-balanced-variable-thresholded-gray2.png" -algorithm balanced-variable-thresholded -quantizer gray-levels -levels 2 -seed 7

# Stochastic and path-based modes.
run_gither "cat-threshold-gray2.png" -algorithm threshold -quantizer gray-levels -levels 2 -threshold 127
run_gither "cat-random-gray2.png" -algorithm random -quantizer gray-levels -levels 2 -seed 7 -random-strength 40
run_gither "cat-riemersma-rgb4.png" -algorithm riemersma -quantizer rgb-levels -levels 4

echo "wrote outputs to $OUTPUT_DIR"
