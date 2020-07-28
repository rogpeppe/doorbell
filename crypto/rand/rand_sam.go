// build tags derived from:
// 	grep -L TRNG device/sam/*.go

// +build at91sam9cn11 at91sam9cn12 at91sam9m10 at91sam9m11 at91sam9n12 atsam3a4c atsam3a8c atsam3x4c atsam3x4e atsam3x8c atsam3x8e atsam3x8h atsama5d31 atsama5d33 atsama5d34 atsama5d35 atsamd51g18a atsamd51g19a atsamd51j18a atsamd51j19a atsamd51j20a atsamd51n19a atsamd51n20a atsamd51p19a atsamd51p20a atsaml11d14a atsaml11d15a atsaml11d16a atsaml11e14a atsaml11e15a atsaml11e16a atsamv71j19 atsamv71j19b atsamv71j20 atsamv71j20b atsamv71j21 atsamv71j21b atsamv71n19 atsamv71n19b atsamv71n20 atsamv71n20b atsamv71n21 atsamv71n21b atsamv71q19 atsamv71q19b atsamv71q20 atsamv71q20b atsamv71q21 atsamv71q21b

// Package rand provides access to a hardware random number generator.
// It's similar to crypto/rand except that it doesn't provide the Int and Prime
// functions.
package rand

import (
	"device/sam"
	"encoding/binary"
	"runtime/interrupt"
	"sync"
)

var (
	initOnce sync.Once
	randc    chan uint32
)

func init() {
	Reader = hwReader{}
}

type hwReader struct{}

func (hwReader) Read(buf []byte) (int, error) {
	var randData [4]byte
	n := 0
	for n < len(buf) {
		x, ok := getUint32(n == 0)
		if !ok {
			// We've failed to get a number, which can only happen
			// if blocking is false, in which case we've already
			// got some data.
			break
		}
		binary.LittleEndian.PutUint32(randData[:], x)
		n += copy(buf[n:], randData[:])
	}
	return n, nil
}

// getUint32 returns a uint32 value generated by the
// hardware random number generator and reports
// whether it succeeded. It will always succeed
// if blocking is true.
func getUint32(blocking bool) (uint32, bool) {
	initOnce.Do(setup)
	select {
	case x := <-randc:
		return x, true
	default:
		if !blocking {
			return 0, false
		}
		// The channel is empty, which might be because
		// we're reading too fast for the generator to
		// keep up, or because we were reading too slow
		// and the channel got full. In either case,
		// ensure that the interrupt is enabled (it's idempotent).
		sam.TRNG.INTENSET.SetBits(sam.TRNG_INTENSET_DATARDY)
		return <-randc, true
	}
}

func setup() {
	randc = make(chan uint32, 8)

	// Enable Main Clock for TRNG.
	sam.MCLK.APBCMASK.SetBits(sam.MCLK_APBCMASK_TRNG_)

	// Set up the TRNG interrupt handler.
	interruptTRNG := interrupt.New(sam.IRQ_TRNG, handleTRNG)
	interruptTRNG.Enable()

	// Enable TRNG.
	sam.TRNG.CTRLA.SetBits(sam.TRNG_CTRLA_ENABLE)
}

func handleTRNG(interrupt.Interrupt) {
	// We've got a random number. Put it into the channel if we can.
	select {
	case randc <- sam.TRNG.DATA.Get():
		// We've put the random value into the stream.
		if len(randc) < cap(randc) {
			// There's space for another value in the channel, so mark the
			// interrupt as ready so we'll get another one.
			sam.TRNG.INTENSET.SetBits(sam.TRNG_INTENSET_DATARDY)
		}
	default:
		// This shouldn't happen, as we only set the data-ready indication
		// when there's room enough in the channel for the next value,
		// but clear it anyway to be sure.
		sam.TRNG.INTENCLR.SetBits(sam.TRNG_INTENCLR_DATARDY)
	}
}
