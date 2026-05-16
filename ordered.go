package gither

import (
	"github.com/pixelkarma/gither/internal/engine"
	"github.com/pixelkarma/gither/internal/maps"
)

type OrderedMap = engine.OrderedMap

func NewOrderedMap(values []uint16, width, height int, strength float32) (OrderedMap, error) {
	return engine.NewOrderedMap(values, width, height, strength)
}

func NewOrderedMapFromU8(values []uint8, width, height int, strength float32) (OrderedMap, error) {
	return engine.NewOrderedMapFromU8(values, width, height, strength)
}

func Bayer2x2(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.Bayer2x2, Width: 2, Height: 2, Strength: opts.WithDefaults().Strength}, opts)
}
func Bayer4x4(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.Bayer4x4, Width: 4, Height: 4, Strength: opts.WithDefaults().Strength}, opts)
}
func Bayer8x8(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.Bayer8x8, Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts)
}
func Bayer16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateBayer16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}
func AdaptiveBayer8x8(img *Image, opts Options) error {
	return engine.ApplyAdaptiveOrdered(img, OrderedMap{Values: maps.Bayer8x8, Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts, 1)
}
func AdaptiveBayer16x16(img *Image, opts Options) error {
	return engine.ApplyAdaptiveOrdered(img, OrderedMap{Values: maps.GenerateBayer16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts, 2)
}
func ClusterDot4x4(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.ClusterDot4x4, Width: 4, Height: 4, Strength: opts.WithDefaults().Strength}, opts)
}
func ClusterDot8x8(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.ClusterDot8x8, Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts)
}
func ClusterDot16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateClusterDot16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}
func StochasticClusterDot16x16(img *Image, opts Options) error {
	return engine.StochasticClusteredDot(img, opts, OrderedMap{Values: maps.GenerateStochasticCluster16x16(opts.WithDefaults().Seed), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength})
}
func Polyomino16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GeneratePolyomino16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}
func SpaceFilling16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateHilbert16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}
func SpaceFillingMorton16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateMorton16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}
func SpaceFillingSerpentine16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateSerpentine16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}
func VoidAndCluster64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateVoidAndCluster64x64(), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}
func BlueNoiseMultitone64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateBlueNoise64x64(), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}
func Yliluoma1(img *Image, opts Options) error { return engine.Yliluoma1(img, opts) }
func Yliluoma2(img *Image, opts Options) error { return engine.Yliluoma2(img, opts) }
func Yliluoma3(img *Image, opts Options) error { return engine.Yliluoma3(img, opts) }
func CustomOrdered(img *Image, ordered OrderedMap, opts Options) error {
	return engine.ApplyOrdered(img, ordered, opts)
}
