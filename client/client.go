package client

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"time"
	"wolfmud.org/utils/broadcaster"
	"wolfmud.org/utils/parser"
	"wolfmud.org/entities/mobile/player"
)

const (
	MAX_RETRIES = 60 // Each retry is 10 seconds
)

type Interface interface {
	Start()
	AttachParser(p parser.Interface)
	DetachParser()
	Send(format string, any ...interface{})
	SendWithoutPrompt(format string, any ...interface{})
}

type Client struct {
	parser       parser.Interface
	name         string
	conn         *net.TCPConn
	bail         bool
	send         chan string
	senderWakeup chan bool
	ending       chan bool
}

func Final(c *Client) {
	log.Printf("+++ Client %s finalized +++\n", c.name)
}

func Spawn(conn *net.TCPConn, world broadcaster.Interface) {

	c := &Client{
		conn:         conn,
		send:         make(chan string, 100),
		senderWakeup: make(chan bool, 1),
		ending:       make(chan bool),
	}

	runtime.SetFinalizer(c, Final)

	c.parser = player.New(c, world)

	go c.receiver()
	go c.sender()

	<-c.ending
	<-c.ending

	c.parser.Destroy()
	c.parser = nil

	if err := c.conn.Close(); err != nil {
		log.Printf("Error closing socket for %s, %s\n", c.name, err)
	}

	log.Printf("Ended Start for %s\n", c.name)
}

func (c *Client) receiver() {

	var inBuffer [255]byte

	c.conn.SetKeepAlive(false)
	c.conn.SetLinger(0)
	idleRetrys := MAX_RETRIES

	for ; !c.bail && idleRetrys > 0; idleRetrys-- {
		c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		if b, err := c.conn.Read(inBuffer[0:254]); err != nil {
			if oe, ok := err.(*net.OpError); !ok || !oe.Timeout() {
				log.Printf("Comms error for: %s, %s\n", c.name, err)
				c.bail = true
			}
		} else {
			input := strings.TrimSpace(string(inBuffer[0:b]))
			c.parser.Parse(input)
			idleRetrys = MAX_RETRIES + 1
		}
	}

	// Connection idle and we ran out of retries?
	if idleRetrys == 0 {
		c.Send("\n\nIdle connection terminated by server.\n\nBye Bye\n\n")
		log.Printf("Closing idle connection for: %s\n", c.name)
		c.bail = true
	}

	log.Printf("Sending wakeup signal for %s\n", c.name)
	c.senderWakeup <- true

	log.Printf("Ending receiver for %s\n", c.name)
	c.ending <- true
}

func (c *Client) Send(format string, any ...interface{}) {
	c.SendWithoutPrompt(format+"\n>", any...)
}

func (c *Client) SendWithoutPrompt(format string, any ...interface{}) {
	if c.bail {
		log.Printf("oops %s dropping message %s\n", c.name, fmt.Sprintf(format, any...))
	} else {
		for i := 0; i < 10 && (cap(c.send)-len(c.send)) < 5; i++ {
			if c.bail {
				log.Printf("reschedule %s dropping message %s\n", c.name, fmt.Sprintf(format, any...))
				return
			}
			log.Printf("reschedule %s\n", c.name)
			runtime.Gosched()
		}
		c.send <- fmt.Sprintf(format, any...)
	}
}

func (c *Client) sender() {

	for !c.bail {
		select {
		case <-c.senderWakeup:
			c.bail = true
			break
		case msg := <-c.send:
			if c.bail {
				log.Printf("oops %s dropping message %s\n", c.name, msg)
			} else {
				if _, err := c.conn.Write([]byte(msg)); err != nil {
					log.Printf("Comms error for: %s, %s\n", c.name, err)
					c.bail = true
					break
				}
			}
		}
	}

	log.Printf("Ending sender for %s\n", c.name)
	c.ending <- true
}
