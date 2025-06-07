package pool

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

func (pool *workerPool) worker(ctx context.Context) {
	var lastJobID uuid.UUID
	id := uuid.New()

	for {
		select {
		case j, ok := <-pool.jobs:
			if !ok {
				fmt.Println("job channel closed, returning with no jobs")
				return
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
			lastJobID = j.ID

			pool.wg.Done()
			continue

		case <-ctx.Done():
			fmt.Printf(
				"\t\tcancelled worker %s. error detail: %v, last done job %s\n",
				id,
				ctx.Err(),
				lastJobID,
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

type workerPool struct {
	wg          sync.WaitGroup
	ec          *errorCollector
	m           *monitor
	count       int
	jobs        chan job
	closed      atomic.Bool
	maxAttempts uint
}

func NewWorkerPool(wc int) Pooler {
	wp := &workerPool{
		wg:          sync.WaitGroup{},
		ec:          newErrorCollector(),
		count:       wc,
		jobs:        make(chan job, wc),
		closed:      atomic.Bool{},
		maxAttempts: 5,
	}

	wp.m = newMonitor(4, wp)
	return wp
}

func (pool *workerPool) Exec(argument string) error {
	if pool == nil {
		return errors.New("pool is unexpectedly nil")
	}

	if pool.closed.Load() {
		return errors.New("pool is already closed")
	}

	pool.wg.Add(1)
	j := createJob(argument)

	select {
	case pool.jobs <- j:
		return nil
	default:
		return errors.New("job channel already closed")
	}
}

func (pool *workerPool) Start(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(pool.count)
	go pool.ec.errorCollecting()

	ctx, cancel := context.WithCancel(ctx)
	pool.m.cancel = cancel

	for i := 0; i < pool.count; i++ {
		go pool.worker(ctx)
		wg.Done()
	}
	wg.Wait()

	go pool.wait()
}

func (pool *workerPool) Stop() {
	pool.closed.CompareAndSwap(false, true)
	pool.m.monitor()
}

func (pool *workerPool) GetErrors() []error {
	var errs []error
	for _, err := range pool.ec.errors {
		errs = append(errs, err)
	}

	return errs
}

func (pool *workerPool) wait() {
	pool.wg.Wait()
}
