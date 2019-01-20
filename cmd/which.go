// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
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

// Syntax: WHICH item...
func init() {
	addHandler(which{}, "WHICH")
}

type which cmd

func (w which) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendGood("You look around for nothing in particular.")
		return
	}

	// Find items either being carried or at location
	matches, unknowns, _ := SearchAll(
		s.words,
		attr.FindInventory(s.actor).Contents(),
		s.where.Everything(),
	)

	s.msg.Actor.SendGood("You look around.")
	for _, m := range matches {
		if attr.FindLocate(m).Where() == s.where {
			s.msg.Actor.Append("\nYou see ", attr.FindName(m).Name("something"), " here.")
		} else {
			s.msg.Actor.Append("\nYou are carrying ", attr.FindName(m).Name("something"), ".")
		}
	}

	if len(unknowns) > 0 {
		for _, unknown := range unknowns {
			s.msg.Actor.SendBad("You see no '", unknown, "' here.")
		}
	}

	s.ok = true
}

// Search is shorthand for SearchLimit with a limit of 1.
func Search(wordList []string, inv ...[]has.Thing) (matches []has.Thing, unknowns, words []string) {
	matches, unknowns, words = SearchLimit(wordList, 1, inv...)
	return
}

// SearchAll is shorthand for SearchLimit with a limit of -1.
func SearchAll(wordList []string, inv ...[]has.Thing) (matches []has.Thing, unknowns, words []string) {
	matches, unknowns, words = SearchLimit(wordList, -1, inv...)
	return
}

// SearchLimit tries to find Things in the given inventories by matching
// aliases and qualifiers in the given word list. If a limit of -1 is specified
// SearchLimit will attempt to consume all of the words in the word list. If a
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
// Search to find the container, then SearchAll to find the remaining items. If
// you wanted to limit the PUT command to "PUT <item> <container>" you would
// call Search to find the container, then Search to find the item. If the word
// list is not zero at the end then too many items have been specified.
// Likewise if the list of matches is not one then more than one container/item
// has been found.
func SearchLimit(wordList []string, limit int, inv ...[]has.Thing) (matches []has.Thing, unknowns, words []string) {

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
		maxInstance := 0
		anInstance := 0

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
							// If qualifier not found can it be used as a count?
							if n, err := strconv.Atoi(word); err == nil {
								maxInstance = n
								numMatch++
							}

							// If qualifier not found and not a count can it be used for a
							// specific instance such as 1st, 2nd, 3rd...
							if split := strings.LastIndexAny(word, "0123456789"); split != -1 {
								split++
								if n, err := strconv.Atoi(word[:split]); err == nil {
									post := word[split:]
									if post == "ST" || post == "ND" || post == "RD" || post == "TH" {
										anInstance = n
										numMatch++
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
					found = append(found, t)
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
			if maxInstance > 0 && x < (len(found)-maxInstance) {
				continue
			}
			if anInstance > 0 && x != (len(found)-anInstance) {
				continue
			}
			for _, m := range matches {
				if f == m {
					break uniqueLoop
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
