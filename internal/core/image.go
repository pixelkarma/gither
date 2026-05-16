package core

type Format uint8

const (
	Gray8 Format = iota
	RGB8
	RGBA8
)

type Image struct {
	Pix    []uint8
	Width  int
	Height int
	Stride int
	Format Format
}

func NewImage(pix []uint8, width, height, stride int, format Format) (*Image, error) {
	img := &Image{Pix: pix, Width: width, Height: height, Stride: stride, Format: format}
	if err := img.Validate(); err != nil {
		return nil, err
	}
	return img, nil
}

func NewPackedImage(pix []uint8, width, height int, format Format) (*Image, error) {
	channels, ok := format.Channels()
	if !ok {
		return nil, ErrInvalidFormat
	}
	return NewImage(pix, width, height, width*channels, format)
}

func (img *Image) Validate() error {
	if img.Width <= 0 || img.Height <= 0 {
		return ErrZeroDimensions
	}
	channels, ok := img.Format.Channels()
	if !ok {
		return ErrInvalidFormat
	}
	minStride := img.Width * channels
	if img.Stride < minStride {
		return ErrStrideTooSmall
	}
	required := img.Stride * img.Height
	if len(img.Pix) < required {
		return ErrPixTooShort
	}
	return nil
}

func (img *Image) Clone() *Image {
	dup := make([]uint8, len(img.Pix))
	copy(dup, img.Pix)
	return &Image{Pix: dup, Width: img.Width, Height: img.Height, Stride: img.Stride, Format: img.Format}
}

func (img *Image) Row(y int) []uint8 {
	start := y * img.Stride
	return img.Pix[start : start+img.Stride]
}

func (img *Image) PixelOffset(x, y int) int {
	return y*img.Stride + x*img.ChannelCount()
}

func (img *Image) ChannelCount() int {
	channels, _ := img.Format.Channels()
	return channels
}

func (img *Image) HasAlpha() bool {
	return img.Format == RGBA8
}

func (f Format) Channels() (int, bool) {
	switch f {
	case Gray8:
		return 1, true
	case RGB8:
		return 3, true
	case RGBA8:
		return 4, true
	default:
		return 0, false
	}
}
