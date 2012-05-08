package entities

import (
	"fmt"
	"net"
	"strings"
)

type Client interface {
	Start()
	AttachPlayer(p Player)
	DetachPlayer()
	SendResponse(format string, any ...interface{})
	SendPlain(format string, any ...interface{})
}

type client struct {
	player      Player
	name        string
	conn        net.Conn
	connBadLock chan bool
	connBad     bool
}

func NewClient(conn net.Conn) Client {
	return &client{
		conn:        conn,
		connBadLock: make(chan bool, 1),
	}
}

func (c *client) AttachPlayer(p Player) {
	c.player = p
	c.name = p.Name()
}

func (c *client) DetachPlayer() {
	c.player = nil
}

func (c *client) setConnBad(err error) {
	c.connBad = true

	fmt.Printf("client.setConnBad: Comms error for: %s, %s\n", c.name, err)
	if err := c.conn.Close(); err != nil {
		fmt.Printf("client.setConnBad: Error closing socket for %s, %s\n", c.name, err)
	}
	return
}

func (c *client) Start() {

	var inBuffer [255]byte

	for {
		c.connBadLock <- true
		if c.connBad {
			<-c.connBadLock
			break
		}
		if b, err := c.conn.Read(inBuffer[0:254]); err != nil {
			c.setConnBad(err)
			<-c.connBadLock
			p := c.player
			p.DetachClient()
			p.Destroy()
			break
		} else {
			<-c.connBadLock
			input := strings.TrimSpace(string(inBuffer[0:b]))
			c.player.Parse(input)
		}
	}

	fmt.Printf("client.Start: Ending for %s\n", c.name)

}

func (c *client) SendResponse(format string, any ...interface{}) {
	c.SendPlain("\n"+format+"\n>", any...)
}

func (c *client) SendPlain(format string, any ...interface{}) {
	c.connBadLock <- true
	if c.connBad {
		<-c.connBadLock
		return
	}

	s := fmt.Sprintf(format, any...)
	if _, err := c.conn.Write([]byte(s)); err != nil {
		c.setConnBad(err)
		<-c.connBadLock
		p := c.player
		p.DetachClient()
		p.Destroy()
		fmt.Printf("client.Send: Comms error for: %s, %s\n", c.name, err)
	} else {
		<-c.connBadLock
	}
	return
}
