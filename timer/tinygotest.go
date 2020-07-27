// +build ignore

// This program benchmarks the overhead of the timer primitive.
// It's designed to be run using TinyGo.
package main

import (
	"time"

	"github.com/rogpeppe/doorbell/timer"
)

const iterations = 1000000

func main() {
	time.Sleep(3 * time.Second)
	println("starting")
	sanityCheck()
	iterations := 10
	var d time.Duration
	for {
		println("iterations", iterations)
		timer := timer.NewTimer()
		t0 := time.Now()
		for i := 0; i < iterations; i++ {
			timer.Reset(0)
			<-timer.C
		}
		d = time.Since(t0)
		timer.Close()
		if d >= time.Second {
			break
		}
		iterations *= 2
	}
	println("iterations ", iterations)
	println("timer overhead: ", (d / time.Duration(iterations)).String())
}

func sanityCheck() {
	timer := timer.NewTimer()
	timer.Reset(0)
	timeout := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(timeout)
	}()
	select {
	case <-timer.C:
	case <-timeout:
		println("timed out waiting for timeout")
	}
	timer.Close()
}
