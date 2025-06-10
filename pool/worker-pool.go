package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type workerPool struct {
	wg          sync.WaitGroup
	ec          *errorCollector
	m           *monitor
	jobs        chan job
	closed      atomic.Bool
	maxAttempts uint
	count       int
	workers     map[uuid.UUID]chan struct{}
}

func NewWorkerPool(wc int) Pooler {
	wp := &workerPool{
		wg:          sync.WaitGroup{},
		ec:          newErrorCollector(),
		jobs:        make(chan job, wc*2),
		closed:      atomic.Bool{},
		maxAttempts: 10,
		count:       wc,
		workers:     make(map[uuid.UUID]chan struct{}, wc),
	}

	for i := 0; i < wc; i++ {
		wp.workers[uuid.New()] = make(chan struct{})
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

	var i uint

	for i = 1; i < pool.maxAttempts; i++ {
		select {
		case pool.jobs <- j:
			return nil
		default:
			time.Sleep(time.Second * 1)
			break
		}
	}

	return errors.New("max attempts exceeded")
}

func (pool *workerPool) Start(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(pool.count)
	go pool.ec.errorCollecting()

	ctx, cancel := context.WithCancel(ctx)
	pool.m.cancel = cancel

	for id, ch := range pool.workers {
		go pool.worker(ctx, id, ch)
		wg.Done()
	}

	wg.Wait()
}

func (pool *workerPool) Stop() {
	fmt.Println("start closing pool")

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

func (pool *workerPool) AddWorkers(ctx context.Context, count int) {
	wg := sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		pool.addWorker(ctx, uuid.New(), make(chan struct{}))
		wg.Done()
	}

	wg.Wait()
}

func (pool *workerPool) addWorker(ctx context.Context, id uuid.UUID, done chan struct{}) {
	pool.workers[id] = done
	go pool.worker(ctx, id, done)
}

func (pool *workerPool) DeleteWorkers(count int) {
	var counter int

	wg := sync.WaitGroup{}
	wg.Add(count)

	for id, ch := range pool.workers {
		delete(pool.workers, id)
		ch <- struct{}{}
		counter++
		wg.Done()

		if counter == count {
			break
		}
	}
}

func (pool *workerPool) Wait() {
	pool.wg.Wait()
}
