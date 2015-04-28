// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
)

const (
	reset = 0
	space = 1
)

// Fold takes a string and reformats it so lines have a maximum length of
// width. Fold will handle multibyte runes. However it cannot handle 'wide'
// runes - those that are wider than a normal single character when displayed.
// This is because the required information is actually contained in the font
// files of the font in use.
//
// For example the Chinese for 9 is 九 (U+4E5D). Even in a monospaced font 九
// will take up the space of two columns.
func Fold(in string, width int) []byte {

	// Can we take a short cut? Note we are just counting bytes here.
	if len(in) <= width {
		return []byte(in)
	}

	// Add extra line feed to end of input. Will cause final word and line to be
	// 'flushed' from the buffers. The extra line feed itself will not be output
	// because it will still be in the buffers - so we don't need to trim it off.
	in += "\n"

	var (
		word = bytes.NewBuffer(make([]byte, 0, 32))
		line = bytes.NewBuffer(make([]byte, 0, width+32))
		page = bytes.NewBuffer(make([]byte, 0, len(in)+24))
	)

	var (
		wordLen, lineLen, pageLen = 0, 0, 0 // word, line and output length in runes
		blank                     = true    // true when line is empty or only blanks
	)

	for _, r := range in {

		if (r != ' ' && r != '\n') || (r == ' ' && blank == true) {
			word.WriteRune(r)
			wordLen++
			blank = r == ' '
			continue
		}

		if lineLen+space+wordLen > width {
			if pageLen != reset {
				page.WriteByte('\n')
				pageLen++
			}
			line.WriteTo(page)
			pageLen += lineLen
			lineLen = reset
			blank = true
		}

		if lineLen != reset {
			line.WriteByte(' ')
			lineLen++
		}
		word.WriteTo(line)
		lineLen += wordLen
		wordLen = reset

		if r == '\n' {
			if pageLen != reset {
				page.WriteByte('\n')
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
