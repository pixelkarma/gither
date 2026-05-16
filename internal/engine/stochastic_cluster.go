package engine

import "github.com/pixelkarma/gither/internal/core"

func StochasticClusteredDot(img *core.Image, opts core.Options, ordered OrderedMap) error {
	return ApplyOrdered(img, ordered, opts)
}
