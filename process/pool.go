package process

import (
	"context"
	"io"
	"maps"
	"strconv"
	"sync"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
	"github.com/goravel/framework/errors"
)

var _ contractsprocess.PoolBuilder = (*PoolBuilder)(nil)
var _ contractsprocess.Pool = (*Pool)(nil)
var _ contractsprocess.PoolCommand = (*PoolCommand)(nil)

type PoolBuilder struct {
	concurrency int
	ctx         context.Context
	onOutput    contractsprocess.OnPoolOutputFunc
	timeout     time.Duration

	poolConfigurer func(pool contractsprocess.Pool)
}

func NewPool() *PoolBuilder {
	return &PoolBuilder{ctx: context.Background()}
}

func (r *PoolBuilder) Concurrency(n int) contractsprocess.PoolBuilder {
	r.concurrency = n
	return r
}

func (r *PoolBuilder) OnOutput(handler contractsprocess.OnPoolOutputFunc) contractsprocess.PoolBuilder {
	r.onOutput = handler
	return r
}

func (r *PoolBuilder) Pool(configurer func(pool contractsprocess.Pool)) contractsprocess.PoolBuilder {
	r.poolConfigurer = configurer
	return r
}

func (r *PoolBuilder) Run() (map[string]contractsprocess.Result, error) {
	run, err := r.start(r.poolConfigurer)
	if err != nil {
		return nil, err
	}
	return run.Wait(), nil
}

func (r *PoolBuilder) Start() (contractsprocess.RunningPool, error) {
	return r.start(r.poolConfigurer)
}

func (r *PoolBuilder) Timeout(timeout time.Duration) contractsprocess.PoolBuilder {
	r.timeout = timeout
	return r
}

func (r *PoolBuilder) WithContext(ctx context.Context) contractsprocess.PoolBuilder {
	if ctx == nil {
		ctx = context.Background()
	}
	r.ctx = ctx
	return r
}

type job struct {
	id      int
	command *PoolCommand
}

type result struct {
	key string
	res contractsprocess.Result
}

// start initiates the execution of all configured commands concurrently but does not wait for them to complete.
// It orchestrates a pool of worker goroutines to process commands up to the specified concurrency limit.
//
// This method is non-blocking. It returns a RunningPool instance immediately, which can be used to
// wait for the completion of all processes and retrieve their results.
//
// The core concurrency pattern is as follows:
//  1. A job channel (`jobCh`) distributes commands to a pool of worker goroutines.
//  2. A result channel (`resultCh`) collects the outcome of each command from a dedicated waiter goroutine.
//  3. A separate "collector" goroutine safely populates the final results map from the result channel.
//  4. WaitGroups synchronize the completion of all workers and the collection of all results
//     before the entire operation is marked as "done".
func (r *PoolBuilder) start(configurer func(contractsprocess.Pool)) (contractsprocess.RunningPool, error) {
	if configurer == nil {
		return nil, errors.ProcessPoolNilConfigurer
	}

	pool := &Pool{}
	configurer(pool)

	commands := pool.commands
	if len(commands) == 0 {
		return nil, errors.ProcessPipelineEmpty
	}

	ctx := r.ctx
	var cancel context.CancelFunc
	if r.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
	}

	concurrency := r.concurrency
	if concurrency <= 0 || concurrency > len(commands) {
		concurrency = len(commands)
	}

	jobCh := make(chan job, len(commands))
	resultCh := make(chan result, len(commands))
	done := make(chan struct{})

	results := make(map[string]contractsprocess.Result, len(commands))
	runningProcesses := make([]contractsprocess.Running, len(commands))
	keys := make([]string, len(commands))

	var resultsWg sync.WaitGroup
	var workersWg sync.WaitGroup
	var startsWg sync.WaitGroup
	var mu sync.Mutex

	// The results collector goroutine centralizes writing to the results map
	// to avoid race conditions, as map writes are not concurrent-safe.
	// It waits for all expected results before exiting.
	resultsWg.Add(len(commands))
	go func() {
		for i := 0; i < len(commands); i++ {
			rc := <-resultCh
			mu.Lock()
			results[rc.key] = rc.res
			mu.Unlock()
			resultsWg.Done()
		}
	}()

	for i := 0; i < concurrency; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			for currentJob := range jobCh {
				command := currentJob.command
				cmdCtx := command.ctx
				if cmdCtx == nil {
					cmdCtx = ctx
				}

				proc := New().WithContext(cmdCtx).Path(command.path).Env(command.env).Input(command.input)
				if command.quietly {
					proc = proc.Quietly()
				}
				if !command.buffering {
					proc = proc.DisableBuffering()
				}
				if command.timeout > 0 {
					proc = proc.Timeout(command.timeout)
				}
				if r.onOutput != nil {
					proc = proc.OnOutput(func(typ contractsprocess.OutputType, line []byte) {
						r.onOutput(typ, line, command.key)
					})
				}

				run, err := proc.Start(command.name, command.args...)

				if err != nil {
					resultCh <- result{key: command.key, res: NewResult(err, -1, command.name, "", "")}
				} else {
					mu.Lock()
					runningProcesses[currentJob.id] = run
					mu.Unlock()

					// Launch a dedicated goroutine to wait for the process to finish.
					// This prevents the worker from being blocked by a long-running process
					// and allows it to immediately pick up the next job from jobCh.
					go func(p contractsprocess.Running, k string) {
						res := p.Wait()
						resultCh <- result{key: k, res: res}
					}(run, command.key)
				}

				// Signal that this process has completed its start attempt
				startsWg.Done()
			}
		}()
	}

	startsWg.Add(len(commands))
	for i, command := range commands {
		keys[i] = command.key
		jobCh <- job{id: i, command: command}
	}
	close(jobCh)

	// Wait for all processes to complete their start attempts before returning.
	// This ensures the runningProcesses slice is fully populated and safe to access.
	startsWg.Wait()

	// This goroutine orchestrates the clean shutdown. It waits for all workers
	// to finish processing jobs, then waits for all results to be collected.
	// Finally, it cancels the context (if a timeout was set) and signals
	// completion by closing the `done` channel.
	go func() {
		workersWg.Wait()
		resultsWg.Wait()
		if cancel != nil {
			cancel()
		}
		close(done)
	}()

	return NewRunningPool(runningProcesses, keys, cancel, results, done), nil
}

type Pool struct {
	commands []*PoolCommand
}

func (r *Pool) Command(name string, args ...string) contractsprocess.PoolCommand {
	command := NewPoolCommand(strconv.Itoa(len(r.commands)), name, args)
	r.commands = append(r.commands, command)
	return command
}

type PoolCommand struct {
	args      []string
	buffering bool
	ctx       context.Context
	env       map[string]string
	input     io.Reader
	key       string
	name      string
	path      string
	quietly   bool
	timeout   time.Duration
}

func NewPoolCommand(key, name string, args []string) *PoolCommand {
	return &PoolCommand{
		key:       key,
		name:      name,
		args:      args,
		buffering: true,
		env:       make(map[string]string),
	}
}

func (r *PoolCommand) As(key string) contractsprocess.PoolCommand {
	r.key = key
	return r
}

func (r *PoolCommand) DisableBuffering() contractsprocess.PoolCommand {
	r.buffering = false
	return r
}

func (r *PoolCommand) Env(vars map[string]string) contractsprocess.PoolCommand {
	maps.Copy(r.env, vars)
	return r
}

func (r *PoolCommand) Input(in io.Reader) contractsprocess.PoolCommand {
	r.input = in
	return r
}

func (r *PoolCommand) Path(path string) contractsprocess.PoolCommand {
	r.path = path
	return r
}

func (r *PoolCommand) Quietly() contractsprocess.PoolCommand {
	r.quietly = true
	return r
}

func (r *PoolCommand) Timeout(timeout time.Duration) contractsprocess.PoolCommand {
	r.timeout = timeout
	return r
}

func (r *PoolCommand) WithContext(ctx context.Context) contractsprocess.PoolCommand {
	r.ctx = ctx
	return r
}
