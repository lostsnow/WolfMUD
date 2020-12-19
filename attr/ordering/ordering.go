// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package ordering provides a list for ordering Attributes when they are
// marshaled.
//
// There is no good reason for defining the attribute names as constants, other
// than the fact the groupings can be documented.
//
// BUG(diddymus): This should really be under attr/ordering.go, however this
// causes the configuration loading to kick in when the init methods are run.
// This is not ideal for simple tools like wrjfmt :(
package ordering

// Attributes defines the prefered sorting or for attributes when marshaled.
var Attributes = []string{
	Ref,
	Zone,
	Author,
	Disabled,
	Name,
	Alias, Aliases,
	Start,
	Exit, Exits,
	ZoneLinks,
	Barrier,
	Door,
	Location,
	Description,
	Body,
	Gender,
	Health,
	Inv, Inventory,
	Holding,
	Wearing,
	Wielding,
	Narrative,
	Holdable,
	Wearable,
	Wieldable,
	Writing,
	Veto, Vetoes,
	Action,
	OnAction,
	Cleanup,
	OnCleanup,
	Reset,
	OnReset,
}

// Identification related attributes:
const (
	Ref     = "Ref"
	Name    = "Name"
	Alias   = "Alias"
	Aliases = "Aliases"
)

// Zone related attributes (only used in zone header records):
const (
	Author   = "Author"
	Disabled = "Disabled"
	Zone     = "Zone"
)

// Location related attributes:
const (
	Barrier   = "Barrier"
	Door      = "Door"
	Exit      = "Exit"
	Exits     = "Exits"
	Location  = "Location"
	Start     = "Start"
	ZoneLinks = "ZoneLinks"
)

// Description attribute when used as a named field and not an unnamed free text section:
const (
	Description = "Description"
)

// Body related attributes:
const (
	Body      = "Body"
	Gender    = "Gender"
	Health    = "Health"
	Holding   = "Holding"
	Inv       = "Inv"
	Inventory = "Inventory"
	Wearing   = "Wearing"
	Wielding  = "Wielding"
)

// Attributes affecting how something is used:
const (
	Holdable  = "Holdable"
	Narrative = "Narrative"
	Veto      = "Veto"
	Vetoes    = "Vetoes"
	Wearable  = "Wearable"
	Wieldable = "Wieldable"
	Writing   = "Writing"
)

// Event related attributes
const (
	Action    = "Action"
	Cleanup   = "Cleanup"
	OnAction  = "OnAction"
	OnCleanup = "OnCleanup"
	OnReset   = "OnReset"
	Reset     = "Reset"
)
