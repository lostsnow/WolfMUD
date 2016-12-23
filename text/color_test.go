// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
	"fmt"
	"testing"
)

var testCasesSubstitute = []struct {
	data string
	want string
}{
	// Simple substitutions
	{"[RESET]", "\033[0m"},
	{"[BOLD]", "\033[1m"},
	{"[NORMAL]", "\033[22m"},
	{"[BLACK]", "\033[30m"},
	{"[RED]", "\033[31m"},
	{"[GREEN]", "\033[32m"},
	{"[YELLOW]", "\033[33m"},
	{"[BROWN]", "\033[33m"},
	{"[BLUE]", "\033[34m"},
	{"[MAGENTA]", "\033[35m"},
	{"[CYAN]", "\033[36m"},
	{"[WHITE]", "\033[37m"},
	{"[BGBLACK]", "\033[40m"},
	{"[BGRED]", "\033[41m"},
	{"[BGGREEN]", "\033[42m"},
	{"[BGYELLOW]", "\033[43m"},
	{"[BGBROWN]", "\033[43m"},
	{"[BGBLUE]", "\033[44m"},
	{"[BGMAGENTA]", "\033[45m"},
	{"[BGCYAN]", "\033[46m"},
	{"[BGWHITE]", "\033[47m"},

	// Multiple substitutions
	{"[RED][GREEN][YELLOW]", "\033[31m\033[32m\033[33m"},
	{"[[RED][GREEN][YELLOW]", "[\033[31m\033[32m\033[33m"},
	{
		"[BLACK]R[RED]A[GREEN]I[YELLOW]N[BLUE]B[MAGENTA]O[CYAN]W[WHITE]![RESET]",
		"\033[30mR\033[31mA\033[32mI\033[33mN\033[34mB\033[35mO\033[36mW\033[37m!\033[0m",
	},

	// Nested and mismatched square braces
	{"[]RED[]", "[]RED[]"},
	{"][RED][", "]\033[31m["},
	{"[[RED]", "[\033[31m"},
	{"[RED]]", "\033[31m]"},
}

func TestColorize(t *testing.T) {
	for _, tc := range testCasesSubstitute {
		t.Run(fmt.Sprintf("Substitute %s", tc.data), func(t *testing.T) {
			have := Colorize([]byte(tc.data))
			if !bytes.Equal(have, []byte(tc.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, tc.want)
			}
		})
	}
}

var benchmarkCasesSubstitute = []string{
	"[RED]WolfMUD[RESET]",
	"[BLACK] R [RED] A [GREEN] I [YELLOW] N [BLUE] B [MAGENTA] O [CYAN] W [WHITE] ! [RESET]",
	`[CYAN][ Fireplace ][WHITE]
You are in the corner of the common room in the dragon's breath tavern. A fire
burns merrily in an ornate fireplace, giving comfort to weary travellers. The
fire causes shadows to flicker and dance around the room, changing darkness to
light and back again. To the south the common room continues and east the common
room leads to the tavern entrance.
[GREEN]
You see a curious brass lattice here.
You see a small green ball here.
You see a small red ball here.
You see an iron bound chest here.

[CYAN]You can see exits [YELLOW]east, southeast and south.
[MAGENTA]>`,
	`[ Fireplace ]
You are in the corner of the common room in the dragon's breath tavern. A fire
burns merrily in an ornate fireplace, giving comfort to weary travellers. The
fire causes shadows to flicker and dance around the room, changing darkness to
light and back again. To the south the common room continues and east the common
room leads to the tavern entrance.

You see a curious brass lattice here.
You see a small green ball here.
You see a small red ball here.
You see an iron bound chest here.

You can see exits [YELLOW]east, southeast and south.
>`,
}

func BenchmarkColorize(b *testing.B) {
	for x, tc := range benchmarkCasesSubstitute {
		data := []byte(tc)
		b.Run(fmt.Sprintf("Test %d", x), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Colorize(data)
			}
		})
	}
}
