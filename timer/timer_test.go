package timer

import (
	"testing"
	"time"
)

func TestSimpleTimer(t *testing.T) {
	timer := NewTimer()
	defer timer.Close()
	t0 := time.Now()
	timer.Reset(10 * time.Millisecond)
	<-timer.C
	dt := time.Since(t0)
	if got, want := dt, 10*time.Millisecond; got < want {
		t.Fatalf("timer fired too soon; got %v want >=%v", got, want)
	}
}

func TestResetShortToLong(t *testing.T) {
	timer := NewTimer()
	defer timer.Close()
	t0 := time.Now()
	timer.Reset(10 * time.Millisecond)
	timer.Reset(20 * time.Millisecond)
	<-timer.C
	assertDuration(t, time.Since(t0), 20*time.Millisecond)
}

func BenchmarkRepeatedTimer(b *testing.B) {
	b.ReportAllocs()
	timer := NewTimer()
	defer timer.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timer.Reset(time.Microsecond)
		<-timer.C
	}
}

const margin = time.Millisecond

func TestResetLongToShort(t *testing.T) {
	timer := NewTimer()
	defer timer.Close()
	t0 := time.Now()
	timer.Reset(20 * time.Millisecond)
	timer.Reset(10 * time.Millisecond)
	<-timer.C
	assertDuration(t, time.Since(t0), 10*time.Millisecond)
	// Check that the original timer doesn't fire:
	select {
	case <-timer.C:
		t.Fatalf("unexpected receive on timer channel at %v", time.Since(t0))
	case <-time.After(30 * time.Millisecond):
	}
}

func TestExpireWithoutReceive(t *testing.T) {
	timer := NewTimer()
	defer timer.Close()
	timer.Reset(time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	t0 := time.Now()
	timer.Reset(10 * time.Millisecond)
	<-timer.C
	assertDuration(t, time.Since(t0), 10*time.Millisecond)
	select {
	case <-timer.C:
		t.Fatalf("unexpected receive on timer channel at %v", time.Since(t0))
	case <-time.After(time.Millisecond):
	}
}

func TestStop(t *testing.T) {
	timer := NewTimer()
	defer timer.Close()
	t0 := time.Now()
	timer.Reset(time.Millisecond)
	timer.Stop()
	select {
	case <-timer.C:
		t.Fatalf("unexpected receive on timer channel at %v", time.Since(t0))
	case <-time.After(time.Millisecond):
	}
}

func TestMultipleResets(t *testing.T) {
	timer := NewTimer()
	defer timer.Close()
	var t0 time.Time
	// This should cause a goroutine to be started each time,
	// because each sleep is shorter than the last one.
	for d := 30 * time.Millisecond; d >= 5*time.Millisecond; d -= time.Millisecond {
		t0 = time.Now()
		timer.Reset(d)
	}
	<-timer.C
	assertDuration(t, time.Since(t0), 5*time.Millisecond)
	// If we reset now, an existing goroutine should be used (hard to check
	// directly, but try to tickle the code path anyway).
	t0 = time.Now()
	timer.Reset(10 * time.Millisecond)
	<-timer.C
	assertDuration(t, time.Since(t0), 10*time.Millisecond)
	select {
	case <-timer.C:
		t.Fatalf("unexpected receive on timer channel at %v", time.Since(t0))
	case <-time.After(40 * time.Millisecond):
	}
	if got, want := timer.idle, int8(maxIdle); got != want {
		t.Fatalf("unexpected idle count; got %v want %v", got, want)
	}
}

func TestStopWithoutPreviousReceive(t *testing.T) {
	t0 := time.Now()
	timer := NewTimer()
	defer timer.Close()
	timer.Reset(time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	timer.Stop()
	select {
	case <-timer.C:
		t.Fatalf("unexpected receive on timer channel at %v", time.Since(t0))
	case <-time.After(time.Millisecond):
	}
}

func assertDuration(t *testing.T, got, want time.Duration) {
	if got < want {
		t.Fatalf("duration too small; got %v want at least %v", got, want)
	}
	if got > want+margin {
		t.Fatalf("duration too large; got %v want not more than %v+margin", got, want)
	}
}
