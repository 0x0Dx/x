package main

type Block struct {
	Char        rune
	Coverage    [4]bool
	CoverageMap string
}

type Symbol uint8

const (
	SymbolHalf Symbol = iota
	SymbolQuarter
	SymbolAll
)

var halfBlocks = []Block{
	{Char: '▀', Coverage: [4]bool{true, true, false, false}, CoverageMap: "██\n  "},
	{Char: '▄', Coverage: [4]bool{false, false, true, true}, CoverageMap: "  \n██"},
	{Char: ' ', Coverage: [4]bool{false, false, false, false}, CoverageMap: "  \n  "},
	{Char: '█', Coverage: [4]bool{true, true, true, true}, CoverageMap: "██\n██"},
}

var quarterBlocks = []Block{
	{Char: '▘', Coverage: [4]bool{true, false, false, false}, CoverageMap: "█ \n  "},
	{Char: '▝', Coverage: [4]bool{false, true, false, false}, CoverageMap: " █\n  "},
	{Char: '▖', Coverage: [4]bool{false, false, true, false}, CoverageMap: "  \n█ "},
	{Char: '▗', Coverage: [4]bool{false, false, false, true}, CoverageMap: "  \n █"},
	{Char: '▌', Coverage: [4]bool{true, false, true, false}, CoverageMap: "█ \n█ "},
	{Char: '▐', Coverage: [4]bool{false, true, false, true}, CoverageMap: " █\n █"},
	{Char: '▀', Coverage: [4]bool{true, true, false, false}, CoverageMap: "██\n  "},
	{Char: '▄', Coverage: [4]bool{false, false, true, true}, CoverageMap: "  \n██"},
}

var complexBlocks = []Block{
	{Char: '▙', Coverage: [4]bool{true, false, true, true}, CoverageMap: "█ \n██"},
	{Char: '▟', Coverage: [4]bool{false, true, true, true}, CoverageMap: " █\n██"},
	{Char: '▛', Coverage: [4]bool{true, true, true, false}, CoverageMap: "██\n█ "},
	{Char: '▜', Coverage: [4]bool{true, true, false, true}, CoverageMap: "██\n █"},
	{Char: '▚', Coverage: [4]bool{true, false, false, true}, CoverageMap: "█ \n █"},
	{Char: '▞', Coverage: [4]bool{false, true, true, false}, CoverageMap: " █\n█ "},
}

func getAvailableBlocks(symbol Symbol) []Block {
	blocks := halfBlocks

	if symbol == SymbolQuarter || symbol == SymbolAll {
		blocks = append(blocks, quarterBlocks...)
	}

	if symbol == SymbolAll {
		blocks = append(blocks, complexBlocks...)
	}

	return blocks
}
