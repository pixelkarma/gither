package gither

import "github.com/pixelkarma/gither/internal/core"

type (
	Options      = core.Options
	Format       = core.Format
	Image        = core.Image
	Color        = core.Color
	Palette      = core.Palette
	QuantizeKind = core.QuantizeKind
	Quantizer    = core.Quantizer
)

const (
	Gray8 = core.Gray8
	RGB8  = core.RGB8
	RGBA8 = core.RGBA8
)

const (
	QuantizeGrayLevels  = core.QuantizeGrayLevels
	QuantizeRGBLevels   = core.QuantizeRGBLevels
	QuantizePalette     = core.QuantizePalette
	QuantizeSingleColor = core.QuantizeSingleColor
)

var (
	ErrZeroDimensions      = core.ErrZeroDimensions
	ErrStrideTooSmall      = core.ErrStrideTooSmall
	ErrPixTooShort         = core.ErrPixTooShort
	ErrInvalidFormat       = core.ErrInvalidFormat
	ErrInvalidLevels       = core.ErrInvalidLevels
	ErrEmptyPalette        = core.ErrEmptyPalette
	ErrPaletteTooLarge     = core.ErrPaletteTooLarge
	ErrInvalidOrderedMap   = core.ErrInvalidOrderedMap
	ErrThresholdOutOfRange = core.ErrThresholdOutOfRange
)

func NewImage(pix []uint8, width, height, stride int, format Format) (*Image, error) {
	return core.NewImage(pix, width, height, stride, format)
}

func NewPackedImage(pix []uint8, width, height int, format Format) (*Image, error) {
	return core.NewPackedImage(pix, width, height, format)
}

func GrayLevels(levels int) Quantizer      { return core.GrayLevels(levels) }
func RGBLevels(levels int) Quantizer       { return core.RGBLevels(levels) }
func PaletteQuantizer(p Palette) Quantizer { return core.PaletteQuantizer(p) }
func SingleColorQuantizer(levels int, color Color) Quantizer {
	return core.SingleColorQuantizer(levels, color)
}
