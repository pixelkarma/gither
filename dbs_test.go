package gither

import "testing"

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
		t.Fatal("expected DBS report to record pass count")
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

func TestDBSFamilyPreservesRGBAAlpha(t *testing.T) {
	pix := rgbaGradient(16, 16)
	alpha := make([]uint8, 16*16)
	for i := range alpha {
		alpha[i] = pix[i*4+3]
	}
	palette := Palette{
		{R: 16, G: 20, B: 24},
		{R: 235, G: 228, B: 210},
		{R: 210, G: 100, B: 48},
		{R: 52, G: 92, B: 120},
	}
	cases := []struct {
		name string
		run  func(*Image) error
	}{
		{
			name: "dbs",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{
					Passes:       1,
					Threshold:    127,
					MoveMode:     DBSMoveHybrid,
					Neighborhood: 1,
					Metric:       DBSMetricBalanced,
					ScanOrder:    DBSScanSerpentine,
					RandomSeed:   7,
				})
			},
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
		},
		{
			name: "multilevel-dbs",
			run: func(img *Image) error {
				return MultiLevelDBS(img, DBSOptions{
					Levels:       4,
					Passes:       1,
					MoveMode:     DBSMoveHybrid,
					Neighborhood: 1,
					Metric:       DBSMetricBalanced,
					ScanOrder:    DBSScanSerpentine,
					RandomSeed:   7,
				})
			},
		},
		{
			name: "color-dbs",
			run: func(img *Image) error {
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
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img, err := NewPackedImage(append([]uint8(nil), pix...), 16, 16, RGBA8)
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.run(img); err != nil {
				t.Fatal(err)
			}
			for i := range alpha {
				if img.Pix[i*4+3] != alpha[i] {
					t.Fatalf("alpha changed at pixel %d", i)
				}
			}
		})
	}
}
