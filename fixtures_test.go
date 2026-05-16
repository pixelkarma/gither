package gither

import "testing"

func TestFixtureAlgorithmsRepeatable(t *testing.T) {
	fixture := mustLoadFixtureForTest(t)
	sourceHash := hashBytes(fixture.Pix)
	cases := []struct {
		name string
		run  func(*Image) error
	}{
		{name: "bayer-8x8", run: func(img *Image) error { return Bayer8x8(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "floyd-steinberg", run: func(img *Image) error { return FloydSteinberg(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "riemersma", run: func(img *Image) error { return Riemersma(img, Options{Quantizer: RGBLevels(4)}) }},
		{name: "balanced-variable", run: func(img *Image) error { return BalancedVariable(img, Options{Quantizer: GrayLevels(2)}) }},
		{
			name: "dbs-fast",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricFast})
			},
		},
		{
			name: "dbs-balanced",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced})
			},
		},
		{
			name: "dbs-perceptual",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricPerceptual})
			},
		},
		{
			name: "clustered-dbs",
			run: func(img *Image) error {
				return ClusteredDBS(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7, ClusterStrength: 0.18, ClusterToneAware: true})
			},
		},
		{
			name: "multilevel-dbs",
			run: func(img *Image) error {
				return MultiLevelDBS(img, DBSOptions{Levels: 4, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
		},
		{
			name: "color-dbs",
			run: func(img *Image) error {
				palette, err := ExtractPaletteWithOptions(img, PaletteExtractOptions{Colors: 6, Method: PaletteMethodMedianCut, Sort: PaletteSortRGB})
				if err != nil {
					return err
				}
				return ColorDBS(img, DBSOptions{Palette: palette, Passes: 1, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			imgA := fixture.Clone()
			imgB := fixture.Clone()
			if err := tc.run(imgA); err != nil {
				t.Fatal(err)
			}
			if err := tc.run(imgB); err != nil {
				t.Fatal(err)
			}
			hashA := hashBytes(imgA.Pix)
			hashB := hashBytes(imgB.Pix)
			if hashA != hashB {
				t.Fatalf("repeatability mismatch: got %d and %d", hashA, hashB)
			}
			if hashA == sourceHash {
				t.Fatal("algorithm output unexpectedly matched the source image")
			}
		})
	}
}

func TestPaletteExtractionOptions(t *testing.T) {
	fixture := mustLoadFixtureForTest(t)
	median, err := ExtractPaletteWithOptions(fixture, PaletteExtractOptions{Colors: 8, Method: PaletteMethodMedianCut, Sort: PaletteSortLuma})
	if err != nil {
		t.Fatal(err)
	}
	popular, err := ExtractPaletteWithOptions(fixture, PaletteExtractOptions{Colors: 8, Method: PaletteMethodPopularity, Sort: PaletteSortFrequency})
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
