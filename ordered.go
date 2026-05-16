package gither

import (
	"github.com/pixelkarma/gither/internal/engine"
	"github.com/pixelkarma/gither/internal/maps"
)

// OrderedMap is a reusable threshold map for ordered dithering.
type OrderedMap = engine.OrderedMap

// NewOrderedMap constructs an ordered map from rank values.
func NewOrderedMap(values []uint16, width, height int, strength float32) (OrderedMap, error) {
	return engine.NewOrderedMap(values, width, height, strength)
}

// NewOrderedMapFromU8 constructs an ordered map from byte-sized rank values.
func NewOrderedMapFromU8(values []uint8, width, height int, strength float32) (OrderedMap, error) {
	return engine.NewOrderedMapFromU8(values, width, height, strength)
}

// Bayer2x2 applies a 2x2 Bayer ordered dither.
func Bayer2x2(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.Bayer2x2, Width: 2, Height: 2, Strength: opts.WithDefaults().Strength}, opts)
}

// Bayer4x4 applies a 4x4 Bayer ordered dither.
func Bayer4x4(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.Bayer4x4, Width: 4, Height: 4, Strength: opts.WithDefaults().Strength}, opts)
}

// Bayer8x8 applies an 8x8 Bayer ordered dither.
func Bayer8x8(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.Bayer8x8, Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts)
}

// Bayer16x16 applies a generated 16x16 Bayer ordered dither.
func Bayer16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateBayer16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}

// AdaptiveBayer8x8 applies an adaptive 8x8 Bayer-style ordered dither.
func AdaptiveBayer8x8(img *Image, opts Options) error {
	return engine.ApplyAdaptiveOrdered(img, OrderedMap{Values: maps.Bayer8x8, Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts, 1)
}

// AdaptiveBayer16x16 applies an adaptive 16x16 Bayer-style ordered dither.
func AdaptiveBayer16x16(img *Image, opts Options) error {
	return engine.ApplyAdaptiveOrdered(img, OrderedMap{Values: maps.GenerateBayer16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts, 2)
}

// ClusterDot4x4 applies a 4x4 clustered-dot ordered dither.
func ClusterDot4x4(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.ClusterDot4x4, Width: 4, Height: 4, Strength: opts.WithDefaults().Strength}, opts)
}

// ClusterDot8x8 applies an 8x8 clustered-dot ordered dither.
func ClusterDot8x8(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.ClusterDot8x8, Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts)
}

// ClusterDot16x16 applies a generated 16x16 clustered-dot ordered dither.
func ClusterDot16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateClusterDot16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}

// StochasticClusterDot16x16 applies a seeded clustered-dot variant with jittered ranking.
func StochasticClusterDot16x16(img *Image, opts Options) error {
	return engine.StochasticClusteredDot(img, opts, OrderedMap{Values: maps.GenerateStochasticCluster16x16(opts.WithDefaults().Seed), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength})
}

// Polyomino16x16 applies a polyomino-style ordered dither map.
func Polyomino16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GeneratePolyomino16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}

// SpaceFilling16x16 applies a Hilbert-style space-filling ordered dither.
func SpaceFilling16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateHilbert16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}

// SpaceFillingMorton16x16 applies a Morton-curve ordered dither.
func SpaceFillingMorton16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateMorton16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}

// SpaceFillingSerpentine16x16 applies a serpentine space-filling ordered dither.
func SpaceFillingSerpentine16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateSerpentine16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}

// VoidAndCluster64x64 applies a void-and-cluster ordered dither.
func VoidAndCluster64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateVoidAndCluster64x64(), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}

// BlueNoiseMultitone64x64 applies the default blue-noise multitone ordered dither.
func BlueNoiseMultitone64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateBlueNoise64x64(), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}

// Yliluoma1 applies the first Yliluoma palette-ordered dither variant.
func Yliluoma1(img *Image, opts Options) error { return engine.Yliluoma1(img, opts) }

// Yliluoma2 applies the second Yliluoma palette-ordered dither variant.
func Yliluoma2(img *Image, opts Options) error { return engine.Yliluoma2(img, opts) }

// Yliluoma3 applies the third Yliluoma palette-ordered dither variant.
func Yliluoma3(img *Image, opts Options) error { return engine.Yliluoma3(img, opts) }

// CustomOrdered applies a caller-provided ordered dither map.
func CustomOrdered(img *Image, ordered OrderedMap, opts Options) error {
	return engine.ApplyOrdered(img, ordered, opts)
}
