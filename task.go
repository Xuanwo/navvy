package navvy

import "context"

// Task is the acceptable input for navvy.
type Task interface {
	Context() context.Context // Context used to get context if no ctx passed when run
	Run(ctx context.Context)
}

type taskWithFunc struct {
	ctx context.Context
	fn  func(ctx context.Context)
}

func (t *taskWithFunc) Run(ctx context.Context) {
	if ctx == nil {
		ctx = t.Context()
	}
	t.fn(ctx)
}

func (t *taskWithFunc) Context() context.Context {
	if t.ctx == nil {
		return context.Background()
	}
	return t.ctx
}

// TaskWrapper will wrapper a func into a valid task.
func TaskWrapper(fn func(ctx context.Context)) Task {
	return &taskWithFunc{nil, fn}
}
