package gither

import (
	"hash/fnv"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func grayRamp(width, height int) []uint8 {
	out := make([]uint8, width*height)
	den := width*height - 1
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out[y*width+x] = uint8(((x + y*width) * 255) / den)
		}
	}
	return out
}

func rgbGradient(width, height int) []uint8 {
	out := make([]uint8, width*height*3)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 3
			out[i] = uint8((x * 255) / testMaxInt(1, width-1))
			out[i+1] = uint8((y * 255) / testMaxInt(1, height-1))
			out[i+2] = uint8(((x + y) * 255) / testMaxInt(1, width+height-2))
		}
	}
	return out
}

func rgbaGradient(width, height int) []uint8 {
	out := make([]uint8, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 4
			out[i] = uint8((x * 255) / testMaxInt(1, width-1))
			out[i+1] = uint8((y * 255) / testMaxInt(1, height-1))
			out[i+2] = uint8(((x + y) * 255) / testMaxInt(1, width+height-2))
			out[i+3] = uint8((x*17 + y*29) % 256)
		}
	}
	return out
}

func hashBytes(data []uint8) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(data)
	return h.Sum64()
}

func mustLoadFixtureForTest(tb testing.TB) *Image {
	tb.Helper()
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("unable to resolve test file path")
	}
	path := filepath.Join(filepath.Dir(callerFile), "images", "test.png")
	f, err := os.Open(path)
	if err != nil {
		tb.Fatal(err)
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		tb.Fatal(err)
	}
	return packedRGBAFromImage(tb, src)
}

func packedRGBAFromImage(tb testing.TB, src image.Image) *Image {
	tb.Helper()
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
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
	img, err := NewPackedImage(pix, width, height, RGBA8)
	if err != nil {
		tb.Fatal(err)
	}
	return img
}
