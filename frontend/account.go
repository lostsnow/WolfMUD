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
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// explainAccountDisplay displays the requirements for new account IDs. It is
// separated from newAccountDisplay so that if there is a problem we can ask
// for the new account ID again without having to have the explanation.
func (f *frontend) explainAccountDisplay() {
	l := strconv.Itoa(config.Login.AccountLength)
	f.buf.WriteJoin("Your account ID can be anything you can remember: an email address, a book title, a film title, a quote. You can use upper and lower case characters, numbers and symbols. The only restriction is it has to be at least ", l, " characters long.\n\nThis is NOT your character's name it is for your account ID for logging in only.\n\n")
	f.newAccountDisplay()
}

// newAccountDisplay asks the player for a new account ID
func (f *frontend) newAccountDisplay() {
	f.buf.WriteString("Enter text to use for your account ID or just press enter to cancel:")
	f.nextFunc = f.newAccountProcess
}

// newAccountProcess takes the current input and stores it in the current state as
// an account ID hash. At this point it is not validated yet, just stored.
func (f *frontend) newAccountProcess() {
	switch l := len(f.input); {
	case l == 0:
		f.buf.WriteString("Account creation cancelled.\n\n")
		f.accountDisplay()
	case l < config.Login.AccountLength:
		l := strconv.Itoa(config.Login.AccountLength)
		f.buf.WriteJoin("Account ID is too short. Needs to be ", l, " characters or longer.\n\n")
		f.newAccountDisplay()
	default:
		hash := md5.Sum(f.input)
		f.account = hex.EncodeToString(hash[:])
		f.newPasswordDisplay()
	}
}

// newPasswordDisplay asks for a password to associate with the account ID.
func (f *frontend) newPasswordDisplay() {
	f.buf.WriteString("Enter a password to use for your account ID or just press enter to cancel:")
	f.nextFunc = f.newPasswordProcess
}

// newPasswordProcess takes the current input and stores it in the current
// state as a hash. The hash is calculated with a random salt that is also
// stored in the current state.
func (f *frontend) newPasswordProcess() {
	switch l := len(f.input); {
	case l == 0:
		f.buf.WriteString("Account creation cancelled.\n\n")
		f.accountDisplay()
	case l < config.Login.PasswordLength:
		l := strconv.Itoa(config.Login.PasswordLength)
		f.buf.WriteJoin("Password is too short. Needs to be ", l, " characters or longer.\n\n")
		f.newPasswordDisplay()
	default:
		salt := salt(config.Login.SaltLength)
		hash := sha512.Sum512(append(salt, f.input...))
		f.stash["SALT"] = salt
		f.stash["HASH"] = hash[:]
		f.confirmPasswordDisplay()
	}
}

// confirmPasswordDisplay asks for the password to be typed again for confirmation.
func (f *frontend) confirmPasswordDisplay() {
	f.buf.WriteString("Enter your password again to confirm or just press enter to cancel:")
	f.nextFunc = f.confirmPasswordProcess
}

// confirmPasswordProcess verifies that the confirmation password matches the
// new password already stored in the current state.
func (f *frontend) confirmPasswordProcess() {
	switch l := len(f.input); {
	case l == 0:
		f.buf.WriteString("Account creation cancelled.\n\n")
		f.accountDisplay()
	default:
		hash := sha512.Sum512(append(f.stash["SALT"], f.input...))
		if !bytes.Equal(hash[:], f.stash["HASH"]) {
			f.buf.WriteJoin("Passwords do not match, please try again.\n\n")
			f.newPasswordDisplay()
			return
		}
		f.nameDisplay()
	}
}

// nameDisplay asks for a player name.
func (f *frontend) nameDisplay() {
	f.buf.WriteString("Enter a name for your character or just press enter to cancel:")
	f.nextFunc = f.nameProcess
}

// nameProcess verifies the player name and stores it in the current state.
func (f *frontend) nameProcess() {
	switch l := len(f.input); {
	case l == 0:
		f.buf.WriteString("Account creation cancelled.\n\n")
		f.accountDisplay()
	case l < 3:
		f.buf.WriteJoin("The name '", string(f.input), "' is too short.\n\n")
		f.nameDisplay()
	default:
		f.stash["NAME"] = f.input
		f.genderDisplay()
	}
}

// genderDisplay asks for the gender of the player.
func (f *frontend) genderDisplay() {
	f.buf.WriteJoin("Would you like ", string(f.stash["NAME"]), " to be male or female?")
	f.nextFunc = f.genderProcess
}

// genderProcess verifies the gender and stores it in the current state.
func (f *frontend) genderProcess() {
	switch string(bytes.ToUpper(f.input)) {
	case "M", "MALE":
		f.stash["GENDER"] = []byte("MALE")
		f.write()
	case "F", "FEMALE":
		f.stash["GENDER"] = []byte("FEMALE")
		f.write()
	default:
		f.buf.WriteString("Please specify male or female.\n\n")
		f.genderDisplay()
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
func (f *frontend) write() {
	jar := recordjar.Jar{}

	// Create account record
	hash := base64.URLEncoding.EncodeToString(f.stash["HASH"])
	rec := recordjar.Record{}
	rec["ACCOUNT"] = recordjar.Encode.String(f.account)
	rec["PASSWORD"] = recordjar.Encode.String(hash)
	rec["SALT"] = recordjar.Encode.Bytes(f.stash["SALT"])
	rec["CREATED"] = recordjar.Encode.DateTime(time.Now())
	jar = append(jar, rec)

	// Create player record
	rec = recordjar.Record{}
	rec["NAME"] = recordjar.Encode.Bytes(f.stash["NAME"])
	rec["ALIAS"] = recordjar.Encode.Keyword(string(f.stash["NAME"]))
	rec["GENDER"] = recordjar.Encode.Keyword(string(f.stash["GENDER"]))
	rec["REF"] = recordjar.Encode.Keyword("R1")
	rec["INVENTORY"] = recordjar.Encode.KeywordList([]string{})
	rec["DESCRIPTION"] = recordjar.Encode.String("This is an adventurer, just like you!")
	jar = append(jar, rec)

	temp := filepath.Join(config.Server.DataDir, "players", f.account+".tmp")
	real := filepath.Join(config.Server.DataDir, "players", f.account+".wrj")

	// Lock accounts to prevent races while manipulating files
	accounts.Lock()
	defer accounts.Unlock()

	// Check if account ID is already registered
	if _, err := os.Stat(real); !os.IsNotExist(err) {
		f.buf.WriteString("The account ID you used is not available.\n\n")
		f.accountDisplay()
		return
	}

	// Write record jar to temporary file
	wrj, err := os.Create(temp)
	if err != nil {
		log.Printf("Error creating account: %s, %s", temp, err)
		return
	}
	jar.Write(wrj, "DESCRIPTION")
	wrj.Close()

	// If all went well rename the temporary file to the real file. The rename
	// should be an atomic operation but is dependant on the underlying file
	// system and operating system being used.
	if err := os.Rename(temp, real); err != nil {
		log.Printf("Error renaming account: %s, %d, %s", temp, real, err)
		return
	}
	log.Printf("New account created: %s", real)

	accounts.inuse[f.account] = struct{}{}

	// Assemble player
	f.player = attr.NewThing()
	f.player.(*attr.Thing).Unmarshal(1, rec)
	f.player.Add(attr.NewLocate(nil))
	f.player.Add(attr.NewPlayer(f.output))

	// Greet new player
	f.buf.WriteString("Welcome ")
	f.buf.WriteString(attr.FindName(f.player).Name("Someone"))
	f.buf.WriteString("!\n")

	f.menuDisplay()
}
