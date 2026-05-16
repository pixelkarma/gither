# Experimental Audit

This file records the current intent for the experimental algorithm surface in `gither`.

## Keep Experimental

These modes are worth shipping, but should stay explicitly experimental until they have had broader image review and longer API stability:

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
- `ostromoukhov`
- `zhou-fang`
- `balanced-variable`
- `balanced-variable-thresholded`
- `smooth-variable`
- `punchy-variable`

## Promotion Criteria

An experimental mode should only move into the stable alpha group if it clears all of the following:

- visually distinct and worthwhile on multiple image classes
- no obvious output regressions on the fixture suite
- CLI and library names are likely to stay as-is
- no known performance footguns beyond what the family already implies

## Current Stable Alpha Baseline

The practical stable alpha baseline remains:

- Bayer
- cluster-dot
- Yliluoma
- classic diffusion kernels
- threshold/random
- Riemersma
- DBS family
