package main

import (
	crypto "crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"time"
)

func main() {

	nbr := flag.Int("n", 10, "number of bot to launch")
	mins := flag.Int("t", 1, "number of minutes to run for")

	flag.Parse()

	fmt.Printf("Launching %d bots for %d minutes\n", *nbr, *mins)

	// Initialise random number generator with random seed
	seed, _ := crypto.Int(crypto.Reader, big.NewInt(0x7FFFFFFFFFFFFFFF))
	rand.Seed(seed.Int64())

	for i := 0; i < *nbr; i++ {
		go NewBot()
		fmt.Print(".")
	}

	fmt.Print("\nRunning...\n")

	// How long to run for?
	time.Sleep(time.Duration(*mins) * time.Minute)

}

func NewBot() {
	// Set base speed so we can have slow and fast bots
	baseSpeed := (rand.Intn(9) + 1) * 10
	steps := 8 + rand.Intn(32)

	for {

		// Connect to server
		conn, _ := net.Dial("tcp", "localhost:4001")

		// Start a reader to absorb data we get back from server
		go func() {
			for {
				var buffer [255]byte
				if b, err := conn.Read(buffer[0:254]); err != nil {
					return
				} else {
					data := buffer[0:b]
					_ = data
				}
			}
		}()

		// Script to 'execute'
		script := []string{"S", "E", "N", "W"}

		// Run script Ad infinitum with slight timing variations
		for i := 0; i < steps; i++ {
			for _, cmd := range script {
				time.Sleep(time.Duration(rand.Intn(250)+baseSpeed) * time.Millisecond)
				conn.Write([]byte(cmd))
			}
		}

		conn.Close()
	}
}
