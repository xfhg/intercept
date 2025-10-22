package tasker

import (
	"context"
	"log"
	"sync"
	"time"
)

// Option mirrors the small subset of options used by intercept.
type Option struct {
	Verbose bool
	Tz      string
}

// Tasker provides a minimal shim of the gronx tasker used by intercept.
//
// Behavior:
//   - Task(schedule, fn): registers the task and runs it once after a short delay.
//     The schedule string is accepted but ignored in this shim.
//   - Run(): blocks forever (or until Stop is called). Tasks already run in their
//     own goroutines so Run only needs to keep the program alive.
type Tasker struct {
	opt   Option
	mu    sync.Mutex
	tasks []struct {
		schedule string
		fn       func(context.Context) (int, error)
	}
	stop chan struct{}
}

// New returns a new Tasker shim.
func New(opt Option) *Tasker {
	return &Tasker{
		opt:  opt,
		stop: make(chan struct{}),
		tasks: make([]struct {
			schedule string
			fn       func(context.Context) (int, error)
		}, 0),
	}
}

// Task registers a task and runs it once after a short delay. The schedule
// string is accepted but ignored in this shim.
func (t *Tasker) Task(schedule string, fn func(context.Context) (int, error)) {
	entry := struct {
		schedule string
		fn       func(context.Context) (int, error)
	}{schedule: schedule, fn: fn}

	t.mu.Lock()
	t.tasks = append(t.tasks, entry)
	t.mu.Unlock()

	// Run task once after a short delay so startup scheduling still triggers.
	go func(e func(context.Context) (int, error)) {
		// Small stagger to avoid races
		time.Sleep(1 * time.Second)
		ctx := context.Background()
		if _, err := e(ctx); err != nil {
			log.Printf("task execution error: %v", err)
		}
	}(fn)
}

// Run blocks until stop is closed. This keeps the process alive.
func (t *Tasker) Run() {
	<-t.stop
}

// Stop stops the Run loop (not used by intercept but provided for completeness).
func (t *Tasker) Stop() {
	select {
	case <-t.stop:
		// already closed
	default:
		close(t.stop)
	}
}
