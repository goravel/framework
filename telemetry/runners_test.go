package telemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
)

func TestTelemetryRunner_Signature(t *testing.T) {
	assert.Equal(t, "telemetry", NewTelemetryRunner(nil, nil).Signature())
}

func TestTelemetryRunner_ShouldRun(t *testing.T) {
	assert.False(t, NewTelemetryRunner(nil, nil).ShouldRun())
	assert.True(t, NewTelemetryRunner(nil, mockstelemetry.NewTelemetry(t)).ShouldRun())
}

func TestTelemetryRunner_ShutdownPriority(t *testing.T) {
	assert.Equal(t, 100, NewTelemetryRunner(nil, nil).ShutdownPriority())
}

func TestTelemetryRunner_Shutdown_UnblocksRun(t *testing.T) {
	mockTelemetry := mockstelemetry.NewTelemetry(t)
	mockTelemetry.EXPECT().Shutdown(mock.Anything).Return(nil).Once()

	runner := NewTelemetryRunner(nil, mockTelemetry)

	ran := make(chan error, 1)
	go func() { ran <- runner.Run() }()

	assert.NoError(t, runner.Shutdown())

	select {
	case err := <-ran:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Run did not unblock after Shutdown")
	}
}

func TestTelemetryRunner_Shutdown_UsesConfiguredTimeout(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetDuration("telemetry.shutdown_timeout", defaultShutdownTimeout).Return(2 * time.Second).Once()

	mockTelemetry := mockstelemetry.NewTelemetry(t)
	mockTelemetry.EXPECT().Shutdown(mock.Anything).Return(assert.AnError).Once()

	runner := NewTelemetryRunner(mockConfig, mockTelemetry)
	assert.Equal(t, assert.AnError, runner.Shutdown())
}
