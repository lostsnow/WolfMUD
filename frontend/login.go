// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/recordjar"

	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
)

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
	jar = jar[1:]

	// Check if account already in use to prevent multiple logins
	accounts.Lock()
	if _, inuse := accounts.inuse[d.account]; inuse {
		log.Printf("Account already logged in: %s", d.account)
		d.buf.WriteString("Acount is already logged in. If your connection to the server was unceramoniously terminated you may need to wait a while for the account to automatically logout.\n")
		d.accountDisplay()
		accounts.Unlock()
		return
	}
	accounts.inuse[d.account] = struct{}{}
	accounts.Unlock()

	// Assemble player
	d.player = attr.NewThing()
	d.player.(*attr.Thing).Unmarshal(0, jar[0])
	d.player.Add(attr.NewLocate(nil))
	d.player.Add(attr.NewPlayer(d.output))

	// Greet returning player
	d.buf.WriteString("Welcome back ")
	d.buf.WriteString(attr.FindName(d.player).Name("Someone"))
	d.buf.WriteString("!\n")

	d.menuDisplay()
}
