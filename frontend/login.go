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

func (f *frontend) accountDisplay() {
	f.buf.WriteString("Please enter your account ID or just press enter to create a new account:")
	f.nextFunc = f.accountProcess
}

func (f *frontend) accountProcess() {
	if len(f.input) == 0 {
		f.buf.WriteString("Account creation unavailable.\n")
		f.accountDisplay()
		return
	}

	hash := md5.Sum(f.input)
	f.account = hex.EncodeToString(hash[:])
	f.passwordDisplay()
}

func (f *frontend) passwordDisplay() {
	f.buf.WriteString("Please enter the password for your account or just press enter to abort:")
	f.nextFunc = f.passwordProcess
}

func (f *frontend) passwordProcess() {
	if len(f.input) == 0 {
		f.buf.WriteString("No Password given.\n")
		f.accountDisplay()
		return
	}

	// Can we get the account file?
	p := filepath.Join(config.Server.DataDir, "players", f.account+".wrj")
	wrj, err := os.Open(p)
	if err != nil {
		log.Printf("Error opening account: %s", err)
		f.buf.WriteString("Acount ID or password is incorrect.\n")
		f.accountDisplay()
		return
	}

	jar := recordjar.Read(wrj, "description")
	if err := wrj.Close(); err != nil {
		log.Printf("Error closing account: %s", err)
		f.buf.WriteString("Acount ID or password is incorrect.\n")
		f.accountDisplay()
		return
	}

	data := jar[0]
	hash := sha512.Sum512(append(data["SALT"], f.input...))
	if (base64.URLEncoding.EncodeToString(hash[:])) != string(data["PASSWORD"]) {
		f.buf.WriteString("Acount ID or password is incorrect.\n")
		f.accountDisplay()
		return
	}
	jar = jar[1:]

	// Check if account already in use to prevent multiple logins
	accounts.Lock()
	if _, inuse := accounts.inuse[f.account]; inuse {
		log.Printf("Account already logged in: %s", f.account)
		f.buf.WriteString("Acount is already logged in. If your connection to the server was unceramoniously terminated you may need to wait a while for the account to automatically logout.\n")
		f.accountDisplay()
		accounts.Unlock()
		return
	}
	accounts.inuse[f.account] = struct{}{}
	accounts.Unlock()

	// Assemble player
	f.player = attr.NewThing()
	f.player.(*attr.Thing).Unmarshal(0, jar[0])
	f.player.Add(attr.NewLocate(nil))
	f.player.Add(attr.NewPlayer(f.output))

	// Greet returning player
	f.buf.WriteString("Welcome back ")
	f.buf.WriteString(attr.FindName(f.player).Name("Someone"))
	f.buf.WriteString("!\n")

	f.menuDisplay()
}
