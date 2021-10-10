// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package client handles all network communication to and from a player.
package client

import (
	"bufio"
	"log"
	"math/rand"
	"net"
	"runtime"
	"time"

	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/mailbox"
	"code.wolfmud.org/WolfMUD.git/text"
)

type client struct {
	*core.Thing
	*net.TCPConn
	err   chan error
	queue <-chan string
	uid   string // Can't touch c.As[core.UID] when not under BWL
}

func New(conn *net.TCPConn) {
	c := &client{
		Thing:   core.NewThing(),
		TCPConn: conn,
		err:     make(chan error, 1),
	}

	c.err <- nil

	c.SetKeepAlive(true)
	c.SetLinger(10)
	c.SetNoDelay(false)
	c.SetWriteBuffer(80 * 24)
	c.SetReadBuffer(80)

	c.queue = mailbox.Add(c.As[core.UID])
	c.uid = c.As[core.UID]

	log.Printf("[%s] connection from: %s", c.uid, c.RemoteAddr())

	go c.messenger()
	c.receive()
}

func (c *client) receive() {

	s := core.NewState(c.Thing)

	c.createPlayer()
	core.BWL.Lock()
	c.Ref[core.Where] = core.WorldStart[rand.Intn(len(core.WorldStart))]
	c.Ref[core.Where].Who[c.uid] = c.Thing
	core.BWL.Unlock()
	cmd := s.Parse("$POOF")

	var input string
	var err error
	r := bufio.NewReaderSize(c, 80)
	for cmd != "QUIT" && c.error() == nil {
		c.SetReadDeadline(time.Now().Add(60 * time.Minute))
		if input, err = r.ReadString('\n'); err != nil {
			log.Printf("[%s] read error: %s", c.uid, err)
			c.setError(err)
			cmd = s.Parse("QUIT")
			break
		}
		if len(c.queue) > 10 {
			continue
		}
		cmd = s.Parse(clean(input))
		runtime.Gosched()
	}

	mailbox.Delete(c.uid)
	c.Free()

	log.Printf("[%s] disconnect from: %s", c.uid, c.RemoteAddr())
	c.CloseRead()
}

func (c *client) messenger() {
	var err error
	var buf []byte

	for {
		select {
		case msg, ok := <-c.queue:
			if ok && c.error() == nil {
				buf = buf[:0]
				if len(msg) > 0 {
					buf = append(buf, text.Reset...)
					buf = append(buf, msg...)
					buf = append(buf, '\n')
				}
				buf = append(buf, text.Magenta...)
				buf = append(buf, '>')
				c.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if _, err = c.Write(text.Fold(buf, 80)); err != nil {
					log.Printf("[%s] conn error: %s", c.uid, err)
					c.setError(err)
					mailbox.Delete(c.uid)
					c.CloseWrite()
				}
			}
			if !ok {
				mailbox.Delete(c.uid)
				c.SetWriteDeadline(time.Now().Add(10 * time.Second))
				c.Write([]byte(text.Reset))
				c.CloseWrite()
				return
			}
		}
	}
}

func (c *client) createPlayer() {
	c.Is = c.Is | core.Player
	c.As[core.Name] = "Player"
	c.As[core.UName] = c.As[core.Name]
	c.As[core.TheName] = c.As[core.Name]
	c.As[core.UTheName] = c.As[core.Name]
	c.As[core.Description] = "An adventurer, just like you."
	c.As[core.DynamicAlias] = "PLAYER"
	c.Any[core.Alias] = []string{"PLAYER"}
	c.Any[core.Body] = []string{
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
}

// error returns the first error raised or nil if there have been no errors.
func (c *client) error() error {
	e := <-c.err
	c.err <- e
	return e
}

// setError records the first error raised only, which can be retrieved by
// calling error.
func (c *client) setError(err error) {
	e := <-c.err
	if e == nil {
		e = err
	}
	c.err <- e
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
