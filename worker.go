package navvy

import (
	"log"
	"runtime"
)

// Worker is the worker for navvy.
type Worker struct {
	// pool who owns this worker.
	pool *Pool

	// task is a job should be done.
	task chan Task
}

func (w *Worker) run() {
	w.pool.incRunning()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				w.pool.decRunning()
				w.pool.workerCache.Put(w)
				log.Printf("worker exits from a panic: %v\n", p)

				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				log.Printf("worker exits from panic: %s\n", string(buf[:n]))
			}
		}()

		for f := range w.task {
			if f == nil {
				w.pool.decRunning()
				w.pool.workerCache.Put(w)
				return
			}

			f.Run()
			w.pool.wg.Done()

			if ok := w.pool.revertWorker(w); !ok {
				break
			}
		}
	}()
}
