package maps

import "sort"

var (
	Bayer2x2 = []uint16{0, 2, 3, 1}
	Bayer4x4 = []uint16{
		0, 8, 2, 10,
		12, 4, 14, 6,
		3, 11, 1, 9,
		15, 7, 13, 5,
	}
	Bayer8x8 = []uint16{
		0, 32, 8, 40, 2, 34, 10, 42,
		48, 16, 56, 24, 50, 18, 58, 26,
		12, 44, 4, 36, 14, 46, 6, 38,
		60, 28, 52, 20, 62, 30, 54, 22,
		3, 35, 11, 43, 1, 33, 9, 41,
		51, 19, 59, 27, 49, 17, 57, 25,
		15, 47, 7, 39, 13, 45, 5, 37,
		63, 31, 55, 23, 61, 29, 53, 21,
	}
	ClusterDot4x4 = []uint16{
		12, 5, 6, 13,
		4, 0, 1, 7,
		11, 3, 2, 8,
		15, 10, 9, 14,
	}
	ClusterDot8x8 = []uint16{
		24, 10, 12, 26, 35, 47, 49, 37,
		8, 0, 2, 14, 45, 59, 61, 51,
		22, 6, 4, 16, 43, 57, 63, 53,
		30, 20, 18, 28, 33, 41, 55, 39,
		34, 46, 48, 36, 25, 11, 13, 27,
		44, 58, 60, 50, 9, 1, 3, 15,
		42, 56, 62, 52, 23, 7, 5, 17,
		32, 40, 54, 38, 31, 21, 19, 29,
	}
)

func GenerateBayer16x16() []uint16 {
	out := make([]uint16, 256)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			base := Bayer8x8[(y%8)*8+(x%8)]
			quadrant := Bayer2x2[(y/8)*2+(x/8)]
			out[y*16+x] = base*4 + quadrant
		}
	}
	return out
}

func GenerateClusterDot16x16() []uint16 {
	type point struct {
		x, y    int
		distSq  int
		diamond int
		checker int
	}
	points := make([]point, 0, 16*16)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			dx2 := 2*x - 15
			dy2 := 2*y - 15
			absDx2 := dx2
			if absDx2 < 0 {
				absDx2 = -absDx2
			}
			absDy2 := dy2
			if absDy2 < 0 {
				absDy2 = -absDy2
			}
			points = append(points, point{
				x:       x,
				y:       y,
				distSq:  dx2*dx2 + dy2*dy2,
				diamond: absDx2 + absDy2,
				checker: (x + y) & 1,
			})
		}
	}
	sort.Slice(points, func(i, j int) bool {
		left, right := points[i], points[j]
		if left.distSq != right.distSq {
			return left.distSq < right.distSq
		}
		if left.diamond != right.diamond {
			return left.diamond < right.diamond
		}
		if left.checker != right.checker {
			return left.checker < right.checker
		}
		if left.y != right.y {
			return left.y < right.y
		}
		return left.x < right.x
	})
	out := make([]uint16, 256)
	for rank, pt := range points {
		out[pt.y*16+pt.x] = uint16(rank)
	}
	return out
}
