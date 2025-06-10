package pool

import (
	"context"
)

type Pooler interface {
	Exec(argument string) error
	GetErrors() []error
	Start(ctx context.Context)
	Stop()
	AddWorkers(ctx context.Context, count int)
	DeleteWorkers(count int)
	Wait()
}
