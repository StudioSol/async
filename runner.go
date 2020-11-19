package async

import (
	"context"
	"sync"
)

type Runner struct {
	sync.Mutex
	tasks      []Task
	errs       []error
	limit      int
	waitErrors bool
}

// NewRunner creates a new task manager to control async functions.
func NewRunner(tasks ...Task) *Runner {
	return &Runner{
		tasks: tasks,
		limit: len(tasks),
	}
}

// WaitErrors tells the runner to wait for the response from all functions instead of cancelling them all when the first error occurs.
func (r *Runner) WaitErrors() *Runner {
	r.waitErrors = true
	return r
}

// WithLimit defines a limit for concurrent tasks execution
func (r *Runner) WithLimit(limit int) *Runner {
	r.limit = limit
	return r
}

// AllErrors returns all errors reported by functions
func (r *Runner) AllErrors() []error {
	return r.errs
}

// registerErr store an error to final report
func (r *Runner) registerErr(err error) {
	r.Lock()
	defer r.Unlock()
	if err != nil {
		r.errs = append(r.errs, err)
	}
}

// wrapperChannel converts a given Task to a channel of errors
func wrapperChannel(ctx context.Context, task Task) chan error {
	cerr := make(chan error, 1)
	go func() {
		defer safePanic(cerr)

		cerr <- task(ctx)
		close(cerr)
	}()
	return cerr
}

// Run starts the task manager and returns the first error or nil if succeed
func (r *Runner) Run(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	cerr := make(chan error, len(r.tasks))
	queue := make(chan struct{}, r.limit)
	var wg sync.WaitGroup
	wg.Add(len(r.tasks))
	for _, task := range r.tasks {
		queue <- struct{}{}
		go func(fn func(context.Context) error) {
			defer func() {
				<-queue
				wg.Done()
			}()

			select {
			case <-parentCtx.Done():
				cerr <- parentCtx.Err()
				r.registerErr(parentCtx.Err())
			case err := <-wrapperChannel(ctx, fn):
				cerr <- err
				r.registerErr(err)
			}
		}(task)
	}

	go func() {
		wg.Wait()
		cancel()
		close(cerr)
	}()

	var firstErr error
	for err := range cerr {
		if err != nil && firstErr == nil {
			firstErr = err
			if r.waitErrors {
				continue
			}
			cancel()
			return firstErr
		}
	}

	return firstErr
}
