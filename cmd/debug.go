// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"log"
	"os"
	"runtime"
	rtdebug "runtime/debug"
	"runtime/pprof"

	"code.wolfmud.org/WolfMUD.git/config"
)

// Syntax: DEBUG
// Syntax: DEBUG MEMPROF		<action>
// Syntax: DEBUG CPUPROF		<action>
// Syntax: DEBUG BLOCKPROF	<action>
// Syntax: DEBUG HEAPDUMP
// Syntax: DEBUG PANIC
//
// The DEBUG command by itself will list the running state of the memory, cpu
// and block profiles.
//
// The action can be used to start or stop a profile. To start a profile action
// can be one of: ON, START or RUN. To stop a profile action can be one of:
// OFF, STOP or END. If no action is specified the current running state of the
// profile will be displayed.
//
// The #DEBUG command is only available if the server is running with the
// configuration option Debug.AllowDebug set to true.
func init() {
	addHandler(debug{}, "#DEBUG")
}

type debug cmd

func (debug) process(s *state) {
	if !config.Debug.AllowDebug {
		s.msg.Actor.SendBad("#DEBUG command is not available. Server not running with configuration option Debug.AllowDebug=true")
		return
	}

	// If no sub-command given list all profile running states
	if len(s.words) == 0 {
		for _, p := range profiles {
			p.startStopStatus(s)
		}
		s.ok = true
		return
	}

	// If sub-command is a profile start or stop it
	if p, ok := profiles[s.words[0]]; ok {
		p.startStopStatus(s)
		s.ok = true
		return
	}

	// Check for other sub-commands
	switch s.words[0] {
	case "HEAPDUMP":
		heapDump(s)
	case "PANIC":
		log.Printf("#DEBUG: panic forced")
		panic("#DEBUG: panic forced")
	default:
		s.msg.Actor.SendBad("Unknown debug sub-command: ", s.words[0])
	}

	s.ok = true
}

// profile represents the name and running state of a profile and provides
// methods to start and stop the profile.
type profile struct {
	name    string
	running bool
	start   func() error
	stop    func() error
}

// profiles is a map of available profiles.
var profiles = map[string]*profile{
	"MEMPROF":   {"Memory profile", false, memStart, memStop},
	"CPUPROF":   {"CPU profile", false, cpuStart, cpuStop},
	"BLOCKPROF": {"Block profile", false, blockStart, blockStop},
}

// startStopStatus will start, stop or display the status of a profile. The
// action to take is specified via state.words[1].
//
// To start a profile state.words[1] can be one of: ON, START, RUN
//
// To stop a profile state.words[1] can be one of: OFF, STOP, END
//
// If state.words[1] is not provided the current status is displayed.
func (p *profile) startStopStatus(s *state) {

	// If we don't have a start/atop action report current state
	if len(s.words) < 2 {
		if p.running {
			s.msg.Actor.SendGood(p.name, " is running.")
			return
		}
		s.msg.Actor.SendInfo(p.name, " is stopped.")
		return
	}

	switch s.words[1] {
	case "ON", "START", "RUN":
		if p.running {
			s.msg.Actor.SendBad(p.name, " already running.")
			return
		}
		if p.start() != nil {
			s.msg.Actor.SendInfo(p.name, " not started, see log for details.")
			return
		}
		p.running = true
		s.msg.Actor.SendInfo(p.name, " started.")
		log.Printf("#DEBUG: %s started", p.name)
	case "OFF", "STOP", "END":
		if !p.running {
			s.msg.Actor.SendBad(p.name, " not running.")
			return
		}
		if p.stop() != nil {
			s.msg.Actor.SendInfo("Error stopping ", p.name, ", see log for details.")
			return
		}
		p.running = false
		s.msg.Actor.SendInfo(p.name, " stopped.")
		log.Printf("#DEBUG: %s stopped", p.name)
	default:
		s.msg.Actor.SendBad(p.name, ", invalid action: ", s.words[1])
	}
}

// memStart starts a memory profile.
func memStart() error {
	runtime.MemProfileRate = 1
	return nil
}

// memStop stops a memory profile.
func memStop() error {
	f, err := newFile("memprof")
	if err != nil {
		return err
	}
	pprof.WriteHeapProfile(f)
	f.Close()
	runtime.MemProfileRate = 0
	return nil
}

// cpuStart starts a cpu profile.
func cpuStart() error {
	f, err := newFile("cpuprof")
	if err != nil {
		return err
	}
	pprof.StartCPUProfile(f)
	return nil
}

// cpuStop stops a cpu profile.
func cpuStop() error {
	pprof.StopCPUProfile()
	return nil
}

// blockStart starts a block profile.
func blockStart() error {
	runtime.SetBlockProfileRate(1)
	return nil
}

// blockStop stops a block profile.
func blockStop() error {
	f, err := newFile("blockprof")
	if err != nil {
		return err
	}
	pprof.Lookup("block").WriteTo(f, 0)
	f.Close()
	runtime.SetBlockProfileRate(0)
	return nil
}

// heapDump produces a heapdump file.
func heapDump(s *state) {
	if f, err := newFile("heapdump"); err != nil {
		s.msg.Actor.SendBad("Heap dump not written, see log for details.")
	} else {
		rtdebug.WriteHeapDump(f.Fd())
		f.Close()
		s.msg.Actor.SendInfo("Heap dumped.")
		log.Printf("#DEBUG: heap dumped")
	}
}

// newFile is a helper to create a new file consistently
func newFile(name string) (f *os.File, err error) {

	// Can we create the file?
	if f, err = os.Create(name); err != nil {
		log.Printf("#DEBUG cannot create %s: %s", name, err)
		return nil, err
	}

	// Can we set the correct permissions?
	if config.Server.SetPermissions {
		if err = f.Chmod(0660); err != nil {
			f.Close()
			log.Printf("#DEBUG cannot chmod %s: %s", name, err)
			return nil, err
		}
	}
	return f, err
}
