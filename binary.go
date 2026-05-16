package gither

import "github.com/pixelkarma/gither/internal/engine"

// Threshold applies threshold-based binary dithering before quantization.
func Threshold(img *Image, opts Options) error { return engine.Threshold(img, opts) }

// Random applies random-threshold binary dithering before quantization.
func Random(img *Image, opts Options) error { return engine.Random(img, opts) }

// Riemersma applies path-based error diffusion along a Hilbert-style traversal.
func Riemersma(img *Image, opts Options) error { return engine.Riemersma(img, opts) }
