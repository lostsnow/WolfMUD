// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

import (
	"code.wolfmud.org/WolfMUD.git/text"
)

// stopWords is the list of words considered to be stop words.
var stopWords = text.Dictionary(
	"a",
	"an",
	"from",
	"in",
	"into",
	"of",
	"out",
	"some",
	"the",
)

// removeStopWords takes a slice of strings (words) and returns a slice of
// strings with the words removed that are stop words. For example:
//
//	take the apple from the bag
//
// becomes:
//
//	take apple bag
//
func RemoveStopWords(in []string) (out []string) {
	for _, word := range in {
		if stopWords.Contains(word) {
			continue
		}
		out = append(out, word)
	}
	return
}
