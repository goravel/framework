package process

import "github.com/goravel/framework/contracts/process"

func NewOutputWriter(typ process.OutputType, handler func(typ process.OutputType, line string)) *OutputWriter {
	return &OutputWriter{
		handler: handler,
		typ:     typ,
	}
}

type OutputWriter struct {
	typ     process.OutputType
	handler func(typ process.OutputType, line string)
}

func (w *OutputWriter) Write(p []byte) (n int, err error) {
	w.handler(w.typ, string(p))
	return len(p), nil
}
