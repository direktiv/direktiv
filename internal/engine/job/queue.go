package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ID represents the unique identifier of a queued job.
type ID = uuid.UUID

// Payload is an opaque job payload. In your engine, this would carry the script,
// mappings, function, input and metadata required to run the JS.
type Payload any

// Result is the result emitted by a job. In your engine, this would be the
// exported JS object or an error.
type Result any

// Runner executes a job. Implement this in your engine to call execJSScript
// (one fresh VM per job). It must be safe for concurrent use by many workers.
type Runner interface {
	Run(ctx context.Context, id ID, p Payload) (Result, error)
}

// HookSet carries optional callbacks for observability and policy.
type HookSet struct {
	// OnEnqueue is called after a job is accepted into the queue.
	OnEnqueue func(id ID, p Payload)

	// OnStart is called when a worker starts running a job.
	OnStart func(id ID, p Payload)

	// OnFinish is called when a worker finishes a job (success or error).
	OnFinish func(id ID, p Payload, res Result, err error, dur time.Duration)

	// OnDrop is called when a job cannot be enqueued due to a full queue and
	// DropIfFull option is enabled.
	OnDrop func(id ID, p Payload)
}

// Options configure the Manager.
type Options struct {
	// Workers is the number of worker goroutines. If zero, defaults to runtime.NumCPU().
	Workers int

	// QueueSize is the bounded capacity of the queue. If zero, defaults to 2*Workers.
	QueueSize int

	// DropIfFull controls behavior when the queue is full:
	//   - if true, Enqueue returns ErrQueueFull immediately.
	//   - if false (default), Enqueue blocks until space is available or ctx is canceled.
	DropIfFull bool

	// Logger is optional; if nil a no-op logger is used.
	Logger *slog.Logger

	// Hooks are optional observability callbacks.
	Hooks HookSet
}

// ErrQueueFull is returned by Enqueue when DropIfFull is true and the queue is full.
var ErrQueueFull = errors.New("job queue full")

// ErrManagerClosed is returned when Enqueue is called after Stop.
var ErrManagerClosed = errors.New("job manager closed")

// ErrManagerNotStarted is returned when manager is not started.
var ErrManagerNotStarted = errors.New("job manager not started")

type job struct {
	id      ID
	payload Payload
	// deadline or timeout is conveyed via ctx.
	ctx context.Context
	// done is closed when the job finishes (used in tests)
	done chan struct{}
	// result fields
	res Result
	err error
}

// Manager is a bounded job queue with a fixed-size worker pool.
type Manager struct {
	runner Runner
	opts   Options

	log *slog.Logger

	mu      sync.RWMutex
	started bool
	closed  bool

	// queue and workers
	queue chan *job
	wg    sync.WaitGroup
	// cancel stops workers
	cancel context.CancelFunc
	// root context for workers
	rootCtx context.Context
}

// NewManager constructs a Manager.
func NewManager(runner Runner, opts Options) *Manager {
	if opts.Workers <= 0 {
		opts.Workers = runtime.NumCPU()
	}
	if opts.QueueSize <= 0 {
		opts.QueueSize = opts.Workers * 2
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}

	return &Manager{
		runner: runner,
		opts:   opts,
		log:    opts.Logger,
		queue:  make(chan *job, opts.QueueSize),
	}
}

// Start launches the worker goroutines. It is safe to call once.
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return
	}
	m.rootCtx, m.cancel = context.WithCancel(context.Background())
	for i := range m.opts.Workers {
		m.wg.Add(1)
		go m.worker(i)
	}
	m.started = true
}

// Stop gracefully stops workers after draining queued jobs.
// No new jobs are accepted after Stop returns.
func (m *Manager) Stop() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true
	cancel := m.cancel
	close(m.queue) // signal no more jobs
	m.mu.Unlock()

	// stop workers
	if cancel != nil {
		cancel()
	}
	m.wg.Wait()
}

// Enqueue submits a job to the queue. If DropIfFull is false, it blocks until
// space is available or ctx is canceled. If DropIfFull is true and the queue is
// full, it returns ErrQueueFull immediately.
func (m *Manager) Enqueue(ctx context.Context, p Payload) (ID, error) {
	j := &job{
		id:      uuid.New(),
		payload: p,
		ctx:     ctx,
		done:    make(chan struct{}),
	}

	m.mu.RLock()
	closed := m.closed
	started := m.started
	m.mu.RUnlock()

	if closed {
		return uuid.Nil, ErrManagerClosed
	}
	if !started {
		return uuid.Nil, ErrManagerNotStarted
	}

	select {
	case m.queue <- j:
		if m.opts.Hooks.OnEnqueue != nil {
			m.opts.Hooks.OnEnqueue(j.id, j.payload)
		}

		return j.id, nil
	default:
		// full
		if m.opts.DropIfFull {
			if m.opts.Hooks.OnDrop != nil {
				m.opts.Hooks.OnDrop(j.id, j.payload)
			}

			return uuid.Nil, ErrQueueFull
		}
		// block until space, or ctx canceled, or manager closed
		select {
		case m.queue <- j:
			if m.opts.Hooks.OnEnqueue != nil {
				m.opts.Hooks.OnEnqueue(j.id, j.payload)
			}

			return j.id, nil
		case <-ctx.Done():
			return uuid.Nil, ctx.Err()
		}
	}
}

// worker executes jobs until Stop is called and the queue is drained.
func (m *Manager) worker(idx int) {
	defer m.wg.Done()
	for {
		select {
		case <-m.rootCtx.Done():
			// Still drain remaining jobs if queue is closed.
			// But if queue is open, we exit on cancel signal.
			if !m.isQueueClosed() {
				return
			}
		default:
		}

		j, ok := <-m.queue
		if !ok {
			// queue closed and drained
			return
		}

		start := time.Now()
		if m.opts.Hooks.OnStart != nil {
			m.opts.Hooks.OnStart(j.id, j.payload)
		}

		// Run with panic containment
		func() {
			defer func() {
				if r := recover(); r != nil {
					j.err = recoverToError(r)
				}
			}()
			ctx := j.ctx
			if ctx == nil {
				ctx = m.rootCtx
			}
			j.res, j.err = m.runner.Run(ctx, j.id, j.payload)
		}()

		if m.opts.Hooks.OnFinish != nil {
			m.opts.Hooks.OnFinish(j.id, j.payload, j.res, j.err, time.Since(start))
		}
		close(j.done)
	}
}

func (m *Manager) isQueueClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.closed
}

func recoverToError(r any) error {
	switch v := r.(type) {
	case error:
		return v
	default:
		return errors.New("panic: " + toString(v))
	}
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmtAny(t)
	}
}

// fmtAny is isolated to avoid fmt import at top-level impacting allocations in hot path.
func fmtAny(v any) string {
	return sprint(v)
}

//go:noinline
func sprint(v any) string {
	return fmt.Sprint(v)
}
