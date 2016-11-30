// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/text"
)

// menu embeds a frontend instance adding fields and methods specific to
// the main menu.
type menu struct {
	*frontend
}

// NewMenu returns a menu with the specified frontend embedded. The returned
// menu can be used for processing the main menu and it's options.
func NewMenu(f *frontend) (m *menu) {
	m = &menu{frontend: f}
	m.menuDisplay()
	return
}

// menuDisplay shows the main menu of options available once a player is logged
// into the system.
func (m *menu) menuDisplay() {
	m.buf.Send(`
  Main Menu
  ---------

  1. Enter game
  0. Quit

Select an option:`)
	m.nextFunc = m.menuProcess
}

// menuProcess takes the curernt input and processes it as a menu option. If
// the option is valid the corresponding action is taken. If the option is not
// valid the player is notified and we wait for another option to be chosen.
func (m *menu) menuProcess() {
	switch string(m.input) {
	case "":
		return
	case "1":
		NewGame(m.frontend)
	case "0":
		m.Close()
	default:
		m.buf.Send(text.Bad, "Invalid option selected.", text.Default)
	}
}
