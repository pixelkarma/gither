package stdimage

import (
	"image"
	"image/color"

	"gither"
)

func FromImage(src image.Image) (*gither.Image, error) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	switch img := src.(type) {
	case *image.Gray:
		return gither.NewImage(img.Pix, width, height, img.Stride, gither.Gray8)
	case *image.RGBA:
		return gither.NewImage(img.Pix, width, height, img.Stride, gither.RGBA8)
	case *image.NRGBA:
		pix := make([]uint8, width*height*4)
		offset := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := src.At(x, y).RGBA()
				pix[offset] = uint8(r >> 8)
				pix[offset+1] = uint8(g >> 8)
				pix[offset+2] = uint8(b >> 8)
				pix[offset+3] = uint8(a >> 8)
				offset += 4
			}
		}
		return gither.NewPackedImage(pix, width, height, gither.RGBA8)
	default:
		pix := make([]uint8, width*height*4)
		offset := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := src.At(x, y).RGBA()
				pix[offset] = uint8(r >> 8)
				pix[offset+1] = uint8(g >> 8)
				pix[offset+2] = uint8(b >> 8)
				pix[offset+3] = uint8(a >> 8)
				offset += 4
			}
		}
		return gither.NewPackedImage(pix, width, height, gither.RGBA8)
	}
}

func ToImage(src *gither.Image) image.Image {
	switch src.Format {
	case gither.Gray8:
		return &image.Gray{
			Pix:    append([]uint8(nil), src.Pix...),
			Stride: src.Stride,
			Rect:   image.Rect(0, 0, src.Width, src.Height),
		}
	case gither.RGB8:
		out := image.NewRGBA(image.Rect(0, 0, src.Width, src.Height))
		for y := 0; y < src.Height; y++ {
			row := src.Row(y)
			for x := 0; x < src.Width; x++ {
				i := x * 3
				out.SetRGBA(x, y, color.RGBA{R: row[i], G: row[i+1], B: row[i+2], A: 255})
			}
		}
		return out
	default:
		return &image.RGBA{
			Pix:    append([]uint8(nil), src.Pix...),
			Stride: src.Stride,
			Rect:   image.Rect(0, 0, src.Width, src.Height),
		}
	}
}
