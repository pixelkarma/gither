package maps

import (
	"math"
	"sort"

	"github.com/pixelkarma/gither/internal/mathx"
)

const (
	voidClusterSide        = 64
	voidClusterLen         = voidClusterSide * voidClusterSide
	voidClusterRadius      = 7
	voidClusterInitialOnes = voidClusterLen / 8
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
	kernel := voidClusterKernel()
	pattern := voidClusterInitialPattern()
	voidClusterRelax(&pattern, kernel)

	ranks := make([]uint16, voidClusterLen)
	for i := range ranks {
		ranks[i] = math.MaxUint16
	}
	assigned := make([]bool, voidClusterLen)

	phase1Pattern := pattern
	phase1Field := voidClusterField(&phase1Pattern, kernel)
	ones := 0
	for _, value := range phase1Pattern {
		if value {
			ones++
		}
	}
	for rank := ones - 1; rank >= 0; rank-- {
		idx := voidClusterFindCluster(&phase1Pattern, phase1Field)
		if idx < 0 {
			break
		}
		ranks[idx] = uint16(rank)
		assigned[idx] = true
		voidClusterSet(&phase1Pattern, phase1Field, kernel, idx, false)
	}

	phase2Pattern := pattern
	phase2Field := voidClusterField(&phase2Pattern, kernel)
	for rank := ones; rank < voidClusterLen; rank++ {
		idx := voidClusterFindVoid(&phase2Pattern, phase2Field)
		if idx < 0 {
			break
		}
		ranks[idx] = uint16(rank)
		assigned[idx] = true
		voidClusterSet(&phase2Pattern, phase2Field, kernel, idx, true)
	}

	used := make([]bool, voidClusterLen)
	for _, rank := range ranks {
		if rank != math.MaxUint16 {
			used[rank] = true
		}
	}
	next := 0
	for idx := range ranks {
		if assigned[idx] {
			continue
		}
		for next < voidClusterLen && used[next] {
			next++
		}
		if next < voidClusterLen {
			ranks[idx] = uint16(next)
			used[next] = true
		}
	}

	return ranks
}

func GenerateBlueNoise64x64() []uint16 {
	base := GenerateVoidAndCluster64x64()
	type scored struct {
		index int
		score float64
	}
	scoredPixels := make([]scored, voidClusterLen)
	for idx := range base {
		x := idx % voidClusterSide
		y := idx / voidClusterSide
		baseUnit := float64(base[idx]) / float64(voidClusterLen-1)
		cluster := float64(ClusterDot8x8[(y%8)*8+(x%8)]) / 63.0
		ring := blueNoiseMicroRingValue(x, y)
		jitter := blueNoiseHash01(x, y, 0x9c4f2a3e5d17801b) * 1.0e-6
		scoredPixels[idx] = scored{
			index: idx,
			score: (baseUnit * 0.86) + (cluster * 0.09) + (ring * 0.05) + jitter,
		}
	}
	sort.Slice(scoredPixels, func(i, j int) bool {
		if scoredPixels[i].score == scoredPixels[j].score {
			return blueNoiseCellHash(scoredPixels[i].index) < blueNoiseCellHash(scoredPixels[j].index)
		}
		return scoredPixels[i].score < scoredPixels[j].score
	})
	out := make([]uint16, voidClusterLen)
	for rank, item := range scoredPixels {
		out[item.index] = uint16(rank)
	}
	return out
}

func voidClusterInitialPattern() [voidClusterLen]bool {
	var ranked [][2]uint64
	ranked = make([][2]uint64, 0, voidClusterLen)
	for idx := 0; idx < voidClusterLen; idx++ {
		ranked = append(ranked, [2]uint64{voidClusterHash(idx), uint64(idx)})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i][0] == ranked[j][0] {
			return ranked[i][1] < ranked[j][1]
		}
		return ranked[i][0] < ranked[j][0]
	})
	var pattern [voidClusterLen]bool
	for _, entry := range ranked[:voidClusterInitialOnes] {
		pattern[entry[1]] = true
	}
	return pattern
}

func voidClusterHash(idx int) uint64 {
	x := uint64(idx % voidClusterSide)
	y := uint64(idx / voidClusterSide)
	mixed := x*0x9e3779b185ebca87 + y*0xc2b2ae3d27d4eb4f + 0x94d049bb133111eb
	return mathx.Mix64(mixed)
}

func voidClusterKernel() [][3]int64 {
	kernel := make([][3]int64, 0, 200)
	radiusSq := voidClusterRadius * voidClusterRadius
	for dy := -voidClusterRadius; dy <= voidClusterRadius; dy++ {
		for dx := -voidClusterRadius; dx <= voidClusterRadius; dx++ {
			distanceSq := dx*dx + dy*dy
			if distanceSq > radiusSq {
				continue
			}
			weight := int64(1024 / (1 + distanceSq))
			if weight < 1 {
				weight = 1
			}
			kernel = append(kernel, [3]int64{int64(dx), int64(dy), weight})
		}
	}
	return kernel
}

func voidClusterField(pattern *[voidClusterLen]bool, kernel [][3]int64) []int64 {
	field := make([]int64, voidClusterLen)
	for idx, value := range pattern {
		if value {
			voidClusterAccumulate(field, kernel, idx, 1)
		}
	}
	return field
}

func voidClusterRelax(pattern *[voidClusterLen]bool, kernel [][3]int64) {
	field := voidClusterField(pattern, kernel)
	for i := 0; i < voidClusterLen*4; i++ {
		cluster := voidClusterFindCluster(pattern, field)
		if cluster < 0 {
			break
		}
		voidIdx := voidClusterFindVoid(pattern, field)
		if voidIdx < 0 {
			break
		}
		if field[cluster] <= field[voidIdx] {
			break
		}
		voidClusterSet(pattern, field, kernel, cluster, false)
		voidClusterSet(pattern, field, kernel, voidIdx, true)
	}
}

func voidClusterSet(pattern *[voidClusterLen]bool, field []int64, kernel [][3]int64, idx int, value bool) {
	if pattern[idx] == value {
		return
	}
	pattern[idx] = value
	delta := int64(-1)
	if value {
		delta = 1
	}
	voidClusterAccumulate(field, kernel, idx, delta)
}

func voidClusterAccumulate(field []int64, kernel [][3]int64, idx int, delta int64) {
	x := idx % voidClusterSide
	y := idx / voidClusterSide
	for _, entry := range kernel {
		nx := voidClusterWrap(x + int(entry[0]))
		ny := voidClusterWrap(y + int(entry[1]))
		target := ny*voidClusterSide + nx
		field[target] += delta * entry[2]
	}
}

func voidClusterFindCluster(pattern *[voidClusterLen]bool, field []int64) int {
	bestIdx := -1
	bestValue := int64(math.MinInt64)
	for idx, value := range pattern {
		if !value {
			continue
		}
		score := field[idx]
		if score > bestValue || (score == bestValue && (bestIdx < 0 || idx < bestIdx)) {
			bestValue = score
			bestIdx = idx
		}
	}
	return bestIdx
}

func voidClusterFindVoid(pattern *[voidClusterLen]bool, field []int64) int {
	bestIdx := -1
	bestValue := int64(math.MaxInt64)
	for idx, value := range pattern {
		if value {
			continue
		}
		score := field[idx]
		if score < bestValue || (score == bestValue && (bestIdx < 0 || idx < bestIdx)) {
			bestValue = score
			bestIdx = idx
		}
	}
	return bestIdx
}

func voidClusterWrap(value int) int {
	value %= voidClusterSide
	if value < 0 {
		value += voidClusterSide
	}
	return value
}

func blueNoiseMicroRingValue(x, y int) float64 {
	tx := float64(x%4) - 1.5
	ty := float64(y%4) - 1.5
	return math.Min(math.Sqrt(tx*tx+ty*ty)*0.47140452, 1.0)
}

func blueNoiseCellHash(idx int) uint64 {
	x := uint64(idx % voidClusterSide)
	y := uint64(idx / voidClusterSide)
	return mathx.Mix64(x*0x9e3779b185ebca87 + y*0xc2b2ae3d27d4eb4f + 0x626c75656e6f6973)
}

func blueNoiseHash01(x, y int, seed uint64) float64 {
	mixed := mathx.Mix64(uint64(x)*0x632be59bd9b4e019 + uint64(y)*0x8cb92baa33a5049f + seed)
	return float64(mixed) / float64(^uint64(0))
}
