// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"
	"testing"
)

func TestLastLeadingDigit(t *testing.T) {
	for _, test := range []struct {
		word string
		want int
	}{
		// Valid
		{"0", 0},
		{"1", 0},
		{"2", 0},
		{"3", 0},
		{"4", 0},
		{"5", 0},
		{"6", 0},
		{"7", 0},
		{"8", 0},
		{"9", 0},
		{"10", 1},
		{"100", 2},

		// Invalid
		{"-1", -1},
		{"a", -1},
		{"/", -1}, // ASCII '0' - 1
		{":", -1}, // ASCII '9' + 1
	} {
		t.Run(test.word, func(t *testing.T) {
			have := lastLeadingDigit(test.word)
			if have != test.want {
				t.Errorf("have: %d, want: %d", have, test.want)
			}
		})
	}
}

func BenchmarkLastLeadingDigit(b *testing.B) {

	for _, test := range []string{
		"1",
		"12",
		"123",
		"1234",
		"12345",
	} {
		b.Run(test, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = lastLeadingDigit(test)
			}
		})
	}
}

func TestSpecialQualifier(t *testing.T) {
	for _, test := range []struct {
		word     string
		minLimit int
		maxLimit int
	}{
		// Valid
		{"all", 0, -1},
		{"0", 0, 0}, // Should this be invalid (-1,-1) ?
		{"1", 0, 1},
		{"2", 0, 2},
		{"3", 0, 3},
		{"01", 0, 1},
		{"1st", 0, 1},
		{"2nd", 1, 2},
		{"3rd", 2, 3},
		{"4th", 3, 4},
		{"1-2", 0, 2},
		{"2-3", 1, 3},
		{"2-1", 0, 2},
		{"3-2", 1, 3},

		// A valid 'invalid'
		{"", -1, -1},

		// Invalid
		{"a", -1, -1},
		{"st", -1, -1},
		{"ast", -1, -1},
		{"1xx", -1, -1},
		{"a-b", -1, -1},
		{"1.0", -1, -1},
		{"1.0-", -1, -1},
		{"1.0-2.0", -1, -1},
		{"2+3", -1, -1},

		// Should these be allowed as meaning from n to end and start to n?
		{"0-", -1, -1},
		{"1-", -1, -1},
		{"2-", -1, -1},
		{"-0", -1, -1},
		{"-1", -1, -1},
		{"-2", -1, -1},
	} {
		t.Run(test.word, func(t *testing.T) {
			word := strings.ToUpper(test.word)
			haveMinLimit, haveMaxLimit := specialQualifier(word)
			if haveMinLimit != test.minLimit || haveMaxLimit != test.maxLimit {
				t.Errorf("\nhave: %d,%d\nwant: %d,%d",
					haveMinLimit, haveMaxLimit, test.minLimit, test.maxLimit,
				)
			}
		})
	}
}

func BenchmarkSpecialQualifier(b *testing.B) {

	for _, test := range []string{
		"ALL",
		"1",
		"12",
		"123456",
		"1-2",
		"12-34",
		"1234-5678",
		"SHORT", // Not a special qualifier
	} {
		b.Run(test, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = specialQualifier(test)
			}
		})
	}
}
