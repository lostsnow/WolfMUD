// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
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
	"unicode/utf8"
)

// String is a helper that returns the value of a header from a Record as a
// string. If the header is not found in the Record an empty string is returned.
func (r Record) String(property string) string {
	if _, ok := r[property]; !ok {
		log.Printf("Property %q not found. Defaulting to empty string", property)
		return ""
	}
	return strings.TrimSpace(r[property])
}

// Keyword is a helper that returns the value of a header from a Record as an
// UPPERCASED string. If the header is not found in the Record an empty string
// is returned.
//
// This function is helpful for Ids and references which are case insensative
// and for consistency when matching are usually uppercased.
func (r Record) Keyword(property string) string {
	return strings.ToUpper(r.String(property))
}

// KeywordList is a helper that returns the value of a header from a Record
// interpreted as whitespace separated keywords. It returns the keywords as a
// slice of uppercased strings. If the header is not found in the Record an
// empty string slice is returned.
func (r Record) KeywordList(property string) []string {
	if _, ok := r[property]; !ok {
		log.Printf("Property %q not found. Defaulting to empty list", property)
		return []string{}
	}
	return strings.Fields(strings.ToUpper(r[property]))
}

// PairList is a helper that returns the value of a header from a Record
// interpreted as whitespace separated pairs of values. The pairs are split
// using the first non-digit and non-letter separator. For example exits could
// be specified as one of:
//
//	Exits: E→L3 SE→L4 S→L2
//	Exits: E=L3 SE=L4 S=L2
//	Exits: E>L3 SE>L4 S>L2
//	Exits: E.L3 SE.L4 S.L2
//
// In the case of multiple non-digits and/or non-letters only the first is used
// as the seperator. For example:
//
//	Exits: E→L1.a // direction = 'E', Location reference = 'L1.a'
//
func (r Record) PairList(property string) (pairs [][2]string) {
	if _, ok := r[property]; !ok {
		log.Printf("Property %q not found. Defaulting to empty pair list", property)
		return
	}

	splitter := func(r rune) bool {
		return !unicode.IsDigit(r) && !unicode.IsLetter(r)
	}

	for _, pair := range strings.Fields(r[property]) {
		//split := strings.FieldsFunc(pair, splitter)
		split := strings.IndexFunc(pair, splitter)
		if split == -1 {
			log.Printf("Ignoring invalid pair: %s", pair)
			continue
		}
		_, runeSize := utf8.DecodeRuneInString(pair[split:])
		pairs = append(pairs, [2]string{pair[:split], pair[split+runeSize:]})
	}
	return
}

// Int is a helper that returns the value of a header from a Record interpreted
// - as parsed by strconv.Atoi - as an integer. If the header is not found
// in the Record or the value cannot be parsed as an integer integer zero is
// returned.
func (r Record) Int(property string) (i int) {
	if _, ok := r[property]; !ok {
		log.Printf("Property %q not found. Defaulting to zero", property)
		return 0
	}
	var err error
	if i, err = strconv.Atoi(r[property]); err != nil {
		log.Printf("Error retrieving %q as type int: %s. Defaulting to zero.", r[property], err)
		return 0
	}
	return i
}

// Duration is helper that returns the value of a header from a Record
// interpreted - as parsed by time.ParseDuration - as a duration of time. If the
// header is not found in the Record or the value cannot be parsed as a duration
// a zero duration is returned.
func (r Record) Duration(property string) (d time.Duration) {
	if _, ok := r[property]; !ok {
		log.Printf("Property %q not found. Defaulting to zero", property)
		return 0
	}
	var err error
	if d, err = time.ParseDuration(r[property]); err != nil {
		log.Printf("Error parsing %q as type time.Duration: %s. Defaulting to zero.", r[property], err)
		return 0
	}
	return d
}
