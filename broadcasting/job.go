package broadcasting

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type BroadcastJob struct {
	config contractsconfig.Config
	app    foundation.Application

	// item is the payload of the task being processed, set by Handle and
	// read by ShouldRetry. BroadcastJob is a shared singleton, so access is
	// guarded by mu.
	//
	// Limitation: the mutex guarantees memory safety, not logical isolation.
	// ShouldRetry has no access to task args by contract, so a concurrent
	// failed task can still overwrite this payload between another task's
	// Handle returning an error and its ShouldRetry call. Clearing item on
	// Handle's success path converts the common interleaving case into the
	// safe single-shot fallback; a full fix requires per-task state passed by
	// the queue worker, which is out of scope.
	mu   sync.Mutex
	item *broadcastItem
}

func (j *BroadcastJob) Signature() string {
	return "goravel_broadcast"
}

// Handle executes the queued broadcast. Retries are governed by ShouldRetry
// using the event's BroadcastTries/BroadcastBackoff captured at dispatch time.
// Broadcasts without BroadcastTries defer to the queue worker's tries config.
// A fresh context is synthesized (the dispatch-time ctx cannot cross the queue
// boundary); if the originating event implemented ShouldBroadcastWithTimeout
// the worker context is bounded accordingly, which the Pusher driver honours
// via WithContext.
func (j *BroadcastJob) Handle(args ...any) error {
	if len(args) != 1 {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return errors.BroadcastInvalidQueuePayload
	}

	raw, ok := args[0].(string)
	if !ok {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return errors.BroadcastInvalidQueuePayload
	}

	var item broadcastItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return errors.BroadcastInvalidQueuePayload
	}

	j.mu.Lock()
	j.item = &item
	j.mu.Unlock()

	cfg, err := NewConfig(j.config)
	if err != nil {
		j.mu.Lock()
		j.item = nil
		j.mu.Unlock()
		return err
	}

	conns := item.Connections
	if len(conns) == 0 {
		conns = []string{cfg.DefaultConnection()}
	}

	ctx, cancel := withTimeout(context.Background(), item.Timeout)
	defer cancel()

	if err := j.broadcastToConns(ctx, cfg, item.Channels, item.Event, item.Payload, conns); err != nil {
		return err
	}

	// A successful task is never consulted by ShouldRetry, so release the
	// payload: an interleaving concurrent failed task then reads the safe
	// single-shot fallback (item == nil) instead of a wrong retry policy.
	j.mu.Lock()
	j.item = nil
	j.mu.Unlock()

	return nil
}

// ShouldRetry controls retries for the queued broadcast. A
// BroadcastTries-declaring event wins; without one (Tries <= 0) the queue
// worker's maxTries governs. With a cap it retries while
// attempt < Tries using the configured per-attempt Backoff (last value
// repeats); backoff applies to worker-driven retries too. item == nil (cleared
// on success/invalid payload) is the safe single-shot fallback — never the
// worker's tries, so an interleaved concurrent task can't inherit another
// task's policy.
func (j *BroadcastJob) ShouldRetry(err error, attempt, maxTries int) (retryable bool, delay time.Duration) {
	j.mu.Lock()
	item := j.item
	j.mu.Unlock()

	// item == nil = no task policy (cleared on success/invalid payload): the
	// safe single-shot fallback — never the worker's tries, so an interleaved
	// concurrent task can't inherit another task's policy.
	if item == nil {
		return false, 0
	}
	if attempt < 1 {
		// Defensive: attempts come from the pop-incremented reservation (or a
		// chain counter starting at 1), so this is unreachable in practice.
		// Returning false avoids an accidental infinite retry loop.
		return false, 0
	}

	// The event's own BroadcastTries wins; without one the worker's
	// maxTries governs.
	effectiveTries := item.Tries
	if effectiveTries <= 0 {
		effectiveTries = maxTries
	}
	if attempt >= effectiveTries {
		return false, 0
	}
	if len(item.Backoff) == 0 {
		return true, 0
	}

	idx := min(attempt-1, len(item.Backoff)-1)
	return true, time.Duration(item.Backoff[idx]) * time.Millisecond
}

func (j *BroadcastJob) broadcastToConns(ctx context.Context, cfg *Config, channels []string, event string, payload map[string]any, conns []string) error {
	for _, conn := range conns {
		cfgConn, err := cfg.Connection(conn)
		if err != nil {
			return err
		}

		driver, err := CreateDriver(cfgConn, j.app)
		if err != nil {
			return err
		}

		if err := driver.Broadcast(ctx, channels, event, payload); err != nil {
			return err
		}
	}

	return nil
}
