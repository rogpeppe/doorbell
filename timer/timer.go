// Package timer provides a timer API designed for use by TinyGo.
package timer

import (
	"sync"
	"time"
)

// maxIdle holds the maximum number of goroutines that will
// sit around waiting for sleep requests.
const maxIdle = 5

// Timer represents a timer designed for use by a single goroutine.
// It is designed to be kept around: in the usual case when
// reused, it does no allocations, although in unusual cases
// (calling Reset with successively smaller deadlines without waiting
// for expiry) it could start an arbitrary number of goroutines.
type Timer struct {
	C <-chan struct{}
	// c holds a copy of C so that callers can't abuse it.
	c chan struct{}
	// sleepc is used to send sleep requests to existing sleeper goroutines.
	sleepc chan time.Time

	// mu guards the fields following it.
	mu sync.Mutex
	// expiry holds the current timer expiration time.
	// If it's zero, the timer is stopped.
	expiry time.Time
	// idle holds the number of idle goroutines.
	idle int8
	// closed holds whether Timer.Close has been called.
	closed bool
	// wakeTimes holds the times that the sleeping
	// goroutines will wake, reverse-ordered.
	wakeTimes []time.Time
}

// NewTimer returns a new stopped timer. It must be closed
// with the Close method when done with, otherwise
// it can leak goroutines.
func NewTimer() *Timer {
	c := make(chan struct{}, 1)
	t := &Timer{
		C:      c,
		c:      c,
		sleepc: make(chan time.Time),
	}
	return t
}

// After resets the timer to d and returns t.C.
// Note that unlike time.After, at most one goroutine
// can use the channel at any one time.
func (t *Timer) After(d time.Duration) <-chan struct{} {
	t.Reset(d)
	return t.c
}

// maxSleepTime holds the maximum duration of a sleep.
// This means that sleeper goroutines will go away in a relatively
// short time even if a very long timer is changed or cancelled.
const maxSleepTime = 500 * time.Millisecond

// Reset starts the timer going, resetting any existing
// timer expiration. After the given duration, a value will be sent on t.C.
func (t *Timer) Reset(d time.Duration) {
	expiry := time.Now().Add(d)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.retract()
	t.expiry = expiry
	firstWakeup := t.firstWakeup()
	if !firstWakeup.IsZero() && !firstWakeup.After(expiry) {
		// There's already a sleeper that will wake up in time,
		// so we can rely on that to do the sending.
		return
	}
	// Don't sleep for longer than maxSleepTime.
	now := time.Now()
	wakeup := expiry
	if wakeup.Sub(now) > maxSleepTime {
		wakeup = now.Add(maxSleepTime)
	}
	t.wakeTimes = append(t.wakeTimes, wakeup)
	// Try to find an idle goroutine to sleep.
	select {
	case t.sleepc <- wakeup:
		t.idle--
		return
	default:
	}
	// No idle goroutine available, so start one.
	go t.sleeper(wakeup)
}

func (t *Timer) sleeper(wakeup time.Time) {
	dt := time.Until(wakeup)
	time.Sleep(dt)
	for {
		t.mu.Lock()
		t.removeWakeTime(wakeup)
		wakeup = t.maybeSend()
		isIdle := wakeup.IsZero()
		if isIdle {
			if t.idle >= maxIdle {
				// Too many idle goroutines; stop this one.
				t.mu.Unlock()
				return
			}
			t.idle++
		} else {
			t.wakeTimes = append(t.wakeTimes, wakeup)
		}
		t.mu.Unlock()
		if isIdle {
			// No need to sleep, so just wait for a sleep request.
			var ok bool
			wakeup, ok = <-t.sleepc
			if !ok {
				// The Timer has been closed.
				return
			}
		}
		time.Sleep(time.Until(wakeup))
	}
}

// removeWakeTime removes the given time from the wakeTime
// slice. Called with t.mu held.
func (t *Timer) removeWakeTime(wakeup time.Time) {
	// Although the shortest sleep should wake up first,
	// that's not guaranteed, so remove the wake time
	// from the slice whereever we find it.
	wakeTimes := t.wakeTimes
	for i := len(wakeTimes) - 1; i >= 0; i-- {
		if wakeTimes[i] == wakeup {
			copy(wakeTimes[i:], wakeTimes[i+1:])
			t.wakeTimes = wakeTimes[0 : len(wakeTimes)-1]
			return
		}
	}
	panic("wakeup time not found in list")
}

// Stop stops the timer. After this returns, no value
// will be received on t.C until the timer is started again.
func (t *Timer) Stop() {
	t.mu.Lock()
	t.retract()
	t.expiry = time.Time{}
	t.mu.Unlock()
}

// Close closes the timer, which should not be used afterwards.
// Close must not be called more than once.
func (t *Timer) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.closed {
		close(t.sleepc)
		t.closed = true
	}
}

// retract takes back a channel send if we've sent it,
// it hasn't been received and the timer has been stopped or reset.
func (t *Timer) retract() {
	select {
	case <-t.c:
	default:
	}
}

// maybeSend sends a value on t.C if the timer deadline has expired
// and returns a new wakeup time (or the zero time if there's nothing
// to wait for).
//
// Called with t.mu held.
func (t *Timer) maybeSend() time.Time {
	expiry := t.expiry
	if expiry.IsZero() {
		// The timer has stopped.
		return expiry
	}
	now := time.Now()
	if now.Before(expiry) {
		// The timer hasn't expired yet, either because the
		// expiry time was changed to be later, or because the
		// timer duration was more than maxSleepTime.
		firstWake := t.firstWakeup()
		if !firstWake.IsZero() && !firstWake.After(expiry) {
			// There's another sleeper that can do the job,
			// so we've nothing to do.
			return time.Time{}
		}
		dt := expiry.Sub(now)
		wakeup := expiry
		if dt > maxSleepTime {
			wakeup = now.Add(maxSleepTime)
		}
		return wakeup
	}
	// The timer has expired so send on the timer channel.
	select {
	case t.c <- struct{}{}:
	default:
	}
	t.expiry = time.Time{}
	return time.Time{}
}

func (t *Timer) firstWakeup() time.Time {
	if len(t.wakeTimes) == 0 {
		return time.Time{}
	}
	return t.wakeTimes[len(t.wakeTimes)-1]
}
