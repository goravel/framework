package telemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
)

func TestTelemetryRunner(t *testing.T) {
	t.Run("signature", func(t *testing.T) {
		runner := &TelemetryRunner{}
		assert.Equal(t, "telemetry", runner.Signature())
	})

	t.Run("shutdown priority", func(t *testing.T) {
		runner := &TelemetryRunner{}
		assert.Equal(t, 100, runner.ShutdownPriority())
	})

	t.Run("should run when telemetry facade set", func(t *testing.T) {
		runner := NewTelemetryRunner(nil, mockstelemetry.NewTelemetry(t))
		assert.True(t, runner.ShouldRun())
	})

	t.Run("should not run when telemetry facade not set", func(t *testing.T) {
		runner := NewTelemetryRunner(nil, nil)
		assert.False(t, runner.ShouldRun())
	})

	t.Run("shutdown unblocks run", func(t *testing.T) {
		telemetry := mockstelemetry.NewTelemetry(t)
		telemetry.EXPECT().Shutdown(mock.Anything).Return(nil).Once()

		runner := NewTelemetryRunner(nil, telemetry)

		ran := make(chan error, 1)
		go func() { ran <- runner.Run() }()

		assert.NoError(t, runner.Shutdown())

		select {
		case err := <-ran:
			assert.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("Run did not unblock after Shutdown")
		}
	})

	t.Run("shutdown uses configured timeout", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		config.EXPECT().GetDuration("telemetry.shutdown_timeout", defaultShutdownTimeout).Return(2 * time.Second).Once()

		telemetry := mockstelemetry.NewTelemetry(t)
		telemetry.EXPECT().Shutdown(mock.Anything).Return(assert.AnError).Once()

		runner := NewTelemetryRunner(config, telemetry)
		assert.Equal(t, assert.AnError, runner.Shutdown())
	})
}
