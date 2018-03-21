// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"
	"unicode"

	"code.wolfmud.org/WolfMUD.git/text"
)

// Jar represents the collection of Records in a recordjar.
type Jar []Record

// Record represents the separate records in a recordjar.
type Record map[string][]byte

// splitLine is a regex to split fields and data in a recordjar.
var splitLine = regexp.MustCompile(`^(?:([^\s:]+):)?\s*(.*?)$`)

var (
	comment   = []byte("//") // Comment marker
	separator = []byte("%%") // Record separator marker
)

// Read takes as input an io.Reader assuming the data to be in the WolfMUD
// recordjar format and the fieldname to use for the free text block. The input
// is parsed into a jar which is then returned.
//
// For details of the recordjar format see the separate package documentation.
func Read(in io.Reader, freetext string) Jar {

	// work variables for the current field and it's data
	var (
		b         *bufio.Reader // Buffered reader for the input
		j         Jar           // Jar being built
		r         Record        // Current jar record
		raw       []byte        // Current raw line
		line      []byte        // Current clean (trimmed) line
		tokens    [][]byte      // Line split into field / data tokens
		field     string        // Current field name
		data      []byte        // Current data
		ok        bool          // Map key checking flag
		blankLine bool          // Flag when previous line is blank
		err       error
	)

	b = bufio.NewReader(in)
	j = Jar{}
	r = Record{}

	// Make sure the name to use for the free text block is uppercased.
	freetext = strings.ToUpper(freetext)

	// Start off assuming last field seen was the special freetext field. If the
	// record does not actually start with free text then the field name found
	// will overwrite this. This has the effect of allowing a record to consist
	// of just a free text block without having to precede it with a blank line.
	lastField := freetext

	// Helper: if current record not empty add it to the jar and create a new
	// record. Also reset lastField to the free text block field - see above.
	addJar := func() {
		if len(r) > 0 {
			j = append(j, r)
			r = Record{}
		}
		lastField = freetext
		blankLine = false
	}

	// Main processing loop
	for err == nil {

		raw, err = b.ReadBytes('\n')

		// If not processing freetext block
		if _, ok = r[freetext]; !ok {

			line = bytes.TrimSpace(raw)

			switch {
			case len(line) == 0:
				// Ignore blank lines caused by errors, otherwise start free text block.
				// Also need to check if first line is a blank line in which case we
				// include it in the freetextblock and not as a separator.
				if err == nil {
					r[freetext] = []byte{}
					blankLine = lastField == freetext
				}
				continue
			case bytes.HasPrefix(line, comment):
				continue
			case bytes.Equal(line, separator):
				addJar()
				continue
			}

			tokens = splitLine.FindSubmatch(line)
			field, data = string(bytes.ToUpper(tokens[1])), tokens[2]

			// If we have no field name and we are processing the freetext block the
			// free text block started on the first line of the record. So we have to
			// store the line for the free text block with only the right side
			// stripped of whitespace.
			if field == "" && lastField == freetext {
				r[freetext] = bytes.TrimRightFunc(raw, unicode.IsSpace)
				continue
			}

			if field != "" {
				lastField = field
			}

			if _, ok = r[lastField]; ok {
				r[lastField] = append(r[lastField], ' ')
			}
			r[lastField] = append(r[lastField], data...)

		} else {
			// Processing freetext block if we get here

			// Only trim right to keep leading whitespace...
			line = bytes.TrimRightFunc(raw, unicode.IsSpace)

			// ... but trim left to check for a record separator
			if bytes.Equal(bytes.TrimLeftFunc(line, unicode.IsSpace), separator) {
				addJar()
				continue
			}

			// If previous line was blank or current line is empty or starts with
			// whitespace append a line feed
			if blankLine || len(line) == 0 || bytes.IndexFunc(line, unicode.IsSpace) == 0 {
				r[freetext] = append(r[freetext], '\n')
			}

			blankLine = len(line) == 0

			// If didn't append a line feed and we already have text append a space
			if l := len(r[freetext]); l != 0 && r[freetext][l-1] != '\n' {
				r[freetext] = append(r[freetext], ' ')
			}

			// Append actual line data
			r[freetext] = append(r[freetext], line...)
		}

	}

	addJar() // Add last record to the jar

	return j
}

// Write writes out a Record Jar to the specified io.Writer. It also takes as
// input the fieldname used for the free text block in the jar.
//
// For details of the recordjar format see the separate package documentation.
func (j Jar) Write(out io.Writer, freetext string) {

	freetext = text.TitleFirst(strings.ToLower(freetext))

	var (
		maxLen int
		buf    bytes.Buffer
		sepLen int = len(": ")
	)

	for _, rec := range j {

		// Find maximum field name length used in the current record
		maxLen = 0
		for field := range rec {
			if field == freetext {
				continue
			}
			if len(field) > maxLen {
				maxLen = len(field)
			}
		}

		for field, data := range rec {
			if field == freetext {
				continue
			}
			data = text.Fold(data, 80-maxLen-sepLen)
			data = bytes.Replace(data, []byte("\r"), []byte(""), -1)
			for i, l := range bytes.Split(data, []byte("\n")) {
				if i == 0 {
					buf.Write(bytes.Repeat([]byte(" "), maxLen-len(field)))
					buf.WriteString(text.TitleFirst(strings.ToLower(field)))
					buf.WriteString(": ")
				} else {
					buf.Write(bytes.Repeat([]byte(" "), maxLen+sepLen))
				}
				buf.Write(l)
				buf.WriteString("\n")
			}
		}
		if data, ok := rec[freetext]; ok {
			data = text.Fold(data, 80)
			data = bytes.Replace(data, []byte("\r"), []byte(""), -1)
			buf.WriteString("\n")
			buf.Write(data)
			buf.WriteString("\n")
		}
		buf.WriteString("%%\n")
		buf.WriteTo(out)
	}
}
