// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"sort"
)

// CrowdSize represents the minimum number of players considered to be a crowd.
// FIXME(diddymus): This needs to be configurable.
const CrowdSize = 11

// Commands maps command strings to the implementing methods.
var commands = map[string]func(*state){
	"":          func(*state) {},
	"QUIT":      (*state).Quit,
	"L":         (*state).Look,
	"LOOK":      (*state).Look,
	"N":         (*state).Move,
	"NORTH":     (*state).Move,
	"NE":        (*state).Move,
	"NORTHEAST": (*state).Move,
	"E":         (*state).Move,
	"EAST":      (*state).Move,
	"SE":        (*state).Move,
	"SOUTHEAST": (*state).Move,
	"S":         (*state).Move,
	"SOUTH":     (*state).Move,
	"SW":        (*state).Move,
	"SOUTHWEST": (*state).Move,
	"W":         (*state).Move,
	"WEST":      (*state).Move,
	"NW":        (*state).Move,
	"U":         (*state).Move,
	"UP":        (*state).Move,
	"D":         (*state).Move,
	"DOWN":      (*state).Move,
	"NORTHWEST": (*state).Move,
	"EXAM":      (*state).Examine,
	"EXAMINE":   (*state).Examine,
	"INV":       (*state).Inventory,
	"INVENTORY": (*state).Inventory,
	"DROP":      (*state).Drop,
	"GET":       (*state).Get,
	"TAKE":      (*state).Take,
	"PUT":       (*state).Put,
	"READ":      (*state).Read,
	"OPEN":      (*state).Open,
	"CLOSE":     (*state).Close,
	"COMMANDS":  nil, // Will be populated by init to avoid initialisation cycle

	// Admin and debugging commands
	"#DUMP":     (*state).Dump,
	"#TELEPORT": (*state).Teleport,
	"#GOTO":     (*state).Teleport,

	// Scripting only commands
	"$POOF": (*state).Poof,
}

// cmdNames is a precomputed, sorted list of available player and admin
// commands.Scripting commands with a '$' prefix are not included.
var cmdNames = func() (names []string) {
	for name := range commands {
		if name != "" && name[0] != '$' {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}()

// init registers the COMMANDS command. COMMANDS can't be in commands when
// cmdNames is initialised as it will create an initialisation cycle.
func init() {
	commands["COMMANDS"] = (*state).Commands
}

func (s *state) Quit() {
	delete(World[s.actor.As[Where]].Who, s.actor.As[UID])
	s.Msg(s.actor, "You leave this world behind.\n\nBye bye!\n")
}

func (s *state) Look() {
	where := World[s.actor.As[Where]]
	auid := s.actor.As[UID]

	switch {
	case where == nil:
		s.Msg(s.actor, "[The Void]\n",
			"You are in a dark void. Around you nothing.",
			"No stars, no light, no heat and no sound.")
	case where.Is&Dark == Dark:
		s.Msg(s.actor, "It's too dark to see anything!")
	default:
		s.Msg(s.actor, "[", where.As[Name], "]")
		s.Msg(s.actor, where.As[Description], "\n")
		mark := s.buf[auid].Len()
		if len(where.Who) < CrowdSize {
			for _, who := range where.Who.Sort() {
				if who.As[UID] == auid {
					continue
				}
				s.Msg(s.actor, "You see ", who.As[Name], " here.")
			}
			for _, item := range where.In.Sort() {
				if item.Is&Narrative == Narrative || item.As[UID] == auid {
					continue
				}
				s.Msg(s.actor, "You see ", item.As[Name], " here.")
			}
			if mark != s.buf[auid].Len() {
				s.Msg(s.actor)
				mark = s.buf[auid].Len()
			}
		} else {
			s.Msg(s.actor, "It's too crowded here to see anything.\n")
			mark = s.buf[auid].Len()
		}
		for dir := North; dir <= Down; dir++ {
			if where.As[dir] != "" {
				if s.buf[auid].Len() == mark {
					s.Msg(s.actor, "You see exits: ", DirToName[dir])
				} else {
					s.MsgAppend(s.actor, ", ", DirToName[dir])
				}
			}
		}
		if mark == s.buf[auid].Len() {
			s.Msg(s.actor, "You see no obvious exits.")
		}
	}

	// Only notify observers if actually looking and not $POOF or entering a
	// location when moving.
	if (s.cmd == "L" || s.cmd == "LOOK") && len(where.Who) < CrowdSize {
		s.Msg(where, s.actor.As[Name], " starts looking around.")
	}
}

func (s *state) Move() {

	dir := NameToDir[s.cmd]
	where := World[s.actor.As[Where]]

	if where.As[dir] == "" {
		s.Msg(s.actor, "You can't go ", DirToName[dir], ".")
		return
	}

	// Try and find first blocker for direction we want to go
	var blocker *Thing
	for _, item := range where.In {
		if item.As[Blocker] == "" {
			continue
		}
		blocking := NameToDir[item.As[Blocker]]
		// If on 'other side' need opposite direction blocked
		if item.As[Where] != s.actor.As[Where] {
			blocking = blocking.ReverseDir()
		}
		if blocking == dir && item.Is&Open != Open {
			blocker = item
			break
		}
	}

	switch {
	case blocker != nil:
		s.Msg(s.actor, "You can't go ", DirToName[dir], ". ",
			blocker.As[Name], " is blocking your way.")
	case World[where.As[dir]] == nil:
		s.Msg(s.actor, "Oops! You can't actually go ", DirToName[dir], ".")
	default:
		delete(World[s.actor.As[Where]].Who, s.actor.As[UID])
		if len(World[s.actor.As[Where]].Who) < CrowdSize {
			s.MsgAppend(World[s.actor.As[Where]], s.actor.As[Name],
				" leaves ", DirToName[dir], ".")
		}
		s.actor.As[Where] = where.As[dir]
		if len(World[s.actor.As[Where]].Who) < CrowdSize {
			s.MsgAppend(World[s.actor.As[Where]], s.actor.As[Name], " enters.")
		}
		World[s.actor.As[Where]].Who[s.actor.As[UID]] = s.actor
		s.Look()
	}
}

func (s *state) Examine() {

	if len(s.word) == 0 {
		s.Msg(s.actor, "You examine this and that, find nothing special.")
		return
	}

	uids := Match(s.word, s.actor, World[s.actor.As[Where]])
	uid := uids[0]
	what := s.actor.In[uid]
	if what == nil {
		what = World[s.actor.As[Where]].In[uid]
	}
	if what == nil {
		what = World[s.actor.As[Where]].Who[uid]
	}

	switch {
	case len(s.word) == 0:
		s.Msg(s.actor, "You examine this and that, find nothing special.")
	case what == nil:
		s.Msg(s.actor, "You see no '", uid, "' to examine.")
	case len(uids) > 1:
		s.Msg(s.actor, "You can only examine one thing at a time.")
	case what.Is&Container != Container || len(what.In) == 0:
		s.Msg(s.actor, "You examine ", what.As[Name], ".\n", what.As[Description])
		// If a blocker, e.g. a door, is it open or closed?
		switch {
		case what.As[Blocker] == "":
		case what.Is&Open == Open:
			s.MsgAppend(s.actor, " It is open.")
		default:
			s.MsgAppend(s.actor, " It is closed.")
		}
	case len(what.In) == 1:
		s.Msg(s.actor, "You examine ", what.As[Name], ".\n", what.As[Description])
		for _, item := range what.In {
			s.MsgAppend(s.actor, " It contains ", item.As[Name], ".")
		}
	default:
		s.Msg(s.actor, "You examine ", what.As[Name], ".\n", what.As[Description])
		s.MsgAppend(s.actor, " It contains: ")
		for _, item := range what.In.Sort() {
			s.Msg(s.actor, "  ", item.As[Name])
		}
	}
}

func (s *state) Inventory() {
	switch {
	case len(s.actor.In) == 0:
		s.Msg(s.actor, "You are not carrying anything.")
	default:
		s.Msg(s.actor, "You are carrying:")
		for _, what := range s.actor.In.Sort() {
			s.Msg(s.actor, "  ", what.As[Name])
		}
	}
}

func (s *state) Drop() {

	if len(s.word) == 0 {
		s.Msg(s.actor, "You go to drop... something?")
		return
	}

	for _, uid := range Match(s.word, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, "You do not have any '", uid, "' to drop.")
		case what.As[VetoDrop] != "":
			s.Msg(s.actor, what.As[VetoDrop])
		default:
			delete(s.actor.In, what.As[UID])
			World[s.actor.As[Where]].In[what.As[UID]] = what
			delete(what.As, DynamicQualifier)
			s.Msg(s.actor, "You drop ", what.As[Name], ".")
		}
	}
}

func (s *state) Get() {

	if len(s.word) == 0 {
		s.Msg(s.actor, "You go to get... something?")
		return
	}

	for _, uid := range Match(s.word, World[s.actor.As[Where]]) {
		what := World[s.actor.As[Where]].In[uid]
		if what == nil {
			what = World[s.actor.As[Where]].Who[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, "You see no '", uid, "' to get.")
		case what.As[VetoGet] != "":
			s.Msg(s.actor, what.As[VetoGet])
		case what.Is&Narrative == Narrative:
			s.Msg(s.actor, "You cannot take ", what.As[Name], ".")
		case what.Is&(NPC|Player) != 0:
			s.Msg(s.actor, what.As[Name], " does not want to be taken!")
		default:
			delete(World[s.actor.As[Where]].In, what.As[UID])
			s.actor.In[what.As[UID]] = what
			what.As[DynamicQualifier] = "MY"
			s.Msg(s.actor, "You get ", what.As[Name], ".")
		}
	}
}

func (s *state) Take() {

	if len(s.word) == 0 {
		s.Msg(s.actor, "You go to take something from something else...")
		return
	}

	uids, words := LimitedMatch(s.word, s.actor, World[s.actor.As[Where]])
	uid := uids[0]
	where := s.actor.In[uid]
	if where == nil {
		where = World[s.actor.As[Where]].In[uid]
	}

	switch {
	case where == nil:
		s.Msg(s.actor, "You see no '", uid, "' to take anything from.")
	case len(uids) > 1:
		s.Msg(s.actor, "You can only take things from one container at a time.")
	case where.Is&Container != Container:
		s.Msg(s.actor, where.As[Name], " is not something you can take anything from.")
	case len(words) == 0:
		s.Msg(s.actor, "You go to take something from ", where.As[Name], ".")
	case where.As[VetoTakeOut] != "":
		s.Msg(s.actor, where.As[VetoTakeOut])
	}
	if s.buf[s.actor.As[UID]] != nil {
		return
	}

	for _, uid := range Match(words, where) {
		what := where.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, where.As[Name], " does not seem to contain '", uid, "'.")
		case what.As[VetoTake] != "":
			s.Msg(s.actor, what.As[VetoTake])
		case where.Is&NPC == NPC:
			s.Msg(s.actor, "You can't take ", what.As[Name], " from ", where.As[Name], ".")
		default:
			delete(where.In, what.As[UID])
			s.actor.In[what.As[UID]] = what
			what.As[DynamicQualifier] = "MY"
			s.Msg(s.actor, "You take ", what.As[Name], " out of ", where.As[Name], ".")
		}
	}
}

func (s *state) Put() {

	if len(s.word) == 0 {
		s.Msg(s.actor, "You go to put something into something else...")
		return
	}

	uids, words := LimitedMatch(s.word, s.actor, World[s.actor.As[Where]])
	uid := uids[0]
	where := s.actor.In[uid]
	if where == nil {
		where = World[s.actor.As[Where]].In[uid]
	}

	switch {
	case where == nil:
		s.Msg(s.actor, "You see no '", uid, "' to put anything into.")
	case len(uids) > 1:
		s.Msg(s.actor, "You can only put things into one container at a time.")
	case where.Is&Container != Container:
		s.Msg(s.actor, where.As[Name], " is not something you can put anything into.")
	case len(words) == 0:
		s.Msg(s.actor, "You go to put something into ", where.As[Name], ".")
	case where.As[VetoPutIn] != "":
		s.Msg(s.actor, where.As[VetoPutIn])
	case where.Is&(NPC|Player) != 0:
		s.Msg(s.actor, "Taxidermist are we?")
	}
	if s.buf[s.actor.As[UID]] != nil {
		return
	}

	for _, uid := range Match(words, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, "You have no '", uid, "' to put into ", where.As[Name], ".")
		case what.As[VetoPut] != "":
			s.Msg(s.actor, what.As[VetoPut])
		case uid == where.As[UID]:
			s.Msg(s.actor, "It might be interesting to put ", what.As[Name],
				" inside itself, but probably paradoxical as well.")
		default:
			delete(s.actor.In, what.As[UID])
			where.In[what.As[UID]] = what
			delete(what.As, DynamicQualifier)
			s.Msg(s.actor, "You put ", what.As[Name], " into ", where.As[Name], ".")
		}
	}
}

func (s *state) Dump() {
	if len(s.word) == 0 {
		s.Msg(s.actor, "What did you want to dump?")
		return
	}
	var uids []string
	if s.word[0] == "@" {
		uids = []string{s.actor.As[Where]}
	} else {
		uids = Match(s.word, s.actor, World[s.actor.As[Where]])
	}
	for _, uid := range uids {
		what := s.actor.In[uid]
		if what == nil {
			what = World[s.actor.As[Where]].In[uid]
		}
		if what == nil {
			what = World[s.actor.As[Where]].Who[uid]
		}
		if what == nil {
			what = World[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, "You see no '", uid, "' to dump.")
		default:
			s.Msg(s.actor, "DUMP: ", uid, "\n")
			what.Dump(s.buf[s.actor.As[UID]], 80)
		}
	}
}

func (s *state) Read() {
	if len(s.word) == 0 {
		s.Msg(s.actor, "You go to read something...")
		return
	}
	for _, uid := range Match(s.word, World[s.actor.As[Where]], s.actor) {
		what := World[s.actor.As[Where]].In[uid]
		if what == nil {
			what = World[s.actor.As[Where]].Who[uid]
		}
		if what == nil {
			what = s.actor.In[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, "You see no '", uid, "' here to read.")
		case what.As[Writing] == "":
			s.Msg(s.actor, "There is nothing on ", what.As[Name], " to read.")
		default:
			s.Msg(s.actor, "You read ", what.As[Name], ". ", what.As[Writing])
		}
	}
}

func (s *state) Open() {
	if len(s.word) == 0 {
		s.Msg(s.actor, "You go to open something...")
		return
	}
	for _, uid := range Match(s.word, World[s.actor.As[Where]]) {
		what := World[s.actor.As[Where]].In[uid]
		if what == nil {
			what = World[s.actor.As[Where]].Who[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, "You see no '", uid, "' to open.")
		case what.As[Blocker] == "":
			s.Msg(s.actor, what.As[Name], " is not something you can open.")
		case what.Is&Open == Open:
			s.Msg(s.actor, what.As[Name], " is already open.")
		default:
			what.Is |= Open
			s.Msg(s.actor, "You open ", what.As[Name], ".")
		}
	}
}

func (s *state) Close() {
	if len(s.word) == 0 {
		s.Msg(s.actor, "You go to close something...")
		return
	}
	for _, uid := range Match(s.word, World[s.actor.As[Where]]) {
		what := World[s.actor.As[Where]].In[uid]
		if what == nil {
			what = World[s.actor.As[Where]].Who[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, "You see no '", uid, "' to close.")
		case what.As[Blocker] == "":
			s.Msg(s.actor, what.As[Name], " is not something you can close.")
		case what.Is&Open == 0:
			s.Msg(s.actor, what.As[Name], " is already closed.")
		default:
			what.Is &^= Open
			s.Msg(s.actor, "You close ", what.As[Name], ".")
		}
	}
}

func (s *state) Commands() {
	cols := 7
	split := (len(cmdNames) / cols) + 1
	pad := "               "
	s.Msg(s.actor, "Commands currently available:\n\n")
	for x := 0; x < split; x++ {
		for y := x; y < len(cmdNames); y += split {
			if y >= len(cmdNames) {
				continue
			}
			s.MsgAppend(s.actor, ("  " + cmdNames[y] + pad)[:12])
		}
		s.Msg(s.actor)
	}
}

func (s *state) Teleport() {
	if len(s.word) == 0 {
		s.Msg(s.actor, "Where do you want to go?")
		return
	}
	where := World[s.word[0]]
	switch {
	case where == nil:
		s.Msg(s.actor, "You don't know where '", s.word[0], "' is.")
	default:
		delete(World[s.actor.As[Where]].In, s.actor.As[UID])
		s.actor.As[Where] = s.word[0]
		World[s.actor.As[Where]].In[s.actor.As[UID]] = s.actor
		s.Msg(s.actor, "There is a loud 'Spang!'...\n")
		s.Look()
	}
}

func (s *state) Poof() {
	if len(World[s.actor.As[Where]].Who) < CrowdSize {
		s.Msg(World[s.actor.As[Where]], "There is a cloud of smoke from which ",
			s.actor.As[Name], " emerges coughing and spluttering.")
	}
	s.Look()
}
