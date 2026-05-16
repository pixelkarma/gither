# DBS Roadmap

## Intent

This document lays out a staged plan for bringing **Direct Binary Search**
halftoning into `gither` as a complete algorithm family, not just a minimal
first pass.

The goal is to arrive at:

- high-quality binary grayscale DBS
- configurable perceptual error models
- multiple local search strategies
- multilevel and color-capable extensions
- benchmarked and regression-tested implementations
- a CLI and library surface that fits the rest of `gither`

This is a longer-term program than ordered dithering or classic diffusion.
DBS needs its own engine model because it is an **iterative optimization**
family rather than a one-pass constructive family.

## What DBS Is

DBS is a local-search halftoning method.

At a high level:

1. choose an initial halftone pattern
2. define an error function between target image and current halftone
3. test candidate moves such as:
   - flipping a pixel
   - swapping a black/white pair
4. accept moves that reduce the error
5. repeat until convergence or budget exhaustion

The defining characteristics are:

- binary or low-level output quality can be extremely high
- quality depends heavily on the error model
- performance depends heavily on incremental update design
- architecture is very different from ordered or diffusion pipelines

## Design Principles

The roadmap assumes these design choices:

- DBS gets its own engine under `internal/engine/dbs.go` and related files
- start grayscale-first, then extend outward
- optimize around **incremental error updates**, not full recomputation
- keep deterministic mode available for testing and regression coverage
- benchmark every stage before expanding the family
- only add color/multilevel variants after grayscale binary DBS is solid

## Package Direction

Recommended eventual structure:

```text
internal/
  dbs/
    core.go
    init.go
    moves.go
    metric.go
    filtered_error.go
    grayscale.go
    multilevel.go
    color.go
    schedule.go
    neighborhood.go
    cache.go
```

Possible public wrappers:

```text
dbs.go
```

Potential public surface:

```go
type DBSOptions struct { ... }
type DBSMetric struct { ... }

func DirectBinarySearch(img *Image, opts DBSOptions) error
func ClusteredDBS(img *Image, opts DBSOptions) error
func MultiLevelDBS(img *Image, opts DBSOptions) error
func ColorDBS(img *Image, opts DBSOptions) error
```

## Stage 0: Research and Scaffolding

### Goal

Prepare the engine architecture so later stages do not need to be rewritten.

### Deliverables

- `DBS_ROADMAP.md`
- dedicated `internal/dbs/` or `internal/engine/dbs_*.go` area
- option and terminology decisions
- benchmark harness placeholders

### Decisions to lock

- output domain for stage 1:
  - binary grayscale only
- move set:
  - pixel flip
  - black/white pair swap
- initial seeds:
  - threshold
  - ordered
  - Floyd-Steinberg
- error model baseline:
  - filtered grayscale squared error
- schedule baseline:
  - greedy descent with multiple passes

### Exit criteria

- the package boundary and file ownership are clear
- the benchmark/test plan is decided before implementation starts

## Stage 1: Binary Grayscale DBS Core

### Goal

Implement a correct grayscale binary DBS engine that can improve a seed image.

### Scope

- grayscale target image only
- binary halftone state only
- global optimization loop
- no color
- no multilevel output
- correctness first, not ultimate speed

### Deliverables

- binary state representation
- seed generation:
  - threshold
  - Bayer
  - Floyd-Steinberg
- move evaluation:
  - single-pixel flip
- acceptance logic:
  - greedy improvement only
- stop conditions:
  - max iterations
  - no-improvement pass cutoff

### Recommended files

```text
internal/dbs/core.go
internal/dbs/init.go
internal/dbs/grayscale.go
internal/dbs/moves.go
```

### Data structures

- binary plane:
  - `[]uint8` or packed bitset later
- target grayscale plane:
  - `[]uint8`
- working filtered image or error response:
  - start simple

### Exit criteria

- produces visibly improved binary halftones over a naive threshold seed
- deterministic when seed/schedule are fixed
- fixture hashes are stable

## Stage 2: Incremental Error Engine

### Goal

Replace naive full-image rescoring with local incremental updates.

This is the stage that makes DBS viable rather than just demonstrable.

### Scope

- maintain filtered error state incrementally
- update only affected neighborhoods after a move
- support fast score delta computation

### Deliverables

- local influence kernel
- filtered response cache
- delta scoring for flip moves
- move application that updates caches in place

### Why it matters

Without this stage, DBS remains too slow for anything beyond tiny fixtures.

### Exit criteria

- substantial speedup versus stage 1
- same output as naive rescoring under equivalent schedule
- benchmark proves incremental scoring is correct and faster

## Stage 3: Pair-Swap DBS

### Goal

Add neighborhood swaps between black and white pixels, not just flips.

### Why

Classic DBS quality depends heavily on allowing local pixel exchanges.
Flip-only search is a weaker family member.

### Scope

- evaluate black/white swaps in a bounded neighborhood
- preserve or control global black-pixel count if desired
- schedule flips and swaps together

### Deliverables

- neighborhood search policy
- swap delta scoring
- configurable search radius
- optional mode:
  - `flip-only`
  - `swap-only`
  - `hybrid`

### Exit criteria

- visible quality improvement over flip-only DBS
- speed remains acceptable under bounded neighborhoods

## Stage 4: Better Error Metrics

### Goal

Move from plain local grayscale error to stronger perceptual scoring.

### Scope

- Gaussian or low-pass filtered grayscale error
- multiple kernel presets
- optional edge-sensitive weighting
- configurable metric presets

### Recommended presets

- `fast`
  - small blur, fast updates
- `balanced`
  - moderate blur
- `perceptual`
  - larger support, slower, highest quality

### Deliverables

- metric abstraction
- kernel preset library
- CLI/API metric options
- regression fixtures per metric

### Exit criteria

- metrics are selectable and benchmarked
- higher-quality presets show a real visible difference

## Stage 5: Scheduling and Convergence Control

### Goal

Add robust search schedules so the engine is tunable for quality versus speed.

### Scope

- pass scheduling
- annealing-style or probabilistic acceptance as optional mode
- random and deterministic scan orders
- restart strategy
- convergence reporting

### Deliverables

- configurable iteration budget
- neighborhood expansion policies
- deterministic mode for tests
- opportunistic randomized mode for quality experiments

### Exit criteria

- quality/speed tradeoffs are measurable and reproducible
- CLI can expose practical presets like:
  - `preview`
  - `balanced`
  - `hq`

## Stage 6: Clustered DBS Variant

### Goal

Add clustered-dot biased DBS, not just dispersed-dot style optimization.

### Why

Some use cases want print-like clustered halftones instead of fine blue-noise
textures.

### Scope

- clustered seed options
- cluster-biased move cost or structural regularization
- optional tone-dependent cluster tendency

### Deliverables

- `ClusteredDBS`
- cluster-preserving move heuristics
- tests showing stable output and different visual character

### Exit criteria

- the family can produce both dispersed and clustered DBS results

## Stage 7: Multilevel DBS

### Goal

Extend DBS beyond pure black/white into multilevel grayscale output.

### Scope

- output levels > 2
- move set becomes:
  - raise/lower level
  - local exchange
- metric must support multilevel state cleanly

### Deliverables

- `MultiLevelDBS`
- level-aware incremental scoring
- compatibility with existing quantizer model where possible

### Important note

This is a major change in state space and should not be rushed.

### Exit criteria

- multilevel DBS gives meaningfully different results from binary DBS
- performance is still controllable

## Stage 8: Color DBS

### Goal

Implement true color DBS, not just grayscale DBS run per channel.

### Scope

- color target support
- palette-aware or direct RGB color state
- perceptual color metric
- optional alpha handling strategy

### Two paths

1. palette-index DBS
   - each pixel stores a palette index
   - moves change or swap indices
   - best fit with `gither` palette workflows

2. direct multichannel DBS
   - larger search space
   - more complex and slower

### Recommendation

Implement palette-index color DBS first.

### Exit criteria

- color DBS is real and useful, not a toy per-channel hack

## Stage 9: Acceleration and Systems Work

### Goal

Make the mature DBS family fast enough for serious offline use.

### Scope

- packed binary storage
- cache-friendly neighborhood layout
- precomputed influence tables
- reduced allocations
- SIMD-friendly loops where practical
- optional worker parallelism for independent candidate evaluation

### Important constraint

Parallelism is harder here than in ordered dithering.
The focus should be:

- better data layout
- better incremental math
- controlled candidate batching

not naive goroutine fan-out.

### Exit criteria

- benchmarks show real improvement on large fixtures
- allocations and memory traffic are under control

## Stage 10: Product Surface and Tooling

### Goal

Expose DBS as a first-class family in the library and CLI.

### Deliverables

- public `DBSOptions`
- CLI algorithms:
  - `dbs`
  - `clustered-dbs`
  - `multilevel-dbs`
  - `color-dbs`
- seed selection flags
- metric flags
- pass budget / neighborhood controls
- optional verbose convergence stats

### Example CLI shape

```bash
gither \
  -algorithm dbs \
  -quantizer gray-levels \
  -levels 2 \
  -dbs-seed floyd-steinberg \
  -dbs-metric balanced \
  -dbs-neighborhood 1 \
  -dbs-passes 12
```

### Exit criteria

- DBS family is usable without custom code
- options are documented and benchmarked

## Stage 11: Verification and Benchmark Program

### Goal

Treat DBS like a product line, not a side algorithm.

### Deliverables

- dedicated fixture suite
- golden-image or hash regression coverage
- `go test -bench` coverage
- quality comparison notes against:
  - threshold
  - Bayer
  - Floyd-Steinberg
  - Riemersma
- memory and runtime reporting

### Recommended fixtures

- photographic grayscale
- line art
- low-contrast smooth gradient
- high-frequency texture
- palette-index color fixture later

### Exit criteria

- quality and performance claims are defensible

## Stage 12: Complete DBS Family

### Goal

Arrive at a complete DBS solution for `gither`.

### Status

Completed on May 16, 2026.

### Completion standard

The DBS family is complete when all of these are true:

- [x] grayscale binary DBS exists and is solid
- [x] pair-swap and flip scheduling are available
- [x] multiple perceptual metrics are available
- [x] clustered DBS exists
- [x] multilevel DBS exists
- [x] color DBS exists
- [x] the CLI supports practical presets
- [x] fixtures and benchmarks cover the family
- [x] performance is optimized enough for offline use on real images

At that point, DBS is no longer a roadmap item. It becomes a maintained major
family alongside:

- ordered dithers
- diffusion dithers
- variable diffusion
- Riemersma
- halftone variants

## Non-Goals For Early Stages

These should not be pulled into stage 1 or 2:

- full color DBS immediately
- alpha-aware compositing DBS
- GPU rewrite
- arbitrary perceptual vision model research
- ultra-general acceptance schedules before the core engine is proven

## Recommended Execution Order

The practical order is:

1. Stage 0
2. Stage 1
3. Stage 2
4. Stage 3
5. Stage 4
6. Stage 5
7. Stage 6
8. Stage 7
9. Stage 8
10. Stage 9
11. Stage 10
12. Stage 11
13. Stage 12

## Immediate Next Implementation Step

If DBS work starts now, the first concrete coding target should be:

- **Stage 1: binary grayscale DBS core**

Specifically:

- create a binary grayscale state representation
- implement threshold/Bayer/Floyd seeders
- add a naive but correct flip-only optimizer
- benchmark it on `images/cat.png`

That gives a correctness baseline before any serious optimization work begins.
