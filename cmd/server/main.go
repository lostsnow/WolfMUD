// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// server is the main WolfMUD game server.
package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"code.wolfmud.org/WolfMUD.git/client"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/mailbox"
	"code.wolfmud.org/WolfMUD.git/quota"
	"code.wolfmud.org/WolfMUD.git/stats"
	"code.wolfmud.org/WolfMUD.git/text"
	"code.wolfmud.org/WolfMUD.git/world"
)

type pkgConfig struct {
	port       string
	host       string
	maxPlayers int
}

// cfg setup by Config and should be treated as immutable and not changed.
var cfg pkgConfig

// Config sets up package configuration for settings that can't be constants.
// It should be called by main, only once, before anything else starts. Once
// the configuration is set it should be treated as immutable an not changed.
func Config(c config.Config) {

	cfg = pkgConfig{
		host:       c.Server.Host,
		port:       c.Server.Port,
		maxPlayers: c.Server.MaxPlayers,
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
		if !c.Debug.LongLog {
			log.SetFlags(log.LstdFlags | log.LUTC)
			log.Printf("Switching to short log format")
		}

		stats.Config(c)
		core.Config(c)
		world.Config(c)
		quota.Config(c, time.Now)
		client.Config(c)
	}

	stats.Start()

	// Stop the world while we are building it
	core.BWL.Lock()
	core.RegisterCommandHandlers()
	world.Load()
	core.BWL.Unlock()

	quota.Status()

	server := net.JoinHostPort(cfg.host, cfg.port)
	addr, _ := net.ResolveTCPAddr("tcp", server)
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("Error setting up listener: %s", err)
		return
	}

	log.Printf("Accepting connections on: %s (max players: %d)",
		addr, cfg.maxPlayers)

	var (
		conn *net.TCPConn
		ip   string
	)

	for {
		conn, err = listener.AcceptTCP()
		if err == nil {
			ip, _, err = net.SplitHostPort(conn.RemoteAddr().String())
		}
		switch {
		case err != nil:
			log.Printf("Error accepting connection: %s", err)
		case !quota.Accept(ip):
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
