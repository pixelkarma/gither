package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pixelkarma/gither"
)

func printStats(cfg config, img *gither.Image, dbsReport *gither.DBSReport, startedAt, processedAt, finishedAt time.Time) {
	processElapsed := processedAt.Sub(startedAt)
	saveElapsed := finishedAt.Sub(processedAt)
	totalElapsed := finishedAt.Sub(startedAt)
	if isDBSAlgorithm(cfg.algorithm) && dbsReport != nil {
		dbsOpts := buildDBSOptions(cfg, nil, gither.Options{})
		fmt.Fprintf(os.Stderr,
			"stats algorithm=%s quantizer=%s size=%dx%d process_ms=%.3f save_ms=%.3f total_ms=%.3f dbs_schedule=%s dbs_metric=%s dbs_scan=%s dbs_cluster_strength=%.3f dbs_passes_run=%d dbs_moves=%d dbs_flips=%d dbs_swaps=%d dbs_restarts=%d output=%s\n",
			cfg.algorithm,
			describeQuantizer(cfg),
			img.Width,
			img.Height,
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
	fmt.Fprintf(os.Stderr, "stats algorithm=%s quantizer=%s size=%dx%d process_ms=%.3f save_ms=%.3f total_ms=%.3f output=%s\n", cfg.algorithm, describeQuantizer(cfg), img.Width, img.Height, float64(processElapsed)/float64(time.Millisecond), float64(saveElapsed)/float64(time.Millisecond), float64(totalElapsed)/float64(time.Millisecond), cfg.out)
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
