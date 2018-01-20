// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package encode implements functions for encoding recordjar fields.
package encode

import (
	"bytes"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// String returns the given string as a []byte.
func String(s string) []byte {
	return []byte(s)
}

// Keyword returns the passed string as an uppercased []byte. This is helpful
// for keeping IDs and references consistent and independent of how they appear
// in e.g. data files.
func Keyword(s string) []byte {
	return bytes.ToUpper([]byte(s))
}

// KeywordList returns the []string data as a whitespace separated, uppercased
// slice of bytes.
func KeywordList(s []string) []byte {
	return bytes.ToUpper([]byte(strings.Join(s, " ")))
}

// PairList returns the passed map of string pairs as an uppercased []byte.
// Each pair of strings is separated with the given delimiter. All of the
// string pairs are then concatenated together separated by whitespace.
//
//	exits := map[string]string{
//		"E":  "L3",
//		"SE": "L4",
//		"S":  "L2",
//	}
//	data := PairList(exits, '→')
//
// Results in data being a byte slice containing "E→L3 SE→L4 S→L2".
func PairList(data map[string]string, delimeter rune) (pairs []byte) {
	d := make([]byte, utf8.RuneLen(delimeter))
	utf8.EncodeRune(d, delimeter)

	for name, value := range data {
		pairs = append(pairs, bytes.ToUpper([]byte(name))...)
		pairs = append(pairs, d...)
		pairs = append(pairs, bytes.ToUpper([]byte(value))...)
		pairs = append(pairs, ' ')
	}
	if len(data) > 0 {
		pairs = pairs[0 : len(pairs)-1]
	}
	return
}

// Bytes returns a copy of the passed []byte. Important so we don't
// accidentally pin a larger backing array in memory via the slice.
func Bytes(dataIn []byte) []byte {
	dataOut := make([]byte, len(dataIn), len(dataIn))
	copy(dataOut, dataIn)
	return dataOut
}

// Duration returns the given time.Duration as a []byte. The byte slice will
// have the format "0h0m0.0s".
func Duration(d time.Duration) []byte {
	return []byte(d.String())
}

// Duration returns the given time.Duration as a []byte. The byte slice will be
// formatted according to RFC1123. For example "Mon, 02 Jan 2006 15:04:05 MST".
func DateTime(t time.Time) []byte {
	return []byte(t.Format(time.RFC1123))
}

// Boolean returns the given boolean as a []byte containing either "TRUE" or
// "FALSE".
func Boolean(b bool) []byte {
	if b {
		return []byte("TRUE")
	}
	return []byte("FALSE")
}

// Integer returns the passed integer value as a stringified []byte.
func Integer(i int) []byte {
	return []byte(strconv.Itoa(i))
}
