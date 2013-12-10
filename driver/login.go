// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package driver

import (
	"code.wolfmud.org/WolfMUD.git/entities/mobile/player"

	"log"
)

// greeting to display to player when they initially connect.
//
// TODO: This needs moving into an easily editable text file. However RecordJar
// has a 'minor feature' which will not let you encode newlines that are
// preserved *sigh*
const greetingText = `

[GREEN]Wolf[WHITE]MUD Copyright 2013 Andrew 'Diddymus' Rolfe

    [GREEN]W[WHITE]orld
    [GREEN]O[WHITE]f
    [GREEN]L[WHITE]iving
    [GREEN]F[WHITE]antasy


`

// login is a driver for logging in players to the server
type login struct {
	*driver
}

// newLogin creates a new login driver from the current driver.
func (d *driver) newLogin() func() {
	l := login{d}
	return l.welcome
}

// welcome displayes the greeting text and asks for the player's name beginning
// the login sequence.
func (l *login) welcome() {
	l.Respond(greetingText)
	l.needName()
}

// needName asks for the player's name and sets the next function to check the
// name entered.
func (l *login) needName() {
	l.Respond("Please enter your character's name.")
	l.next = l.checkName
}

// checkName processes the current input as the player's name. If it is empty
// we ask again for the name otherwise we move on to asking for the password.
func (l *login) checkName() {
	l.name = l.input
	if l.name == "" {
		l.needName()
	} else {
		l.needPassword()
	}
}

// needPassword asks for the player's password and sets the next function to
// check the password.
func (l *login) needPassword() {
	l.Respond("Enter character's password or just press [CYAN]ENTER[WHITE] to abort.")
	l.next = l.checkPassword
}

// checkPassword processes the current input as the player's password. If no
// password is entered we go back to asking for the player's name. Otherwise we
// try loading the player's data file and respond accordingly. If the player's
// data file is loaded and everything is OK we switch to the menu driver.
func (l *login) checkPassword() {

	// If no password entered go back to asking for a name.
	if l.input == "" {
		l.needName()
		return
	}

	// Ignores processing if quit sent from client on timeout
	if l.input == "QUIT" {
		return
	}

	p, err := player.Load(l.name, l.input)

	switch err {

	case player.BadCredentials:
		l.Respond("[RED]Name or password is incorrect. Please try again.")

	case player.BadPlayerFile:
		l.Respond("[RED]An embarrassed sounding little voice squeaks 'Sorry... there seems to be a problem restoring you. Please contact the MUD Admin staff.")

	}

	if err != nil {
		l.needName()
		return
	}

	if err := player.PlayerList.Add(p); err != nil {
		log.Printf("Login error: %s", err)
		l.Respond("[RED]That player is already logged in!")
		l.needName()
		return
	}

	l.player = p
	l.Respond("[GREEN]A loud voice booms 'You have been brought back " + l.name + "'.")
	l.next = l.newMenu()
	log.Printf("Successful login: %s", l.name)
}
