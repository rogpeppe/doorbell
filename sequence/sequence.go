package sequence

import (
	"encoding/binary"
	"sort"
	"time"
)

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

// ActionsForTune reads a sequence of channel activations (solenoid
// pulses) from the following data format. Each entry holds a number
// of milliseconds to delay (2 bytes, big endian) and a channel numer
// to activate after the delay (1 byte).
//
// The returned actions will be sorted in time order.
func ActionsForTune(chanCount int, data []byte, solenoidDuration time.Duration) []Action {
	actions := make([]Action, 0, len(data)/3*2)
	now := time.Duration(0)
	for len(data) > 0 {
		if len(data) < 3 {
			return actions
		}
		duration := binary.BigEndian.Uint16(data[0:2])
		channel := data[2]
		data = data[3:]
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

type actionsByTime []Action

func (s actionsByTime) Less(i, j int) bool {
	return s[i].When < s[j].When
}

func (s actionsByTime) Len() int {
	return len(s)
}

func (s actionsByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
