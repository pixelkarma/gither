package gither

import (
	"gither/internal/engine"
	"gither/internal/kernels"
)

func FloydSteinberg(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.FloydSteinberg)
}
func FalseFloydSteinberg(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.FalseFloydSteinberg)
}
func JarvisJudiceNinke(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.JarvisJudiceNinke)
}
func Stucki(img *Image, opts Options) error { return engine.ApplyDiffusion(img, opts, kernels.Stucki) }
func Burkes(img *Image, opts Options) error { return engine.ApplyDiffusion(img, opts, kernels.Burkes) }
func Sierra(img *Image, opts Options) error { return engine.ApplyDiffusion(img, opts, kernels.Sierra) }
func TwoRowSierra(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.TwoRowSierra)
}
func SierraLite(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.SierraLite)
}
func StevensonArce(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.StevensonArce)
}
func Atkinson(img *Image, opts Options) error {
	return engine.ApplyDiffusion(img, opts, kernels.Atkinson)
}
