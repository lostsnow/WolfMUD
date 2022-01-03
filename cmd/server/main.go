// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// server is the main WolfMUD game server.
package main

import (
	"log"
	"math/bits"
	"math/rand"
	"net"
	"os"
	"time"

	"code.wolfmud.org/WolfMUD.git/client"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/mailbox"
	"code.wolfmud.org/WolfMUD.git/stats"
	"code.wolfmud.org/WolfMUD.git/text"
	"code.wolfmud.org/WolfMUD.git/world"
)

type pkgConfig struct {
	port        string
	host        string
	maxPlayers  int
	quotaSlots  int
	quotaMask   uint64
	quotaWindow time.Duration
}

// cfg setup by Config and should be treated as immutable and not changed.
var cfg pkgConfig

// Config sets up package configuration for settings that can't be constants.
// It should be called by main, only once, before anything else starts. Once
// the configuration is set it should be treated as immutable an not changed.
func Config(c config.Config) {

	// Max limit of 63 quota slots due to bits in uint64 for mask + 1 extra bit
	slots := c.Quota.Slots
	if slots > 63 {
		log.Printf("WARNING: Limiting Quota.Slots to 63, was %d", slots)
		slots = 63
	}

	cfg = pkgConfig{
		host:        c.Server.Host,
		port:        c.Server.Port,
		maxPlayers:  c.Server.MaxPlayers,
		quotaSlots:  slots,
		quotaMask:   uint64((1 << (slots + 1)) - 1),
		quotaWindow: c.Quota.Window,
	}
}

var serverFull = []byte(
	text.Bad +
		"\nServer too busy. Please come back in a short while.\n\n" +
		text.Reset,
)

var tooManyConnections = []byte(
	text.Bad +
		"\nToo many connection attempts, please wait before trying again.\n\n" +
		text.Reset,
)

func main() {

	// Setup global logging format
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds | log.LUTC)
	log.Printf("Server started, logging using UTC timezone.")

	rand.Seed(time.Now().UnixNano())

	// Setup configuration before doing anything else
	{
		var (
			path string
			err  error
		)

		if path = os.Getenv("WOLFMUD_DIR"); path != "" {
			log.Printf("Found enviroment variable WOLFMUD_DIR: %s", path)
		}
		c := config.Default()
		if path == "NONE" {
			log.Print("Using built-in configuration")
		} else {
			if c, err = c.Load(path); err != nil {
				log.Fatalf("Configuration error: %s", err)
			}
		}
		Config(c)
		stats.Config(c)
		core.Config(c)
		world.Config(c)
		client.Config(c)

		if !c.Debug.LongLog {
			log.SetFlags(log.LstdFlags | log.LUTC)
			log.Printf("Switching to short log format")
		}
	}

	stats.Start()

	// Stop the world while we are building it
	core.BWL.Lock()
	core.RegisterCommandHandlers()
	world.Load()
	core.BWL.Unlock()

	server := net.JoinHostPort(cfg.host, cfg.port)
	addr, _ := net.ResolveTCPAddr("tcp", server)
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("Error setting up listener: %s", err)
		return
	}

	if cfg.quotaSlots == 0 || cfg.quotaWindow == 0 {
		log.Printf("IP Quotas disabled, set Quota.Slots and Quota.Window to enable")
	} else {
		log.Printf("IP Quotas enabled, limiting to %d connections per IP address in %s", cfg.quotaSlots, cfg.quotaWindow)
	}

	log.Printf("Accepting connections on: %s (max players: %d)",
		addr, cfg.maxPlayers)

	for {
		conn, err := listener.AcceptTCP()
		switch {
		case err != nil:
			log.Printf("Error accepting connection: %s", err)
		case !quota(conn):
			conn.Write(tooManyConnections)
			conn.Close()
		case mailbox.Len() >= cfg.maxPlayers:
			conn.Write(serverFull)
			conn.Close()
		default:
			go client.New(conn).Play()
		}
	}
}

// quotaCache records connection attempts and the timestamp of the last
// connection attempt. The attempts field contains bits representing an
// interval cfg.quotaWindow apart.
var quotaCache = map[string]struct {
	when     time.Time
	attempts uint64
}{}

// Quota check per IP connection quotas. Quota will return true Quota will
// return true if there are less than Quota.Slots connections in a Quota.Window
// period else false.
//
// NOTE: There is a maximum limit of 63 for Quota.Slots
func quota(conn *net.TCPConn) (allowed bool) {
	if cfg.quotaSlots == 0 || cfg.quotaWindow == 0 {
		return true
	}

	now := time.Now()
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// Purge any expired cache entries
	expiry := now.Add(-time.Duration(cfg.quotaSlots) * cfg.quotaWindow)
	for addr, c := range quotaCache {
		if ip != addr && c.when.Before(expiry) {
			delete(quotaCache, ip)
		}
	}

	c, found := quotaCache[ip]
	if !found {
		c.when, c.attempts = now, 1
		quotaCache[ip] = c
		return true
	}

	s := int(now.Sub(c.when)/cfg.quotaWindow) + 1
	c.when, c.attempts = now, c.attempts<<s
	c.attempts++
	tries := bits.OnesCount64(c.attempts & cfg.quotaMask)
	quotaCache[ip] = c

	return tries <= cfg.quotaSlots
}
