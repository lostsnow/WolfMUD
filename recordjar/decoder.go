// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// decoder is used to convert recordjar []byte values into specific types.
type decoder struct{}

// Decode is a convenient way to access the recordjar decoding functions. For
// example:
//
//  d := recordjar.Decode.Duration(data)
//
var Decode = decoder{}

// String returns the []bytes data as a string.
func (decoder) String(data []byte) string {
	return string(data)
}

// Keyword returns the []bytes data as an uppercased string. This is helpful
// for keeping IDs and references consistent and independent of how they appear
// in e.g. data files.
func (decoder) Keyword(data []byte) string {
	return strings.ToUpper(string(data))
}

// KeywordList returns the []byte data as an uppercased slice of strings. The
// data is split on whitespace, extra whitespace is stripped and the individual
// 'words' are returned in the string slice.
func (decoder) KeywordList(data []byte) []string {
	return strings.Fields(strings.ToUpper(string(data)))
}

// PairList returns the []byte data as uppercassed pairs of strings in a slice.
// The data is first split on whitespace and extra whitespace is stripped. The
// 'words' are then split into pairs on the first non-unicode letter or digit.
// If we take exits as an example:
//
//  Exits: E→L3 SE→L4 S→ W
//
// Results in four pairs:
//
//  [2]string{"E","L3"}
//  [2]string{"SE","L4"}
//  [2]string{"S",""}
//  [2]string{"W",""}
//
// Here the separator used is → but any non-unicode letter or digit may be
// used.
func (decoder) PairList(data []byte) (pairs [][2]string) {
	for _, pair := range strings.Fields(string(data)) {
		runes := []rune(strings.ToUpper(pair))
		split := false
		for i, r := range runes {
			if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
				pairs = append(pairs, [2]string{string(runes[:i]), string(runes[i+1:])})
				split = true
				break
			}
		}
		if !split {
			pairs = append(pairs, [2]string{string(runes[:]), ""})
		}
	}
	return
}

// StringList returns the []byte date as a []string by splitting the data on a
// colon separator.
func (decoder) StringList(data []byte) (s []string) {
	for _, t := range strings.Split(string(data), ":") {
		if w := strings.TrimSpace(t); w != "" {
			s = append(s, w)
		}
	}
	return
}

// KeyedString returns the []byte data as an uppercassed keywrd and a string.
// The keyword is split from the beginning of the []byte on the first
// non-unicode letter or digit. For example:
//
//   input: []byte("GET→You can't get it.")
//  output: [2]string{"GET", "You can't get it."}
//
// Here the separator used is → but any non-unicode letter or digit may be
// used.
func (decoder) KeyedString(data []byte) (pair [2]string) {
	runes := []rune(string(data))
	split := false
	for i, r := range runes {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) && !unicode.IsSpace(r) {
			key := strings.TrimSpace(strings.ToUpper(string(runes[:i])))
			data := strings.TrimSpace(string(runes[i+1:]))
			pair = [2]string{key, data}
			split = true
			break
		}
	}
	if !split {
		pair = [2]string{string(runes[:]), ""}
	}
	return
}

// KeyedStringList splits the []byte on a colon (:) separator and passes each
// result through KeyedString. KeyedStringList is a shorthand for combining
// StringList and KeyedString. For example the follow RecordJar record's data:
//
//  Vetoes:  GET→You can't get it.
//        : DROP→You can't drop it.
//
// Would produce a slice with two entries:
//
//  [2]string{"GET", "You can't get it."}
//  [2]string{"DROP", "You can't drop it."}
//
func (decoder) KeyedStringList(data []byte) (list [][2]string) {
	for _, w := range Decode.StringList(data) {
		list = append(list, Decode.KeyedString([]byte(w)))
	}
	return
}

// Bytes returns a copy of the []byte data. Important so we don't accidentally
// pin a larger backing array in memory via the slice.
func (decoder) Bytes(dataIn []byte) []byte {
	dataOut := make([]byte, len(dataIn), len(dataIn))
	copy(dataOut, dataIn)
	return dataOut
}

// Duration returns the []byte data as a time.Duration. The data is parsed
// using time.ParseDuration and will default to 0 if the data cannot be parsed.
func (decoder) Duration(data []byte) (t time.Duration) {
	var err error
	if t, err = time.ParseDuration(strings.ToLower(string(data))); err != nil {
		log.Printf("Duration field has invalid value %q, using default: %s", data, t)
	}
	return t
}

// Boolean returns the []byte data as a boolean value. The data is parsed using
// strconv.ParseBool and will default to false if the data cannot be parsed.
// Using strconv.parseBool allows true and false to be represented in many ways.
func (decoder) Boolean(data []byte) (b bool) {
	var err error
	if b, err = strconv.ParseBool(string(data)); err != nil {
		log.Printf("Boolean field has invalid value %q, using default: %t", data, b)
	}
	return
}

// Integer returns the []byte data as an integer value. The []byte is parsed
// using strconv.Atoi and will default to 0 if the data cannot be parsed.
func (decoder) Integer(data []byte) (i int) {
	var err error
	if i, err = strconv.Atoi(string(data)); err != nil {
		log.Printf("Integer field has invalid value %q, using default: %d", data, i)
	}
	return
}
