// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.
package text

import (
	"strings"
)

// List returns a slice of strings as a comma separated list with 'and' between
// the last items. If an empty slice is passed the returned string will be
// empty. If a slice with only one element is passed the returned string will
// be equal to the passed element.
//
// For example List([]string{"A", "B", "C"}) returns "A, B and C".
func List(items []string) string {
	switch l := len(items); l {
	case 0:
		return ""
	case 1:
		return items[0]
	default:
		return strings.Join(items[:l-1], ", ") + " and " + items[l-1]
	}
}
