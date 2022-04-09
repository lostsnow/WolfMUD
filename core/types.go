// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"strings"
)

// Type definitions for Thing field keys.
type (
	isKey    uint32 // index for Thing.Is
	asKey    uint32 // index for Thing.As
	anyKey   uint32 // index for Thing.Any
	intKey   uint32 // index for Thing.Int
	refKey   uint32 // index for Thing.Ref
	eventKey uint32 // index for Thing.Events
)

// Constants for use as bitmasks with the Thing.Is field.
const (
	Container isKey = 1 << iota // A container, allows PUT/TAKE
	Dark                        // A dark location
	Freed                       // Thing has been freed for GC
	HasBody                     // Item has a body (Any[Body] can be empty)
	Holding                     // Item is being held
	Location                    // Item is a location
	NPC                         // An NPC
	Narrative                   // A narrative item
	Open                        // An open item (e.g. door)
	Player                      // Is a player
	Spawnable                   // Is item spawnable?
	Start                       // A starting location
	Wait                        // Container reset wait for inventory?
	Wielding                    // Item is being wielded
	Wearing                     // Item is being worn
	_Open                       // Initial open state of item (e.g. door)
)

// Useful masks for groups of constants for checking multiple flags.
const (
	Using isKey = Holding | Wearing | Wielding
)

// isNames maps isKey bits to their string name. See also setName method.
var isNames = []string{
	"Container",
	"Dark",
	"Freed",
	"HasBody",
	"Holding",
	"Location",
	"NPC",
	"Narrative",
	"Open",
	"Player",
	"Spawnable",
	"Start",
	"Wait",
	"Wielding",
	"Wearing",
	"_Open",
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

// Constants for use as keys in a Thing.As field.
const (
	BadAsKey asKey = iota

	// Exit directions should always be consecutive constants in given order
	// These direction keys are only used by the loader.
	_North
	_Northeast
	_East
	_Southeast
	_South
	_Southwest
	_West
	_Northwest
	_Up
	_Down

	Account          // MD5 hash of player's account
	Barrier          // A barrier, value is direction of exit blocked ("E")
	Blocker          // Name of direction being blocked ("E")
	Description      // Item's description
	DynamicAlias     // "PLAYER" or unset, "SELF" for actor performing a command
	DynamicQualifier // Situation dependant e.g. GET sets "MY",DROP deleted "MY"
	Gender           // Gender of a player or NPC
	Name             // Item's name
	OnCleanup        // Custome cleanup message for an item
	OnReset          // Custom reset message for an item
	Password         // Salted SHA512 hash of the account password
	PromptStyle      // Current prompt style
	Ref              // Item's original reference (zone:ref or ref)
	Salt             // Salt used for the account password
	TheName          // Item's name with a/an/some prefix changed to 'the'
	TriggerType      // Type of trigger event to send
	UID              // Item's unique identifier
	UName            // Name with the initial character upper-cased
	UTheName         // TheName with the initial character upper-cased
	VetoClose        // Veto CLOSE command
	VetoCombat       // Veto fighting commands
	VetoDrop         // Veto for DROP command
	VetoGet          // Veto for GET command
	VetoHold         // Veto HOLD command
	VetoJunk         // Veto for JUNK command
	VetoOpen         // Veto OPEN command
	VetoPut          // Veto PUT command for item
	VetoPutIn        // Veto for PUT command into container
	VetoRead         // Veto READ command
	VetoRemove       // Veto REMOVE command
	VetoTake         // Veto TAKE command for item
	VetoTakeOut      // Veto for TAKE command from container
	VetoWear         // Veto WEAR command
	VetoWield        // Veto WIELD command
	Writing          // Description of writing on an item
	Zone             // Zone item's definition loaded from
)

// asNames maps asKey values to their string name.
var asNames = []string{
	"BadAsKey",

	"_North", "_Northeast", "_East", "_Southeast",
	"_South", "_Southwest", "_West", "_Northwest",
	"_Up", "_Down",

	"Account",
	"Barrier",
	"Blocker",
	"Description",
	"DynamicAlias",
	"DynamicQualifier",
	"Gender",
	"Name",
	"OnCleanup",
	"OnReset",
	"Password",
	"PromptStyle",
	"Ref",
	"Salt",
	"TheName",
	"TriggerType",
	"UID",
	"UName",
	"UTheName",
	"VetoClose",
	"VetoCombat",
	"VetoDrop",
	"VetoGet",
	"VetoHold",
	"VetoJunk",
	"VetoOpen",
	"VetoPut",
	"VetoPutIn",
	"VetoRead",
	"VetoRemove",
	"VetoTake",
	"VetoTakeOut",
	"VetoWear",
	"VetoWield",
	"Writing",
	"Zone",
}

var (
	// NameToDir maps a long or short direction name to its Thing.As constant.
	NameToDir = map[string]refKey{
		"N": North, "NE": Northeast, "E": East, "SE": Southeast,
		"S": South, "SW": Southwest, "W": West, "NW": Northwest,
		"U": Up, "D": Down,
		"NORTH": North, "NORTHEAST": Northeast, "EAST": East, "SOUTHEAST": Southeast,
		"SOUTH": South, "SOUTHWEST": Southwest, "WEST": West, "NORTHWEST": Northwest,
		"UP": Up, "DOWN": Down,
	}

	// DirToName maps a Thing.As direction constant to the direction's long name.
	DirToName = map[refKey]string{
		North: "north", Northeast: "northeast", East: "east", Southeast: "southeast",
		South: "south", Southwest: "southwest", West: "west", Northwest: "northwest",
		Up: "up", Down: "down",
	}

	// DirRefToAs maps a Thing.Ref direction to a Thing.As direction
	DirRefToAs = map[refKey]asKey{
		North: _North, Northeast: _Northeast, East: _East, Southeast: _Southeast,
		South: _South, Southwest: _Southwest, West: _West, Northwest: _Northwest,
		Up: _Up, Down: _Down,
	}

	// ReverseDir maps a Thing.Ref direction to its opposite direction
	ReverseDir = map[refKey]refKey{
		North: South, Northeast: Southwest, East: West, Southeast: Northwest,
		South: North, Southwest: Northeast, West: East, Northwest: Southeast,
		Up: Down, Down: Up,
	}
)

// Constants for Thing.Any keys
const (
	BadAnyKey anyKey = iota

	Alias        // Aliases for an item
	BarrierAllow // Aliases allowed to pass barrier
	BarrierDeny  // Aliases denied to pass barrier
	Body         // Body slots available to an item
	Holdable     // Body slots required to hold item
	OnAction     // Actions that can be performed
	Permissions  // Permissions a player has
	Qualifier    // Alias qualifiers
	Wearable     // Body slots required to wear item
	Wieldable    // Body slots required to wield item
	_Holding     // UIDs of items initially held
	_Wearing     // UIDs of items initially worn
	_Wielding    // UIDs of items initially wielded

)

// anyNames maps anyKey values to their string name.
var anyNames = []string{
	"BadAnyKey",

	"Alias",
	"BarrierAllow",
	"BarrierDeny",
	"Body",
	"Holdable",
	"OnAction",
	"Permissions",
	"Qualifier",
	"Wearable",
	"Wieldable",
	"_Holding",
	"_Wearing",
	"_Wielding",
}

// Constants for Thing.Int keys
//
// NOTE: See also comments for eventKey constants.
const (
	BadIntKey intKey = iota

	// Events
	ActionAfter   // How often an action event should occur
	ActionJitter  // Maximum random delay to add to ActionAfter
	ActionDueAt   // Time a scheduled Action is due
	ActionDueIn   // Time remaining for Action
	CleanupAfter  // How soon a clean-up event should occur
	CleanupJitter // Maximum random delay to add to CleanupAfter
	CleanupDueAt  // Time a scheduled clean-up is due
	CleanupDueIn  // Time remaining for clean-up
	HealthAfter   // How soon a healing event should occur
	HealthJitter  // Maximum random delay to add to HealthAfter
	HealthDueAt   // Time a scheduled healing event is due
	HealthDueIn   // Time remaining for healing event
	ResetAfter    // How soon a reset event should occur
	ResetJitter   // Maximum random delay to add to TesetAfter
	ResetDueAt    // Time a scheduled reset is due
	ResetDueIn    // Time remaining for reset
	TriggerAfter  // How soon a trigger should occur
	TriggerJitter // Maximum random delay to add to trigger
	TriggerDueAt  // Time a scheduled trigger event is due
	TriggerDueIn  // Time remaining for trigger event

	// Non-events
	Created       // Timestamp of when item (player) created
	HealthCurrent // Current health of a player/mobile
	HealthMaximum // Maximum health a player/mobile heals up to.
	HealthRestore // Health restored per healing event
)

// intNames maps intKey values to their string name.
var intNames = []string{
	"BadIntKey",

	// Events
	"ActionAfter",
	"ActionJitter",
	"ActionDueAt",
	"ActionDueIn",
	"CleanupAfter",
	"CleanupJitter",
	"CleanupDueAt",
	"CleanupDueIn",
	"HealthAfter",
	"HealthJitter",
	"HealthDueAt",
	"HealthDueIn",
	"ResetAfter",
	"ResetJitter",
	"ResetDueAt",
	"ResetDueIn",
	"TriggerAfter",
	"TriggerJitter",
	"TriggerDueAt",
	"TriggerDueIn",

	// Non-events
	"Created",
	"HealthCurrent",
	"HealthMaximum",
	"HealthRestore",
}

// Standard offsets for Event related values. Given an eventKey we can add the
// offsets to get the After, Jitter, DueAt and DueIn values from Thing.Int for
// an event.
const (
	AfterOffset intKey = iota
	JitterOffset
	DueAtOffset
	DueInOffset
)

// Constants for Thing.Events keys
//
// NOTE: Events map to Thing.Int values. The intKey constants for an event's
// After and Jitter values should be consecutive as we assume After = eventKey
// and Jitter = eventKey+1.
const (
	Action  eventKey = eventKey(ActionAfter)
	Cleanup          = eventKey(CleanupAfter)
	Health           = eventKey(HealthAfter)
	Reset            = eventKey(ResetAfter)
	Trigger          = eventKey(TriggerAfter)
)

// eventNames maps eventKey values to their string name.
var eventNames = map[eventKey]string{
	Action:  "Action",
	Cleanup: "Cleanup",
	Health:  "Health",
	Reset:   "Reset",
	Trigger: "Trigger",
}

// Constants for Thing.Ref keys
const (
	BadRefKey refKey = iota

	// Exit directions should always be consecutive constants in given order
	North
	Northeast
	East
	Southeast
	South
	Southwest
	West
	Northwest
	Up
	Down

	Where  // Where an item is
	Origin // Where a unique item resets to
)

// refNames maps refKey values to their string name.
var refNames = []string{
	"BadRefKey",

	"North", "Northeast", "East", "Southeast",
	"South", "Southwest", "West", "Northwest",
	"Up", "Down",

	"Where",
	"Origin",
}

// preferredOrdering defines the preferred sorting or for attributes when
// marshaled. As much as possible the names used are predefined type names.
var preferredOrdering = []string{
	asNames[Ref],
	asNames[Zone],
	"Author",
	"Disabled",
	asNames[Name],
	anyNames[Alias], "Aliases",
	Start.setNames(),
	"Exit", "Exits",
	"ZoneLinks",
	"Barrier",
	"Door",
	"Location",
	asNames[Description],
	anyNames[Body],
	asNames[Gender],
	eventNames[Health],
	"Inv", "Inventory",
	Holding.setNames(),
	Wearing.setNames(),
	Wielding.setNames(),
	Narrative.setNames(),
	anyNames[Holdable],
	anyNames[Wearable],
	anyNames[Wieldable],
	asNames[Writing],
	"Veto", "Vetoes",
	eventNames[Action],
	"On" + eventNames[Action],
	eventNames[Cleanup],
	"On" + eventNames[Cleanup],
	eventNames[Reset],
	"On" + eventNames[Reset],
}
