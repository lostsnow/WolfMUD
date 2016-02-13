// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

// stopWords is a map of valid stopwords as initialised by init.
var stopWords map[string]struct{}

// Initialise stop words
func init() {
	stopWords = make(map[string]struct{})
	for _, word := range []string{
		"a",
		"an",
		"from",
		"in",
		"into",
		"of",
		"out",
		"some",
		"the",
	} {
		stopWords[word] = struct{}{}
	}
}

// removeStopWords takes a slice of strings (words) and returns a slice of
// strings with the words removed that are stop words. For example:
//
//	take the apple from the bag
//
// becomes:
//
//	take orange bag
//
func RemoveStopWords(in []string) (out []string) {
	for _, word := range in {
		if _, match := stopWords[word]; !match {
			out = append(out, word)
		}
	}
	return
}
