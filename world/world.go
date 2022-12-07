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

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/world/preprocessor"
)

type pkgConfig struct {
	zonePath string
}

// cfg setup by Config and should be treated as immutable and not changed.
var cfg pkgConfig

// Config sets up package configuration for settings that can't be constants.
// It should be called by main, only once, before anything else starts. Once
// the configuration is set it should be treated as immutable an not changed.
func Config(c config.Config) {
	cfg = pkgConfig{
		zonePath: filepath.Join(c.Server.DataPath, "zones", "*.wrj"),
	}
}

// taggedThing is a *Thing with additional information only stored during the
// loading process.
type taggedThing struct {
	*core.Thing
	inventory []string
	location  []string
	zoneLinks map[string]string
}

// Load creates the game world.
//
// BUG(diddymus): Load will populate core.World directly as a side effect of
// being called. The core package can't import the world package as it would
// cause a cyclic import.
func Load() {

	log.Printf("Loading zones from: %s", cfg.zonePath)

	refToUID := make(map[string]string)

	filenames, err := filepath.Glob(cfg.zonePath)
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
		preprocessor.Process(jar)
		jar = jar[1:]

		// Load everything into temporary store
		log.Print("  Loading temporary store")
		store := make(map[string]taggedThing)
		for _, record := range jar {
			ref := decode.Keyword(record["REF"])
			store[ref] = taggedThing{
				Thing:     core.NewThing(),
				inventory: decode.KeywordList(record["INVENTORY"]),
				location:  decode.KeywordList(record["LOCATION"]),
				zoneLinks: decode.PairList(record["ZONELINKS"]),
			}
			store[ref].As[core.Zone] = zref + ":"
			store[ref].Unmarshal(record)
		}

		// Resolve Inventory attributes in the store with pointer references. An
		// Inventory Ref with an exclamation mark '!' prefix indicates the item is
		// disabled and out of play.
		log.Print("  Linking temporary store inventories")
		for _, item := range store {
			for _, ref := range item.inventory {
				disabled := ref[0] == '!'
				if disabled {
					ref = ref[1:]
				}
				if what, ok := store[ref]; ok {
					if disabled {
						item.Out[what.Thing.As[core.UID]] = what.Thing
					} else {
						item.In[what.Thing.As[core.UID]] = what.Thing
					}
				} else {
					log.Printf("load warning, ref not found for inventory: %s\n", ref)
				}
			}
		}

		// Resolve Location attributes in the store with pointer references. A
		// Location Ref with an exclamation mark '!' prefix indicates the item is
		// disabled and out of play.
		log.Print("  Linking temporary store locations")
		for _, item := range store {
			for _, ref := range item.location {
				disabled := ref[0] == '!'
				if disabled {
					ref = ref[1:]
				}
				if where, ok := store[ref]; ok {
					if disabled {
						where.Out[item.Thing.As[core.UID]] = item.Thing
					} else {
						where.In[item.Thing.As[core.UID]] = item.Thing
					}
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
				c := item.Copy(true)
				core.World[c.As[core.UID]] = c
				if c.Is&core.Start == core.Start {
					core.WorldStart = append(core.WorldStart, c)
				}
				refToUID[c.As[core.Ref]] = c.As[core.UID]

				// Apply zonelinks to exits
				for name, ref := range item.zoneLinks {
					if ref != "" {
						c.As[core.DirRefToAs[core.NameToDir[name]]] = ref
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
	log.Print("Resolving exit refs to UIDs")
	for _, loc := range core.World {
		for _, dir := range core.DirRefToAs {
			if loc.As[dir] != "" {
				loc.As[dir] = refToUID[loc.As[dir]]
			}
		}
	}

	// Finish initialising all items in the world - this is done last so that all
	// location references have been resolved and we can have things like doors
	// between zones work properly.
	//
	// NOTE: If we didn't allow one side of a door to be in one zone and the
	// other side of the door to be in a different zone we could initialise when
	// copying to the world.
	log.Print("Final item setup")
	for _, loc := range core.World {
		loc.InitOnce(nil)
	}

	log.Printf("Total world locations: %d, starting locations: %d",
		len(core.World), len(core.WorldStart))
	log.Print("Genesis complete")
	return
}
