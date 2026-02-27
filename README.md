# 0x0D's experimental packages 

Monorepo containing tools.

## Projects

- [`cliimage`](./cliimage) Terminal image viewer.
- [`goserv`](./goserv) HTTP file server.
- [`gitx`](./gitx) Simple git wrapper.
- [`mochii`](./mochii) Package manager inspired by early Nix.

## mochii

A simple, opinionated package manager inspired by early Nix.

### Lua Builder Scripts

mochii supports Lua-based builder scripts (`.lua`) in addition to shell scripts. Lua builders provide a more expressive and type-safe way to define package builds.

```lua
-- Example: hello.lua builder
-- Define environment variables
env("NAME", "world")
env("VERSION", "1.0.0")

-- Add source dependencies
source("https://example.com/hello.tar.gz")

-- Define the derivation
derive("hello", {"input1", "input2"})

-- Run build commands
run("make")
run("make", "install", "PREFIX=" .. mu.install)

-- Define output
output(mu.install .. "/bin/hello")
```

### Available Lua Functions

| Function | Description |
|----------|-------------|
| `derive(name, inputs)` | Define a derivation |
| `source(url)` | Add a source URL |
| `path(s)` | Convert to absolute path |
| `hash(s)` | Compute hash |
| `output(path)` | Define output path |
| `env(key, value)` | Set environment variable |
| `run(cmd, ...)` | Run a command |
| `fetchurl(url)` | Fetch a URL |
| `nar(path)` | Create NAR archive |
| `unnar(archive, dest)` | Extract NAR archive |

### Environment Variables

- `mu` - Table of environment variables
- `mu.install` - Installation directory
- `mu.sources` - Sources directory
- `ARGS` - Command line arguments
- `ENV` - Full environment

### Development

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
