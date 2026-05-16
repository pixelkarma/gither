package engine

import (
	"gither/internal/core"
	"gither/internal/kernels"
	"gither/internal/maps"
	"gither/internal/mathx"
)

type DBSSeed string

const (
	DBSSeedThreshold DBSSeed = "threshold"
	DBSSeedBayer     DBSSeed = "bayer"
	DBSSeedFloyd     DBSSeed = "floyd-steinberg"
)

type DBSOptions struct {
	Seed      DBSSeed
	Passes    int
	Threshold uint8
}

func (o DBSOptions) withDefaults() DBSOptions {
	if o.Seed == "" {
		o.Seed = DBSSeedThreshold
	}
	if o.Passes <= 0 {
		o.Passes = 1
	}
	return o
}

func DirectBinarySearch(img *core.Image, opts DBSOptions) error {
	if err := img.Validate(); err != nil {
		return err
	}
	opts = opts.withDefaults()
	target := grayscalePlane(img)
	targetFiltered := filteredGrayPlane(target, img.Width, img.Height)
	binary, err := dbsSeedPlane(target, img.Width, img.Height, opts)
	if err != nil {
		return err
	}
	for pass := 0; pass < opts.Passes; pass++ {
		improved := false
		for y := 0; y < img.Height; y++ {
			for x := 0; x < img.Width; x++ {
				idx := y*img.Width + x
				before := localFilteredError(binary, targetFiltered, img.Width, img.Height, x, y)
				original := binary[idx]
				if original == 0 {
					binary[idx] = 255
				} else {
					binary[idx] = 0
				}
				after := localFilteredError(binary, targetFiltered, img.Width, img.Height, x, y)
				if after+1e-9 < before {
					improved = true
				} else {
					binary[idx] = original
				}
			}
		}
		if !improved {
			break
		}
	}
	writeBinaryToImage(img, binary)
	return nil
}

func grayscalePlane(img *core.Image) []uint8 {
	out := make([]uint8, img.Width*img.Height)
	channels := img.ChannelCount()
	for y := 0; y < img.Height; y++ {
		row := img.Row(y)
		for x := 0; x < img.Width; x++ {
			offset := x * channels
			switch img.Format {
			case core.Gray8:
				out[y*img.Width+x] = row[offset]
			default:
				out[y*img.Width+x] = mathx.LumaByte(row[offset], row[offset+1], row[offset+2])
			}
		}
	}
	return out
}

func filteredGrayPlane(gray []uint8, width, height int) []float32 {
	out := make([]float32, len(gray))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out[y*width+x] = filteredBinaryValue(gray, width, height, x, y)
		}
	}
	return out
}

func dbsSeedPlane(target []uint8, width, height int, opts DBSOptions) ([]uint8, error) {
	seed := append([]uint8(nil), target...)
	img, err := core.NewPackedImage(seed, width, height, core.Gray8)
	if err != nil {
		return nil, err
	}
	baseOpts := core.Options{Quantizer: core.GrayLevels(2), Threshold: opts.Threshold}
	switch opts.Seed {
	case DBSSeedThreshold:
		if err := Threshold(img, baseOpts); err != nil {
			return nil, err
		}
	case DBSSeedBayer:
		ordered := OrderedMap{Values: maps.Bayer8x8, Width: 8, Height: 8, Strength: core.DefaultOrderedStrength}
		if err := ApplyOrdered(img, ordered, baseOpts); err != nil {
			return nil, err
		}
	case DBSSeedFloyd:
		if err := ApplyDiffusion(img, baseOpts, kernels.FloydSteinberg); err != nil {
			return nil, err
		}
	default:
		if err := Threshold(img, baseOpts); err != nil {
			return nil, err
		}
	}
	return img.Pix, nil
}

func localFilteredError(binary []uint8, targetFiltered []float32, width, height, x, y int) float64 {
	var score float64
	for yy := maxDBSInt(0, y-1); yy <= minDBSInt(height-1, y+1); yy++ {
		for xx := maxDBSInt(0, x-1); xx <= minDBSInt(width-1, x+1); xx++ {
			idx := yy*width + xx
			diff := float64(filteredBinaryValue(binary, width, height, xx, yy) - targetFiltered[idx])
			score += diff * diff
		}
	}
	return score
}

func filteredBinaryValue(binary []uint8, width, height, x, y int) float32 {
	var sum float32
	for ky := -1; ky <= 1; ky++ {
		yy := clampDBSInt(y+ky, 0, height-1)
		for kx := -1; kx <= 1; kx++ {
			xx := clampDBSInt(x+kx, 0, width-1)
			weight := dbsKernelWeight(kx, ky)
			sum += mathx.ByteToUnit(binary[yy*width+xx]) * weight
		}
	}
	return sum / 16.0
}

func dbsKernelWeight(kx, ky int) float32 {
	if kx == 0 && ky == 0 {
		return 4
	}
	if kx == 0 || ky == 0 {
		return 2
	}
	return 1
}

func writeBinaryToImage(img *core.Image, binary []uint8) {
	channels := img.ChannelCount()
	for y := 0; y < img.Height; y++ {
		row := img.Row(y)
		for x := 0; x < img.Width; x++ {
			v := binary[y*img.Width+x]
			offset := x * channels
			switch img.Format {
			case core.Gray8:
				row[offset] = v
			case core.RGB8:
				row[offset], row[offset+1], row[offset+2] = v, v, v
			case core.RGBA8:
				alpha := row[offset+3]
				row[offset], row[offset+1], row[offset+2], row[offset+3] = v, v, v, alpha
			}
		}
	}
}

func clampDBSInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func minDBSInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxDBSInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
