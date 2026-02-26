# Avatar

A CLI tool to change your avatar on various services.

## Installation

```bash
go install github.com/0x0Dx/x/avatar@latest
```

## Usage

```bash
avatar -p <provider> -i <image> -t <token>
```

### Options

- `-p, --provider` - Service provider: `github`, `discord`, `steam`
- `-i, --image` - Path to image file or URL
- `-t, --token` - API token (optional, can use `AVATAR_TOKEN` env var)

### Examples

```bash
# GitHub (requires personal access token with user:email scope)
avatar -p github -i avatar.png -t ghp_xxxxxxxxxxxx

# Discord (requires user token)
avatar -p discord -i avatar.png -t MzIxxxxxxxx

# Steam (requires session cookies)
avatar -p steam -i avatar.png -t "sessionid;steamLoginSecure"

# From URL
avatar -p github -i "https://example.com/avatar.png" -t $GITHUB_TOKEN

# Using environment variable
export AVATAR_TOKEN=ghp_xxxxxxxxxxxx
avatar -p github -i avatar.png
```

## Providers

- **GitHub**: Uses GitHub REST API (requires `user` scope)
- **Discord**: Uses Discord API with user token
- **Steam**: Uses Steam web upload (requires session cookies from browser)

## Steam Cookies

To get Steam cookies:
1. Open Steam in browser
2. Open Developer Tools (F12)
3. Go to Application > Cookies > steamcommunity.com
4. Copy `sessionid` and `steamLoginSecure` values
