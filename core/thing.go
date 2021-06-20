// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
)

// Thing is used to represent any and all items in the game world.
type Thing struct {
	Is  isKey               // Bit flags for capabilities/state
	As  map[asKey]string    // Single value for a key
	Any map[anyKey][]string // One or more values for a key
	In  map[string]*Thing   // Item's in a Thing (inventory)
}

// Type definitions for Thing field keys.
type (
	isKey  uint32 // index for Thing.Is
	asKey  uint32 // index for Thing.As
	anyKey string // index for Thing.Any
)

// Constants for use as bitmasks with the Thing.Is field.
const (
	Container isKey = 1 << iota // A container, allows PUT/TAKE
	Dark                        // A dark location
	Location                    // Item is a location
	NPC                         // An NPC
	Narrative                   // A narrative item
	Open                        // An open item (e.g. door)
	Player                      // Is a player
	Start                       // A starting location
)

// Constants for use as keys in a Thing.As field. Comments provide expected
// values for each constant.
//
// NOTE: The first 10 direction constants are fixed and their values SHOULD NOT
// BE CHANGED. The other constants should be kept in alphabetical order as new
// ones are added.
const (

	// Location reference exit leads to ("L1")
	North asKey = iota
	Northeast
	East
	Southeast
	South
	Southwest
	West
	Northwest
	Up
	Down

	Blocker     // Name of direction being blocked ("E")
	Description // Item's description
	Name        // Item's name
	Ref         // Item's original reference (zone:ref or ref)
	UID         // Item's unique identifier
	VetoDrop    // Veto for DROP command
	VetoGet     // Veto for GET command
	VetoPut     // Veto PUT command for item
	VetoPutIn   // Veto for PUT command into container
	VetoTake    // Veto TAKE command for item
	VetoTakeOut // Veto for TAKE command from container
	Where       // Current location ref ("L1")
	Writing     // Description of writing on an item
	Zone        // Zone item's definition loaded from
)

// Constants for Thing.Any keys
const (
	Alias     anyKey = "ALIAS"     // Aliases for an item
	Qualifier anyKey = "QUALIFIER" // Alias qualifiers
)

// nextUID is used to store the next unique identifier to be used for a new
// Thing. It is setup and initialised via the init function.
var nextUID chan uint

// init is used to setup and initialise the nextUID channel.
func init() {
	nextUID = make(chan uint, 1)
	nextUID <- 0
}

// isNames provides the string names for the Thing.Is bitmasks. The helper
// function IsNames can be used to retrieve a list of names for the bits set in
// a Thing.Is fields.
var isNames = []string{
	"Container",
	"Dark",
	"Location",
	"NPC",
	"Narrative",
	"Open",
	"Player",
	"Start",
}

// setNames returns the names of the set bits in a Thing.Is field. Names are
// separated by the OR (|) symbol. For example: "Narrative|Open".
func (is isKey) setNames() string {
	names := []string{}
	for x := len(isNames) - 1; x >= 0; x-- {
		if is&(1<<x) != 0 {
			names = append(names, isNames[x])
		}
	}
	return strings.Join(names, "|")
}

// asNames provides the string names for the Thing.As field constants. A name
// for a specific Thing.As value can be retrieved by simple indexing. For
// example: asNames[Alias] returns the string "Alias".
var asNames = []string{
	"North", "Northeast", "East", "Southeast",
	"South", "Southwest", "West", "Northwest",
	"Up", "Down",

	"Blocker",
	"Description",
	"Name",
	"Reference",
	"UID",
	"VetoDrop",
	"VetoGet",
	"VetoPut",
	"VetoPutIn",
	"VetoTake",
	"VetoTakeOut",
	"Where",
	"Writing",
	"Zone",
}

var (
	// NameToDir maps a long or short direction name to its Thing.As constant.
	NameToDir = map[string]asKey{
		"N": North, "NE": Northeast, "E": East, "SE": Southeast,
		"S": South, "SW": Southwest, "W": West, "NW": Northwest,
		"U": Up, "D": Down,
		"NORTH": North, "NORTHEAST": Northeast, "EAST": East, "SOUTHEAST": Southeast,
		"SOUTH": South, "SOUTHWEST": Southwest, "WEST": West, "NORTHWEST": Northwest,
		"UP": Up, "DOWN": Down,
	}

	// DirToName maps a Thing.As direction constant to the direction's long name.
	DirToName = map[asKey]string{
		North: "north", Northeast: "northeast", East: "east", Southeast: "southeast",
		South: "south", Southwest: "southwest", West: "west", Northwest: "northwest",
		Up: "up", Down: "down",
	}
)

// ReverseDir returns the reverse or opposite direction. For example if passed
// the constant East it will return West. If the passed value is not one of the
// direction constants it will be returned unchanged.
func (dir asKey) ReverseDir() asKey {
	switch {
	case dir > Down:
		return dir
	case dir < Up:
		return dir ^ 1<<2
	default:
		return dir ^ 1
	}
}

// NewThing returns a new initialised Thing with no properties set.
//
// TODO(diddymus): UID needs adding as an alias once we have multiple aliases.
func NewThing() *Thing {
	uid := <-nextUID
	nextUID <- uid + 1
	t := &Thing{
		As:  make(map[asKey]string),
		Any: make(map[anyKey][]string),
		In:  make(map[string]*Thing),
	}
	t.As[UID] = fmt.Sprintf("#UID-%X", uid)
	t.Any[Alias] = append(t.Any[Alias], t.As[UID])
	return t
}

// Unmarshal loads data from the passed Record into a Thing.
func (t *Thing) Unmarshal(r recordjar.Record) {
	for field, data := range r {
		switch field {
		case "ALIAS", "ALIASES":
			a := make(map[string]struct{})
			q := make(map[string]struct{})
			for _, alias := range decode.KeywordList(r[field]) {
				parts := strings.Split(alias, ":")
				switch {
				case len(parts) == 0:
					// Ignore empty aliases
				case len(parts) == 1 && parts[0][0] == '+':
					q[parts[0][1:]] = struct{}{}
				case len(parts) == 1:
					a[parts[0]] = struct{}{}
				case len(parts) == 2:
					q[alias[1:]] = struct{}{}
					a[parts[1]] = struct{}{}
				}
			}
			for alias := range a {
				t.Any[Alias] = append(t.Any[Alias], alias)
			}
			for qualifier := range q {
				t.Any[Qualifier] = append(t.Any[Qualifier], qualifier)
			}
		case "DESCRIPTION":
			t.As[Description] = decode.String(data)
		case "DOOR":
			for field, data := range decode.PairList(r["DOOR"]) {
				switch field {
				case "EXIT":
					t.As[Blocker] = data
				case "OPEN":
					if decode.Boolean([]byte(data)) {
						t.Is |= Open
					}
				default:
					//fmt.Printf("Unknown attribute: %s\n", field)
				}
			}
		case "EXIT", "EXITS":
			for dir, loc := range decode.PairList(r["EXITS"]) {
				t.As[NameToDir[dir]] = loc
			}
			t.Is |= Location
		case "INV", "INVENTORY":
			t.Is |= Container
		case "LOCATION":
			// Do nothing - only used by loader
		case "NAME":
			t.As[Name] = decode.String(data)
		case "NARRATIVE":
			t.Is |= Narrative
		case "REF":
			t.As[Ref] = decode.Keyword(r[field])
		case "START":
			t.Is |= Start
		case "VETO":
			for cmd, msg := range decode.KeyedStringList(r[field]) {
				switch cmd {
				case "DROP":
					t.As[VetoDrop] = msg
				case "GET":
					t.As[VetoGet] = msg
				case "PUT":
					t.As[VetoPut] = msg
				case "PUTIN":
					t.As[VetoPutIn] = msg
				case "TAKE":
					t.As[VetoTake] = msg
				case "TAKEOUT":
					t.As[VetoTakeOut] = msg
				default:
					//fmt.Printf("Unknown veto: %s, for: %s\n", cmd, t.As[Name])
				}
			}
		case "WRITING":
			t.As[Writing] = decode.String(data)
		case "ZONELINKS":
			// Do nothing - only used by loader
		default:
			//fmt.Printf("Unknown field: %s\n", field)
		}
	}

	// If it's a location it's not a container
	if t.Is&Location == Location {
		t.Is &^= Container
	}

	// If zone information present append it to Ref, any exits, then discard.
	if t.As[Zone] != "" {
		t.As[Ref] = t.As[Zone] + ":" + t.As[Ref]

		for dir := range DirToName {
			if t.As[dir] != "" {
				t.As[dir] = t.As[Zone] + ":" + t.As[dir]
			}
		}
		delete(t.As, Zone)
	}
}

// Copy returns a duplicate of the receiver Thing with only the UID being
// different. The Thing's inventory will be copied recursively.
func (t *Thing) Copy() *Thing {
	T := NewThing()
	T.Is = t.Is
	for k, v := range t.As {
		if k == UID {
			continue
		}
		T.As[k] = v
	}
	for k, v := range t.Any {
		T.Any[k] = v
	}
	for _, item := range t.In {
		c := item.Copy()
		T.In[c.As[UID]] = c
	}

	// Swap old UID alias for copy's new UID
	for x, alias := range T.Any[Alias] {
		if alias == t.As[UID] {
			T.Any[Alias][x] = T.As[UID]
			break
		}
	}

	return T
}

// SortedIn returns the Thing's inventory as a slice with elements sorted by UID.
func (t *Thing) SortedIn() []*Thing {
	if t == nil {
		return nil
	}
	ord := make([]*Thing, 0, len(t.In))
	for _, item := range t.In {
		ord = append(ord, item)
	}
	sort.Slice(ord, func(i, j int) bool { return ord[i].As[UID] < ord[j].As[UID] })
	return ord
}

// Free recursively unlinks everything from a Thing. This is not really
// necessary, but makes it easier for the garbage collector.
func (t *Thing) Free() {
	if t == nil {
		return
	}
	t.Is = 0
	for k := range t.As {
		delete(t.As, k)
	}
	t.As = nil
	for k, item := range t.In {
		item.Free()
		t.In[k] = nil
	}
	t.In = nil
}

// Dump will write a pretty ASCII tree representing the details of a Thing.
// A simple, single item:
//
//	`- 0xc00000e048 *core.Thing - CAT
//	   |- Name - the tavern cat
//	   |- Description - The tavern cat is a ball of fur with one golden eye, the
//	   |                other eye replaced by a large scar. It senses you
//	   |                watching it and returns your gaze with a steady one of
//	   |                its own.
//	   |- Is - 00000000000000000000000000001000 (NPC)
//	   |- As - len: 1
//	   |  `- [11] Alias: CAT
//	   `- In - len: 0, nil: true
//
// A container with an item in its inventory:
//
//	`- 0xc00009c008 *core.Thing - BAG
//	   |- Name - a bag
//	   |- Description - This is a simple cloth bag.
//	   |- Is - 00000000000000000000000000010000 (Container)
//	   |- As - len: 1
//	   |  `- [11] Alias: BAG
//	   `- In - len: 1, nil: false
//	      `- 0xc00009c010 *core.Thing - APPLE
//	         |- Name - an apple
//	         |- Description - This is a red apple.
//	         |- Is - 00000000000000000000000000000000 ()
//	         |- As - len: 1
//	         |  `- [11] Alias: APPLE
//	         `- In - len: 0, nil: true
//
func (t *Thing) Dump(w io.Writer, width int) {
	t.dump(w, width, "", true)
}

// Tree drawing parts for dump method
var tree = map[bool]struct{ i, b string }{
	false: {i: "|- ", b: "|  "}, // tree item/branch when end=false
	true:  {i: "`- ", b: "   "}, // tree item/branch when end=true
}

// dump implements the core functionality of the Dump method which just passes
// some initial values to dump.
func (t *Thing) dump(w io.Writer, width int, indent string, last bool) {
	var b strings.Builder

	p := func(f string, a ...interface{}) {
		b.WriteString(indent)
		fmt.Fprintf(&b, f, a...)
		b.WriteByte('\n')
	}

	p("%s%p %[2]T - UID: %s (%s)", tree[last].i, t, t.As[UID], t.As[Name])
	indent += tree[last].b
	p("%sIs - %032b (%s)", tree[false].i, t.Is, t.Is.setNames())
	lIn, lAs, lAny := len(t.In), len(t.As), len(t.Any)
	p("%sAs - len: %d", tree[false].i, lAs)
	for k, v := range t.As {
		lAs--
		line := simpleFold(v, width-len(indent)-len(asNames[k])-len("|  |- [00] : "))
		pad := strings.Repeat(" ", len(asNames[k])+len("[00] : "))
		p("%s%s[%2d] %s: %s", tree[false].b, tree[lAs == 0].i, k, asNames[k], line[0])
		for _, line := range line[1:] {
			p("%s%s%s%s", tree[false].b, tree[lAs == 0].b, pad, line)
		}
	}
	p("%sAny - len: %d", tree[false].i, lAny)
	for k, v := range t.Any {
		lAny--
		p("%s%s %s: %q", tree[false].b, tree[lAny == 0].i, k, v)
	}
	p("%sIn - len: %d", tree[true].i, lIn)
	w.Write([]byte(b.String()))
	for _, item := range t.In {
		lIn--
		item.dump(w, width, indent+tree[true].b, lIn == 0)
	}
}

// simpleFold folds a line of text returning multiple lines of the given width.
//
// BUG(diddymus): This function is not Unicode safe and embeded line feeds will
// probably produce an undesireable result.
func simpleFold(s string, width int) (lines []string) {
	var b strings.Builder
	for _, word := range strings.Fields(s) {
		if len(word)+b.Len()+1 > width {
			lines = append(lines, b.String())
			b.Reset()
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(word)
	}
	lines = append(lines, b.String())
	return
}
