package gither

import "gither/internal/engine"

type DBSSeed = engine.DBSSeed
type DBSMoveMode = engine.DBSMoveMode
type DBSMetric = engine.DBSMetric
type DBSOptions = engine.DBSOptions

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
)

func DirectBinarySearch(img *Image, opts DBSOptions) error {
	return engine.DirectBinarySearch(img, opts)
}
