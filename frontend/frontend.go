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
	"bytes"
	"io"
	"sync"

	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/log"
	"code.wolfmud.org/WolfMUD.git/message"
	"code.wolfmud.org/WolfMUD.git/stats"
	"code.wolfmud.org/WolfMUD.git/text"
)

// accounts is used to track which (valid) accounts are logged in and in use.
// It's main purpose is to track logged in account IDs to prevent duplicate
// logins.
var accounts struct {
	sync.Mutex
	inuse map[string]struct{}
}

// init is used to initialise the map used in account ID tracking.
func init() {
	accounts.inuse = make(map[string]struct{})
}

// ClosedError represents the fact that Close has been called on a frontend
// instance releasing it's resources and that the instance should be discarded.
type ClosedError struct{}

// Error implements the error interface for errors and returns descriptive text
// for the ClosedError error.
func (ClosedError) Error() string {
	return "frontend closed"
}

// Temporary always returns true for a frontend.ClosedError. A ClosedError is
// considered temporary as recovery can be performed by creating a new frontend
// instance.
func (ClosedError) Temporary() bool {
	return true
}

// frontend represents the current frontend state for a given io.Writer - this
// is typically from a player's network connection.
type frontend struct {
	output   io.Writer       // Writer to send output text to
	buf      *message.Buffer // Buffered messages written with next prompt
	input    []byte          // The input text we are currently processing
	nextFunc func()          // The next frontend function called by Parse
	player   has.Thing       // The current player instance (ingame or not)
	account  string          // The current account hash (also key to accounts)
	err      error           // First error to occur else nil
	log      log.Conn        // Per connection logging
}

// New returns an initialised instance of frontend. The passed log.Conn is used
// by the instance for per-connection logging. The io.Writer is used to write
// data back to the client. The instance is also setup with greetingDisplay as
// the nextFunc to call.
func New(log log.Conn, output io.Writer) *frontend {
	f := &frontend{
		buf:    message.AcquireBuffer(),
		output: output,
		log:    log,
	}
	f.buf.OmitLF(true)
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
	f.err = ClosedError{}

	// If player is still in the game force them to quit
	if stats.Find(f.player) {
		cmd.Parse(f.player, "QUIT")
	}

	// Make sure any remaining messages are sent
	if f.buf != nil {
		f.buf.Deliver(f)
	}

	// Remove account from inuse list
	accounts.Lock()
	delete(accounts.inuse, f.account)
	accounts.Unlock()

	// Free up resources
	message.ReleaseBuffer(f.buf)
	f.buf = nil

	f.output = nil
	f.nextFunc = nil

	if f.player != nil {
		f.player.Free()
	}
	f.player = nil

	f = nil
}

// Parse is the main input/output processing method for frontend. The input is
// stripped of leading and trailing whitespace before being stored in the
// frontend state. Any response from processing the input is written to the
// io.Writer passed to the initial New function that created the frontend. If
// the frontend is closed during processing a frontend.ClosedError will be
// returned else nil.
func (f *frontend) Parse(input []byte) error {

	// If we already have an error just return it
	if f.err != nil {
		return f.err
	}

	// Trim whitespace from input and process it
	f.input = bytes.TrimSpace(input)
	f.nextFunc()

	// If we have a message buffer write out its content and a new prompt
	if f.buf != nil {
		f.buf.Deliver(f)
	}
	return f.err
}

// greetingDisplay displays the welcome message to players when they first
// connect to the server. The text displayed is stored in the server
// configuration file free text area. For more information on specifying the
// welcome message see the config package.
//
// The greeting does not have it's own source file or a NewGreeting function
// like other frontend helpers such as account or login as might be expected.
// This is due to the fact that frontend needs to initialise nextFunc with a
// func() type and greetingDisplay seems a logical choice.
func (f *frontend) greetingDisplay() {
	f.buf.Send(string(config.Server.Greeting))
	NewLogin(f)
}

// Write writes the specified byte slice to the associated client.
func (f *frontend) Write(b []byte) (n int, err error) {
	b = append(b, text.Prompt...)
	b = append(b, '>')
	n, err = f.output.Write(b)
	return
}

// Zero writes zero bytes into the passed slice
func Zero(data []byte) {
	if len(data) > 0 {
		data[0] = 0
		for i := 1; i < len(data); i *= 2 {
			copy(data[i:], data[:i])
		}
	}
}
