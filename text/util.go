// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"unicode"
)

// TitleFirst will return the passed string with the first rune in the string
// Titlecased.
func TitleFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToTitle(r[0])
	return string(r)
}
