package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func worker(ctx context.Context, jobs <-chan Job, errCh chan<- error, done <-chan struct{}) {
	id := uuid.New()

	for {
		select {
		case <-done:
			fmt.Printf(
				"shutting down worker %s\n",
				id.String(),
			)

			return
		case job, ok := <-jobs:
			if !ok {
				fmt.Printf(
					"channel with jobs was closed: worker %s returning...\n",
					id,
				)

				return
			}

			err := job.Exec(ctx, id)

			errCh <- NewHandleError(id, job.ID, "", err)
			continue

		case <-ctx.Done():
			fmt.Printf(
				"cancelled worker %s. error detail: %v\n",
				id,
				ctx.Err(),
			)

			err := ctx.Err()
			errCh <- NewHandleError(id, uuid.Nil, "", err)
			return

		case <-time.After(time.Second * 10):
			fmt.Printf(
				"timeout worker %s. returning with no jobs\n",
				id,
			)

			return
		}
	}
}

type WorkerPool struct {
	count int
	jobs  chan Job
	errCh chan error
	done  chan struct{}
}

func NewWorkerPool(wc int) *WorkerPool {
	return &WorkerPool{
		count: wc,
		jobs:  make(chan Job, wc),
		errCh: make(chan error, wc),
		done:  make(chan struct{}),
	}
}

func (pool *WorkerPool) Exec(argument string) {
	job := createJob(argument)

	go func() {
		select {
		case pool.jobs <- job:
			return
		case <-time.After(time.Second * 2):
			return
		}
	}()
}

func (pool *WorkerPool) Start(ctx context.Context) {
	go pool.errorCollecting()
	for i := 0; i < pool.count; i++ {
		go worker(ctx, pool.jobs, pool.errCh, pool.done)
	}
}

func (pool *WorkerPool) errorCollecting() {
	var err []error
	for {
		select {
		case hErr, ok := <-pool.errCh:
			if !ok {
				return
			}

			err = append(err, hErr)
		}
	}
}

func (pool *WorkerPool) Stop() {
	close(pool.jobs)
	<-pool.done
	close(pool.errCh)
	close(pool.done)
}
