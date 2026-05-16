package core

import "errors"

type Options struct {
	Quantizer      Quantizer
	Strength       float32
	Threshold      uint8
	Seed           uint64
	RandomStrength uint8
}

const DefaultOrderedStrength = 64.0 / 255.0

func (o Options) WithDefaults() Options {
	if o.Strength == 0 {
		o.Strength = DefaultOrderedStrength
	}
	if o.Seed == 0 {
		o.Seed = 1
	}
	return o
}

type Color struct {
	R uint8
	G uint8
	B uint8
}

type Palette []Color

func (p Palette) Validate() error {
	if len(p) == 0 {
		return ErrEmptyPalette
	}
	if len(p) > 256 {
		return ErrPaletteTooLarge
	}
	return nil
}

func (p Palette) Contains(c Color) bool {
	for _, candidate := range p {
		if candidate == c {
			return true
		}
	}
	return false
}

var (
	ErrZeroDimensions      = errors.New("image width and height must be positive")
	ErrStrideTooSmall      = errors.New("image stride is smaller than row width")
	ErrPixTooShort         = errors.New("image pixel buffer is shorter than required length")
	ErrInvalidFormat       = errors.New("invalid image format")
	ErrInvalidLevels       = errors.New("quantization levels must be >= 2")
	ErrEmptyPalette        = errors.New("palette must contain at least one color")
	ErrPaletteTooLarge     = errors.New("palette cannot contain more than 256 colors")
	ErrInvalidOrderedMap   = errors.New("ordered map dimensions or values are invalid")
	ErrThresholdOutOfRange = errors.New("threshold must be in 0..255")
)
