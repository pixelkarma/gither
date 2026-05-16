package main

import (
	"errors"
	"fmt"

	"github.com/pixelkarma/gither"
)

func applyAlgorithm(img *gither.Image, cfg config, opts gither.Options, dbsReport *gither.DBSReport) error {
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
	case "dbs":
		return gither.DirectBinarySearch(img, buildDBSOptions(cfg, dbsReport, opts))
	case "clustered-dbs":
		return gither.ClusteredDBS(img, buildDBSOptions(cfg, dbsReport, opts))
	case "multilevel-dbs":
		return gither.MultiLevelDBS(img, buildDBSOptions(cfg, dbsReport, opts))
	case "color-dbs":
		if opts.Quantizer.Kind != gither.QuantizePalette {
			return errors.New("color-dbs requires -quantizer palette")
		}
		return gither.ColorDBS(img, buildDBSOptions(cfg, dbsReport, opts))
	default:
		return fmt.Errorf("unsupported algorithm %q", cfg.algorithm)
	}
}
