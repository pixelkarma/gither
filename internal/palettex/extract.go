package palettex

import (
	"sort"

	"gither/internal/core"
)

type bucket struct {
	colors []weightedColor
}

type weightedColor struct {
	color  core.Color
	weight int
}

type Method uint8
type SortMode uint8

const (
	MethodMedianCut Method = iota
	MethodPopularity
)

const (
	SortRGB SortMode = iota
	SortLuma
	SortFrequency
)

type Options struct {
	Colors         int
	Method         Method
	Sort           SortMode
	AlphaThreshold uint8
}

func Extract(img *core.Image, colors int) (core.Palette, error) {
	return ExtractWithOptions(img, Options{Colors: colors, Method: MethodMedianCut, Sort: SortRGB})
}

func ExtractWithOptions(img *core.Image, opts Options) (core.Palette, error) {
	if err := img.Validate(); err != nil {
		return nil, err
	}
	if opts.Colors < 1 || opts.Colors > 256 {
		return nil, core.ErrPaletteTooLarge
	}
	hist := buildHistogram(img, opts.AlphaThreshold)
	if len(hist) == 0 {
		return core.Palette{{0, 0, 0}}, nil
	}
	if len(hist) <= opts.Colors {
		out := make(core.Palette, len(hist))
		for i, entry := range hist {
			out[i] = entry.color
		}
		sortPalette(out, hist, opts.Sort)
		return out, nil
	}
	if opts.Method == MethodPopularity {
		palette := make(core.Palette, 0, opts.Colors)
		for i := 0; i < opts.Colors && i < len(hist); i++ {
			palette = append(palette, hist[i].color)
		}
		sortPalette(palette, hist, opts.Sort)
		return palette, nil
	}
	boxes := []bucket{{colors: hist}}
	for len(boxes) < opts.Colors {
		index := pickSplitBucket(boxes)
		if index < 0 {
			break
		}
		left, right, ok := splitBucket(boxes[index])
		if !ok {
			break
		}
		boxes[index] = boxes[len(boxes)-1]
		boxes = boxes[:len(boxes)-1]
		boxes = append(boxes, left, right)
	}
	out := make(core.Palette, 0, len(boxes))
	weights := make([]weightedColor, 0, len(boxes))
	for _, box := range boxes {
		color, weight := averageColor(box.colors)
		out = append(out, color)
		weights = append(weights, weightedColor{color: color, weight: weight})
	}
	sortPalette(out, weights, opts.Sort)
	return out, nil
}

func buildHistogram(img *core.Image, alphaThreshold uint8) []weightedColor {
	counts := make(map[uint32]int, img.Width*img.Height)
	channels := img.ChannelCount()
	for y := 0; y < img.Height; y++ {
		row := img.Row(y)
		for x := 0; x < img.Width; x++ {
			offset := x * channels
			var c core.Color
			switch img.Format {
			case core.Gray8:
				v := row[offset]
				c = core.Color{R: v, G: v, B: v}
			case core.RGB8:
				c = core.Color{R: row[offset], G: row[offset+1], B: row[offset+2]}
			case core.RGBA8:
				if row[offset+3] <= alphaThreshold {
					continue
				}
				c = core.Color{R: row[offset], G: row[offset+1], B: row[offset+2]}
			}
			key := uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B)
			counts[key]++
		}
	}
	out := make([]weightedColor, 0, len(counts))
	for key, weight := range counts {
		out = append(out, weightedColor{
			color: core.Color{
				R: uint8(key >> 16),
				G: uint8(key >> 8),
				B: uint8(key),
			},
			weight: weight,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].weight != out[j].weight {
			return out[i].weight > out[j].weight
		}
		if out[i].color.R != out[j].color.R {
			return out[i].color.R < out[j].color.R
		}
		if out[i].color.G != out[j].color.G {
			return out[i].color.G < out[j].color.G
		}
		return out[i].color.B < out[j].color.B
	})
	return out
}

func pickSplitBucket(boxes []bucket) int {
	best := -1
	bestScore := -1
	for i, box := range boxes {
		if len(box.colors) < 2 {
			continue
		}
		score := bucketRange(box.colors)
		if score > bestScore {
			bestScore = score
			best = i
		}
	}
	return best
}

func splitBucket(box bucket) (bucket, bucket, bool) {
	axis := bucketAxis(box.colors)
	sort.Slice(box.colors, func(i, j int) bool {
		left, right := box.colors[i].color, box.colors[j].color
		switch axis {
		case 0:
			if left.R != right.R {
				return left.R < right.R
			}
		case 1:
			if left.G != right.G {
				return left.G < right.G
			}
		default:
			if left.B != right.B {
				return left.B < right.B
			}
		}
		return box.colors[i].weight > box.colors[j].weight
	})
	total := 0
	for _, entry := range box.colors {
		total += entry.weight
	}
	target := total / 2
	accum := 0
	split := 1
	for split < len(box.colors) {
		accum += box.colors[split-1].weight
		if accum >= target {
			break
		}
		split++
	}
	if split <= 0 || split >= len(box.colors) {
		return bucket{}, bucket{}, false
	}
	left := append([]weightedColor(nil), box.colors[:split]...)
	right := append([]weightedColor(nil), box.colors[split:]...)
	return bucket{colors: left}, bucket{colors: right}, true
}

func bucketAxis(colors []weightedColor) int {
	var rMin, gMin, bMin uint8 = 255, 255, 255
	var rMax, gMax, bMax uint8
	for _, entry := range colors {
		c := entry.color
		if c.R < rMin {
			rMin = c.R
		}
		if c.R > rMax {
			rMax = c.R
		}
		if c.G < gMin {
			gMin = c.G
		}
		if c.G > gMax {
			gMax = c.G
		}
		if c.B < bMin {
			bMin = c.B
		}
		if c.B > bMax {
			bMax = c.B
		}
	}
	rRange := int(rMax) - int(rMin)
	gRange := int(gMax) - int(gMin)
	bRange := int(bMax) - int(bMin)
	if rRange >= gRange && rRange >= bRange {
		return 0
	}
	if gRange >= bRange {
		return 1
	}
	return 2
}

func bucketRange(colors []weightedColor) int {
	var rMin, gMin, bMin uint8 = 255, 255, 255
	var rMax, gMax, bMax uint8
	for _, entry := range colors {
		c := entry.color
		if c.R < rMin {
			rMin = c.R
		}
		if c.R > rMax {
			rMax = c.R
		}
		if c.G < gMin {
			gMin = c.G
		}
		if c.G > gMax {
			gMax = c.G
		}
		if c.B < bMin {
			bMin = c.B
		}
		if c.B > bMax {
			bMax = c.B
		}
	}
	return max3(int(rMax)-int(rMin), int(gMax)-int(gMin), int(bMax)-int(bMin))
}

func averageColor(colors []weightedColor) (core.Color, int) {
	var rSum, gSum, bSum, weightSum int
	for _, entry := range colors {
		rSum += int(entry.color.R) * entry.weight
		gSum += int(entry.color.G) * entry.weight
		bSum += int(entry.color.B) * entry.weight
		weightSum += entry.weight
	}
	if weightSum == 0 {
		return core.Color{}, 0
	}
	return core.Color{
		R: uint8((rSum + weightSum/2) / weightSum),
		G: uint8((gSum + weightSum/2) / weightSum),
		B: uint8((bSum + weightSum/2) / weightSum),
	}, weightSum
}

func max3(a, b, c int) int {
	if a >= b && a >= c {
		return a
	}
	if b >= c {
		return b
	}
	return c
}

func sortPalette(palette core.Palette, weights []weightedColor, mode SortMode) {
	switch mode {
	case SortLuma:
		sort.SliceStable(palette, func(i, j int) bool {
			left := colorWeight(weights, palette[i])
			right := colorWeight(weights, palette[j])
			ll := int(palette[i].R)*299 + int(palette[i].G)*587 + int(palette[i].B)*114
			rl := int(palette[j].R)*299 + int(palette[j].G)*587 + int(palette[j].B)*114
			if ll != rl {
				return ll < rl
			}
			return left > right
		})
	case SortFrequency:
		sort.SliceStable(palette, func(i, j int) bool {
			left := colorWeight(weights, palette[i])
			right := colorWeight(weights, palette[j])
			if left != right {
				return left > right
			}
			return rgbLess(palette[i], palette[j])
		})
	default:
		sort.SliceStable(palette, func(i, j int) bool {
			return rgbLess(palette[i], palette[j])
		})
	}
}

func colorWeight(weights []weightedColor, color core.Color) int {
	for _, entry := range weights {
		if entry.color == color {
			return entry.weight
		}
	}
	return 0
}

func rgbLess(left, right core.Color) bool {
	if left.R != right.R {
		return left.R < right.R
	}
	if left.G != right.G {
		return left.G < right.G
	}
	return left.B < right.B
}
