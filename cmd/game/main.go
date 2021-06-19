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
	"strings"
	"time"

	"code.wolfmud.org/WolfMUD.git/proc"
	"code.wolfmud.org/WolfMUD.git/world"
)

var nextPlayer chan uint64

func main() {

	nextPlayer = make(chan uint64, 1)
	nextPlayer <- 1
	rand.Seed(time.Now().UnixNano())
	world.Load()

	listener, err := net.Listen("tcp", ":4001")
	if err != nil {
		log.Printf("Error setting up listener: %s", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}
		log.Printf("Connection from: %s", conn.RemoteAddr())
		go player(conn)
	}
}

func player(conn net.Conn) {
	start := proc.WorldStart[rand.Intn(len(proc.WorldStart))]

	np := <-nextPlayer
	nextPlayer <- np + 1

	// Setup player
	player := proc.NewThing()
	player.Is = player.Is | proc.Player
	player.As[proc.Name] = "Player" + strconv.FormatUint(np, 10)
	player.As[proc.Description] = "An adventurer, just like you."
	player.As[proc.Where] = start
	player.Any[proc.Alias] = []string{"PLAYER" + strconv.FormatUint(np, 10)}
	uid := player.As[proc.UID]

	s := proc.NewState(conn, player)
	proc.BWL.Lock()
	proc.World[start].In[uid] = player
	proc.BWL.Unlock()
	s.Parse("LOOK")

	var input string
	r := bufio.NewReader(conn)
	for strings.ToUpper(input) != "QUIT\r\n" {
		input, _ = r.ReadString('\n')
		s.Parse(input)
	}
	log.Printf("Disconnect from: %s", conn.RemoteAddr())
	conn.Close()
}
