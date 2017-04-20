// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/config"

	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
)

// Syntax: DEBUG MEMPROF		( ( ON | START ) | ( OFF | STOP ) )
// Syntax: DEBUG CPUPROF		( ( ON | START ) | ( OFF | STOP ) )
// Syntax: DEBUG BLOCKPROF	( ( ON | START ) | ( OFF | STOP ) )
// Syntax: DEBUG HEAPDUMP
// Syntax: DEBUG PANIC
//
// The #DEBUG command is only available if the server is running with the
// configuration option Debug.AllowDebug set to true.
func init() {
	AddHandler(Debug, "#DEBUG")
}

func Debug(s *state) {
	if !config.Debug.AllowDebug {
		s.msg.Actor.SendBad("#DEBUG command is not available. Server not running with configuration option Debug.AllowDebug=true")
		return
	}

	switch s.words[0] {
	case "MEMPROF":
		switch s.words[1] {
		case "ON", "START":
			runtime.MemProfileRate = 1
			s.msg.Actor.SendInfo("Memory profiling turned on.")
		case "OFF", "STOP":
			f, _ := os.Create("memprof")
			pprof.WriteHeapProfile(f)
			f.Close()
			s.msg.Actor.SendInfo("Memory profiling turned off.")
		}
	case "CPUPROF":
		switch s.words[1] {
		case "ON", "START":
			f, _ := os.Create("cpuprof")
			pprof.StartCPUProfile(f)
			s.msg.Actor.SendInfo("CPU profiling turned on.")
		case "OFF", "STOP":
			pprof.StopCPUProfile()
			s.msg.Actor.SendInfo("CPU profiling turned off.")
		}
	case "BLOCKPROF":
		switch s.words[1] {
		case "ON", "START":
			runtime.SetBlockProfileRate(1)
			s.msg.Actor.SendInfo("Block profiling turned on.")
		case "OFF", "STOP":
			f, _ := os.Create("blockprof")
			pprof.Lookup("block").WriteTo(f, 0)
			f.Close()
			runtime.SetBlockProfileRate(0)
			s.msg.Actor.SendInfo("Block profiling turned off.")
		}
	case "HEAPDUMP":
		f, _ := os.Create("heapdump")
		debug.WriteHeapDump(f.Fd())
		f.Close()
		s.msg.Actor.SendInfo("Heap dumped.")
	case "PANIC":
		panic("#DEBUG force panic")
	}
	s.ok = true
}
