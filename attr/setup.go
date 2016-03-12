// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"

	"log"
	"os"
	"path/filepath"
)

// Setup test 'world' with some test data
func Setup() map[string]has.Thing {

	world := map[string]has.Thing{}

	world["cheese"] = NewThing(
		NewName("some cheese"),
		NewDescription("This is a blob of very soft, sticky, smelly cheese."),
		NewAlias("cheese"),
		NewVetoes(
			NewVeto("drop", "You can't drop the sticky cheese!"),
		),
	)

	world["mug"] = NewThing(
		NewName("a mug"),
		NewDescription("This is a large, white, chipped mug."),
		NewWriting("Stay calm and drink more coffee!"),
		NewAlias("mug"),
		NewAlias("cup"),
		NewInventory(
			NewThing(
				NewName("some coffee"),
				NewDescription("This is some hot, strong coffee."),
			),
		),
	)

	world["box"] = NewThing(
		NewName("a box"),
		NewDescription("This is a small, wooden box."),
		NewInventory(),
		NewAlias("box"),
	)

	world["bag"] = NewThing(
		NewName("a bag"),
		NewDescription("This is a small bag."),
		NewAlias("bag"),
		NewInventory(
			NewThing(
				NewName("an apple"),
				NewDescription("This is a juicy red apple."),
				NewAlias("apple"),
			),
			NewThing(
				NewName("an orange"),
				NewDescription("This is a large orange."),
				NewAlias("orange"),
			),
		),
	)

	world["chairs"] = NewThing(
		NewNarrative(),
		NewName("some rough chairs"),
		NewDescription("These chairs(?) are very rough wooden affairs, so rough in fact you decide it's a bad idea to sit on them without some descent rear armour to fend of the splinters."),
		NewAlias("chair", "chairs"),
	)

	world["tables"] = NewThing(
		NewNarrative(),
		NewName("some rough tables"),
		NewDescription("Well you suppose these are tables. If so they must have been made by a blind carpenter having a very bad day."),
		NewAlias("table", "tables"),
	)

	world["plaque"] = NewThing(
		NewNarrative(),
		NewName("a wooden plaque"),
		NewDescription("This is a small wooden plaque."),
		NewAlias("plaque"),
		NewWriting("Please do not read the plaques!"),
		NewVetoes(
			NewVeto("get", "You can't take the plaque. It's firmly nailed to the wall."),
			NewVeto("examine", "You try to examine the plaque but it makes your head hurt."),
		),
	)

	// Define some locations

	world["loc1"] = NewThing(
		NewStart(),
		NewName("Fireplace"),
		NewDescription("You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. There is a small plaque above the fireplace. To the south the common room extends and east the common room leads to the tavern entrance."),
		NewAlias("tavern", "fireplace"),
		NewInventory(
			world["cheese"],
			world["mug"],
			world["box"],
			world["bag"],
			world["plaque"],
			NewThing(
				NewNarrative(),
				NewName("an ornate fireplace"),
				NewDescription("This is a very ornate fireplace carved from marble. Either side a dragon curls downward until the head is below the fire looking upward, giving the impression that they are breathing fire."),
				NewAlias("fireplace"),
				NewVetoes(
					NewVeto("get", "You try and rip the ornate fireplace out of the wall but it's just too heavy."),
				),
			),
		),
		NewExits(),
	)

	world["loc2"] = NewThing(
		NewName("Common Room"),
		NewDescription("You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away."),
		NewAlias("tavern", "common"),
		NewInventory(
			world["chairs"],
			world["tables"],
		),
		NewExits(),
	)

	world["loc3"] = NewThing(
		NewName("Tavern Entrance"),
		NewDescription("You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street."),
		NewAlias("tavern", "entrance"),
		NewInventory(),
		NewExits(),
	)

	world["loc4"] = NewThing(
		NewName("Tavern Bar"),
		NewDescription("You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here."),
		NewAlias("tavern", "bar"),
		NewInventory(),
		NewVetoes(
			NewVeto("drop", "No littering in the bar!"),
		),
		NewExits(),
	)

	world["loc5"] = NewThing(
		NewName("Street between Tavern and Bakers"),
		NewDescription("You are on a well kept cobbled street. Buildings looming up either side of you. To the east the smells of a bakery taunt you, west there is the entrance to a tavern. A sign above the tavern door proclaims it as the Dragon's Breath. The street continues to the north and south."),
		NewAlias("tavern", "bakers", "street"),
		NewInventory(),
		NewExits(),
	)

	world["loc6"] = NewThing(
		NewName("Baker's shop"),
		NewDescription("You are standing in a bakers shop. Low tables show an array of fresh breads, cakes and the like. The smells here are beyond description."),
		NewAlias("SHOP", "BAKERS"),
		NewInventory(),
		NewExits(),
	)

	world["loc7"] = NewThing(
		NewStart(),
		NewName("Street outside pawn shop"),
		NewDescription("You are on a well kept cobbled street that runs north and south. To the east You can see a small Pawn shop. Southward you can see a large fountain and northward the smell of a bakery teases you."),
		NewAlias("STREET", "PAWNSHOP"),
		NewInventory(),
		NewExits(),
	)

	world["loc8"] = NewThing(
		NewName("Pawn shop"),
		NewDescription("You are in small Pawn shop. All around you on shelves are what looks like a load of useless junk."),
		NewAlias("SHOP", "PAWN", "PAWNSHOP"),
		NewInventory(),
		NewExits(),
	)

	world["loc9"] = NewThing(
		NewName("Fountain Square"),
		NewDescription("You are in a small square at the crossing of two roads. In the centre of the square a magnificent fountain has been erected, providing fresh water to any who want it. From here the streets lead off in all directions."),
		NewAlias("FOUNTAIN"),
		NewInventory(),
		NewExits(),
	)

	world["loc10"] = NewThing(
		NewName("Street outside armourer"),
		NewDescription("You are on a well kept cobbled street which runs to the east and west. To the south you can see the shop of an armourer."),
		NewAlias("STREET", "ARMOURER"),
		NewInventory(),
		NewExits(),
	)

	// Link up location exits
	var e has.Exits

	e = FindExits(world["loc1"])
	e.Link(South, FindInventory(world["loc2"]))
	e.Link(East, FindInventory(world["loc3"]))
	e.Link(Southeast, FindInventory(world["loc4"]))

	e = FindExits(world["loc2"])
	e.Link(North, FindInventory(world["loc1"]))
	e.Link(Northeast, FindInventory(world["loc3"]))
	e.Link(East, FindInventory(world["loc4"]))

	e = FindExits(world["loc3"])
	e.Link(West, FindInventory(world["loc1"]))
	e.Link(Southwest, FindInventory(world["loc2"]))
	e.Link(South, FindInventory(world["loc4"]))
	e.Link(East, FindInventory(world["loc5"]))

	e = FindExits(world["loc4"])
	e.Link(North, FindInventory(world["loc3"]))
	e.Link(Northwest, FindInventory(world["loc1"]))
	e.Link(West, FindInventory(world["loc2"]))

	e = FindExits(world["loc5"])
	//e.Link(North, FindInventory(world["loc14"]))
	e.Link(South, FindInventory(world["loc7"]))
	e.Link(East, FindInventory(world["loc6"]))
	e.Link(West, FindInventory(world["loc3"]))

	e = FindExits(world["loc6"])
	e.Link(West, FindInventory(world["loc5"]))

	e = FindExits(world["loc7"])
	e.Link(North, FindInventory(world["loc5"]))
	e.Link(East, FindInventory(world["loc8"]))
	e.Link(South, FindInventory(world["loc9"]))

	e = FindExits(world["loc8"])
	e.Link(West, FindInventory(world["loc7"]))

	e = FindExits(world["loc9"])
	e.Link(North, FindInventory(world["loc7"]))
	//e.Link(South, FindInventory(world["loc50"]))
	//e.Link(East, FindInventory(world["loc12"]))
	e.Link(West, FindInventory(world["loc10"]))

	e = FindExits(world["loc10"])
	//e.Link(South, FindInventory(world["loc11"]))
	e.Link(East, FindInventory(world["loc9"]))
	//e.Link(West, FindInventory(world["loc24"]))

	return world
}

// loadZone loads the recordjar specified by the filename and returns an
// unmarshaled zone.
func loadZone(filename string) (zone map[string]has.Thing) {

	zone = make(map[string]has.Thing)

	// Try and open the data file
	f, err := os.Open(filepath.Join(config.Server.DataDir, filename))
	if err != nil {
		log.Printf("Error loading %s: %s", filename, err)
		return
	}

	// Read the data into a jar and close the data file
	jar := recordjar.Read(f, "description")
	if err := f.Close(); err != nil {
		log.Printf("Error closing %s: %s", filename, err)
		return
	}

	log.Printf("Loading: %s", filename)

	// Go through the records in the jar. For each record unmarshal a Thing
	for i, record := range jar {

		t := NewThing()

		// If we don't have a reference we can't add the Thing to the zone
		if _, ok := record["ref"]; !ok {
			log.Printf("[Record %d]: Not added to zone, no ref.", i)
			continue
		}

		ref := recordjar.Decode.Keyword(record["ref"])
		t.Unmarshal(i, record)

		// Log a warning if we are going to overwrite a Thing by using the same
		// reference more than once - but don't prevent it.
		if _, ok := zone[ref]; ok {
			log.Printf("[Record %d/Ref %s] Warning: duplicate ref, not overwriting.", i, ref)
			continue
		}

		// Finally add Thing to the zone
		zone[ref] = t
	}

	// Post unmarshaling processing

	log.Printf("Checking exits...")
	checkExitsHaveInventory(zone)

	log.Printf("Linking exits...")
	linkupExits(zone, jar)

	log.Printf("Populating inventories...")
	linkupInventory(zone, jar)

	log.Printf("Populating things...")
	linkupThings(zone, jar)

	return
}

// checkExitsHaveInventory makes sure that all locations in a zone have an
// Inventory attribute. Locations are identified as any Thing with an Exits
// attribute. If a location does not also have an Inventory nothing can be put
// into the location.
func checkExitsHaveInventory(zone map[string]has.Thing) {
	for _, t := range zone {
		// If we have no exits we don't have to worry about an inventory
		if !FindExits(t).Found() {
			continue
		}
		// If we have an inventory we don't have to worry about adding one
		if FindInventory(t).Found() {
			continue
		}
		// Add required Inventory
		t.Add(NewInventory())
	}
}

// linkupExits sets up all exits for a zone. For each location in the given
// zone the Exits are linked to the respective destination location Inventory
// attributes. This cannot be done during unmarshaling as we cannot link from
// one location to another if either of them have not been unmarshaled yet.
func linkupExits(zone map[string]has.Thing, jar recordjar.Jar) {

	var (
		data  []byte
		ref   string
		exits [][2]string
		to    has.Thing
		ok    bool
	)

	for _, r := range jar {

		// Get reference from record
		if data, ok = r["ref"]; !ok {
			continue
		}
		ref = recordjar.Decode.Keyword(data)

		// Get exits from record
		if data, ok = r["exits"]; !ok {
			continue
		}
		exits = recordjar.Decode.PairList(data)

		e := FindExits(zone[ref])

		for _, pair := range exits {
			d, r := pair[0], pair[1]
			if to, ok = zone[r]; !ok {
				continue
			}
			dir, _ := e.NormalizeDirection(d)
			e.Link(dir, FindInventory(to))
		}
	}
}

// linkupInventory puts Thing into an Inventory as specified by the inventory
// field in a recordjar record. That is:
//
//	Inventory: O1 O2
//
// This says put Things with references O1 and O2 into this inventory. This
// cannot be done during unmarshaling of the Inventory as the Things may not be
// unmarshaled yet.
//
// BUG(diddymus): linkupInventory does not check if a Thing is adding itself to
// it's own Inventory.
func linkupInventory(zone map[string]has.Thing, jar recordjar.Jar) {

	var (
		data []byte
		ref  string
		inv  []string
		ok   bool
	)

	for _, r := range jar {

		// Get reference from record
		if data, ok = r["ref"]; !ok {
			continue
		}
		ref = recordjar.Decode.Keyword(data)

		// Get inventory from record
		if data, ok = r["inventory"]; !ok {
			continue
		}

		inv = recordjar.Decode.KeywordList(data)
		i := FindInventory(zone[ref])

		for _, ref := range inv {
			i.Add(zone[ref])
		}
	}
}

// linkupThings puts Thing into an Inventory as specified by the location field
// in a recordjar record. That is:
//
//	Location: L1 L2
//
// This says that the Thing with the location field should be put into the
// Inventory that have the references L1 and L2.
//
// BUG(diddymus): linkupThings does not check if a Thing is adding itself to
// it's own Inventory.
func linkupThings(zone map[string]has.Thing, jar recordjar.Jar) {

	var (
		data []byte
		ref  string
		inv  []string
		ok   bool
	)

	for _, r := range jar {

		// Get reference from record
		if data, ok = r["ref"]; !ok {
			continue
		}
		ref = recordjar.Decode.Keyword(data)

		// Get location from record
		if data, ok = r["location"]; !ok {
			continue
		}

		inv = recordjar.Decode.KeywordList(data)

		for _, ref2 := range inv {
			i := FindInventory(zone[ref2])
			i.Add(zone[ref])
		}
	}
}
