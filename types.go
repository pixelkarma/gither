package gither

import "github.com/pixelkarma/gither/internal/core"

type (
	// Options configures non-DBS algorithms.
	Options = core.Options
	// Format describes the pixel layout of an Image.
	Format = core.Format
	// Image is the shared in-memory image type used by the library.
	Image = core.Image
	// Color is an RGB color value used by palette and single-color quantizers.
	Color = core.Color
	// Palette is a list of Colors used for palette-constrained output.
	Palette = core.Palette
	// QuantizeKind identifies the active quantization mode.
	QuantizeKind = core.QuantizeKind
	// Quantizer maps source colors into a constrained output space.
	Quantizer = core.Quantizer
)

const (
	// Gray8 stores one 8-bit channel per pixel.
	Gray8 = core.Gray8
	// RGB8 stores three 8-bit channels per pixel.
	RGB8 = core.RGB8
	// RGBA8 stores three color channels plus alpha.
	RGBA8 = core.RGBA8
)

const (
	// QuantizeGrayLevels quantizes output into evenly spaced gray values.
	QuantizeGrayLevels = core.QuantizeGrayLevels
	// QuantizeRGBLevels quantizes each RGB channel independently.
	QuantizeRGBLevels = core.QuantizeRGBLevels
	// QuantizePalette constrains output to a provided palette.
	QuantizePalette = core.QuantizePalette
	// QuantizeSingleColor scales a single color by luminance.
	QuantizeSingleColor = core.QuantizeSingleColor
)

var (
	// ErrZeroDimensions reports an image with non-positive width or height.
	ErrZeroDimensions = core.ErrZeroDimensions
	// ErrStrideTooSmall reports an image stride smaller than one row of pixels.
	ErrStrideTooSmall = core.ErrStrideTooSmall
	// ErrPixTooShort reports a pixel buffer shorter than the declared image size.
	ErrPixTooShort = core.ErrPixTooShort
	// ErrInvalidFormat reports an unsupported pixel format.
	ErrInvalidFormat = core.ErrInvalidFormat
	// ErrInvalidLevels reports invalid quantization levels.
	ErrInvalidLevels = core.ErrInvalidLevels
	// ErrEmptyPalette reports a missing palette in palette-constrained modes.
	ErrEmptyPalette = core.ErrEmptyPalette
	// ErrPaletteTooLarge reports a palette larger than the supported size.
	ErrPaletteTooLarge = core.ErrPaletteTooLarge
	// ErrInvalidOrderedMap reports invalid ordered-map dimensions or values.
	ErrInvalidOrderedMap = core.ErrInvalidOrderedMap
	// ErrThresholdOutOfRange reports an invalid binary threshold.
	ErrThresholdOutOfRange = core.ErrThresholdOutOfRange
)

// NewImage constructs an Image over an existing pixel buffer.
func NewImage(pix []uint8, width, height, stride int, format Format) (*Image, error) {
	return core.NewImage(pix, width, height, stride, format)
}

// NewPackedImage constructs an Image with a tightly packed row layout.
func NewPackedImage(pix []uint8, width, height int, format Format) (*Image, error) {
	return core.NewPackedImage(pix, width, height, format)
}

// GrayLevels builds a gray-level quantizer with the requested number of levels.
func GrayLevels(levels int) Quantizer { return core.GrayLevels(levels) }

// RGBLevels builds an RGB-level quantizer with the requested number of levels.
func RGBLevels(levels int) Quantizer { return core.RGBLevels(levels) }

// PaletteQuantizer builds a quantizer constrained to the provided palette.
func PaletteQuantizer(p Palette) Quantizer { return core.PaletteQuantizer(p) }

// SingleColorQuantizer scales one color over the requested number of levels.
func SingleColorQuantizer(levels int, color Color) Quantizer {
	return core.SingleColorQuantizer(levels, color)
}
