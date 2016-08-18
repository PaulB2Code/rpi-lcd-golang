package LCD

import (
	"log"
	"testing"
	"time"
)

func TestPrint(t *testing.T) {

	var disp Display
	log.Printf("main: starting lcd")
	disp = NewLcd()
	disp.Display("Coucou     \n     yeah Go!")
	defer func() {
		if e := recover(); e != nil {
			log.Printf("Recover: %v", e)
		}
		disp.Close()
		log.Printf("main.defer: all closed")
	}()
	time.Sleep(1 * time.Second)
}
