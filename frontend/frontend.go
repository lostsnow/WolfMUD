// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/config"
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

// NOTE: The account hash is NOT considered a valid account until it is
// registered as inuse in the frontent.accounts tracking map. I.E. we may have
// the account but not the password yet.
type frontend struct {
	buf      *bytes.Buffer // Buffered text written to output when next prompt written
	output   io.Writer     // Writer to send output text to
	input    []byte        // The input text we are currently processing
	nextFunc func()        // The next frontend function to call to process current input
	player   has.Thing     // The current player instance (may be ingame or not)
	account  string        // The current account hash
	err      error         // Contains the first error to occur else nil
}

func New(output io.Writer) *frontend {
	f := &frontend{
		buf:    new(bytes.Buffer),
		output: output,
	}
	f.nextFunc = f.greetingDisplay
	return f
}

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

func (f *frontend) Parse(input []byte) error {
	f.input = bytes.TrimSpace(input)
	f.nextFunc()
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

// GREETING

func (f *frontend) greetingDisplay() {
	f.buf.Write(config.Server.Greeting)
	f.accountDisplay()
}
