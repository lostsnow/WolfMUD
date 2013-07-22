// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package player

import (
	"code.wolfmud.org/WolfMUD.git/utils/config"
	"code.wolfmud.org/WolfMUD.git/utils/parser"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"code.wolfmud.org/WolfMUD.git/utils/sender"
	"crypto/sha512"
	"encoding/base64"
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
const greeting = `

[GREEN]Wolf[WHITE]MUD Copyright 2013 Andrew 'Diddymus' Rolfe

    [GREEN]W[WHITE]orld
    [GREEN]O[WHITE]f
    [GREEN]L[WHITE]iving
    [GREEN]F[WHITE]antasy


`

// state type lets us easily redefine the base type of states if needed and
// using 'state' instead of 'uint8' is more descriptive. States are used to
// record which state the parser is in such as asking for a name or password.
type state uint8

// The different states the parser can be in.
const (
	connect state = iota
	needName
	needPassword
	badCredentials
	goodCredentials
)

// login is the implementation of a parser for logging into the server :)
type login struct {
	state
	sender    sender.Interface
	name      string
	recordjar *recordjar.RecordJar
	quitting  bool
}

// Login creates and returns a new login parser.
func Login(s sender.Interface) (l *login) {
	l = &login{
		state:    connect,
		sender:   s,
		name:     "unknown",
		quitting: false,
	}

	// The connect state in the Parse method does not have any 'input'
	// processing so any input will be ignored.  However there is 'output'
	// processing which will cause the initial greeting to be displayed.
	l.Parse("")

	return l
}

// Parse takes input and processes it. Parse is made up of two switches with
// each one implementing a simple finite state machine - FSM. The first switch
// processes the current input based on the current state. For example if we
// are in the state needName we parse the input as a name. The processing may
// also transition us to a new state.
//
// The second switch is used mainly to respond to the user based on the new
// state after the first switch has finished processing. Note that the state
// may not necessarily be changed by the first switch.
func (l *login) Parse(input string) {

	if input == "QUIT" {
		l.quitting = true
		return
	}

	// (mostly) INPUT PROCESSING

	// Process the input we expect for the current state we are in, optionally
	// setting a new state as a result.
	switch l.state {

	case connect:
		// THIS CASE INTENTIONALLY LEFT BLANK...

	case needName:
		l.state = l.checkName(input)

	case needPassword:
		l.state = l.checkPassword(input)

	case badCredentials:
		// THIS CASE INTENTIONALLY LEFT BLANK...

	case goodCredentials:
		// THIS CASE INTENTIONALLY LEFT BLANK...

	}

	// (mostly) OUTPUT PROCESSING

	// Respond depending on the state we are now in - probably a right old one
	// eh? ;) - We loop here so that we can potentially display a message, change
	// state, then display another message. An example is greeting the user and
	// asking for their login name.
	for again, msg := true, ""; again; {
		again = false

		switch l.state {

		case connect:
			msg += greeting
			l.state = needName
			again = true

		case needName:
			l.sender.Prompt(sender.PROMPT_DEFAULT)
			msg += "Please enter your character's name or the name for a new character."
			l.name = "unknown"

		case needPassword:
			msg += "Enter character's password or just press [CYAN]ENTER[WHITE] to abort."

		case badCredentials:
			msg += "[RED]Name or password is incorrect. Please try again.[WHITE]\n"
			l.state = needName
			again = true

		case goodCredentials:
			l.sender.Prompt(sender.PROMPT_NONE)
			msg += "[GREEN]A loud voice booms 'You have been brought back " + l.name + "'.[WHITE]\n\n"
			l.quitting = true

		}

		if !again {
			l.sender.Send(msg)
		}
	}

}

// checkName stores the current input as the player's name and then sets the
// current state to ask for the password next. If no input is provided state is
// not changed causing the name to be asked for again.
func (l *login) checkName(input string) state {
	if input == "" {
		return needName
	}
	l.name = input
	return needPassword
}

// checkPassword takes the current input and uses it as the player's password.
//
// First the SHA512 hash of the player's name is calculated and Base64 encoded.
// This gives us a safe 88 character string to be used as the filename storing
// the player's character details. If we just accepted the player's input as
// the filename they could try something like '../config' or something equally
// nasty.
//
// If we cannot open the player's file we change state to badCredentials.
//
// Otherwise we take the salt value from the player's file and append the
// password taken from the current input. The SHA512 of the resulting string is
// calculated and Base64 encoded before being compared with the stored Base64
// endcoded password hash.
//
// If salt+input = password hash in the player's file we have a valid login and
// return a state of goodCredentials. Otherwise we return a state of
// badCredentials.
//
// TODO: Most of this processing should be moved to player.go
//
// NOTE: The 88 character filename + 4 character extension (.wrj) will break
// some file systems such as HFS on Mac OS (Not OS X), Joliet for CD-ROMs and
// maybe others.
//
// BONUS TRIVIA: The Java version of WolfMUD would not compile on Mac OS at one
// point due to a file with a name over 32 characters long... *sigh*
func (l *login) checkPassword(input string) state {

	// If no password entered go back to asking for a name.
	if input == "" {
		return needName
	}

	// Create a hash which can be reused by calling Reset().
	h := sha512.New()

	// Convert name into a 88 character Base64 encoded string.
	io.WriteString(h, l.name)
	filename := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// Can we open the player's file to get the current salt and password hash?
	f, err := os.Open(config.DataDir + "players/" + filename + ".wrj")
	if err != nil {
		return badCredentials
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
	io.WriteString(h, s+input)

	// Million dollar question: does the salt+input password hash match the
	// salt+file password hash?
	if base64.URLEncoding.EncodeToString(h.Sum(nil)) == p {
		l.name = r.String("name")
		l.recordjar = &rj
		log.Printf("Successful login: %s", l.name)
		return goodCredentials
	}

	return badCredentials
}

func (l *login) Name() string {
	return l.name
}

// Destroy for this parser does not need to do anything. It implements part of
// the parser.Interface.
func (l *login) Destroy() {}

// IsQuitting returns true if the parser is trying to quit otherwise false. It
// implements part of the parser.Interface.
func (l *login) IsQuitting() bool {
	return l.quitting
}

// Next returns either a new player parser or if we are wanting the client to
// quit we just return nil for the next parser to use.
func (l *login) Next() parser.Interface {
	if l.recordjar == nil {
		return nil
	}
	return New(l.sender, l.recordjar)
}
