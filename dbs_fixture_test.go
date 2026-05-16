package gither

import (
	"hash/fnv"
	"testing"

	"gither/internal/mathx"
)

func hashFixtureBytes(b []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(b)
	return h.Sum64()
}

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

func TestDBSFixtureSuiteDeterministic(t *testing.T) {
	fixtures := dbsFixtureSet(t)
	cases := []struct {
		name    string
		fixture string
		run     func(*Image) error
		hash    uint64
	}{
		{
			name:    "line-art/dbs-balanced",
			fixture: "line-art",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
			hash: 12968967869221384461,
		},
		{
			name:    "line-art/clustered-dbs",
			fixture: "line-art",
			run: func(img *Image) error {
				return ClusteredDBS(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7, ClusterStrength: 0.18, ClusterToneAware: true})
			},
			hash: 2932870014787304613,
		},
		{
			name:    "low-contrast/multilevel-dbs",
			fixture: "low-contrast",
			run: func(img *Image) error {
				return MultiLevelDBS(img, DBSOptions{Levels: 4, Passes: 1, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
			hash: 16333605137037632989,
		},
		{
			name:    "texture/dbs-balanced",
			fixture: "texture",
			run: func(img *Image) error {
				return DirectBinarySearch(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7})
			},
			hash: 14805613429304380296,
		},
		{
			name:    "texture/clustered-dbs",
			fixture: "texture",
			run: func(img *Image) error {
				return ClusteredDBS(img, DBSOptions{Passes: 1, Threshold: 127, MoveMode: DBSMoveHybrid, Neighborhood: 1, Metric: DBSMetricBalanced, ScanOrder: DBSScanSerpentine, RandomSeed: 7, ClusterStrength: 0.18, ClusterToneAware: true})
			},
			hash: 3542612787489216144,
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
			hash: 7766925434517295867,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img := fixtures[tc.fixture].Clone()
			if err := tc.run(img); err != nil {
				t.Fatal(err)
			}
			if got := hashFixtureBytes(img.Pix); got != tc.hash {
				t.Fatalf("hash mismatch: got %d want %d", got, tc.hash)
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
