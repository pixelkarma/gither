package gither

import "github.com/pixelkarma/gither/internal/engine"

// DBSSeed selects the initialization strategy for DBS modes.
type DBSSeed = engine.DBSSeed

// DBSMoveMode selects which local move types DBS evaluates.
type DBSMoveMode = engine.DBSMoveMode

// DBSMetric selects the error metric used during DBS optimization.
type DBSMetric = engine.DBSMetric

// DBSScanOrder selects the iteration order used by DBS passes.
type DBSScanOrder = engine.DBSScanOrder

// DBSRadiusPolicy controls how DBS move neighborhoods expand.
type DBSRadiusPolicy = engine.DBSRadiusPolicy

// DBSOptions configures Direct Binary Search family algorithms.
type DBSOptions = engine.DBSOptions

// DBSReport captures pass and move statistics from a DBS run.
type DBSReport = engine.DBSReport

const (
	// DBSSeedThreshold starts from simple thresholded output.
	DBSSeedThreshold = engine.DBSSeedThreshold
	// DBSSeedBayer starts from Bayer-ordered output.
	DBSSeedBayer = engine.DBSSeedBayer
	// DBSSeedFloyd starts from Floyd-Steinberg output.
	DBSSeedFloyd = engine.DBSSeedFloyd
	// DBSSeedCluster16 starts from a 16x16 clustered-dot style output.
	DBSSeedCluster16 = engine.DBSSeedCluster16

	// DBSMoveFlip evaluates single-pixel flips.
	DBSMoveFlip = engine.DBSMoveFlip
	// DBSMoveSwap evaluates neighbor swaps.
	DBSMoveSwap = engine.DBSMoveSwap
	// DBSMoveHybrid evaluates both flips and swaps.
	DBSMoveHybrid = engine.DBSMoveHybrid

	// DBSMetricFast uses the fastest DBS metric.
	DBSMetricFast = engine.DBSMetricFast
	// DBSMetricBalanced uses the default DBS metric.
	DBSMetricBalanced = engine.DBSMetricBalanced
	// DBSMetricPerceptual uses the heaviest perceptual DBS metric.
	DBSMetricPerceptual = engine.DBSMetricPerceptual

	// DBSScanRaster scans rows in raster order.
	DBSScanRaster = engine.DBSScanRaster
	// DBSScanSerpentine scans rows in alternating directions.
	DBSScanSerpentine = engine.DBSScanSerpentine
	// DBSScanRandom scans in a seeded randomized order.
	DBSScanRandom = engine.DBSScanRandom

	// DBSRadiusFixed uses a fixed move neighborhood.
	DBSRadiusFixed = engine.DBSRadiusFixed
	// DBSRadiusExpand expands the move neighborhood during the run.
	DBSRadiusExpand = engine.DBSRadiusExpand
)

// DirectBinarySearch applies grayscale binary Direct Binary Search.
func DirectBinarySearch(img *Image, opts DBSOptions) error {
	return engine.DirectBinarySearch(img, opts)
}

// ClusteredDBS applies clustered Direct Binary Search with boundary regularization.
func ClusteredDBS(img *Image, opts DBSOptions) error {
	return engine.ClusteredDBS(img, opts)
}

// MultiLevelDBS applies multilevel grayscale Direct Binary Search.
func MultiLevelDBS(img *Image, opts DBSOptions) error {
	return engine.MultiLevelDBS(img, opts)
}

// ColorDBS applies palette-index color Direct Binary Search.
func ColorDBS(img *Image, opts DBSOptions) error {
	return engine.ColorDBS(img, opts)
}
