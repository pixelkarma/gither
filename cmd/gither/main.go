package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/pixelkarma/gither"
	"github.com/pixelkarma/gither/adapters/stdimage"
)

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "gither: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg config) error {
	startedAt := time.Now()
	if cfg.in == "" || cfg.out == "" {
		return errors.New("both -in and -out are required")
	}
	src, err := loadImage(cfg.in)
	if err != nil {
		return err
	}
	img, err := stdimage.FromImage(src)
	if err != nil {
		return err
	}
	opts, err := buildOptions(cfg, img)
	if err != nil {
		return err
	}
	var dbsReport *gither.DBSReport
	if isDBSAlgorithm(cfg.algorithm) {
		dbsReport = &gither.DBSReport{}
	}
	if err := applyAlgorithm(img, cfg, opts, dbsReport); err != nil {
		return err
	}
	processedAt := time.Now()
	if err := saveImage(cfg.out, stdimage.ToImage(img), cfg.jpegQuality); err != nil {
		return err
	}
	if cfg.verbose {
		printStats(cfg, img, dbsReport, startedAt, processedAt, time.Now())
	}
	return nil
}
