package debounce

import (
	"time"
)

// TODO make this configurable.
const debounceTime = 50 * time.Millisecond

// Debouncer implements button debouncing logic.
// Call Update repeatedly to update the state,
// and use State to access the stable state.
// The zero value of a Debouncer is OK to use.
type Debouncer struct {
	stableState bool
	// stable holds whether the current state is considered stable.
	isStable      bool
	// state holds the most recently updated state.
	state       bool
	// lastChanged holds the time that state changed
	// most recently.
	lastChanged time.Time
}

// State returns the most recently known stable state.
func (d *Debouncer) State() bool {
	return d.stableState
}

// Update updates the debouncer with the latest button state.
func (d *Debouncer) Update(state bool) {
	d.updateAtTime(state, time.Now())
}

func (d *Debouncer) updateAtTime(state bool, now time.Time) {
	switch {
	case state != d.state:
		d.lastChanged = now
		d.state = state
		if d.isStable {
			// When a change happens after a period of stability,
			// we want to generate an event immediately rather
			// than adding the debounce latency, because we
			// know that *something* has happened.
			d.isStable = false
			d.stableState = state
		}
	case now.After(d.lastChanged.Add(debounceTime)):
		d.isStable = true
		d.stableState = state
	}
}
