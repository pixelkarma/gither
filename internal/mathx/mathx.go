package mathx

func ClampFloat32(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func ClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

var byteToUnitLUT = func() [256]float32 {
	var out [256]float32
	for i := 0; i < len(out); i++ {
		out[i] = float32(i) / 255.0
	}
	return out
}()

func ByteToUnit(v uint8) float32 {
	return byteToUnitLUT[v]
}

func UnitToByte(v float32) uint8 {
	return uint8(ClampFloat32(v, 0, 1)*255 + 0.5)
}

func LumaByte(r, g, b uint8) uint8 {
	return uint8((299*uint32(r) + 587*uint32(g) + 114*uint32(b) + 500) / 1000)
}

func RGBDistanceSq(r1, g1, b1, r2, g2, b2 uint8) uint32 {
	dr := int32(r1) - int32(r2)
	dg := int32(g1) - int32(g2)
	db := int32(b1) - int32(b2)
	return uint32(dr*dr + dg*dg + db*db)
}

func QuantizeByte(v uint8, levels int) uint8 {
	if levels <= 1 {
		return v
	}
	steps := levels - 1
	index := int(v)*steps + 127
	index /= 255
	return uint8((index*255 + steps/2) / steps)
}

func ScaleByte(channel, gray uint8) uint8 {
	return uint8((uint16(channel)*uint16(gray) + 127) / 255)
}

func Mix64(v uint64) uint64 {
	v ^= v >> 33
	v *= 0xff51afd7ed558ccd
	v ^= v >> 33
	v *= 0xc4ceb9fe1a85ec53
	v ^= v >> 33
	return v
}

func NextPowerOfTwo(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

func HilbertD2XY(n, d int) (int, int) {
	t := d
	x, y := 0, 0
	for s := 1; s < n; s *= 2 {
		rx := 1 & (t / 2)
		ry := 1 & (t ^ rx)
		if ry == 0 {
			if rx == 1 {
				x = s - 1 - x
				y = s - 1 - y
			}
			x, y = y, x
		}
		x += s * rx
		y += s * ry
		t /= 4
	}
	return x, y
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
