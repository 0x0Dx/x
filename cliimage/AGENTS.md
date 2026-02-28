# AGENTS.md - cliimage

Terminal image viewer and converter.

## Project Overview

- **Type**: CLI tool
- **Entry**: `main.go`
- **Uses**: spf13/cobra

## Testing

```bash
cd cliimage
go test -v -run TestName ./...
go test -cover ./...
```

## Code Structure

```
cliimage/
├── main.go           # Entry point
├── internal/
│   ├── blocks/       # ASCII block symbols
│   ├── config/       # Configuration (cobra)
│   └── renderer/     # Image rendering logic
└── go.mod
```

## Key Functions

- `renderer.Render()` - Main rendering logic
- `blocks.Symbol*` - Block character constants
- Uses builder pattern: `renderer.New().Width(...).Height(...)`
