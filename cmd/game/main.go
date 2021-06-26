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
	"runtime"
	"strconv"
	"time"

	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/mailbox"
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
		go player(conn)
	}
}

func player(conn *net.TCPConn) {

	conn.SetKeepAlive(true)
	conn.SetLinger(10)
	conn.SetNoDelay(false)
	conn.SetWriteBuffer(80 * 24)
	conn.SetReadBuffer(80)

	np := <-nextPlayer
	nextPlayer <- np + 1

	// Setup player
	player := core.NewThing()
	player.Is = player.Is | core.Player
	player.As[core.Name] = "Player" + strconv.FormatUint(np, 10)
	player.As[core.Description] = "An adventurer, just like you."
	player.As[core.Where] = core.WorldStart[rand.Intn(len(core.WorldStart))]
	player.Any[core.Alias] = []string{"PLAYER" + strconv.FormatUint(np, 10)}
	uid := player.As[core.UID]

	log.Printf("[%s] connection from: %s", uid, conn.RemoteAddr())

	q := mailbox.Add(uid)
	s := core.NewState(player)

	errState := make(chan error, 1)
	errState <- nil

	go func() {
		var err error
		var buf []byte
		for {
			select {
			case msg, ok := <-q:
				err = <-errState
				errState <- err

				if ok && err == nil {
					buf = buf[:0]
					if len(msg) > 0 {
						buf = append(buf, msg...)
						buf = append(buf, '\n')
					}
					buf = append(buf, '>')
					conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
					if _, err = conn.Write(buf); err != nil {
						log.Printf("[%s] conn error: %s", uid, err)
						<-errState
						errState <- err
						mailbox.Delete(uid)
						conn.CloseWrite()
						log.Printf("[%s] mailbox deleted, error", uid)
					}
				} else {
					if ok {
						log.Printf("[%s] discarding: %q", uid, msg)
					}
				}
				if !ok {
					mailbox.Delete(uid)
					conn.CloseWrite()
					log.Printf("[%s] mailbox deleted, channel closed", uid)
					return
				}
			}
		}
	}()

	core.BWL.Lock()
	core.World[player.As[core.Where]].In[uid] = player
	core.BWL.Unlock()
	cmd := s.Parse("LOOK")

	var input string
	var err error
	r := bufio.NewReader(conn)
	for cmd != "QUIT" {
		err = <-errState
		if err != nil {
			errState <- err
			log.Printf("[%s] pre-read error: %s", uid, err)
			cmd = s.Parse("QUIT")
			break
		}
		errState <- err
		conn.SetReadDeadline(time.Now().Add(60 * time.Minute))
		input, err = r.ReadString('\n')
		if len(q) > 10 {
			log.Printf("[%s] command dropped: %q", uid, input)
			continue
		}
		if err != nil {
			<-errState
			errState <- err
			log.Printf("[%s] read error: %s", uid, err)
			cmd = s.Parse("QUIT")
			break
		}
		cmd = s.Parse(input)
		runtime.Gosched()
	}

	mailbox.Delete(uid)
	player.Free()

	log.Printf("[%s] disconnect from: %s", uid, conn.RemoteAddr())
	conn.CloseRead()
}
