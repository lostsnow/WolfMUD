// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text_test

import (
	"fmt"
	"regexp"

	"code.wolfmud.org/WolfMUD.git/text"
)

// Example of a commented regular expression and uncommenting it.
func ExampleUncomment_simple() {
	text := text.Uncomment(`
		(?m)         # Match in multi-line mode
		(?:          # Start a non-capture, alternating group
		  \s*#\s.*$  # Match line ending in a  '#' delimited comment
		|            # OR
		  ^\s+       # Leading whitespace
		|            # OR
		  \s*\n      # Optional trailing whitespace, followed by a new line
		)            # End group
	`)
	fmt.Println(text)

	// Output: (?m)(?:\s*#\s.*$|^\s+|\s*\n)
}

// Example of a commented regular expression, uncommenting it and compiling it.
func ExampleUncomment_compile() {
	r := regexp.MustCompile(text.Uncomment(`
		de     # Match 'de'
		.      # Then any character
		[a-z]  # Then a lowercase letter
	`))
	fmt.Println(r.FindString("abcdefghi"))

	// Output: defg
}
