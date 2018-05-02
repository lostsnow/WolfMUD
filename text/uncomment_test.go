// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text_test

import (
	"fmt"
	"testing"

	"code.wolfmud.org/WolfMUD.git/text"
)

func TestUncomment(t *testing.T) {
	for _, test := range []struct {
		input string
		want  string
	}{
		{"", ""},
		{"#", "#"}, // No space after # so not a comment
		{"# ", ""}, // Space after # so is an (empty) comment
		{"# empty string", ""},
		{"^$", "^$"},
		{"^#$", "^#$"},
		{"^$ # match empty string", "^$"},
		{"^abc   \n$", "^abc$"},
		{"^\nabc\n$", "^abc$"},
		{"^\r\nabc\r\n$", "^abc$"},
		{`
			^
			(.*)
			$
			`, "^(.*)$"},
		{`
			^     # Match start of string
			(.*)  # Capture everything
			$     # Match end of string
			`, "^(.*)$"},
		{`
			^     # Match start of string

			(.*)  # Capture everything

			$     # Match end of string
			`, "^(.*)$"},
		{`# Starting comment
			^     # Match start of string
			(.*)  # Capture everything
			$     # Match end of string
			# Ending comment`, "^(.*)$"},
		{`
			^
			# (.*) # Commented out line
			[ ]    # Embedded space
			$
			`, "^[ ]$"},
		{`
			^               # Match start of string
			(?:             # Start non-capture group
			  ([A-Za-z]+)   # Capture first 'word'
			  .*            # Ignore everything else
			)               # End non-capture group
			$               # Match end of string
			`, "^(?:([A-Za-z]+).*)$"},
	} {
		t.Run(test.want, func(t *testing.T) {
			have := text.Uncomment(test.input)
			want := test.want
			if have != want {
				t.Errorf("have: %q, want: %q", have, want)
			}
		})
	}
}

func BenchmarkUncomment(b *testing.B) {
	for x, bench := range []string{`
		(?m)         # Match in multi-line mode
		(?:          # Start a non-capture, alternating group
		  \s*#\s.*$  # Match line ending in a  '#' delimited comment
		|            # OR
		  ^\s+       # Leading whitespace
		|            # OR
		  \s*\n      # Optional trailing whitespace, followed by a new line
		)            # End group
		`, `
		^            # match start of string
		(?:          # non-capture group for 'field:'
		  \s*        # don't capture whitespace before 'field'
		  ([^\s:]+)  # capture 'field' - non-whitespace/non-colon
		  :          # non-capture match of colon as field:value separator
		)?           # match non-captured 'field:' zero or once, prefer once
		\s*          # consume any whitepace - leading or after 'field:' if matched
		(.*?)        # capture everything left umatched, not greedy
		$            # match at end of string
		`,
	} {
		_ = text.Uncomment("make sure re is compiled and hot before benchmarking")
		b.Run(fmt.Sprintf("Uncomment-%d)", x), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = text.Uncomment(bench)
			}
		})
	}
}
