// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package decode implements functions for decoding recordjar fields.
package decode

import (
	"bytes"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

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
// 'words' are returned in the string slice.
func KeywordList(data []byte) []string {
	return strings.Fields(strings.ToUpper(string(data)))
}

// PairList returns the []byte data as uppercassed pairs of strings in a map.
// The data is first split on whitespace and extra whitespace is stripped. The
// 'words' are then split into pairs on the first non-unicode letter or digit.
// If we take exits as an example:
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
// Here the separator used is → but any non-unicode letter or digit may be
// used.
func PairList(data []byte) (pairs map[string]string) {

	pairs = make(map[string]string)

	for _, name := range strings.Fields(string(data)) {
		name = strings.TrimSpace(strings.ToUpper(name))
		value := ""
		if i, l := indexSeparator(name); i != -1 {
			value = name[i+l:]
			name = name[:i]
		}
		pairs[name] = value
	}
	return
}

// StringList returns the []byte data as a []string by splitting the data on a
// colon separator.
func StringList(data []byte) (s []string) {
	for _, t := range strings.Split(string(data), ":") {
		if w := strings.TrimSpace(t); w != "" {
			s = append(s, w)
		}
	}
	return
}

// KeyedString returns the []byte data as an uppercassed keyword and a string
// value. The keyword is split from the beginning of the []byte on the first
// non-unicode letter or digit. For example:
//
//   input: []byte("GET→You can't get it.")
//  output: "GET", "You can't get it."
//
// Here the separator used is → but any non-unicode letter or digit may be
// used.
func KeyedString(data []byte) (name, value string) {

	name = string(data)

	if i, l := indexSeparator(name); i != -1 {
		value = name[i+l:]
		name = name[:i]
	}

	name = strings.TrimSpace(strings.ToUpper(name))
	value = strings.TrimSpace(value)

	return
}

// KeyedStringList splits the []byte data into a map of keywords and strings.
// The []byte data is first split on a colon (:) separator to determine the
// pairs. The keyword is then split from the beginning of each pair on the
// first non-unicode letter or digit. For example:
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
func KeyedStringList(data []byte) (list map[string]string) {
	list = make(map[string]string)
	for _, w := range StringList(data) {
		name, value := KeyedString([]byte(w))
		list[name] = value
	}
	return
}

// Bytes returns a copy of the []byte data. Important so we don't accidentally
// pin a larger backing array in memory via the slice.
func Bytes(dataIn []byte) []byte {
	dataOut := make([]byte, len(dataIn), len(dataIn))
	copy(dataOut, dataIn)
	return dataOut
}

// Duration returns the []byte data as a time.Duration. The data is parsed
// using time.ParseDuration and will default to 0 if the data cannot be parsed.
func Duration(data []byte) (t time.Duration) {
	var err error
	if t, err = time.ParseDuration(strings.ToLower(string(data))); err != nil {
		log.Printf("Duration field has invalid value %q, using default: %s", data, t)
	}
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
	if t, err = time.Parse(time.RFC1123Z, string(data)); err != nil {
		// If not parsed as RFC1123Z try pre WolfMUD v0.0.11 legacy RFC1123
		if t, err = time.Parse(time.RFC1123, string(data)); err != nil {
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
// Using strconv.parseBool allows true and false to be represented in many ways.
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
// using strconv.Atoi and will default to 0 if the data cannot be parsed.
func Integer(data []byte) (i int) {
	var err error
	if i, err = strconv.Atoi(string(data)); err != nil {
		log.Printf("Integer field has invalid value %q, using default: %d", data, i)
	}
	return
}

// indexSeparator returns the position and length in bytes of the first
// separator rune found. If no separator is found the position returned will be
// -1 and the length returned will be 0.
func indexSeparator(s string) (index int, size int) {
	index = strings.IndexFunc(s, func(r rune) bool {
		return !unicode.In(r, unicode.Digit, unicode.Letter, unicode.White_Space)
	})
	if index != -1 {
		_, size = utf8.DecodeRune([]byte(s[index:]))
	}
	return
}
