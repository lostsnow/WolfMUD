package main

import (
	"net"
	"time"
	"math/rand"
	crypto "crypto/rand"
)

func main() {
	println("Bot running")

	buffer := []byte{0,0,0,0,0,0}
	crypto.Reader.Read(buffer)
	seed := int64(buffer[0])
	for s := range buffer {
		if s > 0 {
			seed *= int64(s)
		}
	}
	rand.Seed(seed)
	base := (rand.Intn(9)+1) * 1000

	print("Base: ")
	println(base)
	print("Seed: ")
	println(seed)

	conn, err := net.Dial("tcp", "localhost:4001")
	if err != nil {
		// handle error
	}

	go func() {
		for {
			var buffer [255]byte
			b, _ := conn.Read(buffer[0:254])
			print(string(buffer[0:b]))
		}
	}()

	for {
		conn.Write([]byte("S"))
		time.Sleep(time.Duration(rand.Intn(2000)+base) * time.Millisecond)
		conn.Write([]byte("E"))
		time.Sleep(time.Duration(rand.Intn(2000)+base) * time.Millisecond)
		conn.Write([]byte("N"))
		time.Sleep(time.Duration(rand.Intn(2000)+base) * time.Millisecond)
		conn.Write([]byte("W"))
		time.Sleep(time.Duration(rand.Intn(2000)+base) * time.Millisecond)
	}
}
