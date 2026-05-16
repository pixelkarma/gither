package gither

import "gither/internal/engine"

type DBSSeed = engine.DBSSeed
type DBSMoveMode = engine.DBSMoveMode
type DBSMetric = engine.DBSMetric
type DBSScanOrder = engine.DBSScanOrder
type DBSRadiusPolicy = engine.DBSRadiusPolicy
type DBSOptions = engine.DBSOptions
type DBSReport = engine.DBSReport

const (
	DBSSeedThreshold = engine.DBSSeedThreshold
	DBSSeedBayer     = engine.DBSSeedBayer
	DBSSeedFloyd     = engine.DBSSeedFloyd

	DBSMoveFlip   = engine.DBSMoveFlip
	DBSMoveSwap   = engine.DBSMoveSwap
	DBSMoveHybrid = engine.DBSMoveHybrid

	DBSMetricFast       = engine.DBSMetricFast
	DBSMetricBalanced   = engine.DBSMetricBalanced
	DBSMetricPerceptual = engine.DBSMetricPerceptual

	DBSScanRaster     = engine.DBSScanRaster
	DBSScanSerpentine = engine.DBSScanSerpentine
	DBSScanRandom     = engine.DBSScanRandom

	DBSRadiusFixed  = engine.DBSRadiusFixed
	DBSRadiusExpand = engine.DBSRadiusExpand
)

func DirectBinarySearch(img *Image, opts DBSOptions) error {
	return engine.DirectBinarySearch(img, opts)
}
