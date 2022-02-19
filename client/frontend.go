// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.
package client

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/mailbox"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/text"
)

var verifyName = regexp.MustCompile(`^[a-zA-Z]+$`)

var (
	accountsMux sync.RWMutex
	accounts    = make(map[string]struct{})
)

func (c *client) input() string {
	var input string
	var err error
	r := bufio.NewReaderSize(c, 80)
	c.SetReadDeadline(time.Now().Add(cfg.frontendTimeout))
	if input, err = r.ReadString('\n'); err != nil {
		c.setError(err)
	}
	return clean(input)
}

// frontend implements a question/answer flow with the player. Currently
// implements logon and account+player creation.
func (c *client) frontend() bool {

	// Valid frontend stages
	const (
		welcome = iota
		account
		password
		explainAccount
		newAccount
		newPassword
		verifyPassword
		name
		gender
		create
		cancelCreate
		finished
	)

	var buf []byte
	mailbox.Suffix(c.uid, "\n"+text.Magenta+">")

	for stage := welcome; ; {

		// Write question for current stage to player
		switch stage {
		case welcome:
			buf = append(buf, cfg.greeting...)
			stage = account
			continue

		case account:
			buf = append(buf, "Enter your account ID or just press enter to create a new account, enter QUIT to leave the server:"...)
			delete(c.As, core.Account)
			delete(c.As, core.Password)
			delete(c.As, core.Salt)
			delete(c.Int, core.Created)

		case password:
			buf = append(buf, "Enter the password for your account ID or just press enter to cancel:"...)

		case explainAccount:
			buf = append(buf, "Your account ID can be anything you can remember: an email address, a book title, a film title, a quote. You can use upper and lower case characters, numbers and symbols. The only restriction is it has to be at least "...)
			buf = append(buf, strconv.Itoa(cfg.accountMin)...)
			buf = append(buf, " characters long.\n\nThis is NOT your character's name, it is for your account ID for logging in only.\n\n"...)
			stage = newAccount
			continue

		case newAccount:
			buf = append(buf, "Enter text to use for your new account ID or just press enter to cancel:"...)

		case newPassword:
			buf = append(buf, "Enter a password to use for your account ID or just press enter to cancel:"...)

		case verifyPassword:
			buf = append(buf, "Enter your password again to confirm or just press enter to cancel:"...)

		case name:
			buf = append(buf, "Enter a name for your character or just press enter to cancel:"...)

		case gender:
			buf = append(buf, "Would you like "...)
			buf = append(buf, c.As[core.Name]...)
			buf = append(buf, " to be male or female? Or just press enter to cancel."...)

		case create:
			c.createPlayer()
			mailbox.Suffix(c.uid, "")
			mailbox.Send(c.uid, true, text.Good+"You step into another world...\n\n")
			stage = finished
			continue

		case cancelCreate:
			buf = append(buf, text.Bad...)
			buf = append(buf, "Account creation cancelled.\n\n"...)
			buf = append(buf, text.Reset...)
			stage = account
			continue

		case finished:
			return true
		}

		// Output message to player and get an answer to question
		if len(buf) > 0 {
			mailbox.Send(c.uid, true, string(buf))
			buf = buf[:0]
		}
		input := c.input()
		if c.error() != nil {
			return false
		}

		// Process answer to question for current stage
		switch stage {
		case account:
			if input == "QUIT" {
				return false
			}
			if input == "" {
				stage = explainAccount
				continue
			}
			hash := md5.Sum([]byte(input))
			c.As[core.Account] = hex.EncodeToString(hash[:])
			stage = password

		case password:
			if input == "" {
				buf = append(buf, text.Bad...)
				buf = append(buf, "Account ID or password is incorrect.\n\n"...)
				buf = append(buf, text.Reset...)
				stage = account
				continue
			}
			f := filepath.Join(cfg.playerPath, c.As[core.Account]+".wrj")
			wrj, err := os.Open(f)
			if err != nil {
				buf = append(buf, text.Bad...)
				buf = append(buf, "Account ID or password is incorrect.\n\n"...)
				buf = append(buf, text.Reset...)
				c.Log("Invalid account")
				stage = account
				continue
			}

			jar := recordjar.Read(wrj, "description")
			wrj.Close()
			rec := jar[0]
			c.As[core.Salt] = decode.String(rec["SALT"])
			hash := sha512.Sum512([]byte(c.As[core.Salt] + input))
			c.As[core.Password] = base64.URLEncoding.EncodeToString(hash[:])
			c.Int[core.Created] = decode.DateTime(rec["CREATED"]).Unix()
			if c.As[core.Password] != decode.String(rec["PASSWORD"]) {
				buf = append(buf, text.Bad...)
				buf = append(buf, "Account ID or password is incorrect.\n\n"...)
				buf = append(buf, text.Reset...)
				c.Log("Invalid password for: %s", c.As[core.Account])
				stage = account
				continue
			}

			accountsMux.RLock()
			_, active := accounts[c.As[core.Account]]
			accountsMux.RUnlock()

			if active {
				buf = append(buf, text.Bad...)
				buf = append(buf, "The account ID is already logged in. If your connection to the server was unceramoniously terminated you may need to wait a while for the account to automatically logout.\n\n"...)
				buf = append(buf, text.Reset...)
				stage = account
				continue
			}

			accountsMux.Lock()
			accounts[c.As[core.Account]] = struct{}{}
			accountsMux.Unlock()
			c.Log("Login for: %s", c.As[core.Account])
			c.assemblePlayer(jar[1:])
			mailbox.Suffix(c.uid, "")
			mailbox.Send(c.uid, true, text.Good+"Welcome back "+c.As[core.Name]+"!\n\n")
			stage = finished

		case newAccount:
			if input == "" {
				stage = cancelCreate
				continue
			}
			if len(input) < cfg.accountMin {
				buf = append(buf, text.Bad...)
				buf = append(buf, "Account ID must be at least "...)
				buf = append(buf, strconv.Itoa(cfg.accountMin)...)
				buf = append(buf, " characters long.\n\n"...)
				buf = append(buf, text.Reset...)
				stage = newAccount
				continue
			}
			hash := md5.Sum([]byte(input))
			c.As[core.Account] = hex.EncodeToString(hash[:])
			if _, err := os.Stat(filepath.Join(cfg.playerPath, c.As[core.Account]+".wrj")); err == nil {
				buf = append(buf, text.Bad...)
				buf = append(buf, "The specified Account ID is currently unavailable.\n\n"...)
				buf = append(buf, text.Reset...)
				continue
			}
			stage = newPassword

		case newPassword:
			if input == "" {
				stage = cancelCreate
				continue
			}
			if len(input) < cfg.passwordMin {
				buf = append(buf, text.Bad...)
				buf = append(buf, "Password must be at least "...)
				buf = append(buf, strconv.Itoa(cfg.passwordMin)...)
				buf = append(buf, " characters long.\n\n"...)
				buf = append(buf, text.Reset...)
				stage = newPassword
				continue
			}
			salt := make([]byte, cfg.saltLength)
			rand.Read(salt)
			c.As[core.Salt] = base64.URLEncoding.EncodeToString(salt)
			hash := sha512.Sum512([]byte(c.As[core.Salt] + input))
			c.As[core.Password] = base64.URLEncoding.EncodeToString(hash[:])
			stage = verifyPassword

		case verifyPassword:
			if input == "" {
				stage = cancelCreate
				continue
			}
			hash := sha512.Sum512([]byte(c.As[core.Salt] + input))
			if c.As[core.Password] != base64.URLEncoding.EncodeToString(hash[:]) {
				buf = append(buf, text.Bad...)
				buf = append(buf, "Passwords do not match.\n\n"...)
				buf = append(buf, text.Reset...)
				stage = newPassword
				continue
			}
			stage = name

		case name:
			if input == "" {
				stage = cancelCreate
				continue
			}
			if verifyName.FindString(input) == "" {
				buf = append(buf, text.Bad...)
				buf = append(buf, "A character's name must only contain the upper or lower cased letters 'a' through 'z'. Using other letters, such as those with accents, will make it harder for other players to interact with you if they cannot type your character's name.\n\n"...)
				buf = append(buf, text.Reset...)
				continue
			}
			if len(input) < 3 {
				buf = append(buf, text.Bad...)
				buf = append(buf, "A character's name must be a minimum of 3 letters in length.\n\n"...)
				buf = append(buf, text.Reset...)
				continue
			}
			c.As[core.Name] = input
			stage = gender

		case gender:
			switch strings.ToUpper(input) {
			case "":
				stage = cancelCreate
			case "M", "MALE":
				c.As[core.Gender] = "MALE"
				stage = create
			case "F", "FEMALE":
				c.As[core.Gender] = "FEMALE"
				stage = create
			default:
				buf = append(buf, text.Bad...)
				buf = append(buf, "Please specify male or female.\n\n"...)
				buf = append(buf, text.Reset...)
			}
		}
	}
	return false
}

func (c *client) enterWorld() {
	core.BWL.Lock()
	defer core.BWL.Unlock()
	c.Ref[core.Where] = core.WorldStart[rand.Intn(len(core.WorldStart))]
	c.Ref[core.Where].Who[c.uid] = c.Thing
}

// TODO(diddymus): Need to add a proper player file upgrade path + versions
func (c *client) assemblePlayer(jar recordjar.Jar) {
	store := make(map[string]*core.Thing)
	invs := make(map[string][]string)

	// TODO(diddymus): add bounds cheddcking for broken jar...
	pref := decode.Keyword(jar[0]["REF"])

	// Upgrade and add HELTH if missing
	if _, found := jar[0]["HEALTH"]; !found {
		jar[0]["HEALTH"] = []byte("AFTER→10S MAXIMUM→30 RESTORE→2")
	}
	// If old HEALTH record upgrade fields
	jar[0]["HEALTH"] =
		bytes.ReplaceAll(jar[0]["HEALTH"], []byte("REGENERATES"), []byte("RESTORE"))
	jar[0]["HEALTH"] =
		bytes.ReplaceAll(jar[0]["HEALTH"], []byte("FREQUENCY"), []byte("AFTER"))

	// Load player jar into temporary store
	for _, record := range jar {
		ref := decode.Keyword(record["REF"])
		invs[ref] = decode.KeywordList(record["INVENTORY"])
		store[ref] = core.NewThing()
		store[ref].Unmarshal(record)
	}

	// Link up intentories in temporary store
	for _, item := range store {
		for _, ref := range invs[item.As[core.Ref]] {
			disabled := ref[0] == '!'
			if disabled {
				ref = ref[1:]
			}
			if what, ok := store[ref]; ok {
				if disabled {
					item.Out[what.As[core.UID]] = what
				} else {
					item.In[what.As[core.UID]] = what
				}
			} else {
				c.Log("load warning, ref not found for inventory: %s\n", ref)
			}
		}
	}

	// Some shenanigans to move the player Thing into our current Thing keeping
	// our original UID. This prevents disassociating from our messenger mailbox.
	p := store[pref].Copy(true)
	for x, alias := range p.Any[core.Alias] {
		if alias == p.As[core.UID] {
			p.Any[core.Alias][x] = c.uid
			break
		}
	}
	p.As[core.UID] = c.uid
	p.As[core.Ref] = c.uid
	p.As[core.Account] = c.As[core.Account]
	p.As[core.Password] = c.As[core.Password]
	p.As[core.Salt] = c.As[core.Salt]
	p.Int[core.Created] = c.Int[core.Created]
	p.Is |= core.Player
	p.Is &^= core.NPC
	c.Thing.Free()
	c.Thing = p
	c.InitOnce(nil)
	clearOrigins(c.Thing)

	// Set "MY" dynamic alias for player's immediate inventory items.
	for _, item := range p.In {
		item.As[core.DynamicQualifier] = "MY"
	}

	// Tear down temporary store
	for ref, item := range store {
		item.Free()
		delete(store, ref)
	}

	return
}

// clearOrigins will remove any Thing.Ref[Origin], recursively for inventories,
// for a Thing. By default assembling a player will make all items unique due
// to calling Thing.InitOnce - this function undoes that.
func clearOrigins(item *core.Thing) {
	delete(item.Ref, core.Origin)
	for _, item := range item.In {
		clearOrigins(item)
	}
	for _, item := range item.Out {
		clearOrigins(item)
	}
}

func (c *client) createPlayer() {
	c.Is |= core.Player
	c.As[core.UName] = c.As[core.Name]
	c.As[core.TheName] = c.As[core.Name]
	c.As[core.UTheName] = c.As[core.Name]
	c.As[core.Description] = "An adventurer, just like you."
	c.As[core.DynamicAlias] = "PLAYER"
	c.Any[core.Alias] = []string{"PLAYER", strings.ToUpper(c.As[core.Name])}
	c.Any[core.Body] = []string{
		"HEAD",
		"FACE", "EAR", "EYE", "NOSE", "EYE", "EAR",
		"MOUTH", "UPPER_LIP", "LOWER_LIP",
		"NECK",
		"SHOULDER", "UPPER_ARM", "ELBOW", "LOWER_ARM", "WRIST",
		"HAND", "FINGER", "FINGER", "FINGER", "FINGER", "THUMB",
		"SHOULDER", "UPPER_ARM", "ELBOW", "LOWER_ARM", "WRIST",
		"HAND", "FINGER", "FINGER", "FINGER", "FINGER", "THUMB",
		"BACK", "CHEST",
		"WAIST", "PELVIS",
		"UPPER_LEG", "KNEE", "LOWER_LEG", "ANKLE", "FOOT",
		"UPPER_LEG", "KNEE", "LOWER_LEG", "ANKLE", "FOOT",
	}
	c.Int[core.Created] = time.Now().Unix()
	c.Int[core.HealthAfter] = (10 * time.Second).Nanoseconds()
	c.Int[core.HealthRestore] = 2
	c.Int[core.HealthCurrent] = 30
	c.Int[core.HealthMaximum] = 30
}
