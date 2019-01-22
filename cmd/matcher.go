// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strconv"
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Match is shorthand for MatchLimit with a limit of 1.
func Match(wordList []string, inv ...[]has.Thing) (matches []has.Thing, unknowns, words []string) {
	matches, unknowns, words = MatchLimit(wordList, 1, inv...)
	return
}

// MatchAll is shorthand for MatchLimit with a limit of -1.
func MatchAll(wordList []string, inv ...[]has.Thing) (matches []has.Thing, unknowns, words []string) {
	matches, unknowns, words = MatchLimit(wordList, -1, inv...)
	return
}

// MatchLimit tries to find Things in the given inventories by matching
// aliases and qualifiers in the given word list. If a limit of -1 is specified
// MatchLimit will attempt to consume all of the words in the word list. If a
// specific limit X is given only enough words will be consumed to identify X
// sets of matches. Each set of matches will contain the Things the consumed
// words matched. It should also be noted that words are consumed from the list
// right to left. So given a word list of "apple ball cat", cat will be
// searched for first, then ball then apple.
//
// An example, given "large ball small ball" as the word list. A limit of 1
// could consume two words if "small ball" was found to match some Things. The
// returned matches would include all Things that matched "small ball".
//
// Given the same list with a limit of -1 all words would be consumed and the
// returned matches would contain all the Things matching "large ball" or
// "small ball".
//
// If the word list in the example was replaced with "ball" the matches
// returned would include all large and small balls found even though only one
// word in the word would be consumed.
//
// Words not consumed are returned so that they may be passed to consecutive
// searches for further processing.
//
// For example in the command "PUT <items...> <container>" you would call
// Match to find the container, then MatchAll to find the remaining items. If
// you wanted to limit the PUT command to "PUT <item> <container>" you would
// call Match to find the container, then Match to find the item. If the word
// list is not zero at the end then too many items have been specified.
// Likewise if the list of matches is not one then more than one container/item
// has been found.
func MatchLimit(wordList []string, limit int, inv ...[]has.Thing) (matches []has.Thing, unknowns, words []string) {

	// Get a working list of the words in reverse order
	words = make([]string, len(wordList))
	for x, word := range wordList {
		words[len(wordList)-1-x] = word
	}

	// unknownSet is the list of unknown matches, consecutive unknown matches are
	// grouped together. thingMatched is a flag tracking whether the last run of
	// words matched at least one known Thing or not.
	unknownSet := [][]string{}
	thingMatched := false
	for len(words) > 0 && (limit == -1 || len(matches) < limit) {
		maxMatch := 0
		found := []has.Thing{}
		minLimit, maxLimit := 0, 1

		for _, i := range inv {
			for _, t := range i {

				a := attr.FindAlias(t) // Ignore thing if no aliases
				if !a.Found() {
					continue
				}

				numMatch := 0

				// Try and match alias then as many qualifiers as possible
			wordLoop:
				for x, word := range words {

					// If current word already seen in this run of words stop looking for
					// further matches
					for _, seen := range words[:x] {
						if word == seen {
							break wordLoop
						}
					}

					if (x == 0 && !a.HasAlias(word)) || (x > 0 && !a.HasAlias("+"+word)) {
						if x != 0 {

							// Special qualifier for all matches?
							if word == "ALL" {
								maxLimit = 0
								numMatch++
							}
							// Could word be a special numeric qualifier?
							if split := lastLeadingDigit(word); split != -1 {
								split++
								if n, err := strconv.Atoi(word[:split]); err == nil {
									post := word[split:]
									// Just a number is a limit from 0-n items
									if post == "" {
										maxLimit = n
										numMatch++
									}
									// Number with postfix is a specific instance such as 1st, 2nd
									if post == "ST" || post == "ND" || post == "RD" || post == "TH" {
										minLimit, maxLimit = n, n
										numMatch++
									}
									// Number followed by a hyphen and a number (n-N) is a range
									if strings.HasPrefix(post, "-") {
										if N, err := strconv.Atoi(post[1:]); err == nil {
											if N > n {
												minLimit, maxLimit = n, N
											} else {
												minLimit, maxLimit = N, n
											}
											numMatch++
										}
									}
								}
							}

						}
						break
					}

					numMatch++
				}

				if numMatch != 0 && numMatch >= maxMatch {
					if numMatch > maxMatch {
						found = found[:0]
					}
					maxMatch = numMatch
					found = append([]has.Thing{t}, found...)
				}
			}
		}

		// If matching fails append word to current unknown word group, starting a
		// new group if previous match was good.
		if maxMatch == 0 {
			if thingMatched == true || len(unknownSet) == 0 {
				unknownSet = append([][]string{[]string{}}, unknownSet...)
			}
			unknownSet[0] = append([]string{words[0]}, unknownSet[0]...)
			thingMatched = false
			words = words[1:]
			continue
		}

		thingMatched = true

		// Append any new good matches ignoring anything already in the list of
		// matches.
	uniqueLoop:
		for x, f := range found {
			min := len(found) - minLimit
			max := len(found) - maxLimit
			if maxLimit > 0 && (x > min || x < max) {
				continue
			}
			for _, m := range matches {
				if f == m {
					continue uniqueLoop
				}
			}
			matches = append([]has.Thing{f}, matches...)
		}
		words = words[maxMatch:]

	}

	// Condense groups of unknown words into simple list
	for _, unknown := range unknownSet {
		unknowns = append(unknowns, strings.Join(unknown, " "))
	}

	// Return words in the correct order
	words = wordList[:len(words)]

	return
}

// lastLeadingDigit returns the position of the last leading digit or -1 if
// there are no leading digits.
func lastLeadingDigit(s string) int {
	for x, c := range s {
		if c < '0' || c > '9' {
			return x - 1
		}
	}
	return len(s) - 1
}
