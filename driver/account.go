// Copyright 2014 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package driver

import (
	"code.wolfmud.org/WolfMUD.git/entities/mobile/player"
	"code.wolfmud.org/WolfMUD.git/utils/config"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"

	"log"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// account is a driver for creating new accounts/players
type account struct {
	*driver
	password string
	name     string
	gender   string
}

// newAccount creates a new account driver from the current driver.
func (d *driver) newAccount() func() {
	a := account{driver: d}

	a.Respond("Your account ID can be anything you want: email address, book title, film title, a quote - anything that you can remember. You can use upper and lower case characters, numbers and symbols. The only restriction is it has to be [CYAN]%d[WHITE] characters or more in length. This is [CYAN]NOT[WHITE] your character's name.\n", config.AccountIdMin)
	a.needAccount()
	return a.checkAccount
}

func (a *account) needAccount() {
	a.Respond("Enter your new account ID:")
	a.next = a.checkAccount
}

func (a *account) checkAccount() {
	len := utf8.RuneCountInString(a.input)
	min := config.AccountIdMin

	if len < min {
		a.Respond("[RED]You only entered %d characters, minimum length is %d characters.", len, min)
		a.needAccount()
		return
	}

	a.account = player.HashAccount(a.input)
	a.explainPassword()
}

func (a *account) explainPassword() {
	a.Respond("Passwords must be at least [CYAN]%d[WHITE] charaters long and may contain upper and lower case letters, numbers and symbols. Please [CYAN]DO NOT[WHITE] reuse an existing password from your email, online banking or any other accounts!\n", config.AccountPasswordMin)
	a.needPassword()
}

func (a *account) needPassword() {
	a.Respond("Enter a password for your account:")
	a.next = a.checkPassword
}

func (a *account) checkPassword() {
	len := utf8.RuneCountInString(a.input)
	min := config.AccountPasswordMin

	if len < min {
		a.Respond("[RED]You only entered %d characters, minimum length is %d characters.", len, min)
		a.needPassword()
		return
	}

	a.password = a.input
	a.needVerify()
}

func (a *account) needVerify() {
	a.Respond("Please verify your password by entering it again:")
	a.next = a.checkVerify
}

func (a *account) checkVerify() {
	if a.password != a.input {
		a.Respond("[RED]Passwords do not match. Please try again.")
		a.needPassword()
		return
	}

	a.explainName()
}

func (a *account) explainName() {
	a.Respond("The name for you character must be a single word. It is prefferable that you use upper or lower cased characters in the range A to Z without diacritics as this could make it difficult for others to talk to and interact with you.\n")
	a.needName()
}

func (a *account) needName() {
	a.Respond("Enter a name for your character:")
	a.next = a.checkName
}

func (a *account) checkName() {
	if a.input == "" {
		a.needName()
		return
	}

	for _, r := range []rune(a.input) {
		if !unicode.IsLetter(r) {
			a.Respond("[RED]Names can only contain upper or lower cased letters in the range A to Z.")
			a.needName()
			return
		}
	}

	a.name = a.input
	a.needGender()
}

func (a *account) needGender() {
	a.Respond("What gender would you like your character to be?\n\n\tF - Female\n\tM - Male\n")
	a.next = a.checkGender
}

func (a *account) checkGender() {
	switch strings.ToUpper(a.input) {
	case "M", "MALE":
		a.gender = "male"
	case "F", "FEMALE":
		a.gender = "female"
	default:
		a.Respond("[RED]Please enter M, F, Male or Female.")
		a.needGender()
		return
	}

	a.create()
}

func (a *account) create() {

	password, salt := player.HashPassword(a.password)

	// Setup player
	e := recordjar.Encoder{}
	e.Keyword("type", "player")
	e.Keyword("ref", "player")
	e.String("account", a.account)
	e.String("password", string(password[:]))
	e.String("salt", salt)
	e.String("name", a.name)
	e.Keyword("gender", a.gender)
	e.Time("created", time.Now())

	var err error

	// Write out player file
	if err = player.Save(e); err != nil {
		a.Respond("[RED]Oops, there was an error creating your account :(")
		log.Printf("Error creating account: %s", err)
		a.needName()
		return
	}

	// Load player from written file
	a.player, err = player.Load(a.account, a.password)
	if err != nil {
		a.Respond("[RED]Oops, there was an error setting up your account :(")
		log.Printf("Error setting up account: %s", err)
		a.needName()
		return
	}

	// Log player in
	//
	// NOTE: We could take our encoder, wrap it in a decoder and unmarshal the
	// player. However by using the normal login method we make sure any
	// additional processing is carried out.
	//
	// TODO: Should this be done earlier to 'reserve' the account name? Then we
	// wouldn't go all the way through the account creation process to possibly
	// have it fail.
	if err = a.login(); err != nil {
		a.Respond("[RED]That account is already logged in!")
		log.Printf("Error setting up account: %s", err)
		a.needName()
		return
	}

	a.newMenu()
}
