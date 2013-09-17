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
