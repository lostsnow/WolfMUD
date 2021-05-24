// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

func (s *state) Look() {
	where := World[s.actor.As[Where]]
	switch {
	case where == nil:
		s.buff.WriteString("[The Void]\nYou are in a dark void. Around you nothing. No stars, no light, no heat and no sound.")
	case where.Is&Dark == Dark:
		s.buff.WriteString("It's too dark to see anything!")
	default:
		var things string
		for _, item := range where.In {
			if item.Is&Narrative == Narrative {
				continue
			}
			things += "\nYou see " + item.Name + " here."
		}
		if len(things) > 0 {
			things = "\n" + things
		}
		var exits string
		for dir, text := range DirToName {
			if where.As[dir] != "" {
				exits += text + " "
			}
		}
		if len(exits) > 0 {
			exits = "\n\nExits: " + exits
		}
		s.buff.WriteString("[" + where.Name + "]\n" +
			where.Description + things + exits)
	}
}

func (s *state) Move() {
	dir := NameToDir[s.cmd]
	where := World[s.actor.As[Where]]
	switch {
	case where.As[dir] == "":
		s.buff.WriteString("You can't go " + DirToName[dir] + "!")
	case World[where.As[dir]] == nil:
		s.buff.WriteString("Oops! You can't actually go " + DirToName[dir] + ".")
	default:
		s.actor.As[Where] = where.As[dir]
		s.Look()
	}
}

func (s *state) Examine() {
	var contains string
	for _, where := range []*Thing{s.actor, World[s.actor.As[Where]]} {
		for _, what := range where.In {
			if what.As[Alias] == s.word[0] {
				if len(what.In) > 0 {
					contains = " It contains: "
					for _, item := range what.In {
						contains += item.Name + ", "
					}
					contains = contains[:len(contains)-2] + "."
				}
				s.buff.WriteString("You examine " + what.Name + ".\n" + what.Description + contains)
				return
			}
		}
	}
	s.buff.WriteString("You see no '" + s.word[0] + "' to examine.")
}

func (s *state) Inv() {
	if len(s.actor.In) == 0 {
		s.buff.WriteString("You are not carrying anything.")
		return
	}

	items := "You are carrying:"
	for _, what := range s.actor.In {
		items += "\n  " + what.Name
	}
	s.buff.WriteString(items)
}

func (s *state) Drop() {
	if s.word[0] == "" {
		s.buff.WriteString("You go to drop... something?")
		return
	}

	for idx, what := range s.actor.In {
		if what.As[Alias] == s.word[0] {
			copy(s.actor.In[idx:], s.actor.In[idx+1:])
			s.actor.In[len(s.actor.In)-1] = nil
			s.actor.In = s.actor.In[:len(s.actor.In)-1]

			where := World[s.actor.As[Where]]
			where.In = append(where.In, what)
			s.buff.WriteString("You drop " + what.Name + ".")
			return
		}
	}
	s.buff.WriteString("You do not have any '" + s.word[0] + "' to drop.")
}

func (s *state) Get() {
	if s.word[0] == "" {
		s.buff.WriteString("You go to get... something?")
		return
	}

	where := World[s.actor.As[Where]]

	for idx, what := range where.In {
		if what.As[Alias] == s.word[0] {
			switch {
			case what.Is&Narrative == Narrative:
				s.buff.WriteString("You cannot take " + what.Name + ".")
			case what.Is&NPC == NPC:
				s.buff.WriteString(what.Name + " does not want to be taken!")
			default:
				copy(where.In[idx:], where.In[idx+1:])
				where.In[len(where.In)-1] = nil
				where.In = where.In[:len(where.In)-1]

				s.actor.In = append(s.actor.In, what)
				s.buff.WriteString("You get " + what.Name + ".")
			}
			return
		}
	}
	s.buff.WriteString("You see no '" + s.word[0] + "' to get.")
}

func (s *state) Take() {

	switch {
	case s.word[0] == "":
		s.buff.WriteString("You go to take something from something else...")
		return
	case s.word[1] == "":
		s.buff.WriteString("You go to take '" + s.word[0] + "' from something...")
		return
	}

	var where *Thing

	// Find container
findContainer:
	for _, inv := range []*Thing{s.actor, World[s.actor.As[Where]]} {
		for _, item := range inv.In {
			if item.As[Alias] == s.word[1] {
				where = item
				break findContainer
			}
		}
	}

	if where == nil {
		s.buff.WriteString("You see no '" + s.word[1] + "' to take anything from.")
		return
	}

	// Find item in container
	for idx, what := range where.In {
		if what.As[Alias] == s.word[0] {

			// Remove item from container
			copy(where.In[idx:], where.In[idx+1:])
			where.In[len(where.In)-1] = nil
			where.In = where.In[:len(where.In)-1]

			// Give to player
			s.actor.In = append(s.actor.In, what)

			s.buff.WriteString("You take " + what.Name + " from " + where.Name + ".")
			return
		}
	}

	s.buff.WriteString(where.Name + " does not seem to contain '" + s.word[0] + "'.")
}

func (s *state) Put() {

	switch {
	case s.word[0] == "":
		s.buff.WriteString("You go to put something into something else...")
		return
	case s.word[1] == "":
		s.buff.WriteString("You go to put '" + s.word[0] + "' into something...")
		return
	}

	var where *Thing

	// Find container
findContainer:
	for _, inv := range []*Thing{s.actor, World[s.actor.As[Where]]} {
		for _, item := range inv.In {
			if item.As[Alias] == s.word[1] {
				where = item
				break findContainer
			}
		}
	}

	if where == nil {
		s.buff.WriteString("You have no '" + s.word[1] + "' to put anything in.")
		return
	}

	// Find item (must be carried)
	for idx, what := range s.actor.In {
		if what.As[Alias] == s.word[0] {

			// Remove item from player
			copy(s.actor.In[idx:], s.actor.In[idx+1:])
			s.actor.In[len(s.actor.In)-1] = nil
			s.actor.In = s.actor.In[:len(s.actor.In)-1]

			// Put into container
			where.In = append(where.In, what)

			s.buff.WriteString("You put " + what.Name + " into " + where.Name + ".")
			return
		}
	}

	s.buff.WriteString("You do not have '" + s.word[0] + "' to put into "+where.Name+".")
}
