// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package driver

import (
	"strconv"
)

// menu to display to players.
const menuText = `
[WHITE]Please select one of the following options:

    [GREEN]1[WHITE] - Enter the game
    [GREEN]2[WHITE] - Read latest news
    [GREEN]3[WHITE] - Quit
`

// menu is a driver for the options menu
type menu struct {
	*driver
}

// newMenu creates a new menu driver from the current driver.
func (d *driver) newMenu() func() {
	m := menu{d}
	m.needOption()
	return m.checkOption
}

// needOption displays the menu and asks for the player's choice. The next
// function is then set to checkOption.
func (m *menu) needOption() {
	m.Respond(menuText)
	m.next = m.checkOption
}

// checkOption processes the current input as an option selected from the menu
// displayed by needOption.
func (m *menu) checkOption() {

	option, _ := strconv.Atoi(m.input)

	switch option {

	case 1:
		m.next = m.newGame()

	case 2:
		m.Respond("[YELLOW]Latest news is - we have no news...")
		m.needOption()

	case 3:
		m.next = nil

	case 42:
		m.Respond("[YELLOW]You have found a backdoor! Unfortunatly someone has locked it ;)")
		m.needOption()

	default:
		m.Respond("[RED]Invalid option selected.")
		m.needOption()

	}
}
