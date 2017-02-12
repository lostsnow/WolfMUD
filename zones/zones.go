// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// The zones package implements WolfMUD's high level build blocks. Zones are
// used to group together multiple Thing into manageable groups. Zones can be
// used to put together a universe, a world, a city, a town, a complex building
// or a patch of forest. A zone can be a single Thing or a few thousand but
// typically each zone will have a few hundred Thing which represent everything
// within an area of the game. Each zone is then linked together to make up the
// complete game world.
//
// Each zone is represented as a simple plain text file laid out in the WolfMUD
// Record Jar format. For more details on the format see the recordjar package.
package zones

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"

	"log"
	"os"
	"path/filepath"
	"strconv"
)

// zone represents a self contained collection of Things.
type zone struct {
	ref    string
	name   string
	things map[string]has.Thing
}

// newZone returns a new empty, initialised zone.
func newZone() zone {
	return zone{
		ref:    "ZONE_" + strconv.Itoa(len(zones)),
		name:   "Unknown",
		things: make(map[string]has.Thing),
	}
}

// zones is a collection of all of the currently loaded zones.
var zones = map[string]zone{}

// zonelink contains the bookkeeping information  we need to be able to
// initialise a zone link once all zones are loaded.
type zoneLink struct {
	zone string
	ref  string
	data []byte
}

// zoneLinks contains all of the zone links discovered while loading the zones.
var zoneLinks = []zoneLink{}

// Load loads all of the zone files it can find in the data directory's 'zones'
// subdirectory.
func Load() {

	// Get a list of zone files
	z := filepath.Join(config.Server.DataDir, "zones", "*.wrj")
	log.Printf("Searching for zones: %s", z)
	files, _ := filepath.Glob(z)

	// Load each zone
	for _, f := range files {
		if z := loadZone(f); len(z.things) > 0 {
			zones[z.ref] = z
			continue
		}
	}

	// If no zones loaded create a default void
	if l := len(zones); l > 0 {
		log.Printf("Loaded %d zones", l)
	} else {
		zones["VOID"] = createVoid()
	}

	log.Printf("Linking zones...")
	linkupZones()

	return
}

// loadZone loads the single zone file specified by the passed path and returns
// a zone. A zone will always be returned even if it is empty, in which case
// len(zone.things) == 0.
func loadZone(path string) zone {

	filename := filepath.Base(path)
	z := newZone()

	// Try and open the data file
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error loading %s: %s", filename, err)
		return z
	}

	// Read the data into a jar and close the data file
	jar := recordjar.Read(f, "description")
	if err := f.Close(); err != nil {
		log.Printf("Error closing %s: %s", filename, err)
		return z
	}

	// Did we find an empty jar?
	if len(jar) == 0 {
		log.Printf("Ignoring empty zone: %s", filename)
		return z
	}

	// check if the first record is a zone record
	if name, ok := jar[0]["ZONE"]; ok {
		z.name = recordjar.Decode.String(name)

		if ref, ok := jar[0]["REF"]; ok {
			z.ref = recordjar.Decode.Keyword(ref)
		}

		// Zone record finished with so dispose of it
		jar = jar[1:]
	}

	// Did jar just have a zone record?
	if len(jar) == 0 {
		log.Printf("Ignoring empty zone: %s", filename)
		return z
	}

	log.Printf("Loading %s: %s (%s)", filename, z.name, z.ref)

	// Go through the records in the jar. For each record unmarshal a Thing
	for i, record := range jar {

		t := attr.NewThing()

		// If we don't have a reference we can't add the Thing to the zone
		if _, ok := record["REF"]; !ok {
			log.Printf("[Record %d]: Not added to zone, no ref.", i)
			continue
		}
		ref := recordjar.Decode.Keyword(record["REF"])

		t.Unmarshal(i, record)

		// Log a warning if we are going to overwrite a Thing by using the same
		// reference more than once - but don't prevent it.
		if _, ok := z.things[ref]; ok {
			log.Printf("[Record %d/Ref %s] Warning: duplicate ref, not overwriting.", i, ref)
			continue
		}

		// Finally add Thing to the zone
		z.things[ref] = t

		// If we have any zone links record the details
		if _, ok := record["ZONELINKS"]; ok {
			zoneLinks = append(zoneLinks, zoneLink{z.ref, ref, record["ZONELINKS"]})
		}

	}

	z.zoneBookkeeping(jar)

	return z
}

// zoneBookkeeping takes care of linking and populating inventories. This
// cannot be done until a zone is fully loaded due to dependencies between
// Things. The jar passed should be the one used to create the zone. This is so
// that data from the jar that is only used for initialisation does not have to
// be stored anywhere else.
func (z *zone) zoneBookkeeping(jar recordjar.Jar) {

	//Nothing to do?
	if len(z.things) == 0 {
		return
	}

	log.Printf("Checking exits...")
	z.checkExitsHaveInventory()

	log.Printf("Linking exits...")
	z.linkupExits(jar)

	log.Printf("Populating inventories...")
	z.linkupInventory(jar)

	log.Printf("Populating things...")
	z.linkupThings(jar)

}

// createVoid makes a default zone with a starting location which can be used
// if no other zones are loaded.
func createVoid() zone {
	z := newZone()
	z.things["VOID"] = attr.NewThing(
		attr.NewStart(),
		attr.NewName("The Void"),
		attr.NewDescription("You are in a dark void. Around you nothing. No stars, no light, no heat and no sound."),
		attr.NewInventory(
			attr.NewThing(
				attr.NewNarrative(),
				attr.NewName("the void"),
				attr.NewAlias("VOID"),
				attr.NewVetoes(
					attr.NewVeto("EXAMINE", "You try to examine the void but there is nothing to examine."),
				),
			),
		),
	)
	return z
}

// checkExitsHaveInventory makes sure that all locations in a zone have an
// Inventory attribute. Locations are identified as any Thing with an Exits
// attribute. If a location does not also have an Inventory nothing can be put
// into the location.
func (z *zone) checkExitsHaveInventory() {
	for _, t := range z.things {
		// If we have no exits we don't have to worry about an inventory
		if !attr.FindExits(t).Found() {
			continue
		}
		// If we have an inventory we don't have to worry about adding one
		if attr.FindInventory(t).Found() {
			continue
		}
		// Add required Inventory
		t.Add(attr.NewInventory())
	}
}

// linkupExits sets up all exits for a zone. For each location in the given
// zone the Exits are linked to the respective destination location's Inventory
// attribute.
func (z *zone) linkupExits(jar recordjar.Jar) {

	var (
		data []byte
		to   has.Thing
		ok   bool
	)

	for _, r := range jar {

		// Get reference from record
		if data, ok = r["REF"]; !ok {
			continue
		}
		ref := recordjar.Decode.Keyword(data)

		// Get exits from record
		if data, ok = r["EXITS"]; !ok {
			continue
		}
		exitList := recordjar.Decode.PairList(data)

		from := attr.FindExits(z.things[ref])

		for _, pair := range exitList {
			d, r := pair[0], pair[1]

			dir, err := from.NormalizeDirection(d)
			if err != nil {
				log.Printf("Location %s: invalid direction ignored: %s", ref, d)
				continue
			}

			if to, ok = z.things[r]; !ok {
				log.Printf("Location %s: cannot link %s exit to missing ref %s", ref, from.ToName(dir), r)
				continue
			}

			from.Link(dir, attr.FindInventory(to))
		}
	}
}

// linkupInventory puts Thing into an Inventory as specified by the inventory
// field in a recordjar record. That is:
//
//	Inventory: O1 O2
//
// This says put Things with references O1 and O2 into this inventory.
func (z *zone) linkupInventory(jar recordjar.Jar) {

	var (
		data []byte
		ok   bool
	)

	for _, r := range jar {

		// Get reference from record
		if data, ok = r["REF"]; !ok {
			continue
		}
		ref := recordjar.Decode.Keyword(data)

		// Get inventory from record
		if data, ok = r["INVENTORY"]; !ok {
			continue
		}
		inv := recordjar.Decode.KeywordList(data)

		i := attr.FindInventory(z.things[ref])

		for _, r := range inv {
			if r == ref {
				log.Printf("Ref %s: Cannot put something into its own inventory", ref)
				continue
			}
			if _, ok := z.things[r]; !ok {
				log.Printf("Ref %s: Cannot put into inventory %s, ref not found", r, ref)
				continue
			}

			i.Add(z.things[r])
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
func (z *zone) linkupThings(jar recordjar.Jar) {

	var (
		data []byte
		t    has.Thing
		ok   bool
	)

	for _, r := range jar {

		// Get reference from record
		if data, ok = r["REF"]; !ok {
			continue
		}
		ref := recordjar.Decode.Keyword(data)

		// Get location from record
		if data, ok = r["LOCATION"]; !ok {
			continue
		}
		inv := recordjar.Decode.KeywordList(data)

		for _, i := range inv {
			if i == ref {
				log.Printf("Ref %s: Cannot locate something inside itself", ref)
				continue
			}
			if t, ok = z.things[i]; !ok {
				log.Printf("cannot put %s into %s, ref: %[2]s not found", ref, i)
				continue
			}

			i := attr.FindInventory(t)
			i.Add(z.things[ref])
		}
	}
}

// linkupZones processes zone linking information to link exits between
// different zones. This has to be delayed until all of the zones have been
// loaded and initialised.
func linkupZones() {
	for _, l := range zoneLinks {

		// We have to have Exits where we are linking from
		e := attr.FindExits(zones[l.zone].things[l.ref])
		if !e.Found() {
			e = attr.NewExits()
			zones[l.zone].things[l.ref].Add(e)
		}

		// Split the ZoneLink data into direction and a reference pairs
		for _, pair := range recordjar.Decode.PairList(l.data) {
			d, r := pair[0], pair[1]
			{
				// Split the reference into a zone and reference pair
				pair := recordjar.Decode.PairList([]byte(r))

				// Ignore incomplete zonelinks so that we can leave markers in the zone file
				if len(pair) == 0 {
					continue
				}
				z, r := pair[0][0], pair[0][1]

				// Check destination exists
				if _, ok := zones[z]; !ok {
					log.Printf("Zone %s: Destination zone not found for zone link: %s", l.zone, z)
					continue
				}
				if _, ok := zones[z].things[r]; !ok {
					log.Printf("Zone %s: Destination location not found for zone link: %s", z, r)
					continue
				}

				// We have to have an inventory where we are linking to
				i := attr.FindInventory(zones[z].things[r])
				if !i.Found() {
					i = attr.NewInventory()
					zones[z].things[r].Add(i)
				}

				d, _ := e.NormalizeDirection(d)
				e.Link(d, i)
			}
		}
	}

	// We can't make further use of the zone links so clear them down
	zoneLinks = zoneLinks[0:0]
}
