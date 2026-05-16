package maps

import "gither/internal/mathx"

func GenerateStochasticCluster16x16(seed uint64) []uint16 {
	out := GenerateClusterDot16x16()
	type pair struct {
		i   int
		key uint64
	}
	pairs := make([]pair, len(out))
	for i := range out {
		x := i % 16
		y := i / 16
		pairs[i] = pair{i: i, key: mathx.Mix64(seed + uint64((x+1)*73856093) + uint64((y+1)*19349663))}
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if out[pairs[i].i] != out[pairs[j].i] {
				continue
			}
			if pairs[j].key < pairs[i].key {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	// Deterministic low-amplitude permutation within 8-rank bands.
	ranks := make([]uint16, len(out))
	for i := range ranks {
		ranks[i] = uint16(i)
	}
	for i := 0; i < len(out); i++ {
		perm := int((pairs[i].key >> 8) & 7)
		bandStart := (int(out[i]) / 8) * 8
		out[i] = uint16(bandStart + ((int(out[i]) + perm) & 7))
	}
	return out
}

func GeneratePolyomino16x16() []uint16 {
	out := make([]uint16, 16*16)
	for tileY := 0; tileY < 4; tileY++ {
		for tileX := 0; tileX < 4; tileX++ {
			baseRank := (tileY*4 + tileX) * 16
			rotation := (tileX + tileY) & 3
			for y := 0; y < 4; y++ {
				for x := 0; x < 4; x++ {
					sx, sy := rotate4(x, y, rotation)
					local := ClusterDot4x4[sy*4+sx]
					out[(tileY*4+y)*16+(tileX*4+x)] = uint16(baseRank) + local
				}
			}
		}
	}
	return out
}

func GenerateMorton16x16() []uint16 {
	out := make([]uint16, 16*16)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			out[y*16+x] = uint16(morton2D(x, y))
		}
	}
	return out
}

func GenerateSerpentine16x16() []uint16 {
	out := make([]uint16, 16*16)
	rank := 0
	for y := 0; y < 16; y++ {
		if y%2 == 0 {
			for x := 0; x < 16; x++ {
				out[y*16+x] = uint16(rank)
				rank++
			}
		} else {
			for x := 15; x >= 0; x-- {
				out[y*16+x] = uint16(rank)
				rank++
			}
		}
	}
	return out
}

func rotate4(x, y, rot int) (int, int) {
	switch rot & 3 {
	case 1:
		return 3 - y, x
	case 2:
		return 3 - x, 3 - y
	case 3:
		return y, 3 - x
	default:
		return x, y
	}
}

func morton2D(x, y int) int {
	return part1By1(x) | (part1By1(y) << 1)
}

func part1By1(v int) int {
	v &= 0xffff
	v = (v | (v << 8)) & 0x00FF00FF
	v = (v | (v << 4)) & 0x0F0F0F0F
	v = (v | (v << 2)) & 0x33333333
	v = (v | (v << 1)) & 0x55555555
	return v
}
