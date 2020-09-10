// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

import (
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Vetoes represent one or more Veto allowing commands to be vetoed for a
// Thing. Multiple Veto can be added to veto multiple commands for different
// reasons.
//
// Its default implementation is the attr.Vetoes type.
type Vetoes interface {
	Attribute

	// Check compares the passed commands, issued by the given actor, with any
	// registered Veto commands. If a command for a Thing is vetoed the first
	// matching Veto found is returned. If no matching Veto are found nil is
	// returned.
	Check(actor Thing, cmd ...string) Veto
}

// Veto provides a way for a specific command to be vetoed for a specific
// Thing. Each Veto can veto a single command and provide a message detailing
// why the command was vetoed. Veto should be added to a Vetoes attribute for a
// Thing.
//
// Its default implementation is the attr.Veto type.
type Veto interface {

	// Dump adds information to the passed Node for debugging. The returned Node
	// indicates where addition information may be added.
	Dump(*tree.Node) *tree.Node

	// Command returns the command as an uppercased string that this Veto is for.
	Command() string

	// Message returns the details of why the associated command was vetoed.
	Message() string
}
