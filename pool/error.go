package pool

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type errorCollector struct {
	errCh  chan handlerError
	errors []handlerError
}

func newErrorCollector() *errorCollector {
	return &errorCollector{
		errCh: make(chan handlerError),
	}
}

func (ec *errorCollector) errorCollecting() {
	for {
		select {
		case hErr, ok := <-ec.errCh:
			if !ok {
				return
			}
			fmt.Println("\tcollector received an error")
			ec.errors = append(ec.errors, hErr)
		default:
		}
	}
}

type handlerError struct {
	workerID     uuid.UUID
	jobID        uuid.UUID
	occurredTime time.Time
	err          error
}

func handleError(workerID uuid.UUID, jobID uuid.UUID, err error) handlerError {
	return handlerError{
		workerID:     workerID,
		jobID:        jobID,
		occurredTime: time.Now(),
		err:          err,
	}
}

func (h handlerError) Error() string {
	return fmt.Sprintf(
		"Error: %v, Worker ID: %s, job ID: %s, Occurred: %s",
		h.err, h.workerID, h.jobID, h.occurredTime,
	)
}
