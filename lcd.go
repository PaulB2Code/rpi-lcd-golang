package LCD

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/stianeikeland/go-rpio"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

//Display (output) interface
type Display interface {
	Display(string)
	Close()
}

const (
	//Timing constants
	ePulse = 1 * time.Microsecond
	eDelay = 70 * time.Microsecond

	//BCM http://pinout.xyz/
	lcdRS = 7
	lcdE  = 8
	lcdD4 = 21
	lcdD5 = 20
	lcdD6 = 16
	lcdD7 = 12

	//Define some device constants
	lcdWidth = 16 // Maximum characters per line
	lcdChr   = true
	lcdCmd   = false

	lcdLine1 = 0x80 // LCD RAM address for the 1st line
	lcdLine2 = 0xC0 // LCD RAM address for the 2nd line
)

func removeNlChars(str string) string {
	isOk := func(r rune) bool {
		return r < 32 || r >= 127
	}
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isOk))
	str, _, _ = transform.String(t, str)
	return str
}

//Lcd output
type Lcd struct {
	sync.Mutex
	lcdRS rpio.Pin
	lcdE  rpio.Pin
	lcdD4 rpio.Pin
	lcdD5 rpio.Pin
	lcdD6 rpio.Pin
	lcdD7 rpio.Pin

	line1  string
	line2  string
	active bool

	msg chan string
	end chan bool
}

//NewLcd create and init new lcd output
func NewLcd() (l *Lcd) {

	if err := rpio.Open(); err != nil {
		panic(err)

	}

	l = &Lcd{
		lcdRS:  initPin(lcdRS),
		lcdE:   initPin(lcdE),
		lcdD4:  initPin(lcdD4),
		lcdD5:  initPin(lcdD5),
		lcdD6:  initPin(lcdD6),
		lcdD7:  initPin(lcdD7),
		active: true,
		msg:    make(chan string),
		end:    make(chan bool),
	}
	l.Reset()

	go func() {
		for {
			select {
			case msg := <-l.msg:
				l.display(msg)
			case _ = <-l.end:
				l.close()
				return
			}
		}
	}()
	return l
} //NewLcd()

//Display show some message
func (l *Lcd) Display(msg string) {
	l.msg <- msg
} //Display(str)

//Close LCD
func (l *Lcd) Close() {
	log.Printf("Lcd.Close")
	if l.active {
		l.end <- true
	}
} //Close()

func initPin(pin int) (p rpio.Pin) {
	p = rpio.Pin(pin)
	rpio.PinMode(p, rpio.Output)
	return
} //initPin(pin int) (p rpio.Pin)

func (l *Lcd) Reset() {
	//log.Printf("Lcd.Reset()")
	//l.writeByte(0x33, lcdCmd) // 110011 Initialise
	l.write4Bits(0x3, lcdCmd) // 110011 Initialise
	time.Sleep(5 * time.Millisecond)
	//l.writeByte(0x32, lcdCmd) // 110010 Initialise
	l.write4Bits(0x3, lcdCmd) // 110010 Initialise
	time.Sleep(120 * time.Microsecond)
	//l.writeByte(0x30, lcdCmd) // 110000 Initialise
	l.write4Bits(0x3, lcdCmd) // 110010 Initialise
	time.Sleep(120 * time.Microsecond)

	l.write4Bits(0x2, lcdCmd) // 110010 Initialise
	time.Sleep(120 * time.Microsecond)

	l.writeByte(0x28, lcdCmd) // 101000 Data length, number of lines, font size
	l.writeByte(0x0C, lcdCmd) // 001100 Display On,Cursor Off, Blink Off
	l.writeByte(0x06, lcdCmd) // 000110 Cursor move direction
	l.writeByte(0x01, lcdCmd) // 000001 Clear display
	time.Sleep(5 * time.Millisecond)
	//log.Printf("Lcd.Reset() finished")
} //Reset()

func (l *Lcd) close() error {
	l.Lock()
	defer l.Unlock()

	if !l.active {
		return errors.New(fmt.Sprintf("Lcd.close() already close: %v", l.active))
	}

	l.writeByte(lcdLine1, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		l.writeByte(' ', lcdChr)
	}
	l.writeByte(lcdLine2, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		l.writeByte(' ', lcdChr)
	}
	time.Sleep(1 * time.Second)

	l.writeByte(0x01, lcdCmd) // 000001 Clear display
	l.writeByte(0x0C, lcdCmd) // 001000 Display Off

	l.lcdRS.Low()
	l.lcdE.Low()
	l.lcdD4.Low()
	l.lcdD5.Low()
	l.lcdD6.Low()
	l.lcdD7.Low()
	rpio.Close()

	l.active = false
	close(l.msg)
	close(l.end)
	return nil
} //close()

// writeByte send byte to lcd
func (l *Lcd) writeByte(bits uint8, characterMode bool) {
	if characterMode {
		l.lcdRS.High()
	} else {
		l.lcdRS.Low()
	}

	//High bits
	if bits&0x10 == 0x10 {
		l.lcdD4.High()
	} else {
		l.lcdD4.Low()
	}
	if bits&0x20 == 0x20 {
		l.lcdD5.High()
	} else {
		l.lcdD5.Low()
	}
	if bits&0x40 == 0x40 {
		l.lcdD6.High()
	} else {
		l.lcdD6.Low()
	}
	if bits&0x80 == 0x80 {
		l.lcdD7.High()
	} else {
		l.lcdD7.Low()
	}

	//Toggle 'Enable' pin
	time.Sleep(ePulse)
	l.lcdE.High()
	time.Sleep(ePulse)
	l.lcdE.Low()
	time.Sleep(ePulse)
	//time.Sleep(eDelay)

	//Low bits
	if bits&0x01 == 0x01 {
		l.lcdD4.High()
	} else {
		l.lcdD4.Low()
	}
	if bits&0x02 == 0x02 {
		l.lcdD5.High()
	} else {
		l.lcdD5.Low()
	}
	if bits&0x04 == 0x04 {
		l.lcdD6.High()
	} else {
		l.lcdD6.Low()
	}
	if bits&0x08 == 0x08 {
		l.lcdD7.High()
	} else {
		l.lcdD7.Low()
	}
	//Toggle 'Enable' pin
	time.Sleep(ePulse)
	l.lcdE.High()
	time.Sleep(ePulse)
	l.lcdE.Low()

	time.Sleep(eDelay)
} //writeByte(bits uint8, characterMode bool)

//write4Bits send 4bits to lcd
func (l *Lcd) write4Bits(bits uint8, characterMode bool) {
	if characterMode {
		l.lcdRS.High()
	} else {
		l.lcdRS.Low()
	}

	//Low bits
	if bits&0x01 == 0x01 {
		l.lcdD4.High()
	} else {
		l.lcdD4.Low()
	}
	if bits&0x02 == 0x02 {
		l.lcdD5.High()
	} else {
		l.lcdD5.Low()
	}
	if bits&0x04 == 0x04 {
		l.lcdD6.High()
	} else {
		l.lcdD6.Low()
	}
	if bits&0x08 == 0x08 {
		l.lcdD7.High()
	} else {
		l.lcdD7.Low()
	}
	//Toggle 'Enable' pin
	time.Sleep(ePulse)
	l.lcdE.High()
	time.Sleep(ePulse)
	l.lcdE.Low()

	time.Sleep(eDelay)
} //write4Bits(bits uint8, characterMode bool)

func (l *Lcd) display(msg string) error {
	l.Lock()
	defer l.Unlock()

	if !l.active {
		return errors.New(fmt.Sprintf("Lcd.display() allready close : %v", l.active))
	}

	//log.Printf("Lcd.display('%#v')", msg)

	for line, m := range strings.Split(msg, "\n") {
		if len(m) < lcdWidth {
			m = m + strings.Repeat(" ", lcdWidth-len(m))
		}

		switch line {
		case 0:
			if l.line1 == m {
				continue
			}
			l.line1 = m
			l.writeByte(lcdLine1, lcdCmd)
		case 1:
			if l.line2 == m {
				continue
			}
			l.line2 = m
			l.writeByte(lcdLine2, lcdCmd)
		default:
			return errors.New(fmt.Sprintf("Lcd.display: to many lines %d: '%v'", line, m))
		}

		for i := 0; i < lcdWidth; i++ {
			l.writeByte(byte(m[i]), lcdChr)
		}
	}

	return nil
} //display(msg string)
