// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package decode

import (
	"fmt"
	"testing"
)

func TestIndexSeparator(t *testing.T) {
	for _, test := range []struct {
		data      string
		wantIndex int
		wantSize  int
	}{
		{"", 0, 0},
		{"noseparator", 11, 0},
		{" noseparator", 12, 0},
		{"\tnoseparator", 12, 0},
		{"\nnoseparator", 12, 0},
		{"a:z", 1, 1},
		{" a:z", 2, 1},
		{"\ta:z", 2, 1},
		{"\na:z", 2, 1},
		{"a→z", 1, 3},
		{" a→z", 2, 3},
		{"\ta→z", 2, 3},
		{"\na→z", 2, 3},
		{"a b→z", 3, 3},
		{" a b→z", 4, 3},
		{"\ta b→z", 4, 3},
		{"\na b→z", 4, 3},
		{"a b →z", 4, 3},
		{" a b →z", 5, 3},
		{"\ta b →z", 5, 3},
		{"\na b →z", 5, 3},
		{"a→z:y", 1, 3},
		{" a→z:y", 2, 3},
		{"\ta→z:y", 2, 3},
		{"\na→z:y", 2, 3},
		{"Χαίρετε→hello", 14, 3},
		{"X-Y→hello", 3, 3},
		{"X_Y→hello", 3, 3},
		{"X→Y-hello", 1, 3},
		{"X→Y_hello", 1, 3},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			haveIndex, haveSize := indexSeparator([]byte(test.data))
			if (haveIndex != test.wantIndex) || (haveSize != test.wantSize) {
				t.Errorf("\nhave %d, %d\nwant %d, %d",
					haveIndex, haveSize, test.wantIndex, test.wantSize,
				)
				return
			}
		})
	}
}

func BenchmarkIndexSeparator(b *testing.B) {
	for _, test := range []struct {
		name string
		data string
	}{
		{"Exit", "E→L4"},
		{"Veto", "GET→The rock seems quite immovable."},
		{"Invalid", "Invalid"},
		{"_Body", "UPPER_ARM→2"},
		{"-Body", "UPPER-ARM→2"},
	} {
		data := []byte(test.data)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = indexSeparator(data)
			}
		})
	}
}
