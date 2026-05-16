package maps

import (
	"math"
	"sort"

	"github.com/pixelkarma/gither/internal/mathx"
)

func GenerateHilbert16x16() []uint16 {
	out := make([]uint16, 16*16)
	for d := 0; d < 16*16; d++ {
		x, y := mathx.HilbertD2XY(16, d)
		out[y*16+x] = uint16(d)
	}
	return out
}

func GenerateVoidAndCluster64x64() []uint16 {
	const side = 64
	const total = side * side
	ranks := make([]uint16, total)
	points := make([]int, 0, total)
	used := make([]bool, total)
	state := uint64(0x6d2b79f5)
	first := int(mathx.Mix64(state) % total)
	points = append(points, first)
	used[first] = true
	ranks[first] = 0
	for rank := 1; rank < total; rank++ {
		candidates := 12 + rank/384
		if candidates > 48 {
			candidates = 48
		}
		bestIndex := -1
		bestDistance := -1
		for i := 0; i < candidates; i++ {
			state = mathx.Mix64(state + 0x9e3779b97f4a7c15 + uint64(rank*131+i))
			candidate := int(state % total)
			if used[candidate] {
				continue
			}
			dist := toroidalMinDistanceSq(candidate, points, side)
			if dist > bestDistance {
				bestDistance = dist
				bestIndex = candidate
			}
		}
		if bestIndex < 0 {
			for idx := 0; idx < total; idx++ {
				if !used[idx] {
					bestIndex = idx
					break
				}
			}
		}
		used[bestIndex] = true
		points = append(points, bestIndex)
		ranks[bestIndex] = uint16(rank)
	}
	return ranks
}

func GenerateBlueNoise64x64() []uint16 {
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
		ring := 0.5 + 0.5*math.Sin(math.Sqrt(cx*cx+cy*cy)*9.0)
		jitter := float64(mathx.Mix64(uint64((x+1)*73856093^(y+1)*19349663))&1023) / 1023.0
		scoredPixels[idx] = scored{
			index: idx,
			score: baseUnit*0.72 + cluster*0.18 + ring*0.06 + jitter*0.04,
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

func toroidalMinDistanceSq(candidate int, points []int, side int) int {
	cx := candidate % side
	cy := candidate / side
	best := math.MaxInt
	for _, point := range points {
		px := point % side
		py := point / side
		dx := absInt(cx - px)
		dy := absInt(cy - py)
		if dx > side/2 {
			dx = side - dx
		}
		if dy > side/2 {
			dy = side - dy
		}
		dist := dx*dx + dy*dy
		if dist < best {
			best = dist
		}
	}
	if best == math.MaxInt {
		return side * side
	}
	return best
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
