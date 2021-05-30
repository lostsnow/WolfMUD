// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

import (
	"fmt"
	"io"
	"strings"
)

// Is Attributes
const (
	Start uint32 = 1 << iota
	Narrative
	Dark
	NPC
	Container
	Open
)

// Is value mapping to name.
var isNames = []string{
	"Start", "Narrative", "Dark", "NPC", "Container", "Open",
}

// isNames returns the names of the set flags separated by the OR (|) symbol.
func IsNames(is uint32) string {
	names := []string{}
	for x := len(isNames) - 1; x >= 0; x-- {
		if is&(1<<x) != 0 {
			names = append(names, isNames[x])
		}
	}
	return strings.Join(names, "|")
}

// As value keys
const (
	North uint32 = iota
	Northeast
	East
	Southeast
	South
	Southwest
	West
	Northwest
	Up
	Down
	Where
	Alias
	Writing
	Blocker
)

// As value mappings
var asNames = []string{
	"North", "Northeast", "East", "Southeast",
	"South", "Southwest", "West", "Northwest", "Up", "Down",
	"Where", "Alias", "Writing", "Blocker",
}

var (
	// NameToDir maps a long or short direction name to its As constant.
	NameToDir = map[string]uint32{
		"N": North, "NE": Northeast, "E": East, "SE": Southeast,
		"S": South, "SW": Southwest, "W": West, "NW": Northwest,
		"NORTH": North, "NORTHEAST": Northeast, "EAST": East, "SOUTHEAST": Southeast,
		"SOUTH": South, "SOUTHWEST": Southwest, "WEST": West, "NORTHWEST": Northwest,
		"UP": Up, "DOWN": Down,
	}

	// DirToName maps an As direction constant to the direction's long name.
	DirToName = map[uint32]string{
		North: "north", Northeast: "northeast", East: "east", Southeast: "southeast",
		South: "south", Southwest: "southwest", West: "west", Northwest: "northwest",
		Up: "up", Down: "down",
	}
)

// ReverseDir takes a direction value and returns the reverse or opposite
// direction. For example if passed the constant East it will return West. If
// the passed value is not one of the direction constants it will be returned
// unchanged.
func ReverseDir(dir uint32) uint32 {
	switch {
	case dir > Down:
		return dir
	case dir < Up:
		return dir ^ 1<<2
	default:
		return dir ^ 1
	}
}

var nextUID chan uint32

func init() {
	nextUID = make(chan uint32, 1)
	nextUID <- 0
}

// Thing is a basic one thing fits all type.
type Thing struct {
	Name        string
	Description string
	UID         uint32
	Is          uint32
	As          map[uint32]string
	In          []*Thing
}

func NewThing(name, description string) *Thing {
	uid := <-nextUID
	nextUID <- uid + 1
	return &Thing{
		UID:         uid,
		Name:        name,
		Description: description,
		As:          make(map[uint32]string),
	}
}

// Find looks for a Thing with the given alias in the provided list of Things
// inventories. If a matching Thing is found returns the Thing, the Thing who's
// Inventory it was in and the index in the inventory where it was found. If
// there is not match returns nill for the Thing, nil for the Inventory and an
// index of -1.
func Find(alias string, where ...*Thing) (*Thing, *Thing, int) {
	if alias == "" {
		return nil, nil, -1
	}
	for _, inv := range where {
		if inv == nil {
			continue
		}
		for idx, item := range inv.In {
			if item.As[Alias] == alias {
				return item, inv, idx
			}
		}
	}
	return nil, nil, -1
}

// Dump will write a pretty ASCII tree representing the details of a Thing.
// Some examples:
//
//	`- 0xc00000e048 *proc.Thing - CAT
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
//
//	`- 0xc00009c008 *proc.Thing - BAG
//	   |- Name - a bag
//	   |- Description - This is a simple cloth bag.
//	   |- Is - 00000000000000000000000000010000 (Container)
//	   |- As - len: 1
//	   |  `- [11] Alias: BAG
//	   `- In - len: 1, nil: false
//	      `- 0xc00009c010 *proc.Thing - APPLE
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

	lines := simpleFold(t.Description, width-len(indent)-20)
	p("%s%p %[2]T - UID: %d, %s", tree[last].i, t, t.UID, t.As[Alias])
	indent += tree[last].b
	p("%sName - %s", tree[false].i, t.Name)
	p("%sDescription - %s", tree[false].i, lines[0])
	for _, line := range lines[1:] {
		p("%-17s%s", tree[false].b, line)
	}
	p("%sIs - %032b (%s)", tree[false].i, t.Is, IsNames(t.Is))
	lIn, lAs := len(t.In), len(t.As)
	p("%sAs - len: %d", tree[false].i, lAs)
	for k, v := range t.As {
		lAs--
		p("%s%s[%2d] %2s: %s", tree[false].b, tree[lAs == 0].i, k, asNames[k], v)
	}
	p("%sIn - len: %d, nil: %t", tree[true].i, lIn, t.In == nil)
	w.Write([]byte(b.String()))
	for x, item := range t.In {
		item.dump(w, width, indent+tree[true].b, x == lIn-1)
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
