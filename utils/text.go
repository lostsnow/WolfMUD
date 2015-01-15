// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package utils

import (
	"bytes"
)

const space = 1

func Fold(in string, width int) string {

	// Can we take a short cut?
	if len(in) <= width {
		return in
	}

	// Add extra line feed to end of input. Will cause final word and line to be
	// 'flushed' from the buffers. The extra line feed itself will not be output
	// becuase it will still be in the buffers - so we don't need to trim it off.
	in += "\n"

	bw := bytes.Buffer{} // Buffered current word
	bl := bytes.Buffer{} // Buffered current line
	bo := bytes.Buffer{} // Buffered output

	lb := true // Only leading blanks have been written to a word

	for _, r := range in {

		if (r != ' ' && r != '\n') || (r == ' ' && lb == true) {
			bw.WriteRune(r)
			lb = !(r != ' ')
			continue
		}

		if bl.Len()+space+bw.Len() >= width {
			if bo.Len() != 0 {
				bo.WriteRune('\n')
			}
			bo.WriteString(bl.String())
			bl.Reset()
			lb = true
		}

		if bl.Len() != 0 {
			bl.WriteRune(' ')
		}
		bl.WriteString(bw.String())
		bw.Reset()

		if r == '\n' {
			if bo.Len() != 0 {
				bo.WriteRune('\n')
			}
			bo.WriteString(bl.String())
			bl.Reset()
			lb = true
		}

	}

	return bo.String()
}
