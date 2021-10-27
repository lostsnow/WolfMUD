// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package client handles all network communication to and from a player.
package client

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/mailbox"
	"code.wolfmud.org/WolfMUD.git/text"
)

const (
	frontendTimeout = 5 * time.Minute
	ingameTimeout   = 60 * time.Minute
)

var idleDisconnect = text.Bad + "\nIdle connection terminated by server.\n" +
	text.Reset

type client struct {
	*core.Thing
	*net.TCPConn
	err   chan error
	queue <-chan string
	quit  chan struct{}
	uid   string // Can't touch c.As[core.UID] when not under BWL
}

func New(conn *net.TCPConn) {
	c := &client{
		Thing:   core.NewThing(),
		TCPConn: conn,
		err:     make(chan error, 1),
		quit:    make(chan struct{}, 1),
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
	if c.frontend() {
		c.enterWorld()
		c.receive()
	}

	if err := c.error(); err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			mailbox.Send(c.uid, true, idleDisconnect)
		}
		log.Printf("[%s] client error: %s", c.uid, err)
	}

	log.Printf("[%s] disconnect from: %s", c.uid, c.RemoteAddr())
	mailbox.Send(c.uid, true, text.Good+"\n\nBye bye!\n\n"+text.Reset)

	mailbox.Delete(c.uid)
	<-c.quit
	if c.As[core.Account] != "" {
		accountsMux.Lock()
		delete(accounts, c.As[core.Account])
		accountsMux.Unlock()
	}
	c.Free()
	c.Close()
}

func (c *client) receive() {

	s := core.NewState(c.Thing)
	cmd := s.Parse("$POOF")

	var input string
	var err error
	r := bufio.NewReaderSize(c, 80)
	for cmd != "QUIT" && c.error() == nil {
		c.SetReadDeadline(time.Now().Add(ingameTimeout))
		if input, err = r.ReadString('\n'); err != nil {
			c.setError(err)
			break
		}
		if len(c.queue) > 10 {
			continue
		}
		cmd = s.Parse(clean(input))
		runtime.Gosched()
	}

	if cmd != "QUIT" {
		cmd = s.Parse("QUIT")
	}
}

func (c *client) messenger() {
	var buf []byte

	for {
		select {
		case msg, ok := <-c.queue:
			if !ok {
				c.quit <- struct{}{}
				return
			}

			buf = buf[:0]
			if len(msg) > 0 {
				buf = append(buf, text.Reset...)
				buf = append(buf, msg...)
				buf = append(buf, '\n')
			}
			buf = append(buf, text.Magenta...)
			buf = append(buf, '>')

			c.SetWriteDeadline(time.Now().Add(10 * time.Second))
			c.Write(text.Fold(buf, 80))
		}
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
