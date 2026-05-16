# gither

`gither` is a Go image dithering library and CLI with a broad algorithm surface:

- ordered dithers
- diffusion dithers
- halftone variants
- Yliluoma palette methods
- variable diffusion
- DBS family:
  - `dbs`
  - `clustered-dbs`
  - `multilevel-dbs`
  - `color-dbs`

## Build

```bash
cd /Users/admin/Documents/dither/gither
go build ./cmd/gither
```

## CLI Quick Start

```bash
go run ./cmd/gither \
  -in /Users/admin/Documents/dither/gither/images/cat.png \
  -out /Users/admin/Documents/dither/gither/examples-out/cat-floyd.png \
  -algorithm floyd-steinberg \
  -quantizer rgb-levels \
  -levels 4
```

Palette workflow:

```bash
go run ./cmd/gither \
  -in /Users/admin/Documents/dither/gither/images/cat.png \
  -out /Users/admin/Documents/dither/gither/examples-out/cat-palette.png \
  -algorithm cluster-dot-8x8 \
  -quantizer palette \
  -palette auto \
  -palette-colors 6
```

## DBS Family

The DBS family is a separate optimization-based surface. These modes are slower than
ordered or diffusion dithers, but they expose richer quality tradeoffs.

### Binary DBS

```bash
go run ./cmd/gither \
  -in /Users/admin/Documents/dither/gither/images/cat.png \
  -out /Users/admin/Documents/dither/gither/examples-out/cat-dbs-preview.png \
  -algorithm dbs \
  -quantizer gray-levels \
  -levels 2 \
  -dbs-schedule preview \
  -verbose
```

### Clustered DBS

```bash
go run ./cmd/gither \
  -in /Users/admin/Documents/dither/gither/images/cat.png \
  -out /Users/admin/Documents/dither/gither/examples-out/cat-clustered-dbs.png \
  -algorithm clustered-dbs \
  -quantizer gray-levels \
  -levels 2 \
  -dbs-schedule balanced \
  -dbs-cluster-strength 0.18 \
  -verbose
```

### Multilevel DBS

```bash
go run ./cmd/gither \
  -in /Users/admin/Documents/dither/gither/images/cat.png \
  -out /Users/admin/Documents/dither/gither/examples-out/cat-multilevel-dbs.png \
  -algorithm multilevel-dbs \
  -quantizer gray-levels \
  -levels 4 \
  -dbs-schedule balanced \
  -verbose
```

### Color DBS

`color-dbs` is palette-index DBS. It requires `-quantizer palette`.

```bash
go run ./cmd/gither \
  -in /Users/admin/Documents/dither/gither/images/cat.png \
  -out /Users/admin/Documents/dither/gither/examples-out/cat-color-dbs.png \
  -algorithm color-dbs \
  -quantizer palette \
  -palette auto \
  -palette-colors 6 \
  -dbs-schedule balanced \
  -verbose
```

### DBS Controls

- `-dbs-seed threshold|bayer|floyd-steinberg|cluster-dot-16x16`
- `-dbs-move flip|swap|hybrid`
- `-dbs-metric fast|balanced|perceptual`
- `-dbs-scan raster|serpentine|random`
- `-dbs-radius-policy fixed|expand`
- `-dbs-passes N`
- `-dbs-max-no-improve N`
- `-dbs-restarts N`
- `-dbs-cluster-strength X`
- `-dbs-cluster-tone-aware`

### DBS Schedules

- `preview`: fast binary DBS
- `balanced`: practical default
- `hq`: offline-quality preset
- `custom`: use explicit DBS flags as provided

### Verbose Stats

`-verbose` prints timing and DBS convergence fields such as:

- passes run
- accepted move count
- flip count
- swap count
- restart count

## Example Scripts

Render the full algorithm matrix:

```bash
./scripts/render_examples.sh
```

Render only the DBS family:

```bash
./scripts/render_dbs_examples.sh
```

## Benchmarks

Run the full benchmark suite:

```bash
go test -run '^$' -bench . -benchmem ./...
```

Useful focused DBS benches:

```bash
go test -run '^$' -bench 'BenchmarkAlgorithmsFixtureCat/(dbs-fast|dbs-balanced|dbs-perceptual|clustered-dbs|multilevel-dbs|color-dbs)$' -benchmem -benchtime=1x
go test -run '^$' -bench 'BenchmarkDBSSchedulesFixtureCat/(preview|balanced|hq)$' -benchmem -benchtime=1x
./scripts/benchmark_dbs.sh
```

See [DBS_VERIFICATION.md](/Users/admin/Documents/dither/gither/DBS_VERIFICATION.md:1) for the dedicated fixture suite and comparison benchmark program.
