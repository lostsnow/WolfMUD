// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// TODO: The loader should read text files and parse them creating entities
// that are then loaded into the world. At the moment the file parser has not
// been written and the loader is hardcoded.
package loader

import (
	"code.wolfmud.org/WolfMUD.git/entities/location"
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/entities/thing/item"
	"code.wolfmud.org/WolfMUD.git/utils/config"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	refs map[string]thing.Interface
)

func Load() {
	refs = make(map[string]thing.Interface)
	defer func() { refs = nil }()

	if files, err := filepath.Glob(config.DataDir + "*.wrj"); err != nil {
		log.Printf("Failed to find data files: %s", err)
	} else {
		for _, file := range files {
			if !strings.HasSuffix(file, "config.wrj") {
				load(file)
			}
		}
	}
}

// Load creates entities and adds them to the world.
func load(filename string) {

	f, err := os.Open(filename)
	if err != nil {
		log.Printf("Failed to load data file: %s", err)
		return
	}
	defer f.Close()

	log.Printf("Loading data file: %s", filepath.Base(filename))

	rj, _ := recordjar.Read(f)
	var i thing.Interface
	var r, t string

	for _, rec := range rj {

		r = rec.String("ref")
		t = rec.String("type")

		switch strings.ToLower(t) {
		case "item":
			i = &item.Item{}
		case "basic":
			i = &location.Basic{}
		case "start":
			i = &location.Start{}
		default:
			i = nil
		}

		if i != nil {
			i.Unmarshal(rec)
			refs[r] = i
			log.Printf("Loaded: %s (%s)", refs[r].Name(), t)
		} else {
			log.Printf("Unknown type: %#v\n", t)
		}
	}

	for _, rec := range rj {
		if r, ok := refs[rec.String("ref")]; ok {
			r.Init(rec, refs)
		}
	}
}
