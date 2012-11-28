// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package thing

import (
	"code.wolfmud.org/WolfMUD.git/utils/uid"
	"strings"
	"testing"
)

var testSubjects = []struct {
	name        string
	aliases     []string
	description string
}{
	{"Name", []string{"Alias"}, "Description"},
	{"Thing", []string{"Thing", "Something"}, "I'm a Thing!"},
	{"", []string{}, ""},
	{"", nil, ""},
	{"Duplicate", []string{"Ditto", "Copy"}, "This is a duplicate duplicate"},
	{"Duplicate", []string{"Ditto", "Copy"}, "This is a duplicate duplicate"},
}

// Make sure aliases parameter is not modified by New
//
// NOTE: This MUST come before any other tests otherwise if New IS modifying
// parameters it would have corrupted the test subjects already!
//
// Yes, I spent HOURS debugging this little bugger...
func TestParameters(t *testing.T) {
	for i, s := range testSubjects {

		aliases_copy := make([]string, len(s.aliases))
		copy(aliases_copy, s.aliases)

		_ = New(s.name, s.aliases, s.description)

		for j, have := range s.aliases {
			want := aliases_copy[j]
			if have != want {
				t.Errorf("Alias parameter to New modified: Case %d, have %q want %q", i, have, want)
			}
		}

	}
}

func TestNew(t *testing.T) {
	for i, s := range testSubjects {
		thing := New(s.name, s.aliases, s.description)

		{
			have := thing.name
			want := s.name
			if have != want {
				t.Errorf("Corrupt name: Case %d, have %q wanted %q", i, have, want)
			}
		}

		{
			have := thing.description
			want := s.description
			if have != want {
				t.Errorf("Corrupt description: Case %d, have %q wanted %q", i, have, want)
			}
		}

		{
			have := len(thing.aliases)
			want := len(s.aliases)
			if have != want {
				t.Errorf("Invalid alias length: Case %d, have %d wanted %d", i, have, want)
			}
		}

		for i, have := range thing.aliases {
			want := strings.ToUpper(strings.TrimSpace(s.aliases[i]))
			if have != want {
				t.Errorf("Corrupt alias: Case %d, have %q, wanted %q", i, have, want)
			}
		}
	}
}

func TestName(t *testing.T) {
	for i, s := range testSubjects {
		thing := New(s.name, s.aliases, s.description)
		have := thing.Name()
		want := s.name
		if have != want {
			t.Errorf("Invalid Name: Case %d, have %q wanted %q", i, have, want)
		}
	}
}

func TestDescription(t *testing.T) {
	for i, s := range testSubjects {
		thing := New(s.name, s.aliases, s.description)
		have := thing.Description()
		want := s.description
		if have != want {
			t.Errorf("Invalid Description: Case %d, have %q wanted %q", i, have, want)
		}
	}
}

func TestAliases(t *testing.T) {
	for _, s := range testSubjects {
		thing := New(s.name, s.aliases, s.description)
		for i, have := range thing.Aliases() {
			want := strings.ToUpper(strings.TrimSpace(s.aliases[i]))
			if have != want {
				t.Errorf("Invalid alias: Case %d, have %q wanted %q", i, have, want)
			}
		}
	}
}

func TestIsAlias(t *testing.T) {

	allAliases := make(map[string](map[uid.UID]bool))
	subjects := make([]*Thing, len(testSubjects))

	// Go through the testSubjects and create subjects and a map of aliases that
	// map to unique Ids
	for i, s := range testSubjects {
		subjects[i] = New(s.name, s.aliases, s.description)
		for _, a := range s.aliases {
			if _, ok := allAliases[a]; !ok {
				allAliases[a] = make(map[uid.UID]bool)
			}
			allAliases[a][subjects[i].UniqueId()] = true
		}
	}

	// Go through all aliases and check in the map to see if IsAlias() should
	// return true or false
	for i, s := range subjects {
		for a := range allAliases {
			have := s.IsAlias(a)
			want := allAliases[a][s.UniqueId()]
			if have != want {
				t.Errorf("Corrupt IsAlias %q: Case %d, have %t wanted %t", a, i, have, want)
			}
		}
	}
}
