package telemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
)

func TestRunner_Signature(t *testing.T) {
	assert.Equal(t, "telemetry", NewRunner(nil, nil).Signature())
}

func TestRunner_ShouldRun(t *testing.T) {
	assert.False(t, NewRunner(nil, nil).ShouldRun())
	assert.True(t, NewRunner(nil, mockstelemetry.NewTelemetry(t)).ShouldRun())
}

func TestRunner_ShutdownPriority(t *testing.T) {
	assert.Equal(t, 100, NewRunner(nil, nil).ShutdownPriority())
}

func TestRunner_Shutdown_UnblocksRun(t *testing.T) {
	mockTelemetry := mockstelemetry.NewTelemetry(t)
	mockTelemetry.EXPECT().Shutdown(mock.Anything).Return(nil).Once()

	runner := NewRunner(nil, mockTelemetry)

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

func TestRunner_Shutdown_UsesConfiguredTimeout(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetDuration("telemetry.shutdown_timeout", defaultShutdownTimeout).Return(2 * time.Second).Once()

	mockTelemetry := mockstelemetry.NewTelemetry(t)
	mockTelemetry.EXPECT().Shutdown(mock.Anything).Return(assert.AnError).Once()

	runner := NewRunner(mockConfig, mockTelemetry)
	assert.Equal(t, assert.AnError, runner.Shutdown())
}
