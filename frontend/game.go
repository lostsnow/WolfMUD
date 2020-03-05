// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/message"
	"code.wolfmud.org/WolfMUD.git/stats"
)

// game embeds a frontend instance adding fields and methods specific to
// communicating with the game.
type game struct {
	*frontend
}

// NewGame returns a game with the specified frontend embedded. The returned
// game can be used for processing communication to the actual game.
func NewGame(f *frontend) (g *game) {
	g = &game{frontend: f}
	g.init()
	return
}

// gameInit is used to place the player into the game world. As the game
// backend has it's own output handling we remove the frontend.buf buffer to
// prevent duplicate output. The buffer is restored by gameProcess when the
// player quits the game world.
func (g *game) init() {

	message.ReleaseBuffer(g.buf)
	g.buf = nil

	// Get a random starting location
	start := (*attr.Start)(nil).Pick().Outermost()

	// Lock starting location and player in LockID order to avoid deadlocks
	i1 := start
	i2 := attr.FindInventory(g.player)
	if i1.LockID() > i2.LockID() {
		i1, i2 = i2, i1
	}
	i1.Lock()
	i2.Lock()

	attr.FindPlayer(g.player).SetPromptStyle(has.StyleShort)
	start.Add(g.player)
	start.Enable(g.player)
	stats.Add(g.player)

	// Release locks before calling cmd.Script which will also try and lock the
	// starting location and would cause a deadlock. It's a shame we can't reuse
	// the lock we have already acquired somehow...
	i2.Unlock()
	i1.Unlock()

	cmd.Script(g.player, "$POOF")
	g.nextFunc = g.process
}

// gameProcess hands input to the game backend for processing while the player
// is in the game. When the player is no longer in the world the frontend.buf
// buffer is restored - see gameInit.
func (g *game) process() {
	l := attr.FindLocate(g.player)

	// Only pass command to game parser if still in the world
	if l.Where() != nil {
		cmd.Parse(g.player, string(g.input))
	}

	// If no longer in the world switch to frontend main menu
	if l.Where() == nil {
		g.buf = message.AcquireBuffer()
		g.buf.OmitLF(true)
		NewMenu(g.frontend)
	}
}
