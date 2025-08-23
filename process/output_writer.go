package process

import (
	"bufio"
	"bytes"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func NewOutputWriter(typ contractsprocess.OutputType, handler contractsprocess.OnOutputFunc) *OutputWriter {
	return &OutputWriter{
		typ:     typ,
		handler: handler,
		buffer:  &bytes.Buffer{},
	}
}

type OutputWriter struct {
	typ     contractsprocess.OutputType
	handler contractsprocess.OnOutputFunc
	buffer  *bytes.Buffer
}

func (w *OutputWriter) Write(p []byte) (n int, err error) {
	n, err = w.buffer.Write(p)
	if err != nil {
		return n, err
	}

	scanner := bufio.NewScanner(w.buffer)
	for scanner.Scan() {
		line := scanner.Bytes()
		copied := make([]byte, len(line))
		copy(copied, line)
		w.handler(w.typ, copied)
	}

	rest := w.buffer.Bytes()
	w.buffer.Reset()
	w.buffer.Write(rest)

	return n, err
}
