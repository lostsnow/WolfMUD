package entities

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"
)

type Client interface {
	Start()
	AttachPlayer(p Player)
	DetachPlayer()
	Send(format string, any ...interface{})
	SendWithoutPrompt(format string, any ...interface{})
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
		c.conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
		if b, err := c.conn.Read(inBuffer[0:254]); err != nil {
			if oe, ok := err.(*net.OpError); ok && oe.Timeout() {
				c.Send("\n\nIdle connection terminated by server.\n\nBye Bye\n\n")
				fmt.Printf("client.receiver: Closing idle connection for: %s\n", c.name)
			} else {
				c.receiveFail = true
				fmt.Printf("client.receiver: Comms error for: %s, %s\n", c.name, err)
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

func (c *client) Send(format string, any ...interface{}) {
	c.SendWithoutPrompt("\n"+format+"\n>", any...)
}

func (c *client) SendWithoutPrompt(format string, any ...interface{}) {
	msg := fmt.Sprintf(format, any...)
	if c.sendFail || c.receiveFail {
		//fmt.Printf("client.Send: oops %s dropping message %s\n", c.name, fmt.Sprintf(format, any...))
	} else {
		for i := 0; i < 50 && (cap(c.send)-len(c.send)) < 5; i++ {
			runtime.Gosched()
		}
		c.send <- msg
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
