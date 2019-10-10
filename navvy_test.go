package navvy

import (
	"log"
	"testing"
	"time"
)

type sleepTask struct {
	id int
}

func (st *sleepTask) Run() {
	log.Printf("Task %d start", st.id)
	time.Sleep(time.Second)
	log.Printf("Task %d finished", st.id)
}

func TestNewPool(t *testing.T) {
	p := NewPool(10)
	defer p.Release()

	start := time.Now()

	for i := 0; i < 100; i++ {
		p.Submit(&sleepTask{i})
	}

	p.Wait()
	since := time.Since(start)
	log.Printf("finish 100 tasks in %s", since)
}
