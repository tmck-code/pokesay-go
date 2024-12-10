package pokesay

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/mitchellh/go-wordwrap"
	"github.com/tmck-code/pokesay/src/pokedex"
)

type BoxChars struct {
	HorizontalEdge    string
	VerticalEdge      string
	TopRightCorner    string
	TopLeftCorner     string
	BottomRightCorner string
	BottomLeftCorner  string
	BalloonString     string
	BalloonTether     string
	Separator         string
	RightArrow        string
	CategorySeparator string
}

type Args struct {
	Width          int
	NoWrap         bool
	DrawBubble     bool
	TabSpaces      string
	NoTabSpaces    bool
	NoCategoryInfo bool
	ListCategories bool
	ListNames      bool
	Category       string
	NameToken      string
	JapaneseName   bool
	BoxChars       *BoxChars
	DrawInfoBorder bool
	Help           bool
	Verbose        bool
}

var (
	textStyleItalic *color.Color = color.New(color.Italic)
	textStyleBold   *color.Color = color.New(color.Bold)
	resetColourANSI string       = "\033[0m"
	AsciiBoxChars   *BoxChars    = &BoxChars{
		HorizontalEdge:    "-",
		VerticalEdge:      "|",
		TopRightCorner:    "\\",
		TopLeftCorner:     "/",
		BottomRightCorner: "/",
		BottomLeftCorner:  "\\",
		BalloonString:     "\\",
		BalloonTether:     "¡",
		Separator:         "|",
		RightArrow:        ">",
		CategorySeparator: "/",
	}
	UnicodeBoxChars *BoxChars = &BoxChars{
		HorizontalEdge:    "─",
		VerticalEdge:      "│",
		TopRightCorner:    "╮",
		TopLeftCorner:     "╭",
		BottomRightCorner: "╯",
		BottomLeftCorner:  "╰",
		BalloonString:     "╲",
		BalloonTether:     "╲",
		Separator:         "│",
		RightArrow:        "→",
		CategorySeparator: "/",
	}
	SingleWidthChars map[string]bool = map[string]bool{
		"♀": true,
		"♂": true,
	}
)

func DetermineBoxChars(unicodeBox bool) *BoxChars {
	if unicodeBox {
		return UnicodeBoxChars
	} else {
		return AsciiBoxChars
	}
}

// The main print function! This uses a chosen pokemon's index, names and categories, and an
// embedded filesystem of cowfile data
// 1. The text received from STDIN is printed inside a speech bubble
// 2. The cowfile data is retrieved using the matching index, decompressed (un-gzipped),
// 3. The pokemon is printed along with the name & category information
func Print(args Args, choice int, names []string, categories []string, cows embed.FS) {
	var b1 strings.Builder
	drawSpeechBubble(args.BoxChars, bufio.NewScanner(os.Stdin), args, &b1)
	var b2 strings.Builder
	drawPokemon(args, choice, names, categories, cows, &b2)

	s1, s2 := strings.Split(b1.String(), "\n"), strings.Split(b2.String(), "\n")
	// fmt.Println(b2.String(), len(s2))
	// fmt.Println(b1.String(), len(s1))
	// print side-by-side
	j := 0
	for i := 0; i < len(s1) || i < len(s2); i++ {
		if i < len(s1) {
			if i == len(s1)-4 {
				fmt.Print(s1[i] + "╱     ")
			} else if i == len(s1)-5 {
				fmt.Print(s1[i] + " ╱    ")
			} else if i == len(s1)-6 {
				fmt.Print(s1[i] + "  ╱   ")
			} else if i == len(s1)-7 {
				fmt.Print(s1[i] + "   ╱  ")
			} else if i == len(s1)-8 {
				fmt.Print(s1[i] + "    ╱ ")
			} else if i == len(s1)-9 {
				fmt.Print(s1[i] + "     ╱")
			} else {
				fmt.Print(s1[i] + "      ")
			}
		}
		if i >= len(s1)-len(s2) && j < len(s2) {
			fmt.Print("   " + s2[j])
			j++
		}
		fmt.Println()
	}
}

// Prints text from STDIN, surrounded by a speech bubble.
func drawSpeechBubble(boxChars *BoxChars, scanner *bufio.Scanner, args Args, b *strings.Builder) {
	if args.DrawBubble {
		// fmt.Fprintf(
		// 	b,
		// 	"%s%s%s\n",
		// 	boxChars.TopLeftCorner,
		// 	strings.Repeat(boxChars.HorizontalEdge, args.Width+2),
		// 	boxChars.TopRightCorner,
		// )
		b.WriteString(
			boxChars.TopLeftCorner + strings.Repeat(boxChars.HorizontalEdge, args.Width+2) + boxChars.TopRightCorner + "\n",
		)
	}

	for scanner.Scan() {
		line := scanner.Text()

		if !args.NoTabSpaces {
			line = strings.Replace(line, "\t", args.TabSpaces, -1)
		}
		if args.NoWrap {
			drawSpeechBubbleLine(boxChars, line, args, b)
		} else {
			drawWrappedText(boxChars, line, args, b)
		}
	}

	// bottomBorder := strings.Repeat(boxChars.HorizontalEdge, 6) +
	// 	boxChars.BalloonTether +
	// 	strings.Repeat(boxChars.HorizontalEdge, args.Width+2-7)

	bottomBorder := strings.Repeat(boxChars.HorizontalEdge, args.Width+2)

	if args.DrawBubble {
		// fmt.Fprintf(b, "%s%s%s\n", boxChars.BottomLeftCorner, bottomBorder, boxChars.BottomRightCorner)
		b.WriteString(boxChars.BottomLeftCorner + bottomBorder + boxChars.BottomRightCorner + "\n")
	} else {
		// fmt.Fprintf(b, " %s \n", bottomBorder)
		b.WriteString(" " + bottomBorder + " \n")
	}
	// for i := 0; i < 4; i++ {
	// 	// fmt.Fprintf(b, "%s%s\n", strings.Repeat(" ", i+8), boxChars.BalloonString)
	// 	b.WriteString(strings.Repeat(" ", i+8) + boxChars.BalloonString + "\n")
	// }
}

// Prints a single speech bubble line
func drawSpeechBubbleLine(boxChars *BoxChars, line string, args Args, b *strings.Builder) {
	if !args.DrawBubble {
		fmt.Fprintln(b, line)
		return
	}

	lineLen := UnicodeStringLength(line)
	if lineLen <= args.Width {
		// print the line with padding, the most common case
		// fmt.Fprintf(
		// 	b,
		// 	"%s %s%s%s %s\n",
		// 	boxChars.VerticalEdge, // left-hand side of the bubble
		// 	line, resetColourANSI, // the text
		// 	strings.Repeat(" ", args.Width-lineLen), // padding
		// 	boxChars.VerticalEdge,                   // right-hand side of the bubble
		// )
		b.WriteString(
			boxChars.VerticalEdge + " " + line + resetColourANSI + strings.Repeat(" ", args.Width-lineLen) + " " + boxChars.VerticalEdge + "\n",
		)
	} else if lineLen > args.Width {
		// print the line without padding or right-hand side of the bubble if the line is too long
		// fmt.Fprintf(
		// 	b,
		// 	"%s %s%s\n",
		// 	boxChars.VerticalEdge, // left-hand side of the bubble
		// 	line, resetColourANSI, // the text
		// )
		b.WriteString(boxChars.VerticalEdge + " " + line + resetColourANSI + "\n")
	}
}

// Prints line of text across multiple lines, wrapping it so that it doesn't exceed the desired width.
func drawWrappedText(boxChars *BoxChars, line string, args Args, b *strings.Builder) {
	for _, wline := range strings.Split(wordwrap.WrapString(strings.Replace(line, "\t", args.TabSpaces, -1), uint(args.Width)), "\n") {
		drawSpeechBubbleLine(boxChars, wline, args, b)
	}
}

func nameLength(names []string) int {
	totalLen := 0

	for _, name := range names {
		for _, c := range name {
			// check if ascii or single-width unicode
			if (c < 128) || (SingleWidthChars[string(c)]) {
				totalLen++
			} else {
				totalLen += 2
			}
		}

	}
	return totalLen
}

// Returns the length of a string, taking into account Unicode characters and ANSI escape codes.
func UnicodeStringLength(s string) int {
	nRunes, totalLen, ansiCode := len(s), 0, false

	for i, r := range s {
		if i < nRunes-1 {
			// detect the beginning of an ANSI escape code
			// e.g. "\033[38;5;196m"
			//       ^^^ start    ^ end
			if s[i:i+2] == "\033[" {
				ansiCode = true
			}
		}
		if ansiCode {
			// detect the end of an ANSI escape code
			if r == 'm' {
				ansiCode = false
			}
		} else {
			if r < 128 {
				// if ascii, then use width of 1. this saves some time
				totalLen++
			} else {
				totalLen += runewidth.RuneWidth(r)
			}
		}
	}
	return totalLen
}

// Prints a pokemon with its name & category information.
func drawPokemon(args Args, index int, names []string, categoryKeys []string, GOBCowData embed.FS, b *strings.Builder) {
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
			args.BoxChars.RightArrow, strings.Join(namesFmt, fmt.Sprintf(" %s ", args.BoxChars.Separator)),
		)
	} else {
		infoLine = fmt.Sprintf(
			"%s %s %s %s",
			args.BoxChars.RightArrow,
			strings.Join(namesFmt, fmt.Sprintf(" %s ", args.BoxChars.Separator)),
			args.BoxChars.Separator,
			textStyleItalic.Sprint(strings.Join(categoryKeys, args.BoxChars.CategorySeparator)),
		)
		for _, category := range categoryKeys {
			width += len(category)
		}
		width += len(categoryKeys) - 1 + 1 + 2 // lol why did I do this
	}

	if args.DrawInfoBorder {
		topBorder := fmt.Sprintf(
			"%s%s%s",
			args.BoxChars.TopLeftCorner, strings.Repeat(args.BoxChars.HorizontalEdge, width-2), args.BoxChars.TopRightCorner,
		)
		// b.WriteString(
		// 	args.BoxChars.TopLeftCorner + strings.Repeat(args.BoxChars.HorizontalEdge, width-2) + args.BoxChars.TopRightCorner + "\n",
		// )
		bottomBorder := fmt.Sprintf(
			"%s%s%s",
			args.BoxChars.BottomLeftCorner, strings.Repeat(args.BoxChars.HorizontalEdge, width-2), args.BoxChars.BottomRightCorner,
		)
		// b.WriteString(
		// 	args.BoxChars.BottomLeftCorner + strings.Repeat(args.BoxChars.HorizontalEdge, width-2) + args.BoxChars.BottomRightCorner + "\n",
		// )
		infoLine = fmt.Sprintf(
			"%s\n%s %s %s\n%s\n",
			topBorder, args.BoxChars.VerticalEdge, infoLine, args.BoxChars.VerticalEdge, bottomBorder,
		)
	} else {
		infoLine = fmt.Sprintf("%s\n", infoLine)
	}
	// fmt.Fprintf(b, "%s%s", pokedex.Decompress(d), infoLine)
	b.WriteString(string(pokedex.Decompress(d)) + infoLine)
}
