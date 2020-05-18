// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Body represents the body (and which slots are available) of a player or
// mobile so that Things can be used - held, worn, wielded, etc.
//
// Its default implementation is the attr.Body type.
type Body interface {
	Attribute

	// Wield returns true if the Wieldable is successfully wielded else false. If
	// successfully wielded Body slots will be allocated to the Wieldable and the
	// slots marked as 'wielding', on failure no slots are allocated.
	Wield(Wieldable) bool

	// Wear returns true if the Wearable is successfully worn else false. If
	// successfully worn Body slots will be allocated to the Wearable and the
	// slots marked as 'wearing', on failure no slots are allocated.
	Wear(Wearable) bool

	// Hold returns true if the Holdable is successfully held else false. If
	// successfully held Body slots will be allocated to the Holdable and the
	// slots marked as 'holding', on failure no slots are allocated.
	Hold(Holdable) bool

	// Remove the passed Thing from all Body slots allocated to it. Cleared Body
	// slots will also have the usage flags cleared.
	Remove(Thing)

	// Using returns true if passed Thing is allocated to any Body slot, else
	// false.
	Using(Thing) bool

	// Usage returns a string describing how the passed Thing is being used. For
	// example 'wielding' if the Thing is being wielded. If the passed Thing is
	// not currently being used an empty string is returned.
	Usage(Thing) string

	// UsedBy returns a slice of unique Things, even if allocated to more than
	// one slot, that are using the given slot references.
	UsedBy(refs []string) []Thing

	// Has returns true if the Body has all of the slots specified by the passed
	// refs, else false. Missing slots should never match.
	Has(refs []string) bool
}
