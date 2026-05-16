# Algorithm Guide

This file is a practical map of the current `gither` surface, not a theory document.

## Stable Alpha

Ordered:

- `bayer-2x2`
- `bayer-4x4`
- `bayer-8x8`
- `bayer-16x16`
- `cluster-dot-4x4`
- `cluster-dot-8x8`
- `cluster-dot-16x16`
- `custom-ordered`

Palette-ordered:

- `yliluoma-1`
- `yliluoma-2`
- `yliluoma-3`

Diffusion:

- `floyd-steinberg`
- `false-floyd-steinberg`
- `jjn`
- `stucki`
- `burkes`
- `sierra`
- `two-row-sierra`
- `sierra-lite`
- `stevenson-arce`
- `atkinson`

Other stable alpha:

- `threshold`
- `random`
- `riemersma`
- `dbs`
- `clustered-dbs`
- `multilevel-dbs`
- `color-dbs`

## Experimental Alpha

Ordered and halftone variants:

- `adaptive-bayer-8x8`
- `adaptive-bayer-16x16`
- `stochastic-cluster-dot-16x16`
- `polyomino-16x16`
- `space-filling-16x16`
- `space-filling-morton-16x16`
- `space-filling-serpentine-16x16`
- `void-and-cluster-64x64`
- `blue-noise-multitone-64x64`
- `blue-noise-soft-64x64`
- `blue-noise-hard-64x64`
- `dot-diffusion-8x8`
- `dot-diffusion-diagonal-8x8`
- `am-fm-hybrid-64x64`
- `clustered-am-fm-64x64`

Variable diffusion:

- `ostromoukhov`
- `zhou-fang`
- `balanced-variable`
- `balanced-variable-thresholded`
- `smooth-variable`
- `punchy-variable`

## Selection Hints

- If you want a safe default, start with `floyd-steinberg`, `bayer-8x8`, `cluster-dot-8x8`, or `dbs`.
- If you want explicit palette control, start with `yliluoma-2` or `cluster-dot-8x8`.
- If you want print-like clustered structure, start with the cluster-dot family or `clustered-dbs`.
- If you want exploratory or stylized outputs, the experimental group is where most of the unusual looks live.
