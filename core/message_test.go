// Copyright 2022 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core_test

import (
	"testing"

	"code.wolfmud.org/WolfMUD.git/core"
)

func TestMessage(t *testing.T) {

	female := core.NewThing()
	female.As[core.Gender] = "FEMALE"
	female.As[core.Name] = "Alice"
	female.As[core.UName] = "Alice"
	female.As[core.TheName] = "Alice"
	female.As[core.UTheName] = "Alice"

	male := core.NewThing()
	male.As[core.Gender] = "MALE"
	male.As[core.Name] = "Bob"
	male.As[core.UName] = "Bob"
	male.As[core.TheName] = "Bob"
	male.As[core.UTheName] = "Bob"

	it := core.NewThing()
	it.As[core.Name] = "an imp"
	it.As[core.UName] = "An imp"
	it.As[core.TheName] = "an imp"
	it.As[core.UTheName] = "An imp"

	type sss struct {
		A string
		D string
		O string
	}

	var have sss

	for _, test := range []struct {
		A    *core.Thing
		B    *core.Thing
		msg  string
		want sss
	}{
		{female, male, "", sss{"", "", ""}},
		{female, male, "[", sss{"", "", ""}},
		{female, male, "]", sss{"", "", ""}},
		{female, male, "[]", sss{"", "", ""}},
		{female, male, "[/]", sss{"", "", ""}},
		{female, male, "[//]", sss{"", "", ""}},
		{female, male, "[///]", sss{"", "", ""}},
		{female, male, "[a]", sss{"a", "a", "a"}},
		{female, male, "[/a]", sss{"", "a", "a"}},
		{female, male, "[//a]", sss{"", "", "a"}},
		{female, male, "[///a]", sss{"", "", ""}},
		{female, male, "[a/b]", sss{"a", "b", "b"}},
		{female, male, "[a//b]", sss{"a", "", "b"}},
		{female, male, "[a/b/c]", sss{"a", "b", "c"}},
		{female, male, "[a/b//c]", sss{"a", "b", ""}},
		{female, male, "text", sss{"text", "text", "text"}},
		{it, female, "[%A]", sss{"You", "An imp", "An imp"}},
		{it, female, "[%a]", sss{"you", "an imp", "an imp"}},
		{female, it, "[%D]", sss{"An imp", "You", "An imp"}},
		{female, it, "[%d]", sss{"an imp", "you", "an imp"}},
		{female, it, "[%X]", sss{"%X", "%X", "%X"}},
		{female, it, "[%A] [%X]", sss{"You %X", "Alice %X", "Alice %X"}},
		{female, it, "[%A.they]", sss{"You", "Alice", "Alice"}},
		{male, it, "[%A] [%A.they]", sss{"You You", "Bob He", "Bob He"}},
		{male, it, "[%A] [%a.they]", sss{"You you", "Bob he", "Bob he"}},
		{male, it, "[%A] [%a.them]", sss{"You you", "Bob him", "Bob him"}},
		{male, it, "[%A] [%a.their]", sss{"You you", "Bob his", "Bob his"}},
		{male, it, "[%A] [%a.themself]", sss{"You you", "Bob himself", "Bob himself"}},
		{male, it, "[%A.dummy]", sss{"You", "Bob", "Bob"}},
		{male, it, "[%A] [%A.dummy]", sss{"You You", "Bob He", "Bob He"}},
		{male, it, "[%A.]", sss{"You", "Bob", "Bob"}},
		{male, it, "[%A] [%A.]", sss{"You You", "Bob Bob", "Bob Bob"}},
		{male, female, "[%A] [%D]", sss{"You Alice", "Bob You", "Bob Alice"}},
		{male, female, "[%A] [%d]", sss{"You Alice", "Bob you", "Bob Alice"}},
		{male, female, "[%A] [%d.they]", sss{"You Alice", "Bob you", "Bob Alice"}},
		{male, female, "[%A] [%d.them]", sss{"You Alice", "Bob you", "Bob Alice"}},
		{male, female, "[%A] [%d.their]", sss{"You Alice", "Bob you", "Bob Alice"}},
		{male, female, "[%A] [%d.themself]", sss{"You Alice", "Bob you", "Bob Alice"}},
		{male, female, "[%A] [%a.they]", sss{"You you", "Bob he", "Bob he"}},
		{male, female, "[%A] [%a.them]", sss{"You you", "Bob him", "Bob him"}},
		{male, female, "[%A] [%a.their]", sss{"You you", "Bob his", "Bob his"}},
		{male, female, "[%A] [%a.themself]", sss{"You you", "Bob himself", "Bob himself"}},

		{female, male, "[%A] hit[/s] [%d].",
			sss{"You hit Bob.", "Alice hits you.", "Alice hits Bob."},
		},
		{female, male, "[%A] swing[/s] [your/%a.their] dagger at [%d] cutting [%d.them].", sss{
			"You swing your dagger at Bob cutting him.",
			"Alice swings her dagger at you cutting you.",
			"Alice swings her dagger at Bob cutting him.",
		}},
	} {
		t.Run(test.msg, func(t *testing.T) {
			have.A, have.D, have.O = core.Message(test.A, test.B, test.msg)
			if have.A != test.want.A {
				t.Errorf("\nhave: %q\nwant: %q", have.A, test.want.A)
			}
			if have.D != test.want.D {
				t.Errorf("\nhave: %q\nwant: %q", have.D, test.want.D)
			}
			if have.O != test.want.O {
				t.Errorf("\nhave: %q\nwant: %q", have.O, test.want.O)
			}
		})
	}
}

func BenchmarkMessage(b *testing.B) {

	female := core.NewThing()
	female.As[core.Gender] = "FEMALE"
	female.As[core.Name] = "Alice"
	female.As[core.UName] = "Alice"
	female.As[core.TheName] = "Alice"
	female.As[core.UTheName] = "Alice"

	male := core.NewThing()
	male.As[core.Gender] = "MALE"
	male.As[core.Name] = "Bob"
	male.As[core.UName] = "Bob"
	male.As[core.TheName] = "Bob"
	male.As[core.UTheName] = "Bob"

	msg := "[%A] swing[/s] [your/%a.their] dagger at [%d] cutting [%d.them]."

	b.Run("message", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = core.Message(female, male, msg)
		}
	})
}

func TestPronoun(t *testing.T) {
	for _, test := range []struct {
		gender  string
		pronoun string
		want    string
	}{
		{"MALE", "they", "he"},
		{"FEMALE", "they", "she"},
		{"IT", "they", "it"},
		{"NEUTRAL", "they", "they"},
		{"", "they", "it"},
		{"dummy", "they", "it"},
	} {
		t.Run(test.gender+":"+test.pronoun, func(t *testing.T) {
			what := core.NewThing()
			what.As[core.Gender] = test.gender
			have := core.Pronoun(what, test.pronoun)
			if have != test.want {
				t.Errorf("\nhave: %q\nwant: %q", have, test.want)
			}
		})
	}
}
