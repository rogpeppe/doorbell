// Package rand provides access to a hardware random number generator
// for TinyGo. It's similar to crypto/rand except that it doesn't provide the Int and Prime
// functions.
// Currently only the SAM series is supported; other platforms will cause Read
// to return ErrUnavailable.
package rand

import (
	"errors"
	"io"
)

// Reader is a global, shared instance of a cryptographically secure random number generator.
// It uses the chip-provided random number generator if available.
var Reader io.Reader

func init() {
	// This init function can run before or after any system-specific init
	// function, so don't override Reader if it's running second.
	if Reader == nil {
		Reader = unsupportedReader{}
	}
}

// ErrUnavailable is returned on error on platforms that don't have a supported
// hardware random number generator.
var ErrUnavailable = errors.New("no hardware number generator available")

// Read is a helper function that calls Reader.Read using io.ReadFull.
// On return, n == len(b) if and only if err == nil.
func Read(buf []byte) (int, error) {
	return io.ReadFull(Reader, buf)
}

type unsupportedReader struct{}

func (unsupportedReader) Read(buf []byte) (int, error) {
	return 0, ErrUnavailable
}
