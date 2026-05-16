package gither

import "github.com/pixelkarma/gither/internal/palettex"

// PaletteExtractMethod selects the palette extraction strategy.
type PaletteExtractMethod = palettex.Method

// PaletteSortMode controls how an extracted palette is ordered.
type PaletteSortMode = palettex.SortMode

// PaletteExtractOptions configures automatic palette extraction.
type PaletteExtractOptions = palettex.Options

const (
	// PaletteMethodMedianCut extracts colors using a median-cut style partitioner.
	PaletteMethodMedianCut = palettex.MethodMedianCut
	// PaletteMethodPopularity extracts the most frequent colors first.
	PaletteMethodPopularity = palettex.MethodPopularity
)

const (
	// PaletteSortRGB sorts by raw RGB channel order.
	PaletteSortRGB = palettex.SortRGB
	// PaletteSortLuma sorts darkest-to-lightest by luma.
	PaletteSortLuma = palettex.SortLuma
	// PaletteSortFrequency sorts by observed frequency in the source image.
	PaletteSortFrequency = palettex.SortFrequency
)

// ExtractPalette derives a palette with the requested number of colors.
func ExtractPalette(img *Image, colors int) (Palette, error) {
	return palettex.Extract(img, colors)
}

// ExtractPaletteWithOptions derives a palette using explicit extraction settings.
func ExtractPaletteWithOptions(img *Image, opts PaletteExtractOptions) (Palette, error) {
	return palettex.ExtractWithOptions(img, opts)
}
