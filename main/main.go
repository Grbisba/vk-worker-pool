package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Grbisba/vk-worker-pool/pool"
)

func main() {
	var (
		tasksNum   = 10000
		workersNum = 10
		wg         = sync.WaitGroup{}
	)
	wp := pool.NewWorkerPool(workersNum)

	ctx := context.Background()

	wp.Start(ctx)

	fmt.Println("Starting jobs...")

	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 1)
		wp.Exec(fmt.Sprintf("Job with number - %d", i))
	}
}
