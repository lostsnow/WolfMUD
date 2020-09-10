// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd_test

import (
	"testing"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/text"
)

// TestRead_messages checks messages are output in the correct order with
// the correct color as well as being sent to the right players.
func TestRead_messages(t *testing.T) {

	const OI = "\n" + text.Info  // Observer Info shorthand
	const P = "\n" + text.Prompt // Prompt (StyleNone) shorthand

	for _, test := range []struct {
		params   string
		actor    string
		observer string
	}{
		{
			"", // No item
			text.Info + "Did you want to read something specific?" + P, "",
		}, {
			"frog", // Invalid item
			text.Bad + "You see no 'FROG' to read." + P, "",
		}, {
			"plaque", // Read narrative item at location
			text.Good + "You read the plaque." +
				text.Reset + "\nIt says 'Please do not read this plaque'." + P,
			OI + "You see the actor read a plaque." + P,
		}, {
			"newspaper", // Read item at location
			text.Good + "You read the newspaper." +
				text.Reset + "\nIt's full of depressing news stories." + P,
			OI + "You see the actor read a newspaper." + P,
		}, {
			"token", // Read held item
			text.Good + "You read the token." +
				text.Reset + "\nIt has 'Test Token' written on it." + P,
			OI + "You see the actor read a token." + P,
		}, {
			"rock", // Try to read item with no writing
			text.Bad + "You see no writing on the rock to read." + P, "",
		},
	} {

		world := attr.NewThing(
			attr.NewStart(),
			attr.NewName("Test room A"),
			attr.NewAlias("ROOM_A"),
			attr.NewDescription(
				"This is a room for testing.",
			),
			attr.NewExits(),
			attr.NewInventory(
				attr.NewThing(
					attr.NewName("a rock"),
					attr.NewAlias("ROCK"),
					attr.NewDescription("This is a small rock."),
				),
				attr.NewThing(
					attr.NewName("a newspaper"),
					attr.NewAlias("NEWSPAPER"),
					attr.NewDescription("This is a folded newspaper."),
					attr.NewWriting("It's full of depressing news stories."),
				),
				attr.NewThing(
					attr.NewName("a plaque"),
					attr.NewAlias("PLAQUE"),
					attr.NewDescription("This is a small plaque."),
					attr.NewNarrative(),
					attr.NewWriting("It says 'Please do not read this plaque'."),
				),
			),
		)

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
				attr.NewWriting("It has 'Test Token' written on it."),
			),
		)

		observer := cmd.NewTestPlayer("observer", "OBSERVER")

		c := "read " + test.params
		t.Run(c, func(t *testing.T) {
			cmd.Parse(actor, c)
			if have := actor.Messages(); have != test.actor {
				t.Errorf("Actor for %+q:\nhave: %+q\nwant: %+q", c, have, test.actor)
			}
			if have := observer.Messages(); have != test.observer {
				t.Errorf("Observer for %+q:\nhave: %+q\nwant: %+q", c, have, test.observer)
			}
		})

		world.Free()
	}
}
