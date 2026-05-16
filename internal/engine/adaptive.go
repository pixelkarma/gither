package engine

import (
	"gither/internal/core"
	"gither/internal/mathx"
)

func ApplyAdaptiveOrdered(img *core.Image, ordered OrderedMap, opts core.Options, radius int) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	if ordered.Width <= 0 || ordered.Height <= 0 || len(ordered.Values) != ordered.Width*ordered.Height {
		return core.ErrInvalidOrderedMap
	}
	if radius < 1 {
		radius = 1
	}
	mapMin, mapMax := ordered.Values[0], ordered.Values[0]
	for _, v := range ordered.Values[1:] {
		if v < mapMin {
			mapMin = v
		}
		if v > mapMax {
			mapMax = v
		}
	}
	thresholdDen := mapMax - mapMin + 1
	baseStrength := opts.WithDefaults().Strength
	channels := img.ChannelCount()
	luma := buildLumaPlane(img)
	localMean, localContrast := buildLocalStats(luma, img.Width, img.Height, radius)
	parallelRows(img.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			row := img.Row(y)
			mapRowBase := (y % ordered.Height) * ordered.Width
			for x := 0; x < img.Width; x++ {
				idx := y*img.Width + x
				lumaUnit := mathx.ByteToUnit(localMean[idx])
				contrastUnit := float32(localContrast[idx]) / 128.0
				strength := baseStrength * (0.55 + 0.6*(1.0-lumaUnit) + 0.35*contrastUnit)
				strength = mathx.ClampFloat32(strength, 0.08, 1.15)
				rank := ordered.Values[mapRowBase+(x%ordered.Width)] - mapMin
				threshold := orderedThresholdUnit(rank, thresholdDen, strength)
				offset := x * channels
				switch img.Format {
				case core.Gray8:
					gray := mathx.UnitToByte(mathx.ByteToUnit(row[offset]) + threshold)
					row[offset] = opts.Quantizer.QuantizeGrayFromRGB(gray, gray, gray)
				case core.RGB8:
					r := mathx.UnitToByte(mathx.ByteToUnit(row[offset]) + threshold)
					g := mathx.UnitToByte(mathx.ByteToUnit(row[offset+1]) + threshold)
					b := mathx.UnitToByte(mathx.ByteToUnit(row[offset+2]) + threshold)
					row[offset], row[offset+1], row[offset+2] = opts.Quantizer.QuantizeColor(r, g, b)
				case core.RGBA8:
					alpha := row[offset+3]
					r := mathx.UnitToByte(mathx.ByteToUnit(row[offset]) + threshold)
					g := mathx.UnitToByte(mathx.ByteToUnit(row[offset+1]) + threshold)
					b := mathx.UnitToByte(mathx.ByteToUnit(row[offset+2]) + threshold)
					row[offset], row[offset+1], row[offset+2] = opts.Quantizer.QuantizeColor(r, g, b)
					row[offset+3] = alpha
				}
			}
		}
	})
	return nil
}

func buildLumaPlane(img *core.Image) []uint8 {
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

func buildLocalStats(luma []uint8, width, height, radius int) ([]uint8, []uint8) {
	mean := make([]uint8, width*height)
	contrast := make([]uint8, width*height)
	for y := 0; y < height; y++ {
		y0 := adaptiveMaxInt(0, y-radius)
		y1 := adaptiveMinInt(height-1, y+radius)
		for x := 0; x < width; x++ {
			x0 := adaptiveMaxInt(0, x-radius)
			x1 := adaptiveMinInt(width-1, x+radius)
			sum, count := 0, 0
			minV, maxV := 255, 0
			for yy := y0; yy <= y1; yy++ {
				rowBase := yy * width
				for xx := x0; xx <= x1; xx++ {
					v := int(luma[rowBase+xx])
					sum += v
					count++
					if v < minV {
						minV = v
					}
					if v > maxV {
						maxV = v
					}
				}
			}
			idx := y*width + x
			mean[idx] = uint8(sum / count)
			contrast[idx] = uint8(maxV - minV)
		}
	}
	return mean, contrast
}

func adaptiveMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func adaptiveMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
