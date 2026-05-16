package engine

import (
	"gither/internal/core"
	"gither/internal/kernels"
	"gither/internal/maps"
	"gither/internal/mathx"
)

type DBSSeed string
type DBSMoveMode string

const (
	DBSSeedThreshold DBSSeed = "threshold"
	DBSSeedBayer     DBSSeed = "bayer"
	DBSSeedFloyd     DBSSeed = "floyd-steinberg"

	DBSMoveFlip   DBSMoveMode = "flip"
	DBSMoveSwap   DBSMoveMode = "swap"
	DBSMoveHybrid DBSMoveMode = "hybrid"
)

type DBSOptions struct {
	Seed         DBSSeed
	Passes       int
	Threshold    uint8
	MoveMode     DBSMoveMode
	Neighborhood int
}

func (o DBSOptions) withDefaults() DBSOptions {
	if o.Seed == "" {
		o.Seed = DBSSeedThreshold
	}
	if o.Passes <= 0 {
		o.Passes = 1
	}
	if o.MoveMode == "" {
		o.MoveMode = DBSMoveHybrid
	}
	if o.Neighborhood <= 0 {
		o.Neighborhood = 1
	}
	return o
}

func DirectBinarySearch(img *core.Image, opts DBSOptions) error {
	if err := img.Validate(); err != nil {
		return err
	}
	opts = opts.withDefaults()
	target := grayscalePlane(img)
	targetFiltered := filteredGrayPlane(target, img.Width, img.Height)
	binary, err := dbsSeedPlane(target, img.Width, img.Height, opts)
	if err != nil {
		return err
	}
	filtered := filteredGrayPlane(binary, img.Width, img.Height)
	for pass := 0; pass < opts.Passes; pass++ {
		improved := false
		for y := 0; y < img.Height; y++ {
			for x := 0; x < img.Width; x++ {
				idx := y*img.Width + x
				original := binary[idx]
				bestMode, bestSwapX, bestSwapY, bestBefore, bestAfter := dbsBestMove(binary, filtered, targetFiltered, img.Width, img.Height, x, y, opts)
				if bestAfter+1e-9 < bestBefore {
					improved = true
					switch bestMode {
					case DBSMoveFlip:
						delta := dbsPixelDelta(original)
						binary[idx] = original ^ 0xff
						applyFilteredDelta(filtered, img.Width, img.Height, x, y, delta)
					case DBSMoveSwap:
						dbsApplySwap(binary, filtered, img.Width, img.Height, x, y, bestSwapX, bestSwapY)
					}
				}
			}
		}
		if !improved {
			break
		}
	}
	writeBinaryToImage(img, binary)
	return nil
}

func grayscalePlane(img *core.Image) []uint8 {
	out := make([]uint8, img.Width*img.Height)
	channels := img.ChannelCount()
	for y := 0; y < img.Height; y++ {
		row := img.Row(y)
		for x := 0; x < img.Width; x++ {
			offset := x * channels
			switch img.Format {
			case core.Gray8:
				out[y*img.Width+x] = row[offset]
			default:
				out[y*img.Width+x] = mathx.LumaByte(row[offset], row[offset+1], row[offset+2])
			}
		}
	}
	return out
}

func filteredGrayPlane(gray []uint8, width, height int) []float32 {
	out := make([]float32, len(gray))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out[y*width+x] = filteredBinaryValue(gray, width, height, x, y)
		}
	}
	return out
}

func dbsSeedPlane(target []uint8, width, height int, opts DBSOptions) ([]uint8, error) {
	seed := append([]uint8(nil), target...)
	img, err := core.NewPackedImage(seed, width, height, core.Gray8)
	if err != nil {
		return nil, err
	}
	baseOpts := core.Options{Quantizer: core.GrayLevels(2), Threshold: opts.Threshold}
	switch opts.Seed {
	case DBSSeedThreshold:
		if err := Threshold(img, baseOpts); err != nil {
			return nil, err
		}
	case DBSSeedBayer:
		ordered := OrderedMap{Values: maps.Bayer8x8, Width: 8, Height: 8, Strength: core.DefaultOrderedStrength}
		if err := ApplyOrdered(img, ordered, baseOpts); err != nil {
			return nil, err
		}
	case DBSSeedFloyd:
		if err := ApplyDiffusion(img, baseOpts, kernels.FloydSteinberg); err != nil {
			return nil, err
		}
	default:
		if err := Threshold(img, baseOpts); err != nil {
			return nil, err
		}
	}
	return img.Pix, nil
}

func dbsBestMove(binary []uint8, filtered, targetFiltered []float32, width, height, x, y int, opts DBSOptions) (DBSMoveMode, int, int, float64, float64) {
	bestMode := DBSMoveFlip
	bestSwapX, bestSwapY := -1, -1
	bestBefore, bestAfter := 0.0, 0.0
	hasCandidate := false
	if opts.MoveMode == DBSMoveFlip || opts.MoveMode == DBSMoveHybrid {
		delta := dbsPixelDelta(binary[y*width+x])
		before, after := localFilteredErrorDelta(filtered, targetFiltered, width, height, x, y, delta)
		bestBefore, bestAfter = before, after
		hasCandidate = true
	}
	if opts.MoveMode == DBSMoveSwap || opts.MoveMode == DBSMoveHybrid {
		swapBefore, swapAfter, sx, sy, ok := bestSwapCandidate(binary, filtered, targetFiltered, width, height, x, y, opts.Neighborhood)
		if ok && (!hasCandidate || swapAfter < bestAfter) {
			bestMode = DBSMoveSwap
			bestSwapX, bestSwapY = sx, sy
			bestBefore, bestAfter = swapBefore, swapAfter
			hasCandidate = true
		}
	}
	if !hasCandidate {
		return DBSMoveFlip, -1, -1, 0, 0
	}
	return bestMode, bestSwapX, bestSwapY, bestBefore, bestAfter
}

func bestSwapCandidate(binary []uint8, filtered, targetFiltered []float32, width, height, x, y, radius int) (float64, float64, int, int, bool) {
	origin := binary[y*width+x]
	bestBefore, bestAfter := 0.0, 0.0
	bestX, bestY := -1, -1
	found := false
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := x + dx
			ny := y + dy
			if nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			other := binary[ny*width+nx]
			if other == origin {
				continue
			}
			before, after := localSwapErrorDelta(filtered, targetFiltered, width, height, x, y, nx, ny, dbsPixelDelta(origin), dbsPixelDelta(other))
			if !found || after < bestAfter {
				bestBefore, bestAfter = before, after
				bestX, bestY = nx, ny
				found = true
			}
		}
	}
	return bestBefore, bestAfter, bestX, bestY, found
}

func localFilteredErrorDelta(filtered, targetFiltered []float32, width, height, x, y int, delta float32) (float64, float64) {
	var before, after float64
	for yy := maxDBSInt(0, y-1); yy <= minDBSInt(height-1, y+1); yy++ {
		for xx := maxDBSInt(0, x-1); xx <= minDBSInt(width-1, x+1); xx++ {
			idx := yy*width + xx
			current := filtered[idx]
			diffBefore := float64(current - targetFiltered[idx])
			diffAfter := float64(current + delta*dbsKernelContribution(xx-x, yy-y) - targetFiltered[idx])
			before += diffBefore * diffBefore
			after += diffAfter * diffAfter
		}
	}
	return before, after
}

func filteredBinaryValue(binary []uint8, width, height, x, y int) float32 {
	var sum float32
	for ky := -1; ky <= 1; ky++ {
		yy := clampDBSInt(y+ky, 0, height-1)
		for kx := -1; kx <= 1; kx++ {
			xx := clampDBSInt(x+kx, 0, width-1)
			weight := dbsKernelWeight(kx, ky)
			sum += mathx.ByteToUnit(binary[yy*width+xx]) * weight
		}
	}
	return sum / 16.0
}

func applyFilteredDelta(filtered []float32, width, height, x, y int, delta float32) {
	for yy := maxDBSInt(0, y-1); yy <= minDBSInt(height-1, y+1); yy++ {
		for xx := maxDBSInt(0, x-1); xx <= minDBSInt(width-1, x+1); xx++ {
			filtered[yy*width+xx] += delta * dbsKernelContribution(xx-x, yy-y)
		}
	}
}

func localSwapErrorDelta(filtered, targetFiltered []float32, width, height, x1, y1, x2, y2 int, delta1, delta2 float32) (float64, float64) {
	minX := maxDBSInt(0, minDBSInt(x1, x2)-1)
	maxX := minDBSInt(width-1, maxDBSInt(x1, x2)+1)
	minY := maxDBSInt(0, minDBSInt(y1, y2)-1)
	maxY := minDBSInt(height-1, maxDBSInt(y1, y2)+1)
	var before, after float64
	for yy := minY; yy <= maxY; yy++ {
		for xx := minX; xx <= maxX; xx++ {
			idx := yy*width + xx
			current := filtered[idx]
			updated := current
			if absDBSInt(xx-x1) <= 1 && absDBSInt(yy-y1) <= 1 {
				updated += delta1 * dbsKernelContribution(xx-x1, yy-y1)
			}
			if absDBSInt(xx-x2) <= 1 && absDBSInt(yy-y2) <= 1 {
				updated += delta2 * dbsKernelContribution(xx-x2, yy-y2)
			}
			diffBefore := float64(current - targetFiltered[idx])
			diffAfter := float64(updated - targetFiltered[idx])
			before += diffBefore * diffBefore
			after += diffAfter * diffAfter
		}
	}
	return before, after
}

func dbsApplySwap(binary []uint8, filtered []float32, width, height, x1, y1, x2, y2 int) {
	idx1 := y1*width + x1
	idx2 := y2*width + x2
	v1 := binary[idx1]
	v2 := binary[idx2]
	if v1 == v2 {
		return
	}
	delta1 := dbsPixelDelta(v1)
	delta2 := dbsPixelDelta(v2)
	binary[idx1], binary[idx2] = v2, v1
	applyFilteredDelta(filtered, width, height, x1, y1, delta1)
	applyFilteredDelta(filtered, width, height, x2, y2, delta2)
}

func dbsKernelWeight(kx, ky int) float32 {
	if kx == 0 && ky == 0 {
		return 4
	}
	if kx == 0 || ky == 0 {
		return 2
	}
	return 1
}

func dbsKernelContribution(kx, ky int) float32 {
	return dbsKernelWeight(kx, ky) / 16.0
}

func dbsPixelDelta(original uint8) float32 {
	if original == 0 {
		return 1.0
	}
	return -1.0
}

func writeBinaryToImage(img *core.Image, binary []uint8) {
	channels := img.ChannelCount()
	for y := 0; y < img.Height; y++ {
		row := img.Row(y)
		for x := 0; x < img.Width; x++ {
			v := binary[y*img.Width+x]
			offset := x * channels
			switch img.Format {
			case core.Gray8:
				row[offset] = v
			case core.RGB8:
				row[offset], row[offset+1], row[offset+2] = v, v, v
			case core.RGBA8:
				alpha := row[offset+3]
				row[offset], row[offset+1], row[offset+2], row[offset+3] = v, v, v, alpha
			}
		}
	}
}

func clampDBSInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func minDBSInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxDBSInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absDBSInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
