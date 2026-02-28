# AGENTS.md - gitx

Opinionated git wrapper with shorter commands.

## Project Overview

- **Type**: CLI tool
- **Entry**: `main.go`
- **Commands**: `cmd/` directory

## Testing

```bash
cd gitx
go test -v -run TestName ./...
go test -cover ./...
```

## Code Structure

```
gitx/
├── main.go           # Entry point
├── cmd/
│   ├── root.go       # Root command
│   ├── branch.go
│   ├── clone.go
│   ├── commit.go
│   ├── diff.go
│   ├── reset.go
│   ├── status.go
│   └── sync.go
└── go.mod
```

## Commands

- `gitx branch` - Manage branches
- `gitx clone` - Clone repositories
- `gitx commit` - Create commits
- `gitx diff` - Show differences
- `gitx reset` - Reset HEAD
- `gitx status` - Show status
- `gitx sync` - Sync with remote
