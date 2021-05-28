// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

// Commands maps command strings to the implementing methods.
var commands = map[string]func(*state){
	"":          func(*state) {},
	"QUIT":      func(s *state) { s.Msg("Bye bye!") },
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
	"#DUMP":     (*state).Dump,
}

func (s *state) Quit() {
	s.Msg("Bye bye!")
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
		s.Msg("[", where.Name, "]\n", where.Description, "\n\n")
		mark := s.buff.Len()
		for _, item := range where.In {
			if item.Is&Narrative == Narrative {
				continue
			}
			s.Msg("You see ", item.Name, " here.\n")
		}
		if s.buff.Len() > mark {
			s.Msg("\n")
		}
		s.Msg("You see exits:")
		for dir, text := range DirToName {
			if where.As[dir] != "" {
				s.Msg(" ", text)
			}
		}
	}
}

func (s *state) Move() {

	dir := NameToDir[s.cmd]
	where := World[s.actor.As[Where]]

	switch {
	case where.As[dir] == "":
		s.Msg("You can't go ", DirToName[dir], ".")
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
		s.Msg("You examine ", what.Name, ".\n", what.Description)
	default:
		s.Msg("You examine ", what.Name, ".\n", what.Description)
		s.Msg(" It contains: ", what.In[0].Name)
		for _, item := range what.In[1:] {
			s.Msg(", ", item.Name)
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
			s.Msg("\n  ", what.Name)
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
		s.Msg("You drop ", what.Name, ".")
	}
}

func (s *state) Get() {

	what, where, idx := Find(s.word[0], World[s.actor.As[Where]])

	switch {
	case s.word[0] == "":
		s.Msg("You go to get... something?")
	case what == nil:
		s.Msg("You see no '", s.word[0], "' to get.")
	case what.Is&Narrative == Narrative:
		s.Msg("You cannot take ", what.Name, ".")
	case what.Is&NPC == NPC:
		s.Msg(what.Name, " does not want to be taken!")
	default:
		copy(where.In[idx:], where.In[idx+1:])
		where.In[len(where.In)-1] = nil
		where.In = where.In[:len(where.In)-1]

		s.actor.In = append(s.actor.In, what)
		s.Msg("You get ", what.Name, ".")
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
		s.Msg(where.Name, " does not seem to contain '", s.word[0], "'.")
	case where.Is&(Container|NPC) == NPC :
		s.Msg("You can't take ", what.Name, " from ", where.Name, ".")
	case where.Is&Container == 0:
		s.Msg("You can't take ", what.Name, " out of ", where.Name, ".")
	case what == nil:
	default:
		copy(where.In[idx:], where.In[idx+1:])
		where.In[len(where.In)-1] = nil
		where.In = where.In[:len(where.In)-1]

		s.actor.In = append(s.actor.In, what)
		s.Msg("You take ", what.Name, " from ", where.Name, ".")
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
		s.Msg("You go to put ", what.Name, " into something...")
	case where == nil && what == nil:
		s.Msg("You see no '", s.word[1], "' to put anything in.")
	case where == nil:
		s.Msg("You see no '", s.word[1], "' to put ", what.Name, " in.")
	case where.Is&NPC == NPC:
		s.Msg("Taxidermist are we?")
	case where.Is&Container == 0:
		s.Msg("You can't put ", what.Name, " into ", where.Name, ".")
	case what == nil:
		s.Msg("You have no '", s.word[0], "' to put into ", where.Name, ".")
	default:
		copy(s.actor.In[idx:], s.actor.In[idx+1:])
		s.actor.In[len(s.actor.In)-1] = nil
		s.actor.In = s.actor.In[:len(s.actor.In)-1]

		where.In = append(where.In, what)
		s.Msg("You put ", what.Name, " into ", where.Name, ".")
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
