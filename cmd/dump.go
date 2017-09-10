// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"unsafe"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: #DUMP ( alias | <address> )
//
// The #DUMP command is only available if the server is running with the
// configuration option Debug.AllowDump set to true.
//
// The address can be any address printed using %p that points to a
// *attr.Thing - e.g. 0xc42011fab0.
func init() {
	AddHandler(Dump, "#DUMP")
}

func Dump(s *state) {
	if !config.Debug.AllowDump {
		s.msg.Actor.SendBad("#DUMP command is not available. Server not running with configuration option Debug.AllowDump=true")
		return
	}

	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("%v", p)
			s.msg.Actor.SendBad("Cannot dump ", s.input[0], ": ", err.Error())
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

	// Here be dragons - poking around in random memory locations is ill advised!
	if strings.HasPrefix(name, "0X") {

		// Change faults to panics instead so we can catch them and defer changing
		// them back again. It's easy to cause a fault with an invalid address.
		spof := debug.SetPanicOnFault(true)
		defer debug.SetPanicOnFault(spof)

		n, _ := strconv.ParseUint(name[2:], 16, 64)
		p := (*attr.Thing)(unsafe.Pointer(uintptr(n)))
		what = (*attr.Thing)(p)
	}

	// Was item to dump eventually found?
	if what == nil {
		s.msg.Actor.Send("There is nothing with alias '", name, "' to dump.")
		return
	}

	s.msg.Actor.Send(strings.Join(what.Dump(), "\n"))
	s.ok = true
}
