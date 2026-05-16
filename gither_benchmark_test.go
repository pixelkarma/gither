package gither

import "testing"

func BenchmarkFloydSteinbergRGBA1024(b *testing.B) {
	src := rgbaGradient(1024, 1024)
	opts := Options{Quantizer: PaletteQuantizer(Palette{{0, 0, 0}, {255, 255, 255}, {0, 255, 128}, {255, 64, 64}})}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pix := append([]uint8(nil), src...)
		img, _ := NewPackedImage(pix, 1024, 1024, RGBA8)
		_ = FloydSteinberg(img, opts)
	}
}

func BenchmarkBayer8x8Gray1024(b *testing.B) {
	src := grayRamp(1024, 1024)
	opts := Options{Quantizer: GrayLevels(2)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pix := append([]uint8(nil), src...)
		img, _ := NewPackedImage(pix, 1024, 1024, Gray8)
		_ = Bayer8x8(img, opts)
	}
}
