// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package units

import (
	"testing"
)

var testSubjects = []struct {
	weight      int
	description string
}{
	{0, "nothing"},
	{1, "1oz"},
	{6, "6oz"},
	{7, "7oz"},
	{8, "half a pound"},
	{9, "9oz"},
	{16, "a pound"},
	{17, "1lb and 1oz"},
	{18, "1lb and 2oz"},
	{23, "1lb and 7oz"},
	{24, "1lb and 8oz"},
	{25, "1lb and 9oz"},
	{31, "1lb and 15oz"},
	{32, "2lb"},
	{33, "2lb"},
	{39, "2lb"},
	{40, "2lb"},
	{41, "3lb"},
	{64, "4lb"},
	{65, "4lb"},
	{224, "14lb"}, // 1 stone
	{448, "28lb"}, // 1 quarter
	{1000, "62lb"},
	{1792, "112lb"},   // Hundredweight
	{35840, "2240lb"}, // Ton
}

func TestInt(t *testing.T) {
	for i, s := range testSubjects {
		have := Weight(s.weight).Int()
		want := s.weight
		if have != want {
			t.Errorf("Invalid weight value: Case %d, have %d wanted %d", i, have, want)
		}
	}
}

func TestStringer(t *testing.T) {
	for i, s := range testSubjects {
		have := Weight(s.weight).String()
		want := s.description
		if have != want {
			t.Errorf("Invalid weight string: Case %d, have %v wanted %v", i, have, want)
		}
	}
}
