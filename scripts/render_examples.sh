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

  # Ordered families.
  run_gither "$input_path" "gither-${image_name}-bayer-2x2-rgb4.png" -algorithm bayer-2x2 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-bayer-4x4-rgb4.png" -algorithm bayer-4x4 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-bayer-8x8-rgb4.png" -algorithm bayer-8x8 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-bayer-16x16-rgb4.png" -algorithm bayer-16x16 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-adaptive-bayer-8x8-rgb4.png" -algorithm adaptive-bayer-8x8 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-adaptive-bayer-16x16-rgb4.png" -algorithm adaptive-bayer-16x16 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-cluster-dot-4x4-palette-auto6.png" -algorithm cluster-dot-4x4 -quantizer palette -palette auto -palette-colors 6
  run_gither "$input_path" "gither-${image_name}-cluster-dot-8x8-palette-auto6.png" -algorithm cluster-dot-8x8 -quantizer palette -palette auto -palette-colors 6
  run_gither "$input_path" "gither-${image_name}-cluster-dot-16x16-palette-auto6.png" -algorithm cluster-dot-16x16 -quantizer palette -palette auto -palette-colors 6
  run_gither "$input_path" "gither-${image_name}-stochastic-cluster-dot-16x16-palette-auto6.png" -algorithm stochastic-cluster-dot-16x16 -quantizer palette -palette auto -palette-colors 6 -seed 7
  run_gither "$input_path" "gither-${image_name}-polyomino-16x16-palette-auto6.png" -algorithm polyomino-16x16 -quantizer palette -palette auto -palette-colors 6
  run_gither "$input_path" "gither-${image_name}-space-filling-16x16-rgb4.png" -algorithm space-filling-16x16 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-space-filling-morton-16x16-rgb4.png" -algorithm space-filling-morton-16x16 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-space-filling-serpentine-16x16-rgb4.png" -algorithm space-filling-serpentine-16x16 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-void-and-cluster-64x64-rgb4.png" -algorithm void-and-cluster-64x64 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-blue-noise-multitone-64x64-gray2.png" -algorithm blue-noise-multitone-64x64 -quantizer gray-levels -levels 2
  run_gither "$input_path" "gither-${image_name}-blue-noise-soft-64x64-gray2.png" -algorithm blue-noise-soft-64x64 -quantizer gray-levels -levels 2
  run_gither "$input_path" "gither-${image_name}-blue-noise-hard-64x64-gray2.png" -algorithm blue-noise-hard-64x64 -quantizer gray-levels -levels 2
  run_gither "$input_path" "gither-${image_name}-dot-diffusion-8x8-rgb4.png" -algorithm dot-diffusion-8x8 -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-dot-diffusion-diagonal-8x8-rgb4.png" -algorithm dot-diffusion-diagonal-8x8 -quantizer rgb-levels -levels 4

  # Yliluoma palette-ordered variants.
  run_gither "$input_path" "gither-${image_name}-yliluoma-1-auto6.png" -algorithm yliluoma-1 -quantizer palette -palette auto -palette-colors 6
  run_gither "$input_path" "gither-${image_name}-yliluoma-2-auto6.png" -algorithm yliluoma-2 -quantizer palette -palette auto -palette-colors 6
  run_gither "$input_path" "gither-${image_name}-yliluoma-3-auto6.png" -algorithm yliluoma-3 -quantizer palette -palette auto -palette-colors 6

  # Classic diffusion kernels.
  run_gither "$input_path" "gither-${image_name}-floyd-steinberg-palette-auto6.png" -algorithm floyd-steinberg -quantizer palette -palette auto -palette-colors 6
  run_gither "$input_path" "gither-${image_name}-false-floyd-steinberg-rgb4.png" -algorithm false-floyd-steinberg -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-jjn-rgb4.png" -algorithm jjn -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-stucki-rgb4.png" -algorithm stucki -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-burkes-rgb4.png" -algorithm burkes -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-sierra-rgb4.png" -algorithm sierra -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-two-row-sierra-rgb4.png" -algorithm two-row-sierra -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-sierra-lite-rgb4.png" -algorithm sierra-lite -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-stevenson-arce-rgb4.png" -algorithm stevenson-arce -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-atkinson-rgb4.png" -algorithm atkinson -quantizer rgb-levels -levels 4

  # Variable diffusion.
  run_gither "$input_path" "gither-${image_name}-ostromoukhov-gray2.png" -algorithm ostromoukhov -quantizer gray-levels -levels 2
  run_gither "$input_path" "gither-${image_name}-zhou-fang-gray2.png" -algorithm zhou-fang -quantizer gray-levels -levels 2 -seed 7
  run_gither "$input_path" "gither-${image_name}-balanced-variable-gray2.png" -algorithm balanced-variable -quantizer gray-levels -levels 2
  run_gither "$input_path" "gither-${image_name}-balanced-variable-thresholded-gray2.png" -algorithm balanced-variable-thresholded -quantizer gray-levels -levels 2 -seed 7
  run_gither "$input_path" "gither-${image_name}-smooth-variable-gray2.png" -algorithm smooth-variable -quantizer gray-levels -levels 2
  run_gither "$input_path" "gither-${image_name}-punchy-variable-gray2.png" -algorithm punchy-variable -quantizer gray-levels -levels 2

  # Stochastic and path-based modes.
  run_gither "$input_path" "gither-${image_name}-threshold-gray2.png" -algorithm threshold -quantizer gray-levels -levels 2 -threshold 127
  run_gither "$input_path" "gither-${image_name}-random-gray2.png" -algorithm random -quantizer gray-levels -levels 2 -seed 7 -random-strength 40
  run_gither "$input_path" "gither-${image_name}-riemersma-rgb4.png" -algorithm riemersma -quantizer rgb-levels -levels 4
  run_gither "$input_path" "gither-${image_name}-am-fm-hybrid-64x64-gray2.png" -algorithm am-fm-hybrid-64x64 -quantizer gray-levels -levels 2 -seed 7
  run_gither "$input_path" "gither-${image_name}-clustered-am-fm-64x64-gray2.png" -algorithm clustered-am-fm-64x64 -quantizer gray-levels -levels 2 -seed 7
  run_gither "$input_path" "gither-${image_name}-dbs-preview-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule preview
  run_gither "$input_path" "gither-${image_name}-dbs-balanced-gray2.png" -algorithm dbs -quantizer gray-levels -levels 2 -dbs-schedule balanced
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

echo "wrote outputs to $OUTPUT_DIR"
