package sequence

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

var actionsForTuneTests = []struct {
	testName         string
	chanCount        int
	solenoidDuration time.Duration
	data             []byte
	expect           []Action
}{{
	testName:         "single-action",
	chanCount:        4,
	solenoidDuration: time.Millisecond,
	data: []byte{
		2, 0, 5,
	},
	expect: []Action{{
		Chan: 2,
		On:   true,
		When: 5 * time.Millisecond,
	}, {
		Chan: 2,
		On:   false,
		When: 6 * time.Millisecond,
	}},
}, {
	testName:         "several-actions",
	chanCount:        4,
	solenoidDuration: time.Millisecond,
	data: []byte{
		2, 0, 5,
		2, 0, 3,
		3, 0, 2,
	},
	expect: []Action{{
		Chan: 2,
		On:   true,
		When: 5 * time.Millisecond,
	}, {
		Chan: 2,
		On:   false,
		When: 6 * time.Millisecond,
	}, {
		Chan: 2,
		On:   true,
		When: 8 * time.Millisecond,
	}, {
		Chan: 2,
		On:   false,
		When: 9 * time.Millisecond,
	}, {
		Chan: 3,
		On:   true,
		When: 10 * time.Millisecond,
	}, {
		Chan: 3,
		On:   false,
		When: 11 * time.Millisecond,
	}},
}, {
	testName:         "concurrent-actions",
	chanCount:        6,
	solenoidDuration: time.Millisecond,
	data: []byte{
		2, 0, 5,
		4, 0, 0,
		2, 0, 3,
		4, 0, 0,
		3, 0, 2,
	},
	expect: []Action{{
		Chan: 2,
		On:   true,
		When: 5 * time.Millisecond,
	}, {
		Chan: 4,
		On:   true,
		When: 5 * time.Millisecond,
	}, {
		Chan: 2,
		On:   false,
		When: 6 * time.Millisecond,
	}, {
		Chan: 4,
		On:   false,
		When: 6 * time.Millisecond,
	}, {
		Chan: 2,
		On:   true,
		When: 8 * time.Millisecond,
	}, {
		Chan: 4,
		On:   true,
		When: 8 * time.Millisecond,
	}, {
		Chan: 2,
		On:   false,
		When: 9 * time.Millisecond,
	}, {
		Chan: 4,
		On:   false,
		When: 9 * time.Millisecond,
	}, {
		Chan: 3,
		On:   true,
		When: 10 * time.Millisecond,
	}, {
		Chan: 3,
		On:   false,
		When: 11 * time.Millisecond,
	}},
}}

func TestActionsForTune(t *testing.T) {
	c := qt.New(t)
	for _, test := range sequenceForTuneTests {
		c.Run(test.testName, func(c *qt.C) {
			actions := SequenceForTune(test.chanCount, test.data, test.solenoidDuration)
			c.Assert(actions, qt.DeepEquals, test.expect)
		})
	}
}
