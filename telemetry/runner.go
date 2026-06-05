package telemetry

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/config"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
)

const (
	runnerShutdownPriority = 100
	defaultShutdownTimeout = 15 * time.Second
)

var _ contractsfoundation.RunnerWithShutdownPriority = (*Runner)(nil)

type Runner struct {
	config    config.Config
	telemetry contractstelemetry.Telemetry
	done      chan struct{}
}

func NewRunner(config config.Config, telemetry contractstelemetry.Telemetry) *Runner {
	return &Runner{
		config:    config,
		telemetry: telemetry,
		done:      make(chan struct{}),
	}
}

func (r *Runner) Run() error {
	<-r.done
	return nil
}

func (r *Runner) ShouldRun() bool {
	return r.telemetry != nil
}

func (r *Runner) Shutdown() error {
	defer close(r.done)

	timeout := defaultShutdownTimeout
	if r.config != nil {
		timeout = r.config.GetDuration("telemetry.shutdown_timeout", defaultShutdownTimeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return r.telemetry.Shutdown(ctx)
}

func (r *Runner) ShutdownPriority() int {
	return runnerShutdownPriority
}

func (r *Runner) Signature() string {
	return "telemetry"
}
