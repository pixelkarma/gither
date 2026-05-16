package gither

import "testing"

func TestBayerGrayLevelsBinaryInvariant(t *testing.T) {
	pix := grayRamp(16, 16)
	img, err := NewPackedImage(pix, 16, 16, Gray8)
	if err != nil {
		t.Fatal(err)
	}
	if err := Bayer8x8(img, Options{Quantizer: GrayLevels(2)}); err != nil {
		t.Fatal(err)
	}
	for _, v := range img.Pix[:16*16] {
		if v != 0 && v != 255 {
			t.Fatalf("unexpected gray value %d", v)
		}
	}
}

func TestDiffusionPaletteMembership(t *testing.T) {
	palette := Palette{{0, 0, 0}, {255, 255, 255}, {255, 0, 0}, {0, 128, 255}}
	pix := rgbGradient(12, 12)
	img, err := NewPackedImage(pix, 12, 12, RGB8)
	if err != nil {
		t.Fatal(err)
	}
	if err := FloydSteinberg(img, Options{Quantizer: PaletteQuantizer(palette)}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(img.Pix); i += 3 {
		c := Color{img.Pix[i], img.Pix[i+1], img.Pix[i+2]}
		if !palette.Contains(c) {
			t.Fatalf("pixel %v is not in palette", c)
		}
	}
}

func TestRGBAAlphaPreserved(t *testing.T) {
	pix := rgbaGradient(10, 10)
	originalAlpha := make([]uint8, 100)
	for i := range originalAlpha {
		originalAlpha[i] = pix[i*4+3]
	}
	img, err := NewPackedImage(pix, 10, 10, RGBA8)
	if err != nil {
		t.Fatal(err)
	}
	if err := Atkinson(img, Options{Quantizer: RGBLevels(3)}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		if img.Pix[i*4+3] != originalAlpha[i] {
			t.Fatalf("alpha changed at pixel %d", i)
		}
	}
}

func TestRiemersmaDeterministic(t *testing.T) {
	pixA := rgbGradient(18, 18)
	pixB := rgbGradient(18, 18)
	imgA, _ := NewPackedImage(pixA, 18, 18, RGB8)
	imgB, _ := NewPackedImage(pixB, 18, 18, RGB8)
	opts := Options{Quantizer: RGBLevels(4)}
	if err := Riemersma(imgA, opts); err != nil {
		t.Fatal(err)
	}
	if err := Riemersma(imgB, opts); err != nil {
		t.Fatal(err)
	}
	if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
		t.Fatal("riemersma output should be deterministic")
	}
}

func TestRandomDeterministicBySeed(t *testing.T) {
	pixA := grayRamp(16, 16)
	pixB := grayRamp(16, 16)
	imgA, _ := NewPackedImage(pixA, 16, 16, Gray8)
	imgB, _ := NewPackedImage(pixB, 16, 16, Gray8)
	opts := Options{Quantizer: GrayLevels(2), Seed: 99, RandomStrength: 48}
	if err := Random(imgA, opts); err != nil {
		t.Fatal(err)
	}
	if err := Random(imgB, opts); err != nil {
		t.Fatal(err)
	}
	if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
		t.Fatal("random dithering should match for the same seed")
	}
}

func TestExtractPaletteProducesRequestedCount(t *testing.T) {
	img, err := NewPackedImage(rgbGradient(24, 24), 24, 24, RGB8)
	if err != nil {
		t.Fatal(err)
	}
	palette, err := ExtractPalette(img, 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(palette) != 6 {
		t.Fatalf("expected 6 colors, got %d", len(palette))
	}
}

func TestYliluomaPaletteMembership(t *testing.T) {
	palette := Palette{{0, 0, 0}, {255, 255, 255}, {255, 0, 0}, {0, 128, 255}}
	img, err := NewPackedImage(rgbGradient(10, 10), 10, 10, RGB8)
	if err != nil {
		t.Fatal(err)
	}
	if err := Yliluoma2(img, Options{Quantizer: PaletteQuantizer(palette)}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(img.Pix); i += 3 {
		c := Color{img.Pix[i], img.Pix[i+1], img.Pix[i+2]}
		if !palette.Contains(c) {
			t.Fatalf("pixel %v is not in palette", c)
		}
	}
}

func TestBlueNoiseDeterministic(t *testing.T) {
	imgA, _ := NewPackedImage(grayRamp(32, 32), 32, 32, Gray8)
	imgB, _ := NewPackedImage(grayRamp(32, 32), 32, 32, Gray8)
	opts := Options{Quantizer: GrayLevels(2)}
	if err := BlueNoiseMultitone64x64(imgA, opts); err != nil {
		t.Fatal(err)
	}
	if err := BlueNoiseMultitone64x64(imgB, opts); err != nil {
		t.Fatal(err)
	}
	if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
		t.Fatal("blue-noise output should be deterministic")
	}
}

func TestClusterDot16x16PaletteMembership(t *testing.T) {
	palette := Palette{{0, 0, 0}, {255, 255, 255}, {255, 0, 0}, {0, 128, 255}}
	img, err := NewPackedImage(rgbGradient(12, 12), 12, 12, RGB8)
	if err != nil {
		t.Fatal(err)
	}
	if err := ClusterDot16x16(img, Options{Quantizer: PaletteQuantizer(palette)}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(img.Pix); i += 3 {
		c := Color{img.Pix[i], img.Pix[i+1], img.Pix[i+2]}
		if !palette.Contains(c) {
			t.Fatalf("pixel %v is not in palette", c)
		}
	}
}

func TestAdaptiveOrderedDeterministic(t *testing.T) {
	imgA, _ := NewPackedImage(rgbGradient(18, 18), 18, 18, RGB8)
	imgB, _ := NewPackedImage(rgbGradient(18, 18), 18, 18, RGB8)
	opts := Options{Quantizer: RGBLevels(4)}
	if err := AdaptiveBayer16x16(imgA, opts); err != nil {
		t.Fatal(err)
	}
	if err := AdaptiveBayer16x16(imgB, opts); err != nil {
		t.Fatal(err)
	}
	if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
		t.Fatal("adaptive ordered dithering should be deterministic")
	}
}

func TestExtendedOrderedPaletteMembership(t *testing.T) {
	palette := Palette{{0, 0, 0}, {255, 255, 255}, {255, 0, 0}, {0, 128, 255}}
	cases := []struct {
		name string
		run  func(*Image) error
	}{
		{name: "stochastic-cluster", run: func(img *Image) error {
			return StochasticClusterDot16x16(img, Options{Quantizer: PaletteQuantizer(palette), Seed: 7})
		}},
		{name: "polyomino", run: func(img *Image) error { return Polyomino16x16(img, Options{Quantizer: PaletteQuantizer(palette)}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img, err := NewPackedImage(rgbGradient(12, 12), 12, 12, RGB8)
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.run(img); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < len(img.Pix); i += 3 {
				c := Color{img.Pix[i], img.Pix[i+1], img.Pix[i+2]}
				if !palette.Contains(c) {
					t.Fatalf("pixel %v is not in palette", c)
				}
			}
		})
	}
}

func TestAdditionalVariableCurvesDeterministic(t *testing.T) {
	cases := []struct {
		name string
		run  func(*Image) error
	}{
		{name: "smooth", run: func(img *Image) error { return SmoothVariable(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "punchy", run: func(img *Image) error { return PunchyVariable(img, Options{Quantizer: GrayLevels(2)}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			imgA, _ := NewPackedImage(rgbaGradient(14, 14), 14, 14, RGBA8)
			imgB, _ := NewPackedImage(rgbaGradient(14, 14), 14, 14, RGBA8)
			if err := tc.run(imgA); err != nil {
				t.Fatal(err)
			}
			if err := tc.run(imgB); err != nil {
				t.Fatal(err)
			}
			if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
				t.Fatal("variable curve output should be deterministic")
			}
		})
	}
}

func TestStepTwoOrderedDeterministic(t *testing.T) {
	cases := []struct {
		name string
		run  func(*Image) error
	}{
		{name: "dot-diffusion", run: func(img *Image) error { return DotDiffusion8x8(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "dot-diffusion-diagonal", run: func(img *Image) error { return DotDiffusionDiagonal8x8(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "blue-noise-soft", run: func(img *Image) error { return BlueNoiseSoft64x64(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "blue-noise-hard", run: func(img *Image) error { return BlueNoiseHard64x64(img, Options{Quantizer: GrayLevels(2)}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			imgA, _ := NewPackedImage(rgbGradient(16, 16), 16, 16, RGB8)
			imgB, _ := NewPackedImage(rgbGradient(16, 16), 16, 16, RGB8)
			if err := tc.run(imgA); err != nil {
				t.Fatal(err)
			}
			if err := tc.run(imgB); err != nil {
				t.Fatal(err)
			}
			if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
				t.Fatal("step two ordered mode should be deterministic")
			}
		})
	}
}

func TestAMFMModesDeterministic(t *testing.T) {
	cases := []struct {
		name string
		run  func(*Image) error
	}{
		{name: "am-fm-hybrid", run: func(img *Image) error { return AMFMHybrid64x64(img, Options{Quantizer: GrayLevels(2), Seed: 7}) }},
		{name: "clustered-am-fm", run: func(img *Image) error { return ClusteredAMFM64x64(img, Options{Quantizer: GrayLevels(2), Seed: 7}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			imgA, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
			imgB, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
			if err := tc.run(imgA); err != nil {
				t.Fatal(err)
			}
			if err := tc.run(imgB); err != nil {
				t.Fatal(err)
			}
			if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
				t.Fatal("am/fm mode should be deterministic")
			}
		})
	}
}

func TestVariableDiffusionPreservesRGBAAlpha(t *testing.T) {
	pix := rgbaGradient(12, 12)
	alpha := make([]uint8, 12*12)
	for i := range alpha {
		alpha[i] = pix[i*4+3]
	}
	img, err := NewPackedImage(pix, 12, 12, RGBA8)
	if err != nil {
		t.Fatal(err)
	}
	if err := ZhouFang(img, Options{Quantizer: GrayLevels(2), Seed: 7}); err != nil {
		t.Fatal(err)
	}
	for i := range alpha {
		if img.Pix[i*4+3] != alpha[i] {
			t.Fatalf("alpha changed at pixel %d", i)
		}
	}
}
