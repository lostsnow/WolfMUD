// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
)

// ANSI escape sequences for setting colors. These sequences can be
// concatenated into strings or appended to slice directly. This is preferable
// to calling Colorize with embedded colour place holders due to the slower
// performance of Colorize.
const (
	ESC       = "\x1b"
	CSI       = ESC + "[" // Control Sequence Introducer
	Reset     = CSI + "0m"
	Bold      = CSI + "1m"
	Normal    = CSI + "22m"
	Black     = CSI + "30m"
	Red       = CSI + "31m"
	Green     = CSI + "32m"
	Yellow    = CSI + "33m"
	Blue      = CSI + "34m"
	Magenta   = CSI + "35m"
	Cyan      = CSI + "36m"
	White     = CSI + "37m"
	BGBlack   = CSI + "40m"
	BGRed     = CSI + "41m"
	BGGreen   = CSI + "42m"
	BGYellow  = CSI + "43m"
	BGBlue    = CSI + "44m"
	BGMagenta = CSI + "45m"
	BGCyan    = CSI + "46m"
	BGWhite   = CSI + "47m"

	// Setup brown as an alias for yellow
	Brown   = Yellow
	BGBrown = BGYellow

	// WolfMUD specific meta colors
	Good   = Green
	Info   = Yellow
	Bad    = Red
	Prompt = Magenta
)

// colorTable maps color place holders to color escape sequences. Colorize uses
// this map to to substitute color placeholders of the form [COLOR] with the
// matching ANSI escape sequence.
var colorTable = map[string]string{
	"[RESET]":     Reset,
	"[BOLD]":      Bold,
	"[NORMAL]":    Normal,
	"[BLACK]":     Black,
	"[RED]":       Red,
	"[GREEN]":     Green,
	"[BROWN]":     Brown,
	"[YELLOW]":    Yellow,
	"[BLUE]":      Blue,
	"[MAGENTA]":   Magenta,
	"[CYAN]":      Cyan,
	"[WHITE]":     White,
	"[BGBLACK]":   BGBlack,
	"[BGRED]":     BGRed,
	"[BGGREEN]":   BGGreen,
	"[BGBROWN]":   BGBrown,
	"[BGYELLOW]":  BGYellow,
	"[BGBLUE]":    BGBlue,
	"[BGMAGENTA]": BGMagenta,
	"[BGCYAN]":    BGCyan,
	"[BGWHITE]":   BGWhite,
}

// Colorize returns a []byte with color place holders replaced with their ANSI
// escape sequence equivalent. Color place holders have the format [COLOR]
// where COLOR represents the name of the color (uppercased) to be used. For
// example:
//
//	Colorize([]byte("[RED]Hello [GREEN]World![DEFAULT]"))
//
// Would return a []byte with [RED] and [GREEN] replaced with the ANSI escape
// sequences \x1b[31m and \x1b[32m respectively causing Hello to be displayed
// in red and World! to be displayed in green.
//
// The returned slice is always a copy even it the original contains no colors.
//
// Use of this function is discouraged due to relatively poor performance. It's
// main use is to render text from files, such as those loaded when the server
// is initially started. In code it is better to use the ANSI escape sequence
// constants directly.
func Colorize(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)

	if bytes.IndexByte(out, ']') == -1 {
		return out
	}

	p := 0
	for color, code := range colorTable {
		// Quick exit? Check for ']' as '[' also in replacement text
		if bytes.IndexByte(out, ']') == -1 {
			break
		}
		// Shortcut id we can't find an instance of the current color?
		if p = bytes.Index(out, []byte(color)); p == -1 {
			continue
		}
		// If no shortcut available we can still use the position of the check to
		// shorten the length of the slice we are doing replacements on
		out = append(out[:p], bytes.Replace(out[p:], []byte(color), []byte(code), -1)...)
	}
	return out
}
