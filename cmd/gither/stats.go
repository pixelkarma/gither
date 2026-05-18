package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pixelkarma/gither"
)

type stageTimings struct {
	startedAt   time.Time
	loadedAt    time.Time
	convertedAt time.Time
	preparedAt  time.Time
	processedAt time.Time
	finishedAt  time.Time
}

func printStats(cfg config, img *gither.Image, dbsReport *gither.DBSReport, timings stageTimings) {
	loadElapsed := timings.loadedAt.Sub(timings.startedAt)
	convertElapsed := timings.convertedAt.Sub(timings.loadedAt)
	prepElapsed := timings.preparedAt.Sub(timings.convertedAt)
	ditherElapsed := timings.processedAt.Sub(timings.preparedAt)
	saveElapsed := timings.finishedAt.Sub(timings.processedAt)
	processElapsed := timings.processedAt.Sub(timings.startedAt)
	totalElapsed := timings.finishedAt.Sub(timings.startedAt)
	if isDBSAlgorithm(cfg.algorithm) && dbsReport != nil {
		dbsOpts := buildDBSOptions(cfg, nil, gither.Options{})
		fmt.Fprintf(os.Stderr,
			"stats algorithm=%s quantizer=%s size=%dx%d load_ms=%.3f convert_ms=%.3f prep_ms=%.3f dither_ms=%.3f process_ms=%.3f save_ms=%.3f total_ms=%.3f dbs_schedule=%s dbs_metric=%s dbs_scan=%s dbs_cluster_strength=%.3f dbs_passes_run=%d dbs_moves=%d dbs_flips=%d dbs_swaps=%d dbs_restarts=%d output=%s\n",
			cfg.algorithm,
			describeQuantizer(cfg),
			img.Width,
			img.Height,
			float64(loadElapsed)/float64(time.Millisecond),
			float64(convertElapsed)/float64(time.Millisecond),
			float64(prepElapsed)/float64(time.Millisecond),
			float64(ditherElapsed)/float64(time.Millisecond),
			float64(processElapsed)/float64(time.Millisecond),
			float64(saveElapsed)/float64(time.Millisecond),
			float64(totalElapsed)/float64(time.Millisecond),
			strings.TrimSpace(cfg.dbsSchedule),
			string(dbsOpts.Metric),
			string(dbsOpts.ScanOrder),
			float64(dbsOpts.ClusterStrength),
			dbsReport.PassesRun,
			dbsReport.AcceptedMoves,
			dbsReport.FlipMoves,
			dbsReport.SwapMoves,
			dbsReport.RestartsUsed,
			cfg.out,
		)
		return
	}
	fmt.Fprintf(
		os.Stderr,
		"stats algorithm=%s quantizer=%s size=%dx%d load_ms=%.3f convert_ms=%.3f prep_ms=%.3f dither_ms=%.3f process_ms=%.3f save_ms=%.3f total_ms=%.3f output=%s\n",
		cfg.algorithm,
		describeQuantizer(cfg),
		img.Width,
		img.Height,
		float64(loadElapsed)/float64(time.Millisecond),
		float64(convertElapsed)/float64(time.Millisecond),
		float64(prepElapsed)/float64(time.Millisecond),
		float64(ditherElapsed)/float64(time.Millisecond),
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
		return "palette(explicit)"
	case "single-color":
		return fmt.Sprintf("single-color(levels=%d,color=%s)", cfg.levels, strings.TrimSpace(cfg.singleColor))
	default:
		return cfg.quantizer
	}
}
