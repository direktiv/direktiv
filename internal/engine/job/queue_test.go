package job

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

type testRunner struct {
	delay time.Duration
	panic bool
	err   error
	count atomic.Int64
}

func (tr *testRunner) Run(ctx context.Context, id ID, p Payload) (Result, error) {
	tr.count.Add(1)
	if tr.panic {
		panic("boom")
	}
	select {
	case <-time.After(tr.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if tr.err != nil {
		return nil, tr.err
	}
	return p, nil
}

func TestEnqueueAndRun(t *testing.T) {
	r := &testRunner{}
	m := NewManager(r, Options{Workers: 2, QueueSize: 4})
	m.Start()
	defer m.Stop()

	id, err := m.Enqueue(context.Background(), "hello")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if id == uuid.Nil {
		t.Fatalf("expected non-nil id")
	}
	// give it a moment to run
	time.Sleep(50 * time.Millisecond)
	if r.count.Load() != 1 {
		t.Fatalf("expected 1 run, got %d", r.count.Load())
	}
}

func TestBackpressureBlocking(t *testing.T) {
	r := &testRunner{delay: 100 * time.Millisecond}
	m := NewManager(r, Options{Workers: 1, QueueSize: 1, DropIfFull: false})
	m.Start()
	defer m.Stop()

	_, err := m.Enqueue(context.Background(), "a") // accepted
	if err != nil {
		t.Fatalf("enqueue a: %v", err)
	}
	_, err = m.Enqueue(context.Background(), "b") // accepted
	if err != nil {
		t.Fatalf("enqueue b: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	// queue now has 1 in-flight + 1 buffered. Next enqueue should block and then ctx timeout.
	_, err = m.Enqueue(ctx, "c")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestDropIfFull(t *testing.T) {
	r := &testRunner{delay: 200 * time.Millisecond}
	m := NewManager(r, Options{Workers: 1, QueueSize: 1, DropIfFull: true})
	m.Start()
	defer m.Stop()

	_, err := m.Enqueue(context.Background(), "a")
	if err != nil {
		t.Fatalf("enqueue a: %v", err)
	}
	_, err = m.Enqueue(context.Background(), "b")
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}

func TestTimeoutPropagation(t *testing.T) {
	r := &testRunner{delay: 200 * time.Millisecond}
	m := NewManager(r, Options{Workers: 1, QueueSize: 2})
	m.Start()
	defer m.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := m.Enqueue(ctx, "a")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	// allow runner to pick it up
	time.Sleep(10 * time.Millisecond)
	// The runner should return ctx.Err()
	time.Sleep(60 * time.Millisecond)

	// TODO: yassir, test job result.
}

func TestPanicRecovery(t *testing.T) {
	r := &testRunner{panic: true}
	m := NewManager(r, Options{Workers: 1, QueueSize: 2})
	m.Start()
	defer m.Stop()

	_, err := m.Enqueue(context.Background(), "a")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	// wait a bit; we only assert no crash
	time.Sleep(30 * time.Millisecond)

	// TODO: yassir, test job result.
}
