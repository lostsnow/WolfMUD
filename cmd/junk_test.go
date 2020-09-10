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

// TestJunk_messages checks messages are output in the correct order with
// the correct color as well as being sent to the right players.
func TestJunk_messages(t *testing.T) {

	const OI = "\n" + text.Info  // Observer Info shorthand
	const P = "\n" + text.Prompt // Prompt (StyleNone) shorthand

	for _, test := range []struct {
		params   string
		actor    string
		observer string
	}{
		{
			"", // No item
			text.Info + "You go to junk... something?" + P, "",
		}, {
			"frog", // Invalid item
			text.Bad + "You see no 'FROG' to junk." + P, "",
		}, {
			"rock", // Junk item at location
			text.Good + "You junk the rock." + P,
			OI + "You see the actor junk a rock." + P,
		}, {
			"smell", // Try to junk vetoed item at location
			text.Bad + "How, exactly, would you junk a bad smell?" + P, "",
		}, {
			"door", // Try to junk a narrative at location
			text.Bad + "You cannot junk the door." + P, "",
		}, {
			"observer", // Try to junk a player at location
			text.Bad + "The observer does not want to be junked!" + P, "",
		}, {
			"bucket", // Try to junk a container with item inside at location
			text.Good + "You junk the bucket." + P,
			OI + "You see the actor junk a bucket." + P,
		}, {
			"pouch", // Try to junk a container with vetoing item inside
			text.Bad + "The pouch seems to contain something that cannot be junked." + P, "",
		}, {
			"token", // Junk held item
			text.Good + "You junk the token." + P,
			OI + "You see the actor junk a token." + P,
		}, {
			"cup", // Try to junk a held container with item inside
			text.Good + "You junk the cup." + P,
			OI + "You see the actor junk a cup." + P,
		}, {
			"mug", // Try to junk a held container with vetoing item inside
			text.Bad + "The mug seems to contain something that cannot be junked." + P, "",
		}, {
			"doll", // Try to junk held, nested containers with vetoing item inside
			text.Bad + "The russian doll seems to contain something that cannot be junked." + P, "",
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
					attr.NewName("a bad smell"),
					attr.NewAlias("SMELL"),
					attr.NewDescription("This is a bad smell."),
					attr.NewVetoes(
						attr.NewVeto("JUNK", "How, exactly, would you junk a bad smell?"),
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
					),
				),
				attr.NewThing(
					attr.NewName("a pouch"),
					attr.NewAlias("POUCH"),
					attr.NewDescription("This is a pouch."),
					attr.NewInventory(
						attr.NewThing(
							attr.NewName("a magic bean"),
							attr.NewAlias("BEAN"),
							attr.NewDescription("This is a small magical bean."),
							attr.NewVetoes(
								attr.NewVeto("JUNK", "The magic bean is still here..."),
							),
						),
					),
				),
				attr.NewThing(
					attr.NewName("a rock"),
					attr.NewAlias("ROCK"),
					attr.NewDescription("This is a small rock."),
				),
				attr.NewThing(
					attr.NewName("a russian doll"),
					attr.NewAlias("DOLL"),
					attr.NewDescription("This is a wooden, nesting russian doll."),
					attr.NewInventory(
						attr.NewThing(
							attr.NewName("a russian doll"),
							attr.NewAlias("DOLL"),
							attr.NewDescription("This is a wooden, nesting russian doll."),
							attr.NewInventory(
								attr.NewThing(
									attr.NewName("a baby russian doll"),
									attr.NewAlias("BABY"),
									attr.NewDescription("This is a baby wooden russian doll."),
									attr.NewVetoes(
										attr.NewVeto("JUNK", "Awww... you can;t junk the baby :("),
									),
								),
							),
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
				attr.NewName("a cup"),
				attr.NewAlias("CUP"),
				attr.NewDescription("This is a cup."),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("some tea"),
						attr.NewAlias("TEA"),
						attr.NewDescription("This is some weak, milky, sweet tea."),
					),
				),
			),
			attr.NewThing(
				attr.NewName("a mug"),
				attr.NewAlias("MUG"),
				attr.NewDescription("This is a mug."),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("some coffee"),
						attr.NewAlias("COFFEE"),
						attr.NewDescription("This is some strong, black coffee."),
						attr.NewVetoes(
							attr.NewVeto("JUNK", "You cannot junk perfectly good coffee!"),
						),
					),
				),
			),
		)

		observer := cmd.NewTestPlayer("an observer", "OBSERVER")

		c := "junk " + test.params
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
