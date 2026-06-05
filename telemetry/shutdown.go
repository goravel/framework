package telemetry

import (
	"context"

	"github.com/goravel/framework/errors"
)

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

func callAll(ctx context.Context, fns []func(context.Context) error) error {
	var errs []error

	for _, fn := range fns {
		if fn == nil {
			continue
		}
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
