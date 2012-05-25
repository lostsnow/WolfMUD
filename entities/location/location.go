package location

import (
	"fmt"
	//"log"
	"strings"
	"wolfmud.org/entities/inventory"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/responder"
)

type direction uint8

const (
	N, NORTH direction = iota, iota
	NE, NORTHEAST
	E, EAST
	SE, SOUTHEAST
	S, SOUTH
	SW, SOUTHWEST
	W, WEST
	NW, NORTHWEST
	U, UP
	D, DOWN
)

var directionNames = [...]string{
	N:  "North",
	NE: "Northeast",
	E:  "East",
	SE: "Southeast",
	S:  "South",
	SW: "Southwest",
	W:  "West",
	NW: "Northwest",
	U:  "Up",
	D:  "Down",
}

type Interface interface {
	thing.Interface
	command.Interface
	inventory.Interface
	LinkExit(d direction, to Interface)
	Look(cmd *command.Command) (handled bool)
	Broadcast(ommit []thing.Interface, format string, any ...interface{})
	Lock()
	Unlock()
}

type Locateable interface {
	Relocate(Interface)
	Locate() Interface
}

type Location struct {
	*thing.Thing
	*inventory.Inventory
	exits [len(directionNames)]Interface
	lock  chan bool
}

func New(name string, aliases []string, description string) *Location {
	return &Location{
		Thing:     thing.New(name, aliases, description),
		Inventory: &inventory.Inventory{},
		lock:      make(chan bool, 1),
	}
}

func (l *Location) Lock() {
	//log.Printf("Locking %s", l.Name())
	l.lock <- true
	//log.Printf("Locked %s", l.Name())
}

func (l *Location) Unlock() {
	//log.Printf("Unlocking %s", l.Name())
	<-l.lock
	//log.Printf("Unlocked %s", l.Name())
}

func (l *Location) LinkExit(d direction, to Interface) {
	l.exits[d] = to
}

func (l *Location) Add(thing thing.Interface) {
	if t, ok := thing.(Locateable); ok {
		t.Relocate(l)
	}
	l.Inventory.Add(thing)
}

func (l *Location) Remove(thing thing.Interface) {
	if t, ok := thing.(Locateable); ok {
		t.Relocate(nil)
	}
	l.Inventory.Remove(thing)
}

func (l *Location) Broadcast(ommit []thing.Interface, format string, any ...interface{}) {
	msg := fmt.Sprintf("\n"+format, any...)

	for _, v := range l.Inventory.List(ommit...) {
		if resp, ok := v.(responder.Interface); ok {
			resp.Respond(msg)
		}
	}
}

func (l *Location) Process(cmd *command.Command) (handled bool) {
	switch cmd.Verb {
	case "LOOK", "L":
		handled = l.Look(cmd)
	case "NORTH", "N":
		handled = l.move(cmd, NORTH)
	case "NORTHEAST", "NE":
		handled = l.move(cmd, NORTHEAST)
	case "EAST", "E":
		handled = l.move(cmd, EAST)
	case "SOUTHEAST", "SE":
		handled = l.move(cmd, SOUTHEAST)
	case "SOUTH", "S":
		handled = l.move(cmd, SOUTH)
	case "SOUTHWEST", "SW":
		handled = l.move(cmd, SOUTHWEST)
	case "WEST", "W":
		handled = l.move(cmd, WEST)
	case "NORTHWEST", "NW":
		handled = l.move(cmd, NORTHWEST)
	case "UP":
		handled = l.move(cmd, UP)
	case "DOWN":
		handled = l.move(cmd, DOWN)
	}

	if handled == false {
		//handled = l.thing.Process(cmd)
	}

	if handled == false {
		//handled = l.Inventory.delegate(cmd)
	}

	return handled
}

func (l *Location) Look(cmd *command.Command) (handled bool) {

	thingsHere := []string{}
	for _, o := range l.Inventory.List(cmd.Issuer) {
		thingsHere = append(thingsHere, "You can see "+o.Name()+" here.")
	}

	validExits := []string{}
	for d, l := range l.exits {
		if l != nil {
			validExits = append(validExits, directionNames[d])
		}
	}

	cmd.Respond("%s\n%s\n%s\n\nYou can see exits: %s", l.Name(), l.Description(), strings.Join(thingsHere, "\n"), strings.Join(validExits, ", "))

	return true
}

func (l *Location) Move(d direction) (to Interface) {
	return l.exits[d]
}

func (l *Location) move(cmd *command.Command, d direction) (handled bool) {
	if to := l.exits[d]; to != nil {
		if !cmd.IsLocked(to) {
			cmd.Relock = to
			return true
		}

		l.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s go %s.", cmd.Issuer.Name(), directionNames[d])

		l.Remove(cmd.Issuer)

		cmd.Respond("You go %s.", directionNames[d])
		to.Add(cmd.Issuer)
		to.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s walk in.", cmd.Issuer.Name())

		to.Look(cmd)
	} else {
		cmd.Respond("You can't go %s from here!", directionNames[d])
	}
	return true
}
