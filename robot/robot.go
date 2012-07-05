package main

import (
	"flag"
	"fmt"
	"io"
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
	rand.Seed(time.Now().UnixNano())

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

	launched <- true
	var buffer [255]byte

	for {
		baseSpeed := ((rand.Intn(10) + 1) * 1000) + 1000
		steps := 250 + rand.Intn(250)

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
		script := "SENEEWNENSEEWWWWNSWNNNWDUESSSWWWNNNSWWESSSSNNNNESSEEEEESSEWSWSNWSNEEESNEEWWWNNWW"

		// Run script Ad infinitum with slight timing variations
		for stepsToTake := steps; stepsToTake > 0; {
			for _, cmd := range script {
				stepsToTake--
				if stepsToTake == 0 {
					break
				}
				runtime.Gosched()
				time.Sleep(time.Duration((rand.Intn(10)*1000)+baseSpeed) * time.Millisecond)
				io.WriteString(conn, string(cmd))
				io.WriteString(conn, "\r\n")
			}
		}
		if rand.Intn(100) < 95 {
			runtime.Gosched()
			time.Sleep(time.Duration(rand.Intn(500)+baseSpeed) * time.Millisecond)
			io.WriteString(conn, "QUIT\r\n")
			time.Sleep(5 * time.Second)
		}
		conn.Close()
	}
}
