package main

import (
	"math/rand"

	"github.com/rogpeppe/doorbell/sequence"
)

type tuneSelection struct {
	tunes     [][]sequence.Action
	rand      *rand.Rand
	played    []bool
	numPlayed int
}

func newTuneSelection(tunes [][]sequence.Action, randGen *rand.Rand) *tuneSelection {
	return &tuneSelection{
		tunes:  tunes,
		rand:   randGen,
		played: make([]bool, len(tunes)),
	}
}

func (ts *tuneSelection) reset() {
	for i := range ts.played {
		ts.played[i] = false
	}
	ts.numPlayed = 0
}

func (ts *tuneSelection) choose() []sequence.Action {
	if ts.numPlayed >= len(ts.tunes) {
		// We've played all the tunes, so start again with a new random selection.
		ts.reset()
	}
	n := ts.rand.Intn(len(ts.tunes) - ts.numPlayed)
	for i, played := range ts.played {
		if played {
			continue
		}
		if n == 0 {
			return ts.tunes[i]
		}
		n--
	}
	panic("unreachable")
}
