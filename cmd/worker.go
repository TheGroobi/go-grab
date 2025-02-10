package cmd

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/TheGroobi/go-grab/pkg/workers"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Check if async works",
	Long:  `Check if async workers work`,
	Run: func(cmd *cobra.Command, args []string) {
		tasks := make([]workers.Task, 20)

		for i := 0; i < len(tasks); i++ {
			tasks[i] = workers.Task{ID: i + 1, ExecFunc: MockupWorkerTask}
		}

		wp := workers.WorkerPool{
			Tasks:       tasks,
			Concurrency: len(tasks),
		}

		wp.Run()

		fmt.Println("Jaca praca skonczona")
	},
}

func MockupWorkerTask() {
	int := rand.IntN(10)
	fmt.Printf("running task %d\n", int)

	time.Sleep(time.Second)
	fmt.Printf("task %d has completed\n", int)
}
