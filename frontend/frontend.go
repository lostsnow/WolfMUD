// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package frontend provides interactive processing of front end pages and
// access to the backend game. Note that a 'page' can be anything from a menu
// of options to choose from to a simple request for a password.
//
// The frontend is responsible for coordinating the display of pages to a user
// and processing their responses. These pages cover logging into the server,
// account creation, player creation and other non in-game activities. When the
// player is in-game the frontend will simply pass any input through to the
// game backend for processing.
//
// Pages typically have a pair of methods - a display part and a processing
// part. For example accountDisplay and accountProcess. Sometimes there is only
// a display part, for example greetingDisplay.
//
// The current state is held in an instance of frontend. With frontend.nextFunc
// being the next method to call when input is received - usually an xxxProcess
// method.
//
// Each time input is received Parse will be called. The method in nextFunc
// will be called to handle the input. nextFunc should then call the next
// xxxDisplay method to send a response to the input processing and setup
// nextFunc with the method that will process the next input received. Any
// buffered response will then be sent back before Parse exits. Parse will then
// be called again when more input is received.
package frontend

import (
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/stats"

	"bytes"
	"io"
	"sync"
)

// accounts is used to track which (valid) accounts are logged in and in use.
// It's main purpose is to track logged in account to prevent duplicate logins.
//
// NOTE: A frontend account hash is NOT valid until it has been registered as
// being inuse.
var accounts struct {
	inuse map[string]struct{}
	sync.Mutex
}

// init is used to initialise the map used in account tracking.
func init() {
	accounts.inuse = make(map[string]struct{})
}

// closedError represents the fact that Close has been called on a frontend
// instance releasing it's resources and that the instance should be discarded.
// As interaction with the error is via the standard error and comms.temporary
// interfaces it does not need to be exported.
type closedError struct{}

// Error implements the error interface for errors and returns descriptive text
// for the closedError error.
func (_ closedError) Error() string {
	return "frontend closed"
}

// Temporary always returns true for any closedError. A closedError is
// considered temporary as recovery is easy - create a new frontend instance.
func (_ closedError) Temporary() bool {
	return true
}

// buffer is our extended version of a bytes.Buffer so that we can add some
// convience methods.
type buffer struct {
	*bytes.Buffer
}

// WriteJoin takes a number of strings and writes them into the buffer. It's a
// convenience method to save writing multiple WriteString statements and an
// alternative to additional allocations due to concatenation.
//
// The return value n is the total length of all s, in bytes; err is always nil.
// The underlying bytes.Buffer may panic if it becomes too large.
func (b *buffer) WriteJoin(s ...string) (n int, err error) {
	for _, s := range s {
		x, _ := b.WriteString(s)
		n += x
	}
	return n, nil
}

// frontend represents the current frontend state for a given io.Writer - this
// is typically from a player's network connection.
//
// NOTE: The account hash is NOT considered a valid account until it is
// registered as inuse in the frontent.accounts tracking map. I.E. we may have
// the account but not the password yet.
type frontend struct {
	output   io.Writer // Writer to send output text to
	buf      *buffer   // Buffered text written to output when next prompt written
	input    []byte    // The input text we are currently processing
	nextFunc func()    // The next frontend function to be called by Parse
	player   has.Thing // The current player instance (may be ingame or not)
	account  string    // The current account hash (also key to accounts)
	err      error     // First error to occur else nil
}

// New returns an instance of frontend initialised with the given io.Writer.
// The io.Writer is used to send responses back from calling Parse. The new
// frontend is initialised with an output buffer and nextFunc setup to call
// greetingDisplay.
func New(output io.Writer) *frontend {
	f := &frontend{
		buf:    &buffer{new(bytes.Buffer)},
		output: output,
	}
	f.nextFunc = f.greetingDisplay
	return f
}

// Close makes sure the player is no longer 'in game' and frees up resources
// held the the instance of frontent. If the player is 'in game' a "QUIT"
// command will be issued.
func (f *frontend) Close() {

	// Just return if we already have an error
	if f.err != nil {
		return
	}
	f.err = closedError{}

	// If player is still in the game force them to quit
	if stats.Find(f.player) {
		cmd.Parse(f.player, "QUIT")
	}

	// Remove account from inuse list
	accounts.Lock()
	delete(accounts.inuse, f.account)
	accounts.Unlock()

	// Free up resources
	f.buf = nil
	f.player = nil
	f.output = nil
	f.nextFunc = nil
}

// Parse is the main input/output processing method for frontend. The input is
// stripped of leading and trailing whitespace before being stored in the
// frontend state. Any response from processing the input is written to the
// io.Writer passed to the initial New function that created the frontend. If
// the frontend is closed during processing a closedError will be returned else
// nil.
func (f *frontend) Parse(input []byte) error {

	// If we already have an error just return it
	if f.err != nil {
		return f.err
	}

	// Trim whitespace from input and process it
	f.input = bytes.TrimSpace(input)
	f.nextFunc()

	// If we have an output buffer write out it content
	if f.buf != nil {
		if len(f.input) > 0 || f.buf.Len() > 0 {
			f.buf.WriteByte('\n')
		}
		f.buf.WriteByte('>')
		f.output.Write(f.buf.Bytes())
		f.buf.Reset()
	}
	return f.err
}
