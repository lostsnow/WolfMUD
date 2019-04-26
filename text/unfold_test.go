// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
	"fmt"
	"testing"
)

func TestUnfold(t *testing.T) {
	for _, tc := range []struct {
		data string
		want string
	}{
		{"", ""},
		{" ", ""},
		{"  ", ""},
		{"\n", "\n"},
		{"\n\n", "\n\n"},
		{"  \n", "\n"},
		{"  \n  \n", "\n\n"},
		{"a ", "a"},
		{"a  ", "a"},
		{"a \n", "a\n"},
		{"a  \n", "a\n"},
		{"  \n  \n   abc", "\n\n   abc"},
		{"Sentance one. Sentance two.", "Sentance one. Sentance two."},
		{"Sentance one.\nSentance two.", "Sentance one. Sentance two."},
		{"Sentance one.\n  Sentance two.", "Sentance one.\n  Sentance two."},
		{"Sentance one.\n\nSentance two.", "Sentance one.\n\nSentance two."},
		{"Sentance one.\n\n  Sentance two.", "Sentance one.\n\n  Sentance two."},
		{"Sentance one.  \nSentance two.", "Sentance one. Sentance two."},
		{"\nSentance one.\nSentance two.", "\nSentance one. Sentance two."},
		{
			"Sentance one.\n\x1b[32m  Sentance two.",
			"Sentance one.\n\x1b[32m  Sentance two.",
		},
		{"a\n\x1b[0;0m a", "a\n\x1b[0;0m a"},
		{"a\n\x1b[31m\x1b[32m a", "a\n\x1b[31m\x1b[32m a"},
		{
			"A report on MUDs from December, 1990. Written by Dr Richard Bartle of MUD1\nfame. An interesting read into the past.",
			"A report on MUDs from December, 1990. Written by Dr Richard Bartle of MUD1 fame. An interesting read into the past.",
		},
		{
			"\nWolfMUD Copyright 1984-2018 Andrew 'Diddymus' Rolfe\n\n  World\n  Of\n  Living\n  Fantasy\n\nWelcome to WolfMUD!\n\n",
			"\nWolfMUD Copyright 1984-2018 Andrew 'Diddymus' Rolfe\n\n  World\n  Of\n  Living\n  Fantasy\n\nWelcome to WolfMUD!\n\n",
		},
	} {
		t.Run(fmt.Sprintf("%40q", tc.data), func(t *testing.T) {
			have := Unfold([]byte(tc.data))
			if !bytes.Equal(have, []byte(tc.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, tc.want)
			}
		})
	}
}

func BenchmarkUnfold(b *testing.B) {
	for _, test := range []string{
		"The quick brown fox jumps over the lazy dog.",
		"The quick brown fox\njumps over the lazy dog.",
		"You are in the corner of the common room in the dragon's breath\ntavern. A fire burns merrily in an ornate fireplace, giving comfort to\nweary travellers. The fire causes shadows to flicker and dance around\nthe room, changing darkness to light and back again. To the south the\ncommon room continues and east the common room leads to the tavern\nentrance.",
	} {
		data := []byte(test)
		b.Run(fmt.Sprintf("%s", test[:20]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Unfold(data)
			}
		})
	}
}

func TestConsumeEscape(t *testing.T) {
	for _, tc := range []struct {
		data string
		want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"\x1b[31m", ""},
		{"\x1b[31;32m", ""},
		{"\x1b[31", "\x1b[31"},
		{"\x1b[3", "\x1b[3"},
		{"\x1b[", "\x1b["},
		{"\x1b", "\x1b"},
		{"\x1b[0;", "\x1b[0;"},
		{"\x1b[0m", ""},
		{"\x1b[m", ""},
		{"\x1b[x", "\x1b[x"},
		{"\x1b[0x", "\x1b[0x"},
	} {
		t.Run(fmt.Sprintf("%40q", tc.data), func(t *testing.T) {
			have := string(consumeEscape([]rune(tc.data)))
			if have != tc.want {
				t.Errorf("\nhave %+q\nwant %+q", have, tc.want)
			}
		})
	}
}
