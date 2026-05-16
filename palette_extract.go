package gither

import "gither/internal/palettex"

func ExtractPalette(img *Image, colors int) (Palette, error) {
	return palettex.Extract(img, colors)
}
