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
	"code.wolfmud.org/WolfMUD.git/text"
	"code.wolfmud.org/WolfMUD.git/world"
)

var nextPlayer chan uint64

func main() {

	nextPlayer = make(chan uint64, 1)
	nextPlayer <- 1
	rand.Seed(time.Now().UnixNano())

	core.RegisterCommandHandlers()
	world.Load()

	addr, _ := net.ResolveTCPAddr("tcp", ":4001")
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("Error setting up listener: %s", err)
		return
	}

	log.Printf("Accepting connections on: %s", addr)
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
	player.As[core.DynamicAlias] = "PLAYER"
	player.Any[core.Alias] = []string{"PLAYER" + strconv.FormatUint(np, 10)}
	player.Any[core.Body] = []string{
		"HEAD",
		"FACE", "EAR", "EYE", "NOSE", "EYE", "EAR",
		"MOUTH", "UPPER_LIP", "LOWER_LIP",
		"NECK",
		"SHOULDER", "UPPER_ARM", "ELBOW", "LOWER_ARM", "WRIST",
		"HAND", "FINGER", "FINGER", "FINGER", "FINGER", "THUMB",
		"SHOULDER", "UPPER_ARM", "ELBOW", "LOWER_ARM", "WRIST",
		"HAND", "FINGER", "FINGER", "FINGER", "FINGER", "THUMB",
		"BACK", "CHEST",
		"WAIST", "PELVIS",
		"UPPER_LEG", "KNEE", "LOWER_LEG", "ANKLE", "FOOT",
		"UPPER_LEG", "KNEE", "LOWER_LEG", "ANKLE", "FOOT",
	}
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
						buf = append(buf, text.Reset...)
						buf = append(buf, msg...)
						buf = append(buf, '\n')
					}
					buf = append(buf, text.Magenta...)
					buf = append(buf, '>')
					conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
					if _, err = conn.Write(text.Fold(buf, 80)); err != nil {
						log.Printf("[%s] conn error: %s", uid, err)
						<-errState
						errState <- err
						mailbox.Delete(uid)
						conn.CloseWrite()
					}
				}
				if !ok {
					mailbox.Delete(uid)
					conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
					conn.Write([]byte(text.Reset))
					conn.CloseWrite()
					return
				}
			}
		}
	}()

	core.BWL.Lock()
	player.Ref[core.Where] = core.WorldStart[rand.Intn(len(core.WorldStart))]
	player.Ref[core.Where].Who[uid] = player
	core.BWL.Unlock()
	cmd := s.Parse("$POOF")

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
			continue
		}
		if err != nil {
			<-errState
			errState <- err
			log.Printf("[%s] read error: %s", uid, err)
			cmd = s.Parse("QUIT")
			break
		}
		cmd = s.Parse(clean(input))
		runtime.Gosched()
	}

	mailbox.Delete(uid)
	player.Free()

	log.Printf("[%s] disconnect from: %s", uid, conn.RemoteAddr())
	conn.CloseRead()
}

// clean incoming data. Invalid runes or C0 and C1 control codes are dropped.
// An exception in the C0 control code is backspace ('\b', ASCII 0x08) which
// will erase the previous rune. This can occur when the player's Telnet client
// does not support line editing.
func clean(in string) string {

	o := make([]rune, len(in)) // oversize due to len = bytes
	i := 0
	for _, v := range in {
		switch {
		case v == '\uFFFD':
			// drop invalid runes
		case v == '\b' && i > 0:
			i--
		case v <= 0x1F:
			// drop C0 control codes
		case 0x80 <= v && v <= 0x9F:
			// drop C1 control codes
		default:
			o[i] = v
			i++
		}
	}

	return string(o[:i])
}
