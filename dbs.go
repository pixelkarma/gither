package gither

import "gither/internal/engine"

type DBSSeed = engine.DBSSeed
type DBSOptions = engine.DBSOptions

const (
	DBSSeedThreshold = engine.DBSSeedThreshold
	DBSSeedBayer     = engine.DBSSeedBayer
	DBSSeedFloyd     = engine.DBSSeedFloyd
)

func DirectBinarySearch(img *Image, opts DBSOptions) error {
	return engine.DirectBinarySearch(img, opts)
}
