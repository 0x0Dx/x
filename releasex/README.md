# releasex

A simple release tool for Go projects.

## Installation

```bash
go install github.com/0x0Dx/x/releasex@main
```

## Quick Start

```bash
releasex init           # Create releasex.yaml
releasex build          # Build binaries to ./dist
releasex release       # Build and create GitHub release
```

## Configuration (releasex.yaml)

```yaml
project: myapp
version: v1.0.0

builds:
  - id: main
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    main: ./cmd/app
    binary: myapp
    ldflags: "-s -w"

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
  draft: false
```

## Monorepo Support

Build multiple projects from a monorepo:

```yaml
project: mymonorepo
version: v1.0.0

builds:
  - id: cli
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    main: ./cli
    binary: mycli
  - id: server
    goos: [linux]
    goarch: [amd64]
    main: ./server
    binary: myserver
```

## Commands

| Command | Description |
|---------|-------------|
| `releasex init` | Generate releasex.yaml |
| `releasex build` | Build binaries to ./dist |
| `releasex release` | Build and create GitHub release |

## Environment

- `GITHUB_TOKEN` - GitHub token for releases (or use `--token` flag)
