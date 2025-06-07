package pool

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

type executionFunc func(ctx context.Context, workerID uuid.UUID) error

type jobDescriptor struct {
	metadata map[string]string
}

type job struct {
	ID       uuid.UUID
	Work     executionFunc
	Arg      string
	Attempts uint
	Done     chan struct{}
}

func createJob(argument string) job {
	j := job{
		ID:       uuid.New(),
		Arg:      argument,
		Attempts: 0,
		Done:     make(chan struct{}),
	}
	j.Work = func(ctx context.Context, workerID uuid.UUID) error {
		switch {
		case workerID == uuid.Nil:
			return errors.New("can not work without a worker")
		case j.Arg == "":
			return errors.New("can not work without an argument")
		}

		if randBool() {
			return errors.New("error occurred during execution job")
		}

		fmt.Printf(
			"\t\tjob «%s» was successfully finished by «%s».\n\t\tResult: %s\n",
			j.ID.String(), workerID, j.Arg)

		return nil
	}

	fmt.Println("Initialized job")

	return j
}

func (j *job) exec(ctx context.Context, workerID uuid.UUID) error {
	err := j.Work(ctx, workerID)
	if err != nil {
		return err
	}

	return nil
}

func randBool() bool {
	return rand.Intn(2) == 1
}
