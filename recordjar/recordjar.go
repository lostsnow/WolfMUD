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
	"unicode"
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

	b := bufio.NewReader(in)
	j := Jar{}
	r := Record{}

	// Make sure the name to use for the free text block is uppercased.
	freetext = (string)(bytes.ToUpper([]byte(freetext)))

	// Start off assuming last field seen was the special freetext field. If the
	// record does not actually start with free text then the field name found
	// will overwrite this. This has the effect of allowing a record to consist
	// of just a free text block without having to precede it with a blank line.
	lastField := freetext

	// work variables for the current field and it's data
	var (
		raw    []byte   // Current raw line
		line   []byte   // Current clean (trimmed) line
		tokens [][]byte // Line split into field / data tokens
		field  string   // Current field name
		data   []byte   // Current data
		ok     bool     // Map key checking flag
		err    error
	)

	for {
		raw, err = b.ReadBytes('\n')
		line = bytes.TrimSpace(raw)

		// Work out our field and data parts
		tokens = splitLine.FindSubmatch(line)
		switch len(tokens) {
		case 3:
			field, data = string(bytes.ToUpper(tokens[1])), tokens[2]
		case 2:
			field, data = "", tokens[1]
		default:
			panic("should not happen!")
		}

		// Ignore comments if not processing the free text block
		if _, ok = r[freetext]; !ok {
			if field == "" && bytes.HasPrefix(data, comment) {
				continue
			}
		}

		// If not processing the free text block do we need to start it? There is
		// no point starting a free block at EOF so just exit out of the loop.
		if _, ok = r[freetext]; !ok {
			if field == "" && len(data) == 0 {
				if err == io.EOF {
					break
				}
				r[freetext] = []byte{}
				continue
			}
		}

		// Special handling of comments and leading whitespace when processing the
		// free text block
		if _, ok = r[freetext]; ok {
			field = freetext
			switch {

			// If all the whitespace was stripped from the raw line output a blank line
			case len(raw) != 0 && len(line) == 0:
				// Check if previous line is terminated, if not terminate it
				if l := len(r[freetext]); l != 0 && r[freetext][l-1] != '\n' {
					data = []byte("\n\n")
				} else {
					data = []byte("\n")
				}

			// If raw line started with whitespace keep it but strip from the right still
			case bytes.HasPrefix(raw, []byte(" ")) || bytes.HasPrefix(raw, []byte("\t")):
				if l := len(r[freetext]); l != 0 && r[freetext][l-1] != '\n' {
					data = []byte("\n")
					data = append(data, bytes.TrimRightFunc(raw, unicode.IsSpace)...)
				} else {
					data = bytes.TrimRightFunc(raw, unicode.IsSpace)
				}

			}
		}

		// If we find a record separator start new record after adding current
		// record to the jar - but don't add empty records to the jar.
		if bytes.Equal(line, separator) {
			if len(r) > 0 {
				j = append(j, r)
				r = Record{}
			}
			lastField = freetext
			continue
		}

		// Record a change in the field name. If it doesn't change we are appending
		// data to the last field seen so append a space to it's data first - as
		// long as the previous value does not end with a newline or the new data
		// starts with a newline, otherwise we get trailing whitespace.
		if field != "" && field != lastField {
			lastField = field
		} else {
			if l := len(r[lastField]); l != 0 && r[lastField][l-1] != '\n' {
				if d := len(data); d != 0 && data[0] != '\n' {
					r[lastField] = append(r[lastField], ' ')
				}
			}
		}

		r[lastField] = append(r[lastField], data...)

		if err != nil {
			break
		}
	}

	// Add last record to the jar as long as the record isn't empty
	if len(r) > 0 {
		j = append(j, r)
	}

	return j
}

// Write writes out a Record Jar to the specified io.Writer. It also takes as
// input the fieldname used for the free text block in the jar.
//
// For details of the recordjar format see the separate package documentation.
//
// TODO: Add wrapping of long values.
func (j Jar) Write(out io.Writer, freetext string) {

	freetext = (string)(bytes.ToTitle([]byte(freetext)))

	var (
		maxLen int
		buf    bytes.Buffer
	)

	for _, rec := range j {
		maxLen = 0
		for field, _ := range rec {
			if field == freetext {
				continue
			}
			if len(field) > maxLen {
				maxLen = len(field)
			}
		}
		maxLen++
		for field, data := range rec {
			if field == freetext {
				continue
			}
			buf.Write(bytes.Repeat([]byte(" "), maxLen-len(field)))
			buf.Write(bytes.Title(bytes.ToLower([]byte(field))))
			buf.WriteString(": ")
			buf.Write(data)
			buf.WriteString("\n")
		}
		if data, ok := rec[freetext]; ok {
			buf.WriteString("\n")
			buf.Write(data)
			buf.WriteString("\n")
		}
		buf.WriteString("%%\n")
		buf.WriteTo(out)
	}
}
