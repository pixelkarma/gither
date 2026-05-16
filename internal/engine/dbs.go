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
	DBSSeedCluster16 DBSSeed = "cluster-dot-16x16"

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
	Seed             DBSSeed
	Levels           int
	Palette          core.Palette
	Passes           int
	Threshold        uint8
	MoveMode         DBSMoveMode
	Neighborhood     int
	Metric           DBSMetric
	ScanOrder        DBSScanOrder
	RadiusPolicy     DBSRadiusPolicy
	MaxNoImprove     int
	Restarts         int
	RandomSeed       uint64
	ClusterStrength  float32
	ClusterToneAware bool
	Report           *DBSReport
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
	if o.Levels < 2 {
		o.Levels = 2
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
	return runDBS(img, opts.withDefaults())
}

func ClusteredDBS(img *core.Image, opts DBSOptions) error {
	opts = opts.withDefaults()
	opts.Levels = 2
	if opts.Seed == DBSSeedThreshold {
		opts.Seed = DBSSeedCluster16
	}
	if opts.ClusterStrength <= 0 {
		opts.ClusterStrength = 0.18
	}
	opts.ClusterToneAware = true
	return runDBS(img, opts)
}

func MultiLevelDBS(img *core.Image, opts DBSOptions) error {
	if err := img.Validate(); err != nil {
		return err
	}
	opts = opts.withDefaults()
	target := grayscalePlane(img)
	metric := dbsMetricSpecFor(opts.Metric)
	targetFiltered := filteredGrayPlane(target, img.Width, img.Height, metric)
	errorWeights := dbsErrorWeights(target, img.Width, img.Height, metric)
	levelValues := dbsLevelValues(opts.Levels)
	gray, err := dbsSeedGrayPlane(target, img.Width, img.Height, opts)
	if err != nil {
		return err
	}
	bestGray := append([]uint8(nil), gray...)
	bestFiltered := filteredGrayPlane(bestGray, img.Width, img.Height, metric)
	bestScore := dbsFullScore(bestFiltered, targetFiltered, errorWeights)
	totalReport := DBSReport{}
	for restart := 0; restart <= opts.Restarts; restart++ {
		workingGray := append([]uint8(nil), gray...)
		workingFiltered := filteredGrayPlane(workingGray, img.Width, img.Height, metric)
		report := dbsOptimizeMultiLevel(workingGray, workingFiltered, targetFiltered, errorWeights, img.Width, img.Height, opts, metric, levelValues, restart)
		score := dbsFullScore(workingFiltered, targetFiltered, errorWeights)
		if restart == 0 || score < bestScore {
			bestScore = score
			copy(bestGray, workingGray)
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
	writeGrayToImage(img, bestGray)
	return nil
}

func ColorDBS(img *core.Image, opts DBSOptions) error {
	if err := img.Validate(); err != nil {
		return err
	}
	opts = opts.withDefaults()
	if err := opts.Palette.Validate(); err != nil {
		return err
	}
	targetR, targetG, targetB := colorPlanes(img)
	metric := dbsMetricSpecFor(opts.Metric)
	targetFR := filteredGrayPlane(targetR, img.Width, img.Height, metric)
	targetFG := filteredGrayPlane(targetG, img.Width, img.Height, metric)
	targetFB := filteredGrayPlane(targetB, img.Width, img.Height, metric)
	errorWeights := dbsColorErrorWeights(targetR, targetG, targetB, img.Width, img.Height, metric)
	indexes := dbsSeedPaletteIndexes(targetR, targetG, targetB, opts.Palette)
	bestIndexes := append([]uint8(nil), indexes...)
	bestFR, bestFG, bestFB := filteredPalettePlanes(bestIndexes, opts.Palette, img.Width, img.Height, metric)
	bestScore := dbsColorScore(bestFR, bestFG, bestFB, targetFR, targetFG, targetFB, errorWeights)
	totalReport := DBSReport{}
	for restart := 0; restart <= opts.Restarts; restart++ {
		working := append([]uint8(nil), indexes...)
		fr, fg, fb := filteredPalettePlanes(working, opts.Palette, img.Width, img.Height, metric)
		report := dbsOptimizeColor(working, fr, fg, fb, targetR, targetG, targetB, targetFR, targetFG, targetFB, errorWeights, img.Width, img.Height, opts, metric, restart)
		score := dbsColorScore(fr, fg, fb, targetFR, targetFG, targetFB, errorWeights)
		if restart == 0 || score < bestScore {
			bestScore = score
			copy(bestIndexes, working)
			copy(bestFR, fr)
			copy(bestFG, fg)
			copy(bestFB, fb)
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
	writePaletteToImage(img, bestIndexes, opts.Palette)
	return nil
}

func runDBS(img *core.Image, opts DBSOptions) error {
	if err := img.Validate(); err != nil {
		return err
	}
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
	bestScore := dbsObjectiveScore(bestBinary, bestFiltered, target, targetFiltered, errorWeights, img.Width, img.Height, opts)
	totalReport := DBSReport{}
	for restart := 0; restart <= opts.Restarts; restart++ {
		workingBinary := append([]uint8(nil), binary...)
		workingFiltered := filteredGrayPlane(workingBinary, img.Width, img.Height, metric)
		report := dbsOptimize(workingBinary, workingFiltered, target, targetFiltered, errorWeights, img.Width, img.Height, opts, metric, restart)
		score := dbsObjectiveScore(workingBinary, workingFiltered, target, targetFiltered, errorWeights, img.Width, img.Height, opts)
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

func dbsOptimize(binary []uint8, filtered []float32, target []uint8, targetFiltered, errorWeights []float32, width, height int, opts DBSOptions, metric dbsMetricSpec, restart int) DBSReport {
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
			bestMode, bestSwapX, bestSwapY, bestBefore, bestAfter := dbsBestMove(binary, filtered, target, targetFiltered, errorWeights, width, height, x, y, moveOpts, metric)
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

func dbsObjectiveScore(binary []uint8, filtered []float32, target []uint8, targetFiltered, errorWeights []float32, width, height int, opts DBSOptions) float64 {
	score := dbsFullScore(filtered, targetFiltered, errorWeights)
	if opts.ClusterStrength > 0 {
		score += fullClusterPenalty(binary, target, width, height, opts)
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
	case DBSSeedCluster16:
		ordered := OrderedMap{Values: maps.GenerateClusterDot16x16(), Width: 16, Height: 16, Strength: core.DefaultOrderedStrength}
		if err := ApplyOrdered(img, ordered, baseOpts); err != nil {
			return nil, err
		}
	default:
		if err := Threshold(img, baseOpts); err != nil {
			return nil, err
		}
	}
	return img.Pix, nil
}

func dbsSeedGrayPlane(target []uint8, width, height int, opts DBSOptions) ([]uint8, error) {
	seed := append([]uint8(nil), target...)
	img, err := core.NewPackedImage(seed, width, height, core.Gray8)
	if err != nil {
		return nil, err
	}
	baseOpts := core.Options{Quantizer: core.GrayLevels(opts.Levels), Threshold: opts.Threshold}
	switch opts.Seed {
	case DBSSeedBayer:
		ordered := OrderedMap{Values: maps.Bayer8x8, Width: 8, Height: 8, Strength: core.DefaultOrderedStrength}
		if err := ApplyOrdered(img, ordered, baseOpts); err != nil {
			return nil, err
		}
	case DBSSeedFloyd:
		if err := ApplyDiffusion(img, baseOpts, kernels.FloydSteinberg); err != nil {
			return nil, err
		}
	case DBSSeedCluster16:
		ordered := OrderedMap{Values: maps.GenerateClusterDot16x16(), Width: 16, Height: 16, Strength: core.DefaultOrderedStrength}
		if err := ApplyOrdered(img, ordered, baseOpts); err != nil {
			return nil, err
		}
	default:
		for i := range img.Pix {
			img.Pix[i] = mathx.QuantizeByte(img.Pix[i], opts.Levels)
		}
	}
	return img.Pix, nil
}

func dbsLevelValues(levels int) []uint8 {
	out := make([]uint8, levels)
	for i := 0; i < levels; i++ {
		if levels == 1 {
			out[i] = 0
			continue
		}
		out[i] = uint8((i*255 + (levels-1)/2) / (levels - 1))
	}
	return out
}

func dbsLevelIndex(values []uint8, v uint8) int {
	best := 0
	bestDist := absDBSInt(int(v) - int(values[0]))
	for i := 1; i < len(values); i++ {
		dist := absDBSInt(int(v) - int(values[i]))
		if dist < bestDist {
			best = i
			bestDist = dist
		}
	}
	return best
}

func dbsSeedPaletteIndexes(targetR, targetG, targetB []uint8, palette core.Palette) []uint8 {
	indexes := make([]uint8, len(targetR))
	for i := range indexes {
		indexes[i] = uint8(nearestPaletteIndex(palette, targetR[i], targetG[i], targetB[i]))
	}
	return indexes
}

func nearestPaletteIndex(palette core.Palette, r, g, b uint8) int {
	best := 0
	bestDist := mathx.RGBDistanceSq(r, g, b, palette[0].R, palette[0].G, palette[0].B)
	for i := 1; i < len(palette); i++ {
		c := palette[i]
		dist := mathx.RGBDistanceSq(r, g, b, c.R, c.G, c.B)
		if dist < bestDist {
			best = i
			bestDist = dist
		}
	}
	return best
}

func nearestPaletteCandidates(palette core.Palette, r, g, b uint8, limit int) []int {
	type candidate struct {
		idx  int
		dist uint32
	}
	best := make([]candidate, 0, limit)
	for i, c := range palette {
		dist := mathx.RGBDistanceSq(r, g, b, c.R, c.G, c.B)
		insertAt := len(best)
		for j := 0; j < len(best); j++ {
			if dist < best[j].dist {
				insertAt = j
				break
			}
		}
		if insertAt < limit {
			best = append(best, candidate{})
			copy(best[insertAt+1:], best[insertAt:])
			best[insertAt] = candidate{idx: i, dist: dist}
			if len(best) > limit {
				best = best[:limit]
			}
		} else if len(best) < limit {
			best = append(best, candidate{idx: i, dist: dist})
		}
	}
	out := make([]int, len(best))
	for i := range best {
		out[i] = best[i].idx
	}
	return out
}

func filteredPalettePlanes(indexes []uint8, palette core.Palette, width, height int, metric dbsMetricSpec) ([]float32, []float32, []float32) {
	r := make([]uint8, len(indexes))
	g := make([]uint8, len(indexes))
	b := make([]uint8, len(indexes))
	for i, idx := range indexes {
		c := palette[int(idx)]
		r[i], g[i], b[i] = c.R, c.G, c.B
	}
	return filteredGrayPlane(r, width, height, metric), filteredGrayPlane(g, width, height, metric), filteredGrayPlane(b, width, height, metric)
}

func colorPlanes(img *core.Image) ([]uint8, []uint8, []uint8) {
	r := make([]uint8, img.Width*img.Height)
	g := make([]uint8, img.Width*img.Height)
	b := make([]uint8, img.Width*img.Height)
	channels := img.ChannelCount()
	for y := 0; y < img.Height; y++ {
		row := img.Row(y)
		for x := 0; x < img.Width; x++ {
			offset := x * channels
			switch img.Format {
			case core.Gray8:
				v := row[offset]
				r[y*img.Width+x], g[y*img.Width+x], b[y*img.Width+x] = v, v, v
			default:
				r[y*img.Width+x] = row[offset]
				g[y*img.Width+x] = row[offset+1]
				b[y*img.Width+x] = row[offset+2]
			}
		}
	}
	return r, g, b
}

func dbsColorErrorWeights(targetR, targetG, targetB []uint8, width, height int, metric dbsMetricSpec) []float32 {
	gray := make([]uint8, len(targetR))
	for i := range gray {
		gray[i] = mathx.LumaByte(targetR[i], targetG[i], targetB[i])
	}
	return dbsErrorWeights(gray, width, height, metric)
}

func dbsColorScore(fr, fg, fb, tr, tg, tb, errorWeights []float32) float64 {
	var score float64
	for i := range fr {
		dr := float64(fr[i] - tr[i])
		dg := float64(fg[i] - tg[i])
		db := float64(fb[i] - tb[i])
		weight := float64(errorWeights[i])
		score += weight * (0.30*dr*dr + 0.59*dg*dg + 0.11*db*db)
	}
	return score
}

func dbsOptimizeMultiLevel(gray []uint8, filtered, targetFiltered, errorWeights []float32, width, height int, opts DBSOptions, metric dbsMetricSpec, levels []uint8, restart int) DBSReport {
	report := DBSReport{RestartsUsed: restart}
	noImprovePasses := 0
	for pass := 0; pass < opts.Passes; pass++ {
		improved := false
		report.PassesRun++
		radius := dbsEffectiveRadius(opts, pass)
		dbsIteratePixels(width, height, opts, restart, pass, func(x, y int) {
			bestKind, nx, ny, before, after, newA, newB := dbsBestMultiLevelMove(gray, filtered, targetFiltered, errorWeights, width, height, x, y, radius, metric, levels)
			if after+1e-9 < before {
				improved = true
				report.AcceptedMoves++
				report.LastImprovedPass = report.PassesRun
				idx := y*width + x
				switch bestKind {
				case DBSMoveFlip:
					report.FlipMoves++
					delta := mathx.ByteToUnit(newA) - mathx.ByteToUnit(gray[idx])
					gray[idx] = newA
					applyFilteredDelta(filtered, width, height, x, y, delta, metric)
				case DBSMoveSwap:
					report.SwapMoves++
					applyMultiLevelExchange(gray, filtered, width, height, x, y, nx, ny, newA, newB, metric)
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

func dbsOptimizeColor(indexes []uint8, fr, fg, fb []float32, targetR, targetG, targetB []uint8, targetFR, targetFG, targetFB, errorWeights []float32, width, height int, opts DBSOptions, metric dbsMetricSpec, restart int) DBSReport {
	report := DBSReport{RestartsUsed: restart}
	noImprovePasses := 0
	for pass := 0; pass < opts.Passes; pass++ {
		improved := false
		report.PassesRun++
		radius := dbsEffectiveRadius(opts, pass)
		dbsIteratePixels(width, height, opts, restart, pass, func(x, y int) {
			move, nx, ny, before, after, newA, newB := dbsBestColorMove(indexes, fr, fg, fb, targetR, targetG, targetB, targetFR, targetFG, targetFB, errorWeights, width, height, x, y, radius, opts.Palette, metric)
			if after+1e-9 < before {
				improved = true
				report.AcceptedMoves++
				report.LastImprovedPass = report.PassesRun
				idx := y*width + x
				switch move {
				case DBSMoveFlip:
					report.FlipMoves++
					old := opts.Palette[int(indexes[idx])]
					next := opts.Palette[int(newA)]
					indexes[idx] = newA
					applyFilteredDelta(fr, width, height, x, y, mathx.ByteToUnit(next.R)-mathx.ByteToUnit(old.R), metric)
					applyFilteredDelta(fg, width, height, x, y, mathx.ByteToUnit(next.G)-mathx.ByteToUnit(old.G), metric)
					applyFilteredDelta(fb, width, height, x, y, mathx.ByteToUnit(next.B)-mathx.ByteToUnit(old.B), metric)
				case DBSMoveSwap:
					report.SwapMoves++
					applyColorSwap(indexes, fr, fg, fb, width, height, x, y, nx, ny, newA, newB, opts.Palette, metric)
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

func dbsBestColorMove(indexes []uint8, fr, fg, fb []float32, targetR, targetG, targetB []uint8, targetFR, targetFG, targetFB, errorWeights []float32, width, height, x, y, radius int, palette core.Palette, metric dbsMetricSpec) (DBSMoveMode, int, int, float64, float64, uint8, uint8) {
	idx := y*width + x
	current := int(indexes[idx])
	bestMode := DBSMoveFlip
	bestX, bestY := -1, -1
	bestBefore, bestAfter := 0.0, 0.0
	bestA, bestB := uint8(current), uint8(current)
	hasCandidate := false
	candidates := nearestPaletteCandidates(palette, targetR[idx], targetG[idx], targetB[idx], minDBSInt(4, len(palette)))
	for _, candidate := range candidates {
		if candidate == current {
			continue
		}
		before, after := localFilteredColorErrorDelta(fr, fg, fb, targetFR, targetFG, targetFB, errorWeights, width, height, x, y, palette[current], palette[candidate], metric)
		if !hasCandidate || after < bestAfter {
			bestMode = DBSMoveFlip
			bestBefore, bestAfter = before, after
			bestA, bestB = uint8(candidate), uint8(current)
			hasCandidate = true
		}
	}
	swapBefore, swapAfter, sx, sy, sa, sb, ok := bestColorSwap(indexes, fr, fg, fb, targetFR, targetFG, targetFB, errorWeights, width, height, x, y, radius, palette, metric)
	if ok && (!hasCandidate || swapAfter < bestAfter) {
		bestMode = DBSMoveSwap
		bestX, bestY = sx, sy
		bestBefore, bestAfter = swapBefore, swapAfter
		bestA, bestB = sa, sb
		hasCandidate = true
	}
	if !hasCandidate {
		return DBSMoveFlip, -1, -1, 0, 0, uint8(current), uint8(current)
	}
	return bestMode, bestX, bestY, bestBefore, bestAfter, bestA, bestB
}

func bestColorSwap(indexes []uint8, fr, fg, fb, targetFR, targetFG, targetFB, errorWeights []float32, width, height, x, y, radius int, palette core.Palette, metric dbsMetricSpec) (float64, float64, int, int, uint8, uint8, bool) {
	idx := y*width + x
	current := indexes[idx]
	bestBefore, bestAfter := 0.0, 0.0
	bestX, bestY := -1, -1
	bestA, bestB := current, current
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
			nidx := ny*width + nx
			other := indexes[nidx]
			if other == current {
				continue
			}
			before, after := localFilteredColorSwapErrorDelta(fr, fg, fb, targetFR, targetFG, targetFB, errorWeights, width, height, x, y, nx, ny, palette[int(current)], palette[int(other)], palette[int(other)], palette[int(current)], metric)
			if !found || after < bestAfter {
				bestBefore, bestAfter = before, after
				bestX, bestY = nx, ny
				bestA, bestB = other, current
				found = true
			}
		}
	}
	return bestBefore, bestAfter, bestX, bestY, bestA, bestB, found
}

func dbsBestMultiLevelMove(gray []uint8, filtered, targetFiltered, errorWeights []float32, width, height, x, y, radius int, metric dbsMetricSpec, levels []uint8) (DBSMoveMode, int, int, float64, float64, uint8, uint8) {
	idx := y*width + x
	current := gray[idx]
	levelIdx := dbsLevelIndex(levels, current)
	bestMode := DBSMoveFlip
	bestX, bestY := -1, -1
	bestBefore, bestAfter := 0.0, 0.0
	bestA, bestB := current, current
	hasCandidate := false
	tryLevel := func(candidate uint8) {
		if candidate == current {
			return
		}
		delta := mathx.ByteToUnit(candidate) - mathx.ByteToUnit(current)
		before, after := localFilteredErrorDelta(filtered, targetFiltered, errorWeights, width, height, x, y, delta, metric)
		if !hasCandidate || after < bestAfter {
			bestMode = DBSMoveFlip
			bestBefore, bestAfter = before, after
			bestA, bestB = candidate, current
			hasCandidate = true
		}
	}
	if levelIdx > 0 {
		tryLevel(levels[levelIdx-1])
	}
	if levelIdx+1 < len(levels) {
		tryLevel(levels[levelIdx+1])
	}
	swapBefore, swapAfter, sx, sy, sa, sb, ok := bestMultiLevelExchange(gray, filtered, targetFiltered, errorWeights, width, height, x, y, radius, metric, levels)
	if ok && (!hasCandidate || swapAfter < bestAfter) {
		bestMode = DBSMoveSwap
		bestX, bestY = sx, sy
		bestBefore, bestAfter = swapBefore, swapAfter
		bestA, bestB = sa, sb
		hasCandidate = true
	}
	if !hasCandidate {
		return DBSMoveFlip, -1, -1, 0, 0, current, current
	}
	return bestMode, bestX, bestY, bestBefore, bestAfter, bestA, bestB
}

func bestMultiLevelExchange(gray []uint8, filtered, targetFiltered, errorWeights []float32, width, height, x, y, radius int, metric dbsMetricSpec, levels []uint8) (float64, float64, int, int, uint8, uint8, bool) {
	idx := y*width + x
	current := gray[idx]
	levelIdx := dbsLevelIndex(levels, current)
	bestBefore, bestAfter := 0.0, 0.0
	bestX, bestY := -1, -1
	bestA, bestB := current, current
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
			nidx := ny*width + nx
			other := gray[nidx]
			otherIdx := dbsLevelIndex(levels, other)
			try := func(newIdxA, newIdxB int) {
				if newIdxA < 0 || newIdxA >= len(levels) || newIdxB < 0 || newIdxB >= len(levels) {
					return
				}
				newA, newB := levels[newIdxA], levels[newIdxB]
				if newA == current && newB == other {
					return
				}
				deltaA := mathx.ByteToUnit(newA) - mathx.ByteToUnit(current)
				deltaB := mathx.ByteToUnit(newB) - mathx.ByteToUnit(other)
				before, after := localMultiSwapErrorDelta(filtered, targetFiltered, errorWeights, width, height, x, y, nx, ny, deltaA, deltaB, metric)
				if !found || after < bestAfter {
					bestBefore, bestAfter = before, after
					bestX, bestY = nx, ny
					bestA, bestB = newA, newB
					found = true
				}
			}
			if levelIdx+1 < len(levels) && otherIdx > 0 {
				try(levelIdx+1, otherIdx-1)
			}
			if levelIdx > 0 && otherIdx+1 < len(levels) {
				try(levelIdx-1, otherIdx+1)
			}
		}
	}
	return bestBefore, bestAfter, bestX, bestY, bestA, bestB, found
}

func localMultiSwapErrorDelta(filtered, targetFiltered, errorWeights []float32, width, height, x1, y1, x2, y2 int, delta1, delta2 float32, metric dbsMetricSpec) (float64, float64) {
	return localSwapErrorDelta(filtered, targetFiltered, errorWeights, width, height, x1, y1, x2, y2, delta1, delta2, metric)
}

func localFilteredColorErrorDelta(fr, fg, fb, targetFR, targetFG, targetFB, errorWeights []float32, width, height, x, y int, oldColor, newColor core.Color, metric dbsMetricSpec) (float64, float64) {
	return localFilteredColorSwapErrorDelta(
		fr, fg, fb, targetFR, targetFG, targetFB, errorWeights, width, height,
		x, y, x, y,
		oldColor, oldColor,
		newColor, oldColor,
		metric,
	)
}

func localFilteredColorSwapErrorDelta(fr, fg, fb, targetFR, targetFG, targetFB, errorWeights []float32, width, height, x1, y1, x2, y2 int, oldA, oldB, newA, newB core.Color, metric dbsMetricSpec) (float64, float64) {
	minX := maxDBSInt(0, minDBSInt(x1, x2)-metric.Radius)
	maxX := minDBSInt(width-1, maxDBSInt(x1, x2)+metric.Radius)
	minY := maxDBSInt(0, minDBSInt(y1, y2)-metric.Radius)
	maxY := minDBSInt(height-1, maxDBSInt(y1, y2)+metric.Radius)
	deltaR1 := mathx.ByteToUnit(newA.R) - mathx.ByteToUnit(oldA.R)
	deltaG1 := mathx.ByteToUnit(newA.G) - mathx.ByteToUnit(oldA.G)
	deltaB1 := mathx.ByteToUnit(newA.B) - mathx.ByteToUnit(oldA.B)
	deltaR2 := mathx.ByteToUnit(newB.R) - mathx.ByteToUnit(oldB.R)
	deltaG2 := mathx.ByteToUnit(newB.G) - mathx.ByteToUnit(oldB.G)
	deltaB2 := mathx.ByteToUnit(newB.B) - mathx.ByteToUnit(oldB.B)
	var before, after float64
	for yy := minY; yy <= maxY; yy++ {
		for xx := minX; xx <= maxX; xx++ {
			weight := float64(errorWeights[yy*width+xx])
			curR, curG, curB := fr[yy*width+xx], fg[yy*width+xx], fb[yy*width+xx]
			nextR, nextG, nextB := curR, curG, curB
			if absDBSInt(xx-x1) <= metric.Radius && absDBSInt(yy-y1) <= metric.Radius {
				contrib := metric.contribution(xx-x1, yy-y1)
				nextR += deltaR1 * contrib
				nextG += deltaG1 * contrib
				nextB += deltaB1 * contrib
			}
			if (x2 != x1 || y2 != y1) && absDBSInt(xx-x2) <= metric.Radius && absDBSInt(yy-y2) <= metric.Radius {
				contrib := metric.contribution(xx-x2, yy-y2)
				nextR += deltaR2 * contrib
				nextG += deltaG2 * contrib
				nextB += deltaB2 * contrib
			}
			dr0 := float64(curR - targetFR[yy*width+xx])
			dg0 := float64(curG - targetFG[yy*width+xx])
			db0 := float64(curB - targetFB[yy*width+xx])
			dr1 := float64(nextR - targetFR[yy*width+xx])
			dg1 := float64(nextG - targetFG[yy*width+xx])
			db1 := float64(nextB - targetFB[yy*width+xx])
			before += weight * (0.30*dr0*dr0 + 0.59*dg0*dg0 + 0.11*db0*db0)
			after += weight * (0.30*dr1*dr1 + 0.59*dg1*dg1 + 0.11*db1*db1)
		}
	}
	return before, after
}

func applyMultiLevelExchange(gray []uint8, filtered []float32, width, height, x1, y1, x2, y2 int, newA, newB uint8, metric dbsMetricSpec) {
	idx1 := y1*width + x1
	idx2 := y2*width + x2
	delta1 := mathx.ByteToUnit(newA) - mathx.ByteToUnit(gray[idx1])
	delta2 := mathx.ByteToUnit(newB) - mathx.ByteToUnit(gray[idx2])
	gray[idx1], gray[idx2] = newA, newB
	applyFilteredDelta(filtered, width, height, x1, y1, delta1, metric)
	applyFilteredDelta(filtered, width, height, x2, y2, delta2, metric)
}

func applyColorSwap(indexes []uint8, fr, fg, fb []float32, width, height, x1, y1, x2, y2 int, newA, newB uint8, palette core.Palette, metric dbsMetricSpec) {
	idx1 := y1*width + x1
	idx2 := y2*width + x2
	oldA := palette[int(indexes[idx1])]
	oldB := palette[int(indexes[idx2])]
	nextA := palette[int(newA)]
	nextB := palette[int(newB)]
	indexes[idx1], indexes[idx2] = newA, newB
	applyFilteredDelta(fr, width, height, x1, y1, mathx.ByteToUnit(nextA.R)-mathx.ByteToUnit(oldA.R), metric)
	applyFilteredDelta(fg, width, height, x1, y1, mathx.ByteToUnit(nextA.G)-mathx.ByteToUnit(oldA.G), metric)
	applyFilteredDelta(fb, width, height, x1, y1, mathx.ByteToUnit(nextA.B)-mathx.ByteToUnit(oldA.B), metric)
	applyFilteredDelta(fr, width, height, x2, y2, mathx.ByteToUnit(nextB.R)-mathx.ByteToUnit(oldB.R), metric)
	applyFilteredDelta(fg, width, height, x2, y2, mathx.ByteToUnit(nextB.G)-mathx.ByteToUnit(oldB.G), metric)
	applyFilteredDelta(fb, width, height, x2, y2, mathx.ByteToUnit(nextB.B)-mathx.ByteToUnit(oldB.B), metric)
}

func dbsBestMove(binary []uint8, filtered []float32, target []uint8, targetFiltered, errorWeights []float32, width, height, x, y int, opts DBSOptions, metric dbsMetricSpec) (DBSMoveMode, int, int, float64, float64) {
	bestMode := DBSMoveFlip
	bestSwapX, bestSwapY := -1, -1
	bestBefore, bestAfter := 0.0, 0.0
	hasCandidate := false
	if opts.MoveMode == DBSMoveFlip || opts.MoveMode == DBSMoveHybrid {
		delta := dbsPixelDelta(binary[y*width+x])
		before, after := localFilteredErrorDelta(filtered, targetFiltered, errorWeights, width, height, x, y, delta, metric)
		clusterBefore, clusterAfter := localClusterFlipDelta(binary, target, width, height, x, y, opts)
		before += clusterBefore
		after += clusterAfter
		bestBefore, bestAfter = before, after
		hasCandidate = true
	}
	if opts.MoveMode == DBSMoveSwap || opts.MoveMode == DBSMoveHybrid {
		swapBefore, swapAfter, sx, sy, ok := bestSwapCandidate(binary, filtered, target, targetFiltered, errorWeights, width, height, x, y, opts.Neighborhood, opts, metric)
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

func bestSwapCandidate(binary []uint8, filtered []float32, target []uint8, targetFiltered, errorWeights []float32, width, height, x, y, radius int, opts DBSOptions, metric dbsMetricSpec) (float64, float64, int, int, bool) {
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
			clusterBefore, clusterAfter := localClusterSwapDelta(binary, target, width, height, x, y, nx, ny, opts)
			before += clusterBefore
			after += clusterAfter
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

type dbsPixelChange struct {
	x      int
	y      int
	newVal uint8
}

func localClusterFlipDelta(binary, target []uint8, width, height, x, y int, opts DBSOptions) (float64, float64) {
	if opts.ClusterStrength <= 0 {
		return 0, 0
	}
	idx := y*width + x
	return localClusterPenalty(binary, target, width, height, opts, []dbsPixelChange{{
		x:      x,
		y:      y,
		newVal: binary[idx] ^ 0xff,
	}})
}

func localClusterSwapDelta(binary, target []uint8, width, height, x1, y1, x2, y2 int, opts DBSOptions) (float64, float64) {
	if opts.ClusterStrength <= 0 {
		return 0, 0
	}
	idx1 := y1*width + x1
	idx2 := y2*width + x2
	if binary[idx1] == binary[idx2] {
		return 0, 0
	}
	return localClusterPenalty(binary, target, width, height, opts, []dbsPixelChange{
		{x: x1, y: y1, newVal: binary[idx2]},
		{x: x2, y: y2, newVal: binary[idx1]},
	})
}

func localClusterPenalty(binary, target []uint8, width, height int, opts DBSOptions, changes []dbsPixelChange) (float64, float64) {
	overrides := make(map[int]uint8, len(changes))
	seen := make(map[uint64]struct{}, len(changes)*8)
	for _, change := range changes {
		overrides[change.y*width+change.x] = change.newVal
	}
	var before, after float64
	for _, change := range changes {
		idx := change.y*width + change.x
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx := change.x + dx
				ny := change.y + dy
				if nx < 0 || ny < 0 || nx >= width || ny >= height {
					continue
				}
				nidx := ny*width + nx
				key := dbsEdgeKey(idx, nidx)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				diagonalScale := 1.0
				if dx != 0 && dy != 0 {
					diagonalScale = 0.70710678118
				}
				weight := clusterEdgeWeight(target, idx, nidx, opts) * diagonalScale
				oldA := binary[idx]
				oldB := binary[nidx]
				newA := oldA
				if v, ok := overrides[idx]; ok {
					newA = v
				}
				newB := oldB
				if v, ok := overrides[nidx]; ok {
					newB = v
				}
				if oldA != oldB {
					before += weight
				}
				if newA != newB {
					after += weight
				}
			}
		}
	}
	return before, after
}

func fullClusterPenalty(binary, target []uint8, width, height int, opts DBSOptions) float64 {
	if opts.ClusterStrength <= 0 {
		return 0
	}
	var penalty float64
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			if x+1 < width {
				ridx := y*width + x + 1
				if binary[idx] != binary[ridx] {
					penalty += clusterEdgeWeight(target, idx, ridx, opts)
				}
			}
			if y+1 < height {
				didx := (y+1)*width + x
				if binary[idx] != binary[didx] {
					penalty += clusterEdgeWeight(target, idx, didx, opts)
				}
			}
			if x+1 < width && y+1 < height {
				dridx := (y+1)*width + x + 1
				if binary[idx] != binary[dridx] {
					penalty += clusterEdgeWeight(target, idx, dridx, opts) * 0.70710678118
				}
			}
		}
	}
	return penalty
}

func clusterEdgeWeight(target []uint8, idxA, idxB int, opts DBSOptions) float64 {
	weight := float64(opts.ClusterStrength)
	if !opts.ClusterToneAware {
		return weight
	}
	mid := 0.5 * (midtoneWeight(target[idxA]) + midtoneWeight(target[idxB]))
	return weight * (0.35 + 0.65*mid)
}

func midtoneWeight(v uint8) float64 {
	value := float64(v) / 255.0
	distance := value - 0.5
	weight := 1.0 - 2.0*absFloat64(distance)
	if weight < 0 {
		return 0
	}
	return weight
}

func dbsEdgeKey(a, b int) uint64 {
	if a > b {
		a, b = b, a
	}
	return (uint64(a) << 32) | uint64(uint32(b))
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

func writeGrayToImage(img *core.Image, gray []uint8) {
	channels := img.ChannelCount()
	for y := 0; y < img.Height; y++ {
		row := img.Row(y)
		for x := 0; x < img.Width; x++ {
			v := gray[y*img.Width+x]
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

func writePaletteToImage(img *core.Image, indexes []uint8, palette core.Palette) {
	channels := img.ChannelCount()
	for y := 0; y < img.Height; y++ {
		row := img.Row(y)
		for x := 0; x < img.Width; x++ {
			c := palette[int(indexes[y*img.Width+x])]
			offset := x * channels
			switch img.Format {
			case core.Gray8:
				row[offset] = mathx.LumaByte(c.R, c.G, c.B)
			case core.RGB8:
				row[offset], row[offset+1], row[offset+2] = c.R, c.G, c.B
			case core.RGBA8:
				alpha := row[offset+3]
				row[offset], row[offset+1], row[offset+2], row[offset+3] = c.R, c.G, c.B, alpha
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

func absFloat64(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
