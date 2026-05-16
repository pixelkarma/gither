package engine

import (
	"sync"

	"gither/internal/core"
	"gither/internal/mathx"
)

var historyWeights = [16]float32{1, 1, 2, 2, 3, 4, 5, 6, 8, 10, 12, 14, 16, 19, 23, 27}

type hilbertPoint struct {
	packed uint32
}

var hilbertPathCache sync.Map

func Riemersma(img *core.Image, opts core.Options) error {
	if err := img.Validate(); err != nil {
		return err
	}
	if err := opts.Quantizer.Validate(); err != nil {
		return err
	}
	width, height := img.Width, img.Height
	path := cachedHilbertPath(width, height)
	var history [16][3]float32
	head, filled := 0, 0
	var weighted [3]float32
	channels := img.ChannelCount()
	for _, pt := range path {
		x, y := pt.unpack()
		offset := y*img.Stride + x*channels
		weighted = weightedError(history, head, filled)
		switch img.Format {
		case core.Gray8:
			adjusted := mathx.ClampFloat32(mathx.ByteToUnit(img.Pix[offset])+weighted[0], 0, 1)
			gray := mathx.UnitToByte(adjusted)
			quantized := opts.Quantizer.QuantizeGrayFromRGB(gray, gray, gray)
			img.Pix[offset] = quantized
			pushError(&history, &head, &filled, [3]float32{adjusted - mathx.ByteToUnit(quantized), 0, 0})
		case core.RGB8:
			rAdj := mathx.ClampFloat32(mathx.ByteToUnit(img.Pix[offset])+weighted[0], 0, 1)
			gAdj := mathx.ClampFloat32(mathx.ByteToUnit(img.Pix[offset+1])+weighted[1], 0, 1)
			bAdj := mathx.ClampFloat32(mathx.ByteToUnit(img.Pix[offset+2])+weighted[2], 0, 1)
			qr, qg, qb := opts.Quantizer.QuantizeColor(mathx.UnitToByte(rAdj), mathx.UnitToByte(gAdj), mathx.UnitToByte(bAdj))
			img.Pix[offset], img.Pix[offset+1], img.Pix[offset+2] = qr, qg, qb
			pushError(&history, &head, &filled, [3]float32{rAdj - mathx.ByteToUnit(qr), gAdj - mathx.ByteToUnit(qg), bAdj - mathx.ByteToUnit(qb)})
		case core.RGBA8:
			alpha := img.Pix[offset+3]
			rAdj := mathx.ClampFloat32(mathx.ByteToUnit(img.Pix[offset])+weighted[0], 0, 1)
			gAdj := mathx.ClampFloat32(mathx.ByteToUnit(img.Pix[offset+1])+weighted[1], 0, 1)
			bAdj := mathx.ClampFloat32(mathx.ByteToUnit(img.Pix[offset+2])+weighted[2], 0, 1)
			qr, qg, qb := opts.Quantizer.QuantizeColor(mathx.UnitToByte(rAdj), mathx.UnitToByte(gAdj), mathx.UnitToByte(bAdj))
			img.Pix[offset], img.Pix[offset+1], img.Pix[offset+2], img.Pix[offset+3] = qr, qg, qb, alpha
			pushError(&history, &head, &filled, [3]float32{rAdj - mathx.ByteToUnit(qr), gAdj - mathx.ByteToUnit(qg), bAdj - mathx.ByteToUnit(qb)})
		}
	}
	return nil
}

func weightedError(history [16][3]float32, head, filled int) [3]float32 {
	if filled == 0 {
		return [3]float32{}
	}
	startWeight := len(historyWeights) - filled
	var accum [3]float32
	var sum float32
	for i := 0; i < filled; i++ {
		slot := i
		if filled == len(historyWeights) {
			slot = (head + i) % len(historyWeights)
		}
		weight := historyWeights[startWeight+i]
		sum += weight
		accum[0] += history[slot][0] * weight
		accum[1] += history[slot][1] * weight
		accum[2] += history[slot][2] * weight
	}
	return [3]float32{accum[0] / sum, accum[1] / sum, accum[2] / sum}
}

func pushError(history *[16][3]float32, head, filled *int, err [3]float32) {
	if *filled < len(history) {
		history[*filled] = err
		*filled = *filled + 1
		return
	}
	history[*head] = err
	*head = (*head + 1) % len(history)
}

func cachedHilbertPath(width, height int) []hilbertPoint {
	key := struct {
		width  int
		height int
	}{width: width, height: height}
	if cached, ok := hilbertPathCache.Load(key); ok {
		return cached.([]hilbertPoint)
	}
	side := mathx.NextPowerOfTwo(mathx.MaxInt(width, height))
	total := side * side
	path := make([]hilbertPoint, 0, width*height)
	for d := 0; d < total && len(path) < width*height; d++ {
		x, y := mathx.HilbertD2XY(side, d)
		if x >= width || y >= height {
			continue
		}
		path = append(path, hilbertPoint{packed: uint32(x)<<16 | uint32(y)})
	}
	actual, _ := hilbertPathCache.LoadOrStore(key, path)
	return actual.([]hilbertPoint)
}

func (p hilbertPoint) unpack() (int, int) {
	return int(p.packed >> 16), int(p.packed & 0xffff)
}
