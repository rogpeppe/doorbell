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
	"encoding/binary"
	"math/rand"
	"time"

	cryptorand "github.com/rogpeppe/doorbell/crypto/rand"
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
	actions := SequenceForTune(numSolenoids, data)
	Doorbell(DoorbellParams{
		Solenoids: solenoidPins,
		DoorButtons: &buttonDevice{
			dev:  inputs,
			mask: 1<<len(buttonPinNumbers) - 1,
		},
		Tune: actions,
		Rand: newRandSource(),
	})
}

type Pin = mcp23017.Pin

type buttonDevice struct {
	dev  *mcp23017.Device
	mask mcp23017.Pins
}

func (b *buttonDevice) buttons() mcp23017.Pins {
	// Ignore error because we don't care enough.
	buts, _ := b.dev.GetPins()
	return buts & b.mask
}

type DoorbellParams struct {
	Solenoids   []mcp23017.Pin
	DoorButtons *buttonDevice
	Tune        []Action
	Rand        *rand.Rand
}

func Doorbell(p DoorbellParams) {
	println("starting doorbell")
	pushed := make(chan mcp23017.Pins, 1)
	go buttonPoller(p.DoorButtons, pushed)
	go player(p.Solenoids, p.Tune, pushed)
	select {}
}

// buttonPoller continually polls the buttons and sends any changes
// on pushed.
func buttonPoller(doorButtons *buttonDevice, pushed chan<- mcp23017.Pins) {
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
		println("wait for button")
		// Wait for button to be pushed.
		<-pushed
		println("button pushed")
		// On first push and release, just do a two-note thing.
		Play(timer, solenoids, dingActions, nil, nil)

		// Wait for all buttons to be released, but if they press the
		// button for a long time, play a tune instead of the "dong" sound.
		timer.Reset(750 * time.Millisecond)
	buttonWait:
		for {
			select {
			case state := <-pushed:
				if state == 0 {
					Play(timer, solenoids, dongActions, nil, nil)
					break buttonWait
				}
				// One of the buttons is still pressed.
				// TODO is this actually the right thing to do when other buttons are pushed?
			case <-timer.C:
				// The button's been pushed for a long time: start a tune playing.
				stop := make(chan struct{})
				done := make(chan struct{})
			tuneLoop:
				for {
					go Play(timer, solenoids, tune, stop, done)
					// Wait for all buttons to be released.
					for <-pushed != 0 {
					}
					select {
					case <-pushed:
						// The button has been pushed again while the tune is playing,
						// so stop the tune playing and head around the loop to
						// start another tune.
						select {
						case stop <- struct{}{}:
							<-done
						case <-done:
						}
					case <-done:
						// The tune has finished playing.
						break tuneLoop
					}
				}
			}
		}
	}
}

// Play plays the given sequence of actions, using the given
// pins as channels. It stops if it receives a value on the stop
// channel.
//
// If done is non-nil, a value will be sent on it before Play
// returns.
func Play(timer *timer.Timer, pins []mcp23017.Pin, seq []Action, stop <-chan struct{}, done chan<- struct{}) {
	start := time.Now()
	var active mcp23017.Pins
sequenceLoop:
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
				break sequenceLoop
			}
		}
		println("channel ", a.Chan, a.On)
		pins[a.Chan].Set(a.On)
		active.Set(int(a.Chan), a.On)
	}
	if done != nil {
		done <- struct{}{}
	}
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

func newRandSource() *rand.Rand {
	var seed int64
	var buf [8]byte
	// Use H/W random number seed by default, falling
	// back to RTC if that's unavailable.
	if _, err := cryptorand.Read(buf[:]); err != nil {
		seed = time.Now().UnixNano()
	} else {
		seed = int64(binary.LittleEndian.Uint64(buf[:]))
	}
	return rand.New(rand.NewSource(seed))
}
