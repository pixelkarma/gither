package gither

import (
	"hash/fnv"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
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

func TestDBSDeterministic(t *testing.T) {
	seeds := []DBSSeed{DBSSeedThreshold, DBSSeedBayer, DBSSeedFloyd}
	for _, seed := range seeds {
		t.Run(string(seed), func(t *testing.T) {
			imgA, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
			imgB, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
			opts := DBSOptions{Seed: seed, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced}
			if err := DirectBinarySearch(imgA, opts); err != nil {
				t.Fatal(err)
			}
			if err := DirectBinarySearch(imgB, opts); err != nil {
				t.Fatal(err)
			}
			if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
				t.Fatal("DBS output should be deterministic")
			}
			for _, v := range imgA.Pix {
				if v != 0 && v != 255 {
					t.Fatalf("unexpected binary value %d", v)
				}
			}
		})
	}
}

func TestDBSMoveModesDeterministic(t *testing.T) {
	modes := []DBSMoveMode{DBSMoveFlip, DBSMoveSwap, DBSMoveHybrid}
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			imgA, _ := NewPackedImage(grayRamp(12, 12), 12, 12, Gray8)
			imgB, _ := NewPackedImage(grayRamp(12, 12), 12, 12, Gray8)
			opts := DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       1,
				Threshold:    127,
				MoveMode:     mode,
				Neighborhood: 1,
				Metric:       DBSMetricBalanced,
			}
			if err := DirectBinarySearch(imgA, opts); err != nil {
				t.Fatal(err)
			}
			if err := DirectBinarySearch(imgB, opts); err != nil {
				t.Fatal(err)
			}
			if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
				t.Fatal("DBS move mode should be deterministic")
			}
		})
	}
}

func TestDBSMetricsDeterministic(t *testing.T) {
	metrics := []DBSMetric{DBSMetricFast, DBSMetricBalanced, DBSMetricPerceptual}
	for _, metric := range metrics {
		t.Run(string(metric), func(t *testing.T) {
			imgA, _ := NewPackedImage(grayRamp(12, 12), 12, 12, Gray8)
			imgB, _ := NewPackedImage(grayRamp(12, 12), 12, 12, Gray8)
			opts := DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       1,
				Threshold:    127,
				MoveMode:     DBSMoveHybrid,
				Neighborhood: 1,
				Metric:       metric,
			}
			if err := DirectBinarySearch(imgA, opts); err != nil {
				t.Fatal(err)
			}
			if err := DirectBinarySearch(imgB, opts); err != nil {
				t.Fatal(err)
			}
			if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
				t.Fatal("DBS metric should be deterministic")
			}
		})
	}
}

func TestDBSScanOrdersDeterministic(t *testing.T) {
	orders := []DBSScanOrder{DBSScanRaster, DBSScanSerpentine, DBSScanRandom}
	for _, order := range orders {
		t.Run(string(order), func(t *testing.T) {
			imgA, _ := NewPackedImage(grayRamp(12, 12), 12, 12, Gray8)
			imgB, _ := NewPackedImage(grayRamp(12, 12), 12, 12, Gray8)
			opts := DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       2,
				Threshold:    127,
				MoveMode:     DBSMoveHybrid,
				Neighborhood: 1,
				Metric:       DBSMetricBalanced,
				ScanOrder:    order,
				RandomSeed:   7,
			}
			if err := DirectBinarySearch(imgA, opts); err != nil {
				t.Fatal(err)
			}
			if err := DirectBinarySearch(imgB, opts); err != nil {
				t.Fatal(err)
			}
			if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
				t.Fatal("DBS scan order should be deterministic when seed is fixed")
			}
		})
	}
}

func TestDBSReportPopulated(t *testing.T) {
	img, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
	report := DBSReport{}
	err := DirectBinarySearch(img, DBSOptions{
		Seed:         DBSSeedThreshold,
		Passes:       2,
		Threshold:    127,
		MoveMode:     DBSMoveHybrid,
		Neighborhood: 1,
		Metric:       DBSMetricBalanced,
		ScanOrder:    DBSScanSerpentine,
		RadiusPolicy: DBSRadiusExpand,
		MaxNoImprove: 1,
		Restarts:     1,
		RandomSeed:   5,
		Report:       &report,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.PassesRun == 0 {
		t.Fatal("expected DB S report to record pass count")
	}
	if report.AcceptedMoves == 0 {
		t.Fatal("expected DBS report to record accepted moves")
	}
}

func TestClusteredDBSDeterministic(t *testing.T) {
	imgA, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
	imgB, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
	opts := DBSOptions{
		Passes:           1,
		Threshold:        127,
		MoveMode:         DBSMoveHybrid,
		Neighborhood:     1,
		Metric:           DBSMetricBalanced,
		ClusterStrength:  0.18,
		ClusterToneAware: true,
		ScanOrder:        DBSScanSerpentine,
		RandomSeed:       7,
	}
	if err := ClusteredDBS(imgA, opts); err != nil {
		t.Fatal(err)
	}
	if err := ClusteredDBS(imgB, opts); err != nil {
		t.Fatal(err)
	}
	if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
		t.Fatal("clustered DBS should be deterministic")
	}
}

func TestClusteredDBSDiffersFromDispersed(t *testing.T) {
	src := grayRamp(24, 24)
	dispersed, _ := NewPackedImage(append([]uint8(nil), src...), 24, 24, Gray8)
	clustered, _ := NewPackedImage(append([]uint8(nil), src...), 24, 24, Gray8)
	opts := DBSOptions{
		Passes:       1,
		Threshold:    127,
		MoveMode:     DBSMoveHybrid,
		Neighborhood: 1,
		Metric:       DBSMetricBalanced,
		ScanOrder:    DBSScanSerpentine,
		RandomSeed:   7,
	}
	if err := DirectBinarySearch(dispersed, opts); err != nil {
		t.Fatal(err)
	}
	clusteredOpts := opts
	clusteredOpts.ClusterStrength = 0.18
	clusteredOpts.ClusterToneAware = true
	if err := ClusteredDBS(clustered, clusteredOpts); err != nil {
		t.Fatal(err)
	}
	if hashBytes(dispersed.Pix) == hashBytes(clustered.Pix) {
		t.Fatal("clustered DBS should differ from dispersed DBS")
	}
}

func TestMultiLevelDBSDeterministic(t *testing.T) {
	imgA, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
	imgB, _ := NewPackedImage(grayRamp(16, 16), 16, 16, Gray8)
	opts := DBSOptions{
		Levels:       4,
		Passes:       1,
		Threshold:    127,
		MoveMode:     DBSMoveHybrid,
		Neighborhood: 1,
		Metric:       DBSMetricBalanced,
		ScanOrder:    DBSScanSerpentine,
		RandomSeed:   7,
	}
	if err := MultiLevelDBS(imgA, opts); err != nil {
		t.Fatal(err)
	}
	if err := MultiLevelDBS(imgB, opts); err != nil {
		t.Fatal(err)
	}
	if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
		t.Fatal("multilevel DBS should be deterministic")
	}
}

func TestMultiLevelDBSUsesMultipleLevels(t *testing.T) {
	img, _ := NewPackedImage(grayRamp(24, 24), 24, 24, Gray8)
	opts := DBSOptions{
		Levels:       4,
		Passes:       1,
		Threshold:    127,
		MoveMode:     DBSMoveHybrid,
		Neighborhood: 1,
		Metric:       DBSMetricBalanced,
		ScanOrder:    DBSScanSerpentine,
		RandomSeed:   7,
	}
	if err := MultiLevelDBS(img, opts); err != nil {
		t.Fatal(err)
	}
	seen := map[uint8]struct{}{}
	for _, v := range img.Pix {
		seen[v] = struct{}{}
	}
	if len(seen) < 3 {
		t.Fatalf("expected multilevel DBS to produce 3+ levels, got %d", len(seen))
	}
}

func TestColorDBSDeterministic(t *testing.T) {
	palette := Palette{
		{R: 16, G: 20, B: 24},
		{R: 235, G: 228, B: 210},
		{R: 210, G: 100, B: 48},
		{R: 52, G: 92, B: 120},
	}
	imgA, _ := NewPackedImage(rgbGradient(16, 16), 16, 16, RGB8)
	imgB, _ := NewPackedImage(rgbGradient(16, 16), 16, 16, RGB8)
	opts := DBSOptions{
		Palette:      palette,
		Passes:       1,
		MoveMode:     DBSMoveHybrid,
		Neighborhood: 1,
		Metric:       DBSMetricBalanced,
		ScanOrder:    DBSScanSerpentine,
		RandomSeed:   7,
	}
	if err := ColorDBS(imgA, opts); err != nil {
		t.Fatal(err)
	}
	if err := ColorDBS(imgB, opts); err != nil {
		t.Fatal(err)
	}
	if hashBytes(imgA.Pix) != hashBytes(imgB.Pix) {
		t.Fatal("color DBS should be deterministic")
	}
}

func TestColorDBSUsesPaletteColors(t *testing.T) {
	palette := Palette{
		{R: 16, G: 20, B: 24},
		{R: 235, G: 228, B: 210},
		{R: 210, G: 100, B: 48},
		{R: 52, G: 92, B: 120},
	}
	img, _ := NewPackedImage(rgbGradient(18, 18), 18, 18, RGB8)
	opts := DBSOptions{
		Palette:      palette,
		Passes:       1,
		MoveMode:     DBSMoveHybrid,
		Neighborhood: 1,
		Metric:       DBSMetricBalanced,
		ScanOrder:    DBSScanSerpentine,
		RandomSeed:   7,
	}
	if err := ColorDBS(img, opts); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(img.Pix); i += 3 {
		c := Color{R: img.Pix[i], G: img.Pix[i+1], B: img.Pix[i+2]}
		if !palette.Contains(c) {
			t.Fatalf("output color %#v not found in palette", c)
		}
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

func TestFixtureAlgorithmsDeterministic(t *testing.T) {
	fixture := mustLoadFixtureForTest(t)
	cases := []struct {
		name string
		run  func(*Image) error
		hash uint64
	}{
		{
			name: "bayer-8x8",
			run:  func(img *Image) error { return Bayer8x8(img, Options{Quantizer: RGBLevels(4)}) },
			hash: 2059834114531819937,
		},
		{
			name: "floyd-steinberg",
			run:  func(img *Image) error { return FloydSteinberg(img, Options{Quantizer: RGBLevels(4)}) },
			hash: 16790227691806807573,
		},
		{
			name: "riemersma",
			run:  func(img *Image) error { return Riemersma(img, Options{Quantizer: RGBLevels(4)}) },
			hash: 12317244182241967815,
		},
		{
			name: "balanced-variable",
			run:  func(img *Image) error { return BalancedVariable(img, Options{Quantizer: GrayLevels(2)}) },
			hash: 4432248508225494701,
		},
		{
			name: "dbs-fast",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{
					Seed:         DBSSeedThreshold,
					Passes:       1,
					Threshold:    127,
					MoveMode:     DBSMoveHybrid,
					Neighborhood: 1,
					Metric:       DBSMetricFast,
				})
			},
			hash: 14760988332519077620,
		},
		{
			name: "dbs-balanced",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{
					Seed:         DBSSeedThreshold,
					Passes:       1,
					Threshold:    127,
					MoveMode:     DBSMoveHybrid,
					Neighborhood: 1,
					Metric:       DBSMetricBalanced,
				})
			},
			hash: 7371833554584279101,
		},
		{
			name: "dbs-perceptual",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{
					Seed:         DBSSeedThreshold,
					Passes:       1,
					Threshold:    127,
					MoveMode:     DBSMoveHybrid,
					Neighborhood: 1,
					Metric:       DBSMetricPerceptual,
				})
			},
			hash: 7225122755161863901,
		},
		{
			name: "clustered-dbs",
			run: func(img *Image) error {
				return ClusteredDBS(img, DBSOptions{
					Passes:           1,
					Threshold:        127,
					MoveMode:         DBSMoveHybrid,
					Neighborhood:     1,
					Metric:           DBSMetricBalanced,
					ScanOrder:        DBSScanSerpentine,
					RandomSeed:       7,
					ClusterStrength:  0.18,
					ClusterToneAware: true,
				})
			},
			hash: 10849702519006246172,
		},
		{
			name: "multilevel-dbs",
			run: func(img *Image) error {
				return MultiLevelDBS(img, DBSOptions{
					Levels:       4,
					Passes:       1,
					Threshold:    127,
					MoveMode:     DBSMoveHybrid,
					Neighborhood: 1,
					Metric:       DBSMetricBalanced,
					ScanOrder:    DBSScanSerpentine,
					RandomSeed:   7,
				})
			},
			hash: 10788990793269373098,
		},
		{
			name: "color-dbs",
			run: func(img *Image) error {
				palette, err := ExtractPaletteWithOptions(img, PaletteExtractOptions{
					Colors: 6,
					Method: PaletteMethodMedianCut,
					Sort:   PaletteSortRGB,
				})
				if err != nil {
					return err
				}
				return ColorDBS(img, DBSOptions{
					Palette:      palette,
					Passes:       1,
					MoveMode:     DBSMoveHybrid,
					Neighborhood: 1,
					Metric:       DBSMetricBalanced,
					ScanOrder:    DBSScanSerpentine,
					RandomSeed:   7,
				})
			},
			hash: 15234579066110475956,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img := fixture.Clone()
			if err := tc.run(img); err != nil {
				t.Fatal(err)
			}
			if got := hashBytes(img.Pix); got != tc.hash {
				t.Fatalf("hash mismatch: got %d want %d", got, tc.hash)
			}
		})
	}
}

func TestPaletteExtractionOptions(t *testing.T) {
	fixture := mustLoadFixtureForTest(t)
	median, err := ExtractPaletteWithOptions(fixture, PaletteExtractOptions{
		Colors: 8,
		Method: PaletteMethodMedianCut,
		Sort:   PaletteSortLuma,
	})
	if err != nil {
		t.Fatal(err)
	}
	popular, err := ExtractPaletteWithOptions(fixture, PaletteExtractOptions{
		Colors: 8,
		Method: PaletteMethodPopularity,
		Sort:   PaletteSortFrequency,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(median) != 8 || len(popular) != 8 {
		t.Fatalf("unexpected palette lengths: median=%d popularity=%d", len(median), len(popular))
	}
	if median[0] == popular[0] && median[len(median)-1] == popular[len(popular)-1] {
		t.Fatal("expected palette methods to produce meaningfully different orderings")
	}
}

func mustLoadFixtureForTest(tb testing.TB) *Image {
	tb.Helper()
	path := filepath.Join("/Users/admin/Documents/dither/gither", "images", "cat.png")
	file, err := os.Open(path)
	if err != nil {
		tb.Fatal(err)
	}
	defer file.Close()
	src, _, err := image.Decode(file)
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
