// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package driver

import (
	"code.wolfmud.org/WolfMUD.git/entities/mobile/player"
	"code.wolfmud.org/WolfMUD.git/utils/sender"

	"errors"
	"log"
	"strings"
)

// accounts is a channel that should buffer a single map of logged in accounts
// keyed by the hash of the account. The map then acts as both data and lock.
// To access the accounts you take the lock by removing the map, use it, then
// put the map back into the channel to release the lock. While the map is in
// use other go routines will block until the map is put back and can be read
// again. As maps are reference types only a reference should actually go into
// the channel keeping things lightweight.
var (
	accounts chan map[string]struct{}
)

// Initialise accounts channel and map up front
func init() {
	accounts = make(chan map[string]struct{}, 1)
	accounts <- make(map[string]struct{})
}

func (d *driver) login() error {
	a := <-accounts
	defer func() { accounts <- a }()

	if _, found := a[d.account]; found {
		log.Printf("Duplicate login: %s", d.sender)
		return errors.New("Duplicate login")
	}

	log.Printf("Successful login: %s", d.sender)
	a[d.account] = struct{}{}

	return nil
}

func (d *driver) Logout() {
	a := <-accounts
	defer func() { accounts <- a }()

	// Check if we are already logged out and save time...
	if _, found := a[d.account]; !found {
		return
	}

	if d.player != nil && d.player.Locate() != nil {
		d.player.Parse("QUIT")
	}

	log.Printf("Logout: %s", d.sender)
	delete(a, d.account)
}

// driver is a very simple base type to handle login and menu type frontend
// processing. See login.go and menu.go for examples of drivers.
//
// TODO: Document writing drivers.
type driver struct {
	input   string
	account string
	next    func()
	player  *player.Player
	buff    buffer
	sender  sender.Interface
}

// buffer stores buffered messages sent by Respond. A call to flush flushes the
// buffers and clears them for reuse.
type buffer struct {
	format []string
	any    []interface{}
}

// flush processes the buffered messages sent using Respond and clears the
// buffers for reuse.
func (d *driver) flush() {
	if len(d.buff.format) > 0 {
		format := strings.Join(d.buff.format, "[WHITE]\n")
		d.sender.Send(format, d.buff.any...)
		d.buff.format, d.buff.any = d.buff.format[:0], d.buff.any[:0]
	}
}

// Respond buffers messages to send back to the current client. Send is
// modelled after fmt.Sprintf and takes parameters in the same way. The
// buffered messages are not sent until flush is called.
func (d *driver) Respond(format string, any ...interface{}) {
	d.buff.format = append(d.buff.format, format)
	d.buff.any = append(d.buff.any, any...)
}

// New creates a frontend driver associated with the passed sender.  Initially
// it is setup as a login driver.
func New(s sender.Interface) (d *driver) {
	d = &driver{}
	d.sender = s
	d.next = d.newLogin()
	d.Process("")
	return d
}

// Process takes input and stores it in the current driver. It then invokes the
// next function stored in the driver. When the invoked function completes the
// output buffer is flushed and all output is sent to the current sender.
func (d *driver) Process(input string) {
	d.input = input
	d.next()
	d.flush()
}

// IsQuitting returns true if the driver is trying to quit otherwise false.
func (d *driver) IsQuitting() bool {
	return d.next == nil
}
