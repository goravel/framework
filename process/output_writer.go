package process

import contractsprocess "github.com/goravel/framework/contracts/process"

func NewOutputWriter(typ contractsprocess.OutputType, handler contractsprocess.OnOutputFunc) *OutputWriter {
	return &OutputWriter{
		handler: handler,
		typ:     typ,
	}
}

type OutputWriter struct {
	typ     contractsprocess.OutputType
	handler contractsprocess.OnOutputFunc
}

func (w *OutputWriter) Write(p []byte) (n int, err error) {
	w.handler(w.typ, p)
	return len(p), nil
}
