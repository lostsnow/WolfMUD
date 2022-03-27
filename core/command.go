// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"bytes"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text"
)

// commandHandlers maps command strings to the implementing methods. It is
// typically initialised by calling RegisterCommandHandlers.
var commandHandlers map[string]func(*state)

// commandNames is a precomputed, sorted list of registered player and admin
// commands. It is typically initialised by calling RegisterCommandHandlers.
var commandNames []string

// eventCommands map an eventKey to its associated scripting command handler.
// It is typically initialised by calling RegisterCommandHandlers.
var eventCommands map[eventKey]string

// RegisterCommandHandlers initialises the commandHandlers, commandNames and
// eventCommands. It needs to be called before any player, admin or scripting
// commands are used. RegisterCommandHandlers should be called while core.BWL
// is held.
func RegisterCommandHandlers() {

	commandHandlers = map[string]func(*state){
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
		"COMMANDS":  (*state).Commands,
		"\"":        (*state).Say,
		"SAY":       (*state).Say,
		"SNEEZE":    (*state).Sneeze,
		"SHOUT":     (*state).Shout,
		"JUNK":      (*state).Junk,
		"REMOVE":    (*state).Remove,
		"HOLD":      (*state).Hold,
		"WEAR":      (*state).Wear,
		"WIELD":     (*state).Wield,
		"VERSION":   (*state).Version,
		"SAVE":      (*state).Save,
		"HIT":       (*state).Hit,
		"TELL":      (*state).Tell,
		"TALK":      (*state).Tell,
		"WHISPER":   (*state).Whisper,

		// Out of character commands
		"/WHO":    (*state).Who,
		"/WHOAMI": (*state).WhoAmI,

		// Admin and debugging commands
		"#DUMP":     (*state).Dump,
		"#LDUMP":    (*state).Dump,
		"#TELEPORT": (*state).Teleport,
		"#GOTO":     (*state).Teleport,
		"#DEBUG":    (*state).Debug,

		// Scripting only commands
		"$POOF":    (*state).Poof,
		"$ACT":     (*state).Act,
		"$ACTION":  (*state).Action,
		"$RESET":   (*state).Reset,
		"$CLEANUP": (*state).Cleanup,
		"$TRIGGER": (*state).Trigger,
		"$QUIT":    (*state).Quit,
		"$HEALTH":  (*state).Health,
	}

	eventCommands = map[eventKey]string{
		Action:  "$ACTION",
		Reset:   "$RESET",
		Cleanup: "$CLEANUP",
		Trigger: "$TRIGGER",
		Health:  "$HEALTH",
	}

	// precompute a sorted list of available player and admin commands. Scripting
	// commands with a '$' prefix are not included.
	for name := range commandHandlers {
		if name != "" && name[0] != '$' {
			commandNames = append(commandNames, name)
		}
	}
	sort.Strings(commandNames)

	log.Printf("Registered %d command handlers", len(commandHandlers))
}

// FIXME: We reset usage here in case item is unique, should it go somewhere
// else? Thing.Junk maybe?
func (s *state) Quit() {
	delete(Players, s.actor.As[UID])
	s.Prompt(s.actor, "")

	// If scripting QUIT user has not hit enter so nudge them off the prompt
	if s.cmd == "$QUIT" {
		s.Msg(s.actor, "")
	}

	s.quitUniqueCheck(s.actor)
	s.Save()
	where := s.actor.Ref[Where]
	for _, what := range s.actor.In {
		what.Is &^= Using
		what.Junk()
	}
	delete(where.Who, s.actor.As[UID])

	s.Msg(s.actor, text.Good, "You leave this world behind.\n")
	if len(where.Who) < cfg.crowdSize {
		s.Msg(where, text.Info, s.actor.As[Name],
			" gives a strangled cry of 'Bye Bye', slowly fades away and is gone.")
	}
}

// quitUniqueCheck will force junk any unique items the player is carrying.
// This is recursive and looks for unique items in nested containers.
//
// BUG(diddymus): Junking a unique container junks all its content.
func (s *state) quitUniqueCheck(what *Thing) {
	for _, item := range what.In {
		if item.Ref[Origin] != nil {
			s.Msg(s.actor, text.Red, "You cannot take ", item.As[TheName], " with you.")
			item.Is &^= Using
			item.Junk()
		} else {
			s.quitUniqueCheck(item)
		}
	}
}

func (s *state) Look() {
	where := s.actor.Ref[Where]

	switch {
	case where == nil:
		s.Msg(s.actor, text.Cyan, "[The Void]\n", text.Reset,
			"You are in a dark void. Around you nothing.",
			"No stars, no light, no heat and no sound.")
	case where.Is&Dark == Dark:
		s.Msg(s.actor, text.Bad, "It's too dark to see anything!")
	default:
		s.Msg(s.actor, text.Cyan, "[", where.As[Name], "]", text.Reset)
		s.Msg(s.actor, where.As[Description], "\n")
		mark := s.buf[s.actor].Len()
		if len(where.Who) < cfg.crowdSize {
			for _, who := range where.Who.Sort() {
				if who == s.actor {
					continue
				}
				s.Msg(s.actor, text.Green, "You see ", who.As[Name], " here.")
			}
			for _, item := range where.In.Sort() {
				if item.Is&Narrative == Narrative || item == s.actor {
					continue
				}
				s.Msg(s.actor, text.Yellow, "You see ", item.As[Name], " here.")
			}
			if mark != s.buf[s.actor].Len() {
				s.Msg(s.actor)
				mark = s.buf[s.actor].Len()
			}
		} else {
			s.Msg(s.actor, text.Green, "It's too crowded here to see anything.\n")
			mark = s.buf[s.actor].Len()
		}

		// Get directions in a fixed order
		for dir := North; dir <= Down; dir++ {
			if where.Ref[dir] != nil {
				if s.buf[s.actor].Len() == mark {
					s.Msg(s.actor, text.Cyan, "You see exits: ", text.Reset, DirToName[dir])
				} else {
					s.MsgAppend(s.actor, ", ", DirToName[dir])
				}
			}
		}

		if mark == s.buf[s.actor].Len() {
			s.Msg(s.actor, text.Cyan, "You see no obvious exits.")
		}
	}

	// Only notify observers if actually looking and not $POOF or entering a
	// location when moving.
	if (s.cmd == "L" || s.cmd == "LOOK") && len(where.Who) < cfg.crowdSize {
		s.Msg(where, text.Info, s.actor.As[UTheName], " starts looking around.")
	}
}

func (s *state) Move() {

	dir := NameToDir[s.cmd]
	where := s.actor.Ref[Where]

	if where.Ref[dir] == nil {
		s.Msg(s.actor, text.Bad, "You can't go ", DirToName[dir], ".")
		return
	}

	// Alias is a helper closure that calculates the aliases for the actor when
	// first called. Subsequent calls return the calculated value. For players
	// we don't test barriers against the player's name alias. Otherwise some
	// wise-ass is going to call themselves "NPC", "GUARD", etc.
	aliases := func() func() []string {
		var a []string
		return func() []string {
			if a != nil {
				return a
			}
			a = s.actor.Any[Alias]
			if s.actor.Is&Player == Player {
				a, _ = remainder(a, []string{strings.ToUpper(s.actor.As[Name])})
			}
			return a
		}
	}()

	// See if a location barrier is blocking our way
	if NameToDir[where.As[Barrier]] == dir {
		a, d := where.Any[BarrierAllow], where.Any[BarrierDeny]
		if !intersects(a, aliases()) && (d == nil || intersects(d, aliases())) {
			s.Msg(s.actor, text.Bad, "You can't go ", DirToName[dir], ", something is blocking your way.")
			return
		}
	}

	// See if any item barriers are blocking our way
	for _, item := range where.In {
		if NameToDir[item.As[Barrier]] != dir {
			continue
		}
		a, d := item.Any[BarrierAllow], item.Any[BarrierDeny]
		if !intersects(a, aliases()) && (d == nil || intersects(d, aliases())) {
			s.Msg(s.actor, text.Bad, "You can't go ", DirToName[dir], ", ", or(item.As[TheName], "something"), " is blocking your way.")
			return
		}
	}

	// Try and find first blocker for direction we want to go
	var blocker *Thing
	for _, item := range where.In {
		if item.As[Blocker] == "" {
			continue
		}
		blocking := NameToDir[item.As[Blocker]]
		// If on 'other side' need opposite direction blocked
		if item.Ref[Where] != s.actor.Ref[Where] {
			blocking = ReverseDir[blocking]
		}
		if blocking == dir && item.Is&Open != Open {
			blocker = item
			break
		}
	}

	switch {
	case blocker != nil:
		s.Msg(s.actor, text.Bad, "You can't go ", DirToName[dir], ". ",
			blocker.As[TheName], " is blocking your way.")
	case where.Ref[dir] == nil:
		s.Msg(s.actor, text.Bad, "Oops! You can't actually go ", DirToName[dir], ".")
	case s.actor.Is&Player != Player:
		delete(where.In, s.actor.As[UID])
		if len(where.Who) < cfg.crowdSize {
			s.MsgAppend(where, text.Info, s.actor.As[UTheName], " leaves ", DirToName[dir], ".")
		}

		where = where.Ref[dir]
		s.actor.Ref[Where] = where
		where.In[s.actor.As[UID]] = s.actor
		if len(where.Who) < cfg.crowdSize {
			s.MsgAppend(where, text.Info, s.actor.As[UName], " enters.")
		}
	default:
		delete(where.Who, s.actor.As[UID])
		if len(where.Who) < cfg.crowdSize {
			s.MsgAppend(where, text.Info, s.actor.As[UTheName], " leaves ", DirToName[dir], ".")
		}

		where = where.Ref[dir]
		s.actor.Ref[Where] = where
		where.Who[s.actor.As[UID]] = s.actor
		if len(where.Who) < cfg.crowdSize {
			s.MsgAppend(where, text.Info, s.actor.As[UName], " enters.")
		}
		s.Look()
	}
}

// FIXME(diddymus): At the moment containers can contain narritives. This
// complicates describing a container's inventory. Should this be allowed? What
// could narratives in containers be useful for?
func (s *state) Examine() {

	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You examine this and that, find nothing special.")
		return
	}

	uids := Match(s.word, s.actor.Ref[Where], s.actor)
	uid := uids[0]
	what := s.actor.In[uid]
	if what == nil {
		what = s.actor.Ref[Where].In[uid]
	}
	if what == nil {
		what = s.actor.Ref[Where].Who[uid]
	}

	switch {
	case what == nil:
		s.Msg(s.actor, text.Bad, "You see no '", uid, "' to examine.")
	case len(uids) > 1:
		s.Msg(s.actor, text.Bad, "You can only examine one thing at a time.")
	case uid == s.actor.As[UID]:
		s.Msg(s.actor, text.Good, "Looking fine!")
	default:
		s.Msg(s.actor, text.Good, "You examine ", what.As[TheName], ".\n", text.Reset, what.As[Description])

		// If a blocker, e.g. a door, is it open or closed?
		switch {
		case what.As[Blocker] == "":
		case what.Is&Open == Open:
			s.MsgAppend(s.actor, " It is open.")
		default:
			s.MsgAppend(s.actor, " It is closed.")
		}

		// If a container then count non-narrative items in it. When examining
		// containers we only want to describe non-narrative content.
		itemCount := 0
		if what.Is&Container == Container {
			for _, item := range what.In {
				if item.Is&Narrative == 0 {
					itemCount++
				}
			}
		}

		// If a container, describe its content
		switch {
		case what.Is&Container == 0:
			// Not a container
		case itemCount == 0 && what.Is&Narrative == Narrative:
			// Don't describe empty narrative containers ;)
		case itemCount == 0:
			s.MsgAppend(s.actor, " It is empty.")
		case itemCount == 1:
			for _, item := range what.In {
				if item.Is&Narrative == 0 {
					s.MsgAppend(s.actor, " It contains ", item.As[Name], ".")
				}
			}
		default:
			s.MsgAppend(s.actor, " It contains: ")
			for _, item := range what.In.Sort() {
				s.Msg(s.actor, "  ", item.As[Name])
			}
		}

		// If we have a body describe items being worn, held or wielded.
		if what.Is&HasBody == HasBody {
			var wearing, holding, wielding []string
			for _, item := range what.In {
				switch {
				case item.Is&Holding == Holding:
					holding = append(holding, item.As[Name])
				case item.Is&Wearing == Wearing:
					wearing = append(wearing, item.As[Name])
				case item.Is&Wielding == Wielding:
					wielding = append(wielding, item.As[Name])
				}
			}

			if len(wearing) > 0 {
				s.MsgAppend(s.actor, " They are wearing ", text.List(wearing), ".")
			}

			switch {
			case len(holding) > 0 && len(wielding) > 0:
				s.MsgAppend(s.actor, " They are holding ", text.List(holding),
					" while wielding ", text.List(wielding), ".")
			case len(holding) > 0:
				s.MsgAppend(s.actor, " They are holding ", text.List(holding), ".")
			case len(wielding) > 0:
				s.MsgAppend(s.actor, " They are wielding ", text.List(wielding), ".")
			}
		}

		if len(s.actor.Ref[Where].Who) < cfg.crowdSize {
			if what.Is&Player == Player {
				s.Msg(what, text.Info, s.actor.As[UTheName], " studies you.")
			}
			s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " studies ", what.As[Name], ".")
		}
	}
}

func (s *state) Inventory() {
	switch {
	case len(s.actor.In) == 0:
		s.Msg(s.actor, text.Info, "You are not carrying anything.")
	default:
		s.Msg(s.actor, "You currently have:")
		for _, what := range s.actor.In.Sort() {
			var usage string
			switch {
			case what.Is&Holding == Holding:
				usage = " (held)"
			case what.Is&Wearing == Wearing:
				usage = " (worn)"
			case what.Is&Wielding == Wielding:
				usage = " (wielded)"
			}
			s.Msg(s.actor, "  ", what.As[Name], usage)
		}
		if len(s.actor.Ref[Where].Who) < cfg.crowdSize {
			s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " checks over their gear.")
		}
	}
}

func (s *state) Drop() {

	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to drop... something?")
		return
	}

	notify := len(s.actor.Ref[Where].Who) < cfg.crowdSize

	for _, uid := range Match(s.word, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You do not have any '", uid, "' to drop.")
		case what.As[VetoDrop] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoDrop])
		case what.Is&Using != 0:
			s.Msg(s.actor, text.Bad, "You can't drop ", what.As[TheName], " while using it.")
		default:
			delete(s.actor.In, what.As[UID])
			s.actor.Ref[Where].In[what.As[UID]] = what
			what.Schedule(Action)
			what.Schedule(Cleanup)
			what.Ref[Where] = s.actor.Ref[Where]
			delete(what.As, DynamicQualifier)
			s.Msg(s.actor, text.Good, "You drop ", what.As[TheName], ".")
			if notify {
				s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " drops ", what.As[Name], ".")
			}
		}
	}
}

func (s *state) Get() {

	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to get... something?")
		return
	}

	notify := len(s.actor.Ref[Where].Who) < cfg.crowdSize

	for _, uid := range Match(s.word, s.actor.Ref[Where]) {
		what := s.actor.Ref[Where].In[uid]
		if what == nil {
			what = s.actor.Ref[Where].Who[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You see no '", uid, "' to get.")
		case what.As[VetoGet] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoGet])
		case uid == s.actor.As[UID]:
			s.Msg(s.actor, text.Info, "Trying to pick youreself up by your bootlaces?")
		case what.Is&Narrative == Narrative:
			s.Msg(s.actor, text.Bad, "You cannot take ", what.As[TheName], ".")
		case what.Is&(NPC|Player) != 0 && len(what.Any[Holdable]) == 0:
			s.Msg(s.actor, text.Bad, what.As[UTheName], " does not want to be taken!")
		default:
			what.Suspend(Action)
			what.Cancel(Cleanup)
			delete(s.actor.Ref[Where].In, what.As[UID])
			what = what.Spawn()

			// If item is a container it may have had items put in to it before it
			// was spawned. In which case the items will have pending clean-ups that
			// need to be cancelled.
			for _, what := range what.In {
				what.Cancel(Cleanup)
			}

			s.actor.In[what.As[UID]] = what
			what.Ref[Where] = s.actor
			what.As[DynamicQualifier] = "MY"
			s.Msg(s.actor, text.Good, "You get ", what.As[TheName], ".")
			if notify {
				s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " picks up ", what.As[TheName], ".")
			}
		}
	}
}

func (s *state) Take() {

	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to take something from something else...")
		return
	}

	uids, words := LimitedMatch(s.word, s.actor, s.actor.Ref[Where])
	uid := uids[0]
	where := s.actor.In[uid]
	if where == nil {
		where = s.actor.Ref[Where].In[uid]
	}
	if where == nil {
		where = s.actor.Ref[Where].Who[uid]
	}

	switch {
	case where == nil:
		s.Msg(s.actor, text.Bad, "You see no '", uid, "' to take anything from.")
	case len(uids) > 1:
		s.Msg(s.actor, text.Bad, "You can only take things from one container at a time.")
	case where.Is&(NPC|Player) != 0:
		s.Msg(s.actor, text.Bad, where.As[UTheName], " does not want you taking anything of theirs!")
	case where.Is&Container != Container:
		s.Msg(s.actor, text.Bad, where.As[UTheName], " is not something you can take anything from.")
	case len(words) == 0:
		s.Msg(s.actor, text.Info, "You go to take something from ", where.As[TheName], ".")
	case where.As[VetoTakeOut] != "":
		s.Msg(s.actor, text.Bad, where.As[VetoTakeOut])
	}
	if s.buf[s.actor] != nil {
		return
	}

	notify := false
	for _, uid := range Match(words, where) {
		what := where.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, where.As[UTheName], " does not seem to contain '", uid, "'.")
		case what.As[VetoTake] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoTake])
		case what.Is&Narrative == Narrative:
			s.Msg(s.actor, text.Bad, "You can't take ", what.As[TheName], " from ", where.As[TheName], ".")
		default:
			what.Cancel(Cleanup)
			delete(where.In, what.As[UID])
			what = what.Spawn()
			s.actor.In[what.As[UID]] = what
			what.Ref[Where] = s.actor
			what.As[DynamicQualifier] = "MY"
			s.Msg(s.actor, text.Good, "You take ", what.As[TheName], " out of ", where.As[TheName], ".")
			notify = true
		}

	}
	if notify && len(s.actor.Ref[Where].Who) < cfg.crowdSize {
		if s.actor.In[where.As[UID]] == nil {
			s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " takes something out of ", where.As[TheName], ".")
		} else {
			s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " takes something out of ", where.As[Name], " they are carrying.")
		}
	}
}

func (s *state) Put() {

	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to put something into something else...")
		return
	}

	uids, words := LimitedMatch(s.word, s.actor, s.actor.Ref[Where])
	uid := uids[0]
	where := s.actor.In[uid]
	if where == nil {
		where = s.actor.Ref[Where].In[uid]
	}
	if where == nil {
		where = s.actor.Ref[Where].Who[uid]
	}

	switch {
	case where == nil:
		s.Msg(s.actor, text.Bad, "You see no '", uid, "' to put anything into.")
	case len(uids) > 1:
		s.Msg(s.actor, text.Bad, "You can only put things into one container at a time.")
	case where.Is&(NPC|Player) != 0:
		s.Msg(s.actor, text.Info, "Taxidermist are we?")
	case where.Is&Container != Container:
		s.Msg(s.actor, text.Bad, where.As[UTheName], " is not something you can put anything into.")
	case len(words) == 0:
		s.Msg(s.actor, text.Bad, "You go to put something into ", where.As[TheName], ".")
	case where.As[VetoPutIn] != "":
		s.Msg(s.actor, text.Bad, where.As[VetoPutIn])
	}
	if s.buf[s.actor] != nil {
		return
	}

	parent := where.Ref[Where]

	notify := false
	for _, uid := range Match(words, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You have no '", uid, "' to put into ", where.As[TheName], ".")
		case what.Is&NPC == NPC:
			s.Msg(s.actor, text.Bad, what.As[UTheName], " does not want to be put into ", where.As[TheName], ".")
		case what.As[VetoPut] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoPut])
		case what.Is&Using != 0:
			s.Msg(s.actor, text.Bad, "You can't put ", what.As[TheName], " anywhere while using it.")
		case uid == where.As[UID]:
			s.Msg(s.actor, text.Info, "It might be interesting to put ", what.As[TheName],
				" inside itself, but probably paradoxical as well.")
		default:
			delete(s.actor.In, what.As[UID])
			where.In[what.As[UID]] = what
			what.Ref[Where] = where

			// If container not held by a player and not pending a clean-up then schedule our own
			if (parent == nil || parent.Is&Player == 0) && where.Event[Cleanup] == nil {
				what.Schedule(Cleanup)
			}

			delete(what.As, DynamicQualifier)
			s.Msg(s.actor, text.Good, "You put ", what.As[TheName], " into ", where.As[TheName], ".")
			notify = true
		}
	}

	if notify && len(s.actor.Ref[Where].Who) < cfg.crowdSize {
		if s.actor.In[where.As[UID]] == nil {
			s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " puts something into ", where.As[TheName], ".")
		} else {
			s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " puts something into ", where.As[Name], " they are carrying.")
		}
	}
}

func (s *state) Dump() {
	if !cfg.allowDump {
		s.Msg(s.actor, text.Bad, "The #DUMP and #LDUMP commands are unavailable.")
		return
	}
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "What did you want to dump?")
		return
	}
	var uids []string
	if s.word[0] == "@" {
		uids = []string{s.actor.Ref[Where].As[UID]}
	} else {
		uids = Match(s.word, s.actor.Ref[Where], s.actor)
	}
	for _, uid := range uids {
		what := s.actor.In[uid]
		if what == nil {
			what = s.actor.Ref[Where].In[uid]
		}
		if what == nil {
			what = s.actor.Ref[Where].Who[uid]
		}
		if what == nil {
			what = World[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You see no '", uid, "' to dump.")
		case s.cmd == "#LDUMP":
			buf := &bytes.Buffer{}
			what.Dump(buf, 132)
			log.Printf("DUMP: %s\n%s", uid, buf.Bytes())
			s.Msg(s.actor, text.Info, "Dumped to log: ", uid)
		default:
			s.Msg(s.actor, "DUMP: ", uid, "\n")
			buf := &bytes.Buffer{}
			what.Dump(buf, 80)
			s.buf[s.actor].Write(bytes.ReplaceAll(buf.Bytes(), []byte(" "), []byte("␠")))
		}
	}
}

func (s *state) Read() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to read something...")
		return
	}
	for _, uid := range Match(s.word, s.actor.Ref[Where], s.actor) {
		what := s.actor.Ref[Where].In[uid]
		if what == nil {
			what = s.actor.Ref[Where].Who[uid]
		}
		if what == nil {
			what = s.actor.In[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You see no '", uid, "' here to read.")
		case what.As[Writing] == "":
			s.Msg(s.actor, text.Bad, "There is nothing on ", what.As[TheName], " to read.")
		case what.As[VetoRead] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoRead])
		default:
			s.Msg(s.actor, text.Good, "You read ", what.As[TheName], ". ", what.As[Writing])
			if len(s.actor.Ref[Where].Who) < cfg.crowdSize {
				s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " reads ", what.As[Name], ".")
			}
		}
	}
}

func (s *state) Open() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to open something...")
		return
	}
	for _, uid := range Match(s.word, s.actor.Ref[Where]) {
		what := s.actor.Ref[Where].In[uid]
		if what == nil {
			what = s.actor.Ref[Where].Who[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You see no '", uid, "' to open.")
		case what.As[Blocker] == "":
			s.Msg(s.actor, text.Bad, what.As[UTheName], " is not something you can open.")
		case what.Is&Open == Open:
			s.Msg(s.actor, text.Bad, what.As[UTheName], " is already open.")
		case s.actor != what && what.As[VetoOpen] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoOpen])
		default:
			what.Is |= Open
			if what.Is&_Open == 0 {
				what.Schedule(Trigger)
			} else {
				what.Cancel(Trigger)
			}

			where := s.actor.Ref[Where]

			if s.actor != what {
				s.Msg(s.actor, text.Good, "You open ", what.As[TheName], ".")
			}

			if len(where.Who) < cfg.crowdSize {
				if s.actor != what {
					s.Msg(where, text.Info, s.actor.As[UTheName], " opens ", what.As[TheName], ".")
				} else {
					s.Msg(where, text.Info, what.As[UTheName], " opens.")
				}
			}

			// Find location on other side...
			if where == what.Ref[Where] {
				exit := NameToDir[what.As[Blocker]]
				where = where.Ref[exit]
			} else {
				where = what.Ref[Where]
			}
			if len(where.Who) < cfg.crowdSize {
				s.Msg(where, text.Info, what.As[UTheName], " opens.")
			}
		}
	}
}

func (s *state) Close() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to close something...")
		return
	}
	for _, uid := range Match(s.word, s.actor.Ref[Where]) {
		what := s.actor.Ref[Where].In[uid]
		if what == nil {
			what = s.actor.Ref[Where].Who[uid]
		}
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You see no '", uid, "' to close.")
		case what.As[Blocker] == "":
			s.Msg(s.actor, text.Bad, what.As[UTheName], " is not something you can close.")
		case what.Is&Open == 0:
			s.Msg(s.actor, text.Bad, what.As[UTheName], " is already closed.")
		case s.actor != what && what.As[VetoClose] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoClose])
		default:
			what.Is &^= Open
			if what.Is&_Open == _Open {
				what.Schedule(Trigger)
			} else {
				what.Cancel(Trigger)
			}

			where := s.actor.Ref[Where]

			if s.actor != what {
				s.Msg(s.actor, text.Good, "You close ", what.As[TheName], ".")
			}

			if len(where.Who) < cfg.crowdSize {
				if s.actor != what {
					s.Msg(where, text.Info, s.actor.As[UTheName], " closes ", what.As[TheName], ".")
				} else {
					s.Msg(where, text.Info, what.As[UTheName], " closes.")
				}
			}

			// Find location on other side...
			if where == what.Ref[Where] {
				exit := NameToDir[what.As[Blocker]]
				where = where.Ref[exit]
			} else {
				where = what.Ref[Where]
			}
			if len(where.Who) < cfg.crowdSize {
				s.Msg(where, text.Info, what.As[UTheName], " closes.")
			}
		}
	}
}

func (s *state) Commands() {
	cols := 7
	split := (len(commandNames) / cols) + 1
	pad := []rune("␠␠␠␠␠␠␠␠␠␠␠␠␠")
	s.Msg(s.actor, "Commands currently available:\n\n")
	for x := 0; x < split; x++ {
		for y := x; y < len(commandNames); y += split {
			if y >= len(commandNames) {
				continue
			}
			s.MsgAppend(s.actor, "␠␠", commandNames[y], string(pad[:9-len(commandNames[y])]))
		}
		s.Msg(s.actor)
	}
}

func (s *state) Teleport() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "Where do you want to go?")
		return
	}
	where := World[s.word[0]]
	switch {
	case where == nil:
		s.Msg(s.actor, text.Bad, "You don't know where '", s.word[0], "' is.")
	default:
		delete(s.actor.Ref[Where].Who, s.actor.As[UID])
		if len(s.actor.Ref[Where].Who) < cfg.crowdSize {
			s.Msg(s.actor.Ref[Where], text.Info, "There is a loud 'Spang!' and ", s.actor.As[TheName], " suddenly disappears.")
		}
		s.actor.Ref[Where] = where
		s.actor.Ref[Where].Who[s.actor.As[UID]] = s.actor
		s.Msg(s.actor, text.Good, "There is a loud 'Spang!'...\n")
		s.Look()
		if len(s.actor.Ref[Where].Who) < cfg.crowdSize {
			s.Msg(s.actor.Ref[Where], text.Info, "There is a loud 'Spang!' and ", s.actor.As[Name], " suddenly appears.")
		}
	}
}

func (s *state) Poof() {
	Players[s.actor.As[UID]] = s.actor
	s.buildPrompt(s.actor)
	if s.actor.Int[HealthCurrent] < s.actor.Int[HealthMaximum] {
		s.actor.Schedule(Health)
	}

	if len(s.actor.Ref[Where].Who) < cfg.crowdSize {
		s.Msg(s.actor.Ref[Where], text.Info, "There is a cloud of smoke from which ",
			s.actor.As[Name], " emerges coughing and spluttering.")
	}
	s.Look()
}

func (s *state) buildPrompt(actor *Thing) {
	s.Prompt(actor, "\n%sH:%d/%d%s>",
		text.Blue, actor.Int[HealthCurrent], actor.Int[HealthMaximum], text.Magenta,
	)
}

func (s *state) Act() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "What did you want to act out?")
		return
	}

	s.Msg(s.actor, text.Good, s.actor.As[UTheName], " ", s.input)
	if len(s.actor.Ref[Where].Who) < cfg.crowdSize {
		s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " ", s.input)
	}
}

func (s *state) Say() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "What did you want to say?")
		return
	}

	where := s.actor.Ref[Where]
	l := len(where.Who)

	if l >= cfg.crowdSize {
		s.Msg(s.actor, text.Info, "It's too crowded for you to be heard.")
		return
	}

	if l == 1 {
		s.Msg(s.actor, text.Info, "Talking to yourself again?")
	} else {
		s.Msg(s.actor, text.Good, "You say: ", s.input)
		s.Msg(where, text.Info, s.actor.As[UTheName], " says: ", s.input)
	}

	for _, where := range radius(1, where)[1] {
		if l = len(where.Who); 0 < l && l < cfg.crowdSize {
			s.Msg(where, text.Info, "You hear talking nearby.")
		}
	}
}

func (s *state) Action() {
	s.actor.Cancel(Action)
	l := len(s.actor.Any[OnAction])
	if l == 0 {
		return
	}

	s.subparse(s.actor.Any[OnAction][rand.Intn(l)])
	s.actor.Schedule(Action)
}

// FIXME(diddymus): Currently SNEEZE has very aggressive crowd control to limit
// the amount of broadcasting we do, otherwise network traffic and CPU usage
// goes through the roof.
func (s *state) Sneeze() {

	s.Msg(s.actor, text.Good, "You sneeze. Aaahhhccchhhooo!")

	// Don't propagate sneeze if it's crowded.
	if len(s.actor.Ref[Where].Who) >= cfg.crowdSize {
		return
	}

	s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " sneezes.")

	locs := radius(2, s.actor.Ref[Where])
	for _, where := range locs[1] {
		if l := len(where.Who); 0 < l && l < cfg.crowdSize {
			s.Msg(where, text.Info, "You hear a loud sneeze.")
		}
	}
	for _, where := range locs[2] {
		if l := len(where.Who); 0 < l && l < cfg.crowdSize {
			s.Msg(where, text.Info, "You hear a sneeze.")
		}
	}
}

// FIXME(diddymus): Currently SHOUT has very aggressive crowd control to limit
// the amount of broadcasting we do, otherwise network traffic and CPU usage
// goes through the roof.
func (s *state) Shout() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "What did you want to shout?")
		return
	}

	// Don't propagate shout if it's crowded.
	if len(s.actor.Ref[Where].Who) >= cfg.crowdSize {
		s.Msg(s.actor, text.Info, "Even shouting, it's too crowded for you to be heard.")
		return
	}

	s.Msg(s.actor, text.Good, "You shout: ", s.input)
	s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " shouts: ", s.input)

	locs := radius(2, s.actor.Ref[Where])
	for _, where := range locs[1] {
		if l := len(where.Who); 0 < l && l < cfg.crowdSize {
			s.Msg(where, text.Info, "You hear someone shout: ", s.input)
		}
	}
	for _, where := range locs[2] {
		if l := len(where.Who); 0 < l && l < cfg.crowdSize {
			s.Msg(where, text.Info, "You hear shouting nearby.")
		}
	}
}

// radius returns the locations within size moves of a location. The locations
// are returned as 'rings' around the given location. For example [0][0] is the
// central location, [1][...] are locations within one move, [2][...] are
// locations within two moves, etc. Note that the radius is 3D and includes
// locations above and below.
//
// BUG(diddymus): Blockers, such as doors, are currently ignored.
func radius(size int, where *Thing) [][]*Thing {
	locs := make([][]*Thing, size+1)
	seen := make(map[*Thing]struct{})

	// Add central location
	locs[0] = append(locs[0], where)
	seen[where] = struct{}{}

	var (
		found bool
		dir   refKey
		loc   *Thing
	)
	for r := 1; r <= size; r++ {
		for _, where = range locs[r-1] {
			for dir = range DirToName {
				if loc = where.Ref[dir]; loc == nil {
					continue
				}
				if _, found = seen[loc]; found {
					continue
				}
				locs[r] = append(locs[r], loc)
				seen[loc] = struct{}{}
			}
		}
	}
	return locs
}

func (s *state) Reset() {
	s.actor.Cancel(Reset)

	// If actor should wait and has out of play items don't reset
	if s.actor.Is&Wait == Wait && len(s.actor.Out) > 0 {
		return
	}

	where := s.actor.Ref[Where]
	parent := where.Ref[Where]
	s.actor.Init()
	delete(where.Out, s.actor.As[UID])
	where.In[s.actor.As[UID]] = s.actor

	// Check parent of where reset will happen to see if where is out of play.
	// If where is out of play reset will not be seen. However, if where reset
	// will happen now has no out of play items we can schedule a reset for it.
	if parent != nil && parent.Out[where.As[UID]] != nil {
		if len(where.Out) == 0 {
			where.Schedule(Reset)
		}
		return
	}

	// If reset message supressed we can just bail now
	if msg, ok := s.actor.As[OnReset]; ok && msg == "" {
		return
	}

	// If resetting in a container the reset will not be seen. However, if we
	// have a custom reset message and the parent is a location or player we send
	// the custom message there. This lets custom messages notify players that
	// something has happened in the container.
	if where.Is&Container == Container {
		if s.actor.As[OnReset] != "" && parent.Is&(Player|Location) != 0 {
			where = parent
		} else {
			return
		}
	}

	// If where message is being sent is crowded we won't see it
	if len(where.Who) >= cfg.crowdSize {
		return
	}

	if s.actor.As[OnReset] == "" {
		s.Msg(where, text.Info, "You notice ", s.actor.As[Name], " you didn't see before.")
		return
	}
	s.Msg(where, text.Info, s.actor.As[OnReset])
}

func (s *state) Junk() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "Now what did you want to go and junk?")
		return
	}

	notify := len(s.actor.Ref[Where].Who) < cfg.crowdSize

	for _, uid := range Match(s.word, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You have no '", uid, "' to junk.")
		case what.As[VetoJunk] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoJunk])
		case what.Is&Using != 0:
			s.Msg(s.actor, text.Bad, "You can't junk ", what.As[TheName], " while using it.")
		default:
			s.Msg(s.actor, text.Good, "You junk ", what.As[TheName], ".")
			if notify {
				s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " junks ", what.As[Name], ".")
			}
			what.Junk()
		}
	}
}

func (s *state) Cleanup() {
	s.actor.Cancel(Cleanup)
	defer s.actor.Junk()

	where := s.actor.Ref[Where]

	if where.Is&Container == Container || len(where.Who) >= cfg.crowdSize {
		return
	}

	if msg, ok := s.actor.As[OnCleanup]; ok {
		if msg != "" {
			s.Msg(where, text.Info, msg)
		}
		return
	}

	s.Msg(where, text.Info, "You thought you noticed ", s.actor.As[Name], " here, but you can't see it now.")
}

// Trigger is used to process a trigger event for an item. The trigger type can
// be set via Thing.As[TriggerType]. The item can be in or out of play at the
// time.
func (s *state) Trigger() {
	s.actor.Cancel(Trigger)

	switch s.actor.As[TriggerType] {
	case "BLOCKER":
		if s.actor.Is&_Open == _Open {
			s.subparse("OPEN " + s.actor.As[UID])
		} else {
			s.subparse("CLOSE " + s.actor.As[UID])
		}
	}
}

func (s *state) Remove() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to remove... something?")
		return
	}

	where := s.actor.Ref[Where]
	notify := len(where.Who) < cfg.crowdSize

	for _, uid := range Match(s.word, s.actor) {
		what := s.actor.In[uid]
		var (
			usage string
			slots []string
		)
		switch {
		case what == nil:
			// Do nothing, avoids nil panic accessing a nil what's fields
		case what.Is&Holding == Holding:
			usage = " holding "
			slots = what.Any[Holdable]
		case what.Is&Wearing == Wearing:
			usage = " wearing "
			slots = what.Any[Wearable]
		case what.Is&Wielding == Wielding:
			usage = " wielding "
			slots = what.Any[Wieldable]
		}

		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You have no '", uid, "' to remove.")
		case what.Is&Using == 0:
			s.Msg(s.actor, text.Bad, "You are not using ", what.As[TheName], ".")
		case what.As[VetoRemove] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoRemove])
		default:
			what.Is &^= Using
			s.actor.Any[Body] = append(s.actor.Any[Body], slots...)
			s.Msg(s.actor, text.Good, "You stop", usage, what.As[TheName], ".")
			if notify {
				s.Msg(where, text.Info, s.actor.As[UTheName], " stops", usage, what.As[Name], ".")
			}
		}
	}
}

func (s *state) Hold() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to hold... something?")
		return
	}

	notify := len(s.actor.Ref[Where].Who) < cfg.crowdSize

	for _, uid := range Match(s.word, s.actor) {
		what := s.actor.In[uid]

		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You have no '", uid, "' to hold.")
		case what.Is&Holding == Holding:
			s.Msg(s.actor, text.Info, "You are already holding ", what.As[TheName], ".")
		case what.Any[Holdable] == nil:
			s.Msg(s.actor, text.Bad, what.As[UTheName], " isn't something you can hold.")
		case what.Is&Wearing == Wearing:
			s.Msg(s.actor, text.Bad, "You can't hold ", what.As[TheName], " while wearing it.")
		case what.Is&Wielding == Wielding:
			s.Msg(s.actor, text.Bad, "You can't hold ", what.As[TheName], " while wielding it.")
		case what.As[VetoHold] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoHold])
		case !contains(s.actor.Any[Body], what.Any[Holdable]):
			var whys []string
			for _, item := range s.actor.In {
				if item.Is&Holding != 0 && intersects(item.Any[Holdable], what.Any[Holdable]) {
					whys = append(whys, item.As[Name])
				}
			}
			if len(whys) == 0 {
				s.Msg(s.actor, text.Bad, "You are incapable of holding ", what.As[TheName], " properly.")
				return
			}
			s.Msg(s.actor, text.Bad, "You can't hold ", what.As[TheName], " while holding ", text.List(whys), ".")
		default:
			what.Is |= Holding
			s.actor.Any[Body], _ = remainder(s.actor.Any[Body], what.Any[Holdable])
			s.Msg(s.actor, text.Good, "You hold ", what.As[TheName], ".")
			if notify {
				s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " holds ", what.As[Name], ".")
			}
		}
	}
}

func (s *state) Wear() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to wear... something?")
		return
	}

	notify := len(s.actor.Ref[Where].Who) < cfg.crowdSize

	for _, uid := range Match(s.word, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You have no '", uid, "' to wear.")
		case what.Is&Wearing == Wearing:
			s.Msg(s.actor, text.Info, "You are already wearing ", what.As[TheName], ".")
		case what.Any[Wearable] == nil:
			s.Msg(s.actor, text.Bad, what.As[UTheName], " isn't something you can wear.")
		case what.Is&Holding == Holding:
			s.Msg(s.actor, text.Bad, "You can't wear ", what.As[TheName], " while holding it.")
		case what.Is&Wielding == Wielding:
			s.Msg(s.actor, text.Bad, "You can't wear ", what.As[TheName], " while wielding it.")
		case what.As[VetoWear] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoWear])
		case !contains(s.actor.Any[Body], what.Any[Wearable]):
			var whys []string
			for _, item := range s.actor.In {
				if item.Is&Wearing != 0 && intersects(item.Any[Wearable], what.Any[Wearable]) {
					whys = append(whys, item.As[Name])
				}
			}
			if len(whys) == 0 {
				s.Msg(s.actor, text.Bad, "You are incapable of wearing ", what.As[TheName], " properly.")
				return
			}
			s.Msg(s.actor, text.Bad, "You can't wear ", what.As[TheName], " while wearing ", text.List(whys), ".")
		default:
			what.Is |= Wearing
			s.actor.Any[Body], _ = remainder(s.actor.Any[Body], what.Any[Wearable])
			s.Msg(s.actor, text.Good, "You wear ", what.As[TheName], ".")
			if notify {
				s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " wears ", what.As[Name], ".")
			}
		}
	}
}

func (s *state) Wield() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to wield... something?")
		return
	}

	notify := len(s.actor.Ref[Where].Who) < cfg.crowdSize

	for _, uid := range Match(s.word, s.actor) {
		what := s.actor.In[uid]
		switch {
		case what == nil:
			s.Msg(s.actor, text.Bad, "You have no '", uid, "' to wield.")
		case what.Is&Wielding == Wielding:
			s.Msg(s.actor, text.Info, "You are already wielding ", what.As[TheName], ".")
		case what.Any[Wieldable] == nil:
			s.Msg(s.actor, text.Bad, what.As[UTheName], " isn't something you can wield.")
		case what.Is&Holding == Holding:
			s.Msg(s.actor, text.Bad, "You can't wield ", what.As[TheName], " while holding it.")
		case what.Is&Wearing == Wearing:
			s.Msg(s.actor, text.Bad, "You can't wield ", what.As[TheName], " while wearing it.")
		case what.As[VetoWield] != "":
			s.Msg(s.actor, text.Bad, what.As[VetoWield])
		case !contains(s.actor.Any[Body], what.Any[Wieldable]):
			var whys []string
			for _, item := range s.actor.In {
				if item.Is&Wielding != 0 && intersects(item.Any[Wieldable], what.Any[Wieldable]) {
					whys = append(whys, item.As[Name])
				}
			}
			if len(whys) == 0 {
				s.Msg(s.actor, text.Bad, "You are incapable of wielding ", what.As[TheName], " properly.")
				return
			}
			s.Msg(s.actor, text.Bad, "You can't wield ", what.As[TheName], " while wielding ", text.List(whys), ".")
		default:
			what.Is |= Wielding
			s.actor.Any[Body], _ = remainder(s.actor.Any[Body], what.Any[Wieldable])
			s.Msg(s.actor, text.Good, "You wield ", what.As[TheName], ".")
			if notify {
				s.Msg(s.actor.Ref[Where], text.Info, s.actor.As[UTheName], " wields ", what.As[Name], ".")
			}
		}
	}
}

func (s *state) Version() {
	s.Msg(s.actor, "Version: ", commit, ", built with: ", runtime.Version(), " (", runtime.Compiler, ")")
}

func (s *state) Save() {

	j := &recordjar.Jar{}
	hdr := recordjar.Record{
		"Account":  encode.String(s.actor.As[Account]),
		"Created":  encode.DateTime(time.Unix(s.actor.Int[Created], 0)),
		"Password": encode.String(s.actor.As[Password]),
		"Salt":     encode.String(s.actor.As[Salt]),
		"Player":   encode.String(s.actor.As[UID]),
	}
	*j = append(*j, hdr)
	save(s.actor, j)

	temp := filepath.Join(cfg.playerPath, s.actor.As[Account]+".tmp")
	real := filepath.Join(cfg.playerPath, s.actor.As[Account]+".wrj")

	wrj, _ := os.Create(temp)
	wrj.Chmod(0660)
	j.Write(wrj, "DESCRIPTION", preferredOrdering)
	wrj.Close()
	os.Rename(temp, real)

	s.Msg(s.actor, text.Good, "You have been saved.")
}

func save(t *Thing, j *recordjar.Jar) {
	if t.Ref[Origin] == nil {
		*j = append(*j, t.Marshal())
	}
	for _, item := range t.In {
		save(item, j)
	}
	for _, item := range t.Out {
		save(item, j)
	}
}

var cpuProfile *os.File

func (s *state) Debug() {
	if !cfg.allowDebug {
		s.Msg(s.actor, text.Bad, "The #DEBUG command is unavailable.")
		return
	}

	if len(s.word) < 1 {
		s.Msg(s.actor, text.Info,
			"#DEBUG requires a sub-command: CPUPROF|MEMPROF|PANIC",
		)
		return
	}

	switch s.word[0] {
	case "CPUPROF":
		if len(s.word) < 2 {
			s.Msg(s.actor, text.Info,
				"#DEBUG CPUPROF requires an action: ON|OFF|START|STOP|END",
			)
			return
		}
		switch s.word[1] {
		case "ON", "START":
			if cpuProfile != nil {
				s.Msg(s.actor, text.Bad, "CPU profile already started.")
				return
			}
			cpuProfile, _ = os.Create("./cpuprof")
			cpuProfile.Chmod(0660)
			pprof.StartCPUProfile(cpuProfile)
			s.Msg(s.actor, text.Info, "CPU profile started.")
		case "OFF", "END", "STOP":
			if cpuProfile == nil {
				s.Msg(s.actor, text.Bad, "CPU profile not started.")
				return
			}
			pprof.StopCPUProfile()
			cpuProfile.Close()
			cpuProfile = nil
			s.Msg(s.actor, text.Info, "CPU profile stopped.")
		}
	case "MEMPROF":
		if len(s.word) < 2 {
			s.Msg(s.actor, text.Info,
				"#DEBUG MEMPROF requires an action: ON|OFF|START|STOP|END",
			)
			return
		}
		switch s.word[1] {
		case "ON", "START":
			runtime.MemProfileRate = 1
			s.Msg(s.actor, text.Info, "Memory profile started.")
		case "OFF", "END", "STOP":
			f, _ := os.Create("./memprof")
			f.Chmod(0660)
			runtime.GC()
			pprof.WriteHeapProfile(f)
			runtime.MemProfileRate = 0
			s.Msg(s.actor, text.Info, "Memory profile stopped.")
		}
	case "PANIC":
		s.Msg(s.actor, "You panic!")
		panic("User panicked.")
	}
}

func (s *state) Health() {
	s.actor.Cancel(Health)
	if s.actor.Int[HealthCurrent] >= s.actor.Int[HealthMaximum] {
		return
	}

	s.actor.Int[HealthCurrent] += s.actor.Int[HealthRestore]

	if s.actor.Int[HealthCurrent] >= s.actor.Int[HealthMaximum] {
		s.actor.Int[HealthCurrent] = s.actor.Int[HealthMaximum]
		s.Msg(s.actor, text.Good, "\nYou feel healthy.")
		s.Msg(s.actor.Ref[Where], text.Info,
			s.actor.As[UName], " looks healthy.")
	} else {
		s.actor.Schedule(Health)
	}

	s.buildPrompt(s.actor)
}

func (s *state) Hit() {

	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to hit... someone?")
		return
	}

	damage := 2 + rand.Int63n(2+1)
	damageTxt := strconv.FormatInt(damage, 10)
	where := s.actor.Ref[Where]
	notify := len(where.Who) < cfg.crowdSize

	uids := Match(s.word, where)
	uid := uids[0]
	what := where.Who[uid]
	if what == nil {
		what = where.In[uid]
	}

	switch {
	case what == nil:
		s.Msg(s.actor, text.Bad, "You see no '", uid, "' to hit.")
	case s.actor == what:
		s.Msg(s.actor, text.Good, "You give yourself a slap. Awake now?")
		s.Msg(where, text.Info, s.actor.As[UName], " slaps themself.")
	case where.As[VetoCombat] != "":
		s.Msg(s.actor, text.Bad, where.As[VetoCombat])
	case !notify:
		s.Msg(s.actor, text.Bad, "It's too crowded to start a fight.")
	case what.Int[HealthMaximum] == 0:
		s.Msg(s.actor, text.Bad, "You cannot kill ", what.As[Name], ".")
	case what.Int[HealthCurrent] <= damage:

		// Helper to center text within 80 columns
		center := func(text string) string {
			pad := (80 - len(text)) / 2
			return strings.ReplaceAll(strings.Repeat("␠", pad)+text, " ", "␠")
		}

		s.Msg(s.actor, text.Good, "You kill ", what.As[TheName], " (", damageTxt, ").")
		s.Msg(what, text.Bad, s.actor.As[TheName],
			" kills you (", damageTxt, ").",
			text.Cyan,
			"\n",
			"\n", center(" :==[ Rest In Peace ]==:"),
			"\n",
			"\n", center(what.As[Name]),
			"\n", center("Slain By"),
			"\n", center(s.actor.As[Name]),
			text.Good,
			"\n\nYou must know people in high places, you are to be given another chance...\n",
		)

		s.Msg(where, text.Info,
			"You see ", s.actor.As[TheName], " kill ", what.As[Name], ".")

		// Create and place corpse
		c := createCorpse(what)
		where.In[c.As[UID]] = c
		c.Schedule(Cleanup)

		// Remove original
		if what.Is&Player == 0 {
			what.Int[HealthCurrent] = what.Int[HealthMaximum]
			what.Junk()
		} else {
			what.Int[HealthCurrent] = 1
			delete(where.Who, what.As[UID])
			start := WorldStart[rand.Intn(len(WorldStart))]
			what.Ref[Where] = start
			start.Who[what.As[UID]] = what
			s.subparseFor(what, "$POOF")
		}
		s.buildPrompt(what)

	default:
		what.Int[HealthCurrent] -= damage
		if what.Event[Health] == nil && what.Int[HealthCurrent] < what.Int[HealthMaximum] {
			what.Schedule(Health)
		}
		s.buildPrompt(what)

		s.Msg(s.actor, text.Good, "You hit ", what.As[TheName], " (", damageTxt, ").")
		s.Msg(what, text.Bad, s.actor.As[UTheName], " hits you (", damageTxt, ").")
		s.Msg(where, text.Info,
			"You see ", s.actor.As[Name], " hit ", what.As[Name], ".")

		if what.Int[HealthCurrent] < 4 {
			s.MsgAppend(s.actor, text.Good, " ", what.As[UTheName], " looks nearly dead.")
			s.MsgAppend(what, text.Bad, " You are almost dead.")
			s.MsgAppend(where, text.Info, " ", what.As[UTheName], " is almost dead.")
		}

		locs := radius(1, s.actor.Ref[Where])
		for _, where := range locs[1] {
			if l := len(where.Who); 0 < l && l < cfg.crowdSize {
				s.Msg(where, text.Info, "You hear fighting nearby.")
			}
		}
	}
}

func createCorpse(t *Thing) *Thing {
	c := NewThing()
	c.As[Name] = "a corpse of " + t.As[Name]
	c.As[UName] = "A corpse of " + t.As[Name]
	c.As[TheName] = "the corpse of " + t.As[Name]
	c.As[UTheName] = "The corpse of " + t.As[Name]
	c.As[Description] = t.As[Description]
	c.Any[Alias] = append(c.Any[Alias], t.Any[Alias]...)
	c.Any[Qualifier] = append(c.Any[Qualifier], t.Any[Qualifier]...)
	c.Ref[Where] = t.Ref[Where]
	c.Ref[Where].In[c.As[UID]] = t
	c.Int[CleanupAfter] = time.Duration(60 * time.Second).Nanoseconds()
	c.As[OnCleanup] = c.As[UTheName] + " turns to dust."

	// Replace original UID alias with "CORPSE" (new UID was added by NewThing)
	for x, alias := range c.Any[Alias] {
		if alias == t.As[UID] {
			c.Any[Alias][x] = "CORPSE"
		}
	}

	return c
}

func (s *state) Tell() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to tell someone something...")
		return
	}

	uids := Match(s.word, s.actor.Ref[Where])
	uid := uids[0]

	what := s.actor.Ref[Where].Who[uid]
	if what == nil {
		what = s.actor.Ref[Where].In[uid]
	}

	where := s.actor.Ref[Where]
	l := len(where.Who)

	if l >= cfg.crowdSize {
		s.Msg(s.actor, text.Info, "It's too crowded for you to be heard.")
		return
	}

	switch {
	case what == nil:
		s.Msg(s.actor, text.Bad, "You see no '", s.word[0], "' to talk to.")
	case len(s.word) == 1:
		s.Msg(s.actor, text.Info, "What did you want to say to ", what.As[TheName], "?")
	default:
		txt := StripMatch(what, s.input)
		s.Msg(s.actor, text.Good, "You say to ", what.As[TheName], ": ", txt)
		s.Msg(what, text.Good, s.actor.As[UTheName], " says to you: ", txt)
		s.Msg(where, text.Info, s.actor.As[UTheName], " says to ", what.As[TheName], ": ", txt)

		for _, where := range radius(1, where)[1] {
			if l = len(where.Who); 0 < l && l < cfg.crowdSize {
				s.Msg(where, text.Info, "You hear talking nearby.")
			}
		}
	}
}

func (s *state) Whisper() {
	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to whisper something to someone...")
		return
	}

	uids := Match(s.word, s.actor.Ref[Where])
	uid := uids[0]

	what := s.actor.Ref[Where].Who[uid]
	if what == nil {
		what = s.actor.Ref[Where].In[uid]
	}

	where := s.actor.Ref[Where]
	l := len(where.Who)

	if l >= cfg.crowdSize {
		s.Msg(s.actor, text.Info, "It's too crowded for you to be heard.")
		return
	}

	switch {
	case what == nil:
		s.Msg(s.actor, text.Bad, "You see no '", s.word[0], "' to whisper to.")
	case len(s.word) == 1:
		s.Msg(s.actor, text.Info, "What did you want to whisper to ", what.As[TheName], "?")
	default:
		txt := StripMatch(what, s.input)
		s.Msg(s.actor, text.Good, "You whisper to ", what.As[TheName], ": ", txt)
		s.Msg(what, text.Good, s.actor.As[UTheName], " whispers to you: ", txt)
		s.Msg(where, text.Info, s.actor.As[UTheName], " whispers something to ", what.As[TheName], ".")
	}
}

func (s state) Who() {
	if len(Players) == 1 {
		s.Msg(s.actor, text.Good, "You are all alone in this world.")
		return
	}

	auid := s.actor.As[UID]
	pop := strconv.Itoa(len(Players))

	s.Msg(s.actor, text.Good, "Other players:\n\n", text.Reset)
	for uid, player := range Players {
		if uid == auid {
			continue
		}
		s.MsgAppend(s.actor, "␠␠", player.As[Name], "\n")
	}
	s.Msg(s.actor, text.Good, "Current player population: ", pop)
}

func (s state) WhoAmI() {
	s.Msg(s.actor, text.Good, "You are ", s.actor.As[UName], ".")
}

// intersects returns true if any elements of want are also in have, else false.
func intersects(have, want []string) bool {
	sort.Strings(have)
	sort.Strings(want)
	h, w := 0, 0
	for w < len(want) && h < len(have) {
		switch {
		case have[h] < want[w]:
			h++
		case want[w] < have[h]:
			w++
		default:
			return true
		}
	}
	return false
}

// contains returns true if have contains all elements of want, else false.
func contains(have, want []string) bool {
	sort.Strings(have)
	sort.Strings(want)
	x := 0
	for _, have := range have {
		if want[x] == have {
			x++
			if x == len(want) {
				return true
			}
		}
	}
	return false
}

// remainder returns the remaining elements of have after removing all elements
// of want and true. If have does not contain all elements of want nothing is
// removed and have is returned unmodified as well as false.
func remainder(have, want []string) (rem []string, had bool) {
	if len(want) == 0 {
		return have, false
	}
	sort.Strings(have)
	sort.Strings(want)
	x := 0
	for y, had := range have {
		if want[x] == had {
			x++
			if x == len(want) {
				return append(rem, have[y+1:]...), true
			}
		} else {
			rem = append(rem, had)
		}
	}
	return have, false
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
