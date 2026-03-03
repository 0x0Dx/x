# AGENTS.md - releasex

Simple Go release tool.

## Project Overview

- **Type**: CLI tool
- **Entry**: `main.go`
- **Commands**: `cmd/` directory
- **Uses**: spf13/cobra, gopkg.in/yaml.v3

## Testing

```bash
cd releasex
go test -v -run TestName ./...
go test -cover ./...
```

## Code Structure

```
releasex/
├── main.go           # Entry point
├── cmd/
│   ├── root.go       # Root command
│   ├── build.go      # Build binaries
│   ├── release.go    # Build + GitHub release
│   └── init.go       # Generate config
├── internal/
│   ├── config/       # YAML config parsing
│   ├── builder/      # Multi-platform build
│   ├── archiver/     # Tar/zip creation
│   └── checksums/    # SHA256 generation
└── go.mod
```

## Commands

- `releasex init` - Generate releasex.yaml
- `releasex build` - Build binaries
- `releasex release` - Build + create GitHub release

## Config (releasex.yaml)

```yaml
project: myapp
version: v1.0.0

builds:
  - id: main
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    main: ./cmd/app
    binary: myapp

archives:
  - id: default
    builds: [main]
    format: tar.gz

checksums:
  - ids: [default]

github:
  owner: myorg
  repo: myrepo
  version: tag
```

## Release

- Uses `gh` CLI if installed
- Falls back to API with `GITHUB_TOKEN` if `gh` not available
