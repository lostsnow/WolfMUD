// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
	"unicode"
)

// These constants are not really necessary but make the fold code easier to
// read and understand. If chars or lines are defined too small there will
// potentially be additional allocations needed but wasted space will be
// reduced. If chars or lines are defined too large the allocations will be
// reduced but unused space will be allocated.
//
// TODO: Tune chars and lines at runtime based on average text sizes being
// processed? Will need to set maximum limits to avoid runaway sizing based on
// deliberatly large text being sent by players causing a denial of service.
const (
	reset = 0  // Reset buffer to start (position zero) or test if at start
	space = 1  // Width in bytes of a space
	chars = 32 // Starting number of characters for initial word buffer sizing
	lines = 24 // Starting number of lines for page initial buffer sizing
)

var (
	lf   = []byte("\n")   // End of line used internally
	crlf = []byte("\r\n") // End of line for network data
	esc  = '\033'         // Escape control code, same as 0x1b or ^[
)

// Fold takes a string and reformats it so lines have a maximum length of the
// passed width. Fold will handle multibyte runes. However it cannot handle
// 'wide' runes - those that are wider than a normal single character when
// displayed. This is because the required information is actually contained in
// the font files of the font in use at the 'client' end.
//
// For example the Chinese for 9 is 九 (U+4E5D). Even in a monospaced font 九
// will take up the space of two columns.
//
// For combining characters Fold will assume combining marks are zero width.
// For example 'a' plus a combining grave accent U+0061 U+0300 will be counted
// as a single character. However it is better to use and actual latin small
// letter a with grave 'à' U+00E0. Either should work as expected.
//
// It is expected that the end of line markers for incoming data are Unix line
// feeds (LF, '\n') and outgoing data will have network line endings, carriage
// return + line feed pairs (CR+LF, '\r\n').
func Fold(in []byte, width int) []byte {

	// Can we take a short cut? Counting bytes is fine although we may end up
	// with a string shorter than we think it is if there are multibyte runes.
	if len(in) <= width {
		return bytes.Replace(in, lf, crlf, -1)
	}

	// Add extra line feed to end of input. Will cause final word and line to be
	// 'flushed' from the buffers. The extra line feed itself will not be output
	// because it will still be in the buffers - so we don't need to trim it off.
	in = append(in, '\n')

	var (
		word = bytes.NewBuffer(make([]byte, 0, chars))
		line = bytes.NewBuffer(make([]byte, 0, width+chars))
		page = bytes.NewBuffer(make([]byte, 0, len(in)+lines))
	)

	var (
		wordLen, lineLen, pageLen = 0, 0, 0 // word, line and output length in runes
		blank                     = true    // true when line is empty or only blanks
		control                   = false   // true when processing a control sequence
	)

	for _, r := range bytes.Runes(in) {

		// Are we starting a control sequence?
		if r == esc {
			control = true
		}

		// Control codes are zero width and do not add to the length of the word
		// but are written out. Any character in the range 0x40 - 0x7E (ASCII '@'
		// through to ASCII '~') ends a control sequence.
		if control {
			word.WriteRune(r)
			if '@' <= r && r <= '~' {
				control = false
			}
			continue
		}

		// Consider non-spacing and enclosing marks as zero width. This allows for
		// the use of combining marks. For example a lower case 'a' plus combining
		// grave accent to compose 'à' as well as a literal 'à' will work.
		if r > '~' && unicode.In(r, unicode.Mn, unicode.Me) {
			word.WriteRune(r)
			continue
		}

		if (r != ' ' && r != '\n') || (r == ' ' && blank == true) {
			word.WriteRune(r)
			wordLen++
			blank = blank && r == ' '
			continue
		}

		if lineLen+space+wordLen > width {
			if pageLen != reset {
				page.Write(crlf)
				pageLen++
			}
			line.WriteTo(page)
			pageLen += lineLen
			lineLen = reset
		}

		if lineLen != reset {
			line.WriteByte(' ')
			lineLen++
		}
		word.WriteTo(line)
		lineLen += wordLen
		wordLen = reset

		if r == '\n' {

			// An initial linefeed does not count towards the page length. This is
			// normally used to move output off of the player's prompt line.
			if pageLen == reset && lineLen == reset {
				page.Write(crlf)
				continue
			}

			if pageLen != reset {
				page.Write(crlf)
				pageLen++
			}
			line.WriteTo(page)
			pageLen += lineLen
			lineLen = reset
			blank = true
		}

	}

	return page.Bytes()
}
