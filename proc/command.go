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
	"UP":        (*state).Move,
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
		s.Msg("[", where.As[Name], "]\n", where.As[Description], "\n\n")
		mark := s.buff.Len()
		for _, item := range where.In {
			if item.Is&Narrative == Narrative {
				continue
			}
			s.Msg("You see ", item.As[Name], " here.\n")
		}
		if s.buff.Len() > mark {
			s.Msg("\n")
			mark = s.buff.Len()
		}
		for dir := North; dir <= Down; dir++ {
			if where.As[dir] != "" {
				if s.buff.Len() == mark {
					s.Msg("You see exits:")
				}
				s.Msg(" ", DirToName[dir])
			}
		}
		if s.buff.Len() == mark {
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
			blocking = ReverseDir(blocking)
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

	what, _, _ := Find(s.word[0], s.actor, World[s.actor.As[Where]])

	switch {
	case s.word[0] == "":
		s.Msg("You go to examine... something?")
	case what == nil:
		s.Msg("You see no '", s.word[0], "' to examine.")
	case what.Is&Container != Container || len(what.In) == 0:
		s.Msg("You examine ", what.As[Name], ".\n", what.As[Description])
		// If a blocker, e.g. a door, is it open or closed?
		switch {
		case what.As[Blocker] == "":
		case what.Is&Open == Open:
			s.Msg(" It is open.")
		default:
			s.Msg(" It is closed.")
		}
	default:
		s.Msg("You examine ", what.As[Name], ".\n", what.As[Description])
		s.Msg(" It contains: ", what.In[0].As[Name])
		for _, item := range what.In[1:] {
			s.Msg(", ", item.As[Name])
		}
	}
}

func (s *state) Inventory() {
	switch {
	case len(s.actor.In) == 0:
		s.Msg("You are not carrying anything.")
	default:
		s.Msg("You are carrying:")
		for _, what := range s.actor.In {
			s.Msg("\n  ", what.As[Name])
		}
	}
}

func (s *state) Drop() {

	what, _, idx := Find(s.word[0], s.actor)

	switch {
	case s.word[0] == "":
		s.Msg("You go to drop... something?")
	case what == nil:
		s.Msg("You do not have any '", s.word[0], "' to drop.")
	default:
		copy(s.actor.In[idx:], s.actor.In[idx+1:])
		s.actor.In[len(s.actor.In)-1] = nil
		s.actor.In = s.actor.In[:len(s.actor.In)-1]

		where := World[s.actor.As[Where]]
		where.In = append(where.In, what)
		s.Msg("You drop ", what.As[Name], ".")
	}
}

func (s *state) Get() {

	what, where, idx := Find(s.word[0], World[s.actor.As[Where]])

	switch {
	case s.word[0] == "":
		s.Msg("You go to get... something?")
	case what == nil:
		s.Msg("You see no '", s.word[0], "' to get.")
	case what.Any[Veto+"GET"] != nil:
		s.Msg(what.Any[Veto+"GET"]...)
	case what.Is&Narrative == Narrative:
		s.Msg("You cannot take ", what.As[Name], ".")
	case what.Is&NPC == NPC:
		s.Msg(what.As[Name], " does not want to be taken!")
	default:
		copy(where.In[idx:], where.In[idx+1:])
		where.In[len(where.In)-1] = nil
		where.In = where.In[:len(where.In)-1]

		s.actor.In = append(s.actor.In, what)
		s.Msg("You get ", what.As[Name], ".")
	}
}

func (s *state) Take() {

	// Find container, then item in container
	where, _, _ := Find(s.word[1], s.actor, World[s.actor.As[Where]])
	what, _, idx := Find(s.word[0], where)

	switch {
	case s.word[0] == "":
		s.Msg("You go to take something from something else...")
	case s.word[1] == "":
		s.Msg("You go to take '", s.word[0], "' from something...")
	case where == nil:
		s.Msg("You see no '", s.word[1], "' to take anything from.")
	case what == nil:
		s.Msg(where.As[Name], " does not seem to contain '", s.word[0], "'.")
	case where.Is&(Container|NPC) == NPC:
		s.Msg("You can't take ", what.As[Name], " from ", where.As[Name], ".")
	case where.Is&Container == 0:
		s.Msg("You can't take ", what.As[Name], " out of ", where.As[Name], ".")
	case what == nil:
	default:
		copy(where.In[idx:], where.In[idx+1:])
		where.In[len(where.In)-1] = nil
		where.In = where.In[:len(where.In)-1]

		s.actor.In = append(s.actor.In, what)
		s.Msg("You take ", what.As[Name], " from ", where.As[Name], ".")
	}
}

func (s *state) Put() {

	// Find container, find item (must be carried)
	where, _, _ := Find(s.word[1], s.actor, World[s.actor.As[Where]])
	what, _, idx := Find(s.word[0], s.actor)

	switch {
	case s.word[0] == "":
		s.Msg("You go to put something into something else...")
	case s.word[1] == "" && what == nil:
		s.Msg("You go to put '", s.word[0], "' into something...")
	case s.word[1] == "":
		s.Msg("You go to put ", what.As[Name], " into something...")
	case where == nil && what == nil:
		s.Msg("You see no '", s.word[1], "' to put anything in.")
	case where == nil:
		s.Msg("You see no '", s.word[1], "' to put ", what.As[Name], " in.")
	case where.Is&NPC == NPC:
		s.Msg("Taxidermist are we?")
	case where.Is&Container == 0:
		s.Msg("You can't put ", what.As[Name], " into ", where.As[Name], ".")
	case what == nil:
		s.Msg("You have no '", s.word[0], "' to put into ", where.As[Name], ".")
	default:
		copy(s.actor.In[idx:], s.actor.In[idx+1:])
		s.actor.In[len(s.actor.In)-1] = nil
		s.actor.In = s.actor.In[:len(s.actor.In)-1]

		where.In = append(where.In, what)
		s.Msg("You put ", what.As[Name], " into ", where.As[Name], ".")
	}
}

func (s *state) Dump() {

	var what *Thing
	if s.word[0] == "@" {
		what = World[s.actor.As[Where]]
	} else {
		what, _, _ = Find(s.word[0], s.actor, World[s.actor.As[Where]])
	}

	switch {
	case s.word[0] == "":
		s.Msg("What did you want to dump?")
	case what == nil:
		s.Msg("You see no '", s.word[0], "' to dump.")
	default:
		what.Dump(s.buff, 80)
	}
}

func (s *state) Read() {
	what, _, _ := Find(s.word[0], World[s.actor.As[Where]], s.actor)

	switch {
	case s.word[0] == "":
		s.Msg("You go to read something...")
	case what == nil:
		s.Msg("You see no '", s.word[0], "' here to read.")
	case what.As[Writing] == "":
		s.Msg("There is nothing on ", what.As[Name], " to read.")
	default:
		s.Msg("You read ", what.As[Name], ". ", what.As[Writing])
	}
}

func (s *state) Open() {
	what, _, _ := Find(s.word[0], World[s.actor.As[Where]])

	switch {
	case s.word[0] == "":
		s.Msg("You go to open something...")
	case what == nil:
		s.Msg("You see no '", s.word[0], "' to open.")
	case what.As[Blocker] == "":
		s.Msg(what.As[Name], " is not something you can open.")
	case what.Is&Open == Open:
		s.Msg(what.As[Name], " is already open.")
	default:
		what.Is |= Open
		s.Msg("You open ", what.As[Name], ".")
	}
}

func (s *state) Close() {
	what, _, _ := Find(s.word[0], World[s.actor.As[Where]])

	switch {
	case s.word[0] == "":
		s.Msg("You go to close something...")
	case what == nil:
		s.Msg("You see no '", s.word[0], "' to close.")
	case what.As[Blocker] == "":
		s.Msg(what.As[Name], " is not something you can close.")
	case what.Is&Open == 0:
		s.Msg(what.As[Name], " is already closed.")
	default:
		what.Is &^= Open
		s.Msg("You close ", what.As[Name], ".")
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
