# AGENTS.md - lib

Common utilities for CLI tools in this monorepo.

## Package

- **Import**: `github.com/0x0Dx/x/lib`

## Functions

| Function | Description |
|----------|-------------|
| `lib.Run(cmd, args)` | Execute cobra command with error handling |
| `lib.MustGetEnv(key)` | Get env var or exit with error |
| `lib.ExitOnError(err)` | Print error and exit if not nil |

## Usage

```go
import "github.com/0x0Dx/x/lib"

func main() {
    if err := lib.Run(rootCmd, nil); err != nil {
        lib.ExitOnError(err)
    }
    
    token := lib.MustGetEnv("GITHUB_TOKEN")
}
```
