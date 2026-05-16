package gither

import "github.com/pixelkarma/gither/internal/engine"

type VariableAnchor = engine.VariableAnchor
type VariableCurve = engine.VariableCurve

func NewVariableCurve(anchors []VariableAnchor) (VariableCurve, error) {
	return engine.NewVariableCurve(anchors)
}

func OstromoukhovCurve() VariableCurve { return engine.OstromoukhovCurve() }
func BalancedCurve() VariableCurve     { return engine.BalancedCurve() }
func SmoothCurve() VariableCurve       { return engine.SmoothCurve() }
func PunchyCurve() VariableCurve       { return engine.PunchyCurve() }

func ApplyVariableGray(img *Image, opts Options, curve VariableCurve, thresholded bool) error {
	return engine.ApplyVariableGray(img, opts, curve, thresholded)
}

func Ostromoukhov(img *Image, opts Options) error     { return engine.Ostromoukhov(img, opts) }
func ZhouFang(img *Image, opts Options) error         { return engine.ZhouFang(img, opts) }
func BalancedVariable(img *Image, opts Options) error { return engine.BalancedVariable(img, opts) }
func BalancedVariableThresholded(img *Image, opts Options) error {
	return engine.BalancedVariableThresholded(img, opts)
}
func SmoothVariable(img *Image, opts Options) error { return engine.SmoothVariable(img, opts) }
func PunchyVariable(img *Image, opts Options) error { return engine.PunchyVariable(img, opts) }
