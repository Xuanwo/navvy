package navvy

import (
	"sync"
	"sync/atomic"
)

// Pool is the task pool for navvy.
type Pool struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// release is used to notice the pool to closed itself.
	release int32

	// lock for synchronous operation.
	lock sync.Mutex

	// cond for waiting to get a idle worker.
	cond *sync.Cond

	// once makes sure releasing this pool will just be done for one time.
	once sync.Once

	// workerCache speeds up the obtainment of the an usable worker in function:retrieveWorker.
	workerCache sync.Pool

	// wg is the wait group for all tasks.
	wg *sync.WaitGroup

	// workers is a slice that store the available workers.
	workers []*Worker
}

// NewPool generates an instance of ants pool.
func NewPool(size int) (*Pool, error) {
	if size <= 0 {
		return nil, ErrInvalidPoolSize
	}
	var p *Pool
	p = &Pool{
		capacity: int32(size),
		workers:  make([]*Worker, 0, size),
		wg:       &sync.WaitGroup{},
	}

	p.cond = sync.NewCond(&p.lock)
	return p, nil
}

// Submit submits a task to this pool.
func (p *Pool) Submit(task Task) error {
	if atomic.LoadInt32(&p.release) == CLOSED {
		return ErrPoolClosed
	}
	p.wg.Add(1)
	p.retrieveWorker().task <- task
	return nil
}

// Running returns the number of the currently running goroutines.
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns the available goroutines to work.
func (p *Pool) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

// Cap returns the capacity of this pool.
func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Tune changes the capacity of this pool.
func (p *Pool) Tune(size int) {
	if p.Cap() == size {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
	diff := p.Running() - size
	for i := 0; i < diff; i++ {
		p.retrieveWorker().task <- nil
	}
}

// Release Closes this pool.
func (p *Pool) Release() error {
	p.once.Do(func() {
		atomic.StoreInt32(&p.release, 1)
		p.lock.Lock()
		idleWorkers := p.workers
		for i, w := range idleWorkers {
			w.task <- nil
			idleWorkers[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
	})
	return nil
}

// Wait will wait for all task finished.
func (p *Pool) Wait() {
	p.wg.Wait()
}

// incRunning increases the number of the currently running goroutines.
func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// retrieveWorker returns a available worker to run the tasks.
func (p *Pool) retrieveWorker() *Worker {
	var w *Worker

	p.lock.Lock()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 {
		w = idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}

	if p.Running() < p.Cap() {
		p.lock.Unlock()
		if cacheWorker := p.workerCache.Get(); cacheWorker != nil {
			w = cacheWorker.(*Worker)
		} else {
			w = &Worker{
				pool: p,
				// https://github.com/valyala/fasthttp/blob/master/workerpool.go#L147-L149
				task: make(chan Task, 1),
			}
		}
		w.run()
		return w
	}

	var l int
	for {
		p.cond.Wait()
		l = len(p.workers) - 1
		if l >= 0 {
			break
		}
	}

	w = p.workers[l]
	p.workers[l] = nil
	p.workers = p.workers[:l]
	p.lock.Unlock()
	return w
}

// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *Pool) revertWorker(worker *Worker) bool {
	if atomic.LoadInt32(&p.release) == CLOSED {
		return false
	}
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	// Notify the invoker stuck in 'retrieveWorker()' of there is an available worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()
	return true
}
