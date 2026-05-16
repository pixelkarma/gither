package core

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
