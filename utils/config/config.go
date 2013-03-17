// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package config

import (
	"time"
)

// Some sensible defaults
var (
	DataDir            = "."
	ListenAddress      = "127.0.0.1"
	ListenPort         = "4001"
	MemProfileRate int = 0
	StatsRate          = 5 * time.Minute
)
