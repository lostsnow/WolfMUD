// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd_test

import (
	"testing"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/text"
)

// TestExamine_messages checks messages are output in the correct order with
// the correct color as well as being sent to the right players.
func TestExamine_messages(t *testing.T) {

	const OI = "\n" + text.Info  // Observer Info shorthand
	const P = "\n" + text.Prompt // Prompt (StyleNone) shorthand

	for _, test := range []struct {
		params   string
		actor    string
		observer string
	}{
		{
			"", // No item
			text.Info + "You examine this and that, find nothing special." + P, "",
		}, {
			"frog", // Invalid item
			text.Bad + "You see no 'FROG' to examine." + P, "",
		}, {
			"rock", // Single simple item at location
			text.Good + "You examine the rock." +
				text.Reset + "\nThis is a small rock." + P,
			OI + "The actor studies a rock." + P,
		}, {
			"cup", // Examine empty container at location
			text.Good + "You examine the cup." +
				text.Reset + "\nThis is a cup. It is empty." + P,
			OI + "The actor studies a cup." + P,
		}, {
			"box", // Examine container with single item at location
			text.Good + "You examine the box." +
				text.Reset + "\nThis is a box. It contains a small green ball." + P,
			OI + "The actor studies a box." + P,
		}, {
			"bag", // Examine container with multile items at location
			text.Good + "You examine the bag." +
				text.Reset + "\nThis is a bag. It contains:\n" +
				"  a small green ball\n" +
				"  a small red ball" + P,
			OI + "The actor studies a bag." + P,
		}, {
			"token", // Single simple held item
			text.Good + "You examine the token." +
				text.Reset + "\nThis is a test token." + P,
			OI + "The actor studies a token they are carrying." + P,
		}, {
			"mug", // Examine empty, held container
			text.Good + "You examine the mug." +
				text.Reset + "\nThis is a mug. It is empty." + P,
			OI + "The actor studies a mug they are carrying." + P,
		}, {
			"pouch", // Examine held container with single item
			text.Good + "You examine the pouch." +
				text.Reset + "\nThis is a pouch. It contains a small green ball." + P,
			OI + "The actor studies a pouch they are carrying." + P,
		}, {
			"bucket", // Examine held container with multiple items
			text.Good + "You examine the bucket." +
				text.Reset + "\nThis is a small, plastic bucket. It contains:\n" +
				"  some sand\n" +
				"  a pretty seashell" + P,
			OI + "The actor studies a bucket they are carrying." + P,
		}, {
			"stone", // Examine held item also at location - should pick location item
			text.Good + "You examine the stone." +
				text.Reset + "\nThis is a large stone." + P,
			OI + "The actor studies a stone." + P,
		}, {
			"door", // Examine a closed door
			text.Good + "You examine the door." +
				text.Reset + "\nThis is a door. It is closed." + P,
			OI + "The actor studies a door." + P,
		}, {
			"window", // Examine an open window
			text.Good + "You examine the window." +
				text.Reset + "\nThis is a window. It is open." + P,
			OI + "The actor studies a window." + P,
		}, {
			"parchament", // Examine a vetoed item
			text.Bad + "The text on the ancient parchament swirls before you eyes." + P,
			"",
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
					attr.NewName("a door"),
					attr.NewAlias("DOOR"),
					attr.NewDescription("This is a door."),
					attr.NewDoor(attr.East, false, time.Second, 0),
					attr.NewNarrative(),
				),
				attr.NewThing(
					attr.NewName("a window"),
					attr.NewAlias("WINDOW"),
					attr.NewDescription("This is a window."),
					attr.NewDoor(attr.East, true, time.Second, 0),
					attr.NewNarrative(),
				),
				attr.NewThing(
					attr.NewName("a rock"),
					attr.NewAlias("ROCK"),
					attr.NewDescription("This is a small rock."),
				),
				attr.NewThing(
					attr.NewName("a stone"),
					attr.NewAlias("STONE"),
					attr.NewDescription("This is a large stone."),
				),
				attr.NewThing(
					attr.NewName("a cup"),
					attr.NewAlias("CUP"),
					attr.NewDescription("This is a cup."),
					attr.NewInventory(),
				),
				attr.NewThing(
					attr.NewName("a box"),
					attr.NewAlias("BOX"),
					attr.NewDescription("This is a box."),
					attr.NewInventory(
						attr.NewThing(
							attr.NewName("a small green ball"),
							attr.NewAlias("BALL"),
							attr.NewDescription("This is a small, green rubber ball."),
						),
					),
				),
				attr.NewThing(
					attr.NewName("a bag"),
					attr.NewAlias("BAG"),
					attr.NewDescription("This is a bag."),
					attr.NewInventory(
						attr.NewThing(
							attr.NewName("a small green ball"),
							attr.NewAlias("BALL"),
							attr.NewDescription("This is a small, green rubber ball."),
						),
						attr.NewThing(
							attr.NewName("a small red ball"),
							attr.NewAlias("BALL"),
							attr.NewDescription("This is a small, red rubber ball."),
						),
					),
				),
			),
		)

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a stone"),
				attr.NewAlias("STONE"),
				attr.NewDescription("This is a small stone."),
			),
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
			),
			attr.NewThing(
				attr.NewName("a mug"),
				attr.NewAlias("MUG"),
				attr.NewDescription("This is a mug."),
				attr.NewInventory(),
			),
			attr.NewThing(
				attr.NewName("a pouch"),
				attr.NewAlias("POUCH"),
				attr.NewDescription("This is a pouch."),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("a small green ball"),
						attr.NewAlias("BALL"),
						attr.NewDescription("This is a small, green rubber ball."),
					),
				),
			),
			attr.NewThing(
				attr.NewName("an ancient parchament"),
				attr.NewAlias("PARCHAMENT"),
				attr.NewDescription("This is an ancient parchament."),
				attr.NewVetoes(
					attr.NewVeto("EXAM", "The text on the ancient parchament swirls before you eyes."),
				),
			),
			attr.NewThing(
				attr.NewName("a bucket"),
				attr.NewAlias("BUCKET"),
				attr.NewDescription("This is a small, plastic bucket."),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("some sand"),
						attr.NewAlias("SAND"),
						attr.NewDescription("This is a small amount of sand."),
					),
					attr.NewThing(
						attr.NewName("a pretty seashell"),
						attr.NewAlias("SEASHELL"),
						attr.NewDescription("This is a pretty seashell."),
					),
				),
			),
		)

		observer := cmd.NewTestPlayer("observer", "OBSERVER")

		c := "examine " + test.params
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

// TestExamine_player checks that when studying another player that the player
// becomes the participant of the EXAMINE command and that the participants
// inventory is not revealed.
func TestExamine_player(t *testing.T) {

	const OI = "\n" + text.Info  // Observer Info shorthand
	const P = "\n" + text.Prompt // Prompt (StyleNone) shorthand

	for _, test := range []struct {
		params      string
		actor       string
		observer    string
		participant string
	}{
		{
			"participant",
			text.Good + "You examine the participant." +
				text.Reset + "\nThis is a test player." + P,
			OI + "The actor studies a participant." + P,
			OI + "The actor studies you." + P,
		},
	} {

		world := attr.NewThing(
			attr.NewStart(),
			attr.NewName("Test room A"),
			attr.NewAlias("ROOM_A"),
			attr.NewDescription(
				"This is a room for testing.",
			),
			attr.NewInventory(),
			attr.NewExits(),
		)

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
			),
		)

		observer := cmd.NewTestPlayer("an observer", "OBSERVER")

		participant := cmd.NewTestPlayer("a participant", "PARTICIPANT",
			attr.NewThing(
				attr.NewName("a mug"),
				attr.NewAlias("MUG"),
				attr.NewDescription("This is a mug."),
				attr.NewInventory(),
			),
		)

		c := "examine " + test.params
		t.Run(c, func(t *testing.T) {
			cmd.Parse(actor, c)
			if have := actor.Messages(); have != test.actor {
				t.Errorf("Actor for %+q:\nhave: %+q\nwant: %+q", c, have, test.actor)
			}
			if have := observer.Messages(); have != test.observer {
				t.Errorf("Observer for %+q:\nhave: %+q\nwant: %+q", c, have, test.observer)
			}
			if have := participant.Messages(); have != test.participant {
				t.Errorf("Participant for %+q:\nhave: %+q\nwant: %+q", c, have, test.participant)
			}
		})

		world.Free()
	}
}
