// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package units

import (
	"strconv"
	"testing"
	. "wolfmud.org/utils/test"
)

type testData struct {
	weight      int
	description string
}

var testSubjects = []*testData{
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
	{1000, "62lb"},
}

func TestInt(t *testing.T) {
	for _, s := range testSubjects {
		subject := Weight(s.weight)
		Equal(t, "Int", s.weight, subject.Int())
	}
}

func TestStringer(t *testing.T) {
	for _, s := range testSubjects {
		subject := Weight(s.weight)
		Equal(t, "Stringer with "+strconv.Itoa(s.weight), s.description, subject.String())
	}
}
