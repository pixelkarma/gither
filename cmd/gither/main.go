package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gither"
	"gither/adapters/stdimage"
)

type config struct {
	in             string
	out            string
	algorithm      string
	quantizer      string
	levels         int
	palette        string
	paletteColors  int
	paletteMethod  string
	paletteSort    string
	singleColor    string
	strength       float64
	threshold      int
	seed           uint64
	randomStrength int
	mapPath        string
	mapWidth       int
	mapHeight      int
	jpegQuality    int
	verbose        bool
}

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "gither: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() config {
	cfg := config{}
	flag.StringVar(&cfg.in, "in", "", "input image path")
	flag.StringVar(&cfg.out, "out", "", "output image path")
	flag.StringVar(&cfg.algorithm, "algorithm", "floyd-steinberg", "algorithm name")
	flag.StringVar(&cfg.quantizer, "quantizer", "rgb-levels", "gray-levels|rgb-levels|palette|single-color")
	flag.IntVar(&cfg.levels, "levels", 4, "quantization levels for gray-levels, rgb-levels, or single-color")
	flag.StringVar(&cfg.palette, "palette", "", "palette colors as '#000000,#ffffff,#ff0000' or 'auto'")
	flag.IntVar(&cfg.paletteColors, "palette-colors", 8, "palette size when -palette=auto")
	flag.StringVar(&cfg.paletteMethod, "palette-method", "median-cut", "palette auto method: median-cut|popularity")
	flag.StringVar(&cfg.paletteSort, "palette-sort", "rgb", "palette auto sort: rgb|luma|frequency")
	flag.StringVar(&cfg.singleColor, "single-color", "#2f5d62", "single-color quantizer color as hex")
	flag.Float64Var(&cfg.strength, "strength", 64.0/255.0, "ordered dither strength")
	flag.IntVar(&cfg.threshold, "threshold", 127, "binary threshold in 0..255")
	flag.Uint64Var(&cfg.seed, "seed", 1, "random seed for random binary")
	flag.IntVar(&cfg.randomStrength, "random-strength", 32, "random threshold jitter in 0..127")
	flag.StringVar(&cfg.mapPath, "map", "", "path to custom ordered map file")
	flag.IntVar(&cfg.mapWidth, "map-width", 0, "custom ordered map width")
	flag.IntVar(&cfg.mapHeight, "map-height", 0, "custom ordered map height")
	flag.IntVar(&cfg.jpegQuality, "jpeg-quality", 95, "JPEG quality when output extension is .jpg or .jpeg")
	flag.BoolVar(&cfg.verbose, "verbose", false, "print run stats after conversion")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: gither -in input.png -out output.png [options]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Algorithms:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ordered: bayer-2x2, bayer-4x4, bayer-8x8, bayer-16x16, adaptive-bayer-8x8, adaptive-bayer-16x16, cluster-dot-4x4, cluster-dot-8x8, cluster-dot-16x16, stochastic-cluster-dot-16x16, polyomino-16x16, space-filling-16x16, space-filling-morton-16x16, space-filling-serpentine-16x16, void-and-cluster-64x64, blue-noise-multitone-64x64, blue-noise-soft-64x64, blue-noise-hard-64x64, dot-diffusion-8x8, dot-diffusion-diagonal-8x8, custom-ordered\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  palette-ordered: yliluoma-1, yliluoma-2, yliluoma-3\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  diffusion: floyd-steinberg, false-floyd-steinberg, jjn, stucki, burkes, sierra, two-row-sierra, sierra-lite, stevenson-arce, atkinson\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  variable diffusion: ostromoukhov, zhou-fang, balanced-variable, balanced-variable-thresholded, smooth-variable, punchy-variable\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  stochastic: threshold, random\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  advanced: riemersma, am-fm-hybrid-64x64, clustered-am-fm-64x64\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	return cfg
}

func run(cfg config) error {
	startedAt := time.Now()
	if cfg.in == "" || cfg.out == "" {
		return errors.New("both -in and -out are required")
	}
	src, err := loadImage(cfg.in)
	if err != nil {
		return err
	}
	img, err := stdimage.FromImage(src)
	if err != nil {
		return err
	}
	opts, err := buildOptions(cfg, img)
	if err != nil {
		return err
	}
	if err := applyAlgorithm(img, cfg, opts); err != nil {
		return err
	}
	processedAt := time.Now()
	if err := saveImage(cfg.out, stdimage.ToImage(img), cfg.jpegQuality); err != nil {
		return err
	}
	if cfg.verbose {
		printStats(cfg, img, startedAt, processedAt, time.Now())
	}
	return nil
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	return img, err
}

func saveImage(path string, img image.Image, jpegQuality int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: jpegQuality})
	default:
		return png.Encode(file, img)
	}
}

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

func applyAlgorithm(img *gither.Image, cfg config, opts gither.Options) error {
	switch cfg.algorithm {
	case "bayer-2x2":
		return gither.Bayer2x2(img, opts)
	case "bayer-4x4":
		return gither.Bayer4x4(img, opts)
	case "bayer-8x8":
		return gither.Bayer8x8(img, opts)
	case "bayer-16x16":
		return gither.Bayer16x16(img, opts)
	case "adaptive-bayer-8x8":
		return gither.AdaptiveBayer8x8(img, opts)
	case "adaptive-bayer-16x16":
		return gither.AdaptiveBayer16x16(img, opts)
	case "cluster-dot-4x4":
		return gither.ClusterDot4x4(img, opts)
	case "cluster-dot-8x8":
		return gither.ClusterDot8x8(img, opts)
	case "cluster-dot-16x16":
		return gither.ClusterDot16x16(img, opts)
	case "stochastic-cluster-dot-16x16":
		return gither.StochasticClusterDot16x16(img, opts)
	case "polyomino-16x16":
		return gither.Polyomino16x16(img, opts)
	case "space-filling-16x16":
		return gither.SpaceFilling16x16(img, opts)
	case "space-filling-morton-16x16":
		return gither.SpaceFillingMorton16x16(img, opts)
	case "space-filling-serpentine-16x16":
		return gither.SpaceFillingSerpentine16x16(img, opts)
	case "void-and-cluster-64x64":
		return gither.VoidAndCluster64x64(img, opts)
	case "blue-noise-multitone-64x64":
		return gither.BlueNoiseMultitone64x64(img, opts)
	case "blue-noise-soft-64x64":
		return gither.BlueNoiseSoft64x64(img, opts)
	case "blue-noise-hard-64x64":
		return gither.BlueNoiseHard64x64(img, opts)
	case "dot-diffusion-8x8":
		return gither.DotDiffusion8x8(img, opts)
	case "dot-diffusion-diagonal-8x8":
		return gither.DotDiffusionDiagonal8x8(img, opts)
	case "yliluoma-1":
		return gither.Yliluoma1(img, opts)
	case "yliluoma-2":
		return gither.Yliluoma2(img, opts)
	case "yliluoma-3":
		return gither.Yliluoma3(img, opts)
	case "custom-ordered":
		orderedMap, err := loadOrderedMap(cfg.mapPath, cfg.mapWidth, cfg.mapHeight, float32(cfg.strength))
		if err != nil {
			return err
		}
		return gither.CustomOrdered(img, orderedMap, opts)
	case "floyd-steinberg":
		return gither.FloydSteinberg(img, opts)
	case "false-floyd-steinberg":
		return gither.FalseFloydSteinberg(img, opts)
	case "jjn":
		return gither.JarvisJudiceNinke(img, opts)
	case "stucki":
		return gither.Stucki(img, opts)
	case "burkes":
		return gither.Burkes(img, opts)
	case "sierra":
		return gither.Sierra(img, opts)
	case "two-row-sierra":
		return gither.TwoRowSierra(img, opts)
	case "sierra-lite":
		return gither.SierraLite(img, opts)
	case "stevenson-arce":
		return gither.StevensonArce(img, opts)
	case "atkinson":
		return gither.Atkinson(img, opts)
	case "ostromoukhov":
		return gither.Ostromoukhov(img, opts)
	case "zhou-fang":
		return gither.ZhouFang(img, opts)
	case "balanced-variable":
		return gither.BalancedVariable(img, opts)
	case "balanced-variable-thresholded":
		return gither.BalancedVariableThresholded(img, opts)
	case "smooth-variable":
		return gither.SmoothVariable(img, opts)
	case "punchy-variable":
		return gither.PunchyVariable(img, opts)
	case "threshold":
		return gither.Threshold(img, opts)
	case "random":
		return gither.Random(img, opts)
	case "riemersma":
		return gither.Riemersma(img, opts)
	case "am-fm-hybrid-64x64":
		return gither.AMFMHybrid64x64(img, opts)
	case "clustered-am-fm-64x64":
		return gither.ClusteredAMFM64x64(img, opts)
	default:
		return fmt.Errorf("unsupported algorithm %q", cfg.algorithm)
	}
}

func loadOrderedMap(path string, width, height int, strength float32) (gither.OrderedMap, error) {
	if path == "" {
		return gither.OrderedMap{}, errors.New("custom-ordered requires -map")
	}
	if width <= 0 || height <= 0 {
		return gither.OrderedMap{}, errors.New("custom-ordered requires positive -map-width and -map-height")
	}
	file, err := os.Open(path)
	if err != nil {
		return gither.OrderedMap{}, err
	}
	defer file.Close()

	values := make([]uint8, 0, width*height)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == ';'
		})
		for _, field := range fields {
			if field == "" {
				continue
			}
			n, err := strconv.Atoi(field)
			if err != nil {
				return gither.OrderedMap{}, fmt.Errorf("invalid map value %q", field)
			}
			if n < 0 || n > 255 {
				return gither.OrderedMap{}, fmt.Errorf("map value %d out of range", n)
			}
			values = append(values, uint8(n))
		}
	}
	if err := scanner.Err(); err != nil {
		return gither.OrderedMap{}, err
	}
	return gither.NewOrderedMapFromU8(values, width, height, strength)
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

func printStats(cfg config, img *gither.Image, startedAt, processedAt, finishedAt time.Time) {
	processElapsed := processedAt.Sub(startedAt)
	saveElapsed := finishedAt.Sub(processedAt)
	totalElapsed := finishedAt.Sub(startedAt)
	fmt.Fprintf(os.Stderr,
		"stats algorithm=%s quantizer=%s size=%dx%d process_ms=%.3f save_ms=%.3f total_ms=%.3f output=%s\n",
		cfg.algorithm,
		describeQuantizer(cfg),
		img.Width,
		img.Height,
		float64(processElapsed)/float64(time.Millisecond),
		float64(saveElapsed)/float64(time.Millisecond),
		float64(totalElapsed)/float64(time.Millisecond),
		cfg.out,
	)
}

func describeQuantizer(cfg config) string {
	switch cfg.quantizer {
	case "gray-levels", "rgb-levels":
		return fmt.Sprintf("%s(levels=%d)", cfg.quantizer, cfg.levels)
	case "palette":
		if strings.EqualFold(strings.TrimSpace(cfg.palette), "auto") {
			return fmt.Sprintf("palette(auto:%d,%s,%s)", cfg.paletteColors, cfg.paletteMethod, cfg.paletteSort)
		}
		return fmt.Sprintf("palette(explicit)")
	case "single-color":
		return fmt.Sprintf("single-color(levels=%d,color=%s)", cfg.levels, strings.TrimSpace(cfg.singleColor))
	default:
		return cfg.quantizer
	}
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
