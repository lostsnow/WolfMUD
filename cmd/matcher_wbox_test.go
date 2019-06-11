// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"
	"testing"
)

func TestLeadingDigits(t *testing.T) {
	for _, test := range []struct {
		word    string
		wantInt int
		wantLen int
	}{
		// Valid
		{"0", 0, 1},
		{"1", 1, 1},
		{"2", 2, 1},
		{"3", 3, 1},
		{"4", 4, 1},
		{"5", 5, 1},
		{"6", 6, 1},
		{"7", 7, 1},
		{"8", 8, 1},
		{"9", 9, 1},
		{"10", 10, 2},
		{"100", 100, 3},

		// Invalid
		{"-1", 0, 0},
		{"a", 0, 0},
		{"/", 0, 0}, // ASCII '0' - 1
		{":", 0, 0}, // ASCII '9' + 1
		{"SHORT", 0, 0},
	} {
		t.Run(test.word, func(t *testing.T) {
			haveInt, haveLen := leadingDigits(test.word)
			if haveInt != test.wantInt || haveLen != test.wantLen {
				t.Errorf("have: %d (length %d), want: %d (length %d)",
					haveInt, test.wantInt, haveLen, test.wantInt,
				)
			}
		})
	}
}

func BenchmarkLeadingDigits(b *testing.B) {

	for _, test := range []string{
		"1",
		"12",
		"123",
		"1234",
		"12345",
	} {
		b.Run(test, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = leadingDigits(test)
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
		{"-", -1, -1},
		{"a-b", -1, -1},
		{"1-a", -1, -1},
		{"1-1a", -1, -1},
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
