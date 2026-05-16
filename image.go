package gither

import "github.com/pixelkarma/gither/internal/core"

type Format = core.Format
type Image = core.Image

const (
	Gray8 = core.Gray8
	RGB8  = core.RGB8
	RGBA8 = core.RGBA8
)

func NewImage(pix []uint8, width, height, stride int, format Format) (*Image, error) {
	return core.NewImage(pix, width, height, stride, format)
}

func NewPackedImage(pix []uint8, width, height int, format Format) (*Image, error) {
	return core.NewPackedImage(pix, width, height, format)
}
