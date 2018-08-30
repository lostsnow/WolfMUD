// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package encode implements functions for encoding recordjar fields.
package encode

import (
	"bytes"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// String returns the given string as a []byte with leading and trailing white
// space removed.
func String(s string) []byte {
	return bytes.TrimSpace([]byte(s))
}

// Keyword returns the passed string as an uppercased []byte. This is helpful
// for keeping IDs and references consistent and independent of how they appear
// in e.g. data files. Any white space will be removed, either leading,
// trailing or within the keyword - a keyword with white space would actually
// be two or more keywords.
func Keyword(s string) []byte {

	out := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			out = append(out, unicode.ToUpper(r))
		}
	}

	return []byte(string(out))
}

// KeywordList returns the []string data as a white space separated, uppercased
// slice of bytes. Multiple keywords will have consistent ordering. Duplicate
// keywords will be omitted. Any white space will be removed, either leading,
// trailing or within a keyword - a keyword with white space would actually be
// two or more keywords.
func KeywordList(s []string) []byte {

	u := make([]string, len(s))
	pos := 0

	for x := range s {
		u[pos] = string(Keyword(s[x]))
		if u[pos] == "" {
			continue
		}
		for _, y := range u[0:pos] {
			if y == u[pos] {
				pos--
				break
			}
		}
		pos++
	}
	sort.Strings(u[0:pos])

	return []byte(strings.Join(u[0:pos], " "))
}

// PairList returns the passed map of string name/value pairs as an uppercased
// []byte. Each name/value pair is separated with the given delimiter. Any
// white space will be removed, either leading, trailing or within a name or
// value - a name or value with white space would actually be more than a
// single pair of name/value. All of the string name/delimiter/value pairs are
// then concatenated together separated by white space.
//
//	exits := map[string]string{
//		"E":  "L3",
//		"SE": "L4",
//		"S":  "L2",
//	}
//	data := PairList(exits, '→')
//
// Results in data being a byte slice containing "E→L3 SE→L4 S→L2". Multiple
// name/value pairs will have consistent ordering. It should be noted that
// using a space for a delimiter will result in a keyword list, not a pair
// list, which may cause problems when the field is decoded again.
func PairList(data map[string]string, delimiter rune) []byte {

	pairs := make([]string, len(data))
	pos := 0
	b := strings.Builder{}

	for name, value := range data {

		// Shortcut if just a name and no value
		if len(value) == 0 {
			pairs[pos] = string(Keyword(name))
			pos++
			continue
		}

		b.Write(Keyword(name))
		b.WriteRune(delimiter)
		b.Write(Keyword(value))
		pairs[pos] = b.String()
		pos++
		b.Reset()
	}
	sort.Strings(pairs[0:pos])
	return []byte(strings.Join(pairs[0:pos], " "))
}

// StringList returns a list of strings delimited by a colon separator. Each
// string in the list will start with the delimiter on a new line.
func StringList(data []string) []byte {
	s := make([]string, len(data))
	for x := range data {
		s[x] = strings.TrimSpace(data[x])
	}
	sort.Strings(s)
	return []byte(strings.Join(s, "\n: "))
}

// KeyedString returns the name uppercased and concatenated to the value using
// the delimiter, as a []byte. For example:
//
//  KeyedString("get", "You cannot get that!", '→')
//
// Results in a []byte containing "GET→You cannot get that!".
func KeyedString(name, value string, delimiter rune) (data []byte) {
	d := make([]byte, utf8.RuneLen(delimiter))
	utf8.EncodeRune(d, delimiter)

	data = append(data, Keyword(name)...)

	v := strings.TrimSpace(value)
	if len(v) != 0 {
		data = append(data, d...)
		data = append(data, v...)
	}
	return data
}

// KeyedStringList returns the map of names and strings as a list of colon
// separated keyed strings. For example:
//
//	m := map[string]string{
//		"get":  "You cannot get that!",
//		"look": "Your eyes hurt to look at it!",
//	}
//	data := KeyedStringList(m, '→')
//
// Results in data containing:
//
//  GET→You cannot get that!\n: LOOK→Your eyes hurt to look at it!
//
func KeyedStringList(pairs map[string]string, delimiter rune) (data []byte) {
	s := make([]string, len(pairs))
	pos := 0
	for name, value := range pairs {
		s[pos] = string(KeyedString(name, value, delimiter))
		pos++
	}
	sort.Strings(s)
	return []byte(strings.Join(s, "\n: "))
	return data
}

// Bytes returns a copy of the passed []byte. Important so we don't
// accidentally pin a larger backing array in memory via the slice.
func Bytes(dataIn []byte) []byte {
	dataOut := make([]byte, len(dataIn), len(dataIn))
	copy(dataOut, dataIn)
	return dataOut
}

// Duration returns the given time.Duration as a []byte. The byte slice will
// have the format "0h0m0s" and is rounded (half up) to the nearest second.
// Leading and trailing zero units will be omitted.
func Duration(d time.Duration) []byte {
	b := []byte(d.Round(time.Second).String())
	b = bytes.Replace(b, []byte("m0s"), []byte("m"), 1)
	b = bytes.Replace(b, []byte("h0m"), []byte("h"), 1)
	return b
}

// DateTime returns the given time.Time as a []byte. The byte slice will be
// formatted according to RFC1123 and converted to the UTC timezone.
//
// For example, "Mon, 02 Jan 2006 15:04:05 UTC".
func DateTime(t time.Time) []byte {
	return []byte(t.UTC().Format(time.RFC1123))
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
