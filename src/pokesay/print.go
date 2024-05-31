package pokesay

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/mitchellh/go-wordwrap"
	"github.com/tmck-code/pokesay/src/pokedex"
)

type BoxCharacters struct {
	HorizontalEdge    string
	VerticalEdge      string
	TopRightCorner    string
	TopLeftCorner     string
	BottomRightCorner string
	BottomLeftCorner  string
	BalloonString     string
	Separator         string
	RightArrow        string
	CategorySeparator string
}

type Args struct {
	Width          int
	NoWrap         bool
	TabSpaces      string
	NoTabSpaces    bool
	NoCategoryInfo bool
	ListCategories bool
	ListNames      bool
	ListNameToken  string
	Category       string
	NameToken      string
	IDToken        string
	JapaneseName   bool
	ShowID         bool
	BoxCharacters  *BoxCharacters
	DrawInfoBorder bool
	Help           bool
	Verbose        bool
}

var (
	textStyleItalic    *color.Color   = color.New(color.Italic)
	textStyleBold      *color.Color   = color.New(color.Bold)
	AsciiBoxCharacters *BoxCharacters = &BoxCharacters{
		HorizontalEdge:    "-",
		VerticalEdge:      "|",
		TopRightCorner:    "\\",
		TopLeftCorner:     "/",
		BottomRightCorner: "/",
		BottomLeftCorner:  "\\",
		BalloonString:     "\\",
		Separator:         "|",
		RightArrow:        ">",
		CategorySeparator: "/",
	}
	UnicodeBoxCharacters *BoxCharacters = &BoxCharacters{
		HorizontalEdge:    "─",
		VerticalEdge:      "│",
		TopRightCorner:    "╮",
		TopLeftCorner:     "╭",
		BottomRightCorner: "╯",
		BottomLeftCorner:  "╰",
		BalloonString:     "╲",
		Separator:         "│",
		RightArrow:        "→",
		CategorySeparator: "/",
	}
	SingleWidthCars map[string]bool = map[string]bool{
		"♀": true,
		"♂": true,
	}
)

func DetermineBoxCharacters(unicodeBox bool) *BoxCharacters {
	if unicodeBox {
		return UnicodeBoxCharacters
	} else {
		return AsciiBoxCharacters
	}
}

// The main print function! This uses a chosen pokemon's index, names and categories, and an
// embedded filesystem of cowfile data
// 1. The text received from STDIN is printed inside a speech bubble
// 2. The cowfile data is retrieved using the matching index, decompressed (un-gzipped),
// 3. The pokemon is printed along with the name & category information
func Print(args Args, choice int, names []string, categories []string, cows embed.FS) {
	printSpeechBubble(args.BoxCharacters, bufio.NewScanner(os.Stdin), args.Width, args.NoTabSpaces, args.TabSpaces, args.NoWrap)
	printPokemon(args, choice, names, categories, cows)
}

// Prints text from STDIN, surrounded by a speech bubble.
func printSpeechBubble(boxCharacters *BoxCharacters, scanner *bufio.Scanner, width int, noTabSpaces bool, tabSpaces string, noWrap bool) {
	border := strings.Repeat(boxCharacters.HorizontalEdge, width+2)
	fmt.Println(boxCharacters.TopLeftCorner + border + boxCharacters.TopRightCorner)

	for scanner.Scan() {
		line := scanner.Text()

		if !noTabSpaces {
			line = strings.Replace(line, "\t", tabSpaces, -1)
		}
		if noWrap {
			printSpeechBubbleLine(boxCharacters, line, width)
		} else {
			printWrappedText(boxCharacters, line, width, tabSpaces)
		}
	}
	fmt.Println(boxCharacters.BottomLeftCorner + border + boxCharacters.BottomRightCorner)
	for i := 0; i < 4; i++ {
		fmt.Println(strings.Repeat(" ", i+8), boxCharacters.BalloonString)
	}
}

// Prints a single speech bubble line
func printSpeechBubbleLine(boxCharacters *BoxCharacters, line string, width int) {
	if len(line) > width {
		fmt.Printf("%s %s\n", boxCharacters.VerticalEdge, line)
	} else if len(line) == width {
		fmt.Printf("%s %s %s\n", boxCharacters.VerticalEdge, line, boxCharacters.VerticalEdge)
	} else {
		fmt.Printf(
			"%s %s%s %s\n",
			boxCharacters.VerticalEdge, line, strings.Repeat(" ", width-len(line)), boxCharacters.VerticalEdge,
		)
	}
}

// Prints line of text across multiple lines, wrapping it so that it doesn't exceed the desired width.
func printWrappedText(boxCharacters *BoxCharacters, line string, width int, tabSpaces string) {
	for _, wline := range strings.Split(wordwrap.WrapString(strings.Replace(line, "\t", tabSpaces, -1), uint(width)), "\n") {
		printSpeechBubbleLine(boxCharacters, wline, width)
	}
}

func nameLength(names []string) int {
	totalLen := 0

	for _, name := range names {
		for _, c := range name {
			// check if ascii or single-width unicode
			if (c < 128) || (SingleWidthCars[string(c)]) {
				totalLen++
			} else {
				totalLen += 2
			}
		}
	}
	return totalLen
}

// Prints a pokemon with its name & category information.
func printPokemon(args Args, index int, names []string, categoryKeys []string, GOBCowData embed.FS) {
	d, _ := GOBCowData.ReadFile(pokedex.EntryFpath("build/assets/cows", index))

	width := nameLength(names)
	namesFmt := make([]string, 0)
	for _, name := range names {
		namesFmt = append(namesFmt, textStyleBold.Sprint(name))
	}
	// count name separators
	width += (len(names) - 1) * 3
	width += 2     // for the arrow
	width += 2 + 2 // for the end box characters

	infoLine := ""

	if args.NoCategoryInfo {
		infoLine = fmt.Sprintf(
			"%s %s",
			args.BoxCharacters.RightArrow,
			strings.Join(namesFmt, fmt.Sprintf(" %s ", args.BoxCharacters.Separator)),
		)
	} else {
		infoLine = fmt.Sprintf(
			"%s %s %s %s",
			args.BoxCharacters.RightArrow,
			strings.Join(namesFmt, fmt.Sprintf(" %s ", args.BoxCharacters.Separator)),
			args.BoxCharacters.Separator,
			textStyleItalic.Sprint(strings.Join(categoryKeys, args.BoxCharacters.CategorySeparator)),
		)
		for _, category := range categoryKeys {
			width += len(category)
		}
		width += len(categoryKeys) - 1 + 1 + 2 // lol why did I do this
	}

	if args.DrawInfoBorder {
		topBorder := fmt.Sprintf(
			"%s%s%s",
			args.BoxCharacters.TopLeftCorner,
			strings.Repeat(args.BoxCharacters.HorizontalEdge, width-2),
			args.BoxCharacters.TopRightCorner,
		)
		bottomBorder := fmt.Sprintf(
			"%s%s%s",
			args.BoxCharacters.BottomLeftCorner,
			strings.Repeat(args.BoxCharacters.HorizontalEdge, width-2),
			args.BoxCharacters.BottomRightCorner,
		)
		infoLine = fmt.Sprintf(
			"%s\n%s %s %s\n%s\n",
			topBorder,
			args.BoxCharacters.VerticalEdge,
			infoLine,
			args.BoxCharacters.VerticalEdge,
			bottomBorder,
		)
	} else {
		infoLine = fmt.Sprintf("%s\n", infoLine)
	}
	fmt.Printf("%s%s", pokedex.Decompress(d), infoLine)
}
