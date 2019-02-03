// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

/*

This file contains types and methods useful for writing tests for the cmd
package. The filename ends in _test so that it is only built for testing and
not normal builds. The package used is "cmd" rather then "cmd_test" so that
tests can import "cmd" and use these facilities.

*/

package cmd

import (
	"bytes"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// testPlayer represents a player for testing with a bytes.Buffer to simulate
// the network I/O stream. If Messages, MessagesFull or Reset is not called
// then data will accumulate in the bytes.Buffer.
type testPlayer struct {
	has.Thing
	*bytes.Buffer
}

// NewTestPlayer creates a new testPlayer and adds them into the game world at
// a random Start location. The testPlayer will be added to the game world
// silently, without using $POOF. The player's prompt will be set to StyleNone
// and any passed has.Thing will be added to the player's initial inventory.
// Multiple testPlayer may be created for testing the interactions between
// players and messages received by actors, participants and observers.
func NewTestPlayer(name string, inv ...has.Thing) *testPlayer {
	buf := &bytes.Buffer{}
	p := &testPlayer{
		attr.NewThing(
			attr.NewName(name),
			attr.NewDescription("This is a test player."),
			attr.NewInventory(inv...),
			attr.NewPlayer(buf),
		),
		buf,
	}
	attr.FindPlayer(p).SetPromptStyle(has.StyleNone)

	start := (*attr.Start)(nil).Pick()

	// Make sure we lock in LockID order to avoid deadlocks
	i1 := start
	i2 := attr.FindInventory(p)
	if i1.LockID() > i2.LockID() {
		i1, i2 = i2, i1
	}
	i1.Lock()
	defer i1.Unlock()
	i2.Lock()
	defer i2.Unlock()

	start.Add(p)
	start.Enable(p)

	return p
}

// Messages returns any unread messages from the testPlayer and resets the
// output buffer. The messages will have any ANSI escape sequences stripped via
// text.Monoize making it easier to write tests that compare message text.
func (p *testPlayer) Messages() string {
	where := attr.FindLocate(p).Where()
	where.Lock()
	defer where.Unlock()
	out := p.String()
	p.Reset()
	return out
}
