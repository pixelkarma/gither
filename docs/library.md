# Library Guide

`gither` exposes a flat top-level API so simple use stays simple:

- create or load an image
- choose a `Quantizer`
- call an algorithm function

## Core Types

- `Image`: the shared image container used by the library
- `Options`: common non-DBS options
- `Quantizer`: gray-level, RGB-level, palette, or single-color quantization
- `DBSOptions`: separate option surface for DBS family methods

Adapters for standard library images live in [adapters/stdimage](/Users/admin/Documents/dither/gither/adapters/stdimage:1).

## Typical Flow

```go
img, err := stdimage.LoadPath("input.png")
if err != nil {
	panic(err)
}

err = gither.Bayer8x8(img, gither.Options{
	Quantizer: gither.RGBLevels(4),
})
if err != nil {
	panic(err)
}

err = stdimage.SavePath("output.png", stdimage.ToImage(img), 95)
if err != nil {
	panic(err)
}
```

## Quantizers

- `gither.GrayLevels(n)`
- `gither.RGBLevels(n)`
- `gither.PaletteQuantizer(palette)`
- `gither.SingleColorQuantizer(levels, color)`

For automatic palette extraction:

- `gither.ExtractPalette(...)`
- `gither.ExtractPaletteWithOptions(...)`

## DBS

DBS functions use a separate API because the search, scheduling, and metric controls differ from the rest of the library:

- `gither.DirectBinarySearch`
- `gither.ClusteredDBS`
- `gither.MultiLevelDBS`
- `gither.ColorDBS`

If you only need a practical default, start with the CLI schedule presets and then mirror those settings in code later.
