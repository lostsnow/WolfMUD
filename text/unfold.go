// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
	"unicode"
)

// Unfold unwraps long folded lines in the passed []byte keeping only
// significant whitespace. Line feeds that create blank lines or are before an
// indented line are kept. Trailing whitespace before a line feed is removed.
// For example:
//
//   The quick brown \n
//   fox jumps\n
//   \n
//     over the\n
//   lazy dog.
//
// Would be unfolded to:
//
//   The quick brown fox jumps
//
//     over the lazy dog.
//
// If a line starts with one or more CSIm ANSI escape sequences before any
// indenting whitespace, the escape sequences will be treated as zero width
// and ignored preserving the indent.
func Unfold(in []byte) []byte {
	data := bytes.Runes(in)
	out := make([]rune, len(in), len(in))
	pos := 0
	for x, r := range data {
		if r == '\n' {
			for ; pos > 0 && spaceNotLF(out[pos-1]); pos-- { // Trim WS not LF
			}
			if pos > 0 && out[pos-1] != '\n' && !spacePrefix(data[x+1:]) {
				r = ' '
			}
		}
		out[pos] = r
		pos++
	}
	return []byte(string(out[:pos]))
}

// spaceNotLF tests if the passed rune is Unicode whitespace but not a line
// feed. If the rune is Unicode whitespace and not a line feed true will be
// returned. If the rune is a line feed or non-whitespace then false will be
// returned.
func spaceNotLF(r rune) bool {
	if r == '\n' {
		return false
	}
	return unicode.IsSpace(r)
}

// spacePrefix tests to see if the passed []rune has a whitespace prefix. If
// the []rune has a whitespace prefix true is returned else false. Any leading
// CSIm ANSI escape sequences will be ignored for purposes of the test. For
// example, testing the string "\x1b[31m   abc" would return true.
func spacePrefix(in []rune) bool {
	if len(in) == 0 {
		return false
	}
	in = consumeEscape(in)
	if unicode.IsSpace(in[0]) {
		return true
	}
	return false
}

// consumeEscape will remove leading CSIm ANSI escape sequences from the passed
// []rune. The escape sequence may have the form "\x1b[m", "\x1b[nm" or
// "\x1b[n;nm" where n is one or more digits and "n;" may be repeated more than
// once. For example: "\x1b[m", "\x1b[31m", "\x1b[0;1;31;40m".
func consumeEscape(in []rune) []rune {
start:
	x := 0
	if x > len(in)-1 || in[x] != '\x1b' {
		return in
	}
	x++
	if x > len(in)-1 || in[x] != '[' {
		return in
	}
	x++
digit:
	if x > len(in)-1 {
		return in
	}
	if '0' <= in[x] && in[x] <= '9' {
		x++
		goto digit
	}
	if in[x] != ';' && in[x] != 'm' {
		return in
	}
	if in[x] == ';' {
		x++
		goto digit
	}
	if in[x] != 'm' {
		return in
	}
	x++
	in = in[x:]
	goto start
}
