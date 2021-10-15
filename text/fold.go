// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"math"
	"unicode"
	"unicode/utf8"
)

// Fold takes a []byte and attempts to reformat it so lines have a maximum
// length of the passed width. Fold will only split lines on whitespace when
// reformatting. This may result in lines longer than the given width when a
// word is too long and cannot be split. A width of zero or less indicates no
// folding will be done but line endings will still be changed from Unix '\n'
// to network '\r\n' line endings and trailing whitespace, except line feeds
// '\n', will be removed.
//
// Fold will handle multibyte runes. However it cannot handle 'wide' runes -
// those that are wider than a normal single character when displayed. This is
// because the required information is actually contained in the font files of
// the font in use at the 'client' end.
//
// For example the Chinese for 9 is 九 (U+4E5D). Even in a monospaced font 九
// will take up the space of two columns.
//
// For combining characters Fold will assume combining marks are zero width.
// For example 'a' plus a combining grave accent U+0061 U+0300 will be counted
// as a single character. However it is better to use an actual latin small
// letter a with grave 'à' U+00E0. Either should work as expected.
//
// It is expected that the end of line markers for incoming data are Unix line
// feeds (LF, '\n') and outgoing data will have network line endings, carriage
// return + line feed pairs (CR+LF, '\r\n').
func Fold(in []byte, width int) []byte {

	var (
		// output buffer, twice the input length is the pathalogical case of \n
		// expanding to \r\n for every character in the input buffer.
		o = make([]byte, len(in)*2)

		ip int // Input position
		is int // Input position of last space
		op int // Output position
		os int // Output position of last space
		vc int // Perceived visual count of characters in current word

		pre = true  // Preserve white-space mode
		esc = false // Processing ANSI escape sequence
	)

	// If no wrapping (width < 1) go as wide as possible
	if width < 1 {
		width = math.MaxInt
	}

	for ip = 0; ip < len(in); ip++ {

		switch {

		// Start of CSI "ESC[" ANSI escape sequence
		case !esc && in[ip] == 0x1b && ip < len(in)-1 && in[ip+1] == '[':
			o[op] = 0x1b
			op++
			o[op] = '['
			op++
			ip++
			esc = true

		// Parameter and intermediate ANSI escape sequence bytes
		case esc && ' ' <= in[ip] && in[ip] <= '?':
			o[op] = in[ip]
			op++

		// Terminating ANSI escape sequence byte
		case esc && '@' <= in[ip] && in[ip] <= '~':
			o[op] = in[ip]
			op++
			esc = false

		// Complete UTF-8 multibyte
		case in[ip]&0b11000000 == 0b10000000:
			o[op] = in[ip]
			op++

		// Width exceeded on a space - so break on space
		case vc == width && in[ip] == ' ':
			// Skip trailing white-space
			for ; ip < len(in) && in[ip] == ' '; ip++ {
			}
			ip--

			if ip < len(in)-1 {
				o[op] = '\r'
				op++
				o[op] = '\n'
				op++
				pre = true
			}
			vc, os, is = 0, 0, 0

		// Width exceeded on a non-space - need to break on previous space
		case vc > width && os != 0:
			// Skip trailing white-space
			for ; ip < len(in) && in[ip] == ' '; ip++ {
			}
			ip--

			if ip < len(in)-1 {
				o[os] = '\r'
				os++
				o[os] = '\n'
				os++
				pre = true
			}
			op, ip = os, is
			vc, os, is = 0, 0, 0

		// Substitute '\n' in input with "\r\n" in output
		case in[ip] == '\n':
			o[op] = '\r'
			op++
			o[op] = '\n'
			op++
			vc, os, is = 0, 0, 0
			pre = true

		// Drop/remove consecutive white-space if not in preserve mode
		case in[ip] == ' ' && !pre && op > 0 && o[op-1] == ' ':

		case in[ip] == ' ':
			o[op] = ' '
			if !pre {
				os, is = op, ip
			}
			op++
			vc++

		// Replace preserving hard-space '␠' U+2420, UTF8 bytes: 0xe2 0x90 0xa0
		case in[ip] == 0xe2 && ip < len(in)-2 && in[ip+1] == 0x90 && in[ip+2] == 0xa0:
			o[op] = ' '
			op++
			ip += 2
			vc++
			pre = true

		// Plain ASCII
		case in[ip] <= '~':
			o[op] = in[ip]
			op++
			vc++
			pre = false

		default:
			o[op] = in[ip]
			op++
			pre = false

			// Only count UTF8 start byte and only if not combining
			if in[ip]&0b11000000 == 0b11000000 {
				r, _ := utf8.DecodeRune(in[ip:])
				if !unicode.IsMark(r) {
					vc++
				}
			}
		}
	}

	if vc > width && os != 0 {
		copy(o[os+1:], o[os:])
		o[os] = '\r'
		o[os+1] = '\n'
		op++
	}
	if !pre {
		for ; op > 0 && o[op-1] == ' '; op-- {
		}
	}

	return o[:op]
}
