# gither

`gither` is a Go dithering toolkit with two equally important surfaces:

- a CLI for batch image conversion
- a Go library for embedding dithering directly into your own code

The CLI is the front door. The goal is to make `gither` a serious option when you want broad dithering coverage from a single binary, while the library stays clean enough to use as the Go-native core underneath.

Status: `alpha`

- module path: [`github.com/pixelkarma/gither`](https://github.com/pixelkarma/gither)
- installable CLI: `go install github.com/pixelkarma/gither/cmd/gither@latest`
- license: [The Unlicense](/Users/admin/Documents/dither/gither/LICENSE:1)

## Why gither

- broad algorithm coverage across ordered, diffusion, halftone, Yliluoma, variable diffusion, and DBS families
- usable as both a CLI and a library
- auto-palette and explicit-palette workflows
- dedicated DBS surface, including clustered, multilevel, and palette-index color DBS
- release automation for prebuilt binaries through GitHub Releases

## Install

CLI:

```bash
go install github.com/pixelkarma/gither/cmd/gither@latest
```

Library:

```bash
go get github.com/pixelkarma/gither
```

Build from source:

```bash
git clone https://github.com/pixelkarma/gither.git
cd gither
mkdir -p dist
go build -o ./dist/gither ./cmd/gither
```

## CLI Quick Start

Basic diffusion:

```bash
gither \
  -in ./images/test.png \
  -out ./examples-out/gither-floyd.png \
  -algorithm floyd-steinberg \
  -quantizer rgb-levels \
  -levels 4
```

Auto-palette ordered dither:

```bash
gither \
  -in ./images/test.png \
  -out ./examples-out/gither-cluster-dot.png \
  -algorithm cluster-dot-8x8 \
  -quantizer palette \
  -palette auto \
  -palette-colors 6
```

DBS preview:

```bash
gither \
  -in ./images/test.png \
  -out ./examples-out/gither-dbs-preview.png \
  -algorithm dbs \
  -quantizer gray-levels \
  -levels 2 \
  -dbs-schedule preview \
  -verbose
```

Generate the full example matrix:

```bash
./scripts/render_examples.sh
```

Generate only DBS outputs:

```bash
./scripts/render_dbs_examples.sh
```

Both scripts create [examples-out](/Users/admin/Documents/dither/gither/examples-out:1) if needed and clear old image outputs before rendering.

## Library Quick Start

```go
package main

import (
	"github.com/pixelkarma/gither"
	"github.com/pixelkarma/gither/adapters/stdimage"
)

func main() {
	img, err := stdimage.LoadPath("images/test.png")
	if err != nil {
		panic(err)
	}

	if err := gither.FloydSteinberg(img, gither.Options{
		Quantizer: gither.RGBLevels(4),
	}); err != nil {
		panic(err)
	}

	if err := stdimage.SavePath("out.png", stdimage.ToImage(img), 95); err != nil {
		panic(err)
	}
}
```

More library notes live in:

- [docs/library.md](/Users/admin/Documents/dither/gither/docs/library.md:1)
- [docs/algorithms.md](/Users/admin/Documents/dither/gither/docs/algorithms.md:1)
- [docs/releases.md](/Users/admin/Documents/dither/gither/docs/releases.md:1)

## Stable vs Experimental

This project is alpha. The CLI and library are usable now, but names and APIs can still move before `v1`.

Current intent:

- stable alpha:
  - Bayer
  - cluster-dot
  - Yliluoma
  - classic diffusion kernels
  - threshold/random
  - Riemersma
  - DBS family
- experimental alpha:
  - adaptive ordered
  - polyomino
  - space-filling variants
  - void-and-cluster
  - blue-noise variants
  - dot-diffusion
  - AM/FM hybrids
  - variable diffusion family

`gither -h` reflects the same split.

## Benchmarks

Run the full suite locally:

```bash
go test -run '^$' -bench . -benchmem ./...
```

For focused DBS benchmarking:

```bash
./scripts/benchmark_dbs.sh
```

## Releases

Tagged releases are built automatically with GoReleaser and published to GitHub Releases.

- release automation: [.goreleaser.yaml](/Users/admin/Documents/dither/gither/.goreleaser.yaml:1)
- CI: [.github/workflows/ci.yml](/Users/admin/Documents/dither/gither/.github/workflows/ci.yml:1)
- tagged releases: [.github/workflows/release.yml](/Users/admin/Documents/dither/gither/.github/workflows/release.yml:1)

## Samples

Sample source images are kept in [images](/Users/admin/Documents/dither/gither/images:1) so the scripts and benchmarks have fixture inputs. Generated outputs remain out of git.
