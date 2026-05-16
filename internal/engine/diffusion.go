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
	info := compileKernel(kernel)
	switch img.Format {
	case core.Gray8:
		return applyDiffusionGray(img, opts, info)
	case core.RGB8:
		return applyDiffusionRGB(img, opts, info, false)
	case core.RGBA8:
		return applyDiffusionRGB(img, opts, info, true)
	}
	return nil
}

type compiledTap struct {
	dx, dy int
	weight float32
}

type compiledKernel struct {
	taps  []compiledTap
	maxDY int
}

func compileKernel(kernel kernels.ErrorKernel) compiledKernel {
	out := compiledKernel{taps: make([]compiledTap, len(kernel.Taps))}
	den := float32(kernel.WeightDen)
	for i, tap := range kernel.Taps {
		out.taps[i] = compiledTap{
			dx:     int(tap.DX),
			dy:     int(tap.DY),
			weight: float32(tap.WeightNum) / den,
		}
		if int(tap.DY) > out.maxDY {
			out.maxDY = int(tap.DY)
		}
	}
	return out
}

func applyDiffusionGray(img *core.Image, opts core.Options, kernel compiledKernel) error {
	width, height := img.Width, img.Height
	rowCount := kernel.maxDY + 1
	if rowCount < 1 {
		rowCount = 1
	}
	errors := make([][]float32, rowCount)
	for i := range errors {
		errors[i] = make([]float32, width)
	}
	for y := 0; y < height; y++ {
		row := img.Row(y)
		current := errors[y%rowCount]
		for x := 0; x < width; x++ {
			adjusted := mathx.ClampFloat32(mathx.ByteToUnit(row[x])+current[x], 0, 1)
			gray := mathx.UnitToByte(adjusted)
			quantized := opts.Quantizer.QuantizeGrayFromRGB(gray, gray, gray)
			row[x] = quantized
			residual := adjusted - mathx.ByteToUnit(quantized)
			for _, tap := range kernel.taps {
				ty := y + tap.dy
				tx := x + tap.dx
				if ty < 0 || ty >= height || tx < 0 || tx >= width {
					continue
				}
				errors[ty%rowCount][tx] += residual * tap.weight
			}
		}
		clear(current)
	}
	return nil
}

func applyDiffusionRGB(img *core.Image, opts core.Options, kernel compiledKernel, hasAlpha bool) error {
	width, height := img.Width, img.Height
	rowCount := kernel.maxDY + 1
	if rowCount < 1 {
		rowCount = 1
	}
	errors := make([][]float32, rowCount)
	for i := range errors {
		errors[i] = make([]float32, width*3)
	}
	channels := 3
	if hasAlpha {
		channels = 4
	}
	for y := 0; y < height; y++ {
		row := img.Row(y)
		current := errors[y%rowCount]
		for x := 0; x < width; x++ {
			pix := x * channels
			errBase := x * 3
			rAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[pix])+current[errBase], 0, 1)
			gAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[pix+1])+current[errBase+1], 0, 1)
			bAdj := mathx.ClampFloat32(mathx.ByteToUnit(row[pix+2])+current[errBase+2], 0, 1)
			qr, qg, qb := opts.Quantizer.QuantizeColor(mathx.UnitToByte(rAdj), mathx.UnitToByte(gAdj), mathx.UnitToByte(bAdj))
			if hasAlpha {
				alpha := row[pix+3]
				row[pix], row[pix+1], row[pix+2], row[pix+3] = qr, qg, qb, alpha
			} else {
				row[pix], row[pix+1], row[pix+2] = qr, qg, qb
			}
			rRes := rAdj - mathx.ByteToUnit(qr)
			gRes := gAdj - mathx.ByteToUnit(qg)
			bRes := bAdj - mathx.ByteToUnit(qb)
			for _, tap := range kernel.taps {
				ty := y + tap.dy
				tx := x + tap.dx
				if ty < 0 || ty >= height || tx < 0 || tx >= width {
					continue
				}
				target := errors[ty%rowCount]
				base := tx * 3
				w := tap.weight
				target[base] += rRes * w
				target[base+1] += gRes * w
				target[base+2] += bRes * w
			}
		}
		clear(current)
	}
	return nil
}
