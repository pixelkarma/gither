package engine

import (
	"math"

	"gither/internal/core"
	"gither/internal/mathx"
)

func Ostromoukhov(img *core.Image, opts core.Options) error {
	return applyVariableGray(img, opts, false)
}

func ZhouFang(img *core.Image, opts core.Options) error {
	return applyVariableGray(img, opts, true)
}

func applyVariableGray(img *core.Image, opts core.Options, thresholded bool) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	width, height := img.Width, img.Height
	errors := make([]float32, width*height)
	for y := 0; y < height; y++ {
		leftToRight := y%2 == 0
		xStart, xEnd, xStep := 0, width, 1
		if !leftToRight {
			xStart, xEnd, xStep = width-1, -1, -1
		}
		for x := xStart; x != xEnd; x += xStep {
			offset := img.PixelOffset(x, y)
			adjusted := mathx.ClampFloat32(grayUnitAt(img, offset)+errors[y*width+x], 0, 1)
			decision := adjusted
			if thresholded {
				decision = mathx.ClampFloat32(adjusted+zhouFangJitter(x, y, opts.Seed, adjusted), 0, 1)
			}
			gray := mathx.UnitToByte(decision)
			quantized := opts.Quantizer.QuantizeGrayFromRGB(gray, gray, gray)
			writeGrayAt(img, offset, quantized)
			residual := adjusted - mathx.ByteToUnit(quantized)
			forward, downDiag, down := coefficientsForTone(mathx.UnitToByte(adjusted))
			distributeVariableError(errors, width, height, x, y, xStep, residual, forward, downDiag, down)
		}
	}
	return nil
}

func grayUnitAt(img *core.Image, offset int) float32 {
	switch img.Format {
	case core.Gray8:
		return mathx.ByteToUnit(img.Pix[offset])
	default:
		return mathx.ByteToUnit(mathx.LumaByte(img.Pix[offset], img.Pix[offset+1], img.Pix[offset+2]))
	}
}

func writeGrayAt(img *core.Image, offset int, gray uint8) {
	switch img.Format {
	case core.Gray8:
		img.Pix[offset] = gray
	case core.RGB8:
		img.Pix[offset], img.Pix[offset+1], img.Pix[offset+2] = gray, gray, gray
	case core.RGBA8:
		alpha := img.Pix[offset+3]
		img.Pix[offset], img.Pix[offset+1], img.Pix[offset+2], img.Pix[offset+3] = gray, gray, gray, alpha
	}
}

func distributeVariableError(errors []float32, width, height, x, y, xStep int, residual, forward, downDiag, down float32) {
	pushGrayError(errors, width, height, x+xStep, y, residual*forward)
	pushGrayError(errors, width, height, x+xStep, y+1, residual*downDiag)
	pushGrayError(errors, width, height, x, y+1, residual*down)
}

func pushGrayError(errors []float32, width, height, x, y int, value float32) {
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	errors[y*width+x] += value
}

func coefficientsForTone(tone uint8) (float32, float32, float32) {
	type anchor struct {
		tone int
		f    float32
		dd   float32
		d    float32
	}
	anchors := [...]anchor{
		{0, 0.72, 0.00, 0.28},
		{48, 0.65, 0.08, 0.27},
		{96, 0.56, 0.17, 0.27},
		{128, 0.46, 0.27, 0.27},
		{160, 0.34, 0.36, 0.30},
		{208, 0.20, 0.46, 0.34},
		{255, 0.08, 0.56, 0.36},
	}
	t := int(tone)
	for i := 1; i < len(anchors); i++ {
		if t <= anchors[i].tone {
			left, right := anchors[i-1], anchors[i]
			span := float32(right.tone - left.tone)
			if span == 0 {
				return left.f, left.dd, left.d
			}
			alpha := float32(t-left.tone) / span
			return lerp(left.f, right.f, alpha), lerp(left.dd, right.dd, alpha), lerp(left.d, right.d, alpha)
		}
	}
	last := anchors[len(anchors)-1]
	return last.f, last.dd, last.d
}

func zhouFangJitter(x, y int, seed uint64, tone float32) float32 {
	amplitude := zhouFangAmplitude(tone)
	if amplitude == 0 {
		return 0
	}
	state := mathx.Mix64(seed + uint64((x+1)*73856093) + uint64((y+1)*19349663))
	noise := float32(int(state&1023)-512) / 512.0
	return noise * amplitude / 255.0
}

func zhouFangAmplitude(tone float32) float32 {
	distance := float32(math.Abs(float64(tone*255 - 127.5)))
	midBoost := 1.0 - distance/127.5
	if midBoost < 0 {
		midBoost = 0
	}
	return 4 + midBoost*22
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}
