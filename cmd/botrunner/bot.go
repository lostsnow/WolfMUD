// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

// btoi maps a boolean to an integer
var btoi = map[bool]int{false: 0, true: 1}

// commands that the bots will execute randomly. The commands are split into
// high and low frequency. With high frequency commands executed more often.
var commands = [2][]string{
	{ // High frequency commands
		"N", "NE", "E", "SE", "S", "SW", "W", "NW", "U", "D",
	},
	{ // Low frequency commands
		"SNEEZE",
		"SAY Hi!",
		"EXAMINE ANY PLAYER",
		//"WHISPER ANY PLAYER Hi!",
		"SHOUT Hello!",
		//"TELL ANY PLAYER Nice weather!",
		"HIT ANY CREATURE",
		"HIT ANY PLAYER",
	},
}

// Bot represents a simuated player.
type Bot struct {
	Id        string
	Quit      chan struct{}
	buffer    []byte
	baseSpeed time.Duration
	steps     int
	net.Conn
}

// NewBot sets up a new simulated player.
func NewBot(id string) *Bot {

	// Jitter ranges from 0 to 0.5 seconds
	jitter := time.Duration(rand.Intn(6)*100) * time.Millisecond

	b := &Bot{
		Id:        id,
		Quit:      make(chan struct{}, 1),
		buffer:    make([]byte, 512),
		baseSpeed: 1500*time.Millisecond + jitter, // ranges from 1.5 to 2 seconds
		steps:     rand.Intn(2500) + 500,          // command count before quitting
	}
	return b
}

// Runner causes a bot to connect to the server and then execute random commands.
func (b *Bot) Runner(host, port string) {
	var err error
	for {
		if err = b.connect(host, port); err == nil {
			if !b.do() {
				return
			}
		}
		// If we get an error (usually server full or account already logged in)
		// sleep and additional 20 seconds to give error time to clear before
		// retrying.
		if err != nil {
			log.Printf("[%s] Error: %s", b.Id, err)
			time.Sleep(20 * time.Second)
		}
		time.Sleep(10 * time.Second)
	}
}

// connect to the server and log in the simulated player.
func (b *Bot) connect(host, port string) (err error) {
	server := net.JoinHostPort(host, port)
	b.Conn, err = net.DialTimeout("tcp", server, time.Minute)
	if err != nil {
		return err
	}

	for _, f := range []struct {
		op   func(string) error
		data string
	}{
		{b.recv, "...server:\r\n\x1b[35m>"},
		{b.send, b.Id},
		{b.recv, "\x1b[0mEnter the password...to cancel:\r\n\x1b[35m>"},
		{b.send, b.Id},
		{b.recv, "\x1b[0m\x1b[32mWelcome back..."},
	} {
		if err = f.op(f.data); err != nil {
			b.Close()
			return
		}
	}

	return
}

func (b *Bot) do() bool {
	go b.discard()

	for step := 0; step < b.steps; step++ {

		// Check if bot should be quitting
		select {
		case <-b.Quit:
			b.send("QUIT")
			time.Sleep(time.Second)
			b.Close()
			return false
		default:
		}

		// baseSpeed adjustment ±500ms
		adj := time.Duration((rand.Intn(11)-5)*100) * time.Millisecond

		// Pick a high (90% chance) or low (10% chance) frequence table, then pick
		// a random action to perform from chosen table
		freq := commands[btoi[rand.Intn(100) > 89]]
		action := freq[rand.Intn(len(freq))]

		// Execute command
		if err := b.send(action); err != nil {
			log.Printf("[%s] %s", b.Id, err)
			return true
		}
		time.Sleep(b.baseSpeed + adj)
	}

	// Log off from the server and close connection
	b.send("QUIT")
	b.send("0")
	log.Printf("[%s] Run finished", b.Id)
	time.Sleep(time.Second)
	b.Close()

	return true
}

func (b *Bot) discard() {
	buf := [512]byte{}
	for {
		b.SetReadDeadline(time.Now().Add(60 * time.Second))
		if _, err := b.Read(buf[0:511]); err != nil {
			if err, ok := err.(*net.OpError); !ok || !err.Timeout() {
				return
			}
		}
	}
}

func (b *Bot) send(data string) error {
	b.SetWriteDeadline(time.Now().Add(60 * time.Second))
	_, err := b.Write([]byte(data + "\r\n"))
	return err
}

func (b *Bot) recv(data string) error {
	b.SetReadDeadline(time.Now().Add(60 * time.Second))
	x, err := b.Read(b.buffer[0:512])
	b.buffer = b.buffer[0:x]

	if err != nil {
		b.buffer = b.buffer[:0]
		return err
	}

	if !b.ifeq(data) {
		return fmt.Errorf("[%s] Unexpected response: %q", b.Id, b.buffer)
	}
	return nil
}

func (b *Bot) ifeq(data string) bool {
	part := bytes.Split([]byte(data), []byte("..."))
	start, end := part[0], part[1]

	if len(start) > 0 && !bytes.HasPrefix(b.buffer, start) {
		return false
	}
	if len(end) > 0 && !bytes.HasSuffix(b.buffer, end) {
		return false
	}

	return true
}
