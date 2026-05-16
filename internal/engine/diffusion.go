package engine

import (
	"gither/internal/core"
	"gither/internal/kernels"
	"gither/internal/mathx"
)

func ApplyDiffusion(img *core.Image, opts core.Options, kernel kernels.ErrorKernel) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	if kernel.WeightDen <= 0 {
		return core.ErrInvalidOrderedMap
	}
	width, height := img.Width, img.Height
	errors := make([]float32, width*height*4)
	channels := img.ChannelCount()
	den := float32(kernel.WeightDen)
	for y := 0; y < height; y++ {
		row := img.Row(y)
		for x := 0; x < width; x++ {
			offset := x * channels
			errIdx := (y*width + x) * 4
			switch img.Format {
			case core.Gray8:
				adjusted := mathx.ClampFloat32(mathx.ByteToUnit(row[offset])+errors[errIdx], 0, 1)
				gray := mathx.UnitToByte(adjusted)
				quantized := opts.Quantizer.QuantizeGrayFromRGB(gray, gray, gray)
				row[offset] = quantized
				residual := adjusted - mathx.ByteToUnit(quantized)
				for _, tap := range kernel.Taps {
					diffuseError(errors, width, height, x+int(tap.DX), y+int(tap.DY), residual*float32(tap.WeightNum)/den, 0, 0)
				}
			case core.RGB8:
				rAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[offset])+errors[errIdx], 0, 1)
				gAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[offset+1])+errors[errIdx+1], 0, 1)
				bAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[offset+2])+errors[errIdx+2], 0, 1)
				qr, qg, qb := opts.Quantizer.QuantizeColor(mathx.UnitToByte(rAdj), mathx.UnitToByte(gAdj), mathx.UnitToByte(bAdj))
				row[offset], row[offset+1], row[offset+2] = qr, qg, qb
				rRes := rAdj - mathx.ByteToUnit(qr)
				gRes := gAdj - mathx.ByteToUnit(qg)
				bRes := bAdj - mathx.ByteToUnit(qb)
				for _, tap := range kernel.Taps {
					w := float32(tap.WeightNum) / den
					diffuseError(errors, width, height, x+int(tap.DX), y+int(tap.DY), rRes*w, gRes*w, bRes*w)
				}
			case core.RGBA8:
				alpha := row[offset+3]
				rAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[offset])+errors[errIdx], 0, 1)
				gAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[offset+1])+errors[errIdx+1], 0, 1)
				bAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[offset+2])+errors[errIdx+2], 0, 1)
				qr, qg, qb := opts.Quantizer.QuantizeColor(mathx.UnitToByte(rAdj), mathx.UnitToByte(gAdj), mathx.UnitToByte(bAdj))
				row[offset], row[offset+1], row[offset+2], row[offset+3] = qr, qg, qb, alpha
				rRes := rAdj - mathx.ByteToUnit(qr)
				gRes := gAdj - mathx.ByteToUnit(qg)
				bRes := bAdj - mathx.ByteToUnit(qb)
				for _, tap := range kernel.Taps {
					w := float32(tap.WeightNum) / den
					diffuseError(errors, width, height, x+int(tap.DX), y+int(tap.DY), rRes*w, gRes*w, bRes*w)
				}
			}
		}
	}
	return nil
}

func diffuseError(errors []float32, width, height, x, y int, r, g, b float32) {
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	base := (y*width + x) * 4
	errors[base] += r
	errors[base+1] += g
	errors[base+2] += b
}
