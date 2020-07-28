package main

import (
	"github.com/rogpeppe/doorbell/sequence"
)

const (
	noteC1 = iota
	noteCs1
	noteD1
	noteEb1
	noteE1
	noteF1
	noteFs1
	noteG1
	noteGs1
	noteA1
	noteAs1
	noteB1
	noteC2
	noteCs2
	noteD2
	noteEb2
	noteE2
	noteF2
	noteFs2
	noteG2
	noteGs2
	noteA2
	noteAs2
	noteB2
)

var dingActions = []sequence.Action{{
	Chan: noteC2,
	On:   true,
	When: 0,
}, {
	Chan: noteC2,
	On:   false,
	When: solenoidDuration,
}}

var dongActions = []sequence.Action{{
	Chan: noteG2,
	On:   true,
	When: 0,
}, {
	Chan: noteG2,
	On:   false,
	When: solenoidDuration,
}}

var tunesData = [][]byte{
	sequenceTune,
	happyBirthdayTune,
	rippleTune,
}

// Note: channel, delay before activation (milliseconds, two bytes)
var sequenceTune = []byte{
	0, 0, 0,
	0, 0, 1,
	2, 0, 2,
	2, 0, 3,
	2, 0, 4,
	0, 0, 5,
	0, 0, 6,
	2, 0, 7,
	2, 0, 8,
	2, 0, 9,
	2, 0, 10,
	2, 0, 11,
	2, 0, 12,
	2, 0, 13,
	2, 0, 14,
	2, 0, 15,
	2, 0, 16,
	2, 0, 17,
	2, 0, 18,
	2, 0, 19,
	2, 0, 20,
	2, 0, 21,
	2, 0, 22,
	2, 0, 23,
}

var happyBirthdayTune = []byte{
	0, 0,
	noteG1, 0x2, 0xee,
	noteG1, 0x0, 0xfa,
	noteA1, 0x1, 0xf4,
	noteG1, 0x1, 0xf4,
	noteC2, 0x1, 0xf4,
	noteB2, 0x3, 0xe8,

	noteG1, 0x2, 0xee,
	noteG1, 0x0, 0xfa,
	noteA1, 0x1, 0xf4,
	noteG1, 0x1, 0xf4,
	noteD2, 0x1, 0xf4,
	noteC2, //, 0x3, 0xe8
}

var rippleTune = []byte{
	0, 0,
	0, 0, 0,
	1, 0, 40,
	2, 0, 0,
	3, 0, 40,
	4, 0, 0,
	5, 0, 40,
	6, 0, 0,
	7, 0, 40,
	8, 0, 0,
	9, 0, 40,
	10, 0, 0,
	11, //, 0, 40,
}
