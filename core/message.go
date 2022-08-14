// Copyright 2022 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"math/rand"
	"strings"

	"code.wolfmud.org/WolfMUD.git/text"
)

var views = []byte{'A', 'D', 'O'}

func Message(A, D *Thing, msg string) (a, d, o string) {

	type msb = map[byte]bool

	var (
		who   = map[byte]*Thing{'A': A, 'D': D}
		used  = map[byte]msb{'A': msb{}, 'D': msb{}, 'O': msb{}}
		reply = map[byte]*strings.Builder{
			'A': &strings.Builder{},
			'D': &strings.Builder{},
			'O': &strings.Builder{},
		}
		subs       []string
		tok        string
		Ltok, Rtok string
		lhb, LHB   byte
	)

	for _, m := range split(msg) {

		// Shortcut if not an '[...]' block
		if m[len(m)-1] != ']' {
			reply['A'].WriteString(m)
			reply['D'].WriteString(m)
			reply['O'].WriteString(m)
			continue
		}

		// Separate substitutions in a '[...]' block
		subs = strings.Split(m, "/")

		for vidx, view := range views {

			// Get Nth substitution token from '[...]' block, repeat last
			// substitution if fewer than N available.
			tok = subs[min(len(subs)-1, vidx)]

			if l := len(tok); l > 0 && tok[l-1] == ']' {
				tok = tok[:l-1]
			}

			// Split substition token on a period.
			Ltok, Rtok, _ = strings.Cut(tok, ".")

			// Convert LHS to original (lhb) and upper (LHB) cased byte.
			if lhb, LHB = 0, 0; len(Ltok) == 2 && Ltok[0] == '%' {
				lhb = Ltok[1]
				LHB = lhb & 0b11011111
			}

			switch {
			case LHB != 'A' && LHB != 'D':
			case view == LHB && LHB == lhb:
				tok = "You"
			case view == LHB:
				tok = "you"
			case Rtok == "" && LHB == lhb:
				tok = who[LHB].As[UTheName]
			case Rtok == "":
				tok = who[LHB].As[TheName]
			case !used[view][LHB] && LHB == lhb:
				tok = who[LHB].As[UTheName]
			case !used[view][LHB]:
				tok = who[LHB].As[TheName]
			case LHB == lhb:
				tok = text.TitleFirst(Pronoun(who[LHB], Rtok))
			default:
				tok = Pronoun(who[LHB], Rtok)
			}

			used[view][LHB] = true
			reply[view].WriteString(tok)
		}
	}

	return reply['A'].String(), reply['D'].String(), reply['O'].String()
}

// split takes a message and breaks it into chunks of substitution blocks
// '[...]' and non-substitution blocks. Substitution blocks are those strings
// ending with a ']'. For example:
//
//	split("[%A] swing[/s] [your/%A.his] sword at [%D].")
//
//	[]string{"%A]", " swing", "/s]", " ", "your/%A.his]", " sword at ", "%D]", "."}
func split(m string) []string {
	chunks := strings.Split(m, "[")
	split := make([]string, 0, len(chunks)*2)
	var subchunks []string

	for _, chunk := range chunks {
		subchunks = subchunks[:0]
		subchunks = strings.SplitAfter(chunk, "]")
		for _, sc := range subchunks {
			if len(sc) == 0 {
				continue
			}
			split = append(split, sc)
		}
	}
	return split
}

// pronouns provides a lookup table of pronouns for a gender.
var pronouns = map[string][]string{
	"FEMALE":  {"she", "her", "her", "herself"},
	"IT":      {"it", "it", "its", "itself"},
	"":        {"it", "it", "its", "itself"},
	"MALE":    {"he", "him", "his", "himself"},
	"NEUTRAL": {"they", "them", "their", "themself"},
}

var pidx = map[string]int{"they": 0, "them": 1, "their": 2, "themself": 3}

// Pronoun returns the correct pronoun for a given actor's gender. This allows,
// for example, messages to be defined generally with neutral pronouns. This
// function will then turn the neutral pronoun into one specific to the actor's
// gender. If the actor does not have a gender, or the gender is invalid, a
// generic 'it' is assumed.
//
// NOTE: messages should be defined using neutral pronouns as they are unique.
func Pronoun(actor *Thing, p string) string {
	gender := actor.As[Gender]
	if pronouns[gender] == nil {
		gender = "IT"
	}
	return pronouns[gender][pidx[p]]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func pickMessage(actor *Thing) string {
	var msgs []string
	for _, item := range actor.In {
		if item.Is&(Wielding) != 0 && len(item.Any[OnCombat]) != 0 {
			msgs = append(msgs, item.Any[OnCombat]...)
		}
	}
	if len(msgs) == 0 {
		msgs = actor.Any[OnCombat]
	}
	if len(msgs) == 0 {
		return "[%A] hit[/s] [%d] wounding [%d.them]."
	}
	return msgs[rand.Intn(len(msgs))]
}
