package debounce

import (
	"time"

	"github.com/rogpeppe/doorbell/mcp23017"
)

// TODO we could use a more generic type for this, such as uint32.
type Pins = mcp23017.Pins

// TODO make this configurable.
const debounceTime = 50 * time.Millisecond

// Debouncer implements button debouncing logic.
// Call Update repeatedly to update the state,
// and use State to access the stable state.
// The zero value of a Debouncer is OK to use.
type Debouncer struct {
	stableState Pins
	// stable holds whether the current state is considered stable.
	stable      bool
	state       Pins
	lastChanged time.Time
}

// State returns the most recently known stable state.
func (d *Debouncer) State() Pins {
	return d.stableState
}

// Update updates the debouncer with the latest button state.
func (d *Debouncer) Update(state Pins) {
	d.updateAtTime(state, time.Now())
}

func (d *Debouncer) updateAtTime(state Pins, now time.Time) {
	switch {
	case state != d.state:
		println("debounce state change ", state)
		d.lastChanged = now
		d.state = state
		if d.stable {
			// When a change happens after a period of stability,
			// we want to generate an event immediately rather
			// than adding the debounce latency, because we
			// know that *something* has happened.
			d.stable = false
			d.stableState = state
		}
	case now.After(d.lastChanged.Add(debounceTime)):
		d.stable = true
		d.stableState = state
	}
}
