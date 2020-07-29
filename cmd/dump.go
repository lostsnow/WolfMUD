// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"fmt"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Syntax: #DUMP|#UDUMP|#LDUMP alias
//
// The #DUMP, #UDUMP and #LDUMP commands are only available if the server is
// running with the configuration option Debug.AllowDump set to true.
func init() {
	addHandler(dump{}, "#DUMP")  // Dump ASCII graph to terminal
	addHandler(dump{}, "#UDUMP") // Dump Unicode graph to terminal
	addHandler(dump{}, "#LDUMP") // Dump ASCII graph to server log
}

type dump cmd

func (dump) process(s *state) {
	if !config.Debug.AllowDump {
		s.msg.Actor.SendBad("The #DUMP, #UDUMP and #LDUMP commands are not available. Server not running with configuration option Debug.AllowDump=true")
		return
	}

	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("%v", p)
			s.msg.Actor.SendBad("Error dumping ", s.input[0], ": ", err.Error())
		}
	}()

	if len(s.words) == 0 {
		s.msg.Actor.Send("What do you want to dump?")
		return
	}

	name := s.words[0]

	var what has.Thing

	// If we can, search where we are
	if s.where != nil {
		what = s.where.Search(name)
	}

	// If item still not found try our own inventory
	if what == nil {
		what = attr.FindInventory(s.actor).Search(name)
	}

	// If match still not found try the location itself - as opposed to it's
	// inventory and narratives.
	if what == nil && s.where != nil {
		if attr.FindAlias(s.where.Parent()).HasAlias(name) {
			what = s.where.Parent()
		}
	}

	// If item still not found  and we are nowhere try the actor - normally we
	// would find the actor in the location's inventory, assuming the actor is
	// somewhere. If the actor is nowhere we have to check it specifically.
	if what == nil && s.where == nil {
		if attr.FindAlias(s.actor).HasAlias(name) {
			what = s.actor
		}
	}

	// Was item to dump eventually found?
	if what == nil {
		s.msg.Actor.Send("There is nothing with alias '", name, "' to dump.")
		return
	}

	if s.cmd == "#LDUMP" {
		what.DumpToLog("#LDUMP")
		s.ok = true
		return
	}

	t := tree.Tree{}
	t.Indent, t.Offset, t.Width = 2, 13, 78

	if s.cmd == "#UDUMP" {
		t.Style = tree.StyleUnicode + "␠"
	} else {
		t.Style = tree.StyleASCII + "␠"
	}

	what.Dump(t.Branch())

	s.msg.Actor.Send(t.Render())
	s.ok = true
}
