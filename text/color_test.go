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
	{"[RESET]", "\x1b[0m"},
	{"[BOLD]", "\x1b[1m"},
	{"[NORMAL]", "\x1b[22m"},
	{"[BLACK]", "\x1b[30m"},
	{"[RED]", "\x1b[31m"},
	{"[GREEN]", "\x1b[32m"},
	{"[YELLOW]", "\x1b[33m"},
	{"[BROWN]", "\x1b[33m"},
	{"[BLUE]", "\x1b[34m"},
	{"[MAGENTA]", "\x1b[35m"},
	{"[CYAN]", "\x1b[36m"},
	{"[WHITE]", "\x1b[37m"},
	{"[BGBLACK]", "\x1b[40m"},
	{"[BGRED]", "\x1b[41m"},
	{"[BGGREEN]", "\x1b[42m"},
	{"[BGYELLOW]", "\x1b[43m"},
	{"[BGBROWN]", "\x1b[43m"},
	{"[BGBLUE]", "\x1b[44m"},
	{"[BGMAGENTA]", "\x1b[45m"},
	{"[BGCYAN]", "\x1b[46m"},
	{"[BGWHITE]", "\x1b[47m"},

	// Multiple substitutions
	{"[RED][GREEN][YELLOW]", "\x1b[31m\x1b[32m\x1b[33m"},
	{"[[RED][GREEN][YELLOW]", "[\x1b[31m\x1b[32m\x1b[33m"},
	{
		"[BLACK]R[RED]A[GREEN]I[YELLOW]N[BLUE]B[MAGENTA]O[CYAN]W[WHITE]![RESET]",
		"\x1b[30mR\x1b[31mA\x1b[32mI\x1b[33mN\x1b[34mB\x1b[35mO\x1b[36mW\x1b[37m!\x1b[0m",
	},

	// Nested and mismatched square braces
	{"[]RED[]", "[]RED[]"},
	{"][RED][", "]\x1b[31m["},
	{"[[RED]", "[\x1b[31m"},
	{"[RED]]", "\x1b[31m]"},
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
