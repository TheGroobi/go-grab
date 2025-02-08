package cmd

import (
	"fmt"

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
			tasks[i] = workers.Task{ID: i + 1}
		}

		wp := workers.WorkerPool{
			Tasks:       tasks,
			Concurrency: 5,
		}

		wp.Run()

		fmt.Println("Jaca praca skonczona")
	},
}
