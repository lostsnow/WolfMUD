// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package driver

// game is simple driver for when the player is actually in the game. It simply
// passes input into the game to be handled.
type game struct {
	*driver
}

// newGame creates a new game driver from the current driver. It also places
// the player into the game.
func (d *driver) newGame() func() {
	g := game{d}
	g.player.Start(g.sender)
	return g.forward
}

// forward simply forwards any input to the in-game player to handle. If as a
// result the player is quitting the game we extract the player and change the
// current driver to be the main menu driver.
func (g *game) forward() {
	g.player.Parse(g.input)

	if g.player.IsQuitting() {
		g.next = g.newMenu()
	}
}
