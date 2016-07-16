// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/stats"

	"bytes"
	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
)

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

// ACCOUNT

func (d *Driver) accountDisplay() {
	d.buf.WriteString("Please enter your account ID or just press enter to create a new account:")
	d.nextFunc = d.accountProcess
}

func (d *Driver) accountProcess() {
	if len(d.input) == 0 {
		d.buf.WriteString("Account creation unavailable.\n")
		d.accountDisplay()
		return
	}

	hash := md5.Sum(d.input)
	d.account = hex.EncodeToString(hash[:])
	d.passwordDisplay()
}

// PASSWORD

func (d *Driver) passwordDisplay() {
	d.buf.WriteString("Please enter the password for your account or just press enter to abort:")
	d.nextFunc = d.passwordProcess
}

func (d *Driver) passwordProcess() {
	if len(d.input) == 0 {
		d.buf.WriteString("No Password given.\n")
		d.accountDisplay()
		return
	}

	// Can we get the account file?
	p := filepath.Join(config.Server.DataDir, "players", d.account+".wrj")
	f, err := os.Open(p)
	if err != nil {
		log.Printf("Error opening account: %s", err)
		d.buf.WriteString("Acount ID or password is incorrect.\n")
		d.accountDisplay()
		return
	}

	jar := recordjar.Read(f, "description")
	if err := f.Close(); err != nil {
		log.Printf("Error closing account: %s", err)
		d.buf.WriteString("Acount ID or password is incorrect.\n")
		d.accountDisplay()
		return
	}

	data := jar[0]
	hash := sha512.Sum512(append(data["SALT"], d.input...))
	if (base64.URLEncoding.EncodeToString(hash[:])) != string(data["PASSWORD"]) {
		d.buf.WriteString("Acount ID or password is incorrect.\n")
		d.accountDisplay()
		return
	}

	// Assemble player
	d.player = attr.NewThing()
	d.player.(*attr.Thing).Unmarshal(0, data)
	d.player.Add(attr.NewLocate(nil))
	d.player.Add(attr.NewPlayer(d.output))

	// Greet returning player
	d.buf.WriteString("Welcome back ")
	d.buf.WriteString(attr.FindName(d.player).Name("Someone"))
	d.buf.WriteString("!\n")

	d.menuDisplay()
}

// MENU

func (d *Driver) menuDisplay() {
	d.buf.Write([]byte(`
  Main Menu
  ---------

  1. Enter game
  0. Quit

Select an option:`))
	d.nextFunc = d.menuProcess
}

func (d *Driver) menuProcess() {
	if len(d.input) == 0 {
		return
	}
	switch string(d.input) {
	case "1":
		d.gameSetup()
	case "0":
		d.write = false
		d.err = EndOfDataError{}
	default:
		d.buf.WriteString("Invalid option selected.")
	}
}

// GAME

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
	c := cmd.Parse(d.player, string(d.input))
	if c == "QUIT" {
		d.write = true
		d.menuDisplay()
	}
}
