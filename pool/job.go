package pool

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ExecutionFn func(ctx context.Context, workerID uuid.UUID) error

type JobDescriptor struct {
	metadata map[string]string
}

type Job struct {
	ID   uuid.UUID
	Work ExecutionFn
	Arg  string
	Done chan struct{}
}

func createJob(argument string) Job {
	j := Job{
		ID:   uuid.New(),
		Arg:  argument,
		Done: make(chan struct{}),
	}
	j.Work = func(ctx context.Context, workerID uuid.UUID) error {
		switch {
		case workerID == uuid.Nil:
			return errors.New("can not work without a worker")
		case j.Arg == "":
			return errors.New("can not work without an argument")
		}

		fmt.Printf(
			"Job «%s» was successfully finished by «%s».\n Result: %s\n",
			j.ID.String(), workerID, j.Arg)

		return nil
	}

	fmt.Println("Initialized Job")

	return j
}

func (j *Job) Exec(ctx context.Context, workerID uuid.UUID) error {
	err := j.Work(ctx, workerID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
