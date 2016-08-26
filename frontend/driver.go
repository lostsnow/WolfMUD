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
// NOTE: A Driver.account hash is NOT valid until it has been registered as
// being inuse.
var accounts struct {
	inuse map[string]struct{}
	sync.Mutex
}

// init is used to initialise the map used in account tracking.
func init() {
	accounts.inuse = make(map[string]struct{})
}

// driverClosedError represents the fact that Close has been called on a Driver
// instance releasing it's resources and that the instance should be discarded.
// As interaction with the error is via the standard error and comms.temporary
// interfaces it does not need to be exported.
type driverClosedError struct{}

// Error implements the error interface for errors and returns descriptive text
// for the driverClosedError error.
func (_ driverClosedError) Error() string {
	return "frontend driver closed"
}

// Temporary always returns true for any driverClosedError. A driverClosedError
// does not bring down the network connection to the player - a comms.client
// instance can still send and receive network data directly.
func (_ driverClosedError) Temporary() bool {
	return true
}

// NOTE: The account hash is NOT considered a valid account until it is
// registered as inuse in the frontent.accounts tracking map. I.E. we may have
// the account but not the password yet.
type Driver struct {
	buf      *bytes.Buffer // Buffered text written to output when next prompt written
	output   io.Writer     // Writer to send output text to
	input    []byte        // The input text we are currently processing
	nextFunc func()        // The next driver function to call to process current input
	player   has.Thing     // The current player instance (may be ingame or not)
	account  string        // The current account hash
	err      error         // Contains the first error to occur else nil
}

func NewDriver(output io.Writer) *Driver {
	d := &Driver{
		buf:    new(bytes.Buffer),
		output: output,
	}
	d.nextFunc = d.greetingDisplay
	return d
}

func (d *Driver) Close() {

	// Just return if we already have an error
	if d.err != nil {
		return
	}
	d.err = driverClosedError{}

	// If player is still in the game force them to quit
	if stats.Find(d.player) {
		cmd.Parse(d.player, "QUIT")
	}

	// Remove account from inuse list
	accounts.Lock()
	delete(accounts.inuse, d.account)
	accounts.Unlock()

	// Free up resources
	d.buf = nil
	d.player = nil
	d.output = nil
	d.nextFunc = nil
}

func (d *Driver) Parse(input []byte) error {
	d.input = bytes.TrimSpace(input)
	d.nextFunc()
	if d.buf != nil {
		if len(d.input) > 0 || d.buf.Len() > 0 {
			d.buf.WriteByte('\n')
		}
		d.buf.WriteByte('>')
		d.output.Write(d.buf.Bytes())
		d.buf.Reset()
	}
	return d.err
}

// GREETING

func (d *Driver) greetingDisplay() {
	d.buf.Write(config.Server.Greeting)
	d.accountDisplay()
}
