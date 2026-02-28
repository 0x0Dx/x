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

# 3. Create a declarative config
cat > config.json << 'EOF'
{
  "packages": {
    "hello": {
      "expression": {
        "name": "hello",
        "builder": "/bin/bash",
        "args": ["-c", "echo hello > $out"]
      }
    }
  }
}
EOF

# 4. Install packages from config
./mochii config config.json

# 5. Add to PATH
export PATH=$MOCHII_HOME/profiles/current/bin:$PATH
hello
```

## Declarative Package Management

Mochii uses a declarative configuration file to specify packages:

```json
{
  "packages": {
    "package-name": {
      "expression": {
        "name": "package",
        "builder": "/bin/bash",
        "args": ["-c", "echo $src > $out"],
        "env": {
          "src": "abc123..."
        }
      }
    },
    "vim": {
      "hash": "def456..."
    }
  }
}
```

## Commands

| Command | Description |
|---------|-------------|
| `mochii init` | Initialize the database |
| `mochii verify` | Verify installed packages |
| `mochii switch-config FILE` | Switch to declarative config |
| `mochii config FILE` | Install packages from config |
| `mochii profile` | List profile generations |
| `mochii gc` | Collect garbage |
| `mochii pull URL...` | Pull prebuilt packages |
| `mochii push DIR` | Push prebuilt packages |

## Configuration

Set `MOCHII_HOME` to change the default directory (default: `/var/mochii`).

```bash
export MOCHII_HOME=$HOME/.mochii
```

## Expression Types

- **Str("...")** - String literal
- **True / False** - Boolean values
- **Lam("x", e)** - Lambda abstraction (variable binding)
- **Var("x")** - Variable reference
- **(e1 e2)** - Function application
- **Hash(...)** - Reference to a value by hash
- **External(...)** - Reference to non-expression value
- **Deref(...)** - Dereference an expression
- **(exec platform prog [args])** - Execution primitive

## NAR Archive Format

NAR (Nix ARchive) is a serialization format for directory trees:

- **mochi-archive-1** - Version header
- Content-addressed directory hashing
- Supports regular files, directories, and symlinks

## Directory Structure

```
$MOCHII_HOME/
├── pkginfo.db       # SQLite database
├── config.json      # Current declarative config
├── pkg/             # Installed packages
├── profiles/        # Profile generations
│   └── current/
└── prebuilts/      # Prebuilt packages
```

## Why mochii?

Mochii is a Go port of Nix v0.1.0 (circa 2003), before it became 500k+ lines of C++. It provides:

- **Source-based package management** - build from source
- **Content-addressed storage** - files identified by hash
- **Binary substitutes** - prebuilts for faster installation
- **Atomic upgrades and rollbacks** - switch between generations instantly
- **Garbage collection** - automatic cleanup of unused packages
- **Declarative configurations** - define your system in JSON

Without: sandboxing complexity, daemon setup, or the full Nix expression language.
