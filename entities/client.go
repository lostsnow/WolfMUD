package entities

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Client interface {
	Start()
	AttachPlayer(p Player)
	DetachPlayer()
	SendResponse(format string, any ...interface{})
	SendPlain(format string, any ...interface{})
}

type client struct {
	player       Player
	name         string
	conn         *net.TCPConn
	sendFail     bool
	receiveFail  bool
	send         chan string
	senderWakeup chan bool
}

func NewClient(conn *net.TCPConn) Client {
	return &client{
		conn:         conn,
		send:         make(chan string, 100),
		senderWakeup: make(chan bool, 1),
	}
}

func (c *client) AttachPlayer(p Player) {
	c.player = p
	c.name = p.Name()
}

func (c *client) DetachPlayer() {
	c.player = nil
}

func (c *client) Start() {
	go c.receiver()
	go c.sender()
}

func (c *client) receiver() {

	var inBuffer [255]byte

	c.conn.SetKeepAlive(false)
	c.conn.SetLinger(0)

	for {
		if c.receiveFail || c.sendFail {
			break
		}
		c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if b, err := c.conn.Read(inBuffer[0:254]); err != nil {
			if oe, ok := err.(*net.OpError); ok && oe.Timeout() {
				c.SendPlain("\n\n +++ Connection Idle for X minutes, Logged out by Server +++\n\nBye Bye\n\n")
				fmt.Printf("client.receiver: Closing idle connection for: %s\n", c.name)
			} else {
				c.receiveFail = true
				fmt.Printf("client.receiver: Comms error for: %s, %s, %T\n", c.name, err, err)
			}
			if err := c.conn.Close(); err != nil {
				fmt.Printf("client.receiver: Error closing socket for %s, %s\n", c.name, err)
			}
			break
		} else {
			input := strings.TrimSpace(string(inBuffer[0:b]))
			c.player.Parse(input)
		}
	}

	p := c.player
	p.DetachClient()
	p.Destroy()
	c.senderWakeup <- true

	fmt.Printf("client.receiver: Ending for %s\n", c.name)
}

func (c *client) SendResponse(format string, any ...interface{}) {
	msg := fmt.Sprintf("\n"+format+"\n>", any...)
	if c.sendFail || c.receiveFail {
		//fmt.Printf("client.SendResponse: oops %s dropping message %s\n", c.name, msg)
	} else {
		//fmt.Printf("client.SendResponse: %s adding to queue %d\n", c.name, len(c.send))
		c.send <- msg
	}
}

func (c *client) SendPlain(format string, any ...interface{}) {
	if c.sendFail || c.receiveFail {
		//fmt.Printf("client.SendPlain: oops %s dropping message %s\n", c.name, fmt.Sprintf(format, any...))
	} else {
		c.send <- fmt.Sprintf(format, any...)
	}
}

func (c *client) sender() {

	for {
		if c.receiveFail || c.sendFail {
			break
		}
		select {
		case <-c.senderWakeup:

		case msg := <-c.send:
			if c.sendFail {
				//fmt.Printf("client.sender: oops %s dropping message %s\n", c.name, msg)
			} else {
				if _, err := c.conn.Write([]byte(msg)); err != nil {
					c.sendFail = true
					fmt.Printf("client.sender: Comms error for: %s, %s\n", c.name, err)
					//if err := c.conn.Close(); err != nil {
					//	fmt.Printf("client.sender: Error closing socket for %s, %s\n", c.name, err)
					//}
					break
				}
			}
		}
	}

	fmt.Printf("client.sender: Ending for %s\n", c.name)
}
