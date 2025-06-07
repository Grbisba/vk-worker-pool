package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Grbisba/vk-worker-pool/pool"
)

func main() {
	var (
		tasksNum   = 20
		workersNum = 10
	)

	wp := pool.NewWorkerPool(workersNum)

	ctx := context.Background()

	wp.Start(ctx)

	fmt.Println("\tStarting jobs...")

	for i := 0; i < tasksNum+1; i++ {
		time.Sleep(time.Millisecond * 500)
		err := wp.Exec(fmt.Sprintf("\tJob with number - %d", i))
		if err != nil {
			fmt.Println(err)
		}
	}

	wp.Stop()

	errs := wp.GetErrors()

	for _, err := range errs {
		if err != nil {
			fmt.Println(err)
		}
	}
}
