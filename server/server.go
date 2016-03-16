// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"code.wolfmud.org/WolfMUD.git/comms"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/stats"

	"log"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	config.Load()
	stats.Start()
	comms.Listen(config.Server.Host, config.Server.Port)
}
