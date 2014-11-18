// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package player defines an actual human player in the game.
package entities

import (
	"code.wolfmud.org/WolfMUD.git/entities/location"
	"code.wolfmud.org/WolfMUD.git/entities/mobile"
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"code.wolfmud.org/WolfMUD.git/utils/config"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"code.wolfmud.org/WolfMUD.git/utils/sender"

	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	saltLength = 10 // Length of salt to generate/use for passwords
)

// Player is the implementation of a player. Most of the functionality comes
// from the Mobile type and methods to implement the parser Interface. Apart
// from the parser interface methods Player only contains Player specific code.
type Player struct {
	mobile.Mobile
	sender   sender.Interface
	account  string
	password string
	salt     string
	created  time.Time
	quitting bool
}

// Register zero value instance of Player with the loader.
func init() {
	recordjar.RegisterUnmarshaler("player", &Player{})
}

// Unmarshal a recordjar record into a player
//
// NOTE: Password hash may be split over multiple lines in the data file which
// when read will be concatenated together with spaces - which need removing.
func (p *Player) Unmarshal(d recordjar.Decoder) {
	p.account = d.String("account")
	p.password = d.String("password")
	p.password = strings.Replace(d.String("password"), " ", "", -1)
	p.salt = d.String("salt")
	p.created = d.Time("created")

	p.Mobile.Unmarshal(d)
}

// Start starts a Player off in the world. The player is put into the world at
// a random starting location and the location is described to them.
func (p *Player) Start(s sender.Interface) {
	p.quitting = false
	p.sender = s
	p.add(location.GetStart())
}

// IsQuitting returns true if the player psrser is trying to quit otherwise
// false. It implements part of the parser.Interface.
func (p *Player) IsQuitting() bool {
	return p.quitting
}

// add places a player in the world safely and announces their arrival.  We
// manually build and parse the 'LOOK' command to avoid deadlocking - adding
// the player locks the location as does a normal p.Parse('LOOK'). We could add
// the player and then parse but that would require obtaining the lock twice.
func (p *Player) add(l location.Interface) {
	l.Lock()
	defer l.Unlock()

	l.Add(p)
	PlayerList.Add(p)

	cmd := command.New(p, "LOOK")
	p.Process(cmd)

	if !l.Crowded() {
		cmd.Broadcast([]thing.Interface{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.Name())
	}

	cmd.Flush()
}

// dropInventory drops everything the player is carrying.
func (p *Player) dropInventory(cmd *command.Command) {
	for _, o := range p.Inventory.List() {
		if c, ok := o.(command.Interface); ok {
			if aliases := o.Aliases(); len(aliases) > 0 {
				cmd.New("DROP " + o.Aliases()[0])
				c.Process(cmd)
			}
		}
	}
}

// Parse takes a string and begins the delegation to potential processors. To
// avoid deadlocks, inconsistencies, race conditions and other unmentionables
// we lock the location of the player. However there is a race condition
// between getting the player's location and locking it - they may have moved
// in-between. We therefore get and lock their current location then check it's
// still their current location. If it is not we unlock and try again.
//
// If a command effects more than one location we have to release the current
// lock on the location and relock the locations in Unique Id order before
// trying again. Always locking in a consistent order greatly helps in avoiding
// deadlocks.
//
// MOST of the time we are only interested in a few things: The current player,
// it's location, items at the location, mobiles at the location. We can
// therefore avoid complex fine grained locking on each individual Thing and
// just lock on the whole location. This does mean if there are a LOT of things
// happening in one specific location we will not have as much parallelism as we
// would like.
//
// TODO: If there many clients trying to connect at once - say 250+ simultaneous
// clients connecting - then the starting location becomes a bit of a bottle
// neck (at 1,000+ simultaneous clients connecting is a pain - but once
// connected things smooth out and become playable again). Adding more starting
// locations help to spread the bottle neck. Note that this is just an issue
// with the initial connection and multiple clients all trying to grab the start
// location lock!
func (p *Player) Parse(input string) {

	// If no input respond with nothing so the prompt is redisplayed
	if input == "" {
		p.Respond("")
		return
	}

	cmd := command.New(p, input)
	cmd.AddLock(p.Locate())
	cmd.LocksModified()

	// Another funky looking for loop :)
	for p.parseStage2(cmd) {
	}
}

// parseStage2 is called by Parse to take advantage of defer unwinding. By
// splitting the parsing we can easily obtain the locks we want and defer the
// unlocking. This makes both Parse and parseStage2 very simple.
func (p *Player) parseStage2(cmd *command.Command) (retry bool) {
	for _, l := range cmd.Locks {
		l.Lock()
		defer l.Unlock()
	}

	// If player moved before we locked we need to retry
	if !cmd.CanLock(p.Locate()) {
		return true
	}

	handled := p.Process(cmd)
	retry = cmd.LocksModified()

	if !handled && !retry {
		cmd.Respond("[RED]Eh?")
	}

	if !retry {
		cmd.Flush()
	}

	return
}

// NOTE: We should never have a nil sender as it's deallocated only after the
// player is extracted from the world.
func (p *Player) Respond(format string, any ...interface{}) {
	p.sender.Send(format, any...)
}

// Broadcast implements the broadcaster interface and broadcasts to the
// player's current location.
func (p *Player) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	p.Locate().Broadcast(omit, format, any...)
}

// Process implements the command.Interface to handle player specific commands.
// It also delegates to mobile.Process if the player can't handle the command
// which also does most of the delegating to get commands processed. . As a
// last resort we see if PlayerList can handle the command. PlayerList can't be
// handled by Mobile with everything else as it causes a cyclic import and goes
// BOOM!
func (p *Player) Process(cmd *command.Command) (handled bool) {

	switch cmd.Verb {
	case "CPUSTOP":
		handled = p.cpustop(cmd)
	case "CPUSTART":
		handled = p.cpustart(cmd)
	case "MEMPROF":
		handled = p.memprof(cmd)
	case "QUIT":
		handled = p.quit(cmd)
	case "SAY", "'":
		handled = p.say(cmd)
	case "SNEEZE":
		handled = p.sneeze(cmd)
	}

	if !handled {
		handled = p.Mobile.Process(cmd)
	}

	if !handled {
		handled = PlayerList.Process(cmd)
	}

	return
}

// cpustart implement the 'CPUSTART' command and starts CPU profiling.
//
// TODO: Remove - for debugging only
func (p *Player) cpustart(cmd *command.Command) (handled bool) {
	f, err := os.Create("cpuprof")
	if err != nil {
		cmd.Respond("CPU Profile Not Started: %s", err)
		return false
	}
	pprof.StartCPUProfile(f)

	cmd.Respond("CPU profile started")
	return true
}

// cpustop implements the 'CPUSTOP' command, stops CPU profiling and writes the
// profile to cpuprofile in the bin directory.
//
// TODO: Remove - for debugging only
func (p *Player) cpustop(cmd *command.Command) (handled bool) {
	pprof.StopCPUProfile()
	cmd.Respond("CPU profile stopped")
	return true
}

// memprof implements the 'MEMPROF' command and writes out a memprofile.
//
// NOTE: Need to change the value of MemProfileRate in server.go
// TODO: Remove - for debugging only
func (p *Player) memprof(cmd *command.Command) (handled bool) {
	f, err := os.Create("memprof")
	if err != nil {
		cmd.Respond("Memory Profile Not Dumped: %s", err)
		return false
	}
	pprof.WriteHeapProfile(f)
	f.Close()

	cmd.Respond("Memory profile dumped")
	return true
}

// quit implements the 'QUIT' command.
//
// quit extracts a player from the world cleanly. If the player's location is
// not crowded it also announces their departure - in a crowded location their
// departure will go unnoticed.
func (p *Player) quit(cmd *command.Command) (handled bool) {
	log.Printf("%s is quiting", p.Name())
	p.quitting = true
	p.dropInventory(cmd)

	l := p.Locate()

	if !l.Crowded() {
		cmd.Broadcast([]thing.Interface{p}, "%s gives a strangled cry of 'Bye Bye', and then slowly fades away and is gone.", p.Name())
	}

	cmd.Flush()

	l.Remove(p)
	PlayerList.Remove(p)

	return true
}

// sneeze implements the 'SNEEZE' command.
//
// TODO: Remove - for debugging responders and broadcasters
func (p *Player) sneeze(cmd *command.Command) (handled bool) {
	cmd.Respond("You sneeze. Aaahhhccchhhooo!")
	cmd.Broadcast([]thing.Interface{p}, "You see %s sneeze.", cmd.Issuer.Name())
	PlayerList.Broadcast(p.Locate().List(), "You hear a loud sneeze.")
	return true
}

// say implements the 'SAY' command, alias "'". Say sends a message to all
// other responders at the current location.
func (p *Player) say(cmd *command.Command) (handled bool) {

	if len(cmd.Nouns) == 0 {
		cmd.Respond("[RED]Was there anything in particular you wanted to say?")
	} else {
		message := strings.Join(cmd.Input[1:], ` `)

		cmd.Broadcast([]thing.Interface{cmd.Issuer}, "[CYAN]%s says: %s", cmd.Issuer.Name(), message)
		cmd.Respond("[YELLOW]You say: %s", message)
	}

	return true
}

// Errors that can be returned by Load.
var (
	ErrBadCredentials = errors.New("invalid credentals")
	ErrBadPlayerFile  = errors.New("invalid player file")
	ErrDuplicateLogin = errors.New("player already logged in")
)

// Load loads a player .wrj data file. SetAccount should be called before
// calling Load. SetAccount will generate the account hash, this with .wrj
// appended to it is the name of the data file to be loaded.
//
// If the data file does not exist BadCredentials error will be returned.
// If there is no account hash - SetAccount has not ben called maybe? - a
// BadCredentials error will be returned.
//
// If the data file is found but cannot be opened a ErrBadPlayerFile error is
// returned.
//
// If the data file is opened but the password is incorrect a ErrBadCredentials
// error is returned.
//
// If the data file cannot be unmarshaled a ErrBadPlayerFile error is returned.
//
// TODO: We should be able to use the generic RecordJar.LoadFile at some point?
func (p *Player) Load() (err error) {

	// Clear credentials if we exit with an error
	defer func() {
		if err != nil {
			p.account = ""
			p.password = ""
			p.salt = ""
		}
	}()

	// No account hash?
	if p.account == "" {
		return ErrBadCredentials
	}

	// Can we open the player's file?
	f, err := os.Open(config.DataDir + "players/" + p.account + ".wrj")
	if err != nil {
		if os.IsNotExist(err) {
			return ErrBadCredentials
		} else {
			log.Printf("Error opening player file: %s", err)
			return ErrBadPlayerFile
		}
	}
	defer f.Close()

	rj, err := recordjar.Read(f)

	if err != nil {
		log.Printf("Error reading player jar: %s", err)
		return ErrBadPlayerFile
	}

	if len(rj) == 0 {
		log.Printf("Error loading player file: no records in jar")
		return ErrBadPlayerFile
	}

	d := recordjar.Decoder(rj[0])
	p.Unmarshal(d)

	return nil
}

// Save writes out the current receiver to a *.wrj file. The filename is the
// account hash with .wrj appended to it which can be set by calling SetAccount.
//
// If an error occurs writing the file it will be returned.
//
// TODO: At the moment we manually build the jar using an Encoder. This is
// temporary and needs to be replaced with a proper Marshaler.
func (p *Player) Save() error {

	// If creation date empty populate it now
	if p.created.IsZero() {
		p.created = time.Now()
	}

	e := recordjar.Encoder{}
	e.Keyword("type", "player")
	e.Keyword("ref", "player")
	e.String("account", p.account)
	e.String("password", p.password)
	e.String("salt", p.salt)
	e.String("name", p.Name())
	e.Keyword("gender", p.ItMaleFemale())
	e.Time("created", p.created)

	j := recordjar.Jar{recordjar.Record(e)}

	fileFlags := os.O_CREATE | os.O_EXCL | os.O_WRONLY
	filename := config.DataDir + "players/" + p.account + ".wrj"

	f, err := os.OpenFile(filename+".tmp", fileFlags, 0660)
	if err != nil {
		return err
	}

	recordjar.Write(f, j)

	if err = f.Close(); err != nil {
		return err
	}

	if err = os.Rename(filename+".tmp", filename); err != nil {
		return err
	}

	return nil
}

// Account returns the player's account hash. This hash maps directly to the
// name of the player's data file.
func (p *Player) Account() string {
	return p.account
}

// RuneCountError can be used to return the expected and actual number of runes
// used. This is useful for things like account ID and password validation that
// require a minimum number of characters.
type RuneCountError struct {
	Label     string // Descriptive string like 'account' or 'password'.
	Length    int    // Actual length
	MinLength int    // Minimum length
}

// Error provides a descriptive error message for RuneCountErrors.
func (e RuneCountError) Error() string {
	return fmt.Sprintf("%s too short. Wanted at least %d runes. Only got %d runes.", e.Label, e.MinLength, e.Length)
}

// SetAccount takes a plain account string and calculates it's hash as a string.
// The hash is synonymous with the name of a player's data file with .wrj
// appended to it. The hash can be retrieved by calling Account().
//
// Using a weak, fast hash here is not an issue. The hash is used to convert an
// arbitrarily long, user supplied account name into a hexadecimal, fixed length
// string of the hash. The only consequence of using a weak, fast hash is an
// indication that an account already exists when in fact it's a collision.
//
// Even though the account name may contain characters other than letters and
// digits using a hash results in a string only containing only characters 0-9
// and a-f which are safe to use as the player's data file's name.
func (p *Player) SetAccount(account string) error {
	count := utf8.RuneCountInString(account)
	min := config.AccountIdMin

	if count < min {
		return &RuneCountError{"account", count, min}
	}

	h := md5.Sum([]byte(account))
	p.account = hex.EncodeToString(h[:])

	return nil
}

// SetPassword takes a plain string password and generates it's hash and salt
// as strings. The salt is randomly generated for each password by selecting 10
// random printable ASCII characters in the range 0x21 to 0x7E or '!' to '~'.
//
// The returned hash and salt can be passed to PasswordValid to subsequently
// validate passwords.
func (p *Player) SetPassword(password string) error {
	count := utf8.RuneCountInString(password)
	min := config.AccountIdMin

	if count < min {
		return &RuneCountError{"password", count, min}
	}

	l := saltLength + len(password)
	sp := make([]byte, l, l)

	for i := 0; i < saltLength; i++ {
		sp[i] = byte('!' + rand.Intn('~'-'!'))
	}

	copy(sp[saltLength:], password)

	h := sha512.Sum512(sp)

	p.password = base64.URLEncoding.EncodeToString(h[:])
	p.salt = string(sp[:saltLength])

	return nil
}

// PasswordValid validates that a given password is correct for a given hash and
// salt. The passed hash and salt should be generated by SetPassword.
func (p *Player) PasswordValid(password string) bool {
	l := saltLength + len(password)
	sp := make([]byte, l, l)

	copy(sp[:saltLength], p.salt)
	copy(sp[saltLength:], password)

	h := sha512.Sum512(sp)

	return base64.URLEncoding.EncodeToString(h[:]) == p.password
}

// SetName overrides the otherwise promoted Thing.SetName to add validation
// checking. If validation fails a non-nil error will be returned.
func (p *Player) SetName(name string) error {
	for _, r := range []rune(name) {
		if !unicode.IsLetter(r) {
			return errors.New("name contains invlid characters")
		}
	}

	p.Thing.SetName(name)

	return nil
}
