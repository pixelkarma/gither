package engine

import (
	"math"
	"runtime"
	"sync"

	"github.com/pixelkarma/gither/internal/core"
	"github.com/pixelkarma/gither/internal/maps"
	"github.com/pixelkarma/gither/internal/mathx"
)

type mixingPlan struct {
	first  int
	second int
	ratio  uint8
}

func Yliluoma1(img *core.Image, opts core.Options) error {
	return applyYliluoma(img, opts, 1)
}

func Yliluoma2(img *core.Image, opts core.Options) error {
	return applyYliluoma(img, opts, 2)
}

func Yliluoma3(img *core.Image, opts core.Options) error {
	return applyYliluoma(img, opts, 3)
}

func applyYliluoma(img *core.Image, opts core.Options, mode int) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if opts.Quantizer.Kind != core.QuantizePalette {
		return core.ErrEmptyPalette
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	palette := opts.Quantizer.Palette
	if mode == 1 {
		return applyYliluoma1Fast(img, palette)
	}
	if mode == 2 {
		return applyYliluoma2Fast(img, palette)
	}
	return applyYliluoma3Fast(img, palette)
}

func applyYliluoma1Fast(img *core.Image, palette core.Palette) error {
	lut := buildYliluoma1LUT(palette)
	channels := img.ChannelCount()
	parallelRows(img.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			row := img.Row(y)
			rowRankBase := (y % 8) * 8
			for x := 0; x < img.Width; x++ {
				offset := x * channels
				r, g, b := adjustedColorForYliluoma(img.Format, row, offset)
				plan := lut[yliluoma1LUTKey(r, g, b)]
				c := palette[plan.first]
				if int(maps.Bayer8x8[rowRankBase+(x%8)]) < int(plan.ratio) {
					c = palette[plan.second]
				}
				writeColor(img.Format, row, offset, c)
			}
		}
	})
	return nil
}

func buildYliluoma1LUT(palette core.Palette) []mixingPlan {
	const binsPerChannel = 16
	lut := make([]mixingPlan, binsPerChannel*binsPerChannel*binsPerChannel)
	candidateLimit := len(palette)
	if candidateLimit > 6 {
		candidateLimit = 6
	}
	parallelRows(len(lut), func(start, end int) {
		for i := start; i < end; i++ {
			r := uint8(((i >> 8) & 0x0f) * 17)
			g := uint8(((i >> 4) & 0x0f) * 17)
			b := uint8((i & 0x0f) * 17)
			candidates := nearestPaletteIndexesFixed(palette, r, g, b, candidateLimit)
			lut[i] = buildMixingPlanFromCandidates(r, g, b, palette, candidates)
		}
	})
	return lut
}

func yliluoma1LUTKey(r, g, b uint8) int {
	return int(r>>4)<<8 | int(g>>4)<<4 | int(b>>4)
}

func applyYliluoma2Fast(img *core.Image, palette core.Palette) error {
	lut := buildYliluomaSequenceLUT(palette, false)
	return applyYliluomaSequenceLUT(img, palette, lut)
}

func applyYliluoma3Fast(img *core.Image, palette core.Palette) error {
	lut := buildYliluomaSequenceLUT(palette, true)
	return applyYliluomaSequenceLUT(img, palette, lut)
}

func applyYliluomaSequenceLUT(img *core.Image, palette core.Palette, lut [][64]uint8) error {
	channels := img.ChannelCount()
	parallelRows(img.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			row := img.Row(y)
			rowRankBase := (y % 8) * 8
			for x := 0; x < img.Width; x++ {
				offset := x * channels
				r, g, b := adjustedColorForYliluoma(img.Format, row, offset)
				seq := lut[yliluoma1LUTKey(r, g, b)]
				writeColor(img.Format, row, offset, palette[seq[maps.Bayer8x8[rowRankBase+(x%8)]]])
			}
		}
	})
	return nil
}

func buildYliluomaSequenceLUT(palette core.Palette, gamma bool) [][64]uint8 {
	const binsPerChannel = 16
	lut := make([][64]uint8, binsPerChannel*binsPerChannel*binsPerChannel)
	candidateLimit := len(palette)
	if candidateLimit > 6 {
		candidateLimit = 6
	}
	var linearPalette [][3]float64
	if gamma {
		linearPalette = buildLinearPalette(palette)
	}
	parallelRows(len(lut), func(start, end int) {
		for i := start; i < end; i++ {
			r := uint8(((i >> 8) & 0x0f) * 17)
			g := uint8(((i >> 4) & 0x0f) * 17)
			b := uint8((i & 0x0f) * 17)
			candidates := nearestPaletteIndexesFixed(palette, r, g, b, candidateLimit)
			if gamma {
				lut[i] = buildMixSequenceFromCandidatesGamma(r, g, b, palette, linearPalette, candidates)
			} else {
				lut[i] = buildMixSequenceFromCandidates(r, g, b, palette, candidates)
			}
		}
	})
	return lut
}

func adjustedColorForYliluoma(format core.Format, row []uint8, offset int) (uint8, uint8, uint8) {
	switch format {
	case core.Gray8:
		v := row[offset]
		return v, v, v
	default:
		return row[offset], row[offset+1], row[offset+2]
	}
}

func writeColor(format core.Format, row []uint8, offset int, c core.Color) {
	switch format {
	case core.Gray8:
		row[offset] = mathx.LumaByte(c.R, c.G, c.B)
	case core.RGB8:
		row[offset], row[offset+1], row[offset+2] = c.R, c.G, c.B
	case core.RGBA8:
		alpha := row[offset+3]
		row[offset], row[offset+1], row[offset+2], row[offset+3] = c.R, c.G, c.B, alpha
	}
}

func buildMixingPlan(r, g, b uint8, palette core.Palette) mixingPlan {
	candidates := nearestPaletteIndexesFixed(palette, r, g, b, minInt(8, len(palette)))
	return buildMixingPlanFromCandidates(r, g, b, palette, candidates)
}

func buildMixingPlanFromCandidates(r, g, b uint8, palette core.Palette, candidates []int) mixingPlan {
	best := mixingPlan{first: candidates[0], second: candidates[0]}
	bestErr := math.MaxFloat64
	for _, a := range candidates {
		for _, bIdx := range candidates {
			for ratio := 0; ratio <= 64; ratio++ {
				err := mixedDistance(r, g, b, palette[a], palette[bIdx], ratio, false)
				if err < bestErr {
					bestErr = err
					best = mixingPlan{first: a, second: bIdx, ratio: uint8(ratio)}
				}
			}
		}
	}
	return best
}

func buildMixSequence(r, g, b uint8, palette core.Palette, gamma bool) [64]uint8 {
	candidates := nearestPaletteIndexesFixed(palette, r, g, b, minInt(8, len(palette)))
	if gamma {
		return buildMixSequenceFromCandidatesGamma(r, g, b, palette, buildLinearPalette(palette), candidates)
	}
	return buildMixSequenceFromCandidates(r, g, b, palette, candidates)
}

func buildMixSequenceFromCandidates(r, g, b uint8, palette core.Palette, candidates []int) [64]uint8 {
	var seq [64]uint8
	if len(candidates) == 1 {
		for i := range seq {
			seq[i] = uint8(candidates[0])
		}
		return seq
	}
	target := [3]float64{float64(r), float64(g), float64(b)}
	var accum [3]float64
	for i := 0; i < len(seq); i++ {
		bestIdx := candidates[0]
		bestErr := math.MaxFloat64
		for _, candidate := range candidates {
			c := palette[candidate]
			next := [3]float64{
				(accum[0] + float64(c.R)) / float64(i+1),
				(accum[1] + float64(c.G)) / float64(i+1),
				(accum[2] + float64(c.B)) / float64(i+1),
			}
			err := diffSq(next, target)
			if err < bestErr {
				bestErr = err
				bestIdx = candidate
			}
		}
		chosen := palette[bestIdx]
		seq[i] = uint8(bestIdx)
		accum[0] += float64(chosen.R)
		accum[1] += float64(chosen.G)
		accum[2] += float64(chosen.B)
	}
	return seq
}

func buildMixSequenceFromCandidatesGamma(r, g, b uint8, palette core.Palette, linearPalette [][3]float64, candidates []int) [64]uint8 {
	var seq [64]uint8
	if len(candidates) == 1 {
		for i := range seq {
			seq[i] = uint8(candidates[0])
		}
		return seq
	}
	target := [3]float64{gammaToLinearFast(r), gammaToLinearFast(g), gammaToLinearFast(b)}
	var accum [3]float64
	for i := 0; i < len(seq); i++ {
		bestIdx := candidates[0]
		bestErr := math.MaxFloat64
		for _, candidate := range candidates {
			c := linearPalette[candidate]
			next := [3]float64{
				(accum[0] + c[0]) / float64(i+1),
				(accum[1] + c[1]) / float64(i+1),
				(accum[2] + c[2]) / float64(i+1),
			}
			err := diffSq(next, target)
			if err < bestErr {
				bestErr = err
				bestIdx = candidate
			}
		}
		chosen := linearPalette[bestIdx]
		seq[i] = uint8(bestIdx)
		accum[0] += chosen[0]
		accum[1] += chosen[1]
		accum[2] += chosen[2]
	}
	return seq
}

func nearestPaletteIndexesFixed(palette core.Palette, r, g, b uint8, limit int) []int {
	if limit > len(palette) {
		limit = len(palette)
	}
	bestDist := make([]uint32, limit)
	bestIdx := make([]int, limit)
	for i := 0; i < limit; i++ {
		bestDist[i] = math.MaxUint32
		bestIdx[i] = -1
	}
	for i, c := range palette {
		dist := mathx.RGBDistanceSq(r, g, b, c.R, c.G, c.B)
		insertAt := -1
		for j := 0; j < limit; j++ {
			if dist < bestDist[j] {
				insertAt = j
				break
			}
		}
		if insertAt < 0 {
			continue
		}
		for j := limit - 1; j > insertAt; j-- {
			bestDist[j] = bestDist[j-1]
			bestIdx[j] = bestIdx[j-1]
		}
		bestDist[insertAt] = dist
		bestIdx[insertAt] = i
	}
	out := make([]int, 0, limit)
	for _, idx := range bestIdx {
		if idx >= 0 {
			out = append(out, idx)
		}
	}
	return out
}

func mixedDistance(r, g, b uint8, a, c core.Color, ratio int, gamma bool) float64 {
	weight := float64(ratio) / 64.0
	if gamma {
		ar, ag, ab := gammaToLinear(a.R), gammaToLinear(a.G), gammaToLinear(a.B)
		cr, cg, cb := gammaToLinear(c.R), gammaToLinear(c.G), gammaToLinear(c.B)
		target := [3]float64{gammaToLinear(r), gammaToLinear(g), gammaToLinear(b)}
		mixed := [3]float64{
			ar*(1-weight) + cr*weight,
			ag*(1-weight) + cg*weight,
			ab*(1-weight) + cb*weight,
		}
		return diffSq(mixed, target)
	}
	mixed := [3]float64{
		float64(a.R)*(1-weight) + float64(c.R)*weight,
		float64(a.G)*(1-weight) + float64(c.G)*weight,
		float64(a.B)*(1-weight) + float64(c.B)*weight,
	}
	return diffSq(mixed, [3]float64{float64(r), float64(g), float64(b)})
}

func diffSq(a, b [3]float64) float64 {
	dr := a[0] - b[0]
	dg := a[1] - b[1]
	db := a[2] - b[2]
	return dr*dr + dg*dg + db*db
}

func gammaToLinear(v uint8) float64 {
	return math.Pow(float64(v)/255.0, 2.2)
}

var gammaLUT = func() [256]float64 {
	var out [256]float64
	for i := 0; i < len(out); i++ {
		out[i] = math.Pow(float64(i)/255.0, 2.2)
	}
	return out
}()

func gammaToLinearFast(v uint8) float64 {
	return gammaLUT[v]
}

func buildLinearPalette(palette core.Palette) [][3]float64 {
	out := make([][3]float64, len(palette))
	for i, c := range palette {
		out[i] = [3]float64{
			gammaToLinearFast(c.R),
			gammaToLinearFast(c.G),
			gammaToLinearFast(c.B),
		}
	}
	return out
}

func parallelRows(total int, fn func(start, end int)) {
	if total <= 0 {
		return
	}
	workers := runtime.GOMAXPROCS(0)
	if workers < 2 || total < workers*32 {
		fn(0, total)
		return
	}
	chunk := (total + workers - 1) / workers
	var wg sync.WaitGroup
	for start := 0; start < total; start += chunk {
		end := start + chunk
		if end > total {
			end = total
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			fn(start, end)
		}(start, end)
	}
	wg.Wait()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
