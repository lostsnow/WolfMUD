// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"fmt"
	"testing"
)

var dictionaryCases = []struct {
	words []string
	count int
}{
	{
		[]string{"The", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}, 8,
	},
	{
		[]string{"The", "THE", "the"}, 1,
	},
	{
		[]string{"Café", "Straße", "☺☻☹", "123", ""}, 5,
	},
}

func TestAdd(t *testing.T) {
	for _, dc := range dictionaryCases {
		t.Run(fmt.Sprintf("Words %v", dc.words), func(t *testing.T) {
			d := Dictionary()
			for _, word := range dc.words {
				d.Add(word)
			}
			if have := len(d.words); have != dc.count {
				t.Errorf("\nWord count have: %d want: %d", have, dc.count)
			}
		})
	}
}

func TestContains(t *testing.T) {
	for _, dc := range dictionaryCases {
		t.Run(fmt.Sprintf("Words %v", dc.words), func(t *testing.T) {
			d := Dictionary(dc.words...)
			for _, word := range dc.words {
				if !d.Contains(word) {
					t.Errorf("\nWord missing want: %s", word)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	for _, dc := range dictionaryCases {
		t.Run(fmt.Sprintf("Words %v", dc.words), func(t *testing.T) {
			d := Dictionary(dc.words...)
			for _, word := range dc.words {
				d.Delete(word)
				if d.Contains(word) {
					t.Errorf("\nWord not deleted: %s", word)
				}
			}
		})
	}
}

func TestEmpty(t *testing.T) {
	d := Dictionary()
	if l := len(d.words); l != 0 {
		t.Errorf("\nNew dictionery not empty, len: %d", l)
	}
	d.Add()
	if l := len(d.words); l != 0 {
		t.Errorf("\nDictionery not empty after add, len: %d", l)
	}
	d.Delete()
	if l := len(d.words); l != 0 {
		t.Errorf("\nDictionery not empty after delete, len: %d", l)
	}
}
