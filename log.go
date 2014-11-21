package stretcher

import (
	"io"
	"log"
	"os"
	"bytes"
)

type LogWriter struct {
	Output []io.Writer
}

func (w LogWriter) Write(b []byte) (int, error) {
	written := 0
	for _, o := range w.Output {
		wt, err := o.Write(b)
		if err != nil {
			return written, err
		}
		written += wt
	}
	return written, nil
}

func NewLogger(buf *bytes.Buffer) *log.Logger {
	writer := LogWriter{[]io.Writer{os.Stderr, buf}}
	return log.New(writer, "", log.LstdFlags)
}
