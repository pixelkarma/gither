# Release Notes

`gither` is currently versioned as an alpha series below `v1.0.0`.

## Release Intent

- the CLI is a first-class product surface
- the Go library is also a first-class product surface
- API and CLI naming may still change before `v1`

## Install Paths

CLI:

```bash
go install github.com/pixelkarma/gither/cmd/gither@latest
```

Library:

```bash
go get github.com/pixelkarma/gither
```

## Release Automation

Tagged releases are built with GoReleaser and published to GitHub Releases.

Suggested tag format:

- `v0.1.0`
- `v0.2.0`
- `v0.2.1`

## CI Baseline

The current release pipeline checks:

- `go test ./...`
- `go build -o ./dist/gither ./cmd/gither`
- shell validation for the example scripts

That keeps the maintenance burden low while still protecting the public package and binary.
