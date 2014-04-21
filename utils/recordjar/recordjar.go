// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package recordjar implements the main file format used by WolfMUD.  It is
// based on a combination of RFC5322 and the Cookie Jar format as described by
// Eric Raymond in "The Art of Unix Programming", chapter 5:
//
//	http://www.catb.org/esr/writings/taoup/html/ch05s02.html
//
// It is not an actual implementation of the RFC5322 format just based on it:
//
//	- Unicode is allowed in header names and values
//	- Whitespace handling is more lenient and may proceed header names
//	- Line endings can be CRLF or LF
//	- Comments are lines starting with '//' characters
//	- Multiple records are separated by the '%%' sequence
//
// Here is a simple example of two starting locations:
//
//		//
//		// The Dragon's Breath tavern. L1 to L4
//		//
//				Ref: L1
//			 Type: Start
//			 Name: Fireplace
//		Aliases: TAVERN FIREPLACE
//			Exits: E→L3 SE→L4 S→L2
//
//		You are in the corner of a common room in the Dragon's Breath tavern.
//		There is a fire burning away merrily in an ornate fireplace giving
//		comfort to weary travellers. Shadows flicker around the room, changing
//		light to darkness and back again. To the south the common room extends
//		and east the common room leads to the tavern entrance.
//		%%
//				Ref: L2
//			 Type: Start
//			 Name: Common Room
//		Aliases: TAVERN COMMON
//			Exits: N→L1 NE→L3 E→L4
//
//		You are in a small, cosy common room in the Dragon's Breath tavern.
//		Looking around you see a few chairs and tables for patrons. To the east
//		there is a bar and to the north you can see a merry fireplace burning
//		away.
//
//
// When this example is read you would have a RecordJar which is a slice of
// Records - in this case two of them.
package recordjar

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"
)

// Record represents a section read from a record jar file. A Record is a map of
// strings keyed by header strings.
type Record map[string]string

// RecordJar is a slice of Records. When a file is read a RecordJar contains all
// of the Records from the file.
type RecordJar []Record

// Unmarshaler should be implemented by any type that can take data represented
// by a Record and parse / decode it.
type Unmarshaler interface {
	Unmarshal(Decoder)
	Init(ref Decoder, refs map[string]Unmarshaler)
}

type Marshaler interface {
	Marshal(Encoder)
}

// Constants for header line types
const (
	HS  = ""   // Header separator
	PS  = ":"  // name / data pair separator
	RS  = "%%" // Record separator
	REM = "//" // Remark / comment
)

// splitHeader is a regexp to split the header prefix from a line
var (
	splitHeader = regexp.MustCompile(`^(?:([^\s:]+):)?\s*(.*?)$`)
)

// Read reads data from the passed io.Reader and returns a RecordJar of Records.
// If there is an error returns a nil RecordJar and the error.
//
// TODO: Need to detail specifics of RecordJar, Record and file format.
func Read(reader io.Reader) (rj RecordJar, err error) {

	b := bufio.NewReader(reader)
	r := make(Record)

	currentHeader := ""
	line := ""

RECORDS:
	for {

		// If we have a record on the go store it and allocate a new one
		if len(r) != 0 {
			rj = append(rj, r)
			r = make(Record)
		}

		// Exit at EOF
		if err == io.EOF {
			break RECORDS
		}

		// Process record header lines
	HEADERS:
		for {
			line, err = b.ReadString('\n')
			line = strings.TrimSpace(line)

			switch {
			case line == HS:
				break HEADERS
			case line == RS:
				continue RECORDS
			case len(line) > 1 && line[0:2] == REM:
				continue HEADERS
			}

			tokens := splitHeader.FindStringSubmatch(line)
			newHeader, data := strings.ToLower(tokens[1]), tokens[2]

			if newHeader != "" {
				currentHeader = newHeader
			}

			if _, ok := r[currentHeader]; ok {
				r[currentHeader] += " "
			}
			r[currentHeader] += data

			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				continue RECORDS
			}
		}

		// Process free format data lines - between header separator (HS) and
		// record separator / EOF (RS).
		joiner := ""
		for {
			line, err = b.ReadString('\n')
			line = strings.TrimRightFunc(line, unicode.IsSpace)

			if line == RS || line == "" && err == io.EOF {
				continue RECORDS
			}

			if line == "" {
				r[":data:"] += "\n"
				joiner = "\n"
			} else {
				r[":data:"] += joiner + line
				joiner = " "
			}

			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				continue RECORDS
			}
		}
	}

	return rj, nil
}

// Write writes the passed RecordJar to the passed io.Writer.
func Write(writer io.Writer, rj RecordJar) (err error) {

	b := bufio.NewWriter(writer)

	for _, rec := range rj {

		// Find longest attribute name for pretty formatting
		length, maxlength := 0, 0
		for attr := range rec {
			if attr == ":data:" {
				continue
			}
			if length = len(attr); length > maxlength {
				maxlength = length
			}
		}

		// Write out name/data pairs with left padding on the name so
		// everything aligns
		spaces := strings.Repeat(" ", maxlength)
		padding := ""

		for attr, data := range rec {
			if attr == ":data:" {
				continue
			}
			padding = spaces[:maxlength-len(attr)]
			b.WriteString(padding)
			b.WriteString(strings.Title(attr))
			b.WriteString(PS)
			b.WriteByte(' ')
			b.WriteString(data)
			b.WriteByte('\n')
		}

		if data, found := rec[":data:"]; found {
			b.WriteByte('\n')
			b.WriteString(data)
			b.WriteByte('\n')
		}

		b.WriteString(RS)
		b.WriteByte('\n')
	}

	b.Flush()

	return nil
}
