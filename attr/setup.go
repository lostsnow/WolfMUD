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

// Setup the world
func Setup() map[string]has.Thing {

	// Create the world!
	world := map[string]has.Thing{}

	// Load a zone
	zone := loadZone("zones/zinara.wrj")

	// Add zone to the world
	for r, t := range zone {
		world[r] = t
	}

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
