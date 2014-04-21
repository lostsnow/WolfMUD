// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package player

import (
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"code.wolfmud.org/WolfMUD.git/utils/text"

	"bytes"
	"fmt"
	"strconv"
)

// DuplicateLoginError is raised when a player is already in the PlayerList and
// tries to get added a second time.
type DuplicateLoginError struct {
	player *Player
}

func (e DuplicateLoginError) Error() string {
	return fmt.Sprintf("Player already logged in: %s", e.player.account)
}

// playerList type records the current players on the server. Nothing is
// exported as access should be through the package level PlayerList.
type playerList struct {
	players map[string]*Player
	mutex   chan bool
}

// PlayerList exports the playerList type
var PlayerList playerList

// init makes the mutex channel for locking and sets up the players map.
func init() {
	PlayerList.mutex = make(chan bool, 1)
	PlayerList.mutex <- true

	PlayerList.players = make(map[string]*Player)
}

// lock acquires a lock on the package level playerList instance.
func (l *playerList) lock() {
	<-l.mutex
}

// unlock releases a lock on the package level playerList instance.
func (l *playerList) unlock() {
	l.mutex <- true
}

// Add adds a player to the playerList. If the player is already in the
// playerList DuplicateLoginError will be returned.
func (l *playerList) Add(player *Player) error {
	l.lock()
	defer l.unlock()
	if _, found := l.players[player.account]; !found {
		l.players[player.account] = player
		return nil
	}
	return &DuplicateLoginError{player}
}

// Remove removes a player from the playerList. If the player is not in the
// playerList no error is raised and the method completes normally.
func (l *playerList) Remove(player *Player) {
	l.lock()
	defer l.unlock()
	delete(l.players, player.account)
}

// Length returns the number of players in the playerList
func (l *playerList) Length() int {
	l.lock()
	defer l.unlock()
	return len(l.players)
}

// Broadcast implements the broadcaster interface and sends a message to all
// players currently on the server. The omit parameter specifies things not to
// send the message to. For example if we had a scream command we might send a
// message to everyone else:
//
//	You hear someone scream.
//
// The message you would see might be:
//
//	You scream!
//
// However you would not want the 'You hear someone scream.' message sent to
// yourself.
//
// Note: We are sending directly to players which is OK and needs no locking or
// synchronisation here apart from the playerList itself.
func (l *playerList) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	l.lock()
	defer l.unlock()

	msg := text.Colorize(fmt.Sprintf("\n"+format, any...))

	for _, p := range l.nonLockingList(omit...) {
		p.Respond(msg)
	}
}

// nonLockingList returns the current players on the server with possible
// omissions. This method is non-locking so that other locking methods can
// call it. If this method also locked we would end up deadlocking. The omit
// parameter specifies any things to be omitted from the returned list - handy
// when a player wants to know who else is on the server and not including
// themselves for example.
func (l *playerList) nonLockingList(omit ...thing.Interface) (list []*Player) {

OMIT:
	for _, player := range l.players {
		for i, o := range omit {
			if player.IsAlso(o) {
				omit = append(omit[0:i], omit[i+1:]...)
				continue OMIT
			}
		}
		list = append(list, player)
	}

	return
}

// Process implements the command.Interface to handle playerList specific
// commands.
func (l *playerList) Process(cmd *command.Command) (handled bool) {

	switch cmd.Verb {
	case "WHO":
		handled = l.who(cmd)
	}

	return
}

// who implements the 'WHO' command. This lists all the players currently on
// the server.
func (l *playerList) who(cmd *command.Command) (handled bool) {

	b := new(bytes.Buffer)

	if l.Length() < 2 {
		b.WriteString("You are all alone in this world.")
	} else {

		cmd.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s concentrate.", cmd.Issuer.Name())

		for _, p := range PlayerList.nonLockingList(cmd.Issuer) {
			b.WriteString("  ")
			b.WriteString(p.Name())
			b.WriteString("\n")
		}
		b.WriteString("\nTOTAL PLAYERS: ")
		b.WriteString(strconv.Itoa(l.Length()))

	}
	cmd.Respond(b.String())

	return true
}
