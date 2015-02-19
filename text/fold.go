// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
)

const space = 1

func Fold(in string, width int) []byte {

	// Can we take a short cut?
	if len(in) <= width {
		return []byte(in)
	}

	// Add extra line feed to end of input. Will cause final word and line to be
	// 'flushed' from the buffers. The extra line feed itself will not be output
	// becuase it will still be in the buffers - so we don't need to trim it off.
	in += "\n"

	// Setup some buffers:
	//	bw: word buffer	  - initially up to 32 (arbitrary) characters
	//	bl: line buffer		- initially width + 1 word
	//	bo: output buffer - initially input length + 32 (arbitrary) line breaks
	var (
		bw = bytes.NewBuffer(make([]byte, 0, 32))
		bl = bytes.NewBuffer(make([]byte, 0, width+32))
		bo = bytes.NewBuffer(make([]byte, 0, len(in)+32))
	)

	lb := true // Only leading blanks have been written to a word

	for _, r := range in {

		if (r != ' ' && r != '\n') || (r == ' ' && lb == true) {
			bw.WriteRune(r)
			lb = !(r != ' ')
			continue
		}

		if bl.Len()+space+bw.Len() >= width {
			if bo.Len() != 0 {
				bo.WriteByte('\n')
			}
			bl.WriteTo(bo)
			lb = true
		}

		if bl.Len() != 0 {
			bl.WriteByte(' ')
		}
		bw.WriteTo(bl)

		if r == '\n' {
			if bo.Len() != 0 {
				bo.WriteByte('\n')
			}
			bl.WriteTo(bo)
			lb = true
		}

	}

	return bo.Bytes()
}
