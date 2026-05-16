package gither

import "github.com/pixelkarma/gither/internal/core"

type QuantizeKind = core.QuantizeKind
type Quantizer = core.Quantizer

const (
	QuantizeGrayLevels  = core.QuantizeGrayLevels
	QuantizeRGBLevels   = core.QuantizeRGBLevels
	QuantizePalette     = core.QuantizePalette
	QuantizeSingleColor = core.QuantizeSingleColor
)

func GrayLevels(levels int) Quantizer      { return core.GrayLevels(levels) }
func RGBLevels(levels int) Quantizer       { return core.RGBLevels(levels) }
func PaletteQuantizer(p Palette) Quantizer { return core.PaletteQuantizer(p) }
func SingleColorQuantizer(levels int, color Color) Quantizer {
	return core.SingleColorQuantizer(levels, color)
}
