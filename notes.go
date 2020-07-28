package main

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

var dingActions = []Action{{
	Chan: noteC2,
	On:   true,
	When: 0,
}, {
	Chan: noteC2,
	On:   false,
	When: solenoidDuration,
}}

var dongActions = []Action{{
	Chan: noteG2,
	On:   true,
	When: 0,
}, {
	Chan: noteG2,
	On:   false,
	When: solenoidDuration,
}}

func readTuneData() []byte {
	return tuneData
}

// Note: channel, delay before activation (milliseconds, two bytes)
var tuneData = []byte{
	0, 2, 0,
	1, 2, 0,
	2, 2, 0,
	3, 2, 0,
	4, 2, 0,
	5, 2, 0,
	6, 2, 0,
	7, 2, 0,
	8, 2, 0,
	9, 2, 0,
	10, 2, 0,
	11, 2, 0,
	12, 2, 0,
	13, 2, 0,
	14, 2, 0,
	15, 2, 0,
	16, 2, 0,
	17, 2, 0,
	18, 2, 0,
	19, 2, 0,
	20, 2, 0,
	21, 2, 0,
	22, 2, 0,
	23, 0, 0,
}

//, {
//	noteD1, 0, 0,
//	noteD1, 2, 0,
//	noteA1, 2, 0,
//	noteA1, 2, 0,
//	noteB1, 1, 0,
//	noteCs2, 1, 0,
//	noteD2, 1, 0,
//	noteE2, 1, 0,
//	noteCs2, 1, 0,
//	noteB1, 1, 0,
//}
