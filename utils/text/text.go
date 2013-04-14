// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package text implements some text utilities. At the moment some of the
// utilities are text/TELNET related and may need to be split up more later on.
package text

import (
	"strings"
)

// ANSI Color escape sequences. The sequences are defined in the ECMA-48
// standard or ISO/IEC 6429.
//
// For high traffic constant messages like prompts having:
//
//	COLOR_MAGENTA + ">"
//
// Is a wee bit faster than:
//
//	"[MAGENTA]>"
//
// This is because we don't have to do the colorTable lookups. We also try to
// take a shortcut by checking if the character ']' is even in the format string
// of the message. If it isn't present we don't even attempt the colorTable
// lookups.
//
// TODO: Add more codes like background colors, underline, bold, normal ???
const (
	COLOR_BLACK   = "\033[30m"
	COLOR_RED     = "\033[31m"
	COLOR_GREEN   = "\033[32m"
	COLOR_YELLOW  = "\033[33m"
	COLOR_BLUE    = "\033[34m"
	COLOR_MAGENTA = "\033[35m"
	COLOR_CYAN    = "\033[36m"
	COLOR_WHITE   = "\033[37m"

	COLOR_BROWN = COLOR_YELLOW // Setup brown as an alias for yellow
)

// colorTable maps color names to ANSI escape sequences constants.
var colorTable = map[string]string{
	"[BLACK]":   COLOR_BLACK,
	"[RED]":     COLOR_RED,
	"[GREEN]":   COLOR_GREEN,
	"[BROWN]":   COLOR_BROWN,
	"[YELLOW]":  COLOR_YELLOW,
	"[BLUE]":    COLOR_BLUE,
	"[MAGENTA]": COLOR_MAGENTA,
	"[CYAN]":    COLOR_CYAN,
	"[WHITE]":   COLOR_WHITE,
}

const NOT_FOUND = -1

// BUG(Diddymus): Fold assumes control sequences are 5 bytes long. When we add
// more control sequences they probably won't be 5 bytes long. To fix this the
// two += 5 lines should be a table lookup.

// BUG(Diddymus): Fold assumes a fixed font. Some rune may display as 2 or 3
// cells wide and is dependant on the font used. A rune 2 cells wide in one
// font may only be 1 cell wide in another font. Due to this the folding may be
// inaccurate in some instances.

// Fold takes a string of text and turns it into lines of a certain length
// breaking on spaces. The text may contain ANSI color codes in the format
// \033[xxm - for values of xx see the definition of colorTable. Line endings
// are expected to be Linefeeds only - LF, \n or 0x0A - common on *nix systems.
//
// Leading spaces and original linefeeds should remain untouched.
func Fold(in string, width int) string {

	// Can we take a shortcut?
	if len(in) <= width {
		return in
	}

	lastNL := NOT_FOUND // Last new line position
	lastSP := NOT_FOUND // Last space position
	zeroNL := 0         // Zero width runes since last new line
	zeroSP := 0         // Zero width runes since last space
	cw := 0             // Currently calculated width

	output := []rune(in)

	for i, r := range output {

		switch r {

		// A space in the input text?
		case ' ':
			cw = i - lastNL - zeroNL
			if cw == width+1 {
				output[i] = '\n'
				lastNL, zeroNL = i, 0
			} else {
				if cw > width {
					if lastSP != NOT_FOUND {
						output[lastSP] = '\n'
					}
					lastNL, zeroNL = lastSP, zeroSP
				}
			}
			lastSP, zeroSP = i, 0

		// An original newline in the input text?
		case '\n':
			cw = i - lastNL - zeroNL
			if cw == width+1 {
				output[i] = '\n'
			} else {
				if cw > width {
					if lastSP != NOT_FOUND {
						output[lastSP] = '\n'
					}
				}
			}
			lastNL, zeroNL = i, 0
			lastSP, zeroSP = i, 0

		// Start of a control code?
		case '\033':
			zeroNL += 5
			zeroSP += 5

		}
	}

	// Process remaining runes when loop ends
	cw = len(output) - lastNL - zeroNL
	if cw != width+1 && cw > width {
		output[lastSP] = '\n'
	}

	return string(output)
}

// colorize turns color names into color ANSI codes within a string. This allows
// messages to be colored easily using the color names. For example the message:
//
//	"[RED]Boom![WHITE]"
//
// will be turned into:
//
//	"\033[31mBoom!\033[37m"
//
// Ultimately printing "Boom!" in red. Messages do not need to end in "[WHITE]"
// as this will be added automatically so you can't forget to do it. Colors can
// be changed as many times as you want:
//
//	"[RED]C[GREEN]o[YELLOW]l[BLUE]o[MAGENTA]u[CYAN]r"
//
// Prints "Color" each letter in a different color.
//
// TODO: Extend to include background colors?
func Colorize(in string) (out string) {
	if strings.Index(in, "]") != -1 {
		for color, code := range colorTable {
			in = strings.Replace(in, color, code, -1)
		}
	}
	return in
}

// monochrome strips color names from a string. This function is like colorize
// except the color name is replaced with the empty string instead of the raw
// ANSI escape code - in effect stripping the colors.
func Monochrome(in string) (out string) {
	if strings.Index(in, "]") != -1 {
		for color := range colorTable {
			in = strings.Replace(in, color, "", -1)
		}
	}
	return in
}
