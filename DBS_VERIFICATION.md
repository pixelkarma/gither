# DBS Verification

This file describes the dedicated DBS verification and benchmark program in `gither`.

## Fixture Suite

DBS verification now covers these fixture classes:

- `photo`
  - the existing photographic fixture at [images/test.png](/Users/admin/Documents/dither/gither/images/test.png:1)
- `line-art`
  - high-contrast synthetic strokes and sparse detail
- `low-contrast`
  - smooth grayscale ramp with shallow tonal variation
- `texture`
  - deterministic high-frequency grayscale texture
- `color-texture`
  - deterministic high-frequency RGB texture for palette-index color DBS

## Regression Coverage

The fixture suite includes stable hash checks for:

- binary DBS
- clustered DBS
- multilevel DBS
- color DBS

The goal is not to prove image quality mathematically. The goal is to make behavior drift obvious when the DBS family changes.

## Comparison Baselines

The benchmark matrix includes comparison baselines against:

- `threshold`
- `bayer-8x8`
- `floyd-steinberg`
- `riemersma`

That gives the project a practical quality/speed reference set when evaluating DBS changes.

## Quality Notes

These are the intended qualitative differences to watch for when reviewing outputs:

- `threshold`
  - fastest baseline, harsh contouring, no local optimization
- `bayer-8x8`
  - structured ordered texture, stable but visibly patterned
- `floyd-steinberg`
  - strong dispersed diffusion baseline with directional texture
- `riemersma`
  - path-based diffusion with a distinct traversal texture
- `dbs`
  - highest dispersed-dot quality target in grayscale binary mode
- `clustered-dbs`
  - more print-like clustering and heavier blobs, especially in midtones
- `multilevel-dbs`
  - preserves more tonal structure than binary DBS
- `color-dbs`
  - palette-aware optimization, not per-channel grayscale dithering

## Commands

Run the focused DBS verification tests:

```bash
go test ./... -run 'Test(DBS|ClusteredDBS|MultiLevelDBS|ColorDBS|DBSFixtureSuite)'
```

Run the dedicated DBS fixture comparison benchmarks:

```bash
go test -run '^$' -bench 'BenchmarkDBS(ComparisonFixtures|ColorFixtures|SchedulesFixtureCat)' -benchmem -count=1
```

Use the helper script for a timestamped benchmark report:

```bash
./scripts/benchmark_dbs.sh
```

Optionally write the benchmark output to a file:

```bash
./scripts/benchmark_dbs.sh /tmp/dbs-bench.txt
```

## Interpreting Stage 9+ Numbers

If DBS benchmarks regress, look at these first:

- allocations per op
- bytes per op
- whether the regression is isolated to:
  - clustered DBS
  - multilevel DBS
  - color DBS

Large allocation increases are usually a sign that a per-candidate data structure was reintroduced into a hot path.
