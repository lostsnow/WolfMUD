// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// The zones package implements WolfMUD's high level zone loading. Zones are
// used to group together multiple Thing into manageable groups. Zones can be
// used to put together a universe, a world, a city, a town, a complex building
// or a patch of forest. A zone can be a single Thing or a few thousand but
// typically each zone will have a few hundred Thing which represent everything
// within an area of the game. Each zone is then linked together to make up the
// complete game world.
//
// Each zone is represented as a simple plain text file laid out in the
// WolfMUD Record Jar format. It is the job of the zone package to coordinate
// the loading of these files and to assemble everything in the world. For
// more details on the format see the recordjar package.
package zones

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
)

// zone represents a self contained collection of Things. The Things are split
// into locations - those that have an Exits attribute - and everything else
// which gets added to the temporary store. The locations contain the actual
// game locations that will be used while the store is only populated
// temporally while the zone is being assembled.
type zone struct {
	ref       string
	name      string
	locations map[string]taggedThing // Things with Exit attributes
	store     map[string]taggedThing // Temp store of Things without Exit attributes
}

// taggedThing is a Thing that has been tagged with the Record from the
// recordjar it was created from. This is so that the original creation data is
// easily and readily available during assembly of a zone.
type taggedThing struct {
	has.Thing
	recordjar.Record
}

// newZone returns a new empty, initialised zone.
func newZone() zone {
	return zone{
		ref:       "ZONE_" + strconv.Itoa(len(zones)),
		name:      "Unknown",
		locations: make(map[string]taggedThing),
		store:     make(map[string]taggedThing),
	}
}

// zones is a collection of all of the currently loaded zones - the current
// game world.
var zones = map[string]zone{}

// Load loads all of the zone files.
func Load() {
	log.Printf("Loading zones")

	// Load each zone
	for _, path := range zoneFiles() {
		if z := loadZone(path); len(z.locations)+len(z.store) > 0 {
			zones[z.ref] = z
			continue
		}
	}

	// If no zones loaded create a default void
	if len(zones) == 0 {
		zones["VOID"] = createVoid()
	}

	linkupZones()
	detagLocations()
	checkDoorsHaveOtherSide()

	log.Printf("Finished loading %d zones.", len(zones))

	runtime.GC()

	return
}

// zoneFiles returns a list of zone file paths found in the data directory's
// zones subdirectory.
func zoneFiles() (paths []string) {
	pattern := filepath.Join(config.Server.DataDir, "zones", "*.wrj")
	log.Printf("Searching for zones matching: %s", pattern)
	paths, _ = filepath.Glob(pattern)
	return paths
}

// loadZone loads the single zone file specified by the passed path and
// returns a partially assembled zone. To complete the zone Doors need to be
// checked for the 'other side'.
//
// A zone will always be returned even if it is empty, in which case
// len(zone.locations) + len(zone.store) == 0.
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

	// Go through the records in the jar. For each record unmarshal a Thing and
	// store it with its record as a taggedThing in either zone.locations or
	// zone.store
	for i, record := range jar {

		// If the record has no reference we can't add the Thing to the zone
		if _, ok := record["REF"]; !ok {
			log.Printf("[Record %d]: Not added to zone, no reference found", i)
			continue
		}
		ref := recordjar.Decode.Keyword(record["REF"])

		t := attr.NewThing()
		t.Unmarshal(i, record)

		// Finally add Thing to locations if has Exits else add it to the temporary
		// store. Log a warning if we are going to overwrite a Thing by using the
		// same reference.
		if attr.FindExits(t).Found() {
			if _, ok := z.locations[ref]; ok {
				log.Printf("[Record %d] Warning: overwriting duplicate location reference %s", i, ref)
			}
			z.locations[ref] = taggedThing{t, record}
		} else {
			if _, ok := z.store[ref]; ok {
				log.Printf("[Record %d] Warning: overwriting duplicate reference reference %s", i, ref)
			}
			z.store[ref] = taggedThing{t, record}
		}

	}

	z.assemble()

	log.Printf("Loaded %s: %s (%s)", filename, z.name, z.ref)

	return z
}

// assemble takes care of piecing together a zone from Things. It links exits,
// links and populating inventories and other setup functions that cannot be
// done until a zone is fully unmarshaled due to dependencies between
// different Things. For example to link the exits between locations A and B
// both locations must be unmarshaled first. To put item A into container B
// both must be unmarshaled first.
func (z *zone) assemble() {
	z.linkupStoreInventory()
	z.linkupStoreLocation()
	z.checkExitsHaveInventory()
	z.linkupExits()
	z.linkupInventory()
	z.linkupLocation()
	z.closeStore()
}

// checkExitsHaveInventory makes sure that all locations in a zone have an
// Inventory attribute. Locations are identified as any Thing with an Exits
// attribute. If a location does not also have an Inventory nothing can be put
// into the location.
func (z *zone) checkExitsHaveInventory() {
	log.Printf("  Checking exits")
	for _, l := range z.locations {
		if !attr.FindInventory(l).Found() {
			l.Add(attr.NewInventory())
		}
	}
}

// linkupExits sets up all exits for a zone. For each location in the given
// zone the Exits are linked to the respective destination location's Inventory
// attribute. Linking between zones is handled by linkupZones.
//
// NOTE: Incomplete exit links are ignored for convenience so that markers can
// be left in zone files we are working on as reminders. For example, S and S→
// would be ignored. To generate warnings in the log use an invalid reference
// such as S→X
func (z *zone) linkupExits() {
	log.Printf("  Linking exits")
	for lref, l := range z.locations {
		from := attr.FindExits(l.Thing)
		for _, pair := range recordjar.Decode.PairList(l.Record["EXITS"]) {

			if pair[1] == "" { // Ignore incomplete links
				continue
			}

			edir, eref := pair[0], pair[1]
			ndir, err := from.NormalizeDirection(edir)
			if err != nil {
				log.Printf("Location %s: invalid direction ignored %s", lref, edir)
				continue
			}
			if _, ok := z.locations[eref]; !ok {
				log.Printf("Location %s: cannot link %s exit to missing ref %s", lref, from.ToName(ndir), eref)
				continue
			}
			from.Link(ndir, attr.FindInventory(z.locations[eref]))
		}
	}
}

// linkupStoreInventory links up Things in the store via the Inventory
// recordjar field using references NOT copies. This allows for complex Things
// in Things without having to do a lot of complex recursive searching. When a
// Thing is linked into the world then we make the copies.
func (z *zone) linkupStoreInventory() {
	log.Printf("  Linking temporary store inventories")
	for sref, s := range z.store {
		for _, tref := range recordjar.Decode.KeywordList(s.Record["INVENTORY"]) {
			t, ok := z.store[tref]
			if !ok {
				if _, ok = z.locations[tref]; !ok {
					log.Printf("Invalid Inventory reference: ref not found %s", tref)
				}
				continue
			}
			if isParent(s.Thing, t.Thing) {
				log.Printf("Recursive Inventory reference: cannot put %s into %s", tref, sref)
				continue
			}
			i := attr.FindInventory(s)
			i.AddDisabled(t.Thing)
			i.Enable(t.Thing)
		}
	}
}

// linkupStoreLocation links up Things in the store via the Location recordjar
// field using references NOT copies. This allows for complex Things in Things
// without having to do a lot of complex recursive searching. When a Thing is
// linked into the world then we make the copies. If a Thing is to be put
// somewhere and an Inventory does not exist one will be added automatically.
func (z *zone) linkupStoreLocation() {
	log.Printf("  Linking temporary store locations")
	for sref, s := range z.store {
		for _, tref := range recordjar.Decode.KeywordList(s.Record["LOCATION"]) {
			t, ok := z.store[tref]
			if !ok {
				if _, ok := z.locations[tref]; !ok {
					log.Printf("Invalid Location reference, ref not found %s", tref)
				}
				continue
			}
			if isParent(t.Thing, s.Thing) {
				log.Printf("Recursive Location reference: cannot put %s into %s", sref, tref)
				continue
			}
			i := attr.FindInventory(t)
			if !i.Found() {
				s.Add(attr.NewInventory(s.Thing))
				continue
			}
			i.AddDisabled(s.Thing)
			i.Enable(s.Thing)
		}
	}
}

// linkupInventory puts Things from the store into a location Inventory as
// specified by the inventory field in a recordjar record. That is:
//
//	Inventory: O1 O2
//
// This says put Things with references O1 and O2 into this inventory.
func (z *zone) linkupInventory() {
	log.Printf("  Copying (Inventory)")
	for _, l := range z.locations {
		i := attr.FindInventory(l)
		for _, iref := range recordjar.Decode.KeywordList(l.Record["INVENTORY"]) {
			s, ok := z.store[iref]
			if !ok {
				log.Printf("Invalid Inventory reference: ref not found %s", iref)
				continue
			}
			t := s.Copy()
			i.AddDisabled(t)
			i.Enable(t)
			t.SetOrigins()
		}
	}
}

// linkupLocation puts Thing from the store into a location Inventory as
// specified by the location field in a recordjar record. That is:
//
//	Location: L1 L2
//
// This says that the Thing with the location field should be put into the
// Inventory that have the references L1 and L2.
func (z *zone) linkupLocation() {
	log.Printf("  Copying (Location)")
	for _, s := range z.store {
		for _, ref := range recordjar.Decode.KeywordList(s.Record["LOCATION"]) {
			l, ok := z.locations[ref]
			if !ok {
				if _, ok = z.store[ref]; !ok {
					log.Printf("Invalid Inventory reference: ref not found %s", ref)
				}
				continue
			}
			t := s.Copy()
			i := attr.FindInventory(l)
			i.AddDisabled(t)
			i.Enable(t)
			t.SetOrigins()
		}
	}
}

// closeStore removed all data from the temporary store once a zone is
// assembled.
func (z *zone) closeStore() {
	log.Printf("  Closing down temporary store")
	for _, s := range z.store {
		for ref := range s.Record {
			delete(s.Record, ref)
		}
		s.Record = nil
		s.Thing.Free()
		s.Thing = nil
	}
	z.store = nil
}

// createVoid makes a default zone with a starting location which can be used
// if no other zones are loaded.
func createVoid() zone {
	z := newZone()
	z.locations["VOID"] = taggedThing{
		attr.NewThing(
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
		),
		nil,
	}
	return z
}

// checkDoorsHaveOtherSide creates the 'other side' of a door for Things with a
// Door attribute. This needs to be done after all zones have been loaded so
// that doors between zones work.
func checkDoorsHaveOtherSide() {
	log.Printf("  Checking other side")
	for _, z := range zones {
		for _, l := range z.locations {
			i := attr.FindInventory(l)
			for _, t := range append(i.Contents(), i.Narratives()...) {
				if d := attr.FindDoor(t); d.Found() {
					d.OtherSide()
				}
			}
		}
	}
}

// isParent returns true if adding t to p's Inventory creates a cyclic
// dependency, else false. A cyclic dependency is one where t or one of t's
// children is already in the Inventory of p or in the Inventory of one of p's
// ancestors.
func isParent(p, t has.Thing) bool {

	// parent checks if t is p or one of p's ancestors
	parent := func(p, t has.Thing) bool {
		for p := p; t != p; {
			w := attr.FindLocate(p).Where()
			if w == nil {
				return false
			}
			p = w.Parent()
		}
		return true
	}

	// Check t and t's children are not p or one of p's ancestors
	if !parent(p, t) {
		if i := attr.FindInventory(t); i.Found() {
			for _, t := range append(i.Contents(), i.Narratives()...) {
				if parent(p, t) {
					return true
				}
			}
		}
		return false
	}
	return true
}

// linkupZones processes Zonelinks records to link exits between different
// zones. This has to be delayed until all of the zones have been assembled.
//
// NOTE: Incomplete zonelinks are ignored for convenience so that markers can
// be left in zone files we are working on as reminders. For example, any of
// the following would be ignored: S S→ S→ZINARA S→ZINARA:
//
// To generate warnings in the log use an invalid reference such as S→X:X
func linkupZones() {

	log.Printf("  Linking zones")

	// Go through zones and locations we are linking from
	for fzref, zone := range zones {
		for flref, loc := range zone.locations {
			links, ok := loc.Record["ZONELINKS"]
			if !ok {
				continue
			}
			for _, pair := range recordjar.Decode.PairList(links) {

				if pair[1] == "" { // Ignore incomplete link
					continue
				}
				dir, link := pair[0], pair[1]

				// Check direction is valid
				from := attr.FindExits(loc)
				ndir, err := from.NormalizeDirection(dir)
				if err != nil {
					log.Printf("Cannot zonelink from zone: %s ref: %s, invalid direction: %s", fzref, flref, dir)
					continue
				}

				// split link into zone ref and location ref pairs we are linking to
				for _, pairs := range recordjar.Decode.PairList([]byte(link)) {

					if pairs[1] == "" { // Ignore incomplete link
						continue
					}
					tzref, tlref := pairs[0], pairs[1]

					// Check destination exists
					if _, ok := zones[tzref].locations[tlref]; !ok {
						log.Printf("Cannot zonelink %s %s (%s), destination not found: %s:%s ", fzref, flref, dir, tzref, tlref)
						continue
					}

					to := attr.FindInventory(zones[tzref].locations[tlref])
					from.Link(ndir, to)
				}
			}
		}
	}

}

// detagLocations removes the recordjar data from locations once all zones
// have been assembled.
func detagLocations() {
	log.Printf("  Detagging locations")
	for _, z := range zones {
		for _, l := range z.locations {
			for ref := range l.Record {
				delete(l.Record, ref)
			}
			l.Record = nil
		}
	}
}
