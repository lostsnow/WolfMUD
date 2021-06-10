// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package world

import (
	"fmt"
	"os"
	"runtime"

	"code.wolfmud.org/WolfMUD.git/proc"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
)

// taggedThing is a *Thing with additional information only stored during the
// loading process.
type taggedThing struct {
	*proc.Thing
	inventory []string
	location  []string
	zoneLinks map[string]string
}

// Load creates the game world.
//
// FIXME(diddymus): Hard-coded zone files and paths.
//
// BUG(diddymus): Load will populate proc.World directly as a side effect of
// being called. The proc package can't import the world package as it would
// cause a cyclic import.
func Load() {

	proc.World = make(map[string]*proc.Thing)
	refToUID := make(map[string]string)

	for _, fName := range []string{
		"../data/zones/zinara.wrj",
		"../data/zones/zinara_south.wrj",
		"../data/zones/zinara_caves.wrj",
	} {

		f, err := os.Open(fName)
		if err != nil {
			fmt.Printf("Load error: %s\n", err)
			return
		}
		jar := recordjar.Read(f, "DESCRIPTION")
		f.Close()
		PreProcessor(jar)

		// Find zone header record
		if len(jar) < 1 || len(jar[0]["ZONE"]) == 0 {
			fmt.Printf("load warning, zone header not found, skipping: %s\n", fName)
		}

		zone := decode.String(jar[0]["REF"])
		jar = jar[1:]

		// Load everything into temporary store
		store := make(map[string]taggedThing)
		for _, record := range jar {
			ref := decode.String(record["REF"])
			store[ref] = taggedThing{
				Thing:     proc.NewThing(),
				inventory: decode.KeywordList(record["INVENTORY"]),
				location:  decode.KeywordList(record["LOCATION"]),
				zoneLinks: decode.PairList(record["ZONELINKS"]),
			}
			store[ref].As[proc.Zone] = zone
			store[ref].Unmarshal(record)
		}

		// Resolve inventory attributes in the store with references
		for _, item := range store {
			for _, ref := range item.inventory {
				if what, ok := store[ref]; ok {
					item.In = append(item.In, what.Thing)
				} else {
					fmt.Printf("load warning, ref not found for inventory: %s\n", ref)
				}
			}
		}

		// Resolve location attributes in the store with references
		for _, item := range store {
			for _, ref := range item.location {
				if where, ok := store[ref]; ok {
					where.In = append(where.In, item.Thing)
				} else {
					fmt.Printf("load warning, ref not found for location: %s\n", ref)
				}
			}
		}

		// Copy locations to world, recording any starting locations - copying
		// resolves references as unique things.
		for _, item := range store {
			if item.Is&proc.Location == proc.Location {
				c := item.Copy()
				proc.World[c.As[proc.UID]] = c
				if c.Is&proc.Start == proc.Start {
					proc.WorldStart = append(proc.WorldStart, c.As[proc.UID])
				}
				refToUID[c.As[proc.Ref]] = c.As[proc.UID]

				// Apply zonelinks to exits
				for dir, ref := range item.zoneLinks {
					if ref != "" {
						c.As[proc.NameToDir[dir]] = ref
					}
				}
			}
		}

		// Tear down temporary store
		for ref, item := range store {
			item.Free()
			delete(store, ref)
		}
		runtime.GC()
	}

	// Rewrite exits from Refs to UIDs as Refs only unique within a zone. Then
	// drop zone information as no longer required.
	for _, loc := range proc.World {
		for dir := range proc.DirToName {
			if loc.As[dir] != "" {
				loc.As[dir] = refToUID[loc.As[dir]]
			}
		}
	}

	// Create other side of blockers as references so they share state
	for _, loc := range proc.World {
		for _, item := range loc.In {
			blocking := item.As[proc.Blocker]
			if blocking == "" || item.As[proc.Where] != "" {
				continue
			}
			item.As[proc.Where] = loc.As[proc.UID]
			otherUID := loc.As[proc.NameToDir[blocking]]
			proc.World[otherUID].In = append(proc.World[otherUID].In, item)
		}
	}

	return
}
