package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

func TestCallAll(t *testing.T) {
	t.Run("executes all functions and skips nil", func(t *testing.T) {
		callCount := 0

		fn := func(ctx context.Context) error {
			callCount++
			return nil
		}

		err := callAll(context.Background(), []func(context.Context) error{fn, nil, fn})
		assert.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("aggregates errors", func(t *testing.T) {
		err := callAll(context.Background(), []func(context.Context) error{
			func(ctx context.Context) error { return errors.New("error 1") },
			func(ctx context.Context) error { return nil },
			func(ctx context.Context) error { return errors.New("error 2") },
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error 1")
		assert.Contains(t, err.Error(), "error 2")
	})
}
