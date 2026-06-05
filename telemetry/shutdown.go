package telemetry

import "context"

type (
	ShutdownFunc = func(context.Context) error
	FlushFunc    = func(context.Context) error
)

// NoopShutdown returns a ShutdownFunc that does nothing and returns nil.
func NoopShutdown() ShutdownFunc {
	return func(context.Context) error { return nil }
}

// NoopFlush returns a FlushFunc that does nothing and returns nil.
func NoopFlush() FlushFunc {
	return func(context.Context) error { return nil }
}
