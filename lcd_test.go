package LCD

import (
	"log"
	"fmt"
	"testing"
	"time"
)

func TestPrint(t *testing.T) {

	//var disp Display
	log.Printf("main: starting lcd")
	disp := NewLcd()
        msg := fmt.Sprintf("%v     \n%v","Ligne1",time.Now())	
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
