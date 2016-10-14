// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/recordjar"

	"bytes"
	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
)

// login embeds a frontend instance adding fields and methods specific to
// account logins.
type login struct {
	*frontend
}

// NewLogin returns a login with the specified frontend embedded. The returned
// login can be used for processing the logging in of accounts.
func NewLogin(f *frontend) (l *login) {
	l = &login{f}
	l.accountDisplay()
	return
}

// accountDisplay asks for the player's account ID so that they can log into
// the system.
func (l *login) accountDisplay() {
	l.buf.WriteString("Enter your account ID or just press enter to create a new account, enter QUIT to leave the server:")
	l.nextFunc = l.accountProcess
}

// accountProcess takes the current input and stores it in the current state as
// an account ID hash. At this point it is not validated yet just stored.
func (l *login) accountProcess() {
	switch {
	case len(l.input) == 0:
		NewAccount(l.frontend)
	case bytes.Equal(l.input, []byte("QUIT")):
		l.Close()
	default:
		hash := md5.Sum(l.input)
		l.account = hex.EncodeToString(hash[:])
		l.passwordDisplay()
	}
}

// passwordDisplay asks for the player's password for their account ID.
func (l *login) passwordDisplay() {
	l.buf.WriteString("Enter the password for your account ID or just press enter to cancel:")
	l.nextFunc = l.passwordProcess
}

// passwordProcess takes the current input and treats is as the player's
// password for logging into the system. If no password is entered processing
// goes back to asking for the players account ID. If the account ID is valid
// and the password is correct we load the player data and move on to
// displaying the main menu. If either the account ID or password is invalid we
// go back to asking for an account ID.
func (l *login) passwordProcess() {

	// If no password given go back and ask for an account ID.
	if len(l.input) == 0 {
		l.buf.WriteString("Login cancelled.\n\n")
		l.accountDisplay()
		return
	}

	// Can we open the account file? The filename is the MD5 hash of the account
	// ID. That way the filename is of a known format [0-9a-f]{32}\.wrj and we
	// don't have to trust user input for filenames hitting the filesystem.
	p := filepath.Join(config.Server.DataDir, "players", l.account+".wrj")
	wrj, err := os.Open(p)
	if err != nil {
		log.Printf("Error opening account: %s", err)
		l.buf.WriteString("Acount ID or password is incorrect.\n\n")
		l.accountDisplay()
		return
	}

	// Read the account file as a recordjar
	jar := recordjar.Read(wrj, "description")
	if err := wrj.Close(); err != nil {
		log.Printf("Error closing account: %s", err)
		l.buf.WriteString("Acount ID or password is incorrect.\n\n")
		l.accountDisplay()
		return
	}

	// The recordjar should have at least two records: account header and player.
	// If not something is wrong with the data.
	if len(jar) < 2 {
		log.Printf("Account file corrupted: %s.wrj", l.account)
		l.buf.WriteString("Sorry, there is a problem with your account, please contact the admins.\n\n")
		l.accountDisplay()
		return
	}

	var (
		record   = jar[0]
		salt     = recordjar.Decode.Bytes(record["SALT"])
		password = recordjar.Decode.String(record["PASSWORD"])
		hash     = sha512.Sum512(append(salt, l.input...))
	)

	// Check password is valid
	if (base64.URLEncoding.EncodeToString(hash[:])) != password {
		log.Printf("Password invalid for: %s.wrj", l.account)
		l.buf.WriteString("Acount ID or password is incorrect.\n\n")
		l.accountDisplay()
		return
	}

	// Drop header record from jar
	jar = jar[1:]

	// Check if account already in use to prevent multiple logins
	accounts.Lock()
	if _, inuse := accounts.inuse[l.account]; inuse {
		log.Printf("Account already logged in: %s", l.account)
		l.buf.WriteString("Acount is already logged in. If your connection to the server was unceramoniously terminated you may need to wait a while for the account to automatically logout.\n\n")
		l.accountDisplay()
		accounts.Unlock()
		return
	}
	accounts.inuse[l.account] = struct{}{}
	accounts.Unlock()

	// Assemble player
	record = jar[0]
	l.player = attr.NewThing()
	l.player.(*attr.Thing).Unmarshal(1, record)
	l.player.Add(attr.NewLocate(nil))
	l.player.Add(attr.NewPlayer(l.output))

	// Greet returning player
	l.buf.WriteString("Welcome back ")
	l.buf.WriteString(attr.FindName(l.player).Name("Someone"))
	l.buf.WriteString("!\n")

	NewMenu(l.frontend)
}
