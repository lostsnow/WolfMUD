// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Result is the result of a Match. If the match is successful then Thing will
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

// matcher is used to hold the current matcher state.
type matcher struct {
	words   []string
	alias   string
	things  []has.Thing
	aliases []has.Alias
}

// Match takes a list of alias and qualifier words, and lists of Things and
// returns a subset of the Things that match the aliases and qualifiers as a
// list of Results. See the Result type for details. Match will also return any
// unprocessed words. The word list may contain special limit qualifiers such
// as ALL, 3, 2nd or 2-4. See the specialQualifier function for details.
//
// For example, given the words: {"ALL", "SMALL", "BALL", "ALL", "GREEN", "BALL"}
//
// Match processes the words in reverse order. In this case starting with the
// last word 'BALL'.
//
// Match will try to identify Things in the list with an alias of 'BALL'.
//
// If there are no matches a Result of 'unknown' will be returned.
//
// If there are matches then the word 'BALL' will be consumed. The matches will
// then be reduced to those Things with a qualifier of 'GREEN' - which is now
// the last word in the list.
//
// If the matches cannot be reduced the original matches will be used. If the
// matches are reduced the qualifier 'GREEN' will be consumed.
//
// The matches will then be reduced by limits. The default limit is the first
// match. This may be changed by using a special qualifier such as ALL, 3, 2nd,
// etc. In this example we have 'ALL' so all of the current matches will be
// returned. If there are not enough remaining matches to satisfy the limits
// requested a result of 'not enough' will be returned. If a special qualifier
// is found it will be consumed and removed from the word list.
//
// Assuming we have at least one Thing that matches our example the words
// 'BALL', 'GREEN' and 'ALL' will have been consumed. The remaing, unprocessed
// words in the word list will be returned.
//
// The reaming words can then be used to a subsequent call to Match for further
// matching. See also MatchAll.
func Match(words []string, things ...[]has.Thing) ([]Result, []string) {

	if len(words) == 0 {
		return []Result{}, []string{}
	}

	m := newMatcher(words, things...)
	m.subsetAlias()

	// If no aliases match then return an 'unknown' Result
	if len(m.things) == 0 {
		return []Result{{nil, m.alias, ""}}, m.words
	}

	m.subsetQualifiers()
	m.subsetLimits()

	// If no items left in subset then return a 'not enough' Result
	if len(m.things) == 0 {
		return []Result{{nil, "", m.alias}}, m.words
	}

	return m.subsetAsResults(), m.words
}

// MatchAll repeatedly calls Match until all of the words are consumed. The
// returned Results will only contain unique Things.
func MatchAll(words []string, things ...[]has.Thing) (matches []Result) {

	if len(words) == 0 {
		return
	}

	var results []Result
	var r, m Result

	// Pre-flatten things so it's not done for every newMatcher in Match.
	// However, we do want newMatcher to make a copy still.
	for _, t := range things[1:] {
		things[0] = append(things[0], t...)
	}

	matches, words = Match(words, things[0])

	for len(words) > 0 {
		results, words = Match(words, things[0])
		r = results[0]
		m = matches[0]

		switch {

		// If 1st result Thing not nil merge all results with matches
		case r.Thing != nil:
			matches = mergeUniqueResults(matches, results)

		// If 1st result 'unknown' prepend to existing 'unknown'
		case r.Unknown != "" && m.Unknown != "":
			matches[0].Unknown = r.Unknown + " " + m.Unknown

		// Default is to prepend new error to current matches
		default:
			matches = append(matches, Result{})
			copy(matches[1:], matches[0:])
			matches[0] = r

		}
	}
	return
}

// newMatcher initialises a new matcher.
func newMatcher(words []string, things ...[]has.Thing) *matcher {

	s := 0
	for _, t := range things {
		s += len(t)
	}

	m := &matcher{words, "", make([]has.Thing, s), make([]has.Alias, s)}

	s = 0
	for _, t := range things {
		copy(m.things[s:], t)
		s += len(t)
	}

	return m
}

// nextWord returns the current word from the marcher word list.
func (m *matcher) nextWord() string {
	return m.words[len(m.words)-1]
}

// deleteWord removes the current word from the marcher word list.
func (m *matcher) deleteWord() {
	m.words = m.words[:len(m.words)-1]
}

// mergeUniqueResults takes a set of matches and a set of results, merges new,
// unique results into the matches and returns the new list.
func mergeUniqueResults(matches, results []Result) []Result {

	unique := results[:0]

uniqueLoop:
	for _, r := range results {
		for _, m := range matches {
			if m.Thing == r.Thing {
				continue uniqueLoop
			}
		}
		unique = append(unique, r)
	}
	return append(unique, matches...)
}

// subsetAlias takes a list of Things and returns a list of all of the Things
// matching the specified alias.
func (m *matcher) subsetAlias() {

	m.alias = m.nextWord()
	m.deleteWord()

	subset := m.things[:0]
	aliases := m.aliases[:0]
	for _, t := range m.things {
		if a := attr.FindAlias(t); a.HasAlias(m.alias) {
			subset = append(subset, t)
			aliases = append(aliases, a)
		}
	}

	m.things = m.things[:len(subset)]
	m.aliases = m.aliases[:len(aliases)]
}

// subsetQualifiers takes a list of Things and returns all of the Things
// matching the greatest number of matched qualifiers (taken from words). The
// alias passed is used to determine bound qualifiers. Qualifiers used will be
// removed from the list of words passed with the unmatched words returned.
func (m *matcher) subsetQualifiers() {

	if len(m.words) == 0 {
		return
	}

	qualifier := m.nextWord()
	subset := m.things[:0]
	aliases := m.aliases[:0]

	for x, a := range m.aliases {
		if a.HasQualifierForAlias(m.alias, qualifier) || a.HasQualifier(qualifier) {
			subset = append(subset, m.things[x])
			aliases = append(aliases, a)
		}
	}

	if len(subset) == 0 {
		return
	}

	m.deleteWord()
	m.things = m.things[:len(subset)]
	m.aliases = m.aliases[:len(aliases)]
	m.subsetQualifiers()
}

// subsetLimits takes a list of Things and returns a subset of the list within
// limits. If the last word can be interpreted as limits by specialQualifier
// then those limits will be used and the word consumed, if not a default limit
// of the first thing [0:1] will be used.
func (m *matcher) subsetLimits() {

	// Shortcut: if no limit qualifier posible just return first result
	if len(m.words) == 0 {
		m.things = m.things[:1]
		return
	}

	// Try to interpret special limit qualifier
	minLimit, maxLimit := specialQualifier(m.nextWord())

	// If special qualifier not found use default limit of [0:1]. Otherwise
	// consume word and make sure limits within bounds of things slice.
	if minLimit == -1 && maxLimit == -1 {
		minLimit, maxLimit = 0, 1
	} else {
		m.deleteWord()

		if minLimit < 0 {
			minLimit = 0
		}

		if maxLimit == -1 || maxLimit > len(m.things) {
			maxLimit = len(m.things)
		}
	}

	// If minimum limit is greater than the number of things there is no way to
	// have any matches.
	if minLimit > len(m.things) {
		minLimit, maxLimit = 0, 0
	}

	m.things = m.things[minLimit:maxLimit]
	return
}

// subsetAsResults takes a list of Things and returns them as a list of
// Results.
func (m *matcher) subsetAsResults() []Result {
	results := make([]Result, len(m.things))
	for x := range m.things {
		results[x].Thing = m.things[x]
	}
	return results
}

// specialQualifier takes a word and tries to process it as a special
// qualifier. Special qualifiers can be of the form:
//
//   ALL - all items
//   n - limit matches to 0-n
//   nP - specific instance 1st, 2nd, rd, 4th, etc
//   n-N - a range of items from n to N
//
// The values n and N are restricted to a maximum value of 9,999,999.
//
// specialQualifier will return the new minimum and maximum limits. If the word
// is not a special qualifier then the returned limits will both be -1.
// Note that the minimum limit will be zero based and hence 1 less than might
// be expected. If the maxLimit is -1 it should be treated as 'unbounded'.
//
// If the maxLimit is not -1 then the limits can be used directly for slicing,
// as in slice[minLimit:maxLimit].
func specialQualifier(word string) (minLimit, maxLimit int) {

	// Special qualifier for all matches?
	if word == "ALL" {
		return 0, -1
	}

	// If qualifier has no leading digits just return, not special
	n, l := leadingDigits(word)
	if l == 0 {
		return -1, -1
	}

	// A plain int is a limit from 0 for n items
	if l == len(word) {
		return 0, n
	}

	// Drop initial number now it's been processed
	word = word[l:]

	// If not a leading hyphen for a range is it a postfix? Postfix is a specific
	// instance such as 1st, 2nd, etc. If not a hyphen or a postfix it's not
	// special.
	if word[0] != '-' {
		if word == "ST" || word == "ND" || word == "RD" || word == "TH" {
			return n - 1, n
		}
		return -1, -1
	}

	// We have a hyphen for a range (n-N) so not special if an int does not
	// follow it.
	N, l := leadingDigits(word[1:])
	if l == 0 || l < len(word)-1 {
		return -1, -1
	}

	// Was a reverse range given? If so swap returned results
	if N < n {
		return N - 1, n
	}

	return n - 1, N
}

// leadingDigits returns an integer representing the digits at the beginning of
// a string and a count of the digits used. If the passed string has no leading
// digits then an integer of 0 will be returned with a count of 0. For example
// leadingDigits("123xyz") would return an int of 123 and count 3.
func leadingDigits(s string) (n, count int) {
	for _, b := range []byte(s) {
		if b < '0' || '9' < b {
			return
		}
		n *= 10
		n += int(b - '0')
		if n > 9999999 {
			n = 9999999
		}
		count++
	}
	return
}
