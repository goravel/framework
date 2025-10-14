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
	concurrency int
	ctx         context.Context
	onOutput    contractsprocess.OnPoolOutputFunc
	timeout     time.Duration
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

func (r *PoolBuilder) Run(configure func(contractsprocess.Pool)) (map[string]contractsprocess.Result, error) {
	return r.run(configure)
}

func (r *PoolBuilder) Start(configure func(contractsprocess.Pool)) (contractsprocess.RunningPool, error) {
	return r.start(configure)
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

	steps := pool.steps
	if len(steps) == 0 {
		return nil, errors.ProcessPipelineEmpty
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
				proc := New().WithContext(ctx).Path(step.path).Env(step.env).Input(step.input)
				if step.quietly {
					proc = proc.Quietly()
				}
				if step.timeout > 0 {
					proc = proc.Timeout(step.timeout)
				}
				if r.onOutput != nil {
					proc = proc.OnOutput(func(typ contractsprocess.OutputType, line []byte) {
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
	for k, v := range vars {
		r.env[k] = v
	}
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
