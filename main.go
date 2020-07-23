/*
possible boards:
ItsyBitsy M4 Express (but is this the Feather ?)
https://www.amazon.co.uk/Adafruit-ItsyBitsy-Express-featuring-ATSAMD51/dp/B07GYYYJSB?fbclid=IwAR30jM7HVdfo5q53Q3VeOn_uyPWJGvzJsRquIVeobL_5Y37t6RlWNpEkdEQ

https://pkg.go.dev/github.com/tinygo-org/drivers/flash?tab=doc#Device
https://github.com/bgould/go-littlefs
*/

package main

import (
	"encoding/binary"
	"machine"
	"sort"
	"time"
)

var solenoidPins = []machine.Pin{
	machine.D0,
	machine.D1,
	machine.D4,
	machine.D5,
	machine.D6,
	machine.D8,
	machine.D9,
	machine.D10,
	machine.D11,
	machine.D12,
	machine.D13,
	machine.D21,
	machine.D22,
	machine.D23,
}

var doorButtonPins = []machine.Pin{
	machine.D24,
	machine.D25,
}

func main() {
	data := readTuneData()
	actions := SequenceForTune(16, data)
	Doorbell(solenoidPins, doorButtonPins, actions)
}

var tuneData = []byte{
	0, 0, 0,
	0, 0, 250,
	2, 0, 125,
}

func readTuneData() []byte {
	return tuneData
}

func Doorbell(solenoids []machine.Pin, doorButtons []machine.Pin, tune []Action) {
	for _, pin := range solenoids {
		pin.Configure(machine.PinConfig{
			Mode: machine.PinOutput,
		})
	}
	for _, pin := range tune {
		pin.Configure(machine.PinConfig{
			Mode: machine.PinInputPullup,
		})
	}

	pushed := make(chan struct{}, 1)
	go waitPush(doorButtons, pushed)
	go player(solenoids, tune, pushed)
}

func waitPush(buttons []machine.Pin, pushed chan<- struct{}) {
	state := 0
	for {
		for i, b := range buttons {
			mask := 1 << i
			if b.Get() {
				if state&mask != 0 {
					continue
				}
				// TODO debounce.
				select {
				case pushed <- struct{}{}:
				default:
				}
				state |= mask
			} else {
				state &^= mask
			}
		}
		// TODO can we avoid continuously polling the
		// buttons (e.g. by setting up an interrupt) ?
		time.Sleep(time.Millisecond)
	}
}

func player(solenoids []machine.Pin, tune []Action, pushed <-chan struct{}) {
	for {
		<-pushed
		for !Play(solenoids, tune, pushed) {
		}
	}
}

// Action holds an action to perform on a given channel.
type Action struct {
	// Chan holds the number of the channel for the action.
	Chan uint8
	// On holds whether to turn the channel on or off.
	On bool
	// When holds the time from the start of the sequence
	// that the action should take place.
	When time.Duration
}

// solenoidDuration is the amount of time to pulse the
// solenoid relay for to make the sound.
const solenoidDuration = 50 * time.Millisecond

// SequenceForTune reads a sequence of channel
// activations (solenoid pulses) from the following data format.
// Each entry holds a channel number (1 byte) and a length
// of time to delay before activating that channel (2 bytes, little endian).
//
// The returned actions will be sorted in time order.
func SequenceForTune(chanCount int, data []byte) []Action {
	actions := make([]Action, 0, len(data)/3*2)
	now := time.Duration(0)
	for len(data) > 0 {
		if len(data) < 3 {
			return actions
		}
		channel := data[0]
		duration := binary.LittleEndian.Uint16(data[1:3])
		now += time.Duration(duration) * time.Millisecond
		if int(channel) >= chanCount {
			// ignore out-of-range channels
			continue
		}
		actions = append(actions, Action{
			Chan: channel,
			On:   true,
			When: now,
		}, Action{
			Chan: channel,
			On:   false,
			When: now + solenoidDuration,
		})
	}
	sort.Stable(actionsByTime(actions))
	return actions
}

// Play plays the given sequence of actions, using the given
// pins as channels.
// It reports whether the tune was successfully played without
// being stopped by a button push.
func Play(pins []machine.Pin, seq []Action, stop <-chan struct{}) bool {
	start := time.Now()
	for _, a := range seq {
		select {
		case <-time.After(time.Until(start.Add(a.When))):
		case <-stop:
			return true
		}
		pins[a.Chan].Set(a.On)
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
