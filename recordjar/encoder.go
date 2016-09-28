// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"bytes"
	"strconv"
	"strings"
	"time"
)

// encoder is used to convert specific data types into recordjar []byte values.
type encoder struct{}

// Encode is a convenient way to access the recordjar encoding functions. For
// example:
//
//	d := recordjar.Encode.Duration(data)
//
var Encode = encoder{}

// String returns the given string as a []byte.
func (encoder) String(s string) []byte {
	return []byte(s)
}

// Keyword returns the passed string as an uppercased []byte. This is helpful
// for keeping IDs and references consistent and independent of how they appear
// in e.g. data files.
func (encoder) Keyword(s string) []byte {
	return bytes.ToUpper([]byte(s))
}

// KeywordList returns the []string data as a whitespace separated, uppercased
// slice of bytes.
func (encoder) KeywordList(s []string) []byte {
	return bytes.ToUpper([]byte(strings.Join(s, " ")))
}

// PairList returns the passed slice of string pairs as an uppercased []byte.
// Each pair of strings is separated with the given delimiter. All of the
// string pairs are then concatenated together separated by whitespace.
//
//	exits := [][2]string{
//		[2]string{"E", "L3"},
//		[2]string{"SE", "L4"},
//		[2]string{"S", "L2"},
//	}
//	data := PairList(exits, "→")
//
// Results in data being a byte slice containing "E→L3 SE→L4 S→L2".
func (encoder) PairList(data [][2]string, delimeter string) (pairs []byte) {
	for _, pair := range data {
		pairs = append(pairs, bytes.ToUpper([]byte(pair[0]))...)
		pairs = append(pairs, delimeter...)
		pairs = append(pairs, bytes.ToUpper([]byte(pair[1]))...)
		pairs = append(pairs, ' ')
	}
	if len(data) > 0 {
		pairs = pairs[0 : len(pairs)-1]
	}
	return
}

// Bytes returns a copy of the passed []byte. Important so we don't
// accidentally pin a larger backing array in memory via the slice.
func (encoder) Bytes(dataIn []byte) []byte {
	dataOut := make([]byte, len(dataIn), len(dataIn))
	copy(dataOut, dataIn)
	return dataOut
}

// Duration returns the given time.Duration as a []byte. The byte slice will
// have the format "0h0m0.0s".
func (encoder) Duration(d time.Duration) []byte {
	return []byte(d.String())
}

// Duration returns the given time.Duration as a []byte. The byte slice will be
// formatted according to RFC1123. For example "Mon, 02 Jan 2006 15:04:05 MST".
func (encoder) DateTime(t time.Time) []byte {
	return []byte(t.Format(time.RFC1123))
}

// Boolean returns the given boolean as a []byte containing either "TRUE" or
// "FALSE".
func (encoder) Boolean(b bool) []byte {
	if b {
		return []byte("TRUE")
	}
	return []byte("FALSE")
}

// Integer returns the passed integer value as a stringified []byte.
func (encoder) Integer(i int) []byte {
	return []byte(strconv.Itoa(i))
}
