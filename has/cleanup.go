// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Cleanup provides information on how often a Thing should be cleaned up when
// left laying around.
//
// The default implementation is the attr.Cleanup type.
type Cleanup interface {
	Attribute

	// Cleanup causes the parent Thing to be scheduled for clean up.
	Cleanup()

	// Abort cancels any outstanding clean up events.
	Abort()

	// Active returns true if any of the Inventories the parent Thing is in
	// already have a clean up scheduled, otherwise false.
	Active() bool
}
