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
	As  map[asKey]string    // Single value for a key
	Any map[anyKey][]string // One or more values for a key
	Int map[intKey]int      // Integer values, counts and quantities
	In  Things              // Item's in a Thing (inventory)
	Who Things              // Who is here? Players @ location
	Is  isKey               // Bit flags for capabilities/state
}

// Things represents a group of Thing.
type Things map[string]*Thing

// nextUID is used to store the next unique identifier to be used for a new
// Thing. It is setup and initialised via the init function.
var nextUID chan uint

// init is used to setup and initialise the nextUID channel.
func init() {
	nextUID = make(chan uint, 1)
	nextUID <- 0
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
		Who: make(map[string]*Thing),
		Int: make(map[intKey]int),
	}
	t.As[UID] = fmt.Sprintf("#UID-%X", uid)
	t.Any[Alias] = append(t.Any[Alias], t.As[UID])
	return t
}

// Enable performs final setup for a Thing placed into the world. The Thing
// will have access to its inventory and surroundings. The passed parent is
// the UID of the containing inventory, for locations themselves this will be
// an empty string.
func (t *Thing) Enable(parent string) {
	for _, item := range t.In {
		item.Enable(t.As[UID])
	}

	// If it's a blocker setup the 'other side'
	if t.As[Blocker] != "" && t.As[Where] == "" {
		otherUID := World[parent].As[NameToDir[t.As[Blocker]]]
		World[otherUID].In[t.As[UID]] = t
	}

	// Set Where so that a thing knows where it is.
	if t.As[Where] == "" {
		t.As[Where] = parent
	}
}

// Unmarshal loads data from the passed Record into a Thing.
func (t *Thing) Unmarshal(r recordjar.Record) {
	for field, data := range r {
		switch field {
		case "ACTION":
			for k, v := range decode.PairList(r[field]) {
				b := []byte(v)
				switch k {
				case "AFTER":
					t.Int[ActionAfter] = int(decode.Duration(b).Seconds())
				case "JITTER":
					t.Int[ActionJitter] = int(decode.Duration(b).Seconds())
				}
			}
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
		case "CLEANUP":
			for k, v := range decode.PairList(r[field]) {
				b := []byte(v)
				switch k {
				case "AFTER":
					t.Int[CleanupAfter] = int(decode.Duration(b).Seconds())
				case "JITTER":
					t.Int[CleanupJitter] = int(decode.Duration(b).Seconds())
				}
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
		case "ONACTION":
			t.Any[OnAction] = decode.StringList(r["ONACTION"])
		case "ONCLEANUP":
			t.As[OnCleanup] = decode.String(r["ONCLEANUP"])
		case "ONRESET":
			t.As[OnReset] = decode.String(r["ONRESET"])
		case "REF":
			t.As[Ref] = decode.Keyword(r[field])
		case "RESET":
			for k, v := range decode.PairList(r[field]) {
				b := []byte(v)
				switch k {
				case "AFTER":
					t.Int[ResetAfter] = int(decode.Duration(b).Seconds())
				case "JITTER":
					t.Int[ResetJitter] = int(decode.Duration(b).Seconds())
				}
			}
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

// Copy returns a duplicate of the receiver Thing with only the UID and Who
// being different. We don't duplicate Who as this would duplicate players. The
// Thing's inventory will be copied recursively.
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
	for k, v := range t.Int {
		T.Int[k] = v
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

// Sort returns the receiver Things as a slice of the Things sorted by UID.
func (t Things) Sort() []*Thing {
	if t == nil || len(t) == 0 {
		return nil
	}
	ord := make([]*Thing, len(t))
	keys := make([]string, 0, len(t))
	for key := range t {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for x, key := range keys {
		ord[x] = t[key]
	}
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
	p("%sIs  - %032b (%s)", tree[false].i, t.Is, t.Is.setNames())
	lIn, lAs, lAny, lWho, lInt := len(t.In), len(t.As), len(t.Any), len(t.Who), len(t.Int)
	p("%sAs  - len: %d", tree[false].i, lAs)
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
		p("%s%s%s:", tree[false].b, tree[lAny == 0].i, anyNames[k])
		for kk, vv := range v {
			line := simpleFold(vv, width-len(indent)-len("|  |- [00]"))
			pad := strings.Repeat(" ", len("[00]"))
			p("%s%s%s[%2d] %s",
				tree[false].b, tree[lAny == 0].b, tree[kk == len(v)-1].i, kk, line[0])
			for _, line := range line[1:] {
				p("%s%s%s%s %s",
					tree[false].b, tree[lAny == 0].b, tree[kk == len(v)-1].b, pad, line)
			}
		}
	}
	p("%sInt - len: %d", tree[false].i, lInt)
	for k, v := range t.Int {
		lInt--
		p("%s%s %s: %d", tree[false].b, tree[lInt == 0].i, intNames[k], v)
	}
	p("%sWho - len: %d", tree[false].i, lWho)
	w.Write([]byte(b.String()))
	b.Reset()
	for _, who := range t.Who {
		lWho--
		who.dump(w, width, indent+tree[false].b, lWho == 0)
	}
	p("%sIn  - len: %d", tree[true].i, lIn)
	w.Write([]byte(b.String()))
	b.Reset()
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
