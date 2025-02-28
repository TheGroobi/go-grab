package workers

import (
	"sync"
)

type Task struct {
	ExecFunc func()
	ID       int
}

func (t *Task) Process() {
	if t.ExecFunc != nil {
		t.ExecFunc()
	}
}

type WorkerPool struct {
	tasksch     chan Task
	Tasks       []Task
	wg          sync.WaitGroup
	Concurrency int
}

func (wp *WorkerPool) worker() {
	for task := range wp.tasksch {
		task.Process()
		wp.wg.Done()
	}
}

func (wp *WorkerPool) Run() {
	wp.tasksch = make(chan Task, len(wp.Tasks))

	for i := 0; i < wp.Concurrency; i++ {
		go wp.worker()
	}

	wp.wg.Add(len(wp.Tasks))
	for _, task := range wp.Tasks {
		wp.tasksch <- task
	}

	close(wp.tasksch)

	wp.wg.Wait()
}
