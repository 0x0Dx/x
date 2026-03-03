# cliimage

A terminal image viewer that renders images using Unicode block characters with ANSI colors.

## Installation

```bash
go install github.com/0x0Dx/x/cliimage@main
```

## Usage

```bash
cliimage render -i image.png
```

## Options

| Flag | Description |
|------|-------------|
| `-i, --input` | Input image file (required) |
| `-o, --output` | Output file (default: stdout) |
| `-w, --width` | Output width in characters |
| `-h, --height` | Output height in characters |
| `-t, --threshold` | Luminance threshold (0-255, default: 128) |
| `-p, --pixelation` | Pixelation mode: `half`, `quarter`, `all` |
| `-d, --dither` | Apply Floyd-Steinberg dithering |
| `-b, --noblock` | Use only half blocks |
| `-r, --invert` | Invert colors |
| `--scale` | Scale factor (default: 1) |

## Examples

```bash
# Render image with default settings
cliimage render -i photo.jpg

# Render with custom width
cliimage render -i photo.jpg -w 100

# Use quarter blocks for higher resolution
cliimage render -i photo.jpg -p quarter

# Apply dithering
cliimage render -i photo.jpg -d

# Save to file
cliimage render -i photo.jpg -o output.txt
```
