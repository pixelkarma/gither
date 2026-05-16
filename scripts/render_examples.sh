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

# Ordered families.
run_gither "gither-bayer-2x2-rgb4.png" -algorithm bayer-2x2 -quantizer rgb-levels -levels 4
run_gither "gither-bayer-4x4-rgb4.png" -algorithm bayer-4x4 -quantizer rgb-levels -levels 4
run_gither "gither-bayer-8x8-rgb4.png" -algorithm bayer-8x8 -quantizer rgb-levels -levels 4
run_gither "gither-bayer-16x16-rgb4.png" -algorithm bayer-16x16 -quantizer rgb-levels -levels 4
run_gither "gither-adaptive-bayer-8x8-rgb4.png" -algorithm adaptive-bayer-8x8 -quantizer rgb-levels -levels 4
run_gither "gither-adaptive-bayer-16x16-rgb4.png" -algorithm adaptive-bayer-16x16 -quantizer rgb-levels -levels 4
run_gither "gither-cluster-dot-4x4-palette-auto6.png" -algorithm cluster-dot-4x4 -quantizer palette -palette auto -palette-colors 6
run_gither "gither-cluster-dot-8x8-palette-auto6.png" -algorithm cluster-dot-8x8 -quantizer palette -palette auto -palette-colors 6
run_gither "gither-cluster-dot-16x16-palette-auto6.png" -algorithm cluster-dot-16x16 -quantizer palette -palette auto -palette-colors 6
run_gither "gither-stochastic-cluster-dot-16x16-palette-auto6.png" -algorithm stochastic-cluster-dot-16x16 -quantizer palette -palette auto -palette-colors 6 -seed 7
run_gither "gither-polyomino-16x16-palette-auto6.png" -algorithm polyomino-16x16 -quantizer palette -palette auto -palette-colors 6
run_gither "gither-space-filling-16x16-rgb4.png" -algorithm space-filling-16x16 -quantizer rgb-levels -levels 4
run_gither "gither-space-filling-morton-16x16-rgb4.png" -algorithm space-filling-morton-16x16 -quantizer rgb-levels -levels 4
run_gither "gither-space-filling-serpentine-16x16-rgb4.png" -algorithm space-filling-serpentine-16x16 -quantizer rgb-levels -levels 4
run_gither "gither-void-and-cluster-64x64-rgb4.png" -algorithm void-and-cluster-64x64 -quantizer rgb-levels -levels 4
run_gither "gither-blue-noise-multitone-64x64-gray2.png" -algorithm blue-noise-multitone-64x64 -quantizer gray-levels -levels 2
run_gither "gither-blue-noise-soft-64x64-gray2.png" -algorithm blue-noise-soft-64x64 -quantizer gray-levels -levels 2
run_gither "gither-blue-noise-hard-64x64-gray2.png" -algorithm blue-noise-hard-64x64 -quantizer gray-levels -levels 2
run_gither "gither-dot-diffusion-8x8-rgb4.png" -algorithm dot-diffusion-8x8 -quantizer rgb-levels -levels 4
run_gither "gither-dot-diffusion-diagonal-8x8-rgb4.png" -algorithm dot-diffusion-diagonal-8x8 -quantizer rgb-levels -levels 4

# Yliluoma palette-ordered variants.
run_gither "gither-yliluoma-1-auto6.png" -algorithm yliluoma-1 -quantizer palette -palette auto -palette-colors 6
run_gither "gither-yliluoma-2-auto6.png" -algorithm yliluoma-2 -quantizer palette -palette auto -palette-colors 6
run_gither "gither-yliluoma-3-auto6.png" -algorithm yliluoma-3 -quantizer palette -palette auto -palette-colors 6

# Classic diffusion kernels.
run_gither "gither-floyd-steinberg-palette-auto6.png" -algorithm floyd-steinberg -quantizer palette -palette auto -palette-colors 6
run_gither "gither-false-floyd-steinberg-rgb4.png" -algorithm false-floyd-steinberg -quantizer rgb-levels -levels 4
run_gither "gither-jjn-rgb4.png" -algorithm jjn -quantizer rgb-levels -levels 4
run_gither "gither-stucki-rgb4.png" -algorithm stucki -quantizer rgb-levels -levels 4
run_gither "gither-burkes-rgb4.png" -algorithm burkes -quantizer rgb-levels -levels 4
run_gither "gither-sierra-rgb4.png" -algorithm sierra -quantizer rgb-levels -levels 4
run_gither "gither-two-row-sierra-rgb4.png" -algorithm two-row-sierra -quantizer rgb-levels -levels 4
run_gither "gither-sierra-lite-rgb4.png" -algorithm sierra-lite -quantizer rgb-levels -levels 4
run_gither "gither-stevenson-arce-rgb4.png" -algorithm stevenson-arce -quantizer rgb-levels -levels 4
run_gither "gither-atkinson-rgb4.png" -algorithm atkinson -quantizer rgb-levels -levels 4

# Variable diffusion.
run_gither "gither-ostromoukhov-gray2.png" -algorithm ostromoukhov -quantizer gray-levels -levels 2
run_gither "gither-zhou-fang-gray2.png" -algorithm zhou-fang -quantizer gray-levels -levels 2 -seed 7
run_gither "gither-balanced-variable-gray2.png" -algorithm balanced-variable -quantizer gray-levels -levels 2
run_gither "gither-balanced-variable-thresholded-gray2.png" -algorithm balanced-variable-thresholded -quantizer gray-levels -levels 2 -seed 7
run_gither "gither-smooth-variable-gray2.png" -algorithm smooth-variable -quantizer gray-levels -levels 2
run_gither "gither-punchy-variable-gray2.png" -algorithm punchy-variable -quantizer gray-levels -levels 2

# Stochastic and path-based modes.
run_gither "gither-threshold-gray2.png" -algorithm threshold -quantizer gray-levels -levels 2 -threshold 127
run_gither "gither-random-gray2.png" -algorithm random -quantizer gray-levels -levels 2 -seed 7 -random-strength 40
run_gither "gither-riemersma-rgb4.png" -algorithm riemersma -quantizer rgb-levels -levels 4
run_gither "gither-am-fm-hybrid-64x64-gray2.png" -algorithm am-fm-hybrid-64x64 -quantizer gray-levels -levels 2 -seed 7
run_gither "gither-clustered-am-fm-64x64-gray2.png" -algorithm clustered-am-fm-64x64 -quantizer gray-levels -levels 2 -seed 7
run_gither "gither-dbs-preview-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule preview
run_gither "gither-dbs-balanced-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule balanced
run_gither "gither-clustered-dbs-gray2.png" -algorithm clustered-dbs -quantizer gray-levels -levels 2 -dbs-schedule balanced -dbs-cluster-strength 0.18
run_gither "gither-multilevel-dbs-gray4.png" -algorithm multilevel-dbs -quantizer gray-levels -levels 4 -dbs-schedule balanced
run_gither "gither-color-dbs-auto6.png" -algorithm color-dbs -quantizer palette -palette auto -palette-colors 6 -dbs-schedule balanced

echo "wrote outputs to $OUTPUT_DIR"
