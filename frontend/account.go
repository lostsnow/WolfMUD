// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text"
)

// account embeds a frontend instance adding fields and methods specific to
// account and player creation.
type account struct {
	*frontend
	account  string
	password [sha512.Size]byte
	salt     []byte
	name     string
	gender   string
}

// NewAccount returns an account with the specified frontend embedded. The
// returned account can be used for processing the creation of new accounts and
// players.
func NewAccount(f *frontend) (a *account) {
	a = &account{frontend: f}
	a.explainAccountDisplay()
	return
}

// verifyName is used to test that a players name only uses the letters A-Z,a-z.
var verifyName = regexp.MustCompile(`^[a-zA-Z]+$`)

// explainAccountDisplay displays the requirements for new account IDs. It is
// separated from newAccountDisplay so that if there is a problem we can ask
// for the new account ID again without having to have the explanation as well.
func (a *account) explainAccountDisplay() {
	l := strconv.Itoa(config.Login.AccountLength)
	a.buf.Send("Your account ID can be anything you can remember: an email address, a book title, a film title, a quote. You can use upper and lower case characters, numbers and symbols. The only restriction is it has to be at least ", l, " characters long.\n\nThis is NOT your character's name it is for your account ID for logging in only.\n")
	a.newAccountDisplay()
}

// newAccountDisplay asks the player for a new account ID
func (a *account) newAccountDisplay() {
	a.buf.Send("Enter text to use for your new account ID or just press enter to cancel:")
	a.nextFunc = a.newAccountProcess
}

// newAccountProcess takes the current input and stores it as an account ID
// hash. We don't know if it's already taken yet, we are just storing it.
func (a *account) newAccountProcess() {
	switch l := len(a.input); {
	case l == 0:
		a.buf.Send(text.Info, "Account creation cancelled.\n", text.Reset)
		NewLogin(a.frontend)
	case l < config.Login.AccountLength:
		l := strconv.Itoa(config.Login.AccountLength)
		a.buf.Send(text.Bad, "Account ID is too short. Needs to be ", l, " characters or longer.\n", text.Reset)
		a.newAccountDisplay()
	default:
		hash := md5.Sum(a.input)
		a.account = hex.EncodeToString(hash[:])
		a.newPasswordDisplay()
	}
}

// newPasswordDisplay asks for a password to associate with the account ID.
func (a *account) newPasswordDisplay() {
	a.buf.Send("Enter a password to use for your account ID or just press enter to cancel:")
	a.nextFunc = a.newPasswordProcess
}

// newPasswordProcess takes the current input and stores it in the current
// state as a hash. The hash is calculated with a random salt that is also
// stored in the current state.
func (a *account) newPasswordProcess() {
	switch l := len(a.input); {
	case l == 0:
		a.buf.Send(text.Info, "Account creation cancelled.\n", text.Reset)
		NewLogin(a.frontend)
	case l < config.Login.PasswordLength:
		l := strconv.Itoa(config.Login.PasswordLength)
		a.buf.Send(text.Bad, "Password is too short. Needs to be ", l, " characters or longer.\n", text.Reset)
		a.newPasswordDisplay()
	default:

		// Calculate hash for salt+password then zero buffer used and input
		a.salt = salt(config.Login.SaltLength)
		si := make([]byte, len(a.salt)+len(a.input))
		copy(si[0:], a.salt)
		copy(si[len(a.salt):], a.input)
		a.password = sha512.Sum512(si)
		Zero(si)
		Zero(a.input)

		a.confirmPasswordDisplay()
	}
}

// confirmPasswordDisplay asks for the password to be typed again for confirmation.
func (a *account) confirmPasswordDisplay() {
	a.buf.Send("Enter your password again to confirm or just press enter to cancel:")
	a.nextFunc = a.confirmPasswordProcess
}

// confirmPasswordProcess verifies that the confirmation password matches the
// new password already stored in the current state.
func (a *account) confirmPasswordProcess() {
	switch l := len(a.input); {
	case l == 0:
		a.buf.Send(text.Info, "Account creation cancelled.\n", text.Reset)
		NewLogin(a.frontend)
	default:

		// Calculate hash for salt+password then zero buffer used and input
		si := make([]byte, len(a.salt)+len(a.input))
		copy(si[0:], a.salt)
		copy(si[len(a.salt):], a.input)
		hash := sha512.Sum512(si)
		Zero(si)
		Zero(a.input)

		if hash != a.password {
			a.buf.Send(text.Bad, "Passwords do not match, please try again.\n", text.Reset)
			a.newPasswordDisplay()
			return
		}
		a.nameDisplay()
	}
}

// nameDisplay asks for a player name.
func (a *account) nameDisplay() {
	a.buf.Send("Enter a name for your character or just press enter to cancel:")
	a.nextFunc = a.nameProcess
}

// nameProcess verifies the player name and stores it in the current state.
func (a *account) nameProcess() {
	switch l := len(a.input); {
	case l == 0:
		a.buf.Send(text.Info, "Account creation cancelled.\n", text.Reset)
		NewLogin(a.frontend)
	case l < 3:
		a.buf.Send(text.Bad, "The name '", string(a.input), "' is too short.\n", text.Reset)
		a.nameDisplay()
	case verifyName.Find(a.input) == nil:
		a.buf.Send(text.Bad, "A character's name must only contain the upper or lower cased letters 'a' through 'z'. Using other letters, such as those with accents, will make it harder for other players to interact with you if they cannot type your character's name. \n", text.Reset)
		a.nameDisplay()
	default:
		a.name = string(a.input)
		a.genderDisplay()
	}
}

// genderDisplay asks for the gender of the player.
func (a *account) genderDisplay() {
	a.buf.Send("Would you like ", a.name, " to be male or female?")
	a.nextFunc = a.genderProcess
}

// genderProcess verifies the gender and stores it in the current state.
func (a *account) genderProcess() {
	switch string(bytes.ToUpper(a.input)) {
	case "":
		return
	case "M", "MALE":
		a.gender = "MALE"
		a.write()
	case "F", "FEMALE":
		a.gender = "FEMALE"
		a.write()
	default:
		a.buf.Send(text.Bad, "Please specify male or female.\n", text.Reset)
		a.genderDisplay()
	}
}

// salt returns a []byte containing the given length of random ASCII
// characters. ASCII characters used will be in the range printable range "!"
// (0x21) to "~" (0x7E) - 88 characters total.
func salt(l int) []byte {
	salt := make([]byte, l, l)
	extra := make([]byte, 1, 1)

	rand.Read(salt)

	for x := 0; x < l; x++ {
		// Scale byte value to 0x1F < byte < 0x80
		salt[x] = salt[x]&0x7F | 0x20

		// If byte ASCII Space (0x20) or DEL (0x7F) replace and try again
		if salt[x] == 0x20 || salt[x] == 0x7F {
			rand.Read(extra)
			salt[x] = extra[0]
			x--
		}
	}

	return salt
}

// write creates the player data file and writes it out to the filesystem. The
// player data file is written to DataDir/players where DataDir is set via the
// config.Server.DataDir configuration setting.
//
// BUG(diddymus): write should return any errors so that they can be checked.
// At the moment if there is an error writing the player file the player is
// still let in and their details not saved for next time they log in.
func (a *account) write() {
	jar := recordjar.Jar{}

	// Create account record
	hash := base64.URLEncoding.EncodeToString(a.password[:])
	rec := recordjar.Record{
		"ACCOUNT":  encode.String(a.account),
		"PASSWORD": encode.String(hash),
		"SALT":     encode.Bytes(a.salt),
		"CREATED":  encode.DateTime(time.Now()),
	}
	jar = append(jar, rec)

	// Create player record
	rec = recordjar.Record{
		"NAME":        encode.String(a.name),
		"ALIAS":       encode.Keyword(a.name),
		"GENDER":      encode.Keyword(a.gender),
		"REF":         encode.Keyword("R1"),
		"INVENTORY":   encode.KeywordList([]string{}),
		"DESCRIPTION": encode.String("This is an adventurer, just like you!"),
	}
	jar = append(jar, rec)

	temp := filepath.Join(config.Server.DataDir, "players", a.account+".tmp")
	real := filepath.Join(config.Server.DataDir, "players", a.account+".wrj")

	// Lock accounts to prevent races while manipulating files
	accounts.Lock()
	defer accounts.Unlock()

	// Check if account ID is already registered
	if _, err := os.Stat(real); !os.IsNotExist(err) {
		a.buf.Send(text.Bad, "The account ID you used is not available.\n", text.Reset)
		NewLogin(a.frontend)
		return
	}

	// Write record jar to temporary file
	wrj, err := os.Create(temp)
	if err != nil {
		log.Printf("Error creating account: %s, %s", temp, err)
		return
	}

	if config.Server.SetPermissions {
		err = wrj.Chmod(0660)
		if err != nil {
			wrj.Close()
			log.Printf("Error changing account permissions: %s, %s", temp, err)
			return
		}
	}

	jar.Write(wrj, "DESCRIPTION")
	wrj.Close()

	// If all went well rename the temporary file to the real file. The rename
	// should be an atomic operation but is dependant on the underlying file
	// system and operating system being used.
	if err := os.Rename(temp, real); err != nil {
		log.Printf("Error renaming account: %s, %s, %s", temp, real, err)
		return
	}
	log.Printf("New account created: %s", real)

	a.frontend.account = a.account
	accounts.inuse[a.frontend.account] = struct{}{}

	// Assemble player
	a.player = attr.NewThing()
	a.player.(*attr.Thing).Unmarshal(1, rec)
	a.player.Add(attr.NewPlayer(a.output))

	// Greet new player
	a.buf.Send(text.Good, "Welcome ", attr.FindName(a.player).Name("Someone"), "!", text.Reset)

	NewMenu(a.frontend)
}
