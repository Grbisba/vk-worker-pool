package pool

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func (pool *workerPool) worker(ctx context.Context, id uuid.UUID, done <-chan struct{}) {
	//var attempts uint

	for {
		select {
		case <-done:
			fmt.Printf(
				"worker %s has been removed from pool\n",
				id.String(),
			)
			return
		case j, ok := <-pool.jobs:
			if !ok {
				fmt.Println("worker has been stopped")

				continue
			}

			// Work Imitation
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)+500))

			err := j.exec(ctx, id)
			if err != nil {
				if j.Attempts > pool.maxAttempts {
					pool.ec.errCh <- handleError(id, j.ID, err)
				} else if !pool.closed.Load() {
					pool.jobs <- j
					continue
				}
			}

			pool.wg.Done()
			continue

		case <-ctx.Done():
			fmt.Printf(
				"\t\tcancelled worker %s. error detail: %v\n",
				id,
				ctx.Err(),
			)

			err := ctx.Err()
			pool.ec.errCh <- handleError(id, uuid.Nil, err)
			return

		case <-time.After(time.Second * 10):
			fmt.Printf(
				"\t\ttimeout worker %s. returning with no jobs\n",
				id,
			)

			return
		}
	}
}
