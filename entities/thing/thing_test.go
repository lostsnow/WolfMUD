// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package thing

import (
	"strconv"
	"strings"
	"testing"
	"time"
	. "wolfmud.org/utils/test"
	. "wolfmud.org/utils/uid"
)

type testData struct {
	name        string
	aliases     []string
	description string
}

var testSubjects = []*testData{
	{"Name", []string{"Alias"}, "Description"},
	{"Thing", []string{"Thing", "Something"}, "I'm a Thing!"},
	{"", []string{}, ""},
	{"", nil, ""},
	{"Duplicate", []string{"Ditto", "Copy"}, "This is a duplicate duplicate"},
	{"Duplicate", []string{"Ditto", "Copy"}, "This is a duplicate duplicate"},
}

func TestNew(t *testing.T) {
	for _, s := range testSubjects {
		n := New(s.name, s.aliases, s.description)

		Equal(t, "name", s.name, n.name)
		Equal(t, "description", s.description, n.description)
		Equal(t, "aliases size", len(s.aliases), len(n.aliases))

		for i, expect := range s.aliases {
			expect = strings.ToUpper(strings.TrimSpace(expect))
			Equal(t, "aliases", expect, n.aliases[i])
		}

	}

	// Make sure aliases are copied and don't reference original slice
	aliases := []string{"LOWERCASE"}
	subject := New("", aliases, "")
	NotEqual(t, "aliases parameter referenced", &aliases, &subject.aliases)
}

func TestName(t *testing.T) {
	for _, s := range testSubjects {
		n := New(s.name, s.aliases, s.description)
		Equal(t, "Name()", s.name, n.Name())
	}
}

func TestDescription(t *testing.T) {
	for _, s := range testSubjects {
		n := New(s.name, s.aliases, s.description)
		Equal(t, "Description()", s.description, n.Description())
	}
}

func TestAliases(t *testing.T) {
	for _, s := range testSubjects {
		n := New(s.name, s.aliases, s.description)
		aliases := n.Aliases()
		for i, expect := range s.aliases {
			expect = strings.ToUpper(strings.TrimSpace(expect))
			Equal(t, "Aliases() index "+strconv.Itoa(i), expect, aliases[i])
		}
	}
}

func TestIsAlso(t *testing.T) {

	subjects := make([]*Thing, len(testSubjects))
	for i, s := range testSubjects {
		subjects[i] = New(s.name, s.aliases, s.description)
	}

	// Match each thing with every other thing - should only be itself
	for i1, s1 := range subjects {
		for i2, s2 := range subjects {
			Equal(t, "IsAlso()", s1.IsAlso(s2), i1 == i2)
		}
	}
}

func TestIsAlias(t *testing.T) {

	allAliases := make(map[string](map[UID]bool))
	subjects := make([]*Thing, len(testSubjects))

	// Go through the testSubjects and create subjects and a map of aliases that
	// map to unique Ids
	for i, s := range testSubjects {
		subjects[i] = New(s.name, s.aliases, s.description)
		for _, a := range s.aliases {
			if _, ok := allAliases[a]; !ok {
				allAliases[a] = make(map[UID]bool)
			}
			allAliases[a][subjects[i].uniqueId] = true
		}
	}

	// Go through all aliases and check in the map to see if IsAlias() should
	// return true or false
	for _, s := range subjects {
		for a := range allAliases {
			Equal(t, "IsAlias()", allAliases[a][s.uniqueId], s.IsAlias(a))
		}
	}
}

func TestUniqueId(t *testing.T) {
	for _, s := range testSubjects {
		n := New(s.name, s.aliases, s.description)
		Equal(t, "UniqueId()", n.uniqueId, n.UniqueId())
	}
}

func TestLockUnlock(t *testing.T) {
	subject := New("",nil,"")

	// Check size of mutex channel when locking and unlocking
	subject.Lock()
	Equal(t, "Lock()", 1, len(subject.mutex))
	subject.Unlock()
	Equal(t, "Unlock()", 0, len(subject.mutex))

	// Get start time, lock subject and unlock after 1 second via the goroutine
	start := time.Now()
	subject.Lock()
	go func(){
		defer subject.Unlock()
		time.Sleep(1*time.Second)
	}()

	// While the goroutine is running try and lock a second time which should
	// block for at least a second until the goroutine unlocks
	subject.Lock()
	subject.Unlock()

	// Now get end time and workout how long we blocked for. If it's not at least
	// a second something is wrong.
	delay := time.Now().Sub(start).Seconds()

	if delay < 1 {
		t.Errorf("Lock() & Unlock() delay less than 1 second - %f, locking not working", delay)
	}
}
