package gither

import (
	"github.com/pixelkarma/gither/internal/engine"
	"github.com/pixelkarma/gither/internal/kernels"
)

// FloydSteinberg applies Floyd-Steinberg error diffusion.
func FloydSteinberg(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.FloydSteinberg)
}

// FalseFloydSteinberg applies the False Floyd-Steinberg diffusion kernel.
func FalseFloydSteinberg(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.FalseFloydSteinberg)
}

// JarvisJudiceNinke applies the Jarvis-Judice-Ninke diffusion kernel.
func JarvisJudiceNinke(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.JarvisJudiceNinke)
}

// Stucki applies the Stucki diffusion kernel.
func Stucki(img *Image, opts Options) error { return engine.ApplyDiffusion(img, opts, kernels.Stucki) }

// Burkes applies the Burkes diffusion kernel.
func Burkes(img *Image, opts Options) error { return engine.ApplyDiffusion(img, opts, kernels.Burkes) }

// Sierra applies the Sierra diffusion kernel.
func Sierra(img *Image, opts Options) error { return engine.ApplyDiffusion(img, opts, kernels.Sierra) }

// TwoRowSierra applies the two-row Sierra diffusion kernel.
func TwoRowSierra(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.TwoRowSierra)
}

// SierraLite applies the Sierra Lite diffusion kernel.
func SierraLite(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.SierraLite)
}

// StevensonArce applies the Stevenson-Arce diffusion kernel.
func StevensonArce(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.StevensonArce)
}

// Atkinson applies the Atkinson diffusion kernel.
func Atkinson(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.Atkinson)
}
