package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"runtime"
	"time"
)

func main() {

	nbr := flag.Int("n", 10, "number of bot to launch")
	mins := flag.Int("t", 1, "number of minutes to run for")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("Launching %d bots for %d minutes\n", *nbr, *mins)

	// Initialise random number generator with random seed
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < *nbr; i++ {
		go NewBot(i)
		runtime.Gosched()
		time.Sleep(10000 * time.Microsecond)
	}

	fmt.Print("\nRunning...\n")

	// How long to run for?
	time.Sleep(time.Duration(*mins) * time.Minute)

}

func NewBot(bot int) {
	// Set base speed so we can have slow and fast bots

	var buffer [255]byte

	for {
		baseSpeed := ((rand.Intn(10) + 1) * 1000) + 1000
		steps := 250 + rand.Intn(250)

		// Connect to server
		conn, err := net.DialTimeout("tcp", "127.0.0.1:4001", time.Minute)
		if conn == nil {
			log.Printf("[%d] Connect error: %s\n", bot, err)
			time.Sleep(time.Second)
			continue
		}

		// Start a reader to absorb data we get back from server
		go func() {
			for {
				runtime.Gosched()
				conn.SetReadDeadline(time.Now().Add(time.Minute))
				if _, err := conn.Read(buffer[0:254]); err != nil {
					log.Printf("[%d] Read error: %s\n", bot, err)
					conn.Close()
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
				conn.SetWriteDeadline(time.Now().Add(time.Minute))
				if _, err := conn.Write([]byte(string(cmd) + "\r\n")); err != nil {
					log.Printf("[%d] Write error: %s\n", bot, err)
					conn.Close()
					continue
				}
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
