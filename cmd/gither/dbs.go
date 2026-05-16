package main

import (
	"strings"

	"github.com/pixelkarma/gither"
)

func isDBSAlgorithm(name string) bool {
	switch name {
	case "dbs", "clustered-dbs", "multilevel-dbs", "color-dbs":
		return true
	default:
		return false
	}
}

func buildDBSOptions(cfg config, report *gither.DBSReport, sourceOpts gither.Options) gither.DBSOptions {
	opts := gither.DBSOptions{
		Seed:             parseDBSSeed(cfg.dbsSeed),
		Levels:           cfg.levels,
		Palette:          sourceOpts.Quantizer.Palette,
		Passes:           cfg.dbsPasses,
		Threshold:        uint8(clampInt(cfg.threshold, 0, 255)),
		MoveMode:         parseDBSMoveMode(cfg.dbsMove),
		Neighborhood:     cfg.dbsRadius,
		Metric:           parseDBSMetric(cfg.dbsMetric),
		ScanOrder:        parseDBSScanOrder(cfg.dbsScan),
		RadiusPolicy:     parseDBSRadiusPolicy(cfg.dbsRadiusMode),
		MaxNoImprove:     cfg.dbsNoImprove,
		Restarts:         cfg.dbsRestarts,
		RandomSeed:       cfg.seed,
		ClusterStrength:  float32(cfg.dbsClusterStrength),
		ClusterToneAware: cfg.dbsClusterToneAware,
		Report:           report,
	}
	applyDBSSchedulePreset(&opts, cfg.dbsSchedule)
	return opts
}

func applyDBSSchedulePreset(opts *gither.DBSOptions, schedule string) {
	switch strings.ToLower(strings.TrimSpace(schedule)) {
	case "preview":
		opts.Passes = 1
		opts.MaxNoImprove = 1
		opts.Restarts = 0
		opts.MoveMode = gither.DBSMoveFlip
		opts.Metric = gither.DBSMetricFast
		opts.ScanOrder = gither.DBSScanSerpentine
		opts.RadiusPolicy = gither.DBSRadiusFixed
	case "balanced":
		opts.Passes = 2
		opts.MaxNoImprove = 1
		opts.Restarts = 0
		opts.MoveMode = gither.DBSMoveHybrid
		opts.Metric = gither.DBSMetricBalanced
		opts.ScanOrder = gither.DBSScanSerpentine
		opts.RadiusPolicy = gither.DBSRadiusFixed
	case "hq":
		opts.Passes = 3
		opts.MaxNoImprove = 2
		opts.Restarts = 1
		opts.MoveMode = gither.DBSMoveHybrid
		opts.Metric = gither.DBSMetricPerceptual
		opts.ScanOrder = gither.DBSScanRandom
		opts.RadiusPolicy = gither.DBSRadiusExpand
	}
}

func parseDBSSeed(value string) gither.DBSSeed {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "cluster-dot", "cluster-dot-16x16", "cluster16":
		return gither.DBSSeedCluster16
	case "bayer":
		return gither.DBSSeedBayer
	case "floyd-steinberg", "floyd", "fs":
		return gither.DBSSeedFloyd
	default:
		return gither.DBSSeedThreshold
	}
}

func parseDBSMoveMode(value string) gither.DBSMoveMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "flip":
		return gither.DBSMoveFlip
	case "swap":
		return gither.DBSMoveSwap
	default:
		return gither.DBSMoveHybrid
	}
}

func parseDBSMetric(value string) gither.DBSMetric {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "fast":
		return gither.DBSMetricFast
	case "perceptual":
		return gither.DBSMetricPerceptual
	default:
		return gither.DBSMetricBalanced
	}
}

func parseDBSScanOrder(value string) gither.DBSScanOrder {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "serpentine":
		return gither.DBSScanSerpentine
	case "random":
		return gither.DBSScanRandom
	default:
		return gither.DBSScanRaster
	}
}

func parseDBSRadiusPolicy(value string) gither.DBSRadiusPolicy {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "expand":
		return gither.DBSRadiusExpand
	default:
		return gither.DBSRadiusFixed
	}
}
