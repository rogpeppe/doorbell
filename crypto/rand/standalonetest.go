// +build ignore

// This program exercises the randomness reader.
// It prints random bytes in base64 format.
package main

import (
	"io"
	"os"
	"time"

	"encoding/base64"

	"github.com/rogpeppe/doorbell/crypto/rand"
)

const (
	// fastBytes holds the number of bytes of random output to
	// produce when reading as fast as possible.
	fastBytes = 1024

	// slowReads holds the number of reads to do
	// of size slowBufSize at slowDelay intervals when reading slowly.
	slowReads   = 100
	slowBufSize = 3
	slowDelay   = 30 * time.Millisecond
)

func main() {
	time.Sleep(3 * time.Second)
	println("fast read")
	w := base64.NewEncoder(base64.StdEncoding, &wrappingWriter{
		wrap: 76,
		w:    os.Stdout,
	})
	if _, err := io.Copy(w, io.LimitReader(rand.Reader, 1024)); err != nil {
		println("copy error: ", err)
	} else {
		w.Close()
		os.Stdout.Write(nl)
	}

	if slowReads == 0 {
		return
	}
	time.Sleep(time.Second)
	println("slow read")
	w = base64.NewEncoder(base64.StdEncoding, &wrappingWriter{
		wrap: 76,
		w:    os.Stdout,
	})
	buf := make([]byte, slowBufSize)
	for i := 0; i < slowReads; i++ {
		n, err := rand.Read(buf)
		if err != nil {
			println("read error: ", err.Error())
		}
		if n != len(buf) {
			println("unexpected count: ", n)
		}
		w.Write(buf)
		time.Sleep(slowDelay)
	}
	w.Close()
	os.Stdout.Write(nl)
	println("done")
}

type wrappingWriter struct {
	w    io.Writer
	wrap int
	col  int
}

var nl = []byte("\r\n")

func (w *wrappingWriter) Write(buf []byte) (int, error) {
	total := len(buf)
	for w.col+len(buf) > w.wrap {
		n := w.wrap - w.col
		if _, err := w.w.Write(buf[:n]); err != nil {
			return 0, err
		}
		if _, err := w.w.Write(nl); err != nil {
			return 0, err
		}
		w.col = 0
		buf = buf[n:]
	}
	if len(buf) > 0 {
		if _, err := w.w.Write(buf); err != nil {
			return 0, err
		}
		w.col += len(buf)
	}
	return total, nil
}
