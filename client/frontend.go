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
	"errors"
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
	"code.wolfmud.org/WolfMUD.git/term"
	"code.wolfmud.org/WolfMUD.git/text"
)

var verifyName = regexp.MustCompile(`^[a-zA-Z]+$`)

var (
	accountsMux sync.RWMutex
	accounts    = make(map[string]struct{})
)

func (c *client) read() string {
	var err error
	r := bufio.NewReaderSize(c, inputBufferSize)
retry:
	c.SetReadDeadline(time.Now().Add(cfg.frontendTimeout))
	if c.input, err = r.ReadSlice('\n'); err != nil {
		if errors.Is(err, bufio.ErrBufferFull) {
			for ; errors.Is(err, bufio.ErrBufferFull); _, err = r.ReadSlice('\n') {
			}
			mailbox.Send(c.uid, true, text.Bad+"\nYou type too much! Try again.")
			goto retry
		}
		c.setError(err)
	}
	return clean(c.input)
}

type buffer struct {
	strings.Builder
}

func (b *buffer) Msg(s ...string) {
	b.WriteByte('\n')
	b.WriteString(text.Reset)
	for _, t := range s {
		b.WriteString(t)
	}
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

	buf := &buffer{}

	for stage := welcome; ; {

		// Write question for current stage to player
		switch stage {
		case welcome:
			buf.Msg(cfg.greeting)
			stage = account
			continue

		case account:
			buf.Msg("Enter your account ID or just press enter to create a new account, enter QUIT to leave the server.")
			delete(c.As, core.Account)
			delete(c.As, core.Password)
			delete(c.As, core.Salt)
			delete(c.Int, core.Created)

		case password:
			buf.Msg("Enter the password for your account ID or just press enter to cancel.")

		case explainAccount:
			buf.Msg("Your account ID can be anything you can remember: an email address, a book title, a film title, a quote. You can use upper and lower case characters, numbers and symbols. The only restriction is it has to be at least ", strconv.Itoa(cfg.accountMin), " characters long. This is NOT your character's name, it is for your account ID for logging in only.\n")
			stage = newAccount
			continue

		case newAccount:
			buf.Msg("Enter text to use for your new account ID or just press enter to cancel.")

		case newPassword:
			buf.Msg("Enter a password to use for your account ID or just press enter to cancel.")

		case verifyPassword:
			buf.Msg("Enter your password again to confirm or just press enter to cancel.")

		case name:
			buf.Msg("Enter a name for your character or just press enter to cancel.")

		case gender:
			buf.Msg("Would you like ", c.As[core.Name], " to be male, female or neutral? Or just press enter to cancel.")

		case create:
			c.createPlayer()
			buf.Msg(text.Good, "\nYou step into another world...\n")
			stage = finished
			continue

		case cancelCreate:
			buf.Msg(text.Bad, "Account creation cancelled.")
			stage = account
			continue
		}

		// Output message to player and get an answer to question
		if buf.Len() > 0 {
			mailbox.Send(c.uid, true, buf.String())
			buf.Reset()
		}

		// Finshed?
		if stage == finished {
			return true
		}

		input := c.read()
		if c.error() != nil {
			return false
		}
		buf.Msg(text.Prompt, ">", input)

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
				buf.Msg(text.Bad, "Account ID or password is incorrect.")
				stage = account
				continue
			}
			f := filepath.Join(cfg.playerPath, c.As[core.Account]+".wrj")
			wrj, err := os.Open(f)
			if err != nil {
				buf.Msg(text.Bad, "Account ID or password is incorrect.")
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
			c.Int[core.Created] = decode.DateTime(rec["CREATED"]).UnixNano()
			if len(rec["PERMISSIONS"]) > 0 {
				c.Any[core.Permissions] = decode.KeywordList(rec["PERMISSIONS"])
			}
			if c.As[core.Password] != decode.String(rec["PASSWORD"]) {
				buf.Msg(text.Bad, "Account ID or password is incorrect.")
				c.Log("Invalid password for: %s", c.As[core.Account])
				stage = account
				continue
			}

			accountsMux.RLock()
			_, active := accounts[c.As[core.Account]]
			accountsMux.RUnlock()

			if active {
				buf.Msg(text.Bad, "The account ID is already logged in. If your connection to the server was unceramoniously terminated you may need to wait a while for the account to automatically logout.")
				stage = account
				continue
			}

			accountsMux.Lock()
			accounts[c.As[core.Account]] = struct{}{}
			accountsMux.Unlock()
			c.Log("Login by: %s", c.As[core.Account])
			c.assemblePlayer(jar[1:])
			buf.Msg(text.Good, "\nWelcome back ", c.As[core.Name], "!\n")
			stage = finished

		case newAccount:
			if input == "" {
				stage = cancelCreate
				continue
			}
			if len(input) < cfg.accountMin {
				buf.Msg(text.Bad, "Account ID must be at least ", strconv.Itoa(cfg.accountMin), " characters long.")
				stage = newAccount
				continue
			}
			hash := md5.Sum([]byte(input))
			c.As[core.Account] = hex.EncodeToString(hash[:])
			if _, err := os.Stat(filepath.Join(cfg.playerPath, c.As[core.Account]+".wrj")); err == nil {
				buf.Msg(text.Bad, "The specified Account ID is currently unavailable.")
				continue
			}
			stage = newPassword

		case newPassword:
			if input == "" {
				stage = cancelCreate
				continue
			}
			if len(input) < cfg.passwordMin {
				buf.Msg(text.Bad, "Password must be at least ", strconv.Itoa(cfg.passwordMin), " characters long.")
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
				buf.Msg(text.Bad, "Passwords do not match.")
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
				buf.Msg(text.Bad, "A character's name must only contain the upper or lower cased letters 'a' through 'z'. Using other letters, such as those with accents, will make it harder for other players to interact with you if they cannot type your character's name.")
				continue
			}
			if len(input) < 3 || len(input) > 15 {
				buf.Msg(text.Bad, "A character's name must be a minimum of 3 letters in length and a maximum of 15 letters in length.")
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
			case "N", "NEUTRAL":
				c.As[core.Gender] = "NEUTRAL"
				stage = create
			default:
				buf.Msg(text.Bad, "Please specify male, female or neutral.")
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
		jar[0]["HEALTH"] = []byte("AFTER→1M MAXIMUM→30 RESTORE→2")
	}
	// If old HEALTH record upgrade fields
	jar[0]["HEALTH"] =
		bytes.ReplaceAll(jar[0]["HEALTH"], []byte("REGENERATES"), []byte("RESTORE"))
	jar[0]["HEALTH"] =
		bytes.ReplaceAll(jar[0]["HEALTH"], []byte("FREQUENCY"), []byte("AFTER"))
	// Upgrade if no natural armour (also update old health record)
	if _, found := jar[0]["ARMOUR"]; !found {
		jar[0]["ARMOUR"] = []byte("10")
		jar[0]["HEALTH"] = []byte("AFTER→1M MAXIMUM→30 RESTORE→2")
	}
	// Upgrade if no natural damage
	if _, found := jar[0]["DAMAGE"]; !found {
		jar[0]["DAMAGE"] = []byte("2+2")
	}
	// Upgrade if no combat actions
	if _, found := jar[0]["ONCOMBAT"]; !found {
		jar[0]["ONCOMBAT"] = []byte(`
		  [%A] lash[/es] out at [%d] hitting [%d.them] with random blows.
		: [%A] punch[/es] [%d] winding [%d.them].
		: [%A] punch[/es] [%d], landing a solid blow.
		: [%A] kick[/s] [%d], causing [%d.them] to yell.
		: [%A] headbutt[/s] [%d], stunning [%d.them].
		: [%A] feign[/s] an attack, then swiftly jab[/s] [%d.them].
		: [%D] yell[s//s] as [%a] bite[/s] [%d.them].
		: [%D] stumble[s//s] allowing [%a] to land a heavy blow.
		: [%D] doge[s//s] the wrong way allowing [%a] to hit [%d.them].
		: [%D] dodge[s//s] [%a] opening [%d.themself][/rself/] to a bashing.
		: [%A] slam[/s] [%a.their][r/] body into [%d].
		: [%A] dig[/s] an elbow into [%d].
		: [%A] bring[/s] a knee up hitting [%d].
    `)
	}

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
	if _, ok := c.Any[core.Permissions]; ok {
		p.Any[core.Permissions] = c.Any[core.Permissions]
	}
	p.Int[core.Created] = c.Int[core.Created]
	p.Is |= core.Player
	p.Is &^= core.NPC
	p.As[core.StatusSeq] = string(term.Status(c.height, c.width))
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
	c.Int[core.Created] = time.Now().UnixNano()
	c.Int[core.HealthAfter] = (1 * time.Minute).Nanoseconds()
	c.Int[core.HealthRestore] = 2
	c.Int[core.HealthCurrent] = 30
	c.Int[core.HealthMaximum] = 30
	c.Int[core.Armour] = 10
	c.Int[core.DamageFixed] = 2
	c.Int[core.DamageRandom] = 2
	c.Any[core.OnCombat] = []string{
		"[%A] lash[/es] out at [%d] hitting [%d.them] with random blows.",
		"[%A] punch[/es] [%d] winding [%d.them].",
		"[%A] punch[/es] [%d], landing a solid blow.",
		"[%A] kick[/s] [%d], causing [%d.them] to yell.",
		"[%A] headbutt[/s] [%d], stunning [%d.them].",
		"[%A] feign[/s] an attack, then swiftly jab[/s] [%d.them].",
		"[%D] yell[s//s] as [%a] bite[/s] [%d.them].",
		"[%D] stumble[s//s] allowing [%a] to land a heavy blow.",
		"[%D] doge[s//s] the wrong way allowing [%a] to hit [%d.them].",
		"[%D] dodge[s//s] [%a] opening [%d.themself][/rself/] to a bashing.",
		"[%A] slam[/s] [%a.their][r/] body into [%d].",
		"[%A] dig[/s] an elbow into [%d].",
		"[%A] bring[/s] a knee up hitting [%d].",
	}
	c.As[core.StatusSeq] = string(term.Status(c.height, c.width))
}
