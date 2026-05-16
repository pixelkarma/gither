package engine

import (
	"errors"
	"math"

	"gither/internal/core"
	"gither/internal/mathx"
)

type VariableAnchor struct {
	Tone     uint8
	Forward  float32
	DownDiag float32
	Down     float32
}

type VariableCurve struct {
	coeffs [256][3]float32
}

var (
	errInvalidVariableCurve = errors.New("variable diffusion curve must contain at least two anchors from tone 0 to 255")
	ostromoukhovAnchors     = []VariableAnchor{
		{0, 0.72, 0.00, 0.28},
		{48, 0.65, 0.08, 0.27},
		{96, 0.56, 0.17, 0.27},
		{128, 0.46, 0.27, 0.27},
		{160, 0.34, 0.36, 0.30},
		{208, 0.20, 0.46, 0.34},
		{255, 0.08, 0.56, 0.36},
	}
	balancedAnchors = []VariableAnchor{
		{0, 0.64, 0.08, 0.28},
		{64, 0.57, 0.16, 0.27},
		{128, 0.44, 0.28, 0.28},
		{192, 0.26, 0.42, 0.32},
		{255, 0.12, 0.50, 0.38},
	}
)

func Ostromoukhov(img *core.Image, opts core.Options) error {
	return ApplyVariableGray(img, opts, OstromoukhovCurve(), false)
}

func ZhouFang(img *core.Image, opts core.Options) error {
	return ApplyVariableGray(img, opts, OstromoukhovCurve(), true)
}

func BalancedVariable(img *core.Image, opts core.Options) error {
	return ApplyVariableGray(img, opts, BalancedCurve(), false)
}

func BalancedVariableThresholded(img *core.Image, opts core.Options) error {
	return ApplyVariableGray(img, opts, BalancedCurve(), true)
}

func ApplyVariableGray(img *core.Image, opts core.Options, curve VariableCurve, thresholded bool) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	width, height := img.Width, img.Height
	currentErrors := make([]float32, width)
	nextErrors := make([]float32, width)
	for y := 0; y < height; y++ {
		leftToRight := y%2 == 0
		xStart, xEnd, xStep := 0, width, 1
		if !leftToRight {
			xStart, xEnd, xStep = width-1, -1, -1
		}
		for x := xStart; x != xEnd; x += xStep {
			offset := img.PixelOffset(x, y)
			adjusted := mathx.ClampFloat32(grayUnitAt(img, offset)+currentErrors[x], 0, 1)
			decision := adjusted
			if thresholded {
				decision = mathx.ClampFloat32(adjusted+zhouFangJitter(x, y, opts.Seed, adjusted), 0, 1)
			}
			gray := mathx.UnitToByte(decision)
			quantized := opts.Quantizer.QuantizeGrayFromRGB(gray, gray, gray)
			writeGrayAt(img, offset, quantized)
			residual := adjusted - mathx.ByteToUnit(quantized)
			forward, downDiag, down := curve.Coefficients(mathx.UnitToByte(adjusted))
			if nx := x + xStep; nx >= 0 && nx < width {
				currentErrors[nx] += residual * forward
				nextErrors[nx] += residual * downDiag
			}
			nextErrors[x] += residual * down
		}
		clear(currentErrors)
		currentErrors, nextErrors = nextErrors, currentErrors
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

func NewVariableCurve(anchors []VariableAnchor) (VariableCurve, error) {
	if len(anchors) < 2 || anchors[0].Tone != 0 || anchors[len(anchors)-1].Tone != 255 {
		return VariableCurve{}, errInvalidVariableCurve
	}
	curve := VariableCurve{}
	for i := 1; i < len(anchors); i++ {
		left, right := anchors[i-1], anchors[i]
		if right.Tone < left.Tone {
			return VariableCurve{}, errInvalidVariableCurve
		}
		span := int(right.Tone) - int(left.Tone)
		if span == 0 {
			curve.coeffs[left.Tone] = normalizeCoefficients(left.Forward, left.DownDiag, left.Down)
			continue
		}
		for tone := int(left.Tone); tone <= int(right.Tone); tone++ {
			alpha := float32(tone-int(left.Tone)) / float32(span)
			curve.coeffs[tone] = normalizeCoefficients(
				lerp(left.Forward, right.Forward, alpha),
				lerp(left.DownDiag, right.DownDiag, alpha),
				lerp(left.Down, right.Down, alpha),
			)
		}
	}
	return curve, nil
}

func (c VariableCurve) Coefficients(tone uint8) (float32, float32, float32) {
	coeff := c.coeffs[tone]
	return coeff[0], coeff[1], coeff[2]
}

func OstromoukhovCurve() VariableCurve {
	curve, _ := NewVariableCurve(ostromoukhovAnchors)
	return curve
}

func BalancedCurve() VariableCurve {
	curve, _ := NewVariableCurve(balancedAnchors)
	return curve
}

func normalizeCoefficients(forward, downDiag, down float32) [3]float32 {
	sum := forward + downDiag + down
	if sum <= 0 {
		return [3]float32{0.5, 0.25, 0.25}
	}
	return [3]float32{forward / sum, downDiag / sum, down / sum}
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
