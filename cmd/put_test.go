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

// TestPut_messages checks messages are output in the correct order with the
// correct color as well as being sent to the right players.
func TestPut_messages(t *testing.T) {

	// Observer Reset+Info shorthand
	const ORI = text.Reset + "\n" + text.Info

	for _, test := range []struct {
		params   string
		actor    string
		observer string
	}{
		{
			"", // No item or container
			text.Info + "You go to put something into something else...\n", "",
		}, {
			"ball box", // Held item into held container
			text.Good + "You put a small green ball into a box.\n",
			ORI + "You see the actor put something into a box.\n",
		}, {
			"ball box", // Held item into held container - duplicate, check world reset
			text.Good + "You put a small green ball into a box.\n",
			ORI + "You see the actor put something into a box.\n",
		}, {
			"ball hole", // Held item into container at location
			text.Good + "You put a small green ball into a hole.\n",
			ORI + "You see the actor put something into a hole.\n",
		}, {
			"ball", // Held item, no container
			text.Bad + "What did you want to put a small green ball into?\n", "",
		}, {
			"frog", // Invalid item, no container
			text.Bad + "You have no 'FROG' to put into anything.\n", "",
		}, {
			"ball frog", // Valid held item, invalid container
			text.Bad + "You see no 'FROG' to put a small green ball into.\n", "",
		}, {
			"box box", // Try putting held container inside itself
			text.Info + "It might be interesting to put a box inside itself, " +
				"but probably paradoxical as well.\n",
			ORI + "The actor seems to be trying to turn a box into a paradox.\n",
		}, {
			"hole hole", // Try putting container at location inside itself
			text.Bad + "You have no 'HOLE' to put into anything.\n", "",
		}, {
			"box ball", // Held item into a held non-container
			text.Bad + "You cannot put a box into a small green ball.\n", "",
		}, {
			"ball bag", // Held item into vetoing held container
			text.Bad + "You can't get the bag open.\n", "",
		}, {
			"ball observer", // Held item into a player (treated as vetoing container)
			text.Bad + "You can't put anything into the observer!\n", "",
		}, {
			"hole box", // Held item into container at location
			text.Bad + "You have no 'HOLE' to put into anything.\n", "",
		}, {
			"sticky box", // Vetoing held item into held container
			text.Bad + "You can't let go of something sticky.\n", "",
		}, {
			"rock box", // Non-held item into held container
			text.Bad + "You have no 'ROCK' to put into anything.\n", "",
		},
	} {

		world := attr.Things{
			attr.NewThing(
				attr.NewStart(),
				attr.NewName("Test room A"),
				attr.NewAlias("ROOM_A"),
				attr.NewDescription(
					"This is a room for testing. A large letter 'A' is painted on the wall.",
				),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("a hole"),
						attr.NewAlias("HOLE"),
						attr.NewDescription("This is a hole in the floor."),
						attr.NewInventory(),
						attr.NewNarrative(),
					),
					attr.NewThing(
						attr.NewName("a rock"),
						attr.NewAlias("ROCK"),
						attr.NewDescription("This is a small rock."),
					),
				),
			),
		}

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a box"),
				attr.NewAlias("BOX"),
				attr.NewDescription("This is a box."),
				attr.NewInventory(),
			),
			attr.NewThing(
				attr.NewName("a small green ball"),
				attr.NewAlias("+SMALL", "+GREEN", "BALL"),
				attr.NewDescription("This is a small, green rubber ball."),
			),
			attr.NewThing(
				attr.NewName("a bag"),
				attr.NewAlias("BAG"),
				attr.NewDescription("This is a sealed bag."),
				attr.NewInventory(),
				attr.NewVetoes(
					attr.NewVeto("PUTIN", "You can't get the bag open."),
				),
			),
			attr.NewThing(
				attr.NewName("something sticky"),
				attr.NewAlias("+SOMETHING:STICKY"),
				attr.NewDescription("This is something sticky"),
				attr.NewVetoes(
					attr.NewVeto("PUT", "You can't let go of something sticky."),
				),
			),
		)

		observer := cmd.NewTestPlayer("an observer", "OBSERVER")

		c := "put " + test.params
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

// Make sure we handle the actor not having an Inventory to put things in.
func TestPut_noInventory(t *testing.T) {

	world := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(
			attr.NewThing(
				attr.NewName("a hole"),
				attr.NewAlias("HOLE"),
				attr.NewDescription("This is a hole in the floor."),
				attr.NewInventory(),
				attr.NewNarrative(),
			),
		),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR")

	// Remove and free player's inventory
	inv := attr.FindInventory(actor)
	actor.Remove(inv)
	inv.Free()

	// Try and put an item into a container when player has no inventory
	c := "put ball hole"
	cmd.Parse(actor, c)
	have := actor.Messages()
	want := text.Bad + "You have no 'BALL' to put into anything.\n"
	if have != want {
		t.Errorf("Actor for %+q:\nhave: %+q\nwant: %+q", c, have, want)
	}

	world.Free()
}

// Check to make sure putting an item into a container moves it from the
// actor's Inventory into the container.
func TestPut_inventory(t *testing.T) {

	// Piece together world keeping references

	hole := attr.NewThing(
		attr.NewName("a hole"),
		attr.NewAlias("HOLE"),
		attr.NewDescription("This is a hole in the floor."),
		attr.NewInventory(),
		attr.NewNarrative(),
	)

	world := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(hole),
	)

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
	)

	token := attr.NewThing(
		attr.NewName("a token"),
		attr.NewAlias("+TEST", "TOKEN"),
		attr.NewDescription("This is a test token."),
	)

	box := attr.NewThing(
		attr.NewName("a box"),
		attr.NewAlias("BOX"),
		attr.NewDescription("This is a box."),
		attr.NewInventory(),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR", ball, token, box)

	// Put item into held container
	cmd.Parse(actor, "put ball box")

	// Item should now be in held container
	if attr.FindInventory(box).Search(ball.UID()) == nil {
		t.Errorf("%s: not in container.", attr.FindName(ball).Name("?"))
	}

	// Item should no longer be in Player Inventory
	if attr.FindInventory(actor).Search(ball.UID()) != nil {
		t.Errorf("%s: in player inventory.", attr.FindName(ball).Name("?"))
	}

	// Put item into container at location
	cmd.Parse(actor, "put token hole")

	// Item should now be in container at location
	if attr.FindInventory(hole).Search(token.UID()) == nil {
		t.Errorf("%s: not in container.", attr.FindName(token).Name("?"))
	}

	// Item should no longer be in Player Inventory
	if attr.FindInventory(actor).Search(token.UID()) != nil {
		t.Errorf("%s: in player inventory.", attr.FindName(token).Name("?"))
	}

	world.Free()
}

// TestPut_events tests to make sure action and cleanup events are enabled
// correctly when we put an item.
func TestPut_events(t *testing.T) {

	world := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(
			attr.NewThing(
				attr.NewName("a bag"),
				attr.NewAlias("BAG"),
				attr.NewDescription("This is a bag."),
				attr.NewInventory(),
			),
		),
	)

	cleanup := attr.NewCleanup(time.Hour, 0)
	action := attr.NewAction(time.Hour, 0)

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
		cleanup,
		action,
		attr.NewOnAction([]string{"The ball moves..."}),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR", ball,
		attr.NewThing(
			attr.NewName("a box"),
			attr.NewAlias("BOX"),
			attr.NewDescription("This is a box."),
			attr.NewInventory(),
		),
	)

	// Check clean up and action events inactive
	checkEvent(t, cleanup, inactive)
	checkEvent(t, action, inactive)

	// Put item into carried container - events should stay inactive
	cmd.Parse(actor, "put ball box")
	checkEvent(t, cleanup, inactive)
	checkEvent(t, action, inactive)

	// Take item from container - events should stay inactive
	cmd.Parse(actor, "take ball box")
	checkEvent(t, cleanup, inactive)
	checkEvent(t, action, inactive)

	// Put item into container at location - cleanup event only should become active
	cmd.Parse(actor, "put ball bag")
	checkEvent(t, cleanup, active)
	checkEvent(t, action, inactive)

	world.Free()
}
