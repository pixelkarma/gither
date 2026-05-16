package engine

import (
	"math/rand"

	"gither/internal/core"
	"gither/internal/kernels"
	"gither/internal/maps"
	"gither/internal/mathx"
)

type DBSSeed string
type DBSMoveMode string
type DBSMetric string
type DBSScanOrder string
type DBSRadiusPolicy string

const (
	DBSSeedThreshold DBSSeed = "threshold"
	DBSSeedBayer     DBSSeed = "bayer"
	DBSSeedFloyd     DBSSeed = "floyd-steinberg"

	DBSMoveFlip   DBSMoveMode = "flip"
	DBSMoveSwap   DBSMoveMode = "swap"
	DBSMoveHybrid DBSMoveMode = "hybrid"

	DBSMetricFast       DBSMetric = "fast"
	DBSMetricBalanced   DBSMetric = "balanced"
	DBSMetricPerceptual DBSMetric = "perceptual"

	DBSScanRaster     DBSScanOrder = "raster"
	DBSScanSerpentine DBSScanOrder = "serpentine"
	DBSScanRandom     DBSScanOrder = "random"

	DBSRadiusFixed  DBSRadiusPolicy = "fixed"
	DBSRadiusExpand DBSRadiusPolicy = "expand"
)

type DBSOptions struct {
	Seed         DBSSeed
	Passes       int
	Threshold    uint8
	MoveMode     DBSMoveMode
	Neighborhood int
	Metric       DBSMetric
	ScanOrder    DBSScanOrder
	RadiusPolicy DBSRadiusPolicy
	MaxNoImprove int
	Restarts     int
	RandomSeed   uint64
	Report       *DBSReport
}

type DBSReport struct {
	PassesRun        int
	AcceptedMoves    int
	FlipMoves        int
	SwapMoves        int
	RestartsUsed     int
	LastImprovedPass int
}

type dbsMetricSpec struct {
	Radius       int
	Diameter     int
	Weights      []float32
	EdgeAware    bool
	EdgeStrength float32
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
	if o.Metric == "" {
		o.Metric = DBSMetricBalanced
	}
	if o.ScanOrder == "" {
		o.ScanOrder = DBSScanRaster
	}
	if o.RadiusPolicy == "" {
		o.RadiusPolicy = DBSRadiusFixed
	}
	if o.MaxNoImprove <= 0 {
		o.MaxNoImprove = 1
	}
	if o.RandomSeed == 0 {
		o.RandomSeed = 1
	}
	return o
}

func DirectBinarySearch(img *core.Image, opts DBSOptions) error {
	if err := img.Validate(); err != nil {
		return err
	}
	opts = opts.withDefaults()
	metric := dbsMetricSpecFor(opts.Metric)
	target := grayscalePlane(img)
	targetFiltered := filteredGrayPlane(target, img.Width, img.Height, metric)
	errorWeights := dbsErrorWeights(target, img.Width, img.Height, metric)
	binary, err := dbsSeedPlane(target, img.Width, img.Height, opts)
	if err != nil {
		return err
	}
	bestBinary := append([]uint8(nil), binary...)
	bestFiltered := filteredGrayPlane(bestBinary, img.Width, img.Height, metric)
	bestScore := dbsFullScore(bestFiltered, targetFiltered, errorWeights)
	totalReport := DBSReport{}
	for restart := 0; restart <= opts.Restarts; restart++ {
		workingBinary := append([]uint8(nil), binary...)
		workingFiltered := filteredGrayPlane(workingBinary, img.Width, img.Height, metric)
		report := dbsOptimize(workingBinary, workingFiltered, targetFiltered, errorWeights, img.Width, img.Height, opts, metric, restart)
		score := dbsFullScore(workingFiltered, targetFiltered, errorWeights)
		if restart == 0 || score < bestScore {
			bestScore = score
			copy(bestBinary, workingBinary)
			copy(bestFiltered, workingFiltered)
		}
		totalReport.PassesRun += report.PassesRun
		totalReport.AcceptedMoves += report.AcceptedMoves
		totalReport.FlipMoves += report.FlipMoves
		totalReport.SwapMoves += report.SwapMoves
		if report.LastImprovedPass > 0 {
			totalReport.LastImprovedPass = report.LastImprovedPass
		}
		totalReport.RestartsUsed = restart
	}
	if opts.Report != nil {
		*opts.Report = totalReport
	}
	writeBinaryToImage(img, bestBinary)
	return nil
}

func dbsOptimize(binary []uint8, filtered, targetFiltered, errorWeights []float32, width, height int, opts DBSOptions, metric dbsMetricSpec, restart int) DBSReport {
	report := DBSReport{RestartsUsed: restart}
	noImprovePasses := 0
	for pass := 0; pass < opts.Passes; pass++ {
		improved := false
		report.PassesRun++
		radius := dbsEffectiveRadius(opts, pass)
		dbsIteratePixels(width, height, opts, restart, pass, func(x, y int) {
			idx := y*width + x
			original := binary[idx]
			moveOpts := opts
			moveOpts.Neighborhood = radius
			bestMode, bestSwapX, bestSwapY, bestBefore, bestAfter := dbsBestMove(binary, filtered, targetFiltered, errorWeights, width, height, x, y, moveOpts, metric)
			if bestAfter+1e-9 < bestBefore {
				improved = true
				report.AcceptedMoves++
				report.LastImprovedPass = report.PassesRun
				switch bestMode {
				case DBSMoveFlip:
					report.FlipMoves++
					delta := dbsPixelDelta(original)
					binary[idx] = original ^ 0xff
					applyFilteredDelta(filtered, width, height, x, y, delta, metric)
				case DBSMoveSwap:
					report.SwapMoves++
					dbsApplySwap(binary, filtered, width, height, x, y, bestSwapX, bestSwapY, metric)
				}
			}
		})
		if !improved {
			noImprovePasses++
			if noImprovePasses >= opts.MaxNoImprove {
				break
			}
			continue
		}
		noImprovePasses = 0
	}
	return report
}

func dbsIteratePixels(width, height int, opts DBSOptions, restart, pass int, fn func(x, y int)) {
	switch opts.ScanOrder {
	case DBSScanSerpentine:
		for y := 0; y < height; y++ {
			if y%2 == 0 {
				for x := 0; x < width; x++ {
					fn(x, y)
				}
			} else {
				for x := width - 1; x >= 0; x-- {
					fn(x, y)
				}
			}
		}
	case DBSScanRandom:
		order := make([]int, width*height)
		for i := range order {
			order[i] = i
		}
		rng := rand.New(rand.NewSource(int64(opts.RandomSeed) + int64(restart*131+pass*17)))
		rng.Shuffle(len(order), func(i, j int) {
			order[i], order[j] = order[j], order[i]
		})
		for _, idx := range order {
			fn(idx%width, idx/width)
		}
	default:
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				fn(x, y)
			}
		}
	}
}

func dbsEffectiveRadius(opts DBSOptions, pass int) int {
	if opts.RadiusPolicy != DBSRadiusExpand {
		return opts.Neighborhood
	}
	return opts.Neighborhood + pass
}

func dbsFullScore(filtered, targetFiltered, errorWeights []float32) float64 {
	var score float64
	for i := range filtered {
		diff := float64(filtered[i] - targetFiltered[i])
		score += diff * diff * float64(errorWeights[i])
	}
	return score
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

func filteredGrayPlane(gray []uint8, width, height int, metric dbsMetricSpec) []float32 {
	out := make([]float32, len(gray))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out[y*width+x] = filteredBinaryValue(gray, width, height, x, y, metric)
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

func dbsBestMove(binary []uint8, filtered, targetFiltered, errorWeights []float32, width, height, x, y int, opts DBSOptions, metric dbsMetricSpec) (DBSMoveMode, int, int, float64, float64) {
	bestMode := DBSMoveFlip
	bestSwapX, bestSwapY := -1, -1
	bestBefore, bestAfter := 0.0, 0.0
	hasCandidate := false
	if opts.MoveMode == DBSMoveFlip || opts.MoveMode == DBSMoveHybrid {
		delta := dbsPixelDelta(binary[y*width+x])
		before, after := localFilteredErrorDelta(filtered, targetFiltered, errorWeights, width, height, x, y, delta, metric)
		bestBefore, bestAfter = before, after
		hasCandidate = true
	}
	if opts.MoveMode == DBSMoveSwap || opts.MoveMode == DBSMoveHybrid {
		swapBefore, swapAfter, sx, sy, ok := bestSwapCandidate(binary, filtered, targetFiltered, errorWeights, width, height, x, y, opts.Neighborhood, metric)
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

func bestSwapCandidate(binary []uint8, filtered, targetFiltered, errorWeights []float32, width, height, x, y, radius int, metric dbsMetricSpec) (float64, float64, int, int, bool) {
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
			before, after := localSwapErrorDelta(filtered, targetFiltered, errorWeights, width, height, x, y, nx, ny, dbsPixelDelta(origin), dbsPixelDelta(other), metric)
			if !found || after < bestAfter {
				bestBefore, bestAfter = before, after
				bestX, bestY = nx, ny
				found = true
			}
		}
	}
	return bestBefore, bestAfter, bestX, bestY, found
}

func localFilteredErrorDelta(filtered, targetFiltered, errorWeights []float32, width, height, x, y int, delta float32, metric dbsMetricSpec) (float64, float64) {
	var before, after float64
	for yy := maxDBSInt(0, y-metric.Radius); yy <= minDBSInt(height-1, y+metric.Radius); yy++ {
		for xx := maxDBSInt(0, x-metric.Radius); xx <= minDBSInt(width-1, x+metric.Radius); xx++ {
			idx := yy*width + xx
			current := filtered[idx]
			weight := float64(errorWeights[idx])
			diffBefore := float64(current - targetFiltered[idx])
			diffAfter := float64(current + delta*metric.contribution(xx-x, yy-y) - targetFiltered[idx])
			before += diffBefore * diffBefore * weight
			after += diffAfter * diffAfter * weight
		}
	}
	return before, after
}

func filteredBinaryValue(binary []uint8, width, height, x, y int, metric dbsMetricSpec) float32 {
	var sum float32
	for ky := -metric.Radius; ky <= metric.Radius; ky++ {
		yy := clampDBSInt(y+ky, 0, height-1)
		for kx := -metric.Radius; kx <= metric.Radius; kx++ {
			xx := clampDBSInt(x+kx, 0, width-1)
			weight := metric.contribution(kx, ky)
			sum += mathx.ByteToUnit(binary[yy*width+xx]) * weight
		}
	}
	return sum
}

func applyFilteredDelta(filtered []float32, width, height, x, y int, delta float32, metric dbsMetricSpec) {
	for yy := maxDBSInt(0, y-metric.Radius); yy <= minDBSInt(height-1, y+metric.Radius); yy++ {
		for xx := maxDBSInt(0, x-metric.Radius); xx <= minDBSInt(width-1, x+metric.Radius); xx++ {
			filtered[yy*width+xx] += delta * metric.contribution(xx-x, yy-y)
		}
	}
}

func localSwapErrorDelta(filtered, targetFiltered, errorWeights []float32, width, height, x1, y1, x2, y2 int, delta1, delta2 float32, metric dbsMetricSpec) (float64, float64) {
	minX := maxDBSInt(0, minDBSInt(x1, x2)-metric.Radius)
	maxX := minDBSInt(width-1, maxDBSInt(x1, x2)+metric.Radius)
	minY := maxDBSInt(0, minDBSInt(y1, y2)-metric.Radius)
	maxY := minDBSInt(height-1, maxDBSInt(y1, y2)+metric.Radius)
	var before, after float64
	for yy := minY; yy <= maxY; yy++ {
		for xx := minX; xx <= maxX; xx++ {
			idx := yy*width + xx
			current := filtered[idx]
			updated := current
			if absDBSInt(xx-x1) <= metric.Radius && absDBSInt(yy-y1) <= metric.Radius {
				updated += delta1 * metric.contribution(xx-x1, yy-y1)
			}
			if absDBSInt(xx-x2) <= metric.Radius && absDBSInt(yy-y2) <= metric.Radius {
				updated += delta2 * metric.contribution(xx-x2, yy-y2)
			}
			weight := float64(errorWeights[idx])
			diffBefore := float64(current - targetFiltered[idx])
			diffAfter := float64(updated - targetFiltered[idx])
			before += diffBefore * diffBefore * weight
			after += diffAfter * diffAfter * weight
		}
	}
	return before, after
}

func dbsApplySwap(binary []uint8, filtered []float32, width, height, x1, y1, x2, y2 int, metric dbsMetricSpec) {
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
	applyFilteredDelta(filtered, width, height, x1, y1, delta1, metric)
	applyFilteredDelta(filtered, width, height, x2, y2, delta2, metric)
}

func dbsMetricSpecFor(metric DBSMetric) dbsMetricSpec {
	switch metric {
	case DBSMetricFast:
		return newDBSMetricSpec([][]float32{
			{1, 2, 1},
			{2, 4, 2},
			{1, 2, 1},
		}, false, 0)
	case DBSMetricPerceptual:
		return newDBSMetricSpec([][]float32{
			{1, 6, 15, 20, 15, 6, 1},
			{6, 36, 90, 120, 90, 36, 6},
			{15, 90, 225, 300, 225, 90, 15},
			{20, 120, 300, 400, 300, 120, 20},
			{15, 90, 225, 300, 225, 90, 15},
			{6, 36, 90, 120, 90, 36, 6},
			{1, 6, 15, 20, 15, 6, 1},
		}, true, 1.5)
	default:
		return newDBSMetricSpec([][]float32{
			{1, 4, 6, 4, 1},
			{4, 16, 24, 16, 4},
			{6, 24, 36, 24, 6},
			{4, 16, 24, 16, 4},
			{1, 4, 6, 4, 1},
		}, false, 0)
	}
}

func newDBSMetricSpec(kernel [][]float32, edgeAware bool, edgeStrength float32) dbsMetricSpec {
	diameter := len(kernel)
	weights := make([]float32, 0, diameter*diameter)
	var sum float32
	for _, row := range kernel {
		for _, weight := range row {
			sum += weight
		}
	}
	if sum == 0 {
		sum = 1
	}
	for _, row := range kernel {
		for _, weight := range row {
			weights = append(weights, weight/sum)
		}
	}
	return dbsMetricSpec{
		Radius:       diameter / 2,
		Diameter:     diameter,
		Weights:      weights,
		EdgeAware:    edgeAware,
		EdgeStrength: edgeStrength,
	}
}

func (m dbsMetricSpec) contribution(kx, ky int) float32 {
	if kx < -m.Radius || kx > m.Radius || ky < -m.Radius || ky > m.Radius {
		return 0
	}
	return m.Weights[(ky+m.Radius)*m.Diameter+(kx+m.Radius)]
}

func dbsErrorWeights(target []uint8, width, height int, metric dbsMetricSpec) []float32 {
	weights := make([]float32, width*height)
	for i := range weights {
		weights[i] = 1
	}
	if !metric.EdgeAware {
		return weights
	}
	for y := 0; y < height; y++ {
		up := maxDBSInt(0, y-1)
		down := minDBSInt(height-1, y+1)
		for x := 0; x < width; x++ {
			left := maxDBSInt(0, x-1)
			right := minDBSInt(width-1, x+1)
			gx := absDBSInt(int(target[y*width+right]) - int(target[y*width+left]))
			gy := absDBSInt(int(target[down*width+x]) - int(target[up*width+x]))
			gradient := float32(gx+gy) / 510.0
			weights[y*width+x] = 1 + gradient*metric.EdgeStrength
		}
	}
	return weights
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
