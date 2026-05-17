package engine

import (
	"errors"
	"math"

	"github.com/pixelkarma/gither/internal/core"
	"github.com/pixelkarma/gither/internal/mathx"
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
	smoothAnchors = []VariableAnchor{
		{0, 0.58, 0.10, 0.32},
		{64, 0.50, 0.18, 0.32},
		{128, 0.40, 0.28, 0.32},
		{192, 0.26, 0.38, 0.36},
		{255, 0.14, 0.46, 0.40},
	}
	punchyAnchors = []VariableAnchor{
		{0, 0.76, 0.02, 0.22},
		{64, 0.66, 0.10, 0.24},
		{128, 0.52, 0.22, 0.26},
		{192, 0.28, 0.38, 0.34},
		{255, 0.10, 0.54, 0.36},
	}
)

func Ostromoukhov(img *core.Image, opts core.Options) error {
	return applyReferenceVariableGray(img, opts, nil)
}

func ZhouFang(img *core.Image, opts core.Options) error {
	return applyReferenceVariableGray(img, opts, zhouFangModulation[:])
}

func BalancedVariable(img *core.Image, opts core.Options) error {
	return ApplyVariableGray(img, opts, BalancedCurve(), false)
}

func BalancedVariableThresholded(img *core.Image, opts core.Options) error {
	return ApplyVariableGray(img, opts, BalancedCurve(), true)
}
func SmoothVariable(img *core.Image, opts core.Options) error {
	return ApplyVariableGray(img, opts, SmoothCurve(), false)
}
func PunchyVariable(img *core.Image, opts core.Options) error {
	return ApplyVariableGray(img, opts, PunchyCurve(), false)
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

func applyReferenceVariableGray(img *core.Image, opts core.Options, modulation []uint8) error {
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
		reverse := y%2 == 1
		xStart, xEnd, xStep := 0, width, 1
		if reverse {
			xStart, xEnd, xStep = width-1, -1, -1
		}
		for x := xStart; x != xEnd; x += xStep {
			offset := img.PixelOffset(x, y)
			adjusted := mathx.ClampFloat32(grayUnitAt(img, offset)+currentErrors[x], 0, 1)
			thresholded := adjusted
			if modulation != nil {
				thresholded = zhouFangThresholdedUnit(adjusted, x, y, modulation)
			}
			gray := mathx.UnitToByte(thresholded)
			quantized := opts.Quantizer.QuantizeGrayFromRGB(gray, gray, gray)
			writeGrayAt(img, offset, quantized)

			residual := adjusted - mathx.ByteToUnit(quantized)
			coeff := ostromoukhovCoeffs[mathx.UnitToByte(adjusted)]
			forwardNum, downDiagNum, downNum, den := coeff[0], coeff[1], coeff[2], coeff[3]
			forward := residual * float32(forwardNum) / float32(den)
			downDiag := residual * float32(downDiagNum) / float32(den)
			down := residual * float32(downNum) / float32(den)
			if nx := x + xStep; nx >= 0 && nx < width {
				currentErrors[nx] += forward
			}
			if nx := x - xStep; nx >= 0 && nx < width {
				nextErrors[nx] += downDiag
			}
			nextErrors[x] += down
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
func SmoothCurve() VariableCurve {
	curve, _ := NewVariableCurve(smoothAnchors)
	return curve
}
func PunchyCurve() VariableCurve {
	curve, _ := NewVariableCurve(punchyAnchors)
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

func zhouFangThresholdedUnit(value float32, x, y int, modulation []uint8) float32 {
	luma := mathx.UnitToByte(value)
	amplitude := float32(modulation[luma]) / 255.0
	noise := zhouFangNoiseUnit(x, y)
	return mathx.ClampFloat32(value+(noise-0.5)*amplitude, 0, 1)
}

func zhouFangNoiseUnit(x, y int) float32 {
	state := uint64(x)*0x9E3779B97F4A7C15 ^ uint64(y)*0xBF58476D1CE4E5B9 ^ 0x94D049BB133111EB
	state ^= state >> 30
	state *= 0xBF58476D1CE4E5B9
	state ^= state >> 27
	state *= 0x94D049BB133111EB
	state ^= state >> 31
	return float32(state>>40) / 16777215.0
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

var zhouFangModulation = func() [256]uint8 {
	keyLevels := [...]uint8{0, 32, 64, 96, 112, 128, 144, 160, 192, 255}
	keyAmplitudes := [...]uint8{0, 87, 133, 163, 184, 199, 184, 163, 133, 87}
	var out [256]uint8
	for i := 0; i < 256; i++ {
		x := i
		set := false
		for seg := 0; seg+1 < len(keyLevels); seg++ {
			x0 := int(keyLevels[seg])
			x1 := int(keyLevels[seg+1])
			if x >= x0 && x <= x1 {
				if x1 <= x0 {
					out[i] = keyAmplitudes[seg]
				} else {
					dx := x1 - x0
					tx := x - x0
					dy := int(keyAmplitudes[seg+1]) - int(keyAmplitudes[seg])
					num := dy * tx
					den := dx
					delta := 0
					if num >= 0 {
						delta = (num + den/2) / den
					} else {
						delta = (num - den/2) / den
					}
					value := int(keyAmplitudes[seg]) + delta
					if value < 0 {
						value = 0
					}
					if value > 255 {
						value = 255
					}
					out[i] = uint8(value)
				}
				set = true
				break
			}
		}
		if !set {
			out[i] = keyAmplitudes[len(keyAmplitudes)-1]
		}
	}
	return out
}()

var ostromoukhovCoeffs = [256][4]int16{
	{13, 0, 5, 18}, {13, 0, 5, 18}, {21, 0, 10, 31}, {7, 0, 4, 11}, {8, 0, 5, 13}, {47, 3, 28, 78}, {23, 3, 13, 39}, {15, 3, 8, 26},
	{22, 6, 11, 39}, {43, 15, 20, 78}, {7, 3, 3, 13}, {501, 224, 211, 936}, {249, 116, 103, 468}, {165, 80, 67, 312}, {123, 62, 49, 234}, {489, 256, 191, 936},
	{81, 44, 31, 156}, {483, 272, 181, 936}, {60, 35, 22, 117}, {53, 32, 19, 104}, {237, 148, 83, 468}, {471, 304, 161, 936}, {3, 2, 1, 6}, {481, 314, 185, 980},
	{354, 226, 155, 735}, {1389, 866, 685, 2940}, {227, 138, 125, 490}, {267, 158, 163, 588}, {327, 188, 220, 735}, {61, 34, 45, 140}, {627, 338, 505, 1470}, {1227, 638, 1075, 2940},
	{20, 10, 19, 49}, {1937, 1000, 1767, 4704}, {977, 520, 855, 2352}, {657, 360, 551, 1568}, {71, 40, 57, 168}, {2005, 1160, 1539, 4704}, {337, 200, 247, 784}, {2039, 1240, 1425, 4704},
	{257, 160, 171, 588}, {691, 440, 437, 1568}, {1045, 680, 627, 2352}, {301, 200, 171, 672}, {177, 120, 95, 392}, {2141, 1480, 1083, 4704}, {1079, 760, 513, 2352}, {725, 520, 323, 1568},
	{137, 100, 57, 294}, {2209, 1640, 855, 4704}, {53, 40, 19, 112}, {2243, 1720, 741, 4704}, {565, 440, 171, 1176}, {759, 600, 209, 1568}, {1147, 920, 285, 2352}, {2311, 1880, 513, 4704},
	{97, 80, 19, 196}, {335, 280, 57, 672}, {1181, 1000, 171, 2352}, {793, 680, 95, 1568}, {599, 520, 57, 1176}, {2413, 2120, 171, 4704}, {405, 360, 19, 784}, {2447, 2200, 57, 4704},
	{11, 10, 0, 21}, {158, 151, 3, 312}, {178, 179, 7, 364}, {1030, 1091, 63, 2184}, {248, 277, 21, 546}, {318, 375, 35, 728}, {458, 571, 63, 1092}, {878, 1159, 147, 2184},
	{5, 7, 1, 13}, {172, 181, 37, 390}, {97, 76, 22, 195}, {72, 41, 17, 130}, {119, 47, 29, 195}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6},
	{4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {65, 18, 17, 100}, {95, 29, 26, 150},
	{185, 62, 53, 300}, {30, 11, 9, 50}, {35, 14, 11, 60}, {85, 37, 28, 150}, {55, 26, 19, 100}, {80, 41, 29, 150}, {155, 86, 59, 300}, {5, 3, 2, 10},
	{5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10},
	{5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {305, 176, 119, 600}, {155, 86, 59, 300}, {105, 56, 39, 200}, {80, 41, 29, 150},
	{65, 32, 23, 120}, {55, 26, 19, 100}, {335, 152, 113, 600}, {85, 37, 28, 150}, {115, 48, 37, 200}, {35, 14, 11, 60}, {355, 136, 109, 600}, {30, 11, 9, 50},
	{365, 128, 107, 600}, {185, 62, 53, 300}, {25, 8, 7, 40}, {95, 29, 26, 150}, {385, 112, 103, 600}, {65, 18, 17, 100}, {395, 104, 101, 600}, {4, 1, 1, 6},
	{4, 1, 1, 6}, {395, 104, 101, 600}, {65, 18, 17, 100}, {385, 112, 103, 600}, {95, 29, 26, 150}, {25, 8, 7, 40}, {185, 62, 53, 300}, {365, 128, 107, 600},
	{30, 11, 9, 50}, {355, 136, 109, 600}, {35, 14, 11, 60}, {115, 48, 37, 200}, {85, 37, 28, 150}, {335, 152, 113, 600}, {55, 26, 19, 100}, {65, 32, 23, 120},
	{80, 41, 29, 150}, {105, 56, 39, 200}, {155, 86, 59, 300}, {305, 176, 119, 600}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10},
	{5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {5, 3, 2, 10}, {155, 86, 59, 300},
	{80, 41, 29, 150}, {55, 26, 19, 100}, {85, 37, 28, 150}, {35, 14, 11, 60}, {30, 11, 9, 50}, {185, 62, 53, 300}, {95, 29, 26, 150}, {65, 18, 17, 100}, {4, 1, 1, 6},
	{4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {4, 1, 1, 6}, {119, 47, 29, 195}, {72, 41, 17, 130},
	{97, 76, 22, 195}, {172, 181, 37, 390}, {5, 7, 1, 13}, {878, 1159, 147, 2184}, {458, 571, 63, 1092}, {318, 375, 35, 728}, {248, 277, 21, 546}, {1030, 1091, 63, 2184},
	{178, 179, 7, 364}, {158, 151, 3, 312}, {11, 10, 0, 21}, {2447, 2200, 57, 4704}, {405, 360, 19, 784}, {2413, 2120, 171, 4704}, {599, 520, 57, 1176}, {793, 680, 95, 1568},
	{1181, 1000, 171, 2352}, {335, 280, 57, 672}, {97, 80, 19, 196}, {2311, 1880, 513, 4704}, {1147, 920, 285, 2352}, {759, 600, 209, 1568}, {565, 440, 171, 1176}, {2243, 1720, 741, 4704},
	{53, 40, 19, 112}, {2209, 1640, 855, 4704}, {137, 100, 57, 294}, {725, 520, 323, 1568}, {1079, 760, 513, 2352}, {2141, 1480, 1083, 4704}, {177, 120, 95, 392}, {301, 200, 171, 672},
	{1045, 680, 627, 2352}, {691, 440, 437, 1568}, {257, 160, 171, 588}, {2039, 1240, 1425, 4704}, {337, 200, 247, 784}, {2005, 1160, 1539, 4704}, {71, 40, 57, 168}, {657, 360, 551, 1568},
	{977, 520, 855, 2352}, {1937, 1000, 1767, 4704}, {20, 10, 19, 49}, {1227, 638, 1075, 2940}, {627, 338, 505, 1470}, {61, 34, 45, 140}, {327, 188, 220, 735}, {267, 158, 163, 588},
	{227, 138, 125, 490}, {1389, 866, 685, 2940}, {354, 226, 155, 735}, {481, 314, 185, 980}, {3, 2, 1, 6}, {471, 304, 161, 936}, {237, 148, 83, 468}, {53, 32, 19, 104},
	{60, 35, 22, 117}, {483, 272, 181, 936}, {81, 44, 31, 156}, {489, 256, 191, 936}, {123, 62, 49, 234}, {165, 80, 67, 312}, {249, 116, 103, 468}, {501, 224, 211, 936},
	{7, 3, 3, 13}, {43, 15, 20, 78}, {22, 6, 11, 39}, {15, 3, 8, 26}, {23, 3, 13, 39}, {47, 3, 28, 78}, {8, 0, 5, 13}, {7, 0, 4, 11}, {21, 0, 10, 31}, {13, 0, 5, 18}, {13, 0, 5, 18},
}
