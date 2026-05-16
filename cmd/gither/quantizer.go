package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/pixelkarma/gither"
)

func buildOptions(cfg config, img *gither.Image) (gither.Options, error) {
	quantizer, err := buildQuantizer(cfg, img)
	if err != nil {
		return gither.Options{}, err
	}
	return gither.Options{
		Quantizer:      quantizer,
		Strength:       float32(cfg.strength),
		Threshold:      uint8(clampInt(cfg.threshold, 0, 255)),
		Seed:           cfg.seed,
		RandomStrength: uint8(clampInt(cfg.randomStrength, 0, 127)),
	}, nil
}

func buildQuantizer(cfg config, img *gither.Image) (gither.Quantizer, error) {
	switch cfg.quantizer {
	case "gray-levels":
		return gither.GrayLevels(cfg.levels), nil
	case "rgb-levels":
		return gither.RGBLevels(cfg.levels), nil
	case "palette":
		if strings.EqualFold(strings.TrimSpace(cfg.palette), "auto") {
			palette, err := gither.ExtractPaletteWithOptions(img, gither.PaletteExtractOptions{
				Colors: cfg.paletteColors,
				Method: parsePaletteMethod(cfg.paletteMethod),
				Sort:   parsePaletteSort(cfg.paletteSort),
			})
			if err != nil {
				return gither.Quantizer{}, err
			}
			return gither.PaletteQuantizer(palette), nil
		}
		palette, err := parsePalette(cfg.palette)
		if err != nil {
			return gither.Quantizer{}, err
		}
		return gither.PaletteQuantizer(palette), nil
	case "single-color":
		color, err := parseHexColor(cfg.singleColor)
		if err != nil {
			return gither.Quantizer{}, err
		}
		return gither.SingleColorQuantizer(cfg.levels, color), nil
	default:
		return gither.Quantizer{}, fmt.Errorf("unsupported quantizer %q", cfg.quantizer)
	}
}

func parsePalette(value string) (gither.Palette, error) {
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("palette quantizer requires -palette")
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t'
	})
	palette := make(gither.Palette, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		color, err := parseHexColor(part)
		if err != nil {
			return nil, err
		}
		palette = append(palette, color)
	}
	return palette, nil
}

func parseHexColor(value string) (gither.Color, error) {
	normalized := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(normalized) != 6 {
		return gither.Color{}, fmt.Errorf("invalid hex color %q", value)
	}
	r, err := strconv.ParseUint(normalized[0:2], 16, 8)
	if err != nil {
		return gither.Color{}, fmt.Errorf("invalid hex color %q", value)
	}
	g, err := strconv.ParseUint(normalized[2:4], 16, 8)
	if err != nil {
		return gither.Color{}, fmt.Errorf("invalid hex color %q", value)
	}
	b, err := strconv.ParseUint(normalized[4:6], 16, 8)
	if err != nil {
		return gither.Color{}, fmt.Errorf("invalid hex color %q", value)
	}
	return gither.Color{R: uint8(r), G: uint8(g), B: uint8(b)}, nil
}

func parsePaletteMethod(value string) gither.PaletteExtractMethod {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "popularity":
		return gither.PaletteMethodPopularity
	default:
		return gither.PaletteMethodMedianCut
	}
}

func parsePaletteSort(value string) gither.PaletteSortMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "luma":
		return gither.PaletteSortLuma
	case "frequency":
		return gither.PaletteSortFrequency
	default:
		return gither.PaletteSortRGB
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
