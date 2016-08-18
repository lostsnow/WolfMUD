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

// accounts is used to track which accounts are logged in and in use.
var accounts struct {
	inuse map[string]struct{}
	sync.Mutex
}

// init is used to initialise the map used in account tracking.
func init() {
	accounts.inuse = make(map[string]struct{})
}

// EndOfDataError represents the fact that no more data is expected to be
// returned. For example the QUIT command has been used.
type EndOfDataError struct{}

func (e EndOfDataError) Error() string {
	return "End of data - player quitting"
}

func (e EndOfDataError) Temporary() bool {
	return true
}

type Driver struct {
	buf      *bytes.Buffer
	output   io.Writer
	input    []byte
	nextFunc func()
	write    bool
	player   has.Thing
	name     string
	account  string
	err      error
}

func NewDriver(output io.Writer) *Driver {
	d := &Driver{
		buf:    new(bytes.Buffer),
		output: output,
		write:  true,
	}
	d.nextFunc = d.greetingDisplay

	return d
}

func (d *Driver) Close() {

	if stats.Find(d.player) {
		cmd.Parse(d.player, "QUIT")
	}

	// Remove account from inuse list
	accounts.Lock()
	delete(accounts.inuse, d.account)
	accounts.Unlock()

	d.write = false
	d.buf = nil
	d.player = nil
	d.output = nil
	d.nextFunc = nil
}

func (d *Driver) Parse(input []byte) error {
	d.input = bytes.TrimSpace(input)
	d.nextFunc()
	if d.write {
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
