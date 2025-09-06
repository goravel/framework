package process

import (
	"bytes"
	"io"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

type OutputWriterHandler func(key string, typ contractsprocess.OutputType, line []byte)

func NewOutputWriterForProcess(typ contractsprocess.OutputType, handler contractsprocess.OnOutputFunc) *OutputWriter {
	return NewOutputWriter("", typ, func(_ string, t contractsprocess.OutputType, line []byte) {
		handler(t, line)
	})
}

func NewOutputWriterForPipe(key string, typ contractsprocess.OutputType, h contractsprocess.OnPipeOutputFunc) *OutputWriter {
	return NewOutputWriter(key, typ, func(k string, t contractsprocess.OutputType, line []byte) {
		h(k, t, line)
	})
}

func NewOutputWriter(key string, typ contractsprocess.OutputType, handler OutputWriterHandler) *OutputWriter {
	return &OutputWriter{
		key:     key,
		typ:     typ,
		handler: handler,
		buffer:  bytes.NewBuffer(nil),
	}
}

type OutputWriter struct {
	key     string
	typ     contractsprocess.OutputType
	handler OutputWriterHandler
	buffer  *bytes.Buffer
}

func (w *OutputWriter) Write(p []byte) (n int, err error) {
	n = len(p)

	if _, err := w.buffer.Write(p); err != nil {
		return 0, err
	}

	var line []byte
	for {
		line, err = w.buffer.ReadBytes('\n')

		if err == io.EOF {
			// No complete line found, put data back and return
			w.buffer.Write(line)
			return n, nil
		}

		if err != nil {
			return n, err
		}

		// We have a complete line (including the newline)
		// Remove the trailing newline before sending to handler
		line = line[:len(line)-1]

		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)

		w.handler(w.key, w.typ, lineCopy)
	}
}
