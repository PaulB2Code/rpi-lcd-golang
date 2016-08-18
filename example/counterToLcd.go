package main

import (
	"fmt"
	"log"
	"time"

	"github.com/PaulB2Code/rpi-lcd-golang"
)

func main() {

	log.Println("Start Counter at ", time.Now())

	disp := LCD.NewLcd()
	msg := fmt.Sprintf("%v     \n%v", "Ligne1", time.Now())
	disp.Display(msg)
	defer func() {
		if e := recover(); e != nil {
			log.Printf("Recover: %v", e)
		}
		disp.Close()
		log.Printf("main.defer: all closed")
	}()
	time.Sleep(1 * time.Second)

}
