// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"regexp"
)

// uncomment is a regular expression to remove embedded comments from a string.
// See Uncomment function for details.
var uncomment = regexp.MustCompile(`(?m)(?:\s*#\s.*$|^\s+|\s*\n)`)

// Uncomment takes a string with comments and returns the string with
// whitespace and comments removed. Comments are expected to be delimited with
// a '#' character followed by at least one whitespace. When a string is
// uncommented each line will be stripped of leading and trailing whitespace.
// Comments will be removed from the '#'+whitespace separator to the end of the
// line.
//
// The prime motivation for this is to allow regular expressions to be
// commented inline - in raw string literals - similar to Perl's /x modifier.
func Uncomment(re string) string {
	return uncomment.ReplaceAllLiteralString(re, "")
}
