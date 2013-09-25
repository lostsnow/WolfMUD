// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// TODO: The loader should read text files and parse them creating entities
// that are then loaded into the world. At the moment the file parser has not
// been written and the loader is hardcoded.
package recordjar

import (
	"code.wolfmud.org/WolfMUD.git/entities/is"

	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	loaders map[string]Unmarshaler
)

func init() {
	loaders = make(map[string]Unmarshaler)
}

func Register(name string, u Unmarshaler) {
	name = strings.ToUpper(name)
	if _, ok := loaders[name]; !ok {
		loaders[name] = u
	} else {
		panic("Duplicate loader registering: " + name)
	}
	log.Printf("Registered loader: %s", name)
}

func Load(dir string) {
	if files, err := filepath.Glob(dir + "*.wrj"); err != nil {
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
		log.Printf("Failed to open data file: %s", err)
		return
	}
	defer f.Close()

	log.Printf("Loading data file: %s", filepath.Base(filename))

	rj, err := Read(f)
	if err != nil {
		log.Printf("Failed to load data file: %s", err)
		return
	}

	Unmarshal(&rj)

}

func Unmarshal(rj *RecordJar) map[string]Unmarshaler {

	refs := make(map[string]Unmarshaler)
	var r, t, name string
	var zc Unmarshaler

	for _, rec := range *rj {

		r = rec.Keyword("ref")
		t = rec.Keyword("type")

		if t == "" {
			log.Printf("No type given: %#v", rec)
			continue
		}

		if i, ok := loaders[t]; ok {

			// Create an empty, zero value copy of registered type and unmarshal the
			// current record into it. Then store it in refs so Init functions can
			// refer to it if needed.
			zc = reflect.New(reflect.ValueOf(i).Elem().Type()).Interface().(Unmarshaler)
			zc.Unmarshal(rec)

			if n, ok := zc.(is.Nameable); ok {
				name = n.Name()
			} else {
				name = "Unnamed"
			}

			if r == "" {
				log.Printf("Loaded: %s (%s, not referable)", name, t)
				log.Printf("Init: %s (%s)", name, t)
				zc.Init(rec, refs)
			} else {
				refs[r] = zc
				log.Printf("Loaded: %s (%s, %s)", name, t, r)
			}
		} else {
			log.Printf("Unknown type: %s", t)
		}
	}

	for _, rec := range *rj {
		r = rec.Keyword("ref")

		if zc, ok := refs[r]; ok {

			if zc, ok := zc.(is.Nameable); ok {
				name = zc.Name()
			} else {
				name = "Unnamed"
			}

			log.Printf("Init: %s (%s, %s)", name, t, r)
			zc.Init(rec, refs)
		}
	}

	return refs
}
