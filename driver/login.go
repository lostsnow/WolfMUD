// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package driver

import (
	"code.wolfmud.org/WolfMUD.git/entities/mobile/player"
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

// welcome displayes the greeting text and asks for the player's account
// beginning the login sequence.
func (l *login) welcome() {
	l.Respond(greetingText)
	l.needAccount()
}

// needAccount asks for the player's account and sets the next function to
// check the account entered.
func (l *login) needAccount() {
	l.Respond("Please enter your account ID or just press [CYAN]ENTER[WHITE] to create a new account:")
	l.next = l.checkAccount
}

// checkAccount processes the current input as the player's account. If it is
// empty we ask again for the account otherwise we move on to asking for the
// password.
func (l *login) checkAccount() {

	if l.input == "" {
		l.next = l.newAccount()
		return
	}

	l.account = player.HashAccount(l.input)
	l.needPassword()
}

// needPassword asks for the player's password and sets the next function to
// check the password.
func (l *login) needPassword() {
	l.Respond("Enter character's password or just press [CYAN]ENTER[WHITE] to abort:")
	l.next = l.checkPassword
}

// checkPassword processes the current input as the player's password. If no
// password is entered we go back to asking for the player's account. Otherwise
// we try loading the player's data file and respond accordingly. If the
// player's data file is loaded and everything is OK we switch to the menu
// driver.
func (l *login) checkPassword() {

	// If no password entered go back to asking for an account.
	if l.input == "" {
		l.needAccount()
		return
	}

	// Ignores processing if quit sent from client on timeout
	if l.input == "QUIT" {
		return
	}

	var err error
	l.player, err = player.Load(l.account, l.input)

	switch err {

	case player.BadCredentials:
		l.Respond("[RED]Account or password is incorrect. Please try again.")

	case player.BadPlayerFile:
		l.Respond("[RED]An embarrassed sounding little voice squeaks 'Sorry... there seems to be a problem restoring you. Please contact the MUD Admin staff.")

	}

	if err != nil {
		l.needAccount()
		return
	}

	if err := l.login(); err != nil {
		l.Respond("[RED]That account is already logged in!")
		l.player = nil
		l.needAccount()
		return
	}

	l.Respond("[GREEN]A loud voice booms 'You have been brought back " + l.player.Name() + "'.")
	l.next = l.newMenu()
}
