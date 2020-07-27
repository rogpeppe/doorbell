package main

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

// SequenceForTune reads a sequence of channel activations (solenoid
// pulses) from the following data format. Each entry holds a channel
// number (1 byte) and the number of milliseconds to delay before
// activating that channel (2 bytes, little endian).
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
		duration := binary.BigEndian.Uint16(data[1:3])
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
