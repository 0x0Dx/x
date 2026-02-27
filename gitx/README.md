# gitx

A simple, opinionated git wrapper.

## Installation

```bash
go install github.com/0x0Dx/x/gitx@main
```

## Commands

| Command | Description |
|---------|-------------|
| `gitx clone <repo>` | Clone a repository |
| `gitx co <branch>` | Checkout a branch |
| `gitx co -b <branch>` | Create and checkout new branch |
| `gitx ci <msg>` | Stage all and commit |
| `gitx st` | Show status |
| `gitx br` | List branches |
| `gitx br -d <branch>` | Delete a branch |
| `gitx push` | Push to remote |
| `gitx pull` | Pull from remote |
| `gitx sync` | Pull, rebase, push |

## Shorthands

- `co` = checkout
- `ci` = commit (stage all + commit)
- `st` = status
- `br` = branch
