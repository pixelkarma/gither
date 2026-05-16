package core

import "github.com/pixelkarma/gither/internal/mathx"

type QuantizeKind uint8

const (
	QuantizeGrayLevels QuantizeKind = iota
	QuantizeRGBLevels
	QuantizePalette
	QuantizeSingleColor
)

type Quantizer struct {
	Kind        QuantizeKind
	Levels      int
	Palette     Palette
	SingleColor Color
}

func GrayLevels(levels int) Quantizer      { return Quantizer{Kind: QuantizeGrayLevels, Levels: levels} }
func RGBLevels(levels int) Quantizer       { return Quantizer{Kind: QuantizeRGBLevels, Levels: levels} }
func PaletteQuantizer(p Palette) Quantizer { return Quantizer{Kind: QuantizePalette, Palette: p} }
func SingleColorQuantizer(levels int, c Color) Quantizer {
	return Quantizer{Kind: QuantizeSingleColor, Levels: levels, SingleColor: c}
}

func (q Quantizer) Validate() error {
	switch q.Kind {
	case QuantizeGrayLevels, QuantizeRGBLevels, QuantizeSingleColor:
		if q.Levels < 2 {
			return ErrInvalidLevels
		}
	case QuantizePalette:
		return q.Palette.Validate()
	default:
		return ErrInvalidFormat
	}
	return nil
}

func (q Quantizer) QuantizeColor(r, g, b uint8) (uint8, uint8, uint8) {
	switch q.Kind {
	case QuantizeGrayLevels:
		gv := mathx.QuantizeByte(mathx.LumaByte(r, g, b), q.Levels)
		return gv, gv, gv
	case QuantizeRGBLevels:
		return mathx.QuantizeByte(r, q.Levels), mathx.QuantizeByte(g, q.Levels), mathx.QuantizeByte(b, q.Levels)
	case QuantizePalette:
		c := nearest(q.Palette, r, g, b)
		return c.R, c.G, c.B
	case QuantizeSingleColor:
		gv := mathx.QuantizeByte(mathx.LumaByte(r, g, b), q.Levels)
		return mathx.ScaleByte(q.SingleColor.R, gv), mathx.ScaleByte(q.SingleColor.G, gv), mathx.ScaleByte(q.SingleColor.B, gv)
	default:
		return r, g, b
	}
}

func (q Quantizer) QuantizeGrayFromRGB(r, g, b uint8) uint8 {
	qr, qg, qb := q.QuantizeColor(r, g, b)
	return mathx.LumaByte(qr, qg, qb)
}

func nearest(p Palette, r, g, b uint8) Color {
	best := p[0]
	bestDist := mathx.RGBDistanceSq(r, g, b, best.R, best.G, best.B)
	for i := 1; i < len(p); i++ {
		candidate := p[i]
		dist := mathx.RGBDistanceSq(r, g, b, candidate.R, candidate.G, candidate.B)
		if dist < bestDist {
			bestDist = dist
			best = candidate
		}
	}
	return best
}
