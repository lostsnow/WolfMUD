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

// MatchLimit tries to find Things in the given Inventory by matching aliases
// and qualifiers in the given word list. The limit is the number of groups of
// words to match. A group of words consisting of an alias plus zero or more
// qualifiers. A limit of -1 will consume all words in the list.
//
// Assume the following items, aliases and qualifiers in square brackets and
// qualifiers have a leading '+' symbol:
//
//  a small green ball  [+SMALL +GREEN BALL]
//  a small red ball    [+SMALL +RED BALL]
//  a large green ball  [+LARGE +GREEN BALL]
//  a large red ball    [+LARGE +RED BALL]
//
// All matching processes the word list from right to left. This is because a
// group of words identifying one or more items always ends with an alias and
// can be preceded by one or more qualifiers. By default only the first item
// matching a word group will be returned.
//
// With a word list of 'RED BALL GREEN BALL' a limit of 1 would consume two
// words 'GREEN BALL' and return one Thing matched: a small green ball. In this
// case the alias 'BALL' would match, followed by the qualifier 'GREEN'. As
// 'BALL' preceding 'GREEN' is not a qualifier matching fails and we have
// 'GREEN BALL' as the first matching group of words.
//
// With a word list of 'RED BALL GREEN BALL' a limit of 2 (or -1) would consume
// all of the words and return two Thing matched: a small green ball, a small
// red ball. Note that the limit of 2 if the number of word groups, in this
// case 'RED BALL' and 'GREEN BALL' not the number of items to return.
//
// Within a matching group of words the first qualifier may be special. The
// special qualifiers are:
//
//  ALL - include all items matching the current group
//  n - include up to n items matching the current group
//  nth - include only the nth item (postfix may be ST, ND, RD or TH)
//  n-N - include only matching items n to N for the current group
//
// With a word list of 'ALL GREEN BALL' all items matching the alias of 'BALL'
// with a qualifier of 'GREEN' will be returned. In this case: a small green
// ball, a large green ball.
//
// With a word list of '2 BALL' the first 2 items matching the alias of 'BALL'
// will be returned. In this case: a small green ball, a small red ball.
//
// With a word list of '2ND RED BALL' only the second item matching the alias
// of 'BALL' with a qualifier of 'RED' will be returned. In this case: a large
// red ball.
//
// With a word list of '2-3 BALL' items 2 through to 3 matching the alias of
// 'BALL' would be returned. In this case: a small red ball, a large green
// ball.
//
// Duplicate matching items will be removed from the results. With a word list
// of 'ALL GREEN BALL 2-3 BALL' then 'ALL GREEN BALL' would match: a small green
// ball, a large green ball and '2-3 BALL' would match: a small red ball, a
// large green ball. The results returned would be: a small green ball, a small
// red ball, a large green ball.
//
// When matching an alias and zero or more qualifiers the first non-match will
// end the group. Given the word list 'RED BALL GREEN BALL', after identifying
// the alias 'BALL' and qualifier 'GREEN' the word 'BALL' fails to match as a
// qualifier. Therefore the first matching word group is 'GREEN BALL'. Matching
// then starts again to look for an alias for 'BALL' with a qualifier of 'RED'.
//
// If a word fails to match as an alias, or fails to match as a qualifier after
// an alias match, it is added to the returned unknowns list. Consecutive
// unknown words are grouped together. Given the word list 'GREEN BALL BLUE
// FROG RED BALL', after identifying 'RED BALL' the word 'FROG' is not matched
// and added to the unknowns. Then 'BLUE' is also unmatched and the unknowns
// become 'BLUE FROG' because the unknowns are consecutive. Given the word list
// 'TREE GREEN BALL FROG RED BALL', then 'FROG' and 'TREE' would be unmatched.
// As they are not consecutive - 'GREEN BALL' would match between 'TREE' and
// 'FROG' two unknowns would be returned 'TREE' and 'FROG'.
//
// There may be some instances where different matches need to be carried out
// against different Inventory. For example the PUT command. The PUT command
// takes the form: PUT <item ...> <container>. The container may be carried or
// at the current location, while items to put in it should be carried. In this
// instance the container can be checked by calling Match, the items can then
// be found by calling MatchAll with the word list returned by the initial
// Match to complete the matching. In this way different Inventory can be
// passed to Match and MatchAll.
func MatchLimit(wordList []string, limit int, inv ...[]has.Thing) (matches []has.Thing, unknowns, words []string) {

	// Get a working list of the words in reverse order
	words = make([]string, len(wordList))
	for x, word := range wordList {
		words[len(wordList)-1-x] = word
	}

	// unknownSet is the list of unknown matches, consecutive unknown matches are
	// grouped together. lastMatchGood is a flag tracking whether the last run of
	// words matched at least one known Thing or not.
	unknownSet := [][]string{}
	lastMatchGood := false

	for len(words) > 0 && (limit == -1 || len(matches) < limit) {
		found := []has.Thing{}
		minLimit, maxLimit, maxWordsMatched := 0, 1, 0

		for _, i := range inv {
			for _, t := range i {
				wordsMatched := 0

				a := attr.FindAlias(t) // Ignore thing if no aliases
				if !a.Found() {
					continue
				}

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
							match := 0
							minLimit, maxLimit, match = specialQualifier(word)
							wordsMatched += match
						}
						break
					}

					wordsMatched++
				}

				if wordsMatched != 0 && wordsMatched >= maxWordsMatched {
					if wordsMatched > maxWordsMatched {
						found = found[:0]
					}
					maxWordsMatched = wordsMatched
					found = append([]has.Thing{t}, found...)
				}
			}
		}

		// If matching fails append word to current unknown word group, starting a
		// new group if previous match was good.
		if maxWordsMatched == 0 {
			if lastMatchGood == true || len(unknownSet) == 0 {
				unknownSet = append([][]string{[]string{}}, unknownSet...)
			}
			unknownSet[0] = append([]string{words[0]}, unknownSet[0]...)
			lastMatchGood = false
			words = words[1:]
			continue
		}

		lastMatchGood = true

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
		words = words[maxWordsMatched:]

	}

	// Condense groups of unknown words into simple list
	for _, unknown := range unknownSet {
		unknowns = append(unknowns, strings.Join(unknown, " "))
	}

	// Put remaining words into the correct order
	words = wordList[:len(words)]

	return
}

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

// specialQualifier takes a word and tries to process it as a special
// qualifier. Special qualifiers can be of the form:
//
//   ALL - all items
//   n - limit matches to 0-n
//   nP - specific instance 1st, 2nd, rd, 4th, etc
//   n-N - a range of items from n to N
//
// specialQualifier will return the new minimum and maximum limits. It will
// also return match=0 if there is no match or match=1 if there is a match.
func specialQualifier(word string) (minLimit, maxLimit, match int) {

	// Set default for limit 0-1 and no matches
	minLimit, maxLimit, match = 0, 1, 0

	// Special qualifier for all matches?
	if word == "ALL" {
		minLimit, maxLimit, match = 0, 0, 1
		return
	}

	// If qualifier has no leading digits just return, not special
	split := lastLeadingDigit(word)
	if split == -1 {
		return
	}

	// Move to start of postfix after initial int
	split++

	// If digits cannot be parsed as an int just return, not special
	n, err := strconv.Atoi(word[:split])
	if err != nil {
		return
	}

	// Get postfix for a specific instance or range
	post := word[split:]

	// Just a number is a limit from 0-n items
	if post == "" {
		minLimit, maxLimit, match = 0, n, 1
		return
	}

	// Number with postfix is a specific instance such as 1st, 2nd
	if post == "ST" || post == "ND" || post == "RD" || post == "TH" {
		minLimit, maxLimit, match = n, n, 1
		return
	}

	// Number followed by a hyphen and a number (n-N) is a range
	if strings.HasPrefix(post, "-") {
		N, err := strconv.Atoi(post[1:])

		// Return if digits after hyphen cannot be parsed as an int
		if err != nil {
			return
		}

		if N > n {
			minLimit, maxLimit, match = n, N, 1
		} else {
			minLimit, maxLimit, match = N, n, 1
		}
		return
	}

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
