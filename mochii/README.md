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

# 5. Switch to profile
./mochii switch abc123... $MOCHII_HOME/pkg/abc123...

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
3. Download/fetch source from refs/netsources
4. Extract source to build directory
5. Run builder script
6. Register installed path

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
export PATH=$MOCHII_HOME/profiles/current/bin:$PATH
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

# Switch to profile
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
