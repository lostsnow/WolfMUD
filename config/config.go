// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// TODO: Load from a config file instead of being hardcoded!
package config

import (
	"time"
)

var Server = struct {
	Host        string        // Host for server to listen on
	Port        string        // Port for server to listen on
	IdleTimeout time.Duration // Idle connection disconnect time
}{
	Host:        "127.0.0.1",
	Port:        "4001",
	IdleTimeout: 10 * time.Minute,
}

var Stats = struct {
	Rate time.Duration // Stats collection and display rate
}{
	Rate: 10 * time.Second,
}

// Inventory configuration
const (
	ReclaimFactor = 2  // is capacity > length * reclaimFactor
	ReclaimBuffer = 4  // only reclaim if gain more than reclaimBuffer
	CrowdSize     = 10 // If inventory has more player than this it's a crowd
)
