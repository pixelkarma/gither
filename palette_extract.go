package gither

import "gither/internal/palettex"

type PaletteExtractMethod = palettex.Method
type PaletteSortMode = palettex.SortMode
type PaletteExtractOptions = palettex.Options

const (
	PaletteMethodMedianCut  = palettex.MethodMedianCut
	PaletteMethodPopularity = palettex.MethodPopularity
)

const (
	PaletteSortRGB       = palettex.SortRGB
	PaletteSortLuma      = palettex.SortLuma
	PaletteSortFrequency = palettex.SortFrequency
)

func ExtractPalette(img *Image, colors int) (Palette, error) {
	return palettex.Extract(img, colors)
}

func ExtractPaletteWithOptions(img *Image, opts PaletteExtractOptions) (Palette, error) {
	return palettex.ExtractWithOptions(img, opts)
}
