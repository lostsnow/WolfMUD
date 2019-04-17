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

// Result is the result of a match. If the match is successful then Thing will
// be the matched Thing and Unknown and NotEnough will both be empty strings.
// If the match fails Thing will be nil. If the match fails due to unknown
// words then Unknown will be set to the unknown words. If the match fails
// because there are not enough matching Thing to satisfy a limit then
// NotEnough will be set to the words matched.
type Result struct {
	has.Thing
	Unknown   string
	NotEnough string
}

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
func MatchLimit(wordList []string, limit int, inv ...[]has.Thing) (matches []Result, words []string) {

	// Get a working list of the words in reverse order. We start at the end and
	// work backwards through the words because we know the last word has to be
	// an alias and not a qualifier.
	words = make([]string, len(wordList))
	for x, word := range wordList {
		words[len(wordList)-1-x] = word
	}

	count := 0

	for len(words) > 0 && (limit == -1 || count < limit) {
		count++

		// Loop through Inventory items for those matching alias
		results := []has.Thing{}
		for _, i := range inv {
			for _, t := range i {
				if attr.FindAlias(t).HasAlias(words[0]) {
					results = append(results, t)
				}
			}
		}

		// If no matched aliases add to unknown, consume word, try next word
		if len(results) == 0 {
			if len(matches) == 0 || matches[0].Thing != nil {
				matches = append([]Result{Result{nil, words[0], ""}}, matches...)
			} else {
				matches[0].Unknown = words[0] + " " + matches[0].Unknown
			}
			words = words[1:]
			continue
		}

		// Record word just seen, consume it and flag alias just matched
		wordsSeen := []string{words[0]}
		words = words[1:]

		// Match qualifiers against alias matches until no results left or we run
		// out of words for qualifiers
	qualifierLoop:
		for len(results) > 0 && len(words) > 0 {

			// If current word already seen stop looking for more qualifiers
			for _, seen := range wordsSeen {
				if seen == words[0] {
					break qualifierLoop
				}
			}

			// Loop through matched set looking for matching qualifiers
			subResults := []has.Thing{}
			for _, t := range results {
				if a := attr.FindAlias(t); a.Found() {
					if a.HasQualifierForAlias(wordsSeen[0], words[0]) || a.HasQualifier(words[0]) {
						subResults = append(subResults, t)
					}
				}
			}

			// If no sub-matches left stop checking and don't consume word
			if len(subResults) == 0 {
				break
			}

			// We have matches so record word just seen, consume it and use
			// sub-results as new results
			wordsSeen = append(wordsSeen, words[0])
			words = words[1:]
			results = subResults
		}

		// Set default limits to be first item in results only
		minLimit, maxLimit := 0, 1

		// Check if last word not matched is special. If it is set limits for
		// results.
		if len(words) > 0 {
			min, max := specialQualifier(words[0])
			if !(min == -1 && max == -1) {

				words = words[1:] // Consume special qualifier
				minLimit, maxLimit = min, max

				if minLimit < 0 {
					minLimit = 0
				}

				if maxLimit == -1 || maxLimit > len(results) {
					maxLimit = len(results)
				}
			}
		}

		// If minimum limit beyond result range there is no way to have any matches
		if minLimit > len(results) {
			matches = append([]Result{Result{nil, "", strings.Join(wordsSeen, " ")}}, matches...)
			continue
		}

		// Subset final results by limits
		results = results[minLimit:maxLimit]

		if len(results) == 0 {
			matches = append([]Result{Result{nil, "", strings.Join(wordsSeen, " ")}}, matches...)
			continue
		}

		// Add new results only, in reverse order, to current matches. Order is
		// reversed because we are walking backwards through the words, but for
		// sub-matches they should be added in order.
	uniqueLoop:
		for x := len(results) - 1; x >= 0; x-- {
			for _, m := range matches {
				if m == results[x] {
					continue uniqueLoop
				}
			}
			matches = append([]Result{Result{results[x], "", ""}}, matches...)
		}
	}

	// Put remaining words into the correct order
	words = wordList[:len(words)]

	return
}

// Match is shorthand for MatchLimit with a limit of 1.
func Match(wordList []string, inv ...[]has.Thing) (matches []Result, words []string) {
	matches, words = MatchLimit(wordList, 1, inv...)
	return
}

// MatchAll is shorthand for MatchLimit with a limit of -1. It always consumes
// all words and returns no unused words.
func MatchAll(wordList []string, inv ...[]has.Thing) (matches []Result) {
	matches, _ = MatchLimit(wordList, -1, inv...)
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
// specialQualifier will return the new minimum and maximum limits. If the word
// is not a special qualifier then the returned limits will both be -1.
// Note that the minimum limit will be zero based and hence 1 less than might
// be expected. However the limits can be used directly for slicing, as in
// slice[minLimit:maxLimit].
func specialQualifier(word string) (minLimit, maxLimit int) {

	// Set default for no matches
	minLimit, maxLimit = -1, -1

	// Special qualifier for all matches?
	if word == "ALL" {
		minLimit, maxLimit = 0, -1
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

	// Just a number is a limit from 0 for n items
	if post == "" {
		minLimit, maxLimit = 0, n
		return
	}

	// Number with postfix is a specific instance such as 1st, 2nd
	if post == "ST" || post == "ND" || post == "RD" || post == "TH" {
		minLimit, maxLimit = n-1, n
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
			minLimit, maxLimit = n-1, N
		} else {
			minLimit, maxLimit = N-1, n
		}
		return
	}

	return
}

// lastLeadingDigit returns the position of the last leading digit or -1 if
// there are no leading digits.
func lastLeadingDigit(s string) int {
	for x, c := range s {
		if c < '0' || '9' < c {
			return x - 1
		}
	}
	return len(s) - 1
}
