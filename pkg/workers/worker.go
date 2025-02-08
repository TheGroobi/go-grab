package workers

import (
	"fmt"
	"sync"
)

type Task struct {
	ID int
}

func (t *Task) Process() {
	fmt.Printf("Processing task %d\n", t.ID)
}

type WorkerPool struct {
	tasksChan   chan Task
	Tasks       []Task
	wg          sync.WaitGroup
	Concurrency int
}

func (wp *WorkerPool) worker() {
	for task := range wp.tasksChan {
		task.Process()
		wp.wg.Done()
	}
}

func (wp *WorkerPool) Run() {
	wp.tasksChan = make(chan Task, len(wp.Tasks))

	for i := 0; i < wp.Concurrency; i++ {
		go wp.worker()
	}

	wp.wg.Add(len(wp.Tasks))
	for _, task := range wp.Tasks {
		wp.tasksChan <- task
	}

	close(wp.tasksChan)

	wp.wg.Wait()
}
