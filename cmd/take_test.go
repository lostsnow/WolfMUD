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

// TestTake_messages checks messages are output in the correct order with the
// correct color as well as being sent to the right players.
func TestTake_messages(t *testing.T) {

	const OI = "\n" + text.Info  // Observer Info shorthand
	const P = "\n" + text.Prompt // Prompt (StyleNone) shorthand

	for _, test := range []struct {
		params   string
		actor    string
		observer string
	}{
		{
			"", // No item or container
			text.Info + "You go to take something out of something else..." + P, "",
		}, {
			"ball box", // Item from held container
			text.Good + "You take the small green ball out of the box." + P,
			OI + "You see the actor take something out of a box." + P,
		}, {
			"ball box", // Item from held container - duplicate, check world reset
			text.Good + "You take the small green ball out of the box." + P,
			OI + "You see the actor take something out of a box." + P,
		}, {
			"ball bag box", // 2x item from held container
			text.Good + "You take the small green ball out of the box.\n" +
				text.Good + "You take the bag out of the box." + P,
			OI + "You see the actor take something out of a box." + P,
		}, {
			"rock pebble hole", // 2x item from container at location
			text.Good + "You take the rock out of the hole.\n" +
				text.Good + "You take the pebble out of the hole." + P,
			OI + "You see the actor take something out of a hole." + P,
		}, {
			"rock hole", // Item from container at location
			text.Good + "You take the rock out of the hole." + P,
			OI + "You see the actor take something out of a hole." + P,
		}, {
			"ball", // Item only
			text.Bad + "What did you want to take 'BALL' out of?" + P, "",
		}, {
			"frog", // Invalid item only
			text.Bad + "What did you want to take 'FROG' out of?" + P, "",
		}, {
			"box", // Item only that is a container
			text.Bad + "Did you want to take something from the box?" + P, "",
		}, {
			"2nd token", // Not enough of held item only
			text.Bad + "What did you want to take 'TOKEN' out of?" + P, "",
		}, {
			"ball frog bag box", // valid item + invalid item from held container
			text.Good + "You take the small green ball out of the box.\n" +
				text.Bad + "The box does not seem to contain 'FROG'.\n" +
				text.Good + "You take the bag out of the box." + P,
			OI + "You see the actor rummage around in a box.\n" +
				text.Info + "You see the actor take something out of a box." + P,
		}, {
			"frog ball newt box", // 2x invalid item, make sure only 1 rummage message
			text.Bad + "The box does not seem to contain 'FROG'.\n" +
				text.Good + "You take the small green ball out of the box.\n" +
				text.Bad + "The box does not seem to contain 'NEWT'." + P,
			OI + "You see the actor rummage around in a box.\n" +
				text.Info + "You see the actor take something out of a box." + P,
		}, {
			"ball bag", // Item from vetoing held container
			text.Bad + "You can't get the bag open." + P, "",
		}, {
			"frog bag", // Invalid item from vetoing held container
			text.Bad + "You can't get the bag open." + P, "",
		}, {
			"ball sack", // Item from invalid container
			text.Bad + "You see no 'SACK' to take things out of." + P, "",
		}, {
			"ball token", // Item from a non-container
			text.Bad + "You cannot take anything from the token." + P, "",
		}, {
			"token box", // Item not in container
			text.Bad + "The box does not seem to contain 'TOKEN'." + P,
			OI + "You see the actor rummage around in a box." + P,
		}, {
			"sticky box", // Item with a TAKE veto from container
			text.Bad + "You can't take something sticky." + P, "",
		}, {
			"carving box", // Narrative item from container
			text.Bad + "For some reason you cannot take the carving from the box." + P,
			OI + "You see the actor having trouble removing something from a box." + P,
		}, {
			"carving scratch box", // 2x narrative item from container, 1 trouble msg
			text.Bad + "For some reason you cannot take the carving from the box.\n" +
				text.Bad + "For some reason you cannot take the scratch from the box." + P,
			OI + "You see the actor having trouble removing something from a box." + P,
		}, {
			"2nd ball box", // Not enough items from held container
			text.Bad + "There are not that many 'BALL' to take from the box." + P, "",
		}, {
			"ball 2nd box", // Not enough held containers
			text.Bad + "You don't see that many 'BOX' to take things out of." + P, "",
		}, {
			"ball all container", // Multiple containers
			text.Bad + "You can only take things from one container at a time." + P, "",
		}, {
			"2 container", // 2x items only that are containers
			text.Bad + "You go to take things out of... something?" + P, "",
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
						attr.NewInventory(
							attr.NewThing(
								attr.NewName("a rock"),
								attr.NewAlias("ROCK"),
								attr.NewDescription("This is a small rock."),
							),
							attr.NewThing(
								attr.NewName("a pebble"),
								attr.NewAlias("PEBBLE"),
								attr.NewDescription("This is a small, smooth pebble."),
							),
						),
						attr.NewNarrative(),
					),
				),
			),
		}

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a box"),
				attr.NewAlias("CONTAINER", "BOX"),
				attr.NewDescription("This is a box."),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("a small green ball"),
						attr.NewAlias("+SMALL", "+GREEN", "BALL"),
						attr.NewDescription("This is a small, green rubber ball."),
					),
					attr.NewThing(
						attr.NewName("something sticky"),
						attr.NewAlias("+SOMETHING:STICKY"),
						attr.NewDescription("This is something sticky"),
						attr.NewVetoes(
							attr.NewVeto("TAKE", "You can't take something sticky."),
						),
					),
					attr.NewThing(
						attr.NewName("a carving"),
						attr.NewAlias("CARVING"),
						attr.NewDescription("This is a small, intricate carving."),
						attr.NewNarrative(),
					),
					attr.NewThing(
						attr.NewName("a scratch"),
						attr.NewAlias("SCRATCH"),
						attr.NewDescription("This is a cratch on the box."),
						attr.NewNarrative(),
					),
					attr.NewThing(
						attr.NewName("a bag"),
						attr.NewAlias("BAG"),
						attr.NewDescription("This is a small bag."),
						attr.NewInventory(),
					),
				),
			),
			attr.NewThing(
				attr.NewName("a bag"),
				attr.NewAlias("CONTAINER", "BAG"),
				attr.NewDescription("This is a sealed bag."),
				attr.NewInventory(),
				attr.NewVetoes(
					attr.NewVeto("TAKEOUT", "You can't get the bag open."),
				),
			),
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
			),
		)

		observer := cmd.NewTestPlayer("an observer", "OBSERVER")

		c := "take " + test.params
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
func TestTake_noInventory(t *testing.T) {

	world := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(
			attr.NewThing(
				attr.NewName("a hole"),
				attr.NewAlias("HOLE"),
				attr.NewDescription("This is a hole in the floor."),
				attr.NewNarrative(),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("a small green ball"),
						attr.NewAlias("+SMALL", "+GREEN", "BALL"),
						attr.NewDescription("This is a small, green rubber ball."),
					),
				),
			),
		),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR")

	// Remove and free player's inventory
	inv := attr.FindInventory(actor)
	actor.Remove(inv)
	inv.Free()

	// Try and take an item from a container when player has no inventory
	c := "take ball hole"
	cmd.Parse(actor, c)
	have := actor.Messages()
	want := text.Bad + "You have nowhere to put the small green ball if you remove it from the hole.\n" + text.Prompt
	if have != want {
		t.Errorf("Actor for %+q:\nhave: %+q\nwant: %+q", c, have, want)
	}

	world.Free()
}

// Check to make sure taking an item from a container moves it from the
// container and into the actor's Inventory.
func TestTake_inventory(t *testing.T) {

	// Piece together world keeping references

	token := attr.NewThing(
		attr.NewName("a token"),
		attr.NewAlias("+TEST", "TOKEN"),
		attr.NewDescription("This is a test token."),
	)

	hole := attr.NewThing(
		attr.NewName("a hole"),
		attr.NewAlias("HOLE"),
		attr.NewDescription("This is a hole in the floor."),
		attr.NewInventory(token),
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

	box := attr.NewThing(
		attr.NewName("a box"),
		attr.NewAlias("BOX"),
		attr.NewDescription("This is a box."),
		attr.NewInventory(ball),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR", ball, token, box)

	// Take item from held container
	cmd.Parse(actor, "take ball box")

	// Item should now be in actor's inventory
	if attr.FindInventory(actor).Search(ball.UID()) == nil {
		t.Errorf("%s: not in player inventory.", attr.FindName(ball).Name("?"))
	}

	// Item should no longer be in container at location
	if attr.FindInventory(box).Search(ball.UID()) != nil {
		t.Errorf("%s: in container.", attr.FindName(ball).Name("?"))
	}

	// Take item from container at location
	cmd.Parse(actor, "take token hole")

	// Item should now be in actor's inventory
	if attr.FindInventory(actor).Search(token.UID()) == nil {
		t.Errorf("%s: not in player inventory.", attr.FindName(token).Name("?"))
	}

	// Item should no longer be in container at location
	if attr.FindInventory(hole).Search(token.UID()) != nil {
		t.Errorf("%s: in container at location.", attr.FindName(token).Name("?"))
	}

	world.Free()
}

// TestTake_spawnable checks that when a spawnable item is taken from a
// container up we get a copy of the item and not the original.
func TestTake_spawnable(t *testing.T) {

	// Piece together world keeping references

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
		attr.NewReset(time.Hour, 0, true, false),
	)
	uid := ball.UID()

	inv := attr.NewInventory(ball)

	box := attr.NewThing(
		attr.NewName("a box"),
		attr.NewAlias("BOX"),
		attr.NewDescription("This is a box."),
		inv,
	)

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR", box)

	// Set origins so events work - usually done by the zone loader.
	locA.SetOrigins()

	// Try and take small ball from the box
	cmd.Parse(actor, "take ball box")

	// Small ball should not be found in box's Inventory
	if inv.Search(uid) != nil {
		t.Errorf("original ball still in box's inventory.")
	}

	// Small ball should be disabled waiting for reset
	found := false
	for _, t := range inv.Disabled() {
		if t.UID() == uid {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("original ball not disabled in box.")
	}

	// Copy of small ball should now be in Player Inventory with different UID
	copy := attr.FindInventory(actor).Search("BALL")
	if copy == nil {
		t.Errorf("no ball found in player inventory.")
	}
	if copy != nil && copy.UID() == uid {
		t.Errorf("original ball in player inventory - should be a copy.")
	}

	locA.Free()
}

// TestTake_events tests to make sure action and cleanup events are disabled
// correctly when we take an item.
func TestTake_events(t *testing.T) {

	cleanup := attr.NewCleanup(time.Hour, 0)
	action := attr.NewAction(time.Hour, 0)

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
		cleanup,
		action,
		attr.NewOnAction([]string{"The ball moves..."}),
	)

	box := attr.NewThing(
		attr.NewName("a box"),
		attr.NewAlias("BOX"),
		attr.NewDescription("This is a box."),
		attr.NewInventory(ball),
	)

	world := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(box),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR", ball)

	// Manually start cleanup and action events
	cleanup.Cleanup()
	action.Action()

	// Check clean up and action events are active
	checkEvent(t, cleanup, active)
	checkEvent(t, action, active)

	// Take item from container at location - events should become inactive
	cmd.Parse(actor, "take ball box")
	checkEvent(t, cleanup, inactive)
	checkEvent(t, action, inactive)

	world.Free()
}
