package engine

import (
	"github.com/pixelkarma/gither/internal/core"
	"github.com/pixelkarma/gither/internal/mathx"
)

type OrderedMap struct {
	Values   []uint16
	Width    int
	Height   int
	Strength float32
}

func NewOrderedMap(values []uint16, width, height int, strength float32) (OrderedMap, error) {
	if width <= 0 || height <= 0 || len(values) != width*height {
		return OrderedMap{}, core.ErrInvalidOrderedMap
	}
	return OrderedMap{Values: values, Width: width, Height: height, Strength: strength}, nil
}

func NewOrderedMapFromU8(values []uint8, width, height int, strength float32) (OrderedMap, error) {
	if width <= 0 || height <= 0 || len(values) != width*height {
		return OrderedMap{}, core.ErrInvalidOrderedMap
	}
	limit := width * height
	converted := make([]uint16, limit)
	for i, value := range values {
		if int(value) >= limit {
			return OrderedMap{}, core.ErrInvalidOrderedMap
		}
		converted[i] = uint16(value)
	}
	return OrderedMap{Values: converted, Width: width, Height: height, Strength: strength}, nil
}

func ApplyOrdered(img *core.Image, ordered OrderedMap, opts core.Options) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	if ordered.Width <= 0 || ordered.Height <= 0 || len(ordered.Values) != ordered.Width*ordered.Height {
		return core.ErrInvalidOrderedMap
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
	strength := ordered.Strength
	if strength == 0 {
		strength = core.DefaultOrderedStrength
	}
	channels := img.ChannelCount()
	parallelRows(img.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			row := img.Row(y)
			mapRowBase := (y % ordered.Height) * ordered.Width
			for x := 0; x < img.Width; x++ {
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

func orderedThresholdUnit(rank, den uint16, strength float32) float32 {
	if den <= 1 || strength == 0 {
		return 0
	}
	rng := float32(den - 1)
	centered := float32(rank)*2 - rng
	scaledSteps := float32(int(centered * (strength * 255) / rng))
	return scaledSteps / 255
}

func StochasticClusteredDot(img *core.Image, opts core.Options, ordered OrderedMap) error {
	return ApplyOrdered(img, ordered, opts)
}
