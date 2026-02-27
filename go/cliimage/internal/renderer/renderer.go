// Package renderer provides image rendering for terminal output.
package renderer

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"math"
	"strings"

	"github.com/0x0Dx/x/go/cliimage/internal/blocks"
	"github.com/charmbracelet/x/ansi"
	xdraw "golang.org/x/image/draw"
)

const (
	defaultThreshold = 128
	maxColorValue    = 255
)

// CliImage is the image renderer.
type CliImage struct {
	outputWidth    int
	outputHeight   int
	thresholdLevel uint8
	dither         bool
	noBlockSymbols bool
	invertColors   bool
	scale          int
	symbols        blocks.Symbol
}

// New creates a new CliImage renderer.
func New() CliImage {
	return CliImage{
		outputWidth:    0,
		outputHeight:   0,
		thresholdLevel: defaultThreshold,
		dither:         false,
		noBlockSymbols: false,
		invertColors:   false,
		scale:          1,
		symbols:        blocks.SymbolHalf,
	}
}

type pixelBlock struct {
	Pixels      [2][2]color.Color
	AvgFg       color.Color
	AvgBg       color.Color
	BestSymbol  rune
	BestFgColor color.Color
	BestBgColor color.Color
}

const u8MaxValue = 0xff

type shiftable interface {
	~uint | ~uint16 | ~uint32 | ~uint64
}

func shift[T shiftable](x T) T {
	if x > u8MaxValue {
		x >>= 8
	}
	return x
}

// Scale sets the scale factor for rendering.
func (r CliImage) Scale(scale int) CliImage {
	r.scale = scale
	return r
}

// IgnoreBlockSymbols sets whether to use only half blocks.
func (r CliImage) IgnoreBlockSymbols(fgOnly bool) CliImage {
	r.noBlockSymbols = fgOnly
	return r
}

// Dither sets whether to apply Floyd-Steinberg dithering.
func (r CliImage) Dither(dither bool) CliImage {
	r.dither = dither
	return r
}

// Threshold sets the luminance threshold level.
func (r CliImage) Threshold(threshold int) CliImage {
	if threshold >= 0 && threshold <= u8MaxValue {
		r.thresholdLevel = uint8(threshold)
	}
	return r
}

// InvertColors sets whether to invert colors.
func (r CliImage) InvertColors(invertColors bool) CliImage {
	r.invertColors = invertColors
	return r
}

// Width sets the output width in characters.
func (r CliImage) Width(width int) CliImage {
	r.outputWidth = width
	return r
}

// Height sets the output height in characters.
func (r CliImage) Height(height int) CliImage {
	r.outputHeight = height
	return r
}

// Symbol sets the block symbol mode.
func (r CliImage) Symbol(symbol blocks.Symbol) CliImage {
	r.symbols = symbol
	return r
}

// Render renders an image to an ANSI string.
func Render(img image.Image, width int, height int) string {
	r := New().Width(width).Height(height)
	return r.Render(img)
}

// Render renders the image to an ANSI-styled string.
func (r *CliImage) Render(img image.Image) string {
	bounds := img.Bounds()
	srcWidth := bounds.Max.X - bounds.Min.X
	srcHeight := bounds.Max.Y - bounds.Min.Y

	outWidth := r.outputWidth
	if outWidth <= 0 {
		outWidth = srcWidth
	}

	outHeight := r.outputHeight

	if outHeight <= 0 {
		const divider = 2
		outHeight = max(int(float64(outWidth)*float64(srcHeight)/float64(srcWidth)/divider), 1)
	}

	scaledImg := r.applyScaling(img, outWidth*r.scale, outHeight*r.scale)

	if r.dither {
		scaledImg = r.applyDithering(scaledImg)
	}

	if r.invertColors {
		scaledImg = r.invertImage(scaledImg)
	}

	var output strings.Builder

	imageBounds := scaledImg.Bounds()

	availableBlocks := blocks.GetAvailableBlocks(r.symbols)

	for y := 0; y < imageBounds.Max.Y; y += 2 {
		for x := 0; x < imageBounds.Max.X; x += 2 {
			block := r.createPixelBlock(scaledImg, x, y)

			r.findBestRepresentation(block, availableBlocks)

			output.WriteString(
				ansi.Style{}.ForegroundColor(block.BestFgColor).BackgroundColor(block.BestBgColor).Styled(string(block.BestSymbol)),
			)
		}
		output.WriteString("\n")
	}

	return output.String()
}

func (r *CliImage) createPixelBlock(img image.Image, x, y int) *pixelBlock {
	block := &pixelBlock{}

	for dy := range 2 {
		for dx := range 2 {
			block.Pixels[dy][dx] = r.getPixelSafe(img, x+dx, y+dy)
		}
	}

	return block
}

func (r *CliImage) findBestRepresentation(block *pixelBlock, availableBlocks []blocks.Block) {
	if r.noBlockSymbols {
		block.BestSymbol = '▀'
		block.BestBgColor = r.averageColors(block.Pixels[0][0], block.Pixels[0][1])
		block.BestFgColor = r.averageColors(block.Pixels[1][0], block.Pixels[1][1])
		return
	}

	pixelMask := [2][2]bool{}
	for y := range 2 {
		for x := range 2 {
			luma := rgbaToLuminance(block.Pixels[y][x])
			pixelMask[y][x] = luma >= r.thresholdLevel
		}
	}

	bestChar := ' '
	bestScore := math.MaxFloat64

	for _, blockChar := range availableBlocks {
		score := 0.0
		for i := range 4 {
			y, x := i/2, i%2
			if blockChar.Coverage[i] != pixelMask[y][x] {
				score += 1.0
			}
		}

		if score < bestScore {
			bestScore = score
			bestChar = blockChar.Char
		}
	}

	var fgPixels, bgPixels []color.Color

	var coverage [4]bool
	for _, b := range availableBlocks {
		if b.Char == bestChar {
			coverage = b.Coverage
			break
		}
	}

	for i := range 4 {
		y, x := i/2, i%2
		if coverage[i] {
			fgPixels = append(fgPixels, block.Pixels[y][x])
		} else {
			bgPixels = append(bgPixels, block.Pixels[y][x])
		}
	}

	if len(fgPixels) > 0 {
		block.BestFgColor = r.averageColors(fgPixels...)
	} else {
		block.BestFgColor = color.Black
	}

	if len(bgPixels) > 0 {
		block.BestBgColor = r.averageColors(bgPixels...)
	} else {
		block.BestBgColor = color.Black
	}

	block.BestSymbol = bestChar
}

func (r *CliImage) averageColors(colors ...color.Color) color.Color {
	if len(colors) == 0 {
		return color.Black
	}

	var sumR, sumG, sumB, sumA uint32

	for _, c := range colors {
		colR, colG, colB, colA := c.RGBA()
		colR, colG, colB, colA = shift(colR), shift(colG), shift(colB), shift(colA)
		sumR += colR
		sumG += colG
		sumB += colB
		sumA += colA
	}

	colorCount := len(colors)
	if colorCount <= 0 {
		return color.Black
	}
	//nolint:gosec // G115: false positive - colorCount is always small (0-4)
	count := uint32(colorCount)
	avgR := clampUint8(sumR / count)
	avgG := clampUint8(sumG / count)
	avgB := clampUint8(sumB / count)
	avgA := clampUint8(sumA / count)
	return color.RGBA{
		R: avgR,
		G: avgG,
		B: avgB,
		A: avgA,
	}
}

func clampUint8(v uint32) uint8 {
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func (CliImage) getPixelSafe(img image.Image, x, y int) color.RGBA {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return color.RGBA{0, 0, 0, 255}
	}

	r8, g8, b8, a8 := img.At(x, y).RGBA()
	return color.RGBA{
		R: clampUint8(r8 >> 8),
		G: clampUint8(g8 >> 8),
		B: clampUint8(b8 >> 8),
		A: clampUint8(a8 >> 8),
	}
}

func (r *CliImage) applyScaling(img image.Image, width, height int) image.Image {
	rect := image.Rect(0, 0, width, height)
	dst := image.NewRGBA(rect)
	xdraw.ApproxBiLinear.Scale(dst, rect, img, img.Bounds(), draw.Over, nil)
	return dst
}

func (r *CliImage) applyDithering(img image.Image) image.Image {
	b := img.Bounds()
	pm := image.NewPaletted(b, palette.Plan9)
	FloydSteinberg.Draw(pm, b, img, image.Point{})
	return pm
}

func (r *CliImage) invertImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	result := image.NewRGBA(bounds)
	for y := range height {
		for x := range width {
			r8, g8, b8, a8 := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			invR := clampUint8(maxColorValue - (r8 >> 8))
			invG := clampUint8(maxColorValue - (g8 >> 8))
			invB := clampUint8(maxColorValue - (b8 >> 8))
			alpha := clampUint8(a8 >> 8)
			result.Set(x, y, color.RGBA{
				R: invR,
				G: invG,
				B: invB,
				A: alpha,
			})
		}
	}

	return result
}

func rgbaToLuminance(c color.Color) uint8 {
	r, g, b, _ := c.RGBA()
	r, g, b = shift(r), shift(g), shift(b)
	return uint8(float64(r)*0.299 + float64(g)*0.587 + float64(b)*0.114)
}

// FloydSteinberg is the Floyd-Steinberg dithering algorithm.
var FloydSteinberg = draw.FloydSteinberg
