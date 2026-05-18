package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pixelkarma/gither"
)

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(bufio.NewReaderSize(file, 1<<20))
	return img, err
}

func saveImage(path string, img image.Image, jpegQuality int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriterSize(file, 1<<20)
	var encodeErr error
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		encodeErr = jpeg.Encode(writer, img, &jpeg.Options{Quality: jpegQuality})
	default:
		encodeErr = png.Encode(writer, img)
	}
	if encodeErr != nil {
		return encodeErr
	}
	return writer.Flush()
}

func loadOrderedMap(path string, width, height int, strength float32) (gither.OrderedMap, error) {
	if path == "" {
		return gither.OrderedMap{}, errors.New("custom-ordered requires -map")
	}
	if width <= 0 || height <= 0 {
		return gither.OrderedMap{}, errors.New("custom-ordered requires positive -map-width and -map-height")
	}
	file, err := os.Open(path)
	if err != nil {
		return gither.OrderedMap{}, err
	}
	defer file.Close()

	values := make([]uint8, 0, width*height)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == ';'
		})
		for _, field := range fields {
			if field == "" {
				continue
			}
			n, err := strconv.Atoi(field)
			if err != nil {
				return gither.OrderedMap{}, fmt.Errorf("invalid map value %q", field)
			}
			if n < 0 || n > 255 {
				return gither.OrderedMap{}, fmt.Errorf("map value %d out of range", n)
			}
			values = append(values, uint8(n))
		}
	}
	if err := scanner.Err(); err != nil {
		return gither.OrderedMap{}, err
	}
	return gither.NewOrderedMapFromU8(values, width, height, strength)
}
