# AGENTS.md - Monorepo Development Guide

## Overview

This is a Go monorepo containing multiple CLI tools:
- **goreviewer** - AI-powered code reviewer
- **cliimage** - CLI image processing tool
- **gitx** - Opinionated git wrapper
- **goserv** - Simple HTTP server

## Build Commands

### Task Runner (Recommended)
```bash
task --list                    # List all available tasks
task check                    # Run fmt, lint, test
task build                    # Build all binaries to ./bin
task install                  # Install to $GOPATH/bin
task tidy                     # Run go mod tidy
task fmt                      # Run gofumpt
task lint                     # Run golangci-lint
task lint:fix                 # Run linters with auto-fix
task test                     # Run tests
```

### Direct Go Commands
```bash
# Build specific project
cd <project> && go build .

# Run tests
cd <project> && go test ./...

# Run single test
cd <project> && go test -v -run TestName ./...

# Run tests with coverage
cd <project> && go test -cover ./...

# Lint
cd <project> && golangci-lint run

# Format
cd <project> && gofumpt -w .

# Tidy
cd <project> && go mod tidy
```

### Install Development Tools
```bash
task install:tools
```
Installs: gofumpt, goimports, golangci-lint

## Code Style Guidelines

### Formatting
- Use **gofumpt** for formatting (stricter than gofmt)
- Run `task fmt` before committing
- 120 character line length

### Imports
- Use **goimports** for import management
- Group imports: stdlib → external → internal
- Use explicit import paths, no aliases unless needed

### Naming
- **Variables**: camelCase (e.g., `maxTokens`, `apiKey`)
- **Constants**: camelCase or SCREAMING_SNAKE for exported (e.g., `defaultModel`)
- **Functions**: camelCase, exported functions have doc comments
- **Packages**: lowercase, single word when possible (e.g., `cmd`, `internal/reviewer`)
- **Files**: single lowercase word or `<feature>.go`

### Types
- Use explicit types for public APIs
- Prefer interfaces for dependencies
- Use `time.Duration` for time values
- Use `context.Context` as first parameter for functions that make external calls

### Error Handling
- Use `fmt.Errorf` with `%w` for wrapped errors
- Return errors early, avoid nested error handling
- Add context to errors: `fmt.Errorf("failed to do X: %w", err)`
- Handle errors at call site, don't ignore with `_`

### Testing
- Test files: `<package>_test.go` in same package
- Test functions: `func TestName(t *testing.T)`
- Use table-driven tests when appropriate
- Run single test: `go test -v -run TestFunctionName ./...`

### Documentation
- Package-level doc: `// Package foo provides...`
- Exported function doc: `// FuncName does X.`
- Comments start with identifier name

### Linting
This project uses golangci-lint with these rules:
- revive, gofumpt, goimports (formatters)
- gosec (security), bodyclose, exhaustive
- godot, godox, goconst, goprintffuncname
- misspell, nakedret, nestif, nilerr
- nolintlint, prealloc, rowserrcheck
- sqlclosecheck, tparallel, unconvert
- unparam, whitespace, wrapcheck

### Git Conventions
- Branch naming: `feat/...`, `fix/...`, `docs/...`
- Commit messages: lowercase imperative ("add feature", not "added feature")
- PR titles: same style as commits

### Docker
- Multi-stage builds
- Distroless or alpine base images
- Build with CGO_ENABLED=0 for static binaries

## Project Structure

```
<project>/
├── main.go           # Entry point
├── cmd/              # CLI commands
├── internal/         # Private packages
│   └── <feature>/
├── *.go              # Root-level packages
├── go.mod
└── README.md
```

## Quick Reference

| Task | Command |
|------|---------|
| Build | `task build` |
| Test | `task test` |
| Test (single) | `go test -v -run TestX ./...` |
| Lint | `task lint` |
| Format | `task fmt` |
| Check all | `task check` |
