// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

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

	// Admin and debugging commands
	"#DUMP":     (*state).Dump,
	"#TELEPORT": (*state).Teleport,
	"#GOTO":     (*state).Teleport,
}

func (s *state) Quit() {
	s.Msg("You leave this world behind.\n\nBye bye!\n")
	s.prompt = s.prompt[:0]
}

func (s *state) Look() {
	where := World[s.actor.As[Where]]
	switch {
	case where == nil:
		s.Msg("[The Void]\n",
			"You are in a dark void. Around you nothing.",
			"No stars, no light, no heat and no sound.")
	case where.Is&Dark == Dark:
		s.Msg("It's too dark to see anything!")
	default:
		s.Msg("[", where.As[Name], "]")
		s.Msg(where.As[Description], "\n")
		mark := s.buff.Len()
		for _, item := range where.SortedIn() {
			if item.Is&Narrative == Narrative {
				continue
			}
			s.Msg("You see ", item.As[Name], " here.")
		}
		if mark != s.buff.Len() {
			s.Msg()
			mark = s.buff.Len()
		}
		for dir := North; dir <= Down; dir++ {
			if where.As[dir] != "" {
				if s.buff.Len() == mark {
					s.Msg("You see exits: ", DirToName[dir])
				} else {
					s.MsgAppend(", ", DirToName[dir])
				}
			}
		}
		if mark == s.buff.Len() {
			s.Msg("You see no obvious exits.")
		}
	}
}

func (s *state) Move() {

	dir := NameToDir[s.cmd]
	where := World[s.actor.As[Where]]

	if where.As[dir] == "" {
		s.Msg("You can't go ", DirToName[dir], ".")
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
		s.Msg("You can't go ", DirToName[dir], ". ",
			blocker.As[Name], " is blocking your way.")
	case World[where.As[dir]] == nil:
		s.Msg("Oops! You can't actually go ", DirToName[dir], ".")
	default:
		s.actor.As[Where] = where.As[dir]
		s.Look()
	}
}

func (s *state) Examine() {

	if len(s.word) == 0 {
		s.Msg("You examine this and that, find nothing special.")
		return
	}

	uids := Match(s.word, s.actor, World[s.actor.As[Where]])
	uid := uids[0]
	what := s.actor.In[uid]
	if what == nil {
		what = World[s.actor.As[Where]].In[uid]
	}

	switch {
	case len(s.word) == 0:
		s.Msg("You examine this and that, find nothing special.")
	case what == nil:
		s.Msg("You see no '", uid, "' to examine.")
	case len(uids) > 1:
		s.Msg("You can only examine one thing at a time.")
	case what.Is&Container != Container || len(what.In) == 0:
		s.Msg("You examine ", what.As[Name], ".\n", what.As[Description])
		// If a blocker, e.g. a door, is it open or closed?
		switch {
		case what.As[Blocker] == "":
		case what.Is&Open == Open:
			s.MsgAppend(" It is open.")
		default:
			s.MsgAppend(" It is closed.")
		}
	case len(what.In) == 1:
		s.Msg("You examine ", what.As[Name], ".\n", what.As[Description])
		for _, item := range what.In {
			s.MsgAppend(" It contains ", item.As[Name], ".")
		}
	default:
		s.Msg("You examine ", what.As[Name], ".\n", what.As[Description])
		s.MsgAppend(" It contains: ")
		for _, item := range what.SortedIn() {
			s.Msg("  ", item.As[Name])
		}
	}
}

func (s *state) Inventory() {
	switch {
	case len(s.actor.In) == 0:
		s.Msg("You are not carrying anything.")
	default:
		s.Msg("You are carrying:")
		for _, what := range s.actor.SortedIn() {
			s.Msg("  ", what.As[Name])
		}
	}
}

func (s *state) Drop() {

	if len(s.word) == 0 {
		s.Msg("You go to drop... something?")
		return
	}

	for _, uid := range Match(s.word, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg("You do not have any '", uid, "' to drop.")
		case what.As[VetoDrop] != "":
			s.Msg(what.As[VetoDrop])
		default:
			delete(s.actor.In, what.As[UID])
			World[s.actor.As[Where]].In[what.As[UID]] = what
			s.Msg("You drop ", what.As[Name], ".")
		}
	}
}

func (s *state) Get() {

	if len(s.word) == 0 {
		s.Msg("You go to get... something?")
		return
	}

	for _, uid := range Match(s.word, World[s.actor.As[Where]]) {
		what := World[s.actor.As[Where]].In[uid]
		switch {
		case what == nil:
			s.Msg("You see no '", uid, "' to get.")
		case what.As[VetoGet] != "":
			s.Msg(what.As[VetoGet])
		case what.Is&Narrative == Narrative:
			s.Msg("You cannot take ", what.As[Name], ".")
		case what.Is&NPC == NPC:
			s.Msg(what.As[Name], " does not want to be taken!")
		default:
			delete(World[s.actor.As[Where]].In, what.As[UID])
			s.actor.In[what.As[UID]] = what
			s.Msg("You get ", what.As[Name], ".")
		}
	}
}

func (s *state) Take() {

	if len(s.word) == 0 {
		s.Msg("You go to take something from something else...")
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
		s.Msg("You see no '", uid, "' to take anything from.")
	case len(uids) > 1:
		s.Msg("You can only take things from one container at a time.")
	case where.Is&Container != Container:
		s.Msg(where.As[Name], " is not something you can take anything from.")
	case len(words) == 0:
		s.Msg("You go to take something from ", where.As[Name], ".")
	case where.As[VetoTakeOut] != "":
		s.Msg(where.As[VetoTakeOut])
	}
	if s.buff.Len() > 0 {
		return
	}

	for _, uid := range Match(words, where) {
		what := where.In[uid]
		switch {
		case what == nil:
			s.Msg(where.As[Name], " does not seem to contain '", uid, "'.")
		case what.As[VetoTake] != "":
			s.Msg(what.As[VetoTake])
		case where.Is&NPC == NPC:
			s.Msg("You can't take ", what.As[Name], " from ", where.As[Name], ".")
		default:
			delete(where.In, what.As[UID])
			s.actor.In[what.As[UID]] = what
			s.Msg("You take ", what.As[Name], " out of ", where.As[Name], ".")
		}
	}
}

func (s *state) Put() {

	if len(s.word) == 0 {
		s.Msg("You go to put something into something else...")
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
		s.Msg("You see no '", uid, "' to put anything into.")
	case len(uids) > 1:
		s.Msg("You can only put things into one container at a time.")
	case where.Is&Container != Container:
		s.Msg(where.As[Name], " is not something you can put anything into.")
	case len(words) == 0:
		s.Msg("You go to put something into ", where.As[Name], ".")
	case where.As[VetoPutIn] != "":
		s.Msg(where.As[VetoPutIn])
	case where.Is&NPC == NPC:
		s.Msg("Taxidermist are we?")
	}
	if s.buff.Len() > 0 {
		return
	}

	for _, uid := range Match(words, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg("You have no '", uid, "' to put into ", where.As[Name], ".")
		case what.As[VetoPut] != "":
			s.Msg(what.As[VetoPut])
		case uid == where.As[UID]:
			s.Msg("It might be interesting to put ", what.As[Name],
				" inside itself, but probably paradoxical as well.")
		default:
			delete(s.actor.In, what.As[UID])
			where.In[what.As[UID]] = what
			s.Msg("You put ", what.As[Name], " into ", where.As[Name], ".")
		}
	}
}

func (s *state) Dump() {
	if len(s.word) == 0 {
		s.Msg("What did you want to dump?")
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
			what = World[uid]
		}
		switch {
		case what == nil:
			s.Msg("You see no '", uid, "' to dump.")
		default:
			what.Dump(s.buff, 80)
		}
	}
}

func (s *state) Read() {
	if len(s.word) == 0 {
		s.Msg("You go to read something...")
		return
	}
	for _, uid := range Match(s.word, World[s.actor.As[Where]], s.actor) {
		what := World[s.actor.As[Where]].In[uid]
		if what == nil {
			what = s.actor.In[uid]
		}
		switch {
		case what == nil:
			s.Msg("You see no '", uid, "' here to read.")
		case what.As[Writing] == "":
			s.Msg("There is nothing on ", what.As[Name], " to read.")
		default:
			s.Msg("You read ", what.As[Name], ". ", what.As[Writing])
		}
	}
}

func (s *state) Open() {
	if len(s.word) == 0 {
		s.Msg("You go to open something...")
		return
	}
	for _, uid := range Match(s.word, World[s.actor.As[Where]]) {
		what := World[s.actor.As[Where]].In[uid]
		switch {
		case what == nil:
			s.Msg("You see no '", uid, "' to open.")
		case what.As[Blocker] == "":
			s.Msg(what.As[Name], " is not something you can open.")
		case what.Is&Open == Open:
			s.Msg(what.As[Name], " is already open.")
		default:
			what.Is |= Open
			s.Msg("You open ", what.As[Name], ".")
		}
	}
}

func (s *state) Close() {
	if len(s.word) == 0 {
		s.Msg("You go to close something...")
		return
	}
	for _, uid := range Match(s.word, World[s.actor.As[Where]]) {
		what := World[s.actor.As[Where]].In[uid]
		switch {
		case what == nil:
			s.Msg("You see no '", uid, "' to close.")
		case what.As[Blocker] == "":
			s.Msg(what.As[Name], " is not something you can close.")
		case what.Is&Open == 0:
			s.Msg(what.As[Name], " is already closed.")
		default:
			what.Is &^= Open
			s.Msg("You close ", what.As[Name], ".")
		}
	}
}

func (s *state) Teleport() {
	where := World[s.word[0]]
	switch {
	case where == nil:
		s.Msg("You don't know where '", s.word[0], "' is.")
	default:
		s.actor.As[Where] = s.word[0]
		s.Msg("There is a loud 'Spang!'...\n")
		s.Look()
	}
}
