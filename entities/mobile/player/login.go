// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package player

import (
	"code.wolfmud.org/WolfMUD.git/utils/config"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"

	"crypto/sha512"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"
	"strings"
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

	// TODO: Is this check still needed?
	if l.input == "QUIT" {
		l.quitting = true
		return
	}

	p, err := l.loadPlayer()

	switch err {

	case BadCredentials:
		l.Respond("[RED]Name or password is incorrect. Please try again.")

	case BadPlayerFile:
		l.Respond("[RED]An embarrassed sounding little voice squeaks 'Sorry... there seems to be a problem restoring you. Please contact the MUD Admin staff.")

	}

	if err != nil {
		l.needName()
		return
	}

	if err := PlayerList.Add(p); err != nil {
		log.Printf("Login error: %s", err)
		l.Respond("[RED]That player is already logged in!")
		l.needName()
		return
	}

	l.player = p
	l.player.sender = l.sender
	l.Respond("[GREEN]A loud voice booms 'You have been brought back " + l.name + "'.")
	l.next = l.newMenu()
	log.Printf("Successful login: %s", l.name)
}

// Errors that can be returned by loadPlayer.
var (
	BadCredentials = errors.New("Invalid credentals")
	BadPlayerFile  = errors.New("Invalid player file")
	DuplicateLogin = errors.New("Player already logged in")
)

// loadPlayer loads a player file.
//
// First the SHA512 hash of the player's name is calculated and Base64 encoded.
// This gives us a safe 88 character string to be used as the filename storing
// the player's character details. If we just accepted the player's input as
// the filename they could try something like '../config' or something equally
// nasty.
//
// If we cannot open the player's file we return an error of BadCredentials.
//
// Then we take the salt value from the player's file and append the password
// taken from the current input. The SHA512 of the resulting string is
// calculated and Base64 encoded before being compared with the stored Base64
// endcoded password hash.
//
// If salt+input = password hash in the player's file we have a valid login.
// Otherwise we return an error of BadCredentials.
//
// If the credentials are good but the player's file cannot be loaded we return
// an error of BadPlayerFile.
//
// Note that we are manually opening the player's file, reading it as a
// recordjar, peeking inside it, then unmarshaling it. This is so that we can
// abort at any point - player not found, incorrect password, corrupt player
// file - having done as little work as possible. In this way we are not
// unmarshaling players which may have a lot of dependant stuff (inventory) to
// unmarshal just to validate the login - someone could hit the server and tie
// up processing with invalid logins otherwise if the unmarshaling took a
// significant amount of time.
//
// NOTE: The 88 character filename + 4 character extension (.wrj) will break
// some file systems such as HFS on Mac OS (Not OS X), Joliet for CD-ROMs and
// maybe others.
//
// BONUS TRIVIA: The Java version of WolfMUD would not compile on Mac OS at one
// point due to a file with a name over 32 characters long... *sigh*
func (l *login) loadPlayer() (*Player, error) {

	// Create a hash which can be reused by calling Reset().
	h := sha512.New()

	// Convert name into a 88 character Base64 encoded string.
	io.WriteString(h, l.name)
	filename := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// Can we open the player's file to get the current salt and password hash?
	f, err := os.Open(config.DataDir + "players/" + filename + ".wrj")
	if err != nil {
		return nil, BadCredentials
	}
	defer f.Close()

	rj, _ := recordjar.Read(f)

	r := rj[0]
	s := r.String("salt")
	p := r.String("password")

	// Password hash can be split over multiple lines in the file so try
	// stitching it back together.
	p = strings.Replace(p, " ", "", -1)

	// Reset hash and calculate for salt + password from current input.
	h.Reset()
	io.WriteString(h, s+l.input)

	// Million dollar question: does the salt+input password hash match the
	// salt+file password hash? If so unmarshal the player's file.
	if base64.URLEncoding.EncodeToString(h.Sum(nil)) != p {
		return nil, BadCredentials
	}

	l.name = r.String("name")
	data := recordjar.UnmarshalJar(&rj)

	if data["PLAYER"] == nil {
		log.Printf("Error loading player: %#v", rj)
		return nil, BadPlayerFile
	}

	return data["PLAYER"].(*Player), nil
}
