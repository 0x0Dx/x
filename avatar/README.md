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
- `-d, --debug` - Enable debug output

### Examples

```bash
# GitHub (requires personal access token with user scope)
avatar -p github -i avatar.png -t ghp_xxxxxxxxxxxx

# Discord (requires user token - get from browser devtools)
avatar -p discord -i avatar.png -t MzIxxxxxxxx

# Steam (requires cookies from browser)
avatar -p steam -i avatar.png -t "sessionid_value;steamLoginSecure_value"

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

## Getting Tokens

### GitHub

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Select scope: `user` (full control of user account)
4. Copy the token

### Discord

1. Open Discord in browser (not app)
2. Press F12 to open Developer Tools
3. Go to Network tab
4. Click any request to your user ID (like `/api/v9/users/@me`)
5. Look in "Request Headers" > "Authorization"
6. Copy the token (starts with `Mzi` or `ODY`)

### Steam

You need cookies from steamcommunity.com:

1. Open https://steamcommunity.com/ in browser (must be logged in)
2. Press F12 to open Developer Tools
3. Go to **Application** tab (Chrome) or **Storage** tab (Firefox)
4. Expand **Cookies** > **https://steamcommunity.com**
5. Copy these values:
   - `sessionid`
   - `steamLoginSecure` (contains your SteamID64 before `||`)

Format: `"sessionid_value;steamLoginSecure_value"`

Example:
```bash
avatar -p steam -i avatar.png -t "abc123def456;76561198000000000%7C%7Cabcdef123456..."
```

### Debug Mode

If something isn't working, add `-d` flag to see what's happening:

```bash
avatar -p steam -i avatar.png -t "sessionid;steamLoginSecure" -d
```
