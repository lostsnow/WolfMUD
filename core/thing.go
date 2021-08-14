// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"time"

	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
)

// Thing is used to represent any and all items in the game world.
//
// NOTE: If new fields are added to Thing they should be catered for in the
// NewThing and Free methods.
type Thing struct {
	As    map[asKey]string    // Single value for a key
	Any   map[anyKey][]string // One or more values for a key
	Int   map[intKey]int64    // Integer values, counts and quantities
	Ref   map[refKey]*Thing   // References to other Thing (e.g. Where)
	In    Things              // Item's in a Thing (inventory)
	Out   Things              // Item's out of play in a Thing
	Who   Things              // Who is here? Players @ location
	Is    isKey               // Bit flags for capabilities/state
	Event Events              // In-flight event timers
}

// Things represents a group of Thing indexed by UID.
type Things map[string]*Thing

// Events is used to store currently in-flight events for a Thing.
type Events map[eventKey]*time.Timer

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
		As:    make(map[asKey]string),
		Any:   make(map[anyKey][]string),
		Int:   make(map[intKey]int64),
		Ref:   make(map[refKey]*Thing),
		In:    make(map[string]*Thing),
		Out:   make(map[string]*Thing),
		Who:   make(map[string]*Thing),
		Event: make(map[eventKey]*time.Timer),
	}
	t.As[UID] = fmt.Sprintf("#UID-%X", uid)
	t.Any[Alias] = append(t.Any[Alias], t.As[UID])
	return t
}

// Enable performs final setup for a Thing placed into the world. The Thing
// will have access to its inventory and surroundings. The passed parent is
// the UID of the containing inventory, for locations themselves this will be
// an empty string.
func (t *Thing) Enable(parent *Thing) {

	// If it's a blocker setup the 'other side'
	if t.As[Blocker] != "" && t.Ref[Where] == parent {
		other := parent.Ref[NameToDir[t.As[Blocker]]]
		other.In[t.As[UID]] = t
	}

	// Hard-link exits - convert from Thing.As UIDs to Thing.Ref *Thing
	if t.Is&Location == Location {
		for refDir, asDir := range DirRefToAs {
			if t.As[asDir] != "" {
				t.Ref[refDir] = World[t.As[asDir]]
				delete(t.As, asDir)
			}
		}
	}

	// Check if we need to enable events
	if t.Int[ActionAfter] != 0 || t.Int[ActionJitter] != 0 {
		t.Schedule(Action)
	}

	if t.Is&Spawnable != Spawnable {
		t.Ref[Origin] = parent
	}

	for _, item := range t.In {
		item.Enable(t)
	}
	for _, item := range t.Out {
		item.Enable(t)
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
					t.Int[ActionAfter] = decode.Duration(b).Nanoseconds()
				case "JITTER":
					t.Int[ActionJitter] = decode.Duration(b).Nanoseconds()
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
					t.Int[CleanupAfter] = decode.Duration(b).Nanoseconds()
				case "JITTER":
					t.Int[CleanupJitter] = decode.Duration(b).Nanoseconds()
				}
			}
		case "DESCRIPTION":
			t.As[Description] = decode.String(data)
		case "DOOR":
			for field, data := range decode.PairList(r["DOOR"]) {
				b := []byte(data)
				switch field {
				case "RESET":
					t.Int[TriggerAfter] = decode.Duration(b).Nanoseconds()
				case "EXIT":
					t.As[Blocker] = data
				case "JITTER":
					t.Int[TriggerJitter] = decode.Duration(b).Nanoseconds()
				case "OPEN":
					if decode.Boolean([]byte(data)) {
						t.Is |= Open | _Open
					}
				default:
					//fmt.Printf("Unknown attribute: %s\n", field)
				}
				t.As[TriggerType] = "BLOCKER"
			}
		case "EXIT", "EXITS":
			for name, loc := range decode.PairList(r["EXITS"]) {
				t.As[DirRefToAs[NameToDir[name]]] = loc
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
					t.Int[ResetAfter] = decode.Duration(b).Nanoseconds()
				case "JITTER":
					t.Int[ResetJitter] = decode.Duration(b).Nanoseconds()
				case "SPAWN":
					t.Is |= Spawnable
				case "WAIT":
					t.Is |= Wait

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
				case "JUNK":
					t.As[VetoJunk] = msg
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

		for _, dir := range DirRefToAs {
			if t.As[dir] != "" {
				t.As[dir] = t.As[Zone] + ":" + t.As[dir]
			}
		}
		delete(t.As, Zone)
	}
}

// Copy returns a duplicate of the receiver Thing with only the UID and Who
// being different. We don't duplicate Who as this would duplicate players. If
// deep is set to true the Thing's inventory will be copied recursively.
func (t *Thing) Copy(deep bool) *Thing {
	T := NewThing()

	// Suspend any events in the original
	events := []eventKey{}
	for eventId := range t.Event {
		events = append(events, eventId)
		t.Suspend(eventId)
	}

	T.Is = t.Is
	for k, v := range t.As {
		if k == UID {
			continue
		}
		T.As[k] = v
	}
	for k, v := range t.Any {
		T.Any[k] = make([]string, len(t.Any[k]))
		copy(T.Any[k], v)
	}
	for k, v := range t.Int {
		T.Int[k] = v
	}
	for k, v := range t.Ref {
		T.Ref[k] = v
	}

	if deep {
		for _, item := range t.In {
			c := item.Copy(deep)
			T.In[c.As[UID]] = c
			c.Ref[Where] = T
		}
		for _, item := range t.Out {
			c := item.Copy(deep)
			T.Out[c.As[UID]] = c
			c.Ref[Where] = T
		}
	}

	// Swap old UID alias for copy's new UID
	for x, alias := range T.Any[Alias] {
		if alias == t.As[UID] {
			T.Any[Alias][x] = T.As[UID]
			break
		}
	}

	// enable events in original and copy
	for _, eventId := range events {
		t.Schedule(eventId)
		T.Schedule(eventId)
	}

	return T
}

// Spawn returns a copy of a spawnable Thing or the original if not spawnable.
// The returned copy will not be spawnable. If the original contains any
// spawnable items they will remain spawnable. If the original contains any
// non-spawnable items they will be moved to the copy.
func (t *Thing) Spawn() *Thing {
	if t.Is&Spawnable != Spawnable {
		return t
	}
	t.Ref[Where].Out[t.As[UID]] = t
	T := t.spawn()
	T.Is ^= Spawnable
	delete(T.Int, ResetAfter)
	delete(T.Int, ResetJitter)
	delete(T.Int, ResetDueAt)
	delete(T.Int, ResetDueIn)
	t.Schedule(Reset)
	return T
}

// spawn is a helper for Spawn and recursively spawns a Thing.
func (t *Thing) spawn() *Thing {
	if t.Is&Spawnable != Spawnable {
		return t
	}

	T := t.Copy(false)

	for ref, item := range t.In {
		if item.Is&Spawnable == Spawnable {
			item = item.spawn()
		} else {
			delete(t.In, ref)
		}
		T.In[item.As[UID]] = item
		item.Ref[Where] = T
	}
	for ref, item := range t.Out {
		if item.Is&Spawnable == Spawnable {
			item = item.spawn()
		} else {
			delete(t.Out, ref)
		}
		T.Out[item.As[UID]] = item
		item.Ref[Where] = T
	}
	return T
}

// Junk removes an item from the world and either schedules it to reset if it
// is unique or spawnable, otherwise it is freed for the garbage collector. If
// the item has inventory it is also junked.
func (t *Thing) Junk() {

	// The ordering of Out and In here is important as scheduling a Reset will
	// move items from In to Out and we don't want items processed twice.
	for _, item := range t.Out {
		item.Junk()
	}
	for _, item := range t.In {
		item.Junk()
	}

	delete(t.Ref[Where].In, t.As[UID])

	if t.Ref[Origin] == nil && t.Is&Spawnable != Spawnable {
		t.Free()
		return
	}

	delete(t.As, DynamicQualifier)

	if t.Ref[Origin] != nil {
		t.Ref[Where] = t.Ref[Origin]
	}
	t.Ref[Where].Out[t.As[UID]] = t
	t.Schedule(Reset)
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
	if t == nil || t.Is&Freed == Freed {
		return
	}
	t.Is = Freed

	for eventId := range t.Event {
		t.Suspend(eventId)
		delete(t.Event, eventId)
	}
	t.Event = nil

	for k := range t.As {
		delete(t.As, k)
	}
	t.As = nil

	for k := range t.Any {
		delete(t.Any, k)
	}
	t.Any = nil

	for k := range t.Int {
		delete(t.Int, k)
	}
	t.Int = nil

	for k := range t.Ref {
		delete(t.Ref, k)
	}
	t.Ref = nil

	for k, item := range t.In {
		item.Free()
		delete(t.In, k)
	}
	t.In = nil

	for k, item := range t.Out {
		item.Free()
		delete(t.Out, k)
	}
	t.Out = nil

	for k := range t.Who {
		delete(t.Who, k)
	}
	t.Who = nil
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

	lIn, lOut, lAs, lAny, lWho, lInt, lRef, lEvent :=
		len(t.In), len(t.Out), len(t.As), len(t.Any), len(t.Who), len(t.Int), len(t.Ref), len(t.Event)

	p("%s%p %[2]T - %s (%s)", tree[last].i, t, t.As[UID], t.As[Name])
	indent += tree[last].b
	p("%sIs - %032b (%s)", tree[false].i, t.Is, t.Is.setNames())
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
	p("%sEvents - len: %d", tree[false].i, lEvent)
	for k, event := range t.Event {
		lEvent--
		dueAt, dueIn := "-", "-"
		if at := t.Int[intKey(k)+DueAtOffset]; at > 0 {
			unix := time.Unix(0, int64(at))
			dueAt = unix.Format(time.Stamp)
			dueIn = unix.Sub(time.Now()).Truncate(time.Millisecond).String()
		}
		if in := t.Int[intKey(k)+DueInOffset]; in > 0 {
			dueIn = time.Duration(in).Truncate(time.Millisecond).String()
		}
		p("%s%s%s: %p, at: %s, in: %s",
			tree[false].b, tree[lEvent == 0].i, eventNames[k], event, dueAt, dueIn)
	}
	p("%sInt - len: %d", tree[false].i, lInt)
	for k, v := range t.Int {
		lInt--
		p("%s%s%s: %d", tree[false].b, tree[lInt == 0].i, intNames[k], v)
	}
	p("%sRef - len: %d", tree[false].i, lRef)
	for k, v := range t.Ref {
		lRef--
		if v != nil {
			p("%s%s[%2d] %s: %s (%s)", tree[false].b, tree[lInt == 0].i, k, refNames[k], v.As[UID], v.As[Name])
		} else {
			p("%s%s[%2d] %s: nil", tree[false].b, tree[lInt == 0].i, k, refNames[k])
		}
	}
	p("%sWho - len: %d", tree[false].i, lWho)
	w.Write([]byte(b.String()))
	b.Reset()
	for _, who := range t.Who {
		lWho--
		who.dump(w, width, indent+tree[false].b, lWho == 0)
	}
	p("%sIn - len: %d", tree[false].i, lIn)
	w.Write([]byte(b.String()))
	b.Reset()
	for _, item := range t.In {
		lIn--
		item.dump(w, width, indent+tree[false].b, lIn == 0)
	}
	p("%sOut - len: %d", tree[true].i, lOut)
	w.Write([]byte(b.String()))
	b.Reset()
	for _, item := range t.Out {
		lOut--
		item.dump(w, width, indent+tree[true].b, lOut == 0)
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

// Schedule the specified event for a Thing. If the event is already in-flight
// it will be cancelled and the new event scheduled. A scheduled event may be
// suspended, and then resumed by rescheduling it. A scheduled event may be
// cancelled, in which case rescheduling will cause the timers to start over.
func (t *Thing) Schedule(event eventKey) {

	var (
		idx    = intKey(event)
		delay  = t.Int[idx+AfterOffset]
		jitter = t.Int[idx+JitterOffset]
		dueIn  = t.Int[idx+DueInOffset]
	)

	switch {
	case delay+jitter+dueIn == 0:
		return
	case dueIn != 0:
		delay, jitter = dueIn, 0
		t.Int[idx+DueInOffset] = 0
	case jitter != 0:
		delay += rand.Int63n(jitter)
	}

	wait := time.Duration(delay)
	t.Cancel(event)
	t.Int[idx+DueAtOffset] = time.Now().Add(wait).UnixNano()
	t.Event[event] = time.AfterFunc(
		wait, func() {
			t.Cancel(event)
			NewState(t).Parse(eventCommands[event])
		},
	)
}

// Cancel an event for a Thing. The remaining time for the event is not
// recorded. If the event is rescheduled the timers will start over. A
// suspended event may be subsequently cancelled.
func (t *Thing) Cancel(event eventKey) {
	t.Suspend(event)
	t.Int[intKey(event)+DueInOffset] = 0
}

// Suspend an event for a Thing. If the event is not in-flight no action is
// taken. Suspending an in-flight event will record the time remaining before
// it fires so that the timers can be resumed when the event is rescheduled. A
// suspended event may be subsequently cancelled.
func (t *Thing) Suspend(event eventKey) {
	if t.Event[event] == nil {
		return
	}

	var suspended bool // True if we stop timer before it fires

	if suspended = t.Event[event].Stop(); !suspended {
		select {
		case <-t.Event[event].C:
		default:
		}
	}

	t.Event[event] = nil

	idx := intKey(event)
	dueAt, dueIn := idx+DueAtOffset, idx+DueInOffset

	t.Int[dueIn] = t.Int[dueAt] - time.Now().UnixNano()
	if !suspended || t.Int[dueIn] < 0 {
		t.Int[dueIn] = 0
	}
	t.Int[dueAt] = 0
}
