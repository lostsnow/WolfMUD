// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package world

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
)

// taggedThing is a *Thing with additional information only stored during the
// loading process.
type taggedThing struct {
	*core.Thing
	inventory []string
	location  []string
	zoneLinks map[string]string
}

const zoneDir = "../data/zones/*.wrj"

// Load creates the game world.
//
// FIXME(diddymus): Hard-coded zone files and paths.
//
// BUG(diddymus): Load will populate core.World directly as a side effect of
// being called. The core package can't import the world package as it would
// cause a cyclic import.
func Load() {

	log.Printf("Loading zones from: %s", zoneDir)

	// Stop the world while we are building it
	core.BWL.Lock()
	defer core.BWL.Unlock()

	core.World = make(map[string]*core.Thing)
	refToUID := make(map[string]string)

	filenames, err := filepath.Glob(zoneDir)
	if err != nil || len(filenames) == 0 {
		log.Fatalf("Cannot load any zone files. Server not started.")
		return
	}

	for _, fName := range filenames {

		f, err := os.Open(fName)
		if err != nil {
			log.Printf("Load error: %s\n", err)
			return
		}
		jar := recordjar.Read(f, "DESCRIPTION")
		f.Close()

		// Find zone header record
		if len(jar) < 1 || len(jar[0]["ZONE"]) == 0 {
			log.Printf("load warning, zone header not found, skipping: %s\n", fName)
		}

		zref := decode.String(jar[0]["REF"])
		zone := decode.String(jar[0]["ZONE"])
		disabled := decode.Boolean(jar[0]["DISABLED"])

		if disabled {
			log.Printf("Disabled %s: %s (%s)", filepath.Base(fName), zone, zref)
			continue
		}

		log.Printf("Loading %s: %s (%s)", filepath.Base(fName), zone, zref)
		PreProcessor(jar)
		jar = jar[1:]

		// Load everything into temporary store
		log.Print("  Loading temporary store")
		store := make(map[string]taggedThing)
		for _, record := range jar {
			ref := decode.String(record["REF"])
			store[ref] = taggedThing{
				Thing:     core.NewThing(),
				inventory: decode.KeywordList(record["INVENTORY"]),
				location:  decode.KeywordList(record["LOCATION"]),
				zoneLinks: decode.PairList(record["ZONELINKS"]),
			}
			store[ref].As[core.Zone] = zref
			store[ref].Unmarshal(record)
		}

		// Resolve inventory attributes in the store with references
		log.Print("  Linking temporary store inventories")
		for _, item := range store {
			for _, ref := range item.inventory {
				if what, ok := store[ref]; ok {
					item.In[what.Thing.As[core.UID]] = what.Thing
				} else {
					log.Printf("load warning, ref not found for inventory: %s\n", ref)
				}
			}
		}

		// Resolve location attributes in the store with references
		log.Print("  Linking temporary store locations")
		for _, item := range store {
			for _, ref := range item.location {
				if where, ok := store[ref]; ok {
					where.In[item.Thing.As[core.UID]] = item.Thing
				} else {
					log.Printf("load warning, ref not found for location: %s\n", ref)
				}
			}
		}

		// Copy locations to world, recording any starting locations - copying
		// resolves references as unique things.
		log.Print("  Copying to world")
		for _, item := range store {
			if item.Is&core.Location == core.Location {
				c := item.Copy()
				core.World[c.As[core.UID]] = c
				if c.Is&core.Start == core.Start {
					core.WorldStart = append(core.WorldStart, c.As[core.UID])
				}
				refToUID[c.As[core.Ref]] = c.As[core.UID]

				// Apply zonelinks to exits
				for dir, ref := range item.zoneLinks {
					if ref != "" {
						c.As[core.NameToDir[dir]] = ref
					}
				}
			}
		}

		// Tear down temporary store
		log.Printf("  Closing down temporary store: %d entries", len(store))
		for ref, item := range store {
			item.Free()
			delete(store, ref)
		}
		runtime.GC()
		log.Printf("Loaded %s: %s (%s)", filepath.Base(fName), zone, zref)
	}

	// Rewrite exits from Refs to UIDs as Refs only unique within a zone.
	log.Print("Linking exits")
	for _, loc := range core.World {
		for dir := range core.DirToName {
			if loc.As[dir] != "" {
				loc.As[dir] = refToUID[loc.As[dir]]
			}
		}
	}

	// Create other side of blockers as references so they share state
	log.Print("Checking other side")
	for _, loc := range core.World {
		for _, item := range loc.In {
			blocking := item.As[core.Blocker]
			if blocking == "" || item.As[core.Where] != "" {
				continue
			}
			item.As[core.Where] = loc.As[core.UID]
			otherUID := loc.As[core.NameToDir[blocking]]
			core.World[otherUID].In[item.As[core.UID]] = item
		}
	}

	log.Printf("Total world locations: %d, starting locations: %d",
		len(core.World), len(core.WorldStart))
	log.Print("Genesis complete")
	return
}
