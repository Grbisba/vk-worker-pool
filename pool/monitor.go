package pool

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

type monitor struct {
	ticker *time.Ticker
	pool   *workerPool
	cancel context.CancelFunc
}

func newMonitor(timeIntervalSeconds int, pool *workerPool) *monitor {
	return &monitor{
		ticker: time.NewTicker(time.Duration(timeIntervalSeconds) * time.Second),
		pool:   pool,
	}
}

func (m *monitor) monitor() {
	for {
		select {
		case <-m.ticker.C:
			if len(m.pool.jobs) == 0 {
				close(m.pool.jobs)
				m.cancel()
				time.Sleep(time.Second)
				fmt.Println(runtime.NumGoroutine())
				return
			}

			if len(m.pool.ec.errCh) == 0 {
				close(m.pool.ec.errCh)
				return
			}
		}
	}
}
