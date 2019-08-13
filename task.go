package navvy

// Task is the acceptable input for navvy.
type Task interface {
	Run()
}

type taskWithFunc struct {
	fn func()
}

func (t *taskWithFunc) Run() {
	t.fn()
}

// TaskWrapper will wrapper a func into a valid task.
func TaskWrapper(fn func()) Task {
	return &taskWithFunc{fn}
}
