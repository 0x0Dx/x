# AGENTS.md - goserv

Simple HTTP server.

## Project Overview

- **Type**: CLI tool / HTTP server
- **Entry**: `main.go`

## Testing

```bash
cd goserv
go test -v -run TestName ./...
go test -cover ./...
```

## Code Structure

```
goserv/
├── main.go           # Entry point and server logic
└── go.mod
```

## Usage

```bash
goserv [port]
```

## Notes

- Simple single-file implementation
- Minimal dependencies
