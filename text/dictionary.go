// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"strings"
	"sync"
)

// dictionary procides a convenient list of strings that can be checked to see
// if it contains a specific string. Checking is case insensative. A dictionary
// is basically a map with some helpers. A dictionary is typically created once
// and checked multiple times. A dictionary is safe to use across multiple
// goroutines.
type dictionary struct {
	sync.RWMutex
	words map[string]struct{}
}

// Dictionary returns a new dictionary containing the specified strings. The
// dictionary can then be checked to see if it contains a specific string (case
// insensitive) by calling the Contains method.
//
// NOTE: The strings in the dictionary are converted to uppercase.
func Dictionary(s ...string) (d *dictionary) {
	d = &dictionary{words: map[string]struct{}{}}
	d.Add(s...)
	return
}

// Add adds one or more words to the dictionary.
func (d *dictionary) Add(s ...string) {
	d.Lock()
	for _, s := range s {
		d.words[strings.ToUpper(s)] = struct{}{}
	}
	d.Unlock()
	return
}

// Delete deletes one or more words from the dictionary.
func (d *dictionary) Delete(s ...string) {
	d.Lock()
	for _, s := range s {
		delete(d.words, strings.ToUpper(s))
	}
	d.Unlock()
	return
}

// Contains searches the dictionary to see if it contains the specified string.
// If a match is found true is returned, otherwise false. The search is case
// insensitive.
func (d *dictionary) Contains(s string) (found bool) {
	d.RLock()
	_, found = d.words[strings.ToUpper(s)]
	d.RUnlock()
	return
}
