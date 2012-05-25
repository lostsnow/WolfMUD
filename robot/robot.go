package main

import (
	crypto "crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"runtime"
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

	launched := make(chan bool, 1)

	for i := 0; i < *nbr; i++ {
		go NewBot(launched)
		runtime.Gosched()
		<-launched
		fmt.Print(".")
	}

	fmt.Print("\nRunning...\n")

	// How long to run for?
	time.Sleep(time.Duration(*mins) * time.Minute)

}

func NewBot(launched chan bool) {
	// Set base speed so we can have slow and fast bots
	baseSpeed := ((rand.Intn(5) + 1) * 100) + 1000
	steps := 8 + rand.Intn(8)

	launched <- true
	var buffer [255]byte

	for {

		// Connect to server
		conn, _ := net.Dial("tcp", "localhost:4001")

		// Start a reader to absorb data we get back from server
		go func() {
			for {
				runtime.Gosched()
				if _, err := conn.Read(buffer[0:254]); err != nil {
					return
				}
			}
		}()

		// Script to 'execute'
		script := []string{"S", "E", "N", "E", "E", "W", "S", "E", "W", "S", "N", "N", "W", "W"}

		// Run script Ad infinitum with slight timing variations
		for i := 0; i < steps; i++ {
			for _, cmd := range script {
				runtime.Gosched()
				time.Sleep(time.Duration(rand.Intn(500)+baseSpeed) * time.Millisecond)
				conn.Write([]byte(cmd))
			}
		}

		conn.Close()
	}
}
