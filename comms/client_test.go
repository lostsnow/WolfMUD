// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"bytes"
	"fmt"
	"testing"
)

var cleanData = []struct {
	data string
	want string
}{
	{"", ""},
	{"\b", ""},
	{"\b\b", ""},
	{"abc", "abc"},
	{"abd\bc", "abc"},
	{"\babc", "abc"},
	{"abcd\b", "abc"},
	{"\babcd\b", "abc"},
	{"def\b\b\babc", "abc"},
	{"def\b\b\b\b\babc", "abc"},
	{"\b\b\bdef\b\b\b\b\babc", "abc"},
	{"\bThe quick brown fox jumps over the lazy dog.", "The quick brown fox jumps over the lazy dog."},
	{"The quick brown fox j\bjumps over the lazy dog.", "The quick brown fox jumps over the lazy dog."},
	{"The quick brown fox jumps over the lazy dog..\b", "The quick brown fox jumps over the lazy dog."},
	{"Hello world!\n", "Hello world!"},
	{"That\b\bis is the\b\b\ba test!\b,\b.", "This is a test."},
	{"\b\b\b\btest", "test"},
	{"\b\b\btesting\b\b\b", "test"},
	{"test\b\b\b\b\b\bThis is a test.", "This is a test."},
	{"\b\bThat\b\b\b\b\bt\bThis is a test of an\b\bthe emergency broadcasting system!\b.'\b", "This is a test of the emergency broadcasting system."},
	{"æ\bearth", "earth"},
	{"the æ\bearth", "the earth"},
	{"Mikoł\blaj Hoł\blysz", "Mikolaj Holysz"},
	{"Mikol\błaj Hol\błysz", "Mikołaj Hołysz"},
	{"æ", "æ"},
	{"ææ", "ææ"},
	{"æææ", "æææ"},
	{"ææææ", "ææææ"},
	{"a\u00A3\bbc", "abc"},
	{"a\u2211\bbc", "abc"},
	{"a\U0001f78e\bbc", "abc"},
	{"a\u0061\u0300\bbc", "abc"},
	{"aæ\u00A3\bbc", "aæbc"},
	{"aæ\u2211\bbc", "aæbc"},
	{"aæ\U0001f78e\bbc", "aæbc"},
	{"aæ\u0061\u0300\bbc", "aæbc"},

	{"a\nb\nc", "abc"},       // Embeded line feeds
	{"a\x00b\x00c", "abc"},   // Null bytes
	{"a\x98b", "ab"},         // Embeded ASCII SOS - Start of string
	{"a\xc2\x98b", "ab"},     // Embeded UTF-8 SOS - Start of string
	{"a\u0098b", "ab"},       // Embeded Unicode SOS - Start of string
	{"abc\xc2", "abc"},       // Invalid trailing 0xC2 dropped
	{"abc\x7fd", "abd"},      // DEL not BS
	{"ææææ\b\b\b\b", ""},     //
	{"\x85", ""},             // ASCII C1 control code - NEL, Next line
	{"\x8F\x8F\x8F\x8F", ""}, // Invalid UTF-8 sequence (no non-leading 10xxxxxx)
}

func TestClean(t *testing.T) {
	for i, test := range cleanData {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			have := []byte(test.data)
			clean(&have)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("Have: %q Want %q", have, test.want)
			}
			for _, i := range have[len(have):cap(have)] {
				if i != '\x00' {
					t.Errorf("Have garbage: %q", have[0:cap(have)])
				}
			}
		})
	}

}

func BenchmarkClean(b *testing.B) {
	var have []byte
	for i, test := range cleanData {
		b.Run(fmt.Sprintf("Bench %d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				have = []byte(test.data)
				clean(&have)
			}
		})
	}
}
