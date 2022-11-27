package internal

import (
	"bytes"
	"io"
)

type FlushingWriter struct {
	Flush func(msg string)

	buf    bytes.Buffer
	closed bool
}

func (w *FlushingWriter) Write(bytes []byte) (int, error) {
	if w.closed {
		return 0, io.ErrClosedPipe
	}

	for _, b := range bytes {
		if b != '\n' {
			w.buf.WriteByte(b)
		} else {
			w.Flush(w.buf.String())
			w.buf.Truncate(0)
		}
	}

	return len(bytes), nil
}

func (w *FlushingWriter) Close() error {
	if w.buf.Len() != 0 {
		w.Flush(w.buf.String() + "\n")
	}
	w.closed = true
	return nil
}
