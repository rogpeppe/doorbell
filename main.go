/*
possible boards:
ItsyBitsy M4 Express (but is this the Feather ?)
https://www.amazon.co.uk/Adafruit-ItsyBitsy-Express-featuring-ATSAMD51/dp/B07GYYYJSB?fbclid=IwAR30jM7HVdfo5q53Q3VeOn_uyPWJGvzJsRquIVeobL_5Y37t6RlWNpEkdEQ

https://pkg.go.dev/github.com/tinygo-org/drivers/flash?tab=doc#Device
https://github.com/bgould/go-littlefs
*/

// TODO what should we do if two doorbells are activated at the same time?

package main

import (
	"time"

	"github.com/rogpeppe/doorbell/debounce"
	"github.com/rogpeppe/doorbell/mcp23017"
	"github.com/rogpeppe/doorbell/timer"
)

var buttonPinNumbers = []uint8{
	0,
	1,
	2,
	3,
	4,
}

var solenoidPinNumbers = []uint8{
	// Back left: 0x21 port A
	16,
	17,
	18,
	19,
	20,
	21,
	22,
	23,

	// Back right: 0x20 port B
	8,
	9,
	10,
	11,
	12,
	13,
	14,
	15,

	// Front right: 0x20 Port A
	0,
	1,
	2,
	3,
	4,
	5,
	6,
	7,
}

var numSolenoids = len(solenoidPinNumbers)

// solenoidDuration is the amount of time to pulse the
// solenoid relay for to make the sound.
const solenoidDuration = 200 * time.Millisecond

func main() {
	time.Sleep(3 * time.Second)
	println("starting....")
	devs, err := getDevices(0x20, 0x21, 0x22)
	if err != nil {
		fatal("cannot make new i2c devices: ", err.Error())
	}
	outputs := devs[0:2]
	outputs.SetModes([]mcp23017.PinMode{mcp23017.Output})
	solenoidPins := make([]Pin, len(solenoidPinNumbers))
	for i, n := range solenoidPinNumbers {
		solenoidPins[i] = outputs.Pin(int(n))
	}
	println("pin count ", len(solenoidPins))
	inputs := devs[2]
	if err := inputs.SetModes([]mcp23017.PinMode{mcp23017.Input | mcp23017.Pullup | mcp23017.Invert}); err != nil {
		fatal("cannot set modes: ", err.Error())
	}
	println("set modes etc")
	data := readTuneData()
	actions := SequenceForTune(16, data)
	Doorbell(solenoidPins, &buttonDev{
		dev:  inputs,
		mask: 1<<len(buttonPinNumbers) - 1,
	}, actions)
}

// Note: channel, delay before activation (milliseconds, two bytes)
var tuneData = []byte{
	0, 4, 0,
	1, 4, 0,
	2, 4, 0,
	3, 4, 0,
	4, 4, 0,
	5, 4, 0,
	6, 4, 0,
	7, 4, 0,
	8, 4, 0,
	9, 4, 0,
	10, 4, 0,
	11, 4, 0,
	12, 4, 0,
	13, 4, 0,
	14, 4, 0,
	15, 4, 0,
	16, 4, 0,
	17, 4, 0,
	18, 4, 0,
	19, 4, 0,
	20, 4, 0,
	21, 4, 0,
	22, 4, 0,
	23, 0, 0,

	//	noteD1, 0, 0,
	//	noteD1, 2, 0,
	//	noteA1, 2, 0,
	//	noteA1, 2, 0,
	//	noteB1, 1, 0,
	//	noteCs2, 1, 0,
	//	noteD2, 1, 0,
	//	noteE2, 1, 0,
	//	noteCs2, 1, 0,
	//	noteB1, 1, 0,
}

func readTuneData() []byte {
	return tuneData
}

type Pin = mcp23017.Pin

type buttonDev struct {
	dev  *mcp23017.Device
	mask mcp23017.Pins
}

func (b *buttonDev) buttons() mcp23017.Pins {
	// Ignore error because we don't care enough.
	buts, _ := b.dev.GetPins()
	return buts & b.mask
}

var dingActions = []Action{{
	Chan: noteC2,
	On:   true,
	When: 0,
}, {
	Chan: noteC2,
	On:   false,
	When: solenoidDuration,
}}

var dongActions = []Action{{
	Chan: noteG2,
	On:   true,
	When: 0,
}, {
	Chan: noteG2,
	On:   false,
	When: solenoidDuration,
}}

func Doorbell(solenoids []mcp23017.Pin, doorButtons *buttonDev, tune []Action) {
	println("starting doorbell")
	pushed := make(chan mcp23017.Pins, 1)
	go buttonPoller(doorButtons, pushed)
	go player(solenoids, tune, pushed)
	select {}
}

// buttonPoller continually polls the buttons and sends any changes
// on pushed.
func buttonPoller(doorButtons *buttonDev, pushed chan<- mcp23017.Pins) {
	println("in button poller")
	var debouncer debounce.Debouncer
	var state mcp23017.Pins
	for {
		debouncer.Update(doorButtons.buttons())
		if newState := debouncer.State(); newState != state {
			state = newState
			pushed <- state
		}
		// TODO can we avoid continuously polling the
		// buttons (e.g. by setting up an interrupt) ?
		time.Sleep(time.Millisecond)
	}
}

func player(solenoids []mcp23017.Pin, tune []Action, pushed <-chan mcp23017.Pins) {
	println("in player")
	timer := timer.NewTimer()
	for {
		// Wait for button to be pushed.
		<-pushed
		// On first push and release, just do a two-note thing.
		Play(timer, solenoids, dingActions, nil)

		// Wait for all buttons to be released.
		// TODO is this actually the right thing to do when other buttons are pushed?
		for <-pushed != 0 {
		}
		Play(timer, solenoids, dongActions, nil)
	}
}

// Play plays the given sequence of actions, using the given
// pins as channels.
// It reports whether the tune was successfully played without
// being stopped by a button push.
func Play(timer *timer.Timer, pins []mcp23017.Pin, seq []Action, stop <-chan struct{}) bool {
	start := time.Now()
	var active mcp23017.Pins
	for i, a := range seq {
		if dt := time.Until(start.Add(a.When)); dt > 0 {
			select {
			case <-timer.After(dt):
			case <-stop:
				// We've been stopped; don't stop immediately but play out
				// all the disable events so that we end up with a clean
				// slate and we always activate solenoids for the correct time.
				for _, a := range seq[i:] {
					if active == 0 {
						break
					}
					if !a.On {
						pins[a.Chan].Low()
						active.Low(int(a.Chan))
					}
				}
				return true
			}
		}
		println("channel ", a.Chan, a.On)
		pins[a.Chan].Set(a.On)
		active.Set(int(a.Chan), a.On)
	}
	return false
}

type actionsByTime []Action

func (s actionsByTime) Less(i, j int) bool {
	return s[i].When < s[i].When
}

func (s actionsByTime) Len() int {
	return len(s)
}

func (s actionsByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func fatal(args ...interface{}) {
	print("fatal: ")
	for _, a := range args {
		print(a)
	}
	println()
	select {}
}
