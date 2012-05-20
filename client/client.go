package client

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"
	"wolfmud.org/utils/parser"
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
	sendFail     bool
	receiveFail  bool
	send         chan string
	senderWakeup chan bool
}

func New(conn *net.TCPConn) *Client {
	return &Client{
		conn:         conn,
		send:         make(chan string, 100),
		senderWakeup: make(chan bool, 1),
	}
}

func (c *Client) AttachParser(p parser.Interface) {
	c.parser = p
	c.name = p.Name()
}

func (c *Client) DetachParser() {
	c.parser = nil
}

func (c *Client) Start() {
	go c.receiver()
	go c.sender()
}

func (c *Client) receiver() {

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
			c.parser.Parse(input)
		}
	}

	p := c.parser
	p.DetachClient()
	c.senderWakeup <- true

	fmt.Printf("client.receiver: Ending for %s\n", c.name)
}

func (c *Client) Send(format string, any ...interface{}) {
	c.SendWithoutPrompt(format+"\n>", any...)
}

func (c *Client) SendWithoutPrompt(format string, any ...interface{}) {
	if c.sendFail || c.receiveFail {
		//fmt.Printf("client.Send: oops %s dropping message %s\n", c.name, fmt.Sprintf(format, any...))
	} else {
		for i := 0; i < 50 && (cap(c.send)-len(c.send)) < 5; i++ {
			runtime.Gosched()
		}
		c.send <- fmt.Sprintf(format, any...)
	}
}

func (c *Client) sender() {

	for {
		if c.receiveFail || c.sendFail {
			break
		}
		select {
		case <-c.senderWakeup:
			fmt.Printf("client.sender: send length %d, draining\n", len(c.send))
			if len(c.send) != 0 {
				for _ = range c.send {
				}
			}

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
