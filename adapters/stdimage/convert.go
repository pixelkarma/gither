// Package stdimage bridges gither.Image with the Go standard library image types.
package stdimage

import (
	"bufio"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/pixelkarma/gither"
)

// LoadPath decodes an image file and converts it into a gither.Image.
func LoadPath(path string) (*gither.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(bufio.NewReaderSize(file, 1<<20))
	if err != nil {
		return nil, err
	}
	return FromImage(img)
}

// SavePath encodes an image to disk, using JPEG for .jpg/.jpeg and PNG otherwise.
func SavePath(path string, img image.Image, jpegQuality int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriterSize(file, 1<<20)

	var encodeErr error
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		encodeErr = jpeg.Encode(writer, img, &jpeg.Options{Quality: jpegQuality})
	default:
		encodeErr = png.Encode(writer, img)
	}
	if encodeErr != nil {
		return encodeErr
	}
	return writer.Flush()
}

// FromImage converts a standard library image into a gither.Image.
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

// ToImage converts a gither.Image into a standard library image.Image.
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
			srcRow := src.Row(y)
			dstRow := out.Pix[y*out.Stride : y*out.Stride+src.Width*4]
			for x, srcOffset, dstOffset := 0, 0, 0; x < src.Width; x++ {
				dstRow[dstOffset] = srcRow[srcOffset]
				dstRow[dstOffset+1] = srcRow[srcOffset+1]
				dstRow[dstOffset+2] = srcRow[srcOffset+2]
				dstRow[dstOffset+3] = 255
				srcOffset += 3
				dstOffset += 4
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
