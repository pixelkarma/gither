package gither

import (
	"gither/internal/core"
	"gither/internal/engine"
	"gither/internal/maps"
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
func ClusterDot4x4(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.ClusterDot4x4, Width: 4, Height: 4, Strength: opts.WithDefaults().Strength}, opts)
}
func ClusterDot8x8(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.ClusterDot8x8, Width: 8, Height: 8, Strength: opts.WithDefaults().Strength}, opts)
}
func ClusterDot16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateClusterDot16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}
func SpaceFilling16x16(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateHilbert16x16(), Width: 16, Height: 16, Strength: opts.WithDefaults().Strength}, opts)
}
func VoidAndCluster64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateVoidAndCluster64x64(), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}
func BlueNoiseMultitone64x64(img *Image, opts Options) error {
	return engine.ApplyOrdered(img, OrderedMap{Values: maps.GenerateBlueNoise64x64(), Width: 64, Height: 64, Strength: opts.WithDefaults().Strength}, opts)
}
func CustomOrdered(img *Image, ordered OrderedMap, opts Options) error {
	return engine.ApplyOrdered(img, ordered, opts)
}

var _ = core.DefaultOrderedStrength
