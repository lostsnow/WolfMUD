// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package player defines an actual human player in the game.
package player

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
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"strings"
	"time"
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
func (p *Player) Unmarshal(d recordjar.Decoder) {
	p.account = d.String("account")
	p.password = d.String("password")
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
	BadCredentials = errors.New("Invalid credentals")
	BadPlayerFile  = errors.New("Invalid player file")
	DuplicateLogin = errors.New("Player already logged in")
)

// Load loads a player .wrj data file. The passed account should be a hash
// returned by HashAccount. The name of the data file will be the account hash
// with .wrj appended to it.
//
// If an error is returned a nil *Player will always be returned.
//
// If the data file cannot be opened a BadCredentials error is returned - the
// account is incorrect if the file is not found.
//
// If the data file is opened but the password is incorrect a BadCredentials
// error is returned.
//
// If the data file cannot be unmarshaled a BadPlayerFile error is returned.
//
// NOTE: We are manually opening the player's file, reading it as a recordjar,
// peeking inside it, then unmarshaling it. This is so that we can abort at any
// point - player not found, incorrect password, corrupt player file - having
// done as little work as possible. In this way we are not unmarshaling players
// which may have a lot of dependant stuff (inventory) to unmarshal just to
// validate the login - someone could hit the server and tie up processing with
// invalid logins otherwise if the unmarshaling took a significant amount of
// time.
func Load(account string, password string) (*Player, error) {

	// Can we open the player's file to get the current salt and password hash?
	f, err := os.Open(config.DataDir + "players/" + account + ".wrj")
	if err != nil {
		return nil, BadCredentials
	}
	defer f.Close()

	rj, _ := recordjar.Read(f)

	d := recordjar.Decoder(rj[0])
	p := d.String("password")
	s := d.String("salt")

	// Password hash may be split over multiple lines in the data file which when
	// read will be concatenated together with spaces - which need removing.
	h := strings.Replace(p, " ", "", -1)

	if !PasswordValid(password, s, h) {
		return nil, BadCredentials
	}

	data := recordjar.UnmarshalJar(&rj)

	if data["PLAYER"] == nil {
		log.Printf("Error loading player: %#v", rj)
		return nil, BadPlayerFile
	}

	return data["PLAYER"].(*Player), nil
}

func Save(e recordjar.Encoder) error {

	d := recordjar.Decoder(e)

	account := d.String("account")

	fileFlags := os.O_CREATE | os.O_EXCL | os.O_WRONLY

	f, err := os.OpenFile(config.DataDir+"players/"+account+".wrj", fileFlags, 0660)

	if err != nil {
		return err
	}
	defer f.Close()

	rj := recordjar.RecordJar{}
	rj = append(rj, recordjar.Record(e))
	recordjar.Write(f, rj)

	return nil
}

// HashAccount takes a plain account string and returns it's hash as a string.
// The hash is synonymous with the name of a player's data file with .wrj
// appended to it.
//
// Using a weak, fast hash here is not an issue. The hash is used to convert an
// arbitrarily long, user supplied account name into a hexadecimal, fixed length
// string of the hash. The only consequence of using a weak, fast hash is an
// indication that an account already exists when in fact it's a collision.
//
// Even though the account name may contain characters other than letters and
// digits using a hash results in a string only containing only characters 0-9
// and a-f which are safe to use as the player's data file's name.
func HashAccount(account string) string {
	h := md5.Sum([]byte(account))
	return hex.EncodeToString(h[:])
}

// HashPassword takes a plain string password and returns it's hash and salt as
// strings. The salt is randomly generated for each password by selecting 10
// random printable ASCII characters in the range 0x21 to 0x7E or '!' to '~'.
//
// The returned hash and salt can be passed to PasswordValid to subsequently
// validate passwords.
func HashPassword(password string) (hash, salt string) {
	l := saltLength + len(password)
	sp := make([]byte, l, l)

	for i := 0; i < saltLength; i++ {
		sp[i] = byte('!' + rand.Intn('~'-'!'))
	}

	copy(sp[saltLength:], password)

	h := sha512.Sum512(sp)

	hash = base64.URLEncoding.EncodeToString(h[:])
	salt = string(sp[:saltLength])

	return
}

// PasswordValid validates that a given password is correct for a given hash
// and salt. The passed hash and salt should be generated by HashPassword.
func PasswordValid(password, salt, hash string) bool {
	l := saltLength + len(password)
	sp := make([]byte, l, l)

	copy(sp[:saltLength], salt)
	copy(sp[saltLength:], password)

	h := sha512.Sum512(sp)

	return base64.URLEncoding.EncodeToString(h[:]) == hash
}
