// Package gither provides image dithering algorithms for both library and CLI use.
//
// The public API is intentionally flat: callers choose a quantizer, construct or
// load an Image, and then apply one of the named algorithm functions.
//
// Stable alpha coverage includes Bayer, cluster-dot, Yliluoma, classic diffusion,
// threshold/random, Riemersma, and the DBS family. More stylized halftone and
// variable-diffusion variants remain available but are documented as experimental.
package gither
