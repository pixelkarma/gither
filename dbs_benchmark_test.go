package gither

import "testing"

func BenchmarkDBSSchedulesFixtureTest(b *testing.B) {
	src := mustLoadFixtureForTest(b)
	cases := []benchmarkCase{
		{name: "preview", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 1, Threshold: 127, MoveMode: DBSMoveFlip, Neighborhood: 1, Metric: DBSMetricFast, ScanOrder: DBSScanSerpentine, RadiusPolicy: DBSRadiusFixed, MaxNoImprove: 1, RandomSeed: 7})
		}},
		{name: "balanced", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 2, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RadiusPolicy: DBSRadiusFixed, MaxNoImprove: 1, RandomSeed: 7})
		}},
		{name: "hq", run: func(img *Image) error {
			return DirectBinarySearch(img, DBSOptions{Seed: DBSSeedThreshold, Passes: 3, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricPerceptual, ScanOrder: DBSScanRandom, RadiusPolicy: DBSRadiusExpand, MaxNoImprove: 2, Restarts: 1, RandomSeed: 7})
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

func BenchmarkDBSComparisonFixtures(b *testing.B) {
	fixtures := dbsFixtureSet(b)
	for fixtureName, fixture := range fixtures {
		if fixture.Format != Gray8 {
			continue
		}
		b.Run(fixtureName, func(b *testing.B) {
			cases := []benchmarkCase{
				{name: "threshold", run: func(img *Image) error { return Threshold(img, Options{Quantizer: GrayLevels(2), Threshold: 127}) }},
				{name: "bayer-8x8", run: func(img *Image) error { return Bayer8x8(img, Options{Quantizer: GrayLevels(2)}) }},
				{name: "floyd-steinberg", run: func(img *Image) error { return FloydSteinberg(img, Options{Quantizer: GrayLevels(2)}) }},
				{name: "riemersma", run: func(img *Image) error { return Riemersma(img, Options{Quantizer: GrayLevels(2)}) }},
				{name: "dbs-balanced", run: func(img *Image) error {
					return DirectBinarySearch(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
				}},
				{name: "clustered-dbs", run: func(img *Image) error {
					return ClusteredDBS(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7, ClusterStrength: 0.18, ClusterToneAware: true})
				}},
				{name: "multilevel-dbs", run: func(img *Image) error {
					return MultiLevelDBS(img, DBSOptions{Levels: 4, Passes: 1, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
				}},
			}
			for _, bc := range cases {
				b.Run(bc.name, func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						img := fixture.Clone()
						if err := bc.run(img); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

func BenchmarkDBSColorFixtures(b *testing.B) {
	fixtures := dbsFixtureSet(b)
	for fixtureName, fixture := range fixtures {
		if fixture.Format == Gray8 {
			continue
		}
		b.Run(fixtureName, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				img := fixture.Clone()
				palette, err := ExtractPaletteWithOptions(img, PaletteExtractOptions{Colors: 6, Method: PaletteMethodMedianCut, Sort: PaletteSortRGB})
				if err != nil {
					b.Fatal(err)
				}
				if err := ColorDBS(img, DBSOptions{Palette: palette, Passes: 1, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
