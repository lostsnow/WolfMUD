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
	"sort"
	"strings"
	"unicode"

	"code.wolfmud.org/WolfMUD.git/text"
)

// Jar represents the collection of Records in a recordjar.
type Jar []Record

// Record represents the separate records in a recordjar.
type Record map[string][]byte

// splitLine is a regex to split fields and data in a recordjar .wrj file. The
// result of a FindSubmatch should always be a [][]byte of length 3 consisting
// of: the string matched, the field name, the data.
var splitLine = regexp.MustCompile(text.Uncomment(`
	^            # match start of string
	(?:          # non-capture group for 'field:'
		\s*        # don't capture whitespace before 'field'
	  ([^\s:]+)  # capture 'field' - non-whitespace/non-colon
	  :          # non-capture match of colon as field:value separator
	)?           # match non-captured 'field:' zero or once, prefer once
	\s*          # consume any whitepace - leading or after 'field:' if matched
	(.*?)        # capture everything left umatched, not greedy
	$            # match at end of string
`))

var (
	comment   = []byte("//") // Comment marker
	separator = []byte("%%") // Record separator marker
)

// Read takes as input an io.Reader - assuming the data to be in the WolfMUD
// recordjar format - and the field name to use for the freetext block. The
// input is parsed into a jar which is then returned.
//
// For details of the recordjar format see the separate package documentation.
//
// BUG(diddymus): There is no provision for preserving comments.
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

// Write writes out a Record Jar to the specified io.Writer. The freetext
// string is used to specify which fieldname in a record should be used for the
// freetext block. For example, if the freetext string is 'Description' then
// any fields named description in a record will be written out in the freetext
// block.
//
// For details of the recordjar format see the separate package documentation.
//
// TODO(diddymus): Uppercase character after a hyphen in field names so that
// we can have 'On-Action', 'On-Reset', 'On-Cleanup' automatically.
//
// BUG(diddymus): There is no provision for writing out comments.
// BUG(diddymus): The empty field "" is invalid, currently dropped silently.
// BUG(diddymus): Unicode used in field names not normalised so 'Nаme' with a
// Cyrillic 'а' (U+0430) and 'Name' with a latin 'a' (U+0061) would be
// different fields.
func (j Jar) Write(out io.Writer, freetext string) {

	const maxLineWidth = 80           // Maximum length of a line in a .wrj file
	const separatorLength = len(": ") // Length of field/data separator
	var buf bytes.Buffer              // Temporary buffer for current record

	// A slice of spaces we can re-slice to get variable lengths of padding
	padding := bytes.Repeat([]byte(" "), maxLineWidth-separatorLength)

	// Normalise passed in freetext field name
	freetext = text.TitleFirst(strings.ToLower(freetext))

	for _, rec := range j {

		norm := make(map[string][]byte, len(rec)) // Copy of rec, normalised keys
		keys := make([]string, 0, len(rec))       // List of sortable norm keys
		maxFieldLen := 0                          // Longest normalised field name

		// Copy fields from rec to norm but with normalised keys. As we go through
		// the field names note the length of the longest normalised field name.
		for field, data := range rec {

			if field == "" { // Ignore invalid empty field name
				continue
			}

			field = text.TitleFirst(strings.ToLower(field))
			norm[field], keys = data, append(keys, field)

			if field == freetext { // Ignore freetext field name (never written out)
				continue
			}

			if l := len(field); l > maxFieldLen {
				maxFieldLen = l
			}
		}

		// Write out fields for current record in the order given by the sorted keys
		sort.Strings(keys)
		for _, field := range keys {

			// Ignore the freetext field as it has to be written last
			if field == freetext {
				continue
			}

			// Fold the field data, which will now have network '\r\n' line endings.
			// Strip the '\r' to get Unix line endings. Finally split the data into
			// separate lines using `\n` as the delimiter.
			data := text.Fold(norm[field], maxLineWidth-maxFieldLen-separatorLength)
			data = bytes.Replace(data, []byte("\r"), []byte(""), -1)
			lines := bytes.Split(data, []byte("\n"))

			// Write field name, separator, and first data line
			buf.Write(padding[0 : maxFieldLen-len(field)])
			buf.WriteString(field)
			buf.WriteByte(':')
			if len(lines[0]) != 0 {
				buf.WriteByte(' ')
				buf.Write(lines[0])
			}
			buf.WriteByte('\n')

			// Write continuation data lines
			for _, l := range lines[1:] {
				buf.Write(padding[0 : maxFieldLen+separatorLength])
				buf.Write(l)
				buf.WriteByte('\n')
			}
		}

		// Write out the freetext block, if we have one.
		if data, ok := norm[freetext]; ok {

			// Write separator line if record has fields other than freetext block.
			if len(norm) > 1 {
				buf.WriteByte('\n')
			}

			data = text.Fold(data, maxLineWidth)
			data = bytes.Replace(data, []byte("\r"), []byte(""), -1)
			buf.Write(data)
			buf.WriteByte('\n')
		}

		// If we have written any fields for the record, write a record separator.
		if len(norm) > 0 {
			buf.WriteString("%%\n")
		}
		buf.WriteTo(out)
	}
}
