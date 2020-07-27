package debounce

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestDebouncer(t *testing.T) {
	c := qt.New(t)
	var d Debouncer
	t0 := time.Now()
	now := t0
	d.updateAtTime(0, t0)
	c.Assert(d.State(), qt.Equals, Pins(0))
	now = now.Add(time.Millisecond)
	// The first update after a period of stability should immediately
	// trigger an state change.
	d.updateAtTime(1, now)
	c.Assert(d.State(), qt.Equals, Pins(1))
	// Subsequent fast changes shouldn't change the state.

	now = now.Add(time.Millisecond)
	d.updateAtTime(0, now)
	c.Assert(d.State(), qt.Equals, Pins(1))

	now = now.Add(time.Millisecond)
	d.updateAtTime(0, now)
	c.Assert(d.State(), qt.Equals, Pins(1))

	now = now.Add(time.Millisecond)
	d.updateAtTime(1, now)
	c.Assert(d.State(), qt.Equals, Pins(1))

	now = now.Add(debounceTime + 1)
	d.updateAtTime(1, now)
	c.Assert(d.State(), qt.Equals, Pins(1))

	// The state should be considered stable now, so the first
	// subsequent change should update the state.
	now = now.Add(1)
	d.updateAtTime(0, now)
	c.Assert(d.State(), qt.Equals, Pins(0))
}
