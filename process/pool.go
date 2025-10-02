package process

import (
	"context"
	"io"
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
	ctx         context.Context
	timeout     time.Duration
	concurrency int
	onOutput    contractsprocess.OnPoolOutputFunc
	strategy    contractsprocess.Strategy
}

func NewPool() *PoolBuilder {
	return &PoolBuilder{ctx: context.Background()}
}

func (r *PoolBuilder) WithConcurrency(n int) contractsprocess.PoolBuilder {
	r.concurrency = n
	return r
}

func (r *PoolBuilder) WithTimeout(timeout time.Duration) contractsprocess.PoolBuilder {
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

func (r *PoolBuilder) WithStrategy(strategy contractsprocess.Strategy) contractsprocess.PoolBuilder {
	r.strategy = strategy
	return r
}

func (r *PoolBuilder) WithOutputHandler(handler contractsprocess.OnPoolOutputFunc) contractsprocess.PoolBuilder {
	r.onOutput = handler
	return r
}

func (r *PoolBuilder) Run(configure func(contractsprocess.Pool)) (map[string]contractsprocess.Result, error) {
	return r.run(configure)
}

func (r *PoolBuilder) Start(configure func(contractsprocess.Pool)) (contractsprocess.RunningPool, error) {
	return r.start(configure)
}

func (r *PoolBuilder) run(configure func(pool contractsprocess.Pool)) (map[string]contractsprocess.Result, error) {
	run, err := r.start(configure)
	if err != nil {
		return nil, err
	}
	return run.Wait(), nil
}

func (r *PoolBuilder) start(configure func(contractsprocess.Pool)) (contractsprocess.RunningPool, error) {
	pool := &Pool{}
	configure(pool)

	originalSteps := pool.steps
	if len(originalSteps) == 0 {
		return nil, errors.ProcessPipelineEmpty
	}

	schedules := make([]contractsprocess.Schedulable, len(originalSteps))
	for i, cmd := range originalSteps {
		schedules[i] = cmd
	}
	scheduledSteps := r.strategy.Schedule(schedules)
	steps := make([]*PoolCommand, len(scheduledSteps))
	for i, s := range scheduledSteps {
		steps[i] = s.(*PoolCommand)
	}

	ctx := r.ctx
	var cancel context.CancelFunc
	if r.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
	}

	concurrency := r.concurrency
	if concurrency <= 0 || concurrency > len(steps) {
		concurrency = len(steps)
	}

	type job struct {
		id   int
		step *PoolCommand
	}
	type result struct {
		key string
		res contractsprocess.Result
	}

	jobCh := make(chan job, len(steps))
	resultCh := make(chan result, len(steps))
	done := make(chan struct{})

	results := make(map[string]contractsprocess.Result, len(steps))
	runningProcesses := make([]contractsprocess.Running, len(steps))
	keys := make([]string, len(steps))

	var resultsWg sync.WaitGroup
	var workersWg sync.WaitGroup
	var mu sync.Mutex

	resultsWg.Add(len(steps))
	go func() {
		for i := 0; i < len(steps); i++ {
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
				step := currentJob.step
				proc := New().WithContext(ctx).WithPath(step.path).WithEnv(step.env).WithInput(step.input)
				if step.quietly {
					proc = proc.WithQuiet()
				}
				if step.timeout > 0 {
					proc = proc.WithTimeout(step.timeout)
				}
				if r.onOutput != nil {
					proc = proc.WithOutputHandler(func(typ contractsprocess.OutputType, line []byte) {
						r.onOutput(step.key, typ, line)
					})
				}

				run, err := proc.Start(step.name, step.args...)

				if err != nil {
					resultCh <- result{key: step.key, res: NewResult(err, -1, step.name, "", "")}
				} else {
					mu.Lock()
					runningProcesses[currentJob.id] = run
					mu.Unlock()

					go func(p contractsprocess.Running, k string) {
						res := p.Wait()
						resultCh <- result{key: k, res: res}
					}(run, step.key)
				}
			}
		}()
	}

	for i, step := range steps {
		keys[i] = step.key
		jobCh <- job{id: i, step: step}
	}
	close(jobCh)

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
	steps []*PoolCommand
}

func (r *Pool) Command(name string, args ...string) contractsprocess.PoolCommand {
	step := NewPoolCommand(strconv.Itoa(len(r.steps)), name, args)
	r.steps = append(r.steps, step)
	return step
}

type PoolCommand struct {
	key              string
	name             string
	args             []string
	ctx              context.Context
	timeout          time.Duration
	path             string
	env              map[string]string
	input            io.Reader
	quietly          bool
	disableBuffering bool
	priority         contractsprocess.Priority
}

func NewPoolCommand(key, name string, args []string) *PoolCommand {
	return &PoolCommand{
		key:      key,
		name:     name,
		args:     args,
		priority: contractsprocess.PriorityNormal,
	}
}

func (r *PoolCommand) GetKey() string {
	return r.key
}

func (r *PoolCommand) GetTimeout() time.Duration {
	return r.timeout
}

func (r *PoolCommand) GetPriority() contractsprocess.Priority {
	return r.priority
}

func (r *PoolCommand) WithPriority(priority contractsprocess.Priority) contractsprocess.PoolCommand {
	r.priority = priority
	return r
}

func (r *PoolCommand) WithKey(key string) contractsprocess.PoolCommand {
	r.key = key
	return r
}

func (r *PoolCommand) WithContext(ctx context.Context) contractsprocess.PoolCommand {
	r.ctx = ctx
	return r
}

func (r *PoolCommand) WithTimeout(timeout time.Duration) contractsprocess.PoolCommand {
	r.timeout = timeout
	return r
}

func (r *PoolCommand) WithPath(path string) contractsprocess.PoolCommand {
	r.path = path
	return r
}

func (r *PoolCommand) WithEnv(vars map[string]string) contractsprocess.PoolCommand {
	for k, v := range vars {
		r.env[k] = v
	}
	return r
}

func (r *PoolCommand) WithQuiet() contractsprocess.PoolCommand {
	r.quietly = true
	return r
}

func (r *PoolCommand) WithDisabledBuffering() contractsprocess.PoolCommand {
	r.disableBuffering = true
	return r
}

func (r *PoolCommand) WithInput(in io.Reader) contractsprocess.PoolCommand {
	r.input = in
	return r
}
