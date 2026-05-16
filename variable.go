package gither

import "github.com/pixelkarma/gither/internal/engine"

// VariableAnchor defines one point on a variable-diffusion transfer curve.
type VariableAnchor = engine.VariableAnchor

// VariableCurve is a compiled variable-diffusion curve.
type VariableCurve = engine.VariableCurve

// NewVariableCurve constructs a variable-diffusion curve from anchors.
func NewVariableCurve(anchors []VariableAnchor) (VariableCurve, error) {
	return engine.NewVariableCurve(anchors)
}

// OstromoukhovCurve returns the built-in Ostromoukhov-style curve.
func OstromoukhovCurve() VariableCurve { return engine.OstromoukhovCurve() }

// BalancedCurve returns a balanced general-purpose variable-diffusion curve.
func BalancedCurve() VariableCurve { return engine.BalancedCurve() }

// SmoothCurve returns a lower-noise variable-diffusion curve.
func SmoothCurve() VariableCurve { return engine.SmoothCurve() }

// PunchyCurve returns a higher-contrast variable-diffusion curve.
func PunchyCurve() VariableCurve { return engine.PunchyCurve() }

// ApplyVariableGray applies a custom variable-diffusion curve to grayscale output.
func ApplyVariableGray(img *Image, opts Options, curve VariableCurve, thresholded bool) error {
	return engine.ApplyVariableGray(img, opts, curve, thresholded)
}

// Ostromoukhov applies the built-in Ostromoukhov variable-diffusion mode.
func Ostromoukhov(img *Image, opts Options) error { return engine.Ostromoukhov(img, opts) }

// ZhouFang applies the built-in Zhou-Fang variable-diffusion mode.
func ZhouFang(img *Image, opts Options) error { return engine.ZhouFang(img, opts) }

// BalancedVariable applies the balanced variable-diffusion mode.
func BalancedVariable(img *Image, opts Options) error { return engine.BalancedVariable(img, opts) }

// BalancedVariableThresholded applies the balanced variable-diffusion mode with thresholding.
func BalancedVariableThresholded(img *Image, opts Options) error {
	return engine.BalancedVariableThresholded(img, opts)
}

// SmoothVariable applies the smoother variable-diffusion mode.
func SmoothVariable(img *Image, opts Options) error { return engine.SmoothVariable(img, opts) }

// PunchyVariable applies the punchier variable-diffusion mode.
func PunchyVariable(img *Image, opts Options) error { return engine.PunchyVariable(img, opts) }
