// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"strings"
)

// dictionary procides a convenient list of strings that can be checked to see
// if it contains a specific string. Checking is case insensative. A dictionary
// is basically a map with some helpers. A dictionary is typically created once
// and checked multiple times. If a dictionary is created in an init function
// or as a package level variable it does not to be locked if used for reading
// only. Otherise it must be protected with a lock like a normal map would.
//
// The values in the map are empty structs which use zero bytes.
type dictionary map[string]struct{}

// Dictionary returns a new dictionary containing the specified strings. The
// dictionary can then be checked to see if it contains a specific string (case
// insensitive) by calling the Contains method.
//
// NOTE: The strings in the dictionary are converted to uppercase.
func Dictionary(s ...string) (d dictionary) {
	d = map[string]struct{}{}
	for _, s := range s {
		d[strings.ToUpper(s)] = struct{}{}
	}
	return
}

// Contains searches the dictionary to see if it contains the specified string.
// If a match is found true is returned, otherwise false. The search is case
// insensitive.
func (d dictionary) Contains(s string) (found bool) {
	_, found = d[strings.ToUpper(s)]
	return
}
