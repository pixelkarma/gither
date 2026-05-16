package gither

import "testing"

func TestFixtureAlgorithmsDeterministic(t *testing.T) {
	fixture := mustLoadFixtureForTest(t)
	cases := []struct {
		name string
		run  func(*Image) error
		hash uint64
	}{
		{name: "bayer-8x8", run: func(img *Image) error { return Bayer8x8(img, Options{Quantizer: RGBLevels(4)}) }, hash: 1586559976892664640},
		{name: "floyd-steinberg", run: func(img *Image) error { return FloydSteinberg(img, Options{Quantizer: RGBLevels(4)}) }, hash: 1470673958377565646},
		{name: "riemersma", run: func(img *Image) error { return Riemersma(img, Options{Quantizer: RGBLevels(4)}) }, hash: 12967790854396528878},
		{name: "balanced-variable", run: func(img *Image) error { return BalancedVariable(img, Options{Quantizer: GrayLevels(2)}) }, hash: 7898405161282167868},
		{
			name: "dbs-fast",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricFast})
			},
			hash: 16594214877598800141,
		},
		{
			name: "dbs-balanced",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced})
			},
			hash: 14165239993098868053,
		},
		{
			name: "dbs-perceptual",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricPerceptual})
			},
			hash: 16913546786584952508,
		},
		{
			name: "clustered-dbs",
			run: func(img *Image) error {
				return ClusteredDBS(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7, ClusterStrength: 0.18, ClusterToneAware: true})
			},
			hash: 1602267920084629516,
		},
		{
			name: "multilevel-dbs",
			run: func(img *Image) error {
				return MultiLevelDBS(img, DBSOptions{Levels: 4, Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
			hash: 11413843390288350397,
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
			hash: 3995122937393973226,
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
