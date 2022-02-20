// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"math/rand"
	"strconv"
	"strings"
)

// Match attempts to itentify items in the passed Thing's inventory by matching
// the passed list of words with the item aliases and qualifiers. Match returns
// a slice containing both good and bad matches. Good matches are always single
// words containing the UID of the matched item and always start with '#UID-'.
// Bad matches contain one or more of the input words that could not be
// matched.
//
// For example, assuming we have one red ball, one green ball and two blue
// balls then in the following:
//
//	Match(
//		[]string{"RED", "BALL", "GREEN", "FROG", "ALL", "BLUE", "BALL"}),
//		s.actor,
//	)
//
// Might return:
//
//	[]string{"#UID-106", "GREEN FROG", "#UID-10A", "#UID-10E"}
//
// By default, when multiple items match an alias or qualifier+alias, Match
// will return the first matching item. For example, []string{"BLUE", "BALL"}
// would return the UID of the first blue ball matched. This can be changed
// with the use of special prefix modifiers.
//
// Special Modifiers
//
// A modifier may be used before a qualifier or alias to effect the matches
// that are returned. For example:
//
//	[]string{"ALL", "BALL"}
//	[]string{"ALL", "BLUE", "BALL"}
//
// The currently supported modifiers are:
//
//	ALL   - all matches
//	LAST  - the last item matched
//	ANY   - one random item matched
//	N/Nth - the Nth item matched (or last item if N > matches)
//
// NOTE: For performance reasons, if a thing being searched is considered
// crowded then we don't include everyone in the crowd in the search.
//
// TODO(diddymus): Add 'see also' pointing to docs/ files.
//
// BUG(diddymus): Nth does not care what the suffix is, 2nd and 2rd are both
// seen as valid modifiers.
//
// BUG(diddyus): Does not support ranges yet. For example 'which 2-4 ball' for
// the 2nd, 3rd and 4th balls.
func Match(words []string, where ...*Thing) (results []string) {
	results, _ = match(words, where, false)
	return
}

// LimitedMatch works in the same way as Match except that it stops after the
// first match/non-match and returns the unused, remaining input words. The
// first match/non-match will consume words from the end of the passed input
// word slice.
//
// NOTE: Although LimitedMatch stops after the first match/non-match it may
// produce multiple results. For example, if we have a red, a green and two
// blue balls then:
//
//  LimitedMatch(
//		[]string{"RED", "BALL", "ALL", "BLUE", "BALL"}),
//		s.actor,
//  )
//
// Would return one match, it being "all blue ball" with the results being the
// two blue balls. In this case LimitedMatch would return, for example:
//
//  []string{"#UID-10A", "#UID-10E"} and []string{"RED", "BALL"}
//
func LimitedMatch(words []string, where ...*Thing) (results, remaining []string) {
	return match(words, where, true)
}

// match implements the functionality for Match and LimitedMatch.
func match(words []string, where []*Thing, oneShot bool) ([]string, []string) {

	data := []*Thing{}
	for _, inv := range where {
		// For performance don't include all of the players if there is a crowd.
		if len(inv.Who) < cfg.crowdSize {
			data = append(data, inv.Who.Sort()...)
		}
		data = append(data, inv.In.Sort()...)
	}

	var (
		matches           = make([]*Thing, 0, len(data))
		subset            = make([]*Thing, 0, len(data))
		results           = make([]string, 0, len(data))
		alias, bound, nbr string
		pos, l            int
	)

	for pos = len(words) - 1; pos > -1; pos-- {

		// Filter items by alias
		alias, matches = words[pos], matches[:0]
		for _, item := range data {
			if item.As[DynamicAlias] == alias {
				matches = append(matches, item)
				continue
			}
			for _, a := range item.Any[Alias] {
				if a == alias {
					matches = append(matches, item)
					break
				}
			}
		}
		if len(matches) == 0 {
			if l = len(results) - 1; l == -1 || results[l][0] == '#' {
				results = append(results, alias)
				if oneShot {
					break
				}
			} else {
				results[l] = alias + " " + results[l]
			}
			continue
		}

		// Sub-filter items on qualifier
		alias = ":" + alias
		for pos--; pos > -1; pos-- {
			bound, subset = words[pos]+alias, subset[:0]
			for _, match := range matches {
				if match.As[DynamicQualifier] == words[pos] {
					subset = append(subset, match)
					continue
				}
				for _, qualifier := range match.Any[Qualifier] {
					if qualifier == words[pos] || qualifier == bound {
						subset = append(subset, match)
						break
					}
				}
			}
			// If no sub-matches backtrack so word is retried as an alias
			if len(subset) == 0 {
				pos++
				break
			}
			// Copy sub-matches then repeat and further qualify the items
			matches = matches[:copy(matches, subset)]
		}

		// Handle special modifiers
		switch {
		case len(matches) == 0:
			continue
		case pos < 1: // We can't have a qty
			results = append(results, matches[0].As[UID])
		case words[pos-1] == "ALL":
			for x := len(matches) - 1; x > -1; x-- {
				results = append(results, matches[x].As[UID])
			}
			pos--
		case words[pos-1] == "ANY":
			results = append(results, matches[rand.Intn(len(matches))].As[UID])
			pos--
		case words[pos-1] == "LAST":
			results = append(results, matches[len(matches)-1].As[UID])
			pos--
		case '1' <= words[pos-1][0] && words[pos-1][0] <= '9':
			nbr = words[pos-1]
			if l := len(nbr); l > 2 && strings.Contains("STNDRDTH", nbr[l-2:]) {
				nbr = nbr[:l-2]
			}
			if cnt, err := strconv.Atoi(nbr); err == nil {
				if cnt > len(matches) {
					cnt = len(matches)
				}
				results = append(results, matches[cnt-1].As[UID])
				pos--
			}
		default:
			results = append(results, matches[0].As[UID])
		}

		if oneShot {
			break
		}
	}

	// Reverse final results in-place, as we scan backwards for items
	l = len(results)
	x, y, ok := 0, 0, false
	for x := l/2 - 1; x >= 0; x-- {
		y = l - 1 - x
		results[x], results[y] = results[y], results[x]
	}

	// Filter out duplicate results in-place
	seen := make(map[string]struct{}, len(results))
	x = 0
	for _, result := range results {
		if _, ok = seen[result]; !ok {
			results[x] = result
			seen[result] = struct{}{}
			x++
		}
	}
	if pos < 0 {
		pos = 0
	}
	return results[:x], words[:pos]
}
