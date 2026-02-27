# 0x0D's experimental packages 

Monorepo containing tools.

## Projects

- [`cliimage`](./cliimage) Terminal image viewer.
- [`goserv`](./goserv) HTTP file server.
- [`gitx`](./gitx) Simple git wrapper.

## Installation 
 
```bash
go install github.com/0x0Dx/x/cliimage@main
go install github.com/0x0Dx/x/goserv@main
go install github.com/0x0Dx/x/gitx@main
```

## Development

Requires [Task](https://taskfile.dev) for running common tasks.

```bash
task install:tools  # Install dev tools (gofumpt, goimports, golangci-lint)
task build          # Build all binaries to ./bin
task install        # Install all binaries to $GOPATH/bin
task lint           # Run linters
task lint:fix       # Run linters and fix issues
task fmt            # Format code with gofumpt
task tidy           # Run go mod tidy
task test           # Run tests
```
