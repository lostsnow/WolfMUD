// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text_test

import (
	"fmt"
	"testing"

	"code.wolfmud.org/WolfMUD.git/text"
)

func TestList(t *testing.T) {
	for _, test := range []struct {
		data []string
		want string
	}{
		{[]string{}, ""},
		{[]string{""}, ""},
		{[]string{"A"}, "A"},
		{[]string{"A", "B"}, "A and B"},
		{[]string{"A", "B", "C"}, "A, B and C"},
	} {
		have := text.List(test.data)
		if have != test.want {
			t.Errorf("\nhave %+q\nwant %+q", have, test.want)
		}
	}
}

func BenchmarkList(b *testing.B) {
	for _, data := range [][]string{
		[]string{},
		[]string{""},
		[]string{"A"},
		[]string{"A", "B"},
		[]string{"A", "B", "C"},
		[]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"},
		[]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"},
	} {
		b.Run(fmt.Sprintf("List %v", data), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = text.List(data)
			}
		})
	}
}
