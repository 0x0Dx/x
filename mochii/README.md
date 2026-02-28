# mochii

A simple, opinionated package manager inspired by early Nix.

## Quick Start

```bash
# 1. Build mochii from source
git clone https://github.com/0x0Dx/x.git
cd x/mochii
go build -o mochii .

# 2. Initialize
export MOCHII_HOME=$HOME/.mochii
./mochii init

# 3. Register source code (or URL to source archive)
./mochii regfile /path/to/hello.tar.gz
# -> returns source hash, e.g.: abc123...

# 4. Install (builds from source)
./mochii getpkg abc123...
# -> runs builder script, installs to $MOCHII_HOME/pkg/<hash>

# 5. Switch to profile (point to bin directory, not the package root)
./mochii switch abc123... $MOCHII_HOME/pkg/abc123.../bin

# 6. Add to PATH
export PATH=$MOCHII_HOME/profiles/current/bin:$PATH
hello
```

## Configuration

Set `MOCHII_HOME` to change the default directory (default: `/var/mochii`).

```bash
export MOCHII_HOME=$HOME/.mochii
```

## Commands

| Command | Description |
|---------|-------------|
| `mochii init` | Initialize the database |
| `mochii verify` | Verify installed packages |
| `mochii getpkg HASH` | Build and install package |
| `mochii delpkg HASH` | Remove installed package |
| `mochii listinst` | List installed packages |
| `mochii run HASH [ARGS]` | Run package |
| `mochii regfile FILE` | Register file by hash |
| `mochii regurl HASH URL` | Register URL for hash |
| `mochii fetch URL` | Fetch URL and print hash |
| `mochii profile` | List profile generations |
| `mochii switch HASH PATH` | Switch to package in profile |
| `mochii gc` | Collect garbage |
| `mochii pull URL...` | Pull prebuilt packages |
| `mochii push DIR` | Push prebuilt packages |

## Package Format

### Source Package

A source package is a `.tar.gz` (or `.tar.bz2`, `.tar.xz`, `.zip`) containing source code:

```
hello-1.0/
├── src/
│   └── main.c
├── builder          # build script (REQUIRED)
└── ...
```

### Builder Script

The `builder` script is required. It builds from source and installs:

```bash
#!/bin/bash
set -e

# Unpack source (passed as argument or in current dir)
tar -xf ../src.tar.gz

cd hello-1.0
./configure --prefix=$MOCHII_PREFIX
make
make install
```

Environment variables:
- `MOCHII_PREFIX` - installation directory
- `MOCHII_HASH` - package hash

### Prebuilt Package

Prebuilts are binary substitutes that skip the build step. They are optional.

```
hello-1.0-linux-amd64.tar.gz
```

## Workflow

### 1. Register Source

```bash
# Register source archive (local file)
mochii regfile hello-1.0.tar.gz

# Or register URL to source
mochii regurl <hash> https://example.com/hello-1.0.tar.gz
```

### 2. Build from Source

```bash
mochii getpkg <source-hash>
# Downloads source, extracts, runs builder script
```

### 3. Use Prebuilt (Optional)

Register prebuilts to skip building:

```bash
# After building once, export as prebuilt
mochii push /path/to/prebuilts/

# Pull prebuilts on other machines
mochii pull https://example.com/prebuilts/
```

When a prebuilt exists for a package, `getpkg` will use it instead of building.

## Concepts

### Database Tables

- **refs** - Maps hash to file path (sources)
- **pkginst** - Installed packages (hash -> path)
- **prebuilts** - Binary substitutes (pkg-hash -> prebuilt-hash)
- **netsources** - Network URLs (hash -> URL)

### Build Process

1. Check if package already installed
2. Try prebuilt if available
3. Fetch source from `sources/` directory (or download if from URL)
4. Extract source to temp directory
5. Run builder script
6. Copy result to `pkg/<hash>/`
7. Register installed path

### Profile

A profile tracks generations of installed packages. Each `switch` creates a new generation with symlinks in `profiles/current/bin/`.

### Garbage Collection

`mochii gc` removes packages not referenced by any profile generation.

## Directory Structure

```
$MOCHII_HOME/
├── pkginfo.db       # SQLite database
├── sources/         # Downloaded sources
├── pkg/             # Installed packages
├── logs/            # Build logs
├── profiles/        # Profile generations
│   └── current/
│       └── bin/    # Symlinks to executables
└── prebuilts/      # Prebuilt packages
```

## Troubleshooting

**"permission denied" errors:**
Set `MOCHII_HOME` to a user-writable location:
```bash
export MOCHII_HOME=$HOME/.mochii
```

**Build fails:**
Check build logs: `ls $MOCHII_HOME/logs/`

**Package not in PATH:**
Ensure `$MOCHII_HOME/profiles/current/bin` is in your PATH:
```bash
export MOCHII_HOME=$HOME/.mochii
export PATH=$PATH:$MOCHII_HOME/profiles/current/bin
```

**"no garbage to collect" but disk is full:**
Check `$MOCHII_HOME/pkg/` for unused packages.

## Why mochii?

Mochii is a Go port of Nix v0.1.0 (circa 2003), before it became 500k+ lines of C++. It provides:

- **Source-based package management** - build from source
- **Content-addressed storage** - files identified by hash
- **Binary substitutes** - prebuilts for faster installation
- **Atomic upgrades and rollbacks** - switch between generations instantly
- **Garbage collection** - automatic cleanup of unused packages
- **Self-hosting** - can install itself

Without: sandboxing complexity, daemon setup, or the Nix expression language.

## Architecture

Mochii is organized into several internal packages:

| Package | Purpose |
|---------|---------|
| `builder` | Package building and installation |
| `db` | SQLite database operations |
| `eval` | Expression evaluation for derivations |
| `fetch` | Network source fetching |
| `hasher` | SHA-256 content hashing |
| `helper` | Utility functions |
| `nar` | NAR (Nix ARchive) format support |
| `prebuilts` | Binary substitute management |
| `profile` | Profile generation management |
| `store` | Nix store queries |
| `values` | Value storage and retrieval |

## Derivation Expressions

Mochii supports derivation expressions that describe how to build packages:

```nix
{
  name = "hello";
  buildPlatform = "x86_64-linux";
  builder = "/bin/bash";
  args = ["-c", "echo $src > $out"];
  env = {
    src = "abcdef...";  # reference to a value
  };
}
```

### Expression Types

- **Str("...")** - String literal
- **True / False** - Boolean values
- **Hash(...)** - Reference to a value by hash
- **External(...)** - Reference to non-expression value
- **Deref(...)** - Dereference an expression
- **(lambda x . e)** - Lambda abstraction
- **(e1 e2)** - Function application
- **(exec platform prog [args])** - Execution primitive

### Evaluating Expressions

```go
import "github.com/0x0Dx/x/mochii/internal/eval"

evaluator := eval.New(db, valuesDir, logDir, sourcesDir)
result, err := evaluator.EvalValue(expr)
```

## NAR Archive Format

NAR (Nix ARchive) is a serialization format for directory trees. Mochii uses it for:

- Hashing directories (content-addressed)
- Creating archives of built packages
- Extracting packages from archives

```go
import "github.com/0x0Dx/x/mochii/internal/nar"

// Hash a directory
hash, err := nar.Hash("/path/to/dir")

// Create a NAR archive
w := nar.NewWriter(file)
w.WriteDir("/path/to/dir")

// Extract a NAR archive
r := nar.NewReader(file)
r.Extract("/destination")
```

## Value Storage

Values are content-addressed files stored in the values directory:

```go
import "github.com/0x0Dx/x/mochii/internal/values"

mgr := values.New(db, valuesDir)

// Add a value
hash, err := mgr.AddValue("/path/to/file")

// Query a value path
path, err := mgr.QueryValuePath(hash)
```

## Store Queries

The store provides queries similar to nix-store:

```go
import "github.com/0x0Dx/x/mochii/internal/store"

s := store.New(db, sourcesDir, pkgDir)

// Query store path
path, err := s.QueryStore(hash)

// Verify store integrity
valid, err := s.VerifyStore(hash)

// Query graveyard (deleted paths)
paths, err := s.QueryGraveyard()
```

## Bootstrapping

To bootstrap mochii from source, you need to create a source package with a builder script.

### 1. Create Builder Script

```bash
cat > builder << 'EOF'
#!/bin/bash
set -e

echo "Building mochii..."

go build -o $MOCHII_PREFIX/bin/mochii .

echo "Installed to $MOCHII_PREFIX/bin/mochii"
EOF
chmod +x builder
```

### 2. Create Tarball

```bash
tar --exclude='mochii' -czf mochii-src.tar.gz .
```

### 3. Build and Install

```bash
export MOCHII_HOME=$HOME/.mochii
./mochii init
./mochii regfile mochii-src.tar.gz
# -> returns hash

./mochii getpkg <hash>

# Switch to profile (point to the bin directory, not the binary itself)
./mochii switch <hash> $MOCHII_HOME/pkg/<hash>/bin

# Add to PATH
export PATH=$PATH:$MOCHII_HOME/profiles/current/bin
```

### Directory Structure

Your source tarball should look like:

```
mochii-src/
├── builder          # build script (REQUIRED, must be executable)
├── main.go          # source files
├── go.mod
├── cmd/
└── internal/
```

The builder script runs from the extracted directory, with:
- `$MOCHII_PREFIX` - where to install (e.g., `/home/user/.mochii/pkg/<hash>/bin`)
- `$MOCHII_HASH` - the package hash
