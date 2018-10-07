// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package decode implements functions for decoding recordjar fields.
package decode

import (
	"bytes"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Define separator used for string lists
var listSeparator = []byte(":")

// String returns the []bytes data as a string with leading and trailing white
// space removed. This should only be used to decode fields and not the free
// text section. The decoder.Bytes function is preferred for the free text
// section as the section can contain meaningful leading and/or trailing blank
// lines for formatting.
func String(data []byte) string {
	return string(bytes.TrimSpace(data))
}

// Keyword returns the []bytes data as an uppercased string. This is helpful
// for keeping IDs and references consistent and independent of how they appear
// in e.g. data files. Any white space will be removed, either leading,
// trailing or within the keyword - a keyword with white space would actually
// be two or more keywords.
func Keyword(data []byte) string {
	out := make([]rune, 0, len(data))
	for _, r := range string(data) {
		if !unicode.IsSpace(r) {
			out = append(out, unicode.ToUpper(r))
		}
	}

	return string(out)
}

// KeywordList returns the []byte data as an uppercased slice of strings. The
// data is split on whitespace, extra whitespace is stripped and the individual
// 'words' are returned in the string slice. Duplicate keywrods will be
// removed.
func KeywordList(data []byte) []string {

	f := bytes.Fields(data)
	k := make([]string, len(f))
	pos := 0

	for x := range f {
		k[pos] = Keyword(f[x])
		for _, y := range k[0:pos] {
			if y == k[pos] {
				pos--
				break
			}
		}
		pos++
	}
	sort.Strings(k[0:pos])

	return k[0:pos]
}

// PairList returns the []byte data as uppercassed pairs of strings in a map.
// The data is first split on white space and extra white space is stripped.
// The pairs are then split into a name and value on the first non-letter,
// non-digit. If we take exits as an example:
//
//  Exits: E→L3 SE→L4 S→ W
//
// Results in a map with four pairs:
//
//  map[string]string {
//    "E": "L3",
//    "SE": "L4",
//    "S": "",
//    "W": "",
//  }
//
// Here the separator used is '→' but any non-letter or non-digit may be used.
// If the same name occurs more than once only the first instance will be used.
// A name may appear by itself, as in 'E', or with a separator, as in 'E→' in
// which case the value will be an empty string. If no name is given, for
// example '→L3' any value will be ignored.
func PairList(data []byte) (pairs map[string]string) {

	var i, l int
	var name string

	pairs = make(map[string]string)

	for _, data := range bytes.Fields(data) {
		i, l = indexSeparator(data)
		name = Keyword(data[:i])
		if name == "" {
			continue
		}
		if _, ok := pairs[name]; !ok {
			pairs[name] = Keyword(data[i+l:])
		}
	}
	return
}

// StringList returns the []byte data as a []string by splitting the data on a
// colon separator. Any leading or trailing white space will be removed from
// the returned strings.
func StringList(data []byte) (s []string) {
	var w []byte

	for _, t := range bytes.Split(data, listSeparator) {
		if w = bytes.TrimSpace(t); len(w) > 0 {
			s = append(s, string(w))
		}
	}
	return
}

// KeyedStringList splits the []byte data into a map of keywords and strings.
// The []byte data is first split on a colon (:) separator to determine the
// pairs. The keyword is then split from the beginning of each pair on the
// first non-letter or non-digit. For example:
//
//  Vetoes:  GET→You can't get it.
//        : DROP→You can't drop it.
//
// Would produce a map with two entries:
//
//  map[string]string{
//    "GET": "You can't get it.",
//    "DROP": "You can't drop it.",
//  }
//
// If a keyword is specified more than once only the first instance will be
// used. Leading and trailing whitespace will be removed from the returned
// strings.
func KeyedStringList(data []byte) (list map[string]string) {

	var i, l int // index and length of list separator found
	list = make(map[string]string)

	// Don't reuse StringList as it adds duplicated white space trimming
	for _, s := range bytes.Split(data, listSeparator) {
		i, l = indexSeparator(s)
		name := Keyword(s[:i])
		if len(name) == 0 {
			continue
		}
		if _, ok := list[name]; !ok {
			list[name] = String(s[i+l:])
		}
	}
	return
}

// Bytes returns a copy of the []byte data. Important so we don't accidentally
// pin a larger backing array in memory via the source slice. Any leading or
// trailing white space will be trimmed EXCEPT new lines '\n', which the
// trimming will end at. This is the preferred way to decode a free text
// section as it allows for leading/trailing blank lines.
func Bytes(data []byte) []byte {
	out := make([]byte, len(data), len(data))
	copy(out, data)
	out = bytes.TrimFunc(out, func(r rune) bool {
		if r == '\n' {
			return false
		}
		return unicode.IsSpace(r)
	})
	return out
}

// Duration returns the []byte data as a time.Duration rounded (half up) to the
// nearest second. The data is parsed using time.ParseDuration and will default
// to 0 if the data cannot be parsed.
func Duration(data []byte) (t time.Duration) {
	var err error

	// Lower case passed duration and remove all white space
	d := make([]rune, 0, len(data))
	for _, r := range string(data) {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'z' {
			d = append(d, r)
		} else {
			if !unicode.IsSpace(r) {
				d = append(d, unicode.ToLower(r))
			}
		}
	}

	if t, err = time.ParseDuration(string(d)); err != nil {
		log.Printf("Duration field has invalid value %q, using default: %s", data, t)
	}
	t = t.Round(time.Second)
	return t
}

// DateTime returns the []byte data as a time.Time. The data is parsed using
// time.Parse and is expected to conform to RFC1123Z - as written by
// encode.DateTime. For example: Thu, 20 Sep 2018 20:24:33 +0000
//
// If there is an error parsing the data the date and time will default to the
// current date and time. The returned date/time will use the UTC timezone.
func DateTime(data []byte) (t time.Time) {
	var err error
	stamp := String(data)
	if t, err = time.Parse(time.RFC1123Z, stamp); err != nil {
		// If not parsed as RFC1123Z try pre WolfMUD v0.0.11 legacy RFC1123
		if t, err = time.Parse(time.RFC1123, stamp); err != nil {
			t = time.Now()
		}
	}
	t = t.Truncate(time.Second).UTC()

	if err != nil {
		log.Printf("DateTime field has invalid value %q, using default: %s", data, t)
	}

	return t
}

// Boolean returns the []byte data as a boolean value. The data is parsed using
// strconv.ParseBool and will default to false if the data cannot be parsed.
// Using strconv.parseBool allows true and false to be represented in many
// ways. For example: 0, f, F, false, False, FALSE, 1, t, T, true, True, TRUE.
// As a special case data of length zero will default to true. This allows true
// to be represented as the presence or absence of just a keyword. For example:
//
//  Door: EXIT→E RESET→1m JITTER→1m OPEN
//
// Here OPEN is a boolean and will default to true.
func Boolean(data []byte) (b bool) {
	s := strings.TrimSpace(string(data))
	if len(s) == 0 {
		return true
	}
	var err error
	if b, err = strconv.ParseBool(s); err != nil {
		log.Printf("Boolean field has invalid value %q, using default: %t", s, b)
	}
	return
}

// Integer returns the []byte data as an integer value. The []byte is parsed
// using strconv.Atoi and will default to 0 if the data cannot be parsed. The
// valid range is at least -2147483648 to 2147483647.
func Integer(data []byte) (i int) {
	var err error
	if i, err = strconv.Atoi(string(data)); err != nil {
		log.Printf("Integer field has invalid value %q, using default: %d", data, i)
	}
	return
}

// indexSeparator returns the position (starting at 0) and length in bytes of
// the first separator rune found. If no separator is found the position
// returned will be equal to the length of 'b' and the length returned will be
// 0. The separator is taken to be the first rune found that is a not a letter,
// digit or white space.
func indexSeparator(b []byte) (index int, size int) {
	index = bytes.IndexFunc(b, func(r rune) bool {
		return !unicode.In(r, unicode.Digit, unicode.Letter, unicode.White_Space)
	})
	if index != -1 {
		_, size = utf8.DecodeRune(b[index:])
	} else {
		index, size = len(b), 0
	}
	return
}
