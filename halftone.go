package gither

import (
	"gither/internal/engine"
	"gither/internal/maps"
)

func DotDiffusion8x8(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateDotDiffusion8x8(), Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts)
}

func DotDiffusionDiagonal8x8(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateDotDiffusionDiagonal8x8(), Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts)
}

func BlueNoiseSoft64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateBlueNoise64x64WithProfile(maps.BlueNoiseProfile{BaseWeight: 0.82, ClusterWeight: 0.08, RingWeight: 0.06, JitterWeight: 0.04, Frequency: 6.5}), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}

func BlueNoiseHard64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateBlueNoise64x64WithProfile(maps.BlueNoiseProfile{BaseWeight: 0.58, ClusterWeight: 0.24, RingWeight: 0.12, JitterWeight: 0.06, Frequency: 11.0}), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}

func AMFMHybrid64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateAMFM64x64(maps.AmFmProfile{FMWeight: 0.55, AMWeight: 0.27, ClusterWeight: 0.08, MacroWeight: 0.10}, opts.WithDefaults().Seed^0x6f4d68616c66746f), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}

func ClusteredAMFM64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateAMFM64x64(maps.AmFmProfile{FMWeight: 0.26, AMWeight: 0.48, ClusterWeight: 0.18, MacroWeight: 0.08}, opts.WithDefaults().Seed^0x636c757374657264), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}
