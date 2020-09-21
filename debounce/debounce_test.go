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
	d.updateAtTime(false, t0)
	c.Assert(d.State(), qt.Equals, false)
	now = now.Add(time.Millisecond)
	// The first update after a period of stability should immediately
	// trigger an state change.
	d.updateAtTime(true, now)
	c.Assert(d.State(), qt.Equals, true)
	// Subsequent fast changes shouldn't change the state.

	now = now.Add(time.Millisecond)
	d.updateAtTime(false, now)
	c.Assert(d.State(), qt.Equals, true)

	now = now.Add(time.Millisecond)
	d.updateAtTime(false, now)
	c.Assert(d.State(), qt.Equals, true)

	now = now.Add(time.Millisecond)
	d.updateAtTime(true, now)
	c.Assert(d.State(), qt.Equals, true)

	now = now.Add(debounceTime + 1)
	d.updateAtTime(true, now)
	c.Assert(d.State(), qt.Equals, true)

	// The state should be considered stable now, so the first
	// subsequent change should update the state.
	now = now.Add(1)
	d.updateAtTime(false, now)
	c.Assert(d.State(), qt.Equals, false)
}
