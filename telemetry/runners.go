package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/config"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
)

const (
	runnerShutdownPriority   = 100
	defaultShutdownTimeout   = 15 * time.Second
	configKeyAutoRun         = "app.auto_run"
	configKeyShutdownTimeout = "telemetry.shutdown_timeout"
)

var _ contractsfoundation.RunnerWithShutdownPriority = (*TelemetryRunner)(nil)

type TelemetryRunner struct {
	config    config.Config
	telemetry contractstelemetry.Telemetry
	done      chan struct{}
	closeOnce sync.Once
}

func NewTelemetryRunner(config config.Config, telemetry contractstelemetry.Telemetry) *TelemetryRunner {
	return &TelemetryRunner{
		config:    config,
		telemetry: telemetry,
		done:      make(chan struct{}),
	}
}

func (r *TelemetryRunner) Run() error {
	<-r.done
	return nil
}

func (r *TelemetryRunner) ShouldRun() bool {
	return r.telemetry != nil && r.config != nil &&
		r.config.GetBool(configKeyAutoRun, true)
}

func (r *TelemetryRunner) Shutdown() error {
	defer r.closeOnce.Do(func() { close(r.done) })

	timeout := defaultShutdownTimeout
	if r.config != nil {
		timeout = r.config.GetDuration(configKeyShutdownTimeout, defaultShutdownTimeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return r.telemetry.Shutdown(ctx)
}

func (r *TelemetryRunner) ShutdownPriority() int {
	return runnerShutdownPriority
}

func (r *TelemetryRunner) Signature() string {
	return "telemetry"
}
