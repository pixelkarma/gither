package maps

import (
	"math"
	"sort"

	"gither/internal/mathx"
)

type BlueNoiseProfile struct {
	BaseWeight    float64
	ClusterWeight float64
	RingWeight    float64
	JitterWeight  float64
	Frequency     float64
}

type AmFmProfile struct {
	FMWeight      float64
	AMWeight      float64
	ClusterWeight float64
	MacroWeight   float64
}

func GenerateDotDiffusion8x8() []uint16 {
	class := []uint16{
		0, 48, 12, 60, 3, 51, 15, 63,
		32, 16, 44, 28, 35, 19, 47, 31,
		8, 56, 4, 52, 11, 59, 7, 55,
		40, 24, 36, 20, 43, 27, 39, 23,
		2, 50, 14, 62, 1, 49, 13, 61,
		34, 18, 46, 30, 33, 17, 45, 29,
		10, 58, 6, 54, 9, 57, 5, 53,
		42, 26, 38, 22, 41, 25, 37, 21,
	}
	return class
}

func GenerateDotDiffusionDiagonal8x8() []uint16 {
	out := make([]uint16, 64)
	copy(out, GenerateDotDiffusion8x8())
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			i := y*8 + x
			out[i] = uint16((int(out[i]) + ((x + 2*y) & 7)) & 63)
		}
	}
	return renumberRanks(out)
}

func GenerateBlueNoise64x64WithProfile(profile BlueNoiseProfile) []uint16 {
	const side = 64
	base := GenerateVoidAndCluster64x64()
	type scored struct {
		index int
		score float64
	}
	scoredPixels := make([]scored, side*side)
	for idx := range base {
		x := idx % side
		y := idx / side
		baseUnit := float64(base[idx]) / float64(side*side-1)
		cluster := float64(ClusterDot8x8[(y%8)*8+(x%8)]) / 63.0
		cx := float64(x-side/2) / float64(side/2)
		cy := float64(y-side/2) / float64(side/2)
		ring := 0.5 + 0.5*math.Sin(math.Sqrt(cx*cx+cy*cy)*profile.Frequency)
		jitter := float64(mathx.Mix64(uint64((x+1)*73856093^(y+1)*19349663))&1023) / 1023.0
		scoredPixels[idx] = scored{
			index: idx,
			score: baseUnit*profile.BaseWeight + cluster*profile.ClusterWeight + ring*profile.RingWeight + jitter*profile.JitterWeight,
		}
	}
	sort.Slice(scoredPixels, func(i, j int) bool {
		if scoredPixels[i].score == scoredPixels[j].score {
			return scoredPixels[i].index < scoredPixels[j].index
		}
		return scoredPixels[i].score < scoredPixels[j].score
	})
	out := make([]uint16, side*side)
	for rank, item := range scoredPixels {
		out[item.index] = uint16(rank)
	}
	return out
}

func GenerateAMFM64x64(profile AmFmProfile, seed uint64) []uint16 {
	const side = 64
	const total = side * side
	base := GenerateVoidAndCluster64x64()
	type scored struct {
		index int
		score float64
	}
	scoredPixels := make([]scored, total)
	for idx := 0; idx < total; idx++ {
		x := idx % side
		y := idx / side
		fmNorm := float64(base[idx]) / float64(total-1)
		amNorm := float64(ClusterDot8x8[(y%8)*8+(x%8)]) / 63.0
		cluster := clusterProfile(x, y)
		macro := hash01(x/8, y/8, seed^0xa24b4fbe7b738d41)
		jitter := hash01(x, y, seed^0x5bf0363545fa9f59) * 1.0e-6
		scoredPixels[idx] = scored{
			index: idx,
			score: fmNorm*profile.FMWeight + amNorm*profile.AMWeight + cluster*profile.ClusterWeight + macro*profile.MacroWeight + jitter,
		}
	}
	sort.Slice(scoredPixels, func(i, j int) bool {
		if scoredPixels[i].score == scoredPixels[j].score {
			return scoredPixels[i].index < scoredPixels[j].index
		}
		return scoredPixels[i].score < scoredPixels[j].score
	})
	out := make([]uint16, total)
	for rank, item := range scoredPixels {
		out[item.index] = uint16(rank)
	}
	return out
}

func clusterProfile(x, y int) float64 {
	localX := float64(x%4) - 1.5
	localY := float64(y%4) - 1.5
	distance := math.Sqrt(localX*localX+localY*localY) * 0.47140452
	if distance >= 1 {
		return 0
	}
	return 1 - distance
}

func hash01(x, y int, seed uint64) float64 {
	mixed := mathx.Mix64(uint64(x)*0x632be59bd9b4e019 + uint64(y)*0x8cb92baa33a5049f + seed)
	return float64(mixed) / float64(^uint64(0))
}

func renumberRanks(values []uint16) []uint16 {
	type pair struct {
		index int
		value uint16
	}
	pairs := make([]pair, len(values))
	for i, value := range values {
		pairs[i] = pair{index: i, value: value}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].value == pairs[j].value {
			return pairs[i].index < pairs[j].index
		}
		return pairs[i].value < pairs[j].value
	})
	out := make([]uint16, len(values))
	for rank, pair := range pairs {
		out[pair.index] = uint16(rank)
	}
	return out
}
