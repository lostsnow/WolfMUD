// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"code.wolfmud.org/WolfMUD.git/entities/is"

	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// unmarshalers is a map of Unmarshalers keyed by a string 'type name'. Unmarshalers
// for different types call Register to get added to the map. See Register for
// more details.
var unmarshalers map[string]Unmarshaler

// init makes the unmarshalers map so we don't try referencing a nil map which would
// cause a panic.
func init() {
	unmarshalers = make(map[string]Unmarshaler)
}

// Register is used to register an unmarshaler for a type. When a Record is
// unmarshaled it's type attribute is extracted. This is then used as the key
// for looking up the registered umarshaler which is then passed the Record for
// unmarshaling. The name used for the key is uppercased - in effect making it
// case insensitive.
func Register(name string, u Unmarshaler) {
	name = strings.ToUpper(name)
	if _, ok := unmarshalers[name]; !ok {
		unmarshalers[name] = u
	} else {
		panic("Tried to register duplicate unmarshaler: " + name)
	}
	log.Printf("Unmarshaler registered: %T (%s)", u, name)
}

// LoadDir is a helper to load all WolfMUD recordjar files in a directory and
// unmarshal their content. The configuration file config.wrj is excluded so it
// is not processed twice. Any sub directories found are not processed. The
// passed directory name does not need to end with a directory separator.
// Processing of the found data files is handed over to LoadFile.
func LoadDir(dir string) {

	dir = filepath.Join(dir, "*.wrj")
	dir = filepath.Clean(dir)

	filenames, err := filepath.Glob(dir)

	if err != nil {
		log.Printf("Failed to find data files: %s", err)
		return
	}

	log.Printf("Processing data files: %d found", len(filenames))

	for _, filename := range filenames {
		if filepath.Base(filename) == "config.wrj" {
			log.Printf("Ignoring configuration file: %s", filepath.Base(filename))
			continue
		}
		LoadFile(filename)
	}
}

// LoadFile is a helper to load the specified data file and unmarshal it.
// The unmarshaling is handed over to Unmarshal.
func LoadFile(filename string) {

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

	UnmarshalJar(&rj)
}

// UnmarshalJar unmarshals all of the Record found in a passed RecordJar. Each
// Record in a RecordJar is unmarshaled in two phases. First phase all Record
// are unmarshaled by calling UnmarshalRecord. This instantiates a concrete
// variable of the correct type for the Record. Second phase Init is called on
// each unmarshaled Record. This two phase setup allows unmarshaled Record to
// refer to each other. For example if items are to be put into a location the
// location must exist before items can be put into it. Also the items must
// exist before they can be placed in the location. The UnmarshalRecord creates
// the locations and items and the Init on items puts the items in their defined
// location.
//
// TODO: If an Unmarshaler has no reference one is generated of the form RECn
// where n is the Record index within the RecordJar. This save having processing
// for Unmarshalers with and without references. This probably need reviewing.
//
// TODO: Really hate the way we are passing around refs - needs sorting out.
func UnmarshalJar(rj *RecordJar) map[string]Unmarshaler {

	refs := make(map[string]Unmarshaler)

	// Unmarshal all Record in the RecordJar
	for i, rec := range *rj {
		if ur, err := UnmarshalRecord(rec); err != nil {
			log.Printf("Error unmarshaling: %s", err)
		} else {
			ref := rec.Keyword("ref")
			if ref == "" {
				ref = "REC" + strconv.Itoa(i)
				rec["ref"] = ref
			}
			if _, ok := refs[ref]; ok {
				log.Printf("Warning, overwriting ref: %s", ref)
			}
			refs[ref] = ur
		}
	}

	// Init all unmarshaled instances
	for _, rec := range *rj {
		r := rec.Keyword("ref")

		if zc, ok := refs[r]; ok {
			t := rec.Keyword("type")

			name := "Unnamed"
			if n, ok := zc.(is.Nameable); ok {
				name = n.Name()
			}

			log.Printf("Init: %s (%s, %s)", name, t, r)
			zc.Init(rec, refs)
		}
	}

	return refs
}

// UnmarshalRecord returns an Unmarshaler that has the underlying type that
// was registered using Register. The new Unmarshaler is created using the
// registered Unmarshaler as the 'template'. Unmarshal is then called on the new
// Unmarshaler passing it the Record. Values from the Record are then used to
// fill in the 'template'.
func UnmarshalRecord(r Record) (u Unmarshaler, err error) {

	// Without a type we cannot find the correct unmarshaler
	t := r.Keyword("type")
	if t == "" {
		return nil, &NoTypeError{r}
	}

	// Do we have a type but don't have a registered unmarshaler?
	u, ok := unmarshalers[t]
	if !ok {
		return nil, &UnknownTypeError{t}
	}

	// Create an empty, zero value copy of registered type and unmarshal the
	// current record into it.
	zc := reflect.New(reflect.ValueOf(u).Elem().Type()).Interface().(Unmarshaler)
	zc.Unmarshal(r)

	log.Printf("Loaded: %s (%s)", t, r.Keyword("ref"))
	return zc, nil
}

type NoTypeError struct {
	Record
}

func (e NoTypeError) Error() string {
	return fmt.Sprintf("No type specified: %#v", e.Record)
}

type UnknownTypeError struct {
	string // Type's name
}

func (e UnknownTypeError) Error() string {
	return fmt.Sprintf("Unknow type, unmarshaler not registered: %s", e)
}
