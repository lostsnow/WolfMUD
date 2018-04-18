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

// splitLine is a regex to split name and data in a recordjar.
var splitLine = regexp.MustCompile(`^(?:\s*([^\s:]+):)?\s*(.*?)$`)

var (
	comment   = []byte("//") // Comment marker
	separator = []byte("%%") // Record separator marker
)

// Read takes as input an io.Reader - assuming the data to be in the WolfMUD
// recordjar format - and the field name to use for the freetext block. The
// input is parsed into a jar which is then returned.
//
// For details of the recordjar format see the separate package documentation.
func Read(in io.Reader, freetext string) (j Jar) {

	var (
		b   *bufio.Reader
		ok  bool
		err error

		// Variables for processing current line
		line   []byte   // current line from Reader
		tokens [][]byte // temp vars for name:data pair parsed from line
		name   string   // current name from line
		data   []byte   // current data from line
		field  string   // current field being processed (may differ from name)

		// Some flags to improve code readability
		noName     = false // true if line has no name
		noData     = false // true if line has no data
		noLine     = false // true if line has no name and no data
		noLastLine = false // true if last line had no name and no data
	)

	// If not using a buffered Reader, make it buffered
	if b, ok = in.(*bufio.Reader); !ok {
		b = bufio.NewReader(in)
	}

	// Make sure the field name to use for freetext is uppercased
	freetext = strings.ToUpper(freetext)

	// Setup an initially empty record for the Jar
	r := Record{}

	for err == nil {
		line, err = b.ReadBytes('\n')

		// If we read no data and find EOF continue and let loop exit
		if len(line) == 0 && err == io.EOF {
			continue
		}

		// Read and parse current line
		line = bytes.TrimRightFunc(line, unicode.IsSpace)
		tokens = splitLine.FindSubmatch(line)
		name, data = string(bytes.ToUpper(tokens[1])), tokens[2]

		noName = len(name) == 0
		noData = len(data) == 0
		noLine = noName && noData

		// Ignore comments found outside of freetext block
		if noName && field != freetext && bytes.HasPrefix(data, comment) {
			continue
		}

		// Handle record separator by recording current Record in Jar and setting
		// up a new next record, reset lastField seen and noLastLine flag.
		if noName && bytes.Equal(data, separator) {
			if len(r) > 0 {
				j = append(j, r)
				r = Record{}
			}
			field = ""
			noLastLine = false
			continue
		}

		// If we get a new name store it as the current field being processed
		if !noName {
			field = name
		}

		// Switch to freetext field if empty line and we are not already processing
		// the freetext block. If there was no lastField processed we need to
		// record the blank line so that it is included in the freetext block. This
		// lets us have a record that is freetext only and can start with a blank
		// line, which is not counted as a separator line.
		if noLine && field != freetext {
			if field == "" {
				noLastLine = true
			}
			field = freetext
			continue
		}

		// Handle freetext if already processing the freetext block, or we have no
		// field - in which case assume we are starting the freetext block
		if field == freetext || field == "" {

			// If last line was blank, current line is blank or current line starts
			// with whitespace and we already have some text in the freetext block,
			// then append a new line to terminate the last line and start a new one.
			// If not terminating last line, but we have some data in the freetext
			// already, then append a space before appending the current line.
			if noLastLine || noLine || (len(r[freetext]) != 0 && bytes.IndexFunc(line, unicode.IsSpace) == 0) {
				r[freetext] = append(r[freetext], '\n')
			} else {
				if _, ok = r[freetext]; ok {
					r[freetext] = append(r[freetext], ' ')
				}
			}

			r[freetext] = append(r[freetext], line...)

			noLastLine = noLine
			field = freetext
			continue
		}

		// Handle field. Append a space before appending text if continuation
		if _, ok = r[field]; ok {
			r[field] = append(r[field], ' ')
		}
		r[field] = append(r[field], data...)
	}

	// Append last record to the Jar if we have one
	if len(r) > 0 {
		j = append(j, r)
		r = Record{}
	}

	return
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

		// Find maximum field name length used in the current record.
		// Also record a normalised version of the field names used.
		maxLen = 0
		norm := map[string]string{}
		for field := range rec {
			norm[field] = text.TitleFirst(strings.ToLower(field))
			if norm[field] == freetext {
				continue
			}
			if len(field) > maxLen {
				maxLen = len(field)
			}
		}

		for field, data := range rec {
			if norm[field] == freetext {
				continue
			}
			data = text.Fold(data, 80-maxLen-sepLen)
			data = bytes.Replace(data, []byte("\r"), []byte(""), -1)
			for i, l := range bytes.Split(data, []byte("\n")) {
				if i == 0 {
					buf.Write(bytes.Repeat([]byte(" "), maxLen-len(field)))
					buf.WriteString(norm[field])
					buf.WriteString(": ")
				} else {
					buf.Write(bytes.Repeat([]byte(" "), maxLen+sepLen))
				}
				buf.Write(l)
				buf.WriteString("\n")
			}
		}
		for f, n := range norm {
			if n == freetext {
				data := text.Fold(rec[f], 80)
				data = bytes.Replace(data, []byte("\r"), []byte(""), -1)
				buf.WriteString("\n")
				buf.Write(data)
				buf.WriteString("\n")
				break
			}
		}
		buf.WriteString("%%\n")
		buf.WriteTo(out)
	}
}
