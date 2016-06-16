// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/stats"

	"bytes"
	"io"
)

type Driver struct {
	buf      *bytes.Buffer
	output   io.Writer
	input    []byte
	nextFunc func()
	write    bool
	player   *attr.Thing
	name     string
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
		d.err = cmd.Parse(d.player, "QUIT")
	}

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

func (d *Driver) greetingDisplay() {
	d.buf.Write(config.Server.Greeting)
	d.nameDisplay()
}

func (d *Driver) nameDisplay() {
	d.buf.Write([]byte("Enter a name for your character:"))
	d.nextFunc = d.nameProcess
}

func (d *Driver) nameProcess() {
	if len(d.input) == 0 {
		return
	}
	if len(d.input) < 4 {
		d.buf.Write([]byte("Name needs to be at least 4 characters long."))
		d.nameDisplay()
		return
	}

	d.name = string(d.input)
	d.gameSetup()
}

func (d *Driver) gameSetup() {
	d.player = attr.NewThing(
		attr.NewName(d.name),
		attr.NewDescription("This is an adventurer just like you."),
		attr.NewAlias(d.name),
		attr.NewInventory(),
		attr.NewLocate(nil),
	)
	d.player.Add(attr.NewPlayer(d.output))
	d.write = false

	i := (*attr.Start)(nil).Pick()
	i.Lock()
	i.Add(d.player)
	stats.Add(d.player)
	i.Unlock()

	cmd.Parse(d.player, "LOOK")
	d.nextFunc = d.gameRun
}

func (d *Driver) gameRun() {
	d.err = cmd.Parse(d.player, string(d.input))
}
