package gither

import "gither/internal/engine"

func Threshold(img *Image, opts Options) error { return engine.Threshold(img, opts) }
func Random(img *Image, opts Options) error    { return engine.Random(img, opts) }
