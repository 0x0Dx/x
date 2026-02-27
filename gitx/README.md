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
| `gitx br` | List branches |
| `gitx br -c <branch>` | Create a branch |
| `gitx br -s <branch>` | Switch to a branch |
| `gitx br -d <branch>` | Delete a branch |
| `gitx ci <msg>` | Stage all and commit |
| `gitx st` | Show status |
| `gitx push` | Push to remote |
| `gitx pull` | Pull from remote |
| `gitx sync` | Pull, rebase, push |

## Shorthands

- `ci` = commit (stage all + commit)
- `st` = status
- `br` = branch
