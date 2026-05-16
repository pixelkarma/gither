package engine

import (
	"github.com/pixelkarma/gither/internal/core"
	"github.com/pixelkarma/gither/internal/mathx"
)

func Threshold(img *core.Image, opts core.Options) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	channels := img.ChannelCount()
	parallelRows(img.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			row := img.Row(y)
			for x := 0; x < img.Width; x++ {
				offset := x * channels
				switch img.Format {
				case core.Gray8:
					var binary uint8
					if row[offset] > opts.Threshold {
						binary = 255
					}
					row[offset] = opts.Quantizer.QuantizeGrayFromRGB(binary, binary, binary)
				case core.RGB8:
					var binary uint8
					if mathx.LumaByte(row[offset], row[offset+1], row[offset+2]) > opts.Threshold {
						binary = 255
					}
					row[offset], row[offset+1], row[offset+2] = opts.Quantizer.QuantizeColor(binary, binary, binary)
				case core.RGBA8:
					alpha := row[offset+3]
					var binary uint8
					if mathx.LumaByte(row[offset], row[offset+1], row[offset+2]) > opts.Threshold {
						binary = 255
					}
					row[offset], row[offset+1], row[offset+2] = opts.Quantizer.QuantizeColor(binary, binary, binary)
					row[offset+3] = alpha
				}
			}
		}
	})
	return nil
}

func Random(img *core.Image, opts core.Options) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	opts = opts.WithDefaults()
	channels := img.ChannelCount()
	parallelRows(img.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			row := img.Row(y)
			for x := 0; x < img.Width; x++ {
				offset := x * channels
				threshold := randomThreshold(opts.Seed, x, y, opts.RandomStrength)
				switch img.Format {
				case core.Gray8:
					var binary uint8
					if row[offset] > threshold {
						binary = 255
					}
					row[offset] = opts.Quantizer.QuantizeGrayFromRGB(binary, binary, binary)
				case core.RGB8:
					var binary uint8
					if mathx.LumaByte(row[offset], row[offset+1], row[offset+2]) > threshold {
						binary = 255
					}
					row[offset], row[offset+1], row[offset+2] = opts.Quantizer.QuantizeColor(binary, binary, binary)
				case core.RGBA8:
					alpha := row[offset+3]
					var binary uint8
					if mathx.LumaByte(row[offset], row[offset+1], row[offset+2]) > threshold {
						binary = 255
					}
					row[offset], row[offset+1], row[offset+2] = opts.Quantizer.QuantizeColor(binary, binary, binary)
					row[offset+3] = alpha
				}
			}
		}
	})
	return nil
}

func randomThreshold(seed uint64, x, y int, strength uint8) uint8 {
	if strength == 0 {
		return 127
	}
	span := uint64(strength)*2 + 1
	mixed := coordinateMix(seed, x, y)
	jitter := int(mixed%span) - int(strength)
	return uint8(mathx.ClampInt(127+jitter, 0, 255))
}

func coordinateMix(seed uint64, x, y int) uint64 {
	value := seed ^ (uint64(x) * 0x9e3779b185ebca87) ^ (uint64(y) * 0xc2b2ae3d27d4eb4f)
	return mathx.Mix64(value)
}
