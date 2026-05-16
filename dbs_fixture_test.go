package gither

import (
	"testing"

	"github.com/pixelkarma/gither/internal/mathx"
)

func makeLineArtFixture(width, height int) *Image {
	pix := make([]uint8, width*height)
	for i := range pix {
		pix[i] = 255
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x == y || x == width-1-y || x == width/2 || y == height/2 {
				pix[y*width+x] = 0
			}
			if x > width/4 && x < 3*width/4 && y > height/4 && y < 3*height/4 && (x+y)%9 == 0 {
				pix[y*width+x] = 32
			}
		}
	}
	img, _ := NewPackedImage(pix, width, height, Gray8)
	return img
}

func makeLowContrastFixture(width, height int) *Image {
	pix := make([]uint8, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			base := 104 + (40*x)/(maxIntFixture(1, width-1))
			mod := ((x*3 + y*5) % 11) - 5
			pix[y*width+x] = uint8(clampFixture(base+mod, 0, 255))
		}
	}
	img, _ := NewPackedImage(pix, width, height, Gray8)
	return img
}

func makeTextureFixture(width, height int) *Image {
	pix := make([]uint8, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			v := mathx.Mix64(uint64((x+1)*73856093) ^ uint64((y+1)*19349663))
			pix[y*width+x] = uint8(48 + (v % 160))
		}
	}
	img, _ := NewPackedImage(pix, width, height, Gray8)
	return img
}

func makeColorTextureFixture(width, height int) *Image {
	pix := make([]uint8, width*height*3)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * 3
			a := mathx.Mix64(uint64((x+1)*2654435761) ^ uint64((y+1)*2246822519))
			b := mathx.Mix64(uint64((x+3)*3266489917) ^ uint64((y+7)*668265263))
			c := mathx.Mix64(uint64((x+11)*374761393) ^ uint64((y+13)*1274126177))
			pix[offset] = uint8(24 + (a % 208))
			pix[offset+1] = uint8(24 + (b % 208))
			pix[offset+2] = uint8(24 + (c % 208))
		}
	}
	img, _ := NewPackedImage(pix, width, height, RGB8)
	return img
}

func dbsFixtureSet(tb testing.TB) map[string]*Image {
	tb.Helper()
	return map[string]*Image{
		"photo":         mustLoadFixtureForTest(tb),
		"line-art":      makeLineArtFixture(64, 64),
		"low-contrast":  makeLowContrastFixture(64, 64),
		"texture":       makeTextureFixture(64, 64),
		"color-texture": makeColorTextureFixture(64, 64),
	}
}

func TestDBSFixtureSuiteRepeatable(t *testing.T) {
	fixtures := dbsFixtureSet(t)
	cases := []struct {
		name    string
		fixture string
		run     func(*Image) error
	}{
		{
			name:    "line-art/dbs-balanced",
			fixture: "line-art",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
		},
		{
			name:    "line-art/clustered-dbs",
			fixture: "line-art",
			run: func(img *Image) error {
				return ClusteredDBS(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7, ClusterStrength: 0.18, ClusterToneAware: true})
			},
		},
		{
			name:    "low-contrast/multilevel-dbs",
			fixture: "low-contrast",
			run: func(img *Image) error {
				return MultiLevelDBS(img, DBSOptions{Levels: 4, Passes: 1, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
		},
		{
			name:    "texture/dbs-balanced",
			fixture: "texture",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
		},
		{
			name:    "texture/clustered-dbs",
			fixture: "texture",
			run: func(img *Image) error {
				return ClusteredDBS(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7, ClusterStrength: 0.18, ClusterToneAware: true})
			},
		},
		{
			name:    "color-texture/color-dbs",
			fixture: "color-texture",
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
			sourceHash := hashBytes(fixtures[tc.fixture].Pix)
			imgA := fixtures[tc.fixture].Clone()
			imgB := fixtures[tc.fixture].Clone()
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
				t.Fatal("algorithm output unexpectedly matched the source fixture")
			}
		})
	}
}

func clampFixture(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func maxIntFixture(a, b int) int {
	if a > b {
		return a
	}
	return b
}
