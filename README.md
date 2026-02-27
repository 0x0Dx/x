# 0x0D's experimental packages

Monorepo containing Go CLI tools.

## Projects

- [`cliimage`](./cliimage) Terminal image viewer.
- [`goserv`](./goserv) HTTP file server.

## Installation

```bash
go install github.com/0x0Dx/x/cliimage@main
go install github.com/0x0Dx/x/goserv@main
```

## Development

Requires [Task](https://taskfile.dev) for running common tasks.

```bash
task build    # Build all binaries
task lint     # Run linters
task tidy     # Run go mod tidy
```
