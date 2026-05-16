package main

import (
	"flag"
	"fmt"
)

type config struct {
	in                  string
	out                 string
	algorithm           string
	quantizer           string
	levels              int
	palette             string
	paletteColors       int
	paletteMethod       string
	paletteSort         string
	singleColor         string
	strength            float64
	threshold           int
	seed                uint64
	randomStrength      int
	dbsSeed             string
	dbsPasses           int
	dbsMove             string
	dbsRadius           int
	dbsMetric           string
	dbsSchedule         string
	dbsScan             string
	dbsNoImprove        int
	dbsRestarts         int
	dbsRadiusMode       string
	dbsClusterStrength  float64
	dbsClusterToneAware bool
	mapPath             string
	mapWidth            int
	mapHeight           int
	jpegQuality         int
	verbose             bool
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
	flag.StringVar(&cfg.dbsSeed, "dbs-seed", "threshold", "DBS seed: threshold|bayer|floyd-steinberg")
	flag.IntVar(&cfg.dbsPasses, "dbs-passes", 1, "DBS optimization passes")
	flag.StringVar(&cfg.dbsMove, "dbs-move", "hybrid", "DBS move mode: flip|swap|hybrid")
	flag.IntVar(&cfg.dbsRadius, "dbs-radius", 1, "DBS swap neighborhood radius")
	flag.StringVar(&cfg.dbsMetric, "dbs-metric", "balanced", "DBS metric: fast|balanced|perceptual")
	flag.StringVar(&cfg.dbsSchedule, "dbs-schedule", "custom", "DBS schedule: custom|preview|balanced|hq")
	flag.StringVar(&cfg.dbsScan, "dbs-scan", "raster", "DBS scan order: raster|serpentine|random")
	flag.IntVar(&cfg.dbsNoImprove, "dbs-max-no-improve", 1, "DBS consecutive no-improvement passes before stop")
	flag.IntVar(&cfg.dbsRestarts, "dbs-restarts", 0, "DBS restart count")
	flag.StringVar(&cfg.dbsRadiusMode, "dbs-radius-policy", "fixed", "DBS radius policy: fixed|expand")
	flag.Float64Var(&cfg.dbsClusterStrength, "dbs-cluster-strength", 0.0, "DBS clustered boundary regularization strength")
	flag.BoolVar(&cfg.dbsClusterToneAware, "dbs-cluster-tone-aware", true, "DBS clustered mode uses stronger clustering in midtones")
	flag.StringVar(&cfg.mapPath, "map", "", "path to custom ordered map file")
	flag.IntVar(&cfg.mapWidth, "map-width", 0, "custom ordered map width")
	flag.IntVar(&cfg.mapHeight, "map-height", 0, "custom ordered map height")
	flag.IntVar(&cfg.jpegQuality, "jpeg-quality", 95, "JPEG quality when output extension is .jpg or .jpeg")
	flag.BoolVar(&cfg.verbose, "verbose", false, "print run stats after conversion")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: gither -in input.png -out output.png [options]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Algorithms:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  stable ordered: bayer-2x2, bayer-4x4, bayer-8x8, bayer-16x16, cluster-dot-4x4, cluster-dot-8x8, cluster-dot-16x16, custom-ordered\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  stable palette-ordered: yliluoma-1, yliluoma-2, yliluoma-3\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  stable diffusion: floyd-steinberg, false-floyd-steinberg, jjn, stucki, burkes, sierra, two-row-sierra, sierra-lite, stevenson-arce, atkinson\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  stable stochastic/path: threshold, random, riemersma\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  stable DBS: dbs, clustered-dbs, multilevel-dbs, color-dbs\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  experimental ordered/halftone: adaptive-bayer-8x8, adaptive-bayer-16x16, stochastic-cluster-dot-16x16, polyomino-16x16, space-filling-16x16, space-filling-morton-16x16, space-filling-serpentine-16x16, void-and-cluster-64x64, blue-noise-multitone-64x64, blue-noise-soft-64x64, blue-noise-hard-64x64, dot-diffusion-8x8, dot-diffusion-diagonal-8x8, am-fm-hybrid-64x64, clustered-am-fm-64x64\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  experimental variable diffusion: ostromoukhov, zhou-fang, balanced-variable, balanced-variable-thresholded, smooth-variable, punchy-variable\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	return cfg
}
