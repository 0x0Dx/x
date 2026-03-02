# golicense

CLI tool for generating license files.

## Installation

```bash
go install github.com/0x0Dx/x/golicense@latest
```

## Usage

```bash
# Generate a license to stdout
golicense mit

# Generate a license and write to file
golicense -out mit

# Show all available licenses
golicense -show
```

### Options

- `-name` - Author name (defaults to git config `user.name`)
- `-email` - Author email (defaults to git config `user.email`)
- `-out` - Write to LICENSE file instead of stdout
- `-show` - List all available licenses

## Available Licenses

- **Public Domain**: CC0, Unlicense
- **Permissive**: MIT, Apache-2, BSD-0, BSD-2, BSD-3, ISC
- **GPL**: GPL-2, GPL-3, AGPL-3, LGPL-2
- **Copyleft-ish**: LGPL-2, MPL-2
- **Weird**: Be Gay Do Crimes, WTFPL, YOLO, BOLA, DWTFAWWI, Hot Potato, Curse of Knowledge
