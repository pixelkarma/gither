package gither

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

type benchmarkCase struct {
	name string
	run  func(*Image) error
}

func BenchmarkAlgorithmsFixtureCat(b *testing.B) {
	src := mustLoadFixtureImage(b)
	paletteOpts := PaletteExtractOptions{Colors: 6, Method: PaletteMethodMedianCut, Sort: PaletteSortRGB}
	autoPalette, err := ExtractPaletteWithOptions(src, paletteOpts)
	if err != nil {
		b.Fatal(err)
	}
	cases := []benchmarkCase{
		{name: "bayer-2x2", run: func(img *Image) error { return Bayer2x2(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "bayer-4x4", run: func(img *Image) error { return Bayer4x4(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "bayer-8x8", run: func(img *Image) error { return Bayer8x8(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "bayer-16x16", run: func(img *Image) error { return Bayer16x16(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "adaptive-bayer-8x8", run: func(img *Image) error { return AdaptiveBayer8x8(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "adaptive-bayer-16x16", run: func(img *Image) error { return AdaptiveBayer16x16(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "cluster-dot-4x4", run: func(img *Image) error { return ClusterDot4x4(img, Options{Quantizer: PaletteQuantizer(autoPalette)}) }},
		{name: "cluster-dot-8x8", run: func(img *Image) error { return ClusterDot8x8(img, Options{Quantizer: PaletteQuantizer(autoPalette)}) }},
		{name: "cluster-dot-16x16", run: func(img *Image) error { return ClusterDot16x16(img, Options{Quantizer: PaletteQuantizer(autoPalette)}) }},
		{name: "stochastic-cluster-dot-16x16", run: func(img *Image) error {
			return StochasticClusterDot16x16(img, Options{Quantizer: PaletteQuantizer(autoPalette), Seed: 7})
		}},
		{name: "polyomino-16x16", run: func(img *Image) error { return Polyomino16x16(img, Options{Quantizer: PaletteQuantizer(autoPalette)}) }},
		{name: "space-filling-16x16", run: func(img *Image) error { return SpaceFilling16x16(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "space-filling-morton-16x16", run: func(img *Image) error { return SpaceFillingMorton16x16(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "space-filling-serpentine-16x16", run: func(img *Image) error { return SpaceFillingSerpentine16x16(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "void-and-cluster-64x64", run: func(img *Image) error { return VoidAndCluster64x64(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "blue-noise-multitone-64x64", run: func(img *Image) error { return BlueNoiseMultitone64x64(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "blue-noise-soft-64x64", run: func(img *Image) error { return BlueNoiseSoft64x64(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "blue-noise-hard-64x64", run: func(img *Image) error { return BlueNoiseHard64x64(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "dot-diffusion-8x8", run: func(img *Image) error { return DotDiffusion8x8(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "dot-diffusion-diagonal-8x8", run: func(img *Image) error { return DotDiffusionDiagonal8x8(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "yliluoma-1", run: func(img *Image) error { return Yliluoma1(img, Options{Quantizer: PaletteQuantizer(autoPalette)}) }},
		{name: "yliluoma-2", run: func(img *Image) error { return Yliluoma2(img, Options{Quantizer: PaletteQuantizer(autoPalette)}) }},
		{name: "yliluoma-3", run: func(img *Image) error { return Yliluoma3(img, Options{Quantizer: PaletteQuantizer(autoPalette)}) }},
		{name: "floyd-steinberg", run: func(img *Image) error { return FloydSteinberg(img, Options{Quantizer: PaletteQuantizer(autoPalette)}) }},
		{name: "false-floyd-steinberg", run: func(img *Image) error { return FalseFloydSteinberg(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "jjn", run: func(img *Image) error { return JarvisJudiceNinke(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "stucki", run: func(img *Image) error { return Stucki(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "burkes", run: func(img *Image) error { return Burkes(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "sierra", run: func(img *Image) error { return Sierra(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "two-row-sierra", run: func(img *Image) error { return TwoRowSierra(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "sierra-lite", run: func(img *Image) error { return SierraLite(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "stevenson-arce", run: func(img *Image) error { return StevensonArce(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "atkinson", run: func(img *Image) error { return Atkinson(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "ostromoukhov", run: func(img *Image) error { return Ostromoukhov(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "zhou-fang", run: func(img *Image) error { return ZhouFang(img, Options{Quantizer: GrayLevels(2), Seed: 7}) }},
		{name: "balanced-variable", run: func(img *Image) error { return BalancedVariable(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "smooth-variable", run: func(img *Image) error { return SmoothVariable(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "punchy-variable", run: func(img *Image) error { return PunchyVariable(img, Options{Quantizer: GrayLevels(2)}) }},
		{name: "threshold", run: func(img *Image) error { return Threshold(img, Options{Quantizer: GrayLevels(2), Threshold: 127}) }},
		{name: "random", run: func(img *Image) error {
			return Random(img, Options{Quantizer: GrayLevels(2), Seed: 7, RandomStrength: 40})
		}},
		{name: "riemersma", run: func(img *Image) error { return Riemersma(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "am-fm-hybrid-64x64", run: func(img *Image) error { return AMFMHybrid64x64(img, Options{Quantizer: GrayLevels(2), Seed: 7}) }},
		{name: "clustered-am-fm-64x64", run: func(img *Image) error { return ClusteredAMFM64x64(img, Options{Quantizer: GrayLevels(2), Seed: 7}) }},
		{name: "dbs-fast", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       1,
				Threshold:    127,
				MoveMode:     DBSMoveHybrid,
				Neighborhood: 1,
				Metric:       DBSMetricFast,
			})
		}},
		{name: "dbs-balanced", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       1,
				Threshold:    127,
				MoveMode:     DBSMoveHybrid,
				Neighborhood: 1,
				Metric:       DBSMetricBalanced,
			})
		}},
		{name: "dbs-perceptual", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       1,
				Threshold:    127,
				MoveMode:     DBSMoveHybrid,
				Neighborhood: 1,
				Metric:       DBSMetricPerceptual,
			})
		}},
	}
	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				img := src.Clone()
				if err := bc.run(img); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDBSSchedulesFixtureCat(b *testing.B) {
	src := mustLoadFixtureImage(b)
	cases := []benchmarkCase{
		{name: "preview", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       1,
				Threshold:    127,
				MoveMode:     DBSMoveFlip,
				Neighborhood: 1,
				Metric:       DBSMetricFast,
				ScanOrder:    DBSScanSerpentine,
				RadiusPolicy: DBSRadiusFixed,
				MaxNoImprove: 1,
				RandomSeed:   7,
			})
		}},
		{name: "balanced", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       2,
				Threshold:    127,
				MoveMode:     DBSMoveHybrid,
				Neighborhood: 1,
				Metric:       DBSMetricBalanced,
				ScanOrder:    DBSScanSerpentine,
				RadiusPolicy: DBSRadiusFixed,
				MaxNoImprove: 1,
				RandomSeed:   7,
			})
		}},
		{name: "hq", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{
				Seed:         DBSSeedThreshold,
				Passes:       3,
				Threshold:    127,
				MoveMode:     DBSMoveHybrid,
				Neighborhood: 1,
				Metric:       DBSMetricPerceptual,
				ScanOrder:    DBSScanRandom,
				RadiusPolicy: DBSRadiusExpand,
				MaxNoImprove: 2,
				Restarts:     1,
				RandomSeed:   7,
			})
		}},
	}
	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				img := src.Clone()
				if err := bc.run(img); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkPaletteExtractionFixtureCat(b *testing.B) {
	src := mustLoadFixtureImage(b)
	b.Run("median-cut", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := ExtractPaletteWithOptions(src, PaletteExtractOptions{Colors: 8, Method: PaletteMethodMedianCut, Sort: PaletteSortRGB}); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("popularity", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := ExtractPaletteWithOptions(src, PaletteExtractOptions{Colors: 8, Method: PaletteMethodPopularity, Sort: PaletteSortFrequency}); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func mustLoadFixtureImage(tb testing.TB) *Image {
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
