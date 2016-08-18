package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/PaulB2Code/rpi-lcd-golang"
)

func main() {

	log.Println("Start Counter at ", time.Now())

	disp := LCD.NewLcd()
	msg := fmt.Sprintf("Start Couting\n In One Second")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			log.Println("\nClose Display.\n")
			disp.Close()
			os.Exit(0)
		}
	}()

	i := 0
	for {
		disp.Display(msg)
		time.Sleep(1 * time.Second)
		i++
		msg = fmt.Sprintf("Count %v at \n%v", i, time.Now().Format("02 Jan 15:04:05"))
	}
}
