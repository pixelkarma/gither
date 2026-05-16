package core

import "errors"

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
