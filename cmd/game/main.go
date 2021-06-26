// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"bufio"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/world"
)

var nextPlayer chan uint64

func main() {

	nextPlayer = make(chan uint64, 1)
	nextPlayer <- 1
	rand.Seed(time.Now().UnixNano())
	world.Load()

	addr, _ := net.ResolveTCPAddr("tcp", ":4001")
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("Error setting up listener: %s", err)
		return
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}
		log.Printf("Connection from: %s", conn.RemoteAddr())
		go player(conn)
	}
}

func player(conn *net.TCPConn) {

	conn.SetKeepAlive(true)
	conn.SetLinger(10)
	conn.SetNoDelay(false)
	conn.SetWriteBuffer(80 * 24)
	conn.SetReadBuffer(80)

	start := core.WorldStart[rand.Intn(len(core.WorldStart))]

	np := <-nextPlayer
	nextPlayer <- np + 1

	// Setup player
	player := core.NewThing()
	player.Is = player.Is | core.Player
	player.As[core.Name] = "Player" + strconv.FormatUint(np, 10)
	player.As[core.Description] = "An adventurer, just like you."
	player.As[core.Where] = start
	player.Any[core.Alias] = []string{"PLAYER" + strconv.FormatUint(np, 10)}
	uid := player.As[core.UID]

	s := core.NewState(conn, player)
	core.BWL.Lock()
	core.World[start].In[uid] = player
	core.BWL.Unlock()
	cmd := s.Parse("LOOK")

	var input string
	r := bufio.NewReader(conn)
	for cmd != "QUIT" {
		input, _ = r.ReadString('\n')
		cmd = s.Parse(input)
	}
	log.Printf("Disconnect from: %s", conn.RemoteAddr())
	conn.Close()
}
